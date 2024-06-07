// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime -> tracer API for GC scan tracing.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

type batchHeader uint8

const (
	gcScanBatchSizes batchHeader = iota
	gcScanBatchSpans
	gcScanBatchScan
	gcScanBatchAllocs
	gcScanBatchTypes
)

type heapBitsType uint8

const (
	heapBitsNone   heapBitsType = iota // No scan
	heapBitsPacked                     // Packed at end of span
	heapBitsHeader                     // Pointed to by each object header
	heapBitsOOB                        // Pointed to by span struct (one object per span)
)

type scanType uint8

const (
	scanTypeNone scanType = iota
	scanTypeRoot
	scanTypeObject
)

type evScanPerP struct {
	typ  scanType
	base uintptr

	forceTyp scanType

	n    int
	offs [1024]uintptr
	ptrs [1024]uintptr
}

//go:nosplit
func traceGCScanEnabled() bool {
	return traceEnabled()
}

func (tl traceLocker) GCScanGCStart() {
	assertWorldStopped()

	w := tl.expWriter(traceExperimentGCScan)

	// Make sure we have a fresh buffer
	w = w.refill()

	// Write size class info
	w.byte(byte(gcScanBatchSizes))
	w.varint(uint64(memstats.numgc))
	for cls, elemSize := range class_to_size {
		w.varint(uint64(elemSize))
		w.varint(uint64(class_to_allocnpages[cls]))
		var hbt heapBitsType
		if cls == 0 {
			hbt = heapBitsOOB
		} else if heapBitsInSpan(uintptr(elemSize)) {
			hbt = heapBitsPacked
		} else {
			hbt = heapBitsHeader
		}
		w.byte(byte(hbt))
	}

	// Write spans
	w = w.flush()
	for _, s := range mheap_.allspans {
		if s.state.get() != mSpanInUse {
			continue
		}

		var flushed bool
		w, flushed = w.ensure(2 * traceBytesPerNumber)
		if flushed {
			// Write the header
			w.byte(byte(gcScanBatchSpans))
			w.varint(uint64(memstats.numgc))
		}

		size := uint64(s.spanclass)
		if s.spanclass.sizeclass() == 0 {
			// Large object
			size |= uint64(s.npages) << 8
		}

		w.varint(uint64(s.startAddr))
		w.varint(size)
	}

	w.flush().end()

	// Emit a GC start event. This is necessary for the consumer to be able to
	// find the experimental batches written above.
	tl.eventWriter(traceGoRunning, traceProcRunning).event(traceEvGCScanGCStart, traceArg(memstats.numgc), traceArg(gomaxprocs))
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
	tl.eventWriter(traceGoRunning, traceProcRunning).event(traceEvGCScanSpan, traceArg(s.startAddr), traceArg(size))
}

// traceGCScanForceType forces the type of the next GCScan to be typ.
func traceGCScanForceType(typ scanType) {
	ec := &getg().m.p.ptr().evScan
	if ec.base != 0 {
		throw("scan already in progress on this P")
	}
	ec.forceTyp = typ
}

