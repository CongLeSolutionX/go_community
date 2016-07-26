// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The Core routine for the ROC (request oriented collector) algorithm.
//
// There are three basic operations involved in the ROC algorithm.
//
// startG: This starts a new ROC epoch related to the currently running goroutine.
// The epoch ends when the goroutine exits or the cost of tracking the published
// objects no longer warrants the expense.
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
// currently running G.
// The epoch associated with any previous G will be discarded.
// Spans not currently being allocated from but that are part of
// this epoch are linked through the nextUsedSpan.
// There will be placed on the appropriate empty list so and released.
func (c *mcache) startG() {
	if debug.gcroc == 0 {
		throw("startG called but debug.gcroc == 0")
	}
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
			for s = next; s != nil; s = next {
				s.startindex = s.freeindex // The write barrier uses this to see if a slot is between
				// startindex and freeindex. If so it is in the current
				// rollback eligible area and is local if its allocation
				// bit is not set.
				if s.allocated() != s.nelems {
					// Only the first span (c.alloc[i]) should have free objects.
					throw("s.alloced() should != s.nelems")
				}
				next = s.nextUsedSpan
				s.nextUsedSpan = nil
				mheap_.central[i].mcentral.releaseROCSpan(s) // nextUsedSpan is nil since this can be reused immediately
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
		next := c.alloc[i]
		for s := next; s != nil; s = next {
			next = s.nextUsedSpan
			s.nextUsedSpan = nil
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
		s.allocCount = s.nelems
		s.abortRollbackCount++ // Save some statistics
		s.checkAllocCount(s.freeindex)
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
// since the start of the ROC epoch by marking public objects.
// Local objects that are now no longer reachable will have
// the allocBit cleared and will be available for allocation.
// The actual recycle is done by setting each spans freeindex
// back to the startindex associated with the span and set in
// startG.
// caller must have done an acquirem so this routine can't
// switch Ps.
func (c *mcache) recycleG() {
	if debug.gcroc == 0 {
		throw("in recycleG but debug.gcroc == 0")
	}
	c.recycleNormal()
}

// Keeps track of the total number of bytes ROC recovers.
type rocTrace struct {
	recoveredBytes    uint64
	recoveredBytesAll uint64
}

var rocData rocTrace

// The spans linked by nextUsedSpan are on the empty list but the
// recycle is about to release some objects making them non-empty.
// This means we must pull them off the empty list and put them
// back onto the nonempty list.
func (c *mcache) recycleNormal() {
	if debug.gcroc == 0 {
		throw("recycleNormal called but debug.gcroc < 1")
	}
	mheapsweepgen := mheap_.sweepgen
	_g_ := getg().m.curg
	recycleValid := _g_ != nil && _g_.rocvalid && _g_.rocgcnum == memstats.numgc && isGCoff()

	if !recycleValid {
		systemstack(c.publishG)
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
	recoveredBytes := int64(0)
	for i := range c.alloc {
		if c.alloc[i] == &emptymspan {
			continue
		}
		if !c.alloc[i].incache {
			throw("c.alloc[i].incache should be true")
		}
		next := c.alloc[i]
		for s := next; s != nil; s = next {
			next = s.nextUsedSpan
			s.nextUsedSpan = nil
			ssweepgen := s.sweepgen
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
				println("runtime: c.alloc[i].base()=", hex(c.alloc[i].base()),
					"runtime: s.base()=", hex(s.base()))
				// only the first span is considered in an mcache
				throw("recycleG encounters span that should not be incache")
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
					recoveredBytes += int64(s.elemsize)
				}
			}

			s.smashDebugHelper() // this increases the chance of triggering a bug

			s.freeindex = s.startindex // The actual recycle step.
			s.rollbackCount++
			if debug.gcroc < 2 {
				s.rollbackAllocCount()
			} else {
				s.rollbackAllocCountDebug(s.freeindex, ssweepgen, mheapsweepgen, 253)
			}
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
			recoveredBytes += int64(s.elemsize)
		} else {
			s.startindex = s.freeindex
			s.allocCount = s.nelems
			s.checkAllocCount(s.freeindex)
		}
	}

	atomic.Xadd64(&rocData.recoveredBytes, int64(recoveredBytes)) // TBD make available, perhaps as part of gctrace=2

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
	if debug.gcroc >= 2 {
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
}

// Some debugging support routines. These routines are ad-hoc and provide
// support when debugging ROC. They are meant to be modified so that they
// focus on whatever the current bug is. Their use is triggered only when
// debug.gcroc < 2.

func (s *mspan) checkAllocCount(tmpFreeIndex uintptr) bool {
	if debug.gcroc < 2 {
		return true // short circuit unless gcroc > 1
	}

	if atomic.Load(&gcphase) != _GCoff {
		return true
	}
	oldrollbackCount := s.rollbackCount
	oldabortRollbackCount := s.abortRollbackCount

	oldsweepgen := atomic.Load(&s.sweepgen)
	oldgcphase := gcphase
	badCheck := false
	if debug.gcroc == 0 {
		return true
	}
	if s.freeindex != tmpFreeIndex {
		badCheck = true
	}
	if s.freeindex == s.nelems {
		if uintptr(s.allocCount) != s.nelems {
			if atomic.Load(&gcphase) != _GCoff {
				return true
			}
			println("runtime: s.allocCount=", s.allocCount, "s.freeindex=", s.freeindex, "s.startindex=", s.startindex,
				"\n     s.nelems=", s.nelems, "s.base()=", hex(s.base()))

			for i := uintptr(0); i < s.nelems; i++ {
				if !s.isFree(i) {
					if i == s.freeindex {
						println("+++ s.freeindex---> ")
					}
					println("+++ s.isFree(", i, ")=", s.isFree(i))
				} else {
					if i == s.freeindex {
						println("--- s.freeindex---> ")
					}
					println("--- s.isFree(", i, ")=", s.isFree(i))
				}
			}
			println("runtime: checkAllocCount:544:  uintptr(s.allocCount) != s.nelems ",
				" s.allocCount=", s.allocCount,
				"s.nelems=", s.nelems,
				"s.freeindex=", s.freeindex,
				"tmpFreeIndex=", tmpFreeIndex,
				"\n     s.startindex=", s.startindex,
				"s.isFree(s.startindex)=", s.isFree(s.startindex),
				"s.isFree(s.freeindex)=", s.isFree(s.freeindex),
				"s.allocCache=", hex(s.allocCache),
				"\n     s.rollbackCount=", s.rollbackCount,
				"s.base()=", hex(s.base()),
				"tmpFreeIndex=", tmpFreeIndex,
				"badCheck=", badCheck,
				"s.sweepgen=", atomic.Load(&s.sweepgen),
				"oldsweepgen=", oldsweepgen,
				"\n     s.rollbackCount=", s.rollbackCount,
				"oldrollbackcount=", oldrollbackCount,
				"s.abortRollbackCount=", s.abortRollbackCount,
				"oldabortRollbackCount=", oldabortRollbackCount,
				"oldgcphase=", oldgcphase,
				"gcphase=", gcphase)
			return false
		}
	}
	if s.allocCount > s.nelems {
		if atomic.Load(&gcphase) != _GCoff {
			return true
		}
		println("runtime: checkAllocCount:559: s.allocCount=", s.allocCount, "s.freeindex=", s.freeindex, "s.startindex=", s.startindex,
			"\n     s.nelems=", s.nelems, "s.base()=", hex(s.base()))
		return false
	}
	bitCount := s.freeindex
	for i := s.freeindex; i < s.nelems; i++ {
		if !s.isFree(i) {
			bitCount++
		}
	}
	if uintptr(bitCount) != s.allocCount && gcphase == _GCoff {
		if atomic.Load(&gcphase) != _GCoff {
			return true
		}

		for i := uintptr(0); i < s.nelems; i++ {
			if !s.isFree(i) {
				if i == s.freeindex {
					println("+++ s.freeindex---> ")
				}
				println("+++ s.isFree(", i, ")=", s.isFree(i))
			} else {
				if i == s.freeindex {
					println("--- s.freeindex---> ")
				}
				println("--- s.isFree(", i, ")=", s.isFree(i))
			}
		}

		println("runtime malloc.go:578: bitCount != allocated bitCount bitCount=", bitCount,
			" s.allocCount=", s.allocCount,
			"s.nelems=", s.nelems,
			"s.freeindex=", s.freeindex,
			"tmpFreeIndex=", tmpFreeIndex,
			"\n     s.startindex=", s.startindex,
			"s.isFree(s.startindex)=", s.isFree(s.startindex),
			"s.isFree(s.freeindex)=", s.isFree(s.freeindex),
			"s.allocCache=", hex(s.allocCache),
			"\n     s.rollbackCount=", s.rollbackCount,
			"s.base()=", hex(s.base()),
			"tmpFreeIndex=", tmpFreeIndex,
			"badCheck=", badCheck,
			"s.sweepgen=", atomic.Load(&s.sweepgen),
			"oldsweepgen=", oldsweepgen,
			"s.rollbackCount=", s.rollbackCount,
			"oldrollbackcount=", oldrollbackCount,
			"s.abortRollbackCount=", s.abortRollbackCount,
			"oldabortRollbackCount=", oldabortRollbackCount)

		return false
	}
	return true
}

func dumpBrokenIsPublicState(s *mspan, oldSweepgen uint32, obj uintptr, abits markBits, sg uint32) {
	if debug.gcroc < 2 {
		return // short circuit unless gcroc > 1
	}
	oldnumgc := memstats.numgc
	// oldfreeindex := s.freeindex
	//	oldgcphase := gcphase
	oldstartindex := s.startindex
	oldrollbackCount := s.rollbackCount
	oldabortRollbackCount := s.abortRollbackCount

	newSweepgen := atomic.Load(&s.sweepgen)
	if oldSweepgen != newSweepgen || mheap_.sweepgen != oldSweepgen {
		// mheap_.sweepgen will always be even. If oldSweepgen (and/or newSweepgen) is odd then a sweep is in progress.
		abits := s.allocBitsForAddr(obj)
		if abits.isMarked() {
			return
		} else {
			throw("isPublic has a pointer into a span that just got swept but isMarked is false. which may be OK but we need to check if it is owned by this goroutine.")
		}
	}

	println("the allocation bit index is >=freeindex and it is not set. ",
		"\n if it is local it could be allocated over, if it is global this goroutine should not see it.",
		"\nmbitmap.go:270 isPublic MemStats.NumGC ", memstats.numgc, "abits.index=", abits.index,
		"s.freeindex=", s.freeindex, "obj=", hex(obj),
		"\n     s.elemsize=", s.elemsize, "s.base()=", hex(s.base()),
		"s.nelems=", s.nelems, "gcphase=", gcphase, "s.startindex=", s.startindex,
		"\n     s.rollbackCount=", s.rollbackCount,
		"s.abortRollbackCount", s.abortRollbackCount, "s.spanclass=", s.spanclass,
		"\n     oldnumgc=", oldnumgc, "oldSweepgen=", oldSweepgen, "s.sweepgen=", s.sweepgen,
		"abits.bytep", abits.bytep, "gcphase=", gcphase,
		//"oldnumgc :=", oldnumgc,
		//	"oldfreeindex :=", oldfreeindex,
		//			"oldgcphase :=", oldgcphase,
		"\n     oldstartindex :=", oldstartindex,
		"oldrollbackCount :=", oldrollbackCount,
		"oldabortRollbackCount :=", oldabortRollbackCount,
		"sg=", sg,
		"mheap_.sweepgen=", mheap_.sweepgen,
		"\n     s.allocCount", s.allocCount,
		"\n is s.allocBitsForAddr(obj).isMarked()= still false", s.allocBitsForAddr(obj).isMarked())

	if getg() == nil {
		println("     getg()== nil")
	} else if getg().m == nil {
		println("     getg().m == nil, getg().goid=", getg().goid)
		if getg().m.curg != nil {
			println("     getg().m.curg.goid=", getg().m.curg.goid)
		}
	} else if getg().m.mcache == nil {
		println("     getg().m.mcache == nil")
	} else if getg().m.mcache.alloc[s.spanclass] == &emptymspan {
		println("     getg().m.mcache.alloc[s.spanclass] == &emptyspan")
	} else {
		println("if the span is on this list it is local to this goroutine.")
		for spanList := getg().m.mcache.alloc[s.spanclass]; spanList != &emptymspan; spanList = spanList.nextUsedSpan {
			if spanList == nil {
				println("     spanList == nil")
				break
			}
			println("     spanList.base()=", hex(spanList.base()))
		}
	}
	println("first abits.bytep=", abits.bytep)
	newAbits := s.allocBitsForAddr(obj)
	println("new abits.bytep=", newAbits.bytep, "newAbits.isMarked()=", newAbits.isMarked(), "mheap_.sweepgen=", mheap_.sweepgen)
	// Perhaps check and force a sweep to make sure things are up to date.
}

// Not to be submitted.
func (c *mcentral) releaseROCSpanDebug(s *mspan) {
	if debug.gcroc < 2 {
		return // short circuit unless gcroc > 1
	}

	lock(&c.lock)
	s.incache = false
	s.trace("releaseROCSpan set incache to false")
	// 3 possibilities.
	objectsAllocated := s.allocCount
	/****
		// First check that allocCount is accurate, if it isn't report it as a possible bug
		// but do not throw.

		bitCount := s.freeindex
		for i := s.freeindex; i < s.nelems; i++ {
			if !s.isFree(i) {
				bitCount++
			}
		}

		if uintptr(bitCount) != s.allocCount && atomic.Load(&gcphase) != _GCoff {
			println("releaseROCSpan has bad s.allocCount")
			for i := uintptr(0); i < s.nelems; i++ {
				if !s.isFree(i) {
					if i == s.freeindex {
						println("+++ s.freeindex---> ")
					}
					println("+++ s.isFree(", i, ")=", s.isFree(i))
				} else {
					if i == s.freeindex {
						println("--- s.freeindex---> ")
					}
					println("--- s.isFree(", i, ")=", s.isFree(i))
				}
			}

			println("releaseROCSpan mcentral.go:193: bitCount != allocated bitCount bitCount=", bitCount,
				" s.allocCount=", s.allocCount,
				"s.nelems=", s.nelems,
				"s.freeindex=", s.freeindex,
				"\n     s.startindex=", s.startindex,
				"s.isFree(s.startindex)=", s.isFree(s.startindex),
				"s.isFree(s.freeindex)=", s.isFree(s.freeindex),
				"s.allocCache=", hex(s.allocCache),
				"\n     s.recycleCount=", s.recycleCount,
				"s.base()=", hex(s.base()),
				"s.sweepgen=", atomic.Load(&s.sweepgen),
				"s.recycleCount=", s.recycleCount,
				"s.abortRollbackCount=", s.abortRollbackCount)
			s.allocCount = bitCount // Fix up for now...
		}
	****/
	if objectsAllocated == s.nelems {
		// 1. Free list is empty so put on empty freelist list.
		c.empty.remove(s)
		c.empty.insert(s)
		s.trace("releaseROCSpan leaves span on empty list")
	} else if objectsAllocated > 0 {
		// 2. freelist is not empty but entire span is not free, put on nonempty freelist list
		c.empty.remove(s)
		c.nonempty.insert(s) // nonempty free list
		s.trace("releaseROCSpan moves span to nonempty list")
	} else {
		// 3. entire span is empty, leave it on nonempty freelist list instead
		// of calling freeSpan. The likelihood that it will be needed
		// is high.
		// Another reasonable approach would be to call freeSpan with
		// this span.
		c.empty.remove(s)
		c.nonempty.insert(s) // nonempty free list
		s.trace("releaseROCSpan moves span in no allocations to nonempty list")
	}
	atomic.Xadd64(&memstats.heap_live, -int64(s.nelems-objectsAllocated)*int64(s.elemsize))
	unlock(&c.lock)
}
