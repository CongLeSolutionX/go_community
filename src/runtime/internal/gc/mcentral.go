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

package gc

import (
	_base "runtime/internal/base"
)

// Return span from an MCache.
func mCentral_UncacheSpan(c *_base.Mcentral, s *_base.Mspan) {
	_base.Lock(&c.Lock)

	s.Incache = false

	if s.Ref == 0 {
		_base.Throw("uncaching full span")
	}

	cap := int32((s.Npages << _base.PageShift) / s.Elemsize)
	n := cap - int32(s.Ref)
	if n > 0 {
		_base.MSpanList_Remove(s)
		_base.MSpanList_Insert(&c.Nonempty, s)
	}
	_base.Unlock(&c.Lock)
}

// Free n objects from a span s back into the central free list c.
// Called during sweep.
// Returns true if the span was returned to heap.  Sets sweepgen to
// the latest generation.
// If preserve=true, don't return the span to heap nor relink in MCentral lists;
// caller takes care of it.
func mCentral_FreeSpan(c *_base.Mcentral, s *_base.Mspan, n int32, start _base.Gclinkptr, end _base.Gclinkptr, preserve bool) bool {
	if s.Incache {
		_base.Throw("freespan into cached span")
	}

	// Add the objects back to s's free list.
	wasempty := s.Freelist.Ptr() == nil
	end.Ptr().Next = s.Freelist
	s.Freelist = start
	s.Ref -= uint16(n)

	if preserve {
		// preserve is set only when called from MCentral_CacheSpan above,
		// the span must be in the empty list.
		if s.Next == nil {
			_base.Throw("can't preserve unlinked span")
		}
		_base.Atomicstore(&s.Sweepgen, _base.Mheap_.Sweepgen)
		return false
	}

	_base.Lock(&c.Lock)

	// Move to nonempty if necessary.
	if wasempty {
		_base.MSpanList_Remove(s)
		_base.MSpanList_Insert(&c.Nonempty, s)
	}

	// delay updating sweepgen until here.  This is the signal that
	// the span may be used in an MCache, so it must come after the
	// linked list operations above (actually, just after the
	// lock of c above.)
	_base.Atomicstore(&s.Sweepgen, _base.Mheap_.Sweepgen)

	if s.Ref != 0 {
		_base.Unlock(&c.Lock)
		return false
	}

	// s is completely freed, return it to the heap.
	_base.MSpanList_Remove(s)
	s.Needzero = 1
	s.Freelist = 0
	_base.Unlock(&c.Lock)
	HeapBitsForSpan(s.Base()).InitSpan(s.Layout())
	mHeap_Free(&_base.Mheap_, s, 0)
	return true
}
