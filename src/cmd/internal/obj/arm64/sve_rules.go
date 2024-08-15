// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
)

func p(val uint32, hibit int, lobit int) (uint32, bool) {
	if hibit < lobit {
		panic("need hibit >= lobit")
	}
	var top uint32 = 1 << (hibit - lobit + 1)
	if val >= top {
		panic(fmt.Sprintf("val '%x' too large for %d bit field", val, hibit-lobit))
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
