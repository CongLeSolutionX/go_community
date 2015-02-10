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

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

const (
	TypeDead               = 0
	TypeScalarCheckmarked  = 0
	TypeScalar             = 1
	TypePointer            = 2
	TypePointerCheckmarked = 3

	TypeBitsWidth   = 2 // # of type bits per pointer-sized word
	TypeMask        = 1<<TypeBitsWidth - 1
	TypeBitmapScale = _core.PtrSize * (8 / TypeBitsWidth) // number of data bytes per type bitmap byte

	HeapBitsWidth   = 4
	HeapBitmapScale = _core.PtrSize * (8 / HeapBitsWidth) // number of data bytes per heap bitmap byte
	BitBoundary     = 1
	BitMarked       = 2
	TypeShift       = 2
)

// addb returns the byte pointer p+n.
//go:nowritebarrier
func Addb(p *byte, n uintptr) *byte {
	return (*byte)(_core.Add(unsafe.Pointer(p), n))
}

// subtractb returns the byte pointer p-n.
//go:nowritebarrier
func Subtractb(p *byte, n uintptr) *byte {
	return (*byte)(_core.Add(unsafe.Pointer(p), -n))
}

// mHeap_MapBits is called each time arena_used is extended.
// It maps any additional bitmap memory needed for the new arena memory.
//
//go:nowritebarrier
func mHeap_MapBits(h *_lock.Mheap) {
	// Caller has added extra mappings to the arena.
	// Add extra mappings of bitmap words as needed.
	// We allocate extra bitmap pieces in chunks of bitmapChunk.
	const bitmapChunk = 8192

	n := (_lock.Mheap_.Arena_used - _lock.Mheap_.Arena_start) / HeapBitmapScale
	n = _lock.Round(n, bitmapChunk)
	n = _lock.Round(n, _lock.PhysPageSize)
	if h.Bitmap_mapped >= n {
		return
	}

	SysMap(unsafe.Pointer(h.Arena_start-n), n-h.Bitmap_mapped, h.Arena_reserved, &_lock.Memstats.Gc_sys)
	h.Bitmap_mapped = n
}

// heapBits provides access to the bitmap bits for a single heap word.
// The methods on heapBits take value receivers so that the compiler
// can more easily inline calls to those methods and registerize the
// struct fields independently.
type HeapBits struct {
	Bitp  *uint8
	Shift uint32
}

// heapBitsForAddr returns the heapBits for the address addr.
// The caller must have already checked that addr is in the range [mheap_.arena_start, mheap_.arena_used).
func HeapBitsForAddr(addr uintptr) HeapBits {
	off := (addr - _lock.Mheap_.Arena_start) / _core.PtrSize
	return HeapBits{(*uint8)(unsafe.Pointer(_lock.Mheap_.Arena_start - off/2 - 1)), uint32(4 * (off & 1))}
}

// heapBitsForObject returns the base address for the heap object
// containing the address p, along with the heapBits for base.
// If p does not point into a heap object, heapBitsForObject returns base == 0.
func heapBitsForObject(p uintptr) (base uintptr, hbits HeapBits) {
	if p < _lock.Mheap_.Arena_start || p >= _lock.Mheap_.Arena_used {
		return
	}

	// If heap bits for the pointer-sized word containing p have bitBoundary set,
	// then we know this is the base of the object, and we can stop now.
	// This handles the case where p is the base and, due to rounding
	// when looking up the heap bits, also the case where p points beyond
	// the base but still into the first pointer-sized word of the object.
	hbits = HeapBitsForAddr(p)
	if hbits.isBoundary() {
		base = p &^ (_core.PtrSize - 1)
		return
	}

	// Otherwise, p points into the middle of an object.
	// Consult the span table to find the block beginning.
	// TODO(rsc): Factor this out.
	k := p >> _core.PageShift
	x := k
	x -= _lock.Mheap_.Arena_start >> _core.PageShift
	s := H_spans[x]
	if s == nil || _core.PageID(k) < s.Start || p >= s.Limit || s.State != XMSpanInUse {
		if s != nil && s.State == MSpanStack {
			return
		}

		// The following ensures that we are rigorous about what data
		// structures hold valid pointers.
		// TODO(rsc): Check if this still happens.
		if false {
			// Still happens sometimes. We don't know why.
			Printlock()
			print("runtime:objectstart Span weird: p=", _core.Hex(p), " k=", _core.Hex(k))
			if s == nil {
				print(" s=nil\n")
			} else {
				print(" s.start=", _core.Hex(s.Start<<_core.PageShift), " s.limit=", _core.Hex(s.Limit), " s.state=", s.State, "\n")
			}
			Printunlock()
			_lock.Throw("objectstart: bad pointer in unexpected span")
		}
		return
	}
	base = s.Base()
	if p-base > s.Elemsize {
		base += (p - base) / s.Elemsize * s.Elemsize
	}
	if base == p {
		print("runtime: failed to find block beginning for ", _core.Hex(p), " s=", _core.Hex(s.Start*_core.PageSize), " s.limit=", _core.Hex(s.Limit), "\n")
		_lock.Throw("failed to find block beginning")
	}

	// Now that we know the actual base, compute heapBits to return to caller.
	hbits = HeapBitsForAddr(base)
	if !hbits.isBoundary() {
		_lock.Throw("missing boundary at computed object start")
	}
	return
}

