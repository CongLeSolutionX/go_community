// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// This file defines unfoldTab which is an array of functions. Those functions
// map Go assembly opcode to Arm64 instructions in Arm specification.

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/src"
)

// idx is obj.Addr.Index field.
func isRegShiftOrExt(idx int16) bool {
	return (idx-obj.RBaseARM64)&REG_EXT != 0 || (idx-obj.RBaseARM64)&REG_LSL != 0
}

func isaddcon(v int64) bool {
	/* uimm12 or uimm24? */
	if v < 0 {
		return false
	}
	if (v & 0xFFF) == 0 {
		v >>= 12
	}
	return v <= 0xFFF
}

func isaddcon2(v int64) bool {
	return 0 <= v && v <= 0xFFFFFF
}

/*
 * if v contains a single 16-bit value aligned
 * on a 16-bit field, and thus suitable for movk/movn,
 * return the field index 0 to 3; otherwise return -1
 */
func movcon(v int64) int {
	for s := 0; s < 64; s += 16 {
		if (uint64(v) &^ (uint64(0xFFFF) << uint(s))) == 0 {
			return s / 16
		}
	}
	return -1
}

// isbitcon reports whether a constant can be encoded into a logical instruction.
// bitcon has a binary form of repetition of a bit sequence of length 2, 4, 8, 16, 32, or 64,
// which itself is a rotate (w.r.t. the length of the unit) of a sequence of ones.
// special cases: 0 and -1 are not bitcon.
// this function needs to run against virtually all the constants, so it needs to be fast.
// for this reason, bitcon testing and bitcon encoding are separate functions.
func isbitcon(x uint64) bool {
	if x == 1<<64-1 || x == 0 {
		return false
	}
	// determine the period and sign-extend a unit to 64 bits
	switch {
	case x != x>>32|x<<32:
		// period is 64
		// nothing to do
	case x != x>>16|x<<48:
		// period is 32
		x = uint64(int64(int32(x)))
	case x != x>>8|x<<56:
		// period is 16
		x = uint64(int64(int16(x)))
	case x != x>>4|x<<60:
		// period is 8
		x = uint64(int64(int8(x)))
	default:
		// period is 4 or 2, always true
		// 0001, 0010, 0100, 1000 -- 0001 rotate
		// 0011, 0110, 1100, 1001 -- 0011 rotate
		// 0111, 1011, 1101, 1110 -- 0111 rotate
		// 0101, 1010             -- 01   rotate, repeat
		return true
	}
	return sequenceOfOnes(x) || sequenceOfOnes(^x)
}

// sequenceOfOnes tests whether a constant is a sequence of ones in binary, with leading and trailing zeros
func sequenceOfOnes(x uint64) bool {
	y := x & -x // lowest set bit of x. x is good iff x+y is a power of 2
	y += x
	return (y-1)&y == 0
}

// con64class classifies the 64-bit constant v.
func (c *ctxt7) con64class(v int64) argtype {
	zeroCount := 0
	negCount := 0
	for i := uint(0); i < 4; i++ {
		immh := uint32(v >> (i * 16) & 0xffff)
		if immh == 0 {
			zeroCount++
		} else if immh == 0xffff {
			negCount++
		}
	}
	// Use movz in preference.
	if zeroCount >= 3 {
		return C_MOVCONZ
	} else if negCount >= 3 {
		return C_MOVCONN
	} else if zeroCount == 2 {
		return C_MOVCONZ
	} else if negCount == 2 {
		return C_MOVCONN
	} else if zeroCount == 1 {
		return C_MOVCONZ
	} else if negCount == 1 {
		return C_MOVCONN
	} else {
		return C_MOVCONZ
	}
}

// rclass classifies register r.
func rclass(r int16) argtype {
	switch {
	case REG_R0 <= r && r <= REG_R31, r == REG_RSP:
		return C_REG
	case REG_F0 <= r && r <= REG_F31:
		return C_FREG
	case REG_V0 <= r && r <= REG_V31:
		return C_VREG
	case COND_EQ <= r && r <= COND_NV:
		return C_COND
	case r >= REG_ARNG && r < REG_ELEM:
		return C_ARNG
	case r >= REG_ELEM && r < REG_ELEM_END:
		return C_ELEM
	case r >= REG_UXTB && r < REG_SPECIAL,
		r >= REG_LSL && r < REG_ARNG:
		return C_EXTREG
	case r >= REG_SPECIAL:
		return C_SPR
	}
	return C_NONE
}

// memclass classifies memory offset o.
func memclass(o int64) argtype {
	if o == 0 {
		return C_ZOREG
	}
	if int64(int32(o)) == o {
		return C_LOREG
	}
	return C_VOREG
}

// aclass classifies p's argument a.
func (c *ctxt7) aclass(p *obj.Prog, a *obj.Addr) argtype {
	if a == nil {
		return C_NONE
	}
	switch a.Type {
	case obj.TYPE_NONE:
		return C_NONE

	case obj.TYPE_REG:
		return rclass(a.Reg)

	case obj.TYPE_REGREG:
		return C_PAIR

	case obj.TYPE_SHIFT:
		return C_SHIFT

	case obj.TYPE_REGLIST:
		return C_LIST

	case obj.TYPE_MEM:
		// The base register should be an integer register.
		if int16(REG_F0) <= a.Reg && a.Reg <= int16(REG_V31) {
			break
		}
		c.instoffset = a.Offset
		switch a.Name {
		case obj.NAME_EXTERN, obj.NAME_STATIC:
			if a.Sym == nil {
				break
			}
			// use relocation
			if a.Sym.Type == objabi.STLSBSS {
				if c.ctxt.Flag_shared {
					return C_TLS_IE
				} else {
					return C_TLS_LE
				}
			}
			return C_ADDR

		case obj.NAME_GOTREF:
			return C_GOTADDR

		case obj.NAME_AUTO:
			if a.Reg == REGSP {
				// unset base register for better printing, since
				// a.Offset is still relative to pseudo-FP.
				a.Reg = obj.REG_NONE
			}
			// The frame top 8 or 16 bytes are for FP
			c.instoffset = int64(c.autosize) + a.Offset - int64(c.extrasize)
			return memclass(c.instoffset)

		case obj.NAME_PARAM:
			if a.Reg == REGSP {
				// unset base register for better printing, since
				// a.Offset is still relative to pseudo-FP.
				a.Reg = obj.REG_NONE
			}
			c.instoffset = int64(c.autosize) + a.Offset + 8
			return memclass(c.instoffset)

		case obj.NAME_NONE:
			if a.Index != 0 {
				if a.Offset != 0 {
					if isRegShiftOrExt(a.Index) {
						// extended or shifted register offset, (Rn)(Rm.UXTW<<2) or (Rn)(Rm<<2).
						return C_ROFF
					}
					return C_GOK
				}
				// register offset, (Rn)(Rm)
				return C_ROFF
			}
			return memclass(a.Offset)
		}
		return C_GOK

	case obj.TYPE_FCONST:
		return C_FCON

	case obj.TYPE_TEXTSIZE:
		return C_TEXTSIZE

	case obj.TYPE_CONST:
		return c.con64class(a.Offset)

	case obj.TYPE_ADDR:
		c.instoffset = a.Offset
		switch a.Name {
		case obj.NAME_NONE:
			break

		case obj.NAME_EXTERN, obj.NAME_STATIC:
			if a.Sym == nil {
				return C_GOK
			}
			if a.Sym.Type == objabi.STLSBSS {
				c.ctxt.Diag("taking address of TLS variable is not supported")
			}
			return C_VCONADDR

		case obj.NAME_AUTO:
			// The original offset is relative to the pseudo SP,
			// adjust it to be relative to the RSP register.
			if a.Reg == REG_RSP {
				a.Reg = obj.REG_NONE
			}
			// The frame top 8 or 16 bytes are for FP
			c.instoffset = int64(c.autosize) + a.Offset - int64(c.extrasize)

		case obj.NAME_PARAM:
			// The original offset is relative to the pseudo FP,
			// adjust it to be relative to the RSP register.
			if a.Reg == REG_RSP {
				a.Reg = obj.REG_NONE
			}
			c.instoffset = int64(c.autosize) + a.Offset + 8
		default:
			return C_GOK
		}
		return C_LACON
	case obj.TYPE_BRANCH:
		return C_SBRA
	}

	return C_NONE
}

// newInst constructs a new Inst with the given arguments.
func newInst(as obj.As, idx uint16, pos src.XPos, args []obj.Oprd) obj.Inst {
	return obj.Inst{As: as, Optab: idx, Pos: pos, Args: args}
}

// progToInst constructs a new Inst with p.As, idx, p.Pos and args.
func progToInst(p *obj.Prog, idx uint16, args []obj.Oprd) []obj.Inst {
	return []obj.Inst{newInst(p.As, idx, p.Pos, args)}
}

// op2A returns the corresponding Inst of p which has two operands, and the
// operand order of p is exactly the opposite of that of the arm64 instruction.
func op2A(p *obj.Prog, idx uint16) []obj.Inst {
	return progToInst(p, idx, []obj.Oprd{p.To.ToOprd(), p.From.ToOprd()})
}

// op2ASO returns the corresponding Inst of p which has two operands, and the
// operand order of p is exactly the same as that of the arm64 instruction.
func op2ASO(p *obj.Prog, idx uint16) []obj.Inst {
	return progToInst(p, idx, []obj.Oprd{p.From.ToOprd(), p.To.ToOprd()})
}

// op2D3A returns the corresponding Inst of a three-operand Prog with two
// destination operands, such as STLXR instruction. The second destination
// operand is stored in p.RegTo2.
func op2D3A(p *obj.Prog, idx uint16) []obj.Inst {
	a := []obj.Oprd{{Reg: p.RegTo2, Type: obj.TYPE_REG}, p.From.ToOprd(), p.To.ToOprd()}
	return progToInst(p, idx, a)
}

// op2SA returns the corresponding Inst of a Prog with two sourc operands:
// opcode Rm(or $imm), Rn. The operand order of p must be exactly the opposite
// of the operand order of the corresponding arm64 instruction.
func op2SA(p *obj.Prog, idx uint16) []obj.Inst {
	a := []obj.Oprd{{Reg: p.Reg, Type: obj.TYPE_REG}, p.From.ToOprd()}
	return progToInst(p, idx, a)
}

// op3A returns the corresponding Inst of p which has three operands, and the
// operand order of p is exactly the opposite of that of the arm64 instruction.
func op3A(p *obj.Prog, idx uint16) []obj.Inst {
	// create an Addr for the second operand.
	r := p.Reg
	if r == 0 {
		r = p.To.Reg
	}
	a := []obj.Oprd{p.To.ToOprd(), {Reg: r, Type: obj.TYPE_REG}, p.From.ToOprd()}
	return progToInst(p, idx, a)
}

// op3S4A returns the corresponding Inst of a four-operand Prog with three
// source operands, such as MADD instruction. The three source operands are
// stored in p.RestArg3, p.From and p.Reg respectively.
func op3S4A(p *obj.Prog, idx uint16) []obj.Inst {
	a := []obj.Oprd{p.To.ToOprd(), p.GetFrom3().ToOprd(), p.From.ToOprd(), {Reg: p.Reg, Type: obj.TYPE_REG}}
	return progToInst(p, idx, a)
}

// op4A returns the corresponding Inst of a Prog with four operands:
// p.From, p.Reg, p.GetFrom3() and p.To. The argument order of
// the Go instruction is exactly the opposite of that of the arm64 instruction.
func op4A(p *obj.Prog, idx uint16) []obj.Inst {
	a := []obj.Oprd{p.To.ToOprd(), p.GetFrom3().ToOprd(), {Reg: p.Reg, Type: obj.TYPE_REG}, p.From.ToOprd()}
	return progToInst(p, idx, a)
}

// movcon64 moves the 64-bit constant con64 into the rto register by
// MOVZ/MOVN/MOVK instructions.
func (c *ctxt7) movcon64(con64 int64, rto int16, pos src.XPos) []obj.Inst {
	ctyp := c.con64class(con64)
	d := uint64(con64)
	dn := d
	as := AMOVZ
	zOrN := uint64(0)
	idx := MOVZxis
	if ctyp == C_MOVCONN {
		as = AMOVN
		zOrN = 0xffff
		dn = ^d
		idx = MOVNxis
	}
	rt := obj.Oprd{Reg: rto, Type: obj.TYPE_REG}
	i, imm := 0, uint64(0)
	for ; i < 4; i++ {
		imm = (dn >> uint(i*16)) & 0xffff
		if imm != 0 {
			break
		}
	}
	// Here Oprd.Index is used to store the shift offset of imm.
	args := []obj.Oprd{rt, {Offset: int64(imm), Type: obj.TYPE_CONST, Index: int16(i << 4)}}
	ret := []obj.Inst{newInst(as, idx, pos, args)}
	for i++; i < 4; i++ {
		imm = (d >> uint(i*16)) & 0xffff
		if imm != zOrN {
			args = []obj.Oprd{rt, {Offset: int64(imm), Type: obj.TYPE_CONST, Index: int16(i << 4)}}
			ret = append(ret, newInst(AMOVK, MOVKxis, pos, args))
		}
	}
	return ret
}

// movCon64ToReg moves the 64-bit constant con64 into the rto register by
// MOVZ/MOVN/MOVK instructions.
func (c *ctxt7) movCon64ToReg(con64 int64, rto int16, pos src.XPos) []obj.Inst {
	if movcon(con64) >= 0 {
		// -> MOVZ $imm, rto
		return []obj.Inst{instIR(AMOVD, con64, rto, MOVZxis, pos)}
	} else if movcon(^con64) >= 0 {
		// -> MOVN $imm, rto
		return []obj.Inst{instIR(AMOVD, ^con64, rto, MOVNxis, pos)}
	} else if isbitcon(uint64(con64)) {
		// -> MOV $imm, rto
		return []obj.Inst{instIR(AMOVD, con64, rto, MOVxi_b, pos)}
	}
	// Any other 64-bit integer constant.
	// -> MOVZ/MOVN $imm, rto + (MOVK $imm, rto)+
	return c.movcon64(con64, rto, pos)
}

// movcon32 moves the 32-bit constant con32 into the rto register by one MOVZW
// and one MOVKW instruction.
func (c *ctxt7) movcon32(con32 int64, rto int16, pos src.XPos) []obj.Inst {
	args := []obj.Oprd{{Reg: rto, Type: obj.TYPE_REG}, {Offset: con32 & 0xffff, Type: obj.TYPE_CONST}}
	movzw := newInst(AMOVZW, MOVZwis, pos, args)
	// Here Oprd.Index is used to store the shift number of the offset.
	args = []obj.Oprd{{Reg: rto, Type: obj.TYPE_REG}, {Offset: (con32 >> 16) & 0xffff, Type: obj.TYPE_CONST, Index: 16}}
	movkw := newInst(AMOVKW, MOVKwis, pos, args)
	return []obj.Inst{movzw, movkw}
}

// movCon32ToReg moves the 32-bit constant con32 into the rto register by
// MOVZW/MOVNW/MOVKW instructions. Note that the type of con32 is int64,
// which is for 32-bit BITCON checking.
func (c *ctxt7) movCon32ToReg(con32 int64, rto int16, pos src.XPos) []obj.Inst {
	v := uint32(con32)
	if movcon(int64(v)) >= 0 {
		// -> MOVZW $imm, rto
		return []obj.Inst{instIR(AMOVW, int64(v), rto, MOVZwis, pos)}
	} else if movcon(int64(^v)) >= 0 {
		// -> MOVNW $imm, rto
		return []obj.Inst{instIR(AMOVW, int64(^v), rto, MOVNwis, pos)}
	} else if isbitcon(uint64(con32)) {
		// -> MOVW $imm, rto
		return []obj.Inst{instIR(AMOVW, con32, rto, MOVwi_b, pos)}
	}
	// Any other 32-bit integer constant.
	// -> MOVZW $imm, rto + MOVKW $imm, rto+
	return c.movcon32(con32, rto, pos)
}

// size in log2(bytes)
func movesize(a obj.As) int {
	switch a {
	case AFMOVQ:
		return 4
	case AMOVD, AFMOVD:
		return 3
	case AMOVW, AMOVWU, AFMOVS:
		return 2
	case AMOVH, AMOVHU:
		return 1
	case AMOVB, AMOVBU:
		return 0
	default:
		return -1
	}
}

// bCond returns the corresponding conditon value of the B.cond instruction.
func (c *ctxt7) bCond(as obj.As) int64 {
	switch as {
	case ABEQ:
		return 0
	case ABNE:
		return 1
	case ABCS, ABHS:
		return 2
	case ABCC, ABLO:
		return 3
	case ABMI:
		return 4
	case ABPL:
		return 5
	case ABVS:
		return 6
	case ABVC:
		return 7
	case ABHI:
		return 8
	case ABLS:
		return 9
	case ABGE:
		return 10
	case ABLT:
		return 11
	case ABGT:
		return 12
	case ABLE:
		return 13
		// BAL and BNV are not supported.
	}
	return -1
}

// instLoad generates a Inst for a load instruction, such as MOVD, MOVW.
func instLoad(as obj.As, rf, rt, fromIndex int16, fromOffset int64, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Type: obj.TYPE_REG, Reg: rt}, {Type: obj.TYPE_MEM, Reg: rf, Offset: fromOffset, Index: fromIndex}}
	return newInst(as, idx, pos, args)
}

// instLoadPair generates a Inst for a LDP like instruction, such as LDP, LDPW.
func instLoadPair(as obj.As, rf, rt1, rt2 int16, fromOffset int64, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Type: obj.TYPE_REGREG, Reg: rt1, Offset: int64(rt2)}, {Type: obj.TYPE_MEM, Reg: rf, Offset: fromOffset}}
	return newInst(as, idx, pos, args)
}

// instStore generates a Inst for a store instruction, such as MOVD, MOVW.
func instStore(as obj.As, rf, rt, toIndex int16, toOffset int64, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Type: obj.TYPE_REG, Reg: rf}, {Type: obj.TYPE_MEM, Reg: rt, Offset: toOffset, Index: toIndex}}
	return newInst(as, idx, pos, args)
}

// instStorePair generates a Inst for a STP like instruction, such as STP, STPW.
func instStorePair(as obj.As, rt, rf1, rf2 int16, toOffset int64, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Type: obj.TYPE_REGREG, Reg: rf1, Offset: int64(rf2)}, {Type: obj.TYPE_MEM, Reg: rt, Offset: toOffset}}
	return newInst(as, idx, pos, args)
}

// adrp generates a Inst corresponding to the ADRP instruction.
// p.From.Offset is set to 0.
func adrp(rt int16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Reg: rt, Type: obj.TYPE_REG}, {Type: obj.TYPE_BRANCH, Offset: 0}}
	return newInst(AADRP, ADRPxl, pos, args)
}

// instIR generates a Inst with two parameters, and the first parameter
// is constant type, the second is register types. idx is the corresponding
// index of arm64 instruction in optab.
func instIR(as obj.As, imm int64, rt int16, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Reg: rt, Type: obj.TYPE_REG}, {Type: obj.TYPE_CONST, Offset: imm}}
	return newInst(as, idx, pos, args)
}

// instIRR generates a Inst with three parameters, and the first parameter
// is constant type, the second and third parameters are register types.
// idx is the corresponding index of arm64 instruction in optab.
func instIRR(as obj.As, imm int64, rn, rt int16, idx uint16, pos src.XPos) obj.Inst {
	if rn == 0 {
		rn = rt
	}
	args := []obj.Oprd{{Reg: rt, Type: obj.TYPE_REG}, {Reg: rn, Type: obj.TYPE_REG}, {Type: obj.TYPE_CONST, Offset: imm}}
	return newInst(as, idx, pos, args)
}

// inst2SR generates a Inst with two source parameters of register type.
// The first operand is p.From and the second operand is p.Reg.
// idx is the corresponding index of arm64 instruction in optab.
func inst2SR(as obj.As, rm, rn int16, idx uint16, pos src.XPos) obj.Inst {
	args := []obj.Oprd{{Reg: rn, Type: obj.TYPE_REG}, {Reg: rm, Type: obj.TYPE_REG}}
	return newInst(as, idx, pos, args)
}

