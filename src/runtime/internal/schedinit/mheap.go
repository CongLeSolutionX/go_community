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

package schedinit

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func recordspan(vh unsafe.Pointer, p unsafe.Pointer) {
	h := (*_lock.Mheap)(vh)
	s := (*_core.Mspan)(p)
	if len(_gc.H_allspans) >= cap(_gc.H_allspans) {
		n := 64 * 1024 / _core.PtrSize
		if n < cap(_gc.H_allspans)*3/2 {
			n = cap(_gc.H_allspans) * 3 / 2
		}
		var new []*_core.Mspan
		sp := (*_core.Slice)(unsafe.Pointer(&new))
		sp.Array = (*byte)(_lock.SysAlloc(uintptr(n)*_core.PtrSize, &_lock.Memstats.Other_sys))
		if sp.Array == nil {
			_lock.Gothrow("runtime: cannot allocate memory")
		}
		sp.Len = uint(len(_gc.H_allspans))
		sp.Cap = uint(n)
		if len(_gc.H_allspans) > 0 {
			copy(new, _gc.H_allspans)
			// Don't free the old array if it's referenced by sweep.
			// See the comment in mgc0.c.
			if h.Allspans != _lock.Mheap_.Gcspans {
				_sched.SysFree(unsafe.Pointer(h.Allspans), uintptr(cap(_gc.H_allspans))*_core.PtrSize, &_lock.Memstats.Other_sys)
			}
		}
		_gc.H_allspans = new
		h.Allspans = (**_core.Mspan)(unsafe.Pointer(sp.Array))
	}
	_gc.H_allspans = append(_gc.H_allspans, s)
	h.Nspan = uint32(len(_gc.H_allspans))
}

// Initialize the heap.
func mHeap_Init(h *_lock.Mheap, spans_size uintptr) {
	fixAlloc_Init(&h.Spanalloc, unsafe.Sizeof(_core.Mspan{}), recordspan, unsafe.Pointer(h), &_lock.Memstats.Mspan_sys)
	fixAlloc_Init(&h.Cachealloc, unsafe.Sizeof(_core.Mcache{}), nil, nil, &_lock.Memstats.Mcache_sys)
	fixAlloc_Init(&h.Specialfinalizeralloc, unsafe.Sizeof(_gc.Specialfinalizer{}), nil, nil, &_lock.Memstats.Other_sys)
	fixAlloc_Init(&h.Specialprofilealloc, unsafe.Sizeof(_gc.Specialprofile{}), nil, nil, &_lock.Memstats.Other_sys)

	// h->mapcache needs no init
	for i := range h.Free {
		mSpanList_Init(&h.Free[i])
		mSpanList_Init(&h.Busy[i])
	}

	mSpanList_Init(&h.Freelarge)
	mSpanList_Init(&h.Busylarge)
	for i := range h.Central {
		mCentral_Init(&h.Central[i].Mcentral, int32(i))
	}

	sp := (*_core.Slice)(unsafe.Pointer(&_sched.H_spans))
	sp.Array = (*byte)(unsafe.Pointer(h.Spans))
	sp.Len = uint(spans_size / _core.PtrSize)
	sp.Cap = uint(spans_size / _core.PtrSize)
}

// Initialize an empty doubly-linked list.
func mSpanList_Init(list *_core.Mspan) {
	list.State = _sched.MSpanListHead
	list.Next = list
	list.Prev = list
}
