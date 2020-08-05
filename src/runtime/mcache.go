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
//
//go:notinheap
type mcache struct {
	// The following members are accessed on every malloc,
	// so they are grouped here for better caching.
	next_sample uintptr // trigger heap sample after allocating this many bytes
	localScan   uintptr // bytes of scannable heap allocated

	// Allocator cache for tiny objects w/o pointers.
	// See "Tiny allocator" comment in malloc.go.

	// tiny points to the beginning of the current tiny block, or
	// nil if there is no current tiny block.
	//
	// tiny is a heap pointer. Since mcache is in non-GC'd memory,
	// we handle it by clearing it in releaseAll during mark
	// termination.
	tiny            uintptr
	tinyoffset      uintptr
	localTinyAllocs uintptr // number of tiny allocs not counted in other stats

	// The rest is not accessed on every malloc.

	alloc [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass

	stackcache [_NumStackOrders]stackfreelist

	// Allocator stats (source-of-truth).
	// Only the P that owns this mcache may write to these
	// variables, so it's safe to read non-atomically as long
	// as GOMAXPROCS is prevented from changing, as long ass
	// races are acceptable.
	//
	// When read with stats from other mcaches and with the world
	// stopped, the result will accurately reflect the state of the
	// application.
	localLargeAlloc  uintptr                  // bytes allocated for large objects
	localLargeAllocN uintptr                  // number of large object allocations
	localSmallAllocN [_NumSizeClasses]uintptr // number of allocs for small objects
	localLargeFree   uintptr                  // bytes freed for large objects (>maxSmallSize)
	localLargeFreeN  uintptr                  // number of frees for large objects (>maxSmallSize)
	localSmallFreeN  [_NumSizeClasses]uintptr // number of frees for small objects (<=maxSmallSize)

	// Sharded memory stats (source-of-truth).
	localMemStats memStatsShard

	// flushGen indicates the sweepgen during which this mcache
	// was last flushed. If flushGen != mheap_.sweepgen, the spans
	// in this mcache are stale and need to the flushed so they
	// can be swept. This is done in acquirep.
	flushGen uint32
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

// dummy mspan that contains no free objects.
var emptymspan mspan

func allocmcache() *mcache {
	var c *mcache
	systemstack(func() {
		lock(&mheap_.lock)
		c = (*mcache)(mheap_.cachealloc.alloc())
		c.flushGen = mheap_.sweepgen
		unlock(&mheap_.lock)
	})
	for i := range c.alloc {
		c.alloc[i] = &emptymspan
	}
	c.next_sample = nextSample()
	return c
}

// freemcache releases resources associated with this
// mcache and puts the object onto a free list.
//
// In some cases there is no way to simply release
// resources, such as statistics, so donate them to
// a different mcache (the recipient).
func freemcache(c *mcache, recipient *mcache) {
	systemstack(func() {
		c.releaseAll()
		stackcache_clear(c)

		// NOTE(rsc,rlh): If gcworkbuffree comes back, we need to coordinate
		// with the stealing of gcworkbufs during garbage collection to avoid
		// a race where the workbuf is double-freed.
		// gcworkbuffree(c.gcworkbuf)

		lock(&mheap_.lock)
		// Donate anything else that's left.
		c.donate(recipient)
		mheap_.cachealloc.free(unsafe.Pointer(c))
		unlock(&mheap_.lock)
	})
}

// getMCache is a convenience function which tries to obtain an mcache.
//
// Must be running with a P when called (so the caller must be in a
// non-preemptible state) or must be called during bootstrapping.
func getMCache() *mcache {
	// Grab the mcache, since that's where stats live.
	pp := getg().m.p.ptr()
	var c *mcache
	if pp == nil {
		// We will be called without a P while bootstrapping,
		// in which case we use mcache0, which is set in mallocinit.
		// mcache0 is cleared when bootstrapping is complete,
		// by procresize.
		c = mcache0
		if c == nil {
			throw("allocSpan called with no P or nil mcache")
		}
	} else {
		c = pp.mcache
	}
	return c
}

// donate flushes data and resources which have no global
// pool to another mcache.
func (c *mcache) donate(d *mcache) {
	// localScan is handled separately because it's not
	// like these stats -- it's used for GC pacing.
	d.localLargeAlloc += c.localLargeAlloc
	c.localLargeAlloc = 0
	d.localLargeAllocN += c.localLargeAllocN
	c.localLargeAllocN = 0
	for i := range c.localSmallAllocN {
		d.localSmallAllocN[i] += c.localSmallAllocN[i]
		c.localSmallAllocN[i] = 0
	}
	d.localLargeFree += c.localLargeFree
	c.localLargeFree = 0
	d.localLargeFreeN += c.localLargeFreeN
	c.localLargeFreeN = 0
	for i := range c.localSmallFreeN {
		d.localSmallFreeN[i] += c.localSmallFreeN[i]
		c.localSmallFreeN[i] = 0
	}
	d.localTinyAllocs += c.localTinyAllocs
	c.localTinyAllocs = 0

	// Flush c's local mem stats to cmsd
	// and clear it.
	var cmsd memStatsDelta
	c.localMemStats.unsafeRead(&cmsd)
	c.localMemStats.unsafeClear()

	// Write cmsd to d's local mem stats
	// the usual way.
	dmsd := d.localMemStats.acquire()
	dmsd.merge(&cmsd)
	d.localMemStats.release()
}

// refill acquires a new span of span class spc for c. This span will
// have at least one free object. The current span in c must be full.
//
// Must run in a non-preemptible context since otherwise the owner of
// c could change.
func (c *mcache) refill(spc spanClass) {
	// Return the current cached span to the central lists.
	s := c.alloc[spc]

	if uintptr(s.allocCount) != s.nelems {
		throw("refill of span with free space remaining")
	}
	if s != &emptymspan {
		// Mark this span as no longer cached.
		if s.sweepgen != mheap_.sweepgen+3 {
			throw("bad sweepgen in refill")
		}
		if go115NewMCentralImpl {
			mheap_.central[spc].mcentral.uncacheSpan(s)
		} else {
			atomic.Store(&s.sweepgen, mheap_.sweepgen)
		}
	}

	// Get a new cached span from the central lists.
	s = mheap_.central[spc].mcentral.cacheSpan()
	if s == nil {
		throw("out of memory")
	}

	if uintptr(s.allocCount) == s.nelems {
		throw("span has no free space")
	}

	// Indicate that this span is cached and prevent asynchronous
	// sweeping in the next sweep phase.
	s.sweepgen = mheap_.sweepgen + 3

	// Assume all objects from this span will be allocated in the
	// mcache. If it gets uncached, we'll adjust this.
	c.localSmallAllocN[spc.sizeclass()] += uintptr(s.nelems) - uintptr(s.allocCount)

	// Update heap_live with the same assumption.
	usedBytes := uintptr(s.allocCount) * s.elemsize
	atomic.Xadd64(&memstats.heap_live, int64(s.npages*pageSize)-int64(usedBytes))

	// While we're here, flush localScan, since we have to call
	// revise anyway.
	atomic.Xadd64(&memstats.heap_scan, int64(c.localScan))
	c.localScan = 0

	if trace.enabled {
		// heap_live changed.
		traceHeapAlloc()
	}
	if gcBlackenEnabled != 0 {
		// heap_live and heap_scan changed.
		gcController.revise()
	}

	c.alloc[spc] = s
}

// largeAlloc allocates a span for a large object.
//
// While this method doesn't interact with the mcache directly,
// it helps solidify the mcache as the first layer of allocation.
//
// Furthermore, the statistics for large allocations are stored
// in the mcache.
func (c *mcache) largeAlloc(size uintptr, needzero bool, noscan bool) *mspan {
	if size+_PageSize < size {
		throw("out of memory")
	}
	npages := size >> _PageShift
	if size&_PageMask != 0 {
		npages++
	}

	// Deduct credit for this span allocation and sweep if
	// necessary. mHeap_Alloc will also sweep npages, so this only
	// pays the debt down to npage pages.
	deductSweepCredit(npages*_PageSize, npages)

	spc := makeSpanClass(0, noscan)
	s := mheap_.alloc(npages, spc, needzero)
	if s == nil {
		throw("out of memory")
	}
	c.localLargeAlloc += npages * pageSize
	c.localLargeAllocN++

	// Update heap_live and revise pacing if needed.
	atomic.Xadd64(&memstats.heap_live, int64(npages*pageSize))
	if trace.enabled {
		// Trace that a heap alloc occurred since heap_live changed.
		traceHeapAlloc()
	}
	if gcBlackenEnabled != 0 {
		gcController.revise()
	}

	if go115NewMCentralImpl {
		// Put the large span in the mcentral swept list so that it's
		// visible to the background sweeper.
		mheap_.central[spc].mcentral.fullSwept(mheap_.sweepgen).push(s)
	}
	s.limit = s.base() + size
	heapBitsForAddr(s.base()).initSpan(s)
	return s
}

func (c *mcache) releaseAll() {
	// Take this opportunity to flush localScan.
	atomic.Xadd64(&memstats.heap_scan, int64(c.localScan))
	c.localScan = 0

	sg := mheap_.sweepgen
	for i := range c.alloc {
		s := c.alloc[i]
		if s != &emptymspan {
			// Adjust nsmallalloc in case the span wasn't fully allocated.
			n := uintptr(s.nelems) - uintptr(s.allocCount)
			c.localSmallAllocN[spanClass(i).sizeclass()] -= n
			if s.sweepgen != sg+1 {
				// refill conservatively counted unallocated slots in heap_live.
				// Undo this.
				//
				// If this span was cached before sweep, then
				// heap_live was totally recomputed since
				// caching this span, so we don't do this for
				// stale spans.
				atomic.Xadd64(&memstats.heap_live, -int64(n)*int64(s.elemsize))
			}
			// Release the span to the mcentral.
			mheap_.central[i].mcentral.uncacheSpan(s)
			c.alloc[i] = &emptymspan
		}
	}
	// Clear tinyalloc pool.
	c.tiny = 0
	c.tinyoffset = 0
}

// prepareForSweep flushes c if the system has entered a new sweep phase
// since c was populated. This must happen between the sweep phase
// starting and the first allocation from c.
func (c *mcache) prepareForSweep() {
	// Alternatively, instead of making sure we do this on every P
	// between starting the world and allocating on that P, we
	// could leave allocate-black on, allow allocation to continue
	// as usual, use a ragged barrier at the beginning of sweep to
	// ensure all cached spans are swept, and then disable
	// allocate-black. However, with this approach it's difficult
	// to avoid spilling mark bits into the *next* GC cycle.
	sg := mheap_.sweepgen
	if c.flushGen == sg {
		return
	} else if c.flushGen != sg-2 {
		println("bad flushGen", c.flushGen, "in prepareForSweep; sweepgen", sg)
		throw("bad flushGen")
	}
	c.releaseAll()
	stackcache_clear(c)
	atomic.Store(&c.flushGen, mheap_.sweepgen) // Synchronizes with gcStart
}
