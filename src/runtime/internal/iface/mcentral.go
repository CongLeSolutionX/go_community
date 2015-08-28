// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Central free lists.
//
// See malloc.go for an overview.
//
// The MCentral doesn't actually contain the list of free objects; the MSpan does.
// Each MCentral is two lists of MSpans: those with free objects (c->nonempty)
// and those that are completely allocated (c->empty).

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
)

// Allocate a span to use in an MCache.
func mCentral_CacheSpan(c *_base.Mcentral) *_base.Mspan {
	// Deduct credit for this span allocation and sweep if necessary.
	deductSweepCredit(uintptr(Class_to_size[c.Sizeclass]), 0)

	_base.Lock(&c.Lock)
	sg := _base.Mheap_.Sweepgen
retry:
	var s *_base.Mspan
	for s = c.Nonempty.Next; s != &c.Nonempty; s = s.Next {
		if s.Sweepgen == sg-2 && _base.Cas(&s.Sweepgen, sg-2, sg-1) {
			_base.MSpanList_Remove(s)
			mSpanList_InsertBack(&c.Empty, s)
			_base.Unlock(&c.Lock)
			_gc.MSpan_Sweep(s, true)
			goto havespan
		}
		if s.Sweepgen == sg-1 {
			// the span is being swept by background sweeper, skip
			continue
		}
		// we have a nonempty span that does not require sweeping, allocate from it
		_base.MSpanList_Remove(s)
		mSpanList_InsertBack(&c.Empty, s)
		_base.Unlock(&c.Lock)
		goto havespan
	}

	for s = c.Empty.Next; s != &c.Empty; s = s.Next {
		if s.Sweepgen == sg-2 && _base.Cas(&s.Sweepgen, sg-2, sg-1) {
			// we have an empty span that requires sweeping,
			// sweep it and see if we can free some space in it
			_base.MSpanList_Remove(s)
			// swept spans are at the end of the list
			mSpanList_InsertBack(&c.Empty, s)
			_base.Unlock(&c.Lock)
			_gc.MSpan_Sweep(s, true)
			if s.Freelist.Ptr() != nil {
				goto havespan
			}
			_base.Lock(&c.Lock)
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
	_base.Unlock(&c.Lock)

	// Replenish central list if empty.
	s = mCentral_Grow(c)
	if s == nil {
		return nil
	}
	_base.Lock(&c.Lock)
	mSpanList_InsertBack(&c.Empty, s)
	_base.Unlock(&c.Lock)

	// At this point s is a non-empty span, queued at the end of the empty list,
	// c is unlocked.
havespan:
	cap := int32((s.Npages << _base.PageShift) / s.Elemsize)
	n := cap - int32(s.Ref)
	if n == 0 {
		_base.Throw("empty span")
	}
	if s.Freelist.Ptr() == nil {
		_base.Throw("freelist empty")
	}
	s.Incache = true
	return s
}

// Fetch a new span from the heap and carve into objects for the free list.
func mCentral_Grow(c *_base.Mcentral) *_base.Mspan {
	npages := uintptr(Class_to_allocnpages[c.Sizeclass])
	size := uintptr(Class_to_size[c.Sizeclass])
	n := (npages << _base.PageShift) / size

	s := mHeap_Alloc(&_base.Mheap_, npages, c.Sizeclass, false, true)
	if s == nil {
		return nil
	}

	p := uintptr(s.Start << _base.PageShift)
	s.Limit = p + size*n
	head := _base.Gclinkptr(p)
	tail := _base.Gclinkptr(p)
	// i==0 iteration already done
	for i := uintptr(1); i < n; i++ {
		p += size
		tail.Ptr().Next = _base.Gclinkptr(p)
		tail = _base.Gclinkptr(p)
	}
	if s.Freelist.Ptr() != nil {
		_base.Throw("freelist not empty")
	}
	tail.Ptr().Next = 0
	s.Freelist = head
	_gc.HeapBitsForSpan(s.Base()).InitSpan(s.Layout())
	return s
}
