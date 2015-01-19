// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin nacl netbsd openbsd plan9 solaris windows

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// This implementation depends on OS-specific implementations of
//
//	uintptr runtime·semacreate(void)
//		Create a semaphore, which will be assigned to m->waitsema.
//		The zero value is treated as absence of any semaphore,
//		so be sure to return a non-zero value.
//
//	int32 runtime·semasleep(int64 ns)
//		If ns < 0, acquire m->waitsema and return 0.
//		If ns >= 0, try to acquire m->waitsema for at most ns nanoseconds.
//		Return 0 if the semaphore was acquired, -1 if interrupted or timed out.
//
//	int32 runtime·semawakeup(M *mp)
//		Wake up mp, which is or will soon be sleeping on mp->waitsema.
//
const (
	Locked uintptr = 1

	Active_spin     = 4
	Active_spin_cnt = 30
	Passive_spin    = 1
)

func Lock(l *_core.Mutex) {
	gp := _core.Getg()
	if gp.M.Locks < 0 {
		Throw("runtime·lock: lock count")
	}
	gp.M.Locks++

	// Speculative grab for lock.
	if _core.Casuintptr(&l.Key, 0, Locked) {
		return
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = Semacreate()
	}

	// On uniprocessor's, no point spinning.
	// On multiprocessors, spin for ACTIVE_SPIN attempts.
	spin := 0
	if Ncpu > 1 {
		spin = Active_spin
	}
Loop:
	for i := 0; ; i++ {
		v := _core.Atomicloaduintptr(&l.Key)
		if v&Locked == 0 {
			// Unlocked. Try to lock.
			if _core.Casuintptr(&l.Key, v, v|Locked) {
				return
			}
			i = 0
		}
		if i < spin {
			Procyield(Active_spin_cnt)
		} else if i < spin+Passive_spin {
			_core.Osyield()
		} else {
			// Someone else has it.
			// l->waitm points to a linked list of M's waiting
			// for this lock, chained through m->nextwaitm.
			// Queue this M.
			for {
				gp.M.Nextwaitm = (*_core.M)((unsafe.Pointer)(v &^ Locked))
				if _core.Casuintptr(&l.Key, v, uintptr(unsafe.Pointer(gp.M))|Locked) {
					break
				}
				v = _core.Atomicloaduintptr(&l.Key)
				if v&Locked == 0 {
					continue Loop
				}
			}
			if v&Locked != 0 {
				// Queued.  Wait.
				Semasleep(-1)
				i = 0
			}
		}
	}
}

func Unlock(l *_core.Mutex) {
	gp := _core.Getg()
	var mp *_core.M
	for {
		v := _core.Atomicloaduintptr(&l.Key)
		if v == Locked {
			if _core.Casuintptr(&l.Key, Locked, 0) {
				break
			}
		} else {
			// Other M's are waiting for the lock.
			// Dequeue an M.
			mp = (*_core.M)((unsafe.Pointer)(v &^ Locked))
			if _core.Casuintptr(&l.Key, v, uintptr(unsafe.Pointer(mp.Nextwaitm))) {
				// Dequeued an M.  Wake it.
				Semawakeup(mp)
				break
			}
		}
	}
	gp.M.Locks--
	if gp.M.Locks < 0 {
		Throw("runtime·unlock: lock count")
	}
	if gp.M.Locks == 0 && gp.Preempt { // restore the preemption request in case we've cleared it in newstack
		gp.Stackguard0 = StackPreempt
	}
}
