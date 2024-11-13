// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
	"math"
)

// Validates input parameters to the packing and extracting functions.
func checkBitRange(hibit int, lobit int, max int) {
	if hibit < lobit {
		panic("need hibit >= lobit")
	}
	if lobit < 0 {
		panic("need lobit >= 0")
	}
	if hibit >= max {
		panic(fmt.Sprintf("need hibit < %d", max))
	}
}

// Extract range of bits from int64, interpret as signed integer
func ex(val int64, hibit int, lobit int) int64 {
	checkBitRange(hibit, lobit, 64)
	var mask int64 = (1 << (hibit - lobit + 1)) - 1
	return (val >> lobit) & mask
}

// Pack unsigned integer into uint32 within range [hibit,lobit]
func pu(val uint32, hibit int, lobit int) (uint32, bool) {
	checkBitRange(hibit, lobit, 32)
	var top uint32 = 1 << (hibit - lobit + 1)
	if val >= top {
		Debug("val '%d' too large for %d bit field", val, hibit-lobit+1)
		return 0, false
	}
	return uint32((val & (top - 1)) << lobit), true
}

// Pack signed integer into uint32 within range [hibit,lobit]
func ps(val int32, hibit int, lobit int) (uint32, bool) {
	checkBitRange(hibit, lobit, 32)
	masked := uint32(val & ((1 << (hibit - lobit + 1)) - 1))
	return pu(masked, hibit, lobit)
}

/* Registers */

// Destination register is typically within bits 0-4
func Rd(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 4, 0)
}

// First source register is typically within bits 5-9
func Rn(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 9, 5)
}

// Second source register is typically with bits 16-20
func Rm(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 20, 16)
}

var Rt = Rd
var Rn_20_16 = Rm

// Destination predicate register
func Pd(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 3, 0)
}

// First source predicate register
func Pn(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 8, 5)
}

// Restricted range governing predicate register
func Pg_12_10(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 12, 10)
}

// Governing predicate register
func Pg(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 13, 10)
}

// Second source predicate register
func Pm(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	return pu(uint32(r.Number()), 19, 16)
}

var Pg_8_5 = Pm_8_5
var Pg_19_16 = Pm
var Pm_8_5 = Pn
var Pv = Pn
var Pv_13_10 = Pg
var Pv_12_10 = Pg_12_10

// Destructive source/destination register pair, encoded
// as single destination register.
func Rdn(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	if r1.Number() != r2.Number() {
		return 0, false
	}
	return pu(uint32(r1.Number()), 4, 0)
}

// Destructive source/destination register pair, encoded
// as single destination register.
func Pdn(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	if r1.Number() != r2.Number() {
		return 0, false
	}
	return pu(uint32(r1.Number()), 3, 0)
}

var Zdn = Rdn
var Pdm = Pdn

/* Registers with index */

// 1-bit index shares field with Rm.
func Rmi1(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	v, okv := pu(uint32(r.Number()), 19, 16)
	u, oku := pu(uint32(vals[0].Index), 20, 20)
	if !okv || !oku {
		return 0, false
	}
	return v | u, true
}

// 2-bit index shares field with Rm.
func Rmi2(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	v, okv := pu(uint32(r.Number()), 18, 16)
	u, oku := pu(uint32(vals[0].Index), 20, 19)
	if !okv || !oku {
		return 0, false
	}
	return v | u, true
}

// 3-bit index shares field with Rm and uses another bit lower
// in the opcode.
func Rmi3(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	u, oku := pu(uint32(r.Number()), 18, 16)
	v, okv := pu(uint32(vals[0].Index>>1&0b11), 20, 19)
	w, okt := pu(uint32(vals[0].Index&0b1), 11, 11)
	return u | v | w, oku && okv && okt
}

// 3-bit index shares field with Rm and uses another bit higher in
// the opcode.
func Rmi3h_i3l(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	if vals[0].Index > 7 || vals[0].Index < 0 {
		return 0, false
	}
	u, oku := pu(uint32(r.Number()), 18, 16)
	v, okv := pu(uint32(vals[0].Index&0b11), 20, 19)
	w, okt := pu(uint32(vals[0].Index>>2), 22, 22)
	return u | v | w, oku && okv && okt
}

