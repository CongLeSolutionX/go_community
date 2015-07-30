// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: type and heap bitmaps.
//
// Stack, data, and bss bitmaps
//
// Stack frames and global variables in the data and bss sections are described
// by 1-bit bitmaps in which 0 means uninteresting and 1 means live pointer
// to be visited during GC. The bits in each byte are consumed starting with
// the low bit: 1<<0, 1<<1, and so on.
//
// Heap bitmap
//
// The allocated heap comes from a subset of the memory in the range [start, used),
// where start == mheap_.arena_start and used == mheap_.arena_used.
// The heap bitmap comprises 2 bits for each pointer-sized word in that range,
// stored in bytes indexed backward in memory from start.
// That is, the byte at address start-1 holds the 2-bit entries for the four words
// start through start+3*ptrSize, the byte at start-2 holds the entries for
// start+4*ptrSize through start+7*ptrSize, and so on.
//
// In each 2-bit entry, the lower bit holds the same information as in the 1-bit
// bitmaps: 0 means uninteresting and 1 means live pointer to be visited during GC.
// The meaning of the high bit depends on the position of the word being described
// in its allocated object. In the first word, the high bit is the GC ``marked'' bit.
// In the second word, the high bit is the GC ``checkmarked'' bit (see below).
// In the third and later words, the high bit indicates that the object is still
// being described. In these words, if a bit pair with a high bit 0 is encountered,
// the low bit can also be assumed to be 0, and the object description is over.
// This 00 is called the ``dead'' encoding: it signals that the rest of the words
// in the object are uninteresting to the garbage collector.
//
// The 2-bit entries are split when written into the byte, so that the top half
// of the byte contains 4 mark bits and the bottom half contains 4 pointer bits.
// This form allows a copy from the 1-bit to the 4-bit form to keep the
// pointer bits contiguous, instead of having to space them out.
//
// The code makes use of the fact that the zero value for a heap bitmap
// has no live pointer bit set and is (depending on position), not marked,
// not checkmarked, and is the dead encoding.
// These properties must be preserved when modifying the encoding.
//
// Checkmarks
//
// In a concurrent garbage collector, one worries about failing to mark
// a live object due to mutations without write barriers or bugs in the
// collector implementation. As a sanity check, the GC has a 'checkmark'
// mode that retraverses the object graph with the world stopped, to make
// sure that everything that should be marked is marked.
// In checkmark mode, in the heap bitmap, the high bit of the 2-bit entry
// for the second word of the object holds the checkmark bit.
// When not in checkmark mode, this bit is set to 1.
//
// The smallest possible allocation is 8 bytes. On a 32-bit machine, that
// means every allocated object has two words, so there is room for the
// checkmark bit. On a 64-bit machine, however, the 8-byte allocation is
// just one word, so the second bit pair is not available for encoding the
// checkmark. However, because non-pointer allocations are combined
// into larger 16-byte (maxTinySize) allocations, a plain 8-byte allocation
// must be a pointer, so the type bit in the first word is not actually needed.
// It is still used in general, except in checkmark the type bit is repurposed
// as the checkmark bit and then reinitialized (to 1) as the type bit when
// finished.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	"unsafe"
)

// typeBitsBulkBarrier executes writebarrierptr_nostore
// for every pointer slot in the memory range [p, p+size),
// using the type bitmap to locate those pointer slots.
// The type typ must correspond exactly to [p, p+size).
// This executes the write barriers necessary after a copy.
// Both p and size must be pointer-aligned.
// The type typ must have a plain bitmap, not a GC program.
// The only use of this function is in channel sends, and the
// 64 kB channel element limit takes care of this for us.
//
// Must not be preempted because it typically runs right after memmove,
// and the GC must not complete between those two.
//
//go:nosplit
func typeBitsBulkBarrier(typ *_base.Type, p, size uintptr) {
	if typ == nil {
		_base.Throw("runtime: typeBitsBulkBarrier without type")
	}
	if typ.Size != size {
		println("runtime: typeBitsBulkBarrier with type ", *typ.String, " of size ", typ.Size, " but memory size", size)
		_base.Throw("runtime: invalid typeBitsBulkBarrier")
	}
	if typ.Kind&_iface.KindGCProg != 0 {
		println("runtime: typeBitsBulkBarrier with type ", *typ.String, " with GC prog")
		_base.Throw("runtime: invalid typeBitsBulkBarrier")
	}
	if !_base.WriteBarrierEnabled {
		return
	}
	ptrmask := typ.Gcdata
	var bits uint32
	for i := uintptr(0); i < typ.Ptrdata; i += _base.PtrSize {
		if i&(_base.PtrSize*8-1) == 0 {
			bits = uint32(*ptrmask)
			ptrmask = _gc.Addb(ptrmask, 1)
		} else {
			bits = bits >> 1
		}
		if bits&1 != 0 {
			x := (*uintptr)(unsafe.Pointer(p + i))
			_base.Writebarrierptr_nostore(x, *x)
		}
	}
}

