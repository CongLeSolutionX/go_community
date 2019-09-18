// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Scavenging free pages.
//
// This file implements scavenging (the release of physical pages backing mapped
// memory) of free and unused pages in the heap as a way to deal with page-level
// fragmentation and reduce the RSS of Go applications.
//
// Scavenging in Go happens on two fronts: there's the background
// (asynchronous) scavenger and the heap-growth (synchronous) scavenger.
//
// The former happens on a goroutine much like the background sweeper which is
// soft-capped at using scavengePercent of the mutator's time, based on
// order-of-magnitude estimates of the costs of scavenging. The background
// scavenger's primary goal is to bring the estimated heap RSS of the
// application down to a goal.
//
// That goal is defined as:
//   (retainExtraPercent+100) / 100 * (next_gc / last_next_gc) * last_heap_inuse
//
// Essentially, we wish to have the application's RSS track the heap goal, but
// the heap goal is defined in terms of bytes of objects, rather than pages like
// RSS. As a result, we need to take into account for fragmentation internal to
// spans. next_gc / last_next_gc defines the ratio between the current heap goal
// and the last heap goal, which tells us by how much the heap is growing and
// shrinking. We estimate what the heap will grow to in terms of pages by taking
// this ratio and multiplying it by heap_inuse at the end of the last GC, which
// allows us to account for this additional fragmentation. Note that this
// procedure makes the assumption that the degree of fragmentation won't change
// dramatically over the next GC cycle. Overestimating the amount of
// fragmentation simply results in higher memory use, which will be accounted
// for by the next pacing up date. Underestimating the fragmentation however
// could lead to performance degradation. Handling this case is not within the
// scope of the scavenger. Situations where the amount of fragmentation balloons
// over the course of a single GC cycle should be considered pathologies,
// flagged as bugs, and fixed appropriately.
//
// An additional factor of retainExtraPercent is added as a buffer to help ensure
// that there's more unscavenged memory to allocate out of, since each allocation
// out of scavenged memory incurs a potentially expensive page fault.
//
// The goal is updated after each GC and the scavenger's pacing parameters
// (which live in mheap_) are updated to match. The pacing parameters work much
// like the background sweeping parameters. The parameters define a line whose
// horizontal axis is time and vertical axis is estimated heap RSS, and the
// scavenger attempts to stay below that line at all times.
//
// The synchronous heap-growth scavenging happens whenever the heap grows in
// size, for some definition of heap-growth. The intuition behind this is that
// the application had to grow the heap because existing fragments were
// not sufficiently large to satisfy a page-level memory allocation, so we
// scavenge those fragments eagerly to offset the growth in RSS that results.

package runtime

import (
	"math/bits"
	"runtime/internal/atomic"
	"unsafe"
)

const (
	// The background scavenger is paced according to these parameters.
	//
	// scavengePercent represents the portion of mutator time we're willing
	// to spend on scavenging in percent.
	scavengePercent = 1 // 1%

	// retainExtraPercent represents the amount of memory over the heap goal
	// that the scavenger should keep as a buffer space for the allocator.
	//
	// The purpose of maintaining this overhead is to have a greater pool of
	// unscavenged memory available for allocation (since using scavenged memory
	// incurs an additional cost), to account for heap fragmentation and
	// the ever-changing layout of the heap.
	retainExtraPercent = 10
)

// heapRetained returns an estimate of the current heap RSS.
//
// mheap_.lock must be held or the world must be stopped.
func heapRetained() uint64 {
	return atomic.Load64(&memstats.heap_sys) - atomic.Load64(&memstats.heap_released)
}

