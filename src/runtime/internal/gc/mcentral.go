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

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// Return span from an MCache.
func mCentral_UncacheSpan(c *_lock.Mcentral, s *_core.Mspan) {
	_lock.Lock(&c.Lock)

	s.Incache = false

	if s.Ref == 0 {
		_lock.Throw("uncaching full span")
	}

	cap := int32((s.Npages << _core.PageShift) / s.Elemsize)
	n := cap - int32(s.Ref)
	if n > 0 {
		_sched.MSpanList_Remove(s)
		_sched.MSpanList_Insert(&c.Nonempty, s)
	}
	_lock.Unlock(&c.Lock)
}

// Free n objects from a span s back into the central free list c.
// Called during sweep.
// Returns true if the span was returned to heap.  Sets sweepgen to
// the latest generation.
// If preserve=true, don't return the span to heap nor relink in MCentral lists;
// caller takes care of it.
func mCentral_FreeSpan(c *_lock.Mcentral, s *_core.Mspan, n int32, start _core.Gclinkptr, end _core.Gclinkptr, preserve bool) bool {
	if s.Incache {
		_lock.Throw("freespan into cached span")
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
			_lock.Throw("can't preserve unlinked span")
		}
		_lock.Atomicstore(&s.Sweepgen, _lock.Mheap_.Sweepgen)
		return false
	}

	_lock.Lock(&c.Lock)

	// Move to nonempty if necessary.
	if wasempty {
		_sched.MSpanList_Remove(s)
		_sched.MSpanList_Insert(&c.Nonempty, s)
	}

	// delay updating sweepgen until here.  This is the signal that
	// the span may be used in an MCache, so it must come after the
	// linked list operations above (actually, just after the
	// lock of c above.)
	_lock.Atomicstore(&s.Sweepgen, _lock.Mheap_.Sweepgen)

	if s.Ref != 0 {
		_lock.Unlock(&c.Lock)
		return false
	}

	// s is completely freed, return it to the heap.
	_sched.MSpanList_Remove(s)
	s.Needzero = 1
	s.Freelist = 0
	_lock.Unlock(&c.Lock)
	unmarkspan(uintptr(s.Start)<<_core.PageShift, s.Npages<<_core.PageShift)
	mHeap_Free(&_lock.Mheap_, s, 0)
	return true
}
