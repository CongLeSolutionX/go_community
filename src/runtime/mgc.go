// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(rsc): The code having to do with the heap bitmap needs very serious cleanup.
// It has gotten completely out of control.

// Garbage collector (GC).
//
// The GC runs concurrently with mutator threads, is type accurate (aka precise), allows multiple
// GC thread to run in parallel. It is a concurrent mark and sweep that uses a write barrier. It is
// non-generational and non-compacting. Allocation is done using size segregated per P allocation
// areas to minimize fragmentation while eliminating locks in the common case.
//
// The algorithm decomposes into several steps.
// This is a high level description of the algorithm being used. For an overview of GC a good
// place to start is Richard Jones' gchandbook.org.
//
// The algorithm's intellectual heritage includes Dijkstra's on-the-fly algorithm, see
// Edsger W. Dijkstra, Leslie Lamport, A. J. Martin, C. S. Scholten, and E. F. M. Steffens. 1978.
// On-the-fly garbage collection: an exercise in cooperation. Commun. ACM 21, 11 (November 1978),
// 966-975.
// For journal quality proofs that these steps are complete, correct, and terminate see
// Hudson, R., and Moss, J.E.B. Copying Garbage Collection without stopping the world.
// Concurrency and Computation: Practice and Experience 15(3-5), 2003.
//
//  0. Set phase = GCscan from GCoff.
//  1. Wait for all P's to acknowledge phase change.
//         At this point all goroutines have passed through a GC safepoint and
//         know we are in the GCscan phase.
//  2. GC scans all goroutine stacks, mark and enqueues all encountered pointers
//       (marking avoids most duplicate enqueuing but races may produce benign duplication).
//       Preempted goroutines are scanned before P schedules next goroutine.
//  3. Set phase = GCmark.
//  4. Wait for all P's to acknowledge phase change.
//  5. Now write barrier marks and enqueues black, grey, or white to white pointers.
//       Malloc still allocates white (non-marked) objects.
//  6. Meanwhile GC transitively walks the heap marking reachable objects.
//  7. When GC finishes marking heap, it preempts P's one-by-one and
//       retakes partial wbufs (filled by write barrier or during a stack scan of the goroutine
//       currently scheduled on the P).
//  8. Once the GC has exhausted all available marking work it sets phase = marktermination.
//  9. Wait for all P's to acknowledge phase change.
// 10. Malloc now allocates black objects, so number of unmarked reachable objects
//        monotonically decreases.
// 11. GC preempts P's one-by-one taking partial wbufs and marks all unmarked yet
//        reachable objects.
// 12. When GC completes a full cycle over P's and discovers no new grey
//         objects, (which means all reachable objects are marked) set phase = GCsweep.
// 13. Wait for all P's to acknowledge phase change.
// 14. Now malloc allocates white (but sweeps spans before use).
//         Write barrier becomes nop.
// 15. GC does background sweeping, see description below.
// 16. When sweeping is complete set phase to GCoff.
// 17. When sufficient allocation has taken place replay the sequence starting at 0 above,
//         see discussion of GC rate below.

// Changing phases.
// Phases are changed by setting the gcphase to the next phase and possibly calling ackgcphase.
// All phase action must be benign in the presence of a change.
// Starting with GCoff
// GCoff to GCscan
//     GSscan scans stacks and globals greying them and never marks an object black.
//     Once all the P's are aware of the new phase they will scan gs on preemption.
//     This means that the scanning of preempted gs can't start until all the Ps
//     have acknowledged.
// GCscan to GCmark
//     GCMark turns on the write barrier which also only greys objects. No scanning
//     of objects (making them black) can happen until all the Ps have acknowledged
//     the phase change.
// GCmark to GCmarktermination
//     The only change here is that we start allocating black so the Ps must acknowledge
//     the change before we begin the termination algorithm
// GCmarktermination to GSsweep
//     Object currently on the freelist must be marked black for this to work.
//     Are things on the free lists black or white? How does the sweep phase work?

