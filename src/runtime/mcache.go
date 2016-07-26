// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

// Per-thread (in Go, per-P) cache for small objects.
// No locking needed because it is per-thread (per-P).
//
// mcaches are allocated from non-GC'd memory, so any heap pointers
// must be specially handled.
type mcache struct {
	// The following members are accessed on every malloc,
	// so they are grouped here for better caching.
	next_sample int32   // trigger heap sample after allocating this many bytes
	local_scan  uintptr // bytes of scannable heap allocated

	// Allocator cache for tiny objects w/o pointers.
	// See "Tiny allocator" comment in malloc.go.

	// tiny points to the beginning of the current tiny block, or
	// nil if there is no current tiny block.
	//
	// tiny is a heap pointer. Since mcache is in non-GC'd memory,
	// we handle it by clearing it in releaseAll during mark
	// termination.
	tiny             uintptr
	tinyoffset       uintptr
	local_tinyallocs uintptr // number of tiny allocs not counted in other stats
	rocgoid          int64   // The goid associated with the last roc checkpoint

	// The rest is not accessed on every malloc.
	alloc           [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass
	largeAllocSpans *mspan
	stackcache      [_NumStackOrders]stackfreelist

	// Local allocator stats, flushed during GC.
	local_nlookup    uintptr                  // number of pointer lookups
	local_largefree  uintptr                  // bytes freed for large objects (>maxsmallsize)
	local_nlargefree uintptr                  // number of frees for large objects (>maxsmallsize)
	local_nsmallfree [_NumSizeClasses]uintptr // number of frees for small objects (<=maxsmallsize)
}

// A gclink is a node in a linked list of blocks, like mlink,
// but it is opaque to the garbage collector.
// The GC does not trace the pointers during collection,
// and the compiler does not emit write barriers for assignments
// of gclinkptr values. Code should store references to gclinks
// as gclinkptr, not as *gclink.
type gclink struct {
	next gclinkptr
}

// A gclinkptr is a pointer to a gclink, but it is opaque
// to the garbage collector.
type gclinkptr uintptr

// ptr returns the *gclink form of p.
// The result should be used for accessing fields, not stored
// in other data structures.
func (p gclinkptr) ptr() *gclink {
	return (*gclink)(unsafe.Pointer(p))
}

type stackfreelist struct {
	list gclinkptr // linked list of free stacks
	size uintptr   // total size of stacks in list
}

// dummy MSpan that contains no free objects.
var emptymspan mspan

func allocmcache() *mcache {
	lock(&mheap_.lock)
	c := (*mcache)(mheap_.cachealloc.alloc())
	unlock(&mheap_.lock)
	memclr(unsafe.Pointer(c), unsafe.Sizeof(*c))
	for i := range c.alloc {
		c.alloc[i] = &emptymspan
	}
	c.next_sample = nextSample()
	return c
}

func freemcache(c *mcache) {
	systemstack(func() {
		c.releaseAll()
		stackcache_clear(c)

		// NOTE(rsc,rlh): If gcworkbuffree comes back, we need to coordinate
		// with the stealing of gcworkbufs during garbage collection to avoid
		// a race where the workbuf is double-freed.
		// gcworkbuffree(c.gcworkbuf)

		lock(&mheap_.lock)
		purgecachedstats(c)
		mheap_.cachealloc.free(unsafe.Pointer(c))
		unlock(&mheap_.lock)
	})
}

// Gets a span that has a free object in it and assigns it
// to be the cached span for the given sizeclass. Returns this span.
func (c *mcache) refill(spc spanClass) *mspan {
	_g_ := getg()

	_g_.m.locks++
	// Return the current cached span to the central lists.
	s := c.alloc[spc]

	if uintptr(s.allocCount) != s.nelems {
		throw("refill of span with free space remaining")
	}

	if s != &emptymspan {
		s.incache = false
	}

	// Get a new cached span from the central lists.
	s = mheap_.central[spc].mcentral.cacheSpan()
	if s == nil {
		throw("out of memory")
	}

	if uintptr(s.allocCount) == s.nelems {
		throw("span has no free space")
	}

	c.alloc[spc] = s
	_g_.m.locks--
	return s
}

func (c *mcache) releaseAll() {
	for i := range c.alloc {
		s := c.alloc[i]
		if s != &emptymspan {
			mheap_.central[i].mcentral.uncacheSpan(s)
			c.alloc[i] = &emptymspan
		}
	}
	// Clear tinyalloc pool.
	c.tiny = 0
	c.tinyoffset = 0
}

// rollbackAllocCount recalculates the number of objects allocated in s
// and reflects that in s.allocCount.
func (s *mspan) rollbackAllocCount() {
	s.allocCount = s.allocated()
}

// rollbackAllocCountDebug is used when debug.gcroc >= 2 and contains prints
// out informative messages.
func (s *mspan) rollbackAllocCountDebug(oldfreeindex uintptr, ssweepgen uint32, mheapsweepgen uint32, brand int) {
	if s.startindex != s.freeindex {
		println("runtime: rollbackAllocCount sees span with freeindex != startindex so recycle logic is suspect.")
		throw("startindex not same as freeindex")
	}
	oldsweepgen := atomic.Load(&s.sweepgen)
	traceOn := debug.gcroc >= 3
	if traceOn {
		s.trace("rollbackAllocCount")
		println("s.startindex=", s.startindex, "s.freeindex=", s.freeindex,
			"s.allocCount=", s.allocCount, "s.nelems=", s.nelems, "s.elemsize=", s.elemsize)
	}
	allocated := s.allocated()
	if debug.gcroc >= 2 {
		count := s.startindex // everything before the start index is considered allocated
		for i := s.startindex; i < s.nelems; i++ {
			if !s.isFree(i) {
				count++ // everything between the start index and the last index (s.nelems-1) that is 1 is considered allocated
			}
		}
		if count != s.allocated() {
			freshCount := s.startindex
			for i := s.startindex; i < s.nelems; i++ {
				if !s.isFree(i) {
					freshCount++ // everything between the start index and the last index (s.nelems-1) that is 1 is considered allocated
				}
			}
			if freshCount != count {
				throw("freshCount != count")
			}
			println("runtime: freshCount == count")
			recount := s.allocatedDebug(count)
			println("count=", count, "s.allocated()=", s.allocated(), "s.startindex, s.freeindex=", s.startindex, s.freeindex,
				"s.nelems=", s.nelems, "recount=", recount)
			throw("count != s.allocated")
		}

		if s.startindex != s.freeindex && atomic.Load(&gcphase) == _GCoff {
			if oldsweepgen != atomic.Load(&s.sweepgen) || oldsweepgen != mheap_.sweepgen {
				println("oldsweepgen(", oldsweepgen, ") != atomic.Load(&s.sweepgen)", atomic.Load(&s.sweepgen))
			}
			println("mcache.go:397 busted s.startindex=", s.startindex, "s.freeindex=", s.freeindex, "oldfreeindex=", oldfreeindex,
				"\n     s.allocCount=", s.allocCount, "s.nelems=", s.nelems, "s.elemsize=", s.elemsize, "oldsweepgen=", oldsweepgen,
				"\n     s.sweepgen=", s.sweepgen, "mheap_.sweepgen=", mheap_.sweepgen, "gcphase=", atomic.Load(&gcphase),
				"\n ssweepgen=", ssweepgen, "mheapsweepgen=", mheapsweepgen, "brand=", brand)
			throw(" rollback but s.startindex != s.freeindex")
		}
	}
	s.allocCount = allocated
	if !s.checkAllocCount(s.freeindex) {
		throw("bad checkAllocCount")
	}

	if traceOn {
		println("free objects s.allocCount=", s.allocCount, "s.nelems=", s.nelems)
		s.trace("rollbackAllocCount exits")
	}

}
