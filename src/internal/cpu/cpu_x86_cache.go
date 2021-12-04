// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64

package cpu

const (
	l1IcacheDescType = iota
	l1DcacheDescType
	l2CacheDescType
	l3CacheDescType
	defaultL3Size = 16 << 20

	intelCacheTypeNull  = 0
	intelCacheTypeInstr = 1
	intelCacheTypeData  = 2
)

type CacheData struct {
	l1iSize uint32
	l1dSize uint32
	l2Size  uint32
	l3Size  uint32
}

type IntelCpuidLeaf2Desc struct {
	idx      int
	typ      int
	assoc    int
	linesize int
	size     uint32
}

// From "Table 3-12. Encoding of CPUID Leaf 2 Descriptors" in Intel Manual Vol.2.
var intelDescs = []IntelCpuidLeaf2Desc{
	{0x06, l1IcacheDescType, 4, 32, 8192},
	{0x08, l1IcacheDescType, 4, 32, 16384},
	{0x09, l1IcacheDescType, 4, 32, 32768},
	{0x0a, l1DcacheDescType, 2, 32, 8192},
	{0x0c, l1DcacheDescType, 4, 32, 16384},
	{0x0d, l1DcacheDescType, 4, 64, 16384},
	{0x0e, l1DcacheDescType, 6, 64, 24576},
	{0x21, l2CacheDescType, 8, 64, 262144},
	{0x22, l3CacheDescType, 4, 64, 524288},
	{0x23, l3CacheDescType, 8, 64, 1048576},
	{0x25, l3CacheDescType, 8, 64, 2097152},
	{0x29, l3CacheDescType, 8, 64, 4194304},
	{0x2c, l1DcacheDescType, 8, 64, 32768},
	{0x30, l1IcacheDescType, 8, 64, 32768},
	{0x39, l2CacheDescType, 4, 64, 131072},
	{0x3a, l2CacheDescType, 6, 64, 196608},
	{0x3b, l2CacheDescType, 2, 64, 131072},
	{0x3c, l2CacheDescType, 4, 64, 262144},
	{0x3d, l2CacheDescType, 6, 64, 393216},
	{0x3e, l2CacheDescType, 4, 64, 524288},
	{0x3f, l2CacheDescType, 2, 64, 262144},
	{0x41, l2CacheDescType, 4, 32, 131072},
	{0x42, l2CacheDescType, 4, 32, 262144},
	{0x43, l2CacheDescType, 4, 32, 524288},
	{0x44, l2CacheDescType, 4, 32, 1048576},
	{0x45, l2CacheDescType, 4, 32, 2097152},
	{0x46, l3CacheDescType, 4, 64, 4194304},
	{0x47, l3CacheDescType, 8, 64, 8388608},
	{0x48, l2CacheDescType, 12, 64, 3145728},
	{0x49, l2CacheDescType, 16, 64, 4194304},
	{0x4a, l3CacheDescType, 12, 64, 6291456},
	{0x4b, l3CacheDescType, 16, 64, 8388608},
	{0x4c, l3CacheDescType, 12, 64, 12582912},
	{0x4d, l3CacheDescType, 16, 64, 16777216},
	{0x4e, l2CacheDescType, 24, 64, 6291456},
	{0x60, l1DcacheDescType, 8, 64, 16384},
	{0x66, l1DcacheDescType, 4, 64, 8192},
	{0x67, l1DcacheDescType, 4, 64, 16384},
	{0x68, l1DcacheDescType, 4, 64, 32768},
	{0x78, l2CacheDescType, 8, 64, 1048576},
	{0x79, l2CacheDescType, 8, 64, 131072},
	{0x7a, l2CacheDescType, 8, 64, 262144},
	{0x7b, l2CacheDescType, 8, 64, 524288},
	{0x7c, l2CacheDescType, 8, 64, 1048576},
	{0x7d, l2CacheDescType, 8, 64, 2097152},
	{0x7f, l2CacheDescType, 2, 64, 524288},
	{0x80, l2CacheDescType, 8, 64, 524288},
	{0x82, l2CacheDescType, 8, 32, 262144},
	{0x83, l2CacheDescType, 8, 32, 524288},
	{0x84, l2CacheDescType, 8, 32, 1048576},
	{0x85, l2CacheDescType, 8, 32, 2097152},
	{0x86, l2CacheDescType, 4, 64, 524288},
	{0x87, l2CacheDescType, 8, 64, 1048576},
	{0xd0, l3CacheDescType, 4, 64, 524288},
	{0xd1, l3CacheDescType, 4, 64, 1048576},
	{0xd2, l3CacheDescType, 4, 64, 2097152},
	{0xd6, l3CacheDescType, 8, 64, 1048576},
	{0xd7, l3CacheDescType, 8, 64, 2097152},
	{0xd8, l3CacheDescType, 8, 64, 4194304},
	{0xdc, l3CacheDescType, 12, 64, 2097152},
	{0xdd, l3CacheDescType, 12, 64, 4194304},
	{0xde, l3CacheDescType, 12, 64, 8388608},
	{0xe2, l3CacheDescType, 16, 64, 2097152},
	{0xe3, l3CacheDescType, 16, 64, 4194304},
	{0xe4, l3CacheDescType, 16, 64, 8388608},
	{0xea, l3CacheDescType, 24, 64, 12582912},
	{0xeb, l3CacheDescType, 24, 64, 18874368},
	{0xec, l3CacheDescType, 24, 64, 25165824},
}

