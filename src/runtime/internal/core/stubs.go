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

// in asm_*.s
// not called directly; definitions here supply type information for traceback.
func call16(fn, arg unsafe.Pointer, n, retoffset uint32)
func call32(fn, arg unsafe.Pointer, n, retoffset uint32)
func call64(fn, arg unsafe.Pointer, n, retoffset uint32)
func call128(fn, arg unsafe.Pointer, n, retoffset uint32)
func call256(fn, arg unsafe.Pointer, n, retoffset uint32)
func call512(fn, arg unsafe.Pointer, n, retoffset uint32)
func call1024(fn, arg unsafe.Pointer, n, retoffset uint32)
func call2048(fn, arg unsafe.Pointer, n, retoffset uint32)
func call4096(fn, arg unsafe.Pointer, n, retoffset uint32)
func call8192(fn, arg unsafe.Pointer, n, retoffset uint32)
func call16384(fn, arg unsafe.Pointer, n, retoffset uint32)
func call32768(fn, arg unsafe.Pointer, n, retoffset uint32)
func call65536(fn, arg unsafe.Pointer, n, retoffset uint32)
func call131072(fn, arg unsafe.Pointer, n, retoffset uint32)
func call262144(fn, arg unsafe.Pointer, n, retoffset uint32)
func call524288(fn, arg unsafe.Pointer, n, retoffset uint32)
func call1048576(fn, arg unsafe.Pointer, n, retoffset uint32)
func call2097152(fn, arg unsafe.Pointer, n, retoffset uint32)
func call4194304(fn, arg unsafe.Pointer, n, retoffset uint32)
func call8388608(fn, arg unsafe.Pointer, n, retoffset uint32)
func call16777216(fn, arg unsafe.Pointer, n, retoffset uint32)
func call33554432(fn, arg unsafe.Pointer, n, retoffset uint32)
func call67108864(fn, arg unsafe.Pointer, n, retoffset uint32)
func call134217728(fn, arg unsafe.Pointer, n, retoffset uint32)
func call268435456(fn, arg unsafe.Pointer, n, retoffset uint32)
func call536870912(fn, arg unsafe.Pointer, n, retoffset uint32)
func call1073741824(fn, arg unsafe.Pointer, n, retoffset uint32)
