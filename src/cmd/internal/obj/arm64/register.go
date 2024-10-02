// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"fmt"
	"log"
	"strconv"
)

// The Register type wraps an obj.Addr reference and can be used to read
// and modify register properties, including register number and extension.
type Register struct {
	uint16
}

// Registers are represented as shown in the diagram.
//    1
//    4     9      3   0
// +-+-+-----+------+---+
// |x|1|nnnnn|eeeeee|ttt|
// +-+-+-----+------+---+
// n: Register number
// e: Register extension (suffix) type
// t: Register type

// Bits 9-13 hold the register number in the range 0-31 for R/F/V/Z, 0-15 for P
const numOffset = 9
const numMask = 0b11111 << numOffset

// Bits 3-8 hold the register extension, such as the lane arrangement
const extOffset = 3
const extMask = 0b111111 << extOffset

// Bits 0-2 hold the register type R/F/V/Z/P/...
const groupMask = 0b111

// Types of register extension, add as needed.
const (
	EXT_NONE    = iota << subtypeOffset
	EXT_B       // .B
	EXT_H       // .H
	EXT_S       // .S
	EXT_D       // .D
	EXT_Q       // .Q
	EXT_MERGING // .M
	EXT_ZEROING // .Z
	EXT_END     // End marker
)

// Checks whether the encdoed register follows this representation scheme.
func IsSVECompatibleRegister(data int16) bool {
	return uint16(data)&REG_SVE != 0
}

// Create a register from its enumeration value (e.g. REG_R0) and extension (e.g. EXT_S).
func NewRegister(reg int16, etype int16) Register {
	if reg == REG_RSP {
		reg = REG_R31
	}
	rtype := (reg >> 5) - 1
	if rtype < 0 {
		rtype = REG_R
	}

	r := Register{uint16(reg&31)<<numOffset | uint16(rtype)&7 | REG_SVE}
	r.SetExt(etype)
	return r
}

func AsRegister(data int16) Register {
	if !IsSVECompatibleRegister(data) {
		panic("Addr is not compatible with this register scheme")
	}
	r := Register{uint16(data)}
	if r.Ext() >= EXT_END || r.Ext() < EXT_NONE {
		log.Panicf("Addr has invalid extension: %b", r.Ext())
	}
	return r
}

func (r *Register) Ext() int {
	return int((r.uint16&extMask)>>extOffset) << subtypeOffset
}

func (r *Register) SetExt(ext int16) {
	r.uint16 ^= (r.uint16 & extMask) ^ uint16((ext>>subtypeOffset)<<extOffset)
}

// Get the register number, between 0-31. For P registers between 0-15.
func (r *Register) Number() int16 {
	return int16((r.uint16 & numMask) >> numOffset)
}

func (r *Register) SetNumber(number int16) {
	r.uint16 ^= (r.uint16 & numMask) ^ ((uint16(number) << numOffset) & numMask)
}

func (r *Register) is(rtype uint16) bool {
	return r.uint16&groupMask == rtype
}

func (r *Register) IsR() bool {
	return r.is(REG_R)
}

func (r *Register) IsF() bool {
	return r.is(REG_F)
}

func (r *Register) IsV() bool {
	return r.is(REG_V)
}

func (r *Register) IsZ() bool {
	return r.is(REG_Z)
}

func (r *Register) IsP() bool {
	return r.is(REG_P)
}

// The register number 31 is the zero register or the stack register when used as an operand,
// depending on the instruction context.
func (r *Register) IsSPOrZR() bool {
	return r.IsR() && r.Number() == 31
}

// Returns the string representation of the register name, not including the extension.
func (r *Register) Name() string {
	switch {
	case r.IsR():
		return "R" + strconv.Itoa(int(r.Number()))
	case r.IsF():
		return "F" + strconv.Itoa(int(r.Number()))
	case r.IsV():
		return "V" + strconv.Itoa(int(r.Number()))
	case r.IsZ():
		return "Z" + strconv.Itoa(int(r.Number()))
	case r.IsP():
		return "P" + strconv.Itoa(int(r.Number()))
	default:
		panic("unreachable ARM64 register")
	}
}

// Returns the string representation of the register extension, including any separator punctuation.
func (r *Register) Extension() string {
	switch r.Ext() {
	case EXT_NONE:
		return ""
	case EXT_B:
		return ".B"
	case EXT_H:
		return ".H"
	case EXT_S:
		return ".S"
	case EXT_D:
		return ".D"
	case EXT_Q:
		return ".Q"
	case EXT_MERGING:
		return ".M"
	case EXT_ZEROING:
		return ".Z"
	default:
		panic("unreachable ARM64 extension")
	}
}

func (r *Register) String() string {
	return r.Name() + r.Extension()
}

func (r *Register) Group() int {
	grp := r.uint16 & groupMask
	if grp < REG_R || grp > REG_P_INDEXED {
		panic(fmt.Sprintf("unreachable ARM64 register group: %x", r.uint16))
	}
	return int(grp)
}

func (r *Register) Format() int {
	return r.Group() | r.Ext()
}

func (r *Register) BankSize() int16 {
	switch {
	case r.IsR(), r.IsF(), r.IsV(), r.IsZ():
		return 32
	case r.IsP():
		return 16
	default:
		panic("unreachable ARM64 register group")
	}
}

func (r *Register) ToInt16() int16 {
	return int16(r.uint16)
}

func (r *Register) HasLaneSize() bool {
	return EXT_B <= r.Ext() && r.Ext() <= EXT_D
}
