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

const (
	logPagesPerArena = logHeapArenaBytes - pageShift

	// The maximum amount of arenas that could exist in this address space.
	logMaxArenas = heapAddrBits - logHeapArenaBytes
	maxArenas    = 1 << logMaxArenas

	// The amount of bits (i.e. pages) to consider in the bitmap at once. It may be that each
	// arena's bitmap shard is larger than this, but we divide it into equal sections
	// of allocBitmapChunk to avoid spending too much time running over bitmaps.
	logMallocChunkPages = 9
	logMallocChunkBytes = logMallocChunkPages + pageShift
	mallocChunkPages    = 1 << logMallocChunkPages

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
)
