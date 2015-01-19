// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package runtime

import (
	_lock "runtime/internal/lock"
	_printf "runtime/internal/printf"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

func getenv(s *byte) *byte {
	val := _schedinit.Gogetenv(_lock.Gostringnocopy(s))
	if val == "" {
		return nil
	}
	// Strings found in environment are NUL-terminated.
	return &_printf.Bytes(val)[0]
}

var _cgo_setenv unsafe.Pointer   // pointer to C function
var _cgo_unsetenv unsafe.Pointer // pointer to C function

// Update the C environment if cgo is loaded.
// Called from syscall.Setenv.
//go:linkname syscall_setenv_c syscall.setenv_c
func syscall_setenv_c(k string, v string) {
	if _cgo_setenv == nil {
		return
	}
	arg := [2]unsafe.Pointer{cstring(k), cstring(v)}
	_sched.Asmcgocall(unsafe.Pointer(_cgo_setenv), unsafe.Pointer(&arg))
}

// Update the C environment if cgo is loaded.
// Called from syscall.unsetenv.
//go:linkname syscall_unsetenv_c syscall.unsetenv_c
func syscall_unsetenv_c(k string) {
	if _cgo_unsetenv == nil {
		return
	}
	arg := [1]unsafe.Pointer{cstring(k)}
	_sched.Asmcgocall(unsafe.Pointer(_cgo_unsetenv), unsafe.Pointer(&arg))
}

func cstring(s string) unsafe.Pointer {
	p := make([]byte, len(s)+1)
	sp := (*_printf.String)(unsafe.Pointer(&s))
	_sched.Memmove(unsafe.Pointer(&p[0]), unsafe.Pointer(sp.Str), uintptr(len(s)))
	return unsafe.Pointer(&p[0])
}
