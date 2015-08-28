// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.go for overview.

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

// Sweeps spans in list until reclaims at least npages into heap.
// Returns the actual number of pages reclaimed.
func mHeap_ReclaimList(h *_base.Mheap, list *_base.Mspan, npages uintptr) uintptr {
	n := uintptr(0)
	sg := _base.Mheap_.Sweepgen
retry:
	for s := list.Next; s != list; s = s.Next {
		if s.Sweepgen == sg-2 && _base.Cas(&s.Sweepgen, sg-2, sg-1) {
			_base.MSpanList_Remove(s)
			// swept spans are at the end of the list
			mSpanList_InsertBack(list, s)
			_base.Unlock(&h.Lock)
			snpages := s.Npages
			if _gc.MSpan_Sweep(s, false) {
				n += snpages
			}
			_base.Lock(&h.Lock)
			if n >= npages {
				return n
			}
			// the span could have been moved elsewhere
			goto retry
		}
		if s.Sweepgen == sg-1 {
			// the span is being sweept by background sweeper, skip
			continue
		}
		// already swept empty span,
		// all subsequent ones must also be either swept or in process of sweeping
		break
	}
	return n
}

// Sweeps and reclaims at least npage pages into heap.
// Called before allocating npage pages.
func mHeap_Reclaim(h *_base.Mheap, npage uintptr) {
	// First try to sweep busy spans with large objects of size >= npage,
	// this has good chances of reclaiming the necessary space.
	for i := int(npage); i < len(h.Busy); i++ {
		if mHeap_ReclaimList(h, &h.Busy[i], npage) != 0 {
			return // Bingo!
		}
	}

	// Then -- even larger objects.
	if mHeap_ReclaimList(h, &h.Busylarge, npage) != 0 {
		return // Bingo!
	}

	// Now try smaller objects.
	// One such object is not enough, so we need to reclaim several of them.
	reclaimed := uintptr(0)
	for i := 0; i < int(npage) && i < len(h.Busy); i++ {
		reclaimed += mHeap_ReclaimList(h, &h.Busy[i], npage-reclaimed)
		if reclaimed >= npage {
			return
		}
	}

	// Now sweep everything that is not yet swept.
	_base.Unlock(&h.Lock)
	for {
		n := _gc.Sweepone()
		if n == ^uintptr(0) { // all spans are swept
			break
		}
		reclaimed += n
		if reclaimed >= npage {
			break
		}
	}
	_base.Lock(&h.Lock)
}

// Allocate a new span of npage pages from the heap for GC'd memory
// and record its size class in the HeapMap and HeapMapCache.
func mHeap_Alloc_m(h *_base.Mheap, npage uintptr, sizeclass int32, large bool) *_base.Mspan {
	_g_ := _base.Getg()
	if _g_ != _g_.M.G0 {
		_base.Throw("_mheap_alloc not on g0 stack")
	}
	_base.Lock(&h.Lock)

	// To prevent excessive heap growth, before allocating n pages
	// we need to sweep and reclaim at least n pages.
	if h.Sweepdone == 0 {
		// TODO(austin): This tends to sweep a large number of
		// spans in order to find a few completely free spans
		// (for example, in the garbage benchmark, this sweeps
		// ~30x the number of pages its trying to allocate).
		// If GC kept a bit for whether there were any marks
		// in a span, we could release these free spans
		// at the end of GC and eliminate this entirely.
		mHeap_Reclaim(h, npage)
	}

	// transfer stats from cache to global
	_base.Memstats.Heap_live += uint64(_g_.M.Mcache.Local_cachealloc)
	_g_.M.Mcache.Local_cachealloc = 0
	_base.Memstats.Heap_scan += uint64(_g_.M.Mcache.Local_scan)
	_g_.M.Mcache.Local_scan = 0
	_base.Memstats.Tinyallocs += uint64(_g_.M.Mcache.Local_tinyallocs)
	_g_.M.Mcache.Local_tinyallocs = 0

	_base.GcController.Revise()

	s := _base.MHeap_AllocSpanLocked(h, npage)
	if s != nil {
		// Record span info, because gc needs to be
		// able to map interior pointer to containing span.
		_base.Atomicstore(&s.Sweepgen, h.Sweepgen)
		s.State = _base.XMSpanInUse
		s.Freelist = 0
		s.Ref = 0
		s.Sizeclass = uint8(sizeclass)
		if sizeclass == 0 {
			s.Elemsize = s.Npages << _base.PageShift
			s.DivShift = 0
			s.DivMul = 0
			s.DivShift2 = 0
			s.BaseMask = 0
		} else {
			s.Elemsize = uintptr(Class_to_size[sizeclass])
			m := &Class_to_divmagic[sizeclass]
			s.DivShift = m.Shift
			s.DivMul = m.Mul
			s.DivShift2 = m.Shift2
			s.BaseMask = m.BaseMask
		}

		// update stats, sweep lists
		if large {
			_base.Memstats.Heap_objects++
			_base.Memstats.Heap_live += uint64(npage << _base.PageShift)
			// Swept spans are at the end of lists.
			if s.Npages < uintptr(len(h.Free)) {
				mSpanList_InsertBack(&h.Busy[s.Npages], s)
			} else {
				mSpanList_InsertBack(&h.Busylarge, s)
			}
		}
	}
	if _base.Trace.Enabled {
		_gc.TraceHeapAlloc()
	}

	// h_spans is accessed concurrently without synchronization
	// from other threads. Hence, there must be a store/store
	// barrier here to ensure the writes to h_spans above happen
	// before the caller can publish a pointer p to an object
	// allocated from s. As soon as this happens, the garbage
	// collector running on another processor could read p and
	// look up s in h_spans. The unlock acts as the barrier to
	// order these writes. On the read side, the data dependency
	// between p and the index in h_spans orders the reads.
	_base.Unlock(&h.Lock)
	return s
}

