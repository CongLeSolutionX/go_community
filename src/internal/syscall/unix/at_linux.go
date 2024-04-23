// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"syscall"
	"unsafe"
)

// OpenHow is a struct open_how defined in <linux/openat2.h>.
type OpenHow struct {
	Flags   uint64
	Mode    uint64
	Resolve uint64
}

const (
	RESOLVE_NO_XDEV       = 0x01
	RESOLVE_NO_MAGICLINKS = 0x02
	RESOLVE_NO_SYMLINKS   = 0x04
	RESOLVE_BENEATH       = 0x08
	RESOLVE_IN_ROOT       = 0x10
	RESOLVE_CACHED        = 0x20
)

func Openat2(dirfd int, path string, how OpenHow) (int, error) {
	p, err := syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}

	fd, _, errno := syscall.Syscall6(openat2Trap, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&how)), unsafe.Sizeof(how), 0, 0)
	if errno != 0 {
		return 0, errno
	}

	return int(fd), nil
}
