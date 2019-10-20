// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips64 mips64le

package cpu

const CacheLinePadSize = 32

// These are initialized by archauxv in runtime/os_linux_mips64x.go.
// These should not be changed after they are initialized.
var HWCap uint
var HWCap2 uint

// HWCAP bits. These are exposed by the Linux kernel 5.4.
const (
	// ISA Level
	hwcap_MIPS_R6 = 1 << 0

	// CPU features
	hwcap_MIPS_MSA   = 1 << 1
	hwcap_MIPS_CRC32 = 1 << 2
)

func doinit() {
	options = []option{
		{Name: "r6", Feature: &MIPS64X.IsR6},
		{Name: "msa", Feature: &MIPS64X.HasMSA},
		{Name: "crc32", Feature: &MIPS64X.HasCRC32},
	}

	// HWCAP feature bits
	MIPS64X.IsR6 = isSet(HWCap, hwcap_MIPS_R6)
	MIPS64X.HasMSA = isSet(HWCap, hwcap_MIPS_MSA)
	MIPS64X.HasCRC32 = isSet(HWCap, hwcap_MIPS_CRC32)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