// instRRR generates a Inst with three register type parameters.
// idx is the corresponding index of arm64 instruction in optab.
func instRRR(as obj.As, rm, rn, rt int16, idx uint16, pos src.XPos) obj.Inst {
	if rn == 0 {
		rn = rt
	}
	args := []obj.Oprd{{Reg: rt, Type: obj.TYPE_REG}, {Reg: rn, Type: obj.TYPE_REG}, {Reg: rm, Type: obj.TYPE_REG}}
	return newInst(as, idx, pos, args)
}

/* form offset parameter to SYS; special register number */
func SYSARG4(op1 int, Cn int, Cm int, op2 int) int {
	return op1<<16 | Cn<<12 | Cm<<8 | op2<<5
}

// checkUnpredictable checks whether p will trigger constrained unpredictable behavior.
func checkUnpredictable(isload bool, wback bool, rn int16, rt1 int16, rt2 int16) bool {
	if wback && rn != REGSP && (rn == rt1 || rn == rt2) || isload && rt1 == rt2 {
		return true
	}
	return false
}

// newRelocation returns a newly created relocation with some fields being set with the arguments.
func newRelocation(f *obj.LSym, size uint8, sym *obj.LSym, add int64, typ objabi.RelocType) uint16 {
	idx := obj.Addrel2(f)
	rel := &f.R[idx]
	rel.Siz = size
	rel.Sym = sym
	rel.Add = add
	rel.Type = typ
	return idx + 1
}

// addSub deals with ADD/ADDW/SUB/SUBW/ADDS/ADDSW/SUBS/SUBSW instructions.
// cidx, sidx and eidx are the indexs of the immediate, shifted register
// and extended register format instructions in the optab table, respectively.
func addSub(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16, setflag, is32bit bool, movToReg func(int64, int16, src.XPos) []obj.Inst) []obj.Inst {
	tc := c.aclass(p, &p.To)
	if tc != C_REG || p.RestArgs != nil || p.RegTo2 != 0 {
		return nil
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		v := p.From.Offset
		if is32bit {
			v = int64(uint32(v))
		}
		if isaddcon(v) {
			return op3A(p, cidx) // -> ADD/SUB(immediate)
		} else if !setflag && isaddcon2(v) {
			// -> ADD/SUB(immediate) + ADD/SUB(immediate)
			p1 := instIRR(p.As, v&0xfff, p.Reg, p.To.Reg, cidx, p.Pos)
			p2 := instIRR(p.As, v&0xfff000, p.To.Reg, p.To.Reg, cidx, p.Pos)
			p.Mark |= NOTUSETMP
			return []obj.Inst{p1, p2}
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return nil
		}
		// MOVD $con, Rtmp + ADD/SUB Rtmp, Rn, Rt
		p1 := instRRR(p.As, REGTMP, p.Reg, p.To.Reg, eidx, p.Pos)
		mov := movToReg(p.From.Offset, REGTMP, p.Pos)
		return append(mov, p1)
	case obj.TYPE_SHIFT:
		return op3A(p, sidx)
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		switch ftyp {
		case C_REG:
			if p.Reg == REG_RSP || p.To.Reg == REG_RSP {
				// When Rn or Rd is RSP, should encode with ADD/SUB (extended register).
				return op3A(p, eidx)
			}
			return op3A(p, sidx) // -> ADD/SUB(shift register)
		case C_EXTREG:
			return op3A(p, eidx) // -> ADD/SUB(extended register)
		}
	}
	return nil
}

// bitwiseOp deals with AND/ANDW/EOR/EORW/ORR/ORRW/BIC/BICW/EON/EONW/ORN/ORNW/ANDS/ANDSW/BICS/BICSW instructions.
func bitwiseOp(c *ctxt7, p *obj.Prog, cidx, sidx uint16, neg, supportZR bool, movToReg func(int64, int16, src.XPos) []obj.Inst) []obj.Inst {
	tc := c.aclass(p, &p.To)
	if tc != C_REG {
		return nil
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		v := p.From.Offset
		if isbitcon(uint64(v)) && supportZR {
			if neg {
				p.From.Offset = ^v
			}
			return op3A(p, cidx) // -> Op(immediate)
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return nil
		}
		// MOVD $con, Rtmp + p.As Rtmp, Rn, Rt
		p1 := instRRR(p.As, REGTMP, p.Reg, p.To.Reg, sidx, p.Pos)
		mov := movToReg(v, REGTMP, p.Pos)
		return append(mov, p1)
	case obj.TYPE_SHIFT, obj.TYPE_REG:
		return op3A(p, sidx) // -> Op(shift register)
	}
	return nil
}

// splitAddCon checks if the large offset v of a load/store instruction can be split into
// hi+lo parts of an ADD instruction, and both fit into the instruction. If so then we can
// deal with the constant with an ADD instruction first, for example:
// MOVD $imm(Rs), Ft -> ADD $hi, Rs, Rtmp + LDR lo(Rtmp), Ft
func splitAddCon(c *ctxt7, p *obj.Prog, v int32, s int) (int32, bool) {
	if v < 0 || (v&((1<<uint(s))-1)) != 0 {
		// negative or unaligned offset, use movX
		return 0, false
	}
	hi := v - (v & (0xFFF << uint(s)))
	if hi&0xFFF != 0 {
		c.ctxt.Diag("internal: miscalculated offset %d [%d]: %v", v, s, p)
		return 0, false
	}
	if hi&^0xFFF000 != 0 {
		// hi doesn't fit into an ADD instruction
		return 0, false
	}
	return hi, true
}

// immOffsetStore handles the store of addresses with immediate offset values.
func immOffsetStore(c *ctxt7, p *obj.Prog, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []obj.Inst, checkUnpredicate bool) []obj.Inst {
	a := p.To
	if a.Reg == obj.REG_NONE {
		a.Reg = REG_RSP // convert the pseudo register to RSP
	}
	if p.Scond == C_XPOST || p.Scond == C_XPRE {
		if checkUnpredicate && a.Reg != REG_RSP && p.From.Reg == a.Reg {
			c.ctxt.Diag("constrained unpredictable behavior: %v", p)
			return nil
		}
		if p.To.Reg == 0 { // pseudo registers, like FP, SP
			c.ctxt.Diag("pre and post index format don't support pseudo register: %v", p)
			return nil
		}
		if p.To.Offset < -256 || p.To.Offset > 255 {
			c.ctxt.Diag("offset out of range [-256,255]: %v", p)
			return nil
		}
		if p.Scond == C_XPOST {
			return op2ASO(p, pidx) // -> store (immediate post-index)
		}
		return op2ASO(p, widx) // -> store (immediate pre-index)
	}
	s := movesize(p.As)
	if s < 0 {
		c.ctxt.Diag("unexpected long move, %v", p)
		return nil
	}
	v := c.instoffset
	// fit one store instruction
	if v >= 0 && v <= (0xfff<<uint(s)) && (v&((1<<uint(s))-1) == 0) {
		return []obj.Inst{instStore(p.As, p.From.Reg, a.Reg, 0, v, uidx, p.Pos)}
	}

	// use Store Register (unscaled) instruction if -256 <= c.instoffset < 256
	if v >= -256 && v < 256 {
		return []obj.Inst{instStore(p.As, p.From.Reg, a.Reg, 0, v, idx256, p.Pos)}
	}
	// if offset v can be split into hi+lo, and both fit into instructions, convert
	// to ADD $hi, Rt, Rtmp + store Rs, lo(Rtmp)
	var p1, p2 obj.Inst
	v32 := int32(v)
	hi, ok := splitAddCon(c, p, v32, s)
	if !ok {
		goto storeusemov
	}
	p1 = instIRR(AADD, int64(hi), a.Reg, REGTMP, ADDxxis, p.Pos)
	p2 = instStore(p.As, p.From.Reg, REGTMP, 0, int64(v32-hi), uidx, p.Pos)
	return []obj.Inst{p1, p2}
storeusemov:
	// -> MOVZ/MOVN $imm, Rtmp + (MOVK $imm, Rtmp)+ + store R, (Rt)(Rtmp)
	if p.From.Reg == REGTMP || a.Reg == REGTMP {
		c.ctxt.Diag("REGTMP used in large offset store: %v", p)
		return nil
	}
	mov := movToReg(v, REGTMP, p.Pos)
	p1 = instStore(p.As, p.From.Reg, a.Reg, REGTMP, 0, eidx, p.Pos)
	return append(mov, p1)
}

// immOffsetStore handles the load of addresses with immediate offset values.
func immOffsetLoad(c *ctxt7, p *obj.Prog, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []obj.Inst, checkUnpredicate bool) []obj.Inst {
	a := p.From
	if a.Reg == obj.REG_NONE {
		a.Reg = REG_RSP // convert the pseudo register to RSP
	}
	if p.Scond == C_XPOST || p.Scond == C_XPRE {
		if checkUnpredicate && a.Reg != REG_RSP && a.Reg == p.To.Reg {
			c.ctxt.Diag("constrained unpredictable behavior: %v", p)
			return nil
		}
		if p.From.Reg == 0 { // pseudo registers, like FP, SP
			c.ctxt.Diag("pre and post index format don't support pseudo register: %v", p)
			return nil
		}
		if p.From.Offset < -256 || p.From.Offset > 255 {
			c.ctxt.Diag("offset out of range [-256,255]: %v", p)
			return nil
		}
		if p.Scond == C_XPOST {
			return op2A(p, pidx) // -> load (immediate post-index)
		}
		return op2A(p, widx) // -> load (immediate pre-index)
	}
	s := movesize(p.As)
	if s < 0 {
		c.ctxt.Diag("unexpected long move, %v", p)
		return nil
	}
	v := c.instoffset
	// fit one load instruction
	if v >= 0 && v <= (0xfff<<uint(s)) && (v&((1<<uint(s))-1) == 0) {
		return []obj.Inst{instLoad(p.As, a.Reg, p.To.Reg, 0, v, uidx, p.Pos)}
	}
	// use Load Register (unscaled) instruction if -256 <= v < 256
	if v >= -256 && v < 256 {
		return []obj.Inst{instLoad(p.As, a.Reg, p.To.Reg, 0, v, idx256, p.Pos)}
	}
	// if offset v can be split into hi+lo, and both fit into instructions, do
	// MOVD $imm(Rs), Rt -> ADD $hi, Rs, Rtmp + load lo(Rtmp), Rt
	var p1, p2 obj.Inst
	v32 := int32(v)
	hi, ok := splitAddCon(c, p, v32, s)
	if !ok {
		goto loadusemov
	}
	p1 = instIRR(AADD, int64(hi), a.Reg, REGTMP, ADDxxis, p.Pos)
	p2 = instLoad(p.As, REGTMP, p.To.Reg, 0, int64(v32-hi), uidx, p.Pos)
	return []obj.Inst{p1, p2}
loadusemov:
	// -> MOVZ/MOVN $imm, Rtmp + (MOVK $imm, Rtmp)+ + load (Rs)(Rtmp), R
	if a.Reg == REGTMP {
		c.ctxt.Diag("REGTMP used in large offset load: %v", p)
		return nil
	}
	mov := movToReg(v, REGTMP, p.Pos)
	p1 = instLoad(p.As, a.Reg, p.To.Reg, REGTMP, 0, eidx, p.Pos)
	return append(mov, p1)
}

// generalStore expands the store form of instructions MOVD, MOVW, etc.
// toTyp is the class of p.From, eidx, uidx, pidx, widx and idx256 are the indexs of
// the register, Unsigned offset, Post-index, Pre-index and unscaled format instructions
// in the optab table, respectively. movToReg is c.movCon64ToReg for 64-bit instructions
// and c.movCon32ToReg for 32-bit instructions.
func generalStore(c *ctxt7, p *obj.Prog, toTyp argtype, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []obj.Inst, checkUnpredicate bool) []obj.Inst {
	switch toTyp {
	case C_ROFF:
		return op2ASO(p, eidx) // -> store (register)
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + store (immediate)
		if p.From.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return nil
		}
		p.RelocIdx = newRelocation(c.cursym, 8, p.To.Sym, p.To.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instStore(p.As, p.From.Reg, REGTMP, 0, 0, uidx, p.Pos)
		return []obj.Inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		return immOffsetStore(c, p, eidx, uidx, pidx, widx, idx256, movToReg, checkUnpredicate)
	}
	return nil
}

// generalLoad expands the load form of instructions MOVD, MOVW, etc.
// The meaning of the parameter is the same as generalStore.
func generalLoad(c *ctxt7, p *obj.Prog, fromTyp argtype, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []obj.Inst, checkUnpredicate bool) []obj.Inst {
	switch fromTyp {
	case C_ROFF:
		return op2A(p, eidx) // -> load (register)
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + load (REGTMP), Rt
		p.RelocIdx = newRelocation(c.cursym, 8, p.From.Sym, p.From.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, uidx, p.Pos)
		return []obj.Inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		return immOffsetLoad(c, p, eidx, uidx, pidx, widx, idx256, movToReg, checkUnpredicate)
	}
	return nil
}

