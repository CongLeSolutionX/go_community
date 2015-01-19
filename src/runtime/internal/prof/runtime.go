// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prof

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

//go:generate go run wincallback.go

var ticks struct {
	lock _core.Mutex
	val  uint64
}

// Note: Called by runtime/pprof in addition to runtime code.
func Tickspersecond() int64 {
	r := int64(_sched.Atomicload64(&ticks.val))
	if r != 0 {
		return r
	}
	_lock.Lock(&ticks.lock)
	r = int64(ticks.val)
	if r == 0 {
		t0 := _lock.Nanotime()
		c0 := _sched.Cputicks()
		_core.Usleep(100 * 1000)
		t1 := _lock.Nanotime()
		c1 := _sched.Cputicks()
		if t1 == t0 {
			t1++
		}
		r = (c1 - c0) * 1000 * 1000 * 1000 / (t1 - t0)
		if r == 0 {
			r++
		}
		_sched.Atomicstore64(&ticks.val, uint64(r))
	}
	_lock.Unlock(&ticks.lock)
	return r
}
