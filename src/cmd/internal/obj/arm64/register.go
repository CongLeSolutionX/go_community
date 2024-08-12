// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"log"
	"strconv"
)

// The SVERegister type wraps an obj.Addr reference and can be used to read
// and modify register properties, including register number and extension.
type SVERegister struct {
	uint16
}

// SVE registers are represented as shown in the diagram.
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
const typeMask = 0b111

// Types of register extension, add as needed.
const (
	EXT_NONE    = iota << extOffset
	EXT_B       // .B
	EXT_H       // .H
	EXT_S       // .S
	EXT_D       // .D
	EXT_MERGING // .M
	EXT_ZEROING // .Z
	EXT_END     // End marker
)

const (
	REG_R = iota
	REG_F
	REG_V
	REG_Z
	REG_P
	REG_V_INDEXED
	REG_Z_INDEXED
	REG_P_INDEXED
)

// Bit extractors for Type and Extension from Register.Format()
func getType(fmt int) int {
	return int(uint(fmt) & typeMask)
}

func getExt(fmt int) int {
	return int(uint(fmt) & extMask)
}

// Checks whether the Addr structure represents a register in this scheme.
func IsSVERegister(data int16) bool {
	return uint16(data)&REG_SVE != 0
}

func NewSVERegister(number int16, etype int16) SVERegister {
	r := SVERegister{uint16(number&31)<<numOffset | (uint16(number/32)-1)&7 | REG_SVE}
	r.SetExt(etype)
	return r
}

func AsSVERegister(data int16) SVERegister {
	if !IsSVERegister(data) {
		panic("Addr is not compatible with this register scheme")
	}
	r := SVERegister{uint16(data)}
	if r.Ext() >= EXT_END || r.Ext() < EXT_NONE {
		log.Panicf("Addr has invalid extension: %b", r.Ext())
	}
	return r
}

func (r *SVERegister) Ext() int16 {
	return int16(r.uint16 & extMask)
}

func (r *SVERegister) SetExt(ext int16) {
	r.uint16 ^= (r.uint16 & extMask) ^ uint16(ext)
}

// Get the register number, between 0-31. For P registers between 0-15.
func (r *SVERegister) Number() int16 {
	return int16((r.uint16 & numMask) >> numOffset)
}

func (r *SVERegister) is(rtype uint16) bool {
	return r.uint16&typeMask == rtype
}

func (r *SVERegister) IsR() bool {
	return r.is(REG_R)
}

func (r *SVERegister) IsF() bool {
	return r.is(REG_F)
}

func (r *SVERegister) IsV() bool {
	return r.is(REG_V)
}

func (r *SVERegister) IsZ() bool {
	return r.is(REG_Z)
}

func (r *SVERegister) IsP() bool {
	return r.is(REG_P)
}

// The register number 31 is the zero register or the stack register when used as an operand,
// depending on the instruction context.
func (r *SVERegister) IsSPOrZR() bool {
	return r.IsR() && r.Number() == 31
}

// Returns the string representation of the register name, not including the extension.
func (r *SVERegister) Name() string {
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
func (r *SVERegister) Extension() string {
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
	case EXT_MERGING:
		return ".M"
	case EXT_ZEROING:
		return ".Z"
	default:
		panic("unreachable ARM64 extension")
	}
}

func (r *SVERegister) String() string {
	return r.Name() + r.Extension()
}

func (r *SVERegister) Group() int {
	switch {
	case r.IsR():
		return REG_R
	case r.IsF():
		return REG_F
	case r.IsV():
		return REG_V
	case r.IsZ():
		return REG_Z
	case r.IsP():
		return REG_P
	default:
		panic("unreachable ARM64 register group")
	}
}

func (r *SVERegister) Format() int {
	return r.Group() | int(r.Ext())
}

func (r *SVERegister) ToInt16() int16 {
	return int16(r.uint16)
}

func (r *SVERegister) HasLaneSize() bool {
	return EXT_B <= r.Ext() && r.Ext() <= EXT_D
}
