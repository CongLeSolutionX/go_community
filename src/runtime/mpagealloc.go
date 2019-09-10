// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page allocator.
//
// The page allocator manages mapped pages (defined by pageSize, NOT
// physPageSize) for allocation and re-use. It is embedded into mheap.
//
// Pages are managed using a bitmap that is sharded into chunks.
// In the bitmap, 1 means in-use, and 0 means free. The bitmap spans the
// process's address space. Chunks are allocated using a SLAB allocator
// and pointers to chunks are managed in one large array, which is mapped
// in as needed.
//
// The bitmap is efficiently searched by using a radix tree in combination
// with fast bit-wise intrinsics. Allocation is performed using an address-ordered
// first-fit approach.
//
// Each entry in the radix tree is a summary that describes three properties of
// a particular region of the address space: the number of contiguous free pages
// at the start and end of the region it represents, and the maximum number of
// contiguous free pages found anywhere in that region.
//
// Each level of the radix tree is stored as one contiguous array, which represents
// a different granularity of subdivision of the processes' address space. Thus, this
// radix tree is actually implicit in these large arrays, as opposed to having explicit
// dynamically-allocated pointer-based node structures. Naturally, these arrays may be
// quite large for system with large address spaces, so in these cases they are mapped
// into memory as needed. The leaf summaries of the tree correspond to a bitmap chunk.
//
// The root level (referred to as L0 and index 0 in pageAlloc.summary) has each
// summary represent the largest section of address space (16 GiB on 64-bit systems),
// with each subsequent level representing successively smaller subsections until we
// reach the finest granularity at the leaves, a chunk.
//
// More specifically, each summary in each level (except for leaf summaries)
// represents some number of entries in the following level. For example, each
// summary in the root level may represent a 16 GiB region of address space,
// and in the next level there could be 8 corresponding entries which represent 2
// GiB subsections of that 16 GiB region, each of which could correspond to 8
// entries in the next level which each represent 256 MiB regions, and so on.
//
// Thus, this design only scales to heaps so large, but can always be extended to
// larger heaps by simply adding levels to the radix tree, which mostly costs
// additional virtual address space. The choice of managing large arrays also means
// that a large amount of virtual address space may be reserved by the runtime.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

const (
	// The size of a bitmap chunk, i.e. the amount of bits (that is, pages) to consider
	// in the bitmap at once.
	mallocChunkPages    = 1 << logMallocChunkPages
	mallocChunkBytes    = mallocChunkPages * pageSize
	logMallocChunkPages = 9
	logMallocChunkBytes = logMallocChunkPages + pageShift

	// The number of radix bits for each level.
	//
	// The value of 3 is chosen such that the block of summaries we need to scan at
	// each level in fits in 64 bytes (2^3 summaries * 8 bytes per summary), which is
	// close to the L1 cache line width on many systems. Also, a value of 3 fits 4 tree
	// levels perfectly into the 21-bit mallocBits summary field at the root level.
	//
	// The following equation explains how each of the constants relate:
	// summaryL0Bits + (summaryLevels-1)*summaryLevelBits + logMallocChunkBytes = heapAddrBits
	//
	// summaryLevels is an architecture-dependent value defined in mpagealloc_*.go.
	summaryLevelBits = 3
	summaryL0Bits    = heapAddrBits - logMallocChunkBytes - (summaryLevels-1)*summaryLevelBits

	// Maximum searchAddr value, which indicates that the heap has no free space.
	//
	// We subtract arenaBaseOffset because we want this to represent the maximum
	// value in the shifted address space, but searchAddr is stored as a regular
	// memory address. See arenaBaseOffset for details.
	maxSearchAddr = ^uintptr(0) - arenaBaseOffset

	// Minimum scavAddr value, which indicates that the scavenger is done.
	//
	// minScavAddr + arenaBaseOffset == 0
	minScavAddr = (^uintptr(0) >> logArenaBaseOffset) * arenaBaseOffset
)

// Global chunk index.
//
// Represents an index into the leaf level of the summary structure.
// Similar to arenaIndex, except instead of arenas, it divides the address
// space into chunks.
type chunkIdx uint

// chunkIndex returns the global index of the malloc chunk containing the
// pointer p.
func chunkIndex(p uintptr) chunkIdx {
	return chunkIdx((p + arenaBaseOffset) / mallocChunkBytes)
}

