// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm mips mipsle wasm

// wasm is a treated as a 32-bit architecture for the purposes of the page
// allocator, even though it has 64-bit pointers. This is because any wasm
// pointer always has its top 32 bits as zero, so the effective heap address
// space is only 2^32 bytes in size (see heapAddrBits).

package runtime

const (
	// The number of tiers in the summary structure.
	summaryLevels = 4

	// log2 of the total number of entries in each level.
	logSummaryL3Size = heapAddrBits - logMallocChunkBytes
	logSummaryL2Size = heapAddrBits - logMallocChunkBytes - 1*summaryLevelBits
	logSummaryL1Size = heapAddrBits - logMallocChunkBytes - 2*summaryLevelBits
	logSummaryL0Size = heapAddrBits - logMallocChunkBytes - 3*summaryLevelBits // == summaryL0Bits
)
