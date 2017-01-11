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

	// The rest is not accessed on every malloc.
	alloc           [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass
	largeAllocSpans *mspan
	stackcache      [_NumStackOrders]stackfreelist

	// Local allocator stats, flushed during GC.
	local_nlookup    uintptr                  // number of pointer lookups
	local_largefree  uintptr                  // bytes freed for large objects (>maxsmallsize)
	local_nlargefree uintptr                  // number of frees for large objects (>maxsmallsize)
	local_nsmallfree [_NumSizeClasses]uintptr // number of frees for small objects (<=maxsmallsize)

	// Each mcache is in a rocEpoch that corresponds to this mcache's associations with a
	// specific G. Before an mcache can leave an epoch it must ensure that all reachable
	// local objects associated with previous epoch's G have been made public. Incrementing
	// rocEpoch indicates that this has been done.
	// When a syscall is returned from if the G can reacquire the same mcache and P
	// and the mcache has the same rocEpoch as when the syscall was made then the G can simply
	// continue to use the mcache. On the otherhand if the mcache (and P) have been retaken
	// by another G then the the G returning from the syscall must wait for the rocEpoch to be
	// > than when the syscall was initiated.
	rocEpoch uint64
	rocGoid  int64 // The goid associated with this mcache
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
// to be the cached span for the given spanClass. Returns this span.
func (c *mcache) refill(spc spanClass) *mspan {
	_g_ := getg()

	_g_.m.locks++
	// Return the current cached span to the central lists.
	olds := c.alloc[spc]

	// Get a new cached span from the central lists.

	// ROC:
	// empty means the there is no unused objects that can be used
	// for allocation. Historically this indicated an empty freelist.
	//
	// cacheSpan returns an *mspan to allocate from.
	//
	// If recycleG recovers an object making the span nonempty
	// then the s *mspan is placed on the spanClass's mcentral's nonempty
	// list.
	//
	// Multiple spans from the same spanClass can be used to allocate
	// during a ROC epoch. As spans are used up they are linked onto
	// the new span via the nextUsedSpan field and the new span replaces
	// it in the alloc[spanClass] cache. recycleG and publishG
	// iterated down this list to recycle or publish the objects.

	s := mheap_.central[spc].mcentral.cacheSpan()
	if s == nil {
		throw("out of memory")
	}

	if s.nextUsedSpan != nil {
		println("runtime: s.base()=", hex(s.base()), "s.nextUsedSpan.base()=", hex(s.nextUsedSpan.base()))
		throw("s.nextUsedSpan is not nil.")
	}

	// Update the ROC structures to include this new span.
	s.startindex = s.freeindex
	if c.alloc[spc] != &emptymspan {
		if writeBarrier.roc {
			s.nextUsedSpan = c.alloc[spc]
		}
		if !c.alloc[spc].incache {
			throw("c.alloc[spc].incache is not true.")
		}
		c.alloc[spc].incache = false
	} else {
		s.nextUsedSpan = nil
	}
	s.incache = true
	c.alloc[spc] = s

	if olds != &emptymspan {
		olds.incache = false
		olds.trace("refill sets incache to false")

		if olds.nelems != olds.freeindex {
			println("runtime: refill called with olds.nelems=", olds.nelems, "olds.freeindex=", olds.freeindex)
			throw("olds.nelems != olds.freeindex")
		}
	}

	_g_.m.locks--
	s.trace("span returned from refill")
	return s
}

// publishAllGs publishes all local objects.
// The world is stopped.
func publishAllGs() {
	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.publishAllGsCalls, 1)
	}
	for _, p := range &allp {
		if p == nil || p.mcache == nil {
			continue
		}
		systemstack(p.mcache.publishG)
	}
}

// releaseAll releases all spans associated with this p's mcache.
func (c *mcache) releaseAll() {
	if writeBarrier.roc {
		// publish all local objects in the current ROC epoch.
		systemstack(c.publishG)
	}

	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.releaseAllCalls, 1)
	}
	for i := range c.alloc {
		s := c.alloc[i]
		// ROC can introduce a span without allocated objects in it where
		// this was previously not possible. Such spans are uncached
		// so they can be reused by any spanClass.
		if writeBarrier.roc {
			if s != &emptymspan {
				next := c.alloc[i]
				for s := next; s != nil; s = next {
					next = s.nextUsedSpan
					s.nextUsedSpan = nil
					if s.allocCount == 0 {
						mheap_.central[i].mcentral.uncacheNoAllocsSpan(s)
					} else if s.allocCount < s.nelems {
						mheap_.central[i].mcentral.uncacheSpan(s)
					} else {
						// allocCount == s.elems so the "freelist" is empty.
						if s.allocCount != s.nelems {
							throw(" allocCount should not be > s.nelems")
						}
						mheap_.central[i].mcentral.uncacheSpan(s)
					}
				}
				c.alloc[i] = &emptymspan
			}
		} else {
			if s != &emptymspan {
				mheap_.central[i].mcentral.uncacheSpan(s)
				c.alloc[i] = &emptymspan
			}
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