/* Register lists */

// Typical list of transfer registers for loads/stores.
func RLt(vals ...*obj.Addr) (uint32, bool) {
	rl := ARM64RegisterList{uint64(vals[0].Offset)}
	num := uint32(0)
	if IsSVECompatibleRegister(rl.Base()) {
		r := AsRegister(rl.Base())
		num = uint32(r.Number())
	} else {
		num = uint32(rl.Base() & 31)
	}
	return pu(num, 4, 0)
}

// Typical list of source registers.
func RLn(vals ...*obj.Addr) (uint32, bool) {
	rl := ARM64RegisterList{uint64(vals[0].Offset)}
	num := uint32(0)
	if IsSVECompatibleRegister(rl.Base()) {
		r := AsRegister(rl.Base())
		num = uint32(r.Number())
	} else {
		num = uint32(rl.Base() & 31)
	}
	return pu(num, 9, 5)
}

/* Lane Arrangements/Sizes */

func _sveT(vals ...*obj.Addr) (uint32, bool) {
	baseExt := 0
	switch vals[0].Type {
	case obj.TYPE_REG:
		r := AsRegister(vals[0].Reg)
		if !r.HasLaneSize() {
			return 0, false
		}
		baseExt = r.Ext()
	case obj.TYPE_REGLIST:
		rl := ARM64RegisterList{uint64(vals[0].Offset)}
		baseExt = rl.Ext()
	default:
		return 0, false
	}

	for _, v := range vals[1:] {
		switch v.Type {
		case obj.TYPE_REG:
			r2 := AsRegister(v.Reg)
			if baseExt != r2.Ext() {
				return 0, false
			}
		case obj.TYPE_REGLIST:
			rl := ARM64RegisterList{uint64(v.Offset)}
			if baseExt != rl.Ext() {
				return 0, false
			}
		default:
			return 0, false
		}
	}

	t := uint32(0)
	switch baseExt {
	case EXT_B:
		t = 0
	case EXT_H:
		t = 1
	case EXT_S:
		t = 2
	case EXT_D:
		t = 3
	default:
		panic("unreachable")
	}
	return t, true
}

// Standard SVE lane size (B=0,H=1,S=2,D=3)
func sveT(vals ...*obj.Addr) (uint32, bool) {
	t, oku := _sveT(vals...)
	size, okv := pu(t, 23, 22)
	return size, oku && okv
}

// SVE lane size shifted down 1 bit.
func sveT_22_21(vals ...*obj.Addr) (uint32, bool) {
	t, oku := _sveT(vals...)
	size, okv := pu(t, 22, 21)
	return size, oku && okv
}

// Single bit lane size, for memory addresses.
func svesz(vals ...*obj.Addr) (uint32, bool) {
	r := AsRegister(vals[0].Reg)
	switch r.Ext() {
	case EXT_S:
		return 0, true
	case EXT_D:
		return 1 << 22, true
	default:
		return 0, false
	}
}

/* Immediate values */

// Validator that a floating point constant is equal to 0.
func ImmFP0(vals ...*obj.Addr) (uint32, bool) {
	v := vals[0].Val.(float64)
	if v == 0.0 && !math.Signbit(v) {
		return 0, true
	}
	return 0, false
}

func Imm5(vals ...*obj.Addr) (uint32, bool) {
	return ps(int32(vals[0].Offset), 20, 16)
}

func Uimm7(vals ...*obj.Addr) (uint32, bool) {
	imm7 := vals[0].Offset
	if imm7 < 0 {
		return 0, false
	}
	return pu(uint32(imm7), 20, 14)
}

// vals[0] => TYPE_CONST
func Uimm4(vals ...*obj.Addr) (uint32, bool) {
	return pu(uint32(vals[0].Offset-1), 19, 16)
}

// vals[0] => TYPE_CONST
func Imm6(vals ...*obj.Addr) (uint32, bool) {
	return ps(int32(vals[0].Offset), 10, 5)
}