// next returns the heapBits describing the next pointer-sized word in memory.
// That is, if h describes address p, h.next() describes p+ptrSize.
// Note that next does not modify h. The caller must record the result.
func (h HeapBits) Next() HeapBits {
	if h.Shift == 0 {
		return HeapBits{h.Bitp, 4}
	}
	return HeapBits{Subtractb(h.Bitp, 1), 0}
}

// isMarked reports whether the heap bits have the marked bit set.
func (h HeapBits) IsMarked() bool {
	return *h.Bitp&(BitMarked<<h.Shift) != 0
}

// setMarked sets the marked bit in the heap bits, atomically.
func (h HeapBits) SetMarked() {
	Atomicor8(h.Bitp, BitMarked<<h.Shift)
}

// setMarkedNonAtomic sets the marked bit in the heap bits, non-atomically.
func (h HeapBits) SetMarkedNonAtomic() {
	*h.Bitp |= BitMarked << h.Shift
}

// isBoundary reports whether the heap bits have the boundary bit set.
func (h HeapBits) isBoundary() bool {
	return *h.Bitp&(BitBoundary<<h.Shift) != 0
}

// Note that there is no setBoundary or setBoundaryNonAtomic.
// Boundaries are always in bulk, for the entire span.

// typeBits returns the heap bits' type bits.
func (h HeapBits) TypeBits() uint8 {
	return (*h.Bitp >> (h.Shift + TypeShift)) & TypeMask
}

// isCheckmarked reports whether the heap bits have the checkmarked bit set.
func (h HeapBits) isCheckmarked() bool {
	typ := h.TypeBits()
	return typ == TypeScalarCheckmarked || typ == TypePointerCheckmarked
}

// setCheckmarked sets the checkmarked bit.
func (h HeapBits) setCheckmarked() {
	typ := h.TypeBits()
	if typ == TypeScalar {
		// Clear low type bit to turn 01 into 00.
		atomicand8(h.Bitp, ^((1 << TypeShift) << h.Shift))
	} else if typ == TypePointer {
		// Set low type bit to turn 10 into 11.
		Atomicor8(h.Bitp, (1<<TypeShift)<<h.Shift)
	}
}

// The methods operating on spans all require that h has been returned
// by heapBitsForSpan and that size, n, total are the span layout description
// returned by the mspan's layout method.
// If total > size*n, it means that there is extra leftover memory in the span,
// usually due to rounding.
//
// TODO(rsc): Perhaps introduce a different heapBitsSpan type.

// initSpan initializes the heap bitmap for a span.
func (h HeapBits) InitSpan(size, n, total uintptr) {
	if size == _core.PtrSize {
		// Only possible on 64-bit system, since minimum size is 8.
		// Set all nibbles to bitBoundary using uint64 writes.
		nbyte := n * _core.PtrSize / HeapBitmapScale
		nuint64 := nbyte / 8
		bitp := Subtractb(h.Bitp, nbyte-1)
		for i := uintptr(0); i < nuint64; i++ {
			const boundary64 = BitBoundary |
				BitBoundary<<4 |
				BitBoundary<<8 |
				BitBoundary<<12 |
				BitBoundary<<16 |
				BitBoundary<<20 |
				BitBoundary<<24 |
				BitBoundary<<28 |
				BitBoundary<<32 |
				BitBoundary<<36 |
				BitBoundary<<40 |
				BitBoundary<<44 |
				BitBoundary<<48 |
				BitBoundary<<52 |
				BitBoundary<<56 |
				BitBoundary<<60

			*(*uint64)(unsafe.Pointer(bitp)) = boundary64
			bitp = Addb(bitp, 8)
		}
		return
	}

	if size*n < total {
		// To detect end of object during GC object scan,
		// add boundary just past end of last block.
		// The object scan knows to stop when it reaches
		// the end of the span, but in this case the object
		// ends before the end of the span.
		//
		// TODO(rsc): If the bitmap bits were going to be typeDead
		// otherwise, what's the point of this?
		// Can we delete this logic?
		n++
	}
	step := size / HeapBitmapScale
	bitp := h.Bitp
	for i := uintptr(0); i < n; i++ {
		*bitp = BitBoundary
		bitp = Subtractb(bitp, step)
	}
}

