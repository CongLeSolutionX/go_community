// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin nacl netbsd openbsd plan9 solaris windows

package base

import (
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

func Lock(l *Mutex) {
	gp := Getg()
	if gp.M.Locks < 0 {
		Throw("runtime·lock: lock count")
	}
	gp.M.Locks++

	// Speculative grab for lock.
	if Casuintptr(&l.key, 0, Locked) {
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
		v := Atomicloaduintptr(&l.key)
		if v&Locked == 0 {
			// Unlocked. Try to lock.
			if Casuintptr(&l.key, v, v|Locked) {
				return
			}
			i = 0
		}
		if i < spin {
			Procyield(Active_spin_cnt)
		} else if i < spin+Passive_spin {
			Osyield()
		} else {
			// Someone else has it.
			// l->waitm points to a linked list of M's waiting
			// for this lock, chained through m->nextwaitm.
			// Queue this M.
			for {
				gp.M.nextwaitm = v &^ Locked
				if Casuintptr(&l.key, v, uintptr(unsafe.Pointer(gp.M))|Locked) {
					break
				}
				v = Atomicloaduintptr(&l.key)
				if v&Locked == 0 {
					continue Loop
				}
			}
			if v&Locked != 0 {
				// Queued.  Wait.
				semasleep(-1)
				i = 0
			}
		}
	}
}

//go:nowritebarrier
// We might not be holding a p in this code.
func Unlock(l *Mutex) {
	gp := Getg()
	var mp *M
	for {
		v := Atomicloaduintptr(&l.key)
		if v == Locked {
			if Casuintptr(&l.key, Locked, 0) {
				break
			}
		} else {
			// Other M's are waiting for the lock.
			// Dequeue an M.
			mp = (*M)((unsafe.Pointer)(v &^ Locked))
			if Casuintptr(&l.key, v, mp.nextwaitm) {
				// Dequeued an M.  Wake it.
				semawakeup(mp)
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

// One-time notifications.
func Noteclear(n *Note) {
	n.key = 0
}

func Notewakeup(n *Note) {
	var v uintptr
	for {
		v = Atomicloaduintptr(&n.key)
		if Casuintptr(&n.key, v, Locked) {
			break
		}
	}

	// Successfully set waitm to locked.
	// What was it before?
	switch {
	case v == 0:
		// Nothing was waiting. Done.
	case v == Locked:
		// Two notewakeups!  Not allowed.
		Throw("notewakeup - double wakeup")
	default:
		// Must be the waiting m.  Wake it up.
		semawakeup((*M)(unsafe.Pointer(v)))
	}
}

func Notesleep(n *Note) {
	gp := Getg()
	if gp != gp.M.G0 {
		Throw("notesleep not on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = Semacreate()
	}
	if !Casuintptr(&n.key, 0, uintptr(unsafe.Pointer(gp.M))) {
		// Must be locked (got wakeup).
		if n.key != Locked {
			Throw("notesleep - waitm out of sync")
		}
		return
	}
	// Queued.  Sleep.
	gp.M.blocked = true
	semasleep(-1)
	gp.M.blocked = false
}

//go:nosplit
func Notetsleep_internal(n *Note, ns int64, gp *G, deadline int64) bool {
	// gp and deadline are logically local variables, but they are written
	// as parameters so that the stack space they require is charged
	// to the caller.
	// This reduces the nosplit footprint of notetsleep_internal.
	gp = Getg()

	// Register for wakeup on n->waitm.
	if !Casuintptr(&n.key, 0, uintptr(unsafe.Pointer(gp.M))) {
		// Must be locked (got wakeup).
		if n.key != Locked {
			Throw("notetsleep - waitm out of sync")
		}
		return true
	}
	if ns < 0 {
		// Queued.  Sleep.
		gp.M.blocked = true
		semasleep(-1)
		gp.M.blocked = false
		return true
	}

	deadline = Nanotime() + ns
	for {
		// Registered.  Sleep.
		gp.M.blocked = true
		if semasleep(ns) >= 0 {
			gp.M.blocked = false
			// Acquired semaphore, semawakeup unregistered us.
			// Done.
			return true
		}
		gp.M.blocked = false
		// Interrupted or timed out.  Still registered.  Semaphore not acquired.
		ns = deadline - Nanotime()
		if ns <= 0 {
			break
		}
		// Deadline hasn't arrived.  Keep sleeping.
	}

	// Deadline arrived.  Still registered.  Semaphore not acquired.
	// Want to give up and return, but have to unregister first,
	// so that any notewakeup racing with the return does not
	// try to grant us the semaphore when we don't expect it.
	for {
		v := Atomicloaduintptr(&n.key)
		switch v {
		case uintptr(unsafe.Pointer(gp.M)):
			// No wakeup yet; unregister if possible.
			if Casuintptr(&n.key, v, 0) {
				return false
			}
		case Locked:
			// Wakeup happened so semaphore is available.
			// Grab it to avoid getting out of sync.
			gp.M.blocked = true
			if semasleep(-1) < 0 {
				Throw("runtime: unable to acquire - semaphore out of sync")
			}
			gp.M.blocked = false
			return true
		default:
			Throw("runtime: unexpected waitm - semaphore out of sync")
		}
	}
}

// same as runtime·notetsleep, but called on user g (not g0)
// calls only nosplit functions between entersyscallblock/exitsyscall
func Notetsleepg(n *Note, ns int64) bool {
	gp := Getg()
	if gp == gp.M.G0 {
		Throw("notetsleepg on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = Semacreate()
	}
	entersyscallblock(0)
	ok := Notetsleep_internal(n, ns, nil, 0)
	Exitsyscall(0)
	return ok
}
