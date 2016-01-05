// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/arm"
	"cmd/internal/obj/arm64"
	"cmd/internal/obj/mips"
	"cmd/internal/obj/ppc64"
	"cmd/internal/obj/x86"
	"fmt"
	"strings"
)

// Pseudo-registers whose names are the constant name without the leading R.
const (
	RFP = -(iota + 1)
	RSB
	RSP
	RPC
)

// Arch wraps the link architecture object with more architecture-specific information.
type Arch struct {
	*obj.LinkArch
	// Map of instruction names to enumeration.
	Instructions map[string]int
	// Map of register names to enumeration.
	Register map[string]int16
	// Table of register prefix names. These are things like R for R(0) and SPR for SPR(268).
	RegisterPrefix map[string]bool
	// RegisterNumber converts R(10) into arm.REG_R10.
	RegisterNumber func(string, int16) (int16, bool)
	// Instruction is a jump.
	IsJump func(word string) bool
}

// nilRegisterNumber is the register number function for architectures
// that do not accept the R(N) notation. It always returns failure.
func nilRegisterNumber(name string, n int16) (int16, bool) {
	return 0, false
}

var Pseudos = map[string]int{
	"DATA":     obj.ADATA,
	"FUNCDATA": obj.AFUNCDATA,
	"GLOBL":    obj.AGLOBL,
	"PCDATA":   obj.APCDATA,
	"TEXT":     obj.ATEXT,
}

// Set configures the architecture specified by GOARCH and returns its representation.
// It returns nil if GOARCH is not recognized.
func Set(GOARCH string) *Arch {
	switch GOARCH {
	case "386":
		return archX86(&x86.Link386)
	case "amd64":
		return archX86(&x86.Linkamd64)
	case "amd64p32":
		return archX86(&x86.Linkamd64p32)
	case "arm":
		return archArm()
	case "arm64":
		return archArm64()
	case "mips64":
		a := archMips64()
		a.LinkArch = &mips.Linkmips64
		return a
	case "mips64le":
		a := archMips64()
		a.LinkArch = &mips.Linkmips64le
		return a
	case "ppc64":
		a := archPPC64()
		a.LinkArch = &ppc64.Linkppc64
		return a
	case "ppc64le":
		a := archPPC64()
		a.LinkArch = &ppc64.Linkppc64le
		return a
	}
	return nil
}

func jumpX86(word string) bool {
	return word[0] == 'J' || word == "CALL" || strings.HasPrefix(word, "LOOP") || word == "XBEGIN"
}

var aliasesX86 = map[string]int{
	// alternate names for condition codes
	"JA":   x86.AJHI, /* alternate */
	"JAE":  x86.AJCC, /* alternate */
	"JB":   x86.AJCS, /* alternate */
	"JBE":  x86.AJLS, /* alternate */
	"JC":   x86.AJCS, /* alternate */
	"JCC":  x86.AJCC, /* carry clear (CF = 0) */
	"JCS":  x86.AJCS, /* carry set (CF = 1) */
	"JE":   x86.AJEQ, /* alternate */
	"JEQ":  x86.AJEQ, /* equal (ZF = 1) */
	"JG":   x86.AJGT, /* alternate */
	"JGE":  x86.AJGE, /* greater than or equal (signed) (SF = OF) */
	"JGT":  x86.AJGT, /* greater than (signed) (ZF = 0 && SF = OF) */
	"JHI":  x86.AJHI, /* higher (unsigned) (CF = 0 && ZF = 0) */
	"JHS":  x86.AJCC, /* alternate */
	"JL":   x86.AJLT, /* alternate */
	"JLE":  x86.AJLE, /* less than or equal (signed) (ZF = 1 || SF != OF) */
	"JLO":  x86.AJCS, /* alternate */
	"JLS":  x86.AJLS, /* lower or same (unsigned) (CF = 1 || ZF = 1) */
	"JLT":  x86.AJLT, /* less than (signed) (SF != OF) */
	"JMI":  x86.AJMI, /* negative (minus) (SF = 1) */
	"JNA":  x86.AJLS, /* alternate */
	"JNAE": x86.AJCS, /* alternate */
	"JNB":  x86.AJCC, /* alternate */
	"JNBE": x86.AJHI, /* alternate */
	"JNC":  x86.AJCC, /* alternate */
	"JNE":  x86.AJNE, /* not equal (ZF = 0) */
	"JNG":  x86.AJLE, /* alternate */
	"JNGE": x86.AJLT, /* alternate */
	"JNL":  x86.AJGE, /* alternate */
	"JNLE": x86.AJGT, /* alternate */
	"JNO":  x86.AJOC, /* alternate */
	"JNP":  x86.AJPC, /* alternate */
	"JNS":  x86.AJPL, /* alternate */
	"JNZ":  x86.AJNE, /* alternate */
	"JO":   x86.AJOS, /* alternate */
	"JOC":  x86.AJOC, /* overflow clear (OF = 0) */
	"JOS":  x86.AJOS, /* overflow set (OF = 1) */
	"JP":   x86.AJPS, /* alternate */
	"JPC":  x86.AJPC, /* parity clear (PF = 0) */
	"JPE":  x86.AJPS, /* alternate */
	"JPL":  x86.AJPL, /* non-negative (plus) (SF = 0) */
	"JPO":  x86.AJPC, /* alternate */
	"JPS":  x86.AJPS, /* parity set (PF = 1) */
	"JS":   x86.AJMI, /* alternate */
	"JZ":   x86.AJEQ, /* alternate */

	// Confusion about the 8-byte suffix.
	// In general Intel uses DQ, but Go on x86 uses O (octo).
	// Provide the DQ equivalents where possible.
	"LDDQU":      x86.ALDOU,
	"MASKMOVDQU": x86.AMASKMOVOU,
	"MOVDQ2Q":    x86.AMOVQ,
	"MOVNTDQ":    x86.AMOVNTO,
	"MOVDQU":     x86.AMOVOU,
	"PMULDQ":     x86.APMULO,
	"PSLLDQ":     x86.APSLLO,
	"PSRLDQ":     x86.APSRLO,
	"PUNPCKHQDQ": x86.APUNPCKHQO,
	"PUNPCKLQDQ": x86.APUNPCKLQO,

	// Other architectures use D for 8-byte.
	// Provide MOVD, so that there is a common 8-byte MOV across architectures.
	"MOVD": x86.AMOVQ,

	// MOVO is the aligned form aka MOVOA.
	// And similarly MOVDQA (see note above about DQ).
	"MOVOA":  x86.AMOVO,
	"MOVDQA": x86.AMOVO,

	// Converting a packed 4-byte is PL here,
	// but in the Intel manuals they write DQ.
	"CVTPD2DQ":  x86.ACVTPD2PL,
	"CVTDQ2PD":  x86.ACVTPL2PD,
	"CVTDQ2PS":  x86.ACVTPL2PS,
	"CVTPS2DQ":  x86.ACVTPS2PL,
	"CVTTPD2DQ": x86.ACVTTPD2PL,
	"CVTTPS2DQ": x86.ACVTTPS2PL,

	// Ancient aliases.
	"PF2ID": x86.APF2IL,
	"PI2FD": x86.API2FL,
}