// Logical immediate encoding with vector lane size.
// Vector lane size is implied by the logical immediate encoding pattern.
func sveBitmask(vals ...*obj.Addr) (uint32, bool) {
	reg := AsRegister(vals[0].Reg)
	imm := uint64(vals[1].Offset)

	switch reg.Ext() {
	case EXT_B:
		if imm >= (1 << 8) {
			return 0, false
		}
		imm = (imm << 8) | imm
		imm = (imm << 16) | imm
		imm = (imm << 32) | imm
	case EXT_H:
		if imm >= (1 << 16) {
			return 0, false
		}
		imm = (imm << 16) | imm
		imm = (imm << 32) | imm
	case EXT_S:
		if imm >= (1 << 32) {
			return 0, false
		}
		imm = (imm << 32) | imm
	case EXT_D:
		// Mask is always 64-bit, nothing to do
	default:
		// Invalid extension
		panic("unexpected register extension")
	}

	if !isbitcon(imm) {
		return 0, false
	}

	imm13 := bitconEncode(imm, 64) >> 10
	u, oku := pu(uint32(reg.Number()), 4, 0)
	v, okv := pu(imm13, 17, 5)
	return u | v, oku || okv
}

func imm8(vals ...*obj.Addr) (uint32, bool) {
	return ps(int32(vals[0].Offset), 12, 5)
}

func imm8_FP(vals ...*obj.Addr) (uint32, bool) {
	v, okv := encodeFloat8(vals[0].Val.(float64))
	if !okv {
		return 0, false
	}
	return pu(v, 17, 5)
}

func imm3(vals ...*obj.Addr) (uint32, bool) {
	return pu(uint32(vals[0].Offset), 18, 16)
}

func imm5(vals ...*obj.Addr) (uint32, bool) {
	return ps(int32(vals[0].Offset), 9, 5)
}

func imm5b(vals ...*obj.Addr) (uint32, bool) {
	return ps(int32(vals[0].Offset), 20, 16)
}

func _rot(val int64) (uint32, bool) {
	rot := uint32(0)
	switch val {
	case 0:
		rot = 0
	case 90:
		rot = 1
	case 180:
		rot = 2
	case 270:
		rot = 3
	default:
		return 0, false
	}
	return rot, true
}

func rot(vals ...*obj.Addr) (uint32, bool) {
	rot, ok := _rot(vals[0].Offset)
	return rot << 13, ok
}

func rot_11_10(vals ...*obj.Addr) (uint32, bool) {
	rot, ok := _rot(vals[0].Offset)
	return rot << 10, ok
}

func rot_16_16(vals ...*obj.Addr) (uint32, bool) {
	switch vals[0].Offset {
	case 90:
		return 0, true
	case 270:
		return 1 << 16, true
	case 0, 180:
		// Unsupported rotation for this encoding
		return 0, false
	default:
		return 0, false
	}
}

func _i1_FP(imm float64, zero, one float64) (uint32, bool) {
	switch imm {
	case one:
		return 1, true
	case zero:
		return 0, true
	default:
		return 0, false
	}
}

func i1_FP_0p0_1p0(vals ...*obj.Addr) (uint32, bool) {
	i1, ok := _i1_FP(vals[0].Val.(float64), 0.0, 1.0)
	return i1 << 5, ok
}

func i1_FP_0p5_1p0(vals ...*obj.Addr) (uint32, bool) {
	i1, ok := _i1_FP(vals[0].Val.(float64), 0.5, 1.0)
	return i1 << 5, ok
}

func i1_FP_0p5_2p0(vals ...*obj.Addr) (uint32, bool) {
	i1, ok := _i1_FP(vals[0].Val.(float64), 0.5, 2.0)
	return i1 << 5, ok
}

func imm8h_imm8l(vals ...*obj.Addr) (uint32, bool) {
	imm8 := vals[0].Offset
	if imm8 > 255 || imm8 < 0 {
		return 0, false
	}
	imm8h, oku := pu(uint32(imm8>>3), 20, 16)
	imm8l, okv := pu(uint32(imm8&0b111), 12, 10)

	return imm8h | imm8l, oku || okv
}

