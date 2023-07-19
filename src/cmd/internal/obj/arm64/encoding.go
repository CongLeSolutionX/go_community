// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"log"
	"math/bits"
)

// This file contains the encoding of the each element type.

// encodeElm encodes an element. It is worth noting that the encoding of an element
// is also the checking process of an instruction, returns false if the check fails.
func (c *ctxt7) encodeElm(p *obj.Prog, bin uint32, ag *obj.Addr, instIdx, oprIdx, elmIdx int, checks map[uint32]uint32) (uint32, bool) {
	ai := instTab[instIdx].args[oprIdx]
	e := ai.elms[elmIdx]
	enc := uint32(0)
	reg, offset := uint32(ag.Reg), ag.Offset
	_, typ, _ := DecodeIndex(ag.Index)
	switch e {
	case sa_wt__Rt, sa_xt__Rt:
		if ai.aType == AC_PAIR && reg&1 != 0 {
			// The register must be an even-numbered register for register pair type.
			c.ctxt.Diag("destination register pair must start from even register: %v\n", p)
			return 0, false
		}
		fallthrough
	case sa_wd__Rd, sa_xd__Rd, sa_xd_1__Rd, sa_xt1__Rt, sa_wt1__Rt, sa_xt_1__Rt,
		sa_xt__Xt, sa_xt1__Xt, sa_xd__Xd, sa_xt__Rd, sa_t__Rt:
		enc = reg & 0x1f
	case sa_xd_sp__Rd, sa_xt_sp__Xt, sa_wd_wsp__Rd, sa_xd_sp__Xd:
		if reg == REG_R31 {
			c.ctxt.Diag("illegal destination register: %v\n", p)
			return 0, false
		}
		enc = reg & 0x1f
	case sa_xm_sp__Rm:
		if reg == REG_R31 {
			c.ctxt.Diag("illegal source register: %v\n", p)
			return 0, false
		}
		enc = reg & 0x1f
		if p.As == APACGA {
			enc <<= 16
		}
	case sa_x_s_plus_1, sa_w_s_plus_1:
		reg2 := uint32(ag.Offset) // the second register is saved in the Offset field
		if reg2 != reg+1 {
			c.ctxt.Diag("source register pair must be contiguous: %v\n", p)
			return 0, false
		}
	case sa_x_t_plus_1, sa_w_t_plus_1, sa_xt_plus_1__Rt:
		reg2 := uint32(ag.Offset) // the second register is saved in the Offset field
		if reg2 != reg+1 {
			c.ctxt.Diag("destination register pair must be contiguous: %v\n", p)
			return 0, false
		}
	case sa_xt2__Rt:
		// This element is used in SYSP and TLBIP, which we may have to implemented by optab.
		reg2 := uint32(ag.Offset) // the second register is saved in the Offset field
		if reg == REG_R31 && reg2 != REG_R31 || reg2 != reg+1 {
			c.ctxt.Diag("destination register pair must be contiguous: %v\n", p)
			return 0, false
		}
	case sa_xn_sp__Rn, sa_xn_sp__Xn, sa_wn_wsp__Rn:
		if reg == REG_R31 { // excluding the ZR register
			c.ctxt.Diag("illegal source register: %v\n", p)
			return 0, false
		}
		enc = (reg & 0x1f) << 5
	case sa_xn__Rn:
		if p.As == obj.ACALL {
			rel := obj.Addrel(c.cursym)
			rel.Off = int32(c.pc)
			rel.Siz = 0
			rel.Type = objabi.R_CALLIND
		}
		fallthrough
	case sa_wn__Rn, sa_xn_1__Rn, sa_xn_2__Rn:
		enc = (reg & 0x1f) << 5
	case sa_xn_1__Rn__and__Rm, sa_wn_1__Rn__and__Rm, sa_xs__Rn__and__Rm, sa_ws__Rn__and__Rm:
		enc = (reg&0x1f)<<5 | (reg&0x1f)<<16
	case sa_xt2__Xt2:
		reg = uint32(offset) // the second register is saved in the Offset field
		fallthrough
	case sa_xa__Ra, sa_wa__Ra:
		enc = (reg & 0x1f) << 10
	case sa_xm__Rm:
		if ai.aType == AC_MEMPOSTREG {
			_, r2, _, _, _, _ := c.decodeRegOffReg(p, ag)
			if reg == REGZERO {
				return 0, false
			}
			reg = uint32(r2)
		}
		fallthrough
	case sa_wm__Rm:
		if ai.aType == AC_MEMEXT {
			_, r2, _, _, _, _ := c.decodeRegOffReg(p, ag)
			reg = uint32(r2)
		}
		fallthrough
	case sa_m__Rm, sa_xm_1__Rm, sa_wm_1__Rm, sa_xm__Xm:
		enc = (reg & 0x1f) << 16
	case sa_xm_sp__Xm:
		if reg == REG_R31 {
			c.ctxt.Diag("illegal register: %v\n", p)
			return 0, false
		}
		enc = (reg & 0x1f) << 16
	case sa_ws__Rs:
		// The meaning of sa_ws__Rs is not unique for now. For general instructions,
		// it is encoded in bits 16-20, for SME instructions, it is encoded in bits 13-14.
		// This may be a bug of the xml document to be fixed in a subsequent version.
		if ai.aType != AC_REG && ai.aType != AC_PAIR {
			if !(reg >= REG_R12 && reg <= REG_R15) {
				return 0, false
			}
			return (reg & 3) << 13, true
		}
		fallthrough
	case sa_xs__Rs:
		if ai.aType == AC_PAIR && reg&1 != 0 {
			// The register must be an even-numbered register for register pair type.
			c.ctxt.Diag("source register pair must start from even register: %v\n", p)
			return 0, false
		}
		fallthrough
	case sa_xs_1__Rs:
		enc = (reg & 0x1f) << 16
	case sa_xt2__Rt2, sa_wt2__Rt2:
		reg = uint32(ag.Offset) // the second register is saved in the Offset field
		rt2 := reg & 0x1f
		// The encoding of this element is not unique for now. For some general instructions
		// such as LDP, STP, LDNP etc., it's encoded in bits 10-14, for some instructions added
		// later via features FEAT_LSE128, FEAT_LRCPC3 and FEAT_D128__FEAT_THE, it is encoded
		// in bits 16-20. This may be a bug of the xml document to be fixed in a subsequent version.
		if instTab[instIdx].feature == FEAT_NONE {
			enc = rt2 << 10
		} else {
			enc = rt2 << 16
		}
	case sa_const_REG_X16:
		if reg != REG_R16 {
			return 0, false
		}
	case sa_const_MEMEXT_no_arng1:
		_, _, arng1, _, _, _ := c.decodeRegOffReg(p, ag)
		if typ != RTYP_MEM_ROFF || arng1 != ARNG_NONE {
			return 0, false
		}
	case sa_const_MEMIMM_0, sa_const_MEMPREIMM_no_offset:
		if typ != RTYP_NORMAL || offset != 0 {
			return 0, false
		}
	case sa_const_MEMEXT_no_arng2:
		_, _, _, arng2, _, _ := c.decodeRegOffReg(p, ag)
		if typ != RTYP_MEM_ROFF || arng2 != ARNG_NONE {
			return 0, false
		}
	case sa_const_MEMPOSTIMM_4:
		if offset != 4 {
			c.ctxt.Diag("offset must be 4: %v\n", p)
			return 0, false
		}
	case sa_const_MEMPOSTIMM_8:
		if offset != 8 {
			c.ctxt.Diag("offset must be 8: %v\n", p)
			return 0, false
		}
	case sa_const_MEMPOSTIMM_16:
		if offset != 16 {
			c.ctxt.Diag("offset must be 16: %v\n", p)
			return 0, false
		}
	case sa_const_MEMPREIMM_n4:
		if offset != -4 {
			c.ctxt.Diag("offset must be -4: %v\n", p)
			return 0, false
		}
	case sa_const_MEMPREIMM_n8:
		if offset != -8 {
			c.ctxt.Diag("offset must be -8: %v\n", p)
			return 0, false
		}
	case sa_const_MEMPREIMM_n16:
		if offset != -16 {
			c.ctxt.Diag("offset must be -16: %v\n", p)
			return 0, false
		}
	case sa_const_SPOP_CSYNC:
		spop := SpecialOperand(offset)
		if spop != SPOP_CSYNC {
			c.ctxt.Diag("invalid argument, expect CSYNC: %v\n", p)
			return 0, false
		}
	case sa_const_SPOP_DSYNC:
		spop := SpecialOperand(offset)
		if spop != SPOP_DSYNC {
			c.ctxt.Diag("invalid argument, expect DSYNC: %v\n", p)
			return 0, false
		}
	case sa_const_SPOP_RCTX:
		spop := SpecialOperand(offset)
		if spop != SPOP_RCTX {
			c.ctxt.Diag("invalid argument, expect RCTX: %v\n", p)
			return 0, false
		}
	case sa_amount__imm6:
		_, _, amount := c.decodeRegShift(p, ag)
		if amount < 0 || amount > 31 {
			c.ctxt.Diag("shift amount out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = (amount & 0x3f) << 10
	case sa_amount_1__imm6:
		_, _, amount := c.decodeRegShift(p, ag)
		if amount < 0 || amount > 63 {
			c.ctxt.Diag("shift amount out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = (amount & 0x3f) << 10
	case sa_amount__imm3:
		_, _, amount := c.decodeRegExt(p, ag)
		if amount < 0 || amount > 4 { // in the range 0 to 4
			c.ctxt.Diag("shift amount out of range 0 to 4: %v\n", p)
			return 0, false
		}
		enc = (amount & 7) << 10
	case sa_amount__S:
		_, _, _, _, _, amount := c.decodeRegOffReg(p, ag)
		// <amount> must be #0, encoded in "S" as 0 if omitted, or as 1 if present.
		if amount != 0 {
			c.ctxt.Diag("invalid index shift amount: %v\n", p)
			return 0, false
		}
		// We don't distinguish Rn.<EXT> and Rn.<EXT><<0 in go, so encode as 0 is omitted.
		enc = 0 << 12
	case sa_amount__S__0_1:
		_, _, _, _, _, amount := c.decodeRegOffReg(p, ag)
		if amount&^1 != 0 {
			c.ctxt.Diag("invalid index shift amount: %v\n", p)
			return 0, false
		}
		enc = amount << 12
	case sa_amount__S__0_2:
		_, _, _, _, _, amount := c.decodeRegOffReg(p, ag)
		if amount != 0 && amount != 2 {
			c.ctxt.Diag("invalid index shift amount: %v\n", p)
			return 0, false
		}
		enc = (amount >> 1) << 12
	case sa_amount__S__0_3, sa_amount_1__S__0_3:
		_, _, _, _, _, amount := c.decodeRegOffReg(p, ag)
		if amount != 0 && amount != 3 {
			c.ctxt.Diag("invalid index shift amount: %v\n", p)
			return 0, false
		}
		enc = (amount & 1) << 12

	case sa_const__imm13:
		width := uint32(0)
		_, arng := c.decodeRegARNG(p, &p.To)
		switch arng {
		case ARNG_B:
			width = 8
		case ARNG_H:
			width = 16
		case ARNG_S:
			width = 32
		case ARNG_D:
			width = 64
		}
		n, imms, immr, isImmLogical := encodeBitMask(uint64(offset), width)
		if !isImmLogical {
			c.ctxt.Diag("invalid immediate: %v\n", p)
			return 0, false
		}
		enc = n<<17 | immr<<11 | imms<<5
	case sa_imm__imm5:
		// This element has different meanings in different instructions,
		// although it is encoded in bits 16-20 in all cases.
		imm5 := offset
		switch instTab[instIdx].armOp {
		case A64_CMPEQ, A64_CMPGT, A64_CMPGE, A64_CMPLT, A64_CMPLE, A64_CMPNE,
			A64_INDEX:
			// is the signed immediate operand, in the range -16 to 15.
			if imm5 < -16 || imm5 > 15 {
				c.ctxt.Diag("the immediate out of range -16 to 15: %v\n", p)
				return 0, false
			}
			return (uint32(imm5) & 31) << 16, true
		case A64_CCMN, A64_CCMP, A64_LD1B, A64_LD1SB, A64_LDFF1B, A64_LDFF1SB,
			A64_PRFB, A64_ST1B:
			// Is a five bit unsigned (positive) immediate.
		case A64_LD1H, A64_LD1SH, A64_LDFF1H, A64_LDFF1SH, A64_PRFH, A64_ST1H:
			// Is the optional unsigned immediate byte offset,
			// a multiple of 2 in the range 0 to 62, defaulting to 0.
			if imm5&1 != 0 {
				c.ctxt.Diag("offset must be a multiple of 2: %v\n", p)
				return 0, false
			}
			imm5 = imm5 >> 1
		case A64_LD1W, A64_LD1SW, A64_LDFF1W, A64_LDFF1SW, A64_PRFW, A64_ST1W:
			// Is the optional unsigned immediate byte offset,
			// a multiple of 4 in the range 0 to 124, defaulting to 0.
			if imm5&3 != 0 {
				c.ctxt.Diag("offset must be a multiple of 4: %v\n", p)
				return 0, false
			}
			imm5 = imm5 >> 2
		case A64_LD1D, A64_LDFF1D, A64_PRFD, A64_ST1D:
			// Is the optional unsigned immediate byte offset,
			// a multiple of 8 in the range 0 to 248, defaulting to 0.
			if imm5&7 != 0 {
				c.ctxt.Diag("offset must be a multiple of 8: %v\n", p)
				return 0, false
			}
			imm5 = imm5 >> 3
		default:
			return 0, false
		}
		if imm5&^0x1f != 0 {
			c.ctxt.Diag("offset out of range: %v\n", p)
			return 0, false
		}
		enc = (uint32(imm5) & 31) << 16
	case sa_imm__imm7:
		imm7 := offset
		switch p.As {
		case ALDNPW, ALDPW, ALDPSW, ASTNPW, ASTPW:
			// Is the optional signed immediate byte offset, a multiple of 4 in the range -256 to 252,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/4.
			if imm7&3 != 0 || imm7 < -256 || imm7 > 252 {
				c.ctxt.Diag("offset must be a multiple of 4 in the range -256 to 252: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 2) & 0x7f) << 15
		case AFLDNPD, AFLDPD, AFSTNPD, AFSTPD:
			// Is the optional signed immediate byte offset, a multiple of 8 in the range -512 to 504,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/8.
			if imm7&7 != 0 || imm7 < -512 || imm7 > 504 {
				c.ctxt.Diag("offset must be a multiple of 8 in the range -512 to 504: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 3) & 0x7f) << 15
		default:
			return 0, false
		}
	case sa_imm_1__imm7:
		// <imm> is an immediate value in some instructions, and an address offset
		// value in some other instructions, but it's always saved in the Offset field.
		imm7 := offset
		switch p.As {
		case AZCMPHI, AZCMPHS, AZCMPLO, AZCMPLS:
			// Is the unsigned immediate operand, in the range 0 to 127, encoded in bits 14-20.
			if imm7&^0x7f != 0 {
				c.ctxt.Diag("the immediate out of range 0 to 127: %v\n", p)
				return 0, false
			}
			enc = (uint32(imm7) & 0x7f) << 14
		case ALDPW, ALDPSW, ASTPW:
			// Is the signed immediate byte offset, a multiple of 4 in the range -256 to 252,
			// encoded in bits 15-21 as <imm>/4.
			if imm7&3 != 0 || imm7 < -256 || imm7 > 252 {
				c.ctxt.Diag("offset must be a multiple of 4 in the range -256 to 252: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 2) & 0x7f) << 15
		case ALDNP, AFLDPD, ASTNP, AFSTPD:
			// Is the optional signed immediate byte offset, a multiple of 8 in the range -512 to 504,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/8.
			if imm7&7 != 0 || imm7 < -512 || imm7 > 504 {
				c.ctxt.Diag("offset must be a multiple of 8 in the range -512 to 504: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 3) & 0x7f) << 15
		case AFLDNPQ, AFSTNPQ:
			// Is the optional signed immediate byte offset, a multiple of 16 in the range -1024 to 1008,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/16.
			if imm7&15 != 0 || imm7 < -1024 || imm7 > 1008 {
				c.ctxt.Diag("offset must be a multiple of 16 in the range -1024 to 1008: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 4) & 0x7f) << 15
		default:
			return 0, false
		}
	case sa_imm_2__imm7:
		imm7 := offset
		switch p.As {
		case AFLDNPS, AFSTNPS:
			// Is the signed immediate byte offset, a multiple of 4 in the range -256 to 252,
			// encoded in bits 15-21 as <imm>/4.
			if imm7&3 != 0 || imm7 < -256 || imm7 > 252 {
				c.ctxt.Diag("offset must be a multiple of 4 in the range -256 to 252: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 2) & 0x7f) << 15
		case ALDP, ASTP:
			// Is the optional signed immediate byte offset, a multiple of 8 in the range -512 to 504,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/8.
			if imm7&7 != 0 || imm7 < -512 || imm7 > 504 {
				c.ctxt.Diag("offset must be a multiple of 8 in the range -512 to 504: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 3) & 0x7f) << 15
		case AFLDPQ, AFSTPQ:
			// Is the optional signed immediate byte offset, a multiple of 16 in the range -1024 to 1008,
			// defaulting to 0 and encoded in bits 15-21 as <imm>/16.
			if imm7&15 != 0 || imm7 < -1024 || imm7 > 1008 {
				c.ctxt.Diag("offset must be a multiple of 16 in the range -1024 to 1008: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 4) & 0x7f) << 15
		default:
			return 0, false
		}
	case sa_imm_3__imm7:
		imm7 := offset
		switch p.As {
		case ALDP, ASTP:
			// Is the optional signed immediate byte offset, a multiple of 8 in the range -512 to 504,
			// encoded in bits 15-21 as <imm>/8.
			if imm7&7 != 0 || imm7 < -512 || imm7 > 504 {
				c.ctxt.Diag("offset must be a multiple of 8 in the range -512 to 504: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 3) & 0x7f) << 15
		case AFLDPQ, AFSTPQ:
			// Is the optional signed immediate byte offset, a multiple of 16 in the range -1024 to 1008,
			// encoded in bits 15-21 as <imm>/16.
			if imm7&15 != 0 || imm7 < -1024 || imm7 > 1008 {
				c.ctxt.Diag("offset must be a multiple of 16 in the range -1024 to 1008: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm7) >> 4) & 0x7f) << 15
		default:
			return 0, false
		}
	case sa_imm__imm16:
		imm16 := uint64(offset)
		switch p.As {
		case AMOVKW, AMOVNW, AMOVZW:
			// The immediate may be left shifted by either 0 (the default) or 16.
			if imm16 > 0xffff {
				if imm16&0xffff != 0 {
					c.ctxt.Diag("the immediate out of range: %v\n", p)
					return 0, false
				}
				imm16 >>= 16
			}
		case AMOVK, AMOVN, AMOVZ:
			// The immediate may be left shifted by either 0 (the default), 16, 32 or 48.
			shift := movcon(offset)
			if shift == -1 {
				c.ctxt.Diag("the immediate out of range: %v\n", p)
				return 0, false
			}
			imm16 = imm16 >> shift
		case AUDF:
			if imm16&^0xffff != 0 {
				c.ctxt.Diag("the immediate out of range: %v\n", p)
				return 0, false
			}
			return uint32(offset), true
		}
		if imm16&^0xffff != 0 {
			c.ctxt.Diag("the immediate out of range: %v\n", p)
			return 0, false
		}
		enc = uint32(imm16) << 5
	case sa_imm_1__imm16_hw:
		if offset&^0xffff == 0 {
			enc = uint32(offset) << 5
		} else if offset&^0xffff0000 == 0 {
			enc = (uint32(offset) >> 16) << 5
		} else {
			c.ctxt.Diag("the immediate out of range: %v\n", p)
			return 0, false
		}
	case sa_imm_2__imm16_hw:
		// The immediate may be left shifted by either 0 (the default), 16, 32 or 48.
		shift := movcon(offset)
		if shift == -1 {
			c.ctxt.Diag("the immediate out of range: %v\n", p)
			return 0, false
		}
		enc = (uint32(offset) >> shift) << 5
	case sa_imm__imm12:
		if offset&^0xfff == 0 {
			enc = uint32(offset) << 10
		} else if offset&^0xfff000 == 0 {
			enc = (uint32(offset) >> 12) << 10
		} else {
			c.ctxt.Diag("the immediate out of range: %v\n", p)
			return 0, false
		}
	case sa_imm__imms_immr, sa_imm_2__imms_immr:
		n, imms, immr, isImmLogical := encodeBitMask(uint64(offset), 32)
		if !isImmLogical {
			c.ctxt.Diag("invalid immediate: %v\n", p)
			return 0, false
		}
		enc = n<<22 | immr<<16 | imms<<10
	case sa_imm_1__N_imms_immr, sa_imm_3__N_imms_immr:
		n, imms, immr, isImmLogical := encodeBitMask(uint64(offset), 64)
		if !isImmLogical {
			c.ctxt.Diag("invalid immediate: %v\n", p)
			return 0, false
		}
		enc = n<<22 | immr<<16 | imms<<10
	case sa_imm_1__simm7, sa_imm__simm7:
		simm7 := offset
		if simm7&15 != 0 || simm7 < -1024 || simm7 > 1008 {
			c.ctxt.Diag("offset must be a multiple of 16 in the range -1024 to 1008: %v\n", p)
			return 0, false
		}
		enc = ((uint32(simm7) >> 4) & 0x7f) << 15

	case sa_imm__b5_b40:
		if offset&^0x3f != 0 {
			c.ctxt.Diag("the bit number should be in the range 0 to 63: %v\n", p)
			return 0, false
		}
		b40 := offset & 0x1f
		b5 := (offset >> 5) & 1
		enc = uint32(b5<<31 | b40<<19)

	case sa_imm__CRm:
		imm := offset
		switch p.As {
		case AMSR:
			// Is a 4-bit unsigned immediate, in the range 0 to 15, encoded in bits 8-11.
			// Restricted to the range 0 to 1, encoded in bit 8, when <pstatefield> is ALLINT,
			// PM, SVCRSM, SVCRSMZA, or SVCRZA
			pstatefield := SpecialOperand(p.To.Reg)
			if pstatefield == SPOP_ALLINT || pstatefield == SPOP_PM || pstatefield == SPOP_SVCRSM ||
				pstatefield == SPOP_SVCRSMZA || pstatefield == SPOP_SVCRZA {
				if imm&^1 != 0 {
					c.ctxt.Diag("the immediate must be 0 or 1: %v\n", p)
					return 0, false
				}
			}
			fallthrough
		default:
			// Is a 4-bit unsigned immediate, in the range 0 to 15, encoded in bits 8-11.
			if imm&^0xf != 0 {
				c.ctxt.Diag("the immediate out of range 0 to 15: %v\n", p)
				return 0, false
			}
			enc = uint32(imm) << 8
		}
	case sa_imm__CRm_op2:
		if offset&^0x7f != 0 { // in the range 0 to 127
			c.ctxt.Diag("the immediate out of range 0 to 127: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 5

	case sa_imm5__Rt:
		if offset&^31 != 0 { // in the range 0 to 31
			c.ctxt.Diag("the immediate out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset)
	case sa_imm6__option_2__option_0__S_Rt_2_0:
		if offset&^63 != 0 { // in the range 0 to 63
			c.ctxt.Diag("the immediate out of range 0 to 63: %v\n", p)
			return 0, false
		}
		imm6 := uint32(offset)
		option2, option1 := (imm6>>5)&1, (imm6>>3)&3
		rt := imm6 & 7
		enc = option2<<15 | option1<<12 | rt
	case sa_imms__imms:
		if offset&^0x1f != 0 {
			c.ctxt.Diag("imms out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10
	case sa_imms_1__imms:
		if offset&^0x3f != 0 {
			c.ctxt.Diag("imms out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10
	case sa_immr__immr, sa_shift__immr:
		if offset&^0x1f != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 16
	case sa_immr_1__immr:
		if offset&^0x3f != 0 {
			c.ctxt.Diag("immr out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 16

	case sa_simm__imm8:
		if offset < -128 || offset > 127 {
			c.ctxt.Diag("the immediate out of range -128 to 127: %v\n", p)
			return 0, false
		}
		enc = (uint32(offset) & 0xff) << 10
	case sa_simm__imm9:
		if typ != RTYP_NORMAL || offset < -256 || offset > 255 { // <simm> is memory offset
			c.ctxt.Diag("offset out of range [-256,255]: %v\n", p)
			return 0, false
		}
		switch p.As {
		case ALDG, ASTG, ASTZG, ASTZ2G:
			if offset&0xf != 0 || offset < -4096 || offset > 4080 {
				c.ctxt.Diag("offset out of range [-4096,4080]: %v\n", p)
				return 0, false
			}
			offset >>= 4
		}
		enc = (uint32(offset) & 0x1ff) << 12
	case sa_simm__S_imm9:
		// Is the optional signed immediate byte offset, a multiple of 8 in the range -4096 to 4088,
		// defaulting to 0 and encoded in the "S:imm9" field as <simm>/8.
		if typ != RTYP_NORMAL || offset < -4096 || offset > 4088 || offset&7 != 0 {
			c.ctxt.Diag("offset must be a multiple of 8 in the range -4096 to 4088: %v\n", p)
			return 0, false
		}
		offset >>= 3
		enc = ((uint32(offset)>>9)&1)<<22 | (uint32(offset)&0x1ff)<<12

	case sa_pimm__imm12:
		imm12 := offset
		switch p.As {
		case AFMOVB, AMOVBU, AMOVBW, AMOVB:
			if imm12&^0xfff != 0 {
				c.ctxt.Diag("offset out of range 0 to 4095: %v\n", p)
				return 0, false
			}
			enc = uint32(imm12) << 10
		case AMOVHU, AMOVHW, AMOVH:
			if imm12&1 != 0 || imm12 < 0 || imm12 > 8190 {
				c.ctxt.Diag("offset must be a multiple of 2 in the range 0 to 8190: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm12) >> 1) & 0xfff) << 10
		case AMOVWU, AMOVW:
			if imm12&3 != 0 || imm12 < 0 || imm12 > 16380 {
				c.ctxt.Diag("offset must be a multiple of 4 in the range 0 to 16380: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm12) >> 2) & 0xfff) << 10
		case APRFM:
			if imm12&7 != 0 || imm12 < 0 || imm12 > 32760 {
				c.ctxt.Diag("offset must be a multiple of 8 in the range 0 to 32760: %v\n", p)
				return 0, false
			}
			enc = ((uint32(imm12) >> 3) & 0xfff) << 10
		default:
			return 0, false
		}
	case sa_pimm_1__imm12:
		pimm := offset
		if pimm&7 != 0 || pimm < 0 || pimm > 32760 {
			c.ctxt.Diag("offset must be a multiple of 8 in the range 0 to 32760: %v\n", p)
			return 0, false
		}
		enc = ((uint32(pimm) >> 3) & 0xfff) << 10

	case sa_uimm__imm8:
		uimm := offset
		if uimm&^0xff != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 255: %v\n", p)
			return 0, false
		}
		enc = uint32(uimm) << 10
	case sa_uimm4__uimm4:
		uimm4 := offset
		if uimm4&^0xf != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 15: %v\n", p)
			return 0, false
		}
		enc = uint32(uimm4) << 10
	case sa_uimm6__uimm6:
		uimm6 := offset >> 4
		if offset&0xf != 0 || uimm6&^0x3f != 0 {
			c.ctxt.Diag("the immediate must be a multiple of 16 in the range 0 to 1008: %v\n", p)
			return 0, false
		}
		enc = uint32(uimm6) << 16

	case sa_shift__shift__ASR_LSL_LSR_ROR:
		_, shift, _ := c.decodeRegShift(p, ag)
		if shift < SHIFT_LL || shift > SHIFT_ROR {
			c.ctxt.Diag("unsupported shift operator: %v\n", p)
			return 0, false
		}
		enc = ((shift - SHIFT_LL) & 3) << 22
	case sa_shift__shift__ASR_LSL_LSR:
		_, shift, _ := c.decodeRegShift(p, ag)
		if shift < SHIFT_LL || shift > SHIFT_AR {
			c.ctxt.Diag("unsupported shift operator: %v\n", p)
			return 0, false
		}
		enc = ((shift - SHIFT_LL) & 3) << 22
	case sa_shift__sh__LSL__0_LSL__12:
		sh := uint32(0)
		if offset&^0xfff == 0 {
			sh = 0
		} else if offset&^0xfff000 == 0 {
			sh = 1
		} else {
			return 0, false
		}
		enc = sh << 22
	case sa_shift__hw:
		// The immediate may be left shifted by either 0 (the default) or 16.
		hw := uint32(0)
		if offset != 0 && offset&^0xffff0000 == 0 {
			hw = 1
		}
		enc = hw << 21
	case sa_shift_1__hw:
		shift := movcon(offset)
		if shift == -1 {
			return 0, false
		}
		enc = uint32(shift) << 21
	case sa_shift__imms:
		if offset&^31 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10
	case sa_shift__imm6:
		if offset&^63 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 15
	case sa_shift_1__imms:
		if offset&^63 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10
	case sa_shift_2__immr, sa_shift_1__immr:
		if offset&^63 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 16
	case sa_shift_1:
		if offset&^31 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 31: %v\n", p)
			return 0, false
		}
		imms := uint32(31 - offset)
		immr := uint32(0)
		if offset != 0 {
			immr = uint32(32 - offset)
		}
		enc = imms<<10 | immr<<16
	case sa_shift_3:
		if offset&^63 != 0 {
			c.ctxt.Diag("the immediate out of range 0 to 63: %v\n", p)
			return 0, false
		}
		imms := uint32(63 - offset)
		immr := uint32(0)
		if offset != 0 {
			immr = uint32(64 - offset)
		}
		enc = imms<<10 | immr<<16

	case sa_nzcv__nzcv:
		if offset&^0xf != 0 {
			c.ctxt.Diag("invalid condition: %v\n", p)
			return 0, false
		}
		enc = uint32(offset)
	case sa_cond__cond:
		cond := SpecialOperand(offset)
		if cond < SPOP_EQ || cond > SPOP_NV {
			c.ctxt.Diag("invalid condition: %v\n", p)
			return 0, false
		} else {
			cond -= SPOP_EQ
		}
		enc = uint32(cond&15) << 12
	case sa_cond_1__cond:
		cond := SpecialOperand(offset)
		// Excluding AL and NV.
		if cond < SPOP_EQ || cond > SPOP_NV || (cond == SPOP_AL || cond == SPOP_NV) {
			c.ctxt.Diag("invalid condition: %v\n", p)
			return 0, false
		} else {
			cond -= SPOP_EQ
		}
		cond ^= 1 // invert the least significant bit
		enc = uint32(cond&15) << 12
	case sa_extend__option__LSL_SXTW_SXTX_UXTW, sa_extend__option__SXTW_SXTX_UXTW:
		// For sa_extend__option__SXTW_SXTX_UXTW, we have combined LSL into this element
		// in the xml parser, so it also contains LSL.
		_, _, _, _, ext, _ := c.decodeRegOffReg(p, ag)
		if ext != RTYP_EXT_UXTW && ext != RTYP_EXT_LSL && ext != RTYP_EXT_SXTW && ext != RTYP_EXT_SXTX {
			c.ctxt.Diag("invalid extension: %v\n", p)
			return 0, false
		}
		enc = encodeExtend(ext, true)
	case sa_extend__option__LSL_SXTB_SXTH_SXTW_SXTX_UXTB_UXTH_UXTW_UXTX:
		_, ext, _ := c.decodeRegExt(p, ag)
		if ext < RTYP_EXT_UXTB || ext > RTYP_EXT_LSL {
			c.ctxt.Diag("invalid extension: %v\n", p)
			return 0, false
		}
		enc = encodeExtend(ext, false) // for 32-bit
	case sa_extend_1__option__LSL_SXTB_SXTH_SXTW_SXTX_UXTB_UXTH_UXTW_UXTX:
		_, ext, _ := c.decodeRegExt(p, ag)
		if ext < RTYP_EXT_UXTB || ext > RTYP_EXT_LSL {
			c.ctxt.Diag("invalid extension: %v\n", p)
			return 0, false
		}
		enc = encodeExtend(ext, true) // for 64-bit
	case sa_width:
		// Is the width of the bitfield, in the range 1 to 32-<lsb>, encoded in bits 10-15 as <width>-1.
		// Fetch lsb.
		lsb := int64(0)
		if len(instTab[instIdx].args) == 3 { // BFCW
			lsb = p.GetFrom3().Offset // the second source operand
		} else {
			lsb = p.From.Offset
		}
		if offset < 1 || offset > 32-lsb {
			c.ctxt.Diag("width value out of range: %v\n", p)
			return 0, false
		}
		switch p.As {
		case AUBFXW, ABFXILW, ASBFXW:
			enc = (uint32(lsb+offset-1) & 0x3f) << 10
		case ABFCW, ABFIW, ASBFIZW, AUBFIZW:
			enc = (uint32(offset-1) & 0x3f) << 10
		}
	case sa_width_1:
		// Is the width of the bitfield, in the range 1 to 64-<lsb>, encoded in bits 10-15 as <width>-1.
		// Fetch lsb.
		lsb := int64(0)
		if len(instTab[instIdx].args) == 3 { // BFC
			lsb = p.GetFrom3().Offset // the second source operand
		} else {
			lsb = p.From.Offset
		}
		if offset < 1 || offset > 64-lsb {
			c.ctxt.Diag("width value out of range: %v\n", p)
			return 0, false
		}
		switch p.As {
		case AUBFX, ABFXIL, ASBFX:
			enc = (uint32(lsb+offset-1) & 0x3f) << 10
		case ABFC, ABFI, ASBFIZ, AUBFIZ:
			enc = (uint32(offset-1) & 0x3f) << 10
		}
	case sa_lsb:
		if offset&^0x1f != 0 { // in the range 0 to 31
			c.ctxt.Diag("lsb out of range 0 to 31: %v\n", p)
			return 0, false
		}
		if offset != 0 {
			enc = 32 - uint32(offset)
		}
		enc <<= 16
	case sa_lsb_1:
		if offset&^0x1f != 0 { // in the range 0 to 31
			c.ctxt.Diag("lsb out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 16
	case sa_lsb_2:
		if offset&^0x3f != 0 { // in the range 0 to 63
			c.ctxt.Diag("lsb out of range 0 to 63: %v\n", p)
			return 0, false
		}
		if offset != 0 {
			enc = 64 - uint32(offset)
		}
		enc <<= 16
	case sa_lsb_3:
		if offset&^0x3f != 0 { // in the range 0 to 63
			c.ctxt.Diag("lsb out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 16
	case sa_lsb__imms:
		if offset&^0x1f != 0 {
			c.ctxt.Diag("lsb out of range 0 to 31: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10
	case sa_lsb_1__imms:
		if offset&^0x3f != 0 {
			c.ctxt.Diag("lsb out of range 0 to 63: %v\n", p)
			return 0, false
		}
		enc = uint32(offset) << 10

	case sa_label__immhi_immlo:
		label := uint32(0)
		switch p.As {
		case AADRP:
			// In the range +/-4GB, is encoded as "immhi:immlo" times 4096.
			label = uint32(c.brdist(p, 12, 21, 0))
		case AADR:
			// In the range +/-1MB, is encoded in "immhi:immlo", 19 bits.
			label = uint32(c.brdist(p, 0, 21, 0))
		default:
			return 0, false
		}
		enc = (label&3)<<29 | ((label>>2)&0x7FFFF)<<5
	case sa_label__imm14:
		// In the range +/-32KB, is encoded as "imm14" times 4.
		enc = uint32(c.brdist(p, 0, 14, 2) << 5)
	case sa_label__imm19:
		// In the range +/-1MB, is encoded as "imm19" times 4.
		enc = uint32(c.brdist(p, 0, 19, 2) << 5)
	case sa_label__imm26:
		// In the range +/-128MB, is encoded as "imm26" times 4.
		if ag.Sym == nil {
			enc = uint32(c.brdist(p, 0, 26, 2))
		} else {
			rel := obj.Addrel(c.cursym)
			rel.Off = int32(c.pc)
			rel.Siz = 4
			rel.Sym = ag.Sym
			rel.Add = ag.Offset
			rel.Type = objabi.R_CALLARM64
		}

	case sa_mask__mask:
		if offset&^0xf != 0 {
			c.ctxt.Diag("mask out of range 0 to 15: %v\n", p)
			return 0, false
		}
		enc = uint32(offset)
	case sa_op2__op2, sa_op1__op1, sa_cn__CRn, sa_cm__CRm, sa_op0__o0__2_3:
		// We don't support the operand format S<op0>_<op1>_<Cn>_<Cm>_<op2> in Go, instead we
		// use <systemreg>, so maybe it doesn't matter if these elements are supported or not.
		return 0, false
	case sa_op1_Cn_Cm_op2:
		// #<op1>, <Cn>, <Cm>, #<op2> is combined as one immediate operand, and ruled by this one.
		if offset&^0x7ffe0 != 0 {
			c.ctxt.Diag("invalid immediate: %v\n", p)
			return 0, false
		}
		enc = uint32(offset)
	case sa_at_op__op1_CRm_0__op2:
		op, ok := sysInstAT[SpecialOperand(offset)]
		if !ok {
			c.ctxt.Diag("invalid at_op: %v\n", p)
			return 0, false
		}
		enc = uint32(SYSARG4(int(op.op1), 7, int(op.cm), int(op.op2)))
	case sa_brb_op__op2:
		op := SpecialOperand(offset)
		switch op {
		case SPOP_IALL:
			enc = 4 << 5
		case SPOP_INJ:
			enc = 5 << 5
		default:
			c.ctxt.Diag("invalid brb_op: %v\n", p)
			return 0, false
		}

	case sa_dc_op__op1_CRm_op2:
		op, ok := sysInstTLBIDC[SpecialOperand(p.From.Offset)]
		if !ok || op.cn != 7 {
			c.ctxt.Diag("illegal argument: %v\n", p)
			return 0, false
		}
		enc = uint32(SYSARG4(int(op.op1), int(op.cn), int(op.cm), int(op.op2)))

	case sa_ic_op__op1_CRm_op2:
		op := SpecialOperand(offset)
		op1, crm, op2 := uint32(0), uint32(0), uint32(0)
		switch op {
		case SPOP_IALLUIS:
			op1, crm, op2 = 0, 1, 0
		case SPOP_IALLU:
			op1, crm, op2 = 0, 5, 0
		case SPOP_IVAU:
			op1, crm, op2 = 3, 5, 1
		default:
			c.ctxt.Diag("invalid ic_op: %v\n", p)
			return 0, false
		}
		enc = op1<<16 | crm<<8 | op2<<5
	case sa_prfop__Rt_4_3:
		op := SpecialOperand(offset)
		v, ok := prfopfield[op]
		if !ok { // the prefetch operation may be not supported yet
			c.ctxt.Diag("invalid prf_op: %v\n", p)
			return 0, false
		}
		enc = v & 31
	case sa_rprfop__Rt_0:
		op := SpecialOperand(offset)
		switch op {
		case SPOP_PLDKEEP:
			enc = 0
		case SPOP_PLDSTRM:
			enc = 4
		case SPOP_PSTKEEP:
			enc = 1
		case SPOP_PSTSTRM:
			enc = 5
		default:
			c.ctxt.Diag("invalid rprf_op: %v\n", p)
			return 0, false
		}

	case sa_tlbi_op__op1_CRn_CRm_op2:
		op, ok := sysInstTLBIDC[SpecialOperand(p.From.Offset)]
		if !ok || op.cn == 7 {
			c.ctxt.Diag("illegal argument: %v\n", p)
			return 0, false
		}
		// Some tlbi_ops allow a second operand, some don't.
		if !op.hasOperand2 && len(instTab[instIdx].args) > 1 {
			c.ctxt.Diag("extraneous register at operand 2: %v\n", p)
			return 0, false
		} else if op.hasOperand2 && len(instTab[instIdx].args) < 2 {
			c.ctxt.Diag("missing register at operand 2: %v\n", p)
			return 0, false
		}
		enc = uint32(SYSARG4(int(op.op1), int(op.cn), int(op.cm), int(op.op2)))

	case sa_tlbip_op__op1_CRn_CRm_op2:
		op, ok := sysInstTLBIP[SpecialOperand(offset)]
		if !ok {
			c.ctxt.Diag("invalid tlbip_op: %v\n", p)
			return 0, false
		}
		enc = uint32(SYSARG4(int(op.op1), int(op.cn), int(op.cm), int(op.op2)))

	case sa_systemreg__o0_op1_CRn_CRm_op2:
		_, v, accessFlags := SysRegEnc(ag.Reg)
		if v == 0 { // illegal system register
			c.ctxt.Diag("illegal system register: %v\n", p)
			return 0, false
		}
		switch p.As {
		case AMRRS, AMRS:
			if accessFlags&SR_READ == 0 { // system register is not readable
				c.ctxt.Diag("system register is not readable: %v\n", p)
				return 0, false
			}
		case AMSRR, AMSR:
			if accessFlags&SR_WRITE == 0 { // system register is not writable
				c.ctxt.Diag("system register is not writable: %v\n", p)
				return 0, false
			}
		default:
			return 0, false
		}
		enc = v
	case sa_pstatefield__op1_op2_CRm:
		// Some state fields are classified as system registers.
		op1, op2, crm := uint32(0), uint32(0), uint32(0)
		if ai.aType == AC_SPR {
			switch reg {
			case REG_UAO:
				op1, op2, crm = 0, 3, 0
			case REG_PAN:
				op1, op2, crm = 0, 4, 0
			case REG_SPSel:
				op1, op2, crm = 0, 5, 0
			case REG_SSBS:
				op1, op2, crm = 3, 1, 0
			case REG_DIT:
				op1, op2, crm = 3, 2, 0
			case REG_TCO:
				op1, op2, crm = 3, 4, 0
			default:
				return 0, false
			}
		} else { // AC_SPOP
			op := SpecialOperand(p.To.Offset)
			switch op {
			case SPOP_ALLINT:
				op1, op2, crm = 1, 0, 0
			case SPOP_PM:
				op1, op2, crm = 1, 0, 2
			case SPOP_SVCRSM:
				op1, op2, crm = 3, 3, 2
			case SPOP_SVCRZA:
				op1, op2, crm = 3, 3, 4
			case SPOP_SVCRSMZA:
				op1, op2, crm = 3, 3, 6
			case SPOP_DAIFSet:
				op1, op2, crm = 3, 6, 0
			case SPOP_DAIFClr:
				op1, op2, crm = 3, 7, 0
			default:
				return 0, false
			}
		}
		enc = op1<<16 | crm<<8 | op2<<5

	case sa_r__option__W_X, sa_r__b5__W_X:
		// GO calls X and W registers as R registers collectively, so these elements have no encoding in Go.
	case sa_option__CRm_2_1:
		mode := SpecialOperand(offset)
		switch mode {
		case SPOP_SM:
			enc = 1 << 9
		case SPOP_ZA:
			enc = 2 << 9
		default:
			c.ctxt.Diag("invalid option: %v\n", p)
			return 0, false
		}

	case sa_option_1:
		bop := SpecialOperand(offset)
		switch bop {
		case SPOP_SYnXS:
			enc = 3 << 10
		case SPOP_ISHnXS:
			enc = 2 << 10
		case SPOP_NSHnXS:
			enc = 1 << 10
		case SPOP_OSHnXS:
		default:
			c.ctxt.Diag("invalid option: %v\n", p)
			return 0, false
		}
	case sa_option, sa_option__CRm:
		bop := SpecialOperand(offset)
		if option := encodeBarrierOp(bop); option > 15 || e == sa_option && (bop == SPOP_SSBB || bop == SPOP_PSSBB) {
			c.ctxt.Diag("invalid option: %v\n", p)
			return 0, false
		} else {
			enc = option << 8
		}
	case sa_targets__op2_2_1___omitted__c_j_jc:
		direction := SpecialOperand(offset)
		switch direction {
		case SPOP_C:
			enc = 1
		case SPOP_J:
			enc = 2
		case SPOP_JC:
			enc = 3
		default:
			c.ctxt.Diag("invalid targets: %v\n", p)
			return 0, false
		}
		enc <<= 6
	default:
		log.Fatalf("unimplemented element type %s: %v\n", e, p)
	}
	return enc, true
}

func encodeExtend(ext uint32, is64Bit bool) uint32 {
	option := uint32(0)
	switch ext {
	case RTYP_EXT_UXTB:
		option = 0
	case RTYP_EXT_UXTH:
		option = 1
	case RTYP_EXT_UXTW:
		option = 2
	case RTYP_EXT_UXTX:
		option = 3
	case RTYP_EXT_SXTB:
		option = 4
	case RTYP_EXT_SXTH:
		option = 5
	case RTYP_EXT_SXTW:
		option = 6
	case RTYP_EXT_SXTX:
		option = 7
	case RTYP_EXT_LSL:
		if is64Bit {
			option = 3
		} else {
			option = 2
		}
	default:
		log.Fatalf("unrecognized extension type %v\n", ext)
	}
	return option << 13
}

func encodeBarrierOp(bop SpecialOperand) uint32 {
	crm := uint32(0)
	switch bop {
	case SPOP_SY:
		crm = 15
	case SPOP_ST:
		crm = 14
	case SPOP_LD:
		crm = 13
	case SPOP_ISH:
		crm = 11
	case SPOP_ISHST:
		crm = 10
	case SPOP_ISHLD:
		crm = 9
	case SPOP_NSH:
		crm = 7
	case SPOP_NSHST:
		crm = 6
	case SPOP_NSHLD:
		crm = 5
	case SPOP_PSSBB:
		crm = 4
	case SPOP_OSH:
		crm = 3
	case SPOP_OSHST:
		crm = 2
	case SPOP_OSHLD:
		crm = 1
	case SPOP_SSBB:
		crm = 0
	default:
		return 16
	}
	return crm
}

func lowestSetBit(value uint64) uint64 {
	return value & -value
}

func isPowerOf2(value uint64) bool {
	return (value != 0) && ((value & (value - 1)) == 0)
}

// encodeBitMask tests if a given value can be encoded in the immediate field of a
// logical instruction.
// If it can be encoded, the function returns true, and three fields n, imm_s and
// imm_r required by the logical instruction.
// If it can not be encoded, the function returns false, and the values n, imm_s and
// imm_r are 0.
//
// Logical immediates are encoded using parameters n, imm_s and imm_r using
// the following table:
//
//	N   imms    immr    size        S             R
//	1  ssssss  rrrrrr    64    UInt(ssssss)  UInt(rrrrrr)
//	0  0sssss  xrrrrr    32    UInt(sssss)   UInt(rrrrr)
//	0  10ssss  xxrrrr    16    UInt(ssss)    UInt(rrrr)
//	0  110sss  xxxrrr     8    UInt(sss)     UInt(rrr)
//	0  1110ss  xxxxrr     4    UInt(ss)      UInt(rr)
//	0  11110s  xxxxxr     2    UInt(s)       UInt(r)
//
// (s bits must not be all set)
//
// A pattern is constructed of size bits, where the least significant S+1 bits
// are set. The pattern is rotated right by R, and repeated across a 32 or
// 64-bit value, depending on destination register width.
//
// Put another way: the basic format of a logical immediate is a single
// contiguous stretch of 1 bits, repeated across the whole word at intervals
// given by a power of 2. To identify them quickly, we first locate the
// lowest stretch of 1 bits, then the next 1 bit above that; that combination
// is different for every logical immediate, so it gives us all the
// information we need to identify the only logical immediate that our input
// could be, and then we simply check if that's the value we actually have.
//
// (The rotation parameter does give the possibility of the stretch of 1 bits
// going 'round the end' of the word. To deal with that, we observe that in
// any situation where that happens the bitwise NOT of the value is also a
// valid logical immediate. So we simply invert the input whenever its low bit
// is set, and then we know that the rotated case can't arise.)
func encodeBitMask(value uint64, width uint32) (n, imms, immr uint32, isImmLogical bool) {
	negate := false
	if value&1 == 1 {
		// If the low bit is 1, negate the value, and set a flag to remember that we
		// did (so that we can adjust the return values appropriately).
		negate = true
		value = ^value
	}
	if width <= 32 {
		// To handle 8/16/32-bit logical immediates, repeat the input value to fill a
		// 64-bit word. The correct encoding of that as a logical immediate will also
		// be the correct encoding of the value.
		for bits := width; bits <= 32; bits <<= 1 {
			value <<= bits
			mask := uint64(1)<<bits - 1
			value |= ((value >> bits) & mask)
		}
	}
	a := lowestSetBit(value)
	t1 := value + a
	b := lowestSetBit(t1)
	t2 := t1 - b
	c := lowestSetBit(t2)
	var d, clza int32
	var mask uint64
	if c != 0 {
		clza = int32(bits.LeadingZeros64(a))
		d = clza - int32(bits.LeadingZeros64(c))
		mask = ((uint64(1) << d) - 1)
		n = 0
	} else {
		if a == 0 {
			return 0, 0, 0, false
		} else {
			clza = int32(bits.LeadingZeros64(a))
			d = 64
			mask = ^uint64(0)
			n = 1
		}
	}
	if !isPowerOf2(uint64(d)) {
		return 0, 0, 0, false
	}
	if ((b - a) & (^mask)) != 0 {
		return 0, 0, 0, false
	}
	multipliers := []uint64{
		0x0000000000000001,
		0x0000000100000001,
		0x0001000100010001,
		0x0101010101010101,
		0x1111111111111111,
		0x5555555555555555,
	}
	multiplier := multipliers[bits.LeadingZeros64(uint64(d))-57]
	candidate := (b - a) * multiplier
	if value != candidate {
		return 0, 0, 0, false
	}
	clzb := int32(-1)
	if b != 0 {
		clzb = int32(bits.LeadingZeros64(b))
	}
	s := clza - clzb
	r := int32(0)
	if negate {
		s = d - s
		r = (clzb + 1) & (d - 1)
	} else {
		r = (clza + 1) & (d - 1)
	}
	imms = uint32(((2 * -d) | (s - 1)) & 0x3f)
	immr = uint32(r)
	isImmLogical = true
	return
}

// encodeArgs encodes the argument ag of p.
func (c *ctxt7) encodeArg(p *obj.Prog, bin uint32, ag *obj.Addr, instIdx, oprIdx int, checks map[uint32]uint32) (uint32, bool) {
	ai := instTab[instIdx].args[oprIdx]
	enc := uint32(0)
	for i := range ai.elms {
		if v, ok := c.encodeElm(p, bin, ag, instIdx, oprIdx, i, checks); !ok {
			return 0, false
		} else {
			enc |= v
		}
	}
	return enc, true
}
