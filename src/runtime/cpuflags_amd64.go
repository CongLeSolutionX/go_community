// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

var vendorStringBytes [12]byte
var maxInputValue uint32
var maxExtendedInputValue uint32
var featureFlags uint64
var extendedFeatureFlags uint64
var extraFeatureFlags uint64

var useMemmoveHaswell bool

func hasFeature(feature uint64) bool {
	return (featureFlags & feature) != 0
}

func hasExtendedFeature(feature uint64) bool {
	return (extendedFeatureFlags & feature) != 0
}

func hasExtraFeature(feature uint64) bool {
	return (extraFeatureFlags & feature) != 0
}

func cpuid_low(arg1, arg2 uint32) (eax, ebx, ecx, edx uint32) // implemented in cpuidlow_amd64.s
func xgetbv_low(arg1 uint32) (eax, edx uint32)                // implemented in cpuidlow_amd64.s

func init() {
	var cfOSXSAVE uint64 = 1 << 27
	var cfFMA uint64 = 1 << 12
	var cfMOVBE uint64 = 1 << 22
	var cfAVX2 uint64 = 1 << 5
	var cfAVX uint64 = 1 << 28
	var cfBMI1 uint64 = 1 << 3
	var cfBMI2 uint64 = 1 << 8
	var cfABM uint64 = 1 << 5

	leaf0()
	leaf1()
	leaf7()
	leaf0x80000000()
	leaf0x80000001()

	enabledAVX := false
	// Let's check if OS has set CR4.OSXSAVE[bit 18]
	// to enable XGETBV instruction.
	if hasFeature(cfOSXSAVE) {
		eax, _ := xgetbv_low(0)
		// Let's check that XCR0[2:1] = ‘11b’
		// i.e. XMM state and YMM state are enabled by OS.
		if (eax & 0x6) == 0x6 {
			enabledAVX = true
		}
	}

	// The easiest way to detect 4th generation of Intel Core processor
	// is to check for the new features it supports
	// so let's check for FMA, MOVBE, AVX2, and
	// bit manipulations: BMI1, BMI2, ABM (LZCNT).
	core4thGeneration := hasFeature(cfFMA) &&
		hasFeature(cfMOVBE) &&
		hasExtendedFeature(cfAVX2) &&
		hasExtendedFeature(cfBMI1) &&
		hasExtendedFeature(cfBMI2) &&
		hasExtraFeature(cfABM) &&
		isIntel()

	// Allow to use memmove optimized for Haswell
	// if we have 4th generation of Intel Core processors.
	// AVX should be available and supported by OS.
	if core4thGeneration && hasFeature(cfAVX) && enabledAVX {
		useMemmoveHaswell = true
	}
}

func leaf0() {
	eax, ebx, ecx, edx := cpuid_low(0, 0)
	maxInputValue = eax
	int32ToBytes(ebx, vendorStringBytes[0:4])
	int32ToBytes(edx, vendorStringBytes[4:8])
	int32ToBytes(ecx, vendorStringBytes[8:12])
}

func leaf1() {
	if maxInputValue < 1 {
		return
	}
	_, _, ecx, edx := cpuid_low(1, 0)
	featureFlags = uint64(edx)<<32 | uint64(ecx)
}

func leaf7() {
	if maxInputValue < 7 {
		return
	}
	_, ebx, ecx, _ := cpuid_low(7, 0)
	extendedFeatureFlags = uint64(ecx)<<32 | uint64(ebx)
}

func leaf0x80000000() {
	maxExtendedInputValue, _, _, _ = cpuid_low(0x80000000, 0)
}

func leaf0x80000001() {
	if maxExtendedInputValue < 0x80000001 {
		return
	}
	_, _, ecx, edx := cpuid_low(0x80000001, 0)
	extraFeatureFlags = uint64(edx)<<32 | uint64(ecx)
}

func int32ToBytes(arg uint32, buffer []byte) {
	buffer[3] = byte(arg >> 24)
	buffer[2] = byte(arg >> 16)
	buffer[1] = byte(arg >> 8)
	buffer[0] = byte(arg)
}

func isIntel() bool {
	if slicebytetostringtmp(vendorStringBytes[:]) == "GenuineIntel" {
		return true
	}
	return false
}
