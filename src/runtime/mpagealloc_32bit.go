// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm mips mipsle wasm

package runtime

const (
	// The number of tiers in the summary structure.
	summaryLevels = 4

	// The number of radix bits for each level.
	// summaryL0Bits + (summaryLevels-1)*summaryLevelBits + logMallocChunkBytes = heapAddrBits
	summaryLevelBits = 3
	summaryL0Bits    = heapAddrBits - logMallocChunkBytes - (summaryLevels-1)*summaryLevelBits

	// log2 of the total number of entries in each level.
	logSummaryL3Size = heapAddrBits - logMallocChunkBytes
	logSummaryL2Size = heapAddrBits - logMallocChunkBytes - 1*summaryLevelBits
	logSummaryL1Size = heapAddrBits - logMallocChunkBytes - 2*summaryLevelBits
	logSummaryL0Size = heapAddrBits - logMallocChunkBytes - 3*summaryLevelBits // == summaryL0Bits
)