// chunkIndex returns the base address of the malloc chunk at index ci.
func chunkBase(ci chunkIdx) uintptr {
	return uintptr(ci)*mallocChunkBytes - arenaBaseOffset
}

// chunkPageIndex computes the index of the page that contains p,
// relative to the chunk which contains p.
func chunkPageIndex(p uintptr) uint {
	return uint(p % mallocChunkBytes / pageSize)
}

// addrsToSummaryRange converts base and limit pointers into a range
// of entries for the given summary level.
//
// The returned range is inclusive on the lower bound and exclusive on
// the upper bound.
func addrsToSummaryRange(level int, base, limit uintptr) (lo int, hi int) {
	// This is slightly more nuanced than just a shift for the exclusive
	// upper-bound. Note that the exclusive upper bound may be within a
	// summary at this level, meaning if we just do the obvious computation
	// hi will end up being an inclusive upper bound. Unfortunately, just
	// adding 1 to that is too broad since we might be on the very edge of
	// of a summary's max page count boundary for this level
	// (1 << logLevelPages[level]). So, make limit an inclusive upper bound
	// then shift, then add 1, so we get an exclusive upper bound at the end.
	lo = int((base + arenaBaseOffset) >> levelShift[level])
	hi = int(((limit-1)+arenaBaseOffset)>>levelShift[level]) + 1
	return
}

// blockAlignSummaryRange aligns indices into the given level to that
// level's block width (1 << levelBits[level]). It assumes lo is inclusive
// and hi is exclusive, and so aligns them down and up respectively.
func blockAlignSummaryRange(level int, lo, hi int) (int, int) {
	e := uintptr(1) << levelBits[level]
	return int(alignDown(uintptr(lo), e)), int(alignUp(uintptr(hi), e))
}

type pageAlloc struct {
	// Radix tree of summaries.
	//
	// Each slice's cap represents the whole memory reservation.
	// Each slice's len reflects the allocator's maximum known
	// mapped heap address for that level.
	//
	// The backing store of each summary level is reserved in init
	// and may or may not be committed in grow (small address spaces
	// may commit all the memory in init).
	//
	// The purpose of keeping len <= cap is to enforce bounds checks
	// on the top end of the slice so that instead of an unknown
	// runtime segmentation fault, we get a much friendlier out-of-bounds
	// error.
	//
	// We may still get segmentation faults < len since some of that
	// memory may not be committed yet.
	summary [summaryLevels][]mallocSum

	// chunks is a slice of bitmap chunks.
	//
	// The backing store for chunks is reserved in init and committed
	// by grow.
	//
	// To find the chunk containing a memory address `a`, do:
	//   chunks[chunkIndex(a)]
	//
	// Each chunk is allocated out of mallocDataAlloc.
	chunks          []*mallocData
	mallocDataAlloc fixalloc // allocator for mallocData

	// The address to start an allocation search with.
	//
	// When added with arenaBaseOffset, we guarantee that
	// all valid heap addresses (when also added with
	// arenaBaseOffset) below this value are allocated and
	// not worth searching.
	//
	// Note that adding in arenaBaseOffset transforms addresses
	// to a new address space with a linear view of the full address
	// space on architectures with segmented address spaces.
	searchAddr uintptr

	// The address to start a scavenge candidate search with.
	scavAddr uintptr

	// start and end represent the chunk indices
	// which pageAlloc knows about. It assumes
	// chunks in the range [start, end) are
	// currently ready to use.
	start, end chunkIdx

	// mheap_.lock. This level of indirection makes it possible
	// to test pageAlloc indepedently of the runtime allocator.
	mheapLock *mutex

	// Whether or not this struct is being used in tests.
	test bool
}

func (s *pageAlloc) init(mheapLock *mutex, sysStat *uint64) {
	// System-dependent initialization.
	s.sysInit(sysStat)

	// Start with the searchAddr in a state indicating there's no free memory.
	s.searchAddr = maxSearchAddr

	// Start with the scavAddr in a state indicating there's nothing more to do.
	s.scavAddr = minScavAddr

	// Initialize allocator for mallocBits.
	s.mallocDataAlloc.init(unsafe.Sizeof(mallocData{}), nil, nil, sysStat)

	// Reserve space for pointers to mallocBits and put this reservation
	// into the chunks slice.
	const maxChunks = (1 << heapAddrBits) / mallocChunkBytes
	r := sysReserve(nil, maxChunks*sys.PtrSize)
	sl := notInHeapSlice{(*notInHeap)(r), 0, maxChunks}
	s.chunks = *(*[]*mallocData)(unsafe.Pointer(&sl))

	// Set the mheapLock.
	s.mheapLock = mheapLock
}

