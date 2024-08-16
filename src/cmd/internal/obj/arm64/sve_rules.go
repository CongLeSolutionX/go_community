// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
)

var Rt = Rd

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

func b(val int64, hibit int, lobit int) (int64, bool) {
	checkBitRange(hibit, lobit, 64)
	var mask int64 = (1 << (hibit - lobit + 1)) - 1
	return (val >> lobit) & mask, true
}

func p(val uint32, hibit int, lobit int) (uint32, bool) {
	checkBitRange(hibit, lobit, 32)
	var top uint32 = 1 << (hibit - lobit + 1)
	if val >= top {
		panic(fmt.Sprintf("val '%d' too large for %d bit field", val, hibit-lobit+1))
	}
	return uint32((val & (top - 1)) << lobit), true
}

func Rd(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	return p(uint32(r.Number()), 4, 0)
}

func Rn(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	return p(uint32(r.Number()), 9, 5)
}

func Rm(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	return p(uint32(r.Number()), 20, 16)
}

func Rmi2(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	v, okv := p(uint32(r.Number()), 18, 16)
	u, oku := p(uint32(vals[0].Index), 20, 19)
	if !okv || !oku {
		return 0, false
	}
	return v | u, true
}

func Pg(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	return p(uint32(r.Number()), 12, 10)
}

func Pm(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	return p(uint32(r.Number()), 8, 5)
}

func sveT(vals ...*obj.Addr) (uint32, bool) {
	r := AsSVERegister(vals[0].Reg)
	for _, v := range vals[1:] {
		r2 := AsSVERegister(v.Reg)
		if !r.HasLaneSize() || (r.Ext() != r2.Ext()) {
			return 0, false
		}
	}

	t := uint32(0)
	switch r.Ext() {
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
	return p(t, 23, 22)
}

func Zdn(vals ...*obj.Addr) (uint32, bool) {
	r1 := AsSVERegister(vals[0].Reg)
	r2 := AsSVERegister(vals[1].Reg)
	if r1.Number() != r2.Number() {
		return 0, false
	}
	return p(uint32(r1.Number()), 4, 0)
}

func RnImm9MulVl(vals ...*obj.Addr) (uint32, bool) {
	addr := AsAddress(vals[0])
	imm9l, oku := b(addr.Offset, 2, 0)
	imm9h, okv := b(addr.Offset, 8, 3)
	rn, okw := p(uint32(addr.Reg&31), 9, 5)
	imml, okx := p(uint32(imm9l), 12, 10)
	immh, oky := p(uint32(imm9h), 21, 16)
	return immh | imml | rn, oku && okv && okw && okx && oky
}