// Concurrent sweep.
// The sweep phase proceeds concurrently with normal program execution.
// The heap is swept span-by-span both lazily (when a goroutine needs another span)
// and concurrently in a background goroutine (this helps programs that are not CPU bound).
// However, at the end of the stop-the-world GC phase we don't know the size of the live heap,
// and so next_gc calculation is tricky and happens as follows.
// At the end of the stop-the-world phase next_gc is conservatively set based on total
// heap size; all spans are marked as "needs sweeping".
// Whenever a span is swept, next_gc is decremented by GOGC*newly_freed_memory.
// The background sweeper goroutine simply sweeps spans one-by-one bringing next_gc
// closer to the target value. However, this is not enough to avoid over-allocating memory.
// Consider that a goroutine wants to allocate a new span for a large object and
// there are no free swept spans, but there are small-object unswept spans.
// If the goroutine naively allocates a new span, it can surpass the yet-unknown
// target next_gc value. In order to prevent such cases (1) when a goroutine needs
// to allocate a new small-object span, it sweeps small-object spans for the same
// object size until it frees at least one object; (2) when a goroutine needs to
// allocate large-object span from heap, it sweeps spans until it frees at least
// that many pages into heap. Together these two measures ensure that we don't surpass
// target next_gc value by a large margin. There is an exception: if a goroutine sweeps
// and frees two nonadjacent one-page spans to the heap, it will allocate a new two-page span,
// but there can still be other one-page unswept spans which could be combined into a
// two-page span.
// It's critical to ensure that no operations proceed on unswept spans (that would corrupt
// mark bits in GC bitmap). During GC all mcaches are flushed into the central cache,
// so they are empty. When a goroutine grabs a new span into mcache, it sweeps it.
// When a goroutine explicitly frees an object or sets a finalizer, it ensures that
// the span is swept (either by sweeping it, or by waiting for the concurrent sweep to finish).
// The finalizer goroutine is kicked off only when all spans are swept.
// When the next GC starts, it sweeps all not-yet-swept spans (if any).

// GC rate.
// Next GC is after we've allocated an extra amount of memory proportional to
// the amount already in use. The proportion is controlled by GOGC environment variable
// (100 by default). If GOGC=100 and we're using 4M, we'll GC again when we get to 8M
// (this mark is tracked in next_gc variable). This keeps the GC cost in linear
// proportion to the allocation cost. Adjusting GOGC just changes the linear constant
// (and also the amount of extra memory used).

package runtime

