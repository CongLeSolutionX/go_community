// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

import (
	_base "runtime/internal/base"
)

// Gets a span that has a free object in it and assigns it
// to be the cached span for the given sizeclass.  Returns this span.
func mCache_Refill(c *_base.Mcache, sizeclass int32) *_base.Mspan {
	_g_ := _base.Getg()

	_g_.M.Locks++
	// Return the current cached span to the central lists.
	s := c.Alloc[sizeclass]
	if s.Freelist.Ptr() != nil {
		_base.Throw("refill on a nonempty span")
	}
	if s != &_base.Emptymspan {
		s.Incache = false
	}

	// Get a new cached span from the central lists.
	s = mCentral_CacheSpan(&_base.Mheap_.Central[sizeclass].Mcentral)
	if s == nil {
		_base.Throw("out of memory")
	}
	if s.Freelist.Ptr() == nil {
		println(s.Ref, (s.Npages<<_base.PageShift)/s.Elemsize)
		_base.Throw("empty span")
	}
	c.Alloc[sizeclass] = s
	_g_.M.Locks--
	return s
}
