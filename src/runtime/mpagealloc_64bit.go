// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 arm64 mips64 mips64le ppc64 ppc64le s390x

package runtime

import "unsafe"

const (
	// The number of levels in the radix tree.
	summaryLevels = 5

	// Constants for testing.
	pageAlloc32Bit = 0
	pageAlloc64Bit = 1
)

// levelBits is the number of bits in the radix for a given level in the super summary
// structure.
//
// Combined with levelShift, one can discover the summary at level l related to a
// given pointer p by doing:
//
// p >> levelShift[l]
//
// The sum of all the entries of levelBits should equal heapAddrBits.
var levelBits = [summaryLevels]uint{
	summaryL0Bits,
	summaryLevelBits,
	summaryLevelBits,
	summaryLevelBits,
	summaryLevelBits,
}

// levelShift is the number of bits to shift to acquire the radix for a given level
// in the super summary structure.
//
// See levelBits for usage information and examples.
var levelShift = [summaryLevels]uint{
	heapAddrBits - summaryL0Bits,
	heapAddrBits - summaryL0Bits - 1*summaryLevelBits,
	heapAddrBits - summaryL0Bits - 2*summaryLevelBits,
	heapAddrBits - summaryL0Bits - 3*summaryLevelBits,
	heapAddrBits - summaryL0Bits - 4*summaryLevelBits,
}

// levelLogPages is log2 the maximum number of runtime pages in the address space
// a summary in the given level represents.
//
// The highest tier always represents exactly log2 of 1 chunk's worth of pages.
var levelLogPages = [summaryLevels]uint{
	logMallocChunkPages + 4*summaryLevelBits,
	logMallocChunkPages + 3*summaryLevelBits,
	logMallocChunkPages + 2*summaryLevelBits,
	logMallocChunkPages + 1*summaryLevelBits,
	logMallocChunkPages,
}

func (s *pageAlloc) sysInit(_ *uint64) {
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
}

func (s *pageAlloc) sysGrow(base, size uintptr, sysStat *uint64) {
	// Round up to chunks, since we can't deal with increments smaller
	// than chunks.
	limit := alignUp(base+size, mallocChunkBytes)
	base = alignDown(base, mallocChunkBytes)

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
}
