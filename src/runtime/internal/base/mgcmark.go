// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: marking and scanning

package base

import (
	"unsafe"
)

// gcMaxStackBarriers returns the maximum number of stack barriers
// that can be installed in a stack of stackSize bytes.
func gcMaxStackBarriers(stackSize int) (n int) {
	if FirstStackBarrierOffset == 0 {
		// Special debugging case for inserting stack barriers
		// at every frame. Steal half of the stack for the
		// []stkbar. Technically, if the stack were to consist
		// solely of return PCs we would need two thirds of
		// the stack, but stealing that much breaks things and
		// this doesn't happen in practice.
		return stackSize / 2 / int(unsafe.Sizeof(Stkbar{}))
	}

	offset := FirstStackBarrierOffset
	for offset < stackSize {
		n++
		offset *= 2
	}
	return n + 1
}

// gcPrintStkbars prints a []stkbar for debugging.
func GcPrintStkbars(stkbar []Stkbar) {
	print("[")
	for i, s := range stkbar {
		if i > 0 {
			print(" ")
		}
		print("*", Hex(s.SavedLRPtr), "=", Hex(s.SavedLRVal))
	}
	print("]")
}

// TODO(austin): Can we consolidate the gcDrain* functions?

// gcDrain scans objects in work buffers, blackening grey
// objects until all work buffers have been drained.
// If flushScanCredit != -1, gcDrain flushes accumulated scan work
// credit to gcController.bgScanCredit whenever gcw's local scan work
// credit exceeds flushScanCredit.
//go:nowritebarrier
func GcDrain(gcw *GcWork, flushScanCredit int64) {
	if !WriteBarrierEnabled {
		Throw("gcDrain phase incorrect")
	}

	var lastScanFlush, nextScanFlush int64
	if flushScanCredit != -1 {
		lastScanFlush = gcw.ScanWork
		nextScanFlush = lastScanFlush + flushScanCredit
	} else {
		nextScanFlush = int64(^uint64(0) >> 1)
	}

	for {
		// If another proc wants a pointer, give it some.
		if Work.Nwait > 0 && Work.Full == 0 {
			gcw.Balance()
		}

		b := gcw.get()
		if b == 0 {
			// work barrier reached
			break
		}
		// If the current wbuf is filled by the scan a new wbuf might be
		// returned that could possibly hold only a single object. This
		// could result in each iteration draining only a single object
		// out of the wbuf passed in + a single object placed
		// into an empty wbuf in scanobject so there could be
		// a performance hit as we keep fetching fresh wbufs.
		Scanobject(b, gcw)

		// Flush background scan work credit to the global
		// account if we've accumulated enough locally so
		// mutator assists can draw on it.
		if gcw.ScanWork >= nextScanFlush {
			credit := gcw.ScanWork - lastScanFlush
			Xaddint64(&GcController.BgScanCredit, credit)
			lastScanFlush = gcw.ScanWork
			nextScanFlush = lastScanFlush + flushScanCredit
		}
	}
	if flushScanCredit != -1 {
		credit := gcw.ScanWork - lastScanFlush
		Xaddint64(&GcController.BgScanCredit, credit)
	}
}

// scanobject scans the object starting at b, adding pointers to gcw.
// b must point to the beginning of a heap object; scanobject consults
// the GC bitmap for the pointer mask and the spans for the size of the
// object (it ignores n).
//go:nowritebarrier
func Scanobject(b uintptr, gcw *GcWork) {
	// Note that arena_used may change concurrently during
	// scanobject and hence scanobject may encounter a pointer to
	// a newly allocated heap object that is *not* in
	// [start,used). It will not mark this object; however, we
	// know that it was just installed by a mutator, which means
	// that mutator will execute a write barrier and take care of
	// marking it. This is even more pronounced on relaxed memory
	// architectures since we access arena_used without barriers
	// or synchronization, but the same logic applies.
	arena_start := Mheap_.Arena_start
	arena_used := Mheap_.Arena_used

	// Find bits of the beginning of the object.
	// b must point to the beginning of a heap object, so
	// we can get its bits and span directly.
	hbits := HeapBitsForAddr(b)
	s := SpanOfUnchecked(b)
	n := s.Elemsize
	if n == 0 {
		Throw("scanobject n == 0")
	}

	var i uintptr
	for i = 0; i < n; i += PtrSize {
		// Find bits for this word.
		if i != 0 {
			// Avoid needless hbits.next() on last iteration.
			hbits = hbits.Next()
		}
		// During checkmarking, 1-word objects store the checkmark
		// in the type bit for the one word. The only one-word objects
		// are pointers, or else they'd be merged with other non-pointer
		// data into larger allocations.
		bits := hbits.bits()
		if i >= 2*PtrSize && bits&BitMarked == 0 {
			break // no more pointers in this object
		}
		if bits&BitPointer == 0 {
			continue // not a pointer
		}

		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		obj := *(*uintptr)(unsafe.Pointer(b + i))

		// At this point we have extracted the next potential pointer.
		// Check if it points into heap and not back at the current object.
		if obj != 0 && arena_start <= obj && obj < arena_used && obj-b >= n {
			// Mark the object.
			if obj, hbits, span := HeapBitsForObject(obj); obj != 0 {
				Greyobject(obj, b, i, hbits, span, gcw)
			}
		}
	}
	gcw.bytesMarked += uint64(n)
	gcw.ScanWork += int64(i)
}

