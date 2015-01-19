// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_stackwb "runtime/internal/stackwb"
	"unsafe"
)

var maxstacksize uintptr = 1 << 20 // enough until runtime.main sets it for real

// Called from runtime·morestack when more stack is needed.
// Allocate larger stack and relocate to new stack.
// Stack growth is multiplicative, for constant amortized cost.
//
// g->atomicstatus will be Grunning or Gscanrunning upon entry.
// If the GC is trying to stop this g then it will set preemptscan to true.
func newstack() {
	thisg := _core.Getg()
	// TODO: double check all gp. shouldn't be getg().
	if thisg.M.Morebuf.G.Stackguard0 == _lock.StackFork {
		_lock.Throw("stack growth after fork")
	}
	if thisg.M.Morebuf.G != thisg.M.Curg {
		print("runtime: newstack called from g=", thisg.M.Morebuf.G, "\n"+"\tm=", thisg.M, " m->curg=", thisg.M.Curg, " m->g0=", thisg.M.G0, " m->gsignal=", thisg.M.Gsignal, "\n")
		morebuf := thisg.M.Morebuf
		_lock.Traceback(morebuf.Pc, morebuf.Sp, morebuf.Lr, morebuf.G)
		_lock.Throw("runtime: wrong goroutine in newstack")
	}
	if thisg.M.Curg.Throwsplit {
		gp := thisg.M.Curg
		// Update syscallsp, syscallpc in case traceback uses them.
		morebuf := thisg.M.Morebuf
		gp.Syscallsp = morebuf.Sp
		gp.Syscallpc = morebuf.Pc
		print("runtime: newstack sp=", _core.Hex(gp.Sched.Sp), " stack=[", _core.Hex(gp.Stack.Lo), ", ", _core.Hex(gp.Stack.Hi), "]\n",
			"\tmorebuf={pc:", _core.Hex(morebuf.Pc), " sp:", _core.Hex(morebuf.Sp), " lr:", _core.Hex(morebuf.Lr), "}\n",
			"\tsched={pc:", _core.Hex(gp.Sched.Pc), " sp:", _core.Hex(gp.Sched.Sp), " lr:", _core.Hex(gp.Sched.Lr), " ctxt:", gp.Sched.Ctxt, "}\n")
		_lock.Throw("runtime: stack split at bad time")
	}

	// The goroutine must be executing in order to call newstack,
	// so it must be Grunning or Gscanrunning.

	gp := thisg.M.Curg
	morebuf := thisg.M.Morebuf
	thisg.M.Morebuf.Pc = 0
	thisg.M.Morebuf.Lr = 0
	thisg.M.Morebuf.Sp = 0
	thisg.M.Morebuf.G = nil

	_sched.Casgstatus(gp, _lock.Grunning, _lock.Gwaiting)
	gp.Waitreason = "stack growth"

	rewindmorestack(&gp.Sched)

	if gp.Stack.Lo == 0 {
		_lock.Throw("missing stack in newstack")
	}
	sp := gp.Sched.Sp
	if _lock.Thechar == '6' || _lock.Thechar == '8' {
		// The call to morestack cost a word.
		sp -= _core.PtrSize
	}
	if _sched.StackDebug >= 1 || sp < gp.Stack.Lo {
		print("runtime: newstack sp=", _core.Hex(sp), " stack=[", _core.Hex(gp.Stack.Lo), ", ", _core.Hex(gp.Stack.Hi), "]\n",
			"\tmorebuf={pc:", _core.Hex(morebuf.Pc), " sp:", _core.Hex(morebuf.Sp), " lr:", _core.Hex(morebuf.Lr), "}\n",
			"\tsched={pc:", _core.Hex(gp.Sched.Pc), " sp:", _core.Hex(gp.Sched.Sp), " lr:", _core.Hex(gp.Sched.Lr), " ctxt:", gp.Sched.Ctxt, "}\n")
	}
	if sp < gp.Stack.Lo {
		print("runtime: gp=", gp, ", gp->status=", _core.Hex(_lock.Readgstatus(gp)), "\n ")
		print("runtime: split stack overflow: ", _core.Hex(sp), " < ", _core.Hex(gp.Stack.Lo), "\n")
		_lock.Throw("runtime: split stack overflow")
	}

	if gp.Sched.Ctxt != nil {
		// morestack wrote sched.ctxt on its way in here,
		// without a write barrier. Run the write barrier now.
		// It is not possible to be preempted between then
		// and now, so it's okay.
		_stackwb.Writebarrierptr_nostore((*uintptr)(unsafe.Pointer(&gp.Sched.Ctxt)), uintptr(gp.Sched.Ctxt))
	}

	if gp.Stackguard0 == _lock.StackPreempt {
		if gp == thisg.M.G0 {
			_lock.Throw("runtime: preempt g0")
		}
		if thisg.M.P == nil && thisg.M.Locks == 0 {
			_lock.Throw("runtime: g is running but p is not")
		}
		if gp.Preemptscan {
			for !_gc.Castogscanstatus(gp, _lock.Gwaiting, _lock.Gscanwaiting) {
				// Likely to be racing with the GC as it sees a _Gwaiting and does the stack scan.
				// If so this stack will be scanned twice which does not change correctness.
			}
			_gc.Gcphasework(gp)
			_gc.Casfrom_Gscanstatus(gp, _lock.Gscanwaiting, _lock.Gwaiting)
			_sched.Casgstatus(gp, _lock.Gwaiting, _lock.Grunning)
			gp.Stackguard0 = gp.Stack.Lo + _core.StackGuard
			gp.Preempt = false
			gp.Preemptscan = false // Tells the GC premption was successful.
			_sched.Gogo(&gp.Sched) // never return
		}

		// Be conservative about where we preempt.
		// We are interested in preempting user Go code, not runtime code.
		if thisg.M.Locks != 0 || thisg.M.Mallocing != 0 || thisg.M.Gcing != 0 || thisg.M.P.Status != _lock.Prunning {
			// Let the goroutine keep running for now.
			// gp->preempt is set, so it will be preempted next time.
			gp.Stackguard0 = gp.Stack.Lo + _core.StackGuard
			_sched.Casgstatus(gp, _lock.Gwaiting, _lock.Grunning)
			_sched.Gogo(&gp.Sched) // never return
		}

		// Act like goroutine called runtime.Gosched.
		_sched.Casgstatus(gp, _lock.Gwaiting, _lock.Grunning)
		_sched.Gosched_m(gp) // never return
	}

	// Allocate a bigger segment and move the stack.
	oldsize := int(gp.Stack.Hi - gp.Stack.Lo)
	newsize := oldsize * 2
	if uintptr(newsize) > maxstacksize {
		print("runtime: goroutine stack exceeds ", maxstacksize, "-byte limit\n")
		_lock.Throw("stack overflow")
	}

	_sched.Casgstatus(gp, _lock.Gwaiting, _lock.Gcopystack)

	// The concurrent GC will not scan the stack while we are doing the copy since
	// the gp is in a Gcopystack status.
	_gc.Copystack(gp, uintptr(newsize))
	if _sched.StackDebug >= 1 {
		print("stack grow done\n")
	}
	_sched.Casgstatus(gp, _lock.Gcopystack, _lock.Grunning)
	_sched.Gogo(&gp.Sched)
}

//go:nosplit
func nilfunc() {
	*(*uint8)(nil) = 0
}

// adjust Gobuf as if it executed a call to fn
// and then did an immediate gosave.
func gostartcallfn(gobuf *_core.Gobuf, fv *_core.Funcval) {
	var fn unsafe.Pointer
	if fv != nil {
		fn = (unsafe.Pointer)(fv.Fn)
	} else {
		fn = unsafe.Pointer(_lock.FuncPC(nilfunc))
	}
	gostartcall(gobuf, fn, (unsafe.Pointer)(fv))
}

//go:nosplit
func morestackc() {
	_lock.Systemstack(func() {
		_lock.Throw("attempt to execute C code on Go stack")
	})
}
