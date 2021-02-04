// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// The file contains the arm64 instruction table, which is derived from instFormats
// of https://github.com/golang/arch/blob/master/arm64/arm64asm/tables.go. In the
// future, we'd better make these two tables consistent.

type instArgs [5]argtype

// An Optab describes the format of an instruction encoding.
type Optab struct {
	skeleton uint32 // The known bits of the instruction.
	as       string
	// args describe how to decode the instruction arguments.
	// args is stored as a fixed-size array.
	// If there are fewer than len(args) arguments, args[i] == 0 marks
	// the end of the argument list.
	args instArgs
}

// Explicit optab index, the format is: uppercase arm64 instruction name + lowercase suffix.
// The suffix is to distinguish different formats of the same instruction.
// A rough but not necessarily accurate agreement of the suffix:
// "w": w register.
// "x": x register.
// "r": w or x register.
// "b": 8-bit floating point register.
// "h": 16-bit floating point register.
// "s": 32-bit floating point register or shift.
// "d": 64-bit floating point register.
// "q": 128-bit floating point register.
// "v": v register.
// "sp": sp register.
// "i": immediate number or vector register index.
// "e": extended register format.
// "p": post-index.
// "w": pre-index.
// "l": label.
// "t": vector register arrangement specifier.
// "1","2","3","4": 1, 2, 3 or 4 vetor register.
// The name doesn't matter, just make sure it is distinguishable.
const (
	INull uint16 = iota
	ADCwww
	ADCxxx
	ADCSwww
	ADCSxxx
	ADDwwwe
	ADDxxre
	ADDwwws
	ADDxxxs
	ADDwwis
	ADDxxis
	ADDSxxre
	ADDSwwwe
	ADDSwwis
	ADDSxxis
	ADDSwwws
	ADDSxxxs
	ADRxl
	ADRPxl
	ANDxxxs
	ANDwwws
	ANDwwi
	ANDxxi
	ANDSwwi
	ANDSxxi
	ANDSxxxs
	ANDSwwws
	ASRxxx
	ASRwww
	ASRwwi
	ASRxxi
	ASRVwww
	ASRVxxx
	ATx
	Bl
	Bcl
	BFIxxii
	BFIwwii
	BFMwwii
	BFMxxii
	BFXILwwii
	BFXILxxii
	BICxxxs
	BICwwws
	BICSxxxs
	BICSwwws
	BLl
	BLRx
	BRx
	BRKi
	CASxxx
	CASAxxx
	CASALxxx
	CASLxxx
	CASwwx
	CASAwwx
	CASALwwx
	CASLwwx
	CASHwwx
	CASAHwwx
	CASALHwwx
	CASLHwwx
	CASBwwx
	CASABwwx
	CASALBwwx
	CASLBwwx
	CASPxxx
	CASPAxxx
	CASPALxxx
	CASPLxxx
	CASPwwx
	CASPAwwx
	CASPALwwx
	CASPLwwx
	CBNZxl
	CBNZwl
	CBZxl
	CBZwl
	CCMNwiic
	CCMNwwic
	CCMNxiic
	CCMNxxic
	CCMPwwic
	CCMPxxic
	CCMPwiic
	CCMPxiic
	CINCxxc
	CINCwwc
	CINVxxc
	CINVwwc
	CLREXi
	CLSxx
	CLSww
	CLZxx
	CLZww
	CMNxre
	CMNwwe
	CMNwis
	CMNxis
	CMNxxs
	CMNwws
	CMPxxs
	CMPwws
	CMPxis
	CMPwis
	CMPxre
	CMPwwe
	CNEGxxc
	CNEGwwc
	CRC32Bwww
	CRC32CBwww
	CRC32CHwww
	CRC32CWwww
	CRC32CXwwx
	CRC32Hwww
	CRC32Wwww
	CRC32Xwwx
	CSELxxxc
	CSELwwwc
	CSETxc
	CSETwc
	CSETMwc
	CSETMxc
	CSINCxxxc
	CSINCwwwc
	CSINVxxxc
	CSINVwwwc
	CSNEGxxxc
	CSNEGwwwc
	DCx
	DCPS1i
	DCPS2i
	DCPS3i
	DMBi
	DRPS
	DSBi
	EONxxxs
	EONwwws
	EORwwi
	EORxxi
	EORwwws
	EORxxxs
	ERET
	EXTRxxxi
	EXTRwwwi
	HINTi
	HLTi
	HVCi
	ICix
	ISBi
	LDADDxxx
	LDADDAxxx
	LDADDALxxx
	LDADDLxxx
	LDADDwwx
	LDADDAwwx
	LDADDALwwx
	LDADDLwwx
	LDADDHwwx
	LDADDAHwwx
	LDADDALHwwx
	LDADDLHwwx
	LDADDBwwx
	LDADDABwwx
	LDADDALBwwx
	LDADDLBwwx
	LDARxx
	LDARwx
	LDARBwx
	LDARHwx
	LDAXPwwx
	LDAXPxxx
	LDAXRxx
	LDAXRwx
	LDAXRBwx
	LDAXRHwx
	LDCLRxxx
	LDCLRAxxx
	LDCLRALxxx
	LDCLRLxxx
	LDCLRwwx
	LDCLRAwwx
	LDCLRALwwx
	LDCLRLwwx
	LDCLRHwwx
	LDCLRAHwwx
	LDCLRALHwwx
	LDCLRLHwwx
	LDCLRBwwx
	LDCLRABwwx
	LDCLRALBwwx
	LDCLRLBwwx
	LDEORxxx
	LDEORAxxx
	LDEORALxxx
	LDEORLxxx
	LDEORwwx
	LDEORAwwx
	LDEORALwwx
	LDEORLwwx
	LDEORHwwx
	LDEORAHwwx
	LDEORALHwwx
	LDEORLHwwx
	LDEORBwwx
	LDEORABwwx
	LDEORALBwwx
	LDEORLBwwx
	LDNPxxx
	LDNPwwx
	LDPwwxi_p
	LDPxxxi_p
	LDPwwx_w
	LDPxxx_w
	LDPxxx
	LDPwwx
	LDPSWxxxi_p
	LDPSWxxx_w
	LDPSWxxx
	LDRwxi_p
	LDRxxi_p
	LDRwx_w
	LDRxx_w
	LDRwx
	LDRxx
	LDRwl
	LDRxl
	LDRwxre
	LDRxxre
	LDRBwx
	LDRBwxre
	LDRBwx_w
	LDRBwxi_p
	LDRHwx
	LDRHwxre
	LDRHwxi_p
	LDRHwx_w
	LDRSBwxi_p
	LDRSBxxi_p
	LDRSBxx_w
	LDRSBwx_w
	LDRSBwx
	LDRSBxx
	LDRSBwxre
	LDRSBxxre
	LDRSHxxi_p
	LDRSHwxi_p
	LDRSHxx_w
	LDRSHwx_w
	LDRSHwxre
	LDRSHxxre
	LDRSHwx
	LDRSHxx
	LDRSWxl
	LDRSWxxre
	LDRSWxxi_p
	LDRSWxx_w
	LDRSWxx
	LDSETxxx
	LDSETAxxx
	LDSETALxxx
	LDSETLxxx
	LDSETwwx
	LDSETAwwx
	LDSETALwwx
	LDSETLwwx
	LDSETHwwx
	LDSETAHwwx
	LDSETALHwwx
	LDSETLHwwx
	LDSETBwwx
	LDSETABwwx
	LDSETALBwwx
	LDSETLBwwx
	LDTRwx
	LDTRxx
	LDTRBwx
	LDTRHwx
	LDTRSBwx
	LDTRSBxx
	LDTRSHxx
	LDTRSHwx
	LDTRSWxx
	LDURwx
	LDURxx
	LDURBwx
	LDURHwx
	LDURSBwx
	LDURSBxx
	LDURSHwx
	LDURSHxx
	LDURSWxx
	LDXPwwx
	LDXPxxx
	LDXRwx
	LDXRxx
	LDXRBwx
	LDXRHwx
	LSLwwi
	LSLxxi
	LSLwww
	LSLxxx
	LSLVwww
	LSLVxxx
	LSRwww
	LSRxxx
	LSRxxi
	LSRwwi
	LSRVwww
	LSRVxxx
	MADDxxxx
	MADDwwww
	MNEGxxx
	MNEGwww
	MOVwi_b
	MOVxi_b
	MOVwi_n
	MOVxi_n
	MOVww
	MOVxx
	MOVww_sp
	MOVxx_sp
	MOVwi_z
	MOVxi_z
	MOVKwis
	MOVKxis
	MOVNwis
	MOVNxis
	MOVZwis
	MOVZxis
	MRSx
	MSRi
	MSRx
	MSUBwwww
	MSUBxxxx
	MULxxx
	MULwww
	MVNwws
	MVNxxs
	NEGwws
	NEGxxs
	NEGSxxs
	NEGSwws
	NGCxx
	NGCww
	NGCSxx
	NGCSww
	NOP
	ORNwwws
	ORNxxxs
	ORRwwws
	ORRxxxs
	ORRxxi
	ORRwwi
	PRFMix
	PRFMixre
	PRFMil
	PRFUMix
	RBITxx
	RBITww
	RETx
	REVww
	REVxx
	REV16ww
	REV16xx
	REV32xx
	RORwwi
	RORxxi
	RORxxx
	RORwww
	RORVxxx
	RORVwww
	SBCxxx
	SBCwww
	SBCSxxx
	SBCSwww
	SBFIZwwii
	SBFIZxxii
	SBFMxxii
	SBFMwwii
	SBFXxxii
	SBFXwwii
	SDIVwww
	SDIVxxx
	SEV
	SEVL
	SMADDLxwwx
	SMCi
	SMNEGLxww
	SMSUBLxwwx
	SMULHxxx
	SMULLxww
	STLRwx
	STLRxx
	STLRBwx
	STLRHwx
	STLXPwxxx
	STLXPwwwx
	STLXRwwx
	STLXRwxx
	STLXRBwwx
	STLXRHwwx
	STNPxxx
	STNPwwx
	STPxxxi_p
	STPwwxi_p
	STPxxx_w
	STPwwx_w
	STPxxx
	STPwwx
	STRwxi_p
	STRxxi_p
	STRwx_w
	STRxx_w
	STRwx
	STRxx
	STRwxre
	STRxxre
	STRBwxi_p
	STRBwx_w
	STRBwx
	STRBwxre
	STRHwxi_p
	STRHwx_w
	STRHwx
	STRHwxre
	STTRxx
	STTRwx
	STTRBwx
	STTRHwx
	STURwx
	STURxx
	STURBwx
	STURHwx
	STXPwwwx
	STXPwxxx
	STXRwxx
	STXRwwx
	STXRBwwx
	STXRHwwx
	SUBxxxs
	SUBwwws
	SUBxxis
	SUBwwis
	SUBxxre
	SUBwwwe
	SUBSxxxs
	SUBSwwws
	SUBSwwis
	SUBSxxis
	SUBSxxre
	SUBSwwwe
	SVCi
	SWPxxx
	SWPAxxx
	SWPALxxx
	SWPLxxx
	SWPwwx
	SWPAwwx
	SWPALwwx
	SWPLwwx
	SWPHwwx
	SWPAHwwx
	SWPALHwwx
	SWPLHwwx
	SWPBwwx
	SWPABwwx
	SWPALBwwx
	SWPLBwwx
	SXTBww
	SXTBxw
	SXTHxw
	SXTHww
	SXTWxw
	SYSix
	SYSLix
	TBNZril
	TBZril
	TLBIix
	TSTxi
	TSTwi
	TSTwws
	TSTxxs
	UBFIZwwii
	UBFIZxxii
	UBFMwwii
	UBFMxxii
	UBFXwwii
	UBFXxxii
	UDFi
	UDIVwww
	UDIVxxx
	UMADDLxwwx
	UMNEGLxww
	UMSUBLxwwx
	UMULHxxx
	UMULLxww
	UXTBww
	UXTHww
	WFE
	WFI
	YIELD
	ABSvv
	ABSvv_t
	ADDvvv
	ADDvvv_t
	ADDHNvvv_t
	ADDHN2vvv_t
	ADDPvv_t
	ADDPvvv_t
	ADDVvv_t
	AESDvv
	AESEvv
	AESIMCvv
	AESMCvv
	ANDvvv_t
	BCAXvvv
	BICvis_h
	BICvis_s
	BICvvv_t
	BIFvvv_t
	BITvvv_t
	BSLvvv_t
	CLSvv_t
	CLZvv_t
	CMEQvvv
	CMEQvvv_t
	CMEQvv_t
	CMEQvv
	CMGEvvv
	CMGEvvv_t
	CMGEvv
	CMGEvv_t
	CMGTvvv
	CMGTvvv_t
	CMGTvv
	CMGTvv_t
	CMHIvvv
	CMHIvvv_t
	CMHSvvv
	CMHSvvv_t
	CMLEvv
	CMLEvv_t
	CMLTvv
	CMLTvv_t
	CMTSTvvv
	CMTSTvvv_t
	CNTvv_t
	DUPvr_t
	DUPvv_i
	DUPvv_ti
	EORvvv_t
	EOR3vvv
	EXTvvvi_t
	FABDvvv
	FABDvvv_t
	FABSvv_t
	FABSss
	FABSdd
	FACGEvvv
	FACGEvvv_t
	FACGTvvv_t
	FACGTvvv
	FADDvvv_t
	FADDsss
	FADDddd
	FADDPvvv_t
	FADDPvv_t
	FCCMPddic
	FCCMPssic
	FCCMPEddic
	FCCMPEssic
	FCMEQvv
	FCMEQvv_t
	FCMEQvvv
	FCMEQvvv_t
	FCMGEvv
	FCMGEvv_t
	FCMGEvvv
	FCMGEvvv_t
	FCMGTvv
	FCMGTvv_t
	FCMGTvvv
	FCMGTvvv_t
	FCMLEvv
	FCMLEvv_t
	FCMLTvv
	FCMLTvv_t
	FCMPss
	FCMPdd
	FCMPd0
	FCMPs0
	FCMPEdd
	FCMPEd0
	FCMPEs0
	FCMPEss
	FCSELsssc
	FCSELdddc
	FCVThd
	FCVTsh
	FCVTdh
	FCVThs
	FCVTds
	FCVTsd
	FCVTASvv_t
	FCVTASws
	FCVTASxs
	FCVTASwd
	FCVTASxd
	FCVTASvv
	FCVTAUvv_t
	FCVTAUws
	FCVTAUxs
	FCVTAUwd
	FCVTAUxd
	FCVTAUvv
	FCVTLvv_t
	FCVTL2vv_t
	FCVTMSvv_t
	FCVTMSws
	FCVTMSxs
	FCVTMSwd
	FCVTMSxd
	FCVTMSvv
	FCVTMUvv_t
	FCVTMUws
	FCVTMUxs
	FCVTMUwd
	FCVTMUxd
	FCVTMUvv
	FCVTNvv_t
	FCVTN2vv_t
	FCVTNSvv_t
	FCVTNSws
	FCVTNSxs
	FCVTNSwd
	FCVTNSxd
	FCVTNSvv
	FCVTNUvv_t
	FCVTNUws
	FCVTNUxs
	FCVTNUwd
	FCVTNUxd
	FCVTNUvv
	FCVTPSvv_t
	FCVTPSws
	FCVTPSxs
	FCVTPSwd
	FCVTPSxd
	FCVTPSvv
	FCVTPUvv_t
	FCVTPUws
	FCVTPUxs
	FCVTPUwd
	FCVTPUxd
	FCVTPUvv
	FCVTXNvv_t
	FCVTXNvv
	FCVTXN2vv_t

	FCVTZSwdi
	FCVTZSwsi
	FCVTZSxdi
	FCVTZSxsi
	FCVTZSwd
	FCVTZSws
	FCVTZSxd
	FCVTZSxs
	FCVTZSvvi
	FCVTZSvvi_t
	FCVTZSvv
	FCVTZSvv_t
	FCVTZUwdi
	FCVTZUwsi
	FCVTZUxdi
	FCVTZUxsi
	FCVTZUwd
	FCVTZUws
	FCVTZUxd
	FCVTZUxs
	FCVTZUvvi
	FCVTZUvvi_t
	FCVTZUvv
	FCVTZUvv_t
	FDIVvvv_t
	FDIVsss
	FDIVddd
	FMADDssss
	FMADDdddd
	FMAXsss
	FMAXddd
	FMAXvvv_t
	FMAXNMvvv_t
	FMAXNMsss
	FMAXNMddd
	FMAXNMPvv_t
	FMAXNMPvvv_t
	FMAXNMVvv_t
	FMAXPvv_t
	FMAXPvvv_t
	FMAXVvv_t
	FMINddd
	FMINsss
	FMINvvv_t
	FMINNMddd
	FMINNMsss
	FMINNMvvv_t
	FMINNMPvv_t
	FMINNMPvvv_t
	FMINNMVvv_t
	FMINPvv_t
	FMINPvvv_t
	FMINVvv_t
	FMLAvvv_i
	FMLAvvv_ti
	FMLAvvv_t
	FMLSvvv_t
	FMLSvvv_i
	FMLSvvv_ti
	FMOVsw
	FMOVws
	FMOVdx
	FMOVxd
	FMOVvx
	FMOVxv
	FMOVss
	FMOVdd
	FMOVsi
	FMOVdi
	FMOVvi_t
	FMOVvi_d
	FMSUBssss
	FMSUBdddd
	FMULddd
	FMULvvv_i
	FMULvvv_ti
	FMULsss
	FMULvvv_t
	FMULXvvv_ti
	FMULXvvv
	FMULXvvv_t
	FMULXvvv_i
	FNEGvv_t
	FNEGss
	FNEGdd
	FNMADDssss
	FNMADDdddd
	FNMSUBssss
	FNMSUBdddd
	FNMULddd
	FNMULsss
	FRECPEvv
	FRECPEvv_t
	FRECPSvvv
	FRECPSvvv_t
	FRECPXvv
	FRINTAdd
	FRINTAss
	FRINTAvv_t
	FRINTIdd
	FRINTIvv_t
	FRINTIss
	FRINTMdd
	FRINTMvv_t
	FRINTMss
	FRINTNdd
	FRINTNss
	FRINTNvv_t
	FRINTPss
	FRINTPdd
	FRINTPvv_t
	FRINTXvv_t
	FRINTXss
	FRINTXdd
	FRINTZss
	FRINTZdd
	FRINTZvv_t
	FRSQRTEvv_t
	FRSQRTEvv
	FRSQRTSvvv_t
	FRSQRTSvvv
	FSQRTss
	FSQRTvv_t
	FSQRTdd
	FSUBsss
	FSUBddd
	FSUBvvv_t
	INSvr_i
	INSvv_i
	LD1vx_t1
	LD1vx_t2
	LD1vx_t3
	LD1vx_t4
	LD1vxi_tp1
	LD1vxi_tp2
	LD1vxi_tp3
	LD1vxi_tp4
	LD1vxx_tp1
	LD1vxx_tp2
	LD1vxx_tp3
	LD1vxx_tp4
	LD1vx_bi1
	LD1vx_hi1
	LD1vx_si1
	LD1vx_di1
	LD1vxi_bip1
	LD1vxi_hip1
	LD1vxi_sip1
	LD1vxi_dip1
	LD1vxx_bip1
	LD1vxx_hip1
	LD1vxx_sip1
	LD1vxx_dip1
	LD1Rvx_t1
	LD1Rvxi_tp1
	LD1Rvxx_tp1
	LD2vx_t2
	LD2vxi_tp2
	LD2vxx_tp2
	LD2vx_bi2
	LD2vx_hi2
	LD2vx_si2
	LD2vx_di2
	LD2vx_bip2
	LD2vxx_bip2
	LD2vx_hip2
	LD2vxx_hip2
	LD2vx_sip2
	LD2vxx_sip2
	LD2vx_dip2
	LD2vxx_dip2
	LD2Rvx_t2
	LD2Rvxi_tp2
	LD2Rvxx_tp2
	LD3vx_t3
	LD3vxi_tp3
	LD3vxx_tp3
	LD3vx_bi3
	LD3vx_hi3
	LD3vx_si3
	LD3vx_di3
	LD3vx_bip3
	LD3vxx_bip3
	LD3vx_hip3
	LD3vxx_hip3
	LD3vx_sip3
	LD3vxx_sip3
	LD3vx_dip3
	LD3vxx_dip3
	LD3Rvx_t3
	LD3Rvxi_tp3
	LD3Rvxx_tp3
	LD4vx_t4
	LD4vxi_tp4
	LD4vxx_tp4
	LD4vx_bi4
	LD4vx_hi4
	LD4vx_si4
	LD4vx_di4
	LD4vx_bip4
	LD4vxx_bip4
	LD4vx_hip4
	LD4vxx_hip4
	LD4vx_sip4
	LD4vxx_sip4
	LD4vx_dip4
	LD4vxx_dip4
	LD4Rvx_t4
	LD4Rvxi_tp4
	LD4Rvxx_tp4
	LDNPssx
	LDNPddx
	LDNPqqx
	LDPssx
	LDPddx
	LDPqqx
	LDPssxi_p
	LDPddxi_p
	LDPqqxi_p
	LDPssx_w
	LDPddx_w
	LDPqqx_w
	LDRbxi_p
	LDRhxi_p
	LDRsxi_p
	LDRdxi_p
	LDRqxi_p
	LDRbx_w
	LDRhx_w
	LDRsx_w
	LDRdx_w
	LDRqx_w
	LDRbx
	LDRhx
	LDRsx
	LDRdx
	LDRqx
	LDRsl
	LDRdl
	LDRql
	LDRbxre
	LDRhxre
	LDRsxre
	LDRdxre
	LDRqxre
	LDURbx
	LDURhx
	LDURsx
	LDURdx
	LDURqx
	MLAvvv_t
	MLAvvv_ti
	MLSvvv_t
	MLSvvv_ti
	MOVvv_tii
	MOVvv_ti
	MOVvr_ti
	MOVwv_si
	MOVxv_di
	MOVvv_t
	MOVIvi_tb
	MOVIvi_th
	MOVIvi_ts
	MOVIvii_ts
	MOVIdi
	MOVIvi
	MULvvv_i
	MULvvv_t
	MVNvv_t
	MVNIvi_th
	MVNIvi_ts
	MVNIvii_ts
	NEGvv_t
	NEGvv
	NOTvv_t
	ORNvvv_t
	ORRvvv_t
	ORRvi_th
	ORRvi_ts
	PMULvvv_t
	PMULLvvv_t
	PMULL2vvv_t
	RADDHNvvv_t
	RADDHN2vvv_t
	RAX1vvv
	RBITvv_t
	REV16vv_t
	REV32vv_t
	REV64vv_t
	RSHRNvvi_t
	RSHRN2vvi_t
	RSUBHNvvv_t
	RSUBHN2vvv_t
	SABAvvv_t
	SABALvvv_t
	SABAL2vvv_t
	SABDvvv_t
	SABDLvvv_t
	SABDL2vvv_t
	SADALPvv_t
	SADDLvvv_t
	SADDL2vvv_t
	SADDLPvv_t
	SADDLVvv_t
	SADDWvvv_t
	SADDW2vvv_t
	SCVTFswi
	SCVTFdwi
	SCVTFsxi
	SCVTFdxi
	SCVTFsw
	SCVTFdw
	SCVTFsx
	SCVTFdx
	SCVTFvvi
	SCVTFvvi_t
	SCVTFvv
	SCVTFvv_t
	SHA1Cqsv
	SHA1Hss
	SHA1Mqsv
	SHA1Pqsv
	SHA1SU0vvv
	SHA1SU1vv
	SHA256Hqqv
	SHA256H2qqv
	SHA256SU0vv
	SHA256SU1vvv
	SHA512Hqqv
	SHA512H2qqv
	SHA512SU0vv
	SHA512SU1vvv
	SHADDvvv_t
	SHLvvi_t
	SHLvvi
	SHLLvvi_t
	SHLL2vvi_t
	SHRNvvi_t
	SHRN2vvi_t
	SHSUBvvv_t
	SLIvvi_t
	SLIvvi
	SMAXvvv_t
	SMAXPvvv_t
	SMAXVvv_t
	SMINvvv_t
	SMINPvvv_t
	SMINVvv_t
	SMLALvvv_t
	SMLALvvv_ti
	SMLAL2vvv_ti
	SMLAL2vvv_t
	SMLSLvvv_ti
	SMLSLvvv_t
	SMLSL2vvv_ti
	SMLSL2vvv_t
	SMOVwv_ti
	SMOVxv_ti
	SMULLvvv_t
	SMULLvvv_ti
	SMULL2vvv_ti
	SMULL2vvv_t
	SQABSvv
	SQABSvv_t
	SQADDvvv
	SQADDvvv_t
	SQDMLALvvv_tis
	SQDMLALvvv_tiv
	SQDMLALvvv
	SQDMLALvvv_t
	SQDMLAL2vvv_t
	SQDMLAL2vvv_ti
	SQDMLSLvvv
	SQDMLSLvvv_t
	SQDMLSLvvv_tis
	SQDMLSLvvv_tiv
	SQDMLSL2vvv_t
	SQDMLSL2vvv_ti
	SQDMULHvvv
	SQDMULHvvv_t
	SQDMULHvvv_tis
	SQDMULHvvv_tiv
	SQDMULLvvv
	SQDMULLvvv_t
	SQDMULLvvv_tis
	SQDMULLvvv_tiv
	SQDMULL2vvv_t
	SQDMULL2vvv_ti
	SQNEGvv_t
	SQNEGvv
	SQRDMULHvvv
	SQRDMULHvvv_t
	SQRDMULHvvv_tis
	SQRDMULHvvv_tiv
	SQRSHLvvv
	SQRSHLvvv_t
	SQRSHRNvvi
	SQRSHRNvvi_t
	SQRSHRN2vvi_t
	SQRSHRUNvvi
	SQRSHRUNvvi_t
	SQRSHRUN2vvi_t
	SQSHLvvv
	SQSHLvvv_t
	SQSHLvvi
	SQSHLvvi_t
	SQSHLUvvi_t
	SQSHLUvvi
	SQSHRNvvi_t
	SQSHRNvvi
	SQSHRN2vvi_t
	SQSHRUNvvi_t
	SQSHRUNvvi
	SQSHRUN2vvi_t
	SQSUBvvv
	SQSUBvvv_t
	SQXTNvv
	SQXTNvv_t
	SQXTN2vv_t
	SQXTUNvv
	SQXTUNvv_t
	SQXTUN2vv_t
	SRHADDvvv_t
	SRIvvi
	SRIvvi_t
	SRSHLvvv
	SRSHLvvv_t
	SRSHRvvi
	SRSHRvvi_t
	SRSRAvvi
	SRSRAvvi_t
	SSHLvvv
	SSHLvvv_t
	SSHLLvvi_t
	SSHLL2vvi_t
	SSHRvvi
	SSHRvvi_t
	SSRAvvi_t
	SSRAvvi
	SSUBLvvv_t
	SSUBL2vvv_t
	SSUBWvvv_t
	SSUBW2vvv_t
	ST1vx_t1
	ST1vx_t2
	ST1vx_t3
	ST1vx_t4
	ST1vxi_tp1
	ST1vxi_tp2
	ST1vxi_tp3
	ST1vxi_tp4
	ST1vxx_tp1
	ST1vxx_tp2
	ST1vxx_tp3
	ST1vxx_tp4
	ST1vx_bi1
	ST1vx_hi1
	ST1vx_si1
	ST1vx_di1
	ST1vxi_bip1
	ST1vxi_hip1
	ST1vxi_sip1
	ST1vxi_dip1
	ST1vxx_bip1
	ST1vxx_hip1
	ST1vxx_sip1
	ST1vxx_dip1
	ST2vx_t2
	ST2vxi_tp2
	ST2vxx_tp2
	ST2vx_bi2
	ST2vx_hi2
	ST2vx_si2
	ST2vx_di2
	ST2vx_bip2
	ST2vx_hip2
	ST2vx_sip2
	ST2vx_dip2
	ST2vxx_bip2
	ST2vxx_hip2
	ST2vxx_sip2
	ST2vxx_dip2
	ST3vx_t3
	ST3vxi_tp3
	ST3vxx_tp3
	ST3vx_bi3
	ST3vx_hi3
	ST3vx_si3
	ST3vx_di3
	ST3vx_bip3
	ST3vx_hip3
	ST3vx_sip3
	ST3vx_dip3
	ST3vxx_bip3
	ST3vxx_hip3
	ST3vxx_sip3
	ST3vxx_dip3
	ST4vx_t4
	ST4vxi_tp4
	ST4vxx_tp4
	ST4vx_bi4
	ST4vx_hi4
	ST4vx_si4
	ST4vx_di4
	ST4vx_bip4
	ST4vx_hip4
	ST4vx_sip4
	ST4vx_dip4
	ST4vxx_bip4
	ST4vxx_hip4
	ST4vxx_sip4
	ST4vxx_dip4
	STNPssx
	STNPddx
	STNPqqx
	STPssx
	STPddx
	STPqqx
	STPssxi_p
	STPddxi_p
	STPqqxi_p
	STPssx_w
	STPddx_w
	STPqqx_w
	STRbxi_p
	STRhxi_p
	STRsxi_p
	STRdxi_p
	STRqxi_p
	STRbx_w
	STRhx_w
	STRsx_w
	STRdx_w
	STRqx_w
	STRbx
	STRhx
	STRsx
	STRdx
	STRqx
	STRbxre
	STRhxre
	STRsxre
	STRdxre
	STRqxre
	STURbx
	STURhx
	STURsx
	STURdx
	STURqx
	SUBvvv
	SUBvvv_t
	SUBHNvvv_t
	SUBHN2vvv_t
	SUQADDvv
	SUQADDvv_t
	SXTLvv_t
	SXTL2vv_t
	TBLvvv_t1
	TBLvvv_t2
	TBLvvv_t3
	TBLvvv_t4
	TBXvvv_t1
	TBXvvv_t2
	TBXvvv_t3
	TBXvvv_t4
	TRN1vvv_t
	TRN2vvv_t
	UABAvvv_t
	UABALvvv_t
	UABAL2vvv_t
	UABDvvv_t
	UABDLvvv_t
	UABDL2vvv_t
	UADALPvv_t
	UADDLvvv_t
	UADDL2vvv_t
	UADDLPvv_t
	UADDLVvv_t
	UADDWvvv_t
	UADDW2vvv_t
	UCVTFswi
	UCVTFdwi
	UCVTFsxi
	UCVTFdxi
	UCVTFsw
	UCVTFdw
	UCVTFsx
	UCVTFdx
	UCVTFvvi
	UCVTFvvi_t
	UCVTFvv
	UCVTFvv_t
	UHADDvvv_t
	UHSUBvvv_t
	UMAXvvv_t
	UMAXPvvv_t
	UMAXVvv_t
	UMINvvv_t
	UMINPvvv_t
	UMINVvv_t
	UMLALvvv_ti
	UMLALvvv_t
	UMLAL2vvv_ti
	UMLAL2vvv_t
	UMLSLvvv_ti
	UMLSLvvv_t
	UMLSL2vvv_ti
	UMLSL2vvv_t
	UMOVwv_ti
	UMOVxv_ti
	UMULLvvv_ti
	UMULLvvv_t
	UMULL2vvv_ti
	UMULL2vvv_t
	UQADDvvv
	UQADDvvv_t
	UQRSHLvvv
	UQRSHLvvv_t
	UQRSHRNvvi
	UQRSHRNvvi_t
	UQRSHRN2vvi_t
	UQSHLvvi_t
	UQSHLvvv_t
	UQSHLvvi
	UQSHLvvv
	UQSHRNvvi
	UQSHRNvvi_t
	UQSHRN2vvi_t
	UQSUBvvv
	UQSUBvvv_t
	UQXTNvv_t
	UQXTNvv
	UQXTN2vv_t
	URECPEvv_t
	URHADDvvv_t
	URSHLvvv_t
	URSHLvvv
	URSHRvvi_t
	URSHRvvi
	URSQRTEvv_t
	URSRAvvi
	URSRAvvi_t
	USHLvvv
	USHLvvv_t
	USHLLvvi_t
	USHLL2vvi_t
	USHRvvi_t
	USHRvvi
	USQADDvv
	USQADDvv_t
	USRAvvi
	USRAvvi_t
	USUBLvvv_t
	USUBL2vvv_t
	USUBWvvv_t
	USUBW2vvv_t
	UXTLvv_t
	UXTL2vv_t
	UZP1vvv_t
	UZP2vvv_t
	XARvvvi_t
	XTNvv_t
	XTN2vv_t
	ZIP1vvv_t
	ZIP2vvv_t
)

