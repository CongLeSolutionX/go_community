// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.go for overview.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	"unsafe"
)

func recordspan(vh unsafe.Pointer, p unsafe.Pointer) {
	h := (*_base.Mheap)(vh)
	s := (*_base.Mspan)(p)
	if len(_gc.H_allspans) >= cap(_gc.H_allspans) {
		n := 64 * 1024 / _base.PtrSize
		if n < cap(_gc.H_allspans)*3/2 {
			n = cap(_gc.H_allspans) * 3 / 2
		}
		var new []*_base.Mspan
		sp := (*_base.Slice)(unsafe.Pointer(&new))
		sp.Array = _base.SysAlloc(uintptr(n)*_base.PtrSize, &_base.Memstats.Other_sys)
		if sp.Array == nil {
			_base.Throw("runtime: cannot allocate memory")
		}
		sp.Len = len(_gc.H_allspans)
		sp.Cap = n
		if len(_gc.H_allspans) > 0 {
			copy(new, _gc.H_allspans)
			// Don't free the old array if it's referenced by sweep.
			// See the comment in mgc.go.
			if h.Allspans != _base.Mheap_.Gcspans {
				_base.SysFree(unsafe.Pointer(h.Allspans), uintptr(cap(_gc.H_allspans))*_base.PtrSize, &_base.Memstats.Other_sys)
			}
		}
		_gc.H_allspans = new
		h.Allspans = (**_base.Mspan)(unsafe.Pointer(sp.Array))
	}
	_gc.H_allspans = append(_gc.H_allspans, s)
	h.Nspan = uint32(len(_gc.H_allspans))
}

// TODO: spanOf and spanOfUnchecked are open-coded in a lot of places.
// Use the functions instead.

// spanOf returns the span of p. If p does not point into the heap or
// no span contains p, spanOf returns nil.
func spanOf(p uintptr) *_base.Mspan {
	if p == 0 || p < _base.Mheap_.Arena_start || p >= _base.Mheap_.Arena_used {
		return nil
	}
	return _base.SpanOfUnchecked(p)
}

func mlookup(v uintptr, base *uintptr, size *uintptr, sp **_base.Mspan) int32 {
	_g_ := _base.Getg()

	_g_.M.Mcache.Local_nlookup++
	if _base.PtrSize == 4 && _g_.M.Mcache.Local_nlookup >= 1<<30 {
		// purge cache stats to prevent overflow
		_base.Lock(&_base.Mheap_.Lock)
		_gc.Purgecachedstats(_g_.M.Mcache)
		_base.Unlock(&_base.Mheap_.Lock)
	}

	s := _iface.MHeap_LookupMaybe(&_base.Mheap_, unsafe.Pointer(v))
	if sp != nil {
		*sp = s
	}
	if s == nil {
		if base != nil {
			*base = 0
		}
		if size != nil {
			*size = 0
		}
		return 0
	}

	p := uintptr(s.Start) << _base.PageShift
	if s.Sizeclass == 0 {
		// Large object.
		if base != nil {
			*base = p
		}
		if size != nil {
			*size = s.Npages << _base.PageShift
		}
		return 1
	}

	n := s.Elemsize
	if base != nil {
		i := (uintptr(v) - uintptr(p)) / n
		*base = p + i*n
	}
	if size != nil {
		*size = n
	}

	return 1
}

// Initialize the heap.
func mHeap_Init(h *_base.Mheap, spans_size uintptr) {
	fixAlloc_Init(&h.Spanalloc, unsafe.Sizeof(_base.Mspan{}), recordspan, unsafe.Pointer(h), &_base.Memstats.Mspan_sys)
	fixAlloc_Init(&h.Cachealloc, unsafe.Sizeof(_base.Mcache{}), nil, nil, &_base.Memstats.Mcache_sys)
	fixAlloc_Init(&h.Specialfinalizeralloc, unsafe.Sizeof(_gc.Specialfinalizer{}), nil, nil, &_base.Memstats.Other_sys)
	fixAlloc_Init(&h.Specialprofilealloc, unsafe.Sizeof(_gc.Specialprofile{}), nil, nil, &_base.Memstats.Other_sys)

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

	sp := (*_base.Slice)(unsafe.Pointer(&_base.H_spans))
	sp.Array = unsafe.Pointer(h.Spans)
	sp.Len = int(spans_size / _base.PtrSize)
	sp.Cap = int(spans_size / _base.PtrSize)
}

