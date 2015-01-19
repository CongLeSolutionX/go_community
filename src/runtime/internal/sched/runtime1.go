// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

// Helpers for Go. Must be NOSPLIT, must only call NOSPLIT functions, and must not block.

//go:nosplit
func Acquirem() *_core.M {
	_g_ := _core.Getg()
	_g_.M.Locks++
	return _g_.M
}

//go:nosplit
func Releasem(mp *_core.M) {
	_g_ := _core.Getg()
	mp.Locks--
	if mp.Locks == 0 && _g_.Preempt {
		// restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _lock.StackPreempt
	}
}