// loadPair expands the load form of instructions LDP, LDPW, etc.
// "shift" is the shift value of the offset value.
func loadPair(c *ctxt7, p *obj.Prog, fromTyp argtype, uidx, pidx, widx uint16, shift uint) []obj.Inst {
	switch fromTyp {
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + LDP (Rtmp), (Rt1, Rt2)
		p.RelocIdx = newRelocation(c.cursym, 8, p.From.Sym, p.From.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instLoadPair(p.As, REGTMP, p.To.Reg, int16(p.To.Offset), 0, uidx, p.Pos)
		return []obj.Inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		a := p.From
		if p.Scond == C_XPOST || p.Scond == C_XPRE {
			if a.Reg == obj.REG_NONE { // pseudo registers, like FP, SP
				c.ctxt.Diag("pre and post index format don't support pseudo register: %v", p)
				return nil
			}
			if a.Offset < -64<<shift || a.Offset > 63<<shift {
				c.ctxt.Diag("offset out of range [%d,%d]: %v", -64<<shift, 63<<shift, p)
				return nil
			}
			if a.Offset&(1<<shift-1) != 0 {
				c.ctxt.Diag("offset must be a multiple of %d: %v", 1<<shift, p)
				return nil
			}
			if p.Scond == C_XPOST {
				return op2A(p, pidx) // -> LDP (post-index)
			}
			return op2A(p, widx) // -> LDP (pre-index)
		}
		if a.Reg == obj.REG_NONE {
			a.Reg = REG_RSP // convert the pseudo register to RSP
		}
		v := c.instoffset
		// fit one LDP(signed offset) instruction
		if v >= -64<<shift && v <= 63<<shift && v&(1<<shift-1) == 0 {
			return []obj.Inst{instLoadPair(p.As, a.Reg, p.To.Reg, int16(p.To.Offset), v, uidx, p.Pos)}
		} else if isaddcon(v) || isaddcon(-v) {
			// -> ADD/SUB $imm, Rf, Rtmp + LDP (Rtmp), (Rt1, Rt2)
			as := AADD
			idx := ADDxxis
			if v < 0 {
				as = ASUB
				v = -v
				idx = SUBxxis
			}
			p1 := instIRR(as, v, a.Reg, REGTMP, idx, p.Pos)
			p2 := instLoadPair(p.As, REGTMP, p.To.Reg, int16(p.To.Offset), 0, uidx, p.Pos)
			return []obj.Inst{p1, p2}
		} else {
			// -> MOVZ/MOVN $imm, Rtmp + (MOVK $imm, Rtmp)+ + ADD Rtmp, Rf, Rtmp + LDP (Rtmp), (Rt1, Rt2)
			if a.Reg == REGTMP {
				c.ctxt.Diag("REGTMP used in large offset load: %v", p)
				return nil
			}
			mov := c.movCon64ToReg(v, REGTMP, p.Pos)
			p1 := instRRR(AADD, REGTMP, a.Reg, REGTMP, ADDxxre, p.Pos)
			p2 := instLoadPair(p.As, REGTMP, p.To.Reg, int16(p.To.Offset), 0, uidx, p.Pos)
			mov = append(mov, p1, p2)
			return mov
		}
	}
	return nil
}

// storePair expands the store form of instructions STP, STPW, etc.
func storePair(c *ctxt7, p *obj.Prog, toTyp argtype, uidx, pidx, widx uint16, shift uint) []obj.Inst {
	switch toTyp {
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + STP (Rt1, Rt2), (Rtmp)
		p.RelocIdx = newRelocation(c.cursym, 8, p.To.Sym, p.To.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instStorePair(p.As, REGTMP, p.From.Reg, int16(p.From.Offset), 0, uidx, p.Pos)
		return []obj.Inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		a := p.To
		if a.Reg == obj.REG_NONE {
			a.Reg = REG_RSP // convert the pseudo register to RSP
		}
		if p.Scond == C_XPOST || p.Scond == C_XPRE {
			if checkUnpredictable(false, true, a.Reg, p.From.Reg, int16(p.From.Offset)) {
				c.ctxt.Diag("constrained unpredictable behavior: %v", p)
				return nil
			}
			if p.To.Reg == 0 { // pseudo registers, like FP, SP
				c.ctxt.Diag("pre and post index format don't support pseudo register: %v", p)
				return nil
			}
			if a.Offset < -64<<shift || a.Offset > 63<<shift {
				c.ctxt.Diag("offset out of range [%d,%d]: %v", -64<<shift, 63<<shift, p)
				return nil
			}
			if a.Offset&(1<<shift-1) != 0 {
				c.ctxt.Diag("offset must be a multiple of %d: %v", 1<<shift, p)
				return nil
			}
			if p.Scond == C_XPOST {
				return op2ASO(p, pidx) // -> STP (post-index)
			}
			return op2ASO(p, widx) // -> STP (pre-index)
		}
		v := c.instoffset
		// fit one STP(signed offset) instruction
		if v >= -64<<shift && v <= 63<<shift && v&(1<<shift-1) == 0 {
			return []obj.Inst{instStorePair(p.As, a.Reg, p.From.Reg, int16(p.From.Offset), v, uidx, p.Pos)}
		} else if isaddcon(v) || isaddcon(-v) {
			// -> ADD/SUB $imm, Rt, Rtmp + STP (Rt1, Rt2), (Rtmp)
			as := AADD
			idx := ADDxxis
			if v < 0 {
				as = ASUB
				v = -v
				idx = SUBxxis
			}
			p1 := instIRR(as, v, a.Reg, REGTMP, idx, p.Pos)
			p2 := instStorePair(p.As, REGTMP, p.From.Reg, int16(p.From.Offset), 0, uidx, p.Pos)
			return []obj.Inst{p1, p2}
		} else {
			// -> MOVZ/MOVN $imm, Rtmp + (MOVK $imm, Rtmp)+ + ADD Rtmp, Rt, Rtmp + STP (Rt1, Rt2), (Rtmp)
			if a.Reg == REGTMP {
				c.ctxt.Diag("REGTMP used in large offset store: %v", p)
				return nil
			}
			mov := c.movCon64ToReg(v, REGTMP, p.Pos)
			p1 := instRRR(AADD, REGTMP, a.Reg, REGTMP, ADDxxre, p.Pos)
			p2 := instStorePair(p.As, REGTMP, p.From.Reg, int16(p.From.Offset), 0, uidx, p.Pos)
			mov = append(mov, p1, p2)
			return mov
		}
	}
	return nil
}

// cmpCmn deals with CMP and CMN instructions. cidx, sidx and eidx are the indexs of
// the immediate, shifted register and extended register format instructions in the
// optab table, respectively.
func cmpCmn(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16) []obj.Inst {
	if !(p.Reg != 0 && p.To.Type == obj.TYPE_NONE) {
		return nil
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if !(ftyp == C_REG || ftyp == C_EXTREG) {
			break
		}
		if ftyp == C_EXTREG || p.Reg == REG_RSP {
			return op2SA(p, eidx)
		}
		return op2SA(p, sidx)
	case obj.TYPE_SHIFT:
		return op2SA(p, sidx)
	case obj.TYPE_CONST:
		v := p.From.Offset
		if isaddcon(v) && p.Reg != REGZERO {
			return op2SA(p, cidx) // -> CMP(immediate)
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return nil
		}
		idx := sidx
		if p.Reg == REGSP {
			idx = eidx
		}
		// MOVD $con, Rtmp + CMP Rtmp, R
		p1 := inst2SR(p.As, REGTMP, p.Reg, idx, p.Pos)
		mov := c.movCon64ToReg(v, REGTMP, p.Pos)
		return append(mov, p1)
	}
	return nil
}

// cmpCmn32 deals with CMPW and CMNW instructions.
// The meaning of the parameter is the same as cmpCmn.
func cmpCmn32(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16) []obj.Inst {
	if !(p.Reg != 0 && p.To.Type == obj.TYPE_NONE) {
		return nil
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if !(ftyp == C_REG || ftyp == C_EXTREG) {
			break
		}
		if ftyp == C_EXTREG || p.Reg == REG_RSP {
			return op2SA(p, eidx)
		}
		return op2SA(p, sidx)
	case obj.TYPE_SHIFT:
		return op2SA(p, sidx)
	case obj.TYPE_CONST:
		v := uint32(p.From.Offset)
		if isaddcon(int64(v)) && p.Reg != REGZERO {
			return op2SA(p, cidx) // -> CMPW(immediate)
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return nil
		}
		idx := sidx
		if p.Reg == REGSP {
			idx = eidx
		}
		// MOVW $con, Rtmp + CMPW Rtmp, R
		p1 := inst2SR(p.As, REGTMP, p.Reg, idx, p.Pos)
		mov := c.movCon32ToReg(p.From.Offset, REGTMP, p.Pos)
		return append(mov, p1)
	}
	return nil
}

// opWithCarryIndex returns the index of the arm64 op-with-carry instruction corresponding
// to the opcode as in the optab.
func opWithCarryIndex(as obj.As) uint16 {
	switch as {
	case AADC:
		return ADCxxx
	case AADCW:
		return ADCwww
	case AADCS:
		return ADCSxxx
	case AADCSW:
		return ADCSwww
	case ASBC:
		return SBCxxx
	case ASBCW:
		return SBCwww
	case ASBCS:
		return SBCSxxx
	case ASBCSW:
		return SBCSwww
	case ANGC:
		return NGCxx
	case ANGCW:
		return NGCww
	case ANGCS:
		return NGCSxx
	case ANGCSW:
		return NGCSww
	}
	return 0
}

// bitFieldOpsIndex returns the index of bitfield manipulation instructions
// corresponding to the opcode as in the optab.
func bitFieldOpsIndex(as obj.As) uint16 {
	switch as {
	case ABFM:
		return BFMxxii
	case ABFMW:
		return BFMwwii
	case ABFI:
		return BFIxxii
	case ABFIW:
		return BFIwwii
	case ABFXIL:
		return BFXILxxii
	case ABFXILW:
		return BFXILwwii
	case ASBFM:
		return SBFMxxii
	case ASBFMW:
		return SBFMwwii
	case ASBFIZ:
		return SBFIZxxii
	case ASBFIZW:
		return SBFIZwwii
	case ASBFX:
		return SBFXxxii
	case ASBFXW:
		return SBFXwwii
	case AUBFM:
		return UBFMxxii
	case AUBFMW:
		return UBFMwwii
	case AUBFIZ:
		return UBFIZxxii
	case AUBFIZW:
		return UBFIZwwii
	case AUBFX:
		return UBFXxxii
	case AUBFXW:
		return UBFXwwii
	}
	return 0
}

// condCmpRegOpIndex returns the index of the arm64 conditional comparison
// instruction (register) corresponding to the opcode as in the optab.
func condCmpRegOpIndex(as obj.As) uint16 {
	switch as {
	case ACCMP:
		return CCMPxxic
	case ACCMPW:
		return CCMPwwic
	case ACCMN:
		return CCMNxxic
	case ACCMNW:
		return CCMNwwic
	}
	return 0
}

// condCmpImmOpIndex returns the index of the arm64 conditional comparison
// instruction (immediate) corresponding to the opcode as in the optab.
func condCmpImmOpIndex(as obj.As) uint16 {
	switch as {
	case ACCMP:
		return CCMPxiic
	case ACCMPW:
		return CCMPwiic
	case ACCMN:
		return CCMNxiic
	case ACCMNW:
		return CCMNwiic
	}
	return 0
}

// cselOpIndex returns the index of the arm64 conditional select/increase/negative etc.
// instruction corresponding to the opcode as in the optab.
func cselOpIndex(as obj.As) uint16 {
	switch as {
	case ACSEL:
		return CSELxxxc
	case ACSELW:
		return CSELwwwc
	case ACSINC:
		return CSINCxxxc
	case ACSINCW:
		return CSINCwwwc
	case ACSINV:
		return CSINVxxxc
	case ACSINVW:
		return CSINVwwwc
	case ACSNEG:
		return CSNEGxxxc
	case ACSNEGW:
		return CSNEGwwwc
	case AFCSELD:
		return FCSELdddc
	case AFCSELS:
		return FCSELsssc
	}
	return 0
}

// csetOpIndex returns the index of the arm64 conditional set series
// instruction corresponding to the opcode as in the optab.
func csetOpIndex(as obj.As) uint16 {
	switch as {
	case ACSET:
		return CSETxc
	case ACSETW:
		return CSETwc
	case ACSETM:
		return CSETMxc
	case ACSETMW:
		return CSETMwc
	}
	return 0
}

// CLSOpIndex returns the index of the arm64 CLS/CLZ series
// instruction corresponding to the opcode as in the optab.
func CLSOpIndex(as obj.As) uint16 {
	switch as {
	case ACLS:
		return CLSxx
	case ACLSW:
		return CLSww
	case ACLZ:
		return CLZxx
	case ACLZW:
		return CLZww
	case ARBIT:
		return RBITxx
	case ARBITW:
		return RBITww
	case AREV:
		return REVxx
	case AREVW:
		return REVww
	case AREV16:
		return REV16xx
	case AREV16W:
		return REV16ww
	case AREV32:
		return REV32xx
	}
	return 0
}

// opWithoutArgIndex returns the index of the arm64 instruction without
// arguments corresponding to the opcode as in the optab.
func opWithoutArgIndex(as obj.As) uint16 {
	switch as {
	case AERET:
		return ERET
	case AWFE:
		return WFE
	case AWFI:
		return WFI
	case AYIELD:
		return YIELD
	case ASEV:
		return SEV
	case ASEVL:
		return SEVL
	case ANOOP:
		return NOP
	case ADRPS:
		return DRPS
	}
	return 0
}

// loadAcquireIndex returns the index of the arm64 load acquire instruction
// corresponding to the opcode as in the optab.
func loadAcquireIndex(as obj.As) uint16 {
	switch as {
	case ALDAR:
		return LDARxx
	case ALDARW:
		return LDARwx
	case ALDARH:
		return LDARHwx
	case ALDARB:
		return LDARBwx
	case ALDAXR:
		return LDAXRxx
	case ALDAXRW:
		return LDAXRwx
	case ALDAXRH:
		return LDAXRHwx
	case ALDAXRB:
		return LDAXRBwx
	case ALDXR:
		return LDXRxx
	case ALDXRW:
		return LDXRwx
	case ALDXRH:
		return LDXRHwx
	case ALDXRB:
		return LDXRBwx
	}
	return 0
}

// shiftRegIndex returns the index of the arm64 shift (register) instruction
// corresponding to the opcode as in the optab.
func shiftRegIndex(as obj.As) uint16 {
	switch as {
	case ALSL:
		return LSLxxx
	case ALSLW:
		return LSLwww
	case ALSR:
		return LSRxxx
	case ALSRW:
		return LSRwww
	case AASR:
		return ASRxxx
	case AASRW:
		return ASRwww
	case AROR:
		return RORxxx
	case ARORW:
		return RORwww
	}
	return 0
}

// shiftImmIndex returns the index of the arm64 shift (imm) instruction
// corresponding to the opcode as in the optab.
func shiftImmIndex(as obj.As) uint16 {
	switch as {
	case ALSL:
		return LSLxxi
	case ALSLW:
		return LSLwwi
	case ALSR:
		return LSRxxi
	case ALSRW:
		return LSRwwi
	case AASR:
		return ASRxxi
	case AASRW:
		return ASRwwi
	case AROR:
		return RORxxi
	case ARORW:
		return RORwwi
	}
	return 0
}

// fuseOpIndex returns the index of the arm64 fuse instruction
// corresponding to the opcode as in the optab.
func fuseOpIndex(as obj.As) uint16 {
	switch as {
	case AMADD:
		return MADDxxxx
	case AMADDW:
		return MADDwwww
	case AMSUB:
		return MSUBxxxx
	case AMSUBW:
		return MSUBwwww
	case ASMADDL:
		return SMADDLxwwx
	case ASMSUBL:
		return SMSUBLxwwx
	case AUMADDL:
		return UMADDLxwwx
	case AUMSUBL:
		return UMSUBLxwwx
	case AFMADDD:
		return FMADDdddd
	case AFMADDS:
		return FMADDssss
	case AFMSUBD:
		return FMSUBdddd
	case AFMSUBS:
		return FMSUBssss
	case AFNMADDD:
		return FNMADDdddd
	case AFNMADDS:
		return FNMADDssss
	case AFNMSUBD:
		return FNMSUBdddd
	case AFNMSUBS:
		return FNMSUBssss
	}
	return 0
}

// multipleOpIndex returns the index of the arm64 multiple series instruction
// corresponding to the opcode as in the optab.
func multipleOpIndex(as obj.As) uint16 {
	switch as {
	case AMUL:
		return MULxxx
	case AMULW:
		return MULwww
	case AMNEG:
		return MNEGxxx
	case AMNEGW:
		return MNEGwww
	case ASMNEGL:
		return SMNEGLxww
	case AUMNEGL:
		return UMNEGLxww
	case ASMULH:
		return SMULHxxx
	case ASMULL:
		return SMULLxww
	case AUMULH:
		return UMULHxxx
	case AUMULL:
		return UMULLxww
	}
	return 0
}

// divCRCIndex returns the index of the arm64 SDIV and CRC series instruction
// corresponding to the opcode as in the optab.
func divCRCIndex(as obj.As) uint16 {
	switch as {
	case ASDIV:
		return SDIVxxx
	case ASDIVW:
		return SDIVwww
	case AUDIV:
		return UDIVxxx
	case AUDIVW:
		return UDIVwww
	case ACRC32B:
		return CRC32Bwww
	case ACRC32H:
		return CRC32Hwww
	case ACRC32W:
		return CRC32Wwww
	case ACRC32X:
		return CRC32Xwwx
	case ACRC32CB:
		return CRC32CBwww
	case ACRC32CH:
		return CRC32CHwww
	case ACRC32CW:
		return CRC32CWwww
	case ACRC32CX:
		return CRC32CXwwx
	}
	return 0
}

// extendIndex returns the index of the arm64 extend series instruction
// corresponding to the opcode as in the optab.
func extendIndex(as obj.As) uint16 {
	switch as {
	case ASXTB:
		return SXTBxw
	case ASXTBW:
		return SXTBww
	case ASXTH:
		return SXTHxw
	case ASXTHW:
		return SXTHww
	case ASXTW:
		return SXTWxw
	case AUXTBW:
		return UXTBww
	case AUXTHW:
		return UXTHww
	}
	return 0
}

// atomicIndex returns the index of the arm64 atomic instruction corresponding
// to the opcode as in the optab.
func atomicIndex(as obj.As) uint16 {
	switch as {
	case ACASD:
		return CASxxx
	case ACASW:
		return CASwwx
	case ACASH:
		return CASHwwx
	case ACASB:
		return CASBwwx
	case ACASAD:
		return CASAxxx
	case ACASAW:
		return CASAwwx
	case ACASALD:
		return CASALxxx
	case ACASALW:
		return CASALwwx
	case ACASALH:
		return CASALHwwx
	case ACASALB:
		return CASALBwwx
	case ACASLD:
		return CASLxxx
	case ACASLW:
		return CASLwwx
	case ACASPD:
		return CASPxxx
	case ACASPW:
		return CASPwwx
	case ALDADDD:
		return LDADDxxx
	case ALDADDW:
		return LDADDwwx
	case ALDADDH:
		return LDADDHwwx
	case ALDADDB:
		return LDADDBwwx
	case ALDADDAD:
		return LDADDAxxx
	case ALDADDAW:
		return LDADDAwwx
	case ALDADDAH:
		return LDADDAHwwx
	case ALDADDAB:
		return LDADDABwwx
	case ALDADDALD:
		return LDADDALxxx
	case ALDADDALW:
		return LDADDALwwx
	case ALDADDALH:
		return LDADDALHwwx
	case ALDADDALB:
		return LDADDALBwwx
	case ALDADDLD:
		return LDADDLxxx
	case ALDADDLW:
		return LDADDLwwx
	case ALDADDLH:
		return LDADDLHwwx
	case ALDADDLB:
		return LDADDLBwwx
	case ALDCLRD:
		return LDCLRxxx
	case ALDCLRW:
		return LDCLRwwx
	case ALDCLRH:
		return LDCLRHwwx
	case ALDCLRB:
		return LDCLRBwwx
	case ALDCLRAD:
		return LDCLRAxxx
	case ALDCLRAW:
		return LDCLRAwwx
	case ALDCLRAH:
		return LDCLRAHwwx
	case ALDCLRAB:
		return LDCLRABwwx
	case ALDCLRALD:
		return LDCLRALxxx
	case ALDCLRALW:
		return LDCLRALwwx
	case ALDCLRALH:
		return LDCLRALHwwx
	case ALDCLRALB:
		return LDCLRALBwwx
	case ALDCLRLD:
		return LDCLRLxxx
	case ALDCLRLW:
		return LDCLRLwwx
	case ALDCLRLH:
		return LDCLRLHwwx
	case ALDCLRLB:
		return LDCLRLBwwx
	case ALDEORD:
		return LDEORxxx
	case ALDEORW:
		return LDEORwwx
	case ALDEORH:
		return LDEORHwwx
	case ALDEORB:
		return LDEORBwwx
	case ALDEORAD:
		return LDEORAxxx
	case ALDEORAW:
		return LDEORAwwx
	case ALDEORAH:
		return LDEORAHwwx
	case ALDEORAB:
		return LDEORABwwx
	case ALDEORALD:
		return LDEORALxxx
	case ALDEORALW:
		return LDEORALwwx
	case ALDEORALH:
		return LDEORALHwwx
	case ALDEORALB:
		return LDEORALBwwx
	case ALDEORLD:
		return LDEORLxxx
	case ALDEORLW:
		return LDEORLwwx
	case ALDEORLH:
		return LDEORLHwwx
	case ALDEORLB:
		return LDEORLBwwx
	case ALDORD:
		return LDSETxxx
	case ALDORW:
		return LDSETwwx
	case ALDORH:
		return LDSETHwwx
	case ALDORB:
		return LDSETBwwx
	case ALDORAD:
		return LDSETAxxx
	case ALDORAW:
		return LDSETAwwx
	case ALDORAH:
		return LDSETAHwwx
	case ALDORAB:
		return LDSETABwwx
	case ALDORALD:
		return LDSETALxxx
	case ALDORALW:
		return LDSETALwwx
	case ALDORALH:
		return LDSETALHwwx
	case ALDORALB:
		return LDSETALBwwx
	case ALDORLD:
		return LDSETLxxx
	case ALDORLW:
		return LDSETLwwx
	case ALDORLH:
		return LDSETLHwwx
	case ALDORLB:
		return LDSETLBwwx
	case ASWPD:
		return SWPxxx
	case ASWPW:
		return SWPwwx
	case ASWPH:
		return SWPHwwx
	case ASWPB:
		return SWPBwwx
	case ASWPAD:
		return SWPAxxx
	case ASWPAW:
		return SWPAwwx
	case ASWPAH:
		return SWPAHwwx
	case ASWPAB:
		return SWPABwwx
	case ASWPALD:
		return SWPALxxx
	case ASWPALW:
		return SWPALwwx
	case ASWPALH:
		return SWPALHwwx
	case ASWPALB:
		return SWPALBwwx
	case ASWPLD:
		return SWPLxxx
	case ASWPLW:
		return SWPLwwx
	case ASWPLH:
		return SWPLHwwx
	case ASWPLB:
		return SWPLBwwx
	}
	return 0
}

// floatingOpIndex returns the index of some floating point instructions corresponding
// to the opcode as in the optab.
func floatingOpIndex(as obj.As) uint16 {
	switch as {
	case AFADDD:
		return FADDddd
	case AFADDS:
		return FADDsss
	case AFSUBD:
		return FSUBddd
	case AFSUBS:
		return FSUBsss
	case AFMULD:
		return FMULddd
	case AFMULS:
		return FMULsss
	case AFNMULD:
		return FNMULddd
	case AFNMULS:
		return FNMULsss
	case AFDIVD:
		return FDIVddd
	case AFDIVS:
		return FDIVsss
	case AFMAXD:
		return FMAXddd
	case AFMAXS:
		return FMAXsss
	case AFMIND:
		return FMINddd
	case AFMINS:
		return FMINsss
	case AFMAXNMD:
		return FMAXNMddd
	case AFMAXNMS:
		return FMAXNMsss
	case AFMINNMD:
		return FMINNMddd
	case AFMINNMS:
		return FMINNMsss
	}
	return 0
}

// fcmpRegIndex returns the index of the arm64 floating point compare (register)
// instruction corresponding to the opcode as in the optab.
func fcmpRegIndex(as obj.As) uint16 {
	switch as {
	case AFCMPD:
		return FCMPdd
	case AFCMPS:
		return FCMPss
	case AFCMPED:
		return FCMPEdd
	case AFCMPES:
		return FCMPEss
	}
	return 0
}

// fcmpImmIndex returns the index of the arm64 floating point compare (immediate)
// instruction corresponding to the opcode as in the optab.
func fcmpImmIndex(as obj.As) uint16 {
	switch as {
	case AFCMPD:
		return FCMPd0
	case AFCMPS:
		return FCMPs0
	case AFCMPED:
		return FCMPEd0
	case AFCMPES:
		return FCMPEs0
	}
	return 0
}

// fConvertRoundingIndex returns the index of the arm64 floating point conversion
// and rounding instructions corresponding to the opcode as in the optab.
func fConvertRoundingIndex(as obj.As) uint16 {
	switch as {
	case AFCVTDH:
		return FCVThd
	case AFCVTDS:
		return FCVTsd
	case AFCVTHD:
		return FCVTdh
	case AFCVTHS:
		return FCVTsh
	case AFCVTSD:
		return FCVTds
	case AFCVTSH:
		return FCVThs
	case AFABSD:
		return FABSdd
	case AFABSS:
		return FABSss
	case AFNEGD:
		return FNEGdd
	case AFNEGS:
		return FNEGss
	case AFSQRTD:
		return FSQRTdd
	case AFSQRTS:
		return FSQRTss
	case AFRINTAD:
		return FRINTAdd
	case AFRINTAS:
		return FRINTAss
	case AFRINTID:
		return FRINTIdd
	case AFRINTIS:
		return FRINTIss
	case AFRINTMD:
		return FRINTMdd
	case AFRINTMS:
		return FRINTMss
	case AFRINTND:
		return FRINTNdd
	case AFRINTNS:
		return FRINTNss
	case AFRINTPD:
		return FRINTPdd
	case AFRINTPS:
		return FRINTPss
	case AFRINTXD:
		return FRINTXdd
	case AFRINTXS:
		return FRINTXss
	case AFRINTZD:
		return FRINTZdd
	case AFRINTZS:
		return FRINTZss
	case AFCVTZSD:
		return FCVTZSxd
	case AFCVTZSDW:
		return FCVTZSwd
	case AFCVTZSS:
		return FCVTZSxs
	case AFCVTZSSW:
		return FCVTZSws
	case AFCVTZUD:
		return FCVTZUxd
	case AFCVTZUDW:
		return FCVTZUwd
	case AFCVTZUS:
		return FCVTZUxs
	case AFCVTZUSW:
		return FCVTZUws
	case ASCVTFD:
		return SCVTFdx
	case ASCVTFS:
		return SCVTFsx
	case ASCVTFWD:
		return SCVTFdw
	case ASCVTFWS:
		return SCVTFsw
	case AUCVTFD:
		return UCVTFdx
	case AUCVTFS:
		return UCVTFsx
	case AUCVTFWD:
		return UCVTFdw
	case AUCVTFWS:
		return UCVTFsw
	}
	return 0
}

// cryptoOpIndex returns the index of some crypto instructions corresponding
// to the opcode as in the optab.
func cryptoOpIndex(as obj.As) uint16 {
	switch as {
	case AAESD:
		return AESDvv
	case AAESE:
		return AESEvv
	case AAESIMC:
		return AESIMCvv
	case AAESMC:
		return AESMCvv
	case ASHA1SU1:
		return SHA1SU1vv
	case ASHA256SU0:
		return SHA256SU0vv
	case ASHA512SU0:
		return SHA512SU0vv
	case ASHA1C:
		return SHA1Cqsv
	case ASHA1P:
		return SHA1Pqsv
	case ASHA1M:
		return SHA1Mqsv
	case ASHA256H:
		return SHA256Hqqv
	case ASHA256H2:
		return SHA256H2qqv
	case ASHA512H:
		return SHA512Hqqv
	case ASHA512H2:
		return SHA512H2qqv
	}
	return 0
}

// cryptoOpIndex returns the index of some crypto instructions corresponding
// to the opcode as in the optab.
func arng3Index(as obj.As) uint16 {
	switch as {
	case ASHA1SU0:
		return SHA1SU0vvv
	case ASHA256SU1:
		return SHA256SU1vvv
	case ASHA512SU1:
		return SHA512SU1vvv
	case AVRAX1:
		return RAX1vvv
	case AVADDP:
		return ADDPvvv_t
	case AVAND:
		return ANDvvv_t
	case AVORR:
		return ORRvvv_t
	case AVEOR:
		return EORvvv_t
	case AVBIF:
		return BIFvvv_t
	case AVBIT:
		return BITvvv_t
	case AVBSL:
		return BSLvvv_t
	case AVUMAX:
		return UMAXvvv_t
	case AVUMIN:
		return UMINvvv_t
	case AVUZP1:
		return UZP1vvv_t
	case AVUZP2:
		return UZP2vvv_t
	case AVFMLA:
		return FMLAvvv_t
	case AVFMLS:
		return FMLSvvv_t
	case AVZIP1:
		return ZIP1vvv_t
	case AVZIP2:
		return ZIP2vvv_t
	}
	return 0
}

// asm functions

func asmCall(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	switch tc {
	case C_SBRA:
		// DUFFCOPY/DUFFZERO/BL label -> BL <label>
		p.Insts = progToInst(p, BLl, []obj.Oprd{p.To.ToOprd()})
		p.Mark |= BRANCH26BITS
		if p.To.Sym != nil {
			p.RelocIdx = newRelocation(c.cursym, 4, p.To.Sym, p.To.Offset, objabi.R_CALLARM64)
		}
	case C_REG, C_ZOREG:
		// BL Rn -> BLR <Xn>
		// BL (Rn) or BL 0(Rn) -> BLR <Xn>
		p.Insts = progToInst(p, BLRx, []obj.Oprd{p.To.ToOprd()})
		p.RelocIdx = newRelocation(c.cursym, 0, nil, 0, objabi.R_CALLIND)
	}
}

func asmJMP(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	switch tc {
	case C_SBRA:
		// B label -> B <label>
		p.Insts = progToInst(p, Bl, []obj.Oprd{p.To.ToOprd()})
		p.Mark |= BRANCH26BITS
		if p.To.Sym != nil {
			p.RelocIdx = newRelocation(c.cursym, 4, p.To.Sym, p.To.Offset, objabi.R_CALLARM64)
		}
	case C_REG, C_ZOREG:
		// B Rn -> BR <Xn>
		// B (Rn) or BL 0(Rn) -> BR <Xn>
		p.Insts = progToInst(p, BRx, []obj.Oprd{p.To.ToOprd()})
	}
}

func asmRET(c *ctxt7, p *obj.Prog) {
	// RET -> RET {<Xn>}
	p.Insts = progToInst(p, RETx, []obj.Oprd{p.To.ToOprd()})
}

func asmUNDEF(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.To.Type == obj.TYPE_NONE) {
		return
	}
	p.Insts = progToInst(p, UDFi, nil)
}

// adc/adcs/sbc/sbcs
func asmOpWithCarry(c *ctxt7, p *obj.Prog) {
	fc := c.aclass(p, &p.From)
	tc := c.aclass(p, &p.To)
	if !(fc == C_REG && tc == C_REG) {
		return
	}
	p.Insts = op3A(p, opWithCarryIndex(p.As))
}

// ngc/ngcs
func asmNGCX(c *ctxt7, p *obj.Prog) {
	fc := c.aclass(p, &p.From)
	tc := c.aclass(p, &p.To)
	if !(fc == C_REG && p.Reg == 0 && tc == C_REG) {
		return
	}
	p.Insts = op2A(p, opWithCarryIndex(p.As))
}

// ADD/ADDW/SUB/SUBW/ADDS/ADDSW/SUBS/SUBSW
func asmADD(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, ADDxxis, ADDxxxs, ADDxxre, false, false, c.movCon64ToReg)
}

