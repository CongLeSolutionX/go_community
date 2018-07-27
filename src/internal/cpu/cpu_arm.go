// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

const CacheLineSize = 32

// arm doesn't have a 'cpuid' equivalent, so we rely on HWCAP/HWCAP2.
// These are linknamed in runtime/os_linux_arm.go and are initialized by
// archauxv().
var hwcap uint
var hwcap2 uint

// HWCAP/HWCAP2 bits. These are exposed by Linux.
const (
	hwcap_SWP   = (1 << 0)
	hwcap_HALF  = (1 << 1)
	hwcap_THUMB = (1 << 2)
	// Go does not support ARMv4 and lower.
	// hwcap_26BIT = (1 << 3)

	hwcap_FASTMULT = (1 << 4)
	hwcap_FPA      = (1 << 5)
	hwcap_VFP      = (1 << 6)
	hwcap_EDSP     = (1 << 7)
	hwcap_JAVA     = (1 << 8)
	hwcap_IWMMXT   = (1 << 9)
	hwcap_CRUNCH   = (1 << 10)
	hwcap_THUMBEE  = (1 << 11)
	hwcap_NEON     = (1 << 12)
	hwcap_VFPv3    = (1 << 13)
	hwcap_VFPv3D16 = (1 << 14)
	hwcap_TLS      = (1 << 15)
	hwcap_VFPv4    = (1 << 16)
	hwcap_IDIVA    = (1 << 17)
	hwcap_IDIVT    = (1 << 18)
	hwcap_VFPD32   = (1 << 19)
	hwcap_IDIV     = (hwcap_IDIVA | hwcap_IDIVT)
	hwcap_LPAE     = (1 << 20)
	hwcap_EVTSTRM  = (1 << 21)
	hwcap2_AES     = (1 << 0)
	hwcap2_PMULL   = (1 << 1)
	hwcap2_SHA1    = (1 << 2)
	hwcap2_SHA2    = (1 << 3)
	hwcap2_CRC32   = (1 << 4)
)

func doinit() {
	options = []option{
		// These capabilities should always be enabled on arm:
		// {"swp", &ARM.HasSWP},
		// {"half", &ARM.HasHALF},

		{"thumb", &ARM.HasTHUMB},
		{"fastmult", &ARM.HasFASTMULT},
		{"fpa", &ARM.HasFPA},
		{"vfp", &ARM.HasVFP},
		{"edsp", &ARM.HasEDSP},
		{"java", &ARM.HasJAVA},
		{"iwmmxt", &ARM.HasIWMMXT},
		{"crunch", &ARM.HasCRUNCH},
		{"thumbee", &ARM.HasTHUMBEE},
		{"neon", &ARM.HasNEON},
		{"vfpv3", &ARM.HasVFPv3},
		{"vfpv3d16", &ARM.HasVFPv3D16},
		{"tls", &ARM.HasTLS},
		{"vfpv4", &ARM.HasVFPv4},
		{"idiva", &ARM.HasIDIVA},
		{"idivt", &ARM.HasIDIVT},
		{"vfpd32", &ARM.HasVFPD32},
		{"idiv", &ARM.HasIDIV},
		{"lpae", &ARM.HasLPAE},
		{"evtstrm", &ARM.HasEVTSTRM},
		{"aes", &ARM.HasAES},
		{"pmull", &ARM.HasPMULL},
		{"sha1", &ARM.HasSHA1},
		{"sha2", &ARM.HasSHA2},
		{"crc32", &ARM.HasCRC32},
	}

	// HWCAP feature bits
	ARM.HasSWP = isSet(hwcap, hwcap_SWP)
	ARM.HasHALF = isSet(hwcap, hwcap_HALF)
	ARM.HasTHUMB = isSet(hwcap, hwcap_THUMB)
	ARM.HasFASTMULT = isSet(hwcap, hwcap_FASTMULT)
	ARM.HasFPA = isSet(hwcap, hwcap_FPA)
	ARM.HasVFP = isSet(hwcap, hwcap_VFP)
	ARM.HasEDSP = isSet(hwcap, hwcap_EDSP)
	ARM.HasJAVA = isSet(hwcap, hwcap_JAVA)
	ARM.HasIWMMXT = isSet(hwcap, hwcap_IWMMXT)
	ARM.HasCRUNCH = isSet(hwcap, hwcap_CRUNCH)
	ARM.HasTHUMBEE = isSet(hwcap, hwcap_THUMBEE)
	ARM.HasNEON = isSet(hwcap, hwcap_NEON)
	ARM.HasVFPv3 = isSet(hwcap, hwcap_VFPv3)
	ARM.HasVFPv3D16 = isSet(hwcap, hwcap_VFPv3D16)
	ARM.HasTLS = isSet(hwcap, hwcap_TLS)
	ARM.HasVFPv4 = isSet(hwcap, hwcap_VFPv4)
	ARM.HasIDIVA = isSet(hwcap, hwcap_IDIVA)
	ARM.HasIDIVT = isSet(hwcap, hwcap_IDIVT)
	ARM.HasVFPD32 = isSet(hwcap, hwcap_VFPD32)
	ARM.HasIDIV = isSet(hwcap, hwcap_IDIV)
	ARM.HasLPAE = isSet(hwcap, hwcap_LPAE)
	ARM.HasEVTSTRM = isSet(hwcap, hwcap_EVTSTRM)
	ARM.HasAES = isSet(hwcap2, hwcap2_AES)
	ARM.HasPMULL = isSet(hwcap2, hwcap2_PMULL)
	ARM.HasSHA1 = isSet(hwcap2, hwcap2_SHA1)
	ARM.HasSHA2 = isSet(hwcap2, hwcap2_SHA2)
	ARM.HasCRC32 = isSet(hwcap2, hwcap2_CRC32)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
