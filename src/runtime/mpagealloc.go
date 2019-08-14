// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page allocator.
//
// The page allocator manages mapped pages (defined by pageSize, NOT
// physPageSize) for allocation and re-use and is embedded into mheap.
//
// Pages are managed using a bitmap that is sharded into arenas, with
// each shard stored alongside other arena metadata. In the bitmap, 1
// means in-use, and 0 means free. The bitmap spans the process's
// address space.
//
// The bitmap is efficiently searched by using a tiered summary
// structure. The summaries describe three properties of a particular
// section of the bitmap: the number of contiguous 0 bits at
// the start and end of the bitmap section, and the maximum number of
// contiguous zero bits found anywhere in that section.
//
// The lowest tier has each summary represent the largest section of
// address space (order of TiB on 64-bit systems), with each higher
// tier representing successively smaller tiers representing a
// subsection until we reach the highest granularity in the highest
// tier: an arena.
//
// These tiers effectively form a radix tree. If an entry in the lowest
// tier indicates free space, then one can search that section more
// specifically by looking at entries in the next tier.

package runtime

import (
	"unsafe"
)

const (
	logPagesPerArena = logHeapArenaBytes - pageShift

	// The maximum amount of arenas that could exist in this address space.
	logMaxArenas = heapAddrBits - logHeapArenaBytes
	maxArenas    = 1 << logMaxArenas

	// The amount of bits (i.e. pages) to consider in the bitmap at once. It may be that each
	// arena's bitmap shard is larger than this, but we divide it into equal sections
	// of allocBitmapChunk to avoid spending too much time running over bitmaps.
	logMallocChunkPages  = 9
	logMallocChunkBytes  = logMallocChunkPages + pageShift
	mallocChunkPages     = 1 << logMallocChunkPages
	mallocChunkBytes     = 1 << logMallocChunkBytes
	mallocChunksPerArena = pagesPerArena / mallocChunkPages

	// The number of tiers in the summary structure.
	summaryLevels = 5

	// The number of radix bits for each level.
	// summaryL0Bits + (summaryLevels-1)*summaryLevelBits + logMallocChunkBytes = heapAddrBits
	summaryLevelBits = 3
	summaryL0Bits    = heapAddrBits - logMallocChunkBytes - (summaryLevels-1)*summaryLevelBits

	// log2 of the total number of entries in each level.
	logSummaryL4Size = heapAddrBits - logMallocChunkBytes
	logSummaryL3Size = heapAddrBits - logMallocChunkBytes - 1*summaryLevelBits
	logSummaryL2Size = heapAddrBits - logMallocChunkBytes - 2*summaryLevelBits
	logSummaryL1Size = heapAddrBits - logMallocChunkBytes - 3*summaryLevelBits
	logSummaryL0Size = heapAddrBits - logMallocChunkBytes - 4*summaryLevelBits // == summaryL0Bits

	// Maximum hint value, which indicates that the heap has no free space.
	maxHint = ^uintptr(0) - arenaBaseOffset
)

// TODO(mknyszek): Explain this.
var levelBits = [summaryLevels]uint{
	summaryL0Bits,
	summaryLevelBits,
	summaryLevelBits,
	summaryLevelBits,
	summaryLevelBits,
}

// TODO(mknyszek): Explain this.
var levelShift = [summaryLevels]uint{
	heapAddrBits - logSummaryL0Size,
	heapAddrBits - logSummaryL1Size,
	heapAddrBits - logSummaryL2Size,
	heapAddrBits - logSummaryL3Size,
	heapAddrBits - logSummaryL4Size,
}

// TODO(mknyszek): Explain this.
var levelLogPages = [summaryLevels]uint{
	logMallocChunkPages + 4*summaryLevelBits,
	logMallocChunkPages + 3*summaryLevelBits,
	logMallocChunkPages + 2*summaryLevelBits,
	logMallocChunkPages + 1*summaryLevelBits,
	logMallocChunkPages,
}

// Global chunk index.
type chunkIdx int

func chunkIndex(p uintptr) chunkIdx {
	return chunkIdx((p + arenaBaseOffset) / mallocChunkBytes)
}

func chunkBase(ci chunkIdx) uintptr {
	return uintptr(ci)*mallocChunkBytes - arenaBaseOffset
}

// arenaPageIndex computes the index of the page which contains p,
// relative to the arena which contains p.
func arenaPageIndex(p uintptr) int {
	return int(p % heapArenaBytes / pageSize)
}