func asmADDW(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, ADDwwis, ADDwwws, ADDwwwe, false, true, c.movCon32ToReg)
}

func asmSUB(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, SUBxxis, SUBxxxs, SUBxxre, false, false, c.movCon64ToReg)
}

func asmSUBW(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, SUBwwis, SUBwwws, SUBwwwe, false, true, c.movCon32ToReg)
}

func asmADDS(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, ADDSxxis, ADDSxxxs, ADDSxxre, true, false, c.movCon64ToReg)
}

func asmADDSW(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, ADDSwwis, ADDSwwws, ADDSwwwe, true, true, c.movCon32ToReg)
}

func asmSUBS(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, SUBSxxis, SUBSxxxs, SUBSxxre, true, false, c.movCon64ToReg)
}

func asmSUBSW(c *ctxt7, p *obj.Prog) {
	p.Insts = addSub(c, p, SUBSwwis, SUBSwwws, SUBSwwwe, true, true, c.movCon32ToReg)
}

// asmADRX deals with ADR and ADRP instructions.
func asmADRX(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	if !(tc == C_REG && p.From.Type == obj.TYPE_BRANCH) {
		return
	}
	if p.As == AADR {
		p.Insts = op2A(p, ADRxl)
	} else if p.As == AADRP {
		p.Insts = op2A(p, ADRPxl)
	} else {
		c.ctxt.Diag("invalid opcode: %v", p)
	}
}

// AND/ANDW/EOR/EORW/ORR/ORRW/BIC/BICW/EON/EONW/ORN/ORNW/ANDS/ANDSW/BICS/BICSW
func asmAND(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDxxi, ANDxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmANDW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDwwi, ANDwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmEOR(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, EORxxi, EORxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmEORW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, EORwwi, EORwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmORR(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ORRxxi, ORRxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmORRW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ORRwwi, ORRwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmBIC(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDxxi, BICxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmBICW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDwwi, BICwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmEON(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, EORxxi, EONxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmEONW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, EORwwi, EONwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmORN(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ORRxxi, ORNxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func asmORNW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ORRwwi, ORNwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func asmANDS(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDSxxi, ANDSxxxs, false, true, c.movCon64ToReg)
}

func asmANDSW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDSwwi, ANDSwwws, false, true, c.movCon32ToReg)
}

func asmBICS(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDSxxi, BICSxxxs, true, true, c.movCon64ToReg)
}

func asmBICSW(c *ctxt7, p *obj.Prog) {
	p.Insts = bitwiseOp(c, p, ANDSwwi, BICSwwws, true, true, c.movCon32ToReg)
}

// TST/TSTW
func asmTST(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	if tc != C_NONE {
		return
	}
	cidx, sidx, movConToReg := TSTxi, TSTxxs, c.movCon64ToReg
	if p.As == ATSTW {
		cidx, sidx, movConToReg = TSTwi, TSTwws, c.movCon32ToReg
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		v := p.From.Offset
		if isbitcon(uint64(v)) {
			p.Insts = op2SA(p, cidx) // -> TST/TSTW(immediate)
			return
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v\n", p)
			return
		}
		// MOVD $con, Rtmp + TST/TSTW Rtmp, Rt
		p1 := inst2SR(p.As, REGTMP, p.Reg, sidx, p.Pos)
		mov := movConToReg(v, REGTMP, p.Pos)
		mov = append(mov, p1)
		p.Insts = mov
	case obj.TYPE_SHIFT, obj.TYPE_REG:
		p.Insts = op2SA(p, sidx) // -> TST/TSTW(shift register)
	}
}

// asmBitFieldOps deals with bitfield operation instructions, such as BFM, BFI, etc.
func asmBitFieldOps(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	tc := c.aclass(p, &p.To)
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && tc == C_REG && from3 != nil && (from3.Type == obj.TYPE_CONST || from3.Reg == REGZERO)) {
		return
	}
	switch p.As {
	case ABFXIL, ABFXILW, ASBFX, ASBFXW, AUBFX, AUBFXW:
		// Save p.From.Offset in from3.Index because encoding from3 requires the value.
		from3.Index = int16(p.From.Offset)
	}
	r := p.Reg
	if r == 0 {
		r = p.To.Reg
	}
	a := []obj.Oprd{p.To.ToOprd(), {Reg: r, Type: obj.TYPE_REG}, p.From.ToOprd(), from3.ToOprd()}
	p.Insts = progToInst(p, bitFieldOpsIndex(p.As), a)
}

// load & store
func asmMOVD(c *ctxt7, p *obj.Prog) {
	// MOVD can be translated into several different kinds of instructions,
	// including MOV, LDR, STR, MSR etc.
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_SPR {
			p.Insts = op2A(p, MSRi) // -> MSR (immediate)
			break
		}
		if ttyp != C_REG {
			break
		}
		v := p.From.Offset
		if isbitcon(uint64(v)) && p.To.Reg != REGZERO {
			p.Insts = op2A(p, MOVxi_b) // -> MOV $bitcon, R
		} else if movcon(v) >= 0 {
			p.Insts = op2A(p, MOVZxis) // -> MOVZ $imm, R
		} else if movcon(^v) >= 0 {
			a := []obj.Oprd{p.To.ToOprd(), {Type: obj.TYPE_CONST, Offset: ^v}}
			p.Insts = progToInst(p, MOVNxis, a) // -> MOVN $imm, R
		} else { // Any other 64-bit integer constant.
			// -> MOVZ/MOVN $imm, R + (MOVK $imm, R)+
			mov := c.movcon64(v, p.To.Reg, p.Pos)
			p.Insts = mov
			p.Mark |= NOTUSETMP
		}
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		ttyp := c.aclass(p, &p.To)
		if ftyp == C_SPR && ttyp == C_REG {
			p.Insts = op2A(p, MRSx) // -> MRS
			break
		}
		if ftyp != C_REG {
			break
		}
		if ttyp == C_SPR {
			p.Insts = op2A(p, MSRx) // -> MSR (register)
			break
		}
		if ttyp == C_REG {
			if p.From.Reg == REG_RSP || p.To.Reg == REG_RSP {
				p.Insts = op2A(p, MOVxx_sp) // MOV (to/from SP)
			} else {
				p.Insts = op2A(p, MOVxx) // MOV (register)
			}
			break
		}
		// Store
		p.Insts = generalStore(c, p, ttyp, STRxxre, STRxx, STRxxi_p, STRxx_w, STURxx, c.movCon64ToReg, true)
	case obj.TYPE_ADDR:
		ftyp := c.aclass(p, &p.From)
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		a, v := p.From, c.instoffset
		if a.Reg == obj.REG_NONE {
			a.Reg = REG_RSP
		}
		if ftyp == C_VCONADDR {
			// -> ADRP + ADD + reloc
			p.RelocIdx = newRelocation(c.cursym, 8, p.From.Sym, v, objabi.R_ADDRARM64)
			p1 := adrp(p.To.Reg, p.Pos)
			p2 := instIRR(AADD, 0, 0, p.To.Reg, ADDxxis, p.Pos)
			p.Insts = []obj.Inst{p1, p2}
			p.Mark |= NOTUSETMP
			break
		}
		if ftyp != C_LACON {
			break
		}
		// MOVD $offset(Rf), Rt -> ADD/SUB $offset, Rf, Rt
		// Create a temporary Prog to compute p.Insts.
		q := c.newprog()
		q.Reg = a.Reg
		q.To = p.To
		if v < 0 {
			q.As = ASUB
			q.From = obj.Addr{Type: obj.TYPE_CONST, Offset: -v}
			p.Insts = addSub(c, q, SUBxxis, SUBxxxs, SUBxxre, false, false, c.movCon64ToReg)
		} else {
			q.As = AADD
			q.From = obj.Addr{Type: obj.TYPE_CONST, Offset: v}
			p.Insts = addSub(c, q, ADDxxis, ADDxxxs, ADDxxre, false, false, c.movCon64ToReg)
		}
		if q.Mark&NOTUSETMP != 0 {
			p.Mark |= NOTUSETMP
		}
	case obj.TYPE_MEM:
		// Load
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		switch ftyp {
		case C_GOTADDR:
			// MOVD sym@GOT, Rt -> ADRP + LDR (REGTMP), Rt + relocs
			p.RelocIdx = newRelocation(c.cursym, 8, p.From.Sym, 0, objabi.R_ARM64_GOTPCREL)
			p1 := adrp(REGTMP, p.Pos)
			p2 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, LDRxx, p.Pos)
			p.Insts = []obj.Inst{p1, p2}
		case C_TLS_LE:
			// LE model MOVD $tlsvar, Rt -> MOVZ + reloc
			if p.From.Offset != 0 {
				c.ctxt.Diag("invalid offset on MOVD $tlsvar")
				return
			}
			p.RelocIdx = newRelocation(c.cursym, 4, p.From.Sym, 0, objabi.R_ARM64_TLS_LE)
			p.Insts = progToInst(p, MOVZxis, []obj.Oprd{p.To.ToOprd(), {Type: obj.TYPE_CONST}})
		case C_TLS_IE:
			// IE model MOVD $tlsvar, Rt -> ADRP + LDR (REGTMP), Rt + relocs
			if p.From.Offset != 0 {
				c.ctxt.Diag("invalid offset on MOVD $tlsvar")
				return
			}
			p.RelocIdx = newRelocation(c.cursym, 8, p.From.Sym, 0, objabi.R_ARM64_TLS_IE)
			p1 := adrp(REGTMP, p.Pos)
			p2 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, LDRxx, p.Pos)
			p.Insts = []obj.Inst{p1, p2}
		default:
			p.Insts = generalLoad(c, p, ftyp, LDRxxre, LDRxx, LDRxxi_p, LDRxx_w, LDURxx, c.movCon64ToReg, true)
		}
	}
}

func asmMOVW(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		if isbitcon(uint64(p.From.Offset)) && p.To.Reg != REGZERO {
			p.Insts = op2A(p, MOVwi_b) // -> MOVW $bitcon, R
			break
		}
		v := uint32(p.From.Offset)
		if movcon(int64(v)) >= 0 {
			p.Insts = op2A(p, MOVZwis) // -> MOVZW $imm, R
		} else if movcon(int64(^v)) >= 0 {
			a := []obj.Oprd{p.To.ToOprd(), {Type: obj.TYPE_CONST, Offset: int64(^v)}}
			p.Insts = progToInst(p, MOVNwis, a) // -> MOVNW $imm, R
		} else { // Any other 32-bit integer constant.
			// -> MOVZW $imm, R + MOVKW $imm, R
			p.Insts = c.movcon32(p.From.Offset, p.To.Reg, p.Pos)
			p.Mark |= NOTUSETMP
		}
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			p.Insts = op2A(p, SXTWxw) // -> SXTW
			break
		}
		// Store
		p.Insts = generalStore(c, p, ttyp, STRwxre, STRwx, STRwxi_p, STRwx_w, STURwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRSWxxre, LDRSWxx, LDRSWxxi_p, LDRSWxx_w, LDURSWxx, c.movCon32ToReg, true)
	}
}

func asmMOVWU(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	// For moving the 32-bit immediate value to a register, use MOVW instead, because arm64 doesn't
	// have a mov instruction that can move the sign-extented immediate value to a register.
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			if p.From.Reg == REG_RSP || p.To.Reg == REG_RSP {
				p.Insts = op2A(p, MOVww_sp) // -> MOVW (to/from SP)
			} else {
				p.Insts = op2A(p, MOVww) // -> MOVW (register)
			}
			return
		}
		// Store, same as MOVW
		p.Insts = generalStore(c, p, ttyp, STRwxre, STRwx, STRwxi_p, STRwx_w, STURwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRwxre, LDRwx, LDRwxi_p, LDRwx_w, LDURwx, c.movCon32ToReg, true)
	}
}

func asmMOVH(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			p.Insts = op2A(p, SXTHxw) // -> SXTH
			break
		}
		// Store
		p.Insts = generalStore(c, p, ttyp, STRHwxre, STRHwx, STRHwxi_p, STRHwx_w, STURHwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRSHxxre, LDRSHxx, LDRSHxxi_p, LDRSHxx_w, LDURSHxx, c.movCon32ToReg, true)
	}
}

func asmMOVHU(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			p.Insts = op2A(p, UXTHww) // -> UXTH
			break
		}
		// Store, same as MOVH
		p.Insts = generalStore(c, p, ttyp, STRHwxre, STRHwx, STRHwxi_p, STRHwx_w, STURHwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRHwxre, LDRHwx, LDRHwxi_p, LDRHwx_w, LDURHwx, c.movCon32ToReg, true)
	}
}

func asmMOVB(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			p.Insts = op2A(p, SXTBxw) // -> SXTB
			break
		}
		// Store
		p.Insts = generalStore(c, p, ttyp, STRBwxre, STRBwx, STRBwxi_p, STRBwx_w, STURBwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRSBxxre, LDRSBxx, LDRSBxxi_p, LDRSBxx_w, LDURSBxx, c.movCon32ToReg, true)
	}
}

func asmMOVBU(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			p.Insts = op2A(p, UXTBww) // -> UXTB
			break
		}
		// Store, same as MOVB
		p.Insts = generalStore(c, p, ttyp, STRBwxre, STRBwx, STRBwxi_p, STRBwx_w, STURBwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		p.Insts = generalLoad(c, p, ftyp, LDRBwxre, LDRBwx, LDRBwxi_p, LDRBwx_w, LDURBwx, c.movCon32ToReg, true)
	}
}

// floating point load/store
func asmFMOVQ(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_FREG:
		// store
		switch ttyp {
		case C_ROFF:
			p.Insts = op2ASO(p, STRqxre) // -> STR (register, SIMD&FP)
		case C_ZOREG, C_LOREG, C_VOREG:
			p.Insts = immOffsetStore(c, p, STRqxre, STRqx, STRqxi_p, STRqx_w, STURqx, c.movCon64ToReg, false)
		}
	case C_ROFF:
		if ttyp != C_FREG {
			break
		}
		p.Insts = op2A(p, LDRqxre) // -> LDR (register, SIMD&FP)
	case C_ZOREG, C_LOREG, C_VOREG:
		// Load
		if ttyp != C_FREG {
			break
		}
		p.Insts = immOffsetLoad(c, p, LDRqxre, LDRqx, LDRqxi_p, LDRqx_w, LDURqx, c.movCon64ToReg, false)
	}
}

func asmFMOVD(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_FCON:
		if ttyp != C_FREG {
			break
		}
		p.Insts = op2A(p, FMOVdi)
	case C_REG:
		if ttyp != C_FREG {
			break
		}
		p.Insts = op2A(p, FMOVdx)
	case C_FREG:
		switch ttyp {
		case C_REG:
			p.Insts = op2A(p, FMOVxd)
		case C_FREG:
			p.Insts = op2A(p, FMOVdd)
		default:
			// Store
			p.Insts = generalStore(c, p, ttyp, STRdxre, STRdx, STRdxi_p, STRdx_w, STURdx, c.movCon64ToReg, false)
		}
	default:
		// Load
		if ttyp != C_FREG {
			break
		}
		p.Insts = generalLoad(c, p, ftyp, LDRdxre, LDRdx, LDRdxi_p, LDRdx_w, LDURdx, c.movCon64ToReg, false)
	}
}