// extendMappedRegion ensures that all the memory in the range
// [base+nbase, base+nlimit) is in the Ready state.
// base must refer to the beginning of a memory region in the
// Reserved state. extendMappedRegion assumes that the region
// [base+mbase, base+mlimit) is already mapped.
//
// Note that extendMappedRegion only supports extending
// mappings in one direction. Therefore,
// nbase < mbase && nlimit > mlimit is an invalid input
// and this function will throw.
func extendMappedRegion(base unsafe.Pointer, mbase, mlimit, nbase, nlimit uintptr, sysStat *uint64) {
	// Round the offsets to a physical page.
	mbase = alignDown(mbase, physPageSize)
	nbase = alignDown(nbase, physPageSize)
	mlimit = alignUp(mlimit, physPageSize)
	nlimit = alignUp(nlimit, physPageSize)

	// If none of the region is mapped, don't bother
	// trying to figure out which parts are.
	if mlimit-mbase != 0 {
		// Determine which part of the region actually needs
		// mapping.
		if nbase < mbase && nlimit > mlimit {
			// TODO(mknyszek): Consider supporting this case. It can't
			// ever happen currently in the page allocator, but may be
			// useful in the future. Also, it would make this function's
			// purpose simpler to explain.
			throw("mapped region extended in two directions")
		} else if nbase < mbase && nlimit <= mlimit {
			nlimit = mbase
		} else if nbase >= mbase && nlimit > mlimit {
			nbase = mlimit
		} else {
			return
		}
	}

	// Transition from Reserved to Ready.
	rbase := unsafe.Pointer(uintptr(base) + nbase)
	sysMap(rbase, nlimit-nbase, sysStat)
	sysUsed(rbase, nlimit-nbase)
}

