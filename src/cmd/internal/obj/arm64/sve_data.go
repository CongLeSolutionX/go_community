// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import "cmd/internal/obj"

// Mapping of mnemonic -> list of encoding. Each instruction has multiple potential encodings,
// for each different format of operands possible with the instruction. For example, ADD has
// 3 different formats:
//   - ADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
//   - ADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
//   - ADD <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
//
// Generic lane sizes such as <T> may be expanded into further formats, e.g. we may expand
// <T> into {B,H,S,D} to make it easier for the assembler to find a match with user input.
var instructionTable = map[obj.As][]encoding{
	AZABS: {
		{0x0416a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // ABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZADD: {
		{0x04000000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZAND: {
		{0x041a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // AND <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZASR: {
		{0x04108000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZASRR: {
		{0x04148000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZBFDOT: {
		{0x64608000, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFDOT <Zda>.S, <Zn>.H, <Zm>.H
		{0x64604000, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i2}, // BFDOT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
	AZBIC: {
		{0x041b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // BIC <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZCLS: {
		{0x0418a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CLS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCLZ: {
		{0x0419a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CLZ <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCNOT: {
		{0x041ba000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CNOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCNT: {
		{0x041aa000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CNT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZEOR: {
		{0x04190000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // EOR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABD: {
		{0x65088000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABS: {
		{0x041ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFADD: {
		{0x65008000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFDIV: {
		{0x650d8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFDIVR: {
		{0x650c8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMAX: {
		{0x65068000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMAXNM: {
		{0x65048000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMAXNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMIN: {
		{0x65078000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMINNM: {
		{0x65058000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMINNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMUL: {
		{0x65028000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMULX: {
		{0x650a8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMULX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFNEG: {
		{0x041da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FNEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRECPX: {
		{0x650ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRECPX <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTA: {
		{0x6504a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTA <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTI: {
		{0x6507a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTI <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTM: {
		{0x6502a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTM <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTN: {
		{0x6500a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTN <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTP: {
		{0x6501a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTP <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTX: {
		{0x6506a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTX <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTZ: {
		{0x6503a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTZ <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFSCALE: {
		{0x65098000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSCALE <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSUB: {
		{0x65018000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSUBR: {
		{0x65038000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSQRT: {
		{0x650da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FSQRT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZLSL: {
		{0x04138000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSLR: {
		{0x04178000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSLR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSR: {
		{0x04118000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSRR: {
		{0x04158000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZMUL: {
		{0x04100000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // MUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZNEG: {
		{0x0417a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZNOT: {
		{0x041ea000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZORR: {
		{0x04180000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ORR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZRBIT: {
		{0x05278000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // RBIT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVB: {
		{0x05248000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // REVB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVH: {
		{0x05258000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // REVH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZSABD: {
		{0x040c0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSDIV: {
		{0x04140000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSDIVR: {
		{0x04160000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMAX: {
		{0x04080000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMIN: {
		{0x040a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMULH: {
		{0x04120000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSUB: {
		{0x04010000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSUBR: {
		{0x04030000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSXTB: {
		{0x0410a000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // SXTB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZSXTH: {
		{0x0412a000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // SXTH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUABD: {
		{0x040d0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUDIV: {
		{0x04150000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUDIVR: {
		{0x04170000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMAX: {
		{0x04090000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMIN: {
		{0x040b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMULH: {
		{0x04130000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUXTB: {
		{0x0411a000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // UXTB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUXTH: {
		{0x0413a000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // UXTH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
}

// Key into the format table.
const (
	F_ZdaS_ZnH_ZmH = iota
	F_ZdaS_ZnH_ZmHidx
	F_ZdB_PgM_ZnB
	F_ZdH_PgM_ZnH
	F_ZdS_PgM_ZnS
	F_ZdD_PgM_ZnD
	F_ZdnB_PgM_ZdnB_ZmB
	F_ZdnH_PgM_ZdnH_ZmH
	F_ZdnS_PgM_ZdnS_ZmS
	F_ZdnD_PgM_ZdnD_ZmD
)

// Format groups, common patterns of associated instruction formats. E.g. expansion of the <T> generic lane size.
var FG_ZdT_PgM_ZnT = []int{F_ZdB_PgM_ZnB, F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdT_PgM_ZnT_FP = []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdnT_PgM_ZdnT_ZmT = []int{F_ZdnB_PgM_ZdnB_ZmB, F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}
var FG_ZdnT_PgM_ZdnT_ZmT_FP = []int{F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}

// The format table holds a representation of the operand syntax for an instruction.
var formats = map[int]format{
	F_ZdaS_ZnH_ZmH:      []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z | EXT_H},                      // <Zd>.S, <Zn>.H, <Zm>.H
	F_ZdaS_ZnH_ZmHidx:   []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},              // <Zd>.S, <Zn>.H, <Zm>.H[<imm>]
	F_ZdB_PgM_ZnB:       []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B},                // <Zd>.B, <Pg>/M, <Zn>.B
	F_ZdH_PgM_ZnH:       []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H},                // <Zd>.H, <Pg>/M, <Zn>.H
	F_ZdS_PgM_ZnS:       []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S},                // <Zd>.S, <Pg>/M, <Zn>.S
	F_ZdD_PgM_ZnD:       []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D},                // <Zd>.D, <Pg>/M, <Zn>.D
	F_ZdnB_PgM_ZdnB_ZmB: []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, REG_Z | EXT_B}, // <Zdn>.B, <Pg>/M, <Zdn>.B, <Zm>.B
	F_ZdnH_PgM_ZdnH_ZmH: []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H}, // <Zdn>.H, <Pg>/M, <Zdn>.H, <Zm>.H
	F_ZdnS_PgM_ZdnS_ZmS: []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S}, // <Zdn>.S, <Pg>/M, <Zdn>.S, <Zm>.S
	F_ZdnD_PgM_ZdnD_ZmD: []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D}, // <Zdn>.D, <Pg>/M, <Zdn>.D, <Zm>.D
}

// Key into the encoder table.
const (
	E_Rd_Rn_Rm = iota
	E_Zda_Zn_Zm_i2
	E_size_Pg_Zn_Zd
	E_size_Pg_Zm_Zdn

	// Equivalences
	E_size0_Pg_Zn_Zd = E_size_Pg_Zn_Zd
)

// The encoder table holds a list of encoding schemes for operands. Each scheme contains
// a list of rules and a list of indices that mark which operands need to be fed into the
// rule. Each rule produces a 32-bit number which should be OR'd with the base to create
// an instruction encoding.
var encoders = map[int]encoder{
	E_Rd_Rn_Rm:       {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_Zda_Zn_Zm_i2:   {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_size_Pg_Zn_Zd:  {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}}},
	E_size_Pg_Zm_Zdn: {[]rule{{[]int{0, 2}, Zdn}, {[]int{1}, Pg}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
}