func asmFMOVS(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_FCON:
		if ttyp != C_FREG {
			break
		}
		p.Insts = op2A(p, FMOVsi)
	case C_REG:
		if ttyp != C_FREG {
			break
		}
		p.Insts = op2A(p, FMOVsw)
	case C_FREG:
		switch ttyp {
		case C_REG:
			p.Insts = op2A(p, FMOVws)
		case C_FREG:
			p.Insts = op2A(p, FMOVss)
		default:
			// Store
			p.Insts = generalStore(c, p, ttyp, STRsxre, STRsx, STRsxi_p, STRsx_w, STURsx, c.movCon32ToReg, false)
		}
	default:
		// Load
		if ttyp != C_FREG {
			break
		}
		p.Insts = generalLoad(c, p, ftyp, LDRsxre, LDRsx, LDRsxi_p, LDRsx_w, LDURsx, c.movCon32ToReg, false)
	}
}

// load & store pair
func asmLDP(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPxxx, LDPxxxi_p, LDPxxx_w, 3)
}

func asmLDPW(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPwwx, LDPwwxi_p, LDPwwx_w, 2)
}

func asmLDPSW(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPSWxxx, LDPSWxxxi_p, LDPSWxxx_w, 2)
}

func asmFLDPQ(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPqqx, LDPqqxi_p, LDPqqx_w, 4)
}

func asmFLDPD(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPddx, LDPddxi_p, LDPddx_w, 3)
}

func asmFLDPS(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	p.Insts = loadPair(c, p, ftyp, LDPssx, LDPssxi_p, LDPssx_w, 2)
}

func asmSTP(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	p.Insts = storePair(c, p, ttyp, STPxxx, STPxxxi_p, STPxxx_w, 3)
}

func asmSTPW(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	p.Insts = storePair(c, p, ttyp, STPwwx, STPwwxi_p, STPwwx_w, 2)
}

func asmFSTPQ(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	p.Insts = storePair(c, p, ttyp, STPqqx, STPqqxi_p, STPqqx_w, 4)
}

func asmFSTPD(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	p.Insts = storePair(c, p, ttyp, STPddx, STPddxi_p, STPddx_w, 3)
}

func asmFSTPS(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	p.Insts = storePair(c, p, ttyp, STPssx, STPssxi_p, STPssx_w, 2)
}

// asmCBZX deals with CBZ/CBZW/CBNZ/CBNZW instructions.
func asmCBZX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_SBRA) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ACBZ:
		idx = CBZxl
	case ACBZW:
		idx = CBZwl
	case ACBNZ:
		idx = CBNZxl
	case ACBNZW:
		idx = CBNZwl
	}
	p.Mark |= BRANCH19BITS
	p.Insts = op2ASO(p, idx)
}

// asmCondCMP deals with CCMP, CCMPW, CCMN and CCMNW instructions.
func asmCondCMP(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	from3 := p.GetFrom3()
	if !(ftyp == C_COND && p.Reg != 0 && from3 != nil && (p.To.Type == obj.TYPE_CONST || p.To.Reg == REGZERO)) {
		return
	}
	idx := uint16(0)
	switch from3.Type {
	case obj.TYPE_REG:
		idx = condCmpRegOpIndex(p.As)
	case obj.TYPE_CONST:
		idx = condCmpImmOpIndex(p.As)
	default:
		return
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{{Reg: p.Reg, Type: obj.TYPE_REG}, from3.ToOprd(), p.To.ToOprd(), p.From.ToOprd()})
}

// asmCSELX deals with CSEL, CSINC, CSINV and CSNEG series instructions.
func asmCSELX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	from3 := p.GetFrom3()
	if !(ftyp == C_COND && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	idx := cselOpIndex(p.As)
	p.Insts = progToInst(p, idx, []obj.Oprd{p.To.ToOprd(), {Reg: p.Reg, Type: obj.TYPE_REG}, from3.ToOprd(), p.From.ToOprd()})
}

// asmCSETX deals with CSET/CSETW/CSETM/CSETMW instructions.
func asmCSETX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_COND && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op2A(p, csetOpIndex(p.As))
}

// asmCINCX deals with CINC/CINV/CNEG series instructions.
func asmCINCX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_COND && p.Reg != 0 && p.To.Type == obj.TYPE_REG && p.RestArgs == nil && p.RegTo2 == 0) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ACINC:
		idx = CINCxxc
	case ACINCW:
		idx = CINCwwc
	case ACINV:
		idx = CINVxxc
	case ACINVW:
		idx = CINVwwc
	case ACNEG:
		idx = CNEGxxc
	case ACNEGW:
		idx = CNEGwwc
	}
	p.Insts = op3A(p, idx)
}

// asmFloatingCondCMP deals with FCCMPD/FCCMPED series instructions.
func asmFloatingCondCMP(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	from3 := p.GetFrom3()
	if !(ftyp == C_COND && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && (p.To.Type == obj.TYPE_CONST || p.To.Reg == REGZERO)) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AFCCMPD:
		idx = FCCMPddic
	case AFCCMPS:
		idx = FCCMPssic
	case AFCCMPED:
		idx = FCCMPEddic
	case AFCCMPES:
		idx = FCCMPEssic
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{from3.ToOprd(), {Reg: p.Reg, Type: obj.TYPE_REG}, p.To.ToOprd(), p.From.ToOprd()})
}

//
func asmCLREX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && (p.To.Type == obj.TYPE_CONST || p.To.Reg == REGZERO || p.To.Type == obj.TYPE_NONE)) {
		return
	}
	p.Insts = progToInst(p, CLREXi, []obj.Oprd{p.To.ToOprd()})
}

// asmCLSX deals with CLS/CLZ series instructions.
func asmCLSX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_REG && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op2A(p, CLSOpIndex(p.As))
}

// cmp/cmn
func asmCMP(c *ctxt7, p *obj.Prog) {
	p.Insts = cmpCmn(c, p, CMPxis, CMPxxs, CMPxre)
}

func asmCMPW(c *ctxt7, p *obj.Prog) {
	p.Insts = cmpCmn32(c, p, CMPwis, CMPwws, CMPwwe)
}

func asmCMN(c *ctxt7, p *obj.Prog) {
	p.Insts = cmpCmn(c, p, CMNxis, CMNxxs, CMNxre)
}

func asmCMNW(c *ctxt7, p *obj.Prog) {
	p.Insts = cmpCmn32(c, p, CMNwis, CMNwws, CMNwwe)
}

// asmDMBX deals with DMB/DSB/ISB series instructions.
func asmDMBX(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ADMB:
		idx = DMBi
	case ADSB:
		idx = DSBi
	case AISB:
		idx = ISBi
	case AHINT:
		idx = HINTi
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{p.From.ToOprd()})
}

// asmOpwithoutArg deals with ERET/WFE/WFI series instructions that have no arguments.
func asmOpwithoutArg(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	p.Insts = progToInst(p, opWithoutArgIndex(p.As), nil)
}

// asmEXTRX deals with EXTR and EXTRW instructions.
func asmEXTRX(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AEXTR:
		idx = EXTRxxxi
	case AEXTRW:
		idx = EXTRwwwi
	default:
		c.ctxt.Diag("invalid opcode: %v", p)
		return
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{p.To.ToOprd(), from3.ToOprd(), {Reg: p.Reg, Type: obj.TYPE_REG}, p.From.ToOprd()})
}

// asmLoadAcquire deals with LDAR/LDAXR/LDXR series instructions.
func asmLoadAcquire(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_ZOREG && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op2A(p, loadAcquireIndex(p.As))
}

// asmLDXPX deals with LDXP/LDAXP series instructions.
func asmLDXPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_ZOREG && p.Reg == 0 && ttyp == C_PAIR) {
		return
	}
	if int(p.To.Reg) == int(p.To.Offset) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	idx := uint16(0)
	switch p.As {
	case ALDXP:
		idx = LDXPxxx
	case ALDXPW:
		idx = LDXPwwx
	case ALDAXP:
		idx = LDAXPxxx
	case ALDAXPW:
		idx = LDAXPwwx
	}
	p.Insts = op2A(p, idx)
}

// asmSTLRX deals with STLR series instructions.
func asmSTLRX(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_REG && p.Reg == 0 && ttyp == C_ZOREG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ASTLR:
		idx = STLRxx
	case ASTLRW:
		idx = STLRwx
	case ASTLRH:
		idx = STLRHwx
	case ASTLRB:
		idx = STLRBwx
	}
	p.Insts = op2ASO(p, idx)
}

// asmSTXRX deals with STLXR/STXR series instructions.
func asmSTXRX(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_REG && p.Reg == 0 && ttyp == C_ZOREG && p.RegTo2 != 0) {
		return
	}
	if p.RegTo2 == p.From.Reg || (p.RegTo2 == p.To.Reg && p.To.Reg != REGSP) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	idx := uint16(0)
	switch p.As {
	case ASTXR:
		idx = STXRwxx
	case ASTXRW:
		idx = STXRwwx
	case ASTXRH:
		idx = STXRHwwx
	case ASTXRB:
		idx = STXRBwwx
	case ASTLXR:
		idx = STLXRwxx
	case ASTLXRW:
		idx = STLXRwwx
	case ASTLXRH:
		idx = STLXRHwwx
	case ASTLXRB:
		idx = STLXRBwwx
	}
	p.Insts = op2D3A(p, idx)
}

// asmSTXPX deals with STXP/STLXP series instructions.
func asmSTXPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_PAIR && p.Reg == 0 && ttyp == C_ZOREG && p.RegTo2 != 0) {
		return
	}
	if (p.RegTo2 == p.From.Reg || p.RegTo2 == int16(p.From.Offset)) || (p.RegTo2 == p.To.Reg && p.To.Reg != REGSP) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	idx := uint16(0)
	switch p.As {
	case ASTXP:
		idx = STXPwxxx
	case ASTXPW:
		idx = STXPwwwx
	case ASTLXP:
		idx = STLXPwxxx
	case ASTLXPW:
		idx = STLXPwwwx
	}
	p.Insts = op2D3A(p, idx)
}

// asmShiftOp deals with LSL/LSR/ASR/ROR series instructions.
func asmShiftOp(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_REG || p.From.Type == obj.TYPE_CONST) && p.To.Type == obj.TYPE_REG) {
		return
	}
	// For LSL (immediate), imms != 0b111111, that means the shift number can't be 0.
	// Luckily we have converted $0 to ZR, so we won't encounter this problem.
	if p.From.Type == obj.TYPE_REG {
		p.Insts = op3A(p, shiftRegIndex(p.As))
	} else {
		p.Insts = op3A(p, shiftImmIndex(p.As))
	}
}

// asmFuseOp deals with MADD/MSUB/SMADDL/UMADDL/FMADD/FMSUB/FNMADD/FNMSUB series instructions.
func asmFuseOp(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	if !(p.From.Type == obj.TYPE_REG && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op3S4A(p, fuseOpIndex(p.As))
}

// asmMultipleX deals with MUL/MNEG/SMNEGL/UMNEGL/SMULH/SMULL/UMULH/UMULL series instructions.
func asmMultipleX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op3A(p, multipleOpIndex(p.As))
}

// asmMOVX deals with MOVK/MOVN/MOVZ series instructions.
func asmMOVX(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AMOVK:
		idx = MOVKxis
	case AMOVKW:
		idx = MOVKwis
	case AMOVN:
		idx = MOVNxis
	case AMOVNW:
		idx = MOVNwis
	case AMOVZ:
		idx = MOVZxis
	case AMOVZW:
		idx = MOVZwis
	}
	p.Insts = op2A(p, idx)
}

// mrs
func asmMRS(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_SPR && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	p.Insts = op2A(p, MRSx)
}

// msr
func asmMSR(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(p.Reg == 0 && ttyp == C_SPR) {
		return
	}
	if ftyp == C_REG {
		p.Insts = op2A(p, MSRx)
	} else if p.From.Type == obj.TYPE_CONST {
		p.Insts = op2A(p, MSRi)
	}
}

// asmMVN deals with MVN and MVNW instructions.
func asmMVN(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !((ftyp == C_REG || ftyp == C_SHIFT) && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AMVN:
		idx = MVNxxs
	case AMVNW:
		idx = MVNwws
	}
	p.Insts = op2A(p, idx)
}

// asmNEGX deals with NEG and NEGS series instructions.
func asmNEGX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !((ftyp == C_REG || ftyp == C_SHIFT || ftyp == C_NONE) && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ANEG:
		idx = NEGxxs
	case ANEGW:
		idx = NEGwws
	case ANEGS:
		idx = NEGSxxs
	case ANEGSW:
		idx = NEGSwws
	}
	p.Insts = op2A(p, idx)
}

// prfm
func asmPRFM(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.Reg == 0 && (ttyp == C_SPR || p.To.Type == obj.TYPE_CONST)) {
		return
	}
	ftyp := c.aclass(p, &p.From)
	switch ftyp {
	case C_ZOREG, C_LOREG:
		p.Insts = op2A(p, PRFMix)
	case C_SBRA:
		p.Mark |= BRANCH19BITS
		p.Insts = op2A(p, PRFMil)
	case C_ROFF:
		p.Insts = op2A(p, PRFMixre)
	}
}

// asmREMX deals with REM and UREM series instructions.
func asmREMX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && ttyp == C_REG) {
		return
	}
	// REM Rf, <Rn, > Rt -> SDIV Rf, Rn, Rtmp + MSUB Rf, Rn, Rtmp, Rt
	if p.From.Reg == REGTMP || p.Reg == REGTMP {
		c.ctxt.Diag("cannot use REGTMP as source: %v", p)
		return
	}
	r := p.Reg
	if r == 0 {
		r = p.To.Reg
	}
	p1As, p1Idx := obj.As(0), uint16(0)
	p2As, p2Idx := obj.As(0), uint16(0)
	switch p.As {
	case AREM:
		p1As, p1Idx = ASDIV, SDIVxxx
		p2As, p2Idx = AMSUB, MSUBxxxx
	case AREMW:
		p1As, p1Idx = ASDIVW, SDIVwww
		p2As, p2Idx = AMSUBW, MSUBwwww
	case AUREM:
		p1As, p1Idx = AUDIV, UDIVxxx
		p2As, p2Idx = AMSUB, MSUBxxxx
	case AUREMW:
		p1As, p1Idx = AUDIVW, UDIVwww
		p2As, p2Idx = AMSUBW, MSUBwwww
	}
	p1 := newInst(p1As, p1Idx, p.Pos, []obj.Oprd{{Reg: REGTMP, Type: obj.TYPE_REG}, {Reg: r, Type: obj.TYPE_REG}, p.From.ToOprd()})
	p2Args := []obj.Oprd{{Reg: p.To.Reg, Type: obj.TYPE_REG}, {Reg: REGTMP, Type: obj.TYPE_REG},
		{Reg: p.From.Reg, Type: obj.TYPE_REG}, {Reg: r, Type: obj.TYPE_REG}}
	p2 := newInst(p2As, p2Idx, p.Pos, p2Args)
	p.Insts = []obj.Inst{p1, p2}
}

// asmDIVCRC deals with SDIV and CRC series instructions.
func asmDIVCRC(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && ttyp == C_REG) {
		return
	}
	p.Insts = op3A(p, divCRCIndex(p.As))
}

// asmPE deals with SVC, HVC, HLT, SMC, BRK, DCPS{1, 2, 3} etc. instructions.
func asmPE(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO || p.From.Type == obj.TYPE_NONE) && p.Reg == 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ASVC:
		idx = SVCi
	case AHVC:
		idx = HVCi
	case AHLT:
		idx = HLTi
	case ASMC:
		idx = SMCi
	case ABRK:
		idx = BRKi
	case ADCPS1:
		idx = DCPS1i
	case ADCPS2:
		idx = DCPS2i
	case ADCPS3:
		idx = DCPS3i
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{p.From.ToOprd()})
}

// asmExtend deals with SXTB, SXTH series extend instructions.
func asmExtend(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	p.Insts = op2A(p, extendIndex(p.As))
}

// asmUnsignedExtend deals with UXTB, UXTH and UXTW instructions.
func asmUnsignedExtend(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	off := int64(0)
	switch p.As {
	case AUXTB:
		off = 7 // UXTB Rn, Rd -> UBFM $0, Rn, $7, Rd
	case AUXTH:
		off = 15 // UXTH Rn, Rd -> UBFM $0, Rn, $15, Rd
	case AUXTW:
		off = 31 // UXTW Rn, Rd -> UBFM $0, Rn, $31, Rd
	default:
		c.ctxt.Diag("unexpected opcode: %v", p)
		return
	}
	args := []obj.Oprd{p.To.ToOprd(), {Reg: p.From.Reg, Type: obj.TYPE_REG},
		{Type: obj.TYPE_CONST, Offset: 0}, {Type: obj.TYPE_CONST, Offset: off}}
	p.Insts = []obj.Inst{newInst(AUBFM, bitFieldOpsIndex(p.As), p.Pos, args)}
}

// sys
func asmSYS(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && (p.To.Type == obj.TYPE_NONE || p.To.Type == obj.TYPE_REG)) {
		return
	}
	if (p.From.Offset &^ int64(SYSARG4(0x7, 0xF, 0xF, 0x7))) != 0 {
		c.ctxt.Diag("illegal SYS argument: %v", p)
		return
	}
	// p.From.Offset integrated op1, Cn, Cm, Op2, in order to encode these arguments, p.From.Offset
	// needs to be split into multiple obj.Addr arguments.
	op1 := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 16) & 0x7}
	cn := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 12) & 0xf}
	cm := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 8) & 0xf}
	// op2{, <Xt>}, integrate p.To.Reg into op2 because there's only one argument in optab for them.
	op2 := obj.Oprd{Type: obj.TYPE_CONST, Index: p.To.Reg, Offset: (p.From.Offset >> 5) & 0x7}
	p.Insts = progToInst(p, SYSix, []obj.Oprd{op1, cn, cm, op2})
}

// sysl
func asmSYSL(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	if (p.From.Offset &^ int64(SYSARG4(0x7, 0xF, 0xF, 0x7))) != 0 {
		c.ctxt.Diag("illegal SYSL argument: %v", p)
		return
	}
	// p.From.Offset integrated op1, Cn, Cm, Op2, in order to encode these arguments, p.From.Offset
	// needs to be split into multiple obj.Addr arguments.
	xt := obj.Oprd{Type: obj.TYPE_REG, Reg: p.To.Reg}
	op1 := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 16) & 0x7}
	cn := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 12) & 0xf}
	cm := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 8) & 0xf}
	op2 := obj.Oprd{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 5) & 0x7}
	p.Insts = progToInst(p, SYSLix, []obj.Oprd{xt, op1, cn, cm, op2})
}

// asmSYSAlias deals with SYS alias instructions such as AT, DC, IC and TLBI.
func asmSYSAlias(c *ctxt7, p *obj.Prog) {
	// TODO: The existence of the destination register is based on p.From.Offset.
	// But we have not double-checked the value at the moment.
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && (p.To.Type == obj.TYPE_NONE || p.To.Type == obj.TYPE_REG)) {
		return
	}
	idx, fields := uint16(0), int64(0)
	switch p.As {
	case AAT:
		idx, fields = ATx, 0x7<<16|0<<12|1<<8|0x7<<5
	case ADC:
		idx, fields = DCx, 0x7<<16|0<<12|0xF<<8|0x7<<5
	case AIC:
		idx, fields = ICix, 0x7<<16|0<<12|0xF<<8|0x7<<5
	case ATLBI:
		idx, fields = TLBIix, 0x7<<16|0<<12|0xF<<8|0x7<<5
	}
	if (p.From.Offset &^ fields) != 0 {
		c.ctxt.Diag("illegal system instruction argument: %v", p)
		return
	}
	op := p.From.ToOprd()
	// integrate Xt into op because there is only
	// one corresponding argument in optab.
	op.Index = p.To.Reg
	p.Insts = progToInst(p, idx, []obj.Oprd{op})
}

