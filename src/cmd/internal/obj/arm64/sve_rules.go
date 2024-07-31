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
