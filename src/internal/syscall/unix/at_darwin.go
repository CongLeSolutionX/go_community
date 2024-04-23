// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package unix

import (
	"internal/abi"
	"syscall"
	"unsafe"
)

func libc_readlinkat_trampoline()

//go:cgo_import_dynamic libc_readlinkat readlinkat "/usr/lib/libSystem.B.dylib"

func Readlinkat(dirfd int, path string, buf []byte) (int, error) {
	p, err := syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}
	const readlinkatTrap = 473
	n, _, errno := syscall_syscall6(abi.FuncPCABI0(libc_readlinkat_trampoline),
		uintptr(dirfd),
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0,
		0)
	if errno != 0 {
		return 0, errno
	}
	return int(n), nil
}

func libc_mkdirat_trampoline()

//go:cgo_import_dynamic libc_mkdirat mkdirat "/usr/lib/libSystem.B.dylib"

func Mkdirat(dirfd int, path string, mode uint32) error {
	p, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	const mkdiratTrap = 475
	_, _, errno := syscall_syscall(abi.FuncPCABI0(libc_mkdirat_trampoline),
		uintptr(dirfd),
		uintptr(unsafe.Pointer(p)),
		uintptr(mode))
	if errno != 0 {
		return errno
	}
	return nil
}
