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
)