// grow sets up the metadata for the address range [base, base+size).
// It may allocate metadata, in which case *sysStat will be updated.
//
// s.mheapLock must be held.
func (s *pageAlloc) grow(base, size uintptr, sysStat *uint64) {
	// Round up to chunks, since we can't deal with increments smaller
	// than chunks. Also, sysGrow expects aligned values.
	limit := alignUp(base+size, mallocChunkBytes)
	base = alignDown(base, mallocChunkBytes)

	// Grow the summary levels in a system-dependent manner.
	// We just update a bunch of additional metadata here.
	s.sysGrow(base, limit, sysStat)

	// Update s.start and s.end.
	// If no growth happened yet, start == 0. This is generally
	// safe since the zero page is unmapped.
	oldStart, oldEnd := s.start, s.end
	firstGrowth := s.start == 0
	start, end := chunkIndex(base), chunkIndex(limit)
	if firstGrowth || start < s.start {
		s.start = start
	}
	if end > s.end {
		s.end = end

		// s.end corresponds directly to the length of s.chunks,
		// so just update it here.
		s.chunks = s.chunks[:end]
	}

	// Extend the mapped part of the chunk reservation.
	extendMappedRegion(
		unsafe.Pointer(&s.chunks[0]),
		uintptr(oldStart)*sys.PtrSize,
		uintptr(oldEnd)*sys.PtrSize,
		uintptr(s.start)*sys.PtrSize,
		uintptr(s.end)*sys.PtrSize,
		sysStat,
	)

	// Update the searchAddr since the chunk must have free pages in it (except
	// during tests, but it's always OK if the searchAddr is too low).
	// Compare with arenaBaseOffset added because it gives us a linear view
	// of the heap on architectures with signed address spaces.
	if base+arenaBaseOffset < s.searchAddr+arenaBaseOffset {
		s.searchAddr = base
	}

	// Allocate new sections of the bitmap.
	for c := chunkIndex(base); c < chunkIndex(limit); c++ {
		m := (*mallocData)(s.mallocDataAlloc.alloc())

		// Newly-grown memory is always considered scavenged.
		m.scavengeRange(0, mallocChunkPages)

		// Store without write barrier since m is not in the heap,
		// and we're not allowed to have write barriers in the
		// allocation codepaths.
		atomic.StorepNoWB(unsafe.Pointer(&s.chunks[c]), unsafe.Pointer(m))
	}

	// Update summaries accordingly. The grow acts like a free, so
	// we need to ensure this newly-free memory is visible in the
	// summaries.
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
	sc, ec := chunkIndex(base), chunkIndex(limit)

	// Handle updating the lowest level first.
	if sc == ec {
		// Fast path: the allocation doesn't span more than one chunk,
		// so update this one and if the summary didn't change, return.
		x := s.summary[len(s.summary)-1][sc]
		y := s.chunks[sc].summarize()
		if x == y {
			return
		}
		s.summary[len(s.summary)-1][sc] = y
	} else if contig {
		// Slow contiguous path: the allocation spans more than one chunk
		// and at least one summary is guaranteed to change.
		summary := s.summary[len(s.summary)-1]

		// Update the summary for chunk sc.
		summary[sc] = s.chunks[sc].summarize()

		// Update the summaries for chunks in between, which are
		// either totally allocated or freed.
		whole := s.summary[len(s.summary)-1][sc+1 : ec]
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

		// Update the summary for chunk ec.
		summary[ec] = s.chunks[ec].summarize()
	} else {
		// Slow general path: the allocation spans more than one chunk
		// and at least one summary is guaranteed to change.
		//
		// We can't assume a contiguous allocation happened, so walk over
		// every chunk in the range and manually recompute the summary.
		summary := s.summary[len(s.summary)-1]
		for c := sc; c <= ec; c++ {
			summary[c] = s.chunks[c].summarize()
		}
	}

	// Walk up the radix tree and update the summaries appropriately.
	changed := true
	for l := len(s.summary) - 2; l >= 0 && changed; l-- {
		changed = false

		// "Constants" for the previous level which we
		// need to compute the summary from that level.
		logEntriesPerBlock := levelBits[l+1]
		logMaxPages := levelLogPages[l+1]

		// lo and hi describe all the parts of the level we need to look at.
		// Both lo and hi are aligned down and up to entriesPerBlock respectively.
		lo, hi := addrsToSummaryRange(l, base, limit+1)

		// Iterate over each block, updating the corresponding summary in the less-granular level.
		for i := lo; i < hi; i++ {
			children := s.summary[l+1][i<<logEntriesPerBlock : (i+1)<<logEntriesPerBlock]
			sum := mergeSummaries(children, logMaxPages)
			old := s.summary[l][i]
			if old != sum {
				changed = true
				s.summary[l][i] = sum
			}
		}
	}
}

// allocRange marks the range of memory [base, base+npages*pageSize) as
// allocated.
func (s *pageAlloc) allocRange(base, npages uintptr) uintptr {
	limit := base + npages*pageSize - 1
	sc, ec := chunkIndex(base), chunkIndex(limit)
	si, ei := chunkPageIndex(base), chunkPageIndex(limit)

	scav := uint(0)
	if sc == ec {
		// The range doesn't cross any chunk boundaries.
		scav += s.chunks[sc].allocRange(si, ei+1-si)
	} else {
		// The range crosses at least one chunk boundary.
		scav += s.chunks[sc].allocRange(si, mallocChunkPages-si)
		for c := sc + 1; c < ec; c++ {
			scav += s.chunks[c].allocAll()
		}
		scav += s.chunks[ec].allocRange(0, ei+1)
	}
	return uintptr(scav) * pageSize
}

