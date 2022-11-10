// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"internal/abi"
	"runtime"
	"unsafe"
)

const (
	RTLD_NOW    = 2
	RTLD_GLOBAL = 8
	PATH_MAX    = 1024
)

//go:cgo_import_dynamic libc_dlopen dlopen "/usr/lib/libSystem.B.dylib"
func libc_dlopen_trampoline()

//go:cgo_import_dynamic libc_dlsym dlsym "/usr/lib/libSystem.B.dylib"
func libc_dlsym_trampoline()

//go:cgo_import_dynamic libc_dlerror dlerror "/usr/lib/libSystem.B.dylib"
func libc_dlerror_trampoline()

func Dlopen(path *byte, flags uintptr, err **byte) uintptr {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, _, _ := syscall_syscall(abi.FuncPCABI0(libc_dlopen_trampoline), uintptr(unsafe.Pointer(path)), flags, 0)
	if h == 0 {
		errstr, _, _ := syscall_syscall(abi.FuncPCABI0(libc_dlerror_trampoline), 0, 0, 0)
		*err = (*byte)(unsafe.Pointer(errstr))
	}
	return h
}

func Dlsym(h uintptr, name *byte, err **byte) unsafe.Pointer {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	r, _, _ := syscall_syscall(abi.FuncPCABI0(libc_dlsym_trampoline), h, uintptr(unsafe.Pointer(name)), 0)
	if r == 0 {
		errstr, _, _ := syscall_syscall(abi.FuncPCABI0(libc_dlerror_trampoline), 0, 0, 0)
		*err = (*byte)(unsafe.Pointer(errstr))
	}
	return unsafe.Pointer(r)
}

//go:cgo_import_dynamic libc_realpath realpath "/usr/lib/libSystem.B.dylib"
func libc_realpath_trampoline()

func Realpath(old, new *byte) (*byte, error) {
	ptr, _, errno := syscall_syscallPtr(abi.FuncPCABI0(libc_realpath_trampoline),
		uintptr(unsafe.Pointer(old)),
		uintptr(unsafe.Pointer(new)), 0)
	if errno != 0 {
		return nil, errno
	}
	return (*byte)(unsafe.Pointer(ptr)), nil
}
