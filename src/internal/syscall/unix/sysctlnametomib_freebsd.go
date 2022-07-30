// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"syscall"
	"unsafe"
)

const (
	// From FreeBSD's <sys/sysctl.h>
	_CTL_MAXNAME = 24

	// Undocumented numbers from FreeBSD's lib/libc/gen/sysctlnametomib.c.
	_CTL_QUERY     = 0
	_CTL_QUERY_MIB = 3
)

// SysctlNameToMIB accepts an ASCII representation of the name, looks up the integer name vector,
// and returns the numeric representation or error.
func SysctlNameToMIB(name []byte) ([]uint32, error) {
	oid := [2]uint32{_CTL_QUERY, _CTL_QUERY_MIB}
	mib := [_CTL_MAXNAME]uint32{}
	namelen, miblen := uintptr(len(name)), uintptr(len(mib))

	_, _, err := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&oid[0])), uintptr(len(oid)), uintptr(unsafe.Pointer(&mib[0])), uintptr(unsafe.Pointer(&miblen)), uintptr(unsafe.Pointer(&name[0])), uintptr(namelen))
	if err != 0 {
		return nil, err
	}
	if miblen == 0 {
		return nil, nil
	}
	return mib[:miblen/unsafe.Sizeof(uint32(0))], nil
}