// arenaChunkIndex computes the index of the chunk which contains p,
// relative to the arena which contains p.
func arenaChunkIndex(p uintptr) int {
	return int(p % heapArenaBytes / mallocChunkBytes)
}

// chunkPageIndex computes the index of the page which contains p,
// relative to the chunk which contains p.
func chunkPageIndex(p uintptr) int {
	return int(p % mallocChunkBytes / pageSize)
}

// arenaToChunkIndex returns a global chunk index given an arena index and a
// chunk index relative to that arena.
func arenaToChunkIndex(ai arenaIdx, offset int) chunkIdx {
	return chunkIdx(int(ai)*mallocChunksPerArena + offset)
}

// addrsToSummaryRange converts base and limit pointers into a range
// of entries for the given summary level. base must be the inclusive
// lower bound and limit must be the inclusive upper bound.
func addrsToSummaryRange(level int, base, limit uintptr) (lo int, hi int) {
	// Pull out necessary level-related "constants".
	e := 1 << levelBits[level] // number of entries in an entry block for this level
	sh := levelShift[level]    // radix position (shift count) for this level

	// lo and hi and are the base and the limit rounded down and up
	// to the closest set of entries for this level, respectively.
	//
	// Note that limit represents the inclusive upper bound, so in
	// order to get the rounding right, we need to make sure that
	// after the shift into an index, we increment by 1.
	lo = int(base+arenaBaseOffset) >> sh
	hi = (int(limit+arenaBaseOffset) >> sh) + 1

	// Round down/up to nearest multiple of entries in this level.
	lo = lo &^ (e - 1)
	hi = (hi + e - 1) &^ (e - 1)
	return
}

//go:notinheap
type pageAlloc struct {
	// Summary and super-summary structure.
	//
	// Each slice's cap represents the whole memory reservation,
	// whereas the len the an upper-bound on the available heap.
	summary [summaryLevels][]mallocSum

	// The hint address to start an allocation search with.
	hint uintptr

	// start and end represent the chunk indices
	// which pageAlloc knows about. It assumes
	// chunks in the range [start, end) are
	// currently ready to use.
	start, end chunkIdx

	// Reference to an mheap, used for testing by using
	// a dummy mheap structure.
	mheap *mheap
}

func (s *pageAlloc) init(h *mheap) {
	// Reserve memory for each level. This will get mapped in
	// as R/W by setArenas.
	for l := 0; l < summaryLevels; l++ {
		size := 1 << (heapAddrBits - levelShift[l])

		// Reserve b bytes of memory anywhere in the address space.
		b := alignUp(uintptr(size)*mallocSumBytes, physPageSize)
		r := sysReserve(nil, b)
		if r == nil {
			throw("failed to reserve summary structure space")
		}

		// Put this reservation into a slice.
		sl := notInHeapSlice{(*notInHeap)(r), 0, size}
		s.summary[l] = *(*[]mallocSum)(unsafe.Pointer(&sl))
	}

	// Start with the hint in a state indicating there's no free memory.
	s.hint = maxHint

	// Save a reference to mheap. This extra level of indirection is
	// critical for testing.
	s.mheap = h
}

// arenas returns the heap arena associated with the given arena index.
// It performs this through the mheap pointer in s. This acts as a layer
// of indirection for testing.
func (s *pageAlloc) arenas(ai arenaIdx) *heapArena {
	return s.mheap.arenas[ai.l1()][ai.l2()]
}

