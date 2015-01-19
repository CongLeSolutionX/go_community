// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
)

// Declarations for runtime services implemented in C or assembly.

const PtrSize = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const

// Should be a built-in for unsafe.Pointer?
//go:nosplit
func Add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func Getg() *G

// memclr clears n bytes starting at ptr.
// in memclr_*.s
//go:noescape
func Memclr(ptr unsafe.Pointer, n uintptr)

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to a single xor instruction.
// USE CAREFULLY!
//go:nosplit
func Noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
func Asminit()
func Setg(gg *G)

//go:noescape
func Casuintptr(ptr *uintptr, old, new uintptr) bool

//go:noescape
func atomicstoreuintptr(ptr *uintptr, new uintptr)

//go:noescape
func Atomicloaduintptr(ptr *uintptr) uintptr
