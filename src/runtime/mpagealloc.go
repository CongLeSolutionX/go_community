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
// The bitmap is efficiently searched by using a radix tree in combination
// with fast bit-wise intrinsics. Allocation is performed using an address-ordered
// first-fit approach.
//
// Each entry in the radix tree is a summary which describes three properties of
// a particular region of the address space: the number of contiguous free pages
// at the start and end of the region it represents, and the maximum number of
// contiguous free pages found anywhere in that region. Note that free pages map
// directly to 0s in the bitmap.
//
// Each level of the radix tree is stored as one contiguous array, which represents
// a different granularity of subdivision of the processes' address space. Thus, this
// radix tree is actually implicit in these large arrays, as opposed to having explicit
// dynamically-allocated pointer-based node structures. Naturally, these arrays may be
// quite large for system with large address spaces, so in these cases they are mapped
// into memory as needed.
//
// The leaves of the radix tree are summaries which represent the smallest unit which
// the tree understands, which we call a "malloc chunk". The chunk size is
// independent of the arena size, so there may be one or more chunks per arena.
//
// The root level (referred to as L0 and index 0 in pageAlloc.summary) has each
// summary represent the largest section of address space (16 GiB on 64-bit systems),
// with each subsequent level representing successively smaller subsections until we
// reach the highest granularity at the leaves, a chunk.
//
// More specifically, each summary in each level (except for leaf summaries)
// represent some number of entries in the following level. For example, each
// summary in the lowest level may represent a 16 GiB region of address space,
// and in the next level there could be 8 corresponding entries which represent 2
// GiB subsections of that 16 GiB region, each of which could correspond to 8
// entries in the next level which each represent 256 MiB regions, and so on.
//
// Thus, this design only scales to heaps so large, but can always be extended to
// larger heaps by simply adding levels to the radix tree, which mostly costs
// additional virtual address space. The choice of managing large arrays also means
// that a large amount of virtual address space may be reserved by the runtime.

package runtime

const (
	logPagesPerArena = logHeapArenaBytes - pageShift

	// The amount of bits (i.e. pages) to consider in the bitmap at once. It may be that each
	// arena's bitmap shard is larger than this, but we divide it into equal sections
	// of allocBitmapChunk to avoid spending too much time running over bitmaps.
	logMallocChunkPages  = 9
	logMallocChunkBytes  = logMallocChunkPages + pageShift
	mallocChunkPages     = 1 << logMallocChunkPages
	mallocChunkBytes     = 1 << logMallocChunkBytes
	mallocChunksPerArena = pagesPerArena / mallocChunkPages

	// The number of radix bits for each level.
	// summaryLevels is an architecture-dependent value defined in mpagealloc_*.go.
	// summaryL0Bits + (summaryLevels-1)*summaryLevelBits + logMallocChunkBytes = heapAddrBits
	summaryLevelBits = 3
	summaryL0Bits    = heapAddrBits - logMallocChunkBytes - (summaryLevels-1)*summaryLevelBits

	// Maximum searchAddr value, which indicates that the heap has no free space.
	maxSearchAddr = ^uintptr(0) - arenaBaseOffset
)

// Global chunk index.
//
// Represents an index into the highest tier in the summary structure.
// Similar to arenaIndex, except instead of arenas, it divides the address
// space into chunks.
type chunkIdx int

// chunkIndex returns the global index of the malloc chunk containing the
// pointer p.
func chunkIndex(p uintptr) chunkIdx {
	return chunkIdx((p + arenaBaseOffset) / mallocChunkBytes)
}

// chunkIndex returns the base address of the malloc chunk at index ci.
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
//
// The returned range is inclusive on the lower bound and exclusive on
// the upper bound.
func addrsToSummaryRange(level int, base, limit uintptr) (lo int, hi int) {
	// Pull out necessary level-related "constants".
	e := uintptr(1) << levelBits[level] // number of entries in an entry block for this level
	sh := levelShift[level]             // radix position (shift count) for this level

	// lo and hi and are the base and the limit rounded down and up
	// to the closest set of entries for this level, respectively.
	//
	// Note that limit represents the inclusive upper bound, so in
	// order to get the rounding right, we need to make sure that
	// after the shift into an index, we increment by 1.
	rlo := (base + arenaBaseOffset) >> sh
	rhi := ((limit + arenaBaseOffset) >> sh) + 1

	// Round down/up to nearest multiple of entries in this level.
	return int(alignDown(rlo, e)), int(alignUp(rhi, e))
}

