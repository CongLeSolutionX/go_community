// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sem

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

//go:nosplit
func AcquireSudog() *_core.Sudog {
	c := Gomcache()
	s := c.Sudogcache
	if s != nil {
		if s.Elem != nil {
			_lock.Gothrow("acquireSudog: found s.elem != nil in cache")
		}
		c.Sudogcache = s.Next
		s.Next = nil
		return s
	}

	// Delicate dance: the semaphore implementation calls
	// acquireSudog, acquireSudog calls new(sudog),
	// new calls malloc, malloc can call the garbage collector,
	// and the garbage collector calls the semaphore implementation
	// in stoptheworld.
	// Break the cycle by doing acquirem/releasem around new(sudog).
	// The acquirem/releasem increments m.locks during new(sudog),
	// which keeps the garbage collector from being invoked.
	mp := _sched.Acquirem()
	p := new(_core.Sudog)
	if p.Elem != nil {
		_lock.Gothrow("acquireSudog: found p.elem != nil after new")
	}
	_sched.Releasem(mp)
	return p
}

//go:nosplit
func ReleaseSudog(s *_core.Sudog) {
	if s.Elem != nil {
		_lock.Gothrow("runtime: sudog with non-nil elem")
	}
	if s.Selectdone != nil {
		_lock.Gothrow("runtime: sudog with non-nil selectdone")
	}
	if s.Next != nil {
		_lock.Gothrow("runtime: sudog with non-nil next")
	}
	if s.Prev != nil {
		_lock.Gothrow("runtime: sudog with non-nil prev")
	}
	if s.Waitlink != nil {
		_lock.Gothrow("runtime: sudog with non-nil waitlink")
	}
	gp := _core.Getg()
	if gp.Param != nil {
		_lock.Gothrow("runtime: releaseSudog with non-nil gp.param")
	}
	c := Gomcache()
	s.Next = c.Sudogcache
	c.Sudogcache = s
}
