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

// The mutex state consists of four flags and a pointer. The flag at bit 0,
// mutexLocked, represents the lock itself. Bit 1, mutexSleeping, is a hint that
// the pointer is non-nil. The fast paths for locking and unlocking the mutex
// are based on atomic 8-bit swap operations on the low byte; bits 2 through 7
// are unused.
//
// Bit 8, mutexSpinning, is a try-lock that grants a waiting M permission to
// spin on the state word. Most other Ms must attempt to spend their time
// sleeping to reduce traffic on the cache line. This is the "spin bit" for
// which the implementation is named. (The anti-starvation mechanism also grants
// temporary permission for an M to spin.)
//
// Bit 9, mutexStackLocked, is a try-lock that grants an unlocking M permission
// to inspect the list of waiting Ms and to pop an M off of that stack.
//
// The upper bits hold a (partial) pointer to the M that most recently went to
// sleep. The sleeping Ms form a stack linked by their mWaitList.next fields.
// Because the fast paths use an 8-bit swap on the low byte of the state word,
// we'll need to reconstruct the full M pointer from the bits we have. Most Ms
// are allocated on the heap, and have a known alignment and base offset. (The
// offset is due to mallocgc's allocation headers.) The main program thread uses
// a static M value, m0. We check for m0 specifically and add a known offset
// otherwise.

const (
	active_spin     = 4  // referenced in proc.go for sync.Mutex implementation
	active_spin_cnt = 30 // referenced in proc.go for sync.Mutex implementation
)

const (
	mutexLocked      = 0x001
	mutexSleeping    = 0x002
	mutexSpinning    = 0x100
	mutexStackLocked = 0x200
	mutexMMask       = 0x3FF
	mutexMOffset     = mallocHeaderSize // alignment of heap-allocated Ms (those other than m0)

	mutexActiveSpinCount  = 4
	mutexActiveSpinSize   = 30
	mutexPassiveSpinCount = 1

	mutexTailWakePeriod = 16
)

//go:nosplit
func key8(p *uintptr) *uint8 {
	if goarch.BigEndian {
		return &(*[8]uint8)(unsafe.Pointer(p))[goarch.PtrSize/1-1]
	}
	return &(*[8]uint8)(unsafe.Pointer(p))[0]
}

// mWaitList is part of the M struct, and holds the list of Ms that are waiting
// for a particular runtime.mutex.
//
// When an M is unable to immediately obtain a lock, it adds itself to the list
// of Ms waiting for the lock. It does that via this struct's next field,
// forming a singly-linked list with the mutex's key field pointing to the head
// of the list.
//
// On occasion, unlock2Wake will double-link the list so it can identify the M
// at the end in amortized-constant time.
type mWaitList struct {
	next muintptr // next m waiting for lock
	prev muintptr // previous m waiting for lock (an amortized hint)
	tail muintptr // final m waiting for lock (an amortized hint)
}

// lockVerifyMSize confirms that we can recreate the low bits of the M pointer.
func lockVerifyMSize() {
	size := roundupsize(unsafe.Sizeof(m{}), false) + mallocHeaderSize
	if size&mutexMMask != 0 {
		print("M structure uses sizeclass ", size, "/", hex(size), " bytes; ",
			"incompatible with mutex flag mask ", hex(mutexMMask), "\n")
		throw("runtime.m memory alignment too small for spinbit mutex")
	}
}

// mutexWaitListHead recovers a full muintptr that was missing its low bits.
// With the exception of the static m0 value, it requires allocating runtime.m
// values in a size class with a particular minimum alignment. The 2048-byte
// size class allows recovering the full muintptr value even after overwriting
// the low 11 bits with flags. We can use those 11 bits as 3 flags and an
// atomically-swapped byte.
//
//go:nosplit
func mutexWaitListHead(v uintptr) muintptr {
	if highBits := v &^ mutexMMask; highBits == 0 {
		return 0
	} else if m0bits := muintptr(unsafe.Pointer(&m0)); highBits == uintptr(m0bits)&^mutexMMask {
		return m0bits
	} else {
		return muintptr(highBits + mutexMOffset)
	}
}

