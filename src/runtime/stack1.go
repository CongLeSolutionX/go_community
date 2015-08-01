// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

func stackinit() {
	if _base.StackCacheSize&_base.PageMask != 0 {
		_base.Throw("cache size must be a multiple of page size")
	}
	for i := range _base.Stackpool {
		mSpanList_Init(&_base.Stackpool[i])
	}
	mSpanList_Init(&_gc.StackFreeQueue)
}

var maxstacksize uintptr = 1 << 20 // enough until runtime.main sets it for real

// Called from runtime·morestack when more stack is needed.
// Allocate larger stack and relocate to new stack.
// Stack growth is multiplicative, for constant amortized cost.
//
// g->atomicstatus will be Grunning or Gscanrunning upon entry.
// If the GC is trying to stop this g then it will set preemptscan to true.
func newstack() {
	thisg := _base.Getg()
	// TODO: double check all gp. shouldn't be getg().
	if thisg.M.Morebuf.G.Ptr().Stackguard0 == _base.StackFork {
		_base.Throw("stack growth after fork")
	}
	if thisg.M.Morebuf.G.Ptr() != thisg.M.Curg {
		print("runtime: newstack called from g=", thisg.M.Morebuf.G, "\n"+"\tm=", thisg.M, " m->curg=", thisg.M.Curg, " m->g0=", thisg.M.G0, " m->gsignal=", thisg.M.Gsignal, "\n")
		morebuf := thisg.M.Morebuf
		_base.Traceback(morebuf.Pc, morebuf.Sp, morebuf.Lr, morebuf.G.Ptr())
		_base.Throw("runtime: wrong goroutine in newstack")
	}
	if thisg.M.Curg.Throwsplit {
		gp := thisg.M.Curg
		// Update syscallsp, syscallpc in case traceback uses them.
		morebuf := thisg.M.Morebuf
		gp.Syscallsp = morebuf.Sp
		gp.Syscallpc = morebuf.Pc
		print("runtime: newstack sp=", _base.Hex(gp.Sched.Sp), " stack=[", _base.Hex(gp.Stack.Lo), ", ", _base.Hex(gp.Stack.Hi), "]\n",
			"\tmorebuf={pc:", _base.Hex(morebuf.Pc), " sp:", _base.Hex(morebuf.Sp), " lr:", _base.Hex(morebuf.Lr), "}\n",
			"\tsched={pc:", _base.Hex(gp.Sched.Pc), " sp:", _base.Hex(gp.Sched.Sp), " lr:", _base.Hex(gp.Sched.Lr), " ctxt:", gp.Sched.Ctxt, "}\n")

		_base.Traceback(morebuf.Pc, morebuf.Sp, morebuf.Lr, gp)
		_base.Throw("runtime: stack split at bad time")
	}

	gp := thisg.M.Curg
	morebuf := thisg.M.Morebuf
	thisg.M.Morebuf.Pc = 0
	thisg.M.Morebuf.Lr = 0
	thisg.M.Morebuf.Sp = 0
	thisg.M.Morebuf.G = 0
	rewindmorestack(&gp.Sched)

	// NOTE: stackguard0 may change underfoot, if another thread
	// is about to try to preempt gp. Read it just once and use that same
	// value now and below.
	preempt := _base.Atomicloaduintptr(&gp.Stackguard0) == _base.StackPreempt

	// Be conservative about where we preempt.
	// We are interested in preempting user Go code, not runtime code.
	// If we're holding locks, mallocing, or preemption is disabled, don't
	// preempt.
	// This check is very early in newstack so that even the status change
	// from Grunning to Gwaiting and back doesn't happen in this case.
	// That status change by itself can be viewed as a small preemption,
	// because the GC might change Gwaiting to Gscanwaiting, and then
	// this goroutine has to wait for the GC to finish before continuing.
	// If the GC is in some way dependent on this goroutine (for example,
	// it needs a lock held by the goroutine), that small preemption turns
	// into a real deadlock.
	if preempt {
		if thisg.M.Locks != 0 || thisg.M.Mallocing != 0 || thisg.M.Preemptoff != "" || thisg.M.P.Ptr().Status != _base.Prunning {
			// Let the goroutine keep running for now.
			// gp->preempt is set, so it will be preempted next time.
			gp.Stackguard0 = gp.Stack.Lo + _base.StackGuard
			_base.Gogo(&gp.Sched) // never return
		}
	}

	// The goroutine must be executing in order to call newstack,
	// so it must be Grunning (or Gscanrunning).
	_base.Casgstatus(gp, _base.Grunning, _base.Gwaiting)
	gp.Waitreason = "stack growth"

	if gp.Stack.Lo == 0 {
		_base.Throw("missing stack in newstack")
	}
	sp := gp.Sched.Sp
	if _base.Thechar == '6' || _base.Thechar == '8' {
		// The call to morestack cost a word.
		sp -= _base.PtrSize
	}
	if _base.StackDebug >= 1 || sp < gp.Stack.Lo {
		print("runtime: newstack sp=", _base.Hex(sp), " stack=[", _base.Hex(gp.Stack.Lo), ", ", _base.Hex(gp.Stack.Hi), "]\n",
			"\tmorebuf={pc:", _base.Hex(morebuf.Pc), " sp:", _base.Hex(morebuf.Sp), " lr:", _base.Hex(morebuf.Lr), "}\n",
			"\tsched={pc:", _base.Hex(gp.Sched.Pc), " sp:", _base.Hex(gp.Sched.Sp), " lr:", _base.Hex(gp.Sched.Lr), " ctxt:", gp.Sched.Ctxt, "}\n")
	}
	if sp < gp.Stack.Lo {
		print("runtime: gp=", gp, ", gp->status=", _base.Hex(_base.Readgstatus(gp)), "\n ")
		print("runtime: split stack overflow: ", _base.Hex(sp), " < ", _base.Hex(gp.Stack.Lo), "\n")
		_base.Throw("runtime: split stack overflow")
	}

	if gp.Sched.Ctxt != nil {
		// morestack wrote sched.ctxt on its way in here,
		// without a write barrier. Run the write barrier now.
		// It is not possible to be preempted between then
		// and now, so it's okay.
		_base.Writebarrierptr_nostore((*uintptr)(unsafe.Pointer(&gp.Sched.Ctxt)), uintptr(gp.Sched.Ctxt))
	}

	if preempt {
		if gp == thisg.M.G0 {
			_base.Throw("runtime: preempt g0")
		}
		if thisg.M.P == 0 && thisg.M.Locks == 0 {
			_base.Throw("runtime: g is running but p is not")
		}
		if gp.Preemptscan {
			for !_gc.Castogscanstatus(gp, _base.Gwaiting, _base.Gscanwaiting) {
				// Likely to be racing with the GC as
				// it sees a _Gwaiting and does the
				// stack scan. If so, gcworkdone will
				// be set and gcphasework will simply
				// return.
			}
			if !gp.Gcscandone {
				_gc.Scanstack(gp)
				gp.Gcscandone = true
			}
			gp.Preemptscan = false
			gp.Preempt = false
			_gc.Casfrom_Gscanstatus(gp, _base.Gscanwaiting, _base.Gwaiting)
			_base.Casgstatus(gp, _base.Gwaiting, _base.Grunning)
			gp.Stackguard0 = gp.Stack.Lo + _base.StackGuard
			_base.Gogo(&gp.Sched) // never return
		}

		// Act like goroutine called runtime.Gosched.
		_base.Casgstatus(gp, _base.Gwaiting, _base.Grunning)
		gopreempt_m(gp) // never return
	}

	// Allocate a bigger segment and move the stack.
	oldsize := int(gp.StackAlloc)
	newsize := oldsize * 2
	if uintptr(newsize) > maxstacksize {
		print("runtime: goroutine stack exceeds ", maxstacksize, "-byte limit\n")
		_base.Throw("stack overflow")
	}

	_base.Casgstatus(gp, _base.Gwaiting, _base.Gcopystack)

	// The concurrent GC will not scan the stack while we are doing the copy since
	// the gp is in a Gcopystack status.
	_gc.Copystack(gp, uintptr(newsize))
	if _base.StackDebug >= 1 {
		print("stack grow done\n")
	}
	_base.Casgstatus(gp, _base.Gcopystack, _base.Grunning)
	_base.Gogo(&gp.Sched)
}

//go:nosplit
func nilfunc() {
	*(*uint8)(nil) = 0
}

// adjust Gobuf as if it executed a call to fn
// and then did an immediate gosave.
func gostartcallfn(gobuf *_base.Gobuf, fv *_base.Funcval) {
	var fn unsafe.Pointer
	if fv != nil {
		fn = (unsafe.Pointer)(fv.Fn)
	} else {
		fn = unsafe.Pointer(_base.FuncPC(nilfunc))
	}
	gostartcall(gobuf, fn, (unsafe.Pointer)(fv))
}

//go:nosplit
func morestackc() {
	_base.Systemstack(func() {
		_base.Throw("attempt to execute C code on Go stack")
	})
}
