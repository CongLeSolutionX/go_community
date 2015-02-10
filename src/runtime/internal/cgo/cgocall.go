// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Cgo call and callback support.
//
// To call into the C function f from Go, the cgo-generated code calls
// runtime.cgocall(_cgo_Cfunc_f, frame), where _cgo_Cfunc_f is a
// gcc-compiled function written by cgo.
//
// runtime.cgocall (below) locks g to m, calls entersyscall
// so as not to block other goroutines or the garbage collector,
// and then calls runtime.asmcgocall(_cgo_Cfunc_f, frame).
//
// runtime.asmcgocall (in asm_$GOARCH.s) switches to the m->g0 stack
// (assumed to be an operating system-allocated stack, so safe to run
// gcc-compiled code on) and calls _cgo_Cfunc_f(frame).
//
// _cgo_Cfunc_f invokes the actual C function f with arguments
// taken from the frame structure, records the results in the frame,
// and returns to runtime.asmcgocall.
//
// After it regains control, runtime.asmcgocall switches back to the
// original g (m->curg)'s stack and returns to runtime.cgocall.
//
// After it regains control, runtime.cgocall calls exitsyscall, which blocks
// until this m can run Go code without violating the $GOMAXPROCS limit,
// and then unlocks g from m.
//
// The above description skipped over the possibility of the gcc-compiled
// function f calling back into Go.  If that happens, we continue down
// the rabbit hole during the execution of f.
//
// To make it possible for gcc-compiled C code to call a Go function p.GoF,
// cgo writes a gcc-compiled function named GoF (not p.GoF, since gcc doesn't
// know about packages).  The gcc-compiled C function f calls GoF.
//
// GoF calls crosscall2(_cgoexp_GoF, frame, framesize).  Crosscall2
// (in cgo/gcc_$GOARCH.S, a gcc-compiled assembly file) is a two-argument
// adapter from the gcc function call ABI to the 6c function call ABI.
// It is called from gcc to call 6c functions.  In this case it calls
// _cgoexp_GoF(frame, framesize), still running on m->g0's stack
// and outside the $GOMAXPROCS limit.  Thus, this code cannot yet
// call arbitrary Go code directly and must be careful not to allocate
// memory or use up m->g0's stack.
//
// _cgoexp_GoF calls runtime.cgocallback(p.GoF, frame, framesize).
// (The reason for having _cgoexp_GoF instead of writing a crosscall3
// to make this call directly is that _cgoexp_GoF, because it is compiled
// with 6c instead of gcc, can refer to dotted names like
// runtime.cgocallback and p.GoF.)
//
// runtime.cgocallback (in asm_$GOARCH.s) switches from m->g0's
// stack to the original g (m->curg)'s stack, on which it calls
// runtime.cgocallbackg(p.GoF, frame, framesize).
// As part of the stack switch, runtime.cgocallback saves the current
// SP as m->g0->sched.sp, so that any use of m->g0's stack during the
// execution of the callback will be done below the existing stack frames.
// Before overwriting m->g0->sched.sp, it pushes the old value on the
// m->g0 stack, so that it can be restored later.
//
// runtime.cgocallbackg (below) is now running on a real goroutine
// stack (not an m->g0 stack).  First it calls runtime.exitsyscall, which will
// block until the $GOMAXPROCS limit allows running this goroutine.
// Once exitsyscall has returned, it is safe to do things like call the memory
// allocator or invoke the Go callback function p.GoF.  runtime.cgocallbackg
// first defers a function to unwind m->g0.sched.sp, so that if p.GoF
// panics, m->g0.sched.sp will be restored to its old value: the m->g0 stack
// and the m->curg stack will be unwound in lock step.
// Then it calls p.GoF.  Finally it pops but does not execute the deferred
// function, calls runtime.entersyscall, and returns to runtime.cgocallback.
//
// After it regains control, runtime.cgocallback switches back to
// m->g0's stack (the pointer is still in m->g0.sched.sp), restores the old
// m->g0.sched.sp value from the stack, and returns to _cgoexp_GoF.
//
// _cgoexp_GoF immediately returns to crosscall2, which restores the
// callee-save registers for gcc and returns to GoF, which returns to f.

package cgo

