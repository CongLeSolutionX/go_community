// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

const (
	RL_COUNT1 = (1 + iota) << 16
	RL_COUNT2
	RL_COUNT3
	RL_COUNT4
)

const (
	RL_INC0 = iota << 24
	RL_INC1
)

/* Register lists are stored as a 64-bit integer for ARM64.
 * This is a compact representation of a sequence of registers with
 * length 'count', with register numbers in the sequence increasing
 * by 'increment' from some 'base' register.
 *
 * +----------+------------+--------+--------+
 * |unused(32)|increment(8)|count(8)|base(16)|
 * +----------+------------+--------+--------+
 * increment: int8,
 * count: int8
 * base: int16
 */
type ARM64RegisterList struct {
	uint64
}

func NewARM64RegisterList(base int16, count int8, increment int8) ARM64RegisterList {
	return ARM64RegisterList{
		uint64(base) | (uint64(count)&7)<<16 | (uint64(increment)&3)<<24,
	}
}

func (r ARM64RegisterList) Base() int16 {
	return int16(r.uint64 & ((1 << 16) - 1))
}

func (r ARM64RegisterList) Count() int8 {
	return int8((r.uint64 >> 16) & ((1 << 8) - 1))
}

func (r ARM64RegisterList) Increment() int8 {
	return int8((r.uint64 >> 24) & ((1 << 8) - 1))
}

func (r ARM64RegisterList) GetRegisterAtIndex(index int) int16 {
	if index < 0 {
		panic("negative register list index")
	}

	var reg int16
	if IsSVECompatibleRegister(r.Base()) {
		next := AsRegister(r.Base())
		newNum := (next.Number() + int16(index*int(r.Increment()))) % next.BankSize()
		next.SetNumber(newNum)
		reg = next.ToInt16()
	} else {
		// New register number is old + index * increment
		num := r.Base() & 31
		newNum := (num + int16(index*int(r.Increment()))) % 32

		// Clears the old register number, replaces with the new register number
		reg = r.Base() ^ (r.Base() & 31) ^ (newNum & 31)
	}

	return reg
}

func (r ARM64RegisterList) ToInt64() int64 {
	return int64(r.uint64)
}

func (r ARM64RegisterList) Ext() int {
	if IsSVECompatibleRegister(r.Base()) {
		br := AsRegister(r.Base())
		return br.Ext()
	} else {
		panic("not yet implemented")
	}
}

func (r ARM64RegisterList) Format() int {
	ty := 0
	if IsSVECompatibleRegister(r.Base()) {
		baseReg := AsRegister(r.Base())
		switch baseReg.Group() {
		case REG_R:
			ty = REGLIST_R
		case REG_F:
			ty = REGLIST_F
		case REG_V:
			ty = REGLIST_V
		case REG_Z:
			ty = REGLIST_Z
		case REG_P:
			ty = REGLIST_P
		default:
			panic("unreachable")
		}
	} else {
		switch {
		case REG_R0 >= r.Base() && r.Base() <= REG_R31, r.Base() == REG_RSP:
			ty = REGLIST_R
		case REG_F0 >= r.Base() && r.Base() <= REG_F31:
			ty = REGLIST_F
		case REG_V0 >= r.Base() && r.Base() <= REG_V31:
			ty = REGLIST_V
		case REG_Z0 >= r.Base() && r.Base() <= REG_Z31:
			ty = REGLIST_Z
		case REG_P0 >= r.Base() && r.Base() <= REG_P15:
			ty = REGLIST_P
		default:
			panic("unreachable")
		}
	}
	return ty | r.Ext() | int(r.Count())<<16 | int(r.Increment())<<24
}
