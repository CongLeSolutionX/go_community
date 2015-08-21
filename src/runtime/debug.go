// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	"unsafe"
)

// GOMAXPROCS sets the maximum number of CPUs that can be executing
// simultaneously and returns the previous setting.  If n < 1, it does not
// change the current setting.
// The number of logical CPUs on the local machine can be queried with NumCPU.
// This call will go away when the scheduler improves.
func GOMAXPROCS(n int) int {
	if n > _base.MaxGomaxprocs {
		n = _base.MaxGomaxprocs
	}
	_base.Lock(&_base.Sched.Lock)
	ret := int(_base.Gomaxprocs)
	_base.Unlock(&_base.Sched.Lock)
	if n <= 0 || n == ret {
		return ret
	}

	stopTheWorld("GOMAXPROCS")

	// newprocs will be processed by startTheWorld
	_gc.Newprocs = int32(n)

	startTheWorld()
	return ret
}

// NumCPU returns the number of logical CPUs usable by the current process.
func NumCPU() int {
	return int(_base.Ncpu)
}

// NumCgoCall returns the number of cgo calls made by the current process.
func NumCgoCall() int64 {
	var n int64
	for mp := (*_base.M)(_iface.Atomicloadp(unsafe.Pointer(&_base.Allm))); mp != nil; mp = mp.Alllink {
		n += int64(mp.Ncgocall)
	}
	return n
}

// NumGoroutine returns the number of goroutines that currently exist.
func NumGoroutine() int {
	return int(gcount())
}
