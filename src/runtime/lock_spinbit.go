// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || plan9 || solaris || windows) && goexperiment.spinbitmutex

package runtime

import (
	"internal/goarch"
	"internal/runtime/atomic"
	"unsafe"
)

// This implementation depends on OS-specific implementations of
//
//	func semacreate(mp *m)
//		Create a semaphore for mp, if it does not already have one.
//
//	func semasleep(ns int64) int32
//		If ns < 0, acquire m's semaphore and return 0.
//		If ns >= 0, try to acquire m's semaphore for at most ns nanoseconds.
//		Return 0 if the semaphore was acquired, -1 if interrupted or timed out.
//
//	func semawakeup(mp *m)
//		Wake up mp, which is or will soon be sleeping on its semaphore.

// The mutex state consists of four flags and a pointer. The flag at bit 0
// represents the lock itself. Bit 1 is a hint that the pointer is non-nil. The
// fast paths for locking and unlocking the mutex are based on atomic 8-bit swap
// operations on the low byte; bits 2 through 7 are unused.
//
// Bit 8 is a try-lock that grants a waiting M permission to spin on the state
// word. All other Ms must attempt to spend their time sleeping to reduce
// traffic on the cache line. This is the "spin bit" for which the
// implementation is named.
//
// Bit 9 is a try-lock that grants an unlocking M permission to inspect the list
// of waiting Ms and to pop an M off of that stack.
//
// The upper bits hold a (partial) pointer to the M that most recently went to
// sleep. The sleeping Ms form a stack linked by their nextwaitm fields. Because
// the fast paths use an 8-bit swap on the low byte of the state word, we'll
// need to reconstruct the full M pointer from the bits we have. Most Ms are
// allocated on the heap, and have a known alignment and base offset. (The
// offset is due to mallocgc's allocation headers.) The main program thread uses
// a static M value, m0. We check for m0 specifically and add a known offset
// otherwise.

const (
	mutexLocked    = 0x001
	mutexSleeping  = 0x002
	mutexSpinning  = 0x100
	mutexStackLock = 0x200
	mutexMMask     = 0x3FF
	mutexMOffset   = 0x008 // alignment of heap-allocated Ms (those other than m0)

	active_spin      = 4
	active_spin_cnt  = 30
	passiveSpinCount = 1
)

//go:nosplit
func key8(p *uintptr) *uint8 {
	if goarch.BigEndian {
		return &(*[8]uint8)(unsafe.Pointer(p))[goarch.PtrSize/1-1]
	}
	return &(*[8]uint8)(unsafe.Pointer(p))[0]
}

func mutexContended(l *mutex) bool {
	return atomic.Loaduintptr(&l.key) > mutexLocked
}

func lock(l *mutex) {
	lockWithRank(l, getLockRank(l))
}

func lock2(l *mutex) {
	gp := getg()
	if gp.m.locks < 0 {
		throw("runtime·lock: lock count")
	}
	gp.m.locks++

	k8 := key8(&l.key)

	var v8 uint8
	// Speculative grab for lock.
	v8 = atomic.Xchg8(k8, mutexLocked)
	if v8&mutexLocked == 0 {
		if v8&mutexSleeping != 0 {
			atomic.Or8(k8, mutexSleeping)
		}
		return
	}
	semacreate(gp.m)

	// Verify that we can recreate the low bits of the M pointer
	if offset := uint16(uintptr(unsafe.Pointer(gp.m))) & mutexMMask; (gp.m != &m0) && (offset != mutexMOffset) {
		print("mp.id=", gp.m.id, " mp=", hex(uintptr(unsafe.Pointer(gp.m))), " mask=", hex(mutexMMask), "\n")
		throw("runtime.m memory alignment too small for spinbit mutex")
	}

	timer := &lockTimer{lock: l}
	timer.begin()
	// On uniprocessor's, no point spinning.
	// On multiprocessors, spin for ACTIVE_SPIN attempts.
	spin := 0
	if ncpu > 1 {
		spin = active_spin
	}

	var weSpin bool
	v := atomic.Loaduintptr(&l.key)
tryAcquire:
	for i := 0; ; i++ {
		for v&mutexLocked == 0 {
			if weSpin {
				next := (v &^ mutexMMask) | (v & (mutexMMask &^ mutexSpinning)) | mutexLocked
				if next&^mutexMMask != 0 {
					next |= mutexSleeping
				}
				if atomic.Casuintptr(&l.key, v, next) {
					timer.end()
					return
				}
			} else {
				prev8 := atomic.Xchg8(k8, mutexLocked|mutexSleeping)
				if prev8&mutexLocked == 0 {
					timer.end()
					return
				}
			}
			v = atomic.Loaduintptr(&l.key)
		}

		if !weSpin && v&mutexSpinning == 0 && atomic.Casuintptr(&l.key, v, v|mutexSpinning) {
			v |= mutexSpinning
			weSpin = true
		}

		if weSpin {
			if i < spin {
				procyield(active_spin_cnt)
				v = atomic.Loaduintptr(&l.key)
				continue tryAcquire
			} else if i < spin+passiveSpinCount {
				osyield() // TODO: Consider removing this step. See https://go.dev/issue/69268
				v = atomic.Loaduintptr(&l.key)
				continue tryAcquire
			}
		}

		// Go to sleep
		for v&mutexLocked != 0 {
			// Store the current head of the list of sleeping Ms in our gp.m.nextwaitm field
			if nextMBase := v &^ mutexMMask; nextMBase == uintptr(unsafe.Pointer(&m0))&^mutexMMask {
				gp.m.nextwaitm.set(&m0)
			} else if nextMBase != 0 {
				gp.m.nextwaitm = muintptr(nextMBase + mutexMOffset)
			} else {
				gp.m.nextwaitm = 0
			}

			// Pack a (partial) pointer to this M with the current lock state bits
			next := (uintptr(unsafe.Pointer(gp.m)) &^ mutexMMask) | v&mutexMMask | mutexSleeping
			if weSpin { // If we were spinning, prepare to retire
				next = next &^ mutexSpinning
			}

			if atomic.Casuintptr(&l.key, v, next) {
				weSpin = false
				// We've pushed ourselves onto the stack of waiters. Wait.
				semasleep(-1)
				gp.m.nextwaitm = 0
				i = 0
				v = atomic.Loaduintptr(&l.key)
				continue tryAcquire
			}
			v = atomic.Loaduintptr(&l.key)
		}
	}
}