// Shift immediate literal, either #0 or #8
func sh(vals ...*obj.Addr) (uint32, bool) {
	switch vals[0].Offset {
	case 8:
		return (1 << 13), true
	case 0:
		return 0, true
	default:
		return 0, false
	}
}

func _tszh_tszl_imm3(r1, r2 Register, cons int64, bias bool) (tszh, tszl, imm3 uint32, ok bool) {
	if r1.Ext() != r2.Ext() {
		return 0, 0, 0, false
	}

	esize := 0
	min := 0
	max := 0
	switch r1.Ext() {
	case EXT_B:
		esize = 8
	case EXT_H:
		esize = 16
	case EXT_S:
		esize = 32
	case EXT_D:
		esize = 64
	default:
		// Invalid extension
		return 0, 0, 0, false
	}
	max = esize - 1

	if bias {
		min += 1
		max += 1
	}

	tsize_imm3 := uint32(0)
	if bias {
		tsize_imm3 = uint32(2*esize) - uint32(cons)
	} else {
		tsize_imm3 = uint32(esize) + uint32(cons)
	}

	tszh = tsize_imm3 >> 5
	tszl = (tsize_imm3 >> 3) & 3
	imm3 = tsize_imm3 & 7

	return tszh, tszl, imm3, true
}

// Dependent shift immediate and lane size encoded together.
func tszh_tszl_imm3(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	cons := vals[2].Offset
	tszh, tszl, imm3, ok := _tszh_tszl_imm3(r1, r2, cons, false)
	return tszh<<22 | tszl<<19 | imm3<<16, ok
}

// Dependent shift immediate and lane size encoded together with bias from 1.
func tszh_tszl_imm3_bias1(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	cons := vals[2].Offset
	tszh, tszl, imm3, ok := _tszh_tszl_imm3(r1, r2, cons, true)
	return tszh<<22 | tszl<<19 | imm3<<16, ok
}

func tszh_tszl_imm3_22_8_5(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	cons := vals[2].Offset
	tszh, tszl, imm3, ok := _tszh_tszl_imm3(r1, r2, cons, false)
	return tszh<<22 | tszl<<8 | imm3<<5, ok
}

func tszh_tszl_imm3_bias1_22_8_5(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	cons := vals[2].Offset
	tszh, tszl, imm3, ok := _tszh_tszl_imm3(r1, r2, cons, true)
	return tszh<<22 | tszl<<8 | imm3<<5, ok
}

// Indexing scheme for dup instruction with variable length immediate.
// See DUP (indexed) in instruction reference.
func dup_indexed(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsRegister(vals[0].Reg)
	r2 := AsRegister(vals[1].Reg)
	index := vals[1].Index

	if r1.Ext() != r2.Ext() {
		return 0, false
	}

	pattern := uint32(0)
	switch r1.Ext() {
	case EXT_B:
		if index < 0 || index >= 512/8 {
			return 0, false
		}
		pattern = 0b00001 | uint32(index<<1)
	case EXT_H:
		if index < 0 || index >= 512/16 {
			return 0, false
		}
		pattern = 0b00010 | uint32(index<<2)
	case EXT_S:
		if index < 0 || index >= 512/32 {
			return 0, false
		}
		pattern = 0b00100 | uint32(index<<3)
	case EXT_D:
		if index < 0 || index >= 512/64 {
			return 0, false
		}
		pattern = 0b01000 | uint32(index<<4)
	case EXT_Q:
		if index < 0 || index >= 512/128 {
			return 0, false
		}
		pattern = 0b10000 | uint32(index<<5)
	default:
		return 0, false
	}

	imm2 := pattern >> 5
	tsz := pattern & 31

	zd, okv := Rd(vals[0])
	zn, oku := Rn(vals[1])
	return imm2<<22 | tsz<<16 | zn | zd, okv && oku
}

/* Symbolic operands */