// gcPaceScavenger updates the scavenger's pacing, particularly
// its rate and RSS goal.
//
// The RSS goal is based on the current heap goal with a small overhead
// to accommodate non-determinism in the allocator.
//
// The pacing is based on scavengePageRate, which applies to both regular and
// huge pages. See that constant for more information.
//
// mheap_.lock must be held or the world must be stopped.
func gcPaceScavenger() {
	// If we're called before the first GC completed, disable scavenging.
	// We never scavenge before the 2nd GC cycle anyway (we don't have enough
	// information about the heap yet) so this is fine, and avoids a fault
	// or garbage data later.
	if memstats.last_next_gc == 0 {
		mheap_.scavengeBytesPerNS = 0
		return
	}
	// Compute our scavenging goal.
	goalRatio := float64(memstats.next_gc) / float64(memstats.last_next_gc)
	retainedGoal := uint64(float64(memstats.last_heap_inuse) * goalRatio)
	// Add retainExtraPercent overhead to retainedGoal. This calculation
	// looks strange but the purpose is to arrive at an integer division
	// (e.g. if retainExtraPercent = 12.5, then we get a divisor of 8)
	// that also avoids the overflow from a multiplication.
	retainedGoal += retainedGoal / (1.0 / (retainExtraPercent / 100.0))
	// Align it to a physical page boundary to make the following calculations
	// a bit more exact.
	retainedGoal = (retainedGoal + uint64(physPageSize) - 1) &^ (uint64(physPageSize) - 1)

	// Represents where we are now in the heap's contribution to RSS in bytes.
	//
	// Guaranteed to always be a multiple of physPageSize on systems where
	// physPageSize <= pageSize since we map heap_sys at a rate larger than
	// any physPageSize and released memory in multiples of the physPageSize.
	//
	// However, certain functions recategorize heap_sys as other stats (e.g.
	// stack_sys) and this happens in multiples of pageSize, so on systems
	// where physPageSize > pageSize the calculations below will not be exact.
	// Generally this is OK since we'll be off by at most one regular
	// physical page.
	retainedNow := heapRetained()

	// If we're already below our goal, publish the goal in case it changed
	// then disable the background scavenger.
	if retainedNow <= retainedGoal {
		mheap_.scavengeRetainedGoal = ^uint64(0)
		return
	}

	// Update all the pacing parameters in mheap with scavenge.lock held,
	// so that scavenge.gen is kept in sync with the updated values.
	mheap_.scavengeRetainedGoal = retainedGoal
	mheap_.scavengeGen++ // increase scavenge generation
	mheap_.pages.resetScavengeAddr()
}

// Sleep/wait state of the background scavenger.
var scavenge struct {
	lock   mutex
	g      *g
	parked bool
	timer  *timer

	// Generation counter.
	//
	// It represents the last generation count (as defined by
	// mheap_.scavengeGen) checked by the scavenger and is updated
	// each time the scavenger checks whether it is on-pace.
	//
	// Skew between this field and mheap_.scavengeGen is used to
	// determine whether a new update is available.
	//
	// Protected by mheap_.lock.
	gen uint64
}

// wakeScavenger unparks the scavenger if necessary. It must be called
// after any pacing update.
//
// mheap_.lock and scavenge.lock must not be held.
func wakeScavenger() {
	lock(&scavenge.lock)
	if scavenge.parked {
		// Try to stop the timer but we don't really care if we succeed.
		// It's possible that either a timer was never started, or that
		// we're racing with it.
		// In the case that we're racing with there's the low chance that
		// we experience a spurious wake-up of the scavenger, but that's
		// totally safe.
		stopTimer(scavenge.timer)

		// Unpark the goroutine and tell it that there may have been a pacing
		// change.
		scavenge.parked = false
		ready(scavenge.g, 0, true)
	}
	unlock(&scavenge.lock)
}

// scavengeSleep attempts to put the scavenger to sleep for ns.
//
// Note that this function should only be called by the scavenger.
//
// The scavenger may be woken up earlier by a pacing change, and it may not go
// to sleep at all if there's a pending pacing change.
//
// Returns false if awoken early (i.e. true means a complete sleep).
func scavengeSleep(ns int64) bool {
	lock(&scavenge.lock)

	// First check if there's a pending update.
	// If there is one, don't bother sleeping.
	var hasUpdate bool
	systemstack(func() {
		lock(&mheap_.lock)
		hasUpdate = mheap_.scavengeGen != scavenge.gen
		unlock(&mheap_.lock)
	})
	if hasUpdate {
		unlock(&scavenge.lock)
		return false
	}

	// Set the timer.
	//
	// This must happen here instead of inside gopark
	// because we can't close over any variables without
	// failing escape analysis.
	now := nanotime()
	scavenge.timer.when = now + ns
	startTimer(scavenge.timer)

	// Mark ourself as asleep and go to sleep.
	scavenge.parked = true
	goparkunlock(&scavenge.lock, waitReasonSleep, traceEvGoSleep, 2)

	// Return true if we completed the full sleep.
	return (nanotime() - now) >= ns
}