func unlock(l *mutex) {
	unlockWithRank(l)
}

// We might not be holding a p in this code.
//
//go:nowritebarrier
func unlock2(l *mutex) {
	gp := getg()

	prev8 := atomic.Xchg8(key8(&l.key), 0)
	if prev8&mutexLocked == 0 {
		throw("unlock of unlocked lock")
	}

	if prev8&mutexSleeping != 0 {
		unlock2wake(l)
	}

	gp.m.mLockProfile.recordUnlock(l)
	gp.m.locks--
	if gp.m.locks < 0 {
		throw("runtime·unlock: lock count")
	}
	if gp.m.locks == 0 && gp.preempt { // restore the preemption request in case we've cleared it in newstack
		gp.stackguard0 = stackPreempt
	}
}

// unlock2wake updates the list of Ms waiting on l, waking an M if necessary.
//
//go:nowritebarrier
func unlock2wake(l *mutex) {
	v := atomic.Loaduintptr(&l.key)
	for {
		if v&^mutexMMask == 0 || v&mutexStackLock != 0 {
			// No waiting Ms means nothing to do.
			//
			// If the stack lock is unavailable, its owner would make the same
			// wake decisions that we would, so there's nothing for us to do.
			//
			// Although: This thread may have a different call stack, which
			// would result in a different entry in the mutex contention profile
			// (upon completion of go.dev/issue/66999). That could lead to weird
			// results if a slow critical section ends but another thread
			// quickly takes the lock, finishes its own critical section,
			// releases the lock, and then grabs the stack lock. That quick
			// thread would then take credit (blame) for the delay that this
			// slow thread caused. The alternative is to have more expensive
			// atomic operations (a CAS) on the critical path of unlock2.
			return
		}
		// Other M's are waiting for the lock.
		// Obtain the stack lock, and pop off an M.
		next := v | mutexStackLock
		if atomic.Casuintptr(&l.key, v, next) {
			break
		}
		v = atomic.Loaduintptr(&l.key)
	}

	// We own the mutexStackLock flag. New Ms may push themselves onto the stack
	// concurrently, but we're now the only thread that can remove or modify the
	// Ms that are sleeping in the list.
	//
	// We still go to the trouble of obtaining the stack lock so we can inspect
	// the list of sleeping Ms. First, we may need to take responsibility in the
	// mutex contention profile for the amount of delay we caused. Second, even
	// when there's a designated spinning thread we may wish to wake another
	// thread anyway so sleeping threads don't starve. The M at the bottom of
	// the stack holds necessary information for each of those, and we can only
	// access it if we're sure no other thread will wake that M.
	//
	// Neither of those are done yet, but this structure is prepared for them.

	for {
		headM := v &^ mutexMMask
		flags := v & (mutexMMask &^ mutexStackLock) // preserve low bits, but release stack lock

		var mp *m
		if v&mutexSpinning == 0 {
			if nextMBase := v &^ mutexMMask; nextMBase == uintptr(unsafe.Pointer(&m0))&^mutexMMask {
				mp = &m0
			} else {
				mp = muintptr(nextMBase + mutexMOffset).ptr()
			}
			headM = uintptr(mp.nextwaitm) &^ mutexMMask
		}

		next := headM | flags
		if atomic.Casuintptr(&l.key, v, next) {
			if mp != nil {
				// Popped an M. Wake it.
				semawakeup(mp)
			}
			break
		}

		v = atomic.Loaduintptr(&l.key)
	}
}
