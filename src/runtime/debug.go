// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	"unsafe"
)

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
