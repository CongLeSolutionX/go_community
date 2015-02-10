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

package channels

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	_sched "runtime/internal/sched"
	"unsafe"
)

// typeBitmapInHeapBitmapFormat returns a bitmap holding
// the type bits for the type typ, but expanded into heap bitmap format
// to make it easier to copy them into the heap bitmap.
// TODO(rsc): Change clients to use the type bitmap format instead,
// which can be stored more densely (especially if we drop to 1 bit per pointer).
//
// To make it easier to replicate the bits when filling out the heap
// bitmap for an array of typ, if typ holds an odd number of words
// (meaning the heap bitmap would stop halfway through a byte),
// typeBitmapInHeapBitmapFormat returns the bitmap for two instances
// of typ in a row.
// TODO(rsc): Remove doubling.
func TypeBitmapInHeapBitmapFormat(typ *_core.Type) []uint8 {
	var ptrmask *uint8
	nptr := (uintptr(typ.Size) + _core.PtrSize - 1) / _core.PtrSize
	if typ.Kind&KindGCProg != 0 {
		masksize := nptr
		if masksize%2 != 0 {
			masksize *= 2 // repeated
		}
		const typeBitsPerByte = 8 / _sched.TypeBitsWidth
		masksize = masksize * typeBitsPerByte / 8 // 4 bits per word
		masksize++                                // unroll flag in the beginning
		if masksize > MaxGCMask && typ.Gc[1] != 0 {
			// write barriers have not been updated to deal with this case yet.
			_lock.Throw("maxGCMask too small for now")
		}
		ptrmask = (*uint8)(unsafe.Pointer(uintptr(typ.Gc[0])))
		// Check whether the program is already unrolled
		// by checking if the unroll flag byte is set
		maskword := uintptr(_prof.Atomicloadp(unsafe.Pointer(ptrmask)))
		if *(*uint8)(unsafe.Pointer(&maskword)) == 0 {
			_lock.Systemstack(func() {
				Unrollgcprog_m(typ)
			})
		}
		ptrmask = (*uint8)(_core.Add(unsafe.Pointer(ptrmask), 1)) // skip the unroll flag byte
	} else {
		ptrmask = (*uint8)(unsafe.Pointer(typ.Gc[0])) // pointer to unrolled mask
	}
	return (*[1 << 30]byte)(unsafe.Pointer(ptrmask))[:(nptr+1)/2]
}

// GC type info programs
//
// TODO(rsc): Clean up and enable.

const (
	// GC type info programs.
	// The programs allow to store type info required for GC in a compact form.
	// Most importantly arrays take O(1) space instead of O(n).
	// The program grammar is:
	//
	// Program = {Block} "insEnd"
	// Block = Data | Array
	// Data = "insData" DataSize DataBlock
	// DataSize = int // size of the DataBlock in bit pairs, 1 byte
	// DataBlock = binary // dense GC mask (2 bits per word) of size ]DataSize/4[ bytes
	// Array = "insArray" ArrayLen Block "insArrayEnd"
	// ArrayLen = int // length of the array, 8 bytes (4 bytes for 32-bit arch)
	//
	// Each instruction (insData, insArray, etc) is 1 byte.
	// For example, for type struct { x []byte; y [20]struct{ z int; w *byte }; }
	// the program looks as:
	//
	// insData 3 (typePointer typeScalar typeScalar)
	//	insArray 20 insData 2 (typeScalar typePointer) insArrayEnd insEnd
	//
	// Total size of the program is 17 bytes (13 bytes on 32-bits).
	// The corresponding GC mask would take 43 bytes (it would be repeated
	// because the type has odd number of words).
	InsData = 1 + iota
	InsArray
	InsArrayEnd
	InsEnd

	// 64 bytes cover objects of size 1024/512 on 64/32 bits, respectively.
	MaxGCMask = 65536 // TODO(rsc): change back to 64
)

