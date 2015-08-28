// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: sweeping

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

var Sweep Sweepdata

// State of background sweep.
type Sweepdata struct {
	Lock    _base.Mutex
	G       *_base.G
	Parked  bool
	started bool

	spanidx uint32 // background sweeper position

	Nbgsweep    uint32
	npausesweep uint32
}

//go:nowritebarrier
func finishsweep_m() {
	// The world is stopped so we should be able to complete the sweeps
	// quickly.
	for Sweepone() != ^uintptr(0) {
		Sweep.npausesweep++
	}

	// There may be some other spans being swept concurrently that
	// we need to wait for. If finishsweep_m is done with the world stopped
	// this code is not required.
	sg := _base.Mheap_.Sweepgen
	for _, s := range _base.Work.Spans {
		if s.Sweepgen != sg && s.State == _base.XMSpanInUse {
			MSpan_EnsureSwept(s)
		}
	}
}

// sweeps one span
// returns number of pages returned to heap, or ^uintptr(0) if there is nothing to sweep
//go:nowritebarrier
func Sweepone() uintptr {
	_g_ := _base.Getg()

	// increment locks to ensure that the goroutine is not preempted
	// in the middle of sweep thus leaving the span in an inconsistent state for next GC
	_g_.M.Locks++
	sg := _base.Mheap_.Sweepgen
	for {
		idx := _base.Xadd(&Sweep.spanidx, 1) - 1
		if idx >= uint32(len(_base.Work.Spans)) {
			_base.Mheap_.Sweepdone = 1
			_g_.M.Locks--
			return ^uintptr(0)
		}
		s := _base.Work.Spans[idx]
		if s.State != _base.MSpanInUse {
			s.Sweepgen = sg
			continue
		}
		if s.Sweepgen != sg-2 || !_base.Cas(&s.Sweepgen, sg-2, sg-1) {
			continue
		}
		npages := s.Npages
		if !MSpan_Sweep(s, false) {
			npages = 0
		}
		_g_.M.Locks--
		return npages
	}
}

//go:nowritebarrier
func Gosweepone() uintptr {
	var ret uintptr
	_base.Systemstack(func() {
		ret = Sweepone()
	})
	return ret
}

// Returns only when span s has been swept.
//go:nowritebarrier
func MSpan_EnsureSwept(s *_base.Mspan) {
	// Caller must disable preemption.
	// Otherwise when this function returns the span can become unswept again
	// (if GC is triggered on another goroutine).
	_g_ := _base.Getg()
	if _g_.M.Locks == 0 && _g_.M.Mallocing == 0 && _g_ != _g_.M.G0 {
		_base.Throw("MSpan_EnsureSwept: m is not locked")
	}

	sg := _base.Mheap_.Sweepgen
	if _base.Atomicload(&s.Sweepgen) == sg {
		return
	}
	// The caller must be sure that the span is a MSpanInUse span.
	if _base.Cas(&s.Sweepgen, sg-2, sg-1) {
		MSpan_Sweep(s, false)
		return
	}
	// unfortunate condition, and we don't have efficient means to wait
	for _base.Atomicload(&s.Sweepgen) != sg {
		_base.Osyield()
	}
}