// Background scavenger.
//
// The background scavenger maintains the RSS of the application below
// the line described by the proportional scavenging statistics in
// the mheap struct.
func bgscavenge(c chan int) {
	scavenge.g = getg()

	lock(&scavenge.lock)
	scavenge.parked = true

	scavenge.timer = new(timer)
	scavenge.timer.f = func(_ interface{}, _ uintptr) {
		wakeScavenger()
	}

	c <- 1
	goparkunlock(&scavenge.lock, waitReasonGCScavengeWait, traceEvGoBlock, 1)

	// Exponentially-weighted moving average of the actual scavenge CPU use. It
	// represents a measure of scheduling overheads which might extend the sleep
	// or the critical time beyond what's expected. Assume no overhead to begin with.
	scavengeEWMA := float64(scavengePercent / 100.0)

	// Copied from CL 186924.
	var trace struct {
		nextTime int64   // earliest nanotime of next print
		released uintptr // memory released since last print
	}
	const printLimitNS = 30e9 // print at most every 30 seconds
	trace.nextTime = nanotime() + printLimitNS

	for {
		park := false
		crit := int64(0)

		// Run on the system stack since we grab the heap lock,
		// and a stack growth with the heap lock means a deadlock.
		systemstack(func() {
			lock(&mheap_.lock)

			// Update the last generation count that the scavenger has handled.
			scavenge.gen = mheap_.scavengeGen

			// If background scavenging is disabled or if there's no work to do just park.
			retained, goal := heapRetained(), mheap_.scavengeRetainedGoal
			if retained <= goal {
				unlock(&mheap_.lock)
				park = true
				return
			}
			unlock(&mheap_.lock)

			// Scavenge one page, and measure the amount of time spent scavenging.
			start := nanotime()
			released := mheap_.pages.scavengeone(physPageSize, false)
			crit = nanotime() - start

			// If we failed to release anything, we know we won't be able to make
			// any more progress until the next pacing update.
			if released == 0 {
				park = true
			} else {
				trace.released += released
			}
		})

		if park {
			lock(&scavenge.lock)
			scavenge.parked = true
			goparkunlock(&scavenge.lock, waitReasonGCScavengeWait, traceEvGoBlock, 1)
			continue
		}

		if debug.gctrace > 0 && nanotime() > trace.nextTime {
			if trace.released > 0 {
				print("scvg: ", trace.released>>10, " KB released\n")
			}
			print("scvg: inuse: ", memstats.heap_inuse>>20, ", idle: ", memstats.heap_idle>>20, ", sys: ", memstats.heap_sys>>20, ", released: ", memstats.heap_released>>20, ", consumed: ", (memstats.heap_sys-memstats.heap_released)>>20, " (MB)\n")

			trace.nextTime = nanotime() + printLimitNS
			trace.released = 0
		}

		// Compute the amount of time to sleep, assuming we want to use at most
		// scavengePercent of CPU time. Take into account scheduling overheads
		// that may extend the length of our sleep.
		sleepTime := int64((float64(crit) * scavengeEWMA) / (scavengePercent / 100.0))

		// Go to sleep, measuring how long we actually sleep for.
		start := nanotime()
		scavengeSleep(sleepTime)
		slept := nanotime() - start

		// Update scavengeEWMA by merging in the new crit/slept ratio.
		scavengeEWMA /= 2
		scavengeEWMA += float64(crit) / float64(slept) / 2
	}
}

// scavenge scavenges nbytes worth of free pages, starting with the
// highest address first. Successive calls continue from where it left
// off until the heap is exhausted. Call resetScavengeAddr to bring it
// back to the top of the heap.
//
// Returns the amount of memory scavenged in bytes.
//
// If locked == false, s.mheap must not be locked. If locked == true,
// s.mheap must be locked.
//
// Must run on the system stack because scavengeone must run on the
// system stack.
//
//go:systemstack
func (s *pageAlloc) scavenge(nbytes uintptr, locked bool) uintptr {
	released := uintptr(0)
	for released < nbytes {
		r := s.scavengeone(nbytes-released, locked)
		if r == 0 {
			// Nothing left to scavenge! Give up.
			break
		}
		released += r
	}
	return released
}

// resetScavengeAddr sets the scavenge hint to the top of the heap's
// address space. This should be called each time the scavenger's pacing
// changes.
//
// s.mheap.lock must be held.
func (s *pageAlloc) resetScavengeAddr() {
	s.scavAddr = uintptr(s.end+1)*heapArenaBytes - 1
}