// grow sets up the metadata for the address range [base, base+size).
// It may cause an increase in metadata allocation, and sysStat will be
// updated appropriately to reflect this increase.
func (s *pageAlloc) grow(base, size uintptr, sysStat *uint64) {
	// Round up to chunks, since we can't deal with increments smaller
	// than chunks.
	base = alignDown(base, mallocChunkBytes)
	size = alignUp(size, mallocChunkBytes)
	limit := base + size

	// Walk up the radix tree and map summaries in as needed.
	cbase, climit := chunkBase(s.start), chunkBase(s.end)
	for l := len(s.summary) - 1; l >= 0; l-- {
		// Figure out what part of the summary array is already mapped.
		mlo, mhi := addrsToSummaryRange(l, cbase, climit-1)
		mappedBase := alignDown(uintptr(mlo)*mallocSumBytes, physPageSize)
		mappedLimit := alignUp(uintptr(mhi)*mallocSumBytes, physPageSize)

		// Figure out what part of the summary array this new address space needs.
		lo, hi := addrsToSummaryRange(l, base, limit-1)
		needBase := alignDown(uintptr(lo)*mallocSumBytes, physPageSize)
		needLimit := alignUp(uintptr(hi)*mallocSumBytes, physPageSize)

		// Update the summary slices with a new upper-bound. This ensures
		// we get tight bounds checks on at least the top bound.
		//
		// We must do this regardless of whether we map new memory, because we
		// may be extending further into the mapped memory.
		if hi > len(s.summary[l]) {
			s.summary[l] = s.summary[l][:hi]
		}

		if s.start != 0 {
			if needBase < mappedBase && needLimit > mappedLimit {
				throw("re-initialization of chunk")
			} else if needBase < mappedBase && needLimit <= mappedLimit {
				needLimit = mappedBase
			} else if needBase >= mappedBase && needLimit > mappedLimit {
				needBase = mappedLimit
			} else {
				// The new arenas are completely contained in an already-mapped
				// region for the summaries for this level, so keep walking up.
				continue
			}
		}

		// Commit memory for the relevant summaries at level l.
		// Note that this memory must be physical page aligned.
		sbase := unsafe.Pointer(uintptr(unsafe.Pointer(&s.summary[l][0])) + needBase)
		sysMap(sbase, needLimit-needBase, sysStat)
		sysUsed(sbase, needLimit-needBase)
	}

	// Update s.start and s.end.
	// If no arenas set yet, start == 0. This is generally
	// safe since the zero page is unmapped, and arenas are
	// naturally aligned.
	start, end := chunkIndex(base), chunkIndex(limit)
	if s.start == 0 || start < s.start {
		s.start = start
	}
	if end > s.end {
		s.end = end
	}

	// Update the hint since the arena must have free pages in it (except
	// during tests, but it's always OK if the hint is too low).
	// Compare with arenaBaseOffset added because this gives us a more
	// consistent view of the heap.
	if base+arenaBaseOffset < s.hint+arenaBaseOffset {
		s.hint = base
	}

	// Update summaries accordingly. This operation acts kind of
	// like a free, since it suddenly adds a lot more free memory
	// to the structure.
	s.update(base, size/pageSize, true, false)
}

// update updates heap metadata. It must be called each time the bitmap
// is updated.
//
// If contig is true, update does some optimizations assuming that there was
// a contiguous allocation or free between addr and addr+npages. alloc indicates
// whether the operation performed was an allocation or a free.
func (s *pageAlloc) update(base, npages uintptr, contig, alloc bool) {
	// base, limit, start, and end are inclusive.
	limit := base + npages*pageSize - 1
	sa, ea := arenaIndex(base), arenaIndex(limit)
	sc, ec := arenaChunkIndex(base), arenaChunkIndex(limit)

	// Handle updating the lowest level first.
	if sa == ea && sc == ec {
		// Fast path: the allocation doesn't span more than one arena,
		// so update this one and if the summary didn't change, return.
		x := s.summary[len(s.summary)-1][arenaToChunkIndex(sa, sc)]
		y := s.arenas(sa).pageAlloc.summarize(sc)
		if x == y {
			return
		}
		s.summary[len(s.summary)-1][arenaToChunkIndex(sa, sc)] = y
	} else if contig {
		// Slow contiguous path: the allocation spans more than one chunk
		// and/or arena and at least one summary is guaranteed to change.
		summary := s.summary[len(s.summary)-1]

		// Update the summary for chunk (sa, sc).
		summary[arenaToChunkIndex(sa, sc)] = s.arenas(sa).pageAlloc.summarize(sc)

		// Update the summaries for chunks in between, which are
		// either totally allocated or freed.
		whole := s.summary[len(s.summary)-1][arenaToChunkIndex(sa, sc)+1 : arenaToChunkIndex(ea, ec)]
		if alloc {
			// Should optimize into a memclr.
			for i := range whole {
				whole[i] = 0
			}
		} else {
			for i := range whole {
				whole[i] = freeChunkSum
			}
		}

		// Update the summary for chunk (ea, ec).
		summary[arenaToChunkIndex(ea, ec)] = s.arenas(ea).pageAlloc.summarize(ec)
	} else {
		// Slow general path: the allocation spans more than one chunk
		// and/or arena and at least one summary is guaranteed to change.
		//
		// We can't assume a contiguous allocation happened, so walk over
		// every chunk in the range and manually recompute the summary.
		summary := s.summary[len(s.summary)-1]
		if sa == ea {
			// The range is within an arena.
			a := s.arenas(sa)
			for j := sc; j <= ec; j++ {
				summary[arenaToChunkIndex(sa, j)] = a.pageAlloc.summarize(j)
			}
		} else {
			// The range spans multiple arenas.
			for j := sc; j < mallocChunksPerArena; j++ {
				summary[arenaToChunkIndex(sa, j)] = s.arenas(sa).pageAlloc.summarize(j)
			}
			for i := sa + 1; i < ea; i++ {
				a := s.arenas(i)
				for j := 0; j < mallocChunksPerArena; j++ {
					summary[arenaToChunkIndex(i, j)] = a.pageAlloc.summarize(j)
				}
			}
			for j := 0; j <= ec; j++ {
				summary[arenaToChunkIndex(ea, j)] = s.arenas(ea).pageAlloc.summarize(j)
			}
		}
	}

	// Walk up the radix tree and update the summaries appropriately.
	changed := true
	for l := len(s.summary) - 1; l >= 1 && changed; l-- {
		changed = false
		b := levelBits[l]     // log entries in this level
		e := 1 << b           // entries in this level
		p := levelLogPages[l] // pages per entry in this level

		// level is all the parts of the level we need to look at and
		// modify. This may include multiple blocks of entries at a
		// given level.
		lo, hi := addrsToSummaryRange(l, base, limit)
		level := s.summary[l][lo:hi]
		for i := 0; i < len(level); i += e {
			// Compute summary for each set of entries and
			// propagate to the next level.
			start, max, end := level[i].unpack()
			for j := i + 1; j < i+e; j++ {
				si, mi, ei := level[j].unpack()

				// Compute start by checking if we've collected
				// exactly 2^p pages for each entry in this block so far.
				// This indicates that we have a run of free pages all the
				// way to the beginning of this block of entries.
				if start == (j-i)<<p {
					start += si
				}

				// Compute max.
				if end+si > max {
					max = end + si
				}
				if mi > max {
					max = mi
				}

				// Compute end by checking if this entry has free pages
				// connecting all the way through (the entire entry is
				// free) in which case we keep a running count. Otherwise
				// it's just going to be what ei was.
				if ei == 1<<p {
					end += 1 << p
				} else {
					end = ei
				}
			}
			u := (lo + i) >> b
			old, sum := s.summary[l-1][u], packMallocSum(start, max, end)
			if old != sum {
				changed = true
				s.summary[l-1][u] = sum
			}
		}
	}
}

