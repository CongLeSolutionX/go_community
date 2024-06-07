// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime -> tracer API for GC scan tracing.

package runtime

type scanType uint8

const (
	scanTypeBlock scanType = iota
	scanTypeObject
	scanTypeConservative
	scanTypeTinyAllocs
)

type evScanPerP struct {
	typ  scanType
	base uintptr

	n    int
	offs [1024]uintptr
	ptrs [1024]uintptr
}

//go:nosplit
func traceGCScanEnabled() bool {
	return traceEnabled()
}

func (tl traceLocker) GCScanHeader() {
	for cls := range class_to_size {
		tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanClass, traceArg(class_to_size[cls]), traceArg(class_to_allocnpages[cls]))
	}
}

func (tl traceLocker) GCScanSpan(s *mspan) {
	// Only heap spans.
	if s.state.get() != mSpanInUse {
		return
	}

	size := uint64(s.spanclass)
	if s.spanclass.sizeclass() == 0 {
		// Large object
		size |= uint64(s.npages) << 8
	}
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanSpan, traceArg(s.startAddr), traceArg(size))
}

func (tl traceLocker) GCScan(b, n uintptr, typ scanType) {
	if b&0b11 != 0 {
		throw("unaligned base")
	}
	ec := &getg().m.p.ptr().evScan
	if ec.base != 0 {
		throw("scan already in progress on this P")
	}
	ec.typ = typ
	ec.base = b
}

func (tl traceLocker) GCScanEnd() {
	ec := &getg().m.p.ptr().evScan
	tl.gcScanFlush(ec)
	ec.base = 0
}

func (tl traceLocker) GCScanPointer(p uintptr, offset uintptr, found bool) {
	p &^= 0b11
	p |= uintptr(bool2int(found))

	ec := &getg().m.p.ptr().evScan
	if ec.n == len(ec.ptrs) {
		tl.gcScanFlush(ec)
	}
	ec.ptrs[ec.n] = p
	ec.offs[ec.n] = offset
	ec.n++
}

func (tl traceLocker) gcScanFlush(ec *evScanPerP) {
	// Compute the length of the message. We go to this extra trouble because
	// it's long and many of the varints will pack very small.
	msgLen := 3 * traceBytesPerNumber

	var prevOff uintptr
	for i := range ec.n {
		msgLen += varintLen(uint64(ec.offs[i] - prevOff))
		msgLen += varintLen(uint64(ec.ptrs[i]))
		prevOff = ec.offs[i]
	}

	// TODO: This would be much simpler if there were just an easy way for an
	// event to refer to a particular location in an experimental batch.

	// Acquire space.
	w := tl.expWriter(traceExperimentGCScan)
	w, flushed := w.ensure(msgLen)
	if flushed {
		// Write the batch header
		w.varint(uint64(memstats.numgc))
	}

	// Write the message.
	w.varint(uint64(ec.base) | uint64(ec.typ))
	w.varint(uint64(ec.n))
	prevOff = 0
	for i := range ec.n {
		w.varint(uint64(ec.offs[i] - prevOff))
		w.varint(uint64(ec.ptrs[i]))
		prevOff = ec.offs[i]
	}

	ec.n = 0

	w.end()
}

func (tl traceLocker) GCScanWB(p uintptr, found bool) {
	p &^= 0b11
	if found {
		p |= 1
	}
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanWB, traceArg(p))
}

func (tl traceLocker) GCScanToAllocBlack(b uintptr) {
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanAllocBlack, traceArg(b))
}

func (tl traceLocker) GCScanGCDone(gen uint32) {
	// Flush experimental batches
	lock(&sched.lock)
	for mp := allm; mp != nil; mp = mp.alllink {
		bufp := &mp.trace.buf[tl.gen%2][traceExperimentGCScan]
		if *bufp != nil {
			w := unsafeTraceExpWriter(tl.gen, *bufp, traceExperimentGCScan)
			w.flush().end()
			*bufp = nil
		}
	}
	unlock(&sched.lock)

	// Mark the end of the cycle.
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanGCDone, traceArg(gen))
}
