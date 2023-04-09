// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build loong64

package cpu

// CacheLinePadSize is used to prevent false sharing of cache lines.
// We choose 64 because Loongson 3A5000 the L1 Dcache is 4-way 256-line 64-byte-per-line.
const CacheLinePadSize = 64

func doinit() {}

func read_cpucfg(reg uint32) uint32

const (
	LOONG64_CPUCFG_REG0  = 0
	LOONG64_CPUCFG_REG1  = 1
	LOONG64_CPUCFG_REG2  = 2
	LOONG64_CPUCFG_REG3  = 3
	LOONG64_CPUCFG_REG4  = 4
	LOONG64_CPUCFG_REG5  = 5
	LOONG64_CPUCFG_REG6  = 6
	LOONG64_CPUCFG_REG7  = 7
	LOONG64_CPUCFG_REG8  = 8
	LOONG64_CPUCFG_REG9  = 9
	LOONG64_CPUCFG_REG10 = 10
	LOONG64_CPUCFG_REG11 = 11
	LOONG64_CPUCFG_REG12 = 12
	LOONG64_CPUCFG_REG13 = 13
)

func GetPrid() uint32 {
	return read_cpucfg(LOONG64_CPUCFG_REG0)
}

// Description of the prid field in the linux arch/loongarch/include/asm/cpu.h
//
// As described in LoongArch specs from Loongson Technology, the PRID
// register (CPUCFG.00) has the following layout:
// +---------------+----------------+------------+--------------------+
// | Reserved      | Company ID     | Series ID  |  Product ID        |
// +---------------+----------------+------------+--------------------+
//
// 31            24 23            16 15        12 11                  0
const (
	PRID_COMP_MASK     uint32 = 0x00ff0000
	PRID_COMP_LOONGSON        = 0x00140000

	PRID_SERIES_MASK  = 0x0000f000
	PRID_SERIES_LA132 = 0x00008000 // Loongson 32bit
	PRID_SERIES_LA264 = 0x0000a000 // Loongson 64bit, 2-issue
	PRID_SERIES_LA364 = 0x0000b000 // Loongson 64bit，3-issue
	PRID_SERIES_LA464 = 0x0000c000 // Loongson 64bit, 4-issue
	PRID_SERIES_LA664 = 0x0000d000 // Loongson 64bit, 6-issue

	PRID_PRODUCT_MASK = 0x00000fff
)

// Name returns the CPU name given by the vendor. If the CPU name
// can not be determined an empty string is returned.
func Name() string {
	prid := GetPrid()

	switch prid & PRID_COMP_MASK {
	case PRID_COMP_LOONGSON:
		switch prid & PRID_SERIES_MASK {
		case PRID_SERIES_LA132:
			return "32-bit Loongson Processor (LA132 Core)"
		case PRID_SERIES_LA264:
			return "32-bit Loongson Processor (LA264 Core)"
		case PRID_SERIES_LA364:
			return "64-bit Loongson Processor (LA364 Core)"
		case PRID_SERIES_LA464:
			return "64-bit Loongson Processor (LA464 Core)"
		case PRID_SERIES_LA664:
			return "64-bit Loongson Processor (LA664 Core)"
		}
	default:
	}

	return ""
}
