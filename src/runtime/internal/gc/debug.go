// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sem "runtime/internal/sem"
)

// GOMAXPROCS sets the maximum number of CPUs that can be executing
// simultaneously and returns the previous setting.  If n < 1, it does not
// change the current setting.
// The number of logical CPUs on the local machine can be queried with NumCPU.
// This call will go away when the scheduler improves.
func GOMAXPROCS(n int) int {
	if n > _lock.MaxGomaxprocs {
		n = _lock.MaxGomaxprocs
	}
	_lock.Lock(&_core.Sched.Lock)
	ret := int(_lock.Gomaxprocs)
	_lock.Unlock(&_core.Sched.Lock)
	if n <= 0 || n == ret {
		return ret
	}

	_sem.Semacquire(&Worldsema, false)
	gp := _core.Getg()
	gp.M.Preemptoff = "GOMAXPROCS"
	_lock.Systemstack(Stoptheworld)

	// newprocs will be processed by starttheworld
	newprocs = int32(n)

	gp.M.Preemptoff = ""
	_sem.Semrelease(&Worldsema)
	_lock.Systemstack(Starttheworld)
	return ret
}
