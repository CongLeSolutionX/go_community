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

package finalize

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Removes the Special record of the given kind for the object p.
// Returns the record if the record existed, nil otherwise.
// The caller must FixAlloc_Free the result.
func removespecial(p unsafe.Pointer, kind uint8) *_core.Special {
	span := _maps.MHeap_LookupMaybe(&_lock.Mheap_, p)
	if span == nil {
		_lock.Throw("removespecial on invalid pointer")
	}

	// Ensure that the span is swept.
	// GC accesses specials list w/o locks. And it's just much safer.
	mp := _sched.Acquirem()
	_gc.MSpan_EnsureSwept(span)

	offset := uintptr(p) - uintptr(span.Start<<_core.PageShift)

	_lock.Lock(&span.Speciallock)
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
			_lock.Unlock(&span.Speciallock)
			_sched.Releasem(mp)
			return s
		}
		t = &s.Next
	}
	_lock.Unlock(&span.Speciallock)
	_sched.Releasem(mp)
	return nil
}

// Adds a finalizer to the object p.  Returns true if it succeeded.
func Addfinalizer(p unsafe.Pointer, f *_core.Funcval, nret uintptr, fint *_core.Type, ot *_gc.Ptrtype) bool {
	_lock.Lock(&_lock.Mheap_.Speciallock)
	s := (*_gc.Specialfinalizer)(_lock.FixAlloc_Alloc(&_lock.Mheap_.Specialfinalizeralloc))
	_lock.Unlock(&_lock.Mheap_.Speciallock)
	s.Special.Kind = _gc.KindSpecialFinalizer
	s.Fn = f
	s.Nret = nret
	s.Fint = fint
	s.Ot = ot
	if _maps.Addspecial(p, &s.Special) {
		return true
	}

	// There was an old finalizer
	_lock.Lock(&_lock.Mheap_.Speciallock)
	_sched.FixAlloc_Free(&_lock.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(s))
	_lock.Unlock(&_lock.Mheap_.Speciallock)
	return false
}

// Removes the finalizer (if any) from the object p.
func Removefinalizer(p unsafe.Pointer) {
	s := (*_gc.Specialfinalizer)(unsafe.Pointer(removespecial(p, _gc.KindSpecialFinalizer)))
	if s == nil {
		return // there wasn't a finalizer to remove
	}
	_lock.Lock(&_lock.Mheap_.Speciallock)
	_sched.FixAlloc_Free(&_lock.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(s))
	_lock.Unlock(&_lock.Mheap_.Speciallock)
}
