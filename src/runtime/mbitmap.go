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

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Testing.

func getgcmaskcb(frame *_lock.Stkframe, ctxt unsafe.Pointer) bool {
	target := (*_lock.Stkframe)(ctxt)
	if frame.Sp <= target.Sp && target.Sp < frame.Varp {
		*target = *frame
		return false
	}
	return true
}

// Returns GC type info for object p for testing.
func getgcmask(p unsafe.Pointer, t *_core.Type, mask **byte, len *uintptr) {
	*mask = nil
	*len = 0

	const typeBitsPerByte = 8 / _sched.TypeBitsWidth

	// data
	if uintptr(unsafe.Pointer(&_gc.Data)) <= uintptr(p) && uintptr(p) < uintptr(unsafe.Pointer(&_gc.Edata)) {
		n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			off := (uintptr(p) + i - uintptr(unsafe.Pointer(&_gc.Data))) / _core.PtrSize
			bits := (*(*byte)(_core.Add(unsafe.Pointer(_gc.Gcdatamask.Bytedata), off/typeBitsPerByte)) >> ((off % typeBitsPerByte) * _sched.TypeBitsWidth)) & _sched.TypeMask
			*(*byte)(_core.Add(unsafe.Pointer(*mask), i/_core.PtrSize)) = bits
		}
		return
	}

	// bss
	if uintptr(unsafe.Pointer(&_gc.Bss)) <= uintptr(p) && uintptr(p) < uintptr(unsafe.Pointer(&_gc.Ebss)) {
		n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			off := (uintptr(p) + i - uintptr(unsafe.Pointer(&_gc.Bss))) / _core.PtrSize
			bits := (*(*byte)(_core.Add(unsafe.Pointer(_gc.Gcbssmask.Bytedata), off/typeBitsPerByte)) >> ((off % typeBitsPerByte) * _sched.TypeBitsWidth)) & _sched.TypeMask
			*(*byte)(_core.Add(unsafe.Pointer(*mask), i/_core.PtrSize)) = bits
		}
		return
	}

	// heap
	var n uintptr
	var base uintptr
	if mlookup(uintptr(p), &base, &n, nil) != 0 {
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			bits := _sched.HeapBitsForAddr(base + i).TypeBits()
			*(*byte)(_core.Add(unsafe.Pointer(*mask), i/_core.PtrSize)) = bits
		}
		return
	}

	// stack
	var frame _lock.Stkframe
	frame.Sp = uintptr(p)
	_g_ := _core.Getg()
	_lock.Gentraceback(_g_.M.Curg.Sched.Pc, _g_.M.Curg.Sched.Sp, 0, _g_.M.Curg, 0, nil, 1000, getgcmaskcb, _core.Noescape(unsafe.Pointer(&frame)), 0)
	if frame.Fn != nil {
		f := frame.Fn
		targetpc := frame.Continpc
		if targetpc == 0 {
			return
		}
		if targetpc != f.Entry {
			targetpc--
		}
		pcdata := _gc.Pcdatavalue(f, _lock.PCDATA_StackMapIndex, targetpc)
		if pcdata == -1 {
			return
		}
		stkmap := (*_gc.Stackmap)(_gc.Funcdata(f, _lock.FUNCDATA_LocalsPointerMaps))
		if stkmap == nil || stkmap.N <= 0 {
			return
		}
		bv := _gc.Stackmapdata(stkmap, pcdata)
		size := uintptr(bv.N) / _sched.TypeBitsWidth * _core.PtrSize
		n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			off := (uintptr(p) + i - frame.Varp + size) / _core.PtrSize
			bits := ((*(*byte)(_core.Add(unsafe.Pointer(bv.Bytedata), off*_sched.TypeBitsWidth/8))) >> ((off * _sched.TypeBitsWidth) % 8)) & _sched.TypeMask
			*(*byte)(_core.Add(unsafe.Pointer(*mask), i/_core.PtrSize)) = bits
		}
	}
}