func (tl traceLocker) GCScan(b, n uintptr, typ scanType) {
	if b&0b11 != 0 {
		throw("unaligned base")
	}
	ec := &getg().m.p.ptr().evScan
	if ec.base != 0 {
		throw("scan already in progress on this P")
	}
	if ec.forceTyp != scanTypeNone {
		typ = ec.forceTyp
		ec.forceTyp = scanTypeNone
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
		w.byte(byte(gcScanBatchScan))
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
	tl.eventWriter(traceGoRunning, traceProcRunning).event(traceEvGCScanWB, traceArg(p))
}

func (tl traceLocker) GCScanAllocBlack(b uintptr) {
	tl.eventWriter(traceGoRunning, traceProcRunning).event(traceEvGCScanAllocBlack, traceArg(b))
}

func (tl traceLocker) GCScanGCDone(gen uint32) {
	// Write out final heap metadata.
	tl.gcScanAllAllocs(gen)

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
	tl.eventWriter(traceGoRunning, traceProcRunning).event(traceEvGCScanGCDone, traceArg(gen))
}

// Allocation data

func (tl traceLocker) gcScanAllAllocs(gen uint32) {
	assertWorldStopped()

	w := tl.expWriter(traceExperimentGCScan)

	// Write allocations in each span
	w = w.flush() // Make sure we have a fresh buffer
	for _, s := range mheap_.allspans {
		if s.state.get() != mSpanInUse {
			continue
		}

		ensure := 1 * traceBytesPerNumber         // Header
		ensure += 8 * ((int(s.nelems) + 63) / 64) // Alloc bits
		if s.spanclass.noscan() {
			ensure += 1
		} else if heapBitsInSpan(s.elemsize) {
			// Span-level heap bits
			ensure += 1 + 8*len(s.heapBits())
		} else {
			// Per-object heap bits
			ensure += 1 + int(s.nelems)*traceBytesPerNumber
		}
		var flushed bool
		w, flushed = w.ensure(ensure)
		if flushed {
			// Write the header
			w.byte(byte(gcScanBatchAllocs))
			w.varint(uint64(gen))
		}

		// Write general span info.
		w.varint(uint64(s.startAddr))
		w.varint(uint64(s.nelems))

		// Write the alloc bits.
		abits := s.allocBitsForIndex(0)
		for i := 0; i < int(s.nelems); i += 64 {
			var word uint64
			for j := 0; j < 64 && i+j < int(s.nelems); j++ {
				if abits.index < uintptr(s.freeindex) || abits.isMarked() {
					word |= 1 << j
				}
				abits.advance()
			}
			w.uint64(word)
		}

		if s.spanclass.noscan() {
			// No heap bits.
			w.byte(0)
		} else if heapBitsInSpan(s.elemsize) {
			// The heap bits are packed in the span, so write them out as one block.
			heapBits := s.heapBits()
			if goarch.PtrSize != 8 {
				throw("not implemented: PtrSize != 8")
			}
			if len(heapBits) >= 0xff {
				throw("len(heapBits) >= 0xff not supported")
			}
			w.byte(byte(len(heapBits)))
			for _, b := range heapBits {
				w.uint64(uint64(b))
			}
		} else {
			// Write out the headers for each object separately. These may be
			// stored in object headers or in the span.
			if s.spanclass.sizeclass() == 0 {
				// Out-of-band
				w.byte(0xfe)
			} else {
				w.byte(0xff)
			}
			abits := s.allocBitsForIndex(0)
			for i := uintptr(0); i < uintptr(s.nelems); i++ {
				if abits.index < uintptr(s.freeindex) || abits.isMarked() {
					x := s.base() + i*s.elemsize
					hdr := s.typePointersOfUnchecked(x).typ
					id := theTypeBitsTable.put(hdr)
					if id == 0 {
						// ID 0 is used for "no type", but is implicitly
						// reserved in the trace map implementation.
						throw("got type ID 0, which is reserved")
					}
					//w.uint64(uint64(uintptr(unsafe.Pointer(hdr))))
					w.varint(id)
				}
				abits.advance()
			}
		}
	}

	w.flush().end()

	theTypeBitsTable.dump(w, gen)
	theTypeBitsTable.tab.reset()
}

// Type data

var theTypeBitsTable traceTypeBitsTable

type traceTypeBitsTable struct {
	tab traceMap
}

func (t *traceTypeBitsTable) put(typ *abi.Type) uint64 {
	if typ == nil {
		return 0
	}
	// Insert the pointer to the type itself.
	id, _ := t.tab.put(noescape(unsafe.Pointer(&typ)), goarch.PtrSize)
	return id
}

func (t *traceTypeBitsTable) dump(w traceWriter, gen uint32) {
	w = w.flush()

	for id, typ := range t.tab.all {
		typ := (*abi.Type)(typ)
		if goarch.BigEndian {
			throw("big endian not implemented")
		}
		if goarch.PtrSize != 8 {
			// The GCData is only guaranteed to be padded out to a uintptr.
			throw("32-bit not implemented")
		}
		ptrMask := unsafe.Slice((*uint64)(unsafe.Pointer(typ.GCData)), divRoundUp(typ.PtrBytes/goarch.PtrSize, 64))

		// Split up large pointer masks that may not fit even in a fresh batch.
		const ptrMaskBlock = 1 << 10
		for maskOffset := 0; maskOffset < len(ptrMask); maskOffset += ptrMaskBlock {
			var flushed bool
			w, flushed = w.ensure(3*traceBytesPerNumber + len(ptrMask))
			if flushed {
				// Write the header
				w.byte(byte(gcScanBatchTypes))
				w.varint(uint64(gen))
			}

			partial := 0
			ptrMaskPart := ptrMask[maskOffset:]
			if len(ptrMaskPart) > ptrMaskBlock {
				ptrMaskPart = ptrMaskPart[:ptrMaskBlock]
				partial = 1
			}

			w.varint(id)
			if maskOffset == 0 {
				w.varint(uint64(typ.Size_))
				w.varint(uint64(typ.PtrBytes / goarch.PtrSize))
			} else {
				w.varint(0)
				w.varint(uint64(maskOffset))
			}
			w.varint(uint64((len(ptrMaskPart) << 1) | partial))
			for _, v := range ptrMaskPart {
				w.uint64(v)
			}
		}
	}

	w.flush().end()
}