// allocRange marks the range of memory [addr, addr+npages*pageSize) as
// allocated.
func (s *pageAlloc) allocRange(addr, npages uintptr) {
	limit := addr + npages*pageSize - 1
	si, ei := arenaIndex(addr), arenaIndex(limit)
	sp, ep := arenaPageIndex(addr), arenaPageIndex(limit)

	if si == ei {
		// The range doesn't cross any arena boundaries.
		s.arenas(si).pageAlloc.allocRange(sp, ep+1-sp)
		return
	}

	// The range crosses at least one arena boundary.
	s.arenas(si).pageAlloc.allocRange(sp, pagesPerArena-sp)
	for i := si + 1; i < ei; i++ {
		s.arenas(i).pageAlloc.allocAll()
	}
	s.arenas(ei).pageAlloc.allocRange(0, ep+1)
}

func (s *pageAlloc) allocSlow(npages uintptr) (uintptr, uintptr) {
	// Search algorithm
	//
	// This algorithm walks each level l of the radix tree from top to bottom.
	// It iterates over at most 1 << levelBits[l] of entries in a given level
	// in the radix tree, and uses the summary information to find either:
	//  1) That the next level contains a large enough contiguous region, at
	//     which point it continues on the next level, or
	//  2) That there are enough contiguous boundary-crossing bits to satisfy
	//     the allocation, at which point it knows exactly where to start
	//     allocating from and calls allocRange.
	//
	// i tracks the index into the current level l's structure for the
	// contiguous 1 << levelBits[l] entries we're actually interested in.
	//
	// NOTE: Technically this search could allocate a region which crosses
	// the arenaBaseOffset boundary, which when arenaBaseOffset != 0, is
	// a discontinuity. However, the only way this could happen is if the
	// page at the zero address is mapped, and this is impossible on
	// virtually every system we support. So, the discontinuity is already
	// encoded in the fact that the OS will never map the zero page for us,
	// and this function doesn't try to handle this case in any way.
	i := 0
	// hint is the best hint we could glean from our search.
	// Zero indicates we have no hint yet.
	hint := uintptr(0)
nextLevel:
	for l := 0; l < len(s.summary); l++ {
		b := levelBits[l]
		e := 1 << b
		i <<= b
		level := s.summary[l][i : i+e]
		p := levelLogPages[l]
		// Determine j0, the first index we should start iterating from.
		// The hint may help us eliminate iterations if we followed the
		// hint on the previous level, in which case the top bits of the
		// hint address should be the same as i, after levelShift.
		j0 := 0
		if hintIdx := int((s.hint + arenaBaseOffset) >> levelShift[l]); hintIdx&^(e-1) == i {
			j0 = hintIdx & (e - 1)
		}
		// Run over the level entries looking for
		// a contiguous run of at least npages either
		// within an entry or across entries.
		//
		// start contains the page index (relative to
		// the first entry's first page) of the currently
		// considered run of consecutive pages.
		//
		// size contains the size of the currently considered
		// run of consecutive pages.
		start, size := 0, 0
		for j := j0; j < len(level); j++ {
			sum := level[j]
			if sum == 0 {
				// A full entry means we broke any streak and
				// that we should skip it altogether.
				size = 0
				continue
			}
			s := sum.start()
			if size != 0 && size+s >= int(npages) {
				// We hit npages; we're done!
				size += s
				break
			}
			if sum.max() >= int(npages) {
				// The entry itself contains npages contiguous
				// free pages, so drop down into the next level.
				i += j
				continue nextLevel
			}
			if hint == 0 {
				// When we've reached this point and haven't found
				// the latest non-zero summary, that means this is the
				// latest non-zero summary that we can possibly find
				// during this search. At this point we may look further
				// in this level, but we won't be able to gain any new
				// new information even if we drop down again.
				hint = uintptr((i+j)<<levelShift[l]) - arenaBaseOffset
			}
			if size == 0 || s < 1<<p {
				// We either don't have a current run started, or this entry
				// isn't totally free (meaning we can't continue the current
				// one), so try to begin a new run by setting size and start
				// based on sum.end.
				size = sum.end()
				start = (j+1)<<p - size
				continue
			}
			// The entry is completely free, so continue the run.
			size += 1 << p
		}
		if size >= int(npages) {
			// We found some range which crosses boundaries, just go and mark it
			// directly.
			addr := uintptr(i<<levelShift[l]) - arenaBaseOffset + uintptr(start)*pageSize
			s.allocRange(addr, npages)
			// If at this point we still haven't found the first free open space,
			// this is it.
			if hint == 0 {
				hint = addr + npages*pageSize
			}
			return addr, hint
		}
		if l != 0 {
			throw("bad summary data")
		}
		// We're at level zero, so that means we've exhausted our search.
		return 0, ^uintptr(0)
	}
	// Since we've gotten to this point, that means we haven't found a
	// sufficiently-sized free region straddling some boundary (arena or larger),
	// so just allocate directly out of the arena without a hint.
	//
	// After iterating over all levels, i must contain an arena index which
	// is what the final level represents.
	ai := arenaIdx(i / mallocChunksPerArena)
	j, h := s.arenas(ai).pageAlloc.alloc(npages, (i%mallocChunksPerArena)*mallocChunkPages)
	if j < 0 {
		throw("bad summary data")
	}
	addr := arenaBase(ai) + uintptr(j)*pageSize
	// If at this point we still haven't found the first free space, that
	// means this is it!
	if hint == 0 {
		hint = arenaBase(ai) + uintptr(h)*pageSize
	}
	return addr, hint
}

