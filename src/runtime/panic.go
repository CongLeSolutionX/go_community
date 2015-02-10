// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

var indexError = error(_sched.ErrorString("index out of range"))

func panicindex() {
	panic(indexError)
}

var sliceError = error(_sched.ErrorString("slice bounds out of range"))

func panicslice() {
	panic(sliceError)
}

func throwreturn() {
	_lock.Throw("no return at end of a typed function - compiler is broken")
}

func throwinit() {
	_lock.Throw("recursive call during initialization - linker skew")
}

// Create a new deferred function fn with siz bytes of arguments.
// The compiler turns a defer statement into a call to this.
//go:nosplit
func deferproc(siz int32, fn *_core.Funcval) { // arguments of fn follow fn
	if _core.Getg().M.Curg != _core.Getg() {
		// go code on the system stack can't defer
		_lock.Throw("defer on system stack")
	}

	// the arguments of fn are in a perilous state.  The stack map
	// for deferproc does not describe them.  So we can't let garbage
	// collection or stack copying trigger until we've copied them out
	// to somewhere safe.  The memmove below does that.
	// Until the copy completes, we can only call nosplit routines.
	sp := _lock.Getcallersp(unsafe.Pointer(&siz))
	argp := uintptr(unsafe.Pointer(&fn)) + unsafe.Sizeof(fn)
	callerpc := _lock.Getcallerpc(unsafe.Pointer(&siz))

	_lock.Systemstack(func() {
		d := newdefer(siz)
		if d.Panic != nil {
			_lock.Throw("deferproc: d.panic != nil after newdefer")
		}
		d.Fn = fn
		d.Pc = callerpc
		d.Sp = sp
		_sched.Memmove(_core.Add(unsafe.Pointer(d), unsafe.Sizeof(*d)), unsafe.Pointer(argp), uintptr(siz))
	})

	// deferproc returns 0 normally.
	// a deferred func that stops a panic
	// makes the deferproc return 1.
	// the code the compiler generates always
	// checks the return value and jumps to the
	// end of the function if deferproc returns != 0.
	return0()
	// No code can go here - the C return register has
	// been set and must not be clobbered.
}

// Allocate a Defer, usually using per-P pool.
// Each defer must be released with freedefer.
// Note: runs on g0 stack
func newdefer(siz int32) *_core.Defer {
	var d *_core.Defer
	sc := _schedinit.Deferclass(uintptr(siz))
	mp := _sched.Acquirem()
	if sc < uintptr(len(_core.P{}.Deferpool)) {
		pp := mp.P
		d = pp.Deferpool[sc]
		if d != nil {
			pp.Deferpool[sc] = d.Link
		}
	}
	if d == nil {
		// Allocate new defer+args.
		total := _schedinit.Roundupsize(_schedinit.Totaldefersize(uintptr(siz)))
		d = (*_core.Defer)(_maps.Mallocgc(total, _maps.DeferType, 0))
	}
	d.Siz = siz
	if _lock.Mheap_.Shadow_enabled {
		// This memory will be written directly, with no write barrier,
		// and then scanned like stacks during collection.
		// Unlike real stacks, it is from heap spans, so mark the
		// shadow as explicitly unusable.
		p := _gc.DeferArgs(d)
		for i := uintptr(0); i+_core.PtrSize <= uintptr(siz); i += _core.PtrSize {
			_sched.Writebarrierptr_noshadow((*uintptr)(_core.Add(p, i)))
		}
	}
	gp := mp.Curg
	d.Link = gp.Defer
	gp.Defer = d
	_sched.Releasem(mp)
	return d
}

// The implementation of the predeclared function recover.
// Cannot split the stack because it needs to reliably
// find the stack segment of its caller.
//
// TODO(rsc): Once we commit to CopyStackAlways,
// this doesn't need to be nosplit.
//go:nosplit
func gorecover(argp uintptr) interface{} {
	// Must be in a function running as part of a deferred call during the panic.
	// Must be called from the topmost function of the call
	// (the function used in the defer statement).
	// p.argp is the argument pointer of that topmost deferred function call.
	// Compare against argp reported by caller.
	// If they match, the caller is the one who can recover.
	gp := _core.Getg()
	p := gp.Panic
	if p != nil && !p.Recovered && argp == uintptr(p.Argp) {
		p.Recovered = true
		return p.Arg
	}
	return nil
}
