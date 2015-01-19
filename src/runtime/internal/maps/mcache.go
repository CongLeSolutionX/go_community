// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Per-P malloc cache for small objects.
//
// See malloc.h for an overview.

package maps

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

// Gets a span that has a free object in it and assigns it
// to be the cached span for the given sizeclass.  Returns this span.
func mCache_Refill(c *_core.Mcache, sizeclass int32) *_core.Mspan {
	_g_ := _core.Getg()

	_g_.M.Locks++
	// Return the current cached span to the central lists.
	s := c.Alloc[sizeclass]
	if s.Freelist.Ptr() != nil {
		_lock.Throw("refill on a nonempty span")
	}
	if s != &_lock.Emptymspan {
		s.Incache = false
	}

	// Get a new cached span from the central lists.
	s = mCentral_CacheSpan(&_lock.Mheap_.Central[sizeclass].Mcentral)
	if s == nil {
		_lock.Throw("out of memory")
	}
	if s.Freelist.Ptr() == nil {
		println(s.Ref, (s.Npages<<_core.PageShift)/s.Elemsize)
		_lock.Throw("empty span")
	}
	c.Alloc[sizeclass] = s
	_g_.M.Locks--
	return s
}