func (s *pageAlloc) alloc(npages uintptr) uintptr {
	// If the hint refers to a region which has a higher address than
	// any known arena, then we know we're out of memory.
	if chunkIndex(s.hint) >= s.end {
		return 0
	}

	// If npages has a chance of fitting in the chunk where the hint is,
	// try to allocate from it directly.
	var addr, hint uintptr
	if mallocChunkPages-chunkPageIndex(s.hint) >= int(npages) {
		// npages is guaranteed to be no greater than pagesPerArena here.
		ci := chunkIndex(s.hint)
		if s.summary[len(s.summary)-1][ci].max() >= int(npages) {
			i := arenaIndex(s.hint)
			j, h := s.arenas(i).pageAlloc.alloc(npages, arenaPageIndex(s.hint))
			if j < 0 {
				throw("bad summary data")
			}
			addr = arenaBase(i) + uintptr(j)*pageSize
			hint = arenaBase(i) + uintptr(h)*pageSize
			goto done
		}
	}
	// We failed to use a hint for one reason or another, so try
	// the slow path.
	addr, hint = s.allocSlow(npages)
	if addr == 0 {
		if npages == 1 {
			// We failed to find a single free page, the smallest unit
			// of allocation. This means we know the heap is completely
			// exhausted. Otherwise, the heap still might have free
			// space in it, just not enough contiguous space to
			// accommodate npages.
			s.hint = maxHint
		}
		return 0
	}
done:
	// If we found a better hint, update our hint.
	if hint+arenaBaseOffset > s.hint+arenaBaseOffset {
		s.hint = hint
	}
	// We have a non-zero address, so update and return.
	s.update(addr, npages, true, true)
	return addr
}

