// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

package cpu

const CacheLinePadSize = 128

// ppc64x doesn't have a 'cpuid' equivalent, so we rely on HWCAP/HWCAP2.
// These are initialized by archauxv in runtime/os_linux_ppc64x.go.
// These should not be changed after they are initialized.
var HWCap uint
var HWCap2 uint

// HWCAP/HWCAP2 bits. These are exposed by the kernel.
const (
	// ISA Level
	_PPC_FEATURE2_ARCH_2_07 = 0x80000000
)

func doinit() {
	options = []option{
		{Name: "ispower8", Feature: &PPC64.IsPOWER8, Required: true},
	}

	// HWCAP2 feature bits
	PPC64.IsPOWER8 = isSet(HWCap2, _PPC_FEATURE2_ARCH_2_07)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
