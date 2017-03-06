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
	"runtime/internal/sys"
	"unsafe"
)

// startG establishes a new ROC epoch associated with the
// currently running G.
// The epoch associated with any previous G will be discarded.
// Spans not currently being allocated from but that are part of
// this epoch are linked through the nextUsedSpan.
// There will be placed on the appropriate empty list so and released.
func (c *mcache) startG() {
	if !writeBarrier.roc {
		throw("startG called but writeBarrier.roc is false")
	}
	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.startGCalls, 1)
	}
	mp := acquirem()
	_g_ := mp.curg
	if _g_ == nil {
		_g_ = getg()
	}
	_g_.rocvalid = isGCoff()      // ROC optimizations are invalidated if a GC is in process.
	_g_.rocgcnum = memstats.numgc // thread safe since it is only modified in mark termination which is STW

	//	_g_.rocTries++
	//	_g_.rocOn = _g_rocTries <= 2  // Turn roc off if we are not seeing any success.

	for i := range c.alloc {
		s := c.alloc[i]
		if s != &emptymspan {
			s.startindex = s.freeindex
			// Since we are only remembering the last goroutine
			// we need to undo the nextUsedSpan list.
			next := s.nextUsedSpan
			s.nextUsedSpan = nil
			for s = next; s != nil; s = next {
				s.startindex = s.freeindex
				// The write barrier uses this to see if a slot is between
				// startindex and freeindex. If so it is in the current
				// rollback eligible area and is local if its allocation
				// bit is not set.
				if s.allocated() != s.nelems {
					// Only the first span (c.alloc[i]) should have free objects.
					throw("s.alloced() should != s.nelems")
				}
				next = s.nextUsedSpan
				s.nextUsedSpan = nil
				// Note that this does not release the span in c.alloc[i] which remains in the mcache.
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
	afterFirstStartg = true
}

// publishG is called when the ROC epoch need to have all
// the its local objects published. This happens when it
// is no longer feasible to track the local objects.
func (c *mcache) publishG() {
	if !writeBarrier.roc {
		throw("publishG called but writeBarrier.roc is false")
	}

	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.publishGCalls, 1)
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
	}
	c.largeAllocSpans = nil
	if _g_ != nil {
		_g_.rocvalid = true
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
	if !writeBarrier.roc {
		throw("in recycleG but writeBarrier.roc is false")
	}
	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.recycleGCalls, 1)
	}

	c.recycleNormal()
}

// Keeps track of the total number of bytes ROC recovers.
type rocTrace struct {
	recoveredBytes                   uint64
	recoveredBytesAll                uint64
	recycleGCalls                    uint64
	recycleGCallsFailure             uint64
	failureDueTogisnil               uint64
	failureDueTognotrocvalid         uint64
	failureDueTogbadnumgc            uint64
	failureDueTogcoff                uint64
	publishGCalls                    uint64
	startGCalls                      uint64
	releaseAllCalls                  uint64
	publishAllGsCalls                uint64
	dropgCalls                       uint64
	parkCalls                        uint64
	goschedImplCalls                 uint64
	entersyscallCalls                uint64
	exitsyscall0Calls                uint64
	goexit0Calls                     uint64
	publicToLocal                    uint64
	publicToPublic                   uint64
	localToPublic                    uint64
	localToLocal                     uint64
	mumbleToMumble                   uint64
	makePublicCount                  uint64
	makePublicAlreadyMarked          uint64
	stacksPublished                  uint64
	framesPublished                  uint64
	maxFramesPublished               uint64
	oneFramesPublished               uint64
	twoToTenFramesPublished          uint64
	elevenToFiftyFramesPublished     uint64
	fiftyOneToHundredFramesPublished uint64
	overHundredFramesPublished       uint64
}

// rocReasonType are the reasons that recycleG and publishG are being called.
// Typically they are used to maintain the statistics needed to understand
// the  programs actions.
type rocReasonType int

const (
	rocExit rocReasonType = iota
	rocPark
	rocSyscall
	rocPreempt
)

var rocData rocTrace
var rocDataPrevious rocTrace

var gisnil, gnotrocvalid, gbadnumgc, gcoff uint64