// Testing.

func getgcmaskcb(frame *_base.Stkframe, ctxt unsafe.Pointer) bool {
	target := (*_base.Stkframe)(ctxt)
	if frame.Sp <= target.Sp && target.Sp < frame.Varp {
		*target = *frame
		return false
	}
	return true
}

// gcbits returns the GC type info for x, for testing.
// The result is the bitmap entries (0 or 1), one entry per byte.
//go:linkname reflect_gcbits reflect.gcbits
func reflect_gcbits(x interface{}) []byte {
	ret := getgcmask(x)
	typ := (*_gc.Ptrtype)(unsafe.Pointer((*_iface.Eface)(unsafe.Pointer(&x)).Type)).Elem
	nptr := typ.Ptrdata / _base.PtrSize
	for uintptr(len(ret)) > nptr && ret[len(ret)-1] == 0 {
		ret = ret[:len(ret)-1]
	}
	return ret
}

// Returns GC type info for object p for testing.
func getgcmask(ep interface{}) (mask []byte) {
	e := *(*_iface.Eface)(unsafe.Pointer(&ep))
	p := e.Data
	t := e.Type
	// data or bss
	for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
		// data
		if datap.Data <= uintptr(p) && uintptr(p) < datap.Edata {
			bitmap := datap.Gcdatamask.Bytedata
			n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
			mask = make([]byte, n/_base.PtrSize)
			for i := uintptr(0); i < n; i += _base.PtrSize {
				off := (uintptr(p) + i - datap.Data) / _base.PtrSize
				mask[i/_base.PtrSize] = (*_gc.Addb(bitmap, off/8) >> (off % 8)) & 1
			}
			return
		}

		// bss
		if datap.Bss <= uintptr(p) && uintptr(p) < datap.Ebss {
			bitmap := datap.Gcbssmask.Bytedata
			n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
			mask = make([]byte, n/_base.PtrSize)
			for i := uintptr(0); i < n; i += _base.PtrSize {
				off := (uintptr(p) + i - datap.Bss) / _base.PtrSize
				mask[i/_base.PtrSize] = (*_gc.Addb(bitmap, off/8) >> (off % 8)) & 1
			}
			return
		}
	}

	// heap
	var n uintptr
	var base uintptr
	if mlookup(uintptr(p), &base, &n, nil) != 0 {
		mask = make([]byte, n/_base.PtrSize)
		for i := uintptr(0); i < n; i += _base.PtrSize {
			hbits := _base.HeapBitsForAddr(base + i)
			if hbits.IsPointer() {
				mask[i/_base.PtrSize] = 1
			}
			if i >= 2*_base.PtrSize && !hbits.IsMarked() {
				mask = mask[:i/_base.PtrSize]
				break
			}
		}
		return
	}

	// stack
	if _g_ := _base.Getg(); _g_.M.Curg.Stack.Lo <= uintptr(p) && uintptr(p) < _g_.M.Curg.Stack.Hi {
		var frame _base.Stkframe
		frame.Sp = uintptr(p)
		_g_ := _base.Getg()
		_base.Gentraceback(_g_.M.Curg.Sched.Pc, _g_.M.Curg.Sched.Sp, 0, _g_.M.Curg, 0, nil, 1000, getgcmaskcb, _base.Noescape(unsafe.Pointer(&frame)), 0)
		if frame.Fn != nil {
			f := frame.Fn
			targetpc := frame.Continpc
			if targetpc == 0 {
				return
			}
			if targetpc != f.Entry {
				targetpc--
			}
			pcdata := _gc.Pcdatavalue(f, _base.PCDATA_StackMapIndex, targetpc)
			if pcdata == -1 {
				return
			}
			stkmap := (*_gc.Stackmap)(_gc.Funcdata(f, _base.FUNCDATA_LocalsPointerMaps))
			if stkmap == nil || stkmap.N <= 0 {
				return
			}
			bv := _gc.Stackmapdata(stkmap, pcdata)
			size := uintptr(bv.N) * _base.PtrSize
			n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
			mask = make([]byte, n/_base.PtrSize)
			for i := uintptr(0); i < n; i += _base.PtrSize {
				bitmap := bv.Bytedata
				off := (uintptr(p) + i - frame.Varp + size) / _base.PtrSize
				mask[i/_base.PtrSize] = (*_gc.Addb(bitmap, off/8) >> (off % 8)) & 1
			}
		}
		return
	}

	// otherwise, not something the GC knows about.
	// possibly read-only data, like malloc(0).
	// must not have pointers
	return
}
