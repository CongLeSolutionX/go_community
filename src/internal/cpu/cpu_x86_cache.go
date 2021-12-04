// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64

package cpu

const (
	// The default cache sizes are adapted for old x86-CPUs which do not
	// support implmented techniques to determine the cache size.
	defaultL1ISize = 32 << 10  // 32kb
	defaultL1DSize = 32 << 10  // 32kb
	defaultL2Size  = 256 << 10 // 256kb
	defaultL3Size  = 8 << 20   // 8mb
)

type intelCacheType uint32

const (
	Null intelCacheType = iota
	Instruction
	Data
)

type CacheInfo struct {
	l1iSize uint32
	l1dSize uint32
	l2Size  uint32
	l3Size  uint32
}

func extractBits(arg uint32, l int, r int) uint32 {
	if l > r {
		return 0
	}
	return (arg >> l) & ((1 << (r - l + 1)) - 1)
}

func getDefaultCacheInfo() CacheInfo {
	return CacheInfo{defaultL1ISize, defaultL1DSize, defaultL2Size, defaultL3Size}
}

// The current implementation for Intel CPUs rely on data from CPUID instruciton
// with EAX set to 04H. This cpuid leaf is available on all modern Intel CPUs (first
// mention of the leaf in the Intel Manual is dated from January 2004). If the leaf is not
// available the default sizes are used.
func getCacheInfoIntel() CacheInfo {
	cacheInfo := getDefaultCacheInfo()
	if maxFunctionInformation < 4 {
		return cacheInfo
	}

	for nxtecx := uint32(0); ; nxtecx++ {
		eax4, ebx4, ecx4, _ := cpuid(4, nxtecx)

		// Extracting Cache Type Field (see Table 3-8. Information Returned by CPUID Instruction in Intel Manual Vol.2).
		cacheType := intelCacheType(extractBits(eax4, 0, 4))

		// The parameters are reported until the null type is met.
		if cacheType == Null {
			return cacheInfo
		}

		// Extracting Cache Level (see Table 3-8. Information Returned by CPUID Instruction in Intel Manual Vol.2).
		cacheLevel := extractBits(eax4, 5, 7)

		// Extracting cache size in bytes = (Ways + 1) * (Partitions + 1) * (Line_Size + 1) * (Sets + 1)
		cacheSize := (extractBits(ebx4, 22, 31) + 1) * (extractBits(ebx4, 12, 21) + 1) * (extractBits(ebx4, 0, 11) + 1) * (ecx4 + 1)
		if cacheLevel == 1 && cacheType == Instruction {
			cacheInfo.l1iSize = cacheSize
		} else if cacheLevel == 1 && cacheType == Data {
			cacheInfo.l1dSize = cacheSize
		} else if cacheLevel == 2 {
			cacheInfo.l2Size = cacheSize
		} else if cacheLevel == 3 {
			cacheInfo.l3Size = cacheSize
		}
	}
	return cacheInfo
}

func getCacheInfoAMD() CacheInfo {
	cacheInfo := getDefaultCacheInfo()
	if maxExtendedFunctionInformation < 0x80000006 {
		return cacheInfo
	}

	_, _, ecx5, edx5 := cpuid(0x80000005, 0)
	_, _, ecx6, edx6 := cpuid(0x80000006, 0)

	// The size is return in kb, turning into bytes.
	cacheInfo.l1iSize = extractBits(edx5, 24, 31) << 10
	cacheInfo.l1dSize = extractBits(ecx5, 24, 31) << 10

	// Check that L2 cache is present.
	l2Assoc := extractBits(ecx6, 12, 15)
	if l2Assoc != 0 {
		cacheInfo.l2Size = extractBits(ecx6, 16, 31) << 10
	}

	// Check that L3 cache is present.
	l3Assoc := extractBits(edx6, 12, 15)
	if l3Assoc != 0 {
		// Specifies the L3 cache size is within the following range:
		// (L3Size[31:18] * 512KB) <= L3 cache size < ((L3Size[31:18]+1) * 512KB).
		cacheInfo.l3Size = extractBits(edx6, 18, 31) * (512 << 10)
	}

	return cacheInfo
}

func getCacheInfo() CacheInfo {
	cacheInfo := getDefaultCacheInfo()

	switch cpuVendor {
	case Intel:
		cacheInfo = getCacheInfoIntel()
	case AMD:
		cacheInfo = getCacheInfoAMD()
	}
	return cacheInfo
}
