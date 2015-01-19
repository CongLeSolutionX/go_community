// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

func badsystemstack() {
	_lock.Gothrow("systemstack called from unexpected goroutine")
}

//go:linkname reflect_memmove reflect.memmove
func reflect_memmove(to, from unsafe.Pointer, n uintptr) {
	_sched.Memmove(to, from, n)
}

// exported value for testing
var hashLoad = _maps.LoadFactor

func cgocallback(fn, frame unsafe.Pointer, framesize uintptr)
func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32
func exit1(code int32)
func breakpoint()
func cgocallback_gofunc(fv *_core.Funcval, frame unsafe.Pointer, framesize uintptr)

// return0 is a stub used to return 0 from deferproc.
// It is called at the very end of deferproc to signal
// the calling Go function that it should not jump
// to deferreturn.
// in asm_*.s
func return0()

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
