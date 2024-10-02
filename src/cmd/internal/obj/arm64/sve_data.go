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
	AZBFDOT: {
		{0x64608000, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFDOT <Zda>.S, <Zn>.H, <Zm>.H
		{0x64604000, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i2}, // BFDOT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
}

// Key into the format table.
const (
	F_none = iota
	F_ZdaS_ZnH_ZmH
	F_ZdaS_ZnH_ZmHidx
)

// Format groups, common patterns of associated instruction formats. E.g. expansion of the <T> generic lane size.
var FG_none = []int{F_none} // No arguments.

// The format table holds a representation of the operand syntax for an instruction.
var formats = map[int]format{
	F_none:            []int{},                                                    // No arguments.
	F_ZdaS_ZnH_ZmH:    []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z | EXT_H},         // <Zd>.S, <Zn>.H, <Zm>.H
	F_ZdaS_ZnH_ZmHidx: []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H}, // <Zd>.S, <Zn>.H, <Zm>.H[<imm>]
}

// Key into the encoder table.
const (
	E_none = iota
	E_Rd_Rn_Rm
	E_Zda_Zn_Zm_i2
)

// The encoder table holds a list of encoding schemes for operands. Each scheme contains
// a list of rules and a list of indices that mark which operands need to be fed into the
// rule. Each rule produces a 32-bit number which should be OR'd with the base to create
// an instruction encoding.
var encoders = map[int]encoder{
	E_none:         {},
	E_Rd_Rn_Rm:     {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_Zda_Zn_Zm_i2: {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
}
