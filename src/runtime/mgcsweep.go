// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: sweeping

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
)

func bgsweep(c chan int) {
	_gc.Sweep.G = _base.Getg()

	_base.Lock(&_gc.Sweep.Lock)
	_gc.Sweep.Parked = true
	c <- 1
	_base.Goparkunlock(&_gc.Sweep.Lock, "GC sweep wait", _base.TraceEvGoBlock, 1)

	for {
		for _gc.Gosweepone() != ^uintptr(0) {
			_gc.Sweep.Nbgsweep++
			_gc.Gosched()
		}
		_base.Lock(&_gc.Sweep.Lock)
		if !gosweepdone() {
			// This can happen if a GC runs between
			// gosweepone returning ^0 above
			// and the lock being acquired.
			_base.Unlock(&_gc.Sweep.Lock)
			continue
		}
		_gc.Sweep.Parked = true
		_base.Goparkunlock(&_gc.Sweep.Lock, "GC sweep wait", _base.TraceEvGoBlock, 1)
	}
}

//go:nowritebarrier
func gosweepdone() bool {
	return _base.Mheap_.Sweepdone != 0
}