func extractBits(arg uint32, l int, r int) uint32 {
	if l > r {
		return 0
	}
	return (arg >> l) & ((1 << (r - l + 1)) - 1)
}

func parseIntelDescs(value uint32, cacheData *CacheData) {
	if (value & 0x80000000) != 0 {
		return
	}

	for value != 0 {
		descIdx := int(value) & 0xff
		switch descIdx {
		case 0xff:
			for nxtecx := uint32(0); ; nxtecx++ {
				eax4, ebx4, ecx4, _ := cpuid(4, nxtecx)
				cacheType := eax4 & 0x1f
				if cacheType == intelCacheTypeNull {
					return
				}

				cacheLevel := (eax4 >> 5) & 0x7
				cacheSize := (extractBits(ebx4, 22, 31) + 1) * (extractBits(ebx4, 12, 21) + 1) * (extractBits(ebx4, 0, 11) + 1) * (ecx4 + 1)
				if cacheLevel == 1 && cacheType == intelCacheTypeInstr {
					cacheData.l1iSize = cacheSize
				} else if cacheLevel == 1 && cacheType == intelCacheTypeData {
					cacheData.l1dSize = cacheSize
				} else if cacheLevel == 2 {
					cacheData.l2Size = cacheSize
				} else if cacheLevel == 3 {
					cacheData.l3Size = cacheSize
				}
			}
		default:
			l, r := 0, len(intelDescs)
			for l < r {
				m := int(uint(l+r) >> 1)
				if intelDescs[m].idx < descIdx {
					l = m + 1
				} else {
					r = m
				}
			}
			if l < len(intelDescs) && intelDescs[l].idx == descIdx {
				desc := intelDescs[l]

				// Intel reuses this value for a cpu 15'6 to describe L3 cache.
				// Just pretending this is an L3 descriptor.
				// See "Table 3-12. Encoding of CPUID Leaf 2 Descriptors" in Intel Manual Vol.2.
				if desc.idx == 0x49 && family == 15 && model == 6 {
					desc.typ = l3CacheDescType
				}

				switch desc.typ {
				case l1IcacheDescType:
					cacheData.l1iSize = desc.size
				case l1DcacheDescType:
					cacheData.l1dSize = desc.size
				case l2CacheDescType:
					cacheData.l2Size = desc.size
				case l3CacheDescType:
					cacheData.l3Size = desc.size
				}
			}
		}

		value >>= 8
	}
}

func getCacheSizeIntel() CacheData {
	cacheData := CacheData{0, 0, 0, 0}
	rounds := 1
	for i := 0; i < rounds; i++ {
		eax2, ebx2, ecx2, edx2 := cpuid(2, 0)

		// The least-significant byte in register EAX (register AL) indicates the number of times the
		// CPUID instruction must be executed with an input value of 2 to get a complete description
		// of the processor's caches and TLBs.
		if rounds == 1 {
			rounds = int(eax2) & 0xff
			eax2 &= 0xffffff00
		}

		parseIntelDescs(eax2, &cacheData)
		parseIntelDescs(ebx2, &cacheData)
		parseIntelDescs(ecx2, &cacheData)
		parseIntelDescs(edx2, &cacheData)
	}

	return cacheData
}

func getCacheSizeAMD() CacheData {
	cacheData := CacheData{0, 0, 0, 0}
	if maxExtendedFunctionInformation < 0x80000006 {
		return cacheData
	}

	_, _, ecx5, edx5 := cpuid(0x80000005, 0)
	cacheData.l1iSize = (edx5 >> 14) & 0x3fc00
	cacheData.l1dSize = (ecx5 >> 14) & 0x3fc00

	_, _, ecx6, edx6 := cpuid(0x80000006, 0)
	if (ecx6 & 0xf000) != 0 {
		cacheData.l2Size = (ecx6 >> 6) & 0x3fffc00
	}
	if (edx6 & 0xf000) != 0 {
		cacheData.l3Size = (edx6 & 0x3ffc0000) << 1
	}

	return cacheData
}

func getLLCSize() uint32 {
	cacheData := CacheData{0, 0, 0, 0}

	switch vendor {
	case vendorIntel:
		cacheData = getCacheSizeIntel()
	case vendorAMD:
		cacheData = getCacheSizeAMD()
	}

	if cacheData.l3Size == 0 {
		return defaultL3Size
	}
	return cacheData.l3Size
}
