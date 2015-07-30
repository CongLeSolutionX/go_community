// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package runtime

import (
	_base "runtime/internal/base"
	_print "runtime/internal/print"
	"unsafe"
)

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
	_base.Asmcgocall(unsafe.Pointer(_cgo_setenv), unsafe.Pointer(&arg))
}

// Update the C environment if cgo is loaded.
// Called from syscall.unsetenv.
//go:linkname syscall_unsetenv_c syscall.unsetenv_c
func syscall_unsetenv_c(k string) {
	if _cgo_unsetenv == nil {
		return
	}
	arg := [1]unsafe.Pointer{cstring(k)}
	_base.Asmcgocall(unsafe.Pointer(_cgo_unsetenv), unsafe.Pointer(&arg))
}

func cstring(s string) unsafe.Pointer {
	p := make([]byte, len(s)+1)
	sp := (*_print.String)(unsafe.Pointer(&s))
	_base.Memmove(unsafe.Pointer(&p[0]), unsafe.Pointer(sp.Str), uintptr(len(s)))
	return unsafe.Pointer(&p[0])
}
