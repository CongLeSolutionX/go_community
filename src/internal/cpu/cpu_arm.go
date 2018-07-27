// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

const CacheLineSize = 32

// arm doesn't have a 'cpuid' equivalent, so we rely on HWCAP/HWCAP2.
// These are linknamed in runtime/os_(linux|freebsd)_arm.go and are
// initialized by archauxv().
// These should not be changed after they are initialized.
var HWCap uint
var HWCap2 uint

// HWCAP/HWCAP2 bits. These are exposed by Linux and FreeBSD.
const (
	hwcap_NEON   = 1 << 12
	hwcap_VFPv4  = 1 << 16
	hwcap_IDIVA  = 1 << 17
	hwcap_VFPD32 = 1 << 19
)

func doinit() {
	options = []option{
		{"neon", &ARM.HasNEON},
		{"vfpv4", &ARM.HasVFPv4},
		{"idiva", &ARM.HasIDIVA},
		{"vfpd32", &ARM.HasVFPD32},
	}

	// HWCAP feature bits
	ARM.HasNEON = isSet(hwcap, hwcap_NEON)
	ARM.HasVFPv4 = isSet(hwcap, hwcap_VFPv4)
	ARM.HasIDIVA = isSet(HWCap, hwcap_IDIVA)
	ARM.HasVFPD32 = isSet(hwcap, hwcap_VFPD32)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
