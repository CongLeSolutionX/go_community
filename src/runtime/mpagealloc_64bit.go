// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 arm64 mips64 mips64le ppc64 ppc64le s390x

package runtime

const (
	// The number of tiers in the summary structure.
	summaryLevels = 5

	// log2 of the total number of entries in each level.
	logSummaryL4Size = heapAddrBits - logMallocChunkBytes
	logSummaryL3Size = heapAddrBits - logMallocChunkBytes - 1*summaryLevelBits
	logSummaryL2Size = heapAddrBits - logMallocChunkBytes - 2*summaryLevelBits
	logSummaryL1Size = heapAddrBits - logMallocChunkBytes - 3*summaryLevelBits
	logSummaryL0Size = heapAddrBits - logMallocChunkBytes - 4*summaryLevelBits // == summaryL0Bits
)
