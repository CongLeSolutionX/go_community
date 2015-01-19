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

package maps

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// Sweeps spans in list until reclaims at least npages into heap.
// Returns the actual number of pages reclaimed.
func mHeap_ReclaimList(h *_lock.Mheap, list *_core.Mspan, npages uintptr) uintptr {
	n := uintptr(0)
	sg := _lock.Mheap_.Sweepgen
retry:
	for s := list.Next; s != list; s = s.Next {
		if s.Sweepgen == sg-2 && _sched.Cas(&s.Sweepgen, sg-2, sg-1) {
			_sched.MSpanList_Remove(s)
			// swept spans are at the end of the list
			mSpanList_InsertBack(list, s)
			_lock.Unlock(&h.Lock)
			if _gc.MSpan_Sweep(s, false) {
				// TODO(rsc,dvyukov): This is probably wrong.
				// It is undercounting the number of pages reclaimed.
				// See golang.org/issue/9048.
				// Note that if we want to add the true count of s's pages,
				// we must record that before calling mSpan_Sweep,
				// because if mSpan_Sweep returns true the span has
				// been
				n++
			}
			_lock.Lock(&h.Lock)
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
func mHeap_Reclaim(h *_lock.Mheap, npage uintptr) {
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
	_lock.Unlock(&h.Lock)
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
	_lock.Lock(&h.Lock)
}

// Allocate a new span of npage pages from the heap for GC'd memory
// and record its size class in the HeapMap and HeapMapCache.
func mHeap_Alloc_m(h *_lock.Mheap, npage uintptr, sizeclass int32, large bool) *_core.Mspan {
	_g_ := _core.Getg()
	if _g_ != _g_.M.G0 {
		_lock.Gothrow("_mheap_alloc not on g0 stack")
	}
	_lock.Lock(&h.Lock)

	// To prevent excessive heap growth, before allocating n pages
	// we need to sweep and reclaim at least n pages.
	if h.Sweepdone == 0 {
		mHeap_Reclaim(h, npage)
	}

	// transfer stats from cache to global
	_lock.Memstats.Heap_alloc += uint64(_g_.M.Mcache.Local_cachealloc)
	_g_.M.Mcache.Local_cachealloc = 0
	_lock.Memstats.Tinyallocs += uint64(_g_.M.Mcache.Local_tinyallocs)
	_g_.M.Mcache.Local_tinyallocs = 0

	s := _sched.MHeap_AllocSpanLocked(h, npage)
	if s != nil {
		// Record span info, because gc needs to be
		// able to map interior pointer to containing span.
		_lock.Atomicstore(&s.Sweepgen, h.Sweepgen)
		s.State = _sched.MSpanInUse
		s.Freelist = 0
		s.Ref = 0
		s.Sizeclass = uint8(sizeclass)
		if sizeclass == 0 {
			s.Elemsize = s.Npages << _core.PageShift
		} else {
			s.Elemsize = uintptr(_gc.Class_to_size[sizeclass])
		}

		// update stats, sweep lists
		if large {
			_lock.Memstats.Heap_objects++
			_lock.Memstats.Heap_alloc += uint64(npage << _core.PageShift)
			// Swept spans are at the end of lists.
			if s.Npages < uintptr(len(h.Free)) {
				mSpanList_InsertBack(&h.Busy[s.Npages], s)
			} else {
				mSpanList_InsertBack(&h.Busylarge, s)
			}
		}
	}
	_lock.Unlock(&h.Lock)
	return s
}

func mHeap_Alloc(h *_lock.Mheap, npage uintptr, sizeclass int32, large bool, needzero bool) *_core.Mspan {
	// Don't do any operations that lock the heap on the G stack.
	// It might trigger stack growth, and the stack growth code needs
	// to be able to allocate heap.
	var s *_core.Mspan
	_lock.Systemstack(func() {
		s = mHeap_Alloc_m(h, npage, sizeclass, large)
	})

	if s != nil {
		if needzero && s.Needzero != 0 {
			_core.Memclr(unsafe.Pointer(s.Start<<_core.PageShift), s.Npages<<_core.PageShift)
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
func MHeap_LookupMaybe(h *_lock.Mheap, v unsafe.Pointer) *_core.Mspan {
	if uintptr(v) < uintptr(unsafe.Pointer(h.Arena_start)) || uintptr(v) >= uintptr(unsafe.Pointer(h.Arena_used)) {
		return nil
	}
	p := uintptr(v) >> _core.PageShift
	q := p
	q -= uintptr(unsafe.Pointer(h.Arena_start)) >> _core.PageShift
	s := _sched.H_spans[q]
	if s == nil || p < uintptr(s.Start) || uintptr(v) >= uintptr(unsafe.Pointer(s.Limit)) || s.State != _sched.MSpanInUse {
		return nil
	}
	return s
}

func mSpanList_InsertBack(list *_core.Mspan, span *_core.Mspan) {
	if span.Next != nil || span.Prev != nil {
		println("failed MSpanList_InsertBack", span, span.Next, span.Prev)
		_lock.Gothrow("MSpanList_InsertBack")
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
func Addspecial(p unsafe.Pointer, s *_core.Special) bool {
	span := MHeap_LookupMaybe(&_lock.Mheap_, p)
	if span == nil {
		_lock.Gothrow("addspecial on invalid pointer")
	}

	// Ensure that the span is swept.
	// GC accesses specials list w/o locks. And it's just much safer.
	mp := _sched.Acquirem()
	_gc.MSpan_EnsureSwept(span)

	offset := uintptr(p) - uintptr(span.Start<<_core.PageShift)
	kind := s.Kind

	_lock.Lock(&span.Speciallock)

	// Find splice point, check for existing record.
	t := &span.Specials
	for {
		x := *t
		if x == nil {
			break
		}
		if offset == uintptr(x.Offset) && kind == x.Kind {
			_lock.Unlock(&span.Speciallock)
			_sched.Releasem(mp)
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
	_lock.Unlock(&span.Speciallock)
	_sched.Releasem(mp)

	return true
}

// Set the heap profile bucket associated with addr to b.
func setprofilebucket(p unsafe.Pointer, b *_sem.Bucket) {
	_lock.Lock(&_lock.Mheap_.Speciallock)
	s := (*_gc.Specialprofile)(_lock.FixAlloc_Alloc(&_lock.Mheap_.Specialprofilealloc))
	_lock.Unlock(&_lock.Mheap_.Speciallock)
	s.Special.Kind = _gc.KindSpecialProfile
	s.B = b
	if !Addspecial(p, &s.Special) {
		_lock.Gothrow("setprofilebucket: profile already set")
	}
}