// The spans linked by nextUsedSpan are on the empty list but the
// recycle is about to release some objects making them non-empty.
// This means we must pull them off the empty list and put them
// back onto the nonempty list.
func (c *mcache) recycleNormal() {
	if !writeBarrier.roc {
		throw("in recycleNormal but writeBarrier.roc is false")
	}
	_g_ := getg().m.curg

	recycleValid := _g_ != nil && _g_.rocvalid && _g_.rocgcnum == memstats.numgc && isGCoff()

	if !recycleValid {
		if debug.gcroc == 2 {
			atomic.Xadd64(&rocData.recycleGCallsFailure, 1)
			if _g_ == nil {
				atomic.Xadd64(&rocData.failureDueTogisnil, 1)
			} else if !_g_.rocvalid {
				atomic.Xadd64(&rocData.failureDueTognotrocvalid, 1)
			} else if _g_.rocgcnum != memstats.numgc {
				atomic.Xadd64(&rocData.failureDueTogbadnumgc, 1)
			} else if isGCoff() {
				atomic.Xadd64(&rocData.failureDueTogcoff, 1)
			}
		}

		systemstack(c.publishG)
		if _g_ != nil {
			_g_.rocvalid = false // reset in startG
			_g_.rocgcnum = 42411 // reset in startG
		}
		return
	}
	// This is broken... and needs to be debugged.
	if c.rocgoid != _g_.goid {
		println("c.rocgoid=", c.rocgoid, "_g_.goid=", _g_.goid,
			"\ngetg().goid=", getg().goid, "getg().m.mcache.rocgoid=", getg().m.mcache.rocgoid, "c=", c,
			"\ngetg().m.curg.goid=", getg().m.curg.goid)
		throw("c.rocgoid != _g_.goid")
	}

	// Count of the number of bytes recovered using ROC
	recoveredBytes := int64(0)
	// Count of the number of bytes recovered by ROC that are returned from the mcache.
	heapLiveRecovered := int64(0)
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

			oldAllocCount := s.allocCount

			s.smashDebugHelper() // this increases the chance of triggering a bug

			s.freeindex = s.startindex // The actual recycle step.
			s.rollbackCount++
			s.rollbackAllocCount()

			recycled := oldAllocCount - s.allocCount
			recoveredBytes += int64(recycled) * int64(s.elemsize)

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
				// Not not count recovered bytes still remaining in the mcache.
				heapLiveRecovered += int64(recycled) * int64(s.elemsize)
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
		if s.isFree(0) {
			s.allocCount = 0
			s.nelems = 1
			recoveredBytes += int64(s.elemsize)
			heapLiveRecovered += int64(s.elemsize)
			atomic.Xadd64(&memstats.heap_live, -heapLiveRecovered) // releaseROCSpan adjusts this for non-largeAllocSpans
			mheap_.freeSpan(s, 1)
		} else {
			s.startindex = s.freeindex
			s.allocCount = s.nelems
		}
	}

	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.recoveredBytes, int64(recoveredBytes)) // TBD make available, perhaps as part of gctrace=2
	}

	c.largeAllocSpans = nil

	if _g_ != nil {
		_g_.rocvalid = true
	}
	// clean up tiny logic
	c.tiny = 0
	c.tinyoffset = 0
}

// publishStack scans a freshly create gp's stack, publishing all pointers
// found on the stack. While the g is fresh, the initial routine's arguments,
// potentially including pointers, are already on the stack.
// Since all of these pointers originated on the parent G
// and are being used by the offspring G they need to be published.
// Since this is newly created g there is only a single
// frame that needs to have its pointers published.
//
// The implementation of publishStack follows closely the implementation of
// scanstack so any change to scanstack is likely to require a change to publish
// stack.
//
// publishStack is marked go:systemstack because it must not be preempted
// while using a workbuf.
//
//go:nowritebarrier
//go:systemstack
func publishStack(gp *g) {
	if gp == getg() {
		throw("can't publish our own stack")
	}

	if gcphase != _GCoff {
		println("runtime: gcphase=", gcphase)
		throw("publishStack called during a GC")
	}

	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.stacksPublished, 1)
	}
	// Scan the stack.
	var cache pcvalueCache
	n := int64(0)

	// When creating a new goroutine only a single frame needs to be published.
	// When a stack is being preempted, say when it becomes dormant waiting
	// for new request we if we publish the entire stack ROC can
	// recycle the unreachable local objects.
	// This dormant goroutine optimzation is TBD.
	publishframe := func(frame *stkframe, unused unsafe.Pointer) bool {
		publishFrameWorker(frame, &cache)
		n++
		return true
	}

	gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, publishframe, nil, 0)
	tracebackdefers(gp, publishframe, nil)
	if debug.gcroc == 2 {
		atomic.Xadd64(&rocData.framesPublished, n)
		tmp := atomic.Load64(&rocData.maxFramesPublished)
		if tmp < uint64(n) {
			for !atomic.Cas64(&rocData.maxFramesPublished, tmp, uint64(n)) {
				tmp := atomic.Load64(&rocData.maxFramesPublished)
				if tmp >= uint64(n) {
					break
				}
			}
		}
		if n == 0 {
			throw("why do we have a stack with no frames.")
		} else if n == 1 {
			atomic.Xadd64(&rocData.oneFramesPublished, 1)
		} else if n <= 10 {
			atomic.Xadd64(&rocData.twoToTenFramesPublished, 1)
		} else if n <= 100 {
			atomic.Xadd64(&rocData.elevenToFiftyFramesPublished, 1)
		} else if n <= 100 {
			atomic.Xadd64(&rocData.fiftyOneToHundredFramesPublished, 1)
		} else {
			atomic.Xadd64(&rocData.overHundredFramesPublished, 1)
		}

	}
}

