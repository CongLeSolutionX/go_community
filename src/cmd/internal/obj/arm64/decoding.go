// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"strconv"
	"strings"
)

func (inst *inst) GoSyntax(code uint32) string {
	text := inst.goOp.String()
	var args []string
	for i, arg := range inst.args {
		args = append(args, arg.GoSyntax(inst, code, i))
	}
	if args != nil {
		text += " " + strings.Join(args, ", ")
	}
	return text
}

func (inst *inst) GNUSyntax(code uint32) string {
	text := inst.armOp.String()
	var args []string
	for i, arg := range inst.gnuOrderArgs() {
		args = append(args, arg.GNUSyntax(inst, code, i))
	}
	if args != nil {
		text += " " + strings.Join(args, ", ")
	}
	return text
}

// specialArgs reports whether the order of arguments of inst is special,
// and if so, adjusts the order of the returned arg slice to be consistent
// with the order of the arguments of the corresponding arm64 instruction.
func (inst *inst) specialArgs() ([]arg, bool) {
	special := true
	var args = make([]arg, len(inst.args))
	// arg.elms is not copied, but we don't change it, so it doesn't matter.
	copy(args, inst.args)
	switch inst.armOp {
	case A64_CBNZ, A64_CBZ,
		A64_GCSSTR, A64_GCSSTTR, A64_ST2G,
		A64_ST64B,
		A64_STADD, A64_STADDL, A64_STADDB, A64_STADDLB, A64_STADDH, A64_STADDLH,
		A64_STCLR, A64_STCLRL, A64_STCLRB, A64_STCLRLB, A64_STCLRH, A64_STCLRLH,
		A64_STEOR, A64_STEORL, A64_STEORB, A64_STEORLB, A64_STEORH, A64_STEORLH,
		A64_STSET, A64_STSETL, A64_STSETB, A64_STSETLB, A64_STSETH, A64_STSETLH,
		A64_STSMAX, A64_STSMAXL, A64_STSMAXB, A64_STSMAXLB, A64_STSMAXH, A64_STSMAXLH,
		A64_STSMIN, A64_STSMINL, A64_STSMINB, A64_STSMINLB, A64_STSMINH, A64_STSMINLH,
		A64_STUMAX, A64_STUMAXL, A64_STUMAXB, A64_STUMAXLB, A64_STUMAXH, A64_STUMAXLH,
		A64_STUMIN, A64_STUMINL, A64_STUMINB, A64_STUMINLB, A64_STUMINH, A64_STUMINLH,
		A64_STG, A64_STGM, A64_STGP,
		A64_STILP, A64_STLLR, A64_STLLRB, A64_STLLRH,
		A64_STLR, A64_STLRB, A64_STLRH,
		A64_STLUR, A64_STLURB, A64_STLURH,
		A64_STNP, A64_STP,
		A64_STR, A64_STRB, A64_STRH,
		A64_STTR, A64_STTRB, A64_STTRH,
		A64_STUR, A64_STURB, A64_STURH,
		A64_STZ2G, A64_STZG, A64_STZGM,
		A64_ST1, A64_ST2, A64_ST3, A64_ST4, A64_STL1,
		A64_ST1B, A64_ST1D, A64_ST1H, A64_ST1Q, A64_ST1W,
		A64_ST2B, A64_ST2D, A64_ST2H, A64_ST2Q, A64_ST2W,
		A64_ST3B, A64_ST3D, A64_ST3H, A64_ST3Q, A64_ST3W,
		A64_ST4B, A64_ST4D, A64_ST4H, A64_ST4Q, A64_ST4W,
		A64_STNT1B, A64_STNT1D, A64_STNT1H, A64_STNT1W:
		// Argument order is the same as the ARM64 syntax:
		// cbz, cbnz and store instructions, do nothing.
	case A64_STLXR, A64_STLXRB, A64_STLXRH, A64_STXR, A64_STXRB, A64_STXRH, A64_STLXP, A64_STXP:
		// stlxr w16, xzr, [x15] <=> STLXR ZR, (R15), R16
		// stlxp w5, x17, x19, [x4] <=> STLXP (R17, R19), (R4), R5
		args[0], args[1] = args[1], args[0]
		args[0], args[2] = args[2], args[0]
	case A64_MADD, A64_MSUB, A64_SMADDL, A64_SMSUBL, A64_UMADDL, A64_UMSUBL,
		A64_FMADD, A64_FMSUB, A64_FNMADD, A64_FNMSUB:
		// madd x1, x2, x3, x4 <=> MADD R3, R4, R2, R1
		// fmadd d1, d2, d3, d4 <=> FMADDD F3, F4, F2, F1
		args[0], args[2] = args[2], args[0]
		args[1], args[3] = args[3], args[1]
		args[0], args[1] = args[1], args[0]
	case A64_BFI, A64_BFM, A64_BFXIL,
		A64_SBFM, A64_SBFIZ, A64_SBFX,
		A64_UBFM, A64_UBFIZ, A64_UBFX:
		// bfi w0, w20, #16, #6 <=> BFIW $16, R20, $6, R0
		args[0], args[2] = args[2], args[0]
		args[0], args[3] = args[3], args[0]
	case A64_FCCMP, A64_FCCMPE:
		// fccmp d26, d8, #0x0, al <=> FCCMPD AL, F8, F26, $0
		args[0], args[3] = args[3], args[0]
		args[0], args[2] = args[2], args[0]
	case A64_CCMP, A64_CCMN:
		// ccmp w19, w14, #0xb, cs <=> CCMPW HS, R19, R14, $11
		args[0], args[1] = args[1], args[0]
		args[2], args[3] = args[3], args[2]
		args[1], args[3] = args[3], args[1]
	case A64_CSEL, A64_CSINC, A64_CSINV, A64_CSNEG, A64_FCSEL:
		// csel x1, x0, x19, gt <=> CSEL GT, R0, R19, R1
		args[0], args[3] = args[3], args[0]
	case A64_TBNZ, A64_TBZ:
		// tbz x1, #4, loop <=> TBZ $4, R1, loop
		args[0], args[1] = args[1], args[0]
	case A64_MOVI, A64_MVNI:
		if args[1].elms[0] == sa_amount_2__cmode_0__8_16 { // for "MSL #<amount>"
			// movi <Vd>.<T>, #<imm8>, MSL #<amount> <=> MOVI #<imm8>, MSL #<amount>, <Vd>.<T>
			args[1], args[2] = args[2], args[1]
			args[0], args[1] = args[1], args[0]
		} else {
			special = false
		}
	default:
		// atomic instructions with 3 operands, ST64BV and ST64BV0.
		// swpa x5, x7, [x6]  <=> SWPAD	R5, (R6), R7
		// cas  w5, w6, [x7]  <=> CASW	R5, (R7), R6
		elem := args[0].elms[0]
		if (elem == sa_xs__Rs || elem == sa_ws__Rs) && inst.feature != FEAT_MOPS && len(args) > 2 {
			args[1], args[2] = args[2], args[1]
		} else {
			special = false
		}
	}
	return args, special
}

