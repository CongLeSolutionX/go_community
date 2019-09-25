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
// The highest level of the radix tree (i.e. the leaves) contains the smallest
// unit summaries represent, which we call a "malloc chunk". The chunk size is
// independent of the arena size, so there may be one or more chunks per arena.
//
// The lowest level (referred to as L0 and index 0 in pageAlloc.summary) has each
// summary represent the largest section of address space (16 GiB on 64-bit systems),
// with each higher level representing successively smaller subsection until we reach
// the highest granularity in the highest level, which is a chunk.
//
// More specifically, each summary in each level (except for leaf summaries)
// represent some number of entries in the following level. For example, each
// summary in the lowest level may represent a 16 GiB region of address space,
// and in the next level there could be 8 corresponding entries which represent 2
// GiB subsections of that 16 GiB region, each of which could correspond to 8
// entries in the next level which each represent 256 MiB regions, and so on.
//
// Thus, this design only scales to heaps so large, but can always be extended to
// larger heaps by simply adding levels to the radix tree, which mostly cost additional
// virtual address space. The choice of managing large arrays also means that
// a large amount of virtual address space may be reserved by the runtime.

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
)

const (
	mallocSumBytes    = 8
	logMaxPackedValue = logMallocChunkPages + (summaryLevels-1)*summaryLevelBits
	maxPackedValue    = 1 << logMaxPackedValue
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