import (
	_core "runtime/internal/core"
	_finalize "runtime/internal/finalize"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Call from Go to C.
//go:nosplit
func cgocall(fn, arg unsafe.Pointer) {
	cgocall_errno(fn, arg)
}

//go:nosplit
func cgocall_errno(fn, arg unsafe.Pointer) int32 {
	if !_sched.Iscgo && _lock.GOOS != "solaris" && _lock.GOOS != "windows" {
		_lock.Throw("cgocall unavailable")
	}

	if fn == nil {
		_lock.Throw("cgocall nil")
	}

	if _sched.Raceenabled {
		racereleasemerge(unsafe.Pointer(&racecgosync))
	}

	// Create an extra M for callbacks on threads not created by Go on first cgo call.
	if _core.Needextram == 1 && _sched.Cas(&_core.Needextram, 1, 0) {
		_lock.Systemstack(newextram)
	}

	/*
	 * Lock g to m to ensure we stay on the same stack if we do a
	 * cgo callback. Add entry to defer stack in case of panic.
	 */
	LockOSThread()
	mp := _core.Getg().M
	mp.Ncgocall++
	mp.Ncgo++
	defer endcgo(mp)

	/*
	 * Announce we are entering a system call
	 * so that the scheduler knows to create another
	 * M to run goroutines while we are in the
	 * foreign code.
	 *
	 * The call to asmcgocall is guaranteed not to
	 * split the stack and does not allocate memory,
	 * so it is safe to call while "in a system call", outside
	 * the $GOMAXPROCS accounting.
	 */
	entersyscall(0)
	errno := asmcgocall_errno(fn, arg)
	_sched.Exitsyscall(0)

	return errno
}

//go:nosplit
func endcgo(mp *_core.M) {
	mp.Ncgo--
	if mp.Ncgo == 0 {
		// We are going back to Go and are not in a recursive
		// call.  Let the GC collect any memory allocated via
		// _cgo_allocate that is no longer referenced.
		mp.Cgomal = nil
	}

	if _sched.Raceenabled {
		_sched.Raceacquire(unsafe.Pointer(&racecgosync))
	}

	UnlockOSThread() // invalidates mp
}

// Helper functions for cgo code.

func cmalloc(n uintptr) unsafe.Pointer {
	var args struct {
		n   uint64
		ret unsafe.Pointer
	}
	args.n = uint64(n)
	cgocall(Cgo_malloc, unsafe.Pointer(&args))
	if args.ret == nil {
		_lock.Throw("C malloc failed")
	}
	return args.ret
}

func cfree(p unsafe.Pointer) {
	cgocall(Cgo_free, p)
}

// Call from C back to Go.
//go:nosplit
func cgocallbackg() {
	gp := _core.Getg()
	if gp != gp.M.Curg {
		println("runtime: bad g in cgocallback")
		_core.Exit(2)
	}

	// entersyscall saves the caller's SP to allow the GC to trace the Go
	// stack. However, since we're returning to an earlier stack frame and
	// need to pair with the entersyscall() call made by cgocall, we must
	// save syscall* and let reentersyscall restore them.
	savedsp := unsafe.Pointer(gp.Syscallsp)
	savedpc := gp.Syscallpc
	_sched.Exitsyscall(0) // coming out of cgo call
	cgocallbackg1()
	// going back to cgo call
	reentersyscall(savedpc, uintptr(savedsp))
}

func cgocallbackg1() {
	gp := _core.Getg()
	if gp.M.Needextram {
		gp.M.Needextram = false
		_lock.Systemstack(newextram)
	}

	// Add entry to defer stack in case of panic.
	restore := true
	defer unwindm(&restore)

	if _sched.Raceenabled {
		_sched.Raceacquire(unsafe.Pointer(&racecgosync))
	}

	type args struct {
		fn      *_core.Funcval
		arg     unsafe.Pointer
		argsize uintptr
	}
	var cb *args

	// Location of callback arguments depends on stack frame layout
	// and size of stack frame of cgocallback_gofunc.
	sp := gp.M.G0.Sched.Sp
	switch _lock.GOARCH {
	default:
		_lock.Throw("cgocallbackg is unimplemented on arch")
	case "arm":
		// On arm, stack frame is two words and there's a saved LR between
		// SP and the stack frame and between the stack frame and the arguments.
		cb = (*args)(unsafe.Pointer(sp + 4*_core.PtrSize))
	case "amd64":
		// On amd64, stack frame is one word, plus caller PC.
		if _lock.Framepointer_enabled {
			// In this case, there's also saved BP.
			cb = (*args)(unsafe.Pointer(sp + 3*_core.PtrSize))
			break
		}
		cb = (*args)(unsafe.Pointer(sp + 2*_core.PtrSize))
	case "386":
		// On 386, stack frame is three words, plus caller PC.
		cb = (*args)(unsafe.Pointer(sp + 4*_core.PtrSize))
	case "ppc64", "ppc64le":
		// On ppc64, stack frame is two words and there's a
		// saved LR between SP and the stack frame and between
		// the stack frame and the arguments.
		cb = (*args)(unsafe.Pointer(sp + 4*_core.PtrSize))
	}

	// Invoke callback.
	// NOTE(rsc): passing nil for argtype means that the copying of the
	// results back into cb.arg happens without any corresponding write barriers.
	// For cgo, cb.arg points into a C stack frame and therefore doesn't
	// hold any pointers that the GC can find anyway - the write barrier
	// would be a no-op.
	_finalize.Reflectcall(nil, unsafe.Pointer(cb.fn), unsafe.Pointer(cb.arg), uint32(cb.argsize), 0)

	if _sched.Raceenabled {
		racereleasemerge(unsafe.Pointer(&racecgosync))
	}

	// Do not unwind m->g0->sched.sp.
	// Our caller, cgocallback, will do that.
	restore = false
}

func unwindm(restore *bool) {
	if !*restore {
		return
	}
	// Restore sp saved by cgocallback during
	// unwind of g's stack (see comment at top of file).
	mp := _sched.Acquirem()
	sched := &mp.G0.Sched
	switch _lock.GOARCH {
	default:
		_lock.Throw("unwindm not implemented")
	case "386", "amd64":
		sched.Sp = *(*uintptr)(unsafe.Pointer(sched.Sp))
	case "arm":
		sched.Sp = *(*uintptr)(unsafe.Pointer(sched.Sp + 4))
	case "ppc64", "ppc64le":
		sched.Sp = *(*uintptr)(unsafe.Pointer(sched.Sp + 8))
	}
	_sched.Releasem(mp)
}

var racecgosync uint64 // represents possible synchronization in C code