// Sweep frees or collects finalizers for blocks not marked in the mark phase.
// It clears the mark bits in preparation for the next GC round.
// Returns true if the span was returned to heap.
// If preserve=true, don't return it to heap nor relink in MCentral lists;
// caller takes care of it.
//TODO go:nowritebarrier
func MSpan_Sweep(s *_base.Mspan, preserve bool) bool {
	// It's critical that we enter this function with preemption disabled,
	// GC must not start while we are in the middle of this function.
	_g_ := _base.Getg()
	if _g_.M.Locks == 0 && _g_.M.Mallocing == 0 && _g_ != _g_.M.G0 {
		_base.Throw("MSpan_Sweep: m is not locked")
	}
	sweepgen := _base.Mheap_.Sweepgen
	if s.State != _base.MSpanInUse || s.Sweepgen != sweepgen-1 {
		print("MSpan_Sweep: state=", s.State, " sweepgen=", s.Sweepgen, " mheap.sweepgen=", sweepgen, "\n")
		_base.Throw("MSpan_Sweep: bad span state")
	}

	if _base.Trace.Enabled {
		traceGCSweepStart()
	}

	_base.Xadd64(&_base.Mheap_.PagesSwept, int64(s.Npages))

	cl := s.Sizeclass
	size := s.Elemsize
	res := false
	nfree := 0

	var head, end _base.Gclinkptr

	c := _g_.M.Mcache
	freeToHeap := false

	// Mark any free objects in this span so we don't collect them.
	sstart := uintptr(s.Start << _base.PageShift)
	for link := s.Freelist; link.Ptr() != nil; link = link.Ptr().Next {
		if uintptr(link) < sstart || s.Limit <= uintptr(link) {
			// Free list is corrupted.
			dumpFreeList(s)
			_base.Throw("free list corrupted")
		}
		_base.HeapBitsForAddr(uintptr(link)).SetMarkedNonAtomic()
	}

	// Unlink & free special records for any objects we're about to free.
	specialp := &s.Specials
	special := *specialp
	for special != nil {
		// A finalizer can be set for an inner byte of an object, find object beginning.
		p := uintptr(s.Start<<_base.PageShift) + uintptr(special.Offset)/size*size
		hbits := _base.HeapBitsForAddr(p)
		if !hbits.IsMarked() {
			// Find the exact byte for which the special was setup
			// (as opposed to object beginning).
			p := uintptr(s.Start<<_base.PageShift) + uintptr(special.Offset)
			// about to free object: splice out special record
			y := special
			special = special.Next
			*specialp = special
			if !freespecial(y, unsafe.Pointer(p), size, false) {
				// stop freeing of object if it has a finalizer
				hbits.SetMarkedNonAtomic()
			}
		} else {
			// object is still live: keep special record
			specialp = &special.Next
			special = *specialp
		}
	}

	// Sweep through n objects of given size starting at p.
	// This thread owns the span now, so it can manipulate
	// the block bitmap without atomic operations.

	size, n, _ := s.Layout()
	heapBitsSweepSpan(s.Base(), size, n, func(p uintptr) {
		// At this point we know that we are looking at garbage object
		// that needs to be collected.
		if _base.Debug.Allocfreetrace != 0 {
			tracefree(unsafe.Pointer(p), size)
		}

		// Reset to allocated+noscan.
		if cl == 0 {
			// Free large span.
			if preserve {
				_base.Throw("can't preserve large span")
			}
			HeapBitsForSpan(p).InitSpan(s.Layout())
			s.Needzero = 1

			// important to set sweepgen before returning it to heap
			_base.Atomicstore(&s.Sweepgen, sweepgen)

			// Free the span after heapBitsSweepSpan
			// returns, since it's not done with the span.
			freeToHeap = true
		} else {
			// Free small object.
			if size > 2*_base.PtrSize {
				*(*uintptr)(unsafe.Pointer(p + _base.PtrSize)) = _base.UintptrMask & 0xdeaddeaddeaddead // mark as "needs to be zeroed"
			} else if size > _base.PtrSize {
				*(*uintptr)(unsafe.Pointer(p + _base.PtrSize)) = 0
			}
			if head.Ptr() == nil {
				head = _base.Gclinkptr(p)
			} else {
				end.Ptr().Next = _base.Gclinkptr(p)
			}
			end = _base.Gclinkptr(p)
			end.Ptr().Next = _base.Gclinkptr(0x0bade5)
			nfree++
		}
	})

	// We need to set s.sweepgen = h.sweepgen only when all blocks are swept,
	// because of the potential for a concurrent free/SetFinalizer.
	// But we need to set it before we make the span available for allocation
	// (return it to heap or mcentral), because allocation code assumes that a
	// span is already swept if available for allocation.
	//
	// TODO(austin): Clean this up by consolidating atomicstore in
	// large span path above with this.
	if !freeToHeap && nfree == 0 {
		// The span must be in our exclusive ownership until we update sweepgen,
		// check for potential races.
		if s.State != _base.MSpanInUse || s.Sweepgen != sweepgen-1 {
			print("MSpan_Sweep: state=", s.State, " sweepgen=", s.Sweepgen, " mheap.sweepgen=", sweepgen, "\n")
			_base.Throw("MSpan_Sweep: bad span state after sweep")
		}
		_base.Atomicstore(&s.Sweepgen, sweepgen)
	}
	if nfree > 0 {
		c.Local_nsmallfree[cl] += uintptr(nfree)
		res = mCentral_FreeSpan(&_base.Mheap_.Central[cl].Mcentral, s, int32(nfree), head, end, preserve)
		// MCentral_FreeSpan updates sweepgen
	} else if freeToHeap {
		// Free large span to heap

		// NOTE(rsc,dvyukov): The original implementation of efence
		// in CL 22060046 used SysFree instead of SysFault, so that
		// the operating system would eventually give the memory
		// back to us again, so that an efence program could run
		// longer without running out of memory. Unfortunately,
		// calling SysFree here without any kind of adjustment of the
		// heap data structures means that when the memory does
		// come back to us, we have the wrong metadata for it, either in
		// the MSpan structures or in the garbage collection bitmap.
		// Using SysFault here means that the program will run out of
		// memory fairly quickly in efence mode, but at least it won't
		// have mysterious crashes due to confused memory reuse.
		// It should be possible to switch back to SysFree if we also
		// implement and then call some kind of MHeap_DeleteSpan.
		if _base.Debug.Efence > 0 {
			s.Limit = 0 // prevent mlookup from finding this span
			sysFault(unsafe.Pointer(uintptr(s.Start<<_base.PageShift)), size)
		} else {
			mHeap_Free(&_base.Mheap_, s, 1)
		}
		c.Local_nlargefree++
		c.Local_largefree += size
		res = true
	}
	if _base.Trace.Enabled {
		traceGCSweepDone()
	}
	return res
}

func dumpFreeList(s *_base.Mspan) {
	_base.Printlock()
	print("runtime: free list of span ", s, ":\n")
	sstart := uintptr(s.Start << _base.PageShift)
	link := s.Freelist
	for i := 0; i < int(s.Npages*_base.PageSize/s.Elemsize); i++ {
		if i != 0 {
			print(" -> ")
		}
		print(_base.Hex(link))
		if link.Ptr() == nil {
			break
		}
		if uintptr(link) < sstart || s.Limit <= uintptr(link) {
			// Bad link. Stop walking before we crash.
			print(" (BAD)")
			break
		}
		link = link.Ptr().Next
	}
	print("\n")
	_base.Printunlock()
}
