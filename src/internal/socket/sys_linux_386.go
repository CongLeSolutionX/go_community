// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"syscall"
	"unsafe"
)

const (
	sysGETSOCKNAME = 0x6
	sysGETPEERNAME = 0x7
	sysSETSOCKOPT  = 0xe
	sysGETSOCKOPT  = 0xf
)

func socketcall(call, a0, a1, a2, a3, a4, a5 uintptr) (uintptr, syscall.Errno)
func rawsocketcall(call, a0, a1, a2, a3, a4, a5 uintptr) (uintptr, syscall.Errno)

func getsockname(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	_, errno := rawsocketcall(sysGETSOCKNAME, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0, 0, 0)
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return b[:l], nil
}

func getpeername(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	_, errno := rawsocketcall(sysGETPEERNAME, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0, 0, 0)
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return b[:l], nil
}

func getsockopt(s uintptr, level, name int, b []byte) (int, error) {
	l := uint32(len(b))
	_, errno := socketcall(sysGETSOCKOPT, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0)
	if errno != 0 {
		return 0, errnoErr(errno)
	}
	return int(l), nil
}

func setsockopt(s uintptr, level, name int, b []byte) error {
	_, errno := socketcall(sysSETSOCKOPT, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)
	return errnoErr(errno)
}
