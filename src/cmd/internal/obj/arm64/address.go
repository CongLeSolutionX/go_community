// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
)

const (
	MEM_BASE = iota << subtypeOffset
	MEM_OFFSET_REG
	MEM_OFFSET_IMM
)

type Address struct {
	*obj.Addr
}

func NewAddress(addr *obj.Addr, base SVERegister, ext int) Address {
	a := Address{addr}
	a.Type = obj.TYPE_MEM
	a.SetBase(base)
	return a
}

func IsAddrAddress(addr *obj.Addr) bool {
	return addr.Type == obj.TYPE_ADDR
}

func AsAddress(addr *obj.Addr) Address {
	if !IsAddrAddress(addr) {
		panic("Addr is not compatible with this encoding scheme")
	}
	return Address{addr}
}

func (addr *Address) SetBase(r SVERegister) {
	addr.Reg = r.ToInt16()
}

func (addr *Address) SetOffsetReg(r SVERegister) {
	addr.Index = r.ToInt16()
}

func (addr *Address) SetOffsetImm(offs int64) {
	addr.Offset = offs
}

func (addr *Address) Subtype() int {
	if addr.Offset != 0 {
		return MEM_OFFSET_IMM
	} else if addr.Index != 0 {
		return MEM_OFFSET_REG
	}
	return MEM_BASE
}

func (addr *Address) Format() int {
	return MEM_ADDR | addr.Subtype()
}

func (addr *Address) String() string {
	r := AsSVERegister(addr.Reg)
	switch addr.Subtype() {
	case MEM_BASE:
		return fmt.Sprintf("(%s)", r.String())
	case MEM_OFFSET_IMM:
		return fmt.Sprintf("$%d(%s)", addr.Val.(int64), r.String())
	case MEM_OFFSET_REG:
		return fmt.Sprintf("(%s)(%s)", addr.Val.(*SVERegister).String(), r.String())
	default:
		panic("unsupported ARM64 address type")
	}
}