// asmTBZX deals with TBZ and TBNZ instructions.
func asmTBZX(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg != 0 && p.To.Type == obj.TYPE_BRANCH) {
		return
	}
	p.Mark |= BRANCH14BITS
	idx := TBZril
	if p.As == ATBNZ {
		idx = TBNZril
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{{Reg: p.Reg, Type: obj.TYPE_REG}, p.From.ToOprd(), p.To.ToOprd()})
}

// asmAtomicLoadOpStore deals with LDADD, LDEOR series atomic instructions.
func asmAtomicLoadOpStore(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	t2typ := rclass(p.RegTo2)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_ZOREG && t2typ == C_REG) {
		return
	}
	args := []obj.Oprd{p.From.ToOprd(), {Reg: p.RegTo2, Type: obj.TYPE_REG}, p.To.ToOprd()}
	p.Insts = progToInst(p, atomicIndex(p.As), args)
}

// asmAtomicLoadOpStore deals with CASP series atomic instructions.
func asmCASPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	to2 := p.GetTo2()
	t2typ := c.aclass(p, to2)
	if !(ftyp == C_PAIR && p.Reg == 0 && ttyp == C_ZOREG && t2typ == C_PAIR) {
		return
	}
	args := []obj.Oprd{p.From.ToOprd(), to2.ToOprd(), p.To.ToOprd()}
	p.Insts = progToInst(p, atomicIndex(p.As), args)
}

// asmBCOND deals with B.<cond> series instructions.
func asmBCOND(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && p.To.Type == obj.TYPE_BRANCH) {
		return
	}
	p.Mark |= BRANCH19BITS
	args := []obj.Oprd{{Offset: c.bCond(p.As)}, p.To.ToOprd()}
	p.Insts = progToInst(p, Bcl, args)
}

// asmFloatingOp deals with some floating point instructions such as FADD, FSUB etc.
func asmFloatingOp(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && ttyp == C_FREG) {
		return
	}
	p.Insts = op3A(p, floatingOpIndex(p.As))
}

// asmFCMPX deals with FCMPD, FCMPED series instructions.
func asmFCMPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(p.Reg != 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	if ftyp == C_FREG {
		p.Insts = op2SA(p, fcmpRegIndex(p.As))
	} else if ftyp == C_FCON {
		p.Insts = op2SA(p, fcmpImmIndex(p.As))
	} else {
		c.ctxt.Diag("illegal combination: %v", p)
	}
}

// asmFConvertRounding deals with floating point conversion and rounding series instructions.
func asmFConvertRounding(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && p.Reg == 0 && ttyp == C_FREG) {
		return
	}
	p.Insts = op2A(p, fConvertRoundingIndex(p.As))
}

// asmFConvertToFixed deals with floating point convert to fixed-point series instructions.
func asmFConvertToFixed(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	p.Insts = op2A(p, fConvertRoundingIndex(p.As))
}

// asmFixedToFloating deals with fixed-point convert to floating point series instructions.
func asmFixedToFloating(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_FREG) {
		return
	}
	p.Insts = op2A(p, fConvertRoundingIndex(p.As))
}

// asmAESSHA deals with some AES and SHA instructions that support
// 'opcode VREG, VREG' and 'opcode ARNG, ARNG' formats.
func asmAESSHA(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	af, at := int((p.From.Reg>>5)&15), int((p.To.Reg>>5)&15)
	// support the C_VREG format for compatibility with old code.
	if !((ftyp == C_VREG && ttyp == C_VREG || ftyp == C_ARNG && ttyp == C_ARNG && af == at) && p.Reg == 0) {
		return
	}
	p.Insts = op2A(p, cryptoOpIndex(p.As))
}

// asmFBitwiseOp deals with some floating point bitwise operation instructions such as VREV16 and VCNT.
func asmFBitwiseOp(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	af, at := int((p.From.Reg>>5)&15), int((p.To.Reg>>5)&15)
	if !(ftyp == C_ARNG && ttyp == C_ARNG && af == at && p.Reg == 0) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AVREV16:
		idx = REV16vv_t
	case AVREV32:
		idx = REV32vv_t
	case AVREV64:
		idx = REV64vv_t
	case AVCNT:
		idx = CNTvv_t
	case AVRBIT:
		idx = RBITvv_t
	}
	p.Insts = op2A(p, idx)
}

// sha1h
func asmSHA1H(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_VREG && ttyp == C_VREG && p.Reg == 0) {
		return
	}
	from, to := p.From.ToOprd(), p.To.ToOprd()
	// SHA1H uses the V register in the Go assembly syntax, but in fact it uses the
	// floating point register in the Arm assembly syntax.
	from.Reg = p.From.Reg - REG_V0 + REG_F0
	to.Reg = p.To.Reg - REG_V0 + REG_F0
	p.Insts = progToInst(p, SHA1Hss, []obj.Oprd{to, from})
}

// asmSHAX deals with some SHA algorithm related instructions.
func asmSHAX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	// SHA1C once supported the "ASHA1C C_VREG, C_REG, C_VREG" format, but it is obviously
	// very strange. Do we need to continue to support it for compatibility ?
	if !(ftyp == C_ARNG && f2typ == C_VREG && ttyp == C_VREG) {
		return
	}
	from, to := p.From.ToOprd(), p.To.ToOprd()
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax for encoding.
	to.Reg = p.To.Reg - REG_V0 + REG_F0
	p.Insts = progToInst(p, cryptoOpIndex(p.As), []obj.Oprd{to, {Type: obj.TYPE_REG, Reg: p.Reg - REG_V0 + REG_F0}, from})
}

// asmARNG3 deals with instructions with three ARNG format operands.
func asmARNG3(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	af, an, at := int((p.From.Reg>>5)&15), int((p.Reg>>5)&15), int((p.To.Reg>>5)&15)
	if !(ftyp == C_ARNG && f2typ == C_ARNG && ttyp == C_ARNG && af == an && an == at) {
		return
	}
	p.Insts = op3A(p, arng3Index(p.As))
}

// asmVPMULLX deals with VPMULL and VPMULL2 instructions.
func asmVPMULLX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	tb, tb2, ta := (p.From.Reg>>5)&15, (p.Reg>>5)&15, (p.To.Reg>>5)&15
	if !(ftyp == C_ARNG && f2typ == C_ARNG && ttyp == C_ARNG && tb == tb2) {
		return
	}
	idx, Q := PMULLvvv_t, int16(0)
	if p.As == AVPMULL2 {
		idx, Q = PMULL2vvv_t, 1
	}
	if !sizeTaMatchTb2(ta, tb, Q) {
		c.ctxt.Diag("arrangement dismatch: %v", p)
		return
	}
	p.Insts = op3A(p, idx)
}

// asmVUADDWX deals with VUADDW and VUADDW2 instructions.
func asmVUADDWX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	tb, ta, ta2 := (p.From.Reg>>5)&15, (p.Reg>>5)&15, (p.To.Reg>>5)&15
	if !(ftyp == C_ARNG && f2typ == C_ARNG && ttyp == C_ARNG && ta == ta2) {
		return
	}
	idx, Q := UADDWvvv_t, int16(0)
	if p.As == AVUADDW2 {
		idx, Q = UADDW2vvv_t, 1
	}
	if _, match := immhTaMatchTb(ta, tb, Q); !match {
		c.ctxt.Diag("arrangement dismatch: %v", p)
		return
	}
	p.Insts = op3A(p, idx)
}

// asmVReg3OrARNG3 deals with instructions with three VREG or ARNG format operands.
func asmVReg3OrARNG3(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_VREG:
		if !((f2typ == C_NONE || f2typ == C_VREG) && ttyp == C_VREG) {
			break
		}
		idx := uint16(0)
		switch p.As {
		case AVADD:
			idx = ADDvvv
		case AVSUB:
			idx = SUBvvv
		case AVCMEQ:
			idx = CMEQvvv
		case AVCMTST:
			idx = CMTSTvvv
		}
		p.Insts = op3A(p, idx)
	case C_ARNG:
		af, an, at := int((p.From.Reg>>5)&15), int((p.Reg>>5)&15), int((p.To.Reg>>5)&15)
		if !(f2typ == C_ARNG && ttyp == C_ARNG && af == an && an == at) {
			break
		}
		idx := uint16(0)
		switch p.As {
		case AVADD:
			idx = ADDvvv_t
		case AVSUB:
			idx = SUBvvv_t
		case AVCMEQ:
			idx = CMEQvvv_t
		case AVCMTST:
			idx = CMTSTvvv_t
		}
		p.Insts = op3A(p, idx)
	}
}

// asmARNG4 deals with instructions with four ARNG format operands.
func asmARNG4(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	f3typ := c.aclass(p, p.GetFrom3())
	ttyp := c.aclass(p, &p.To)
	aa, ma, na, ta := (p.From.Reg>>5)&15, (p.Reg>>5)&15, (p.GetFrom3().Reg>>5)&15, (p.To.Reg>>5)&15
	if !(ftyp == C_ARNG && f2typ == C_ARNG && f3typ == C_ARNG && ttyp == C_ARNG && aa == ma && aa == na && aa == ta) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AVEOR3:
		idx = EOR3vvv
	case AVBCAX:
		idx = BCAXvvv
	}
	p.Insts = op4A(p, idx)
}

// ext
func asmVEXT(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	f3typ := c.aclass(p, p.GetFrom3())
	ttyp := c.aclass(p, &p.To)
	at, an, af3 := (p.To.Reg>>5)&15, (p.Reg>>5)&15, (p.GetFrom3().Reg>>5)&15
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && f3typ == C_ARNG && ttyp == C_ARNG && at == an && at == af3) {
		return
	}
	p.Insts = op4A(p, EXTvvvi_t)
	// record the arrangement specifier in p.From.Index so that we know the value when encoding.
	p.Insts[0].Args[3].Index = at
}

// xar
func asmVXAR(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	f3typ := c.aclass(p, p.GetFrom3())
	ttyp := c.aclass(p, &p.To)
	at, an, af3 := (p.To.Reg>>5)&15, (p.Reg>>5)&15, (p.GetFrom3().Reg>>5)&15
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && f3typ == C_ARNG && ttyp == C_ARNG && at == an && at == af3) {
		return
	}
	p.Insts = op4A(p, XARvvvi_t)
}

// vmov
func asmVMOV(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_REG:
		switch ttyp {
		case C_ARNG:
			p.Insts = op2A(p, DUPvr_t)
		case C_ELEM:
			p.Insts = op2A(p, MOVvr_ti)
		}
	case C_ARNG:
		if ttyp != C_ARNG {
			break
		}
		// VMOV <Vn>.<T>, <Vd>.<T> -> VORR <Vn>.<T>, <Vn>.<T>, <Vd>.<T>
		p.Insts = []obj.Inst{newInst(AVORR, ORRvvv_t, p.Pos, []obj.Oprd{p.To.ToOprd(), {Reg: p.From.Reg, Type: obj.TYPE_REG}, p.From.ToOprd()})}
	case C_ELEM:
		switch ttyp {
		case C_REG:
			switch (p.From.Reg >> 5) & 15 {
			case ARNG_B, ARNG_H:
				p.Insts = op2A(p, UMOVwv_ti)
			case ARNG_S:
				p.Insts = op2A(p, MOVwv_si)
			case ARNG_D:
				p.Insts = op2A(p, MOVxv_di)
			}
		case C_VREG:
			p.Insts = op2A(p, MOVvv_ti)
		case C_ELEM:
			if ((p.From.Reg >> 5) & 15) != ((p.To.Reg >> 5) & 15) {
				c.ctxt.Diag("operand mismatch: %v", p)
				return
			}
			p.Insts = op2A(p, MOVvv_tii)
		}
	}
}

// vmovq/vmovd/vmovs
func asmVMOVQ(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && from3 != nil && from3.Type == obj.TYPE_CONST && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM128
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	p.Insts = progToInst(p, LDRql, []obj.Oprd{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, p.From.ToOprd()})
}

func asmVMOVD(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	p.Insts = progToInst(p, LDRdl, []obj.Oprd{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, p.From.ToOprd()})
}

func asmVMOVS(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	p.Insts = progToInst(p, LDRsl, []obj.Oprd{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, p.From.ToOprd()})
}

// ld1/ld2/ld3/ld4/st1/st2/st3/st4
func asmVLD1(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ttyp {
	case C_LIST:
		// multiple structures
		opcode := (p.To.Offset >> 12) & 15
		switch ftyp {
		case C_ZOREG:
			switch opcode {
			case 0x7:
				if p.Scond == C_XPOST {
					p.Insts = op2A(p, LD1vxi_tp1) // Post-index
				} else {
					p.Insts = op2A(p, LD1vx_t1) // one register
				}
			case 0xa:
				if p.Scond == C_XPOST {
					p.Insts = op2A(p, LD1vxi_tp2)
				} else {
					p.Insts = op2A(p, LD1vx_t2) // two registers
				}
			case 0x6:
				if p.Scond == C_XPOST {
					p.Insts = op2A(p, LD1vxi_tp3)
				} else {
					p.Insts = op2A(p, LD1vx_t3) // three registers
				}
			case 0x2:
				if p.Scond == C_XPOST {
					p.Insts = op2A(p, LD1vxi_tp4)
				} else {
					p.Insts = op2A(p, LD1vx_t4) // four registers
				}
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			switch opcode {
			case 0x7:
				p.Insts = op2A(p, LD1vxx_tp1) // one register
			case 0xa:
				p.Insts = op2A(p, LD1vxx_tp2) // two registers
			case 0x6:
				p.Insts = op2A(p, LD1vxx_tp3) // three registers
			case 0x2:
				p.Insts = op2A(p, LD1vxx_tp4) // four registers
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
			}
			return
		case C_LOREG:
			regSize, q, index := int64(0), (p.To.Offset>>30)&1, uint16(0)
			switch opcode {
			case 0x7:
				regSize, index = 1, LD1vxi_tp1 // one register
			case 0xa:
				regSize, index = 2, LD1vxi_tp2 // two registers
			case 0x6:
				regSize, index = 3, LD1vxi_tp3 // three registers
			case 0x2:
				regSize, index = 4, LD1vxi_tp4 // four registers
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
				return
			}
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			p.Insts = op2A(p, index)
		}
	case C_ELEM:
		// single structure
		at := (p.To.Reg >> 5) & 15
		switch ftyp {
		case C_ZOREG:
			switch at {
			case ARNG_B:
				p.Insts = op2A(p, LD1vx_bi1)
			case ARNG_H:
				p.Insts = op2A(p, LD1vx_hi1)
			case ARNG_S:
				p.Insts = op2A(p, LD1vx_si1)
			case ARNG_D:
				p.Insts = op2A(p, LD1vx_di1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			switch at {
			case ARNG_B:
				p.Insts = op2A(p, LD1vxx_bip1)
			case ARNG_H:
				p.Insts = op2A(p, LD1vxx_hip1)
			case ARNG_S:
				p.Insts = op2A(p, LD1vxx_sip1)
			case ARNG_D:
				p.Insts = op2A(p, LD1vxx_dip1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement: %v", p)
			}
		case C_LOREG:
			if p.Scond != C_XPOST {
				break
			}
			switch at {
			case ARNG_B:
				p.Insts = op2A(p, LD1vxi_bip1)
			case ARNG_H:
				p.Insts = op2A(p, LD1vxi_hip1)
			case ARNG_S:
				p.Insts = op2A(p, LD1vxi_sip1)
			case ARNG_D:
				p.Insts = op2A(p, LD1vxi_dip1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement: %v", p)
			}
		}
	}
}

func asmVST1(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_LIST:
		// multiple structures
		opcode := (p.From.Offset >> 12) & 15
		switch ttyp {
		case C_ZOREG:
			switch opcode {
			case 0x7:
				if p.Scond == C_XPOST {
					p.Insts = op2ASO(p, ST1vxi_tp1) // Post-index
				} else {
					p.Insts = op2ASO(p, ST1vx_t1) // one register
				}
			case 0xa:
				if p.Scond == C_XPOST {
					p.Insts = op2ASO(p, ST1vxi_tp2)
				} else {
					p.Insts = op2ASO(p, ST1vx_t2) // two registers
				}
			case 0x6:
				if p.Scond == C_XPOST {
					p.Insts = op2ASO(p, ST1vxi_tp3)
				} else {
					p.Insts = op2ASO(p, ST1vx_t3) // three registers
				}
			case 0x2:
				if p.Scond == C_XPOST {
					p.Insts = op2ASO(p, ST1vxi_tp4)
				} else {
					p.Insts = op2ASO(p, ST1vx_t4) // four registers
				}
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			switch opcode {
			case 0x7:
				p.Insts = op2ASO(p, ST1vxx_tp1) // one register
			case 0xa:
				p.Insts = op2ASO(p, ST1vxx_tp2) // two registers
			case 0x6:
				p.Insts = op2ASO(p, ST1vxx_tp3) // three registers
			case 0x2:
				p.Insts = op2ASO(p, ST1vxx_tp4) // four registers
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
			}
		case C_LOREG:
			regSize, q, index := int64(0), (p.From.Offset>>30)&1, uint16(0)
			switch opcode {
			case 0x7:
				regSize, index = 1, ST1vxi_tp1 // one register
			case 0xa:
				regSize, index = 2, ST1vxi_tp2 // two registers
			case 0x6:
				regSize, index = 3, ST1vxi_tp3 // three registers
			case 0x2:
				regSize, index = 4, ST1vxi_tp4 // four registers
			default:
				c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
				return
			}
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			p.Insts = op2ASO(p, index)
		}
	case C_ELEM:
		// single structure
		af := (p.From.Reg >> 5) & 15
		switch ttyp {
		case C_ZOREG:
			switch af {
			case ARNG_B:
				p.Insts = op2ASO(p, ST1vx_bi1)
			case ARNG_H:
				p.Insts = op2ASO(p, ST1vx_hi1)
			case ARNG_S:
				p.Insts = op2ASO(p, ST1vx_si1)
			case ARNG_D:
				p.Insts = op2ASO(p, ST1vx_di1)
			default:
				c.ctxt.Diag("illegal source operand arrangement: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			switch af {
			case ARNG_B:
				p.Insts = op2ASO(p, ST1vxx_bip1)
			case ARNG_H:
				p.Insts = op2ASO(p, ST1vxx_hip1)
			case ARNG_S:
				p.Insts = op2ASO(p, ST1vxx_sip1)
			case ARNG_D:
				p.Insts = op2ASO(p, ST1vxx_dip1)
			default:
				c.ctxt.Diag("illegal source operand arrangement: %v", p)
			}
		case C_LOREG:
			if p.Scond != C_XPOST {
				break
			}
			switch af {
			case ARNG_B:
				p.Insts = op2ASO(p, ST1vxi_bip1)
			case ARNG_H:
				p.Insts = op2ASO(p, ST1vxi_hip1)
			case ARNG_S:
				p.Insts = op2ASO(p, ST1vxi_sip1)
			case ARNG_D:
				p.Insts = op2ASO(p, ST1vxi_dip1)
			default:
				c.ctxt.Diag("illegal source operand arrangement: %v", p)
			}
		}
	}
}

func asmVLD2(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0xa) {
		return
	} // two registers
	switch ttyp {
	case C_LIST:
		// multiple structures
		switch ftyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2A(p, LD2vxi_tp2) // Post-index
			} else {
				p.Insts = op2A(p, LD2vx_t2)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			p.Insts = op2A(p, LD2vxx_tp2)
		case C_LOREG:
			regSize, q := int64(2), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			p.Insts = op2A(p, LD2vxi_tp2)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func asmVLD3(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0x6) {
		return
	} // three registers
	switch ttyp {
	case C_LIST:
		// multiple structures
		switch ftyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2A(p, LD3vxi_tp3) // Post-index
			} else {
				p.Insts = op2A(p, LD3vx_t3)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			p.Insts = op2A(p, LD3vxx_tp3)
		case C_LOREG:
			regSize, q := int64(3), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			p.Insts = op2A(p, LD3vxi_tp3)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func asmVLD4(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0x2) {
		return
	} // four registers
	switch ttyp {
	case C_LIST:
		// multiple structures
		switch ftyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2A(p, LD4vxi_tp4) // Post-index
			} else {
				p.Insts = op2A(p, LD4vx_t4)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			p.Insts = op2A(p, LD4vxx_tp4)
		case C_LOREG:
			regSize, q := int64(4), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			p.Insts = op2A(p, LD4vxi_tp4)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func asmVST2(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.From.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0xa) {
		return
	} // two registers
	switch ftyp {
	case C_LIST:
		// multiple structures
		switch ttyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2ASO(p, ST2vxi_tp2) // Post-index
			} else {
				p.Insts = op2ASO(p, ST2vx_t2)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			p.Insts = op2ASO(p, ST2vxx_tp2)
		case C_LOREG:
			regSize, q := int64(2), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			p.Insts = op2ASO(p, ST2vxi_tp2)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func asmVST3(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.From.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0x6) {
		return
	} // three registers
	switch ftyp {
	case C_LIST:
		// multiple structures
		switch ttyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2ASO(p, ST3vxi_tp3) // Post-index
			} else {
				p.Insts = op2ASO(p, ST3vx_t3)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			p.Insts = op2ASO(p, ST3vxx_tp3)
		case C_LOREG:
			regSize, q := int64(3), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			p.Insts = op2ASO(p, ST3vxi_tp3)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func asmVST4(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.From.Offset >> 12) & 15
	if !(p.Reg == 0 && opcode == 0x2) {
		return
	} // four registers
	switch ftyp {
	case C_LIST:
		// multiple structures
		switch ttyp {
		case C_ZOREG:
			if p.Scond == C_XPOST {
				p.Insts = op2ASO(p, ST4vxi_tp4) // Post-index
			} else {
				p.Insts = op2ASO(p, ST4vx_t4)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			p.Insts = op2ASO(p, ST4vxx_tp4)
		case C_LOREG:
			regSize, q := int64(4), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			p.Insts = op2ASO(p, ST4vxi_tp4)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

// ld1r/ld2r/ld3r/ld4r
func asmVLD1R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x7) {
		return
	}
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			p.Insts = op2A(p, LD1Rvxi_tp1) // Post-index
		} else {
			p.Insts = op2A(p, LD1Rvx_t1)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		p.Insts = op2A(p, LD1Rvxx_tp1)
	case C_LOREG:
		regSize, size := int64(1), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		p.Insts = op2A(p, LD1Rvxi_tp1)
	}
}

func asmVLD2R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0xa) {
		return
	} // two registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			p.Insts = op2A(p, LD2Rvxi_tp2) // Post-index
		} else {
			p.Insts = op2A(p, LD2Rvx_t2)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		p.Insts = op2A(p, LD2Rvxx_tp2)
	case C_LOREG:
		regSize, size := int64(2), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		p.Insts = op2A(p, LD2Rvxi_tp2)
	}
}

func asmVLD3R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x6) {
		return
	} // three registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			p.Insts = op2A(p, LD3Rvxi_tp3) // Post-index
		} else {
			p.Insts = op2A(p, LD3Rvx_t3)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		p.Insts = op2A(p, LD3Rvxx_tp3)
	case C_LOREG:
		regSize, size := int64(3), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		p.Insts = op2A(p, LD3Rvxi_tp3)
	}
}

func asmVLD4R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x2) {
		return
	} // four registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			p.Insts = op2A(p, LD4Rvxi_tp4) // Post-index
		} else {
			p.Insts = op2A(p, LD4Rvx_t4)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		p.Insts = op2A(p, LD4Rvxx_tp4)
	case C_LOREG:
		regSize, size := int64(4), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		p.Insts = op2A(p, LD4Rvxi_tp4)
	}
}

// vdup
func asmVDUP(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_REG:
		if ttyp != C_ARNG {
			break
		}
		p.Insts = op2A(p, DUPvr_t)
	case C_ELEM:
		switch ttyp {
		case C_VREG:
			p.Insts = op2A(p, DUPvv_i)
		case C_ARNG:
			Ts, T := (p.From.Reg>>5)&15, (p.To.Reg>>5)&15
			if !arngTsMatchT(Ts, T) {
				c.ctxt.Diag("arrangement dismatch: %v", p)
				return
			}
			p.Insts = op2A(p, DUPvv_ti)
		}
	}
}

// uxtl/uxtl2
func asmVUXTLX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_ARNG && ttyp == C_ARNG && p.Reg == 0) {
		return
	}
	Tb, Ta := (p.From.Reg>>5)&15, (p.To.Reg>>5)&15
	idx, Q := UXTLvv_t, int16(0)
	if p.As == AVUXTL2 {
		idx, Q = UXTL2vv_t, 1
	}
	if _, match := immhTaMatchTb(Ta, Tb, Q); !match {
		c.ctxt.Diag("arrangement dismatch: %v", p)
		return
	}
	p.Insts = op2A(p, idx)
}

