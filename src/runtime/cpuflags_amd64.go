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
var vendorId int

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

const (
	cvUnknown = iota
	cvAMD
	cvCentaur
	cvCyrix
	cvIntel
	cvTransmeta
	cvNationalSemiconductor
	cvNexgen
	cvRise
	cvSIS
	cvUMC
	cvVIA
	cvVortex
	cvKVM
	cvHyperV
	cvVMware
	cvXen
)

func init() {
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
		(vendorId == cvIntel)

	// Allow to use memmove optimized for Haswell
	// if we have 4th generation of Intel Core processors.
	// AVX should be available and supported by OS.
	if core4thGeneration && hasFeature(cfAVX) && enabledAVX {
		useMemmoveHaswell = true
	}
}

const (
	cfSSE3 = uint64(1) << iota
	cfPCLMULQDQ
	cfDTES64
	cfMONITOR
	cfDSI_CPL
	cfVMX
	cfSMX
	cfEST
	cfTM2
	cfSSSE3
	cfCNXT_ID
	cfSDBG
	cfFMA
	cfCX16
	cfXTPR
	cfPDCM
	_
	cfPCID
	cfDCA
	cfSSE4_1
	cfSSE4_2
	cfX2APIC
	cfMOVBE
	cfPOPCNT
	cfTSC_DEADLINE
	cfAES
	cfXSAVE
	cfOSXSAVE
	cfAVX
	cfF16C
	cfRDRND
	cfHYPERVISOR
	cfFPU
	cfVME
	cfDE
	cfPSE
	cfTSC
	cfMSR
	cfPAE
	cfMCE
	cfCX8
	cfAPIC
	_
	cfSEP
	cfMTRR
	cfPGE
	cfMCA
	cfCMOV
	cfPAT
	cfPSE_36
	cfPSN
	cfCLFSH
	_
	cfDS
	cfACPI
	cfMMX
	cfFXSR
	cfSSE
	cfSSE2
	cfSS
	cfHTT
	cfTM
	cfIA64
	cfPBE
)

const (
	cfFSGSBASE = uint64(1) << iota
	cfIA32_TSC_ADJUST
	_
	cfBMI1
	cfHLE
	cfAVX2
	_
	cfSMEP
	cfBMI2
	cfERMS
	cfINVPCID
	cfRTM
	cfPQM
	cfDFPUCDS
	cfMPX
	cfPQE
	cfAVX512F
	cfAVX512DQ
	cfRDSEED
	cfADX
	cfSMAP
	cfAVX512IFMA
	cfPCOMMIT
	cfCLFLUSHOPT
	cfCLWB
	cfINTEL_PROCESSOR_TRACE
	cfAVX512PF
	cfAVX512ER
	cfAVX512CD
	cfSHA
	cfAVX512BW
	cfAVX512VL
	// ECX's const from there
	cfPREFETCHWT1
	cfAVX512VBMI
)

const (
	cfLAHF_LM = uint64(1) << iota
	cfCMP_LEGACY
	cfSVM
	cfEXTAPIC
	cfCR8_LEGACY
	cfABM
	cfSSE4A
	cfMISALIGNSSE
	cfPREFETCHW
	cfOSVW
	cfIBS
	cfXOP
	cfSKINIT
	cfWDT
	_
	cfLWP
	cfFMA4
	cfTCE
	_
	cfNODEID_MSR
	_
	cfTBM
	cfTOPOEXT
	cfPERFCTR_CORE
	cfPERFCTR_NB
	cfSPM
	cfDBX
	cfPERFTSC
	cfPCX_L2I
	_
	_
	_
	// EDX features from there
	cfFPU_2
	cfVME_2
	cfDE_2
	cfPSE_2
	cfTSC_2
	cfMSR_2
	cfPAE_2
	cfMCE_2
	cfCX8_2
	cfAPIC_2
	_
	cfSYSCALL
	cfMTRR_2
	cfPGE_2
	cfMCA_2
	cfCMOV_2
	cfPAT_2
	cfPSE36
	_
	cfMP
	cfNX
	_
	cfMMXEXT
	cfMMX_2
	cfFXSR_2
	cfFXSR_OPT
	cfPDPE1GB
	cfRDTSCP
	_
	cfLM
	cf3DNOWEXT
	cf3DNOW
)

func leaf0() {
	eax, ebx, ecx, edx := cpuid_low(0, 0)
	maxInputValue = eax
	int32ToBytes(ebx, vendorStringBytes[0:4])
	int32ToBytes(edx, vendorStringBytes[4:8])
	int32ToBytes(ecx, vendorStringBytes[8:12])
	vendorId = determineVendor()
}

func leaf1() {
	if maxInputValue < 1 {
		return
	}
	_, _, ecx, edx := cpuid_low(1, 0)
	featureFlags = (uint64(edx) << 32) | uint64(ecx)
}

func leaf7() {
	if maxInputValue < 7 {
		return
	}
	_, ebx, ecx, _ := cpuid_low(7, 0)
	extendedFeatureFlags = (uint64(ecx) << 32) | uint64(ebx)
}

func leaf0x80000000() {
	maxExtendedInputValue, _, _, _ = cpuid_low(0x80000000, 0)
}

func leaf0x80000001() {
	if maxExtendedInputValue < 0x80000001 {
		return
	}
	_, _, ecx, edx := cpuid_low(0x80000001, 0)
	extraFeatureFlags = (uint64(edx) << 32) | uint64(ecx)
}

func int32ToBytes(arg uint32, buffer []byte) {
	buffer[3] = byte((arg >> 24) & 0xFF)
	buffer[2] = byte((arg >> 16) & 0xFF)
	buffer[1] = byte((arg >> 8) & 0xFF)
	buffer[0] = byte((arg) & 0xFF)
}

func determineVendor() int {
	switch slicebytetostringtmp(vendorStringBytes[:]) {
	case "AMDisbetter!":
		return cvAMD
	case "AuthenticAMD":
		return cvAMD
	case "CentaurHauls":
		return cvCentaur
	case "CyrixInstead":
		return cvCyrix
	case "GenuineIntel":
		return cvIntel
	case "TransmetaCPU":
		return cvTransmeta
	case "GenuineTMx86":
		return cvTransmeta
	case "Geode by NSC":
		return cvNationalSemiconductor
	case "NexGenDriven":
		return cvNexgen
	case "RiseRiseRise":
		return cvRise
	case "SiS SiS SiS ":
		return cvSIS
	case "UMC UMC UMC ":
		return cvUMC
	case "VIA VIA VIA ":
		return cvVIA
	case "Vortex86 SoC":
		return cvVortex
	case "KVMKVMKVM":
		return cvKVM
	case "Microsoft Hv":
		return cvHyperV
	case "VMwareVMware":
		return cvVMware
	case "XenVMMXenVMM":
		return cvXen
	}
	return cvUnknown
}