func archX86(linkArch *obj.LinkArch) *Arch {
	register := make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	for i, s := range x86.Register {
		register[s] = int16(i + x86.REG_AL)
	}
	// Pseudo-registers.
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	// Register prefix not used on this architecture.

	instructions := make(map[string]int)
	for i, s := range obj.Anames {
		instructions[s] = i
	}
	for i, s := range x86.Anames {
		if i >= obj.A_ARCHSPECIFIC {
			instructions[s] = i + obj.ABaseAMD64
		}
	}
	// Annoying aliases.
	for name, alt := range aliasesX86 {
		instructions[name] = alt
	}

	return &Arch{
		LinkArch:       linkArch,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: nil,
		RegisterNumber: nilRegisterNumber,
		IsJump:         jumpX86,
	}
}

func archArm() *Arch {
	register := make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	// Note that there is no list of names as there is for x86.
	for i := arm.REG_R0; i < arm.REG_SPSR; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	// Avoid unintentionally clobbering g using R10.
	delete(register, "R10")
	register["g"] = arm.REG_R10
	for i := 0; i < 16; i++ {
		register[fmt.Sprintf("C%d", i)] = int16(i)
	}

	// Pseudo-registers.
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	register["SP"] = RSP
	registerPrefix := map[string]bool{
		"F": true,
		"R": true,
	}

	instructions := make(map[string]int)
	for i, s := range obj.Anames {
		instructions[s] = i
	}
	for i, s := range arm.Anames {
		if i >= obj.A_ARCHSPECIFIC {
			instructions[s] = i + obj.ABaseARM
		}
	}
	// Annoying aliases.
	instructions["B"] = obj.AJMP
	instructions["BL"] = obj.ACALL
	// MCR differs from MRC by the way fields of the word are encoded.
	// (Details in arm.go). Here we add the instruction so parse will find
	// it, but give it an opcode number known only to us.
	instructions["MCR"] = aMCR

	return &Arch{
		LinkArch:       &arm.Linkarm,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: armRegisterNumber,
		IsJump:         jumpArm,
	}
}

