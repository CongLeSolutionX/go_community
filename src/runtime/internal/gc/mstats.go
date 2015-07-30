// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory statistics

package gc

import (
	_base "runtime/internal/base"
)

//go:nowritebarrier
func Cachestats() {
	for i := 0; ; i++ {
		p := _base.Allp[i]
		if p == nil {
			break
		}
		c := p.Mcache
		if c == nil {
			continue
		}
		Purgecachedstats(c)
	}
}

//go:nowritebarrier
func Flushallmcaches() {
	for i := 0; ; i++ {
		p := _base.Allp[i]
		if p == nil {
			break
		}
		c := p.Mcache
		if c == nil {
			continue
		}
		mCache_ReleaseAll(c)
		stackcache_clear(c)
	}
}

//go:nosplit
func Purgecachedstats(c *_base.Mcache) {
	// Protected by either heap or GC lock.
	h := &_base.Mheap_
	_base.Memstats.Heap_live += uint64(c.Local_cachealloc)
	c.Local_cachealloc = 0
	if _base.Trace.Enabled {
		TraceHeapAlloc()
	}
	_base.Memstats.Heap_scan += uint64(c.Local_scan)
	c.Local_scan = 0
	_base.Memstats.Tinyallocs += uint64(c.Local_tinyallocs)
	c.Local_tinyallocs = 0
	_base.Memstats.Nlookup += uint64(c.Local_nlookup)
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
