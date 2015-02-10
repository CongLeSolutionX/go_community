// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: type and heap bitmaps.
//
// Type bitmaps
//
// The global variables (in the data and bss sections) and types that aren't too large
// record information about the layout of their memory words using a type bitmap.
// The bitmap holds two bits for each pointer-sized word. The two-bit values are:
//
// 	00 - typeDead: not a pointer, and no pointers in the rest of the object
//	01 - typeScalar: not a pointer
//	10 - typePointer: a pointer that GC should trace
//	11 - unused
//
// typeDead only appears in type bitmaps in Go type descriptors
// and in type bitmaps embedded in the heap bitmap (see below).
// It is not used in the type bitmap for the global variables.
//
// Heap bitmap
//
// The allocated heap comes from a subset of the memory in the range [start, used),
// where start == mheap_.arena_start and used == mheap_.arena_used.
// The heap bitmap comprises 4 bits for each pointer-sized word in that range,
// stored in bytes indexed backward in memory from start.
// That is, the byte at address start-1 holds the 4-bit entries for the two words
// start, start+ptrSize, the byte at start-2 holds the entries for start+2*ptrSize,
// start+3*ptrSize, and so on.
// In the byte holding the entries for addresses p and p+ptrSize, the low 4 bits
// describe p and the high 4 bits describe p+ptrSize.
//
// The 4 bits for each word are:
//	0001 - bitBoundary: this is the start of an object
//	0010 - bitMarked: this object has been marked by GC
//	tt00 - word type bits, as in a type bitmap.
//
// The code makes use of the fact that the zero value for a heap bitmap nibble
// has no boundary bit set, no marked bit set, and type bits == typeDead.
// These properties must be preserved when modifying the encoding.
//
// Checkmarks
//
// In a concurrent garbage collector, one worries about failing to mark
// a live object due to mutations without write barriers or bugs in the
// collector implementation. As a sanity check, the GC has a 'checkmark'
// mode that retraverses the object graph with the world stopped, to make
// sure that everything that should be marked is marked.
// In checkmark mode, in the heap bitmap, the type bits for the first word
// of an object are redefined:
//
//	00 - typeScalarCheckmarked // typeScalar, checkmarked
//	01 - typeScalar // typeScalar, not checkmarked
//	10 - typePointer // typePointer, not checkmarked
//	11 - typePointerCheckmarked // typePointer, checkmarked
//
// That is, typeDead is redefined to be typeScalar + a checkmark, and the
// previously unused 11 pattern is redefined to be typePointer + a checkmark.
// To prepare for this mode, we must move any typeDead in the first word of
// a multiword object to the second word.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// heapBitsForSpan returns the heapBits for the span base address base.
func HeapBitsForSpan(base uintptr) (hbits _sched.HeapBits) {
	if base < _lock.Mheap_.Arena_start || base >= _lock.Mheap_.Arena_end {
		_lock.Throw("heapBitsForSpan: base out of range")
	}
	hbits = _sched.HeapBitsForAddr(base)
	if hbits.Shift != 0 {
		_lock.Throw("heapBitsForSpan: unaligned start")
	}
	return hbits
}

// heapBitsSweepSpan coordinates the sweeping of a span by reading
// and updating the corresponding heap bitmap entries.
// For each free object in the span, heapBitsSweepSpan sets the type
// bits for the first two words (or one for single-word objects) to typeDead
// and then calls f(p), where p is the object's base address.
// f is expected to add the object to a free list.
func heapBitsSweepSpan(base, size, n uintptr, f func(uintptr)) {
	h := HeapBitsForSpan(base)
	if size == _core.PtrSize {
		// Only possible on 64-bit system, since minimum size is 8.
		// Must read and update both top and bottom nibble of each byte.
		bitp := h.Bitp
		for i := uintptr(0); i < n; i += 2 {
			x := int(*bitp)
			if x&_sched.BitMarked != 0 {
				x &^= _sched.BitMarked
			} else {
				x &^= _sched.TypeMask << _sched.TypeShift
				f(base + i*_core.PtrSize)
			}
			if x&(_sched.BitMarked<<4) != 0 {
				x &^= _sched.BitMarked << 4
			} else {
				x &^= _sched.TypeMask << (4 + _sched.TypeShift)
				f(base + (i+1)*_core.PtrSize)
			}
			*bitp = uint8(x)
			bitp = _sched.Subtractb(bitp, 1)
		}
		return
	}

	bitp := h.Bitp
	step := size / _sched.HeapBitmapScale
	for i := uintptr(0); i < n; i++ {
		x := int(*bitp)
		if x&_sched.BitMarked != 0 {
			x &^= _sched.BitMarked
		} else {
			x = _sched.BitBoundary // clear marked bit, set type bits to typeDead
			f(base + i*size)
		}
		*bitp = uint8(x)
		bitp = _sched.Subtractb(bitp, step)
	}
}