func mHeap_Alloc(h *_base.Mheap, npage uintptr, sizeclass int32, large bool, needzero bool) *_base.Mspan {
	// Don't do any operations that lock the heap on the G stack.
	// It might trigger stack growth, and the stack growth code needs
	// to be able to allocate heap.
	var s *_base.Mspan
	_base.Systemstack(func() {
		s = mHeap_Alloc_m(h, npage, sizeclass, large)
	})

	if s != nil {
		if needzero && s.Needzero != 0 {
			_base.Memclr(unsafe.Pointer(s.Start<<_base.PageShift), s.Npages<<_base.PageShift)
		}
		s.Needzero = 0
	}
	return s
}

// Look up the span at the given address.
// Address is *not* guaranteed to be in map
// and may be anywhere in the span.
// Map entries for the middle of a span are only
// valid for allocated spans.  Free spans may have
// other garbage in their middles, so we have to
// check for that.
func MHeap_LookupMaybe(h *_base.Mheap, v unsafe.Pointer) *_base.Mspan {
	if uintptr(v) < uintptr(unsafe.Pointer(h.Arena_start)) || uintptr(v) >= uintptr(unsafe.Pointer(h.Arena_used)) {
		return nil
	}
	p := uintptr(v) >> _base.PageShift
	q := p
	q -= uintptr(unsafe.Pointer(h.Arena_start)) >> _base.PageShift
	s := _base.H_spans[q]
	if s == nil || p < uintptr(s.Start) || uintptr(v) >= uintptr(unsafe.Pointer(s.Limit)) || s.State != _base.XMSpanInUse {
		return nil
	}
	return s
}

func mSpanList_InsertBack(list *_base.Mspan, span *_base.Mspan) {
	if span.Next != nil || span.Prev != nil {
		println("failed MSpanList_InsertBack", span, span.Next, span.Prev)
		_base.Throw("MSpanList_InsertBack")
	}
	span.Next = list
	span.Prev = list.Prev
	span.Next.Prev = span
	span.Prev.Next = span
}

// Adds the special record s to the list of special records for
// the object p.  All fields of s should be filled in except for
// offset & next, which this routine will fill in.
// Returns true if the special was successfully added, false otherwise.
// (The add will fail only if a record with the same p and s->kind
//  already exists.)
func Addspecial(p unsafe.Pointer, s *_base.Special) bool {
	span := MHeap_LookupMaybe(&_base.Mheap_, p)
	if span == nil {
		_base.Throw("addspecial on invalid pointer")
	}

	// Ensure that the span is swept.
	// GC accesses specials list w/o locks. And it's just much safer.
	mp := _base.Acquirem()
	_gc.MSpan_EnsureSwept(span)

	offset := uintptr(p) - uintptr(span.Start<<_base.PageShift)
	kind := s.Kind

	_base.Lock(&span.Speciallock)

	// Find splice point, check for existing record.
	t := &span.Specials
	for {
		x := *t
		if x == nil {
			break
		}
		if offset == uintptr(x.Offset) && kind == x.Kind {
			_base.Unlock(&span.Speciallock)
			_base.Releasem(mp)
			return false // already exists
		}
		if offset < uintptr(x.Offset) || (offset == uintptr(x.Offset) && kind < x.Kind) {
			break
		}
		t = &x.Next
	}

	// Splice in record, fill in offset.
	s.Offset = uint16(offset)
	s.Next = *t
	*t = s
	_base.Unlock(&span.Speciallock)
	_base.Releasem(mp)

	return true
}

// Set the heap profile bucket associated with addr to b.
func setprofilebucket(p unsafe.Pointer, b *_gc.Bucket) {
	_base.Lock(&_base.Mheap_.Speciallock)
	s := (*_gc.Specialprofile)(_base.FixAlloc_Alloc(&_base.Mheap_.Specialprofilealloc))
	_base.Unlock(&_base.Mheap_.Speciallock)
	s.Special.Kind = _gc.KindSpecialProfile
	s.B = b
	if !Addspecial(p, &s.Special) {
		_base.Throw("setprofilebucket: profile already set")
	}
}
