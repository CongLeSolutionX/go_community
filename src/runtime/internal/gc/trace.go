// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go execution tracer.
// The tracer captures a wide range of execution events like goroutine
// creation/blocking/unblocking, syscall enter/exit/block, GC-related events,
// changes of heap size, processor start/stop, etc and writes them to a buffer
// in a compact form. A precise nanosecond-precision timestamp and a stack
// trace is captured for most events.
// See http://golang.org/s/go15trace for more info.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// traceProcFree frees trace buffer associated with pp.
func traceProcFree(pp *_core.P) {
	buf := pp.Tracebuf
	pp.Tracebuf = nil
	if buf == nil {
		return
	}
	_lock.Lock(&_sched.Trace.Lock)
	_sched.TraceFullQueue(buf)
	_lock.Unlock(&_sched.Trace.Lock)
}

// The following functions write specific events to trace.

func traceGomaxprocs(procs int32) {
	_sched.TraceEvent(_sched.TraceEvGomaxprocs, true, uint64(procs))
}

func traceGCStart() {
	_sched.TraceEvent(_sched.TraceEvGCStart, true)
}

func traceGCDone() {
	_sched.TraceEvent(_sched.TraceEvGCDone, false)
}

func traceGCSweepStart() {
	_sched.TraceEvent(_sched.TraceEvGCSweepStart, true)
}

func traceGCSweepDone() {
	_sched.TraceEvent(_sched.TraceEvGCSweepDone, false)
}

func TraceHeapAlloc() {
	_sched.TraceEvent(_sched.TraceEvHeapAlloc, false, _lock.Memstats.Heap_alloc)
}

func traceNextGC() {
	_sched.TraceEvent(_sched.TraceEvNextGC, false, _lock.Memstats.Next_gc)
}
