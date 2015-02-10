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

package maps

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	_sched "runtime/internal/sched"
	"unsafe"
)

// TODO(rsc): Clean up the next two functions.

// heapBitsSetType records that the new allocation [x, x+size)
// holds in [x, x+dataSize) one or more values of type typ.
// (The number of values is given by dataSize / typ.size.)
// If dataSize < size, the fragment [x+dataSize, x+size) is
// recorded as non-pointer data.
func heapBitsSetType(x, size, dataSize uintptr, typ *_core.Type) {
	// From here till marked label marking the object as allocated
	// and storing type info in the GC bitmap.
	h := _sched.HeapBitsForAddr(x)
	if _sched.DebugMalloc && (*h.Bitp>>h.Shift)&0x0f != _sched.BitBoundary {
		println("runtime: bits =", (*h.Bitp>>h.Shift)&0x0f)
		_lock.Throw("bad bits in markallocated")
	}

	var ti, te uintptr
	var ptrmask *uint8
	if size == _core.PtrSize {
		// It's one word and it has pointers, it must be a pointer.
		// The bitmap byte is shared with the one-word object
		// next to it, and concurrent GC might be marking that
		// object, so we must use an atomic update.
		_sched.Atomicor8(h.Bitp, _sched.TypePointer<<(_sched.TypeShift+h.Shift))
		return
	}
	if typ.Kind&_channels.KindGCProg != 0 {
		nptr := (uintptr(typ.Size) + _core.PtrSize - 1) / _core.PtrSize
		masksize := nptr
		if masksize%2 != 0 {
			masksize *= 2 // repeated
		}
		const typeBitsPerByte = 8 / _sched.TypeBitsWidth
		masksize = masksize * typeBitsPerByte / 8 // 4 bits per word
		masksize++                                // unroll flag in the beginning
		if masksize > _channels.MaxGCMask && typ.Gc[1] != 0 {
			// write barriers have not been updated to deal with this case yet.
			_lock.Throw("maxGCMask too small for now")
			// If the mask is too large, unroll the program directly
			// into the GC bitmap. It's 7 times slower than copying
			// from the pre-unrolled mask, but saves 1/16 of type size
			// memory for the mask.
			_lock.Systemstack(func() {
				unrollgcproginplace_m(unsafe.Pointer(x), typ, size, dataSize)
			})
			return
		}
		ptrmask = (*uint8)(unsafe.Pointer(uintptr(typ.Gc[0])))
		// Check whether the program is already unrolled
		// by checking if the unroll flag byte is set
		maskword := uintptr(_prof.Atomicloadp(unsafe.Pointer(ptrmask)))
		if *(*uint8)(unsafe.Pointer(&maskword)) == 0 {
			_lock.Systemstack(func() {
				_channels.Unrollgcprog_m(typ)
			})
		}
		ptrmask = (*uint8)(_core.Add(unsafe.Pointer(ptrmask), 1)) // skip the unroll flag byte
	} else {
		ptrmask = (*uint8)(unsafe.Pointer(typ.Gc[0])) // pointer to unrolled mask
	}
	if size == 2*_core.PtrSize {
		*h.Bitp = *ptrmask | _sched.BitBoundary
		return
	}
	te = uintptr(typ.Size) / _core.PtrSize
	// If the type occupies odd number of words, its mask is repeated.
	if te%2 == 0 {
		te /= 2
	}
	// Copy pointer bitmask into the bitmap.
	for i := uintptr(0); i < dataSize; i += 2 * _core.PtrSize {
		v := *(*uint8)(_core.Add(unsafe.Pointer(ptrmask), ti))
		ti++
		if ti == te {
			ti = 0
		}
		if i == 0 {
			v |= _sched.BitBoundary
		}
		if i+_core.PtrSize == dataSize {
			v &^= _sched.TypeMask << (4 + _sched.TypeShift)
		}

		*h.Bitp = v
		h.Bitp = _sched.Subtractb(h.Bitp, 1)
	}
	if dataSize%(2*_core.PtrSize) == 0 && dataSize < size {
		// Mark the word after last object's word as typeDead.
		*h.Bitp = 0
	}
}

func unrollgcproginplace_m(v unsafe.Pointer, typ *_core.Type, size, size0 uintptr) {
	// TODO(rsc): Explain why these non-atomic updates are okay.
	pos := uintptr(0)
	prog := (*byte)(unsafe.Pointer(uintptr(typ.Gc[1])))
	for pos != size0 {
		_channels.Unrollgcprog1((*byte)(v), prog, &pos, true, true)
	}

	// Mark first word as bitAllocated.
	// Mark word after last as typeDead.
	// TODO(rsc): Explain why we need to set this boundary.
	// Aren't the boundaries always set for the whole span?
	// Did unrollgcproc1 overwrite the boundary bit?
	// Is that okay?
	h := _sched.HeapBitsForAddr(uintptr(v))
	*h.Bitp |= _sched.BitBoundary << h.Shift
	if size0 < size {
		h := _sched.HeapBitsForAddr(uintptr(v) + size0)
		*h.Bitp &^= _sched.TypeMask << _sched.TypeShift
	}
}
