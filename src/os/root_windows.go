// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package os

import (
	"errors"
	"internal/filepathlite"
	"internal/stringslite"
	"internal/syscall/windows"
	"runtime"
	"syscall"
	"unsafe"
)

// rootCleanPath uses GetFullPathName to perform lexical path cleaning.
//
// On Windows, file names are lexically cleaned at the start of a file operation.
// For example, on Windows the path `a\..\b` is exactly equivalent to `b` alone,
// even if `a` does not exist or is not a directory.
//
// We use the Windows API function GetFullPathName to perform this cleaning.
// We could do this ourselves, but there are a number of subtle behaviors here,
// and deferring to the OS maintains consistency.
// (For example, `a\.\` cleans to `a\`.)
//
// GetFullPathName operates on absolute paths, and our input path is relative.
// We make the path absolute by prepending a fixed prefix of \\?\?\.
//
// We want to detect paths which use .. components to escape the root.
// We do this by ensuring the cleaned path still begins with \\?\?\.
// We catch the corner case of a path which includes a ..\?\. component
// by rejecting any input paths which contain a ?, which is not a valid character
// in a Windows filename.
func rootCleanPath(s string, prefix, suffix []string) (string, error) {
	// Reject paths which include a ? component (see above).
	if stringslite.IndexByte(s, '?') >= 0 {
		return "", errPathEscapes // TODO: better error
	}

	const fixedPrefix = `\\?\?`
	buf := []byte(fixedPrefix)
	for _, p := range prefix {
		buf = append(buf, '\\')
		buf = append(buf, []byte(p)...)
	}
	buf = append(buf, '\\')
	buf = append(buf, []byte(s)...)
	for _, p := range suffix {
		buf = append(buf, '\\')
		buf = append(buf, []byte(p)...)
	}
	s = string(buf)

	s, err := syscall.FullPath(s)
	if err != nil {
		return "", err
	}

	s, ok := stringslite.CutPrefix(s, fixedPrefix)
	if !ok {
		return "", errPathEscapes
	}
	s = stringslite.TrimPrefix(s, `\`)
	if s == "" {
		s = "."
	}

	if !filepathlite.IsLocal(s) {
		return "", errors.New("path escapes: '" + s + "'")
	}

	return s, nil
}

type sysfdType = syscall.Handle

// openRootNolog is OpenRoot.
func openRootNolog(name string) (*Root, error) {
	if name == "" {
		return nil, &PathError{Op: "open", Path: name, Err: syscall.ENOENT}
	}
	path := fixLongPath(name)
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, &PathError{Op: "open", Path: name, Err: err}
	}
	return newRoot(fd, name)
}

// newRoot returns a new Root.
// If fd is not a directory, it closes it and returns an error.
func newRoot(fd syscall.Handle, name string) (*Root, error) {
	// Check that this is a directory.
	//
	// If we get any errors here, ignore them; worst case we create a Root
	// which returns errors when you try to use it.
	var fi syscall.ByHandleFileInformation
	err := syscall.GetFileInformationByHandle(fd, &fi)
	if err == nil && fi.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY == 0 {
		syscall.CloseHandle(fd)
		return nil, &PathError{Op: "open", Path: name, Err: errors.New("not a directory")}
	}

	r := &Root{root{
		fd:   fd,
		name: name,
	}}
	runtime.SetFinalizer(&r.root, (*root).Close)
	return r, nil
}

// openRootInRoot is Root.OpenRoot.
func openRootInRoot(r *Root, name string) (*Root, error) {
	fd, err := doInRoot(r, name, rootOpenDir)
	if err != nil {
		return nil, &PathError{Op: "openat", Path: name, Err: err}
	}
	return newRoot(fd, name)
}

// rootOpenFileNolog is Root.OpenFile.
func rootOpenFileNolog(root *Root, name string, flag int, perm FileMode) (*File, error) {
	fd, err := doInRoot(root, name, func(parent syscall.Handle, name string) (syscall.Handle, error) {
		return openat(parent, name, flag, 0, perm)
	})
	if err != nil {
		return nil, &PathError{Op: "openat", Path: name, Err: err}
	}
	return newFile(fd, joinPath(root.Name(), name), "file"), nil
}

func openat(dirfd syscall.Handle, name string, flag int, options uint32, perm FileMode) (_ syscall.Handle, e1 error) {
	if len(name) == 0 {
		return syscall.InvalidHandle, syscall.ERROR_FILE_NOT_FOUND
	}

	var access uint32
	switch flag & (O_RDONLY | O_WRONLY | O_RDWR) {
	case O_RDONLY:
		access = windows.FILE_GENERIC_READ
	case O_WRONLY:
		access = windows.FILE_GENERIC_WRITE
		options |= windows.FILE_NON_DIRECTORY_FILE
	case O_RDWR:
		access = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE
		options |= windows.FILE_NON_DIRECTORY_FILE
	}
	if flag&O_CREATE != 0 {
		access |= windows.FILE_GENERIC_WRITE
	}
	if flag&O_APPEND != 0 {
		access &^= windows.FILE_WRITE_DATA
		access |= windows.FILE_APPEND_DATA
	}
	// Allow File.Stat.
	//
	// We don't need to request FILE_LIST_DIRECTORY,
	// because it's the same bit as FILE_GENERIC_READ.
	// If we're opening the file O_WRONLY,
	// we return an error if it's a directory anyway.
	access |= windows.STANDARD_RIGHTS_READ | windows.FILE_READ_ATTRIBUTES | windows.FILE_READ_EA

	if name == "." {
		name = ""
	}
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	objAttrs := &windows.OBJECT_ATTRIBUTES{
		ObjectName: objectName,
		Attributes: windows.OBJ_DONT_REPARSE,
	}
	if dirfd != syscall.InvalidHandle {
		objAttrs.RootDirectory = dirfd
	}
	objAttrs.Length = uint32(unsafe.Sizeof(*objAttrs))

	// We don't use FILE_OVERWRITE/FILE_OVERWRITE_IF, because when opening
	// a file with FILE_ATTRIBUTE_READONLY these will replace an existing
	// file with a new, read-only one.
	//
	// Instead, we ftruncate the file after opening when O_TRUNC is set.
	var disposition uint32
	switch {
	case flag&(O_CREATE|O_EXCL) == (O_CREATE | O_EXCL):
		disposition = windows.FILE_CREATE
	case flag&O_CREATE == O_CREATE:
		disposition = windows.FILE_OPEN_IF
	default:
		disposition = windows.FILE_OPEN
	}

	fileAttrs := uint32(syscall.FILE_ATTRIBUTE_NORMAL)
	if perm&syscall.S_IWRITE == 0 {
		fileAttrs = syscall.FILE_ATTRIBUTE_READONLY
	}

	var h syscall.Handle
	err = windows.NtCreateFile(
		&h,
		access,
		objAttrs,
		&windows.IO_STATUS_BLOCK{},
		nil,
		fileAttrs,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		disposition,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|options,
		0,
		0,
	)
	if err == nil {
		if flag&O_TRUNC == O_TRUNC {
			err = syscall.Ftruncate(h, 0)
			if err != nil {
				syscall.CloseHandle(h)
				return syscall.InvalidHandle, err
			}
		}
		return h, nil
	}
	s, ok := err.(windows.NTStatus)
	if !ok {
		// Shouldn't really be possible, NtCreateFile always returns NTStatus.
		return syscall.InvalidHandle, err
	}
	// Map some NTStatus errors to specific syscall errnos.
	//
	// Handle surrogate reparse point resolution (symlinks, mount points, etc.).
	// The NtCreateFile call may have failed due to encountering a symlink,
	// mount point, or other redirect. If so, look up the destination of the
	// link and return errSymlink.
	//
	// This lookup suffers from a TOCTOU race, where the file may have been
	// replaced between the NtCreateFile call above and the readReparseLinkAt
	// call below. This is fine; worst case, we return an error.
	switch s {
	case windows.STATUS_REPARSE_POINT_ENCOUNTERED:
		// The OBJ_DONT_REPARSE attribute prevented NtCreateFile
		// from resolving a surrogate, such as a symlink or mount point.
		if link, err := readReparseLinkAt(objAttrs); err == nil {
			return syscall.InvalidHandle, errSymlink(link)
		}
	case windows.STATUS_NOT_A_DIRECTORY:
		// The caller provided a FILE_DIRECTORY_FILE option and
		// the file is not a directory.
		//
		// The file might be a surrogate reparse point, so attempt to
		// resolve it as one. If that fails, map the error to ENOTDIR.
		if link, err := readReparseLinkAt(objAttrs); err == nil {
			return syscall.InvalidHandle, errSymlink(link)
		}
		return syscall.InvalidHandle, syscall.ENOTDIR
	case windows.STATUS_FILE_IS_A_DIRECTORY:
		// We tried to open a directory with write access.
		// This gets mapped to syscall.ERROR_ACCESS_DENIED by
		// NTStatus.Errno; convert it to EISDIR here instead.
		return syscall.InvalidHandle, syscall.EISDIR
	}
	return h, s.Errno()
}

func readReparseLinkAt(objAttrs *windows.OBJECT_ATTRIBUTES) (string, error) {
	var h syscall.Handle
	err := windows.NtCreateFile(
		&h,
		windows.FILE_GENERIC_READ|windows.STANDARD_RIGHTS_READ|windows.FILE_READ_ATTRIBUTES|windows.FILE_READ_EA,
		objAttrs,
		&windows.IO_STATUS_BLOCK{},
		nil,
		uint32(syscall.FILE_ATTRIBUTE_NORMAL),
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT,
		0,
		0,
	)
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(h)
	return readReparseLinkHandle(h)
}

func rootOpenDir(parent syscall.Handle, name string) (syscall.Handle, error) {
	return openat(parent, name, syscall.O_CLOEXEC|syscall.O_RDONLY, windows.FILE_DIRECTORY_FILE, 0)
}

func mkdirat(dirfd syscall.Handle, name string, perm FileMode) error {
	h, err := openat(dirfd, name, O_CREATE|O_EXCL, windows.FILE_DIRECTORY_FILE, perm)
	if err != nil {
		return err
	}
	syscall.CloseHandle(h)
	return nil
}
