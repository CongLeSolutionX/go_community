// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.h for overview.
//
// When a MSpan is in the heap free list, state == MSpanFree
// and heapmap(s->start) == span, heapmap(s->start+s->npages-1) == span.
//
// When a MSpan is allocated, state == MSpanInUse or MSpanStack
// and heapmap(i) == span for all s->start <= i < s->start+s->npages.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

var H_spans []*_core.Mspan // TODO: make this h.spans once mheap can be defined in Go

func mHeap_MapSpans(h *_lock.Mheap) {
	// Map spans array, PageSize at a time.
	n := uintptr(unsafe.Pointer(h.Arena_used))
	n -= uintptr(unsafe.Pointer(h.Arena_start))
	n = n / _core.PageSize * _core.PtrSize
	n = Round(n, _lock.PhysPageSize)
	if h.Spans_mapped >= n {
		return
	}
	sysMap(_core.Add(unsafe.Pointer(h.Spans), h.Spans_mapped), n-h.Spans_mapped, h.Arena_reserved, &_lock.Memstats.Other_sys)
	h.Spans_mapped = n
}

func mHeap_AllocStack(h *_lock.Mheap, npage uintptr) *_core.Mspan {
	_g_ := _core.Getg()
	if _g_ != _g_.M.G0 {
		_lock.Gothrow("mheap_allocstack not on g0 stack")
	}
	_lock.Lock(&h.Lock)
	s := MHeap_AllocSpanLocked(h, npage)
	if s != nil {
		s.State = MSpanStack
		s.Freelist = 0
		s.Ref = 0
		_lock.Memstats.Stacks_inuse += uint64(s.Npages << _core.PageShift)
	}
	_lock.Unlock(&h.Lock)
	return s
}

// Allocates a span of the given size.  h must be locked.
// The returned span has been removed from the
// free list, but its state is still MSpanFree.
func MHeap_AllocSpanLocked(h *_lock.Mheap, npage uintptr) *_core.Mspan {
	var s *_core.Mspan

	// Try in fixed-size lists up to max.
	for i := int(npage); i < len(h.Free); i++ {
		if !MSpanList_IsEmpty(&h.Free[i]) {
			s = h.Free[i].Next
			goto HaveSpan
		}
	}

	// Best fit in list of large spans.
	s = mHeap_AllocLarge(h, npage)
	if s == nil {
		if !mHeap_Grow(h, npage) {
			return nil
		}
		s = mHeap_AllocLarge(h, npage)
		if s == nil {
			return nil
		}
	}

HaveSpan:
	// Mark span in use.
	if s.State != MSpanFree {
		_lock.Gothrow("MHeap_AllocLocked - MSpan not free")
	}
	if s.Npages < npage {
		_lock.Gothrow("MHeap_AllocLocked - bad npages")
	}
	MSpanList_Remove(s)
	if s.Next != nil || s.Prev != nil {
		_lock.Gothrow("still in list")
	}
	if s.Npreleased > 0 {
		sysUsed((unsafe.Pointer)(s.Start<<_core.PageShift), s.Npages<<_core.PageShift)
		_lock.Memstats.Heap_released -= uint64(s.Npreleased << _core.PageShift)
		s.Npreleased = 0
	}

	if s.Npages > npage {
		// Trim extra and put it back in the heap.
		t := (*_core.Mspan)(_lock.FixAlloc_Alloc(&h.Spanalloc))
		mSpan_Init(t, s.Start+_core.PageID(npage), s.Npages-npage)
		s.Npages = npage
		p := uintptr(t.Start)
		p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> _core.PageShift)
		if p > 0 {
			H_spans[p-1] = s
		}
		H_spans[p] = t
		H_spans[p+t.Npages-1] = t
		t.Needzero = s.Needzero
		s.State = MSpanStack // prevent coalescing with s
		t.State = MSpanStack
		MHeap_FreeSpanLocked(h, t, false, false)
		t.Unusedsince = s.Unusedsince // preserve age (TODO: wrong: t is possibly merged and/or deallocated at this point)
		s.State = MSpanFree
	}
	s.Unusedsince = 0

	p := uintptr(s.Start)
	p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> _core.PageShift)
	for n := uintptr(0); n < npage; n++ {
		H_spans[p+n] = s
	}

	_lock.Memstats.Heap_inuse += uint64(npage << _core.PageShift)
	_lock.Memstats.Heap_idle -= uint64(npage << _core.PageShift)

	//println("spanalloc", hex(s.start<<_PageShift))
	if s.Next != nil || s.Prev != nil {
		_lock.Gothrow("still in list")
	}
	return s
}

// Allocate a span of exactly npage pages from the list of large spans.
func mHeap_AllocLarge(h *_lock.Mheap, npage uintptr) *_core.Mspan {
	return bestFit(&h.Freelarge, npage, nil)
}

// Search list for smallest span with >= npage pages.
// If there are multiple smallest spans, take the one
// with the earliest starting address.
func bestFit(list *_core.Mspan, npage uintptr, best *_core.Mspan) *_core.Mspan {
	for s := list.Next; s != list; s = s.Next {
		if s.Npages < npage {
			continue
		}
		if best == nil || s.Npages < best.Npages || (s.Npages == best.Npages && s.Start < best.Start) {
			best = s
		}
	}
	return best
}

