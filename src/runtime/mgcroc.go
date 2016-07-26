// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The Core routine for the ROC (request oriented collector) algorithm.
//
// There are three basic operations involved in the ROC algorithm.
//
// startG: This starts a new ROC epoch related to the currently running goroutine.
// The epoch ends when the goroutine exits or the cost of tracking the published
// objects no longer warrents the expense.
//
// recycleG: recycle the spans, allocating reusing the space left by allocated
// but not published objects. The goroutine is exiting or (TBD) becomes dormant
// when this is called.
//
// publishG: publish all objects allocated by this G. The cost of maintaining
// which objects are local and which are public exceeds the value of recycling
// them. The next GC cycle will reclaim the unreachable objects.
//

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

// startG establishes a new ROC epoch associated with the
// currently runnging G.
// The epoch associated with any previous G will be discared.
// Spans not currently being allocated from but that are part of
// this epoch are linked through the nextUsedSpan.
// There will be placed on the appropriate empty list so and released.
func (c *mcache) startG() {
	mp := acquirem()
	_g_ := mp.curg
	if _g_ == nil {
		_g_ = getg()
	}
	_g_.rocvalid = isGCoff()      // ROC optimizations are invalidated if a GC is in process.
	_g_.rocgcnum = memstats.numgc // thread safe since it is only modified in mark termination which is STW

	for i := range c.alloc {
		s := c.alloc[i]
		if s != &emptymspan {
			s.startindex = s.freeindex
			// Since we are only remembering the last goroutine
			// we need to undo the nextUsedSpan list.
			next := s.nextUsedSpan
			s.nextUsedSpan = nil
			for s = next; s != nil; s, s.nextUsedSpan = s.nextUsedSpan, nil {
				s.startindex = s.freeindex // The write barrier uses this to see if a slot is between
				// startindex and freeindex. If so it is in the current
				// rollback eligible area and is local if its allocation
				// bit is not set.
				if s.allocated() != s.nelems {
					// Only the first span (c.alloc[i]) should have free objecs.
					throw("s.alloced() should != s.nelems")
				}
				mheap_.central[i].mcentral.releaseROCSpan(s)
			}
		}
	}
	// clean up tiny logic
	c.tiny = 0
	c.tinyoffset = 0
	if _g_.rocvalid {
		_g_.rocvalid = isGCoff()
	}
	c.rocgoid = _g_.goid // To support debugging.
	releasem(mp)
}

// publishG is called when the ROC epoch need to have all
// the its local objects published. This happens when it
// is no longer feasible to track the local objects.
func (c *mcache) publishG() {
	if debug.gcroc == 0 {
		throw("publishG called but debug.gcroc < 1")
	}
	mp := acquirem()
	_g_ := getg().m.curg
	for i := range c.alloc {
		if c.alloc[i] == &emptymspan {
			continue
		}
		if !c.alloc[i].incache {
			println("runtime: i=", i, "gcphase=", gcphase, "mheap_.sweepgen=", mheap_.sweepgen, "c.alloc[i].sweepgen=", c.alloc[i].sweepgen)
			throw("c.alloc[i].incache should be true")
		}
		for s := c.alloc[i]; s != nil; s, s.nextUsedSpan = s.nextUsedSpan, nil {
			if s == &emptymspan {
				throw("s == &emptymspan")
			}
			if s.elemsize == 0 {
				throw("s.elemsize == 0")
			}
			if s != c.alloc[i] {
				if s.incache {
					// only the first span is considered in an mcache
					throw("publishG encounters span that should not be marked incache")
				}
				if s.freeindex != s.nelems {
					throw("s.freeindex != s.nelems and span is on ROC used list.")
				}
				s.startindex = s.freeindex
				mheap_.central[i].mcentral.releaseROCSpan(s) // empty free list so leave s on empty list....
			} else {
				// This is the active span in the mcache.
				s.startindex = s.freeindex
			}
			s.abortRollbackCount++ // Save some statistics
		}
	}
	// publish all the largeAllocSpans
	for s := c.largeAllocSpans; s != nil; s, s.nextUsedSpan = s.nextUsedSpan, nil {
		// aborting rollback so just release the spans after adjusting allocCount to s.nelems.
		if s.freeindex != s.nelems {
			throw("s.freeindex != s.nelems and span is on ROC incache used largeAllocSpan list.")
		}
		s.startindex = s.freeindex
	}
	c.largeAllocSpans = nil
	if _g_ != nil {
		_g_.rocvalid = false
		_g_.rocgcnum = 0
	}
	// clean up tiny logic
	c.tiny = 0
	c.tinyoffset = 0
	releasem(mp)
}

// recycleG recycles spans that were used for allocation in the
// ROC epoch that is ending. The spans allocBits have been updated
// since the start of the ROC epoch by marking publich objects.
// Local objects that are now no longer reachable will have
// the allocBit cleared and will be available for allocation.
// The actual recycle is done by setting each spans freeindex
// back to the startindex associated with the span and set in
// startG.

