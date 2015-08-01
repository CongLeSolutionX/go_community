// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.go for overview.

package base

import (
	"unsafe"
)

// Main malloc heap.
// The heap itself is the "free[]" and "large" arrays,
// but all the other global data is here too.
type Mheap struct {
	Lock      Mutex
	Free      [MaxMHeapList]Mspan // free lists of given length
	Freelarge Mspan               // free lists length >= _MaxMHeapList
	Busy      [MaxMHeapList]Mspan // busy lists of large objects of given length
	Busylarge Mspan               // busy lists of large objects length >= _MaxMHeapList
	Allspans  **Mspan             // all spans out there
	Gcspans   **Mspan             // copy of allspans referenced by gc marker or sweeper
	Nspan     uint32
	Sweepgen  uint32 // sweep generation, see comment in mspan
	Sweepdone uint32 // all spans are swept
	// span lookup
	Spans        **Mspan
	spans_mapped uintptr

	// Proportional sweep
	PagesSwept        uint64  // pages swept this cycle; updated atomically
	SweepPagesPerByte float64 // proportional sweep ratio; written with lock, read without

	// Malloc stats.
	Largefree  uint64                 // bytes freed for large objects (>maxsmallsize)
	Nlargefree uint64                 // number of frees for large objects (>maxsmallsize)
	Nsmallfree [NumSizeClasses]uint64 // number of frees for small objects (<=maxsmallsize)

	// range of addresses we might see in the heap
	Bitmap         uintptr
	bitmap_mapped  uintptr
	Arena_start    uintptr
	Arena_used     uintptr // always mHeap_Map{Bits,Spans} before updating
	Arena_end      uintptr
	Arena_reserved bool

	// central free lists for small size classes.
	// the padding makes sure that the MCentrals are
	// spaced CacheLineSize bytes apart, so that each MCentral.lock
	// gets its own cache line.
	Central [NumSizeClasses]struct {
		Mcentral Mcentral
		pad      [CacheLineSize]byte
	}

	Spanalloc             Fixalloc // allocator for span*
	Cachealloc            Fixalloc // allocator for mcache*
	Specialfinalizeralloc Fixalloc // allocator for specialfinalizer*
	Specialprofilealloc   Fixalloc // allocator for specialprofile*
	Speciallock           Mutex    // lock for special record allocators.
}

var Mheap_ Mheap

// An MSpan is a run of pages.
//
// When a MSpan is in the heap free list, state == MSpanFree
// and heapmap(s->start) == span, heapmap(s->start+s->npages-1) == span.
//
// When a MSpan is allocated, state == MSpanInUse or MSpanStack
// and heapmap(i) == span for all s->start <= i < s->start+s->npages.

// Every MSpan is in one doubly-linked list,
// either one of the MHeap's free lists or one of the
// MCentral's span lists.  We use empty MSpan structures as list heads.

// An MSpan representing actual memory has state _MSpanInUse,
// _MSpanStack, or _MSpanFree. Transitions between these states are
// constrained as follows:
//
// * A span may transition from free to in-use or stack during any GC
//   phase.
//
// * During sweeping (gcphase == _GCoff), a span may transition from
//   in-use to free (as a result of sweeping) or stack to free (as a
//   result of stacks being freed).
//
// * During GC (gcphase != _GCoff), a span *must not* transition from
//   stack or in-use to free. Because concurrent GC may read a pointer
//   and then look up its span, the span state must be monotonic.
const (
	MSpanInUse = iota // allocated for garbage collected heap
	MSpanStack        // allocated for use by stack allocator
	MSpanFree
	MSpanListHead
	MSpanDead
)

type Mspan struct {
	Next     *Mspan    // in a span linked list
	Prev     *Mspan    // in a span linked list
	Start    pageID    // starting page number
	Npages   uintptr   // number of pages in span
	Freelist Gclinkptr // list of free objects
	// sweep generation:
	// if sweepgen == h->sweepgen - 2, the span needs sweeping
	// if sweepgen == h->sweepgen - 1, the span is currently being swept
	// if sweepgen == h->sweepgen, the span is swept and ready to use
	// h->sweepgen is incremented by 2 after every GC

	Sweepgen    uint32
	DivMul      uint32   // for divide by elemsize - divMagic.mul
	Ref         uint16   // capacity - number of objects in freelist
	Sizeclass   uint8    // size class
	Incache     bool     // being used by an mcache
	State       uint8    // mspaninuse etc
	Needzero    uint8    // needs to be zeroed before allocation
	DivShift    uint8    // for divide by elemsize - divMagic.shift
	DivShift2   uint8    // for divide by elemsize - divMagic.shift2
	Elemsize    uintptr  // computed from sizeclass or from npages
	Unusedsince int64    // first time spotted by gc in mspanfree state
	Npreleased  uintptr  // number of pages released to the os
	Limit       uintptr  // end of data in span
	Speciallock Mutex    // guards specials list
	Specials    *Special // linked list of special records sorted by offset.
	BaseMask    uintptr  // if non-0, elemsize is a power of 2, & this will get object allocation base
}

