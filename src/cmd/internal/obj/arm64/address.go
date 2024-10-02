// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
)

const (
	MEM_INVALID = iota << subtypeOffset
	MEM_RSP
	MEM_ZS
	MEM_ZS_IMM
	MEM_ZS_ZS
	MEM_ZS_ZS_LSL
	MEM_ZD
	MEM_ZD_IMM
	MEM_ZD_ZD
	MEM_ZD_ZD_LSL
	MEM_ZD_ZD_SXTW
	MEM_ZD_ZD_UXTW
	MEM_RSP_R
	MEM_RSP_R_LSL1
	MEM_RSP_R_LSL2
	MEM_RSP_R_LSL3
	MEM_RSP_IMM
	MEM_RSP_ZD
	MEM_RSP_ZD_LSL1
	MEM_RSP_ZD_LSL2
	MEM_RSP_ZD_LSL3
	MEM_RSP_ZD_SXTW
	MEM_RSP_ZD_SXTW1
	MEM_RSP_ZD_SXTW2
	MEM_RSP_ZD_SXTW3
	MEM_RSP_ZD_UXTW
	MEM_RSP_ZD_UXTW1
	MEM_RSP_ZD_UXTW2
	MEM_RSP_ZD_UXTW3
	MEM_RSP_ZS
	MEM_RSP_ZS_SXTW
	MEM_RSP_ZS_SXTW1
	MEM_RSP_ZS_SXTW2
	MEM_RSP_ZS_SXTW3
	MEM_RSP_ZS_UXTW
	MEM_RSP_ZS_UXTW1
	MEM_RSP_ZS_UXTW2
	MEM_RSP_ZS_UXTW3
)

type Address struct {
	*obj.Addr
}

func NewAddress(addr *obj.Addr, base Register, ext int) Address {
	a := Address{addr}
	a.Type = obj.TYPE_MEM
	a.SetBase(base)
	return a
}

func IsAddrAddress(addr *obj.Addr) bool {
	return addr.Type == obj.TYPE_MEM
}

func AsAddress(addr *obj.Addr) Address {
	if !IsAddrAddress(addr) {
		panic("Addr is not compatible with this encoding scheme")
	}
	return Address{addr}
}

func (addr *Address) SetBase(r Register) {
	addr.Reg = r.ToInt16()
}

func (addr *Address) SetOffsetReg(r Register) {
	addr.Index = r.ToInt16()
}

func (addr *Address) SetOffsetImm(offs int64) {
	addr.Offset = offs
}

func (addr Address) IndexMod() int {
	return int(addr.Scale & (0b1111 << 4))
}

func (addr Address) IndexScale() int {
	return int(addr.Scale & ((1 << 4) - 1))
}