// Helper for alloc. Represents the slow path and the full radix tree search.
//
// Returns a base address for npages contiguous free pages, a
// new potential searchAddr, and the number scavenged bytes in the newly-allocated
// region. The searchAddr need not be better than s.searchAddr.
//
// Returns a base address of 0 on failure, in which case the potential
// searchAddr is invalid and must be ignored.
func (s *pageAlloc) find(npages uintptr) (uintptr, uintptr) {
	// Search algorithm.
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
	// be a valid "best searchAddr". This case is especially common on 32-bit
	// architectures.
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
		var start, size uint
		for j := j0; j < len(level); j++ {
			sum := level[j]
			if sum == 0 {
				// A full entry means we broke any streak and
				// that we should skip it altogether.
				size = 0
				continue
			}
			s := sum.start()
			if size != 0 && size+s >= uint(npages) {
				// We hit npages; we're done!
				size += s
				break
			}
			if sum.max() >= uint(npages) {
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
				start = uint(j+1)<<p - size
				continue
			}
			// The entry is completely free, so continue the run.
			size += 1 << p
		}
		if size >= uint(npages) {
			// We found some range which crosses boundaries, just go ahead and
			// return it.
			addr := uintptr(i<<levelShift[l]) - arenaBaseOffset + uintptr(start)*pageSize
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
			}
			throw("bad summary data")
		}
		// We're at level zero, so that means we've exhausted our search.
		return 0, maxSearchAddr
	}
	// Since we've gotten to this point, that means we haven't found a
	// sufficiently-sized free region straddling some boundary (chunk or larger),
	// so just allocate directly out of the chunk without a searchAddr.
	//
	// After iterating over all levels, i must contain a chunk index which
	// is what the final level represents.
	ci := chunkIdx(i)
	j, h := s.chunks[ci].find(npages, 0)
	if j < 0 {
		max := s.summary[len(s.summary)-1][i].max()
		print("runtime: max = ", max, ", npages = ", npages, "\n")
		throw("bad summary data")
	}
	addr := chunkBase(ci) + uintptr(j)*pageSize
	// If at this point we still haven't found the first free space, that
	// means this is it!
	if !foundBestSearchAddr {
		searchAddr = chunkBase(ci) + uintptr(h)*pageSize
		foundBestSearchAddr = true
	}
	return addr, searchAddr
}

// alloc allocates npages worth of memory from the page heap, returning the base
// address for the allocation and the amount of scavenged memory in bytes
// contained in the region [base address, base address + npages*pageSize).
//
// Returns a 0 base address on failure, in which case other returned values
// should be ignored.
func (s *pageAlloc) alloc(npages uintptr) (uintptr, uintptr) {
	// If the searchAddr refers to a region which has a higher address than
	// any known chunk, then we know we're out of memory.
	if chunkIndex(s.searchAddr) >= s.end {
		return 0, 0
	}

	// If npages has a chance of fitting in the chunk where the searchAddr is,
	// search it directly.
	var addr, searchAddr uintptr
	if mallocChunkPages-chunkPageIndex(s.searchAddr) >= uint(npages) {
		// npages is guaranteed to be no greater than pagesPerArena here.
		i := chunkIndex(s.searchAddr)
		if max := s.summary[len(s.summary)-1][i].max(); max >= uint(npages) {
			j, h := s.chunks[i].find(npages, chunkPageIndex(s.searchAddr))
			if j < 0 {
				print("runtime: max = ", max, ", npages = ", npages, "\n")
				print("runtime: searchAddrIndex = ", chunkPageIndex(s.searchAddr), ", s.searchAddr = ", hex(s.searchAddr), "\n")
				throw("bad summary data")
			}
			addr = chunkBase(i) + uintptr(j)*pageSize
			searchAddr = chunkBase(i) + uintptr(h)*pageSize
			goto Found
		}
	}
	// We failed to use a searchAddr for one reason or another, so try
	// the slow path.
	addr, searchAddr = s.find(npages)
	if addr == 0 {
		if npages == 1 {
			// We failed to find a single free page, the smallest unit
			// of allocation. This means we know the heap is completely
			// exhausted. Otherwise, the heap still might have free
			// space in it, just not enough contiguous space to
			// accommodate npages.
			s.searchAddr = maxSearchAddr
		}
		return 0, 0
	}
Found:
	// Go ahead and actually mark the bits now that we have an address.
	scav := s.allocRange(addr, npages)

	// If we found a better searchAddr, update our searchAddr.
	if searchAddr+arenaBaseOffset > s.searchAddr+arenaBaseOffset {
		s.searchAddr = searchAddr
	}

	// We have a non-zero address, so update and return.
	s.update(addr, npages, true, true)
	return addr, scav
}

// free returns npages worth of memory starting at base back to the page heap.
func (s *pageAlloc) free(base, npages uintptr) {
	// Because of split address spaces, we must compare against the searchAddr address
	// with both sides containing arenaBaseOffset, because this gives us a linear
	// view of the address space and ensures that we can do a simple comparison.
	if base+arenaBaseOffset < s.searchAddr+arenaBaseOffset {
		s.searchAddr = base
	}
	if npages == 1 {
		// Fast path: we're clearing a single bit, and we know exactly
		// where it is, so mark it directly.
		s.chunks[chunkIndex(base)].free1(chunkPageIndex(base))
	} else {
		// Slow path: we're clearing more bits so we may need to iterate.
		limit := base + npages*pageSize - 1
		sc, ec := chunkIndex(base), chunkIndex(limit)
		si, ei := chunkPageIndex(base), chunkPageIndex(limit)

		if sc == ec {
			// The range doesn't cross any chunk boundaries.
			s.chunks[sc].free(si, ei+1-si)
		} else {
			// The range crosses one chunk boundary.
			s.chunks[sc].free(si, mallocChunkPages-si)
			for c := sc + 1; c < ec; c++ {
				s.chunks[c].freeAll()
			}
			s.chunks[ec].free(0, ei+1)
		}
	}
	s.update(base, npages, true, false)
}

