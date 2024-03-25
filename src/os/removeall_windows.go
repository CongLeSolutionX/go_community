// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package os

import (
	"internal/syscall/windows"
	"io"
	"syscall"
	"unsafe"
)

func removeAll(path string) error {
	parentDir, base := splitPath(path)
	parentDirW, err := syscall.UTF16PtrFromString(parentDir)
	if err != nil {
		return &PathError{Op: "removeAll", Path: path, Err: err}
	}
	parent, err := syscall.CreateFile(parentDirW,
		syscall.FILE_LIST_DIRECTORY,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT,
		0)
	if err != nil {
		if IsNotExist(err) {
			// If parent does not exist, base cannot exist. Fail silently.
			return nil
		}
		return &PathError{Op: "removeAll", Path: path, Err: err}
	}
	defer syscall.CloseHandle(parent)

	if err := removeAllFrom(parent, base); err != nil {
		if pathErr, ok := err.(*PathError); ok {
			pathErr.Path = parentDir + string(PathSeparator) + pathErr.Path
			err = pathErr
		}
		return err
	}
	return nil
}

// removeAllFrom removes all entries from the directory.
// It returns true if all entries are removed successfully.
func removeAllChildsFrom(parent syscall.Handle, base string) (bool, error) {
	h, err := openFdAt(parent, base, syscall.FILE_LIST_DIRECTORY)
	if err != nil {
		if IsNotExist(err) {
			return true, nil
		}
		return false, &PathError{Op: "openFdAt", Path: base, Err: err}
	}
	file := newFile(h, base, "file")
	defer file.Close()

	var dirinfo dirInfo
	dirinfo.init(h)
	defer dirinfo.close()

	const reqSize = 1024
	var respSize int
	var recurseErr error
	for {
		numErr := 0
		names, _, _, readErr := file.readdir(reqSize, readdirName)
		// Errors other than EOF should stop us from continuing.
		if readErr != nil && readErr != io.EOF {
			if IsNotExist(readErr) {
				return true, nil
			}
			return false, &PathError{Op: "readdir", Path: base, Err: readErr}
		}
		respSize = len(names)
		for _, name := range names {
			if err := removeAllFrom(h, name); err != nil {
				if pathErr, ok := err.(*PathError); ok {
					pathErr.Path = base + string(PathSeparator) + pathErr.Path
				}
				numErr++
				if recurseErr == nil {
					recurseErr = err
				}
			}
		}

		// If we can delete any entry, break to start new iteration.
		// Otherwise, we discard current names, get next entries and try deleting them.
		if numErr != reqSize {
			break
		}
	}

	return respSize < reqSize, recurseErr
}

func removeAllFrom(parent syscall.Handle, base string) error {
	child, err := openFdAt(parent, base, windows.DELETE|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		if IsNotExist(err) {
			return nil
		}
		return &PathError{Op: "openFdAt", Path: base, Err: err}
	}
	defer func() {
		if child != 0 {
			syscall.CloseHandle(child)
		}
	}()
	// Simple case: if removeByHandle works, we're done.
	err = removeByHandle(child)
	if err == nil || IsNotExist(err) {
		return nil
	}

	// ERROR_DIR_NOT_EMPTY means that we have a directory, and we need to
	// remove its contents.
	if err != syscall.ERROR_DIR_NOT_EMPTY {
		return &PathError{Op: "removeByHandle", Path: base, Err: err}
	}

	var d syscall.ByHandleFileInformation
	statErr := syscall.GetFileInformationByHandle(child, &d)
	if statErr != nil {
		if IsNotExist(statErr) {
			return nil
		}
		return &PathError{Op: "GetFileInformationByHandle", Path: base, Err: err}
	}

	if d.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY == 0 {
		// Not a directory; return the error from removeByHandle.
		return &PathError{Op: "removeByHandle", Path: base, Err: err}
	}

	var recurseErr error
	for {
		done, err := removeAllChildsFrom(parent, base)
		if recurseErr == nil {
			recurseErr = err
		}
		if done {
			break
		}
		// Removing files from the directory may have caused
		// the OS to reshuffle it. Simply calling Readdirnames
		// again may skip some entries. The only reliable way
		// to avoid this is to close and re-open the
		// directory. See issue go.dev/issue/20841.
	}

	// Remove the directory itself.
	unlinkError := removeByHandle(child)
	if unlinkError == nil || IsNotExist(unlinkError) {
		return nil
	}

	if recurseErr != nil {
		return recurseErr
	}
	return &PathError{Op: "removeAll", Path: base, Err: unlinkError}
}

func removeByHandle(file syscall.Handle) error {
	var du windows.FILE_DISPOSITION_INFORMATION_EX
	du.Flags = windows.FILE_DISPOSITION_DELETE | windows.FILE_DISPOSITION_POSIX_SEMANTICS | windows.FILE_DISPOSITION_IGNORE_READONLY_ATTRIBUTE
	return windows.SetFileInformationByHandle(file, windows.FileDispositionInfoEx, unsafe.Pointer(&du), uint32(unsafe.Sizeof(du)))
}

func openFdAt(parent syscall.Handle, name string, access uint32) (syscall.Handle, error) {
	namew, err := windows.NewUnicodeString(name)
	if err != nil {
		return 0, err
	}
	attrs := windows.OBJECT_ATTRIBUTES{
		RootDirectory: parent,
		ObjectName:    &namew,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	attrs.Length = uint32(unsafe.Sizeof(attrs))
	var h syscall.Handle
	err = windows.NtCreateFile(&h,
		access|syscall.SYNCHRONIZE,
		&attrs,
		&windows.IO_STATUS_BLOCK{},
		nil,
		syscall.FILE_ATTRIBUTE_NORMAL,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_FOR_BACKUP_INTENT|syscall.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
		0,
	)
	if err != nil {
		if nterr, ok := err.(windows.NTStatus); ok {
			err = nterr.Errno()
		}
		return 0, err
	}
	return h, nil
}
