// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"internal/syscall/windows/sysdll"
	"syscall"
	"unsafe"
)

const (
	sizeofSockaddrStorage = 0x80
)

var (
	modws2_32 = syscall.NewLazyDLL(sysdll.Add("ws2_32.dll"))

	procGetsockname = modws2_32.NewProc("getsockname")
	procGetpeername = modws2_32.NewProc("getpeername")
	procGetsockopt  = modws2_32.NewProc("getsockopt")
	procSetsockopt  = modws2_32.NewProc("setsockopt")
)

func getsockname(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	wserr, _, errno := syscall.Syscall(procGetsockname.Addr(), 3, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)))
	if wserr == uintptr(^uint32(0)) {
		if errno != 0 {
			return nil, errnoErr(errno)
		}
		return nil, syscall.EINVAL
	}
	return b[:l], nil
}

func getpeername(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	wserr, _, errno := syscall.Syscall(procGetpeername.Addr(), 3, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)))
	if wserr == uintptr(^uint32(0)) {
		if errno != 0 {
			return nil, errnoErr(errno)
		}
		return nil, syscall.EINVAL
	}
	return b[:l], nil
}

func getsockopt(s uintptr, level, name int, b []byte) (int, error) {
	l := uint32(len(b))
	wserr, _, errno := syscall.Syscall6(procGetsockopt.Addr(), 5, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0)
	if wserr == uintptr(^uint32(0)) {
		if errno != 0 {
			return 0, errnoErr(errno)
		}
		return 0, syscall.EINVAL
	}
	return int(l), nil
}

func setsockopt(s uintptr, level, name int, b []byte) error {
	wserr, _, errno := syscall.Syscall6(procSetsockopt.Addr(), 5, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)
	if wserr == uintptr(^uint32(0)) {
		if errno != 0 {
			return errnoErr(errno)
		}
		return syscall.EINVAL
	}
	return nil
}