// Scan a stack frame: local variables and function arguments/results.
//
// publishStack is marked go:systemstack because it must not be preempted
// while using a workbuf.
//
//go:systemstack
//go:nowritebarrier
func publishFrameWorker(frame *stkframe, cache *pcvalueCache) {

	f := frame.fn
	targetpc := frame.continpc
	if targetpc == 0 {
		// Frame is dead.
		return
	}
	if _DebugGC > 1 {
		print("publishFrameWorker ", funcname(f), "\n")
	}
	if targetpc != f.entry {
		targetpc--
	}
	pcdata := pcdatavalue(f, _PCDATA_StackMapIndex, targetpc, cache)
	if pcdata == -1 {
		// We do not have a valid pcdata value but there might be a
		// stackmap for this function. It is likely that we are looking
		// at the function prologue, assume so and hope for the best.
		pcdata = 0
	}

	// Scan local variables if stack frame has been allocated.
	size := frame.varp - frame.sp
	var minsize uintptr
	switch sys.ArchFamily {
	case sys.ARM64:
		minsize = sys.SpAlign
	default:
		minsize = sys.MinFrameSize
	}
	if size > minsize {
		stkmap := (*stackmap)(funcdata(f, _FUNCDATA_LocalsPointerMaps))
		if stkmap == nil || stkmap.n <= 0 {
			print("runtime: frame ", funcname(f), " untyped locals ", hex(frame.varp-size), "+", hex(size), "\n")
			throw("missing stackmap")
		}

		// Locals bitmap information, scan just the pointers in locals.
		if pcdata < 0 || pcdata >= stkmap.n {
			// don't know where we are
			print("runtime: pcdata is ", pcdata, " and ", stkmap.n, " locals stack map entries for ", funcname(f), " (targetpc=", targetpc, ")\n")
			throw("publishframe: bad symbol table")
		}
		bv := stackmapdata(stkmap, pcdata)
		size = uintptr(bv.n) * sys.PtrSize
		publishStackBlock(frame.varp-size, size, bv.bytedata)
	}

	// Scan arguments.
	if frame.arglen > 0 {
		var bv bitvector
		if frame.argmap != nil {
			bv = *frame.argmap
		} else {
			stkmap := (*stackmap)(funcdata(f, _FUNCDATA_ArgsPointerMaps))
			if stkmap == nil || stkmap.n <= 0 {
				print("runtime: frame ", funcname(f), " untyped args ", hex(frame.argp), "+", hex(frame.arglen), "\n")
				throw("missing stackmap")
			}
			if pcdata < 0 || pcdata >= stkmap.n {
				// don't know where we are
				print("runtime: pcdata is ", pcdata, " and ", stkmap.n, " args stack map entries for ", funcname(f), " (targetpc=", targetpc, ")\n")
				throw("scanframe: bad symbol table")
			}
			bv = stackmapdata(stkmap, pcdata)
		}
		publishStackBlock(frame.argp, uintptr(bv.n)*sys.PtrSize, bv.bytedata)
	}
}

// publishStackBlock scans b as scanobject would, but using an explicit
// pointer bitmap instead of the heap bitmap.
//
//go:nowritebarrier
func publishStackBlock(b0, n0 uintptr, ptrmask *uint8) {
	// Use local copies of original parameters, so that a stack trace
	// due to one of the throws below shows the original block
	// base and extent.
	b := b0
	n := n0

	arena_start := mheap_.arena_start
	arena_used := mheap_.arena_used

	for i := uintptr(0); i < n; {
		// Find bits for the next word.
		bits := uint32(*addb(ptrmask, i/(sys.PtrSize*8)))
		if bits == 0 {
			i += sys.PtrSize * 8
			continue
		}
		for j := 0; j < 8 && i < n; j++ {
			if bits&1 != 0 {
				// Same work as in publishobject; see comments there.
				ptr := *(*uintptr)(unsafe.Pointer(b + i))
				if ptr != 0 && arena_start <= ptr && ptr < arena_used {
					// Get base of object ptr points to as well as the span.
					if obj, _, span, _ := heapBitsForObject(ptr, b, i); obj != 0 {
						if !isPublic(obj) {
							makePublic(obj, span)
							publish(obj)
						}
					}
				}
			}
			bits >>= 1
			i += sys.PtrSize
		}
	}
}

// end of publishStack routines

// Code below this point is for debugging and deserves only light review.

var afterFirstStartg = false

//go:nosplit
func CheckRocGoids() {
	//return immediately if we haven't done a startG
	if !afterFirstStartg {
		return
	}
	if !writeBarrier.roc {
		return
	}
	_g_ := getg()
	if _g_ == nil {
		return
	}
	if _g_.m == nil {
		return
	}
	gp := _g_.m.curg
	if gp == nil {
		return
	}
	c := gp.m.mcache
	if c == nil {
		return
	}
	if gp.goid != c.rocgoid {
		println("CheckRocGoids:mgcroc.go:55 getg().goid=", getg().goid, "getg().m.curg.goid=", getg().m.curg.goid,
			"\ngetg().m.mcache.rocgoid=", getg().m.mcache.rocgoid)
		throw("checkRocGoids fails")
	}
}

// smash will obliterate the contents of any free objects in the hopes that
// this will cause the program to abort quickly and debugging easier.
func (s *mspan) smashDebugHelper() {
	if writeBarrier.roc {
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