func (c *mcache) recycleG() {
	mp := acquirem()
	c.recycleNormal()
	releasem(mp)
}

// Keeps track of the total number of bytes ROC recovers.
var rocRecoveredBytes uint64

// The spans linked by nextUsedSpan are on the empty list but the
// recycle is about to release some objects making them non-empty.
// This means we must pull them off the empty list and put them
// back onto the nonempty list.
func (c *mcache) recycleNormal() {
	if debug.gcroc == 0 {
		throw("recycleNormal called but debug.gcroc < 1")
	}

	_g_ := getg().m.curg
	recycleValid := _g_ != nil && _g_.rocvalid && _g_.rocgcnum == memstats.numgc && isGCoff()

	if !recycleValid {
		c.publishG()
		if _g_ != nil {
			_g_.rocvalid = false // reset in startG
			_g_.rocgcnum = 0
		}
		return
	}

	if c.rocgoid != _g_.goid {
		println("c.rocgoid=", c.rocgoid, "_g_.goid=", _g_.goid)
		throw("c.rocgoid != _g_.goid")
	}

	// Count of the number of objects recovered using ROC
	recoveredBytes := uintptr(0)
	for i := range c.alloc {
		if c.alloc[i] == &emptymspan {
			continue
		}
		if !c.alloc[i].incache {
			throw("c.alloc[i].incache should be true")
		}
		for s := c.alloc[i]; s != nil; s, s.nextUsedSpan = s.nextUsedSpan, nil {
			if s == &emptymspan {
				throw("s == &emptymspan")
			}
			if s == nil {
				throw("s is == for some reason.")
			}
			if s.elemsize == 0 {
				throw("s.elemsize == 0")
			}
			if s.allocBits == nil {
				throw("s.allocBits == nil")
			}
			if s != c.alloc[i] && s.incache {
				// only the first span is considered in an mcache
				throw("recycleG encounters span that should no be incache")
			}

			// As an optimization move s.startindex past all objects that are now public
			for ii := s.startindex; ii < s.freeindex; ii++ {
				if s.isFree(ii) {
					break
				}
				s.startindex++ // no sense in rolling back over public objects, set startindex and then freeindex to first free object.
			}

			for ii := s.startindex; ii < s.freeindex; ii++ {
				if s.isFree(ii) {
					recoveredBytes += s.elemsize
				}
			}

			// s.smashDebugHelper() // this increases the chance of triggering a bug

			s.freeindex = s.startindex
			s.recycleCount++
			if s.freeindex == s.nelems {
				s.allocCache = 0 // Clear it since this span is full.
			} else {
				// Reset allocCache
				if s.freeindex > s.nelems {
					throw("s.freeindx > s.nelems")
				}
				freeByteBase := s.freeindex &^ (64 - 1)
				whichByte := freeByteBase / 8
				if whichByte > s.nelems/8 {
					throw("whichByte > s.nelems / 8")
				}
				s.refillAllocCache(whichByte)
				// adjust the allocCache so that s.freeindex corresponds to the low bit in
				// s.allocCache
				s.allocCache >>= s.freeindex % 64
			}
			// If this span is not the active alloc span
			// either free it if has no alloced objects or simply uncache
			// it if it has available space for new objects.
			if s != c.alloc[i] {
				mheap_.central[i].mcentral.releaseROCSpan(s)
			}
		}
	}

	// Large objects, one per span.
	// abort rollback of largeAllocSpans
	for s := c.largeAllocSpans; s != nil; s, s.nextUsedSpan = s.nextUsedSpan, nil {
		// aborting rollback so just release the spans after adjusting allocCount to s.nelems.
		if s.freeindex != s.nelems {
			throw("s.freeindex != s.nelems and span is on ROC incache used largeAllocSpan list.")
		}
		if s.allocated() != s.nelems {
			throw("s.allocated() != s.nelems and span is on ROC incache used largerAllocSpan list.")
		}
		if false && s.isFree(0) {
			// TBD: Make entire span available.
			recoveredBytes += s.elemsize
		} else {
			s.startindex = s.freeindex
		}
	}

	atomic.Xadd64(&rocRecoveredBytes, int64(recoveredBytes)) // TBD make available, perhaps as part of gctrace=2

	c.largeAllocSpans = nil

	if _g_ != nil {
		_g_.rocvalid = false
		_g_.rocgcnum = 0
	}
	// clean up tiny logic
	c.tiny = 0
	c.tinyoffset = 0
}

// smash will obliterate the contents of any free objects in the hopes that
// this will cause the program to abort quickly and debugging easier.
func (s *mspan) smashDebugHelper() {
	// Smash object between s.startindex and freeindex
	for i := s.startindex; i < s.freeindex; i++ {
		if s.isFree(i) {
			words := s.elemsize / unsafe.Sizeof(uintptr(0))
			for j := uintptr(0); j < words; j++ {
				ptr := (*uintptr)(unsafe.Pointer(s.base() + i*s.elemsize + j*unsafe.Sizeof(uintptr(0))))
				*ptr = uintptr(0xdeada11c)
			}
		}
	}
}