// scavengeone starts from s.scavAddr and walks down the heap until it finds
// a contiguous run of pages to scavenge. It will try to scavenge at most
// max bytes at once, but will avoid breaking huge pages. Once it scavenges
// some memory it returns how much it scavenged and updates s.scavAddr
// appropriately. s.scavAddr must be reset manually and externally.
//
// Should it exhaust the heap, it will return 0 and set s.scavAddr to 0.
//
// If locked == false, s.mheap must not be locked.
// If locked == true, s.mheap must be locked.
//
// Must be run on the system stack because it either acquires the heap lock
// or executes with the heap lock acquired.
//
//go:systemstack
func (s *pageAlloc) scavengeone(max uintptr, locked bool) uintptr {
	maxPages := int(alignUp(max, pageSize) / pageSize)

	// Helpers for locking and unlocking only if locked == false.
	lockHeap := func() {
		if !locked {
			lock(&s.mheap.lock)
		}
	}
	unlockHeap := func() {
		if !locked {
			unlock(&s.mheap.lock)
		}
	}

	lockHeap()
	if s.scavAddr == 0 {
		// A zero hint means there are no more free and unscavenged pages. Quit.
		unlockHeap()
		return 0
	}

	// Check the arena containing the scav addr, starting at the addr
	// and see if there are any free and unscavenged pages.
	top := arenaIdx(s.scavAddr / heapArenaBytes)
	a := s.arenas(top)
	if a != nil {
		base, npages := a.pageAlloc.findScavengeCandidate(arenaPageIndex(s.scavAddr), maxPages)

		// If we found something, scavenge it and return!
		if npages != 0 {
			s.scavengeRangeLocked(top, a, base, npages)
			unlockHeap()
			return uintptr(npages) * pageSize
		}
	}
	unlockHeap()

	// Slow path: iterate optimistically looking for any free and unscavenged page.
	// If we think we see something, stop and verify it!
arenaLoop:
	for i := top - 1; i >= s.start; i-- {
		for j := mallocChunksPerArena - 1; j >= 0; j-- {
			// If this arena is totally in-use don't bother doing
			// a more sophisticated check.
			//
			// Note we're accessing this without a lock, but that's fine.
			// We're being optimistic anyway.
			if s.summary[len(s.summary)-1][chunkIndex(i, j)].max() == 0 {
				continue
			}

			// Look over the chunk for this arena and see if there are any
			// free and unscavenged pages.
			a := s.arenas(i)
			if a == nil || !a.pageAlloc.hasScavengeCandidate(j) {
				continue
			}

			// We found a candidate, so let's lock and verify it.
			lockHeap()

			// Find, verify, and scavenge if we can.
			base, npages := a.pageAlloc.findScavengeCandidate((j+1)*mallocChunkPages-1, maxPages)
			if npages == 0 {
				// We were fooled, let's take this opportunity to mark all the
				// memory up to here as scavenged for future calls and continue.
				// Note that findScavengeCandidate will attempt to search the rest
				// of the arena's chunks, so we can just start at the next arena.
				s.scavAddr = uintptr(i)*heapArenaBytes - 1
				unlockHeap()
				continue arenaLoop
			}
			s.scavengeRangeLocked(i, a, base, npages)
			unlockHeap()

			return uintptr(npages) * pageSize
		}
	}

	lockHeap()
	// We couldn't find anything, so signal that there's nothing left
	// to scavenge if the hint hasn't been updated at all.
	s.scavAddr = 0
	unlockHeap()

	return 0
}

// scavengeRangeLocked scavenges the given region of memory.
//
// s.mheap must be locked.
func (s *pageAlloc) scavengeRangeLocked(i arenaIdx, a *heapArena, base, npages int) {
	a.pageAlloc.scavengeRange(base, npages)

	// Update the scav pointer.
	s.scavAddr = uintptr(i)*heapArenaBytes + uintptr(base)*pageSize

	if !s.test {
		start := arenaBase(i) + uintptr(base)*pageSize

		// Only perform the actual scavenging if we're not in a test.
		// It's dangerous to do so otherwise.
		sysUnused(unsafe.Pointer(start), uintptr(npages)*pageSize)

		// Update global accounting only when not in test, otherwise
		// the runtime's accounting will be wrong.
		mSysStatInc(&memstats.heap_released, uintptr(npages)*pageSize)
	}
}

// hasScavengeCandidate returns if there's any free-and-unscavenged memory
// in the region represented by this mallocData.
//
// chunk indicates the chunk to search.
func (m *mallocData) hasScavengeCandidate(chunk int) bool {
	// The goal of this search is to see if the arena contains any free and unscavenged memory.
	for i := (chunk+1)*mallocChunkPages/64 - 1; i >= chunk*mallocChunkPages/64; i-- {
		// 1s are scavenged OR non-free => 0s are unscavenged AND free
		x := m.scavenged[i] | m.mallocBits[i]

		// Quickly skip over chunks of non-free or scavenged pages.
		if x == ^uint64(0) {
			continue
		}
		return true
	}
	return false
}