func svePrfop(vals ...*obj.Addr) (uint32, bool) {
	arg := vals[0]
	if arg.Type == obj.TYPE_SPECIAL {
		specialOp := arg.Offset

		if specialOp >= int64(SPOP_PLDL1KEEP) && specialOp <= int64(SPOP_PLDL3STRM) {
			return uint32(specialOp), true
		} else if specialOp >= int64(SPOP_PSTL1KEEP) && specialOp <= int64(SPOP_PSTL3STRM) {
			return uint32(specialOp) - uint32(SPOP_PSTL1KEEP) + 8, true
		}
	} else if arg.Type == obj.TYPE_CONST {
		if (arg.Offset & 0b0110) == 0b0110 {
			return uint32(arg.Offset & 15), true
		}
	}
	Debug("Invalid SVE prefetch operand: %v", arg)
	return 0, false
}

// vals[0] => TYPE_SPECIAL, SPOP_POW2 - SPOP_ALL | TYPE_CONST
func pattern(vals ...*obj.Addr) (uint32, bool) {
	switch vals[0].Type {
	case obj.TYPE_SPECIAL:
		enc := map[SpecialOperand]uint32{
			SPOP_POW2:  0b00000,
			SPOP_VL1:   0b00001,
			SPOP_VL2:   0b00010,
			SPOP_VL3:   0b00011,
			SPOP_VL4:   0b00100,
			SPOP_VL5:   0b00101,
			SPOP_VL6:   0b00110,
			SPOP_VL7:   0b00111,
			SPOP_VL8:   0b01000,
			SPOP_VL16:  0b01001,
			SPOP_VL32:  0b01010,
			SPOP_VL64:  0b01011,
			SPOP_VL128: 0b01100,
			SPOP_VL256: 0b01101,
			SPOP_MUL4:  0b11101,
			SPOP_MUL3:  0b11110,
			SPOP_ALL:   0b11111,
		}

		v, ok := enc[SpecialOperand(vals[0].Offset)]
		if ok {
			return pu(v, 9, 5)
		}
	case obj.TYPE_CONST:
		immPatterns := []uint32{
			0b01110,
			0b10101,
			0b10110,
			0b10001,
			0b10010,
			0b10000,
		}

		for _, pat := range immPatterns {
			if vals[0].Offset < 32 && (pat&uint32(vals[0].Offset)) == pat {
				return pu(uint32(vals[0].Offset), 9, 5)
			}
		}
	}
	return 0, false
}

/* Memory addressing modes */

func RnImm9MulVl(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	imm9l := ex(addr.Offset, 2, 0)
	imm9h := ex(addr.Offset, 8, 3)
	rn, okw := pu(uint32(addr.Reg&31), 9, 5)
	imml, okx := pu(uint32(imm9l), 12, 10)
	immh, oky := pu(uint32(imm9h), 21, 16)
	return immh | imml | rn, okw && okx && oky
}

