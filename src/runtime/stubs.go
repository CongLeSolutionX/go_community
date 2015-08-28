// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

func badsystemstack() {
	_base.Throw("systemstack called from unexpected goroutine")
}

//go:linkname reflect_memclr reflect.memclr
func reflect_memclr(ptr unsafe.Pointer, n uintptr) {
	_base.Memclr(ptr, n)
}

//go:linkname reflect_memmove reflect.memmove
func reflect_memmove(to, from unsafe.Pointer, n uintptr) {
	_base.Memmove(to, from, n)
}

// exported value for testing
var hashLoad = loadFactor

// in asm_*.s
//go:noescape
func memeq(a, b unsafe.Pointer, size uintptr) bool

func cgocallback(fn, frame unsafe.Pointer, framesize uintptr)
func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32

//go:noescape
func jmpdefer(fv *_base.Funcval, argp uintptr)
func exit1(code int32)
func setg(gg *_base.G)
func breakpoint()

// reflectcall calls fn with a copy of the n argument bytes pointed at by arg.
// After fn returns, reflectcall copies n-retoffset result bytes
// back into arg+retoffset before returning. If copying result bytes back,
// the caller should pass the argument frame type as argtype, so that
// call can execute appropriate write barriers during the copy.
// Package reflect passes a frame type. In package runtime, there is only
// one call that copies results back, in cgocallbackg1, and it does NOT pass a
// frame type, meaning there are no write barriers invoked. See that call
// site for justification.
func reflectcall(argtype *_base.Type, fn, arg unsafe.Pointer, argsize uint32, retoffset uint32)

// Not all cgocallback_gofunc frames are actually cgocallback_gofunc,
// so not all have these arguments. Mark them uintptr so that the GC
// does not misinterpret memory when the arguments are not present.
// cgocallback_gofunc is not called from go, only from cgocallback,
// so the arguments will be found via cgocallback's pointer-declared arguments.
// See the assembly implementations for more details.
func cgocallback_gofunc(fv uintptr, frame uintptr, framesize uintptr)

// NO go:noescape annotation; see atomic_pointer.go.
func casp1(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool

// TODO: Write native implementations of int64 atomic ops (or improve
// inliner). These portable ones can't be inlined right now, so we're
// taking an extra function call hit.

func atomicstoreint64(ptr *int64, new int64) {
	_base.Atomicstore64((*uint64)(unsafe.Pointer(ptr)), uint64(new))
}

//go:noescape
func setcallerpc(argp unsafe.Pointer, pc uintptr)

func morestack()
func rt0_go()

// stackBarrier records that the stack has been unwound past a certain
// point. It is installed over a return PC on the stack. It must
// retrieve the original return PC from g.stkbuf, increment
// g.stkbufPos to record that the barrier was hit, and jump to the
// original return PC.
func stackBarrier()

// return0 is a stub used to return 0 from deferproc.
// It is called at the very end of deferproc to signal
// the calling Go function that it should not jump
// to deferreturn.
// in asm_*.s
func return0()

// in asm_*.s
// not called directly; definitions here supply type information for traceback.
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

func systemstack_switch()

func prefetcht0(addr uintptr)
func prefetcht1(addr uintptr)
func prefetcht2(addr uintptr)