func (s *pageAlloc) freeSlow(base, npages uintptr) {
	limit := base + npages*pageSize - 1
	baseOffset := arenaPageIndex(base)
	start, end := arenaIndex(base), arenaIndex(limit)
	if start == end {
		// Contained within one arena, so just update bits in this arena and we're done.
		s.arenas(start).pageAlloc.free(baseOffset, int(npages))
		return
	}
	_ = s.arenas(start)
	_ = s.arenas(end)
	// base + npages*pageSize spans more than one arena, so go over all of them.
	s.arenas(start).pageAlloc.free(baseOffset, pagesPerArena-baseOffset)
	for i := start + 1; i < end; i++ {
		s.arenas(i).pageAlloc.freeAll()
	}
	limitOffset := arenaPageIndex(limit)
	s.arenas(end).pageAlloc.free(0, limitOffset+1 /* add 1 since this is a length */)
}

func (s *pageAlloc) free(base, npages uintptr) {
	// Because of split address spaces, we must compare against the hint address
	// with both sides containing arenaBaseOffset, because this gives us a linear
	// view of the address space and ensures that we can do a simple comparison.
	if base+arenaBaseOffset < s.hint+arenaBaseOffset {
		// We need to convert a pointer into an offset-pointer.
		s.hint = base
	}
	if npages == 1 {
		// Fast path: we're clearing a single bit, and we know exactly
		// where it is, so mark it directly.
		s.arenas(arenaIndex(base)).pageAlloc.free1(arenaPageIndex(base))
	} else {
		// Slow path: we're clearing more bits so we may need to iterate.
		s.freeSlow(base, npages)
	}
	s.update(base, npages, true, false)
}

const (
	mallocSumBytes    = 8
	logMaxPackedValue = logMallocChunkPages + (summaryLevels-1)*summaryLevelBits
	maxPackedValue    = 1 << logMaxPackedValue

	freeChunkSum = mallocSum(uint64(mallocChunkPages) |
		uint64(mallocChunkPages<<logMaxPackedValue) |
		uint64(mallocChunkPages<<(2*logMaxPackedValue)))
)

// mallocSum is a packed summary type which packs three numbers: start, max,
// and end into a single 8-byte value. Each of these values are a summary of
// a bitmap and are thus counts, each of which may have a maximum value of
// 2^21. This value doesn't fit into 21 bits, but we exploit the fact that
// start == max && max == end if any of them have their maximum value, so we
// use the 64th bit to represent this case.
type mallocSum uint64

// packMallocSum takes a start, max, and end value and produces a mallocSum.
func packMallocSum(start, max, end int) mallocSum {
	if max == maxPackedValue {
		return mallocSum(uint64(1 << 63))
	}
	return mallocSum((uint64(start) & (maxPackedValue - 1)) |
		((uint64(max) & (maxPackedValue - 1)) << logMaxPackedValue) |
		((uint64(end) & (maxPackedValue - 1)) << (2 * logMaxPackedValue)))
}

// start extracts the start value from a packed sum.
func (p mallocSum) start() int {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return int(uint64(p) & (maxPackedValue - 1))
}

// max extracts the max value from a packed sum.
func (p mallocSum) max() int {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return int((uint64(p) >> logMaxPackedValue) & (maxPackedValue - 1))
}

// end extracts the end value from a packed sum.
func (p mallocSum) end() int {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return int((uint64(p) >> (2 * logMaxPackedValue)) & (maxPackedValue - 1))
}

// unpack unpacks all three values from the summary.
func (p mallocSum) unpack() (int, int, int) {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue, maxPackedValue, maxPackedValue
	}
	return int(uint64(p) & (maxPackedValue - 1)),
		int((uint64(p) >> logMaxPackedValue) & (maxPackedValue - 1)),
		int((uint64(p) >> (2 * logMaxPackedValue)) & (maxPackedValue - 1))
}
