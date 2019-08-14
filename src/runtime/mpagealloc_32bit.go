// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm mips mipsle wasm

// wasm is a treated as a 32-bit architecture for the purposes of the page
// allocator, even though it has 64-bit pointers. This is because any wasm
// pointer always has its top 32 bits as zero, so the effective heap address
// space is only 2^32 bytes in size (see heapAddrBits).

package runtime

import "unsafe"

const (
	// The number of tiers in the summary structure.
	summaryLevels = 4

	// log2 of the total number of entries in each level.
	logSummaryL3Size = heapAddrBits - logMallocChunkBytes
	logSummaryL2Size = heapAddrBits - logMallocChunkBytes - 1*summaryLevelBits
	logSummaryL1Size = heapAddrBits - logMallocChunkBytes - 2*summaryLevelBits
	logSummaryL0Size = heapAddrBits - logMallocChunkBytes - 3*summaryLevelBits // == summaryL0Bits

	// Constants for testing.
	pageAlloc32Bit = 1
	pageAlloc64Bit = 0
)

// See comment in mpagealloc_64bit.go.
var levelBits = [summaryLevels]uint{
	summaryL0Bits,
	summaryLevelBits,
	summaryLevelBits,
	summaryLevelBits,
}

// See comment in mpagealloc_64bit.go.
var levelShift = [summaryLevels]uint{
	heapAddrBits - logSummaryL0Size,
	heapAddrBits - logSummaryL1Size,
	heapAddrBits - logSummaryL2Size,
	heapAddrBits - logSummaryL3Size,
}

// See comment in mpagealloc_64bit.go.
var levelLogPages = [summaryLevels]uint{
	logMallocChunkPages + 3*summaryLevelBits,
	logMallocChunkPages + 2*summaryLevelBits,
	logMallocChunkPages + 1*summaryLevelBits,
	logMallocChunkPages,
}

func (s *pageAlloc) sysInit(sysStat *uint64) {
	// Calculate how much memory all our entries will take up.
	//
	// This should be around 12 KiB or less.
	totalSize := uintptr(0)
	for l := 0; l < summaryLevels; l++ {
		totalSize += (uintptr(1) << (heapAddrBits - levelShift[l])) * mallocSumBytes
	}
	totalSize = alignUp(totalSize, physPageSize)

	// Reserve memory for all levels in one go. There shouldn't be much for 32-bit.
	reservation := sysReserve(nil, totalSize)
	if reservation == nil {
		throw("failed to reserve summary structure space")
	}
	// There isn't much. Just map it and mark it as used immediately.
	sysMap(reservation, totalSize, sysStat)
	sysUsed(reservation, totalSize)

	// Iterate over the reservation and cut it up into slices.
	//
	// Maintain i as the byte offset from reservation where
	// the new slice should start.
	i := uintptr(0)
	for l := 0; l < summaryLevels; l++ {
		r := unsafe.Pointer(uintptr(reservation) + i)
		e := 1 << (heapAddrBits - levelShift[l])

		// Put this reservation into a slice.
		sl := notInHeapSlice{(*notInHeap)(r), 0, e}
		s.summary[l] = *(*[]mallocSum)(unsafe.Pointer(&sl))

		i += uintptr(e) * mallocSumBytes
	}
}

func (s *pageAlloc) sysGrow(base, size uintptr, _ *uint64) {
	// Round up to chunks, since we can't deal with increments smaller
	// than chunks.
	base = alignDown(base, mallocChunkBytes)
	size = alignUp(size, mallocChunkBytes)

	// Walk up the tree and update the summary slices.
	for l := len(s.summary) - 1; l >= 0; l-- {
		// Update the summary slices with a new upper-bound. This ensures
		// we get tight bounds checks on at least the top bound.
		_, hi := addrsToSummaryRange(l, base, base+size-1)
		if hi > len(s.summary[l]) {
			s.summary[l] = s.summary[l][:hi]
		}
	}
}
