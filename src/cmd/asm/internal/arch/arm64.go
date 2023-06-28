// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the ARM64
// instruction set, to minimize its interaction with the core of the
// assembler.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/arm64"
	"errors"
)

var arm64LS = map[string]uint8{
	"P": arm64.C_XPOST,
	"W": arm64.C_XPRE,
}

var arm64Jump = map[string]bool{
	"B":     true,
	"BL":    true,
	"BEQ":   true,
	"BNE":   true,
	"BCS":   true,
	"BHS":   true,
	"BCC":   true,
	"BLO":   true,
	"BMI":   true,
	"BPL":   true,
	"BVS":   true,
	"BVC":   true,
	"BHI":   true,
	"BLS":   true,
	"BGE":   true,
	"BLT":   true,
	"BGT":   true,
	"BLE":   true,
	"CALL":  true,
	"CBZ":   true,
	"CBZW":  true,
	"CBNZ":  true,
	"CBNZW": true,
	"JMP":   true,
	"TBNZ":  true,
	"TBZ":   true,

	// ADR isn't really a jump, but it takes a PC or label reference,
	// which needs to patched like a jump.
	"ADR":  true,
	"ADRP": true,
}

func jumpArm64(word string) bool {
	return arm64Jump[word]
}

var arm64SpecialOperand map[string]arm64.SpecialOperand

// GetARM64SpecialOperand returns the internal representation of a special operand.
func GetARM64SpecialOperand(name string) arm64.SpecialOperand {
	if arm64SpecialOperand == nil {
		// Generate the mapping automatically when the first time the function is called.
		arm64SpecialOperand = map[string]arm64.SpecialOperand{}
		for opd := arm64.SPOP_BEGIN; opd < arm64.SPOP_END; opd++ {
			arm64SpecialOperand[opd.String()] = opd
		}

		// Handle some special cases.
		specialMapping := map[string]arm64.SpecialOperand{
			// The internal representation of CS(CC) and HS(LO) are the same.
			"CS": arm64.SPOP_HS,
			"CC": arm64.SPOP_LO,
		}
		for s, opd := range specialMapping {
			arm64SpecialOperand[s] = opd
		}
	}
	if opd, ok := arm64SpecialOperand[name]; ok {
		return opd
	}
	return arm64.SPOP_END
}

// IsARM64ADR reports whether the op (as defined by an arm64.A* constant) is
// one of the comparison instructions that require special handling.
func IsARM64ADR(op obj.As) bool {
	switch op {
	case arm64.AADR, arm64.AADRP:
		return true
	}
	return false
}

// IsARM64CMP reports whether the op (as defined by an arm64.A* constant) is
// one of the comparison instructions that require special handling.
func IsARM64CMP(op obj.As) bool {
	switch op {
	case arm64.ACMN, arm64.ACMP, arm64.ATST,
		arm64.ACMNW, arm64.ACMPW, arm64.ATSTW,
		arm64.AFCMPS, arm64.AFCMPD,
		arm64.AFCMPES, arm64.AFCMPED:
		return true
	}
	return false
}

// IsARM64STLXR reports whether the op (as defined by an arm64.A*
// constant) is one of the STLXR-like instructions that require special
// handling.
func IsARM64STLXR(op obj.As) bool {
	switch op {
	case arm64.ASTLXRB, arm64.ASTLXRH, arm64.ASTLXRW, arm64.ASTLXR,
		arm64.ASTXRB, arm64.ASTXRH, arm64.ASTXRW, arm64.ASTXR,
		arm64.ASTXP, arm64.ASTXPW, arm64.ASTLXP, arm64.ASTLXPW:
		return true
	}
	// LDADDx/SWPx/CASx atomic instructions
	return arm64.IsAtomicInstruction(op)
}

// IsARM64TBL reports whether the op (as defined by an arm64.A*
// constant) is one of the TBL-like instructions and one of its
// inputs does not fit into prog.Reg, so require special handling.
func IsARM64TBL(op obj.As) bool {
	switch op {
	case arm64.AVTBL, arm64.AVTBX, arm64.AVMOVQ:
		return true
	}
	return false
}

