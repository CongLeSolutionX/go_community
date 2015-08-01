// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Semaphore implementation exposed to Go.
// Intended use is provide a sleep and wakeup
// primitive that can be used in the contended case
// of other synchronization primitives.
// Thus it targets the same goal as Linux's futex,
// but it has much simpler semantics.
//
// That is, don't think of these as semaphores.
// Think of them as a way to implement sleep and wakeup
// such that every sleep is paired with a single wakeup,
// even if, due to races, the wakeup happens before the sleep.
//
// See Mullender and Cox, ``Semaphores in Plan 9,''
// http://swtch.com/semaphore.pdf

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

// Asynchronous semaphore for sync.Mutex.

type semaRoot struct {
	lock  _base.Mutex
	head  *_base.Sudog
	tail  *_base.Sudog
	nwait uint32 // Number of waiters. Read w/o the lock.
}

// Prime to not correlate with any user patterns.
const semTabSize = 251

var semtable [semTabSize]struct {
	root semaRoot
	pad  [_base.CacheLineSize - unsafe.Sizeof(semaRoot{})]byte
}

// Called from runtime.
func Semacquire(addr *uint32, profile bool) {
	gp := _base.Getg()
	if gp != gp.M.Curg {
		_base.Throw("semacquire not on the G stack")
	}

	// Easy case.
	if cansemacquire(addr) {
		return
	}

	// Harder case:
	//	increment waiter count
	//	try cansemacquire one more time, return if succeeded
	//	enqueue itself as a waiter
	//	sleep
	//	(waiter descriptor is dequeued by signaler)
	s := AcquireSudog()
	root := semroot(addr)
	t0 := int64(0)
	s.Releasetime = 0
	if profile && Blockprofilerate > 0 {
		t0 = _base.Cputicks()
		s.Releasetime = -1
	}
	for {
		_base.Lock(&root.lock)
		// Add ourselves to nwait to disable "easy case" in semrelease.
		_base.Xadd(&root.nwait, 1)
		// Check cansemacquire to avoid missed wakeup.
		if cansemacquire(addr) {
			_base.Xadd(&root.nwait, -1)
			_base.Unlock(&root.lock)
			break
		}
		// Any semrelease after the cansemacquire knows we're waiting
		// (we set nwait above), so go to sleep.
		root.queue(addr, s)
		_base.Goparkunlock(&root.lock, "semacquire", _base.TraceEvGoBlockSync, 4)
		if cansemacquire(addr) {
			break
		}
	}
	if s.Releasetime > 0 {
		Blockevent(int64(s.Releasetime)-t0, 3)
	}
	ReleaseSudog(s)
}

func Semrelease(addr *uint32) {
	root := semroot(addr)
	_base.Xadd(addr, 1)

	// Easy case: no waiters?
	// This check must happen after the xadd, to avoid a missed wakeup
	// (see loop in semacquire).
	if _base.Atomicload(&root.nwait) == 0 {
		return
	}

	// Harder case: search for a waiter and wake it.
	_base.Lock(&root.lock)
	if _base.Atomicload(&root.nwait) == 0 {
		// The count is already consumed by another goroutine,
		// so no need to wake up another goroutine.
		_base.Unlock(&root.lock)
		return
	}
	s := root.head
	for ; s != nil; s = s.Next {
		if s.Elem == unsafe.Pointer(addr) {
			_base.Xadd(&root.nwait, -1)
			root.dequeue(s)
			break
		}
	}
	_base.Unlock(&root.lock)
	if s != nil {
		if s.Releasetime != 0 {
			s.Releasetime = _base.Cputicks()
		}
		_base.Goready(s.G, 4)
	}
}

func semroot(addr *uint32) *semaRoot {
	return &semtable[(uintptr(unsafe.Pointer(addr))>>3)%semTabSize].root
}

func cansemacquire(addr *uint32) bool {
	for {
		v := _base.Atomicload(addr)
		if v == 0 {
			return false
		}
		if _base.Cas(addr, v, v-1) {
			return true
		}
	}
}

func (root *semaRoot) queue(addr *uint32, s *_base.Sudog) {
	s.G = _base.Getg()
	s.Elem = unsafe.Pointer(addr)
	s.Next = nil
	s.Prev = root.tail
	if root.tail != nil {
		root.tail.Next = s
	} else {
		root.head = s
	}
	root.tail = s
}

func (root *semaRoot) dequeue(s *_base.Sudog) {
	if s.Next != nil {
		s.Next.Prev = s.Prev
	} else {
		root.tail = s.Prev
	}
	if s.Prev != nil {
		s.Prev.Next = s.Next
	} else {
		root.head = s.Next
	}
	s.Elem = nil
	s.Next = nil
	s.Prev = nil
}