func archArm64() *Arch {
	register := make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	// Note that there is no list of names as there is for 386 and amd64.
	register[arm64.Rconv(arm64.REGSP)] = int16(arm64.REGSP)
	for i := arm64.REG_R0; i <= arm64.REG_R31; i++ {
		register[arm64.Rconv(i)] = int16(i)
	}
	for i := arm64.REG_F0; i <= arm64.REG_F31; i++ {
		register[arm64.Rconv(i)] = int16(i)
	}
	for i := arm64.REG_V0; i <= arm64.REG_V31; i++ {
		register[arm64.Rconv(i)] = int16(i)
	}
	register["LR"] = arm64.REGLINK
	register["DAIF"] = arm64.REG_DAIF
	register["NZCV"] = arm64.REG_NZCV
	register["FPSR"] = arm64.REG_FPSR
	register["FPCR"] = arm64.REG_FPCR
	register["SPSR_EL1"] = arm64.REG_SPSR_EL1
	register["ELR_EL1"] = arm64.REG_ELR_EL1
	register["SPSR_EL2"] = arm64.REG_SPSR_EL2
	register["ELR_EL2"] = arm64.REG_ELR_EL2
	register["CurrentEL"] = arm64.REG_CurrentEL
	register["SP_EL0"] = arm64.REG_SP_EL0
	register["SPSel"] = arm64.REG_SPSel
	register["DAIFSet"] = arm64.REG_DAIFSet
	register["DAIFClr"] = arm64.REG_DAIFClr
	// Conditional operators, like EQ, NE, etc.
	register["EQ"] = arm64.COND_EQ
	register["NE"] = arm64.COND_NE
	register["HS"] = arm64.COND_HS
	register["CS"] = arm64.COND_HS
	register["LO"] = arm64.COND_LO
	register["CC"] = arm64.COND_LO
	register["MI"] = arm64.COND_MI
	register["PL"] = arm64.COND_PL
	register["VS"] = arm64.COND_VS
	register["VC"] = arm64.COND_VC
	register["HI"] = arm64.COND_HI
	register["LS"] = arm64.COND_LS
	register["GE"] = arm64.COND_GE
	register["LT"] = arm64.COND_LT
	register["GT"] = arm64.COND_GT
	register["LE"] = arm64.COND_LE
	register["AL"] = arm64.COND_AL
	register["NV"] = arm64.COND_NV
	// Pseudo-registers.
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	register["SP"] = RSP
	// Avoid unintentionally clobbering g using R28.
	delete(register, "R28")
	register["g"] = arm64.REG_R28
	registerPrefix := map[string]bool{
		"F": true,
		"R": true,
		"V": true,
	}

	instructions := make(map[string]int)
	for i, s := range obj.Anames {
		instructions[s] = i
	}
	for i, s := range arm64.Anames {
		if i >= obj.A_ARCHSPECIFIC {
			instructions[s] = i + obj.ABaseARM64
		}
	}
	// Annoying aliases.
	instructions["B"] = arm64.AB
	instructions["BL"] = arm64.ABL

	return &Arch{
		LinkArch:       &arm64.Linkarm64,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: arm64RegisterNumber,
		IsJump:         jumpArm64,
	}

}

func archPPC64() *Arch {
	register := make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	// Note that there is no list of names as there is for x86.
	for i := ppc64.REG_R0; i <= ppc64.REG_R31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := ppc64.REG_F0; i <= ppc64.REG_F31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := ppc64.REG_CR0; i <= ppc64.REG_CR7; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := ppc64.REG_MSR; i <= ppc64.REG_CR; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	register["CR"] = ppc64.REG_CR
	register["XER"] = ppc64.REG_XER
	register["LR"] = ppc64.REG_LR
	register["CTR"] = ppc64.REG_CTR
	register["FPSCR"] = ppc64.REG_FPSCR
	register["MSR"] = ppc64.REG_MSR
	// Pseudo-registers.
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	// Avoid unintentionally clobbering g using R30.
	delete(register, "R30")
	register["g"] = ppc64.REG_R30
	registerPrefix := map[string]bool{
		"CR":  true,
		"F":   true,
		"R":   true,
		"SPR": true,
	}

	instructions := make(map[string]int)
	for i, s := range obj.Anames {
		instructions[s] = i
	}
	for i, s := range ppc64.Anames {
		if i >= obj.A_ARCHSPECIFIC {
			instructions[s] = i + obj.ABasePPC64
		}
	}
	// Annoying aliases.
	instructions["BR"] = ppc64.ABR
	instructions["BL"] = ppc64.ABL

	return &Arch{
		LinkArch:       &ppc64.Linkppc64,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: ppc64RegisterNumber,
		IsJump:         jumpPPC64,
	}
}

func archMips64() *Arch {
	register := make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	// Note that there is no list of names as there is for x86.
	for i := mips.REG_R0; i <= mips.REG_R31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := mips.REG_F0; i <= mips.REG_F31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := mips.REG_M0; i <= mips.REG_M31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := mips.REG_FCR0; i <= mips.REG_FCR31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	register["HI"] = mips.REG_HI
	register["LO"] = mips.REG_LO
	// Pseudo-registers.
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	// Avoid unintentionally clobbering g using R30.
	delete(register, "R30")
	register["g"] = mips.REG_R30
	registerPrefix := map[string]bool{
		"F":   true,
		"FCR": true,
		"M":   true,
		"R":   true,
	}

	instructions := make(map[string]int)
	for i, s := range obj.Anames {
		instructions[s] = i
	}
	for i, s := range mips.Anames {
		if i >= obj.A_ARCHSPECIFIC {
			instructions[s] = i + obj.ABaseMIPS64
		}
	}
	// Annoying alias.
	instructions["JAL"] = mips.AJAL

	return &Arch{
		LinkArch:       &mips.Linkmips64,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: mipsRegisterNumber,
		IsJump:         jumpMIPS64,
	}
}