// asmVUSHLLX deals with VUSHLL and VUSHLL2 instructions.
func asmVUSHLLX(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && ttyp == C_ARNG) {
		return
	}
	Tb, Ta := (p.Reg>>5)&15, (p.To.Reg>>5)&15
	shift, max := int(p.From.Offset), 0
	match := false
	idx, Q := USHLLvvi_t, int16(0)
	if p.As == AVUSHLL2 {
		idx, Q = USHLL2vvi_t, 1
	}
	if max, match = immhTaMatchTb(Ta, Tb, Q); !match {
		c.ctxt.Diag("arrangement dismatch: %v", p)
		return
	}
	if shift < 0 || shift > max {
		c.ctxt.Diag("shift amount out of range: %v\n", p)
		return
	}
	p.Insts = op3A(p, idx)
}

// asmVectorShift deals with some vector shift instructions such as SHL, SLI.
func asmVectorShift(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	at, an := (p.To.Reg>>5)&15, (p.Reg>>5)&15
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && ttyp == C_ARNG && at == an) {
		return
	}
	// record the arrangement specifier in p.From.Index so that we know the value when encoding.
	p.From.Index = at
	idx := uint16(0)
	switch p.As {
	case AVSHL:
		idx = SHLvvi_t
	case AVSLI:
		idx = SLIvvi_t
	case AVSRI:
		idx = SRIvvi_t
	case AVUSRA:
		idx = USRAvvi_t
	case AVUSHR:
		idx = USHRvvi_t
	}
	p.Insts = op3A(p, idx)
}

// tbl
func asmVTBL(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	from3 := p.GetFrom3()
	af, at := (p.From.Reg>>5)&15, (p.To.Reg>>5)&15
	if !(ftyp == C_ARNG && p.Reg == 0 && from3 != nil && from3.Type == obj.TYPE_REGLIST && ttyp == C_ARNG && af == at) {
		return
	}
	regCntCode := (from3.Offset >> 12) & 15
	idx := uint16(0)
	switch regCntCode {
	case 0x7:
		idx = TBLvvv_t1 // one register
	case 0xa:
		idx = TBLvvv_t2 // two register
	case 0x6:
		idx = TBLvvv_t3 // three registers
	case 0x2:
		idx = TBLvvv_t4 // four registers
	default:
		c.ctxt.Diag("invalid register numbers in ARM64 register list: %v", p)
		return
	}
	p.Insts = progToInst(p, idx, []obj.Oprd{p.To.ToOprd(), from3.ToOprd(), p.From.ToOprd()})
}

// asmVADDVX deals with VADDV and VUADDLV instructions.
func asmVADDVX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_ARNG && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	if p.As == AVADDV {
		p.Insts = op2A(p, ADDVvv_t)
	} else {
		p.Insts = op2A(p, UADDLVvv_t)
	}
}

// vmovi
func asmVMOVI(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	switch ttyp {
	case C_FREG: // MOVI $imm, Fd
		p.Insts = op2A(p, MOVIdi)
	case C_ARNG:
		from3 := p.GetFrom3()
		if from3 != nil {
			if from3.Type != obj.TYPE_CONST {
				break
			}
			// VMOVI $<imm8>, $<amount>, <Vd>.<T> -> MOVI <Vd>.<T>, #<imm8>, MSL #<amount>
			p.Insts = progToInst(p, MOVIvii_ts, []obj.Oprd{p.To.ToOprd(), p.From.ToOprd(), from3.ToOprd()})
			break
		}
		at := (p.To.Reg >> 5) & 15
		switch at {
		case ARNG_2D:
			p.Insts = op2A(p, MOVIvi)
		case ARNG_2S, ARNG_4S:
			p.Insts = op2A(p, MOVIvi_ts)
		case ARNG_4H, ARNG_8H:
			p.Insts = op2A(p, MOVIvi_th)
		case ARNG_8B, ARNG_16B:
			p.Insts = op2A(p, MOVIvi_tb)
		default:
			c.ctxt.Diag("illegal destination operand arrangement: %v", p)
		}
	}
}

type asmFunc func(c *ctxt7, p *obj.Prog)