// Recursively unrolls GC program in prog.
// mask is where to store the result.
// If inplace is true, store the result not in mask but in the heap bitmap for mask.
// ppos is a pointer to position in mask, in bits.
// sparse says to generate 4-bits per word mask for heap (2-bits for data/bss otherwise).
//go:nowritebarrier
func Unrollgcprog1(maskp *byte, prog *byte, ppos *uintptr, inplace, sparse bool) *byte {
	pos := *ppos
	mask := (*[1 << 30]byte)(unsafe.Pointer(maskp))
	for {
		switch *prog {
		default:
			_lock.Throw("unrollgcprog: unknown instruction")

		case InsData:
			prog = _sched.Addb(prog, 1)
			siz := int(*prog)
			prog = _sched.Addb(prog, 1)
			p := (*[1 << 30]byte)(unsafe.Pointer(prog))
			for i := 0; i < siz; i++ {
				const typeBitsPerByte = 8 / _sched.TypeBitsWidth
				v := p[i/typeBitsPerByte]
				v >>= (uint(i) % typeBitsPerByte) * _sched.TypeBitsWidth
				v &= _sched.TypeMask
				if inplace {
					// Store directly into GC bitmap.
					h := _sched.HeapBitsForAddr(uintptr(unsafe.Pointer(&mask[pos])))
					if h.Shift == 0 {
						*h.Bitp = v << _sched.TypeShift
					} else {
						*h.Bitp |= v << (4 + _sched.TypeShift)
					}
					pos += _core.PtrSize
				} else if sparse {
					// 4-bits per word, type bits in high bits
					v <<= (pos % 8) + _sched.TypeShift
					mask[pos/8] |= v
					pos += _sched.HeapBitsWidth
				} else {
					// 2-bits per word
					v <<= pos % 8
					mask[pos/8] |= v
					pos += _sched.TypeBitsWidth
				}
			}
			prog = _sched.Addb(prog, _lock.Round(uintptr(siz)*_sched.TypeBitsWidth, 8)/8)

		case InsArray:
			prog = (*byte)(_core.Add(unsafe.Pointer(prog), 1))
			siz := uintptr(0)
			for i := uintptr(0); i < _core.PtrSize; i++ {
				siz = (siz << 8) + uintptr(*(*byte)(_core.Add(unsafe.Pointer(prog), _core.PtrSize-i-1)))
			}
			prog = (*byte)(_core.Add(unsafe.Pointer(prog), _core.PtrSize))
			var prog1 *byte
			for i := uintptr(0); i < siz; i++ {
				prog1 = Unrollgcprog1(&mask[0], prog, &pos, inplace, sparse)
			}
			if *prog1 != InsArrayEnd {
				_lock.Throw("unrollgcprog: array does not end with insArrayEnd")
			}
			prog = (*byte)(_core.Add(unsafe.Pointer(prog1), 1))

		case InsArrayEnd, InsEnd:
			*ppos = pos
			return prog
		}
	}
}

var unroll _core.Mutex

// Unrolls GC program in typ.gc[1] into typ.gc[0]
//go:nowritebarrier
func Unrollgcprog_m(typ *_core.Type) {
	_lock.Lock(&unroll)
	mask := (*byte)(unsafe.Pointer(uintptr(typ.Gc[0])))
	if *mask == 0 {
		pos := uintptr(8) // skip the unroll flag
		prog := (*byte)(unsafe.Pointer(uintptr(typ.Gc[1])))
		prog = Unrollgcprog1(mask, prog, &pos, false, true)
		if *prog != InsEnd {
			_lock.Throw("unrollgcprog: program does not end with insEnd")
		}
		if typ.Size/_core.PtrSize%2 != 0 {
			// repeat the program
			prog := (*byte)(unsafe.Pointer(uintptr(typ.Gc[1])))
			Unrollgcprog1(mask, prog, &pos, false, true)
		}

		// atomic way to say mask[0] = 1
		_sched.Atomicor8(mask, 1)
	}
	_lock.Unlock(&unroll)
}
