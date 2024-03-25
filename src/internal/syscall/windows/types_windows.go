// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package windows

import (
	"syscall"
)

// NTStatus corresponds with NTSTATUS, error values returned by ntdll.dll and
// other native functions.
type NTStatus uint32

func (s NTStatus) Errno() syscall.Errno {
	return rtlNtStatusToDosError(s)
}

func (s NTStatus) Error() string {
	return s.Errno().Error()
}

// Socket related.
const (
	TCP_KEEPIDLE  = 0x03
	TCP_KEEPCNT   = 0x10
	TCP_KEEPINTVL = 0x11
)

const (
	FILE_DISPOSITION_DELETE                    = 0x00000001
	FILE_DISPOSITION_POSIX_SEMANTICS           = 0x00000002
	FILE_DISPOSITION_IGNORE_READONLY_ATTRIBUTE = 0x00000010
)

type FILE_DISPOSITION_INFORMATION_EX struct {
	Flags uint32
}

type OBJECT_ATTRIBUTES struct {
	Length             uint32
	RootDirectory      syscall.Handle
	ObjectName         *UnicodeString
	Attributes         uint32
	SecurityDescriptor uintptr
	SecurityQoS        uintptr
}

// Values for the Attributes member of OBJECT_ATTRIBUTES.
const (
	OBJ_CASE_INSENSITIVE = 0x00000040
)

type UnicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

func NewUnicodeString(s string) (UnicodeString, error) {
	namew, err := syscall.UTF16FromString(s)
	if err != nil {
		return UnicodeString{}, err
	}
	n := uint16(len(namew) * 2)
	return UnicodeString{
		Length:        n - 2, // subtract 2 bytes for the NULL terminator
		MaximumLength: n,
		Buffer:        &namew[0],
	}, nil
}

type IO_STATUS_BLOCK struct {
	Status      NTStatus
	Information uintptr
}

// CreateDisposition flags for NtCreateFile and NtCreateNamedPipeFile.
const (
	FILE_OPEN = 0x00000001
)

// CreateOptions flags for NtCreateFile and NtCreateNamedPipeFile.
const (
	FILE_SYNCHRONOUS_IO_NONALERT = 0x00000020
	FILE_OPEN_FOR_BACKUP_INTENT  = 0x00004000
)

type ACCESS_MASK uint32

// Constants for type ACCESS_MASK
const (
	FILE_READ_ATTRIBUTES = 0x00000080
	FILE_WRITE_EA        = 0x00000010
	DELETE               = 0x00010000
)
