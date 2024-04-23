// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package windows_test

import (
	"internal/syscall/windows"
	"os"
	"syscall"
	"testing"
	"unsafe"
)

func TestNTStatusString(t *testing.T) {
	const STATUS_TOO_MANY_NAMES windows.NTStatus = 0xC00000CD
	want := "The name limit for the local computer network adapter card was exceeded."
	got := STATUS_TOO_MANY_NAMES.Error()
	if want != got {
		t.Errorf("NTStatus.Error did not return an expected error string - want %q; got %q", want, got)
	}
	t.Log(got)
}

func TestNTCreateFile(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(dir+"/f", []byte("hello"))

	objectName, err := windows.NewNTUnicodeString(dir)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	oa := &windows.OBJECT_ATTRIBUTES{
		ObjectName: objectName,
	}
	if dirfd != syscall.InvalidHandle {
		oa.RootDirectory = dirfd
	}
	oa.Length = uint32(unsafe.Sizeof(*oa))

	var access uint32
	switch flag & (O_RDONLY | O_WRONLY | O_RDWR) {
	case O_RDONLY:
		access = syscall.GENERIC_READ
	case O_WRONLY:
		access = syscall.GENERIC_WRITE
	case O_RDWR:
		access = syscall.GENERIC_READ | syscall.GENERIC_WRITE
	}
	if flag&O_CREATE != 0 {
		access |= syscall.GENERIC_WRITE
	}
	if flag&O_APPEND != 0 {
		access &^= syscall.GENERIC_WRITE
		access |= syscall.FILE_APPEND_DATA
	}

	var disposition uint32
	switch {
	case flag&(O_CREATE|O_EXCL) == (O_CREATE | O_EXCL):
		disposition = windows.FILE_CREATE
	case flag&(O_CREATE|O_TRUNC) == (O_CREATE | O_TRUNC):
		disposition = windows.FILE_OVERWRITE_IF
	case flag&O_CREATE == O_CREATE:
		disposition = windows.FILE_OPEN_IF
	case flag&O_TRUNC == O_TRUNC:
		disposition = windows.FILE_OVERWRITE
	default:
		disposition = windows.FILE_OPEN
	}

	// CLOEXEC?

	var (
		h         syscall.Handle
		iosb      windows.IO_STATUS_BLOCK
		allocSize int64
	)
	err = windows.NtCreateFile(
		&h,
		access,
		oa,
		&iosb,
		&allocSize,
		syscall.FILE_ATTRIBUTE_NORMAL,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		disposition,
		windows.FILE_RANDOM_ACCESS|windows.FILE_NON_DIRECTORY_FILE|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT,
		//windows.FILE_SYNCHRONOUS_IO_NONALERT, //options|windows.FILE_OPEN_REPARSE_POINT,
		0,
		0,
	)
	if err != nil {
		println("=> ", err.Error())
	}
	return h, err
}