// findScavengeCandidate returns a start index and a size for this mallocData
// segment which represents a contiguous region of free and unscavenged memory.
//
// hint indicates a point at which to start the search, but note that
// findScavengeCandidate searches backwards through the mallocData. That is, it
// will return the highest scavenge candidate.
//
// max is a hint for how big of a region is desired. If max >= pagesPerArena, then
// findScavengeCandidate effectively returns entire free and unscavenged regions.
// If max < pagesPerArena, it may truncate the returned region such that size is
// max. However, findScavengeCandidate may still return a larger region if, for
// example, it chooses to preserve huge pages. That is, even if max is small,
// size is not guaranteed to be equal to max.
//
// TODO(mknyszek): This function is... a bit hard to follow. It should be refactored.
func (m *mallocData) findScavengeCandidate(hint, max int) (start int, size int) {
	splitFreeRegion := func(start, size int) (int, int) {
		bottom := start
		start = start + size - max
		size = max
		// If we don't have huge pages, just return the split down to max.
		if physHugePageSize <= pageSize {
			return start, size
		}
		// Compute the huge page boundary above our candidate.
		pagesPerHugePage := uintptr(physHugePageSize / pageSize)
		hugePageAbove := int(alignUp(uintptr(start), pagesPerHugePage))
		// If that boundary is within our current candidate, then we may be breaking
		// a huge page.
		if hugePageAbove <= start+size {
			hugePageBelow := int(alignDown(uintptr(start), pagesPerHugePage))
			// If start is on a 64 page boundary there could still be more to
			// find. In particular, we might find a huge page. Iterate a little
			// further, at most up to the next huge page boundary.
			if bottom%64 == 0 {
				for i := bottom/64 - 1; i >= 0 && bottom > hugePageBelow; i-- {
					// 1s are scavenged OR non-free => 0s are unscavenged AND free
					x := m.scavenged[i] | m.mallocBits[i]
					// Count leading 1s.
					z1 := bits.LeadingZeros64(^x)
					if z1 != 0 {
						// If there are any leading 1s we can't make any progress.
						break
					}
					// There are no leading 1s, that means there must be leading
					// zeros. No matter what, we want to include them.
					z0 := bits.LeadingZeros64(x)
					bottom -= z0
					// If, however, these zeros don't reach the bottom of the this
					// 64 bit chunk, then that means that we are blocked from
					// continuing.
					if z0 != 64 {
						break
					}
				}
			}
			if hugePageBelow >= bottom {
				// We're in danger of breaking apart a huge page since start+size crosses
				// a huge page boundary and rounding down start to the nearest huge
				// page boundary is valid. Include the entire huge page in the bound by
				// rounding down to the huge page size.
				size = max + start - hugePageBelow
				start = hugePageBelow
			}
		}
		return start, size
	}

	// The goal of this search is to find the last contiguous run
	// of free and unscavenged pages which is at most max pages in size.
	// If we find a larger one, we just take the top max pages from it.
	for i := hint / 64; i >= 0; i-- {
		// 1s are scavenged OR non-free => 0s are unscavenged AND free
		x := m.scavenged[i] | m.mallocBits[i]

		// Quickly skip over blocks of non-free or scavenged pages.
		if x == ^uint64(0) {
			start = 0
			size = 0
			continue
		}

		// z1 counts the number of leading 1s.
		z1 := bits.LeadingZeros64(^x)

		// z0 counts the number of zeroes after that.
		// z0 may be 64 even if z1 != 0 because when we
		// shift we might shift out all the 1s.
		z0 := bits.LeadingZeros64(x << z1)

		if z0 != 64 {
			// Does not reach bottom edge.
			// This could mean one of two things:
			// 1) The top z0 < 64 bits of x are zero, e.g. 1101...111000.
			//    In this case z1 == 0.
			// 2) The top z1 bits of x are ones, and the next z0 bits
			//    are zero, e.g. 1101...111001. In this case z1 != 0.
			// In both of these cases, there could be more pockets of
			// zeroes, but that doesn't matter. We just want to get the
			// first pocket. The same math applies in both cases.
			start = i*64 + 64 - z0 - z1
			size += z0
			if size > max {
				start, size = splitFreeRegion(start, size)
			}
			return start, size
		}
		// Reaches bottom edge.
		// Again there are two cases here:
		// 1) x == 0 and z1 == 0.
		// 2) The bottom 64-z1 bits are zero and z1 != 0.
		start = i * 64
		if z1 == 0 {
			size += 64
		} else {
			size = 64 - z1
		}
		if size >= max {
			return splitFreeRegion(start, size)
		}
	}
	return start, size
}

// scavengeRange unconditionally marks the scavenge bits for
// the given bit range in the mallocData.
func (m *mallocData) scavengeRange(base, npages int) {
	m.scavenged.setRange(base, npages)
}