const (
	mallocSumBytes = unsafe.Sizeof(mallocSum(0))

	// maxPackedValue is the maximum value that any of the three fields in
	// the mallocSum may take on.
	maxPackedValue    = 1 << logMaxPackedValue
	logMaxPackedValue = logMallocChunkPages + (summaryLevels-1)*summaryLevelBits

	freeChunkSum = mallocSum(uint64(mallocChunkPages) |
		uint64(mallocChunkPages<<logMaxPackedValue) |
		uint64(mallocChunkPages<<(2*logMaxPackedValue)))
)

// mallocSum is a packed summary type which packs three numbers: start, max,
// and end into a single 8-byte value. Each of these values are a summary of
// a bitmap and are thus counts, each of which may have a maximum value of
// 2^21 - 1, or all three may be equal to 2^21. The latter case is represented
// by just setting the 64th bit.
type mallocSum uint64

// packMallocSum takes a start, max, and end value and produces a mallocSum.
func packMallocSum(start, max, end uint) mallocSum {
	if max == maxPackedValue {
		return mallocSum(uint64(1 << 63))
	}
	return mallocSum((uint64(start) & (maxPackedValue - 1)) |
		((uint64(max) & (maxPackedValue - 1)) << logMaxPackedValue) |
		((uint64(end) & (maxPackedValue - 1)) << (2 * logMaxPackedValue)))
}

// start extracts the start value from a packed sum.
func (p mallocSum) start() uint {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return uint(uint64(p) & (maxPackedValue - 1))
}

// max extracts the max value from a packed sum.
func (p mallocSum) max() uint {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return uint((uint64(p) >> logMaxPackedValue) & (maxPackedValue - 1))
}

// end extracts the end value from a packed sum.
func (p mallocSum) end() uint {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue
	}
	return uint((uint64(p) >> (2 * logMaxPackedValue)) & (maxPackedValue - 1))
}

// unpack unpacks all three values from the summary.
func (p mallocSum) unpack() (uint, uint, uint) {
	if uint64(p)&uint64(1<<63) != 0 {
		return maxPackedValue, maxPackedValue, maxPackedValue
	}
	return uint(uint64(p) & (maxPackedValue - 1)),
		uint((uint64(p) >> logMaxPackedValue) & (maxPackedValue - 1)),
		uint((uint64(p) >> (2 * logMaxPackedValue)) & (maxPackedValue - 1))
}

// mergeSummaries merges consecutive summaries which may each represent at
// most 1 << logMaxPagesPerSum pages each together into one.
func mergeSummaries(sums []mallocSum, logMaxPagesPerSum uint) mallocSum {
	// Merge the summaries in sums into one.
	//
	// We do this by keeping a running summary representing the merged
	// summaries of sums[:i] in start, max, and end.
	start, max, end := sums[0].unpack()
	for i := 1; i < len(sums); i++ {
		// Merge in sums[i].
		si, mi, ei := sums[i].unpack()

		// Merge in sums[i].start only if the running summary is
		// completely free, otherwise this summary's start
		// plays no role in the combined sum.
		if start == uint(i)<<logMaxPagesPerSum {
			start += si
		}

		// Recompute the max value of the running sum by looking
		// across the boundary between the running sum and sums[i]
		// and at the max sums[i], taking the greatest of those two
		// and the max of the running sum.
		if end+si > max {
			max = end + si
		}
		if mi > max {
			max = mi
		}

		// Merge in end by checking if this new summary is totally
		// free. If it is, then we want to extend the the running
		// sum's end by the new summary. If not, then we have some
		// alloc'd pages in there and we just want to take the end
		// value in sums[i].
		if ei == 1<<logMaxPagesPerSum {
			end += 1 << logMaxPagesPerSum
		} else {
			end = ei
		}
	}
	return packMallocSum(start, max, end)
}
