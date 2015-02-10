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

package core

import (
	"unsafe"
)

const (
	// Timestamps in trace are cputicks/traceTickDiv.
	// This makes absolute values of timestamp diffs smaller,
	// and so they are encoded in less number of bytes.
	// 64 is somewhat arbitrary (one tick is ~20ns on a 3GHz machine).
	TraceTickDiv = 64
	// Maximum number of PCs in a single stack trace.
	// Since events contain only stack id rather than whole stack trace,
	// we can allow quite large values here.
	TraceStackSize = 128
	// Identifier of a fake P that is used when we trace without a real P.
	TraceGlobProc = -1
	// Maximum number of bytes to encode uint64 in base-128.
	TraceBytesPerNumber = 10
	// Shift of the number of arguments in the first event byte.
	TraceArgCountShift = 6
)

// traceBufHeader is per-P tracing buffer.
type TraceBufHeader struct {
	Link      *TraceBuf               // in trace.empty/full
	LastTicks uint64                  // when we wrote the last event
	Buf       []byte                  // trace data, always points to traceBuf.arr
	Stk       [TraceStackSize]uintptr // scratch buffer for traceback
}

// traceBuf is per-P tracing buffer.
type TraceBuf struct {
	TraceBufHeader
	Arr [64<<10 - unsafe.Sizeof(TraceBufHeader{})]byte // underlying buffer for traceBufHeader.buf
}
