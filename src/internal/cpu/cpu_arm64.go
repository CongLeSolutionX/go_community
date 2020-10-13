// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

const CacheLinePadSize = 64

// arm64 doesn't have a 'cpuid' equivalent, so we rely on HWCAP/HWCAP2.
// These are initialized by archauxv and should not be changed after they are
// initialized.
var HWCap uint
var HWCap2 uint

func doinit() {
	options = []option{
		{Name: "evtstrm", Feature: &ARM64.HasEVTSTRM},
		{Name: "aes", Feature: &ARM64.HasAES},
		{Name: "pmull", Feature: &ARM64.HasPMULL},
		{Name: "sha1", Feature: &ARM64.HasSHA1},
		{Name: "sha2", Feature: &ARM64.HasSHA2},
		{Name: "crc32", Feature: &ARM64.HasCRC32},
		{Name: "atomics", Feature: &ARM64.HasATOMICS},
		{Name: "fphp", Feature: &ARM64.HasFPHP},
		{Name: "asimdhp", Feature: &ARM64.HasASIMDHP},
		{Name: "cpuid", Feature: &ARM64.HasCPUID},
		{Name: "asimdrdm", Feature: &ARM64.HasASIMDRDM},
		{Name: "jscvt", Feature: &ARM64.HasJSCVT},
		{Name: "fcma", Feature: &ARM64.HasFCMA},
		{Name: "lrcpc", Feature: &ARM64.HasLRCPC},
		{Name: "dcpop", Feature: &ARM64.HasDCPOP},
		{Name: "sha3", Feature: &ARM64.HasSHA3},
		{Name: "sm3", Feature: &ARM64.HasSM3},
		{Name: "sm4", Feature: &ARM64.HasSM4},
		{Name: "asimddp", Feature: &ARM64.HasASIMDDP},
		{Name: "sha512", Feature: &ARM64.HasSHA512},
		{Name: "sve", Feature: &ARM64.HasSVE},
		{Name: "asimdfhm", Feature: &ARM64.HasASIMDFHM},

		// These capabilities should always be enabled on arm64:
		{Name: "fp", Feature: &ARM64.HasFP, Required: true},
		{Name: "asimd", Feature: &ARM64.HasASIMD, Required: true},
	}

	// HWCAP feature bits
	ARM64.HasFP = isSet(HWCap, hwcap_FP)
	ARM64.HasASIMD = isSet(HWCap, hwcap_ASIMD)
	ARM64.HasEVTSTRM = isSet(HWCap, hwcap_EVTSTRM)
	ARM64.HasAES = isSet(HWCap, hwcap_AES)
	ARM64.HasPMULL = isSet(HWCap, hwcap_PMULL)
	ARM64.HasSHA1 = isSet(HWCap, hwcap_SHA1)
	ARM64.HasSHA2 = isSet(HWCap, hwcap_SHA2)
	ARM64.HasCRC32 = isSet(HWCap, hwcap_CRC32)
	ARM64.HasATOMICS = isSet(HWCap, hwcap_ATOMICS)
	ARM64.HasFPHP = isSet(HWCap, hwcap_FPHP)
	ARM64.HasASIMDHP = isSet(HWCap, hwcap_ASIMDHP)
	ARM64.HasCPUID = isSet(HWCap, hwcap_CPUID)
	ARM64.HasASIMDRDM = isSet(HWCap, hwcap_ASIMDRDM)
	ARM64.HasJSCVT = isSet(HWCap, hwcap_JSCVT)
	ARM64.HasFCMA = isSet(HWCap, hwcap_FCMA)
	ARM64.HasLRCPC = isSet(HWCap, hwcap_LRCPC)
	ARM64.HasDCPOP = isSet(HWCap, hwcap_DCPOP)
	ARM64.HasSHA3 = isSet(HWCap, hwcap_SHA3)
	ARM64.HasSM3 = isSet(HWCap, hwcap_SM3)
	ARM64.HasSM4 = isSet(HWCap, hwcap_SM4)
	ARM64.HasASIMDDP = isSet(HWCap, hwcap_ASIMDDP)
	ARM64.HasSHA512 = isSet(HWCap, hwcap_SHA512)
	ARM64.HasSVE = isSet(HWCap, hwcap_SVE)
	ARM64.HasASIMDFHM = isSet(HWCap, hwcap_ASIMDFHM)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
