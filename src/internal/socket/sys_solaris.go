// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"syscall"
	"unsafe"
)

//go:cgo_import_dynamic libc_getsockname getsockname "libsocket.so"
//go:cgo_import_dynamic libc_getpeername getpeername "libsocket.so"
//go:cgo_import_dynamic libc___xnet_getsockopt __xnet_getsockopt "libsocket.so"
//go:cgo_import_dynamic libc_setsockopt setsockopt "libsocket.so"

//go:linkname procGetsockname libc_getsockname
//go:linkname procGetpeername libc_getpeername
//go:linkname procGetsockopt libc___xnet_getsockopt
//go:linkname procSetsockopt libc_setsockopt

var (
	procGetsockname uintptr
	procGetpeername uintptr
	procGetsockopt  uintptr
	procSetsockopt  uintptr
)

func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (uintptr, uintptr, syscall.Errno)
func rawSysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (uintptr, uintptr, syscall.Errno)

func getsockname(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	_, _, errno := rawSysvicall6(uintptr(unsafe.Pointer(&procGetsockname)), 3, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0, 0, 0)
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return b[:l], nil
}

func getpeername(s uintptr) ([]byte, error) {
	b := make([]byte, sizeofSockaddrStorage)
	l := uint32(sizeofSockaddrStorage)
	_, _, errno := rawSysvicall6(uintptr(unsafe.Pointer(&procGetpeername)), 3, s, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0, 0, 0)
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return b, nil
}

func getsockopt(s uintptr, level, name int, b []byte) (int, error) {
	l := uint32(len(b))
	_, _, errno := sysvicall6(uintptr(unsafe.Pointer(&procGetsockopt)), 5, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&l)), 0)
	if errno != 0 {
		return 0, errnoErr(errno)
	}
	return int(l), nil
}

func setsockopt(s uintptr, level, name int, b []byte) error {
	_, _, errno := sysvicall6(uintptr(unsafe.Pointer(&procSetsockopt)), 5, s, uintptr(level), uintptr(name), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)
	return errnoErr(errno)
}