func (s *Mspan) Base() uintptr {
	return uintptr(s.Start << PageShift)
}

func (s *Mspan) Layout() (size, n, total uintptr) {
	total = s.Npages << PageShift
	size = s.Elemsize
	if size > 0 {
		n = total / size
	}
	return
}

// h_spans is a lookup table to map virtual address page IDs to *mspan.
// For allocated spans, their pages map to the span itself.
// For free spans, only the lowest and highest pages map to the span itself.  Internal
// pages map to an arbitrary span.
// For pages that have never been allocated, h_spans entries are nil.
var H_spans []*Mspan // TODO: make this h.spans once mheap can be defined in Go

// inheap reports whether b is a pointer into a (potentially dead) heap object.
// It returns false for pointers into stack spans.
// Non-preemptible because it is used by write barriers.
//go:nowritebarrier
//go:nosplit
func Inheap(b uintptr) bool {
	if b == 0 || b < Mheap_.Arena_start || b >= Mheap_.Arena_used {
		return false
	}
	// Not a beginning of a block, consult span table to find the block beginning.
	k := b >> PageShift
	x := k
	x -= Mheap_.Arena_start >> PageShift
	s := H_spans[x]
	if s == nil || pageID(k) < s.Start || b >= s.Limit || s.State != XMSpanInUse {
		return false
	}
	return true
}

// spanOfUnchecked is equivalent to spanOf, but the caller must ensure
// that p points into the heap (that is, mheap_.arena_start <= p <
// mheap_.arena_used).
func SpanOfUnchecked(p uintptr) *Mspan {
	return H_spans[(p-Mheap_.Arena_start)>>PageShift]
}

// mHeap_MapSpans makes sure that the spans are mapped
// up to the new value of arena_used.
//
// It must be called with the expected new value of arena_used,
// *before* h.arena_used has been updated.
// Waiting to update arena_used until after the memory has been mapped
// avoids faults when other threads try access the bitmap immediately
// after observing the change to arena_used.
func mHeap_MapSpans(h *Mheap, arena_used uintptr) {
	// Map spans array, PageSize at a time.
	n := arena_used
	n -= h.Arena_start
	n = n / PageSize * PtrSize
	n = Round(n, PhysPageSize)
	if h.spans_mapped >= n {
		return
	}
	sysMap(Add(unsafe.Pointer(h.Spans), h.spans_mapped), n-h.spans_mapped, h.Arena_reserved, &Memstats.Other_sys)
	h.spans_mapped = n
}

func mHeap_AllocStack(h *Mheap, npage uintptr) *Mspan {
	_g_ := Getg()
	if _g_ != _g_.M.G0 {
		Throw("mheap_allocstack not on g0 stack")
	}
	Lock(&h.Lock)
	s := MHeap_AllocSpanLocked(h, npage)
	if s != nil {
		s.State = MSpanStack
		s.Freelist = 0
		s.Ref = 0
		Memstats.Stacks_inuse += uint64(s.Npages << PageShift)
	}

	// This unlock acts as a release barrier. See mHeap_Alloc_m.
	Unlock(&h.Lock)
	return s
}

// Allocates a span of the given size.  h must be locked.
// The returned span has been removed from the
// free list, but its state is still MSpanFree.
func MHeap_AllocSpanLocked(h *Mheap, npage uintptr) *Mspan {
	var s *Mspan

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
		Throw("MHeap_AllocLocked - MSpan not free")
	}
	if s.Npages < npage {
		Throw("MHeap_AllocLocked - bad npages")
	}
	MSpanList_Remove(s)
	if s.Next != nil || s.Prev != nil {
		Throw("still in list")
	}
	if s.Npreleased > 0 {
		sysUsed((unsafe.Pointer)(s.Start<<PageShift), s.Npages<<PageShift)
		Memstats.Heap_released -= uint64(s.Npreleased << PageShift)
		s.Npreleased = 0
	}

	if s.Npages > npage {
		// Trim extra and put it back in the heap.
		t := (*Mspan)(FixAlloc_Alloc(&h.Spanalloc))
		mSpan_Init(t, s.Start+pageID(npage), s.Npages-npage)
		s.Npages = npage
		p := uintptr(t.Start)
		p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> PageShift)
		if p > 0 {
			H_spans[p-1] = s
		}
		H_spans[p] = t
		H_spans[p+t.Npages-1] = t
		t.Needzero = s.Needzero
		s.State = MSpanStack // prevent coalescing with s
		t.State = MSpanStack
		MHeap_FreeSpanLocked(h, t, false, false, s.Unusedsince)
		s.State = MSpanFree
	}
	s.Unusedsince = 0

	p := uintptr(s.Start)
	p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> PageShift)
	for n := uintptr(0); n < npage; n++ {
		H_spans[p+n] = s
	}

	Memstats.Heap_inuse += uint64(npage << PageShift)
	Memstats.Heap_idle -= uint64(npage << PageShift)

	//println("spanalloc", hex(s.start<<_PageShift))
	if s.Next != nil || s.Prev != nil {
		Throw("still in list")
	}
	return s
}

