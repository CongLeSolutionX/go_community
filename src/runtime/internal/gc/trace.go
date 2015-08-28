// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go execution tracer.
// The tracer captures a wide range of execution events like goroutine
// creation/blocking/unblocking, syscall enter/exit/block, GC-related events,
// changes of heap size, processor start/stop, etc and writes them to a buffer
// in a compact form. A precise nanosecond-precision timestamp and a stack
// trace is captured for most events.
// See https://golang.org/s/go15trace for more info.

package gc

import (
	_base "runtime/internal/base"
)

// traceProcFree frees trace buffer associated with pp.
func traceProcFree(pp *_base.P) {
	buf := pp.Tracebuf
	pp.Tracebuf = nil
	if buf == nil {
		return
	}
	_base.Lock(&_base.Trace.Lock)
	_base.TraceFullQueue(buf)
	_base.Unlock(&_base.Trace.Lock)
}

// The following functions write specific events to trace.

func traceGomaxprocs(procs int32) {
	_base.TraceEvent(_base.TraceEvGomaxprocs, 1, uint64(procs))
}

func traceGCStart() {
	_base.TraceEvent(_base.TraceEvGCStart, 4)
}

func traceGCDone() {
	_base.TraceEvent(_base.TraceEvGCDone, -1)
}

func traceGCSweepStart() {
	_base.TraceEvent(_base.TraceEvGCSweepStart, 1)
}

func traceGCSweepDone() {
	_base.TraceEvent(_base.TraceEvGCSweepDone, -1)
}

func TraceGoSched() {
	_base.TraceEvent(_base.TraceEvGoSched, 1)
}

func TraceHeapAlloc() {
	_base.TraceEvent(_base.TraceEvHeapAlloc, -1, _base.Memstats.Heap_live)
}

func traceNextGC() {
	_base.TraceEvent(_base.TraceEvNextGC, -1, _base.Memstats.Next_gc)
}