// vals[0] => [<Xn|SP>]
func MemXnSP(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	return pu(uint32(addr.Reg&31), 9, 5)
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	imm6, okv := ps(int32(addr.Offset), 21, 16)
	return rn | imm6, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset_19_16(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	imm4, okv := ps(int32(addr.Offset), 19, 16)
	return rn | imm4, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset16_19_16(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%16 != 0 {
		return 0, false
	}
	imm4, okv := ps(int32(addr.Offset/16), 19, 16)
	return rn | imm4, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset32_19_16(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%32 != 0 {
		return 0, false
	}
	imm4, okv := ps(int32(addr.Offset/32), 19, 16)
	return rn | imm4, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset2(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%2 != 0 {
		return 0, false
	}
	imm4, okv := ps(int32(addr.Offset/2), 19, 16)
	return rn | imm4, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset3(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%3 != 0 {
		return 0, false
	}
	imm4, okv := ps(int32(addr.Offset/3), 19, 16)
	return rn | imm4, oku && okv
}

// vals[0] => [<Xn|SP>, #<imm>]
func MemXnSPOffset4(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%4 != 0 {
		return 0, false
	}
	imm4, okv := ps(int32(addr.Offset/4), 19, 16)
	return rn | imm4, oku && okv
}

var MemXnSPimm6 = MemXnSPOffset

func MemXnSP2imm6(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%2 != 0 {
		return 0, false
	}
	imm6, okv := pu(uint32(addr.Offset/2), 21, 16)
	return rn | imm6, oku && okv
}

func MemXnSP4imm6(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%4 != 0 {
		return 0, false
	}
	imm6, okv := pu(uint32(addr.Offset/4), 21, 16)
	return rn | imm6, oku && okv
}

func MemXnSP8imm6(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, oku := MemXnSP(vals[0])
	if addr.Offset%8 != 0 {
		return 0, false
	}
	imm6, okv := pu(uint32(addr.Offset/8), 21, 16)
	return rn | imm6, oku && okv
}

// vals[0] => [<Xn|SP>, <Xm>]
func MemXnSPXmOffset(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	rn, okv := pu(uint32(addr.Reg&31), 9, 5)
	rm, oku := pu(uint32(addr.Index&31), 20, 16)
	return rm | rn, oku && okv
}

// vals[0] => [<Xn|SP>, <Zm>]
func MemXnSPZmOffset(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	index := AsRegister(addr.Index)
	rn, okv := pu(uint32(addr.Reg&31), 9, 5)
	rm, oku := pu(uint32(index.Number()), 20, 16)
	return rm | rn, oku && okv
}

// vals[0] => [<Zn>]
func MemZn(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	base := AsRegister(addr.Reg)
	return pu(uint32(base.Number()), 9, 5)
}

// vals[0] => [<Zn>, #<imm>]
func MemZnOffset(vals ...*obj.Addr) (uint32, bool) {
	rn, oku := MemZn(vals[0])
	idx := vals[0].Offset
	imm5, okv := pu(uint32(idx), 20, 16)
	return rn | imm5, oku && okv
}

// vals[0] => [<Zn>, #<imm>]
func MemZnOffset2(vals ...*obj.Addr) (uint32, bool) {
	rn, oku := MemZn(vals[0])
	idx := vals[0].Offset
	if idx%2 != 0 {
		return 0, false
	}
	imm5, okv := pu(uint32(idx/2), 20, 16)
	return rn | imm5, oku && okv
}

// vals[0] => [<Zn>, #<imm>]
func MemZnOffset4(vals ...*obj.Addr) (uint32, bool) {
	rn, oku := MemZn(vals[0])
	idx := vals[0].Offset
	if idx%4 != 0 {
		return 0, false
	}
	imm5, okv := pu(uint32(idx/4), 20, 16)
	return rn | imm5, oku && okv
}

// vals[0] => [<Zn>, #<imm>]
func MemZnOffset8(vals ...*obj.Addr) (uint32, bool) {
	rn, oku := MemZn(vals[0])
	idx := vals[0].Offset
	if idx%8 != 0 {
		return 0, false
	}
	imm5, okv := pu(uint32(idx/8), 20, 16)
	return rn | imm5, oku && okv
}

func _xs(addr Address) (uint32, bool) {
	xs := uint32(0)
	switch addr.IndexMod() {
	case MOD_SXTW:
		xs = 1
	case MOD_UXTW:
		xs = 0
	default:
		return 0, false
	}
	return xs, true
}

// vals[0] => [<Xn|SP>, <Zm>.<T>, <mod>{ #n}]
func MemXnSPZmDmod(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	zm_rn, oku := MemXnSPZmOffset(vals[0])
	xs, okv := _xs(addr)
	return zm_rn | xs<<14, oku && okv
}

// vals[0] => [<Xn|SP>, <Zm>.<T>, <mod>{ #n}]
func MemXnSPZmDmod_22(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	zm_rn, oku := MemXnSPZmOffset(vals[0])
	xs, okv := _xs(addr)
	return zm_rn | xs<<22, oku && okv
}

func Zm_msz_Zn(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	if !IsSVECompatibleRegister(addr.Index) {
		return 0, false
	}
	base := AsRegister(addr.Reg)
	index := AsRegister(addr.Index)

	msz, oku := pu(uint32(addr.IndexScale()), 11, 10)
	zn, okv := pu(uint32(base.Number()), 9, 5)
	zm, okw := pu(uint32(index.Number()), 20, 16)

	return zm | msz | zn, oku && okv && okw
}