// Shade the object if it isn't already.
// The object is not nil and known to be in the heap.
// Preemption must be disabled.
//go:nowritebarrier
func shade(b uintptr) {
	if obj, hbits, span := HeapBitsForObject(b); obj != 0 {
		gcw := &Getg().M.P.Ptr().Gcw
		Greyobject(obj, 0, 0, hbits, span, gcw)
		if Gcphase == GCmarktermination || GcBlackenPromptly {
			// Ps aren't allowed to cache work during mark
			// termination.
			gcw.Dispose()
		}
	}
}

// obj is the start of an object with mark mbits.
// If it isn't already marked, mark it and enqueue into gcw.
// base and off are for debugging only and could be removed.
//go:nowritebarrier
func Greyobject(obj, base, off uintptr, hbits HeapBits, span *Mspan, gcw *GcWork) {
	// obj should be start of allocation, and so must be at least pointer-aligned.
	if obj&(PtrSize-1) != 0 {
		Throw("greyobject: obj not pointer-aligned")
	}

	if UseCheckmark {
		if !hbits.IsMarked() {
			Printlock()
			print("runtime:greyobject: checkmarks finds unexpected unmarked object obj=", Hex(obj), "\n")
			print("runtime: found obj at *(", Hex(base), "+", Hex(off), ")\n")

			// Dump the source (base) object
			gcDumpObject("base", base, off)

			// Dump the object
			gcDumpObject("obj", obj, ^uintptr(0))

			Throw("checkmark found unmarked object")
		}
		if hbits.isCheckmarked(span.Elemsize) {
			return
		}
		hbits.setCheckmarked(span.Elemsize)
		if !hbits.isCheckmarked(span.Elemsize) {
			Throw("setCheckmarked and isCheckmarked disagree")
		}
	} else {
		// If marked we have nothing to do.
		if hbits.IsMarked() {
			return
		}
		hbits.SetMarked()

		// If this is a noscan object, fast-track it to black
		// instead of greying it.
		if !hbits.hasPointers(span.Elemsize) {
			gcw.bytesMarked += uint64(span.Elemsize)
			return
		}
	}

	// Queue the obj for scanning. The PREFETCH(obj) logic has been removed but
	// seems like a nice optimization that can be added back in.
	// There needs to be time between the PREFETCH and the use.
	// Previously we put the obj in an 8 element buffer that is drained at a rate
	// to give the PREFETCH time to do its work.
	// Use of PREFETCHNTA might be more appropriate than PREFETCH

	gcw.put(obj)
}

// gcDumpObject dumps the contents of obj for debugging and marks the
// field at byte offset off in obj.
func gcDumpObject(label string, obj, off uintptr) {
	if obj < Mheap_.Arena_start || obj >= Mheap_.Arena_used {
		print(label, "=", Hex(obj), " is not a heap object\n")
		return
	}
	k := obj >> PageShift
	x := k
	x -= Mheap_.Arena_start >> PageShift
	s := H_spans[x]
	print(label, "=", Hex(obj), " k=", Hex(k))
	if s == nil {
		print(" s=nil\n")
		return
	}
	print(" s.start*_PageSize=", Hex(s.Start*PageSize), " s.limit=", Hex(s.Limit), " s.sizeclass=", s.Sizeclass, " s.elemsize=", s.Elemsize, "\n")
	for i := uintptr(0); i < s.Elemsize; i += PtrSize {
		print(" *(", label, "+", i, ") = ", Hex(*(*uintptr)(unsafe.Pointer(obj + uintptr(i)))))
		if i == off {
			print(" <==")
		}
		print("\n")
	}
}

// Checkmarking

// To help debug the concurrent GC we remark with the world
// stopped ensuring that any object encountered has their normal
// mark bit set. To do this we use an orthogonal bit
// pattern to indicate the object is marked. The following pattern
// uses the upper two bits in the object's boundary nibble.
// 01: scalar  not marked
// 10: pointer not marked
// 11: pointer     marked
// 00: scalar      marked
// Xoring with 01 will flip the pattern from marked to unmarked and vica versa.
// The higher bit is 1 for pointers and 0 for scalars, whether the object
// is marked or not.
// The first nibble no longer holds the typeDead pattern indicating that the
// there are no more pointers in the object. This information is held
// in the second nibble.

// If useCheckmark is true, marking of an object uses the
// checkmark bits (encoding above) instead of the standard
// mark bits.
var UseCheckmark = false
