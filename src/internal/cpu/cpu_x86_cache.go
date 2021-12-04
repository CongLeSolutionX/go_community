// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64

package cpu

const (
	defaultL1ISize = 32 << 10
	defaultL1DSize = 32 << 10
	defaultL2Size  = 256 << 10
	defaultL3Size  = 16 << 20

	intelCacheTypeNull  = 0
	intelCacheTypeInstr = 1
	intelCacheTypeData  = 2
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

func getCacheInfoIntel() CacheInfo {
	cacheInfo := getDefaultCacheInfo()
	if maxFunctionInformation < 4 {
		return cacheInfo
	}

	for nxtecx := uint32(0); ; nxtecx++ {
		eax4, ebx4, ecx4, _ := cpuid(4, nxtecx)
		cacheType := eax4 & 0x1f
		if cacheType == intelCacheTypeNull {
			return cacheInfo
		}

		cacheLevel := (eax4 >> 5) & 0x7
		cacheSize := (extractBits(ebx4, 22, 31) + 1) * (extractBits(ebx4, 12, 21) + 1) * (extractBits(ebx4, 0, 11) + 1) * (ecx4 + 1)
		if cacheLevel == 1 && cacheType == intelCacheTypeInstr {
			cacheInfo.l1iSize = cacheSize
		} else if cacheLevel == 1 && cacheType == intelCacheTypeData {
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
	cacheInfo.l1iSize = (edx5 >> 14) & 0x3fc00
	cacheInfo.l1dSize = (ecx5 >> 14) & 0x3fc00

	_, _, ecx6, edx6 := cpuid(0x80000006, 0)
	if (ecx6 & 0xf000) != 0 {
		cacheInfo.l2Size = (ecx6 >> 6) & 0x3fffc00
	}
	if (edx6 & 0xf000) != 0 {
		cacheInfo.l3Size = (edx6 & 0x3ffc0000) << 1
	}

	return cacheInfo
}

func getCacheInfo() CacheInfo {
	cacheInfo := getDefaultCacheInfo()

	switch vendor {
	case vendorIntel:
		cacheInfo = getCacheInfoIntel()
	case vendorAMD:
		cacheInfo = getCacheInfoAMD()
	}
	return cacheInfo
}