func (addr *Address) Subtype() int {
	var base Register
	if IsSVECompatibleRegister(addr.Reg) {
		base = AsRegister(addr.Reg)
	} else {
		base = NewRegister(addr.Reg, EXT_NONE)
	}
	switch {
	case addr.Offset == 0 && addr.Index == 0 && addr.Scale == 0:
		switch {
		case base.IsR():
			// [<R><n>]
			return MEM_RSP
		case base.IsZ():
			switch base.Ext() {
			case EXT_S:
				// [Z<n>.S]
				return MEM_ZS
			case EXT_D:
				// [Z<n>.D]
				return MEM_ZD
			}
		}
	case addr.Offset != 0 && addr.Index == 0 && addr.Scale == 0:
		switch {
		case base.IsR():
			// [<R><n>, #<imm>]
			return MEM_RSP_IMM
		case base.IsZ():
			switch base.Ext() {
			case EXT_S:
				// [Z<n>.S, #<imm>]
				return MEM_ZS_IMM
			case EXT_D:
				// [Z<n>.D, #<imm>]
				return MEM_ZD_IMM
			}
		}
	case addr.Index != 0:
		var index Register
		if IsSVECompatibleRegister(addr.Index) {
			index = AsRegister(addr.Index)
		} else {
			num := addr.Index
			// Fix for compatibility with the original REG_LSL form of shift registers.
			// Place the shift modifier into the addr.Scale field and process the register
			// as if it was a normal register constant.
			if addr.Index&REG_LSL != 0 {
				addr.Scale = MOD_LSL | (addr.Index>>5)&7
				// Remove the REG_LSL bit and the shift constant from the register
				// to get the register ID.
				num = addr.Index ^ REG_LSL ^ (addr.Index & (7 << 5))
			}
			index = NewRegister(num, EXT_NONE)
		}
		switch {
		case base.IsR() && index.IsR() && addr.Scale == 0:
			// [<R><n>, <R><m>]
			return MEM_RSP_R
		case base.IsR() && index.IsR() && (addr.Scale&MOD_LSL == MOD_LSL):
			switch addr.Scale & 0xf {
			case 1:
				// [<R><n>, <R><m>, LSL #1]
				return MEM_RSP_R_LSL1
			case 2:
				// [<R><n>, <R><m>, LSL #2]
				return MEM_RSP_R_LSL2
			case 3:
				// [<R><n>, <R><m>, LSL #3]
				return MEM_RSP_R_LSL3
			}
		case base.IsR() && index.IsZ() && addr.Scale == 0:
			switch index.Ext() {
			case EXT_S:
				// [<R><n>, Z<m>.S]
				return MEM_RSP_ZS
			case EXT_D:
				// [<R><n>, Z<m>.D]
				return MEM_RSP_ZD
			}
		case base.IsR() && index.IsZ() && (addr.Scale&MOD_LSL == MOD_LSL):
			switch index.Ext() {
			case EXT_D:
				switch addr.Scale & 0xf {
				case 1:
					// [<R><n>, Z<m>.D, LSL #1]
					return MEM_RSP_ZD_LSL1
				case 2:
					// [<R><n>, Z<m>.D, LSL #2]
					return MEM_RSP_ZD_LSL2
				case 3:
					// [<R><n>, Z<m>.D, LSL #3]
					return MEM_RSP_ZD_LSL3
				}
			}
		case base.IsR() && index.IsZ() && (addr.Scale&MOD_SXTW == MOD_SXTW):
			switch index.Ext() {
			case EXT_S:
				switch addr.Scale & 0xf {
				case 0:
					// [<R><n>, Z<m>.S, SXTW]
					return MEM_RSP_ZS_SXTW
				case 1:
					// [<R><n>, Z<m>.S, SXTW #1]
					return MEM_RSP_ZS_SXTW1
				case 2:
					// [<R><n>, Z<m>.S, SXTW #2]
					return MEM_RSP_ZS_SXTW2
				case 3:
					// [<R><n>, Z<m>.S, SXTW #3]
					return MEM_RSP_ZS_SXTW3
				}
			case EXT_D:
				switch addr.Scale & 0xf {
				case 0:
					// [<R><n>, Z<m>.D, SXTW]
					return MEM_RSP_ZD_SXTW
				case 1:
					// [<R><n>, Z<m>.D, SXTW #1]
					return MEM_RSP_ZD_SXTW1
				case 2:
					// [<R><n>, Z<m>.D, SXTW #2]
					return MEM_RSP_ZD_SXTW2
				case 3:
					// [<R><n>, Z<m>.D, SXTW #3]
					return MEM_RSP_ZD_SXTW3
				}
			}
		case base.IsR() && index.IsZ() && (addr.Scale&MOD_UXTW == MOD_UXTW):
			switch index.Ext() {
			case EXT_S:
				switch addr.Scale & 0xf {
				case 0:
					// [<R><n>, Z<m>.S, UXTW]
					return MEM_RSP_ZS_UXTW
				case 1:
					// [<R><n>, Z<m>.S, UXTW #1]
					return MEM_RSP_ZS_UXTW1
				case 2:
					// [<R><n>, Z<m>.S, UXTW #2]
					return MEM_RSP_ZS_UXTW2
				case 3:
					// [<R><n>, Z<m>.S, UXTW #3]
					return MEM_RSP_ZS_UXTW3
				}
			case EXT_D:
				switch addr.Scale & 0xf {
				case 0:
					// [<R><n>, Z<m>.D, UXTW]
					return MEM_RSP_ZD_UXTW
				case 1:
					// [<R><n>, Z<m>.D, UXTW #1]
					return MEM_RSP_ZD_UXTW1
				case 2:
					// [<R><n>, Z<m>.D, UXTW #2]
					return MEM_RSP_ZD_UXTW2
				case 3:
					// [<R><n>, Z<m>.D, UXTW #3]
					return MEM_RSP_ZD_UXTW3
				}
			}
		case base.IsZ() && index.IsZ() && addr.Scale == 0:
			switch base.Ext() {
			case EXT_S:
				// [Z<n>.S, Z<m>.S]
				return MEM_ZS_ZS
			case EXT_D:
				// [Z<n>.D, Z<m>.D]
				return MEM_ZD_ZD
			}
		case base.IsZ() && index.IsZ() && (addr.Scale&MOD_LSL == MOD_LSL):
			switch base.Ext() {
			case EXT_S:
				// [Z<n>.S, Z<m>.S, LSL #<imm>]
				return MEM_ZS_ZS_LSL
			case EXT_D:
				// [Z<n>.D, Z<m>.D, LSL #<imm>]
				return MEM_ZD_ZD_LSL
			}
		case base.IsZ() && index.IsZ() && (addr.Scale&MOD_SXTW == MOD_SXTW):
			switch base.Ext() {
			case EXT_D:
				// [Z<n>.D, Z<m>.D, SXTW{ #<imm>}]
				return MEM_ZD_ZD_SXTW
			}
		case base.IsZ() && index.IsZ() && (addr.Scale&MOD_UXTW == MOD_UXTW):
			switch base.Ext() {
			case EXT_D:
				// [Z<n>.D, Z<m>.D, UXTW{ #<imm>}]
				return MEM_ZD_ZD_UXTW
			}
		}
	}

	return MEM_INVALID
}

func (addr *Address) Format() int {
	return MEM_ADDR | addr.Subtype()
}