// Try to add at least npage pages of memory to the heap,
// returning whether it worked.
func mHeap_Grow(h *_lock.Mheap, npage uintptr) bool {
	// Ask for a big chunk, to reduce the number of mappings
	// the operating system needs to track; also amortizes
	// the overhead of an operating system mapping.
	// Allocate a multiple of 64kB.
	npage = Round(npage, (64<<10)/_core.PageSize)
	ask := npage << _core.PageShift
	if ask < _core.HeapAllocChunk {
		ask = _core.HeapAllocChunk
	}

	v := mHeap_SysAlloc(h, ask)
	if v == nil {
		if ask > npage<<_core.PageShift {
			ask = npage << _core.PageShift
			v = mHeap_SysAlloc(h, ask)
		}
		if v == nil {
			print("runtime: out of memory: cannot allocate ", ask, "-byte block (", _lock.Memstats.Heap_sys, " in use)\n")
			return false
		}
	}

	// Create a fake "in use" span and free it, so that the
	// right coalescing happens.
	s := (*_core.Mspan)(_lock.FixAlloc_Alloc(&h.Spanalloc))
	mSpan_Init(s, _core.PageID(uintptr(v)>>_core.PageShift), ask>>_core.PageShift)
	p := uintptr(s.Start)
	p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> _core.PageShift)
	H_spans[p] = s
	H_spans[p+s.Npages-1] = s
	_lock.Atomicstore(&s.Sweepgen, h.Sweepgen)
	s.State = MSpanInUse
	MHeap_FreeSpanLocked(h, s, false, true)
	return true
}

func MHeap_FreeSpanLocked(h *_lock.Mheap, s *_core.Mspan, acctinuse, acctidle bool) {
	switch s.State {
	case MSpanStack:
		if s.Ref != 0 {
			_lock.Gothrow("MHeap_FreeSpanLocked - invalid stack free")
		}
	case MSpanInUse:
		if s.Ref != 0 || s.Sweepgen != h.Sweepgen {
			print("MHeap_FreeSpanLocked - span ", s, " ptr ", _core.Hex(s.Start<<_core.PageShift), " ref ", s.Ref, " sweepgen ", s.Sweepgen, "/", h.Sweepgen, "\n")
			_lock.Gothrow("MHeap_FreeSpanLocked - invalid free")
		}
	default:
		_lock.Gothrow("MHeap_FreeSpanLocked - invalid span state")
	}

	if acctinuse {
		_lock.Memstats.Heap_inuse -= uint64(s.Npages << _core.PageShift)
	}
	if acctidle {
		_lock.Memstats.Heap_idle += uint64(s.Npages << _core.PageShift)
	}
	s.State = MSpanFree
	MSpanList_Remove(s)

	// Stamp newly unused spans. The scavenger will use that
	// info to potentially give back some pages to the OS.
	s.Unusedsince = _lock.Nanotime()
	s.Npreleased = 0

	// Coalesce with earlier, later spans.
	p := uintptr(s.Start)
	p -= uintptr(unsafe.Pointer(h.Arena_start)) >> _core.PageShift
	if p > 0 {
		t := H_spans[p-1]
		if t != nil && t.State != MSpanInUse && t.State != MSpanStack {
			s.Start = t.Start
			s.Npages += t.Npages
			s.Npreleased = t.Npreleased // absorb released pages
			s.Needzero |= t.Needzero
			p -= t.Npages
			H_spans[p] = s
			MSpanList_Remove(t)
			t.State = MSpanDead
			FixAlloc_Free(&h.Spanalloc, (unsafe.Pointer)(t))
		}
	}
	if (p+s.Npages)*_core.PtrSize < h.Spans_mapped {
		t := H_spans[p+s.Npages]
		if t != nil && t.State != MSpanInUse && t.State != MSpanStack {
			s.Npages += t.Npages
			s.Npreleased += t.Npreleased
			s.Needzero |= t.Needzero
			H_spans[p+s.Npages-1] = s
			MSpanList_Remove(t)
			t.State = MSpanDead
			FixAlloc_Free(&h.Spanalloc, (unsafe.Pointer)(t))
		}
	}

	// Insert s into appropriate list.
	if s.Npages < uintptr(len(h.Free)) {
		MSpanList_Insert(&h.Free[s.Npages], s)
	} else {
		MSpanList_Insert(&h.Freelarge, s)
	}
}

// Initialize a new span with the given start and npages.
func mSpan_Init(span *_core.Mspan, start _core.PageID, npages uintptr) {
	span.Next = nil
	span.Prev = nil
	span.Start = start
	span.Npages = npages
	span.Freelist = 0
	span.Ref = 0
	span.Sizeclass = 0
	span.Incache = false
	span.Elemsize = 0
	span.State = MSpanDead
	span.Unusedsince = 0
	span.Npreleased = 0
	span.Speciallock.Key = 0
	span.Specials = nil
	span.Needzero = 0
}

func MSpanList_Remove(span *_core.Mspan) {
	if span.Prev == nil && span.Next == nil {
		return
	}
	span.Prev.Next = span.Next
	span.Next.Prev = span.Prev
	span.Prev = nil
	span.Next = nil
}

func MSpanList_IsEmpty(list *_core.Mspan) bool {
	return list.Next == list
}

func MSpanList_Insert(list *_core.Mspan, span *_core.Mspan) {
	if span.Next != nil || span.Prev != nil {
		println("failed MSpanList_Insert", span, span.Next, span.Prev)
		_lock.Gothrow("MSpanList_Insert")
	}
	span.Next = list.Next
	span.Prev = list
	span.Next.Prev = span
	span.Prev.Next = span
}
