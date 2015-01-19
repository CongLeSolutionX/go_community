// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
)

// Similar to stoptheworld but best-effort and can be called several times.
// There is no reverse operation, used during crashing.
// This function must not lock any mutexes.
func freezetheworld() {
	if Gomaxprocs == 1 {
		return
	}
	// stopwait and preemption requests can be lost
	// due to races with concurrently executing threads,
	// so try several times
	for i := 0; i < 5; i++ {
		// this should tell the scheduler to not start any new goroutines
		_core.Sched.Stopwait = 0x7fffffff
		Atomicstore(&_core.Sched.Gcwaiting, 1)
		// this should stop running goroutines
		if !Preemptall() {
			break // no running goroutines
		}
		_core.Usleep(1000)
	}
	// to be sure
	_core.Usleep(1000)
	Preemptall()
	_core.Usleep(1000)
}

// All reads and writes of g's status go through readgstatus, casgstatus
// castogscanstatus, casfrom_Gscanstatus.
//go:nosplit
func Readgstatus(gp *_core.G) uint32 {
	return Atomicload(&gp.Atomicstatus)
}

// Tell all goroutines that they have been preempted and they should stop.
// This function is purely best-effort.  It can fail to inform a goroutine if a
// processor just started running it.
// No locks need to be held.
// Returns true if preemption request was issued to at least one goroutine.
func Preemptall() bool {
	res := false
	for i := int32(0); i < Gomaxprocs; i++ {
		_p_ := Allp[i]
		if _p_ == nil || _p_.Status != Prunning {
			continue
		}
		if Preemptone(_p_) {
			res = true
		}
	}
	return res
}

// Tell the goroutine running on processor P to stop.
// This function is purely best-effort.  It can incorrectly fail to inform the
// goroutine.  It can send inform the wrong goroutine.  Even if it informs the
// correct goroutine, that goroutine might ignore the request if it is
// simultaneously executing newstack.
// No lock needs to be held.
// Returns true if preemption request was issued.
// The actual preemption will happen at some point in the future
// and will be indicated by the gp->status no longer being
// Grunning
func Preemptone(_p_ *_core.P) bool {
	mp := _p_.M
	if mp == nil || mp == _core.Getg().M {
		return false
	}
	gp := mp.Curg
	if gp == nil || gp == mp.G0 {
		return false
	}

	gp.Preempt = true

	// Every call in a go routine checks for stack overflow by
	// comparing the current stack pointer to gp->stackguard0.
	// Setting gp->stackguard0 to StackPreempt folds
	// preemption into the normal stack overflow check.
	gp.Stackguard0 = StackPreempt
	return true
}

var starttime int64

func Schedtrace(detailed bool) {
	now := Nanotime()
	if starttime == 0 {
		starttime = now
	}

	Lock(&_core.Sched.Lock)
	print("SCHED ", (now-starttime)/1e6, "ms: gomaxprocs=", Gomaxprocs, " idleprocs=", _core.Sched.Npidle, " threads=", _core.Sched.Mcount, " spinningthreads=", _core.Sched.Nmspinning, " idlethreads=", _core.Sched.Nmidle, " runqueue=", _core.Sched.Runqsize)
	if detailed {
		print(" gcwaiting=", _core.Sched.Gcwaiting, " nmidlelocked=", _core.Sched.Nmidlelocked, " stopwait=", _core.Sched.Stopwait, " sysmonwait=", _core.Sched.Sysmonwait, "\n")
	}
	// We must be careful while reading data from P's, M's and G's.
	// Even if we hold schedlock, most data can be changed concurrently.
	// E.g. (p->m ? p->m->id : -1) can crash if p->m changes from non-nil to nil.
	for i := int32(0); i < Gomaxprocs; i++ {
		_p_ := Allp[i]
		if _p_ == nil {
			continue
		}
		mp := _p_.M
		h := Atomicload(&_p_.Runqhead)
		t := Atomicload(&_p_.Runqtail)
		if detailed {
			id := int32(-1)
			if mp != nil {
				id = mp.Id
			}
			print("  P", i, ": status=", _p_.Status, " schedtick=", _p_.Schedtick, " syscalltick=", _p_.Syscalltick, " m=", id, " runqsize=", t-h, " gfreecnt=", _p_.Gfreecnt, "\n")
		} else {
			// In non-detailed mode format lengths of per-P run queues as:
			// [len1 len2 len3 len4]
			print(" ")
			if i == 0 {
				print("[")
			}
			print(t - h)
			if i == Gomaxprocs-1 {
				print("]\n")
			}
		}
	}

	if !detailed {
		Unlock(&_core.Sched.Lock)
		return
	}

	for mp := Allm; mp != nil; mp = mp.Alllink {
		_p_ := mp.P
		gp := mp.Curg
		lockedg := mp.Lockedg
		id1 := int32(-1)
		if _p_ != nil {
			id1 = _p_.Id
		}
		id2 := int64(-1)
		if gp != nil {
			id2 = gp.Goid
		}
		id3 := int64(-1)
		if lockedg != nil {
			id3 = lockedg.Goid
		}
		print("  M", mp.Id, ": p=", id1, " curg=", id2, " mallocing=", mp.Mallocing, " throwing=", mp.Throwing, " gcing=", mp.Gcing, ""+" locks=", mp.Locks, " dying=", mp.Dying, " helpgc=", mp.Helpgc, " spinning=", mp.Spinning, " blocked=", _core.Getg().M.Blocked, " lockedg=", id3, "\n")
	}

	Lock(&Allglock)
	for gi := 0; gi < len(Allgs); gi++ {
		gp := Allgs[gi]
		mp := gp.M
		lockedm := gp.Lockedm
		id1 := int32(-1)
		if mp != nil {
			id1 = mp.Id
		}
		id2 := int32(-1)
		if lockedm != nil {
			id2 = lockedm.Id
		}
		print("  G", gp.Goid, ": status=", Readgstatus(gp), "(", gp.Waitreason, ") m=", id1, " lockedm=", id2, "\n")
	}
	Unlock(&Allglock)
	Unlock(&_core.Sched.Lock)
}
