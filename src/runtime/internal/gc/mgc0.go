// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

var Poolcleanup func()

func clearpools() {
	// clear sync.Pools
	if Poolcleanup != nil {
		Poolcleanup()
	}

	for _, p := range &_lock.Allp {
		if p == nil {
			break
		}
		// clear tinyalloc pool
		if c := p.Mcache; c != nil {
			c.Tiny = nil
			c.Tinyoffset = 0

			// disconnect cached list before dropping it on the floor,
			// so that a dangling ref to one entry does not pin all of them.
			var sg, sgnext *_core.Sudog
			for sg = c.Sudogcache; sg != nil; sg = sgnext {
				sgnext = sg.Next
				sg.Next = nil
			}
			c.Sudogcache = nil
		}

		// clear defer pools
		for i := range p.Deferpool {
			// disconnect cached list before dropping it on the floor,
			// so that a dangling ref to one entry does not pin all of them.
			var d, dlink *_core.Defer
			for d = p.Deferpool[i]; d != nil; d = dlink {
				dlink = d.Link
				d.Link = nil
			}
			p.Deferpool[i] = nil
		}
	}
}

// backgroundgc is running in a goroutine and does the concurrent GC work.
// bggc holds the state of the backgroundgc.
func backgroundgc() {
	Bggc.g = _core.Getg()
	Bggc.g.Issystem = true
	for {
		gcwork(0)
		_lock.Lock(&Bggc.lock)
		Bggc.Working = 0
		_sched.Goparkunlock(&Bggc.lock, "Concurrent GC wait", _sched.TraceEvGoBlock)
	}
}

func bgsweep() {
	sweep.g = _core.Getg()
	_core.Getg().Issystem = true
	for {
		for gosweepone() != ^uintptr(0) {
			sweep.nbgsweep++
			Gosched()
		}
		_lock.Lock(&gclock)
		if !gosweepdone() {
			// This can happen if a GC runs between
			// gosweepone returning ^0 above
			// and the lock being acquired.
			_lock.Unlock(&gclock)
			continue
		}
		sweep.parked = true
		_sched.Goparkunlock(&gclock, "GC sweep wait", _sched.TraceEvGoBlock)
	}
}