// IsARM64CASP reports whether the op (as defined by an arm64.A*
// constant) is one of the CASP-like instructions, and its 2nd
// destination is a register pair that require special handling.
func IsARM64CASP(op obj.As) bool {
	switch op {
	case arm64.ACASPD, arm64.ACASPW:
		return true
	}
	return false
}

// ARM64Suffix handles the special suffix for the ARM64.
// It returns a boolean to indicate success; failure means
// cond was unrecognized.
func ARM64Suffix(prog *obj.Prog, cond string) bool {
	if cond == "" {
		return true
	}
	bits, ok := parseARM64Suffix(cond)
	if !ok {
		return false
	}
	prog.Scond = bits
	return true
}

// parseARM64Suffix parses the suffix attached to an ARM64 instruction.
// The input is a single string consisting of period-separated condition
// codes, such as ".P.W". An initial period is ignored.
func parseARM64Suffix(cond string) (uint8, bool) {
	if cond == "" {
		return 0, true
	}
	return parseARMCondition(cond, arm64LS, nil)
}

func arm64RegisterNumber(name string, n int16) (int16, bool) {
	switch name {
	case "F":
		if 0 <= n && n <= 31 {
			return arm64.REG_F0 + n, true
		}
	case "R":
		if 0 <= n && n <= 30 { // not 31
			return arm64.REG_R0 + n, true
		}
	case "V":
		if 0 <= n && n <= 31 {
			return arm64.REG_V0 + n, true
		}
	}
	return 0, false
}

// ARM64RegisterExtension constructs an ARM64 register with extension or arrangement.
func ARM64RegisterExtension(a *obj.Addr, ext string, reg, num int16, isAmount, isIndex bool) error {
	if isAmount {
		if num < 0 || num > 7 {
			return errors.New("index shift amount is out of range")
		}
	}
	if reg <= arm64.REG_R31 && reg >= arm64.REG_R0 {
		if !isAmount {
			return errors.New("invalid register extension")
		}
		switch ext {
		case "UXTB":
			if a.Type == obj.TYPE_MEM {
				return errors.New("invalid shift for the register offset addressing mode")
			}
			a.Reg = reg
			a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_UXTB, num)
		case "UXTH":
			if a.Type == obj.TYPE_MEM {
				return errors.New("invalid shift for the register offset addressing mode")
			}
			a.Reg = reg
			a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_UXTH, num)
		case "UXTW":
			// effective address of memory is a base register value and an offset register value.
			if a.Type == obj.TYPE_MEM {
				a.Offset = int64(arm64.EncodeIndex(0, arm64.RTYP_EXT_UXTW, num))<<16 | int64(reg)
				a.Index |= arm64.RTYP_MEM_ROFF << 6
			} else {
				a.Reg = reg
				a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_UXTW, num)
			}
		case "UXTX":
			if a.Type == obj.TYPE_MEM {
				return errors.New("invalid shift for the register offset addressing mode")
			}
			a.Reg = reg
			a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_UXTX, num)
		case "SXTB":
			if a.Type == obj.TYPE_MEM {
				return errors.New("invalid shift for the register offset addressing mode")
			}
			a.Reg = reg
			a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTB, num)
		case "SXTH":
			if a.Type == obj.TYPE_MEM {
				return errors.New("invalid shift for the register offset addressing mode")
			}
			a.Reg = reg
			a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTH, num)
		case "SXTW":
			if a.Type == obj.TYPE_MEM {
				a.Offset = int64(arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTW, num))<<16 | int64(reg)
				a.Index |= arm64.RTYP_MEM_ROFF << 6
			} else {
				a.Reg = reg
				a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTW, num)
			}
		case "SXTX":
			if a.Type == obj.TYPE_MEM {
				a.Offset = int64(arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTX, num))<<16 | int64(reg)
				a.Index |= arm64.RTYP_MEM_ROFF << 6
			} else {
				a.Reg = reg
				a.Index = arm64.EncodeIndex(0, arm64.RTYP_EXT_SXTX, num)
			}
		case "LSL":
			a.Offset = int64(arm64.EncodeIndex(0, arm64.RTYP_EXT_LSL, num))<<16 | int64(reg)
			a.Index |= arm64.RTYP_MEM_ROFF << 6
		default:
			return errors.New("unsupported general register extension type: " + ext)

		}
	} else if reg <= arm64.REG_V31 && reg >= arm64.REG_V0 {
		a.Reg = reg
		switch ext {
		case "B8":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_8B, 0, 0)
		case "B16":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_16B, 0, 0)
		case "H4":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_4H, 0, 0)
		case "H8":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_8H, 0, 0)
		case "S2":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_2S, 0, 0)
		case "S4":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_4S, 0, 0)
		case "D1":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_1D, 0, 0)
		case "D2":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_2D, 0, 0)
		case "Q1":
			if isIndex {
				return errors.New("invalid register extension")
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_1Q, 0, 0)
		case "B":
			if !isIndex {
				return nil
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_B, arm64.RTYP_INDEX, num)
		case "H":
			if !isIndex {
				return nil
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_H, arm64.RTYP_INDEX, num)
		case "S":
			if !isIndex {
				return nil
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_S, arm64.RTYP_INDEX, num)
		case "D":
			if !isIndex {
				return nil
			}
			a.Index = arm64.EncodeIndex(arm64.ARNG_D, arm64.RTYP_INDEX, num)
		case "B4":
			if !isIndex {
				a.Index = arm64.EncodeIndex(arm64.ARNG_4B, 0, 0)
			} else {
				a.Index = arm64.EncodeIndex(arm64.ARNG_4B, arm64.RTYP_INDEX, num)
			}
		case "H2":
			if !isIndex {
				a.Index = arm64.EncodeIndex(arm64.ARNG_2H, 0, 0)
			} else {
				a.Index = arm64.EncodeIndex(arm64.ARNG_2H, arm64.RTYP_INDEX, num)
			}
		default:
			return errors.New("unsupported simd register extension type: " + ext)
		}
	} else {
		return errors.New("invalid register and extension combination")
	}
	return nil
}

