// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_base "runtime/internal/base"
)

//go:nosplit
func AcquireSudog() *_base.Sudog {
	// Delicate dance: the semaphore implementation calls
	// acquireSudog, acquireSudog calls new(sudog),
	// new calls malloc, malloc can call the garbage collector,
	// and the garbage collector calls the semaphore implementation
	// in stopTheWorld.
	// Break the cycle by doing acquirem/releasem around new(sudog).
	// The acquirem/releasem increments m.locks during new(sudog),
	// which keeps the garbage collector from being invoked.
	mp := _base.Acquirem()
	pp := mp.P.Ptr()
	if len(pp.Sudogcache) == 0 {
		_base.Lock(&_base.Sched.Sudoglock)
		// First, try to grab a batch from central cache.
		for len(pp.Sudogcache) < cap(pp.Sudogcache)/2 && _base.Sched.Sudogcache != nil {
			s := _base.Sched.Sudogcache
			_base.Sched.Sudogcache = s.Next
			s.Next = nil
			pp.Sudogcache = append(pp.Sudogcache, s)
		}
		_base.Unlock(&_base.Sched.Sudoglock)
		// If the central cache is empty, allocate a new one.
		if len(pp.Sudogcache) == 0 {
			pp.Sudogcache = append(pp.Sudogcache, new(_base.Sudog))
		}
	}
	n := len(pp.Sudogcache)
	s := pp.Sudogcache[n-1]
	pp.Sudogcache[n-1] = nil
	pp.Sudogcache = pp.Sudogcache[:n-1]
	if s.Elem != nil {
		_base.Throw("acquireSudog: found s.elem != nil in cache")
	}
	_base.Releasem(mp)
	return s
}

//go:nosplit
func ReleaseSudog(s *_base.Sudog) {
	if s.Elem != nil {
		_base.Throw("runtime: sudog with non-nil elem")
	}
	if s.Selectdone != nil {
		_base.Throw("runtime: sudog with non-nil selectdone")
	}
	if s.Next != nil {
		_base.Throw("runtime: sudog with non-nil next")
	}
	if s.Prev != nil {
		_base.Throw("runtime: sudog with non-nil prev")
	}
	if s.Waitlink != nil {
		_base.Throw("runtime: sudog with non-nil waitlink")
	}
	gp := _base.Getg()
	if gp.Param != nil {
		_base.Throw("runtime: releaseSudog with non-nil gp.param")
	}
	mp := _base.Acquirem() // avoid rescheduling to another P
	pp := mp.P.Ptr()
	if len(pp.Sudogcache) == cap(pp.Sudogcache) {
		// Transfer half of local cache to the central cache.
		var first, last *_base.Sudog
		for len(pp.Sudogcache) > cap(pp.Sudogcache)/2 {
			n := len(pp.Sudogcache)
			p := pp.Sudogcache[n-1]
			pp.Sudogcache[n-1] = nil
			pp.Sudogcache = pp.Sudogcache[:n-1]
			if first == nil {
				first = p
			} else {
				last.Next = p
			}
			last = p
		}
		_base.Lock(&_base.Sched.Sudoglock)
		last.Next = _base.Sched.Sudogcache
		_base.Sched.Sudogcache = first
		_base.Unlock(&_base.Sched.Sudoglock)
	}
	pp.Sudogcache = append(pp.Sudogcache, s)
	_base.Releasem(mp)
}