// Allocate a span of exactly npage pages from the list of large spans.
func mHeap_AllocLarge(h *Mheap, npage uintptr) *Mspan {
	return bestFit(&h.Freelarge, npage, nil)
}

// Search list for smallest span with >= npage pages.
// If there are multiple smallest spans, take the one
// with the earliest starting address.
func bestFit(list *Mspan, npage uintptr, best *Mspan) *Mspan {
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
func mHeap_Grow(h *Mheap, npage uintptr) bool {
	// Ask for a big chunk, to reduce the number of mappings
	// the operating system needs to track; also amortizes
	// the overhead of an operating system mapping.
	// Allocate a multiple of 64kB.
	npage = Round(npage, (64<<10)/PageSize)
	ask := npage << PageShift
	if ask < HeapAllocChunk {
		ask = HeapAllocChunk
	}

	v := mHeap_SysAlloc(h, ask)
	if v == nil {
		if ask > npage<<PageShift {
			ask = npage << PageShift
			v = mHeap_SysAlloc(h, ask)
		}
		if v == nil {
			print("runtime: out of memory: cannot allocate ", ask, "-byte block (", Memstats.Heap_sys, " in use)\n")
			return false
		}
	}

	// Create a fake "in use" span and free it, so that the
	// right coalescing happens.
	s := (*Mspan)(FixAlloc_Alloc(&h.Spanalloc))
	mSpan_Init(s, pageID(uintptr(v)>>PageShift), ask>>PageShift)
	p := uintptr(s.Start)
	p -= (uintptr(unsafe.Pointer(h.Arena_start)) >> PageShift)
	for i := p; i < p+s.Npages; i++ {
		H_spans[i] = s
	}
	Atomicstore(&s.Sweepgen, h.Sweepgen)
	s.State = MSpanInUse
	MHeap_FreeSpanLocked(h, s, false, true, 0)
	return true
}

func MHeap_FreeSpanLocked(h *Mheap, s *Mspan, acctinuse, acctidle bool, unusedsince int64) {
	switch s.State {
	case MSpanStack:
		if s.Ref != 0 {
			Throw("MHeap_FreeSpanLocked - invalid stack free")
		}
	case MSpanInUse:
		if s.Ref != 0 || s.Sweepgen != h.Sweepgen {
			print("MHeap_FreeSpanLocked - span ", s, " ptr ", Hex(s.Start<<PageShift), " ref ", s.Ref, " sweepgen ", s.Sweepgen, "/", h.Sweepgen, "\n")
			Throw("MHeap_FreeSpanLocked - invalid free")
		}
	default:
		Throw("MHeap_FreeSpanLocked - invalid span state")
	}

	if acctinuse {
		Memstats.Heap_inuse -= uint64(s.Npages << PageShift)
	}
	if acctidle {
		Memstats.Heap_idle += uint64(s.Npages << PageShift)
	}
	s.State = MSpanFree
	MSpanList_Remove(s)

	// Stamp newly unused spans. The scavenger will use that
	// info to potentially give back some pages to the OS.
	s.Unusedsince = unusedsince
	if unusedsince == 0 {
		s.Unusedsince = Nanotime()
	}
	s.Npreleased = 0

	// Coalesce with earlier, later spans.
	p := uintptr(s.Start)
	p -= uintptr(unsafe.Pointer(h.Arena_start)) >> PageShift
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
	if (p+s.Npages)*PtrSize < h.spans_mapped {
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
func mSpan_Init(span *Mspan, start pageID, npages uintptr) {
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
	span.Speciallock.key = 0
	span.Specials = nil
	span.Needzero = 0
}

func MSpanList_Remove(span *Mspan) {
	if span.Prev == nil && span.Next == nil {
		return
	}
	span.Prev.Next = span.Next
	span.Next.Prev = span.Prev
	span.Prev = nil
	span.Next = nil
}

func MSpanList_IsEmpty(list *Mspan) bool {
	return list.Next == list
}

func MSpanList_Insert(list *Mspan, span *Mspan) {
	if span.Next != nil || span.Prev != nil {
		println("failed MSpanList_Insert", span, span.Next, span.Prev)
		Throw("MSpanList_Insert")
	}
	span.Next = list.Next
	span.Prev = list
	span.Next.Prev = span
	span.Prev.Next = span
}

type Special struct {
	Next   *Special // linked list in span
	Offset uint16   // span offset of object
	Kind   byte     // kind of special
}