// optab is arm64 instruction table.
var optab = []Optab{
	// Not a valid instruction
	INull: {0x1a000000, "NULL", instArgs{}},
	// ADC <Wd>, <Wn>, <Wm>
	ADCwww: {0x1a000000, "ADC", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// ADC <Xd>, <Xn>, <Xm>
	ADCxxx: {0x9a000000, "ADC", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// ADCS <Wd>, <Wn>, <Wm>
	ADCSwww: {0x3a000000, "ADCS", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// ADCS <Xd>, <Xn>, <Xm>
	ADCSxxx: {0xba000000, "ADCS", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// ADD <Wd|WSP>, <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	ADDwwwe: {0x0b200000, "ADD", instArgs{arg_Wds, arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// ADD <Xd|SP>, <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	ADDxxre: {0x8b200000, "ADD", instArgs{arg_Xds, arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// ADD <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ADDwwws: {0x0b000000, "ADD", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// ADD <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ADDxxxs: {0x8b000000, "ADD", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// ADD <Wd|WSP>, <Wn|WSP>, #<imm>{, <shift>}
	ADDwwis: {0x11000000, "ADD", instArgs{arg_Wds, arg_Wns, arg_IAddSub}},
	// ADD <Xd|SP>, <Xn|SP>, #<imm>{, <shift>}
	ADDxxis: {0x91000000, "ADD", instArgs{arg_Xds, arg_Xns, arg_IAddSub}},
	// ADDS <Xd>, <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	ADDSxxre: {0xab200000, "ADDS", instArgs{arg_Xd, arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// ADDS <Wd>, <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	ADDSwwwe: {0x2b200000, "ADDS", instArgs{arg_Wd, arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// ADDS <Wd>, <Wn|WSP>, #<imm>{, <shift>}
	ADDSwwis: {0x31000000, "ADDS", instArgs{arg_Wd, arg_Wns, arg_IAddSub}},
	// ADDS <Xd>, <Xn|SP>, #<imm>{, <shift>}
	ADDSxxis: {0xb1000000, "ADDS", instArgs{arg_Xd, arg_Xns, arg_IAddSub}},
	// ADDS <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ADDSwwws: {0x2b000000, "ADDS", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// ADDS <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ADDSxxxs: {0xab000000, "ADDS", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// ADR <Xd>, <label>
	ADRxl: {0x10000000, "ADR", instArgs{arg_Xd, arg_slabel_immhi_immlo_0}},
	// ADRP <Xd>, <label>
	ADRPxl: {0x90000000, "ADRP", instArgs{arg_Xd, arg_slabel_immhi_immlo_12}},
	// AND <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ANDxxxs: {0x8a000000, "AND", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// AND <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ANDwwws: {0x0a000000, "AND", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// AND <Wd|WSP>, <Wn>, #<imm>
	ANDwwi: {0x12000000, "AND", instArgs{arg_Wds, arg_Wn, arg_immediate_bitmask_32_imms_immr}},
	// AND <Xd|SP>, <Xn>, #<imm>
	ANDxxi: {0x92000000, "AND", instArgs{arg_Xds, arg_Xn, arg_immediate_bitmask_64_N_imms_immr}},
	// ANDS <Wd>, <Wn>, #<imm>
	ANDSwwi: {0x72000000, "ANDS", instArgs{arg_Wd, arg_Wn, arg_immediate_bitmask_32_imms_immr}},
	// ANDS <Xd>, <Xn>, #<imm>
	ANDSxxi: {0xf2000000, "ANDS", instArgs{arg_Xd, arg_Xn, arg_immediate_bitmask_64_N_imms_immr}},
	// ANDS <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ANDSxxxs: {0xea000000, "ANDS", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// ANDS <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ANDSwwws: {0x6a000000, "ANDS", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// ASR <Xd>, <Xn>, <Xm>
	ASRxxx: {0x9ac02800, "ASR", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// ASR <Wd>, <Wn>, <Wm>
	ASRwww: {0x1ac02800, "ASR", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// ASR <Wd>, <Wn>, #<shift>
	ASRwwi: {0x13007c00, "ASR", instArgs{arg_Wd, arg_Wn, arg_immediate_ASR_SBFM_32M_bitfield_0_31_immr}},
	// ASR <Xd>, <Xn>, #<shift>
	ASRxxi: {0x9340fc00, "ASR", instArgs{arg_Xd, arg_Xn, arg_immediate_ASR_SBFM_64M_bitfield_0_63_immr}},
	// ASRV <Wd>, <Wn>, <Wm>
	ASRVwww: {0x1ac02800, "ASRV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// ASRV <Xd>, <Xn>, <Xm>
	ASRVxxx: {0x9ac02800, "ASRV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// AT <at>, <Xt>
	ATx: {0xd5087800, "AT", instArgs{arg_sysop_AT_SYS_CR_system}},
	// B <label>
	Bl: {0x14000000, "B", instArgs{arg_slabel_imm26_2}},
	// B<c> <label>
	Bcl: {0x54000000, "B", instArgs{arg_conditional, arg_slabel_imm19_2}},
	// BFI <Xd>, <Xn>, #<lsb>, #<width>
	BFIxxii: {0xb3400000, "BFI", instArgs{arg_Xd, arg_Xn, arg_immediate_BFI_BFM_64M_bitfield_lsb_64_immr, arg_immediate_BFI_BFM_64M_bitfield_width_64_imms}},
	// BFI <Wd>, <Wn>, #<lsb>, #<width>
	BFIwwii: {0x33000000, "BFI", instArgs{arg_Wd, arg_Wn, arg_immediate_BFI_BFM_32M_bitfield_lsb_32_immr, arg_immediate_BFI_BFM_32M_bitfield_width_32_imms}},
	// BFM <Wd>, <Wn>, #<immr>, #<imms>
	BFMwwii: {0x33000000, "BFM", instArgs{arg_Wd, arg_Wn, arg_immediate_0_31_immr, arg_immediate_0_31_imms}},
	// BFM <Xd>, <Xn>, #<immr>, #<imms>
	BFMxxii: {0xb3400000, "BFM", instArgs{arg_Xd, arg_Xn, arg_immediate_0_63_immr, arg_immediate_0_63_imms}},
	// BFXIL <Wd>, <Wn>, #<lsb>, #<width>
	BFXILwwii: {0x33000000, "BFXIL", instArgs{arg_Wd, arg_Wn, arg_immediate_BFXIL_BFM_32M_bitfield_lsb_32_immr, arg_immediate_BFXIL_BFM_32M_bitfield_width_32_imms}},
	// BFXIL <Xd>, <Xn>, #<lsb>, #<width>
	BFXILxxii: {0xb3400000, "BFXIL", instArgs{arg_Xd, arg_Xn, arg_immediate_BFXIL_BFM_64M_bitfield_lsb_64_immr, arg_immediate_BFXIL_BFM_64M_bitfield_width_64_imms}},
	// BIC <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	BICxxxs: {0x8a200000, "BIC", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// BIC <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	BICwwws: {0x0a200000, "BIC", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// BICS <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	BICSxxxs: {0xea200000, "BICS", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// BICS <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	BICSwwws: {0x6a200000, "BICS", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// BL <label>
	BLl: {0x94000000, "BL", instArgs{arg_slabel_imm26_2}},
	// BLR <Xn>
	BLRx: {0xd63f0000, "BLR", instArgs{arg_Xn}},
	// BR <Xn>
	BRx: {0xd61f0000, "BR", instArgs{arg_Xn}},
	// BRK #<imm>
	BRKi: {0xd4200000, "BRK", instArgs{arg_immediate_0_65535_imm16}},
	// CAS <Xs>, <Xt>, [<Xn|SP>{, #0}]
	CASxxx: {0xc8a07c00, "CAS", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASA <Xs>, <Xt>, [<Xn|SP>{, #0}]
	CASAxxx: {0xc8e07c00, "CASA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASAL <Xs>, <Xt>, [<Xn|SP>{, #0}]
	CASALxxx: {0xc8e0fc00, "CASAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASL <Xs>, <Xt>, [<Xn|SP>{, #0}]
	CASLxxx: {0xc8a0fc00, "CASL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CAS <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASwwx: {0x88a07c00, "CAS", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASA <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASAwwx: {0x88e07c00, "CASA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASAL <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASALwwx: {0x88e0fc00, "CASAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASL <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASLwwx: {0x88a0fc00, "CASL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASHwwx: {0x48a07c00, "CASH", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASAH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASAHwwx: {0x48e07c00, "CASAH", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASALH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASALHwwx: {0x48e0fc00, "CASALH", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASLH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASLHwwx: {0x48a0fc00, "CASLH", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASBwwx: {0x08a07c00, "CASB", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASAB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASABwwx: {0x08e07c00, "CASAB", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASALB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASALBwwx: {0x08e0fc00, "CASALB", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASLB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	CASLBwwx: {0x08a0fc00, "CASLB", instArgs{arg_Xs, arg_Xt, arg_Xns_mem}},
	// CASP <Xs>, <X(s+1)>, <Xt>, <X(t+1)>, [<Xn|SP>{, #0}]
	CASPxxx: {0x48207c00, "CASP", instArgs{arg_Xs_2_pair_even, arg_Xt_2_pair_even, arg_Xns_mem}},
	// CASPA <Xs>, <X(s+1)>, <Xt>, <X(t+1)>, [<Xn|SP>{, #0}]
	CASPAxxx: {0x48607c00, "CASPA", instArgs{arg_Xs_2_pair_even, arg_Xt_2_pair_even, arg_Xns_mem}},
	// CASPAL <Xs>, <X(s+1)>, <Xt>, <X(t+1)>, [<Xn|SP>{, #0}]
	CASPALxxx: {0x4860fc00, "CASPAL", instArgs{arg_Xs_2_pair_even, arg_Xt_2_pair_even, arg_Xns_mem}},
	// CASPL <Xs>, <X(s+1)>, <Xt>, <X(t+1)>, [<Xn|SP>{, #0}]
	CASPLxxx: {0x4820fc00, "CASPL", instArgs{arg_Xs_2_pair_even, arg_Xt_2_pair_even, arg_Xns_mem}},
	// CASP <Ws>, <W(s+1)>, <Wt>, <W(t+1)>, [<Xn|SP>{, #0}]
	CASPwwx: {0x08207c00, "CASP", instArgs{arg_Ws_2_pair_even, arg_Wt_2_pair_even, arg_Xns_mem}},
	// CASPA <Ws>, <W(s+1)>, <Wt>, <W(t+1)>, [<Xn|SP>{, #0}]
	CASPAwwx: {0x08607c00, "CASPA", instArgs{arg_Ws_2_pair_even, arg_Wt_2_pair_even, arg_Xns_mem}},
	// CASPAL <Ws>, <W(s+1)>, <Wt>, <W(t+1)>, [<Xn|SP>{, #0}]
	CASPALwwx: {0x0860fc00, "CASPAL", instArgs{arg_Ws_2_pair_even, arg_Wt_2_pair_even, arg_Xns_mem}},
	// CASPL <Ws>, <W(s+1)>, <Wt>, <W(t+1)>, [<Xn|SP>{, #0}]
	CASPLwwx: {0x0820fc00, "CASPL", instArgs{arg_Ws_2_pair_even, arg_Wt_2_pair_even, arg_Xns_mem}},
	// CBNZ <Xt>, <label>
	CBNZxl: {0xb5000000, "CBNZ", instArgs{arg_Xt, arg_slabel_imm19_2}},
	// CBNZ <Wt>, <label>
	CBNZwl: {0x35000000, "CBNZ", instArgs{arg_Wt, arg_slabel_imm19_2}},
	// CBZ <Xt>, <label>
	CBZxl: {0xb4000000, "CBZ", instArgs{arg_Xt, arg_slabel_imm19_2}},
	// CBZ <Wt>, <label>
	CBZwl: {0x34000000, "CBZ", instArgs{arg_Wt, arg_slabel_imm19_2}},
	// CCMN <Wn>, #<imm>, #<nzcv>, <cond>
	CCMNwiic: {0x3a400800, "CCMN", instArgs{arg_Wn, arg_immediate_0_31_imm5, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMN <Wn>, <Wm>, #<nzcv>, <cond>
	CCMNwwic: {0x3a400000, "CCMN", instArgs{arg_Wn, arg_Wm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMN <Xn>, #<imm>, #<nzcv>, <cond>
	CCMNxiic: {0xba400800, "CCMN", instArgs{arg_Xn, arg_immediate_0_31_imm5, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMN <Xn>, <Xm>, #<nzcv>, <cond>
	CCMNxxic: {0xba400000, "CCMN", instArgs{arg_Xn, arg_Xm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMP <Wn>, <Wm>, #<nzcv>, <cond>
	CCMPwwic: {0x7a400000, "CCMP", instArgs{arg_Wn, arg_Wm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMP <Xn>, <Xm>, #<nzcv>, <cond>
	CCMPxxic: {0xfa400000, "CCMP", instArgs{arg_Xn, arg_Xm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMP <Wn>, #<imm>, #<nzcv>, <cond>
	CCMPwiic: {0x7a400800, "CCMP", instArgs{arg_Wn, arg_immediate_0_31_imm5, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CCMP <Xn>, #<imm>, #<nzcv>, <cond>
	CCMPxiic: {0xfa400800, "CCMP", instArgs{arg_Xn, arg_immediate_0_31_imm5, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// CINC <Xd>, <Xn>, <cond>
	CINCxxc: {0x9a800400, "CINC", instArgs{arg_Xd, arg_Xmn, arg_cond_NotAllowALNV_Invert}},
	// CINC <Wd>, <Wn>, <cond>
	CINCwwc: {0x1a800400, "CINC", instArgs{arg_Wd, arg_Wmn, arg_cond_NotAllowALNV_Invert}},
	// CINV <Xd>, <Xn>, <cond>
	CINVxxc: {0xda800000, "CINV", instArgs{arg_Xd, arg_Xmn, arg_cond_NotAllowALNV_Invert}},
	// CINV <Wd>, <Wn>, <cond>
	CINVwwc: {0x5a800000, "CINV", instArgs{arg_Wd, arg_Wmn, arg_cond_NotAllowALNV_Invert}},
	// CLREX {#<imm>}
	CLREXi: {0xd503305f, "CLREX", instArgs{arg_immediate_optional_0_15_CRm}},
	// CLS <Xd>, <Xn>
	CLSxx: {0xdac01400, "CLS", instArgs{arg_Xd, arg_Xn}},
	// CLS <Wd>, <Wn>
	CLSww: {0x5ac01400, "CLS", instArgs{arg_Wd, arg_Wn}},
	// CLZ <Xd>, <Xn>
	CLZxx: {0xdac01000, "CLZ", instArgs{arg_Xd, arg_Xn}},
	// CLZ <Wd>, <Wn>
	CLZww: {0x5ac01000, "CLZ", instArgs{arg_Wd, arg_Wn}},
	// CMN <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	CMNxre: {0xab20001f, "CMN", instArgs{arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// CMN <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	CMNwwe: {0x2b20001f, "CMN", instArgs{arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// CMN <Wn|WSP>, #<imm>{, <shift>}
	CMNwis: {0x3100001f, "CMN", instArgs{arg_Wns, arg_IAddSub}},
	// CMN <Xn|SP>, #<imm>{, <shift>}
	CMNxis: {0xb100001f, "CMN", instArgs{arg_Xns, arg_IAddSub}},
	// CMN <Xn>, <Xm> {, <shift> #<amount> }
	CMNxxs: {0xab00001f, "CMN", instArgs{arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// CMN <Wn>, <Wm> {, <shift> #<amount> }
	CMNwws: {0x2b00001f, "CMN", instArgs{arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// CMP <Xn>, <Xm> {, <shift> #<amount> }
	CMPxxs: {0xeb00001f, "CMP", instArgs{arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// CMP <Wn>, <Wm> {, <shift> #<amount> }
	CMPwws: {0x6b00001f, "CMP", instArgs{arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// CMP <Xn|SP>, #<imm>{, <shift>}
	CMPxis: {0xf100001f, "CMP", instArgs{arg_Xns, arg_IAddSub}},
	// CMP <Wn|WSP>, #<imm>{, <shift>}
	CMPwis: {0x7100001f, "CMP", instArgs{arg_Wns, arg_IAddSub}},
	// CMP <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	CMPxre: {0xeb20001f, "CMP", instArgs{arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// CMP <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	CMPwwe: {0x6b20001f, "CMP", instArgs{arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// CNEG <Xd>, <Xn>, <cond>
	CNEGxxc: {0xda800400, "CNEG", instArgs{arg_Xd, arg_Xmn, arg_cond_NotAllowALNV_Invert}},
	// CNEG <Wd>, <Wn>, <cond>
	CNEGwwc: {0x5a800400, "CNEG", instArgs{arg_Wd, arg_Wmn, arg_cond_NotAllowALNV_Invert}},
	// CRC32B <Wd>, <Wn>, <Wm>
	CRC32Bwww: {0x1ac04000, "CRC32B", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32CB <Wd>, <Wn>, <Wm>
	CRC32CBwww: {0x1ac05000, "CRC32CB", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32CH <Wd>, <Wn>, <Wm>
	CRC32CHwww: {0x1ac05400, "CRC32CH", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32CW <Wd>, <Wn>, <Wm>
	CRC32CWwww: {0x1ac05800, "CRC32CW", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32CX <Wd>, <Wn>, <Xm>
	CRC32CXwwx: {0x9ac05c00, "CRC32CX", instArgs{arg_Wd, arg_Wn, arg_Xm}},
	// CRC32H <Wd>, <Wn>, <Wm>
	CRC32Hwww: {0x1ac04400, "CRC32H", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32W <Wd>, <Wn>, <Wm>
	CRC32Wwww: {0x1ac04800, "CRC32W", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// CRC32X <Wd>, <Wn>, <Xm>
	CRC32Xwwx: {0x9ac04c00, "CRC32X", instArgs{arg_Wd, arg_Wn, arg_Xm}},
	// CSEL <Xd>, <Xn>, <Xm>, <cond>
	CSELxxxc: {0x9a800000, "CSEL", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_cond_AllowALNV_Normal}},
	// CSEL <Wd>, <Wn>, <Wm>, <cond>
	CSELwwwc: {0x1a800000, "CSEL", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_cond_AllowALNV_Normal}},
	// CSET <Xd>, <cond>
	CSETxc: {0x9a9f07e0, "CSET", instArgs{arg_Xd, arg_cond_NotAllowALNV_Invert}},
	// CSET <Wd>, <cond>
	CSETwc: {0x1a9f07e0, "CSET", instArgs{arg_Wd, arg_cond_NotAllowALNV_Invert}},
	// CSETM <Wd>, <cond>
	CSETMwc: {0x5a9f03e0, "CSETM", instArgs{arg_Wd, arg_cond_NotAllowALNV_Invert}},
	// CSETM <Xd>, <cond>
	CSETMxc: {0xda9f03e0, "CSETM", instArgs{arg_Xd, arg_cond_NotAllowALNV_Invert}},
	// CSINC <Xd>, <Xn>, <Xm>, <cond>
	CSINCxxxc: {0x9a800400, "CSINC", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_cond_AllowALNV_Normal}},
	// CSINC <Wd>, <Wn>, <Wm>, <cond>
	CSINCwwwc: {0x1a800400, "CSINC", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_cond_AllowALNV_Normal}},
	// CSINV <Xd>, <Xn>, <Xm>, <cond>
	CSINVxxxc: {0xda800000, "CSINV", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_cond_AllowALNV_Normal}},
	// CSINV <Wd>, <Wn>, <Wm>, <cond>
	CSINVwwwc: {0x5a800000, "CSINV", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_cond_AllowALNV_Normal}},
	// CSNEG <Xd>, <Xn>, <Xm>, <cond>
	CSNEGxxxc: {0xda800400, "CSNEG", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_cond_AllowALNV_Normal}},
	// CSNEG <Wd>, <Wn>, <Wm>, <cond>
	CSNEGwwwc: {0x5a800400, "CSNEG", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_cond_AllowALNV_Normal}},
	// DC <dc>, <Xt>
	DCx: {0xd5087000, "DC", instArgs{arg_sysop_DC_SYS_CR_system}},
	// DCPS1 {#<imm>}
	DCPS1i: {0xd4a00001, "DCPS1", instArgs{arg_immediate_optional_0_65535_imm16}},
	// DCPS2 {#<imm>}
	DCPS2i: {0xd4a00002, "DCPS2", instArgs{arg_immediate_optional_0_65535_imm16}},
	// DCPS3 {#<imm>}
	DCPS3i: {0xd4a00003, "DCPS3", instArgs{arg_immediate_optional_0_65535_imm16}},
	// DMB <option>|<imm>
	DMBi: {0xd50330bf, "DMB", instArgs{arg_option_DMB_BO_system_CRm}},
	// DRPS
	DRPS: {0xd6bf03e0, "DRPS", instArgs{}},
	// DSB <option>|<imm>
	DSBi: {0xd503309f, "DSB", instArgs{arg_option_DSB_BO_system_CRm}},
	// EON <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	EONxxxs: {0xca200000, "EON", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// EON <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	EONwwws: {0x4a200000, "EON", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// EOR <Wd|WSP>, <Wn>, #<imm>
	EORwwi: {0x52000000, "EOR", instArgs{arg_Wds, arg_Wn, arg_immediate_bitmask_32_imms_immr}},
	// EOR <Xd|SP>, <Xn>, #<imm>
	EORxxi: {0xd2000000, "EOR", instArgs{arg_Xds, arg_Xn, arg_immediate_bitmask_64_N_imms_immr}},
	// EOR <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	EORwwws: {0x4a000000, "EOR", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// EOR <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	EORxxxs: {0xca000000, "EOR", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// ERET
	ERET: {0xd69f03e0, "ERET", instArgs{}},
	// EXTR <Xd>, <Xn>, <Xm>, #<lsb>
	EXTRxxxi: {0x93c00000, "EXTR", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_immediate_0_63_imms}},
	// EXTR <Wd>, <Wn>, <Wm>, #<lsb>
	EXTRwwwi: {0x13800000, "EXTR", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_immediate_0_31_imms}},
	// HINT #<imm>
	HINTi: {0xd503201f, "HINT", instArgs{arg_immediate_0_127_CRm_op2}},
	// HLT #<imm>
	HLTi: {0xd4400000, "HLT", instArgs{arg_immediate_0_65535_imm16}},
	// HVC #<imm>
	HVCi: {0xd4000002, "HVC", instArgs{arg_immediate_0_65535_imm16}},
	// IC <ic>, {<Xt>}
	ICix: {0xd5087000, "IC", instArgs{arg_sysop_IC_SYS_CR_system}},
	// ISB {<option>|<imm>}
	ISBi: {0xd50330df, "ISB", instArgs{arg_option_ISB_BI_system_CRm}},
	// LDADD <Xs>, <Xt>, [<Xn|SP>]
	LDADDxxx: {0xf8200000, "LDADD", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDADDA <Xs>, <Xt>, [<Xn|SP>]
	LDADDAxxx: {0xf8a00000, "LDADDA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDADDAL <Xs>, <Xt>, [<Xn|SP>]
	LDADDALxxx: {0xf8e00000, "LDADDAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDADDL <Xs>, <Xt>, [<Xn|SP>]
	LDADDLxxx: {0xf8600000, "LDADDL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDADD <Ws>, <Wt>, [<Xn|SP>]
	LDADDwwx: {0xb8200000, "LDADD", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDA <Ws>, <Wt>, [<Xn|SP>]
	LDADDAwwx: {0xb8a00000, "LDADDA", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDAL <Ws>, <Wt>, [<Xn|SP>]
	LDADDALwwx: {0xb8e00000, "LDADDAL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDL <Ws>, <Wt>, [<Xn|SP>]
	LDADDLwwx: {0xb8600000, "LDADDL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDH <Ws>, <Wt>, [<Xn|SP>]
	LDADDHwwx: {0x78200000, "LDADDH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDAH <Ws>, <Wt>, [<Xn|SP>]
	LDADDAHwwx: {0x78a00000, "LDADDAH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDALH <Ws>, <Wt>, [<Xn|SP>]
	LDADDALHwwx: {0x78e00000, "LDADDALH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDLH <Ws>, <Wt>, [<Xn|SP>]
	LDADDLHwwx: {0x78600000, "LDADDLH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDB <Ws>, <Wt>, [<Xn|SP>]
	LDADDBwwx: {0x38200000, "LDADDB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDAB <Ws>, <Wt>, [<Xn|SP>]
	LDADDABwwx: {0x38a00000, "LDADDAB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDALB <Ws>, <Wt>, [<Xn|SP>]
	LDADDALBwwx: {0x38e00000, "LDADDALB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDADDLB <Ws>, <Wt>, [<Xn|SP>]
	LDADDLBwwx: {0x38600000, "LDADDLB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDAR <Xt>, [<Xn|SP>{, #0}]
	LDARxx: {0xc8dffc00, "LDAR", instArgs{arg_Xt, arg_Xns_mem}},
	// LDAR <Wt>, [<Xn|SP>{, #0}]
	LDARwx: {0x88dffc00, "LDAR", instArgs{arg_Wt, arg_Xns_mem}},
	// LDARB <Wt>, [<Xn|SP>{, #0}]
	LDARBwx: {0x08dffc00, "LDARB", instArgs{arg_Wt, arg_Xns_mem}},
	// LDARH <Wt>, [<Xn|SP>{, #0}]
	LDARHwx: {0x48dffc00, "LDARH", instArgs{arg_Wt, arg_Xns_mem}},
	// LDAXP <Wt>, <Wt2>, [<Xn|SP>{, #0}]
	LDAXPwwx: {0x887f8000, "LDAXP", instArgs{arg_Wt_pair, arg_Xns_mem}},
	// LDAXP <Xt>, <Xt2>, [<Xn|SP>{, #0}]
	LDAXPxxx: {0xc87f8000, "LDAXP", instArgs{arg_Xt_pair, arg_Xns_mem}},
	// LDAXR <Xt>, [<Xn|SP>{, #0}]
	LDAXRxx: {0xc85ffc00, "LDAXR", instArgs{arg_Xt, arg_Xns_mem}},
	// LDAXR <Wt>, [<Xn|SP>{, #0}]
	LDAXRwx: {0x885ffc00, "LDAXR", instArgs{arg_Wt, arg_Xns_mem}},
	// LDAXRB <Wt>, [<Xn|SP>{, #0}]
	LDAXRBwx: {0x085ffc00, "LDAXRB", instArgs{arg_Wt, arg_Xns_mem}},
	// LDAXRH <Wt>, [<Xn|SP>{, #0}]
	LDAXRHwx: {0x485ffc00, "LDAXRH", instArgs{arg_Wt, arg_Xns_mem}},
	// LDCLR <Xs>, <Xt>, [<Xn|SP>]
	LDCLRxxx: {0xf8201000, "LDCLR", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDCLRA <Xs>, <Xt>, [<Xn|SP>]
	LDCLRAxxx: {0xf8a01000, "LDCLRA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDCLRAL <Xs>, <Xt>, [<Xn|SP>]
	LDCLRALxxx: {0xf8e01000, "LDCLRAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDCLRL <Xs>, <Xt>, [<Xn|SP>]
	LDCLRLxxx: {0xf8601000, "LDCLRL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDCLR <Ws>, <Wt>, [<Xn|SP>]
	LDCLRwwx: {0xb8201000, "LDCLR", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRA <Ws>, <Wt>, [<Xn|SP>]
	LDCLRAwwx: {0xb8a01000, "LDCLRA", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRAL <Ws>, <Wt>, [<Xn|SP>]
	LDCLRALwwx: {0xb8e01000, "LDCLRAL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRL <Ws>, <Wt>, [<Xn|SP>]
	LDCLRLwwx: {0xb8601000, "LDCLRL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRH <Ws>, <Wt>, [<Xn|SP>]
	LDCLRHwwx: {0x78201000, "LDCLRH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRAH <Ws>, <Wt>, [<Xn|SP>]
	LDCLRAHwwx: {0x78a01000, "LDCLRAH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRALH <Ws>, <Wt>, [<Xn|SP>]
	LDCLRALHwwx: {0x78e01000, "LDCLRALH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRLH <Ws>, <Wt>, [<Xn|SP>]
	LDCLRLHwwx: {0x78601000, "LDCLRLH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRB <Ws>, <Wt>, [<Xn|SP>]
	LDCLRBwwx: {0x38201000, "LDCLRB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRAB <Ws>, <Wt>, [<Xn|SP>]
	LDCLRABwwx: {0x38a01000, "LDCLRAB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRALB <Ws>, <Wt>, [<Xn|SP>]
	LDCLRALBwwx: {0x38e01000, "LDCLRALB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDCLRLB <Ws>, <Wt>, [<Xn|SP>]
	LDCLRLBwwx: {0x38601000, "LDCLRLB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEOR <Xs>, <Xt>, [<Xn|SP>]
	LDEORxxx: {0xf8202000, "LDEOR", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDEORA <Xs>, <Xt>, [<Xn|SP>]
	LDEORAxxx: {0xf8a02000, "LDEORA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDEORAL <Xs>, <Xt>, [<Xn|SP>]
	LDEORALxxx: {0xf8e02000, "LDEORAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDEORL <Xs>, <Xt>, [<Xn|SP>]
	LDEORLxxx: {0xf8602000, "LDEORL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDEOR <Ws>, <Wt>, [<Xn|SP>]
	LDEORwwx: {0xb8202000, "LDEOR", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORA <Ws>, <Wt>, [<Xn|SP>]
	LDEORAwwx: {0xb8a02000, "LDEORA", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORAL <Ws>, <Wt>, [<Xn|SP>]
	LDEORALwwx: {0xb8e02000, "LDEORAL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORL <Ws>, <Wt>, [<Xn|SP>]
	LDEORLwwx: {0xb8602000, "LDEORL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORH <Ws>, <Wt>, [<Xn|SP>]
	LDEORHwwx: {0x78202000, "LDEORH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORAH <Ws>, <Wt>, [<Xn|SP>]
	LDEORAHwwx: {0x78a02000, "LDEORAH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORALH <Ws>, <Wt>, [<Xn|SP>]
	LDEORALHwwx: {0x78e02000, "LDEORALH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORLH <Ws>, <Wt>, [<Xn|SP>]
	LDEORLHwwx: {0x78602000, "LDEORLH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORB <Ws>, <Wt>, [<Xn|SP>]
	LDEORBwwx: {0x38202000, "LDEORB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORAB <Ws>, <Wt>, [<Xn|SP>]
	LDEORABwwx: {0x38a02000, "LDEORAB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORALB <Ws>, <Wt>, [<Xn|SP>]
	LDEORALBwwx: {0x38e02000, "LDEORALB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDEORLB <Ws>, <Wt>, [<Xn|SP>]
	LDEORLBwwx: {0x38602000, "LDEORLB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDNP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_1>}]
	LDNPxxx: {0xa8400000, "LDNP", instArgs{arg_Xt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// LDNP <Wt>, <Wt2>, [<Xn|SP>{, #<imm>}]
	LDNPwwx: {0x28400000, "LDNP", instArgs{arg_Wt_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// LDP <Wt>, <Wt2>, [<Xn|SP>], #<imm_1>
	LDPwwxi_p: {0x28c00000, "LDP", instArgs{arg_Wt_pair, arg_Xns_mem_post_imm7_4_signed}},
	// LDP <Xt>, <Xt2>, [<Xn|SP>], #<imm_3>
	LDPxxxi_p: {0xa8c00000, "LDP", instArgs{arg_Xt_pair, arg_Xns_mem_post_imm7_8_signed}},
	// LDP <Wt>, <Wt2>, [<Xn|SP>{, #<imm_1>}]!
	LDPwwx_w: {0x29c00000, "LDP", instArgs{arg_Wt_pair, arg_Xns_mem_wb_imm7_4_signed}},
	// LDP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_3>}]!
	LDPxxx_w: {0xa9c00000, "LDP", instArgs{arg_Xt_pair, arg_Xns_mem_wb_imm7_8_signed}},
	// LDP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_2>}]
	LDPxxx: {0xa9400000, "LDP", instArgs{arg_Xt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// LDP <Wt>, <Wt2>, [<Xn|SP>{, #<imm>}]
	LDPwwx: {0x29400000, "LDP", instArgs{arg_Wt_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// LDPSW <Xt>, <Xt2>, [<Xn|SP>], #<imm_1>
	LDPSWxxxi_p: {0x68c00000, "LDPSW", instArgs{arg_Xt_pair, arg_Xns_mem_post_imm7_4_signed}},
	// LDPSW <Xt>, <Xt2>, [<Xn|SP>{, #<imm_1>}]!
	LDPSWxxx_w: {0x69c00000, "LDPSW", instArgs{arg_Xt_pair, arg_Xns_mem_wb_imm7_4_signed}},
	// LDPSW <Xt>, <Xt2>, [<Xn|SP>{, #<imm>}]
	LDPSWxxx: {0x69400000, "LDPSW", instArgs{arg_Xt_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// LDR <Wt>, [<Xn|SP>], #<simm>
	LDRwxi_p: {0xb8400400, "LDR", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Xt>, [<Xn|SP>], #<simm>
	LDRxxi_p: {0xf8400400, "LDR", instArgs{arg_Xt, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Wt>, [<Xn|SP>{, #<simm>}]!
	LDRwx_w: {0xb8400c00, "LDR", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Xt>, [<Xn|SP>{, #<simm>}]!
	LDRxx_w: {0xf8400c00, "LDR", instArgs{arg_Xt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Wt>, [<Xn|SP>{, #<pimm>}]
	LDRwx: {0xb9400000, "LDR", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_4_unsigned}},
	// LDR <Xt>, [<Xn|SP>{, #<pimm_1>}]
	LDRxx: {0xf9400000, "LDR", instArgs{arg_Xt, arg_Xns_mem_optional_imm12_8_unsigned}},
	// LDR <Wt>, <label>
	LDRwl: {0x18000000, "LDR", instArgs{arg_Wt, arg_slabel_imm19_2}},
	// LDR <Xt>, <label>
	LDRxl: {0x58000000, "LDR", instArgs{arg_Xt, arg_slabel_imm19_2}},
	// LDR <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRwxre: {0xb8600800, "LDR", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1}},
	// LDR <Xt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRxxre: {0xf8600800, "LDR", instArgs{arg_Xt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1}},
	// LDRB <Wt>, [<Xn|SP>{, #<pimm>}]
	LDRBwx: {0x39400000, "LDRB", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// LDRB <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRBwxre: {0x38600800, "LDRB", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// LDRB <Wt>, [<Xn|SP>{, #<simm>}]!
	LDRBwx_w: {0x38400c00, "LDRB", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRB <Wt>, [<Xn|SP>], #<simm>
	LDRBwxi_p: {0x38400400, "LDRB", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRH <Wt>, [<Xn|SP>{, #<pimm>}]
	LDRHwx: {0x79400000, "LDRH", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_2_unsigned}},
	// LDRH <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRHwxre: {0x78600800, "LDRH", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// LDRH <Wt>, [<Xn|SP>], #<simm>
	LDRHwxi_p: {0x78400400, "LDRH", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRH <Wt>, [<Xn|SP>{, #<simm>}]!
	LDRHwx_w: {0x78400c00, "LDRH", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSB <Wt>, [<Xn|SP>], #<simm>
	LDRSBwxi_p: {0x38c00400, "LDRSB", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRSB <Xt>, [<Xn|SP>], #<simm>
	LDRSBxxi_p: {0x38800400, "LDRSB", instArgs{arg_Xt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRSB <Xt>, [<Xn|SP>{, #<simm>}]!
	LDRSBxx_w: {0x38800c00, "LDRSB", instArgs{arg_Xt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSB <Wt>, [<Xn|SP>{, #<simm>}]!
	LDRSBwx_w: {0x38c00c00, "LDRSB", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSB <Wt>, [<Xn|SP>{, #<pimm>}]
	LDRSBwx: {0x39c00000, "LDRSB", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// LDRSB <Xt>, [<Xn|SP>{, #<pimm>}]
	LDRSBxx: {0x39800000, "LDRSB", instArgs{arg_Xt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// LDRSB <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRSBwxre: {0x38e00800, "LDRSB", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// LDRSB <Xt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRSBxxre: {0x38a00800, "LDRSB", instArgs{arg_Xt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// LDRSH <Xt>, [<Xn|SP>], #<simm>
	LDRSHxxi_p: {0x78800400, "LDRSH", instArgs{arg_Xt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRSH <Wt>, [<Xn|SP>], #<simm>
	LDRSHwxi_p: {0x78c00400, "LDRSH", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRSH <Xt>, [<Xn|SP>{, #<simm>}]!
	LDRSHxx_w: {0x78800c00, "LDRSH", instArgs{arg_Xt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSH <Wt>, [<Xn|SP>{, #<simm>}]!
	LDRSHwx_w: {0x78c00c00, "LDRSH", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSH <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRSHwxre: {0x78e00800, "LDRSH", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// LDRSH <Xt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRSHxxre: {0x78a00800, "LDRSH", instArgs{arg_Xt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// LDRSH <Wt>, [<Xn|SP>{, #<pimm>}]
	LDRSHwx: {0x79c00000, "LDRSH", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_2_unsigned}},
	// LDRSH <Xt>, [<Xn|SP>{, #<pimm>}]
	LDRSHxx: {0x79800000, "LDRSH", instArgs{arg_Xt, arg_Xns_mem_optional_imm12_2_unsigned}},
	// LDRSW <Xt>, <label>
	LDRSWxl: {0x98000000, "LDRSW", instArgs{arg_Xt, arg_slabel_imm19_2}},
	// LDRSW <Xt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRSWxxre: {0xb8a00800, "LDRSW", instArgs{arg_Xt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1}},
	// LDRSW <Xt>, [<Xn|SP>], #<simm>
	LDRSWxxi_p: {0xb8800400, "LDRSW", instArgs{arg_Xt, arg_Xns_mem_post_imm9_1_signed}},
	// LDRSW <Xt>, [<Xn|SP>{, #<simm>}]!
	LDRSWxx_w: {0xb8800c00, "LDRSW", instArgs{arg_Xt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDRSW <Xt>, [<Xn|SP>{, #<pimm>}]
	LDRSWxx: {0xb9800000, "LDRSW", instArgs{arg_Xt, arg_Xns_mem_optional_imm12_4_unsigned}},
	// LDSET <Xs>, <Xt>, [<Xn|SP>]
	LDSETxxx: {0xf8203000, "LDSET", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDSETA <Xs>, <Xt>, [<Xn|SP>]
	LDSETAxxx: {0xf8a03000, "LDSETA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDSETAL <Xs>, <Xt>, [<Xn|SP>]
	LDSETALxxx: {0xf8e03000, "LDSETAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDSETL <Xs>, <Xt>, [<Xn|SP>]
	LDSETLxxx: {0xf8603000, "LDSETL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// LDSET <Ws>, <Wt>, [<Xn|SP>]
	LDSETwwx: {0xb8203000, "LDSET", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETA <Ws>, <Wt>, [<Xn|SP>]
	LDSETAwwx: {0xb8a03000, "LDSETA", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETAL <Ws>, <Wt>, [<Xn|SP>]
	LDSETALwwx: {0xb8e03000, "LDSETAL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETL <Ws>, <Wt>, [<Xn|SP>]
	LDSETLwwx: {0xb8603000, "LDSETL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETH <Ws>, <Wt>, [<Xn|SP>]
	LDSETHwwx: {0x78203000, "LDSETH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETAH <Ws>, <Wt>, [<Xn|SP>]
	LDSETAHwwx: {0x78a03000, "LDSETAH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETALH <Ws>, <Wt>, [<Xn|SP>]
	LDSETALHwwx: {0x78e03000, "LDSETALH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETLH <Ws>, <Wt>, [<Xn|SP>]
	LDSETLHwwx: {0x78603000, "LDSETLH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETB <Ws>, <Wt>, [<Xn|SP>]
	LDSETBwwx: {0x38203000, "LDSETB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETAB <Ws>, <Wt>, [<Xn|SP>]
	LDSETABwwx: {0x38a03000, "LDSETAB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETALB <Ws>, <Wt>, [<Xn|SP>]
	LDSETALBwwx: {0x38e03000, "LDSETALB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDSETLB <Ws>, <Wt>, [<Xn|SP>]
	LDSETLBwwx: {0x38603000, "LDSETLB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// LDTR <Wt>, [<Xn|SP>{, #<simm>}]
	LDTRwx: {0xb8400800, "LDTR", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTR <Xt>, [<Xn|SP>{, #<simm>}]
	LDTRxx: {0xf8400800, "LDTR", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRB <Wt>, [<Xn|SP>{, #<simm>}]
	LDTRBwx: {0x38400800, "LDTRB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRH <Wt>, [<Xn|SP>{, #<simm>}]
	LDTRHwx: {0x78400800, "LDTRH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRSB <Wt>, [<Xn|SP>{, #<simm>}]
	LDTRSBwx: {0x38c00800, "LDTRSB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRSB <Xt>, [<Xn|SP>{, #<simm>}]
	LDTRSBxx: {0x38800800, "LDTRSB", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRSH <Xt>, [<Xn|SP>{, #<simm>}]
	LDTRSHxx: {0x78800800, "LDTRSH", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRSH <Wt>, [<Xn|SP>{, #<simm>}]
	LDTRSHwx: {0x78c00800, "LDTRSH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDTRSW <Xt>, [<Xn|SP>{, #<simm>}]
	LDTRSWxx: {0xb8800800, "LDTRSW", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <Wt>, [<Xn|SP>{, #<simm>}]
	LDURwx: {0xb8400000, "LDUR", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <Xt>, [<Xn|SP>{, #<simm>}]
	LDURxx: {0xf8400000, "LDUR", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURB <Wt>, [<Xn|SP>{, #<simm>}]
	LDURBwx: {0x38400000, "LDURB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURH <Wt>, [<Xn|SP>{, #<simm>}]
	LDURHwx: {0x78400000, "LDURH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURSB <Wt>, [<Xn|SP>{, #<simm>}]
	LDURSBwx: {0x38c00000, "LDURSB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURSB <Xt>, [<Xn|SP>{, #<simm>}]
	LDURSBxx: {0x38800000, "LDURSB", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURSH <Wt>, [<Xn|SP>{, #<simm>}]
	LDURSHwx: {0x78c00000, "LDURSH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURSH <Xt>, [<Xn|SP>{, #<simm>}]
	LDURSHxx: {0x78800000, "LDURSH", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDURSW <Xt>, [<Xn|SP>{, #<simm>}]
	LDURSWxx: {0xb8800000, "LDURSW", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDXP <Wt>, <Wt2>, [<Xn|SP>{, #0}]
	LDXPwwx: {0x887f0000, "LDXP", instArgs{arg_Wt_pair, arg_Xns_mem}},
	// LDXP <Xt>, <Xt2>, [<Xn|SP>{, #0}]
	LDXPxxx: {0xc87f0000, "LDXP", instArgs{arg_Xt_pair, arg_Xns_mem}},
	// LDXR <Wt>, [<Xn|SP>{, #0}]
	LDXRwx: {0x885f7c00, "LDXR", instArgs{arg_Wt, arg_Xns_mem}},
	// LDXR <Xt>, [<Xn|SP>{, #0}]
	LDXRxx: {0xc85f7c00, "LDXR", instArgs{arg_Xt, arg_Xns_mem}},
	// LDXRB <Wt>, [<Xn|SP>{, #0}]
	LDXRBwx: {0x085f7c00, "LDXRB", instArgs{arg_Wt, arg_Xns_mem}},
	// LDXRH <Wt>, [<Xn|SP>{, #0}]
	LDXRHwx: {0x485f7c00, "LDXRH", instArgs{arg_Wt, arg_Xns_mem}},
	// LSL <Wd>, <Wn>, #<shift>
	LSLwwi: {0x53000000, "LSL", instArgs{arg_Wd, arg_Wn, arg_immediate_LSL_UBFM_32M_bitfield_0_31_immr}},
	// LSL <Xd>, <Xn>, #<shift>
	LSLxxi: {0xd3400000, "LSL", instArgs{arg_Xd, arg_Xn, arg_immediate_LSL_UBFM_64M_bitfield_0_63_immr}},
	// LSL <Wd>, <Wn>, <Wm>
	LSLwww: {0x1ac02000, "LSL", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// LSL <Xd>, <Xn>, <Xm>
	LSLxxx: {0x9ac02000, "LSL", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// LSLV <Wd>, <Wn>, <Wm>
	LSLVwww: {0x1ac02000, "LSLV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// LSLV <Xd>, <Xn>, <Xm>
	LSLVxxx: {0x9ac02000, "LSLV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// LSR <Wd>, <Wn>, <Wm>
	LSRwww: {0x1ac02400, "LSR", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// LSR <Xd>, <Xn>, <Xm>
	LSRxxx: {0x9ac02400, "LSR", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// LSR <Xd>, <Xn>, #<shift>
	LSRxxi: {0xd340fc00, "LSR", instArgs{arg_Xd, arg_Xn, arg_immediate_LSR_UBFM_64M_bitfield_0_63_immr}},
	// LSR <Wd>, <Wn>, #<shift>
	LSRwwi: {0x53007c00, "LSR", instArgs{arg_Wd, arg_Wn, arg_immediate_LSR_UBFM_32M_bitfield_0_31_immr}},
	// LSRV <Wd>, <Wn>, <Wm>
	LSRVwww: {0x1ac02400, "LSRV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// LSRV <Xd>, <Xn>, <Xm>
	LSRVxxx: {0x9ac02400, "LSRV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// MADD <Xd>, <Xn>, <Xm>, <Xa>
	MADDxxxx: {0x9b000000, "MADD", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_Xa}},
	// MADD <Wd>, <Wn>, <Wm>, <Wa>
	MADDwwww: {0x1b000000, "MADD", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_Wa}},
	// MNEG <Xd>, <Xn>, <Xm>
	MNEGxxx: {0x9b00fc00, "MNEG", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// MNEG <Wd>, <Wn>, <Wm>
	MNEGwww: {0x1b00fc00, "MNEG", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// MOV <Wd|WSP>, #<imm>
	MOVwi_b: {0x320003e0, "MOV", instArgs{arg_Wds, arg_immediate_bitmask_32_imms_immr}},
	// MOV <Xd|SP>, #<imm>
	MOVxi_b: {0xb20003e0, "MOV", instArgs{arg_Xds, arg_immediate_bitmask_64_N_imms_immr}},
	// MOV <Wd>, #<imm>
	MOVwi_n: {0x12800000, "MOV", instArgs{arg_Wd, arg_immediate_shift_32_implicit_inverse_imm16_hw}},
	// MOV <Xd>, #<imm>
	MOVxi_n: {0x92800000, "MOV", instArgs{arg_Xd, arg_immediate_shift_64_implicit_inverse_imm16_hw}},
	// MOV <Wd>, <Wm>
	MOVww: {0x2a0003e0, "MOV", instArgs{arg_Wd, arg_Wm}},
	// MOV <Xd>, <Xm>
	MOVxx: {0xaa0003e0, "MOV", instArgs{arg_Xd, arg_Xm}},
	// MOV <Wd|WSP>, <Wn|WSP>
	MOVww_sp: {0x11000000, "MOV", instArgs{arg_Wds, arg_Wns}},
	// MOV <Xd|SP>, <Xn|SP>
	MOVxx_sp: {0x91000000, "MOV", instArgs{arg_Xds, arg_Xns}},
	// MOV <Wd>, #<imm>
	MOVwi_z: {0x52800000, "MOV", instArgs{arg_Wd, arg_immediate_shift_32_implicit_imm16_hw}},
	// MOV <Xd>, #<imm>
	MOVxi_z: {0xd2800000, "MOV", instArgs{arg_Xd, arg_immediate_shift_64_implicit_imm16_hw}},
	// MOVK <Wd>, #<imm>{, LSL #<shift>}
	MOVKwis: {0x72800000, "MOVK", instArgs{arg_Wd, arg_immediate_OptLSL_amount_16_0_16}},
	// MOVK <Xd>, #<imm>{, LSL #<shift>}
	MOVKxis: {0xf2800000, "MOVK", instArgs{arg_Xd, arg_immediate_OptLSL_amount_16_0_48}},
	// MOVN <Wd>, #<imm>{, LSL #<shift>}
	MOVNwis: {0x12800000, "MOVN", instArgs{arg_Wd, arg_immediate_OptLSL_amount_16_0_16}},
	// MOVN <Xd>, #<imm>{, LSL #<shift>}
	MOVNxis: {0x92800000, "MOVN", instArgs{arg_Xd, arg_immediate_OptLSL_amount_16_0_48}},
	// MOVZ <Wd>, #<imm>{, LSL #<shift>}
	MOVZwis: {0x52800000, "MOVZ", instArgs{arg_Wd, arg_immediate_OptLSL_amount_16_0_16}},
	// MOVZ <Xd>, #<imm>{, LSL #<shift>}
	MOVZxis: {0xd2800000, "MOVZ", instArgs{arg_Xd, arg_immediate_OptLSL_amount_16_0_48}},
	// MRS <Xt>, <systemreg>
	MRSx: {0xd5300000, "MRS", instArgs{arg_Xt, arg_sysreg_o0_op1_CRn_CRm_op2}},
	// MSR <pstatefield>, #<imm>
	MSRi: {0xd500401f, "MSR", instArgs{arg_pstatefield_op1_op2__SPSel_05__DAIFSet_36__DAIFClr_37, arg_immediate_0_15_CRm}},
	// MSR <systemreg>, <Xt>
	MSRx: {0xd5100000, "MSR", instArgs{arg_sysreg_o0_op1_CRn_CRm_op2, arg_Xt}},
	// MSUB <Wd>, <Wn>, <Wm>, <Wa>
	MSUBwwww: {0x1b008000, "MSUB", instArgs{arg_Wd, arg_Wn, arg_Wm, arg_Wa}},
	// MSUB <Xd>, <Xn>, <Xm>, <Xa>
	MSUBxxxx: {0x9b008000, "MSUB", instArgs{arg_Xd, arg_Xn, arg_Xm, arg_Xa}},
	// MUL <Xd>, <Xn>, <Xm>
	MULxxx: {0x9b007c00, "MUL", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// MUL <Wd>, <Wn>, <Wm>
	MULwww: {0x1b007c00, "MUL", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// MVN <Wd>, <Wm> {, <shift> #<amount> }
	MVNwws: {0x2a2003e0, "MVN", instArgs{arg_Wd, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// MVN <Xd>, <Xm> {, <shift> #<amount> }
	MVNxxs: {0xaa2003e0, "MVN", instArgs{arg_Xd, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// NEG <Wd>, <Wm> {, <shift> #<amount> }
	NEGwws: {0x4b0003e0, "NEG", instArgs{arg_Wd, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// NEG <Xd>, <Xm> {, <shift> #<amount> }
	NEGxxs: {0xcb0003e0, "NEG", instArgs{arg_Xd, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// NEGS <Xd>, <Xm> {, <shift> #<amount> }
	NEGSxxs: {0xeb0003e0, "NEGS", instArgs{arg_Xd, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// NEGS <Wd>, <Wm> {, <shift> #<amount> }
	NEGSwws: {0x6b0003e0, "NEGS", instArgs{arg_Wd, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// NGC <Xd>, <Xm>
	NGCxx: {0xda0003e0, "NGC", instArgs{arg_Xd, arg_Xm}},
	// NGC <Wd>, <Wm>
	NGCww: {0x5a0003e0, "NGC", instArgs{arg_Wd, arg_Wm}},
	// NGCS <Xd>, <Xm>
	NGCSxx: {0xfa0003e0, "NGCS", instArgs{arg_Xd, arg_Xm}},
	// NGCS <Wd>, <Wm>
	NGCSww: {0x7a0003e0, "NGCS", instArgs{arg_Wd, arg_Wm}},
	// NOP
	NOP: {0xd503201f, "NOP", instArgs{}},
	// ORN <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ORNwwws: {0x2a200000, "ORN", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// ORN <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ORNxxxs: {0xaa200000, "ORN", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// ORR <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	ORRwwws: {0x2a000000, "ORR", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// ORR <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	ORRxxxs: {0xaa000000, "ORR", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// ORR <Xd|SP>, <Xn>, #<imm>
	ORRxxi: {0xb2000000, "ORR", instArgs{arg_Xds, arg_Xn, arg_immediate_bitmask_64_N_imms_immr}},
	// ORR <Wd|WSP>, <Wn>, #<imm>
	ORRwwi: {0x32000000, "ORR", instArgs{arg_Wds, arg_Wn, arg_immediate_bitmask_32_imms_immr}},
	// PRFM <prfop>, [<Xn|SP>{, #<pimm>}]
	PRFMix: {0xf9800000, "PRFM", instArgs{arg_prfop_Rt, arg_Xns_mem_optional_imm12_8_unsigned}},
	// PRFM <prfop>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	PRFMixre: {0xf8a00800, "PRFM", instArgs{arg_prfop_Rt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1}},
	// PRFM <prfop>, <label>
	PRFMil: {0xd8000000, "PRFM", instArgs{arg_prfop_Rt, arg_slabel_imm19_2}},
	// PRFUM <prfop>, [<Xn|SP>{, #<simm>}]
	PRFUMix: {0xf8800000, "PRFUM", instArgs{arg_prfop_Rt, arg_Xns_mem_optional_imm9_1_signed}},
	// RBIT <Xd>, <Xn>
	RBITxx: {0xdac00000, "RBIT", instArgs{arg_Xd, arg_Xn}},
	// RBIT <Wd>, <Wn>
	RBITww: {0x5ac00000, "RBIT", instArgs{arg_Wd, arg_Wn}},
	// RET {<Xn>}
	RETx: {0xd65f0000, "RET", instArgs{arg_Xn}},
	// REV <Wd>, <Wn>
	REVww: {0x5ac00800, "REV", instArgs{arg_Wd, arg_Wn}},
	// REV <Xd>, <Xn>
	REVxx: {0xdac00c00, "REV", instArgs{arg_Xd, arg_Xn}},
	// REV16 <Wd>, <Wn>
	REV16ww: {0x5ac00400, "REV16", instArgs{arg_Wd, arg_Wn}},
	// REV16 <Xd>, <Xn>
	REV16xx: {0xdac00400, "REV16", instArgs{arg_Xd, arg_Xn}},
	// REV32 <Xd>, <Xn>
	REV32xx: {0xdac00800, "REV32", instArgs{arg_Xd, arg_Xn}},
	// ROR <Wd>, <Ws>, #<shift>
	RORwwi: {0x13800000, "ROR", instArgs{arg_Wd, arg_Wmn, arg_immediate_0_31_imms}},
	// ROR <Xd>, <Xs>, #<shift>
	RORxxi: {0x93c00000, "ROR", instArgs{arg_Xd, arg_Xmn, arg_immediate_0_63_imms}},
	// ROR <Xd>, <Xn>, <Xm>
	RORxxx: {0x9ac02c00, "ROR", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// ROR <Wd>, <Wn>, <Wm>
	RORwww: {0x1ac02c00, "ROR", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// RORV <Xd>, <Xn>, <Xm>
	RORVxxx: {0x9ac02c00, "RORV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// RORV <Wd>, <Wn>, <Wm>
	RORVwww: {0x1ac02c00, "RORV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// SBC <Xd>, <Xn>, <Xm>
	SBCxxx: {0xda000000, "SBC", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// SBC <Wd>, <Wn>, <Wm>
	SBCwww: {0x5a000000, "SBC", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// SBCS <Xd>, <Xn>, <Xm>
	SBCSxxx: {0xfa000000, "SBCS", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// SBCS <Wd>, <Wn>, <Wm>
	SBCSwww: {0x7a000000, "SBCS", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// SBFIZ <Wd>, <Wn>, #<lsb>, #<width>
	SBFIZwwii: {0x13000000, "SBFIZ", instArgs{arg_Wd, arg_Wn, arg_immediate_SBFIZ_SBFM_32M_bitfield_lsb_32_immr, arg_immediate_SBFIZ_SBFM_32M_bitfield_width_32_imms}},
	// SBFIZ <Xd>, <Xn>, #<lsb>, #<width>
	SBFIZxxii: {0x93400000, "SBFIZ", instArgs{arg_Xd, arg_Xn, arg_immediate_SBFIZ_SBFM_64M_bitfield_lsb_64_immr, arg_immediate_SBFIZ_SBFM_64M_bitfield_width_64_imms}},
	// SBFM <Xd>, <Xn>, #<immr>, #<imms>
	SBFMxxii: {0x93400000, "SBFM", instArgs{arg_Xd, arg_Xn, arg_immediate_0_63_immr, arg_immediate_0_63_imms}},
	// SBFM <Wd>, <Wn>, #<immr>, #<imms>
	SBFMwwii: {0x13000000, "SBFM", instArgs{arg_Wd, arg_Wn, arg_immediate_0_31_immr, arg_immediate_0_31_imms}},
	// SBFX <Xd>, <Xn>, #<lsb>, #<width>
	SBFXxxii: {0x93400000, "SBFX", instArgs{arg_Xd, arg_Xn, arg_immediate_SBFX_SBFM_64M_bitfield_lsb_64_immr, arg_immediate_SBFX_SBFM_64M_bitfield_width_64_imms}},
	// SBFX <Wd>, <Wn>, #<lsb>, #<width>
	SBFXwwii: {0x13000000, "SBFX", instArgs{arg_Wd, arg_Wn, arg_immediate_SBFX_SBFM_32M_bitfield_lsb_32_immr, arg_immediate_SBFX_SBFM_32M_bitfield_width_32_imms}},
	// SDIV <Wd>, <Wn>, <Wm>
	SDIVwww: {0x1ac00c00, "SDIV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// SDIV <Xd>, <Xn>, <Xm>
	SDIVxxx: {0x9ac00c00, "SDIV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// SEV
	SEV: {0xd503209f, "SEV", instArgs{}},
	// SEVL
	SEVL: {0xd50320bf, "SEVL", instArgs{}},
	// SMADDL <Xd>, <Wn>, <Wm>, <Xa>
	SMADDLxwwx: {0x9b200000, "SMADDL", instArgs{arg_Xd, arg_Wn, arg_Wm, arg_Xa}},
	// SMC #<imm>
	SMCi: {0xd4000003, "SMC", instArgs{arg_immediate_0_65535_imm16}},
	// SMNEGL <Xd>, <Wn>, <Wm>
	SMNEGLxww: {0x9b20fc00, "SMNEGL", instArgs{arg_Xd, arg_Wn, arg_Wm}},
	// SMSUBL <Xd>, <Wn>, <Wm>, <Xa>
	SMSUBLxwwx: {0x9b208000, "SMSUBL", instArgs{arg_Xd, arg_Wn, arg_Wm, arg_Xa}},
	// SMULH <Xd>, <Xn>, <Xm>
	SMULHxxx: {0x9b407c00, "SMULH", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// SMULL <Xd>, <Wn>, <Wm>
	SMULLxww: {0x9b207c00, "SMULL", instArgs{arg_Xd, arg_Wn, arg_Wm}},
	// STLR <Wt>, [<Xn|SP>{, #0}]
	STLRwx: {0x889ffc00, "STLR", instArgs{arg_Wt, arg_Xns_mem}},
	// STLR <Xt>, [<Xn|SP>{, #0}]
	STLRxx: {0xc89ffc00, "STLR", instArgs{arg_Xt, arg_Xns_mem}},
	// STLRB <Wt>, [<Xn|SP>{, #0}]
	STLRBwx: {0x089ffc00, "STLRB", instArgs{arg_Wt, arg_Xns_mem}},
	// STLRH <Wt>, [<Xn|SP>{, #0}]
	STLRHwx: {0x489ffc00, "STLRH", instArgs{arg_Wt, arg_Xns_mem}},
	// STLXP <Ws>, <Xt>, <Xt2>, [<Xn|SP>{, #0}]
	STLXPwxxx: {0xc8208000, "STLXP", instArgs{arg_Ws, arg_Xt_pair, arg_Xns_mem}},
	// STLXP <Ws>, <Wt>, <Wt2>, [<Xn|SP>{, #0}]
	STLXPwwwx: {0x88208000, "STLXP", instArgs{arg_Ws, arg_Wt_pair, arg_Xns_mem}},
	// STLXR <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STLXRwwx: {0x8800fc00, "STLXR", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// STLXR <Ws>, <Xt>, [<Xn|SP>{, #0}]
	STLXRwxx: {0xc800fc00, "STLXR", instArgs{arg_Ws, arg_Xt, arg_Xns_mem}},
	// STLXRB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STLXRBwwx: {0x0800fc00, "STLXRB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// STLXRH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STLXRHwwx: {0x4800fc00, "STLXRH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// STNP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_1>}]
	STNPxxx: {0xa8000000, "STNP", instArgs{arg_Xt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// STNP <Wt>, <Wt2>, [<Xn|SP>{, #<imm>}]
	STNPwwx: {0x28000000, "STNP", instArgs{arg_Wt_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// STP <Xt>, <Xt2>, [<Xn|SP>], #<imm_3>
	STPxxxi_p: {0xa8800000, "STP", instArgs{arg_Xt_pair, arg_Xns_mem_post_imm7_8_signed}},
	// STP <Wt>, <Wt2>, [<Xn|SP>], #<imm_1>
	STPwwxi_p: {0x28800000, "STP", instArgs{arg_Wt_pair, arg_Xns_mem_post_imm7_4_signed}},
	// STP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_3>}]!
	STPxxx_w: {0xa9800000, "STP", instArgs{arg_Xt_pair, arg_Xns_mem_wb_imm7_8_signed}},
	// STP <Wt>, <Wt2>, [<Xn|SP>{, #<imm_1>}]!
	STPwwx_w: {0x29800000, "STP", instArgs{arg_Wt_pair, arg_Xns_mem_wb_imm7_4_signed}},
	// STP <Xt>, <Xt2>, [<Xn|SP>{, #<imm_2>}]
	STPxxx: {0xa9000000, "STP", instArgs{arg_Xt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// STP <Wt>, <Wt2>, [<Xn|SP>{, #<imm>}]
	STPwwx: {0x29000000, "STP", instArgs{arg_Wt_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// STR <Wt>, [<Xn|SP>], #<simm>
	STRwxi_p: {0xb8000400, "STR", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Xt>, [<Xn|SP>], #<simm>
	STRxxi_p: {0xf8000400, "STR", instArgs{arg_Xt, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Wt>, [<Xn|SP>{, #<simm>}]!
	STRwx_w: {0xb8000c00, "STR", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Xt>, [<Xn|SP>{, #<simm>}]!
	STRxx_w: {0xf8000c00, "STR", instArgs{arg_Xt, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Wt>, [<Xn|SP>{, #<pimm>}]
	STRwx: {0xb9000000, "STR", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_4_unsigned}},
	// STR <Xt>, [<Xn|SP>{, #<pimm_1>}]
	STRxx: {0xf9000000, "STR", instArgs{arg_Xt, arg_Xns_mem_optional_imm12_8_unsigned}},
	// STR <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRwxre: {0xb8200800, "STR", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1}},
	// STR <Xt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRxxre: {0xf8200800, "STR", instArgs{arg_Xt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1}},
	// STRB <Wt>, [<Xn|SP>], #<simm>
	STRBwxi_p: {0x38000400, "STRB", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// STRB <Wt>, [<Xn|SP>{, #<simm>}]!
	STRBwx_w: {0x38000c00, "STRB", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// STRB <Wt>, [<Xn|SP>{, #<pimm>}]
	STRBwx: {0x39000000, "STRB", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// STRB <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRBwxre: {0x38200800, "STRB", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// STRH <Wt>, [<Xn|SP>], #<simm>
	STRHwxi_p: {0x78000400, "STRH", instArgs{arg_Wt, arg_Xns_mem_post_imm9_1_signed}},
	// STRH <Wt>, [<Xn|SP>{, #<simm>}]!
	STRHwx_w: {0x78000c00, "STRH", instArgs{arg_Wt, arg_Xns_mem_wb_imm9_1_signed}},
	// STRH <Wt>, [<Xn|SP>{, #<pimm>}]
	STRHwx: {0x79000000, "STRH", instArgs{arg_Wt, arg_Xns_mem_optional_imm12_2_unsigned}},
	// STRH <Wt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRHwxre: {0x78200800, "STRH", instArgs{arg_Wt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// STTR <Xt>, [<Xn|SP>{, #<simm>}]
	STTRxx: {0xf8000800, "STTR", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// STTR <Wt>, [<Xn|SP>{, #<simm>}]
	STTRwx: {0xb8000800, "STTR", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STTRB <Wt>, [<Xn|SP>{, #<simm>}]
	STTRBwx: {0x38000800, "STTRB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STTRH <Wt>, [<Xn|SP>{, #<simm>}]
	STTRHwx: {0x78000800, "STTRH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <Wt>, [<Xn|SP>{, #<simm>}]
	STURwx: {0xb8000000, "STUR", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <Xt>, [<Xn|SP>{, #<simm>}]
	STURxx: {0xf8000000, "STUR", instArgs{arg_Xt, arg_Xns_mem_optional_imm9_1_signed}},
	// STURB <Wt>, [<Xn|SP>{, #<simm>}]
	STURBwx: {0x38000000, "STURB", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STURH <Wt>, [<Xn|SP>{, #<simm>}]
	STURHwx: {0x78000000, "STURH", instArgs{arg_Wt, arg_Xns_mem_optional_imm9_1_signed}},
	// STXP <Ws>, <Wt>, <Wt2>, [<Xn|SP>{, #0}]
	STXPwwwx: {0x88200000, "STXP", instArgs{arg_Ws, arg_Wt_pair, arg_Xns_mem}},
	// STXP <Ws>, <Xt>, <Xt2>, [<Xn|SP>{, #0}]
	STXPwxxx: {0xc8200000, "STXP", instArgs{arg_Ws, arg_Xt_pair, arg_Xns_mem}},
	// STXR <Ws>, <Xt>, [<Xn|SP>{, #0}]
	STXRwxx: {0xc8007c00, "STXR", instArgs{arg_Ws, arg_Xt, arg_Xns_mem}},
	// STXR <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STXRwwx: {0x88007c00, "STXR", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// STXRB <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STXRBwwx: {0x08007c00, "STXRB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// STXRH <Ws>, <Wt>, [<Xn|SP>{, #0}]
	STXRHwwx: {0x48007c00, "STXRH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem}},
	// SUB <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	SUBxxxs: {0xcb000000, "SUB", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// SUB <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	SUBwwws: {0x4b000000, "SUB", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// SUB <Xd|SP>, <Xn|SP>, #<imm>{, <shift>}
	SUBxxis: {0xd1000000, "SUB", instArgs{arg_Xds, arg_Xns, arg_IAddSub}},
	// SUB <Wd|WSP>, <Wn|WSP>, #<imm>{, <shift>}
	SUBwwis: {0x51000000, "SUB", instArgs{arg_Wds, arg_Wns, arg_IAddSub}},
	// SUB <Xd|SP>, <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	SUBxxre: {0xcb200000, "SUB", instArgs{arg_Xds, arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// SUB <Wd|WSP>, <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	SUBwwwe: {0x4b200000, "SUB", instArgs{arg_Wds, arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// SUBS <Xd>, <Xn>, <Xm> {, <shift> #<amount> }
	SUBSxxxs: {0xeb000000, "SUBS", instArgs{arg_Xd, arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63}},
	// SUBS <Wd>, <Wn>, <Wm> {, <shift> #<amount> }
	SUBSwwws: {0x6b000000, "SUBS", instArgs{arg_Wd, arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31}},
	// SUBS <Wd>, <Wn|WSP>, #<imm>{, <shift>}
	SUBSwwis: {0x71000000, "SUBS", instArgs{arg_Wd, arg_Wns, arg_IAddSub}},
	// SUBS <Xd>, <Xn|SP>, #<imm>{, <shift>}
	SUBSxxis: {0xf1000000, "SUBS", instArgs{arg_Xd, arg_Xns, arg_IAddSub}},
	// SUBS <Xd>, <Xn|SP>, <R><m>{, <extend_1> {#<amount>}}
	SUBSxxre: {0xeb200000, "SUBS", instArgs{arg_Xd, arg_Xns, arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// SUBS <Wd>, <Wn|WSP>, <Wm>{, <extend> {#<amount>}}
	SUBSwwwe: {0x6b200000, "SUBS", instArgs{arg_Wd, arg_Wns, arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4}},
	// SVC #<imm>
	SVCi: {0xd4000001, "SVC", instArgs{arg_immediate_0_65535_imm16}},
	// SWP <Xs>, <Xt>, [<Xn|SP>]
	SWPxxx: {0xf8208000, "SWP", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// SWPA <Xs>, <Xt>, [<Xn|SP>]
	SWPAxxx: {0xf8a08000, "SWPA", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// SWPAL <Xs>, <Xt>, [<Xn|SP>]
	SWPALxxx: {0xf8e08000, "SWPAL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// SWPL <Xs>, <Xt>, [<Xn|SP>]
	SWPLxxx: {0xf8608000, "SWPL", instArgs{arg_Xs, arg_Xt, arg_Xns_mem_offset}},
	// SWP <Ws>, <Wt>, [<Xn|SP>]
	SWPwwx: {0xb8208000, "SWP", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPA <Ws>, <Wt>, [<Xn|SP>]
	SWPAwwx: {0xb8a08000, "SWPA", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPAL <Ws>, <Wt>, [<Xn|SP>]
	SWPALwwx: {0xb8e08000, "SWPAL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPL <Ws>, <Wt>, [<Xn|SP>]
	SWPLwwx: {0xb8608000, "SWPL", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPH <Ws>, <Wt>, [<Xn|SP>]
	SWPHwwx: {0x78208000, "SWPH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPAH <Ws>, <Wt>, [<Xn|SP>]
	SWPAHwwx: {0x78a08000, "SWPAH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPALH <Ws>, <Wt>, [<Xn|SP>]
	SWPALHwwx: {0x78e08000, "SWPALH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPLH <Ws>, <Wt>, [<Xn|SP>]
	SWPLHwwx: {0x78608000, "SWPLH", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPB <Ws>, <Wt>, [<Xn|SP>]
	SWPBwwx: {0x38208000, "SWPB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPAB <Ws>, <Wt>, [<Xn|SP>]
	SWPABwwx: {0x38a08000, "SWPAB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPALB <Ws>, <Wt>, [<Xn|SP>]
	SWPALBwwx: {0x38e08000, "SWPALB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SWPLB <Ws>, <Wt>, [<Xn|SP>]
	SWPLBwwx: {0x38608000, "SWPLB", instArgs{arg_Ws, arg_Wt, arg_Xns_mem_offset}},
	// SXTB <Wd>, <Wn>
	SXTBww: {0x13001c00, "SXTB", instArgs{arg_Wd, arg_Wn}},
	// SXTB <Xd>, <Wn>
	SXTBxw: {0x93401c00, "SXTB", instArgs{arg_Xd, arg_Wn}},
	// SXTH <Xd>, <Wn>
	SXTHxw: {0x93403c00, "SXTH", instArgs{arg_Xd, arg_Wn}},
	// SXTH <Wd>, <Wn>
	SXTHww: {0x13003c00, "SXTH", instArgs{arg_Wd, arg_Wn}},
	// SXTW <Xd>, <Wn>
	SXTWxw: {0x93407c00, "SXTW", instArgs{arg_Xd, arg_Wn}},
	// SYS #<op1>, <Cn>, <Cm>, <op>, {<Xt>}
	SYSix: {0xd5080000, "SYS", instArgs{arg_immediate_0_7_op1, arg_Cn, arg_Cm, arg_sysop_SYS_CR_system}},
	// SYSL <Xt>, #<op1>, <Cn>, <Cm>, #<op2>
	SYSLix: {0xd5280000, "SYSL", instArgs{arg_Xt, arg_immediate_0_7_op1, arg_Cn, arg_Cm, arg_immediate_0_7_op2}},
	// TBNZ <R><t>, #<imm>, <label>
	TBNZril: {0x37000000, "TBNZ", instArgs{arg_Rt_31_1__W_0__X_1, arg_immediate_0_63_b5_b40, arg_slabel_imm14_2}},
	// TBZ <R><t>, #<imm>, <label>
	TBZril: {0x36000000, "TBZ", instArgs{arg_Rt_31_1__W_0__X_1, arg_immediate_0_63_b5_b40, arg_slabel_imm14_2}},
	// TLBI <tlbi>, {<Xt>}
	TLBIix: {0xd5088000, "TLBI", instArgs{arg_sysop_TLBI_SYS_CR_system}},
	// TST <Xn>, #<imm>
	TSTxi: {0xf200001f, "TST", instArgs{arg_Xn, arg_immediate_bitmask_64_N_imms_immr}},
	// TST <Wn>, #<imm>
	TSTwi: {0x7200001f, "TST", instArgs{arg_Wn, arg_immediate_bitmask_32_imms_immr}},
	// TST <Wn>, <Wm> {, <shift> #<amount> }
	TSTwws: {0x6a00001f, "TST", instArgs{arg_Wn, arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31}},
	// TST <Xn>, <Xm> {, <shift> #<amount> }
	TSTxxs: {0xea00001f, "TST", instArgs{arg_Xn, arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63}},
	// UBFIZ <Wd>, <Wn>, #<lsb>, #<width>
	UBFIZwwii: {0x53000000, "UBFIZ", instArgs{arg_Wd, arg_Wn, arg_immediate_UBFIZ_UBFM_32M_bitfield_lsb_32_immr, arg_immediate_UBFIZ_UBFM_32M_bitfield_width_32_imms}},
	// UBFIZ <Xd>, <Xn>, #<lsb>, #<width>
	UBFIZxxii: {0xd3400000, "UBFIZ", instArgs{arg_Xd, arg_Xn, arg_immediate_UBFIZ_UBFM_64M_bitfield_lsb_64_immr, arg_immediate_UBFIZ_UBFM_64M_bitfield_width_64_imms}},
	// UBFM <Wd>, <Wn>, #<immr>, #<imms>
	UBFMwwii: {0x53000000, "UBFM", instArgs{arg_Wd, arg_Wn, arg_immediate_0_31_immr, arg_immediate_0_31_imms}},
	// UBFM <Xd>, <Xn>, #<immr>, #<imms>
	UBFMxxii: {0xd3400000, "UBFM", instArgs{arg_Xd, arg_Xn, arg_immediate_0_63_immr, arg_immediate_0_63_imms}},
	// UBFX <Wd>, <Wn>, #<lsb>, #<width>
	UBFXwwii: {0x53000000, "UBFX", instArgs{arg_Wd, arg_Wn, arg_immediate_UBFX_UBFM_32M_bitfield_lsb_32_immr, arg_immediate_UBFX_UBFM_32M_bitfield_width_32_imms}},
	// UBFX <Xd>, <Xn>, #<lsb>, #<width>
	UBFXxxii: {0xd3400000, "UBFX", instArgs{arg_Xd, arg_Xn, arg_immediate_UBFX_UBFM_64M_bitfield_lsb_64_immr, arg_immediate_UBFX_UBFM_64M_bitfield_width_64_imms}},
	// UDF #<imm>
	UDFi: {0x0000ffff, "UDF", instArgs{}},
	// UDIV <Wd>, <Wn>, <Wm>
	UDIVwww: {0x1ac00800, "UDIV", instArgs{arg_Wd, arg_Wn, arg_Wm}},
	// UDIV <Xd>, <Xn>, <Xm>
	UDIVxxx: {0x9ac00800, "UDIV", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// UMADDL <Xd>, <Wn>, <Wm>, <Xa>
	UMADDLxwwx: {0x9ba00000, "UMADDL", instArgs{arg_Xd, arg_Wn, arg_Wm, arg_Xa}},
	// UMNEGL <Xd>, <Wn>, <Wm>
	UMNEGLxww: {0x9ba0fc00, "UMNEGL", instArgs{arg_Xd, arg_Wn, arg_Wm}},
	// UMSUBL <Xd>, <Wn>, <Wm>, <Xa>
	UMSUBLxwwx: {0x9ba08000, "UMSUBL", instArgs{arg_Xd, arg_Wn, arg_Wm, arg_Xa}},
	// UMULH <Xd>, <Xn>, <Xm>
	UMULHxxx: {0x9bc07c00, "UMULH", instArgs{arg_Xd, arg_Xn, arg_Xm}},
	// UMULL <Xd>, <Wn>, <Wm>
	UMULLxww: {0x9ba07c00, "UMULL", instArgs{arg_Xd, arg_Wn, arg_Wm}},
	// UXTB <Wd>, <Wn>
	UXTBww: {0x53001c00, "UXTB", instArgs{arg_Wd, arg_Wn}},
	// UXTH <Wd>, <Wn>
	UXTHww: {0x53003c00, "UXTH", instArgs{arg_Wd, arg_Wn}},
	// WFE
	WFE: {0xd503205f, "WFE", instArgs{}},
	// WFI
	WFI: {0xd503207f, "WFI", instArgs{}},
	// YIELD
	YIELD: {0xd503203f, "YIELD", instArgs{}},

	// SIMD&FP instructions
	// ABS <V><d>, <V><n>
	ABSvv: {0x5e20b800, "ABS", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3}},
	// ABS <Vd>.<t>, <Vn>.<t>
	ABSvv_t: {0x0e20b800, "ABS", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// ADD <V><d>, <V><n>, <V><m>
	ADDvvv: {0x5e208400, "ADD", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// ADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ADDvvv_t: {0x0e208400, "ADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// ADDHN <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	ADDHNvvv_t: {0x0e204000, "ADDHN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// ADDHN2 <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	ADDHN2vvv_t: {0x4e204000, "ADDHN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// ADDP <V><d>, <Vn>.<t>
	ADDPvv_t: {0x5e31b800, "ADDP", instArgs{arg_Vd_22_2__D_3, arg_Vn_arrangement_size___2D_3}},
	// ADDP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ADDPvvv_t: {0x0e20bc00, "ADDP", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// ADDV <V><d>, <Vn>.<t>
	ADDVvv_t: {0x0e31b800, "ADDV", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// AESD <Vd>.16B, <Vn>.16B
	AESDvv: {0x4e285800, "AESD", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B}},
	// AESE <Vd>.16B, <Vn>.16B
	AESEvv: {0x4e284800, "AESE", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B}},
	// AESIMC <Vd>.16B, <Vn>.16B
	AESIMCvv: {0x4e287800, "AESIMC", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B}},
	// AESMC <Vd>.16B, <Vn>.16B
	AESMCvv: {0x4e286800, "AESMC", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B}},
	// AND <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ANDvvv_t: {0x0e201c00, "AND", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// BCAX <Vd>.16B, <Vn>.16B, <Vm>.16B, <Va>.16B
	BCAXvvv: {0xce200000, "BCAX", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B, arg_Vm_arrangement_16B, arg_Va_arrangement_16B}},
	// BIC <Vd>.<t>, #<imm8>{, LSL #<amount>}
	BICvis_h: {0x2f009400, "BIC", instArgs{arg_Vd_arrangement_Q___4H_0__8H_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1}},
	// BIC <Vd>.<t_1>, #<imm8>{, LSL #<amount>}
	BICvis_s: {0x2f001400, "BIC", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1__16_2__24_3}},
	// BIC <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	BICvvv_t: {0x0e601c00, "BIC", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// BIF <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	BIFvvv_t: {0x2ee01c00, "BIF", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// BIT <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	BITvvv_t: {0x2ea01c00, "BIT", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// BSL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	BSLvvv_t: {0x2e601c00, "BSL", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// CLS <Vd>.<t>, <Vn>.<t>
	CLSvv_t: {0x0e204800, "CLS", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// CLZ <Vd>.<t>, <Vn>.<t>
	CLZvv_t: {0x2e204800, "CLZ", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// CMEQ <V><d>, <V><n>, <V><m>
	CMEQvvv: {0x7e208c00, "CMEQ", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMEQ <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMEQvvv_t: {0x2e208c00, "CMEQ", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CMEQ <Vd>.<t>, <Vn>.<t>, #0
	CMEQvv_t: {0x0e209800, "CMEQ", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_immediate_zero}},
	// CMEQ <V><d>, <V><n>, #0
	CMEQvv: {0x5e209800, "CMEQ", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_immediate_zero}},
	// CMGE <V><d>, <V><n>, <V><m>
	CMGEvvv: {0x5e203c00, "CMGE", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMGE <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMGEvvv_t: {0x0e203c00, "CMGE", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CMGE <V><d>, <V><n>, #0
	CMGEvv: {0x7e208800, "CMGE", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_immediate_zero}},
	// CMGE <Vd>.<t>, <Vn>.<t>, #0
	CMGEvv_t: {0x2e208800, "CMGE", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_immediate_zero}},
	// CMGT <V><d>, <V><n>, <V><m>
	CMGTvvv: {0x5e203400, "CMGT", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMGT <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMGTvvv_t: {0x0e203400, "CMGT", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CMGT <V><d>, <V><n>, #0
	CMGTvv: {0x5e208800, "CMGT", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_immediate_zero}},
	// CMGT <Vd>.<t>, <Vn>.<t>, #0
	CMGTvv_t: {0x0e208800, "CMGT", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_immediate_zero}},
	// CMHI <V><d>, <V><n>, <V><m>
	CMHIvvv: {0x7e203400, "CMHI", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMHI <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMHIvvv_t: {0x2e203400, "CMHI", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CMHS <V><d>, <V><n>, <V><m>
	CMHSvvv: {0x7e203c00, "CMHS", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMHS <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMHSvvv_t: {0x2e203c00, "CMHS", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CMLE <V><d>, <V><n>, #0
	CMLEvv: {0x7e209800, "CMLE", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_immediate_zero}},
	// CMLE <Vd>.<t>, <Vn>.<t>, #0
	CMLEvv_t: {0x2e209800, "CMLE", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_immediate_zero}},
	// CMLT <V><d>, <V><n>, #0
	CMLTvv: {0x5e20a800, "CMLT", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_immediate_zero}},
	// CMLT <Vd>.<t>, <Vn>.<t>, #0
	CMLTvv_t: {0x0e20a800, "CMLT", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_immediate_zero}},
	// CMTST <V><d>, <V><n>, <V><m>
	CMTSTvvv: {0x5e208c00, "CMTST", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// CMTST <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	CMTSTvvv_t: {0x0e208c00, "CMTST", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// CNT <Vd>.<t>, <Vn>.<t>
	CNTvv_t: {0x0e205800, "CNT", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01, arg_Vn_arrangement_size_Q___8B_00__16B_01}},
	// DUP <Vd>.<t>, <R><n>
	DUPvr_t: {0x0e000c00, "DUP", instArgs{arg_Vd_arrangement_imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Rn_16_5__W_1__W_2__W_4__X_8}},
	// DUP <V><d>, <Vn>.<t_1>[<index>]
	DUPvv_i: {0x5e000400, "DUP", instArgs{arg_Vd_16_5__B_1__H_2__S_4__D_8, arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1}},
	// DUP <Vd>.<t>, <Vn>.<ts>[<index>]
	DUPvv_ti: {0x0e000400, "DUP", instArgs{arg_Vd_arrangement_imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1}},
	// EOR <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	EORvvv_t: {0x2e201c00, "EOR", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// EOR3 <Vd>.16B, <Vn>.16B, <Vm>.16B, <Va>.16B
	EOR3vvv: {0xce000000, "EOR3", instArgs{arg_Vd_arrangement_16B, arg_Vn_arrangement_16B, arg_Vm_arrangement_16B, arg_Va_arrangement_16B}},
	// EXT <Vd>.<t>, <Vn>.<t>, <Vm>.<t>, #<index>
	EXTvvvi_t: {0x2e000000, "EXT", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1, arg_immediate_index_Q_imm4__imm4lt20gt_00__imm4_10}},
	// FABD <V><d>, <V><n>, <V><m>
	FABDvvv: {0x7ea0d400, "FABD", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FABD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FABDvvv_t: {0x2ea0d400, "FABD", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FABS <Vd>.<t>, <Vn>.<t>
	FABSvv_t: {0x0ea0f800, "FABS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FABS <Sd>, <Sn>
	FABSss: {0x1e20c000, "FABS", instArgs{arg_Sd, arg_Sn}},
	// FABS <Dd>, <Dn>
	FABSdd: {0x1e60c000, "FABS", instArgs{arg_Dd, arg_Dn}},
	// FACGE <V><d>, <V><n>, <V><m>
	FACGEvvv: {0x7e20ec00, "FACGE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FACGE <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FACGEvvv_t: {0x2e20ec00, "FACGE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FACGT <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FACGTvvv_t: {0x2ea0ec00, "FACGT", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FACGT <V><d>, <V><n>, <V><m>
	FACGTvvv: {0x7ea0ec00, "FACGT", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FADDvvv_t: {0x0e20d400, "FADD", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FADD <Sd>, <Sn>, <Sm>
	FADDsss: {0x1e202800, "FADD", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FADD <Dd>, <Dn>, <Dm>
	FADDddd: {0x1e602800, "FADD", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FADDP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FADDPvvv_t: {0x2e20d400, "FADDP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FADDP <V><d>, <Vn>.<t>
	FADDPvv_t: {0x7e30d800, "FADDP", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_arrangement_sz___2S_0__2D_1}},
	// FCCMP <Dn>, <Dm>, #<nzcv>, <cond>
	FCCMPddic: {0x1e600400, "FCCMP", instArgs{arg_Dn, arg_Dm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// FCCMP <Sn>, <Sm>, #<nzcv>, <cond>
	FCCMPssic: {0x1e200400, "FCCMP", instArgs{arg_Sn, arg_Sm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// FCCMPE <Dn>, <Dm>, #<nzcv>, <cond>
	FCCMPEddic: {0x1e600410, "FCCMPE", instArgs{arg_Dn, arg_Dm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// FCCMPE <Sn>, <Sm>, #<nzcv>, <cond>
	FCCMPEssic: {0x1e200410, "FCCMPE", instArgs{arg_Sn, arg_Sm, arg_immediate_0_15_nzcv, arg_cond_AllowALNV_Normal}},
	// FCMEQ <V><d>, <V><n>, #0.0
	FCMEQvv: {0x5ea0d800, "FCMEQ", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_immediate_floatzero}},
	// FCMEQ <Vd>.<t>, <Vn>.<t>, #0.0
	FCMEQvv_t: {0x0ea0d800, "FCMEQ", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_immediate_floatzero}},
	// FCMEQ <V><d>, <V><n>, <V><m>
	FCMEQvvv: {0x5e20e400, "FCMEQ", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FCMEQ <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FCMEQvvv_t: {0x0e20e400, "FCMEQ", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCMGE <V><d>, <V><n>, #0.0
	FCMGEvv: {0x7ea0c800, "FCMGE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_immediate_floatzero}},
	// FCMGE <Vd>.<t>, <Vn>.<t>, #0.0
	FCMGEvv_t: {0x2ea0c800, "FCMGE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_immediate_floatzero}},
	// FCMGE <V><d>, <V><n>, <V><m>
	FCMGEvvv: {0x7e20e400, "FCMGE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FCMGE <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FCMGEvvv_t: {0x2e20e400, "FCMGE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCMGT <V><d>, <V><n>, #0.0
	FCMGTvv: {0x5ea0c800, "FCMGT", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_immediate_floatzero}},
	// FCMGT <Vd>.<t>, <Vn>.<t>, #0.0
	FCMGTvv_t: {0x0ea0c800, "FCMGT", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_immediate_floatzero}},
	// FCMGT <V><d>, <V><n>, <V><m>
	FCMGTvvv: {0x7ea0e400, "FCMGT", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FCMGT <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FCMGTvvv_t: {0x2ea0e400, "FCMGT", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCMLE <V><d>, <V><n>, #0.0
	FCMLEvv: {0x7ea0d800, "FCMLE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_immediate_floatzero}},
	// FCMLE <Vd>.<t>, <Vn>.<t>, #0.0
	FCMLEvv_t: {0x2ea0d800, "FCMLE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_immediate_floatzero}},
	// FCMLT <V><d>, <V><n>, #0.0
	FCMLTvv: {0x5ea0e800, "FCMLT", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_immediate_floatzero}},
	// FCMLT <Vd>.<t>, <Vn>.<t>, #0.0
	FCMLTvv_t: {0x0ea0e800, "FCMLT", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_immediate_floatzero}},
	// FCMP <Sn>, <Sm>
	FCMPss: {0x1e202000, "FCMP", instArgs{arg_Sn, arg_Sm}},
	// FCMP <Dn>, <Dm>
	FCMPdd: {0x1e602000, "FCMP", instArgs{arg_Dn, arg_Dm}},
	// FCMP <Dn>, #0.0
	FCMPd0: {0x1e602008, "FCMP", instArgs{arg_Dn, arg_immediate_floatzero}},
	// FCMP <Sn>, #0.0
	FCMPs0: {0x1e202008, "FCMP", instArgs{arg_Sn, arg_immediate_floatzero}},
	// FCMPE <Dn>, <Dm>
	FCMPEdd: {0x1e602010, "FCMPE", instArgs{arg_Dn, arg_Dm}},
	// FCMPE <Dn>, #0.0
	FCMPEd0: {0x1e602018, "FCMPE", instArgs{arg_Dn, arg_immediate_floatzero}},
	// FCMPE <Sn>, #0.0
	FCMPEs0: {0x1e202018, "FCMPE", instArgs{arg_Sn, arg_immediate_floatzero}},
	// FCMPE <Sn>, <Sm>
	FCMPEss: {0x1e202010, "FCMPE", instArgs{arg_Sn, arg_Sm}},
	// FCSEL <Sd>, <Sn>, <Sm>, <cond>
	FCSELsssc: {0x1e200c00, "FCSEL", instArgs{arg_Sd, arg_Sn, arg_Sm, arg_cond_AllowALNV_Normal}},
	// FCSEL <Dd>, <Dn>, <Dm>, <cond>
	FCSELdddc: {0x1e600c00, "FCSEL", instArgs{arg_Dd, arg_Dn, arg_Dm, arg_cond_AllowALNV_Normal}},
	// FCVT <Hd>, <Dn>
	FCVThd: {0x1e63c000, "FCVT", instArgs{arg_Hd, arg_Dn}},
	// FCVT <Sd>, <Hn>
	FCVTsh: {0x1ee24000, "FCVT", instArgs{arg_Sd, arg_Hn}},
	// FCVT <Dd>, <Hn>
	FCVTdh: {0x1ee2c000, "FCVT", instArgs{arg_Dd, arg_Hn}},
	// FCVT <Hd>, <Sn>
	FCVThs: {0x1e23c000, "FCVT", instArgs{arg_Hd, arg_Sn}},
	// FCVT <Dd>, <Sn>
	FCVTds: {0x1e22c000, "FCVT", instArgs{arg_Dd, arg_Sn}},
	// FCVT <Sd>, <Dn>
	FCVTsd: {0x1e624000, "FCVT", instArgs{arg_Sd, arg_Dn}},
	// FCVTAS <Vd>.<t>, <Vn>.<t>
	FCVTASvv_t: {0x0e21c800, "FCVTAS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTAS <Wd>, <Sn>
	FCVTASws: {0x1e240000, "FCVTAS", instArgs{arg_Wd, arg_Sn}},
	// FCVTAS <Xd>, <Sn>
	FCVTASxs: {0x9e240000, "FCVTAS", instArgs{arg_Xd, arg_Sn}},
	// FCVTAS <Wd>, <Dn>
	FCVTASwd: {0x1e640000, "FCVTAS", instArgs{arg_Wd, arg_Dn}},
	// FCVTAS <Xd>, <Dn>
	FCVTASxd: {0x9e640000, "FCVTAS", instArgs{arg_Xd, arg_Dn}},
	// FCVTAS <V><d>, <V><n>
	FCVTASvv: {0x5e21c800, "FCVTAS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTAU <Vd>.<t>, <Vn>.<t>
	FCVTAUvv_t: {0x2e21c800, "FCVTAU", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTAU <Wd>, <Sn>
	FCVTAUws: {0x1e250000, "FCVTAU", instArgs{arg_Wd, arg_Sn}},
	// FCVTAU <Xd>, <Sn>
	FCVTAUxs: {0x9e250000, "FCVTAU", instArgs{arg_Xd, arg_Sn}},
	// FCVTAU <Wd>, <Dn>
	FCVTAUwd: {0x1e650000, "FCVTAU", instArgs{arg_Wd, arg_Dn}},
	// FCVTAU <Xd>, <Dn>
	FCVTAUxd: {0x9e650000, "FCVTAU", instArgs{arg_Xd, arg_Dn}},
	// FCVTAU <V><d>, <V><n>
	FCVTAUvv: {0x7e21c800, "FCVTAU", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTL <Vd>.<ta>, <Vn>.<tb>
	FCVTLvv_t: {0x0e217800, "FCVTL", instArgs{arg_Vd_arrangement_sz___4S_0__2D_1, arg_Vn_arrangement_sz_Q___4H_00__8H_01__2S_10__4S_11}},
	// FCVTL2 <Vd>.<ta>, <Vn>.<tb>
	FCVTL2vv_t: {0x4e217800, "FCVTL2", instArgs{arg_Vd_arrangement_sz___4S_0__2D_1, arg_Vn_arrangement_sz_Q___4H_00__8H_01__2S_10__4S_11}},
	// FCVTMS <Vd>.<t>, <Vn>.<t>
	FCVTMSvv_t: {0x0e21b800, "FCVTMS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTMS <Wd>, <Sn>
	FCVTMSws: {0x1e300000, "FCVTMS", instArgs{arg_Wd, arg_Sn}},
	// FCVTMS <Xd>, <Sn>
	FCVTMSxs: {0x9e300000, "FCVTMS", instArgs{arg_Xd, arg_Sn}},
	// FCVTMS <Wd>, <Dn>
	FCVTMSwd: {0x1e700000, "FCVTMS", instArgs{arg_Wd, arg_Dn}},
	// FCVTMS <Xd>, <Dn>
	FCVTMSxd: {0x9e700000, "FCVTMS", instArgs{arg_Xd, arg_Dn}},
	// FCVTMS <V><d>, <V><n>
	FCVTMSvv: {0x5e21b800, "FCVTMS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTMU <Vd>.<t>, <Vn>.<t>
	FCVTMUvv_t: {0x2e21b800, "FCVTMU", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTMU <Wd>, <Sn>
	FCVTMUws: {0x1e310000, "FCVTMU", instArgs{arg_Wd, arg_Sn}},
	// FCVTMU <Xd>, <Sn>
	FCVTMUxs: {0x9e310000, "FCVTMU", instArgs{arg_Xd, arg_Sn}},
	// FCVTMU <Wd>, <Dn>
	FCVTMUwd: {0x1e710000, "FCVTMU", instArgs{arg_Wd, arg_Dn}},
	// FCVTMU <Xd>, <Dn>
	FCVTMUxd: {0x9e710000, "FCVTMU", instArgs{arg_Xd, arg_Dn}},
	// FCVTMU <V><d>, <V><n>
	FCVTMUvv: {0x7e21b800, "FCVTMU", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTN <Vd>.<tb>, <Vn>.<ta>
	FCVTNvv_t: {0x0e216800, "FCVTN", instArgs{arg_Vd_arrangement_sz_Q___4H_00__8H_01__2S_10__4S_11, arg_Vn_arrangement_sz___4S_0__2D_1}},
	// FCVTN2 <Vd>.<tb>, <Vn>.<ta>
	FCVTN2vv_t: {0x4e216800, "FCVTN2", instArgs{arg_Vd_arrangement_sz_Q___4H_00__8H_01__2S_10__4S_11, arg_Vn_arrangement_sz___4S_0__2D_1}},
	// FCVTNS <Vd>.<t>, <Vn>.<t>
	FCVTNSvv_t: {0x0e21a800, "FCVTNS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTNS <Wd>, <Sn>
	FCVTNSws: {0x1e200000, "FCVTNS", instArgs{arg_Wd, arg_Sn}},
	// FCVTNS <Xd>, <Sn>
	FCVTNSxs: {0x9e200000, "FCVTNS", instArgs{arg_Xd, arg_Sn}},
	// FCVTNS <Wd>, <Dn>
	FCVTNSwd: {0x1e600000, "FCVTNS", instArgs{arg_Wd, arg_Dn}},
	// FCVTNS <Xd>, <Dn>
	FCVTNSxd: {0x9e600000, "FCVTNS", instArgs{arg_Xd, arg_Dn}},
	// FCVTNS <V><d>, <V><n>
	FCVTNSvv: {0x5e21a800, "FCVTNS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTNU <Vd>.<t>, <Vn>.<t>
	FCVTNUvv_t: {0x2e21a800, "FCVTNU", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTNU <Wd>, <Sn>
	FCVTNUws: {0x1e210000, "FCVTNU", instArgs{arg_Wd, arg_Sn}},
	// FCVTNU <Xd>, <Sn>
	FCVTNUxs: {0x9e210000, "FCVTNU", instArgs{arg_Xd, arg_Sn}},
	// FCVTNU <Wd>, <Dn>
	FCVTNUwd: {0x1e610000, "FCVTNU", instArgs{arg_Wd, arg_Dn}},
	// FCVTNU <Xd>, <Dn>
	FCVTNUxd: {0x9e610000, "FCVTNU", instArgs{arg_Xd, arg_Dn}},
	// FCVTNU <V><d>, <V><n>
	FCVTNUvv: {0x7e21a800, "FCVTNU", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTPS <Vd>.<t>, <Vn>.<t>
	FCVTPSvv_t: {0x0ea1a800, "FCVTPS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTPS <Wd>, <Sn>
	FCVTPSws: {0x1e280000, "FCVTPS", instArgs{arg_Wd, arg_Sn}},
	// FCVTPS <Xd>, <Sn>
	FCVTPSxs: {0x9e280000, "FCVTPS", instArgs{arg_Xd, arg_Sn}},
	// FCVTPS <Wd>, <Dn>
	FCVTPSwd: {0x1e680000, "FCVTPS", instArgs{arg_Wd, arg_Dn}},
	// FCVTPS <Xd>, <Dn>
	FCVTPSxd: {0x9e680000, "FCVTPS", instArgs{arg_Xd, arg_Dn}},
	// FCVTPS <V><d>, <V><n>
	FCVTPSvv: {0x5ea1a800, "FCVTPS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTPU <Vd>.<t>, <Vn>.<t>
	FCVTPUvv_t: {0x2ea1a800, "FCVTPU", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTPU <Wd>, <Sn>
	FCVTPUws: {0x1e290000, "FCVTPU", instArgs{arg_Wd, arg_Sn}},
	// FCVTPU <Xd>, <Sn>
	FCVTPUxs: {0x9e290000, "FCVTPU", instArgs{arg_Xd, arg_Sn}},
	// FCVTPU <Wd>, <Dn>
	FCVTPUwd: {0x1e690000, "FCVTPU", instArgs{arg_Wd, arg_Dn}},
	// FCVTPU <Xd>, <Dn>
	FCVTPUxd: {0x9e690000, "FCVTPU", instArgs{arg_Xd, arg_Dn}},
	// FCVTPU <V><d>, <V><n>
	FCVTPUvv: {0x7ea1a800, "FCVTPU", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTXN <Vd>.<tb>, <Vn>.<ta>
	FCVTXNvv_t: {0x2e216800, "FCVTXN", instArgs{arg_Vd_arrangement_sz_Q___2S_10__4S_11, arg_Vn_arrangement_sz___2D_1}},
	// FCVTXN <V><d>, <V><n>
	FCVTXNvv: {0x7e216800, "FCVTXN", instArgs{arg_Vd_22_1__S_1, arg_Vn_22_1__D_1}},
	// FCVTXN2 <Vd>.<tb>, <Vn>.<ta>
	FCVTXN2vv_t: {0x6e216800, "FCVTXN2", instArgs{arg_Vd_arrangement_sz_Q___2S_10__4S_11, arg_Vn_arrangement_sz___2D_1}},
	// FCVTZS <Wd>, <Dn>, #<fbits>
	FCVTZSwdi: {0x1e580000, "FCVTZS", instArgs{arg_Wd, arg_Dn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// FCVTZS <Wd>, <Sn>, #<fbits>
	FCVTZSwsi: {0x1e180000, "FCVTZS", instArgs{arg_Wd, arg_Sn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// FCVTZS <Xd>, <Dn>, #<fbits>
	FCVTZSxdi: {0x9e580000, "FCVTZS", instArgs{arg_Xd, arg_Dn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// FCVTZS <Xd>, <Sn>, #<fbits>
	FCVTZSxsi: {0x9e180000, "FCVTZS", instArgs{arg_Xd, arg_Sn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// FCVTZS <Wd>, <Dn>
	FCVTZSwd: {0x1e780000, "FCVTZS", instArgs{arg_Wd, arg_Dn}},
	// FCVTZS <Wd>, <Sn>
	FCVTZSws: {0x1e380000, "FCVTZS", instArgs{arg_Wd, arg_Sn}},
	// FCVTZS <Xd>, <Dn>
	FCVTZSxd: {0x9e780000, "FCVTZS", instArgs{arg_Xd, arg_Dn}},
	// FCVTZS <Xd>, <Sn>
	FCVTZSxs: {0x9e380000, "FCVTZS", instArgs{arg_Xd, arg_Sn}},
	// FCVTZS <V><d>, <V><n>, #<fbits>
	FCVTZSvvi: {0x5f00fc00, "FCVTZS", instArgs{arg_Vd_19_4__S_4__D_8, arg_Vn_19_4__S_4__D_8, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// FCVTZS <Vd>.<t>, <Vn>.<t>, #<fbits>
	FCVTZSvvi_t: {0x0f00fc00, "FCVTZS", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// FCVTZS <V><d>, <V><n>
	FCVTZSvv: {0x5ea1b800, "FCVTZS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTZS <Vd>.<t>, <Vn>.<t>
	FCVTZSvv_t: {0x0ea1b800, "FCVTZS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FCVTZU <Wd>, <Dn>, #<fbits>
	FCVTZUwdi: {0x1e590000, "FCVTZU", instArgs{arg_Wd, arg_Dn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// FCVTZU <Wd>, <Sn>, #<fbits>
	FCVTZUwsi: {0x1e190000, "FCVTZU", instArgs{arg_Wd, arg_Sn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// FCVTZU <Xd>, <Dn>, #<fbits>
	FCVTZUxdi: {0x9e590000, "FCVTZU", instArgs{arg_Xd, arg_Dn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// FCVTZU <Xd>, <Sn>, #<fbits>
	FCVTZUxsi: {0x9e190000, "FCVTZU", instArgs{arg_Xd, arg_Sn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// FCVTZU <Wd>, <Dn>
	FCVTZUwd: {0x1e790000, "FCVTZU", instArgs{arg_Wd, arg_Dn}},
	// FCVTZU <Wd>, <Sn>
	FCVTZUws: {0x1e390000, "FCVTZU", instArgs{arg_Wd, arg_Sn}},
	// FCVTZU <Xd>, <Dn>
	FCVTZUxd: {0x9e790000, "FCVTZU", instArgs{arg_Xd, arg_Dn}},
	// FCVTZU <Xd>, <Sn>
	FCVTZUxs: {0x9e390000, "FCVTZU", instArgs{arg_Xd, arg_Sn}},
	// FCVTZU <V><d>, <V><n>, #<fbits>
	FCVTZUvvi: {0x7f00fc00, "FCVTZU", instArgs{arg_Vd_19_4__S_4__D_8, arg_Vn_19_4__S_4__D_8, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// FCVTZU <Vd>.<t>, <Vn>.<t>, #<fbits>
	FCVTZUvvi_t: {0x2f00fc00, "FCVTZU", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// FCVTZU <V><d>, <V><n>
	FCVTZUvv: {0x7ea1b800, "FCVTZU", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FCVTZU <Vd>.<t>, <Vn>.<t>
	FCVTZUvv_t: {0x2ea1b800, "FCVTZU", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FDIV <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FDIVvvv_t: {0x2e20fc00, "FDIV", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FDIV <Sd>, <Sn>, <Sm>
	FDIVsss: {0x1e201800, "FDIV", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FDIV <Dd>, <Dn>, <Dm>
	FDIVddd: {0x1e601800, "FDIV", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMADD <Sd>, <Sn>, <Sm>, <Sa>
	FMADDssss: {0x1f000000, "FMADD", instArgs{arg_Sd, arg_Sn, arg_Sm, arg_Sa}},
	// FMADD <Dd>, <Dn>, <Dm>, <Da>
	FMADDdddd: {0x1f400000, "FMADD", instArgs{arg_Dd, arg_Dn, arg_Dm, arg_Da}},
	// FMAX <Sd>, <Sn>, <Sm>
	FMAXsss: {0x1e204800, "FMAX", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FMAX <Dd>, <Dn>, <Dm>
	FMAXddd: {0x1e604800, "FMAX", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMAX <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMAXvvv_t: {0x0e20f400, "FMAX", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMAXNM <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMAXNMvvv_t: {0x0e20c400, "FMAXNM", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMAXNM <Sd>, <Sn>, <Sm>
	FMAXNMsss: {0x1e206800, "FMAXNM", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FMAXNM <Dd>, <Dn>, <Dm>
	FMAXNMddd: {0x1e606800, "FMAXNM", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMAXNMP <V><d>, <Vn>.<t>
	FMAXNMPvv_t: {0x7e30c800, "FMAXNMP", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_arrangement_sz___2S_0__2D_1}},
	// FMAXNMP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMAXNMPvvv_t: {0x2e20c400, "FMAXNMP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMAXNMV <V><d>, <Vn>.<t>
	FMAXNMVvv_t: {0x2e30c800, "FMAXNMV", instArgs{arg_Vd_22_1__S_0, arg_Vn_arrangement_Q_sz___4S_10}},
	// FMAXP <V><d>, <Vn>.<t>
	FMAXPvv_t: {0x7e30f800, "FMAXP", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_arrangement_sz___2S_0__2D_1}},
	// FMAXP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMAXPvvv_t: {0x2e20f400, "FMAXP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMAXV <V><d>, <Vn>.<t>
	FMAXVvv_t: {0x2e30f800, "FMAXV", instArgs{arg_Vd_22_1__S_0, arg_Vn_arrangement_Q_sz___4S_10}},
	// FMIN <Dd>, <Dn>, <Dm>
	FMINddd: {0x1e605800, "FMIN", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMIN <Sd>, <Sn>, <Sm>
	FMINsss: {0x1e205800, "FMIN", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FMIN <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMINvvv_t: {0x0ea0f400, "FMIN", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMINNM <Dd>, <Dn>, <Dm>
	FMINNMddd: {0x1e607800, "FMINNM", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMINNM <Sd>, <Sn>, <Sm>
	FMINNMsss: {0x1e207800, "FMINNM", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FMINNM <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMINNMvvv_t: {0x0ea0c400, "FMINNM", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMINNMP <V><d>, <Vn>.<t>
	FMINNMPvv_t: {0x7eb0c800, "FMINNMP", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_arrangement_sz___2S_0__2D_1}},
	// FMINNMP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMINNMPvvv_t: {0x2ea0c400, "FMINNMP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMINNMV <V><d>, <Vn>.<t>
	FMINNMVvv_t: {0x2eb0c800, "FMINNMV", instArgs{arg_Vd_22_1__S_0, arg_Vn_arrangement_Q_sz___4S_10}},
	// FMINP <V><d>, <Vn>.<t>
	FMINPvv_t: {0x7eb0f800, "FMINP", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_arrangement_sz___2S_0__2D_1}},
	// FMINP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMINPvvv_t: {0x2ea0f400, "FMINP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMINV <V><d>, <Vn>.<t>
	FMINVvv_t: {0x2eb0f800, "FMINV", instArgs{arg_Vd_22_1__S_0, arg_Vn_arrangement_Q_sz___4S_10}},
	// FMLA <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	FMLAvvv_i: {0x5f801000, "FMLA", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMLA <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	FMLAvvv_ti: {0x0f801000, "FMLA", instArgs{arg_Vd_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vn_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMLA <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMLAvvv_t: {0x0e20cc00, "FMLA", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMLS <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMLSvvv_t: {0x0ea0cc00, "FMLS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMLS <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	FMLSvvv_i: {0x5f805000, "FMLS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMLS <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	FMLSvvv_ti: {0x0f805000, "FMLS", instArgs{arg_Vd_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vn_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMOV <Sd>, <Wn>
	FMOVsw: {0x1e270000, "FMOV", instArgs{arg_Sd, arg_Wn}},
	// FMOV <Wd>, <Sn>
	FMOVws: {0x1e260000, "FMOV", instArgs{arg_Wd, arg_Sn}},
	// FMOV <Dd>, <Xn>
	FMOVdx: {0x9e670000, "FMOV", instArgs{arg_Dd, arg_Xn}},
	// FMOV <Xd>, <Dn>
	FMOVxd: {0x9e660000, "FMOV", instArgs{arg_Xd, arg_Dn}},
	// FMOV <Vd>.D[1], <Xn>
	FMOVvx: {0x9eaf0000, "FMOV", instArgs{arg_Vd_arrangement_D_index__1, arg_Xn}},
	// FMOV <Xd>, <Vn>.D[1]
	FMOVxv: {0x9eae0000, "FMOV", instArgs{arg_Xd, arg_Vn_arrangement_D_index__1}},
	// FMOV <Sd>, <Sn>
	FMOVss: {0x1e204000, "FMOV", instArgs{arg_Sd, arg_Sn}},
	// FMOV <Dd>, <Dn>
	FMOVdd: {0x1e604000, "FMOV", instArgs{arg_Dd, arg_Dn}},
	// FMOV <Sd>, #<imm>
	FMOVsi: {0x1e201000, "FMOV", instArgs{arg_Sd, arg_immediate_exp_3_pre_4_imm8}},
	// FMOV <Dd>, #<imm>
	FMOVdi: {0x1e601000, "FMOV", instArgs{arg_Dd, arg_immediate_exp_3_pre_4_imm8}},
	// FMOV <Vd>.<t>, #<imm>
	FMOVvi_t: {0x0f00f400, "FMOV", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_exp_3_pre_4_a_b_c_d_e_f_g_h}},
	// FMOV <Vd>.2D, #<imm>
	FMOVvi_d: {0x6f00f400, "FMOV", instArgs{arg_Vd_arrangement_2D, arg_immediate_exp_3_pre_4_a_b_c_d_e_f_g_h}},
	// FMSUB <Sd>, <Sn>, <Sm>, <Sa>
	FMSUBssss: {0x1f008000, "FMSUB", instArgs{arg_Sd, arg_Sn, arg_Sm, arg_Sa}},
	// FMSUB <Dd>, <Dn>, <Dm>, <Da>
	FMSUBdddd: {0x1f408000, "FMSUB", instArgs{arg_Dd, arg_Dn, arg_Dm, arg_Da}},
	// FMUL <Dd>, <Dn>, <Dm>
	FMULddd: {0x1e600800, "FMUL", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FMUL <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	FMULvvv_i: {0x5f809000, "FMUL", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMUL <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	FMULvvv_ti: {0x0f809000, "FMUL", instArgs{arg_Vd_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vn_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMUL <Sd>, <Sn>, <Sm>
	FMULsss: {0x1e200800, "FMUL", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FMUL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMULvvv_t: {0x2e20dc00, "FMUL", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMULX <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	FMULXvvv_ti: {0x2f809000, "FMULX", instArgs{arg_Vd_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vn_arrangement_Q_sz___2S_00__4S_10__2D_11, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FMULX <V><d>, <V><n>, <V><m>
	FMULXvvv: {0x5e20dc00, "FMULX", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FMULX <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FMULXvvv_t: {0x0e20dc00, "FMULX", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FMULX <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	FMULXvvv_i: {0x7f809000, "FMULX", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_arrangement_sz___S_0__D_1_index__sz_L_H__HL_00__H_10_1}},
	// FNEG <Vd>.<t>, <Vn>.<t>
	FNEGvv_t: {0x2ea0f800, "FNEG", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FNEG <Sd>, <Sn>
	FNEGss: {0x1e214000, "FNEG", instArgs{arg_Sd, arg_Sn}},
	// FNEG <Dd>, <Dn>
	FNEGdd: {0x1e614000, "FNEG", instArgs{arg_Dd, arg_Dn}},
	// FNMADD <Sd>, <Sn>, <Sm>, <Sa>
	FNMADDssss: {0x1f200000, "FNMADD", instArgs{arg_Sd, arg_Sn, arg_Sm, arg_Sa}},
	// FNMADD <Dd>, <Dn>, <Dm>, <Da>
	FNMADDdddd: {0x1f600000, "FNMADD", instArgs{arg_Dd, arg_Dn, arg_Dm, arg_Da}},
	// FNMSUB <Sd>, <Sn>, <Sm>, <Sa>
	FNMSUBssss: {0x1f208000, "FNMSUB", instArgs{arg_Sd, arg_Sn, arg_Sm, arg_Sa}},
	// FNMSUB <Dd>, <Dn>, <Dm>, <Da>
	FNMSUBdddd: {0x1f608000, "FNMSUB", instArgs{arg_Dd, arg_Dn, arg_Dm, arg_Da}},
	// FNMUL <Dd>, <Dn>, <Dm>
	FNMULddd: {0x1e608800, "FNMUL", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FNMUL <Sd>, <Sn>, <Sm>
	FNMULsss: {0x1e208800, "FNMUL", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FRECPE <V><d>, <V><n>
	FRECPEvv: {0x5ea1d800, "FRECPE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FRECPE <Vd>.<t>, <Vn>.<t>
	FRECPEvv_t: {0x0ea1d800, "FRECPE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRECPS <V><d>, <V><n>, <V><m>
	FRECPSvvv: {0x5e20fc00, "FRECPS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FRECPS <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FRECPSvvv_t: {0x0e20fc00, "FRECPS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRECPX <V><d>, <V><n>
	FRECPXvv: {0x5ea1f800, "FRECPX", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FRINTA <Dd>, <Dn>
	FRINTAdd: {0x1e664000, "FRINTA", instArgs{arg_Dd, arg_Dn}},
	// FRINTA <Sd>, <Sn>
	FRINTAss: {0x1e264000, "FRINTA", instArgs{arg_Sd, arg_Sn}},
	// FRINTA <Vd>.<t>, <Vn>.<t>
	FRINTAvv_t: {0x2e218800, "FRINTA", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTI <Dd>, <Dn>
	FRINTIdd: {0x1e67c000, "FRINTI", instArgs{arg_Dd, arg_Dn}},
	// FRINTI <Vd>.<t>, <Vn>.<t>
	FRINTIvv_t: {0x2ea19800, "FRINTI", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTI <Sd>, <Sn>
	FRINTIss: {0x1e27c000, "FRINTI", instArgs{arg_Sd, arg_Sn}},
	// FRINTM <Dd>, <Dn>
	FRINTMdd: {0x1e654000, "FRINTM", instArgs{arg_Dd, arg_Dn}},
	// FRINTM <Vd>.<t>, <Vn>.<t>
	FRINTMvv_t: {0x0e219800, "FRINTM", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTM <Sd>, <Sn>
	FRINTMss: {0x1e254000, "FRINTM", instArgs{arg_Sd, arg_Sn}},
	// FRINTN <Dd>, <Dn>
	FRINTNdd: {0x1e644000, "FRINTN", instArgs{arg_Dd, arg_Dn}},
	// FRINTN <Sd>, <Sn>
	FRINTNss: {0x1e244000, "FRINTN", instArgs{arg_Sd, arg_Sn}},
	// FRINTN <Vd>.<t>, <Vn>.<t>
	FRINTNvv_t: {0x0e218800, "FRINTN", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTP <Sd>, <Sn>
	FRINTPss: {0x1e24c000, "FRINTP", instArgs{arg_Sd, arg_Sn}},
	// FRINTP <Dd>, <Dn>
	FRINTPdd: {0x1e64c000, "FRINTP", instArgs{arg_Dd, arg_Dn}},
	// FRINTP <Vd>.<t>, <Vn>.<t>
	FRINTPvv_t: {0x0ea18800, "FRINTP", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTX <Vd>.<t>, <Vn>.<t>
	FRINTXvv_t: {0x2e219800, "FRINTX", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRINTX <Sd>, <Sn>
	FRINTXss: {0x1e274000, "FRINTX", instArgs{arg_Sd, arg_Sn}},
	// FRINTX <Dd>, <Dn>
	FRINTXdd: {0x1e674000, "FRINTX", instArgs{arg_Dd, arg_Dn}},
	// FRINTZ <Sd>, <Sn>
	FRINTZss: {0x1e25c000, "FRINTZ", instArgs{arg_Sd, arg_Sn}},
	// FRINTZ <Dd>, <Dn>
	FRINTZdd: {0x1e65c000, "FRINTZ", instArgs{arg_Dd, arg_Dn}},
	// FRINTZ <Vd>.<t>, <Vn>.<t>
	FRINTZvv_t: {0x0ea19800, "FRINTZ", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRSQRTE <Vd>.<t>, <Vn>.<t>
	FRSQRTEvv_t: {0x2ea1d800, "FRSQRTE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRSQRTE <V><d>, <V><n>
	FRSQRTEvv: {0x7ea1d800, "FRSQRTE", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// FRSQRTS <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FRSQRTSvvv_t: {0x0ea0fc00, "FRSQRTS", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FRSQRTS <V><d>, <V><n>, <V><m>
	FRSQRTSvvv: {0x5ea0fc00, "FRSQRTS", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1, arg_Vm_22_1__S_0__D_1}},
	// FSQRT <Sd>, <Sn>
	FSQRTss: {0x1e21c000, "FSQRT", instArgs{arg_Sd, arg_Sn}},
	// FSQRT <Vd>.<t>, <Vn>.<t>
	FSQRTvv_t: {0x2ea1f800, "FSQRT", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// FSQRT <Dd>, <Dn>
	FSQRTdd: {0x1e61c000, "FSQRT", instArgs{arg_Dd, arg_Dn}},
	// FSUB <Sd>, <Sn>, <Sm>
	FSUBsss: {0x1e203800, "FSUB", instArgs{arg_Sd, arg_Sn, arg_Sm}},
	// FSUB <Dd>, <Dn>, <Dm>
	FSUBddd: {0x1e603800, "FSUB", instArgs{arg_Dd, arg_Dn, arg_Dm}},
	// FSUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	FSUBvvv_t: {0x0ea0d400, "FSUB", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// INS <Vd>.<ts>[<index>], <R><n>
	INSvr_i: {0x4e001c00, "INS", instArgs{arg_Vd_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1, arg_Rn_16_5__W_1__W_2__W_4__X_8}},
	// INS <Vd>.<ts>[<index1>], <Vn>.<ts>[<index2>]
	INSvv_i: {0x6e000400, "INS", instArgs{arg_Vd_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1, arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1}},
	// LD1 <Vt>.<t>, [<Xn|SP>]
	LD1vx_t1: {0x0c407000, "LD1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD1 <Vt>.<t>, [<Xn|SP>]
	LD1vx_t2: {0x0c40a000, "LD1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD1 <Vt>.<t>, [<Xn|SP>]
	LD1vx_t3: {0x0c406000, "LD1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD1 <Vt>.<t>, [<Xn|SP>]
	LD1vx_t4: {0x0c402000, "LD1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<imm>
	LD1vxi_tp1: {0x0cdf7000, "LD1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__8_0__16_1}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<imm_1>
	LD1vxi_tp2: {0x0cdfa000, "LD1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__16_0__32_1}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<imm_2>
	LD1vxi_tp3: {0x0cdf6000, "LD1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__24_0__48_1}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<imm_3>
	LD1vxi_tp4: {0x0cdf2000, "LD1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__32_0__64_1}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD1vxx_tp1: {0x0cc07000, "LD1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD1vxx_tp2: {0x0cc0a000, "LD1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD1vxx_tp3: {0x0cc06000, "LD1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD1vxx_tp4: {0x0cc02000, "LD1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.B[<index>], [<Xn|SP>]
	LD1vx_bi1: {0x0d400000, "LD1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD1 <Vt>.H[<index_2>], [<Xn|SP>]
	LD1vx_hi1: {0x0d404000, "LD1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD1 <Vt>.S[<index_3>], [<Xn|SP>]
	LD1vx_si1: {0x0d408000, "LD1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// LD1 <Vt>.D[<index_1>], [<Xn|SP>]
	LD1vx_di1: {0x0d408400, "LD1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// LD1 <Vt>.B[<index>], [<Xn|SP>], #1
	LD1vxi_bip1: {0x0ddf0000, "LD1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_1}},
	// LD1 <Vt>.H[<index_2>], [<Xn|SP>], #2
	LD1vxi_hip1: {0x0ddf4000, "LD1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_2}},
	// LD1 <Vt>.S[<index_3>], [<Xn|SP>], #4
	LD1vxi_sip1: {0x0ddf8000, "LD1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_4}},
	// LD1 <Vt>.D[<index_1>], [<Xn|SP>], #8
	LD1vxi_dip1: {0x0ddf8400, "LD1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_8}},
	// LD1 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	LD1vxx_bip1: {0x0dc00000, "LD1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	LD1vxx_hip1: {0x0dc04000, "LD1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	LD1vxx_sip1: {0x0dc08000, "LD1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// LD1 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	LD1vxx_dip1: {0x0dc08400, "LD1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// LD1R <Vt>.<t>, [<Xn|SP>]
	LD1Rvx_t1: {0x0d40c000, "LD1R", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD1R <Vt>.<t>, [<Xn|SP>], #<imm>
	LD1Rvxi_tp1: {0x0ddfc000, "LD1R", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_size__1_0__2_1__4_2__8_3}},
	// LD1R <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD1Rvxx_tp1: {0x0dc0c000, "LD1R", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD2 <Vt>.<t>, [<Xn|SP>]
	LD2vx_t2: {0x0c408000, "LD2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// LD2 <Vt>.<t>, [<Xn|SP>], #<imm>
	LD2vxi_tp2: {0x0cdf8000, "LD2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__16_0__32_1}},
	// LD2 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD2vxx_tp2: {0x0cc08000, "LD2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// LD2 <Vt>.B[<index>], [<Xn|SP>]
	LD2vx_bi2: {0x0d600000, "LD2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD2 <Vt>.H[<index_2>], [<Xn|SP>]
	LD2vx_hi2: {0x0d604000, "LD2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD2 <Vt>.S[<index_3>], [<Xn|SP>]
	LD2vx_si2: {0x0d608000, "LD2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// LD2 <Vt>.D[<index_1>], [<Xn|SP>]
	LD2vx_di2: {0x0d608400, "LD2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// LD2 <Vt>.B[<index>], [<Xn|SP>], #2
	LD2vx_bip2: {0x0dff0000, "LD2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_2}},
	// LD2 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	LD2vxx_bip2: {0x0de00000, "LD2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD2 <Vt>.H[<index_2>], [<Xn|SP>], #4
	LD2vx_hip2: {0x0dff4000, "LD2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_4}},
	// LD2 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	LD2vxx_hip2: {0x0de04000, "LD2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD2 <Vt>.S[<index_3>], [<Xn|SP>], #8
	LD2vx_sip2: {0x0dff8000, "LD2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_8}},
	// LD2 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	LD2vxx_sip2: {0x0de08000, "LD2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// LD2 <Vt>.D[<index_1>], [<Xn|SP>], #16
	LD2vx_dip2: {0x0dff8400, "LD2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_16}},
	// LD2 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	LD2vxx_dip2: {0x0de08400, "LD2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// LD2R <Vt>.<t>, [<Xn|SP>]
	LD2Rvx_t2: {0x0d60c000, "LD2R", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD2R <Vt>.<t>, [<Xn|SP>], #<imm>
	LD2Rvxi_tp2: {0x0dffc000, "LD2R", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_size__2_0__4_1__8_2__16_3}},
	// LD2R <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD2Rvxx_tp2: {0x0de0c000, "LD2R", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD3 <Vt>.<t>, [<Xn|SP>]
	LD3vx_t3: {0x0c404000, "LD3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// LD3 <Vt>.<t>, [<Xn|SP>], #<imm>
	LD3vxi_tp3: {0x0cdf4000, "LD3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__24_0__48_1}},
	// LD3 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD3vxx_tp3: {0x0cc04000, "LD3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// LD3 <Vt>.B[<index>], [<Xn|SP>]
	LD3vx_bi3: {0x0d402000, "LD3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD3 <Vt>.H[<index_2>], [<Xn|SP>]
	LD3vx_hi3: {0x0d406000, "LD3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD3 <Vt>.S[<index_3>], [<Xn|SP>]
	LD3vx_si3: {0x0d40a000, "LD3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// LD3 <Vt>.D[<index_1>], [<Xn|SP>]
	LD3vx_di3: {0x0d40a400, "LD3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// LD3 <Vt>.B[<index>], [<Xn|SP>], #3
	LD3vx_bip3: {0x0ddf2000, "LD3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_3}},
	// LD3 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	LD3vxx_bip3: {0x0dc02000, "LD3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD3 <Vt>.H[<index_2>], [<Xn|SP>], #6
	LD3vx_hip3: {0x0ddf6000, "LD3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_6}},
	// LD3 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	LD3vxx_hip3: {0x0dc06000, "LD3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD3 <Vt>.S[<index_3>], [<Xn|SP>], #12
	LD3vx_sip3: {0x0ddfa000, "LD3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_12}},
	// LD3 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	LD3vxx_sip3: {0x0dc0a000, "LD3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// LD3 <Vt>.D[<index_1>], [<Xn|SP>], #24
	LD3vx_dip3: {0x0ddfa400, "LD3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_24}},
	// LD3 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	LD3vxx_dip3: {0x0dc0a400, "LD3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// LD3R <Vt>.<t>, [<Xn|SP>]
	LD3Rvx_t3: {0x0d40e000, "LD3R", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD3R <Vt>.<t>, [<Xn|SP>], #<imm>
	LD3Rvxi_tp3: {0x0ddfe000, "LD3R", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_size__3_0__6_1__12_2__24_3}},
	// LD3R <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD3Rvxx_tp3: {0x0dc0e000, "LD3R", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LD4 <Vt>.<t>, [<Xn|SP>]
	LD4vx_t4: {0x0c400000, "LD4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// LD4 <Vt>.<t>, [<Xn|SP>], #<imm>
	LD4vxi_tp4: {0x0cdf0000, "LD4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__32_0__64_1}},
	// LD4 <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD4vxx_tp4: {0x0cc00000, "LD4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// LD4 <Vt>.B[<index>], [<Xn|SP>]
	LD4vx_bi4: {0x0d602000, "LD4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD4 <Vt>.H[<index_2>], [<Xn|SP>]
	LD4vx_hi4: {0x0d606000, "LD4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// LD4 <Vt>.S[<index_3>], [<Xn|SP>]
	LD4vx_si4: {0x0d60a000, "LD4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// LD4 <Vt>.D[<index_1>], [<Xn|SP>]
	LD4vx_di4: {0x0d60a400, "LD4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// LD4 <Vt>.B[<index>], [<Xn|SP>], #4
	LD4vx_bip4: {0x0dff2000, "LD4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_4}},
	// LD4 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	LD4vxx_bip4: {0x0de02000, "LD4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD4 <Vt>.H[<index_2>], [<Xn|SP>], #8
	LD4vx_hip4: {0x0dff6000, "LD4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_8}},
	// LD4 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	LD4vxx_hip4: {0x0de06000, "LD4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// LD4 <Vt>.S[<index_3>], [<Xn|SP>], #16
	LD4vx_sip4: {0x0dffa000, "LD4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_16}},
	// LD4 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	LD4vxx_sip4: {0x0de0a000, "LD4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// LD4 <Vt>.D[<index_1>], [<Xn|SP>], #32
	LD4vx_dip4: {0x0dffa400, "LD4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_32}},
	// LD4 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	LD4vxx_dip4: {0x0de0a400, "LD4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// LD4R <Vt>.<t>, [<Xn|SP>]
	LD4Rvx_t4: {0x0d60e000, "LD4R", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// LD4R <Vt>.<t>, [<Xn|SP>], #<imm>
	LD4Rvxi_tp4: {0x0dffe000, "LD4R", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_size__4_0__8_1__16_2__32_3}},
	// LD4R <Vt>.<t>, [<Xn|SP>], #<Xm>
	LD4Rvxx_tp4: {0x0de0e000, "LD4R", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// LDNP <St>, <St2>, [<Xn|SP>{, #<imm_2>}]
	LDNPssx: {0x2c400000, "LDNP", instArgs{arg_St_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// LDNP <Dt>, <Dt2>, [<Xn|SP>{, #<imm>}]
	LDNPddx: {0x6c400000, "LDNP", instArgs{arg_Dt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// LDNP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_1>}]
	LDNPqqx: {0xac400000, "LDNP", instArgs{arg_Qt_pair, arg_Xns_mem_optional_imm7_16_signed}},
	// LDP <St>, <St2>, [<Xn|SP>{, #<imm_4>}]
	LDPssx: {0x2d400000, "LDP", instArgs{arg_St_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// LDP <Dt>, <Dt2>, [<Xn|SP>{, #<imm>}]
	LDPddx: {0x6d400000, "LDP", instArgs{arg_Dt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// LDP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_2>}]
	LDPqqx: {0xad400000, "LDP", instArgs{arg_Qt_pair, arg_Xns_mem_optional_imm7_16_signed}},
	// LDP <St>, <St2>, [<Xn|SP>], #<imm_5>
	LDPssxi_p: {0x2cc00000, "LDP", instArgs{arg_St_pair, arg_Xns_mem_post_imm7_4_signed}},
	// LDP <Dt>, <Dt2>, [<Xn|SP>], #<imm_1>
	LDPddxi_p: {0x6cc00000, "LDP", instArgs{arg_Dt_pair, arg_Xns_mem_post_imm7_8_signed}},
	// LDP <Qt>, <Qt2>, [<Xn|SP>], #<imm_3>
	LDPqqxi_p: {0xacc00000, "LDP", instArgs{arg_Qt_pair, arg_Xns_mem_post_imm7_16_signed}},
	// LDP <St>, <St2>, [<Xn|SP>{, #<imm_5>}]!
	LDPssx_w: {0x2dc00000, "LDP", instArgs{arg_St_pair, arg_Xns_mem_wb_imm7_4_signed}},
	// LDP <Dt>, <Dt2>, [<Xn|SP>{, #<imm_1>}]!
	LDPddx_w: {0x6dc00000, "LDP", instArgs{arg_Dt_pair, arg_Xns_mem_wb_imm7_8_signed}},
	// LDP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_3>}]!
	LDPqqx_w: {0xadc00000, "LDP", instArgs{arg_Qt_pair, arg_Xns_mem_wb_imm7_16_signed}},
	// LDR <Bt>, [<Xn|SP>], #<simm>
	LDRbxi_p: {0x3c400400, "LDR", instArgs{arg_Bt, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Ht>, [<Xn|SP>], #<simm>
	LDRhxi_p: {0x7c400400, "LDR", instArgs{arg_Ht, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <St>, [<Xn|SP>], #<simm>
	LDRsxi_p: {0xbc400400, "LDR", instArgs{arg_St, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Dt>, [<Xn|SP>], #<simm>
	LDRdxi_p: {0xfc400400, "LDR", instArgs{arg_Dt, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Qt>, [<Xn|SP>], #<simm>
	LDRqxi_p: {0x3cc00400, "LDR", instArgs{arg_Qt, arg_Xns_mem_post_imm9_1_signed}},
	// LDR <Bt>, [<Xn|SP>{, #<simm>}]!
	LDRbx_w: {0x3c400c00, "LDR", instArgs{arg_Bt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Ht>, [<Xn|SP>{, #<simm>}]!
	LDRhx_w: {0x7c400c00, "LDR", instArgs{arg_Ht, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <St>, [<Xn|SP>{, #<simm>}]!
	LDRsx_w: {0xbc400c00, "LDR", instArgs{arg_St, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Dt>, [<Xn|SP>{, #<simm>}]!
	LDRdx_w: {0xfc400c00, "LDR", instArgs{arg_Dt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Qt>, [<Xn|SP>{, #<simm>}]!
	LDRqx_w: {0x3cc00c00, "LDR", instArgs{arg_Qt, arg_Xns_mem_wb_imm9_1_signed}},
	// LDR <Bt>, [<Xn|SP>{, #<pimm>}]
	LDRbx: {0x3d400000, "LDR", instArgs{arg_Bt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// LDR <Ht>, [<Xn|SP>{, #<pimm_2>}]
	LDRhx: {0x7d400000, "LDR", instArgs{arg_Ht, arg_Xns_mem_optional_imm12_2_unsigned}},
	// LDR <St>, [<Xn|SP>{, #<pimm_4>}]
	LDRsx: {0xbd400000, "LDR", instArgs{arg_St, arg_Xns_mem_optional_imm12_4_unsigned}},
	// LDR <Dt>, [<Xn|SP>{, #<pimm_1>}]
	LDRdx: {0xfd400000, "LDR", instArgs{arg_Dt, arg_Xns_mem_optional_imm12_8_unsigned}},
	// LDR <Qt>, [<Xn|SP>{, #<pimm_3>}]
	LDRqx: {0x3dc00000, "LDR", instArgs{arg_Qt, arg_Xns_mem_optional_imm12_16_unsigned}},
	// LDR <St>, <label>
	LDRsl: {0x1c000000, "LDR", instArgs{arg_St, arg_slabel_imm19_2}},
	// LDR <Dt>, <label>
	LDRdl: {0x5c000000, "LDR", instArgs{arg_Dt, arg_slabel_imm19_2}},
	// LDR <Qt>, <label>
	LDRql: {0x9c000000, "LDR", instArgs{arg_Qt, arg_slabel_imm19_2}},
	// LDR <Bt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRbxre: {0x3c600800, "LDR", instArgs{arg_Bt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// LDR <Ht>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRhxre: {0x7c600800, "LDR", instArgs{arg_Ht, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// LDR <St>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRsxre: {0xbc600800, "LDR", instArgs{arg_St, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1}},
	// LDR <Dt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRdxre: {0xfc600800, "LDR", instArgs{arg_Dt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1}},
	// LDR <Qt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	LDRqxre: {0x3ce00800, "LDR", instArgs{arg_Qt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__4_1}},
	// LDUR <Bt>, [<Xn|SP>{, #<simm>}]
	LDURbx: {0x3c400000, "LDUR", instArgs{arg_Bt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <Ht>, [<Xn|SP>{, #<simm>}]
	LDURhx: {0x7c400000, "LDUR", instArgs{arg_Ht, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <St>, [<Xn|SP>{, #<simm>}]
	LDURsx: {0xbc400000, "LDUR", instArgs{arg_St, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <Dt>, [<Xn|SP>{, #<simm>}]
	LDURdx: {0xfc400000, "LDUR", instArgs{arg_Dt, arg_Xns_mem_optional_imm9_1_signed}},
	// LDUR <Qt>, [<Xn|SP>{, #<simm>}]
	LDURqx: {0x3cc00000, "LDUR", instArgs{arg_Qt, arg_Xns_mem_optional_imm9_1_signed}},
	// MLA <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	MLAvvv_t: {0x0e209400, "MLA", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// MLA <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	MLAvvv_ti: {0x2f000000, "MLA", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// MLS <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	MLSvvv_t: {0x2e209400, "MLS", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// MLS <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	MLSvvv_ti: {0x2f004000, "MLS", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// MOV <Vd>.<ts>[<index1>], <Vn>.<ts>[<index2>]
	MOVvv_tii: {0x6e000400, "MOV", instArgs{arg_Vd_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1, arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1}},
	// MOV <V><d>, <Vn>.<t_1>[<index>]
	MOVvv_ti: {0x5e000400, "MOV", instArgs{arg_Vd_16_5__B_1__H_2__S_4__D_8, arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1}},
	// MOV <Vd>.<ts>[<index>], <R><n>
	MOVvr_ti: {0x4e001c00, "MOV", instArgs{arg_Vd_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1, arg_Rn_16_5__W_1__W_2__W_4__X_8}},
	// MOV <Wd>, <Vn>.S[<index>]
	MOVwv_si: {0x0e003c00, "MOV", instArgs{arg_Wd, arg_Vn_arrangement_S_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1}},
	// MOV <Xd>, <Vn>.D[<index_1>]
	MOVxv_di: {0x4e003c00, "MOV", instArgs{arg_Xd, arg_Vn_arrangement_D_index__imm5_1}},
	// MOV <Vd>.<t>, <Vn>.<t>
	MOVvv_t: {0x0ea01c00, "MOV", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1}},
	// MOVI <Vd>.<t_2>, #<imm8>{, LSL #0}
	MOVIvi_tb: {0x0f00e400, "MOVI", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_immediate_OptLSLZero__a_b_c_d_e_f_g_h}},
	// MOVI <Vd>.<t>, #<imm8>{, LSL #<amount>}
	MOVIvi_th: {0x0f008400, "MOVI", instArgs{arg_Vd_arrangement_Q___4H_0__8H_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1}},
	// MOVI <Vd>.<t_1>, #<imm8>{, LSL #<amount>}
	MOVIvi_ts: {0x0f000400, "MOVI", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1__16_2__24_3}},
	// MOVI <Vd>.<t_1>, #<imm8>, MSL #<amount>
	MOVIvii_ts: {0x0f00c400, "MOVI", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_MSL__a_b_c_d_e_f_g_h, arg_immediate_cmode__8_0__16_1}},
	// MOVI <Dd>, #<imm>
	MOVIdi: {0x2f00e400, "MOVI", instArgs{arg_Dd, arg_immediate_8x8_a_b_c_d_e_f_g_h}},
	// MOVI <Vd>.2D, #<imm>
	MOVIvi: {0x6f00e400, "MOVI", instArgs{arg_Vd_arrangement_2D, arg_immediate_8x8_a_b_c_d_e_f_g_h}},
	// MUL <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	MULvvv_i: {0x0f008000, "MUL", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// MUL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	MULvvv_t: {0x0e209c00, "MUL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// MVN <Vd>.<t>, <Vn>.<t>
	MVNvv_t: {0x2e205800, "MVN", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1}},
	// MVNI <Vd>.<t>, #<imm8>{, LSL #<amount>}
	MVNIvi_th: {0x2f008400, "MVNI", instArgs{arg_Vd_arrangement_Q___4H_0__8H_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1}},
	// MVNI <Vd>.<t_1>, #<imm8>{, LSL #<amount>}
	MVNIvi_ts: {0x2f000400, "MVNI", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1__16_2__24_3}},
	// MVNI <Vd>.<t_1>, #<imm8>, MSL #<amount>
	MVNIvii_ts: {0x2f00c400, "MVNI", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_MSL__a_b_c_d_e_f_g_h, arg_immediate_cmode__8_0__16_1}},
	// NEG <Vd>.<t>, <Vn>.<t>
	NEGvv_t: {0x2e20b800, "NEG", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// NEG <V><d>, <V><n>
	NEGvv: {0x7e20b800, "NEG", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3}},
	// NOT <Vd>.<t>, <Vn>.<t>
	NOTvv_t: {0x2e205800, "NOT", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1}},
	// ORN <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ORNvvv_t: {0x0ee01c00, "ORN", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// ORR <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ORRvvv_t: {0x0ea01c00, "ORR", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// ORR <Vd>.<t>, #<imm8>{, LSL #<amount>}
	ORRvi_th: {0x0f009400, "ORR", instArgs{arg_Vd_arrangement_Q___4H_0__8H_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1}},
	// ORR <Vd>.<t_1>, #<imm8>{, LSL #<amount>}
	ORRvi_ts: {0x0f001400, "ORR", instArgs{arg_Vd_arrangement_Q___2S_0__4S_1, arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1__16_2__24_3}},
	// PMUL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	PMULvvv_t: {0x2e209c00, "PMUL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01, arg_Vn_arrangement_size_Q___8B_00__16B_01, arg_Vm_arrangement_size_Q___8B_00__16B_01}},
	// PMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	PMULLvvv_t: {0x0e20e000, "PMULL", instArgs{arg_Vd_arrangement_size___8H_0__1Q_3, arg_Vn_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31}},
	// PMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	PMULL2vvv_t: {0x4e20e000, "PMULL2", instArgs{arg_Vd_arrangement_size___8H_0__1Q_3, arg_Vn_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31}},
	// RADDHN <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	RADDHNvvv_t: {0x2e204000, "RADDHN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// RADDHN2 <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	RADDHN2vvv_t: {0x6e204000, "RADDHN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// RAX1 <Vd>.2D, <Vn>.2D, <Vm>.2D
	RAX1vvv: {0xce608c00, "RAX1", instArgs{arg_Vd_arrangement_2D, arg_Vn_arrangement_2D, arg_Vm_arrangement_2D}},
	// RBIT <Vd>.<t>, <Vn>.<t>
	RBITvv_t: {0x2e605800, "RBIT", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_arrangement_Q___8B_0__16B_1}},
	// REV16 <Vd>.<t>, <Vn>.<t>
	REV16vv_t: {0x0e201800, "REV16", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01, arg_Vn_arrangement_size_Q___8B_00__16B_01}},
	// REV32 <Vd>.<t>, <Vn>.<t>
	REV32vv_t: {0x2e200800, "REV32", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11}},
	// REV64 <Vd>.<t>, <Vn>.<t>
	REV64vv_t: {0x0e200800, "REV64", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// RSHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	RSHRNvvi_t: {0x0f008c00, "RSHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// RSHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	RSHRN2vvi_t: {0x4f008c00, "RSHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// RSUBHN <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	RSUBHNvvv_t: {0x2e206000, "RSUBHN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// RSUBHN2 <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	RSUBHN2vvv_t: {0x6e206000, "RSUBHN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// SABA <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SABAvvv_t: {0x0e207c00, "SABA", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SABAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SABALvvv_t: {0x0e205000, "SABAL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SABAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SABAL2vvv_t: {0x4e205000, "SABAL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SABD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SABDvvv_t: {0x0e207400, "SABD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SABDL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SABDLvvv_t: {0x0e207000, "SABDL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SABDL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SABDL2vvv_t: {0x4e207000, "SABDL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADALP <Vd>.<ta>, <Vn>.<tb>
	SADALPvv_t: {0x0e206800, "SADALP", instArgs{arg_Vd_arrangement_size_Q___4H_00__8H_01__2S_10__4S_11__1D_20__2D_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADDL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SADDLvvv_t: {0x0e200000, "SADDL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADDL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SADDL2vvv_t: {0x4e200000, "SADDL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADDLP <Vd>.<ta>, <Vn>.<tb>
	SADDLPvv_t: {0x0e202800, "SADDLP", instArgs{arg_Vd_arrangement_size_Q___4H_00__8H_01__2S_10__4S_11__1D_20__2D_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADDLV <V><d>, <Vn>.<t>
	SADDLVvv_t: {0x0e303800, "SADDLV", instArgs{arg_Vd_22_2__H_0__S_1__D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// SADDW <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	SADDWvvv_t: {0x0e201000, "SADDW", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SADDW2 <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	SADDW2vvv_t: {0x4e201000, "SADDW2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SCVTF <Sd>, <Wn>, #<fbits>
	SCVTFswi: {0x1e020000, "SCVTF", instArgs{arg_Sd, arg_Wn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// SCVTF <Dd>, <Wn>, #<fbits>
	SCVTFdwi: {0x1e420000, "SCVTF", instArgs{arg_Dd, arg_Wn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// SCVTF <Sd>, <Xn>, #<fbits>
	SCVTFsxi: {0x9e020000, "SCVTF", instArgs{arg_Sd, arg_Xn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// SCVTF <Dd>, <Xn>, #<fbits>
	SCVTFdxi: {0x9e420000, "SCVTF", instArgs{arg_Dd, arg_Xn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// SCVTF <Sd>, <Wn>
	SCVTFsw: {0x1e220000, "SCVTF", instArgs{arg_Sd, arg_Wn}},
	// SCVTF <Dd>, <Wn>
	SCVTFdw: {0x1e620000, "SCVTF", instArgs{arg_Dd, arg_Wn}},
	// SCVTF <Sd>, <Xn>
	SCVTFsx: {0x9e220000, "SCVTF", instArgs{arg_Sd, arg_Xn}},
	// SCVTF <Dd>, <Xn>
	SCVTFdx: {0x9e620000, "SCVTF", instArgs{arg_Dd, arg_Xn}},
	// SCVTF <V><d>, <V><n>, #<fbits>
	SCVTFvvi: {0x5f00e400, "SCVTF", instArgs{arg_Vd_19_4__S_4__D_8, arg_Vn_19_4__S_4__D_8, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SCVTF <Vd>.<t>, <Vn>.<t>, #<fbits>
	SCVTFvvi_t: {0x0f00e400, "SCVTF", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SCVTF <V><d>, <V><n>
	SCVTFvv: {0x5e21d800, "SCVTF", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// SCVTF <Vd>.<t>, <Vn>.<t>
	SCVTFvv_t: {0x0e21d800, "SCVTF", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// SHA1C <Qd>, <Sn>, <Vm>.4S
	SHA1Cqsv: {0x5e000000, "SHA1C", instArgs{arg_Qd, arg_Sn, arg_Vm_arrangement_4S}},
	// SHA1H <Sd>, <Sn>
	SHA1Hss: {0x5e280800, "SHA1H", instArgs{arg_Sd, arg_Sn}},
	// SHA1M <Qd>, <Sn>, <Vm>.4S
	SHA1Mqsv: {0x5e002000, "SHA1M", instArgs{arg_Qd, arg_Sn, arg_Vm_arrangement_4S}},
	// SHA1P <Qd>, <Sn>, <Vm>.4S
	SHA1Pqsv: {0x5e001000, "SHA1P", instArgs{arg_Qd, arg_Sn, arg_Vm_arrangement_4S}},
	// SHA1SU0 <Vd>.4S, <Vn>.4S, <Vm>.4S
	SHA1SU0vvv: {0x5e003000, "SHA1SU0", instArgs{arg_Vd_arrangement_4S, arg_Vn_arrangement_4S, arg_Vm_arrangement_4S}},
	// SHA1SU1 <Vd>.4S, <Vn>.4S
	SHA1SU1vv: {0x5e281800, "SHA1SU1", instArgs{arg_Vd_arrangement_4S, arg_Vn_arrangement_4S}},
	// SHA256H <Qd>, <Qn>, <Vm>.4S
	SHA256Hqqv: {0x5e004000, "SHA256H", instArgs{arg_Qd, arg_Qn, arg_Vm_arrangement_4S}},
	// SHA256H2 <Qd>, <Qn>, <Vm>.4S
	SHA256H2qqv: {0x5e005000, "SHA256H2", instArgs{arg_Qd, arg_Qn, arg_Vm_arrangement_4S}},
	// SHA256SU0 <Vd>.4S, <Vn>.4S
	SHA256SU0vv: {0x5e282800, "SHA256SU0", instArgs{arg_Vd_arrangement_4S, arg_Vn_arrangement_4S}},
	// SHA256SU1 <Vd>.4S, <Vn>.4S, <Vm>.4S
	SHA256SU1vvv: {0x5e006000, "SHA256SU1", instArgs{arg_Vd_arrangement_4S, arg_Vn_arrangement_4S, arg_Vm_arrangement_4S}},
	// SHA512H <Qd>, <Qn>, <Vm>.2D
	SHA512Hqqv: {0xce608000, "SHA512H", instArgs{arg_Qd, arg_Qn, arg_Vm_arrangement_2D}},
	// SHA512H2 <Qd>, <Qn>, <Vm>.2D
	SHA512H2qqv: {0xce608400, "SHA512H2", instArgs{arg_Qd, arg_Qn, arg_Vm_arrangement_2D}},
	// SHA512SU0 <Vd>.2D, <Vn>.2D
	SHA512SU0vv: {0xcec08000, "SHA512SU0", instArgs{arg_Vd_arrangement_2D, arg_Vn_arrangement_2D}},
	// SHA512SU1 <Vd>.2D, <Vn>.2D, <Vm>.2D
	SHA512SU1vvv: {0xce608800, "SHA512SU1", instArgs{arg_Vd_arrangement_2D, arg_Vn_arrangement_2D, arg_Vm_arrangement_2D}},
	// SHADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SHADDvvv_t: {0x0e200400, "SHADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SHL <Vd>.<t>, <Vn>.<t>, #<shift>
	SHLvvi_t: {0x0f005400, "SHL", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SHL <V><d>, <V><n>, #<shift>
	SHLvvi: {0x5f005400, "SHL", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_0_63_immh_immb__UIntimmhimmb64_8}},
	// SHLL <Vd>.<ta>, <Vn>.<tb>, #<shift>
	SHLLvvi_t: {0x2e213800, "SHLL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_immediate_0_width_size__8_0__16_1__32_2}},
	// SHLL2 <Vd>.<ta>, <Vn>.<tb>, #<shift>
	SHLL2vvi_t: {0x6e213800, "SHLL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_immediate_0_width_size__8_0__16_1__32_2}},
	// SHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SHRNvvi_t: {0x0f008400, "SHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SHRN2vvi_t: {0x4f008400, "SHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SHSUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SHSUBvvv_t: {0x0e202400, "SHSUB", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SLI <Vd>.<t>, <Vn>.<t>, #<shift>
	SLIvvi_t: {0x2f005400, "SLI", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SLI <V><d>, <V><n>, #<shift>
	SLIvvi: {0x7f005400, "SLI", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_0_63_immh_immb__UIntimmhimmb64_8}},
	// SMAX <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SMAXvvv_t: {0x0e206400, "SMAX", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMAXP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SMAXPvvv_t: {0x0e20a400, "SMAXP", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMAXV <V><d>, <Vn>.<t>
	SMAXVvv_t: {0x0e30a800, "SMAXV", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// SMIN <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SMINvvv_t: {0x0e206c00, "SMIN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMINP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SMINPvvv_t: {0x0e20ac00, "SMINP", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMINV <V><d>, <Vn>.<t>
	SMINVvv_t: {0x0e31a800, "SMINV", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// SMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMLALvvv_t: {0x0e208000, "SMLAL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMLALvvv_ti: {0x0f002000, "SMLAL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMLAL2vvv_ti: {0x4f002000, "SMLAL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMLAL2vvv_t: {0x4e208000, "SMLAL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMLSLvvv_ti: {0x0f006000, "SMLSL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMLSLvvv_t: {0x0e20a000, "SMLSL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMLSL2vvv_ti: {0x4f006000, "SMLSL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMLSL2vvv_t: {0x4e20a000, "SMLSL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMOV <Wd>, <Vn>.<ts>[<index>]
	SMOVwv_ti: {0x0e002c00, "SMOV", instArgs{arg_Wd, arg_Vn_arrangement_imm5___B_1__H_2_index__imm5__imm5lt41gt_1__imm5lt42gt_2_1}},
	// SMOV <Xd>, <Vn>.<ts_1>[<index_1>]
	SMOVxv_ti: {0x4e002c00, "SMOV", instArgs{arg_Xd, arg_Vn_arrangement_imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1}},
	// SMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMULLvvv_t: {0x0e20c000, "SMULL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMULLvvv_ti: {0x0f00a000, "SMULL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SMULL2vvv_ti: {0x4f00a000, "SMULL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SMULL2vvv_t: {0x4e20c000, "SMULL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SQABS <V><d>, <V><n>
	SQABSvv: {0x5e207800, "SQABS", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3}},
	// SQABS <Vd>.<t>, <Vn>.<t>
	SQABSvv_t: {0x0e207800, "SQABS", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQADD <V><d>, <V><n>, <V><m>
	SQADDvvv: {0x5e200c00, "SQADD", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// SQADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQADDvvv_t: {0x0e200c00, "SQADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQDMLAL <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	SQDMLALvvv_tis: {0x5f003000, "SQDMLAL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMLALvvv_tiv: {0x0f003000, "SQDMLAL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMLAL <V><d>, <V><n>, <V><m>
	SQDMLALvvv: {0x5e209000, "SQDMLAL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_22_2__H_1__S_2}},
	// SQDMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMLALvvv_t: {0x0e209000, "SQDMLAL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMLAL2vvv_t: {0x4e209000, "SQDMLAL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMLAL2vvv_ti: {0x4f003000, "SQDMLAL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMLSL <V><d>, <V><n>, <V><m>
	SQDMLSLvvv: {0x5e20b000, "SQDMLSL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_22_2__H_1__S_2}},
	// SQDMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMLSLvvv_t: {0x0e20b000, "SQDMLSL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMLSL <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	SQDMLSLvvv_tis: {0x5f007000, "SQDMLSL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMLSLvvv_tiv: {0x0f007000, "SQDMLSL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMLSL2vvv_t: {0x4e20b000, "SQDMLSL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMLSL2vvv_ti: {0x4f007000, "SQDMLSL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMULH <V><d>, <V><n>, <V><m>
	SQDMULHvvv: {0x5e20b400, "SQDMULH", instArgs{arg_Vd_22_2__H_1__S_2, arg_Vn_22_2__H_1__S_2, arg_Vm_22_2__H_1__S_2}},
	// SQDMULH <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQDMULHvvv_t: {0x0e20b400, "SQDMULH", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMULH <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	SQDMULHvvv_tis: {0x5f00c000, "SQDMULH", instArgs{arg_Vd_22_2__H_1__S_2, arg_Vn_22_2__H_1__S_2, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMULH <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	SQDMULHvvv_tiv: {0x0f00c000, "SQDMULH", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMULL <V><d>, <V><n>, <V><m>
	SQDMULLvvv: {0x5e20d000, "SQDMULL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_22_2__H_1__S_2}},
	// SQDMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMULLvvv_t: {0x0e20d000, "SQDMULL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMULL <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	SQDMULLvvv_tis: {0x5f00b000, "SQDMULL", instArgs{arg_Vd_22_2__S_1__D_2, arg_Vn_22_2__H_1__S_2, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMULLvvv_tiv: {0x0f00b000, "SQDMULL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQDMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SQDMULL2vvv_t: {0x4e20d000, "SQDMULL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQDMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	SQDMULL2vvv_ti: {0x4f00b000, "SQDMULL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQNEG <Vd>.<t>, <Vn>.<t>
	SQNEGvv_t: {0x2e207800, "SQNEG", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQNEG <V><d>, <V><n>
	SQNEGvv: {0x7e207800, "SQNEG", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3}},
	// SQRDMULH <V><d>, <V><n>, <V><m>
	SQRDMULHvvv: {0x7e20b400, "SQRDMULH", instArgs{arg_Vd_22_2__H_1__S_2, arg_Vn_22_2__H_1__S_2, arg_Vm_22_2__H_1__S_2}},
	// SQRDMULH <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQRDMULHvvv_t: {0x2e20b400, "SQRDMULH", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21}},
	// SQRDMULH <V><d>, <V><n>, <Vm>.<ts_1>[<index_1>]
	SQRDMULHvvv_tis: {0x5f00d000, "SQRDMULH", instArgs{arg_Vd_22_2__H_1__S_2, arg_Vn_22_2__H_1__S_2, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQRDMULH <Vd>.<t>, <Vn>.<t>, <Vm>.<ts>[<index>]
	SQRDMULHvvv_tiv: {0x0f00d000, "SQRDMULH", instArgs{arg_Vd_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// SQRSHL <V><d>, <V><n>, <V><m>
	SQRSHLvvv: {0x5e205c00, "SQRSHL", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// SQRSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQRSHLvvv_t: {0x0e205c00, "SQRSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQRSHRN <V><d>, <V><n>, #<shift>
	SQRSHRNvvi: {0x5f009c00, "SQRSHRN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQRSHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQRSHRNvvi_t: {0x0f009c00, "SQRSHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQRSHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQRSHRN2vvi_t: {0x4f009c00, "SQRSHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQRSHRUN <V><d>, <V><n>, #<shift>
	SQRSHRUNvvi: {0x7f008c00, "SQRSHRUN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQRSHRUN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQRSHRUNvvi_t: {0x2f008c00, "SQRSHRUN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQRSHRUN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQRSHRUN2vvi_t: {0x6f008c00, "SQRSHRUN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHL <V><d>, <V><n>, <V><m>
	SQSHLvvv: {0x5e204c00, "SQSHL", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// SQSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQSHLvvv_t: {0x0e204c00, "SQSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQSHL <V><d>, <V><n>, #<shift>
	SQSHLvvi: {0x5f007400, "SQSHL", instArgs{arg_Vd_19_4__B_1__H_2__S_4__D_8, arg_Vn_19_4__B_1__H_2__S_4__D_8, arg_immediate_0_width_m1_immh_immb__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SQSHL <Vd>.<t>, <Vn>.<t>, #<shift>
	SQSHLvvi_t: {0x0f007400, "SQSHL", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SQSHLU <Vd>.<t>, <Vn>.<t>, #<shift>
	SQSHLUvvi_t: {0x2f006400, "SQSHLU", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SQSHLU <V><d>, <V><n>, #<shift>
	SQSHLUvvi: {0x7f006400, "SQSHLU", instArgs{arg_Vd_19_4__B_1__H_2__S_4__D_8, arg_Vn_19_4__B_1__H_2__S_4__D_8, arg_immediate_0_width_m1_immh_immb__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// SQSHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQSHRNvvi_t: {0x0f009400, "SQSHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHRN <V><d>, <V><n>, #<shift>
	SQSHRNvvi: {0x5f009400, "SQSHRN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQSHRN2vvi_t: {0x4f009400, "SQSHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHRUN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQSHRUNvvi_t: {0x2f008400, "SQSHRUN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHRUN <V><d>, <V><n>, #<shift>
	SQSHRUNvvi: {0x7f008400, "SQSHRUN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSHRUN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	SQSHRUN2vvi_t: {0x6f008400, "SQSHRUN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// SQSUB <V><d>, <V><n>, <V><m>
	SQSUBvvv: {0x5e202c00, "SQSUB", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// SQSUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SQSUBvvv_t: {0x0e202c00, "SQSUB", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SQXTN <V><d>, <V><n>
	SQXTNvv: {0x5e214800, "SQXTN", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_22_2__H_0__S_1__D_2}},
	// SQXTN <Vd>.<tb>, <Vn>.<ta>
	SQXTNvv_t: {0x0e214800, "SQXTN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// SQXTN2 <Vd>.<tb>, <Vn>.<ta>
	SQXTN2vv_t: {0x4e214800, "SQXTN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// SQXTUN <V><d>, <V><n>
	SQXTUNvv: {0x7e212800, "SQXTUN", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_22_2__H_0__S_1__D_2}},
	// SQXTUN <Vd>.<tb>, <Vn>.<ta>
	SQXTUNvv_t: {0x2e212800, "SQXTUN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// SQXTUN2 <Vd>.<tb>, <Vn>.<ta>
	SQXTUN2vv_t: {0x6e212800, "SQXTUN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// SRHADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SRHADDvvv_t: {0x0e201400, "SRHADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SRI <V><d>, <V><n>, #<shift>
	SRIvvi: {0x7f004400, "SRI", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// SRI <Vd>.<t>, <Vn>.<t>, #<shift>
	SRIvvi_t: {0x2f004400, "SRI", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SRSHL <V><d>, <V><n>, <V><m>
	SRSHLvvv: {0x5e205400, "SRSHL", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// SRSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SRSHLvvv_t: {0x0e205400, "SRSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SRSHR <V><d>, <V><n>, #<shift>
	SRSHRvvi: {0x5f002400, "SRSHR", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// SRSHR <Vd>.<t>, <Vn>.<t>, #<shift>
	SRSHRvvi_t: {0x0f002400, "SRSHR", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SRSRA <V><d>, <V><n>, #<shift>
	SRSRAvvi: {0x5f003400, "SRSRA", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// SRSRA <Vd>.<t>, <Vn>.<t>, #<shift>
	SRSRAvvi_t: {0x0f003400, "SRSRA", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SSHL <V><d>, <V><n>, <V><m>
	SSHLvvv: {0x5e204400, "SSHL", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// SSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SSHLvvv_t: {0x0e204400, "SSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SSHLL <Vd>.<ta>, <Vn>.<tb>, #<shift>
	SSHLLvvi_t: {0x0f00a400, "SSHLL", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4}},
	// SSHLL2 <Vd>.<ta>, <Vn>.<tb>, #<shift>
	SSHLL2vvi_t: {0x4f00a400, "SSHLL2", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4}},
	// SSHR <V><d>, <V><n>, #<shift>
	SSHRvvi: {0x5f000400, "SSHR", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// SSHR <Vd>.<t>, <Vn>.<t>, #<shift>
	SSHRvvi_t: {0x0f000400, "SSHR", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SSRA <Vd>.<t>, <Vn>.<t>, #<shift>
	SSRAvvi_t: {0x0f001400, "SSRA", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// SSRA <V><d>, <V><n>, #<shift>
	SSRAvvi: {0x5f001400, "SSRA", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// SSUBL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SSUBLvvv_t: {0x0e202000, "SSUBL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SSUBL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	SSUBL2vvv_t: {0x4e202000, "SSUBL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SSUBW <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	SSUBWvvv_t: {0x0e203000, "SSUBW", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// SSUBW2 <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	SSUBW2vvv_t: {0x4e203000, "SSUBW2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// ST1 <Vt>.<t>, [<Xn|SP>]
	ST1vx_t1: {0x0c007000, "ST1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// ST1 <Vt>.<t>, [<Xn|SP>]
	ST1vx_t2: {0x0c00a000, "ST1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// ST1 <Vt>.<t>, [<Xn|SP>]
	ST1vx_t3: {0x0c006000, "ST1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// ST1 <Vt>.<t>, [<Xn|SP>]
	ST1vx_t4: {0x0c002000, "ST1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_offset}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<imm>
	ST1vxi_tp1: {0x0c9f7000, "ST1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__8_0__16_1}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<imm_1>
	ST1vxi_tp2: {0x0c9fa000, "ST1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__16_0__32_1}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<imm_2>
	ST1vxi_tp3: {0x0c9f6000, "ST1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__24_0__48_1}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<imm_3>
	ST1vxi_tp4: {0x0c9f2000, "ST1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Q__32_0__64_1}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST1vxx_tp1: {0x0c807000, "ST1", instArgs{arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST1vxx_tp2: {0x0c80a000, "ST1", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST1vxx_tp3: {0x0c806000, "ST1", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST1vxx_tp4: {0x0c802000, "ST1", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.B[<index>], [<Xn|SP>]
	ST1vx_bi1: {0x0d000000, "ST1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST1 <Vt>.H[<index_2>], [<Xn|SP>]
	ST1vx_hi1: {0x0d004000, "ST1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST1 <Vt>.S[<index_3>], [<Xn|SP>]
	ST1vx_si1: {0x0d008000, "ST1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// ST1 <Vt>.D[<index_1>], [<Xn|SP>]
	ST1vx_di1: {0x0d008400, "ST1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// ST1 <Vt>.B[<index>], [<Xn|SP>], #1
	ST1vxi_bip1: {0x0d9f0000, "ST1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_1}},
	// ST1 <Vt>.H[<index_2>], [<Xn|SP>], #2
	ST1vxi_hip1: {0x0d9f4000, "ST1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_2}},
	// ST1 <Vt>.S[<index_3>], [<Xn|SP>], #4
	ST1vxi_sip1: {0x0d9f8000, "ST1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_4}},
	// ST1 <Vt>.D[<index_1>], [<Xn|SP>], #8
	ST1vxi_dip1: {0x0d9f8400, "ST1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_8}},
	// ST1 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	ST1vxx_bip1: {0x0d800000, "ST1", instArgs{arg_Vt_1_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	ST1vxx_hip1: {0x0d804000, "ST1", instArgs{arg_Vt_1_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	ST1vxx_sip1: {0x0d808000, "ST1", instArgs{arg_Vt_1_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// ST1 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	ST1vxx_dip1: {0x0d808400, "ST1", instArgs{arg_Vt_1_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// ST2 <Vt>.<t>, [<Xn|SP>]
	ST2vx_t2: {0x0c008000, "ST2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// ST2 <Vt>.<t>, [<Xn|SP>], #<imm>
	ST2vxi_tp2: {0x0c9f8000, "ST2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__16_0__32_1}},
	// ST2 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST2vxx_tp2: {0x0c808000, "ST2", instArgs{arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// ST2 <Vt>.B[<index>], [<Xn|SP>]
	ST2vx_bi2: {0x0d200000, "ST2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST2 <Vt>.H[<index_2>], [<Xn|SP>]
	ST2vx_hi2: {0x0d204000, "ST2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST2 <Vt>.S[<index_3>], [<Xn|SP>]
	ST2vx_si2: {0x0d208000, "ST2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// ST2 <Vt>.D[<index_1>], [<Xn|SP>]
	ST2vx_di2: {0x0d208400, "ST2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// ST2 <Vt>.B[<index>], [<Xn|SP>], #2
	ST2vx_bip2: {0x0dbf0000, "ST2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_2}},
	// ST2 <Vt>.H[<index_2>], [<Xn|SP>], #4
	ST2vx_hip2: {0x0dbf4000, "ST2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_4}},
	// ST2 <Vt>.S[<index_3>], [<Xn|SP>], #8
	ST2vx_sip2: {0x0dbf8000, "ST2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_8}},
	// ST2 <Vt>.D[<index_1>], [<Xn|SP>], #16
	ST2vx_dip2: {0x0dbf8400, "ST2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_16}},
	// ST2 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	ST2vxx_bip2: {0x0da00000, "ST2", instArgs{arg_Vt_2_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST2 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	ST2vxx_hip2: {0x0da04000, "ST2", instArgs{arg_Vt_2_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST2 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	ST2vxx_sip2: {0x0da08000, "ST2", instArgs{arg_Vt_2_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// ST2 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	ST2vxx_dip2: {0x0da08400, "ST2", instArgs{arg_Vt_2_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// ST3 <Vt>.<t>, [<Xn|SP>]
	ST3vx_t3: {0x0c004000, "ST3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// ST3 <Vt>.<t>, [<Xn|SP>], #<imm>
	ST3vxi_tp3: {0x0c9f4000, "ST3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__24_0__48_1}},
	// ST3 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST3vxx_tp3: {0x0c804000, "ST3", instArgs{arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// ST3 <Vt>.B[<index>], [<Xn|SP>]
	ST3vx_bi3: {0x0d002000, "ST3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST3 <Vt>.H[<index_2>], [<Xn|SP>]
	ST3vx_hi3: {0x0d006000, "ST3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST3 <Vt>.S[<index_3>], [<Xn|SP>]
	ST3vx_si3: {0x0d00a000, "ST3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// ST3 <Vt>.D[<index_1>], [<Xn|SP>]
	ST3vx_di3: {0x0d00a400, "ST3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// ST3 <Vt>.B[<index>], [<Xn|SP>], #3
	ST3vx_bip3: {0x0d9f2000, "ST3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_3}},
	// ST3 <Vt>.H[<index_2>], [<Xn|SP>], #6
	ST3vx_hip3: {0x0d9f6000, "ST3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_6}},
	// ST3 <Vt>.S[<index_3>], [<Xn|SP>], #12
	ST3vx_sip3: {0x0d9fa000, "ST3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_12}},
	// ST3 <Vt>.D[<index_1>], [<Xn|SP>], #24
	ST3vx_dip3: {0x0d9fa400, "ST3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_24}},
	// ST3 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	ST3vxx_bip3: {0x0d802000, "ST3", instArgs{arg_Vt_3_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST3 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	ST3vxx_hip3: {0x0d806000, "ST3", instArgs{arg_Vt_3_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST3 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	ST3vxx_sip3: {0x0d80a000, "ST3", instArgs{arg_Vt_3_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// ST3 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	ST3vxx_dip3: {0x0d80a400, "ST3", instArgs{arg_Vt_3_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// ST4 <Vt>.<t>, [<Xn|SP>]
	ST4vx_t4: {0x0c000000, "ST4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_offset}},
	// ST4 <Vt>.<t>, [<Xn|SP>], #<imm>
	ST4vxi_tp4: {0x0c9f0000, "ST4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Q__32_0__64_1}},
	// ST4 <Vt>.<t>, [<Xn|SP>], #<Xm>
	ST4vxx_tp4: {0x0c800000, "ST4", instArgs{arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Xns_mem_post_Xm}},
	// ST4 <Vt>.B[<index>], [<Xn|SP>]
	ST4vx_bi4: {0x0d202000, "ST4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST4 <Vt>.H[<index_2>], [<Xn|SP>]
	ST4vx_hi4: {0x0d206000, "ST4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_offset}},
	// ST4 <Vt>.S[<index_3>], [<Xn|SP>]
	ST4vx_si4: {0x0d20a000, "ST4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_offset}},
	// ST4 <Vt>.D[<index_1>], [<Xn|SP>]
	ST4vx_di4: {0x0d20a400, "ST4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_offset}},
	// ST4 <Vt>.B[<index>], [<Xn|SP>], #4
	ST4vx_bip4: {0x0dbf2000, "ST4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_4}},
	// ST4 <Vt>.H[<index_2>], [<Xn|SP>], #8
	ST4vx_hip4: {0x0dbf6000, "ST4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_fixedimm_8}},
	// ST4 <Vt>.S[<index_3>], [<Xn|SP>], #16
	ST4vx_sip4: {0x0dbfa000, "ST4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_post_fixedimm_16}},
	// ST4 <Vt>.D[<index_1>], [<Xn|SP>], #32
	ST4vx_dip4: {0x0dbfa400, "ST4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_post_fixedimm_32}},
	// ST4 <Vt>.B[<index>], [<Xn|SP>], #<Xm>
	ST4vxx_bip4: {0x0da02000, "ST4", instArgs{arg_Vt_4_arrangement_B_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST4 <Vt>.H[<index_2>], [<Xn|SP>], #<Xm>
	ST4vxx_hip4: {0x0da06000, "ST4", instArgs{arg_Vt_4_arrangement_H_index__Q_S_size_1, arg_Xns_mem_post_Xm}},
	// ST4 <Vt>.S[<index_3>], [<Xn|SP>], #<Xm>
	ST4vxx_sip4: {0x0da0a000, "ST4", instArgs{arg_Vt_4_arrangement_S_index__Q_S_1, arg_Xns_mem_post_Xm}},
	// ST4 <Vt>.D[<index_1>], [<Xn|SP>], #<Xm>
	ST4vxx_dip4: {0x0da0a400, "ST4", instArgs{arg_Vt_4_arrangement_D_index__Q_1, arg_Xns_mem_post_Xm}},
	// STNP <St>, <St2>, [<Xn|SP>{, #<imm_2>}]
	STNPssx: {0x2c000000, "STNP", instArgs{arg_St_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// STNP <Dt>, <Dt2>, [<Xn|SP>{, #<imm>}]
	STNPddx: {0x6c000000, "STNP", instArgs{arg_Dt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// STNP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_1>}]
	STNPqqx: {0xac000000, "STNP", instArgs{arg_Qt_pair, arg_Xns_mem_optional_imm7_16_signed}},
	// STP <St>, <St2>, [<Xn|SP>{, #<imm_4>}]
	STPssx: {0x2d000000, "STP", instArgs{arg_St_pair, arg_Xns_mem_optional_imm7_4_signed}},
	// STP <Dt>, <Dt2>, [<Xn|SP>{, #<imm>}]
	STPddx: {0x6d000000, "STP", instArgs{arg_Dt_pair, arg_Xns_mem_optional_imm7_8_signed}},
	// STP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_2>}]
	STPqqx: {0xad000000, "STP", instArgs{arg_Qt_pair, arg_Xns_mem_optional_imm7_16_signed}},
	// STP <St>, <St2>, [<Xn|SP>], #<imm_5>
	STPssxi_p: {0x2c800000, "STP", instArgs{arg_St_pair, arg_Xns_mem_post_imm7_4_signed}},
	// STP <Dt>, <Dt2>, [<Xn|SP>], #<imm_1>
	STPddxi_p: {0x6c800000, "STP", instArgs{arg_Dt_pair, arg_Xns_mem_post_imm7_8_signed}},
	// STP <Qt>, <Qt2>, [<Xn|SP>], #<imm_3>
	STPqqxi_p: {0xac800000, "STP", instArgs{arg_Qt_pair, arg_Xns_mem_post_imm7_16_signed}},
	// STP <St>, <St2>, [<Xn|SP>{, #<imm_5>}]!
	STPssx_w: {0x2d800000, "STP", instArgs{arg_St_pair, arg_Xns_mem_wb_imm7_4_signed}},
	// STP <Dt>, <Dt2>, [<Xn|SP>{, #<imm_1>}]!
	STPddx_w: {0x6d800000, "STP", instArgs{arg_Dt_pair, arg_Xns_mem_wb_imm7_8_signed}},
	// STP <Qt>, <Qt2>, [<Xn|SP>{, #<imm_3>}]!
	STPqqx_w: {0xad800000, "STP", instArgs{arg_Qt_pair, arg_Xns_mem_wb_imm7_16_signed}},
	// STR <Bt>, [<Xn|SP>], #<simm>
	STRbxi_p: {0x3c000400, "STR", instArgs{arg_Bt, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Ht>, [<Xn|SP>], #<simm>
	STRhxi_p: {0x7c000400, "STR", instArgs{arg_Ht, arg_Xns_mem_post_imm9_1_signed}},
	// STR <St>, [<Xn|SP>], #<simm>
	STRsxi_p: {0xbc000400, "STR", instArgs{arg_St, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Dt>, [<Xn|SP>], #<simm>
	STRdxi_p: {0xfc000400, "STR", instArgs{arg_Dt, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Qt>, [<Xn|SP>], #<simm>
	STRqxi_p: {0x3c800400, "STR", instArgs{arg_Qt, arg_Xns_mem_post_imm9_1_signed}},
	// STR <Bt>, [<Xn|SP>{, #<simm>}]!
	STRbx_w: {0x3c000c00, "STR", instArgs{arg_Bt, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Ht>, [<Xn|SP>{, #<simm>}]!
	STRhx_w: {0x7c000c00, "STR", instArgs{arg_Ht, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <St>, [<Xn|SP>{, #<simm>}]!
	STRsx_w: {0xbc000c00, "STR", instArgs{arg_St, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Dt>, [<Xn|SP>{, #<simm>}]!
	STRdx_w: {0xfc000c00, "STR", instArgs{arg_Dt, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Qt>, [<Xn|SP>{, #<simm>}]!
	STRqx_w: {0x3c800c00, "STR", instArgs{arg_Qt, arg_Xns_mem_wb_imm9_1_signed}},
	// STR <Bt>, [<Xn|SP>{, #<pimm>}]
	STRbx: {0x3d000000, "STR", instArgs{arg_Bt, arg_Xns_mem_optional_imm12_1_unsigned}},
	// STR <Ht>, [<Xn|SP>{, #<pimm_2>}]
	STRhx: {0x7d000000, "STR", instArgs{arg_Ht, arg_Xns_mem_optional_imm12_2_unsigned}},
	// STR <St>, [<Xn|SP>{, #<pimm_4>}]
	STRsx: {0xbd000000, "STR", instArgs{arg_St, arg_Xns_mem_optional_imm12_4_unsigned}},
	// STR <Dt>, [<Xn|SP>{, #<pimm_1>}]
	STRdx: {0xfd000000, "STR", instArgs{arg_Dt, arg_Xns_mem_optional_imm12_8_unsigned}},
	// STR <Qt>, [<Xn|SP>{, #<pimm_3>}]
	STRqx: {0x3d800000, "STR", instArgs{arg_Qt, arg_Xns_mem_optional_imm12_16_unsigned}},
	// STR <Bt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRbxre: {0x3c200800, "STR", instArgs{arg_Bt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1}},
	// STR <Ht>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRhxre: {0x7c200800, "STR", instArgs{arg_Ht, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1}},
	// STR <St>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRsxre: {0xbc200800, "STR", instArgs{arg_St, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1}},
	// STR <Dt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRdxre: {0xfc200800, "STR", instArgs{arg_Dt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1}},
	// STR <Qt>, [<Xn|SP>, (<Wm>|<Xm>) {, <extend> {<amount>}}]
	STRqxre: {0x3ca00800, "STR", instArgs{arg_Qt, arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__4_1}},
	// STUR <Bt>, [<Xn|SP>{, #<simm>}]
	STURbx: {0x3c000000, "STUR", instArgs{arg_Bt, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <Ht>, [<Xn|SP>{, #<simm>}]
	STURhx: {0x7c000000, "STUR", instArgs{arg_Ht, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <St>, [<Xn|SP>{, #<simm>}]
	STURsx: {0xbc000000, "STUR", instArgs{arg_St, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <Dt>, [<Xn|SP>{, #<simm>}]
	STURdx: {0xfc000000, "STUR", instArgs{arg_Dt, arg_Xns_mem_optional_imm9_1_signed}},
	// STUR <Qt>, [<Xn|SP>{, #<simm>}]
	STURqx: {0x3c800000, "STUR", instArgs{arg_Qt, arg_Xns_mem_optional_imm9_1_signed}},
	// SUB <V><d>, <V><n>, <V><m>
	SUBvvv: {0x7e208400, "SUB", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// SUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	SUBvvv_t: {0x2e208400, "SUB", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SUBHN <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	SUBHNvvv_t: {0x0e206000, "SUBHN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// SUBHN2 <Vd>.<tb>, <Vn>.<ta>, <Vm>.<ta>
	SUBHN2vvv_t: {0x4e206000, "SUBHN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size___8H_0__4S_1__2D_2}},
	// SUQADD <V><d>, <V><n>
	SUQADDvv: {0x5e203800, "SUQADD", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3}},
	// SUQADD <Vd>.<t>, <Vn>.<t>
	SUQADDvv_t: {0x0e203800, "SUQADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// SXTL <Vd>.<ta>, <Vn>.<tb>
	SXTLvv_t: {0x0f00a400, "SXTL", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41}},
	// SXTL2 <Vd>.<ta>, <Vn>.<tb>
	SXTL2vv_t: {0x4f00a400, "SXTL2", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41}},
	// TBL <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBLvvv_t1: {0x0e000000, "TBL", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_1_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBL <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBLvvv_t2: {0x0e002000, "TBL", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_2_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBL <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBLvvv_t3: {0x0e004000, "TBL", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_3_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBL <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBLvvv_t4: {0x0e006000, "TBL", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_4_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBX <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBXvvv_t1: {0x0e001000, "TBX", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_1_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBX <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBXvvv_t2: {0x0e003000, "TBX", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_2_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBX <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBXvvv_t3: {0x0e005000, "TBX", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_3_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TBX <Vd>.<ta>, <Vn>.16B, <Vm>.<ta>
	TBXvvv_t4: {0x0e007000, "TBX", instArgs{arg_Vd_arrangement_Q___8B_0__16B_1, arg_Vn_4_arrangement_16B, arg_Vm_arrangement_Q___8B_0__16B_1}},
	// TRN1 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	TRN1vvv_t: {0x0e002800, "TRN1", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// TRN2 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	TRN2vvv_t: {0x0e006800, "TRN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UABA <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UABAvvv_t: {0x2e207c00, "UABA", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UABAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UABALvvv_t: {0x2e205000, "UABAL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UABAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UABAL2vvv_t: {0x6e205000, "UABAL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UABD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UABDvvv_t: {0x2e207400, "UABD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UABDL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UABDLvvv_t: {0x2e207000, "UABDL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UABDL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UABDL2vvv_t: {0x6e207000, "UABDL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADALP <Vd>.<ta>, <Vn>.<tb>
	UADALPvv_t: {0x2e206800, "UADALP", instArgs{arg_Vd_arrangement_size_Q___4H_00__8H_01__2S_10__4S_11__1D_20__2D_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADDL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UADDLvvv_t: {0x2e200000, "UADDL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADDL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UADDL2vvv_t: {0x6e200000, "UADDL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADDLP <Vd>.<ta>, <Vn>.<tb>
	UADDLPvv_t: {0x2e202800, "UADDLP", instArgs{arg_Vd_arrangement_size_Q___4H_00__8H_01__2S_10__4S_11__1D_20__2D_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADDLV <V><d>, <Vn>.<t>
	UADDLVvv_t: {0x2e303800, "UADDLV", instArgs{arg_Vd_22_2__H_0__S_1__D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// UADDW <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	UADDWvvv_t: {0x2e201000, "UADDW", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UADDW2 <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	UADDW2vvv_t: {0x6e201000, "UADDW2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UCVTF <Sd>, <Wn>, #<fbits>
	UCVTFswi: {0x1e030000, "UCVTF", instArgs{arg_Sd, arg_Wn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// UCVTF <Dd>, <Wn>, #<fbits>
	UCVTFdwi: {0x1e430000, "UCVTF", instArgs{arg_Dd, arg_Wn, arg_immediate_fbits_min_1_max_32_sub_64_scale}},
	// UCVTF <Sd>, <Xn>, #<fbits>
	UCVTFsxi: {0x9e030000, "UCVTF", instArgs{arg_Sd, arg_Xn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// UCVTF <Dd>, <Xn>, #<fbits>
	UCVTFdxi: {0x9e430000, "UCVTF", instArgs{arg_Dd, arg_Xn, arg_immediate_fbits_min_1_max_64_sub_64_scale}},
	// UCVTF <Sd>, <Wn>
	UCVTFsw: {0x1e230000, "UCVTF", instArgs{arg_Sd, arg_Wn}},
	// UCVTF <Dd>, <Wn>
	UCVTFdw: {0x1e630000, "UCVTF", instArgs{arg_Dd, arg_Wn}},
	// UCVTF <Sd>, <Xn>
	UCVTFsx: {0x9e230000, "UCVTF", instArgs{arg_Sd, arg_Xn}},
	// UCVTF <Dd>, <Xn>
	UCVTFdx: {0x9e630000, "UCVTF", instArgs{arg_Dd, arg_Xn}},
	// UCVTF <V><d>, <V><n>, #<fbits>
	UCVTFvvi: {0x7f00e400, "UCVTF", instArgs{arg_Vd_19_4__S_4__D_8, arg_Vn_19_4__S_4__D_8, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// UCVTF <Vd>.<t>, <Vn>.<t>, #<fbits>
	UCVTFvvi_t: {0x2f00e400, "UCVTF", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__2S_40__4S_41__2D_81, arg_immediate_fbits_min_1_max_0_sub_0_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// UCVTF <V><d>, <V><n>
	UCVTFvv: {0x7e21d800, "UCVTF", instArgs{arg_Vd_22_1__S_0__D_1, arg_Vn_22_1__S_0__D_1}},
	// UCVTF <Vd>.<t>, <Vn>.<t>
	UCVTFvv_t: {0x2e21d800, "UCVTF", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11, arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11}},
	// UHADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UHADDvvv_t: {0x2e200400, "UHADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UHSUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UHSUBvvv_t: {0x2e202400, "UHSUB", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMAX <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UMAXvvv_t: {0x2e206400, "UMAX", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMAXP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UMAXPvvv_t: {0x2e20a400, "UMAXP", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMAXV <V><d>, <Vn>.<t>
	UMAXVvv_t: {0x2e30a800, "UMAXV", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// UMIN <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UMINvvv_t: {0x2e206c00, "UMIN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMINP <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UMINPvvv_t: {0x2e20ac00, "UMINP", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMINV <V><d>, <Vn>.<t>
	UMINVvv_t: {0x2e31a800, "UMINV", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21}},
	// UMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMLALvvv_ti: {0x2f002000, "UMLAL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMLAL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMLALvvv_t: {0x2e208000, "UMLAL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMLAL2vvv_ti: {0x6f002000, "UMLAL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMLAL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMLAL2vvv_t: {0x6e208000, "UMLAL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMLSLvvv_ti: {0x2f006000, "UMLSL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMLSL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMLSLvvv_t: {0x2e20a000, "UMLSL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMLSL2vvv_ti: {0x6f006000, "UMLSL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMLSL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMLSL2vvv_t: {0x6e20a000, "UMLSL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMOV <Wd>, <Vn>.<ts>[<index>]
	UMOVwv_ti: {0x0e003c00, "UMOV", instArgs{arg_Wd, arg_Vn_arrangement_imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1}},
	// UMOV <Xd>, <Vn>.<ts_1>[<index_1>]
	UMOVxv_ti: {0x4e003c00, "UMOV", instArgs{arg_Xd, arg_Vn_arrangement_imm5___D_8_index__imm5_1}},
	// UMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMULLvvv_ti: {0x2f00a000, "UMULL", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMULL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMULLvvv_t: {0x2e20c000, "UMULL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<ts>[<index>]
	UMULL2vvv_ti: {0x6f00a000, "UMULL2", instArgs{arg_Vd_arrangement_size___4S_1__2D_2, arg_Vn_arrangement_size_Q___4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size___H_1__S_2_index__size_L_H_M__HLM_1__HL_2_1}},
	// UMULL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	UMULL2vvv_t: {0x6e20c000, "UMULL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UQADD <V><d>, <V><n>, <V><m>
	UQADDvvv: {0x7e200c00, "UQADD", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// UQADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UQADDvvv_t: {0x2e200c00, "UQADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UQRSHL <V><d>, <V><n>, <V><m>
	UQRSHLvvv: {0x7e205c00, "UQRSHL", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// UQRSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UQRSHLvvv_t: {0x2e205c00, "UQRSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UQRSHRN <V><d>, <V><n>, #<shift>
	UQRSHRNvvi: {0x7f009c00, "UQRSHRN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQRSHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	UQRSHRNvvi_t: {0x2f009c00, "UQRSHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQRSHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	UQRSHRN2vvi_t: {0x6f009c00, "UQRSHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQSHL <Vd>.<t>, <Vn>.<t>, #<shift>
	UQSHLvvi_t: {0x2f007400, "UQSHL", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// UQSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UQSHLvvv_t: {0x2e204c00, "UQSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UQSHL <V><d>, <V><n>, #<shift>
	UQSHLvvi: {0x7f007400, "UQSHL", instArgs{arg_Vd_19_4__B_1__H_2__S_4__D_8, arg_Vn_19_4__B_1__H_2__S_4__D_8, arg_immediate_0_width_m1_immh_immb__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8}},
	// UQSHL <V><d>, <V><n>, <V><m>
	UQSHLvvv: {0x7e204c00, "UQSHL", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// UQSHRN <V><d>, <V><n>, #<shift>
	UQSHRNvvi: {0x7f009400, "UQSHRN", instArgs{arg_Vd_19_4__B_1__H_2__S_4, arg_Vn_19_4__H_1__S_2__D_4, arg_immediate_1_width_immh_immb__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQSHRN <Vd>.<tb>, <Vn>.<ta>, #<shift>
	UQSHRNvvi_t: {0x2f009400, "UQSHRN", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQSHRN2 <Vd>.<tb>, <Vn>.<ta>, #<shift>
	UQSHRN2vvi_t: {0x6f009400, "UQSHRN2", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_Vn_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4}},
	// UQSUB <V><d>, <V><n>, <V><m>
	UQSUBvvv: {0x7e202c00, "UQSUB", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3, arg_Vm_22_2__B_0__H_1__S_2__D_3}},
	// UQSUB <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UQSUBvvv_t: {0x2e202c00, "UQSUB", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UQXTN <Vd>.<tb>, <Vn>.<ta>
	UQXTNvv_t: {0x2e214800, "UQXTN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// UQXTN <V><d>, <V><n>
	UQXTNvv: {0x7e214800, "UQXTN", instArgs{arg_Vd_22_2__B_0__H_1__S_2, arg_Vn_22_2__H_0__S_1__D_2}},
	// UQXTN2 <Vd>.<tb>, <Vn>.<ta>
	UQXTN2vv_t: {0x6e214800, "UQXTN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// URECPE <Vd>.<t>, <Vn>.<t>
	URECPEvv_t: {0x0ea1c800, "URECPE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01, arg_Vn_arrangement_sz_Q___2S_00__4S_01}},
	// URHADD <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	URHADDvvv_t: {0x2e201400, "URHADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// URSHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	URSHLvvv_t: {0x2e205400, "URSHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// URSHL <V><d>, <V><n>, <V><m>
	URSHLvvv: {0x7e205400, "URSHL", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// URSHR <Vd>.<t>, <Vn>.<t>, #<shift>
	URSHRvvi_t: {0x2f002400, "URSHR", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// URSHR <V><d>, <V><n>, #<shift>
	URSHRvvi: {0x7f002400, "URSHR", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// URSQRTE <Vd>.<t>, <Vn>.<t>
	URSQRTEvv_t: {0x2ea1c800, "URSQRTE", instArgs{arg_Vd_arrangement_sz_Q___2S_00__4S_01, arg_Vn_arrangement_sz_Q___2S_00__4S_01}},
	// URSRA <V><d>, <V><n>, #<shift>
	URSRAvvi: {0x7f003400, "URSRA", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// URSRA <Vd>.<t>, <Vn>.<t>, #<shift>
	URSRAvvi_t: {0x2f003400, "URSRA", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// USHL <V><d>, <V><n>, <V><m>
	USHLvvv: {0x7e204400, "USHL", instArgs{arg_Vd_22_2__D_3, arg_Vn_22_2__D_3, arg_Vm_22_2__D_3}},
	// USHL <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	USHLvvv_t: {0x2e204400, "USHL", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// USHLL <Vd>.<ta>, <Vn>.<tb>, #<shift>
	USHLLvvi_t: {0x2f00a400, "USHLL", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4}},
	// USHLL2 <Vd>.<ta>, <Vn>.<tb>, #<shift>
	USHLL2vvi_t: {0x6f00a400, "USHLL2", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41, arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4}},
	// USHR <Vd>.<t>, <Vn>.<t>, #<shift>
	USHRvvi_t: {0x2f000400, "USHR", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// USHR <V><d>, <V><n>, #<shift>
	USHRvvi: {0x7f000400, "USHR", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// USQADD <V><d>, <V><n>
	USQADDvv: {0x7e203800, "USQADD", instArgs{arg_Vd_22_2__B_0__H_1__S_2__D_3, arg_Vn_22_2__B_0__H_1__S_2__D_3}},
	// USQADD <Vd>.<t>, <Vn>.<t>
	USQADDvv_t: {0x2e203800, "USQADD", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// USRA <V><d>, <V><n>, #<shift>
	USRAvvi: {0x7f001400, "USRA", instArgs{arg_Vd_19_4__D_8, arg_Vn_19_4__D_8, arg_immediate_1_64_immh_immb__128UIntimmhimmb_8}},
	// USRA <Vd>.<t>, <Vn>.<t>, #<shift>
	USRAvvi_t: {0x2f001400, "USRA", instArgs{arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81, arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8}},
	// USUBL <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	USUBLvvv_t: {0x2e202000, "USUBL", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// USUBL2 <Vd>.<ta>, <Vn>.<tb>, <Vm>.<tb>
	USUBL2vvv_t: {0x6e202000, "USUBL2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// USUBW <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	USUBWvvv_t: {0x2e203000, "USUBW", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// USUBW2 <Vd>.<ta>, <Vn>.<ta>, <Vm>.<tb>
	USUBW2vvv_t: {0x6e203000, "USUBW2", instArgs{arg_Vd_arrangement_size___8H_0__4S_1__2D_2, arg_Vn_arrangement_size___8H_0__4S_1__2D_2, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21}},
	// UXTL <Vd>.<ta>, <Vn>.<tb>
	UXTLvv_t: {0x2f00a400, "UXTL", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41}},
	// UXTL2 <Vd>.<ta>, <Vn>.<tb>
	UXTL2vv_t: {0x6f00a400, "UXTL2", instArgs{arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4, arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41}},
	// UZP1 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UZP1vvv_t: {0x0e001800, "UZP1", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// UZP2 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	UZP2vvv_t: {0x0e005800, "UZP2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// XAR <Vd>.2D, <Vn>.2D, <Vm>.2D, #<imm6>
	XARvvvi_t: {0xce800000, "XAR", instArgs{arg_Vd_arrangement_2D, arg_Vn_arrangement_2D, arg_Vm_arrangement_2D, arg_immediate_0_63_imm6}},
	// XTN <Vd>.<tb>, <Vn>.<ta>
	XTNvv_t: {0x0e212800, "XTN", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// XTN2 <Vd>.<tb>, <Vn>.<ta>
	XTN2vv_t: {0x4e212800, "XTN2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21, arg_Vn_arrangement_size___8H_0__4S_1__2D_2}},
	// ZIP1 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ZIP1vvv_t: {0x0e003800, "ZIP1", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
	// ZIP2 <Vd>.<t>, <Vn>.<t>, <Vm>.<t>
	ZIP2vvv_t: {0x0e007800, "ZIP2", instArgs{arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31, arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31}},
}
