// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See malloc.h for overview.
//
// TODO(rsc): double-check stats.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

//go:nosplit
func Purgecachedstats(c *_core.Mcache) {
	// Protected by either heap or GC lock.
	h := &_lock.Mheap_
	_lock.Memstats.Heap_alloc += uint64(c.Local_cachealloc)
	c.Local_cachealloc = 0
	_lock.Memstats.Tinyallocs += uint64(c.Local_tinyallocs)
	c.Local_tinyallocs = 0
	_lock.Memstats.Nlookup += uint64(c.Local_nlookup)
	c.Local_nlookup = 0
	h.Largefree += uint64(c.Local_largefree)
	c.Local_largefree = 0
	h.Nlargefree += uint64(c.Local_nlargefree)
	c.Local_nlargefree = 0
	for i := 0; i < len(c.Local_nsmallfree); i++ {
		h.Nsmallfree[i] += uint64(c.Local_nsmallfree[i])
		c.Local_nsmallfree[i] = 0
	}
}
