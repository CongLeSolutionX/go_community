// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package defers

import (
	_core "runtime/internal/core"
	_finalize "runtime/internal/finalize"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

func init() {
	var x interface{}
	x = (*_core.Defer)(nil)
	_maps.DeferType = (*(**_gc.Ptrtype)(unsafe.Pointer(&x))).Elem
}

// Free the given defer.
// The defer cannot be used after this call.
//go:nosplit
func freedefer(d *_core.Defer) {
	if d.Panic != nil {
		freedeferpanic()
	}
	if d.Fn != nil {
		freedeferfn()
	}
	if _lock.Mheap_.Shadow_enabled {
		// Undo the marking in newdefer.
		_lock.Systemstack(func() {
			_maps.Clearshadow(uintptr(_gc.DeferArgs(d)), uintptr(d.Siz))
		})
	}
	sc := _schedinit.Deferclass(uintptr(d.Siz))
	if sc < uintptr(len(_core.P{}.Deferpool)) {
		mp := _sched.Acquirem()
		pp := mp.P
		*d = _core.Defer{}
		d.Link = pp.Deferpool[sc]
		pp.Deferpool[sc] = d
		_sched.Releasem(mp)
	}
}

// Separate function so that it can split stack.
// Windows otherwise runs out of stack space.
func freedeferpanic() {
	// _panic must be cleared before d is unlinked from gp.
	_lock.Throw("freedefer with d._panic != nil")
}

func freedeferfn() {
	// fn must be cleared before d is unlinked from gp.
	_lock.Throw("freedefer with d.fn != nil")
}

// Run a deferred function if there is one.
// The compiler inserts a call to this at the end of any
// function which calls defer.
// If there is a deferred function, this will call runtime·jmpdefer,
// which will jump to the deferred function such that it appears
// to have been called by the caller of deferreturn at the point
// just before deferreturn was called.  The effect is that deferreturn
// is called again and again until there are no more deferred functions.
// Cannot split the stack because we reuse the caller's frame to
// call the deferred function.

// The single argument isn't actually used - it just has its address
// taken so it can be matched against pending defers.
//go:nosplit
func deferreturn(arg0 uintptr) {
	gp := _core.Getg()
	d := gp.Defer
	if d == nil {
		return
	}
	sp := _lock.Getcallersp(unsafe.Pointer(&arg0))
	if d.Sp != sp {
		return
	}

	// Moving arguments around.
	// Do not allow preemption here, because the garbage collector
	// won't know the form of the arguments until the jmpdefer can
	// flip the PC over to fn.
	mp := _sched.Acquirem()
	_sched.Memmove(unsafe.Pointer(&arg0), _gc.DeferArgs(d), uintptr(d.Siz))
	fn := d.Fn
	d.Fn = nil
	gp.Defer = d.Link
	freedefer(d)
	_sched.Releasem(mp)
	_schedinit.Jmpdefer(fn, uintptr(unsafe.Pointer(&arg0)))
}

// Goexit terminates the goroutine that calls it.  No other goroutine is affected.
// Goexit runs all deferred calls before terminating the goroutine.  Because Goexit
// is not panic, however, any recover calls in those deferred functions will return nil.
//
// Calling Goexit from the main goroutine terminates that goroutine
// without func main returning. Since func main has not returned,
// the program continues execution of other goroutines.
// If all other goroutines exit, the program crashes.
func Goexit() {
	// Run all deferred functions for the current goroutine.
	// This code is similar to gopanic, see that implementation
	// for detailed comments.
	gp := _core.Getg()
	for {
		d := gp.Defer
		if d == nil {
			break
		}
		if d.Started {
			if d.Panic != nil {
				d.Panic.Aborted = true
				d.Panic = nil
			}
			d.Fn = nil
			gp.Defer = d.Link
			freedefer(d)
			continue
		}
		d.Started = true
		_finalize.Reflectcall(nil, unsafe.Pointer(d.Fn), _gc.DeferArgs(d), uint32(d.Siz), uint32(d.Siz))
		if gp.Defer != d {
			_lock.Throw("bad defer entry in Goexit")
		}
		d.Panic = nil
		d.Fn = nil
		gp.Defer = d.Link
		freedefer(d)
		// Note: we ignore recovers here because Goexit isn't a panic
	}
	_schedinit.Goexit()
}

// Print all currently active panics.  Used when crashing.
func printpanics(p *_core.Panic) {
	if p.Link != nil {
		printpanics(p.Link)
		print("\t")
	}
	print("panic: ")
	printany(p.Arg)
	if p.Recovered {
		print(" [recovered]")
	}
	print("\n")
}

// The implementation of the predeclared function panic.
func gopanic(e interface{}) {
	gp := _core.Getg()
	if gp.M.Curg != gp {
		print("panic: ")
		printany(e)
		print("\n")
		_lock.Throw("panic on system stack")
	}

	// m.softfloat is set during software floating point.
	// It increments m.locks to avoid preemption.
	// We moved the memory loads out, so there shouldn't be
	// any reason for it to panic anymore.
	if gp.M.Softfloat != 0 {
		gp.M.Locks--
		gp.M.Softfloat = 0
		_lock.Throw("panic during softfloat")
	}
	if gp.M.Mallocing != 0 {
		print("panic: ")
		printany(e)
		print("\n")
		_lock.Throw("panic during malloc")
	}
	if gp.M.Preemptoff != "" {
		print("panic: ")
		printany(e)
		print("\n")
		print("preempt off reason: ")
		print(gp.M.Preemptoff)
		print("\n")
		_lock.Throw("panic during preemptoff")
	}
	if gp.M.Locks != 0 {
		print("panic: ")
		printany(e)
		print("\n")
		_lock.Throw("panic holding locks")
	}

	var p _core.Panic
	p.Arg = e
	p.Link = gp.Panic
	gp.Panic = (*_core.Panic)(_core.Noescape(unsafe.Pointer(&p)))

	for {
		d := gp.Defer
		if d == nil {
			break
		}

		// If defer was started by earlier panic or Goexit (and, since we're back here, that triggered a new panic),
		// take defer off list. The earlier panic or Goexit will not continue running.
		if d.Started {
			if d.Panic != nil {
				d.Panic.Aborted = true
			}
			d.Panic = nil
			d.Fn = nil
			gp.Defer = d.Link
			freedefer(d)
			continue
		}

		// Mark defer as started, but keep on list, so that traceback
		// can find and update the defer's argument frame if stack growth
		// or a garbage collection hapens before reflectcall starts executing d.fn.
		d.Started = true

		// Record the panic that is running the defer.
		// If there is a new panic during the deferred call, that panic
		// will find d in the list and will mark d._panic (this panic) aborted.
		d.Panic = (*_core.Panic)(_core.Noescape((unsafe.Pointer)(&p)))

		p.Argp = unsafe.Pointer(getargp(0))
		_finalize.Reflectcall(nil, unsafe.Pointer(d.Fn), _gc.DeferArgs(d), uint32(d.Siz), uint32(d.Siz))
		p.Argp = nil

		// reflectcall did not panic. Remove d.
		if gp.Defer != d {
			_lock.Throw("bad defer entry in panic")
		}
		d.Panic = nil
		d.Fn = nil
		gp.Defer = d.Link

		// trigger shrinkage to test stack copy.  See stack_test.go:TestStackPanic
		//GC()

		pc := d.Pc
		sp := unsafe.Pointer(d.Sp) // must be pointer so it gets adjusted during stack copy
		freedefer(d)
		if p.Recovered {
			gp.Panic = p.Link
			// Aborted panics are marked but remain on the g.panic list.
			// Remove them from the list.
			for gp.Panic != nil && gp.Panic.Aborted {
				gp.Panic = gp.Panic.Link
			}
			if gp.Panic == nil { // must be done with signal
				gp.Sig = 0
			}
			// Pass information about recovering frame to recovery.
			gp.Sigcode0 = uintptr(sp)
			gp.Sigcode1 = pc
			_sched.Mcall(recovery)
			_lock.Throw("recovery failed") // mcall should not return
		}
	}

	// ran out of deferred calls - old-school panic now
	_lock.Startpanic()
	printpanics(gp.Panic)
	_lock.Dopanic(0) // should not return
	*(*int)(nil) = 0 // not reached
}

// getargp returns the location where the caller
// writes outgoing function call arguments.
//go:nosplit
func getargp(x int) uintptr {
	// x is an argument mainly so that we can return its address.
	// However, we need to make the function complex enough
	// that it won't be inlined. We always pass x = 0, so this code
	// does nothing other than keep the compiler from thinking
	// the function is simple enough to inline.
	if x > 0 {
		return _lock.Getcallersp(unsafe.Pointer(&x)) * 0
	}
	return uintptr(_core.Noescape(unsafe.Pointer(&x)))
}
