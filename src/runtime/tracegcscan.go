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
	b |= uintptr(typ)
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScan, traceArg(b), traceArg(n))
}

func (tl traceLocker) GCScanEnd() {
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanEnd)
}

func (tl traceLocker) GCScanPointer(p uintptr, offset uintptr, found bool) {
	p &^= 0b11
	if found {
		p |= 1
	}
	tl.eventWriter(traceGoRunning, traceProcRunning).commit(traceEvGCScanPointer, traceArg(p), traceArg(offset))
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