// clearSpan clears the heap bitmap bytes for the span.
func (h HeapBits) ClearSpan(size, n, total uintptr) {
	if total%HeapBitmapScale != 0 {
		_lock.Throw("clearSpan: unaligned length")
	}
	nbyte := total / HeapBitmapScale
	_core.Memclr(unsafe.Pointer(Subtractb(h.Bitp, nbyte-1)), nbyte)
}

// initCheckmarkSpan initializes a span for being checkmarked.
// This would be a no-op except that we need to rewrite any
// typeDead bits in the first word of the object into typeScalar
// followed by a typeDead in the second word of the object.
func (h HeapBits) InitCheckmarkSpan(size, n, total uintptr) {
	if size == _core.PtrSize {
		// Only possible on 64-bit system, since minimum size is 8.
		// Must update both top and bottom nibble of each byte.
		// There is no second word in these objects, so all we have
		// to do is rewrite typeDead to typeScalar by adding the 1<<typeShift bit.
		bitp := h.Bitp
		for i := uintptr(0); i < n; i += 2 {
			x := int(*bitp)
			if x&0x11 != 0x11 {
				_lock.Throw("missing bitBoundary")
			}
			if (x>>TypeShift)&TypeMask == TypeDead {
				x += (TypeScalar - TypeDead) << TypeShift
			}
			if (x>>(4+TypeShift))&TypeMask == TypeDead {
				x += (TypeScalar - TypeDead) << (4 + TypeShift)
			}
			*bitp = uint8(x)
			bitp = Subtractb(bitp, 1)
		}
		return
	}

	// Update bottom nibble for first word of each object.
	// If the bottom nibble says typeDead, change to typeScalar
	// and clear top nibble to mark as typeDead.
	bitp := h.Bitp
	step := size / HeapBitmapScale
	for i := uintptr(0); i < n; i++ {
		if *bitp&BitBoundary == 0 {
			_lock.Throw("missing bitBoundary")
		}
		x := *bitp
		if (x>>TypeShift)&TypeMask == TypeDead {
			x += (TypeScalar - TypeDead) << TypeShift
			x &= 0x0f // clear top nibble to typeDead
		}
		bitp = Subtractb(bitp, step)
	}
}

// clearCheckmarkSpan removes all the checkmarks from a span.
// If it finds a multiword object starting with typeScalar typeDead,
// it rewrites the heap bits to the simpler typeDead typeDead.
func (h HeapBits) ClearCheckmarkSpan(size, n, total uintptr) {
	if size == _core.PtrSize {
		// Only possible on 64-bit system, since minimum size is 8.
		// Must update both top and bottom nibble of each byte.
		// typeScalarCheckmarked can be left as typeDead,
		// but we want to change typeScalar back to typeDead.
		bitp := h.Bitp
		for i := uintptr(0); i < n; i += 2 {
			x := int(*bitp)
			if x&(BitBoundary|BitBoundary<<4) != (BitBoundary | BitBoundary<<4) {
				_lock.Throw("missing bitBoundary")
			}

			switch typ := (x >> TypeShift) & TypeMask; typ {
			case TypeScalar:
				x += (TypeDead - TypeScalar) << TypeShift
			case TypePointerCheckmarked:
				x += (TypePointer - TypePointerCheckmarked) << TypeShift
			}

			switch typ := (x >> (4 + TypeShift)) & TypeMask; typ {
			case TypeScalar:
				x += (TypeDead - TypeScalar) << (4 + TypeShift)
			case TypePointerCheckmarked:
				x += (TypePointer - TypePointerCheckmarked) << (4 + TypeShift)
			}

			*bitp = uint8(x)
			bitp = Subtractb(bitp, 1)
		}
		return
	}

	// Update bottom nibble for first word of each object.
	// If the bottom nibble says typeScalarCheckmarked and the top is not typeDead,
	// change to typeScalar. Otherwise leave, since typeScalarCheckmarked == typeDead.
	// If the bottom nibble says typePointerCheckmarked, change to typePointer.
	bitp := h.Bitp
	step := size / HeapBitmapScale
	for i := uintptr(0); i < n; i++ {
		x := int(*bitp)
		if x&BitBoundary == 0 {
			_lock.Throw("missing bitBoundary")
		}

		switch typ := (x >> TypeShift) & TypeMask; {
		case typ == TypeScalarCheckmarked && (x>>(4+TypeShift))&TypeMask != TypeDead:
			x += (TypeScalar - TypeScalarCheckmarked) << TypeShift
		case typ == TypePointerCheckmarked:
			x += (TypePointer - TypePointerCheckmarked) << TypeShift
		}

		*bitp = uint8(x)
		bitp = Subtractb(bitp, step)
	}
}
