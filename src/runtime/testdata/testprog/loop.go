// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"sync/atomic"
)

func init() {
	register("LoopPreemption", LoopPreemption)
}

func LoopPreemption() {
	var start, stop uint32
	go func() {
		for atomic.LoadUint32(&start) == 0 {
		}
		// Force a loop preemption.
		runtime.GC()
		atomic.StoreUint32(&stop, 1)
	}()
	for atomic.LoadUint32(&stop) == 0 {
		atomic.StoreUint32(&start, 1)
	}
	println("OK")
}
