// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

var loopUnmapSink int

func BenchmarkLoopUnmap(b *testing.B) {
	runtime.LoopSignalCount = 0
	for i := 0; i < b.N; i++ {
		runtime.LoopUnmap()
		loopUnmapSink = *(*int)(unsafe.Pointer(runtime.PreemptAddress))
	}
	b.Log(runtime.LoopSignalCount, "preemption signals")
}

func BenchmarkLoopUnmapIPIBroadcast(b *testing.B) {
	// BenchmarkLoopUnmap doesn't require any IPIs for remote TLB
	// shootdown. This benchmark fires up threads on all of the
	// cores, forcing the kernel to perform TLB shootdowns on all
	// cores.
	var done uint32
	runtime.LockOSThread()
	for i := 1; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			runtime.LockOSThread()
			for atomic.LoadUint32(&done) == 0 {
			}
		}()
	}
	defer func() {
		b.StopTimer()
		atomic.StoreUint32(&done, 1)
		time.Sleep(100 * time.Millisecond)
		runtime.UnlockOSThread()
	}()

	BenchmarkLoopUnmap(b)
}
