// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Central free lists.
//
// See malloc.h for an overview.
//
// The MCentral doesn't actually contain the list of free objects; the MSpan does.
// Each MCentral is two lists of MSpans: those with free objects (c->nonempty)
// and those that are completely allocated (c->empty).

package maps

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Allocate a span to use in an MCache.
func mCentral_CacheSpan(c *_lock.Mcentral) *_core.Mspan {
	_lock.Lock(&c.Lock)
	sg := _lock.Mheap_.Sweepgen
retry:
	var s *_core.Mspan
	for s = c.Nonempty.Next; s != &c.Nonempty; s = s.Next {
		if s.Sweepgen == sg-2 && _sched.Cas(&s.Sweepgen, sg-2, sg-1) {
			_sched.MSpanList_Remove(s)
			mSpanList_InsertBack(&c.Empty, s)
			_lock.Unlock(&c.Lock)
			_gc.MSpan_Sweep(s, true)
			goto havespan
		}
		if s.Sweepgen == sg-1 {
			// the span is being swept by background sweeper, skip
			continue
		}
		// we have a nonempty span that does not require sweeping, allocate from it
		_sched.MSpanList_Remove(s)
		mSpanList_InsertBack(&c.Empty, s)
		_lock.Unlock(&c.Lock)
		goto havespan
	}

	for s = c.Empty.Next; s != &c.Empty; s = s.Next {
		if s.Sweepgen == sg-2 && _sched.Cas(&s.Sweepgen, sg-2, sg-1) {
			// we have an empty span that requires sweeping,
			// sweep it and see if we can free some space in it
			_sched.MSpanList_Remove(s)
			// swept spans are at the end of the list
			mSpanList_InsertBack(&c.Empty, s)
			_lock.Unlock(&c.Lock)
			_gc.MSpan_Sweep(s, true)
			if s.Freelist.Ptr() != nil {
				goto havespan
			}
			_lock.Lock(&c.Lock)
			// the span is still empty after sweep
			// it is already in the empty list, so just retry
			goto retry
		}
		if s.Sweepgen == sg-1 {
			// the span is being swept by background sweeper, skip
			continue
		}
		// already swept empty span,
		// all subsequent ones must also be either swept or in process of sweeping
		break
	}
	_lock.Unlock(&c.Lock)

	// Replenish central list if empty.
	s = mCentral_Grow(c)
	if s == nil {
		return nil
	}
	_lock.Lock(&c.Lock)
	mSpanList_InsertBack(&c.Empty, s)
	_lock.Unlock(&c.Lock)

	// At this point s is a non-empty span, queued at the end of the empty list,
	// c is unlocked.
havespan:
	cap := int32((s.Npages << _core.PageShift) / s.Elemsize)
	n := cap - int32(s.Ref)
	if n == 0 {
		_lock.Gothrow("empty span")
	}
	if s.Freelist.Ptr() == nil {
		_lock.Gothrow("freelist empty")
	}
	s.Incache = true
	return s
}

// Fetch a new span from the heap and carve into objects for the free list.
func mCentral_Grow(c *_lock.Mcentral) *_core.Mspan {
	npages := uintptr(_gc.Class_to_allocnpages[c.Sizeclass])
	size := uintptr(_gc.Class_to_size[c.Sizeclass])
	n := (npages << _core.PageShift) / size

	s := mHeap_Alloc(&_lock.Mheap_, npages, c.Sizeclass, false, true)
	if s == nil {
		return nil
	}

	p := uintptr(s.Start << _core.PageShift)
	s.Limit = p + size*n
	head := _core.Gclinkptr(p)
	tail := _core.Gclinkptr(p)
	// i==0 iteration already done
	for i := uintptr(1); i < n; i++ {
		p += size
		tail.Ptr().Next = _core.Gclinkptr(p)
		tail = _core.Gclinkptr(p)
	}
	if s.Freelist.Ptr() != nil {
		_lock.Gothrow("freelist not empty")
	}
	tail.Ptr().Next = 0
	s.Freelist = head
	markspan(unsafe.Pointer(uintptr(s.Start)<<_core.PageShift), size, n, size*n < s.Npages<<_core.PageShift)
	return s
}
