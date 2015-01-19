// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	_sem "runtime/internal/sem"
	"unsafe"
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

	_sem.Semacquire(&_gc.Worldsema, false)
	gp := _core.Getg()
	gp.M.Gcing = 1
	_lock.Systemstack(_gc.Stoptheworld)

	// newprocs will be processed by starttheworld
	_gc.Newprocs = int32(n)

	gp.M.Gcing = 0
	_sem.Semrelease(&_gc.Worldsema)
	_lock.Systemstack(_gc.Starttheworld)
	return ret
}

// NumCPU returns the number of logical CPUs on the local machine.
func NumCPU() int {
	return int(_lock.Ncpu)
}

// NumCgoCall returns the number of cgo calls made by the current process.
func NumCgoCall() int64 {
	var n int64
	for mp := (*_core.M)(_prof.Atomicloadp(unsafe.Pointer(&_lock.Allm))); mp != nil; mp = mp.Alllink {
		n += int64(mp.Ncgocall)
	}
	return n
}

// NumGoroutine returns the number of goroutines that currently exist.
func NumGoroutine() int {
	return int(_gc.Gcount())
}
