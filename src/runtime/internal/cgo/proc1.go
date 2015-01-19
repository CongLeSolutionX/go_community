// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgo

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

// newextram allocates an m and puts it on the extra list.
// It is called with a working local m, so that it can do things
// like call schedlock and allocate.
func newextram() {
	// Create extra goroutine locked to extra m.
	// The goroutine is the context in which the cgo callback will run.
	// The sched.pc will never be returned to, but setting it to
	// goexit makes clear to the traceback routines where
	// the goroutine stack ends.
	mp := _sched.Allocm(nil)
	gp := _sched.Malg(4096)
	gp.Sched.Pc = _lock.FuncPC(_schedinit.Goexit) + _lock.PCQuantum
	gp.Sched.Sp = gp.Stack.Hi
	gp.Sched.Sp -= 4 * _lock.RegSize // extra space in case of reads slightly beyond frame
	gp.Sched.Lr = 0
	gp.Sched.G = gp
	gp.Syscallpc = gp.Sched.Pc
	gp.Syscallsp = gp.Sched.Sp
	// malg returns status as Gidle, change to Gsyscall before adding to allg
	// where GC will see it.
	_sched.Casgstatus(gp, _lock.Gidle, _lock.Gsyscall)
	gp.M = mp
	mp.Curg = gp
	mp.Locked = LockInternal
	mp.Lockedg = gp
	gp.Lockedm = mp
	gp.Goid = int64(_lock.Xadd64(&_core.Sched.Goidgen, 1))
	if _sched.Raceenabled {
		gp.Racectx = Racegostart(_lock.FuncPC(newextram))
	}
	// put on allg for garbage collector
	Allgadd(gp)

	// Add m to the extra list.
	mnext := _core.Lockextra(true)
	mp.Schedlink = mnext
	_core.Unlockextra(mp)
}

// The goroutine g is about to enter a system call.
// Record that it's not using the cpu anymore.
// This is called only from the go syscall library and cgocall,
// not from the low-level system calls used by the
//
// Entersyscall cannot split the stack: the gosave must
// make g->sched refer to the caller's stack segment, because
// entersyscall is going to return immediately after.
//
// Nothing entersyscall calls can split the stack either.
// We cannot safely move the stack during an active call to syscall,
// because we do not know which of the uintptr arguments are
// really pointers (back into the stack).
// In practice, this means that we make the fast path run through
// entersyscall doing no-split things, and the slow path has to use systemstack
// to run bigger things on the system stack.
//
// reentersyscall is the entry point used by cgo callbacks, where explicitly
// saved SP and PC are restored. This is needed when exitsyscall will be called
// from a function further up in the call stack than the parent, as g->syscallsp
// must always point to a valid stack frame. entersyscall below is the normal
// entry point for syscalls, which obtains the SP and PC from the caller.
//go:nosplit
func reentersyscall(pc, sp uintptr) {
	_g_ := _core.Getg()

	// Disable preemption because during this function g is in Gsyscall status,
	// but can have inconsistent g->sched, do not let GC observe it.
	_g_.M.Locks++

	// Entersyscall must not call any function that might split/grow the stack.
	// (See details in comment above.)
	// Catch calls that might, by replacing the stack guard with something that
	// will trip any stack check and leaving a flag to tell newstack to die.
	_g_.Stackguard0 = _lock.StackPreempt
	_g_.Throwsplit = true

	// Leave SP around for GC and traceback.
	_sched.Save(pc, sp)
	_g_.Syscallsp = sp
	_g_.Syscallpc = pc
	_sched.Casgstatus(_g_, _lock.Grunning, _lock.Gsyscall)
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		_lock.Systemstack(func() {
			print("entersyscall inconsistent ", _core.Hex(_g_.Syscallsp), " [", _core.Hex(_g_.Stack.Lo), ",", _core.Hex(_g_.Stack.Hi), "]\n")
			_lock.Gothrow("entersyscall")
		})
	}

	if _lock.Atomicload(&_core.Sched.Sysmonwait) != 0 { // TODO: fast atomic
		_lock.Systemstack(entersyscall_sysmon)
		_sched.Save(pc, sp)
	}

	_g_.M.Mcache = nil
	_g_.M.P.M = nil
	_lock.Atomicstore(&_g_.M.P.Status, _lock.Psyscall)
	if _core.Sched.Gcwaiting != 0 {
		_lock.Systemstack(entersyscall_gcwait)
		_sched.Save(pc, sp)
	}

	// Goroutines must not split stacks in Gsyscall status (it would corrupt g->sched).
	// We set _StackGuard to StackPreempt so that first split stack check calls morestack.
	// Morestack detects this case and throws.
	_g_.Stackguard0 = _lock.StackPreempt
	_g_.M.Locks--
}

// Standard syscall entry used by the go syscall library and normal cgo calls.
//go:nosplit
func entersyscall(dummy int32) {
	reentersyscall(_lock.Getcallerpc(unsafe.Pointer(&dummy)), _lock.Getcallersp(unsafe.Pointer(&dummy)))
}

func entersyscall_sysmon() {
	_lock.Lock(&_core.Sched.Lock)
	if _lock.Atomicload(&_core.Sched.Sysmonwait) != 0 {
		_lock.Atomicstore(&_core.Sched.Sysmonwait, 0)
		_sched.Notewakeup(&_core.Sched.Sysmonnote)
	}
	_lock.Unlock(&_core.Sched.Lock)
}

func entersyscall_gcwait() {
	_g_ := _core.Getg()

	_lock.Lock(&_core.Sched.Lock)
	if _core.Sched.Stopwait > 0 && _sched.Cas(&_g_.M.P.Status, _lock.Psyscall, _lock.Pgcstop) {
		if _core.Sched.Stopwait--; _core.Sched.Stopwait == 0 {
			_sched.Notewakeup(&_core.Sched.Stopnote)
		}
	}
	_lock.Unlock(&_core.Sched.Lock)
}

// dolockOSThread is called by LockOSThread and lockOSThread below
// after they modify m.locked. Do not allow preemption during this call,
// or else the m might be different in this function than in the caller.
//go:nosplit
func DolockOSThread() {
	_g_ := _core.Getg()
	_g_.M.Lockedg = _g_
	_g_.Lockedm = _g_.M
}

//go:nosplit
func LockOSThread() {
	_core.Getg().M.Locked += LockInternal
	DolockOSThread()
}

// dounlockOSThread is called by UnlockOSThread and unlockOSThread below
// after they update m->locked. Do not allow preemption during this call,
// or else the m might be in different in this function than in the caller.
//go:nosplit
func DounlockOSThread() {
	_g_ := _core.Getg()
	if _g_.M.Locked != 0 {
		return
	}
	_g_.M.Lockedg = nil
	_g_.Lockedm = nil
}

//go:nosplit
func UnlockOSThread() {
	_g_ := _core.Getg()
	if _g_.M.Locked < LockInternal {
		_lock.Systemstack(badunlockosthread)
	}
	_g_.M.Locked -= LockInternal
	DounlockOSThread()
}

func badunlockosthread() {
	_lock.Gothrow("runtime: internal error: misuse of lockOSThread/unlockOSThread")
}