// ARM64RegisterArrangement resolves an ARM64 vector register arrangement.
func ARM64RegisterArrangement(reg int16, name, arng string) (int, error) {
	if name[0] != 'V' {
		return 0, errors.New("expect V0 through V31; found: " + name)
	}
	if reg < 0 {
		return 0, errors.New("invalid register number: " + name)
	}
	switch arng {
	case "B8":
		return arm64.ARNG_8B, nil
	case "B16":
		return arm64.ARNG_16B, nil
	case "H4":
		return arm64.ARNG_4H, nil
	case "H8":
		return arm64.ARNG_8H, nil
	case "S2":
		return arm64.ARNG_2S, nil
	case "S4":
		return arm64.ARNG_4S, nil
	case "D1":
		return arm64.ARNG_1D, nil
	case "D2":
		return arm64.ARNG_2D, nil
	case "B":
		return arm64.ARNG_B, nil
	case "H":
		return arm64.ARNG_H, nil
	case "S":
		return arm64.ARNG_S, nil
	case "D":
		return arm64.ARNG_D, nil
	default:
		return 0, errors.New("invalid arrangement in ARM64 register list")
	}
}

// ARM64RegisterListOffset generates offset encoding
func ARM64RegisterListOffset(a *obj.Addr, firstreg, regCnt, arrangement, index, scale int) error {
	if regCnt < 1 || regCnt > 4 {
		return errors.New("invalid register numbers in ARM64 register list")
	}
	// Arm64 register list encoding scheme
	// For more details, refer to: obj/arm64/list7.go
	a.Reg = int16(firstreg)
	if index == 0 {
		a.Index = arm64.EncodeIndex(int16(arrangement), arm64.RTYP_NORMAL, int16(index))
	} else {
		a.Index = arm64.EncodeIndex(int16(arrangement), arm64.RTYP_INDEX, int16(index))
	}
	a.Scale = int16(scale)
	a.Offset = int64(regCnt)
	return nil
}