// asm function table
var unfoldTab = []asmFunc{
	obj.ACALL:     asmCall,
	obj.ADUFFCOPY: asmCall,
	obj.ADUFFZERO: asmCall,
	obj.AJMP:      asmJMP,
	obj.ARET:      asmRET,
	obj.AUNDEF:    asmUNDEF,

	AADC - obj.ABaseARM64:   asmOpWithCarry,
	AADCW - obj.ABaseARM64:  asmOpWithCarry,
	AADCS - obj.ABaseARM64:  asmOpWithCarry,
	AADCSW - obj.ABaseARM64: asmOpWithCarry,
	ASBC - obj.ABaseARM64:   asmOpWithCarry,
	ASBCW - obj.ABaseARM64:  asmOpWithCarry,
	ASBCS - obj.ABaseARM64:  asmOpWithCarry,
	ASBCSW - obj.ABaseARM64: asmOpWithCarry,

	ANGC - obj.ABaseARM64:   asmNGCX,
	ANGCW - obj.ABaseARM64:  asmNGCX,
	ANGCS - obj.ABaseARM64:  asmNGCX,
	ANGCSW - obj.ABaseARM64: asmNGCX,

	AADD - obj.ABaseARM64:   asmADD,
	AADDW - obj.ABaseARM64:  asmADDW,
	AADDS - obj.ABaseARM64:  asmADDS,
	AADDSW - obj.ABaseARM64: asmADDSW,

	ASUB - obj.ABaseARM64:   asmSUB,
	ASUBW - obj.ABaseARM64:  asmSUBW,
	ASUBS - obj.ABaseARM64:  asmSUBS,
	ASUBSW - obj.ABaseARM64: asmSUBSW,

	AADR - obj.ABaseARM64:  asmADRX,
	AADRP - obj.ABaseARM64: asmADRX,

	AAND - obj.ABaseARM64:  asmAND,
	AANDW - obj.ABaseARM64: asmANDW,
	AEOR - obj.ABaseARM64:  asmEOR,
	AEORW - obj.ABaseARM64: asmEORW,
	AORR - obj.ABaseARM64:  asmORR,
	AORRW - obj.ABaseARM64: asmORRW,
	ABIC - obj.ABaseARM64:  asmBIC,
	ABICW - obj.ABaseARM64: asmBICW,
	AEON - obj.ABaseARM64:  asmEON,
	AEONW - obj.ABaseARM64: asmEONW,
	AORN - obj.ABaseARM64:  asmORN,
	AORNW - obj.ABaseARM64: asmORNW,

	AANDS - obj.ABaseARM64:  asmANDS,
	AANDSW - obj.ABaseARM64: asmANDSW,
	ABICS - obj.ABaseARM64:  asmBICS,
	ABICSW - obj.ABaseARM64: asmBICSW,

	ATST - obj.ABaseARM64:  asmTST,
	ATSTW - obj.ABaseARM64: asmTST,

	ABFM - obj.ABaseARM64:    asmBitFieldOps,
	ABFMW - obj.ABaseARM64:   asmBitFieldOps,
	ABFI - obj.ABaseARM64:    asmBitFieldOps,
	ABFIW - obj.ABaseARM64:   asmBitFieldOps,
	ABFXIL - obj.ABaseARM64:  asmBitFieldOps,
	ABFXILW - obj.ABaseARM64: asmBitFieldOps,
	ASBFM - obj.ABaseARM64:   asmBitFieldOps,
	ASBFMW - obj.ABaseARM64:  asmBitFieldOps,
	ASBFIZ - obj.ABaseARM64:  asmBitFieldOps,
	ASBFIZW - obj.ABaseARM64: asmBitFieldOps,
	ASBFX - obj.ABaseARM64:   asmBitFieldOps,
	ASBFXW - obj.ABaseARM64:  asmBitFieldOps,
	AUBFM - obj.ABaseARM64:   asmBitFieldOps,
	AUBFMW - obj.ABaseARM64:  asmBitFieldOps,
	AUBFIZ - obj.ABaseARM64:  asmBitFieldOps,
	AUBFIZW - obj.ABaseARM64: asmBitFieldOps,
	AUBFX - obj.ABaseARM64:   asmBitFieldOps,
	AUBFXW - obj.ABaseARM64:  asmBitFieldOps,

	AMOVD - obj.ABaseARM64:  asmMOVD,
	AMOVW - obj.ABaseARM64:  asmMOVW,
	AMOVWU - obj.ABaseARM64: asmMOVWU,
	AMOVH - obj.ABaseARM64:  asmMOVH,
	AMOVHU - obj.ABaseARM64: asmMOVHU,
	AMOVB - obj.ABaseARM64:  asmMOVB,
	AMOVBU - obj.ABaseARM64: asmMOVBU,

	AFMOVQ - obj.ABaseARM64: asmFMOVQ,
	AFMOVD - obj.ABaseARM64: asmFMOVD,
	AFMOVS - obj.ABaseARM64: asmFMOVS,

	ALDP - obj.ABaseARM64:   asmLDP,
	ALDPW - obj.ABaseARM64:  asmLDPW,
	ALDPSW - obj.ABaseARM64: asmLDPSW,
	AFLDPQ - obj.ABaseARM64: asmFLDPQ,
	AFLDPD - obj.ABaseARM64: asmFLDPD,
	AFLDPS - obj.ABaseARM64: asmFLDPS,
	ASTP - obj.ABaseARM64:   asmSTP,
	ASTPW - obj.ABaseARM64:  asmSTPW,
	AFSTPQ - obj.ABaseARM64: asmFSTPQ,
	AFSTPD - obj.ABaseARM64: asmFSTPD,
	AFSTPS - obj.ABaseARM64: asmFSTPS,

	ACBZ - obj.ABaseARM64:   asmCBZX,
	ACBZW - obj.ABaseARM64:  asmCBZX,
	ACBNZ - obj.ABaseARM64:  asmCBZX,
	ACBNZW - obj.ABaseARM64: asmCBZX,

	ACCMN - obj.ABaseARM64:  asmCondCMP,
	ACCMNW - obj.ABaseARM64: asmCondCMP,
	ACCMP - obj.ABaseARM64:  asmCondCMP,
	ACCMPW - obj.ABaseARM64: asmCondCMP,

	ACSEL - obj.ABaseARM64:   asmCSELX,
	ACSELW - obj.ABaseARM64:  asmCSELX,
	ACSINC - obj.ABaseARM64:  asmCSELX,
	ACSINCW - obj.ABaseARM64: asmCSELX,
	ACSINV - obj.ABaseARM64:  asmCSELX,
	ACSINVW - obj.ABaseARM64: asmCSELX,
	ACSNEG - obj.ABaseARM64:  asmCSELX,
	ACSNEGW - obj.ABaseARM64: asmCSELX,
	AFCSELD - obj.ABaseARM64: asmCSELX,
	AFCSELS - obj.ABaseARM64: asmCSELX,

	ACSET - obj.ABaseARM64:   asmCSETX,
	ACSETW - obj.ABaseARM64:  asmCSETX,
	ACSETM - obj.ABaseARM64:  asmCSETX,
	ACSETMW - obj.ABaseARM64: asmCSETX,

	ACINC - obj.ABaseARM64:  asmCINCX,
	ACINCW - obj.ABaseARM64: asmCINCX,
	ACINV - obj.ABaseARM64:  asmCINCX,
	ACINVW - obj.ABaseARM64: asmCINCX,
	ACNEG - obj.ABaseARM64:  asmCINCX,
	ACNEGW - obj.ABaseARM64: asmCINCX,

	AFCCMPD - obj.ABaseARM64:  asmFloatingCondCMP,
	AFCCMPS - obj.ABaseARM64:  asmFloatingCondCMP,
	AFCCMPED - obj.ABaseARM64: asmFloatingCondCMP,
	AFCCMPES - obj.ABaseARM64: asmFloatingCondCMP,

	ACLREX - obj.ABaseARM64: asmCLREX,

	ACLS - obj.ABaseARM64:    asmCLSX,
	ACLSW - obj.ABaseARM64:   asmCLSX,
	ACLZ - obj.ABaseARM64:    asmCLSX,
	ACLZW - obj.ABaseARM64:   asmCLSX,
	ARBIT - obj.ABaseARM64:   asmCLSX,
	ARBITW - obj.ABaseARM64:  asmCLSX,
	AREV - obj.ABaseARM64:    asmCLSX,
	AREVW - obj.ABaseARM64:   asmCLSX,
	AREV16 - obj.ABaseARM64:  asmCLSX,
	AREV16W - obj.ABaseARM64: asmCLSX,
	AREV32 - obj.ABaseARM64:  asmCLSX,

	ACMP - obj.ABaseARM64:  asmCMP,
	ACMPW - obj.ABaseARM64: asmCMPW,
	ACMN - obj.ABaseARM64:  asmCMN,
	ACMNW - obj.ABaseARM64: asmCMNW,

	ADMB - obj.ABaseARM64:  asmDMBX,
	ADSB - obj.ABaseARM64:  asmDMBX,
	AISB - obj.ABaseARM64:  asmDMBX,
	AHINT - obj.ABaseARM64: asmDMBX,

	AERET - obj.ABaseARM64:  asmOpwithoutArg,
	AWFE - obj.ABaseARM64:   asmOpwithoutArg,
	AWFI - obj.ABaseARM64:   asmOpwithoutArg,
	AYIELD - obj.ABaseARM64: asmOpwithoutArg,
	ASEV - obj.ABaseARM64:   asmOpwithoutArg,
	ASEVL - obj.ABaseARM64:  asmOpwithoutArg,
	ANOOP - obj.ABaseARM64:  asmOpwithoutArg,
	ADRPS - obj.ABaseARM64:  asmOpwithoutArg,

	AEXTR - obj.ABaseARM64:  asmEXTRX,
	AEXTRW - obj.ABaseARM64: asmEXTRX,

	ALDAR - obj.ABaseARM64:   asmLoadAcquire,
	ALDARW - obj.ABaseARM64:  asmLoadAcquire,
	ALDARH - obj.ABaseARM64:  asmLoadAcquire,
	ALDARB - obj.ABaseARM64:  asmLoadAcquire,
	ALDAXR - obj.ABaseARM64:  asmLoadAcquire,
	ALDAXRW - obj.ABaseARM64: asmLoadAcquire,
	ALDAXRH - obj.ABaseARM64: asmLoadAcquire,
	ALDAXRB - obj.ABaseARM64: asmLoadAcquire,
	ALDXR - obj.ABaseARM64:   asmLoadAcquire,
	ALDXRW - obj.ABaseARM64:  asmLoadAcquire,
	ALDXRH - obj.ABaseARM64:  asmLoadAcquire,
	ALDXRB - obj.ABaseARM64:  asmLoadAcquire,

	ALDXP - obj.ABaseARM64:   asmLDXPX,
	ALDXPW - obj.ABaseARM64:  asmLDXPX,
	ALDAXP - obj.ABaseARM64:  asmLDXPX,
	ALDAXPW - obj.ABaseARM64: asmLDXPX,

	ASTLR - obj.ABaseARM64:  asmSTLRX,
	ASTLRW - obj.ABaseARM64: asmSTLRX,
	ASTLRH - obj.ABaseARM64: asmSTLRX,
	ASTLRB - obj.ABaseARM64: asmSTLRX,

	ASTLXR - obj.ABaseARM64:  asmSTXRX,
	ASTLXRW - obj.ABaseARM64: asmSTXRX,
	ASTLXRH - obj.ABaseARM64: asmSTXRX,
	ASTLXRB - obj.ABaseARM64: asmSTXRX,
	ASTXR - obj.ABaseARM64:   asmSTXRX,
	ASTXRW - obj.ABaseARM64:  asmSTXRX,
	ASTXRH - obj.ABaseARM64:  asmSTXRX,
	ASTXRB - obj.ABaseARM64:  asmSTXRX,

	ASTXP - obj.ABaseARM64:   asmSTXPX,
	ASTXPW - obj.ABaseARM64:  asmSTXPX,
	ASTLXP - obj.ABaseARM64:  asmSTXPX,
	ASTLXPW - obj.ABaseARM64: asmSTXPX,

	ALSL - obj.ABaseARM64:  asmShiftOp,
	ALSLW - obj.ABaseARM64: asmShiftOp,
	ALSR - obj.ABaseARM64:  asmShiftOp,
	ALSRW - obj.ABaseARM64: asmShiftOp,
	AASR - obj.ABaseARM64:  asmShiftOp,
	AASRW - obj.ABaseARM64: asmShiftOp,
	AROR - obj.ABaseARM64:  asmShiftOp,
	ARORW - obj.ABaseARM64: asmShiftOp,

	AMADD - obj.ABaseARM64:    asmFuseOp,
	AMADDW - obj.ABaseARM64:   asmFuseOp,
	AMSUB - obj.ABaseARM64:    asmFuseOp,
	AMSUBW - obj.ABaseARM64:   asmFuseOp,
	ASMADDL - obj.ABaseARM64:  asmFuseOp,
	ASMSUBL - obj.ABaseARM64:  asmFuseOp,
	AUMADDL - obj.ABaseARM64:  asmFuseOp,
	AUMSUBL - obj.ABaseARM64:  asmFuseOp,
	AFMADDD - obj.ABaseARM64:  asmFuseOp,
	AFMADDS - obj.ABaseARM64:  asmFuseOp,
	AFMSUBD - obj.ABaseARM64:  asmFuseOp,
	AFMSUBS - obj.ABaseARM64:  asmFuseOp,
	AFNMADDD - obj.ABaseARM64: asmFuseOp,
	AFNMADDS - obj.ABaseARM64: asmFuseOp,
	AFNMSUBD - obj.ABaseARM64: asmFuseOp,
	AFNMSUBS - obj.ABaseARM64: asmFuseOp,

	AMUL - obj.ABaseARM64:    asmMultipleX,
	AMULW - obj.ABaseARM64:   asmMultipleX,
	AMNEG - obj.ABaseARM64:   asmMultipleX,
	AMNEGW - obj.ABaseARM64:  asmMultipleX,
	ASMNEGL - obj.ABaseARM64: asmMultipleX,
	AUMNEGL - obj.ABaseARM64: asmMultipleX,
	ASMULH - obj.ABaseARM64:  asmMultipleX,
	ASMULL - obj.ABaseARM64:  asmMultipleX,
	AUMULH - obj.ABaseARM64:  asmMultipleX,
	AUMULL - obj.ABaseARM64:  asmMultipleX,

	AMOVK - obj.ABaseARM64:  asmMOVX,
	AMOVKW - obj.ABaseARM64: asmMOVX,
	AMOVN - obj.ABaseARM64:  asmMOVX,
	AMOVNW - obj.ABaseARM64: asmMOVX,
	AMOVZ - obj.ABaseARM64:  asmMOVX,
	AMOVZW - obj.ABaseARM64: asmMOVX,

	AMRS - obj.ABaseARM64: asmMRS,
	AMSR - obj.ABaseARM64: asmMSR,

	AMVN - obj.ABaseARM64:  asmMVN,
	AMVNW - obj.ABaseARM64: asmMVN,

	ANEG - obj.ABaseARM64:   asmNEGX,
	ANEGW - obj.ABaseARM64:  asmNEGX,
	ANEGS - obj.ABaseARM64:  asmNEGX,
	ANEGSW - obj.ABaseARM64: asmNEGX,

	APRFM - obj.ABaseARM64: asmPRFM,

	AREM - obj.ABaseARM64:   asmREMX,
	AREMW - obj.ABaseARM64:  asmREMX,
	AUREM - obj.ABaseARM64:  asmREMX,
	AUREMW - obj.ABaseARM64: asmREMX,

	ASDIV - obj.ABaseARM64:    asmDIVCRC,
	ASDIVW - obj.ABaseARM64:   asmDIVCRC,
	AUDIV - obj.ABaseARM64:    asmDIVCRC,
	AUDIVW - obj.ABaseARM64:   asmDIVCRC,
	ACRC32B - obj.ABaseARM64:  asmDIVCRC,
	ACRC32H - obj.ABaseARM64:  asmDIVCRC,
	ACRC32W - obj.ABaseARM64:  asmDIVCRC,
	ACRC32X - obj.ABaseARM64:  asmDIVCRC,
	ACRC32CB - obj.ABaseARM64: asmDIVCRC,
	ACRC32CH - obj.ABaseARM64: asmDIVCRC,
	ACRC32CW - obj.ABaseARM64: asmDIVCRC,
	ACRC32CX - obj.ABaseARM64: asmDIVCRC,

	ASVC - obj.ABaseARM64:   asmPE,
	AHVC - obj.ABaseARM64:   asmPE,
	AHLT - obj.ABaseARM64:   asmPE,
	ASMC - obj.ABaseARM64:   asmPE,
	ABRK - obj.ABaseARM64:   asmPE,
	ADCPS1 - obj.ABaseARM64: asmPE,
	ADCPS2 - obj.ABaseARM64: asmPE,
	ADCPS3 - obj.ABaseARM64: asmPE,

	ASXTB - obj.ABaseARM64:  asmExtend,
	ASXTBW - obj.ABaseARM64: asmExtend,
	ASXTH - obj.ABaseARM64:  asmExtend,
	ASXTHW - obj.ABaseARM64: asmExtend,
	ASXTW - obj.ABaseARM64:  asmExtend,
	AUXTBW - obj.ABaseARM64: asmExtend,
	AUXTHW - obj.ABaseARM64: asmExtend,

	AUXTB - obj.ABaseARM64: asmUnsignedExtend,
	AUXTH - obj.ABaseARM64: asmUnsignedExtend,
	AUXTW - obj.ABaseARM64: asmUnsignedExtend,

	ASYS - obj.ABaseARM64:  asmSYS,
	ASYSL - obj.ABaseARM64: asmSYSL,
	AAT - obj.ABaseARM64:   asmSYSAlias,
	ADC - obj.ABaseARM64:   asmSYSAlias,
	AIC - obj.ABaseARM64:   asmSYSAlias,
	ATLBI - obj.ABaseARM64: asmSYSAlias,

	ATBNZ - obj.ABaseARM64: asmTBZX,
	ATBZ - obj.ABaseARM64:  asmTBZX,

	ACASD - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ACASW - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ACASH - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ACASB - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ACASAD - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ACASAW - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ACASALD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ACASALW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ACASALH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ACASALB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ACASLD - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ACASLW - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ALDADDD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDADDW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDADDH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDADDB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDADDAD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDAW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDAH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDAB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDALD - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDADDALW - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDADDALH - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDADDALB - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDADDLD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDLW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDLH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDADDLB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDCLRW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDCLRH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDCLRB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDCLRAD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRAW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRAH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRAB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRALD - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDCLRALW - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDCLRALH - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDCLRALB - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDCLRLD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRLW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRLH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDCLRLB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDEORW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDEORH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDEORB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDEORAD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORAW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORAH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORAB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORALD - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDEORALW - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDEORALH - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDEORALB - obj.ABaseARM64: asmAtomicLoadOpStore,
	ALDEORLD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORLW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORLH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDEORLB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDORD - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ALDORW - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ALDORH - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ALDORB - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ALDORAD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORAW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORAH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORAB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORALD - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDORALW - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDORALH - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDORALB - obj.ABaseARM64:  asmAtomicLoadOpStore,
	ALDORLD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORLW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORLH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ALDORLB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ASWPD - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ASWPW - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ASWPH - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ASWPB - obj.ABaseARM64:     asmAtomicLoadOpStore,
	ASWPAD - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPAW - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPAH - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPAB - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPALD - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ASWPALW - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ASWPALH - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ASWPALB - obj.ABaseARM64:   asmAtomicLoadOpStore,
	ASWPLD - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPLW - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPLH - obj.ABaseARM64:    asmAtomicLoadOpStore,
	ASWPLB - obj.ABaseARM64:    asmAtomicLoadOpStore,

	ACASPD - obj.ABaseARM64: asmCASPX,
	ACASPW - obj.ABaseARM64: asmCASPX,

	ABEQ - obj.ABaseARM64: asmBCOND,
	ABNE - obj.ABaseARM64: asmBCOND,
	ABCS - obj.ABaseARM64: asmBCOND,
	ABHS - obj.ABaseARM64: asmBCOND,
	ABCC - obj.ABaseARM64: asmBCOND,
	ABLO - obj.ABaseARM64: asmBCOND,
	ABMI - obj.ABaseARM64: asmBCOND,
	ABPL - obj.ABaseARM64: asmBCOND,
	ABVS - obj.ABaseARM64: asmBCOND,
	ABVC - obj.ABaseARM64: asmBCOND,
	ABHI - obj.ABaseARM64: asmBCOND,
	ABLS - obj.ABaseARM64: asmBCOND,
	ABGE - obj.ABaseARM64: asmBCOND,
	ABLT - obj.ABaseARM64: asmBCOND,
	ABGT - obj.ABaseARM64: asmBCOND,
	ABLE - obj.ABaseARM64: asmBCOND,

	AFADDD - obj.ABaseARM64:   asmFloatingOp,
	AFADDS - obj.ABaseARM64:   asmFloatingOp,
	AFSUBD - obj.ABaseARM64:   asmFloatingOp,
	AFSUBS - obj.ABaseARM64:   asmFloatingOp,
	AFMULD - obj.ABaseARM64:   asmFloatingOp,
	AFMULS - obj.ABaseARM64:   asmFloatingOp,
	AFNMULD - obj.ABaseARM64:  asmFloatingOp,
	AFNMULS - obj.ABaseARM64:  asmFloatingOp,
	AFDIVD - obj.ABaseARM64:   asmFloatingOp,
	AFDIVS - obj.ABaseARM64:   asmFloatingOp,
	AFMAXD - obj.ABaseARM64:   asmFloatingOp,
	AFMAXS - obj.ABaseARM64:   asmFloatingOp,
	AFMIND - obj.ABaseARM64:   asmFloatingOp,
	AFMINS - obj.ABaseARM64:   asmFloatingOp,
	AFMAXNMD - obj.ABaseARM64: asmFloatingOp,
	AFMAXNMS - obj.ABaseARM64: asmFloatingOp,
	AFMINNMD - obj.ABaseARM64: asmFloatingOp,
	AFMINNMS - obj.ABaseARM64: asmFloatingOp,

	AFCMPD - obj.ABaseARM64:  asmFCMPX,
	AFCMPS - obj.ABaseARM64:  asmFCMPX,
	AFCMPED - obj.ABaseARM64: asmFCMPX,
	AFCMPES - obj.ABaseARM64: asmFCMPX,

	AFCVTDH - obj.ABaseARM64:  asmFConvertRounding,
	AFCVTDS - obj.ABaseARM64:  asmFConvertRounding,
	AFCVTHD - obj.ABaseARM64:  asmFConvertRounding,
	AFCVTHS - obj.ABaseARM64:  asmFConvertRounding,
	AFCVTSD - obj.ABaseARM64:  asmFConvertRounding,
	AFCVTSH - obj.ABaseARM64:  asmFConvertRounding,
	AFABSD - obj.ABaseARM64:   asmFConvertRounding,
	AFABSS - obj.ABaseARM64:   asmFConvertRounding,
	AFNEGD - obj.ABaseARM64:   asmFConvertRounding,
	AFNEGS - obj.ABaseARM64:   asmFConvertRounding,
	AFSQRTD - obj.ABaseARM64:  asmFConvertRounding,
	AFSQRTS - obj.ABaseARM64:  asmFConvertRounding,
	AFRINTAD - obj.ABaseARM64: asmFConvertRounding,
	AFRINTAS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTID - obj.ABaseARM64: asmFConvertRounding,
	AFRINTIS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTMD - obj.ABaseARM64: asmFConvertRounding,
	AFRINTMS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTND - obj.ABaseARM64: asmFConvertRounding,
	AFRINTNS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTPD - obj.ABaseARM64: asmFConvertRounding,
	AFRINTPS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTXD - obj.ABaseARM64: asmFConvertRounding,
	AFRINTXS - obj.ABaseARM64: asmFConvertRounding,
	AFRINTZD - obj.ABaseARM64: asmFConvertRounding,
	AFRINTZS - obj.ABaseARM64: asmFConvertRounding,

	AFCVTZSD - obj.ABaseARM64:  asmFConvertToFixed,
	AFCVTZSDW - obj.ABaseARM64: asmFConvertToFixed,
	AFCVTZSS - obj.ABaseARM64:  asmFConvertToFixed,
	AFCVTZSSW - obj.ABaseARM64: asmFConvertToFixed,
	AFCVTZUD - obj.ABaseARM64:  asmFConvertToFixed,
	AFCVTZUDW - obj.ABaseARM64: asmFConvertToFixed,
	AFCVTZUS - obj.ABaseARM64:  asmFConvertToFixed,
	AFCVTZUSW - obj.ABaseARM64: asmFConvertToFixed,

	ASCVTFD - obj.ABaseARM64:  asmFixedToFloating,
	ASCVTFS - obj.ABaseARM64:  asmFixedToFloating,
	ASCVTFWD - obj.ABaseARM64: asmFixedToFloating,
	ASCVTFWS - obj.ABaseARM64: asmFixedToFloating,
	AUCVTFD - obj.ABaseARM64:  asmFixedToFloating,
	AUCVTFS - obj.ABaseARM64:  asmFixedToFloating,
	AUCVTFWD - obj.ABaseARM64: asmFixedToFloating,
	AUCVTFWS - obj.ABaseARM64: asmFixedToFloating,

	AAESD - obj.ABaseARM64:      asmAESSHA,
	AAESE - obj.ABaseARM64:      asmAESSHA,
	AAESIMC - obj.ABaseARM64:    asmAESSHA,
	AAESMC - obj.ABaseARM64:     asmAESSHA,
	ASHA1SU1 - obj.ABaseARM64:   asmAESSHA,
	ASHA256SU0 - obj.ABaseARM64: asmAESSHA,
	ASHA512SU0 - obj.ABaseARM64: asmAESSHA,

	AVREV16 - obj.ABaseARM64: asmFBitwiseOp,
	AVREV32 - obj.ABaseARM64: asmFBitwiseOp,
	AVREV64 - obj.ABaseARM64: asmFBitwiseOp,
	AVCNT - obj.ABaseARM64:   asmFBitwiseOp,
	AVRBIT - obj.ABaseARM64:  asmFBitwiseOp,

	ASHA1H - obj.ABaseARM64: asmSHA1H,

	ASHA1C - obj.ABaseARM64:    asmSHAX,
	ASHA1P - obj.ABaseARM64:    asmSHAX,
	ASHA1M - obj.ABaseARM64:    asmSHAX,
	ASHA256H - obj.ABaseARM64:  asmSHAX,
	ASHA256H2 - obj.ABaseARM64: asmSHAX,
	ASHA512H - obj.ABaseARM64:  asmSHAX,
	ASHA512H2 - obj.ABaseARM64: asmSHAX,

	ASHA1SU0 - obj.ABaseARM64:   asmARNG3,
	ASHA256SU1 - obj.ABaseARM64: asmARNG3,
	ASHA512SU1 - obj.ABaseARM64: asmARNG3,
	AVRAX1 - obj.ABaseARM64:     asmARNG3,
	AVADDP - obj.ABaseARM64:     asmARNG3,
	AVAND - obj.ABaseARM64:      asmARNG3,
	AVORR - obj.ABaseARM64:      asmARNG3,
	AVEOR - obj.ABaseARM64:      asmARNG3,
	AVBIF - obj.ABaseARM64:      asmARNG3,
	AVBIT - obj.ABaseARM64:      asmARNG3,
	AVBSL - obj.ABaseARM64:      asmARNG3,
	AVUMAX - obj.ABaseARM64:     asmARNG3,
	AVUMIN - obj.ABaseARM64:     asmARNG3,
	AVUZP1 - obj.ABaseARM64:     asmARNG3,
	AVUZP2 - obj.ABaseARM64:     asmARNG3,
	AVFMLA - obj.ABaseARM64:     asmARNG3,
	AVFMLS - obj.ABaseARM64:     asmARNG3,
	AVZIP1 - obj.ABaseARM64:     asmARNG3,
	AVZIP2 - obj.ABaseARM64:     asmARNG3,

	AVPMULL - obj.ABaseARM64:  asmVPMULLX,
	AVPMULL2 - obj.ABaseARM64: asmVPMULLX,

	AVUADDW - obj.ABaseARM64:  asmVUADDWX,
	AVUADDW2 - obj.ABaseARM64: asmVUADDWX,

	AVADD - obj.ABaseARM64:   asmVReg3OrARNG3,
	AVSUB - obj.ABaseARM64:   asmVReg3OrARNG3,
	AVCMEQ - obj.ABaseARM64:  asmVReg3OrARNG3,
	AVCMTST - obj.ABaseARM64: asmVReg3OrARNG3,

	AVEOR3 - obj.ABaseARM64: asmARNG4,
	AVBCAX - obj.ABaseARM64: asmARNG4,

	AVEXT - obj.ABaseARM64: asmVEXT,

	AVXAR - obj.ABaseARM64: asmVXAR,

	AVMOV - obj.ABaseARM64: asmVMOV,

	AVMOVQ - obj.ABaseARM64: asmVMOVQ,
	AVMOVD - obj.ABaseARM64: asmVMOVD,
	AVMOVS - obj.ABaseARM64: asmVMOVS,

	AVLD1 - obj.ABaseARM64: asmVLD1,
	AVST1 - obj.ABaseARM64: asmVST1,
	AVLD2 - obj.ABaseARM64: asmVLD2,
	AVLD3 - obj.ABaseARM64: asmVLD3,
	AVLD4 - obj.ABaseARM64: asmVLD4,
	AVST2 - obj.ABaseARM64: asmVST2,
	AVST3 - obj.ABaseARM64: asmVST3,
	AVST4 - obj.ABaseARM64: asmVST4,

	AVLD1R - obj.ABaseARM64: asmVLD1R,
	AVLD2R - obj.ABaseARM64: asmVLD2R,
	AVLD3R - obj.ABaseARM64: asmVLD3R,
	AVLD4R - obj.ABaseARM64: asmVLD4R,

	AVDUP - obj.ABaseARM64: asmVDUP,

	AVUXTL - obj.ABaseARM64:  asmVUXTLX,
	AVUXTL2 - obj.ABaseARM64: asmVUXTLX,

	AVUSHLL - obj.ABaseARM64:  asmVUSHLLX,
	AVUSHLL2 - obj.ABaseARM64: asmVUSHLLX,

	AVSHL - obj.ABaseARM64:  asmVectorShift,
	AVSLI - obj.ABaseARM64:  asmVectorShift,
	AVSRI - obj.ABaseARM64:  asmVectorShift,
	AVUSRA - obj.ABaseARM64: asmVectorShift,
	AVUSHR - obj.ABaseARM64: asmVectorShift,

	AVTBL - obj.ABaseARM64: asmVTBL,

	AVADDV - obj.ABaseARM64:   asmVADDVX,
	AVUADDLV - obj.ABaseARM64: asmVADDVX,

	AVMOVI - obj.ABaseARM64: asmVMOVI,
}