func (inst *inst) gnuOrderArgs() []arg {
	if args, isSpecial := inst.specialArgs(); isSpecial {
		return args
	}
	var args = make([]arg, len(inst.args))
	// arg.elms is not copied, but we don't change it, so it doesn't matter.
	copy(args, inst.args)
	// For general cases, just reverse args.
	for i, j := 0, len(args)-1; i < j; i, j = i+1, j-1 {
		args[i], args[j] = args[j], args[i]
	}
	return args
}

func (arg arg) GoSyntax(inst *inst, code uint32, idx int) string {
	text := ""
	switch arg.aType {
	case AC_REG:
		switch arg.elms[0] {
		case sa_wa__Ra:
			ra := (code >> 10) & 0x1f
			if ra == 31 {
				text = "ZR"
			} else {
				text = "R" + strconv.Itoa(int(ra))
			}
		}
	}
	return text
}

func (arg arg) GNUSyntax(inst *inst, code uint32, idx int) string {
	text := ""
	switch arg.aType {
	case AC_REG:
		switch arg.elms[0] {
		case sa_wa__Ra:
			ra := (code >> 10) & 0x1f
			if ra == 31 {
				text = "wzr"
			} else {
				text = "w" + strconv.Itoa(int(ra))
			}
		}
	}
	return text
}