func scavengelist(list *_base.Mspan, now, limit uint64) uintptr {
	if _base.PhysPageSize > _base.PageSize {
		// golang.org/issue/9993
		// If the physical page size of the machine is larger than
		// our logical heap page size the kernel may round up the
		// amount to be freed to its page size and corrupt the heap
		// pages surrounding the unused block.
		return 0
	}

	if _base.MSpanList_IsEmpty(list) {
		return 0
	}

	var sumreleased uintptr
	for s := list.Next; s != list; s = s.Next {
		if (now-uint64(s.Unusedsince)) > limit && s.Npreleased != s.Npages {
			released := (s.Npages - s.Npreleased) << _base.PageShift
			_base.Memstats.Heap_released += uint64(released)
			sumreleased += released
			s.Npreleased = s.Npages
			sysUnused((unsafe.Pointer)(s.Start<<_base.PageShift), s.Npages<<_base.PageShift)
		}
	}
	return sumreleased
}

func mHeap_Scavenge(k int32, now, limit uint64) {
	h := &_base.Mheap_
	_base.Lock(&h.Lock)
	var sumreleased uintptr
	for i := 0; i < len(h.Free); i++ {
		sumreleased += scavengelist(&h.Free[i], now, limit)
	}
	sumreleased += scavengelist(&h.Freelarge, now, limit)
	_base.Unlock(&h.Lock)

	if _base.Debug.Gctrace > 0 {
		if sumreleased > 0 {
			print("scvg", k, ": ", sumreleased>>20, " MB released\n")
		}
		// TODO(dvyukov): these stats are incorrect as we don't subtract stack usage from heap.
		// But we can't call ReadMemStats on g0 holding locks.
		print("scvg", k, ": inuse: ", _base.Memstats.Heap_inuse>>20, ", idle: ", _base.Memstats.Heap_idle>>20, ", sys: ", _base.Memstats.Heap_sys>>20, ", released: ", _base.Memstats.Heap_released>>20, ", consumed: ", (_base.Memstats.Heap_sys-_base.Memstats.Heap_released)>>20, " (MB)\n")
	}
}

//go:linkname runtime_debug_freeOSMemory runtime/debug.freeOSMemory
func runtime_debug_freeOSMemory() {
	_iface.StartGC(_gc.GcForceBlockMode)
	_base.Systemstack(func() { mHeap_Scavenge(-1, ^uint64(0), 0) })
}

// Initialize an empty doubly-linked list.
func mSpanList_Init(list *_base.Mspan) {
	list.State = _base.MSpanListHead
	list.Next = list
	list.Prev = list
}

// Removes the Special record of the given kind for the object p.
// Returns the record if the record existed, nil otherwise.
// The caller must FixAlloc_Free the result.
func removespecial(p unsafe.Pointer, kind uint8) *_base.Special {
	span := _iface.MHeap_LookupMaybe(&_base.Mheap_, p)
	if span == nil {
		_base.Throw("removespecial on invalid pointer")
	}

	// Ensure that the span is swept.
	// GC accesses specials list w/o locks. And it's just much safer.
	mp := _base.Acquirem()
	_gc.MSpan_EnsureSwept(span)

	offset := uintptr(p) - uintptr(span.Start<<_base.PageShift)

	_base.Lock(&span.Speciallock)
	t := &span.Specials
	for {
		s := *t
		if s == nil {
			break
		}
		// This function is used for finalizers only, so we don't check for
		// "interior" specials (p must be exactly equal to s->offset).
		if offset == uintptr(s.Offset) && kind == s.Kind {
			*t = s.Next
			_base.Unlock(&span.Speciallock)
			_base.Releasem(mp)
			return s
		}
		t = &s.Next
	}
	_base.Unlock(&span.Speciallock)
	_base.Releasem(mp)
	return nil
}

// Adds a finalizer to the object p.  Returns true if it succeeded.
func addfinalizer(p unsafe.Pointer, f *_base.Funcval, nret uintptr, fint *_base.Type, ot *_gc.Ptrtype) bool {
	_base.Lock(&_base.Mheap_.Speciallock)
	s := (*_gc.Specialfinalizer)(_base.FixAlloc_Alloc(&_base.Mheap_.Specialfinalizeralloc))
	_base.Unlock(&_base.Mheap_.Speciallock)
	s.Special.Kind = _gc.KindSpecialFinalizer
	s.Fn = f
	s.Nret = nret
	s.Fint = fint
	s.Ot = ot
	if _iface.Addspecial(p, &s.Special) {
		return true
	}

	// There was an old finalizer
	_base.Lock(&_base.Mheap_.Speciallock)
	_base.FixAlloc_Free(&_base.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(s))
	_base.Unlock(&_base.Mheap_.Speciallock)
	return false
}

// Removes the finalizer (if any) from the object p.
func removefinalizer(p unsafe.Pointer) {
	s := (*_gc.Specialfinalizer)(unsafe.Pointer(removespecial(p, _gc.KindSpecialFinalizer)))
	if s == nil {
		return // there wasn't a finalizer to remove
	}
	_base.Lock(&_base.Mheap_.Speciallock)
	_base.FixAlloc_Free(&_base.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(s))
	_base.Unlock(&_base.Mheap_.Speciallock)
}