import (
	_cgo "runtime/internal/cgo"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

var badblock [1024]uintptr
var nbadblock int32

// Is _cgo_allocate linked into the binary?
//go:nowritebarrier
func have_cgo_allocate() bool {
	return &_cgo.Weak_cgo_allocate != nil
}

// If the slot is grey or black return true, if white return false.
// If the slot is not in the known heap and thus does not have a valid GC bitmap then
// it is considered grey. Globals and stacks can hold such slots.
// The slot is grey if its mark bit is set and it is enqueued to be scanned.
// The slot is black if it has already been scanned.
// It is white if it has a valid mark bit and the bit is not set.
//go:nowritebarrier
func shaded(slot uintptr) bool {
	if !_sched.Inheap(slot) { // non-heap slots considered grey
		return true
	}

	var mbits _sched.Markbits
	valid := _sched.Objectstart(slot, &mbits)
	if valid == 0 {
		return true
	}

	if _sched.Checkmark {
		return _sched.Ischeckmarked(&mbits)
	}

	return mbits.Bits&_sched.BitMarked != 0
}

//go:nowritebarrier
func iterate_finq(callback func(*_core.Funcval, unsafe.Pointer, uintptr, *_core.Type, *_gc.Ptrtype)) {
	for fb := _gc.Allfin; fb != nil; fb = fb.Alllink {
		for i := int32(0); i < fb.Cnt; i++ {
			f := &fb.Fin[i]
			callback(f.Fn, f.Arg, f.Nret, f.Fint, f.Ot)
		}
	}
}

//go:nowritebarrier
func gccheckmarkenable_m() {
	_gc.Gccheckmarkenable = true
}

//go:nowritebarrier
func gccheckmarkdisable_m() {
	_gc.Gccheckmarkenable = false
}

// For now this must be bracketed with a stoptheworld and a starttheworld to ensure
// all go routines see the new barrier.
//go:nowritebarrier
func gcinstallmarkwb_m() {
	_sched.Gcphase = _sched.GCmark
}

//go:linkname readGCStats runtime/debug.readGCStats
func readGCStats(pauses *[]uint64) {
	_lock.Systemstack(func() {
		readGCStats_m(pauses)
	})
}

func readGCStats_m(pauses *[]uint64) {
	p := *pauses
	// Calling code in runtime/debug should make the slice large enough.
	if cap(p) < len(_lock.Memstats.Pause_ns)+3 {
		_lock.Throw("runtime: short slice passed to readGCStats")
	}

	// Pass back: pauses, pause ends, last gc (absolute time), number of gc, total pause ns.
	_lock.Lock(&_lock.Mheap_.Lock)

	n := _lock.Memstats.Numgc
	if n > uint32(len(_lock.Memstats.Pause_ns)) {
		n = uint32(len(_lock.Memstats.Pause_ns))
	}

	// The pause buffer is circular. The most recent pause is at
	// pause_ns[(numgc-1)%len(pause_ns)], and then backward
	// from there to go back farther in time. We deliver the times
	// most recent first (in p[0]).
	p = p[:cap(p)]
	for i := uint32(0); i < n; i++ {
		j := (_lock.Memstats.Numgc - 1 - i) % uint32(len(_lock.Memstats.Pause_ns))
		p[i] = _lock.Memstats.Pause_ns[j]
		p[n+i] = _lock.Memstats.Pause_end[j]
	}

	p[n+n] = _lock.Memstats.Last_gc
	p[n+n+1] = uint64(_lock.Memstats.Numgc)
	p[n+n+2] = _lock.Memstats.Pause_total_ns
	_lock.Unlock(&_lock.Mheap_.Lock)
	*pauses = p[:n+n+3]
}

func setGCPercent(in int32) (out int32) {
	_lock.Lock(&_lock.Mheap_.Lock)
	out = _gc.Gcpercent
	if in < 0 {
		in = -1
	}
	_gc.Gcpercent = in
	_lock.Unlock(&_lock.Mheap_.Lock)
	return out
}

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

	// data
	if uintptr(unsafe.Pointer(&_gc.Data)) <= uintptr(p) && uintptr(p) < uintptr(unsafe.Pointer(&_gc.Edata)) {
		n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			off := (uintptr(p) + i - uintptr(unsafe.Pointer(&_gc.Data))) / _core.PtrSize
			bits := (*(*byte)(_core.Add(unsafe.Pointer(_gc.Gcdatamask.Bytedata), off/_sched.XPointersPerByte)) >> ((off % _sched.XPointersPerByte) * _sched.XBitsPerPointer)) & _sched.XBitsMask
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
			bits := (*(*byte)(_core.Add(unsafe.Pointer(_gc.Gcbssmask.Bytedata), off/_sched.XPointersPerByte)) >> ((off % _sched.XPointersPerByte) * _sched.XBitsPerPointer)) & _sched.XBitsMask
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
			off := (uintptr(base) + i - _lock.Mheap_.Arena_start) / _core.PtrSize
			b := _lock.Mheap_.Arena_start - off/_sched.WordsPerBitmapByte - 1
			shift := (off % _sched.WordsPerBitmapByte) * _sched.GcBits
			bits := (*(*byte)(unsafe.Pointer(b)) >> (shift + 2)) & _sched.XBitsMask
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
		size := uintptr(bv.N) / _sched.XBitsPerPointer * _core.PtrSize
		n := (*_gc.Ptrtype)(unsafe.Pointer(t)).Elem.Size
		*len = n / _core.PtrSize
		*mask = &make([]byte, *len)[0]
		for i := uintptr(0); i < n; i += _core.PtrSize {
			off := (uintptr(p) + i - frame.Varp + size) / _core.PtrSize
			bits := ((*(*byte)(_core.Add(unsafe.Pointer(bv.Bytedata), off*_sched.XBitsPerPointer/8))) >> ((off * _sched.XBitsPerPointer) % 8)) & _sched.XBitsMask
			*(*byte)(_core.Add(unsafe.Pointer(*mask), i/_core.PtrSize)) = bits
		}
	}
}