//go:notinheap
type pageAlloc struct {
	// Radix tree of summaries.
	//
	// Each slice's cap represents the whole memory reservation,
	// whereas the len the an upper-bound on the available heap.
	summary [summaryLevels][]mallocSum

	// The address to start an allocation search with.
	searchAddr uintptr

	// start and end represent the chunk indices
	// which pageAlloc knows about. It assumes
	// chunks in the range [start, end) are
	// currently ready to use.
	start, end chunkIdx

	// Reference to an mheap, used for testing by using
	// a dummy mheap structure.
	mheap *mheap
}

func (s *pageAlloc) init(h *mheap, sysStat *uint64) {
	// System-dependent initialization.
	s.sysInit(sysStat)

	// Start with the searchAddr in a state indicating there's no free memory.
	s.searchAddr = maxSearchAddr

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
	// Perform the actual growth in a system-dependent manner.
	// We just update a bunch of additional metadata here.
	s.sysGrow(base, size, sysStat)

	// Round up to chunks, since we can't deal with increments smaller
	// than chunks.
	base = alignDown(base, mallocChunkBytes)
	size = alignUp(size, mallocChunkBytes)
	limit := base + size

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

	// Update the searchAddr since the arena must have free pages in it (except
	// during tests, but it's always OK if the searchAddr is too low).
	// Compare with arenaBaseOffset added because this gives us a more
	// consistent view of the heap.
	if base+arenaBaseOffset < s.searchAddr+arenaBaseOffset {
		s.searchAddr = base
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

// Helper for alloc. Represents the slow path and the full summary
// structure search.
//
// Returns a base address for npages contiguous free pages and a
// new potential searchAddr. This searchAddr may not be better than s.searchAddr.
//
// Returns a base address of 0 on failure, in which case the potential
// searchAddr is invalid and must be ignored.
func (s *pageAlloc) allocSlow(npages uintptr) (uintptr, uintptr) {
	// Search algorithm
	//
	// This algorithm walks each level l of the radix tree from the root level
	// to the leaf level. It iterates over at most 1 << levelBits[l] of entries
	// in a given level in the radix tree, and uses the summary information to
	// find either:
	//  1) That a given region contains a large enough contiguous region, at
	//     which point it continues iterating on the next level, or
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
	// searchAddr is the best searchAddr we could glean from our search.
	// foundBestSearchAddr indicates whether it's been found. Note that 0 may
	// be a valid "best searchAddr", especially on 32-bit architectures.
	searchAddr := uintptr(0)
	foundBestSearchAddr := false
	lastsum := packMallocSum(0, 0, 0)
	lastidx := -1
nextLevel:
	for l := 0; l < len(s.summary); l++ {
		b := levelBits[l]
		e := 1 << b
		i <<= b
		level := s.summary[l][i : i+e]
		p := levelLogPages[l]
		// Determine j0, the first index we should start iterating from.
		// The searchAddr may help us eliminate iterations if we followed the
		// searchAddr on the previous level, in which case the top bits of the
		// searchAddr address should be the same as i, after levelShift.
		j0 := 0
		if searchAddrIdx := int((s.searchAddr + arenaBaseOffset) >> levelShift[l]); searchAddrIdx&^(e-1) == i {
			j0 = searchAddrIdx & (e - 1)
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
				lastidx = i + j
				lastsum = sum
				i += j
				continue nextLevel
			}
			if !foundBestSearchAddr {
				// When we've reached this point and haven't found
				// the latest non-zero summary, that means this is the
				// latest non-zero summary that we can possibly find
				// during this search. At this point we may look further
				// in this level, but we won't be able to gain any new
				// new information even if we drop down again.
				searchAddr = uintptr((i+j)<<levelShift[l]) - arenaBaseOffset
				foundBestSearchAddr = true
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
			if !foundBestSearchAddr {
				searchAddr = addr + npages*pageSize
				foundBestSearchAddr = true
			}
			return addr, searchAddr
		}
		if l != 0 {
			print("runtime: summary[", l-1, "][", lastidx, "] = ", lastsum.start(), ", ", lastsum.max(), ", ", lastsum.end(), "\n")
			print("runtime: level = ", l, ", npages = ", npages, ", j0 = ", j0, "\n")
			print("runtime: s.searchAddr = ", hex(s.searchAddr), ", i = ", i, ", levelShift[level] = ", levelShift[l], ", e = ", e, "\n")
			for j := 0; j < len(level); j++ {
				sum := level[j]
				print("runtime: summary[", l, "][", i+j, "] = (", sum.start(), ", ", sum.max(), ", ", sum.end(), ")\n")
				if l == len(s.summary)-1 {
					if a := s.arenas(arenaIdx(i + j)); a != nil {
						for z, b := range a.pageAlloc {
							print("runtime: mallocbits[", z, "] = ", hex(uintptr(b)), "\n")
						}
					}
				}
			}
			throw("bad summary data")
		}
		// We're at level zero, so that means we've exhausted our search.
		return 0, maxSearchAddr
	}
	// Since we've gotten to this point, that means we haven't found a
	// sufficiently-sized free region straddling some boundary (arena or larger),
	// so just allocate directly out of the arena without a searchAddr.
	//
	// After iterating over all levels, i must contain an arena index which
	// is what the final level represents.
	ai := arenaIdx(i / mallocChunksPerArena)
	j, h := s.arenas(ai).pageAlloc.alloc(npages, (i%mallocChunksPerArena)*mallocChunkPages)
	if j < 0 {
		max := s.summary[len(s.summary)-1][i].max()
		print("runtime: max = ", max, ", npages = ", npages, "\n")
		for z, b := range s.arenas(arenaIdx(i)).pageAlloc {
			print("runtime: mallocbits[", z, "] = ", hex(uintptr(b)), "\n")
		}
		throw("bad summary data")
	}
	addr := arenaBase(ai) + uintptr(j)*pageSize
	// If at this point we still haven't found the first free space, that
	// means this is it!
	if !foundBestSearchAddr {
		searchAddr = arenaBase(ai) + uintptr(h)*pageSize
		foundBestSearchAddr = true
	}
	return addr, searchAddr
}

// alloc allocates npages worth of memory from the page heap, returning the base
// address for the allocation.
//
// Returns 0 on failure.
func (s *pageAlloc) alloc(npages uintptr) uintptr {
	// If the searchAddr refers to a region which has a higher address than
	// any known arena, then we know we're out of memory.
	if chunkIndex(s.searchAddr) >= s.end {
		return 0
	}

	// If npages has a chance of fitting in the chunk where the searchAddr is,
	// try to allocate from it directly.
	var addr, searchAddr uintptr
	if mallocChunkPages-chunkPageIndex(s.searchAddr) >= int(npages) {
		// npages is guaranteed to be no greater than pagesPerArena here.
		ci := chunkIndex(s.searchAddr)
		if max := s.summary[len(s.summary)-1][ci].max(); max >= int(npages) {
			i := arenaIndex(s.searchAddr)
			j, h := s.arenas(i).pageAlloc.alloc(npages, arenaPageIndex(s.searchAddr))
			if j < 0 {
				print("runtime: max = ", max, ", npages = ", npages, "\n")
				print("runtime: searchAddrIndex = ", arenaPageIndex(s.searchAddr), ", s.searchAddr = ", hex(s.searchAddr), "\n")
				for z, b := range s.arenas(i).pageAlloc {
					print("runtime: mallocbits[", z, "] = ", hex(uintptr(b)), "\n")
				}
				throw("bad summary data")
			}
			addr = arenaBase(i) + uintptr(j)*pageSize
			searchAddr = arenaBase(i) + uintptr(h)*pageSize
			goto done
		}
	}
	// We failed to use a searchAddr for one reason or another, so try
	// the slow path.
	addr, searchAddr = s.allocSlow(npages)
	if addr == 0 {
		if npages == 1 {
			// We failed to find a single free page, the smallest unit
			// of allocation. This means we know the heap is completely
			// exhausted. Otherwise, the heap still might have free
			// space in it, just not enough contiguous space to
			// accommodate npages.
			s.searchAddr = maxSearchAddr
		}
		return 0
	}
done:
	// If we found a better searchAddr, update our searchAddr.
	if searchAddr+arenaBaseOffset > s.searchAddr+arenaBaseOffset {
		s.searchAddr = searchAddr
	}
	// We have a non-zero address, so update and return.
	s.update(addr, npages, true, true)
	return addr
}

// Helper for free. Represents the slow path for freeing, that is,
// having to free memory that spans more than one arena.
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

// free returns npages worth of memory starting at base back to the page heap.
func (s *pageAlloc) free(base, npages uintptr) {
	// Because of split address spaces, we must compare against the searchAddr address
	// with both sides containing arenaBaseOffset, because this gives us a linear
	// view of the address space and ensures that we can do a simple comparison.
	if base+arenaBaseOffset < s.searchAddr+arenaBaseOffset {
		// We need to convert a pointer into an offset-pointer.
		s.searchAddr = base
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
