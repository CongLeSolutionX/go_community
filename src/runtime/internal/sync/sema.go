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

package sync

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// Synchronous semaphore for sync.Cond.
type syncSema struct {
	lock _core.Mutex
	head *_core.Sudog
	tail *_core.Sudog
}

// syncsemacquire waits for a pairing syncsemrelease on the same semaphore s.
//go:linkname syncsemacquire sync.runtime_Syncsemacquire
func syncsemacquire(s *syncSema) {
	_lock.Lock(&s.lock)
	if s.head != nil && s.head.Nrelease > 0 {
		// Have pending release, consume it.
		var wake *_core.Sudog
		s.head.Nrelease--
		if s.head.Nrelease == 0 {
			wake = s.head
			s.head = wake.Next
			if s.head == nil {
				s.tail = nil
			}
		}
		_lock.Unlock(&s.lock)
		if wake != nil {
			wake.Next = nil
			_sched.Goready(wake.G)
		}
	} else {
		// Enqueue itself.
		w := _sem.AcquireSudog()
		w.G = _core.Getg()
		w.Nrelease = -1
		w.Next = nil
		w.Releasetime = 0
		t0 := int64(0)
		if _sem.Blockprofilerate > 0 {
			t0 = _sched.Cputicks()
			w.Releasetime = -1
		}
		if s.tail == nil {
			s.head = w
		} else {
			s.tail.Next = w
		}
		s.tail = w
		_sched.Goparkunlock(&s.lock, "semacquire")
		if t0 != 0 {
			_sem.Blockevent(int64(w.Releasetime)-t0, 2)
		}
		_sem.ReleaseSudog(w)
	}
}

// syncsemrelease waits for n pairing syncsemacquire on the same semaphore s.
//go:linkname syncsemrelease sync.runtime_Syncsemrelease
func syncsemrelease(s *syncSema, n uint32) {
	_lock.Lock(&s.lock)
	for n > 0 && s.head != nil && s.head.Nrelease < 0 {
		// Have pending acquire, satisfy it.
		wake := s.head
		s.head = wake.Next
		if s.head == nil {
			s.tail = nil
		}
		if wake.Releasetime != 0 {
			wake.Releasetime = _sched.Cputicks()
		}
		wake.Next = nil
		_sched.Goready(wake.G)
		n--
	}
	if n > 0 {
		// enqueue itself
		w := _sem.AcquireSudog()
		w.G = _core.Getg()
		w.Nrelease = int32(n)
		w.Next = nil
		w.Releasetime = 0
		if s.tail == nil {
			s.head = w
		} else {
			s.tail.Next = w
		}
		s.tail = w
		_sched.Goparkunlock(&s.lock, "semarelease")
		_sem.ReleaseSudog(w)
	} else {
		_lock.Unlock(&s.lock)
	}
}

//go:linkname syncsemcheck sync.runtime_Syncsemcheck
func syncsemcheck(sz uintptr) {
	if sz != unsafe.Sizeof(syncSema{}) {
		print("runtime: bad syncSema size - sync=", sz, " runtime=", unsafe.Sizeof(syncSema{}), "\n")
		_lock.Gothrow("bad syncSema size")
	}
}