// mutexPreferLowLatency reports if this mutex prefers low latency at the risk
// of performance collapse. If so, we can allow all waiting threads to spin on
// the state word rather than go to sleep.
//
// TODO: We could have the waiting Ms each spin on their own private cache line,
// especially if we can put a bound on the on-CPU time that would consume.
//
// TODO: If there's a small set of mutex values with special requirements, they
// could make use of a more specialized lock2/unlock2 implementation. Otherwise,
// we're constrained to what we can fit within a single uintptr with no
// additional storage on the M for each lock held.
//
//go:nosplit
func mutexPreferLowLatency(l *mutex) bool {
	switch l {
	default:
		return false
	case &sched.lock:
		// We often expect sched.lock to pass quickly between Ms in a way that
		// each M has unique work to do: for instance when we stop-the-world
		// (bringing each P to idle) or add new netpoller-triggered work to the
		// global run queue.
		return true
	}
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

	timer := &lockTimer{lock: l}
	timer.begin()
	// On uniprocessors, no point spinning.
	// On multiprocessors, spin for mutexActiveSpinCount attempts.
	spin := 0
	if ncpu > 1 {
		spin = mutexActiveSpinCount
	}

	var weSpin, atTail bool
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

		if weSpin || atTail || mutexPreferLowLatency(l) {
			if i < spin {
				procyield(mutexActiveSpinSize)
				v = atomic.Loaduintptr(&l.key)
				continue tryAcquire
			} else if i < spin+mutexPassiveSpinCount {
				osyield() // TODO: Consider removing this step. See https://go.dev/issue/69268
				v = atomic.Loaduintptr(&l.key)
				continue tryAcquire
			}
		}

		// Go to sleep
		for v&mutexLocked != 0 {
			// Store the current head of the list of sleeping Ms in our gp.m.mWaitList.next field
			gp.m.mWaitList.next = mutexWaitListHead(v)

			// Pack a (partial) pointer to this M with the current lock state bits
			next := (uintptr(unsafe.Pointer(gp.m)) &^ mutexMMask) | v&mutexMMask | mutexSleeping
			if weSpin { // If we were spinning, prepare to retire
				next = next &^ mutexSpinning
			}

			if atomic.Casuintptr(&l.key, v, next) {
				weSpin = false
				// We've pushed ourselves onto the stack of waiters. Wait.
				semasleep(-1)
				atTail = gp.m.mWaitList.next == 0 // we were at risk of starving
				gp.m.mWaitList.next = 0
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
		unlock2Wake(l)
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

// unlock2Wake updates the list of Ms waiting on l, waking an M if necessary.
//
//go:nowritebarrier
func unlock2Wake(l *mutex) {
	v := atomic.Loaduintptr(&l.key)

	// On occasion, seek out and wake the M at the bottom of the stack so it
	// doesn't starve.
	antiStarve := cheaprandn(mutexTailWakePeriod) == 0
	if !(antiStarve || // avoiding starvation may require a wake
		v&mutexSpinning == 0 || // no spinners means we must wake
		mutexPreferLowLatency(l)) { // prefer waiters be awake as much as possible
		return
	}

	for {
		if v&^mutexMMask == 0 || v&mutexStackLocked != 0 {
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
		next := v | mutexStackLocked
		if atomic.Casuintptr(&l.key, v, next) {
			break
		}
		v = atomic.Loaduintptr(&l.key)
	}

	// We own the mutexStackLocked flag. New Ms may push themselves onto the
	// stack concurrently, but we're now the only thread that can remove or
	// modify the Ms that are sleeping in the list.

	var committed *m // If we choose an M within the stack, we've made a promise to wake it
	for {
		flags := v & (mutexMMask &^ mutexStackLocked) // preserve low bits, but release stack lock
		head := mutexWaitListHead(v)
		fixMutexWaitList(head)

		wakem := committed
		if committed == nil {
			if v&mutexSpinning == 0 || mutexPreferLowLatency(l) {
				wakem = head.ptr()
			}
			if antiStarve {
				wakem = head.ptr().mWaitList.tail.ptr()
			}

			if wakem != nil {
				if wakem != head.ptr() {
					committed = wakem
				}
				head = removeMutexWaitList(head, wakem)
			}
		}

		next := (uintptr(head) &^ mutexMMask) | flags
		if atomic.Casuintptr(&l.key, v, next) {
			if wakem != nil {
				// Claimed an M. Wake it.
				wakem.mWaitList.clearLinks()
				semawakeup(wakem) // no use of wakem after this point; it's awake
			}
			break
		}

		v = atomic.Loaduintptr(&l.key)
	}
}

// clearLinks resets the fields related to the M's position in the list of Ms
// waiting for a mutex.
func (l *mWaitList) clearLinks() {
	l.next = 0
	l.prev = 0
	l.tail = 0
}

// verifyMutexWaitList instructs fixMutexWaitList to confirm that the mutex wait
// list invariants are intact. Operations on the list are typically
// amortized-constant; but when active, these extra checks require visiting
// every other M that is waiting for the lock.
const verifyMutexWaitList = false

// fixMutexWaitList restores the invariants of the linked list of Ms waiting for
// a particular mutex.
//
// It takes as an argument the muintptr that is stored in the mutex's key. (The
// caller is responsible for adjusting the low bits so the pointer is either
// valid or nil.)
//
// On return, the list will be doubly-linked, and the head of the list (if not
// nil) will point to an M where mWaitList.tail points to the end of the linked
// list.
//
// The caller must have exclusive access for editing elements of the list.
func fixMutexWaitList(head muintptr) {
	if head == 0 {
		return
	}
	hp := head.ptr()
	node := hp

	var tail *m
	for {
		// For amortized-constant cost, stop searching once we reach part of the
		// list that's been visited before. Identify it by the presence of a
		// tail pointer.
		if node.mWaitList.tail.ptr() != nil {
			tail = node.mWaitList.tail.ptr()
			break
		}

		next := node.mWaitList.next.ptr()
		if next == nil {
			break
		}
		next.mWaitList.prev.set(node)

		node = next
	}
	if tail == nil {
		tail = node
	}
	hp.mWaitList.tail.set(tail)

	if verifyMutexWaitList {
		var reTail *m
		for node := hp; node != nil; node = node.mWaitList.next.ptr() {
			reTail = node
		}

		if reTail != tail {
			throw("incorrect mutex wait list tail")
		}
	}
}

// removeMutexWaitList removes mp from the list of Ms waiting for a particular
// mutex. It relies on (and keeps up to date) the invariants that
// fixMutexWaitList establishes and repairs.
//
// It modifies the nodes that are to remain in the list. It returns the value to
// assign as the head of the list, with the caller responsible for ensuring that
// the (atomic, contended) head assignment worked and subsequently clearing the
// list-related fields of mp.
//
// The only change it makes to mp is to clear the tail field -- so a subsequent
// call to fixMutexWaitList will be able to re-establish the prev link from its
// next node (just in time for another removeMutexWaitList call to clear it
// again).
//
// The caller must have exclusive access for editing elements of the list.
func removeMutexWaitList(head muintptr, mp *m) muintptr {
	if head == 0 {
		return 0
	}
	hp := head.ptr()
	tail := hp.mWaitList.tail

	mp.mWaitList.tail = 0

	if head.ptr() == mp {
		// mp is the head
		if mp.mWaitList.prev.ptr() != nil {
			throw("removeMutexWaitList node at head of list, but has prev field set")
		}
		head = mp.mWaitList.next
	} else {
		// mp is not the head
		if mp.mWaitList.prev.ptr() == nil {
			throw("removeMutexWaitList node not in list (not at head, no prev pointer)")
		}
		mp.mWaitList.prev.ptr().mWaitList.next = mp.mWaitList.next
		if tail.ptr() == mp {
			// mp is the tail
			if mp.mWaitList.next.ptr() != nil {
				throw("removeMutexWaitList node at tail of list, but has next field set")
			}
			tail = mp.mWaitList.prev
		} else {
			if mp.mWaitList.next.ptr() == nil {
				throw("removeMutexWaitList node in body of list, but without next field set")
			}
			mp.mWaitList.next.ptr().mWaitList.prev = mp.mWaitList.prev
		}
	}

	if hp := head.ptr(); hp != nil {
		hp.mWaitList.prev = 0
		hp.mWaitList.tail = tail
	}
	return head
}
