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

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

var H_allspans []*_core.Mspan // TODO: make this h.allspans once mheap can be defined in Go

// Look up the span at the given address.
// Address is guaranteed to be in map
// and is guaranteed to be start or end of span.
func mHeap_Lookup(h *_lock.Mheap, v unsafe.Pointer) *_core.Mspan {
	p := uintptr(v)
	p -= uintptr(unsafe.Pointer(h.Arena_start))
	return _sched.H_spans[p>>_core.PageShift]
}

// Free the span back into the heap.
func mHeap_Free(h *_lock.Mheap, s *_core.Mspan, acct int32) {
	_lock.Systemstack(func() {
		mp := _core.Getg().M
		_lock.Lock(&h.Lock)
		_lock.Memstats.Heap_alloc += uint64(mp.Mcache.Local_cachealloc)
		mp.Mcache.Local_cachealloc = 0
		_lock.Memstats.Tinyallocs += uint64(mp.Mcache.Local_tinyallocs)
		mp.Mcache.Local_tinyallocs = 0
		if acct != 0 {
			_lock.Memstats.Heap_alloc -= uint64(s.Npages << _core.PageShift)
			_lock.Memstats.Heap_objects--
		}
		_sched.MHeap_FreeSpanLocked(h, s, true, true)
		_lock.Unlock(&h.Lock)
	})
}

func mHeap_FreeStack(h *_lock.Mheap, s *_core.Mspan) {
	_g_ := _core.Getg()
	if _g_ != _g_.M.G0 {
		_lock.Throw("mheap_freestack not on g0 stack")
	}
	s.Needzero = 1
	_lock.Lock(&h.Lock)
	_lock.Memstats.Stacks_inuse -= uint64(s.Npages << _core.PageShift)
	_sched.MHeap_FreeSpanLocked(h, s, true, true)
	_lock.Unlock(&h.Lock)
}

// Do whatever cleanup needs to be done to deallocate s.  It has
// already been unlinked from the MSpan specials list.
// Returns true if we should keep working on deallocating p.
func freespecial(s *_core.Special, p unsafe.Pointer, size uintptr, freed bool) bool {
	switch s.Kind {
	case KindSpecialFinalizer:
		sf := (*Specialfinalizer)(unsafe.Pointer(s))
		queuefinalizer(p, sf.Fn, sf.Nret, sf.Fint, sf.Ot)
		_lock.Lock(&_lock.Mheap_.Speciallock)
		_sched.FixAlloc_Free(&_lock.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(sf))
		_lock.Unlock(&_lock.Mheap_.Speciallock)
		return false // don't free p until finalizer is done
	case KindSpecialProfile:
		sp := (*Specialprofile)(unsafe.Pointer(s))
		mProf_Free(sp.B, size, freed)
		_lock.Lock(&_lock.Mheap_.Speciallock)
		_sched.FixAlloc_Free(&_lock.Mheap_.Specialprofilealloc, (unsafe.Pointer)(sp))
		_lock.Unlock(&_lock.Mheap_.Speciallock)
		return true
	default:
		_lock.Throw("bad special kind")
		panic("not reached")
	}
}
