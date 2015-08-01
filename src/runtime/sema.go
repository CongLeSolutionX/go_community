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

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

//go:linkname sync_runtime_Semacquire sync.runtime_Semacquire
func sync_runtime_Semacquire(addr *uint32) {
	_gc.Semacquire(addr, true)
}

//go:linkname net_runtime_Semacquire net.runtime_Semacquire
func net_runtime_Semacquire(addr *uint32) {
	_gc.Semacquire(addr, true)
}

//go:linkname sync_runtime_Semrelease sync.runtime_Semrelease
func sync_runtime_Semrelease(addr *uint32) {
	_gc.Semrelease(addr)
}

//go:linkname net_runtime_Semrelease net.runtime_Semrelease
func net_runtime_Semrelease(addr *uint32) {
	_gc.Semrelease(addr)
}

// Synchronous semaphore for sync.Cond.
type syncSema struct {
	lock _base.Mutex
	head *_base.Sudog
	tail *_base.Sudog
}

// syncsemacquire waits for a pairing syncsemrelease on the same semaphore s.
//go:linkname syncsemacquire sync.runtime_Syncsemacquire
func syncsemacquire(s *syncSema) {
	_base.Lock(&s.lock)
	if s.head != nil && s.head.Nrelease > 0 {
		// Have pending release, consume it.
		var wake *_base.Sudog
		s.head.Nrelease--
		if s.head.Nrelease == 0 {
			wake = s.head
			s.head = wake.Next
			if s.head == nil {
				s.tail = nil
			}
		}
		_base.Unlock(&s.lock)
		if wake != nil {
			wake.Next = nil
			_base.Goready(wake.G, 4)
		}
	} else {
		// Enqueue itself.
		w := _gc.AcquireSudog()
		w.G = _base.Getg()
		w.Nrelease = -1
		w.Next = nil
		w.Releasetime = 0
		t0 := int64(0)
		if _gc.Blockprofilerate > 0 {
			t0 = _base.Cputicks()
			w.Releasetime = -1
		}
		if s.tail == nil {
			s.head = w
		} else {
			s.tail.Next = w
		}
		s.tail = w
		_base.Goparkunlock(&s.lock, "semacquire", _base.TraceEvGoBlockCond, 3)
		if t0 != 0 {
			_gc.Blockevent(int64(w.Releasetime)-t0, 2)
		}
		_gc.ReleaseSudog(w)
	}
}

// syncsemrelease waits for n pairing syncsemacquire on the same semaphore s.
//go:linkname syncsemrelease sync.runtime_Syncsemrelease
func syncsemrelease(s *syncSema, n uint32) {
	_base.Lock(&s.lock)
	for n > 0 && s.head != nil && s.head.Nrelease < 0 {
		// Have pending acquire, satisfy it.
		wake := s.head
		s.head = wake.Next
		if s.head == nil {
			s.tail = nil
		}
		if wake.Releasetime != 0 {
			wake.Releasetime = _base.Cputicks()
		}
		wake.Next = nil
		_base.Goready(wake.G, 4)
		n--
	}
	if n > 0 {
		// enqueue itself
		w := _gc.AcquireSudog()
		w.G = _base.Getg()
		w.Nrelease = int32(n)
		w.Next = nil
		w.Releasetime = 0
		if s.tail == nil {
			s.head = w
		} else {
			s.tail.Next = w
		}
		s.tail = w
		_base.Goparkunlock(&s.lock, "semarelease", _base.TraceEvGoBlockCond, 3)
		_gc.ReleaseSudog(w)
	} else {
		_base.Unlock(&s.lock)
	}
}

//go:linkname syncsemcheck sync.runtime_Syncsemcheck
func syncsemcheck(sz uintptr) {
	if sz != unsafe.Sizeof(syncSema{}) {
		print("runtime: bad syncSema size - sync=", sz, " runtime=", unsafe.Sizeof(syncSema{}), "\n")
		_base.Throw("bad syncSema size")
	}
}
