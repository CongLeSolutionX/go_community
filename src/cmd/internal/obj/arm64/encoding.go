// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// This file contains the encoding implementation of the argument type.

import (
	"cmd/internal/obj"
)

// isRegOrZR checks that REG_R0 <= r <= REG_R30 or r == REGZERO.
func (c *ctxt7) isRegOrZR(p *obj.Prog, r int16) {
	if !(r >= REG_R0 && r <= REG_R31) {
		c.ctxt.Diag("illegal register %s: %v", rconv(int(r)), p)
	}
}

// isRegOrRSP checks that REG_R0 <= r <= REG_R30 or r == REG_RSP.
func (c *ctxt7) isRegOrRSP(p *obj.Prog, r int16) {
	if !(r >= REG_R0 && r < REG_R31 || r == REG_RSP) {
		c.ctxt.Diag("illegal register %s: %v", rconv(int(r)), p)
	}
}

// isFReg checks that REG_F0 <= r <= REG_F31.
func (c *ctxt7) isFReg(p *obj.Prog, r int16) {
	if !(r >= REG_F0 && r <= REG_F31) {
		c.ctxt.Diag("illegal register %s: %v", rconv(int(r)), p)
	}
}

// isVReg checks that REG_V0 <= r <= REG_V31.
func (c *ctxt7) isVReg(p *obj.Prog, r int16) {
	if !(r >= REG_V0 && r <= REG_V31) {
		c.ctxt.Diag("illegal register %s: %v", rconv(int(r)), p)
	}
}

func (c *ctxt7) mem_optional_imm12_unsigned(p *obj.Prog, arg *arg, shift uint) uint32 {
	c.isRegOrRSP(p, arg.Reg)
	vs := arg.Offset >> shift
	if vs<<shift != arg.Offset {
		c.ctxt.Diag("odd offset: %d\n%v", arg.Offset, p)
	}
	if vs < 0 || vs >= (1<<12) {
		c.ctxt.Diag("offset out of range: %d\n%v", vs, p)
	}
	return uint32(vs&0xFFF)<<10 | uint32(arg.Reg&31)<<5
}

func (c *ctxt7) mem_imm7_signed(p *obj.Prog, arg *arg, shift uint) uint32 {
	c.isRegOrRSP(p, arg.Reg)
	vo := arg.Offset
	if vo < -64<<shift || vo > 63<<shift || vo&(1<<shift-1) != 0 {
		c.ctxt.Diag("invalid offset %v\n", p)
	}
	return uint32((vo>>shift)&0x7f)<<15 | uint32(arg.Reg&31)<<5
}

// arngTsMatchT checks if the arrangement specifier Ts matches T.
func arngTsMatchT(Ts, T int16) bool {
	switch T {
	case ARNG_8B, ARNG_16B:
		return Ts == ARNG_B
	case ARNG_4H, ARNG_8H:
		return Ts == ARNG_H
	case ARNG_2S, ARNG_4S:
		return Ts == ARNG_S
	case ARNG_2D:
		return Ts == ARNG_D
	}
	return false
}

// immRegSizeTMatch checks if the post-index immediate offset imm matches register size regSize
// and Q. This is for VLD1 and VST1 series instructions.
func immRegSizeQMatch(imm int64, regSize int64, Q int64) bool {
	return (regSize<<3)<<uint64(Q) == imm
}

// immhTaMatchTb checks if the arrangement specifier Ta matches Tb and Q when Tb is encoded in immh:Q,
// and return the maximum shift value allowed by different Ta.
func immhTaMatchTb(Ta, Tb, Q int16) (int, bool) {
	maxShift := 0
	match := false
	switch Ta {
	case ARNG_8H:
		maxShift = 7
		match = Tb == ARNG_8B && Q == 0 || Tb == ARNG_16B && Q == 1
	case ARNG_4S:
		maxShift = 15
		match = Tb == ARNG_4H && Q == 0 || Tb == ARNG_8H && Q == 1
	case ARNG_2D:
		maxShift = 31
		match = Tb == ARNG_2S && Q == 0 || Tb == ARNG_4S && Q == 1
	}
	return maxShift, match
}

// sizeTaMatchTb2 checks if the arrangement specifier Ta matches Tb, Tb and Q when Tb is encoded in size:Q.
func sizeTaMatchTb2(Ta, Tb, Q int16) bool {
	switch Ta {
	case ARNG_8H:
		return Tb == ARNG_8B && Q == 0 || Tb == ARNG_16B && Q == 1
	case ARNG_1Q:
		return Tb == ARNG_1D && Q == 0 || Tb == ARNG_2D && Q == 1
	}
	return false
}

// tMaxShift returns the maximum shift value corresponding to the arrangement specifier T.
func tMaxShift(T int16) int64 {
	maxShift := int64(0)
	switch T {
	case ARNG_8B, ARNG_16B:
		maxShift = 8
	case ARNG_4H, ARNG_8H:
		maxShift = 16
	case ARNG_2S, ARNG_4S:
		maxShift = 32
	case ARNG_2D:
		maxShift = 64
	}
	return maxShift
}

// imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1 returns the imm5 and imm4 values corresponding to the <Ts>[<index2>] combination.
func (c *ctxt7) imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1(p *obj.Prog, arg *arg) (uint32, uint32) {
	imm5, imm4 := uint32(0), uint32(0)
	index2 := arg.Index
	switch (arg.Reg >> 5) & 15 {
	case ARNG_B:
		c.checkindex(p, index2, 15)
		imm5 |= 1
		imm4 |= uint32(index2)
	case ARNG_H:
		c.checkindex(p, index2, 7)
		imm5 |= 2
		imm4 |= uint32(index2) << 1
	case ARNG_S:
		c.checkindex(p, index2, 3)
		imm5 |= 4
		imm4 |= uint32(index2) << 2
	case ARNG_D:
		c.checkindex(p, index2, 1)
		imm5 |= 8
		imm4 |= uint32(index2) << 3
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return imm5, imm4
}

// imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1 returns the imm5 value corresponding to the <Ts>[<index>] combination.
func (c *ctxt7) imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1(p *obj.Prog, arg *arg) uint32 {
	imm5 := uint32(0)
	switch (arg.Reg >> 5) & 15 {
	case ARNG_D:
		imm5 = c.imm5___D_8_index__imm5_1(p, arg)
	default:
		imm5 = c.imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1(p, arg)
	}
	return imm5
}

// imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1 returns the imm5 value corresponding to the <Ts>[<index>] combination.
func (c *ctxt7) imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1(p *obj.Prog, arg *arg) uint32 {
	imm5 := uint32(0)
	index := arg.Index
	switch (arg.Reg >> 5) & 15 {
	case ARNG_B:
		c.checkindex(p, index, 15)
		imm5 |= 1
		imm5 |= uint32(index) << 1
	case ARNG_H:
		c.checkindex(p, index, 7)
		imm5 |= 2
		imm5 |= uint32(index) << 2
	case ARNG_S:
		c.checkindex(p, index, 3)
		imm5 |= 4
		imm5 |= uint32(index) << 3
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return imm5
}

// imm5___D_8_index__imm5_1 returns the imm5 value corresponding to the <Ts>[<index>] combination.
func (c *ctxt7) imm5___D_8_index__imm5_1(p *obj.Prog, arg *arg) uint32 {
	af := (arg.Reg >> 5) & 15
	if af != ARNG_D {
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	index := arg.Index
	c.checkindex(p, index, 1)
	imm5 := uint32(8)
	imm5 |= uint32(index) << 4
	return imm5
}

// imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81 returns the imm5 and Q value corresponding to the arrangement specifier.
func (c *ctxt7) imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81(p *obj.Prog, arng int) (uint32, uint32) {
	imm5, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_8B:
		imm5, Q = 1, 0
	case ARNG_16B:
		imm5, Q = 1, 1
	case ARNG_4H:
		imm5, Q = 2, 0
	case ARNG_8H:
		imm5, Q = 2, 1
	case ARNG_2S:
		imm5, Q = 4, 0
	case ARNG_4S:
		imm5, Q = 4, 1
	case ARNG_2D:
		imm5, Q = 8, 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return imm5, Q
}

// immediate_8x8_a_b_c_d_e_f_g_h encodes a 64-bit immediate imm in
// 'aaaaaaaabbbbbbbbccccccccddddddddeeeeeeeeffffffffgggggggghhhhhhhh' format as "a:b:c:d:e:f:g:h".
func (c *ctxt7) immediate_8x8_a_b_c_d_e_f_g_h(p *obj.Prog, imm int64) uint32 {
	ret := 0
	for i := 0; i < 8; i++ {
		imm >>= uint(i << 3)
		tmp := imm & 0xff
		if tmp == 0xff {
			ret |= (1 << uint(i))
		} else if tmp != 0x00 {
			c.ctxt.Diag("invalid immediate: %v", p)
			return 0
		}
	}
	return uint32(ret)
}

// immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4 returns the imm5 and Q value corresponding to the arrangement specifier.
func (c *ctxt7) immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4(p *obj.Prog, arng int) uint32 {
	immh := uint32(0)
	switch arng {
	case ARNG_8H:
		immh = 1
	case ARNG_4S:
		immh = 2
	case ARNG_2D:
		immh = 4
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return immh
}

// immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41 returns the immh and Q value corresponding to the arrangement specifier.
func (c *ctxt7) immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41(p *obj.Prog, arng int) (uint32, uint32) {
	immh, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_8B:
		immh, Q = 1, 0
	case ARNG_16B:
		immh, Q = 1, 1
	case ARNG_4H:
		immh, Q = 2, 0
	case ARNG_8H:
		immh, Q = 2, 1
	case ARNG_2S:
		immh, Q = 4, 0
	case ARNG_4S:
		immh, Q = 4, 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return immh, Q
}

// immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81 returns the immh and Q value corresponding to the arrangement specifier.
func (c *ctxt7) immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81(p *obj.Prog, arng int) (uint32, uint32) {
	immh, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_2D:
		immh, Q = 8, 1
	default:
		immh, Q = c.immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41(p, arng)
	}
	return immh, Q
}

// Q___2S_0__4S_1 returns the Q value corresponding to the arrangement specifier.
func (c *ctxt7) Q___2S_0__4S_1(p *obj.Prog, arng int) uint32 {
	Q := uint32(0)
	switch arng {
	case ARNG_2S:
		Q = 0
	case ARNG_4S:
		Q = 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return Q
}

// Q___4H_0__8H_1 returns the Q value corresponding to the arrangement specifier.
func (c *ctxt7) Q___4H_0__8H_1(p *obj.Prog, arng int) uint32 {
	Q := uint32(0)
	switch arng {
	case ARNG_4H:
		Q = 0
	case ARNG_8H:
		Q = 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return Q
}

// Q___8B_0__16B_1 returns the Q value corresponding to the arrangement specifier.
func (c *ctxt7) Q___8B_0__16B_1(p *obj.Prog, arng int) uint32 {
	Q := uint32(0)
	switch arng {
	case ARNG_8B:
		Q = 0
	case ARNG_16B:
		Q = 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return Q
}

// size_imm returns the size value corresponding to the post-index immediate offset imm.
func (c *ctxt7) size_imm(p *obj.Prog, imm, mul int64) uint32 {
	switch imm {
	case 0, mul:
		return 0
	case mul << 1:
		return 1
	case mul << 2:
		return 2
	case mul << 3:
		return 3
	default:
		c.ctxt.Diag("invalid immediate, expected %d, %d, %d or %d: %v", mul, mul<<1, mul<<2, mul<<3, p)
	}
	return 0
}

// size_Q___8B_00__16B_01 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_8B:
		size, Q = 0, 0
	case ARNG_16B:
		size, Q = 0, 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__1D_30__2D_31 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__1D_30__2D_31(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_8B:
		size, Q = 0, 0
	case ARNG_16B:
		size, Q = 0, 1
	case ARNG_1D:
		size, Q = 3, 0
	case ARNG_2D:
		size, Q = 3, 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__4H_10__8H_11__4S_21 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__4H_10__8H_11__4S_21(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_4S:
		size, Q = 2, 1
	default:
		return c.size_Q___8B_00__16B_01__4H_10__8H_11(p, arng)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__4H_10__8H_11 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__4H_10__8H_11(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_8B:
		size, Q = 0, 0
	case ARNG_16B:
		size, Q = 0, 1
	case ARNG_4H:
		size, Q = 1, 0
	case ARNG_8H:
		size, Q = 1, 1
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_1D:
		size, Q = 3, 0
	default:
		return c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31(p, arng)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_2S:
		size, Q = 2, 0
	default:
		return c.size_Q___8B_00__16B_01__4H_10__8H_11__4S_21(p, arng)
	}
	return size, Q
}

// size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31(p *obj.Prog, arng int) (uint32, uint32) {
	size, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_2D:
		size, Q = 3, 1
	default:
		return c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21(p, arng)
	}
	return size, Q
}

// size___8H_0__1Q_3 returns the size value corresponding to the arrangement specifier.
func (c *ctxt7) size___8H_0__1Q_3(p *obj.Prog, arng int) uint32 {
	size := uint32(0)
	switch arng {
	case ARNG_8H:
		size = 0
	case ARNG_1Q:
		size = 3
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return size
}

// size___8H_0__4S_1__2D_2 returns the size value corresponding to the arrangement specifier.
func (c *ctxt7) size___8H_0__4S_1__2D_2(p *obj.Prog, arng int) uint32 {
	size := uint32(0)
	switch arng {
	case ARNG_8H:
		size = 0
	case ARNG_4S:
		size = 1
	case ARNG_2D:
		size = 2
	default:
		c.ctxt.Diag("invalid arrangement: %v", p)
	}
	return size
}

// sz_Q___2S_00__4S_01__2D_11 returns the size and Q value corresponding to the arrangement specifier.
func (c *ctxt7) sz_Q___2S_00__4S_01__2D_11(p *obj.Prog, arng int) (uint32, uint32) {
	sz, Q := uint32(0), uint32(0)
	switch arng {
	case ARNG_2S:
		sz, Q = 0, 0
	case ARNG_4S:
		sz, Q = 0, 1
	case ARNG_2D:
		sz, Q = 1, 1
	default:
		c.ctxt.Diag("invalid arrangement specifier: %v", p)
	}
	return sz, Q
}

// encodeArgs encodes the argument arg of p.
func (c *ctxt7) encodeArg(p *obj.Prog, arg *arg, atyp argtype) uint32 {
	if arg == nil {
		return 0
	}
	switch atyp {
	case arg_Wd, arg_Xd, arg_Wt, arg_Xt, arg_Rt_31_1__W_0__X_1:
		c.isRegOrZR(p, arg.Reg)
		return uint32(arg.Reg) & 0x1f

	case arg_Wds, arg_Xds:
		c.isRegOrRSP(p, arg.Reg)
		return uint32(arg.Reg) & 0x1f

	case arg_Wn, arg_Xn, arg_Rn_16_5__W_1__W_2__W_4__X_8:
		c.isRegOrZR(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Wns, arg_Xns, arg_Xns_mem, arg_Xns_mem_offset:
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Wt2, arg_Xt2, arg_Wa, arg_Xa:
		c.isRegOrZR(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 10

	case arg_Wm, arg_Ws, arg_Xm, arg_Xs:
		c.isRegOrZR(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 16

	case arg_Wmn, arg_Xmn:
		c.isRegOrZR(p, arg.Reg)
		return (uint32(arg.Reg)&0x1f)<<16 | (uint32(arg.Reg)&0x1f)<<5

	case arg_Ws_2_pair_even, arg_Xs_2_pair_even:
		rs, rs2 := arg.Reg, int16(arg.Offset)
		if rs&1 != 0 {
			c.ctxt.Diag("source register pair must start from even register: %v\n", p)
		}
		if rs != rs2-1 {
			c.ctxt.Diag("source register pair must be contiguous: %v\n", p)
		}
		c.isRegOrZR(p, rs)
		return (uint32(rs) & 0x1f) << 16

	case arg_Wt_2_pair_even, arg_Xt_2_pair_even:
		rt, rt2 := arg.Reg, int16(arg.Offset)
		if rt&1 != 0 {
			c.ctxt.Diag("destination register pair must start from even register: %v\n", p)
		}
		if rt != rt2-1 {
			c.ctxt.Diag("destination register pair must be contiguous: %v\n", p)
		}
		c.isRegOrZR(p, rt)
		return uint32(rt) & 0x1f

	case arg_Wt_pair, arg_Xt_pair:
		rt, rt2 := arg.Reg, int16(arg.Offset)
		c.isRegOrZR(p, rt)
		c.isRegOrZR(p, rt2)
		return uint32(rt)&0x1f | (uint32(rt2)&0x1f)<<10

	case arg_St_pair, arg_Dt_pair, arg_Qt_pair:
		rt, rt2 := arg.Reg, int16(arg.Offset)
		c.isFReg(p, rt)
		c.isFReg(p, rt2)
		return uint32(rt)&0x1f | (uint32(rt2)&0x1f)<<10

	case arg_Xns_mem_post_fixedimm_1:
		rm := arg.Offset
		if rm != 1 {
			c.ctxt.Diag("invalid post offset, expected 1: %v", p)
		}
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_fixedimm_2:
		rm := arg.Offset
		if rm != 2 {
			c.ctxt.Diag("invalid post offset, expected 2: %v", p)
		}
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_fixedimm_4:
		rm := arg.Offset
		if rm != 4 {
			c.ctxt.Diag("invalid post offset, expected 4: %v", p)
		}
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_fixedimm_8:
		rm := arg.Offset
		if rm != 8 {
			c.ctxt.Diag("invalid post offset, expected 8: %v", p)
		}
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_Q__8_0__16_1:
		c.isRegOrRSP(p, arg.Reg)
		// we have checked whether the offset value is legal in the unfold function.
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_Q__16_0__32_1:
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_Q__24_0__48_1:
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_Q__32_0__64_1:
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Xns_mem_post_size__1_0__2_1__4_2__8_3:
		c.isRegOrRSP(p, arg.Reg)
		size := c.size_imm(p, arg.Offset, 1)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<10

	case arg_Xns_mem_post_size__2_0__4_1__8_2__16_3:
		c.isRegOrRSP(p, arg.Reg)
		size := c.size_imm(p, arg.Offset, 2)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<10

	case arg_Xns_mem_post_size__3_0__6_1__12_2__24_3:
		c.isRegOrRSP(p, arg.Reg)
		size := c.size_imm(p, arg.Offset, 3)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<10

	case arg_Xns_mem_post_size__4_0__8_1__16_2__32_3:
		c.isRegOrRSP(p, arg.Reg)
		size := c.size_imm(p, arg.Offset, 4)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<10

	case arg_Xns_mem_post_Xm:
		rm := arg.Index & 0x1f
		if rm == 31 {
			c.ctxt.Diag("invalid register ZR: %v", p)
		}
		c.isRegOrRSP(p, arg.Reg)
		return (uint32(arg.Reg)&0x1f)<<5 | uint32(rm)<<16

	case arg_St, arg_Dt, arg_Qt, arg_Sd, arg_Hd, arg_Dd, arg_Qd:
		c.isFReg(p, arg.Reg)
		return uint32(arg.Reg) & 0x1f

	case arg_Sn, arg_Hn, arg_Dn, arg_Qn:
		c.isFReg(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Da, arg_Sa:
		c.isFReg(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 10

	case arg_Sm, arg_Dm:
		c.isFReg(p, arg.Reg)
		return (uint32(arg.Reg) & 0x1f) << 16

	case arg_Vd_16_5__B_1__H_2__S_4__D_8, arg_Vd_22_2__B_0__H_1__S_2, arg_Vd_22_2__H_0__S_1__D_2:
		c.isVReg(p, arg.Reg)
		return uint32(arg.Reg) & 0x1f

	case arg_Vd_22_2__D_3:
		c.isVReg(p, arg.Reg)
		return uint32(arg.Reg)&0x1f | 3<<22

	case arg_Vn_22_2__D_3:
		c.isVReg(p, arg.Reg)
		return (uint32(arg.Reg)&0x1f)<<5 | 3<<22

	case arg_Vm_22_2__D_3:
		c.isVReg(p, arg.Reg)
		return (uint32(arg.Reg)&0x1f)<<16 | 3<<22

	case arg_Vn_1_arrangement_16B:
		Vn := int16(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		opcode := (arg.Offset >> 12) & 15 // opcode
		Q := int(arg.Offset>>30) & 1      // Q is encoded in arg.Offset[30]
		size := int(arg.Offset>>10) & 3   // size is encoded in arg.Offset[11:10]
		if opcode != 0x7 || Q != 1 || size != 0 {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(Vn) << 5

	case arg_Vn_2_arrangement_16B:
		Vn := int16(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		opcode := (arg.Offset >> 12) & 15 // opcode
		Q := int(arg.Offset>>30) & 1      // Q is encoded in arg.Offset[30]
		size := int(arg.Offset>>10) & 3   // size is encoded in arg.Offset[11:10]
		if opcode != 0xa || Q != 1 || size != 0 {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(Vn)<<5 | 1<<13

	case arg_Vn_3_arrangement_16B:
		Vn := int16(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		opcode := (arg.Offset >> 12) & 15 // opcode
		Q := int(arg.Offset>>30) & 1      // Q is encoded in arg.Offset[30]
		size := int(arg.Offset>>10) & 3   // size is encoded in arg.Offset[11:10]
		if opcode != 0x6 || Q != 1 || size != 0 {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(Vn)<<5 | 2<<13

	case arg_Vn_4_arrangement_16B:
		Vn := int16(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		opcode := (arg.Offset >> 12) & 15 // opcode
		Q := int(arg.Offset>>30) & 1      // Q is encoded in arg.Offset[30]
		size := int(arg.Offset>>10) & 3   // size is encoded in arg.Offset[11:10]
		if opcode != 0x2 || Q != 1 || size != 0 {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(Vn)<<5 | 3<<13

	case arg_Vt_1_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31,
		arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31,
		arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31,
		arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__1D_30__2D_31:
		// Have checked the register size in fold function.
		Vn := uint32(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		Q := uint32(arg.Offset>>30) & 1    // Q is encoded in arg.Offset[30]
		size := uint32(arg.Offset>>10) & 3 // size is encoded in arg.Offset[11:10]
		return Vn | size<<10 | Q<<30

	case arg_Vt_2_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31,
		arg_Vt_3_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31,
		arg_Vt_4_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31:
		// Have checked the register size in fold function.
		Vn := uint32(arg.Offset) & 0x1f    // register is encoded in arg.Offset[4:0]
		Q := uint32(arg.Offset>>30) & 1    // Q is encoded in arg.Offset[30]
		size := uint32(arg.Offset>>10) & 3 // size is encoded in arg.Offset[11:10]
		if size == 3 && Q == 0 {           // doesn't include 1D
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return Vn | size<<10 | Q<<30

	case arg_Vt_1_arrangement_B_index__Q_S_size_1:
		at := (arg.Reg >> 5) & 15
		if at != ARNG_B {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		index := arg.Index
		if index < 0 || index > 15 {
			c.ctxt.Diag("index out of range 0 to 15: %v", p)
		}
		Q := uint32(index>>3) & 1
		S_size := uint32(index) & 7
		return uint32(arg.Reg)&0x1f | Q<<30 | S_size<<10

	case arg_Vt_1_arrangement_H_index__Q_S_size_1:
		at := (arg.Reg >> 5) & 15
		if at != ARNG_H {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		index := arg.Index
		if index < 0 || index > 7 {
			c.ctxt.Diag("index out of range 0 to 7: %v", p)
		}
		Q := uint32(index>>2) & 1
		S_size := uint32(index) & 3
		return uint32(arg.Reg)&0x1f | Q<<30 | S_size<<11

	case arg_Vt_1_arrangement_S_index__Q_S_1:
		at := (arg.Reg >> 5) & 15
		if at != ARNG_S {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		index := arg.Index
		if index < 0 || index > 3 {
			c.ctxt.Diag("index out of range 0 to 3: %v", p)
		}
		Q := uint32(index>>1) & 1
		S_size := uint32(index) & 1
		return uint32(arg.Reg)&0x1f | Q<<30 | S_size<<12

	case arg_Vt_1_arrangement_D_index__Q_1:
		at := (arg.Reg >> 5) & 15
		if at != ARNG_D {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		index := arg.Index
		if index < 0 || index > 1 {
			c.ctxt.Diag("index out of range 0 to 1: %v", p)
		}
		Q := uint32(index) & 1
		return uint32(arg.Reg)&0x1f | Q<<30

	case arg_Va_arrangement_16B:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_16B {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 10

	case arg_Vd_arrangement_16B:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_16B {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(arg.Reg) & 0x1f

	case arg_Vd_arrangement_2D:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_2D {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(arg.Reg) & 0x1f

	case arg_Vd_arrangement_4S:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_4S {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return uint32(arg.Reg) & 0x1f

	case arg_Vn_arrangement_16B:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_16B {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Vn_arrangement_2D:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_2D {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Vn_arrangement_4S:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_4S {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 5

	case arg_Vm_arrangement_16B:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_16B {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 16

	case arg_Vm_arrangement_2D:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_2D {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 16

	case arg_Vm_arrangement_4S:
		arng := int((arg.Reg >> 5) & 15)
		if arng != ARNG_4S {
			c.ctxt.Diag("invalid arrangement specifier: %v", p)
		}
		return (uint32(arg.Reg) & 0x1f) << 16

	case arg_Vd_arrangement_imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81:
		imm5, q := c.imm5_Q___8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | imm5<<16 | q<<30

	case arg_Vd_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1:
		imm5 := c.imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1(p, arg)
		return uint32(arg.Reg)&0x1f | imm5<<16

	case arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1:
		imm5 := c.imm5___B_1__H_2__S_4__D_8_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4__imm5lt4gt_8_1(p, arg)
		return (uint32(arg.Reg)&0x1f)<<5 | imm5<<16

	case arg_Vn_arrangement_imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1:
		imm5, imm4 := c.imm5___B_1__H_2__S_4__D_8_index__imm5_imm4__imm4lt30gt_1__imm4lt31gt_2__imm4lt32gt_4__imm4lt3gt_8_1(p, arg)
		return (uint32(arg.Reg)&0x1f)<<5 | imm5<<16 | imm4<<11

	case arg_Vn_arrangement_imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1:
		imm5 := c.imm5___B_1__H_2__S_4_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1(p, arg)
		return (uint32(arg.Reg)&0x1f)<<5 | imm5<<16

	case arg_Vn_arrangement_imm5___D_8_index__imm5_1, arg_Vn_arrangement_D_index__imm5_1:
		imm5 := c.imm5___D_8_index__imm5_1(p, arg)
		return (uint32(arg.Reg)&0x1f)<<5 | imm5<<16

	case arg_Vd_arrangement_immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4:
		immh := c.immh___SEEAdvancedSIMDmodifiedimmediate_0__8H_1__4S_2__2D_4(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | immh<<19

	case arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41:
		immh, q := c.immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | immh<<19 | q<<30

	case arg_Vd_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81:
		immh, q := c.immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | immh<<19 | q<<30

	case arg_Vn_arrangement_immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81:
		immh, q := c.immh_Q___SEEAdvancedSIMDmodifiedimmediate_00__8B_10__16B_11__4H_20__8H_21__2S_40__4S_41__2D_81(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | immh<<19 | q<<30

	case arg_Vn_arrangement_S_index__imm5__imm5lt41gt_1__imm5lt42gt_2__imm5lt43gt_4_1:
		af := (arg.Reg >> 5) & 15
		if af != ARNG_S {
			c.ctxt.Diag("invalid arrangement: %v", p)
		}
		index := arg.Index
		c.checkindex(p, index, 3)
		imm5 := uint32(4)
		imm5 |= uint32(index) << 3
		return (uint32(arg.Reg)&0x1f)<<5 | imm5<<16

	case arg_Vd_arrangement_Q___2S_0__4S_1:
		q := c.Q___2S_0__4S_1(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | q<<30

	case arg_Vd_arrangement_Q___4H_0__8H_1:
		q := c.Q___4H_0__8H_1(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | q<<30

	case arg_Vd_arrangement_Q___8B_0__16B_1:
		q := c.Q___8B_0__16B_1(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | q<<30

	case arg_Vn_arrangement_Q___8B_0__16B_1:
		q := c.Q___8B_0__16B_1(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | q<<30

	case arg_Vm_arrangement_Q___8B_0__16B_1:
		q := c.Q___8B_0__16B_1(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<16 | q<<30

	case arg_Vd_arrangement_size_Q___8B_00__16B_01:
		size, q := c.size_Q___8B_00__16B_01(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22 | q<<30

	case arg_Vn_arrangement_size_Q___8B_00__16B_01:
		size, q := c.size_Q___8B_00__16B_01(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22 | q<<30

	case arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22 | q<<30

	case arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22 | q<<30

	case arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22 | q<<30

	case arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22 | q<<30

	case arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<16 | size<<22 | q<<30

	case arg_Vd_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22 | q<<30

	case arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22 | q<<30

	case arg_Vm_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__2S_20__4S_21__2D_31(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<16 | size<<22 | q<<30

	case arg_Vn_arrangement_size_Q___8B_00__16B_01__4H_10__8H_11__4S_21:
		size, q := c.size_Q___8B_00__16B_01__4H_10__8H_11__4S_21(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22 | q<<30

	case arg_Vd_arrangement_size___8H_0__1Q_3:
		size := c.size___8H_0__1Q_3(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22

	case arg_Vd_arrangement_size___8H_0__4S_1__2D_2:
		size := c.size___8H_0__4S_1__2D_2(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | size<<22

	case arg_Vn_arrangement_size___8H_0__4S_1__2D_2:
		size := c.size___8H_0__4S_1__2D_2(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22

	case arg_Vn_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31:
		size, _ := c.size_Q___8B_00__16B_01__1D_30__2D_31(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | size<<22

	case arg_Vm_arrangement_size_Q___8B_00__16B_01__1D_30__2D_31:
		size, _ := c.size_Q___8B_00__16B_01__1D_30__2D_31(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<16 | size<<22

	case arg_Vd_arrangement_sz_Q___2S_00__4S_01__2D_11:
		sz, q := c.sz_Q___2S_00__4S_01__2D_11(p, int(arg.Reg>>5)&15)
		return uint32(arg.Reg)&0x1f | sz<<22 | q<<30

	case arg_Vn_arrangement_sz_Q___2S_00__4S_01__2D_11:
		sz, q := c.sz_Q___2S_00__4S_01__2D_11(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<5 | sz<<22 | q<<30

	case arg_Vm_arrangement_sz_Q___2S_00__4S_01__2D_11:
		sz, q := c.sz_Q___2S_00__4S_01__2D_11(p, int(arg.Reg>>5)&15)
		return (uint32(arg.Reg)&0x1f)<<16 | sz<<22 | q<<30

	case arg_Xns_mem_post_imm7_4_signed,
		arg_Xns_mem_wb_imm7_4_signed,
		arg_Xns_mem_optional_imm7_4_signed:
		return c.mem_imm7_signed(p, arg, 2)

	case arg_Xns_mem_post_imm7_8_signed,
		arg_Xns_mem_wb_imm7_8_signed,
		arg_Xns_mem_optional_imm7_8_signed:
		return c.mem_imm7_signed(p, arg, 3)

	case arg_Xns_mem_post_imm7_16_signed,
		arg_Xns_mem_wb_imm7_16_signed,
		arg_Xns_mem_optional_imm7_16_signed:
		return c.mem_imm7_signed(p, arg, 4)

	case arg_Xns_mem_post_imm9_1_signed,
		arg_Xns_mem_wb_imm9_1_signed:
		c.isRegOrRSP(p, arg.Reg)
		return uint32(arg.Offset&0x1FF)<<12 | uint32(arg.Reg&31)<<5

	case arg_Xns_mem_optional_imm12_1_unsigned:
		return c.mem_optional_imm12_unsigned(p, arg, 0)

	case arg_Xns_mem_optional_imm12_2_unsigned:
		return c.mem_optional_imm12_unsigned(p, arg, 1)

	case arg_Xns_mem_optional_imm12_4_unsigned:
		return c.mem_optional_imm12_unsigned(p, arg, 2)

	case arg_Xns_mem_optional_imm12_8_unsigned:
		return c.mem_optional_imm12_unsigned(p, arg, 3)

	case arg_Xns_mem_optional_imm12_16_unsigned:
		return c.mem_optional_imm12_unsigned(p, arg, 4)

	case arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__1_1,
		arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__2_1,
		arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__0_0__3_1,
		arg_Xns_mem_extend_m__UXTW_2__LSL_3__SXTW_6__SXTX_7__absent_0__0_1:
		c.isRegOrRSP(p, arg.Reg)
		// don't need to check the supported extensions because we have done it in the parsing process.
		if isRegShiftOrExt(arg.Index) {
			// extended or shifted offset register.
			c.checkShiftAmount(p, arg)
			return uint32(arg.Reg&31)<<5 | c.encRegShiftOrExt(arg.Type, arg.Index, false) /* includes reg, op, etc */
		}
		// (Rn)(Rm), no extension or shift.
		return uint32(arg.Reg&31)<<5 | uint32(0x6)<<12 | uint32(arg.Index&31)<<16

	case arg_Xns_mem_optional_imm9_1_signed:
		c.isRegOrRSP(p, arg.Reg)
		v := arg.Offset
		if v < -256 || v > 255 {
			c.ctxt.Diag("offset out of range: %d\n%v", v, p)
		}
		return uint32(arg.Reg&31)<<5 | uint32(v&0x1FF)<<12

	case arg_Wm_extend__UXTB_0__UXTH_1__LSL_UXTW_2__UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4:
		if arg.Reg <= REG_R31 {
			// without extension.
			c.isRegOrZR(p, arg.Reg)
			return (uint32(arg.Reg&0x1f))<<16 | 0x2<<13
		}
		amount := (arg.Reg >> 5) & 7 // arg.Reg[6:8] is the extended amount.
		if amount > 4 {
			c.ctxt.Diag("shift amount out of range 0 to 4: %v", p)
		}
		return c.encRegShiftOrExt(arg.Type, arg.Reg, true)

	case arg_Rm_extend__UXTB_0__UXTH_1__UXTW_2__LSL_UXTX_3__SXTB_4__SXTH_5__SXTW_6__SXTX_7__0_4:
		if arg.Reg <= REG_R31 {
			// without extension.
			c.isRegOrZR(p, arg.Reg)
			return (uint32(arg.Reg&0x1f))<<16 | 0x3<<13
		}
		amount := (arg.Reg >> 5) & 7 // arg.Reg[5:7] is the extended amount.
		if amount > 4 {
			c.ctxt.Diag("shift amount out of range 0 to 4: %v", p)
		}
		return c.encRegShiftOrExt(arg.Type, arg.Reg, false)

	case arg_Wm_shift__LSL_0__LSR_1__ASR_2__0_31:
		if arg.Reg != 0 {
			// without shift.
			c.isRegOrZR(p, arg.Reg)
			return uint32(arg.Reg&31) << 16
		}
		amount := (arg.Offset >> 10) & 63
		if amount >= 32 {
			c.ctxt.Diag("shift amount out of range 0 to 31: %v", p)
		}
		shift := (arg.Offset >> 22) & 3
		if shift > 2 || shift < 0 {
			c.ctxt.Diag("unsupported shift operator: %v", p)
		}
		// For obj.TYPE_SHIFT, register number, shift and amount has been set in the right
		// position of the arg.Offset field, and arg.Reg maybe 0.
		// For obj.Type_REG, register number is recored in ar.Reg field.
		return uint32(arg.Offset)

	case arg_Xm_shift__LSL_0__LSR_1__ASR_2__0_63:
		if arg.Reg != 0 {
			// without shift.
			c.isRegOrZR(p, arg.Reg)
			return uint32(arg.Reg&31) << 16
		}
		amount := (arg.Offset >> 10) & 63
		if amount >= 64 {
			c.ctxt.Diag("shift amount out of range 0 to 64: %v", p)
		}
		shift := (arg.Offset >> 22) & 3
		if shift > 2 || shift < 0 {
			c.ctxt.Diag("unsupported shift operator: %v", p)
		}
		return uint32(arg.Offset)

	case arg_Wm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_31:
		if arg.Reg != 0 {
			// without shift.
			c.isRegOrZR(p, arg.Reg)
			return uint32(arg.Reg&31) << 16
		}
		amount := (arg.Offset >> 10) & 63
		if amount > 31 {
			c.ctxt.Diag("shift amount out of range 0 to 31: %v", p)
		}
		return uint32(arg.Offset)

	case arg_Xm_shift__LSL_0__LSR_1__ASR_2__ROR_3__0_63:
		if arg.Reg != 0 {
			// without shift.
			c.isRegOrZR(p, arg.Reg)
			return uint32(arg.Reg&31) << 16
		}
		amount := (arg.Offset >> 10) & 63
		if amount > 63 {
			c.ctxt.Diag("shift amount out of range 0 to 63: %v", p)
		}
		return uint32(arg.Offset)

	case arg_IAddSub:
		v := arg.Offset
		if (v & 0xFFF000) != 0 {
			if v&0xFFF != 0 {
				c.ctxt.Diag("invalid arg_%s type constant %d: %v", argNames[atyp], v, p)
			}
			v >>= 12
			return uint32(1<<22 | (v&0xfff)<<10)
		}
		return uint32(v&0xfff) << 10

	case arg_immediate_bitmask_64_N_imms_immr:
		return bitconEncode(uint64(arg.Offset), 64)

	case arg_immediate_bitmask_32_imms_immr:
		return bitconEncode(uint64(arg.Offset), 32)

	case arg_immediate_cmode__8_0__16_1:
		cmode0 := 0
		if arg.Offset == 16 {
			cmode0 = 1
		} else if arg.Offset != 8 {
			c.ctxt.Diag("invalid immediate, expected 8 or 16: %v", p)
		}
		return uint32(cmode0) & 1 << 12

	case arg_immediate_exp_3_pre_4_imm8:
		rf := c.chipfloat7(arg.Val.(float64))
		if rf < 0 {
			c.ctxt.Diag("invalid floating-point immediate: %v", p)
		}
		return (uint32(rf&0xff) << 13)

	case arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4:
		// check whether the value is valid in the unfold function.
		return uint32(arg.Offset) << 16

	case arg_immediate_1_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__16UIntimmhimmb_1__32UIntimmhimmb_2__64UIntimmhimmb_4__128UIntimmhimmb_8:
		// arg.Index stores the arrangement specifier of the instruction, which is set in the unfold function.
		shift, max := arg.Offset, tMaxShift(arg.Index)
		if shift < 1 || shift > max {
			c.ctxt.Diag("shift amount out of range: %v\n", p)
		}
		return uint32(max-shift) << 16

	case arg_immediate_0_width_immh_immb__SEEAdvancedSIMDmodifiedimmediate_0__UIntimmhimmb8_1__UIntimmhimmb16_2__UIntimmhimmb32_4__UIntimmhimmb64_8:
		// arg.Index stores the arrangement specifier of the instruction, which is set in the unfold function.
		shift, max := arg.Offset, tMaxShift(arg.Index)
		if shift < 0 || shift > max-1 {
			c.ctxt.Diag("shift amount out of range: %v\n", p)
		}
		return uint32(shift) << 16

	case arg_immediate_shift_64_implicit_imm16_hw,
		arg_immediate_OptLSL_amount_16_0_48,
		arg_immediate_shift_64_implicit_inverse_imm16_hw:
		imm := arg.Offset
		if arg.Index > 0 { // the shift value has been set.
			return uint32((arg.Index>>4)&3)<<21 | (uint32(imm)&0xFFFF)<<5
		}
		hw := movcon(imm)
		if hw < 0 || hw > 3 {
			c.ctxt.Diag("invalid constant %d: %v", imm, p)
		}
		return uint32(hw<<21) | (uint32((imm>>uint(hw*16)))&0xFFFF)<<5

	case arg_immediate_shift_32_implicit_imm16_hw,
		arg_immediate_OptLSL_amount_16_0_16:
		imm := arg.Offset
		if arg.Index > 0 { // the shift value has been set.
			return uint32((arg.Index>>4)&3)<<21 | (uint32(imm)&0xFFFF)<<5
		}
		hw := movcon(int64(uint32(imm)))
		if hw < 0 || hw > 1 {
			c.ctxt.Diag("invalid constant %d: %v", imm, p)
		}
		return uint32(hw<<21) | (uint32((imm>>uint(hw*16)))&0xFFFF)<<5

	case arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1:
		offset := uint64(arg.Offset)
		cmode21 := uint32(0)
		if offset&^0xff == 0 {
			cmode21 = 0
		} else if (offset>>8)&^0xff == 0 {
			offset >>= 8
			cmode21 = 1
		} else {
			c.ctxt.Diag("invalid immediate %d: %v", offset, p)
		}
		return cmode21<<13 | (uint32(offset)>>5)<<16 | (uint32(offset)&0x1f)<<5

	case arg_immediate_OptLSL__a_b_c_d_e_f_g_h_cmode__0_0__8_1__16_2__24_3:
		offset := uint64(arg.Offset)
		cmode21 := uint32(0)
		if offset&^0xff == 0 {
			cmode21 = 0
		} else if (offset>>8)&^0xff == 0 {
			offset >>= 8
			cmode21 = 1
		} else if (offset>>16)&^0xff == 0 {
			offset >>= 16
			cmode21 = 2
		} else if (offset>>24)&^0xff == 0 {
			offset >>= 24
			cmode21 = 3
		} else {
			c.ctxt.Diag("invalid immediate %d: %v", offset, p)
		}
		return cmode21<<13 | (uint32(offset)>>5)<<16 | (uint32(offset)&0x1f)<<5

	case arg_immediate_shift_32_implicit_inverse_imm16_hw:
		imm := arg.Offset
		if arg.Index > 0 { // the shift value has been set.
			return uint32((arg.Index>>4)&3)<<21 | (uint32(imm)&0xFFFF)<<5
		}
		hw := movcon(int64(uint32(imm)))
		// excluding 0xffff0000 and 0x0000ffff.
		if hw < 0 || hw > 1 || uint32((imm>>uint(hw*16)))&0xFFFF == 0xffff {
			c.ctxt.Diag("invalid constant %d: %v", imm, p)
		}
		return uint32(hw<<21) | (uint32((imm>>uint(hw*16)))&0xFFFF)<<5

	case arg_immediate_0_7_op1:
		if (uint32(arg.Offset) &^ uint32(0x7)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 7: %v", p)
		}
		return (uint32(arg.Offset) & 0x7) << 16

	case arg_immediate_0_7_op2:
		if (uint64(arg.Offset) &^ uint64(0x7)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 7: %v", p)
		}
		return uint32((arg.Offset & 0x7) << 5)

	case arg_immediate_0_127_CRm_op2:
		if (uint64(arg.Offset) &^ uint64(0x7F)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 127: %v", p)
		}
		return uint32((arg.Offset & 0x7F) << 5) /* CRm:op2 */

	case arg_immediate_0_15_CRm, arg_immediate_optional_0_15_CRm:
		if (uint64(arg.Offset) &^ uint64(0xF)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 15: %v", p)
		}
		return uint32((arg.Offset & 0xF) << 8) /* Crm */

	case arg_immediate_0_65535_imm16, arg_immediate_optional_0_65535_imm16:
		if arg.Offset > 65535 || arg.Offset < 0 {
			c.ctxt.Diag("immediate out of range 0 to 65535: %v", p)
		}
		return uint32(arg.Offset) << 5

	case arg_immediate_BFI_BFM_64M_bitfield_lsb_64_immr,
		arg_immediate_SBFIZ_SBFM_64M_bitfield_lsb_64_immr,
		arg_immediate_UBFIZ_UBFM_64M_bitfield_lsb_64_immr:
		off := uint(arg.Offset)
		if off != 0 {
			off = 64 - off
		}
		return c.opbfm(p, off, 16, 63)

	case arg_immediate_BFI_BFM_32M_bitfield_lsb_32_immr,
		arg_immediate_SBFIZ_SBFM_32M_bitfield_lsb_32_immr,
		arg_immediate_UBFIZ_UBFM_32M_bitfield_lsb_32_immr:
		off := uint(arg.Offset)
		if off != 0 {
			off = 32 - off
		}
		return c.opbfm(p, off, 16, 31)

	case arg_immediate_BFI_BFM_64M_bitfield_width_64_imms,
		arg_immediate_SBFIZ_SBFM_64M_bitfield_width_64_imms,
		arg_immediate_UBFIZ_UBFM_64M_bitfield_width_64_imms:
		return c.opbfm(p, uint(arg.Offset-1), 10, 63)

	case arg_immediate_BFI_BFM_32M_bitfield_width_32_imms,
		arg_immediate_SBFIZ_SBFM_32M_bitfield_width_32_imms,
		arg_immediate_UBFIZ_UBFM_32M_bitfield_width_32_imms:
		return c.opbfm(p, uint(arg.Offset-1), 10, 31)

	case arg_immediate_BFXIL_BFM_64M_bitfield_width_64_imms,
		arg_immediate_SBFX_SBFM_64M_bitfield_width_64_imms,
		arg_immediate_UBFX_UBFM_64M_bitfield_width_64_imms:
		// arg.Index (lsb) is saved in arg.Index when unfolding.
		off := uint(int64(arg.Index) + arg.Offset - 1)
		return c.opbfm(p, off, 10, 63)

	case arg_immediate_BFXIL_BFM_32M_bitfield_width_32_imms,
		arg_immediate_SBFX_SBFM_32M_bitfield_width_32_imms,
		arg_immediate_UBFX_UBFM_32M_bitfield_width_32_imms:
		// arg.Index (lsb) is saved in arg.Index when unfolding.
		off := uint(int64(arg.Index) + arg.Offset - 1)
		return c.opbfm(p, off, 10, 31)

	case arg_immediate_0_63_immr,
		arg_immediate_BFXIL_BFM_64M_bitfield_lsb_64_immr,
		arg_immediate_SBFX_SBFM_64M_bitfield_lsb_64_immr,
		arg_immediate_UBFX_UBFM_64M_bitfield_lsb_64_immr:
		return c.opbfm(p, uint(arg.Offset), 16, 63)

	case arg_immediate_0_31_immr,
		arg_immediate_BFXIL_BFM_32M_bitfield_lsb_32_immr,
		arg_immediate_SBFX_SBFM_32M_bitfield_lsb_32_immr,
		arg_immediate_UBFX_UBFM_32M_bitfield_lsb_32_immr:
		return c.opbfm(p, uint(arg.Offset), 16, 31)

	case arg_immediate_0_63_imms:
		return c.opbfm(p, uint(arg.Offset), 10, 63)

	case arg_immediate_0_31_imms:
		return c.opbfm(p, uint(arg.Offset), 10, 31)

	case arg_immediate_LSL_UBFM_32M_bitfield_0_31_immr:
		return c.opbfm(p, uint(32-arg.Offset), 16, 31) | c.opbfm(p, uint(31-arg.Offset), 10, 30)

	case arg_immediate_LSL_UBFM_64M_bitfield_0_63_immr:
		return c.opbfm(p, uint(64-arg.Offset), 16, 63) | c.opbfm(p, uint(63-arg.Offset), 10, 62)

	case arg_immediate_LSR_UBFM_32M_bitfield_0_31_immr, arg_immediate_ASR_SBFM_32M_bitfield_0_31_immr:
		return c.opbfm(p, uint(arg.Offset), 16, 31) | c.opbfm(p, 31, 10, 31)

	case arg_immediate_LSR_UBFM_64M_bitfield_0_63_immr, arg_immediate_ASR_SBFM_64M_bitfield_0_63_immr:
		return c.opbfm(p, uint(arg.Offset), 16, 63) | c.opbfm(p, 63, 10, 63)

	case arg_immediate_MSL__a_b_c_d_e_f_g_h, arg_immediate_OptLSLZero__a_b_c_d_e_f_g_h:
		imm := (arg.Offset)
		if imm < -128 || imm > 255 {
			c.ctxt.Diag("immediate out of range -128 to 255: %v", p)
		}
		return ((uint32(imm)>>5)&7)<<16 | (uint32(imm)&0x1f)<<5

	case arg_immediate_0_31_imm5:
		v := arg.Offset
		if (uint64(arg.Offset) &^ uint64(0x1F)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 31: %v", p)
		}
		return uint32(v) << 16

	case arg_immediate_0_15_nzcv:
		v := arg.Offset
		if (uint64(arg.Offset) &^ uint64(0xF)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 15: %v", p)
		}
		return uint32(v)

	case arg_immediate_0_63_b5_b40:
		if (uint64(arg.Offset) &^ uint64(0x3F)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 63: %v", p)
		}
		return uint32((arg.Offset&0x20)<<26 | (arg.Offset&0x1f)<<19)

	case arg_immediate_0_63_imm6:
		if (uint64(arg.Offset) &^ uint64(0x3F)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 63: %v", p)
		}
		return uint32(arg.Offset) << 10

	case arg_immediate_8x8_a_b_c_d_e_f_g_h:
		imm := c.immediate_8x8_a_b_c_d_e_f_g_h(p, arg.Offset)
		return ((imm>>5)&7)<<16 | (imm&0x1f)<<5

	case arg_immediate_floatzero:
		if arg.Val.(float64) != 0.0 {
			c.ctxt.Diag("immediate must be 0.0: %v", p)
		}
		return 0

	case arg_immediate_index_Q_imm4__imm4lt20gt_00__imm4_10:
		index := arg.Offset
		var Q uint32
		var max int64
		af := arg.Index
		if af == ARNG_8B {
			Q = 0
			max = 7
		} else if af == ARNG_16B {
			Q = 1
			max = 15
		} else {
			c.ctxt.Diag("invalid arrangement, should be B8 or B16: %v", p)
		}
		if index < 0 || index > max {
			c.ctxt.Diag("illegal offset: %v", p)
		}
		return Q<<30 | (uint32(index&15) << 11)

	case arg_Cn:
		if (uint32(arg.Offset) &^ uint32(0xF)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 15: %v", p)
		}
		return (uint32(arg.Offset) & 0xf) << 12

	case arg_Cm:
		if (uint32(arg.Offset) &^ uint32(0xF)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 15: %v", p)
		}
		return (uint32(arg.Offset) & 0xf) << 8

	case arg_option_DMB_BO_system_CRm, arg_option_DSB_BO_system_CRm, arg_option_ISB_BI_system_CRm:
		if (uint64(arg.Offset) &^ uint64(0xF)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 15: %v", p)
		}
		return uint32(arg.Offset) << 8

	case arg_prfop_Rt:
		v := uint32(0xff)
		if arg.Type == obj.TYPE_CONST {
			v = uint32(arg.Offset)
			if v > 31 {
				c.ctxt.Diag("illegal prefetch operation: %v", p)
			}
		} else {
			val, ok := prfopfield[SpecialOperand(arg.Offset)]
			if !ok {
				c.ctxt.Diag("illegal prefetch operation:\n%v", p)
			}
			v = val
		}
		return v & 31

	case arg_sysreg_o0_op1_CRn_CRm_op2:
		// SysRegEnc function returns the system register encoding and accessFlags.
		_, v, accessFlags := SysRegEnc(arg.Reg)
		if v == 0 {
			c.ctxt.Diag("illegal system register: %v", p)
		}
		if (optab[uint32(p.Optab)].skeleton & (v &^ (3 << 19))) != 0 {
			c.ctxt.Diag("register value overlap: %v", p)
		}
		if p.As == AMRS && accessFlags&SR_READ == 0 {
			c.ctxt.Diag("system register is not readable: %v", p)
		} else if p.As == AMSR && accessFlags&SR_WRITE == 0 {
			c.ctxt.Diag("system register is not writable: %v", p)
		}
		return v

	case arg_sysop_SYS_CR_system:
		if (uint32(arg.Offset) &^ uint32(0x7)) != 0 {
			c.ctxt.Diag("immediate out of range 0 to 7: %v", p)
		}
		v := (uint32(arg.Offset) & 7) << 5
		if arg.Index == 0 { // Xt is integrated in arg.Index.
			v |= 0x1f
		} else {
			c.isRegOrZR(p, arg.Index)
			v |= uint32(arg.Index) & 0x1f
		}
		return v

	case arg_sysop_AT_SYS_CR_system, arg_sysop_IC_SYS_CR_system:
		// TODO: check if each sysop is valid
		v := uint32(arg.Offset)
		if arg.Index == 0 { // Xt is integrated in arg.Index.
			v |= 0x1f
		} else {
			c.isRegOrZR(p, arg.Index)
			v |= uint32(arg.Index)
		}
		return v

	case arg_sysop_DC_SYS_CR_system, arg_sysop_TLBI_SYS_CR_system:
		v := uint32(0)
		op, ok := sysInstFields[SpecialOperand(arg.Offset)]
		if !ok || (p.As == ATLBI && op.cn != 8) || (p.As == ADC && op.cn != 7) {
			c.ctxt.Diag("illegal argument: %v\n", p)
		}
		if op.hasOperand2 {
			if arg.Index == 0 { // Xt is integrated in arg.Index.
				c.ctxt.Diag("missing register at operand 2: %v\n", p)
			}
			c.isRegOrZR(p, arg.Index)
			v |= uint32(arg.Index & 0x1F)
		} else {
			if arg.Index != 0 {
				c.ctxt.Diag("extraneous register at operand 2: %v\n", p)
			}
			v |= uint32(0x1F)
		}
		return v | uint32(SYSARG4(int(op.op1), int(op.cn), int(op.cm), int(op.op2)))

	case arg_pstatefield_op1_op2__SPSel_05__DAIFSet_36__DAIFClr_37:
		v := uint32(0)
		// PSTATEfield can be special registers and special operands.
		if arg.Type == obj.TYPE_REG && arg.Reg == REG_SPSel {
			v = 5 << 5
		} else if arg.Type == obj.TYPE_SPECIAL {
			opd := SpecialOperand(arg.Offset)
			for _, val := range pstatefield {
				if val.opd == opd {
					v = val.enc
					break
				}
			}
		}
		if v == 0 {
			c.ctxt.Diag("illegal PSTATE field for immediate move: %v", p)
		}
		return v

	case arg_slabel_immhi_immlo_0:
		d := c.brdist(p, arg, 0, 21, 0)
		return uint32((d&3)<<29 | ((d>>2)&0x7FFFF)<<5)

	case arg_slabel_immhi_immlo_12:
		d := c.brdist(p, arg, 12, 21, 0)
		return uint32((d&3)<<29 | ((d>>2)&0x7FFFF)<<5)

	case arg_slabel_imm14_2:
		return uint32(c.brdist(p, arg, 0, 14, 2) << 5)

	case arg_slabel_imm19_2:
		d := uint32(c.brdist(p, arg, 0, 19, 2))
		return (d & 0x7FFFF) << 5

	case arg_slabel_imm26_2:
		return uint32(c.brdist(p, arg, 0, 26, 2))

	case arg_conditional:
		// BAL and BNV are not supported.
		if arg.Offset < 0 || arg.Offset > 13 {
			c.ctxt.Diag("invalid branch conditional instruction: %v", p)
		}
		return uint32(arg.Offset)

	case arg_cond_AllowALNV_Normal:
		cond := SpecialOperand(arg.Offset)
		if cond < SPOP_EQ || cond > SPOP_NV {
			c.ctxt.Diag("invalid condition: %v", p)
		} else {
			cond -= SPOP_EQ
		}
		return uint32(cond&15) << 12

	case arg_cond_NotAllowALNV_Invert:
		cond := SpecialOperand(arg.Offset)
		// AL and NV are not allowed.
		if cond < SPOP_EQ || cond > SPOP_LE {
			c.ctxt.Diag("invalid condition: %v", p)
		} else {
			// Invert the least significant bit.
			cond = (cond - SPOP_EQ) ^ 1
		}
		return uint32(cond&15) << 12
	}
	// TODO: enable more types.
	c.ctxt.Diag("unimplemented argument type: %v", p)
	return 0
}

// encodeOpcode encodes the opcode. The opcode of some special
// instructions affects the encoding, for most instructions
// this function just return 0.
// TODO enable this function when necessary.
/*
func (c *ctxt7) encodeOpcode(a obj.As) uint32 {
	switch a {
	default:
		return 0
	}
}
*/
