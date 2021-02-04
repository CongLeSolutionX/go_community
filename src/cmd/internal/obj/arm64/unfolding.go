// Copyright 2022 The Go Authors. All rights reserved.
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

// newInst constructs a new inst with the given arguments.
func newInst(as obj.As, idx uint16, pos src.XPos, args []arg) inst {
	return inst{As: as, Optab: idx, Pos: pos, Args: args}
}

// progToInst constructs a new inst with p.As, idx, p.Pos and args.
func progToInst(p *obj.Prog, idx uint16, args []arg) []inst {
	return []inst{newInst(p.As, idx, p.Pos, args)}
}

// op2A returns the corresponding inst of p which has two operands, and the
// operand order of p is exactly the opposite of that of the arm64 instruction.
func op2A(p *obj.Prog, idx uint16) []inst {
	return progToInst(p, idx, []arg{addrToArg(p.To), addrToArg(p.From)})
}

// op2ASO returns the corresponding inst of p which has two operands, and the
// operand order of p is exactly the same as that of the arm64 instruction.
func op2ASO(p *obj.Prog, idx uint16) []inst {
	return progToInst(p, idx, []arg{addrToArg(p.From), addrToArg(p.To)})
}

// op2D3A returns the corresponding inst of a three-operand Prog with two
// destination operands, such as STLXR instruction. The second destination
// operand is stored in p.RegTo2.
func op2D3A(p *obj.Prog, idx uint16) []inst {
	a := []arg{{Reg: p.RegTo2, Type: obj.TYPE_REG}, addrToArg(p.From), addrToArg(p.To)}
	return progToInst(p, idx, a)
}

// op2SA returns the corresponding inst of a Prog with two sourc operands:
// opcode Rm(or $imm), Rn. The operand order of p must be exactly the opposite
// of the operand order of the corresponding arm64 instruction.
func op2SA(p *obj.Prog, idx uint16) []inst {
	a := []arg{{Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(p.From)}
	return progToInst(p, idx, a)
}

// op3A returns the corresponding inst of p which has three operands, and the
// operand order of p is exactly the opposite of that of the arm64 instruction.
func op3A(p *obj.Prog, idx uint16) []inst {
	// create an Addr for the second operand.
	r := p.Reg
	if r == 0 {
		r = p.To.Reg
	}
	a := []arg{addrToArg(p.To), {Reg: r, Type: obj.TYPE_REG}, addrToArg(p.From)}
	return progToInst(p, idx, a)
}

// op3S4A returns the corresponding inst of a four-operand Prog with three
// source operands, such as MADD instruction. The three source operands are
// stored in p.RestArg3, p.From and p.Reg respectively.
func op3S4A(p *obj.Prog, idx uint16) []inst {
	a := []arg{addrToArg(p.To), addrToArg(*p.GetFrom3()), addrToArg(p.From), {Reg: p.Reg, Type: obj.TYPE_REG}}
	return progToInst(p, idx, a)
}

// op4A returns the corresponding inst of a Prog with four operands:
// p.From, p.Reg, p.GetFrom3() and p.To. The argument order of
// the Go instruction is exactly the opposite of that of the arm64 instruction.
func op4A(p *obj.Prog, idx uint16) []inst {
	a := []arg{addrToArg(p.To), addrToArg(*p.GetFrom3()), {Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(p.From)}
	return progToInst(p, idx, a)
}

// movcon64 moves the 64-bit constant con64 into the rto register by
// MOVZ/MOVN/MOVK instructions.
func (c *ctxt7) movcon64(con64 int64, rto int16, pos src.XPos) []inst {
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
	rt := arg{Reg: rto, Type: obj.TYPE_REG}
	i, imm := 0, uint64(0)
	for ; i < 4; i++ {
		imm = (dn >> uint(i*16)) & 0xffff
		if imm != 0 {
			break
		}
	}
	// Here arg.Index is used to store the shift offset of imm.
	args := []arg{rt, {Offset: int64(imm), Type: obj.TYPE_CONST, Index: int16(i << 4)}}
	ret := []inst{newInst(as, idx, pos, args)}
	for i++; i < 4; i++ {
		imm = (d >> uint(i*16)) & 0xffff
		if imm != zOrN {
			args = []arg{rt, {Offset: int64(imm), Type: obj.TYPE_CONST, Index: int16(i << 4)}}
			ret = append(ret, newInst(AMOVK, MOVKxis, pos, args))
		}
	}
	return ret
}

// movCon64ToReg moves the 64-bit constant con64 into the rto register by
// MOVZ/MOVN/MOVK instructions.
func (c *ctxt7) movCon64ToReg(con64 int64, rto int16, pos src.XPos) []inst {
	if isbitcon(uint64(con64)) {
		// -> MOV $imm, rto
		return []inst{instIR(AMOVD, con64, rto, MOVxi_b, pos)}
	} else if movcon(con64) >= 0 {
		// -> MOVZ $imm, rto
		return []inst{instIR(AMOVD, con64, rto, MOVZxis, pos)}
	} else if movcon(^con64) >= 0 {
		// -> MOVN $imm, rto
		return []inst{instIR(AMOVD, ^con64, rto, MOVNxis, pos)}
	}
	// Any other 64-bit integer constant.
	// -> MOVZ/MOVN $imm, rto + (MOVK $imm, rto)+
	return c.movcon64(con64, rto, pos)
}

// movcon32 moves the 32-bit constant con32 into the rto register by one MOVZW
// and one MOVKW instruction.
func (c *ctxt7) movcon32(con32 int64, rto int16, pos src.XPos) []inst {
	args := []arg{{Reg: rto, Type: obj.TYPE_REG}, {Offset: con32 & 0xffff, Type: obj.TYPE_CONST}}
	movzw := newInst(AMOVZW, MOVZwis, pos, args)
	// Here arg.Index is used to store the shift number of the offset.
	args = []arg{{Reg: rto, Type: obj.TYPE_REG}, {Offset: (con32 >> 16) & 0xffff, Type: obj.TYPE_CONST, Index: 16}}
	movkw := newInst(AMOVKW, MOVKwis, pos, args)
	return []inst{movzw, movkw}
}

// movCon32ToReg moves the 32-bit constant con32 into the rto register by
// MOVZW/MOVNW/MOVKW instructions. Note that the type of con32 is int64,
// which is for 32-bit BITCON checking.
func (c *ctxt7) movCon32ToReg(con32 int64, rto int16, pos src.XPos) []inst {
	v := uint32(con32)
	if isbitcon(uint64(con32)) {
		// -> MOVW $imm, rto
		return []inst{instIR(AMOVW, con32, rto, MOVwi_b, pos)}
	} else if movcon(int64(v)) >= 0 {
		// -> MOVZW $imm, rto
		return []inst{instIR(AMOVW, int64(v), rto, MOVZwis, pos)}
	} else if movcon(int64(^v)) >= 0 {
		// -> MOVNW $imm, rto
		return []inst{instIR(AMOVW, int64(^v), rto, MOVNwis, pos)}
	}
	// Any other 32-bit integer constant.
	// -> MOVZW $imm, rto + MOVKW $imm, rto+
	return c.movcon32(con32, rto, pos)
}

// bCond returns the corresponding conditon value of the B.cond instruction.
func (c *ctxt7) bCond(as obj.As) SpecialOperand {
	switch as {
	case ABEQ:
		return SPOP_EQ
	case ABNE:
		return SPOP_NE
	case ABCS, ABHS:
		return SPOP_HS
	case ABCC, ABLO:
		return SPOP_LO
	case ABMI:
		return SPOP_MI
	case ABPL:
		return SPOP_PL
	case ABVS:
		return SPOP_VS
	case ABVC:
		return SPOP_VC
	case ABHI:
		return SPOP_HI
	case ABLS:
		return SPOP_LS
	case ABGE:
		return SPOP_GE
	case ABLT:
		return SPOP_LT
	case ABGT:
		return SPOP_GT
	case ABLE:
		return SPOP_LE
		// BAL and BNV are not supported.
	}
	return SPOP_END
}

// instLoad generates a inst for a load instruction, such as MOVD, MOVW.
func instLoad(as obj.As, rf, rt, fromIndex int16, fromOffset int64, idx uint16, pos src.XPos) inst {
	args := []arg{{Type: obj.TYPE_REG, Reg: rt}, {Type: obj.TYPE_MEM, Reg: rf, Offset: fromOffset, Index: fromIndex}}
	return newInst(as, idx, pos, args)
}

// instLoadPair generates a inst for a LDP like instruction, such as LDP, LDPW.
func instLoadPair(as obj.As, rf, rt1, rt2 int16, fromOffset int64, idx uint16, pos src.XPos) inst {
	args := []arg{{Type: obj.TYPE_REGREG, Reg: rt1, Offset: int64(rt2)}, {Type: obj.TYPE_MEM, Reg: rf, Offset: fromOffset}}
	return newInst(as, idx, pos, args)
}

// instStore generates a inst for a store instruction, such as MOVD, MOVW.
func instStore(as obj.As, rf, rt, toIndex int16, toOffset int64, idx uint16, pos src.XPos) inst {
	args := []arg{{Type: obj.TYPE_REG, Reg: rf}, {Type: obj.TYPE_MEM, Reg: rt, Offset: toOffset, Index: toIndex}}
	return newInst(as, idx, pos, args)
}

// instStorePair generates a inst for a STP like instruction, such as STP, STPW.
func instStorePair(as obj.As, rt, rf1, rf2 int16, toOffset int64, idx uint16, pos src.XPos) inst {
	args := []arg{{Type: obj.TYPE_REGREG, Reg: rf1, Offset: int64(rf2)}, {Type: obj.TYPE_MEM, Reg: rt, Offset: toOffset}}
	return newInst(as, idx, pos, args)
}

// adrp generates a inst corresponding to the ADRP instruction.
// p.From.Offset is set to 0.
func adrp(rt int16, pos src.XPos) inst {
	args := []arg{{Reg: rt, Type: obj.TYPE_REG}, {Type: obj.TYPE_BRANCH, Offset: 0}}
	return newInst(AADRP, ADRPxl, pos, args)
}

// instIR generates a inst with two parameters, and the first parameter
// is constant type, the second is register types. idx is the corresponding
// index of arm64 instruction in optab.
func instIR(as obj.As, imm int64, rt int16, idx uint16, pos src.XPos) inst {
	args := []arg{{Reg: rt, Type: obj.TYPE_REG}, {Type: obj.TYPE_CONST, Offset: imm}}
	return newInst(as, idx, pos, args)
}

// instIRR generates a inst with three parameters, and the first parameter
// is constant type, the second and third parameters are register types.
// idx is the corresponding index of arm64 instruction in optab.
func instIRR(as obj.As, imm int64, rn, rt int16, idx uint16, pos src.XPos) inst {
	if rn == 0 {
		rn = rt
	}
	args := []arg{{Reg: rt, Type: obj.TYPE_REG}, {Reg: rn, Type: obj.TYPE_REG}, {Type: obj.TYPE_CONST, Offset: imm}}
	return newInst(as, idx, pos, args)
}

// inst2SR generates a inst with two source parameters of register type.
// The first operand is p.From and the second operand is p.Reg.
// idx is the corresponding index of arm64 instruction in optab.
func inst2SR(as obj.As, rm, rn int16, idx uint16, pos src.XPos) inst {
	args := []arg{{Reg: rn, Type: obj.TYPE_REG}, {Reg: rm, Type: obj.TYPE_REG}}
	return newInst(as, idx, pos, args)
}

// instRRR generates a inst with three register type parameters.
// idx is the corresponding index of arm64 instruction in optab.
func instRRR(as obj.As, rm, rn, rt int16, idx uint16, pos src.XPos) inst {
	if rn == 0 {
		rn = rt
	}
	args := []arg{{Reg: rt, Type: obj.TYPE_REG}, {Reg: rn, Type: obj.TYPE_REG}, {Reg: rm, Type: obj.TYPE_REG}}
	return newInst(as, idx, pos, args)
}

// newRelocation returns a newly created relocation with some fields being set with the arguments.
func newRelocation(f *obj.LSym, size uint8, sym *obj.LSym, add int64, typ objabi.RelocType) uint32 {
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
func addSub(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16, setflag, is32bit bool, movToReg func(int64, int16, src.XPos) []inst) []inst {
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
			return []inst{p1, p2}
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
			return nil
		}
		// MOVD $con, Rtmp + ADD/SUB Rtmp, Rn, Rt
		idx := sidx
		if p.Reg == REG_RSP || p.To.Reg == REG_RSP {
			idx = eidx
		}
		p1 := instRRR(p.As, REGTMP, p.Reg, p.To.Reg, idx, p.Pos)
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
func bitwiseOp(c *ctxt7, p *obj.Prog, cidx, sidx uint16, neg, supportZR bool, movToReg func(int64, int16, src.XPos) []inst) []inst {
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
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
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
func immOffsetStore(c *ctxt7, p *obj.Prog, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []inst, checkUnpredicate bool) []inst {
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
			c.ctxt.Diag("offset %d out of range [-256,255]: %v", p.To.Offset, p)
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
		return []inst{instStore(p.As, p.From.Reg, a.Reg, 0, v, uidx, p.Pos)}
	}

	// use Store Register (unscaled) instruction if -256 <= c.instoffset < 256
	if v >= -256 && v < 256 {
		return []inst{instStore(p.As, p.From.Reg, a.Reg, 0, v, idx256, p.Pos)}
	}
	// if offset v can be split into hi+lo, and both fit into instructions, convert
	// to ADD $hi, Rt, Rtmp + store Rs, lo(Rtmp)
	var p1, p2 inst
	v32 := int32(v)
	hi, ok := splitAddCon(c, p, v32, s)
	if !ok {
		goto storeusemov
	}
	p1 = instIRR(AADD, int64(hi), a.Reg, REGTMP, ADDxxis, p.Pos)
	p2 = instStore(p.As, p.From.Reg, REGTMP, 0, int64(v32-hi), uidx, p.Pos)
	return []inst{p1, p2}
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
func immOffsetLoad(c *ctxt7, p *obj.Prog, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []inst, checkUnpredicate bool) []inst {
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
			c.ctxt.Diag("offset %d out of range [-256,255]: %v", p.From.Offset, p)
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
		return []inst{instLoad(p.As, a.Reg, p.To.Reg, 0, v, uidx, p.Pos)}
	}
	// use Load Register (unscaled) instruction if -256 <= v < 256
	if v >= -256 && v < 256 {
		return []inst{instLoad(p.As, a.Reg, p.To.Reg, 0, v, idx256, p.Pos)}
	}
	// if offset v can be split into hi+lo, and both fit into instructions, do
	// MOVD $imm(Rs), Rt -> ADD $hi, Rs, Rtmp + load lo(Rtmp), Rt
	var p1, p2 inst
	v32 := int32(v)
	hi, ok := splitAddCon(c, p, v32, s)
	if !ok {
		goto loadusemov
	}
	p1 = instIRR(AADD, int64(hi), a.Reg, REGTMP, ADDxxis, p.Pos)
	p2 = instLoad(p.As, REGTMP, p.To.Reg, 0, int64(v32-hi), uidx, p.Pos)
	return []inst{p1, p2}
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
func generalStore(c *ctxt7, p *obj.Prog, toTyp argtype, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []inst, checkUnpredicate bool) []inst {
	switch toTyp {
	case C_ROFF:
		return op2ASO(p, eidx) // -> store (register)
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + store (immediate)
		if p.From.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
			return nil
		}
		c.rtab[p] = newRelocation(c.cursym, 8, p.To.Sym, p.To.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instStore(p.As, p.From.Reg, REGTMP, 0, 0, uidx, p.Pos)
		return []inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		return immOffsetStore(c, p, eidx, uidx, pidx, widx, idx256, movToReg, checkUnpredicate)
	}
	return nil
}

// generalLoad expands the load form of instructions MOVD, MOVW, etc.
// The meaning of the parameter is the same as generalStore.
func generalLoad(c *ctxt7, p *obj.Prog, fromTyp argtype, eidx, uidx, pidx, widx, idx256 uint16, movToReg func(int64, int16, src.XPos) []inst, checkUnpredicate bool) []inst {
	switch fromTyp {
	case C_ROFF:
		return op2A(p, eidx) // -> load (register)
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + load (REGTMP), Rt
		c.rtab[p] = newRelocation(c.cursym, 8, p.From.Sym, p.From.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, uidx, p.Pos)
		return []inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		return immOffsetLoad(c, p, eidx, uidx, pidx, widx, idx256, movToReg, checkUnpredicate)
	}
	return nil
}

// loadPair expands the load form of instructions LDP, LDPW, etc.
// "shift" is the shift value of the offset value.
func loadPair(c *ctxt7, p *obj.Prog, fromTyp argtype, uidx, pidx, widx uint16, shift uint) []inst {
	switch fromTyp {
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + LDP (Rtmp), (Rt1, Rt2)
		c.rtab[p] = newRelocation(c.cursym, 8, p.From.Sym, p.From.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instLoadPair(p.As, REGTMP, p.To.Reg, int16(p.To.Offset), 0, uidx, p.Pos)
		return []inst{p1, p2, p3}
	case C_ZOREG, C_LOREG, C_VOREG:
		a := p.From
		if p.Scond == C_XPOST || p.Scond == C_XPRE {
			if a.Reg == obj.REG_NONE { // pseudo registers, like FP, SP
				c.ctxt.Diag("pre and post index format don't support pseudo register: %v", p)
				return nil
			}
			if a.Offset < -64<<shift || a.Offset > 63<<shift {
				c.ctxt.Diag("offset %d out of range [%d,%d]: %v", a.Offset, -64<<shift, 63<<shift, p)
				return nil
			}
			if a.Offset&(1<<shift-1) != 0 {
				c.ctxt.Diag("offset %d must be a multiple of %d: %v", a.Offset, 1<<shift, p)
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
			return []inst{instLoadPair(p.As, a.Reg, p.To.Reg, int16(p.To.Offset), v, uidx, p.Pos)}
		}
		if p.To.Reg == REGTMP || int(p.To.Offset) == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
			return nil
		}
		if isaddcon(v) || isaddcon(-v) {
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
			return []inst{p1, p2}
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
func storePair(c *ctxt7, p *obj.Prog, toTyp argtype, uidx, pidx, widx uint16, shift uint) []inst {
	switch toTyp {
	case C_ADDR:
		// -> ADRP + ADD (immediate) + reloc + STP (Rt1, Rt2), (Rtmp)
		c.rtab[p] = newRelocation(c.cursym, 8, p.To.Sym, p.To.Offset, objabi.R_ADDRARM64)
		p1 := adrp(REGTMP, p.Pos)
		p2 := instIRR(AADD, 0, REGTMP, REGTMP, ADDxxis, p.Pos)
		p3 := instStorePair(p.As, REGTMP, p.From.Reg, int16(p.From.Offset), 0, uidx, p.Pos)
		return []inst{p1, p2, p3}
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
				c.ctxt.Diag("offset %d out of range [%d,%d]: %v", a.Offset, -64<<shift, 63<<shift, p)
				return nil
			}
			if a.Offset&(1<<shift-1) != 0 {
				c.ctxt.Diag("offset %d must be a multiple of %d: %v", a.Offset, 1<<shift, p)
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
			return []inst{instStorePair(p.As, a.Reg, p.From.Reg, int16(p.From.Offset), v, uidx, p.Pos)}
		}
		if p.From.Reg == REGTMP || int(p.From.Offset) == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
			return nil
		}
		if isaddcon(v) || isaddcon(-v) {
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
			return []inst{p1, p2}
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
func cmpCmn(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16) []inst {
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
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
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
func cmpCmn32(c *ctxt7, p *obj.Prog, cidx, sidx, eidx uint16) []inst {
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
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
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
	case AVTRN1:
		return TRN1vvv_t
	case AVTRN2:
		return TRN2vvv_t
	}
	return 0
}

// unfolding functions

func unfoldCall(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	switch tc {
	case C_SBRA:
		// DUFFCOPY/DUFFZERO/BL label -> BL <label>
		c.ctab[p] = progToInst(p, BLl, []arg{addrToArg(p.To)})
		p.Mark |= BRANCH26BITS
		if p.To.Sym != nil {
			c.rtab[p] = newRelocation(c.cursym, 4, p.To.Sym, p.To.Offset, objabi.R_CALLARM64)
		}
	case C_REG, C_ZOREG:
		// BL Rn -> BLR <Xn>
		// BL (Rn) or BL 0(Rn) -> BLR <Xn>
		c.ctab[p] = progToInst(p, BLRx, []arg{addrToArg(p.To)})
		c.rtab[p] = newRelocation(c.cursym, 0, nil, 0, objabi.R_CALLIND)
	}
}

func unfoldJMP(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	switch tc {
	case C_SBRA:
		// B label -> B <label>
		c.ctab[p] = progToInst(p, Bl, []arg{addrToArg(p.To)})
		p.Mark |= BRANCH26BITS
		if p.To.Sym != nil {
			c.rtab[p] = newRelocation(c.cursym, 4, p.To.Sym, p.To.Offset, objabi.R_CALLARM64)
		}
	case C_REG, C_ZOREG:
		// B Rn -> BR <Xn>
		// B (Rn) or BL 0(Rn) -> BR <Xn>
		c.ctab[p] = progToInst(p, BRx, []arg{addrToArg(p.To)})
	}
}

func unfoldRET(c *ctxt7, p *obj.Prog) {
	// RET -> RET {<Xn>}
	c.ctab[p] = progToInst(p, RETx, []arg{addrToArg(p.To)})
}

func unfoldUNDEF(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.To.Type == obj.TYPE_NONE) {
		return
	}
	c.ctab[p] = progToInst(p, UDFi, nil)
}

// adc/adcs/sbc/sbcs
func unfoldOpWithCarry(c *ctxt7, p *obj.Prog) {
	fc := c.aclass(p, &p.From)
	tc := c.aclass(p, &p.To)
	if !(fc == C_REG && tc == C_REG) {
		return
	}
	c.ctab[p] = op3A(p, opWithCarryIndex(p.As))
}

// ngc/ngcs
func unfoldNGCX(c *ctxt7, p *obj.Prog) {
	fc := c.aclass(p, &p.From)
	tc := c.aclass(p, &p.To)
	if !(fc == C_REG && p.Reg == 0 && tc == C_REG) {
		return
	}
	c.ctab[p] = op2A(p, opWithCarryIndex(p.As))
}

// ADD/ADDW/SUB/SUBW/ADDS/ADDSW/SUBS/SUBSW
func unfoldADD(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, ADDxxis, ADDxxxs, ADDxxre, false, false, c.movCon64ToReg)
}

func unfoldADDW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, ADDwwis, ADDwwws, ADDwwwe, false, true, c.movCon32ToReg)
}

func unfoldSUB(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, SUBxxis, SUBxxxs, SUBxxre, false, false, c.movCon64ToReg)
}

func unfoldSUBW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, SUBwwis, SUBwwws, SUBwwwe, false, true, c.movCon32ToReg)
}

func unfoldADDS(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, ADDSxxis, ADDSxxxs, ADDSxxre, true, false, c.movCon64ToReg)
}

func unfoldADDSW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, ADDSwwis, ADDSwwws, ADDSwwwe, true, true, c.movCon32ToReg)
}

func unfoldSUBS(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, SUBSxxis, SUBSxxxs, SUBSxxre, true, false, c.movCon64ToReg)
}

func unfoldSUBSW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = addSub(c, p, SUBSwwis, SUBSwwws, SUBSwwwe, true, true, c.movCon32ToReg)
}

// unfoldADRX deals with ADR and ADRP instructions.
func unfoldADRX(c *ctxt7, p *obj.Prog) {
	tc := c.aclass(p, &p.To)
	if !(tc == C_REG && p.From.Type == obj.TYPE_BRANCH) {
		return
	}
	if p.As == AADR {
		c.ctab[p] = op2A(p, ADRxl)
	} else if p.As == AADRP {
		c.ctab[p] = op2A(p, ADRPxl)
	}
}

// AND/ANDW/EOR/EORW/ORR/ORRW/BIC/BICW/EON/EONW/ORN/ORNW/ANDS/ANDSW/BICS/BICSW
func unfoldAND(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDxxi, ANDxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldANDW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDwwi, ANDwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldEOR(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, EORxxi, EORxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldEORW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, EORwwi, EORwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldORR(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ORRxxi, ORRxxxs, false, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldORRW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ORRwwi, ORRwwws, false, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldBIC(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDxxi, BICxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldBICW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDwwi, BICwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldEON(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, EORxxi, EONxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldEONW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, EORwwi, EONwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldORN(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ORRxxi, ORNxxxs, true, p.To.Reg != REGZERO, c.movCon64ToReg)
}

func unfoldORNW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ORRwwi, ORNwwws, true, p.To.Reg != REGZERO, c.movCon32ToReg)
}

func unfoldANDS(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDSxxi, ANDSxxxs, false, true, c.movCon64ToReg)
}

func unfoldANDSW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDSwwi, ANDSwwws, false, true, c.movCon32ToReg)
}

func unfoldBICS(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDSxxi, BICSxxxs, true, true, c.movCon64ToReg)
}

func unfoldBICSW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = bitwiseOp(c, p, ANDSwwi, BICSwwws, true, true, c.movCon32ToReg)
}

// TST/TSTW
func unfoldTST(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2SA(p, cidx) // -> TST/TSTW(immediate)
			return
		}
		if p.Reg == REGTMP {
			c.ctxt.Diag("cannot use REGTMP as source: %v", p)
			return
		}
		// MOVD $con, Rtmp + TST/TSTW Rtmp, Rt
		p1 := inst2SR(p.As, REGTMP, p.Reg, sidx, p.Pos)
		mov := movConToReg(v, REGTMP, p.Pos)
		mov = append(mov, p1)
		c.ctab[p] = mov
	case obj.TYPE_SHIFT, obj.TYPE_REG:
		c.ctab[p] = op2SA(p, sidx) // -> TST/TSTW(shift register)
	}
}

// unfoldBitFieldOps deals with bitfield operation instructions, such as BFM, BFI, etc.
func unfoldBitFieldOps(c *ctxt7, p *obj.Prog) {
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
	a := []arg{addrToArg(p.To), {Reg: r, Type: obj.TYPE_REG}, addrToArg(p.From), addrToArg(*from3)}
	c.ctab[p] = progToInst(p, bitFieldOpsIndex(p.As), a)
}

// load & store
func unfoldMOVD(c *ctxt7, p *obj.Prog) {
	// MOVD can be translated into several different kinds of instructions,
	// including MOV, LDR, STR, MSR etc.
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_SPR {
			c.ctab[p] = op2A(p, MSRi) // -> MSR (immediate)
			break
		}
		if ttyp != C_REG {
			break
		}
		v := p.From.Offset
		if movcon(v) >= 0 {
			c.ctab[p] = op2A(p, MOVZxis) // -> MOVZ $imm, R
		} else if movcon(^v) >= 0 {
			a := []arg{addrToArg(p.To), {Type: obj.TYPE_CONST, Offset: ^v}}
			c.ctab[p] = progToInst(p, MOVNxis, a) // -> MOVN $imm, R
		} else if isbitcon(uint64(v)) && p.To.Reg != REGZERO {
			c.ctab[p] = op2A(p, MOVxi_b) // -> MOV $bitcon, R
		} else { // Any other 64-bit integer constant.
			// -> MOVZ/MOVN $imm, R + (MOVK $imm, R)+
			mov := c.movcon64(v, p.To.Reg, p.Pos)
			c.ctab[p] = mov
			p.Mark |= NOTUSETMP
		}
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		ttyp := c.aclass(p, &p.To)
		if ftyp == C_SPR && ttyp == C_REG {
			c.ctab[p] = op2A(p, MRSx) // -> MRS
			break
		}
		if ftyp != C_REG {
			break
		}
		if ttyp == C_SPR {
			c.ctab[p] = op2A(p, MSRx) // -> MSR (register)
			break
		}
		if ttyp == C_REG {
			if p.From.Reg == REG_RSP || p.To.Reg == REG_RSP {
				c.ctab[p] = op2A(p, MOVxx_sp) // MOV (to/from SP)
			} else {
				c.ctab[p] = op2A(p, MOVxx) // MOV (register)
			}
			break
		}
		// Store
		c.ctab[p] = generalStore(c, p, ttyp, STRxxre, STRxx, STRxxi_p, STRxx_w, STURxx, c.movCon64ToReg, true)
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
			c.rtab[p] = newRelocation(c.cursym, 8, p.From.Sym, v, objabi.R_ADDRARM64)
			p1 := adrp(p.To.Reg, p.Pos)
			p2 := instIRR(AADD, 0, 0, p.To.Reg, ADDxxis, p.Pos)
			c.ctab[p] = []inst{p1, p2}
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
			c.ctab[p] = addSub(c, q, SUBxxis, SUBxxxs, SUBxxre, false, false, c.movCon64ToReg)
		} else {
			q.As = AADD
			q.From = obj.Addr{Type: obj.TYPE_CONST, Offset: v}
			c.ctab[p] = addSub(c, q, ADDxxis, ADDxxxs, ADDxxre, false, false, c.movCon64ToReg)
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
			c.rtab[p] = newRelocation(c.cursym, 8, p.From.Sym, 0, objabi.R_ARM64_GOTPCREL)
			p1 := adrp(REGTMP, p.Pos)
			p2 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, LDRxx, p.Pos)
			c.ctab[p] = []inst{p1, p2}
		case C_TLS_LE:
			// LE model MOVD $tlsvar, Rt -> MOVZ + reloc
			if p.From.Offset != 0 {
				c.ctxt.Diag("invalid offset %d: %v", p.From.Offset, p)
				return
			}
			c.rtab[p] = newRelocation(c.cursym, 4, p.From.Sym, 0, objabi.R_ARM64_TLS_LE)
			c.ctab[p] = progToInst(p, MOVZxis, []arg{addrToArg(p.To), {Type: obj.TYPE_CONST}})
		case C_TLS_IE:
			// IE model MOVD $tlsvar, Rt -> ADRP + LDR (REGTMP), Rt + relocs
			if p.From.Offset != 0 {
				c.ctxt.Diag("invalid offset %d: %v", p.From.Offset, p)
				return
			}
			c.rtab[p] = newRelocation(c.cursym, 8, p.From.Sym, 0, objabi.R_ARM64_TLS_IE)
			p1 := adrp(REGTMP, p.Pos)
			p2 := instLoad(p.As, REGTMP, p.To.Reg, 0, 0, LDRxx, p.Pos)
			c.ctab[p] = []inst{p1, p2}
		default:
			c.ctab[p] = generalLoad(c, p, ftyp, LDRxxre, LDRxx, LDRxxi_p, LDRxx_w, LDURxx, c.movCon64ToReg, true)
		}
	}
}

func unfoldMOVW(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	switch p.From.Type {
	case obj.TYPE_CONST:
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		v := uint32(p.From.Offset)
		if movcon(int64(v)) >= 0 {
			c.ctab[p] = op2A(p, MOVZwis) // -> MOVZW $imm, R
		} else if movcon(int64(^v)) >= 0 {
			a := []arg{addrToArg(p.To), {Type: obj.TYPE_CONST, Offset: int64(^v)}}
			c.ctab[p] = progToInst(p, MOVNwis, a) // -> MOVNW $imm, R
		} else if isbitcon(uint64(p.From.Offset)) && p.To.Reg != REGZERO {
			c.ctab[p] = op2A(p, MOVwi_b) // -> MOVW $bitcon, R
		} else { // Any other 32-bit integer constant.
			// -> MOVZW $imm, R + MOVKW $imm, R
			c.ctab[p] = c.movcon32(p.From.Offset, p.To.Reg, p.Pos)
			p.Mark |= NOTUSETMP
		}
	case obj.TYPE_REG:
		ftyp := rclass(p.From.Reg)
		if ftyp != C_REG {
			break
		}
		ttyp := c.aclass(p, &p.To)
		if ttyp == C_REG {
			c.ctab[p] = op2A(p, SXTWxw) // -> SXTW
			break
		}
		// Store
		c.ctab[p] = generalStore(c, p, ttyp, STRwxre, STRwx, STRwxi_p, STRwx_w, STURwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRSWxxre, LDRSWxx, LDRSWxxi_p, LDRSWxx_w, LDURSWxx, c.movCon32ToReg, true)
	}
}

func unfoldMOVWU(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2A(p, MOVww_sp) // -> MOVW (to/from SP)
			} else {
				c.ctab[p] = op2A(p, MOVww) // -> MOVW (register)
			}
			return
		}
		// Store, same as MOVW
		c.ctab[p] = generalStore(c, p, ttyp, STRwxre, STRwx, STRwxi_p, STRwx_w, STURwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRwxre, LDRwx, LDRwxi_p, LDRwx_w, LDURwx, c.movCon32ToReg, true)
	}
}

func unfoldMOVH(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2A(p, SXTHxw) // -> SXTH
			break
		}
		// Store
		c.ctab[p] = generalStore(c, p, ttyp, STRHwxre, STRHwx, STRHwxi_p, STRHwx_w, STURHwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRSHxxre, LDRSHxx, LDRSHxxi_p, LDRSHxx_w, LDURSHxx, c.movCon32ToReg, true)
	}
}

func unfoldMOVHU(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2A(p, UXTHww) // -> UXTH
			break
		}
		// Store, same as MOVH
		c.ctab[p] = generalStore(c, p, ttyp, STRHwxre, STRHwx, STRHwxi_p, STRHwx_w, STURHwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		ttyp := c.aclass(p, &p.To)
		if ttyp != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRHwxre, LDRHwx, LDRHwxi_p, LDRHwx_w, LDURHwx, c.movCon32ToReg, true)
	}
}

func unfoldMOVB(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2A(p, SXTBxw) // -> SXTB
			break
		}
		// Store
		c.ctab[p] = generalStore(c, p, ttyp, STRBwxre, STRBwx, STRBwxi_p, STRBwx_w, STURBwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRSBxxre, LDRSBxx, LDRSBxxi_p, LDRSBxx_w, LDURSBxx, c.movCon32ToReg, true)
	}
}

func unfoldMOVBU(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2A(p, UXTBww) // -> UXTB
			break
		}
		// Store, same as MOVB
		c.ctab[p] = generalStore(c, p, ttyp, STRBwxre, STRBwx, STRBwxi_p, STRBwx_w, STURBwx, c.movCon32ToReg, true)
	case obj.TYPE_MEM:
		// Load
		if c.aclass(p, &p.To) != C_REG {
			break
		}
		ftyp := c.aclass(p, &p.From)
		c.ctab[p] = generalLoad(c, p, ftyp, LDRBwxre, LDRBwx, LDRBwxi_p, LDRBwx_w, LDURBwx, c.movCon32ToReg, true)
	}
}

// floating point load/store
func unfoldFMOVQ(c *ctxt7, p *obj.Prog) {
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
			c.ctab[p] = op2ASO(p, STRqxre) // -> STR (register, SIMD&FP)
		case C_ZOREG, C_LOREG, C_VOREG:
			c.ctab[p] = immOffsetStore(c, p, STRqxre, STRqx, STRqxi_p, STRqx_w, STURqx, c.movCon64ToReg, false)
		}
	case C_ROFF:
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = op2A(p, LDRqxre) // -> LDR (register, SIMD&FP)
	case C_ZOREG, C_LOREG, C_VOREG:
		// Load
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = immOffsetLoad(c, p, LDRqxre, LDRqx, LDRqxi_p, LDRqx_w, LDURqx, c.movCon64ToReg, false)
	}
}

func unfoldFMOVD(c *ctxt7, p *obj.Prog) {
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
		c.ctab[p] = op2A(p, FMOVdi)
	case C_REG:
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = op2A(p, FMOVdx)
	case C_FREG:
		switch ttyp {
		case C_REG:
			c.ctab[p] = op2A(p, FMOVxd)
		case C_FREG:
			c.ctab[p] = op2A(p, FMOVdd)
		default:
			// Store
			c.ctab[p] = generalStore(c, p, ttyp, STRdxre, STRdx, STRdxi_p, STRdx_w, STURdx, c.movCon64ToReg, false)
		}
	default:
		// Load
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = generalLoad(c, p, ftyp, LDRdxre, LDRdx, LDRdxi_p, LDRdx_w, LDURdx, c.movCon64ToReg, false)
	}
}

func unfoldFMOVS(c *ctxt7, p *obj.Prog) {
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
		c.ctab[p] = op2A(p, FMOVsi)
	case C_REG:
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = op2A(p, FMOVsw)
	case C_FREG:
		switch ttyp {
		case C_REG:
			c.ctab[p] = op2A(p, FMOVws)
		case C_FREG:
			c.ctab[p] = op2A(p, FMOVss)
		default:
			// Store
			c.ctab[p] = generalStore(c, p, ttyp, STRsxre, STRsx, STRsxi_p, STRsx_w, STURsx, c.movCon32ToReg, false)
		}
	default:
		// Load
		if ttyp != C_FREG {
			break
		}
		c.ctab[p] = generalLoad(c, p, ftyp, LDRsxre, LDRsx, LDRsxi_p, LDRsx_w, LDURsx, c.movCon32ToReg, false)
	}
}

// load & store pair
func unfoldLDP(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPxxx, LDPxxxi_p, LDPxxx_w, 3)
}

func unfoldLDPW(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPwwx, LDPwwxi_p, LDPwwx_w, 2)
}

func unfoldLDPSW(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, p.Scond == C_XPOST || p.Scond == C_XPRE, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPSWxxx, LDPSWxxxi_p, LDPSWxxx_w, 2)
}

func unfoldFLDPQ(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPqqx, LDPqqxi_p, LDPqqx_w, 4)
}

func unfoldFLDPD(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPddx, LDPddxi_p, LDPddx_w, 3)
}

func unfoldFLDPS(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(ttyp == C_PAIR && p.Reg == 0) {
		return
	}
	if checkUnpredictable(true, false, p.From.Reg, p.To.Reg, int16(p.To.Offset)) {
		c.ctxt.Diag("constrained unpredictable behavior: %v", p)
		return
	}
	ftyp := c.aclass(p, &p.From)
	c.ctab[p] = loadPair(c, p, ftyp, LDPssx, LDPssxi_p, LDPssx_w, 2)
}

func unfoldSTP(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	c.ctab[p] = storePair(c, p, ttyp, STPxxx, STPxxxi_p, STPxxx_w, 3)
}

func unfoldSTPW(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	c.ctab[p] = storePair(c, p, ttyp, STPwwx, STPwwxi_p, STPwwx_w, 2)
}

func unfoldFSTPQ(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	c.ctab[p] = storePair(c, p, ttyp, STPqqx, STPqqxi_p, STPqqx_w, 4)
}

func unfoldFSTPD(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	c.ctab[p] = storePair(c, p, ttyp, STPddx, STPddxi_p, STPddx_w, 3)
}

func unfoldFSTPS(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_PAIR && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	c.ctab[p] = storePair(c, p, ttyp, STPssx, STPssxi_p, STPssx_w, 2)
}

// unfoldCBZX deals with CBZ/CBZW/CBNZ/CBNZW instructions.
func unfoldCBZX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2ASO(p, idx)
}

// unfoldCondCMP deals with CCMP, CCMPW, CCMN and CCMNW instructions.
func unfoldCondCMP(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = progToInst(p, idx, []arg{{Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(*from3), addrToArg(p.To), addrToArg(p.From)})
}

// unfoldCSELX deals with CSEL, CSINC, CSINV and CSNEG series instructions.
func unfoldCSELX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	from3 := p.GetFrom3()
	if !(ftyp == C_COND && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	idx := cselOpIndex(p.As)
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(p.To), {Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(*from3), addrToArg(p.From)})
}

// unfoldCSETX deals with CSET/CSETW/CSETM/CSETMW instructions.
func unfoldCSETX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_COND && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op2A(p, csetOpIndex(p.As))
}

// unfoldCINCX deals with CINC/CINV/CNEG series instructions.
func unfoldCINCX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op3A(p, idx)
}

// unfoldFloatingCondCMP deals with FCCMPD/FCCMPED series instructions.
func unfoldFloatingCondCMP(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(*from3), {Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(p.To), addrToArg(p.From)})
}
func unfoldCLREX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && (p.To.Type == obj.TYPE_CONST || p.To.Reg == REGZERO || p.To.Type == obj.TYPE_NONE)) {
		return
	}
	c.ctab[p] = progToInst(p, CLREXi, []arg{addrToArg(p.To)})
}

// unfoldCLSX deals with CLS/CLZ series instructions.
func unfoldCLSX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_REG && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op2A(p, CLSOpIndex(p.As))
}

// cmp/cmn
func unfoldCMP(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = cmpCmn(c, p, CMPxis, CMPxxs, CMPxre)
}

func unfoldCMPW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = cmpCmn32(c, p, CMPwis, CMPwws, CMPwwe)
}

func unfoldCMN(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = cmpCmn(c, p, CMNxis, CMNxxs, CMNxre)
}

func unfoldCMNW(c *ctxt7, p *obj.Prog) {
	c.ctab[p] = cmpCmn32(c, p, CMNwis, CMNwws, CMNwwe)
}

// unfoldDMBX deals with DMB/DSB/ISB series instructions.
func unfoldDMBX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(p.From)})
}

// unfoldOpwithoutArg deals with ERET/WFE/WFI series instructions that have no arguments.
func unfoldOpwithoutArg(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	c.ctab[p] = progToInst(p, opWithoutArgIndex(p.As), nil)
}

// unfoldEXTRX deals with EXTR and EXTRW instructions.
func unfoldEXTRX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(p.To), addrToArg(*from3), {Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(p.From)})
}

// unfoldLoadAcquire deals with LDAR/LDAXR/LDXR series instructions.
func unfoldLoadAcquire(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_ZOREG && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op2A(p, loadAcquireIndex(p.As))
}

// unfoldLDXPX deals with LDXP/LDAXP series instructions.
func unfoldLDXPX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2A(p, idx)
}

// unfoldSTLRX deals with STLR series instructions.
func unfoldSTLRX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2ASO(p, idx)
}

// unfoldSTXRX deals with STLXR/STXR series instructions.
func unfoldSTXRX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2D3A(p, idx)
}

// unfoldSTXPX deals with STXP/STLXP series instructions.
func unfoldSTXPX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2D3A(p, idx)
}

// unfoldShiftOp deals with LSL/LSR/ASR/ROR series instructions.
func unfoldShiftOp(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_REG || p.From.Type == obj.TYPE_CONST) && p.To.Type == obj.TYPE_REG) {
		return
	}
	// For LSL (immediate), imms != 0b111111, that means the shift number can't be 0.
	// Luckily we have converted $0 to ZR, so we won't encounter this problem.
	if p.From.Type == obj.TYPE_REG {
		c.ctab[p] = op3A(p, shiftRegIndex(p.As))
	} else {
		c.ctab[p] = op3A(p, shiftImmIndex(p.As))
	}
}

// unfoldFuseOp deals with MADD/MSUB/SMADDL/UMADDL/FMADD/FMSUB/FNMADD/FNMSUB series instructions.
func unfoldFuseOp(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	if !(p.From.Type == obj.TYPE_REG && p.Reg != 0 && from3 != nil && from3.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op3S4A(p, fuseOpIndex(p.As))
}

// unfoldMultipleX deals with MUL/MNEG/SMNEGL/UMNEGL/SMULH/SMULL/UMULH/UMULL series instructions.
func unfoldMultipleX(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_REG && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op3A(p, multipleOpIndex(p.As))
}

// unfoldMOVX deals with MOVK/MOVN/MOVZ series instructions.
func unfoldMOVX(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case AMOVK:
		// For details of this restriction, see CL 275812.
		if p.From.Offset == 0 {
			c.ctxt.Diag("zero shifts cannot be handled correctly: %v", p)
		}
		idx = MOVKxis
	case AMOVKW:
		// For details of this restriction, see CL 275812.
		if p.From.Offset == 0 {
			c.ctxt.Diag("zero shifts cannot be handled correctly: %v", p)
		}
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
	c.ctab[p] = op2A(p, idx)
}

// mrs
func unfoldMRS(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(ftyp == C_SPR && p.Reg == 0 && p.To.Type == obj.TYPE_REG) {
		return
	}
	c.ctab[p] = op2A(p, MRSx)
}

// msr
func unfoldMSR(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(p.Reg == 0 && (ttyp == C_SPR || ttyp == C_SPOP)) {
		return
	}
	if ftyp == C_REG {
		c.ctab[p] = op2A(p, MSRx)
	} else if p.From.Type == obj.TYPE_CONST {
		c.ctab[p] = op2A(p, MSRi)
	}
}

// unfoldMVN deals with MVN and MVNW instructions.
func unfoldMVN(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2A(p, idx)
}

// unfoldNEGX deals with NEG and NEGS series instructions.
func unfoldNEGX(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2A(p, idx)
}

// prfm
func unfoldPRFM(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.Reg == 0 && (ttyp == C_SPOP || p.To.Type == obj.TYPE_CONST)) {
		return
	}
	ftyp := c.aclass(p, &p.From)
	switch ftyp {
	case C_ZOREG, C_LOREG:
		c.ctab[p] = op2A(p, PRFMix)
	case C_SBRA:
		p.Mark |= BRANCH19BITS
		c.ctab[p] = op2A(p, PRFMil)
	case C_ROFF:
		c.ctab[p] = op2A(p, PRFMixre)
	}
}

// unfoldREMX deals with REM and UREM series instructions.
func unfoldREMX(c *ctxt7, p *obj.Prog) {
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
	p1 := newInst(p1As, p1Idx, p.Pos, []arg{{Reg: REGTMP, Type: obj.TYPE_REG}, {Reg: r, Type: obj.TYPE_REG}, addrToArg(p.From)})
	p2Args := []arg{{Reg: p.To.Reg, Type: obj.TYPE_REG}, {Reg: REGTMP, Type: obj.TYPE_REG},
		{Reg: p.From.Reg, Type: obj.TYPE_REG}, {Reg: r, Type: obj.TYPE_REG}}
	p2 := newInst(p2As, p2Idx, p.Pos, p2Args)
	c.ctab[p] = []inst{p1, p2}
}

// unfoldDIVCRC deals with SDIV and CRC series instructions.
func unfoldDIVCRC(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && ttyp == C_REG) {
		return
	}
	c.ctab[p] = op3A(p, divCRCIndex(p.As))
}

// unfoldPE deals with SVC, HVC, HLT, SMC, BRK, DCPS{1, 2, 3} etc. instructions.
func unfoldPE(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(p.From)})
}

// unfoldExtend deals with SXTB, SXTH series extend instructions.
func unfoldExtend(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	c.ctab[p] = op2A(p, extendIndex(p.As))
}

// unfoldUnsignedExtend deals with UXTB, UXTH and UXTW instructions.
func unfoldUnsignedExtend(c *ctxt7, p *obj.Prog) {
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
	args := []arg{addrToArg(p.To), {Reg: p.From.Reg, Type: obj.TYPE_REG},
		{Type: obj.TYPE_CONST, Offset: 0}, {Type: obj.TYPE_CONST, Offset: off}}
	c.ctab[p] = []inst{newInst(AUBFM, bitFieldOpsIndex(p.As), p.Pos, args)}
}

// sys
func unfoldSYS(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && (p.To.Type == obj.TYPE_NONE || p.To.Type == obj.TYPE_REG)) {
		return
	}
	if (p.From.Offset &^ int64(SYSARG4(0x7, 0xF, 0xF, 0x7))) != 0 {
		c.ctxt.Diag("illegal SYS argument %d: %v", p.From.Offset, p)
		return
	}
	// p.From.Offset integrated op1, Cn, Cm, Op2, in order to encode these arguments, p.From.Offset
	// needs to be split into multiple obj.Addr arguments.
	op1 := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 16) & 0x7}
	cn := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 12) & 0xf}
	cm := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 8) & 0xf}
	// op2{, <Xt>}, integrate p.To.Reg into op2 because there's only one argument in optab for them.
	op2 := arg{Type: obj.TYPE_CONST, Index: p.To.Reg, Offset: (p.From.Offset >> 5) & 0x7}
	c.ctab[p] = progToInst(p, SYSix, []arg{op1, cn, cm, op2})
}

// sysl
func unfoldSYSL(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	if (p.From.Offset &^ int64(SYSARG4(0x7, 0xF, 0xF, 0x7))) != 0 {
		c.ctxt.Diag("illegal SYSL argument %d: %v", p.From.Offset, p)
		return
	}
	// p.From.Offset integrated op1, Cn, Cm, Op2, in order to encode these arguments, p.From.Offset
	// needs to be split into multiple obj.Addr arguments.
	xt := arg{Type: obj.TYPE_REG, Reg: p.To.Reg}
	op1 := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 16) & 0x7}
	cn := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 12) & 0xf}
	cm := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 8) & 0xf}
	op2 := arg{Type: obj.TYPE_CONST, Offset: (p.From.Offset >> 5) & 0x7}
	c.ctab[p] = progToInst(p, SYSLix, []arg{xt, op1, cn, cm, op2})
}

// unfoldATIC deals with SYS alias instructions AT and IC.
func unfoldATIC(c *ctxt7, p *obj.Prog) {
	// TODO: The existence of the destination register is based on p.From.Offset.
	// But we have not double-checked the value at the moment.
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0 && (p.To.Type == obj.TYPE_NONE || p.To.Type == obj.TYPE_REG)) {
		return
	}
	idx, fields := uint16(0), int64(0)
	switch p.As {
	case AAT:
		idx, fields = ATx, 0x7<<16|0<<12|1<<8|0x7<<5
	case AIC:
		idx, fields = ICix, 0x7<<16|0<<12|0xF<<8|0x7<<5
	}
	if (p.From.Offset &^ fields) != 0 {
		c.ctxt.Diag("illegal system instruction argument %d: %v", p.From.Offset, p)
		return
	}
	op := addrToArg(p.From)
	// integrate Xt into op because there is only
	// one corresponding argument in optab.
	op.Index = p.To.Reg
	c.ctab[p] = progToInst(p, idx, []arg{op})
}

// unfoldSYSAlias deals with SYS alias instructions DC and TLBI.
func unfoldSYSAlias(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !((ftyp == C_SPOP) && p.Reg == 0 && (p.To.Type == obj.TYPE_NONE || p.To.Type == obj.TYPE_REG)) {
		return
	}
	idx := uint16(0)
	switch p.As {
	case ADC:
		idx = DCx
	case ATLBI:
		idx = TLBIix
	}
	op := addrToArg(p.From)
	// integrate Xt into op because there is only
	// one corresponding argument in optab.
	op.Index = p.To.Reg
	c.ctab[p] = progToInst(p, idx, []arg{op})
}

// unfoldTBZX deals with TBZ and TBNZ instructions.
func unfoldTBZX(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg != 0 && p.To.Type == obj.TYPE_BRANCH) {
		return
	}
	p.Mark |= BRANCH14BITS
	idx := TBZril
	if p.As == ATBNZ {
		idx = TBNZril
	}
	c.ctab[p] = progToInst(p, idx, []arg{{Reg: p.Reg, Type: obj.TYPE_REG}, addrToArg(p.From), addrToArg(p.To)})
}

// unfoldAtomicLoadOpStore deals with LDADD, LDEOR series atomic instructions.
func unfoldAtomicLoadOpStore(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	t2typ := rclass(p.RegTo2)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_ZOREG && t2typ == C_REG) {
		return
	}
	args := []arg{addrToArg(p.From), {Reg: p.RegTo2, Type: obj.TYPE_REG}, addrToArg(p.To)}
	c.ctab[p] = progToInst(p, atomicIndex(p.As), args)
}

// unfoldAtomicLoadOpStore deals with CASP series atomic instructions.
func unfoldCASPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	to2 := p.GetTo2()
	t2typ := c.aclass(p, to2)
	if !(ftyp == C_PAIR && p.Reg == 0 && ttyp == C_ZOREG && t2typ == C_PAIR) {
		return
	}
	args := []arg{addrToArg(p.From), addrToArg(*to2), addrToArg(p.To)}
	c.ctab[p] = progToInst(p, atomicIndex(p.As), args)
}

// unfoldBCOND deals with B.<cond> series instructions.
func unfoldBCOND(c *ctxt7, p *obj.Prog) {
	if !(p.From.Type == obj.TYPE_NONE && p.Reg == 0 && p.To.Type == obj.TYPE_BRANCH) {
		return
	}
	p.Mark |= BRANCH19BITS
	args := []arg{{Offset: int64(c.bCond(p.As))}, addrToArg(p.To)}
	c.ctab[p] = progToInst(p, Bcl, args)
}

// unfoldFloatingOp deals with some floating point instructions such as FADD, FSUB etc.
func unfoldFloatingOp(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && ttyp == C_FREG) {
		return
	}
	c.ctab[p] = op3A(p, floatingOpIndex(p.As))
}

// unfoldFCMPX deals with FCMPD, FCMPED series instructions.
func unfoldFCMPX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	if !(p.Reg != 0 && p.To.Type == obj.TYPE_NONE) {
		return
	}
	if ftyp == C_FREG {
		c.ctab[p] = op2SA(p, fcmpRegIndex(p.As))
	} else if ftyp == C_FCON {
		c.ctab[p] = op2SA(p, fcmpImmIndex(p.As))
	}
}

// unfoldFConvertRounding deals with floating point conversion and rounding series instructions.
func unfoldFConvertRounding(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && p.Reg == 0 && ttyp == C_FREG) {
		return
	}
	c.ctab[p] = op2A(p, fConvertRoundingIndex(p.As))
}

// unfoldFConvertToFixed deals with floating point convert to fixed-point series instructions.
func unfoldFConvertToFixed(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_FREG && p.Reg == 0 && ttyp == C_REG) {
		return
	}
	c.ctab[p] = op2A(p, fConvertRoundingIndex(p.As))
}

// unfoldFixedToFloating deals with fixed-point convert to floating point series instructions.
func unfoldFixedToFloating(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_REG && p.Reg == 0 && ttyp == C_FREG) {
		return
	}
	c.ctab[p] = op2A(p, fConvertRoundingIndex(p.As))
}

// unfoldAESSHA deals with some AES and SHA instructions that support
// 'opcode VREG, VREG' and 'opcode ARNG, ARNG' formats.
func unfoldAESSHA(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	af, at := int((p.From.Reg>>5)&15), int((p.To.Reg>>5)&15)
	// support the C_VREG format for compatibility with old code.
	if !((ftyp == C_VREG && ttyp == C_VREG || ftyp == C_ARNG && ttyp == C_ARNG && af == at) && p.Reg == 0) {
		return
	}
	c.ctab[p] = op2A(p, cryptoOpIndex(p.As))
}

// unfoldFBitwiseOp deals with some floating point bitwise operation instructions such as VREV16 and VCNT.
func unfoldFBitwiseOp(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op2A(p, idx)
}

// sha1h
func unfoldSHA1H(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_VREG && ttyp == C_VREG && p.Reg == 0) {
		return
	}
	from, to := addrToArg(p.From), addrToArg(p.To)
	// SHA1H uses the V register in the Go assembly syntax, but in fact it uses the
	// floating point register in the Arm assembly syntax.
	from.Reg = p.From.Reg - REG_V0 + REG_F0
	to.Reg = p.To.Reg - REG_V0 + REG_F0
	c.ctab[p] = progToInst(p, SHA1Hss, []arg{to, from})
}

// unfoldSHAX deals with some SHA algorithm related instructions.
func unfoldSHAX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	// SHA1C once supported the "ASHA1C C_VREG, C_REG, C_VREG" format, but it is obviously
	// very strange. Do we need to continue to support it for compatibility ?
	if !(ftyp == C_ARNG && f2typ == C_VREG && ttyp == C_VREG) {
		return
	}
	from, to := addrToArg(p.From), addrToArg(p.To)
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax for encoding.
	to.Reg = p.To.Reg - REG_V0 + REG_F0
	c.ctab[p] = progToInst(p, cryptoOpIndex(p.As), []arg{to, {Type: obj.TYPE_REG, Reg: p.Reg - REG_V0 + REG_F0}, from})
}

// unfoldARNG3 deals with instructions with three ARNG format operands.
func unfoldARNG3(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	f2typ := rclass(p.Reg)
	ttyp := c.aclass(p, &p.To)
	af, an, at := int((p.From.Reg>>5)&15), int((p.Reg>>5)&15), int((p.To.Reg>>5)&15)
	if !(ftyp == C_ARNG && f2typ == C_ARNG && ttyp == C_ARNG && af == an && an == at) {
		return
	}
	c.ctab[p] = op3A(p, arng3Index(p.As))
}

// unfoldVPMULLX deals with VPMULL and VPMULL2 instructions.
func unfoldVPMULLX(c *ctxt7, p *obj.Prog) {
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
		c.ctxt.Diag("arrangements %v and %v dismatch: %v", arrange(int(ta)), arrange(int(tb)), p)
		return
	}
	c.ctab[p] = op3A(p, idx)
}

// unfoldVUADDWX deals with VUADDW and VUADDW2 instructions.
func unfoldVUADDWX(c *ctxt7, p *obj.Prog) {
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
		c.ctxt.Diag("arrangements %v and %v dismatch: %v", arrange(int(ta)), arrange(int(tb)), p)
		return
	}
	c.ctab[p] = op3A(p, idx)
}

// unfoldVReg3OrARNG3 deals with instructions with three VREG or ARNG format operands.
func unfoldVReg3OrARNG3(c *ctxt7, p *obj.Prog) {
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
		c.ctab[p] = op3A(p, idx)
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
		c.ctab[p] = op3A(p, idx)
	}
}

// unfoldARNG4 deals with instructions with four ARNG format operands.
func unfoldARNG4(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op4A(p, idx)
}

// ext
func unfoldVEXT(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	f3typ := c.aclass(p, p.GetFrom3())
	ttyp := c.aclass(p, &p.To)
	at, an, af3 := (p.To.Reg>>5)&15, (p.Reg>>5)&15, (p.GetFrom3().Reg>>5)&15
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && f3typ == C_ARNG && ttyp == C_ARNG && at == an && at == af3) {
		return
	}
	c.ctab[p] = op4A(p, EXTvvvi_t)
	// record the arrangement specifier in p.From.Index so that we know the value when encoding.
	c.ctab[p][0].Args[3].Index = at
}

// xar
func unfoldVXAR(c *ctxt7, p *obj.Prog) {
	ntyp := rclass(p.Reg)
	f3typ := c.aclass(p, p.GetFrom3())
	ttyp := c.aclass(p, &p.To)
	at, an, af3 := (p.To.Reg>>5)&15, (p.Reg>>5)&15, (p.GetFrom3().Reg>>5)&15
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && ntyp == C_ARNG && f3typ == C_ARNG && ttyp == C_ARNG && at == an && at == af3) {
		return
	}
	c.ctab[p] = op4A(p, XARvvvi_t)
}

// vmov
func unfoldVMOV(c *ctxt7, p *obj.Prog) {
	if p.Reg != 0 {
		return
	}
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	switch ftyp {
	case C_REG:
		switch ttyp {
		case C_ARNG:
			c.ctab[p] = op2A(p, DUPvr_t)
		case C_ELEM:
			c.ctab[p] = op2A(p, MOVvr_ti)
		}
	case C_ARNG:
		if ttyp != C_ARNG {
			break
		}
		// VMOV <Vn>.<T>, <Vd>.<T> -> VORR <Vn>.<T>, <Vn>.<T>, <Vd>.<T>
		c.ctab[p] = []inst{newInst(AVORR, ORRvvv_t, p.Pos, []arg{addrToArg(p.To), {Reg: p.From.Reg, Type: obj.TYPE_REG}, addrToArg(p.From)})}
	case C_ELEM:
		switch ttyp {
		case C_REG:
			switch (p.From.Reg >> 5) & 15 {
			case ARNG_B, ARNG_H:
				c.ctab[p] = op2A(p, UMOVwv_ti)
			case ARNG_S:
				c.ctab[p] = op2A(p, MOVwv_si)
			case ARNG_D:
				c.ctab[p] = op2A(p, MOVxv_di)
			}
		case C_VREG:
			c.ctab[p] = op2A(p, MOVvv_ti)
		case C_ELEM:
			ta, tb := int((p.From.Reg>>5)&15), int((p.To.Reg>>5)&15)
			if ta != tb {
				c.ctxt.Diag("arrangements %v and %v mismatch: %v", arrange(ta), arrange(tb), p)
				return
			}
			c.ctab[p] = op2A(p, MOVvv_tii)
		}
	}
}

// vmovq/vmovd/vmovs
func unfoldVMOVQ(c *ctxt7, p *obj.Prog) {
	from3 := p.GetFrom3()
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && from3 != nil && from3.Type == obj.TYPE_CONST && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM128
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	c.ctab[p] = progToInst(p, LDRql, []arg{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, addrToArg(p.From)})
}

func unfoldVMOVD(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	c.ctab[p] = progToInst(p, LDRdl, []arg{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, addrToArg(p.From)})
}

func unfoldVMOVS(c *ctxt7, p *obj.Prog) {
	ttyp := c.aclass(p, &p.To)
	if !(p.From.Type == obj.TYPE_CONST && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	p.Mark |= LFROM
	// Adjust the registers used in Go assembly syntax to be consistent with the registers
	// used in Arm assembly syntax.
	c.ctab[p] = progToInst(p, LDRsl, []arg{{Reg: p.To.Reg - REG_V0 + REG_F0, Type: obj.TYPE_REG}, addrToArg(p.From)})
}

// ld1/ld2/ld3/ld4/st1/st2/st3/st4
func unfoldVLD1(c *ctxt7, p *obj.Prog) {
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
					c.ctab[p] = op2A(p, LD1vxi_tp1) // Post-index
				} else {
					c.ctab[p] = op2A(p, LD1vx_t1) // one register
				}
			case 0xa:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2A(p, LD1vxi_tp2)
				} else {
					c.ctab[p] = op2A(p, LD1vx_t2) // two registers
				}
			case 0x6:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2A(p, LD1vxi_tp3)
				} else {
					c.ctab[p] = op2A(p, LD1vx_t3) // three registers
				}
			case 0x2:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2A(p, LD1vxi_tp4)
				} else {
					c.ctab[p] = op2A(p, LD1vx_t4) // four registers
				}
			default:
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			switch opcode {
			case 0x7:
				c.ctab[p] = op2A(p, LD1vxx_tp1) // one register
			case 0xa:
				c.ctab[p] = op2A(p, LD1vxx_tp2) // two registers
			case 0x6:
				c.ctab[p] = op2A(p, LD1vxx_tp3) // three registers
			case 0x2:
				c.ctab[p] = op2A(p, LD1vxx_tp4) // four registers
			default:
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
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
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
				return
			}
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2A(p, index)
		}
	case C_ELEM:
		// single structure
		at := (p.To.Reg >> 5) & 15
		switch ftyp {
		case C_ZOREG:
			switch at {
			case ARNG_B:
				c.ctab[p] = op2A(p, LD1vx_bi1)
			case ARNG_H:
				c.ctab[p] = op2A(p, LD1vx_hi1)
			case ARNG_S:
				c.ctab[p] = op2A(p, LD1vx_si1)
			case ARNG_D:
				c.ctab[p] = op2A(p, LD1vx_di1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement %v: %v", arrange(int(at)), p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			switch at {
			case ARNG_B:
				c.ctab[p] = op2A(p, LD1vxx_bip1)
			case ARNG_H:
				c.ctab[p] = op2A(p, LD1vxx_hip1)
			case ARNG_S:
				c.ctab[p] = op2A(p, LD1vxx_sip1)
			case ARNG_D:
				c.ctab[p] = op2A(p, LD1vxx_dip1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement %v: %v", arrange(int(at)), p)
			}
		case C_LOREG:
			if p.Scond != C_XPOST {
				break
			}
			switch at {
			case ARNG_B:
				c.ctab[p] = op2A(p, LD1vxi_bip1)
			case ARNG_H:
				c.ctab[p] = op2A(p, LD1vxi_hip1)
			case ARNG_S:
				c.ctab[p] = op2A(p, LD1vxi_sip1)
			case ARNG_D:
				c.ctab[p] = op2A(p, LD1vxi_dip1)
			default:
				c.ctxt.Diag("illegal destination operand arrangement %v: %v", arrange(int(at)), p)
			}
		}
	}
}

func unfoldVST1(c *ctxt7, p *obj.Prog) {
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
					c.ctab[p] = op2ASO(p, ST1vxi_tp1) // Post-index
				} else {
					c.ctab[p] = op2ASO(p, ST1vx_t1) // one register
				}
			case 0xa:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2ASO(p, ST1vxi_tp2)
				} else {
					c.ctab[p] = op2ASO(p, ST1vx_t2) // two registers
				}
			case 0x6:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2ASO(p, ST1vxi_tp3)
				} else {
					c.ctab[p] = op2ASO(p, ST1vx_t3) // three registers
				}
			case 0x2:
				if p.Scond == C_XPOST {
					c.ctab[p] = op2ASO(p, ST1vxi_tp4)
				} else {
					c.ctab[p] = op2ASO(p, ST1vx_t4) // four registers
				}
			default:
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			switch opcode {
			case 0x7:
				c.ctab[p] = op2ASO(p, ST1vxx_tp1) // one register
			case 0xa:
				c.ctab[p] = op2ASO(p, ST1vxx_tp2) // two registers
			case 0x6:
				c.ctab[p] = op2ASO(p, ST1vxx_tp3) // three registers
			case 0x2:
				c.ctab[p] = op2ASO(p, ST1vxx_tp4) // four registers
			default:
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
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
				c.ctxt.Diag("invalid register numbers in register list: %v", p)
				return
			}
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2ASO(p, index)
		}
	case C_ELEM:
		// single structure
		af := (p.From.Reg >> 5) & 15
		switch ttyp {
		case C_ZOREG:
			switch af {
			case ARNG_B:
				c.ctab[p] = op2ASO(p, ST1vx_bi1)
			case ARNG_H:
				c.ctab[p] = op2ASO(p, ST1vx_hi1)
			case ARNG_S:
				c.ctab[p] = op2ASO(p, ST1vx_si1)
			case ARNG_D:
				c.ctab[p] = op2ASO(p, ST1vx_di1)
			default:
				c.ctxt.Diag("illegal source operand arrangement %v: %v", arrange(int(af)), p)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			switch af {
			case ARNG_B:
				c.ctab[p] = op2ASO(p, ST1vxx_bip1)
			case ARNG_H:
				c.ctab[p] = op2ASO(p, ST1vxx_hip1)
			case ARNG_S:
				c.ctab[p] = op2ASO(p, ST1vxx_sip1)
			case ARNG_D:
				c.ctab[p] = op2ASO(p, ST1vxx_dip1)
			default:
				c.ctxt.Diag("illegal source operand arrangement %v: %v", arrange(int(af)), p)
			}
		case C_LOREG:
			if p.Scond != C_XPOST {
				break
			}
			switch af {
			case ARNG_B:
				c.ctab[p] = op2ASO(p, ST1vxi_bip1)
			case ARNG_H:
				c.ctab[p] = op2ASO(p, ST1vxi_hip1)
			case ARNG_S:
				c.ctab[p] = op2ASO(p, ST1vxi_sip1)
			case ARNG_D:
				c.ctab[p] = op2ASO(p, ST1vxi_dip1)
			default:
				c.ctxt.Diag("illegal source operand arrangement %v: %v", arrange(int(af)), p)
			}
		}
	}
}

func unfoldVLD2(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2A(p, LD2vxi_tp2) // Post-index
			} else {
				c.ctab[p] = op2A(p, LD2vx_t2)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			c.ctab[p] = op2A(p, LD2vxx_tp2)
		case C_LOREG:
			regSize, q := int64(2), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2A(p, LD2vxi_tp2)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func unfoldVLD3(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2A(p, LD3vxi_tp3) // Post-index
			} else {
				c.ctab[p] = op2A(p, LD3vx_t3)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			c.ctab[p] = op2A(p, LD3vxx_tp3)
		case C_LOREG:
			regSize, q := int64(3), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2A(p, LD3vxi_tp3)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func unfoldVLD4(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2A(p, LD4vxi_tp4) // Post-index
			} else {
				c.ctab[p] = op2A(p, LD4vx_t4)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
				break
			}
			c.ctab[p] = op2A(p, LD4vxx_tp4)
		case C_LOREG:
			regSize, q := int64(4), (p.To.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.From.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2A(p, LD4vxi_tp4)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func unfoldVST2(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2ASO(p, ST2vxi_tp2) // Post-index
			} else {
				c.ctab[p] = op2ASO(p, ST2vx_t2)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			c.ctab[p] = op2ASO(p, ST2vxx_tp2)
		case C_LOREG:
			regSize, q := int64(2), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2ASO(p, ST2vxi_tp2)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func unfoldVST3(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2ASO(p, ST3vxi_tp3) // Post-index
			} else {
				c.ctab[p] = op2ASO(p, ST3vx_t3)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			c.ctab[p] = op2ASO(p, ST3vxx_tp3)
		case C_LOREG:
			regSize, q := int64(3), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2ASO(p, ST3vxi_tp3)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

func unfoldVST4(c *ctxt7, p *obj.Prog) {
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
				c.ctab[p] = op2ASO(p, ST4vxi_tp4) // Post-index
			} else {
				c.ctab[p] = op2ASO(p, ST4vx_t4)
			}
		case C_ROFF:
			if p.Scond != C_XPOST || isRegShiftOrExt(p.To.Index) {
				break
			}
			c.ctab[p] = op2ASO(p, ST4vxx_tp4)
		case C_LOREG:
			regSize, q := int64(4), (p.From.Offset>>30)&1
			if !(p.Scond == C_XPOST && immRegSizeQMatch(p.To.Offset, regSize, q)) {
				break
			}
			c.ctab[p] = op2ASO(p, ST4vxi_tp4)
		}
	case C_ELEM:
		// TODO: single structure
	}
}

// ld1r/ld2r/ld3r/ld4r
func unfoldVLD1R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x7) {
		return
	}
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			c.ctab[p] = op2A(p, LD1Rvxi_tp1) // Post-index
		} else {
			c.ctab[p] = op2A(p, LD1Rvx_t1)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		c.ctab[p] = op2A(p, LD1Rvxx_tp1)
	case C_LOREG:
		regSize, size := int64(1), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		c.ctab[p] = op2A(p, LD1Rvxi_tp1)
	}
}

func unfoldVLD2R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0xa) {
		return
	} // two registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			c.ctab[p] = op2A(p, LD2Rvxi_tp2) // Post-index
		} else {
			c.ctab[p] = op2A(p, LD2Rvx_t2)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		c.ctab[p] = op2A(p, LD2Rvxx_tp2)
	case C_LOREG:
		regSize, size := int64(2), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		c.ctab[p] = op2A(p, LD2Rvxi_tp2)
	}
}

func unfoldVLD3R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x6) {
		return
	} // three registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			c.ctab[p] = op2A(p, LD3Rvxi_tp3) // Post-index
		} else {
			c.ctab[p] = op2A(p, LD3Rvx_t3)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		c.ctab[p] = op2A(p, LD3Rvxx_tp3)
	case C_LOREG:
		regSize, size := int64(3), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		c.ctab[p] = op2A(p, LD3Rvxi_tp3)
	}
}

func unfoldVLD4R(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	opcode := (p.To.Offset >> 12) & 15
	if !(p.Reg == 0 && ttyp == C_LIST && opcode == 0x2) {
		return
	} // four registers
	switch ftyp {
	case C_ZOREG:
		if p.Scond == C_XPOST {
			c.ctab[p] = op2A(p, LD4Rvxi_tp4) // Post-index
		} else {
			c.ctab[p] = op2A(p, LD4Rvx_t4)
		}
	case C_ROFF:
		if p.Scond != C_XPOST || isRegShiftOrExt(p.From.Index) {
			break
		}
		c.ctab[p] = op2A(p, LD4Rvxx_tp4)
	case C_LOREG:
		regSize, size := int64(4), uint(p.To.Offset>>10)&3
		if !(p.Scond == C_XPOST && p.From.Offset == regSize<<size) {
			break
		}
		c.ctab[p] = op2A(p, LD4Rvxi_tp4)
	}
}

// vdup
func unfoldVDUP(c *ctxt7, p *obj.Prog) {
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
		c.ctab[p] = op2A(p, DUPvr_t)
	case C_ELEM:
		switch ttyp {
		case C_VREG:
			c.ctab[p] = op2A(p, DUPvv_i)
		case C_ARNG:
			Ts, T := (p.From.Reg>>5)&15, (p.To.Reg>>5)&15
			if !arngTsMatchT(Ts, T) {
				c.ctxt.Diag("arrangements %v and %v dismatch: %v", arrange(int(Ts)), arrange(int(T)), p)
				return
			}
			c.ctab[p] = op2A(p, DUPvv_ti)
		}
	}
}

// uxtl/uxtl2
func unfoldVUXTLX(c *ctxt7, p *obj.Prog) {
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
		c.ctxt.Diag("arrangements %v and %v dismatch: %v", arrange(int(Ta)), arrange(int(Tb)), p)
		return
	}
	c.ctab[p] = op2A(p, idx)
}

// unfoldVUSHLLX deals with VUSHLL and VUSHLL2 instructions.
func unfoldVUSHLLX(c *ctxt7, p *obj.Prog) {
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
		c.ctxt.Diag("arrangements %v and %v dismatch: %v", arrange(int(Ta)), arrange(int(Tb)), p)
		return
	}
	if shift < 0 || shift > max {
		c.ctxt.Diag("shift amount out of range: %v", p)
		return
	}
	c.ctab[p] = op3A(p, idx)
}

// unfoldVectorShift deals with some vector shift instructions such as SHL, SLI.
func unfoldVectorShift(c *ctxt7, p *obj.Prog) {
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
	c.ctab[p] = op3A(p, idx)
}

// tbl
func unfoldVTBL(c *ctxt7, p *obj.Prog) {
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
		c.ctxt.Diag("invalid register numbers in register list: %v", p)
		return
	}
	c.ctab[p] = progToInst(p, idx, []arg{addrToArg(p.To), addrToArg(*from3), addrToArg(p.From)})
}

// unfoldVADDVX deals with VADDV and VUADDLV instructions.
func unfoldVADDVX(c *ctxt7, p *obj.Prog) {
	ftyp := c.aclass(p, &p.From)
	ttyp := c.aclass(p, &p.To)
	if !(ftyp == C_ARNG && p.Reg == 0 && ttyp == C_VREG) {
		return
	}
	if p.As == AVADDV {
		c.ctab[p] = op2A(p, ADDVvv_t)
	} else {
		c.ctab[p] = op2A(p, UADDLVvv_t)
	}
}

// vmovi
func unfoldVMOVI(c *ctxt7, p *obj.Prog) {
	if !((p.From.Type == obj.TYPE_CONST || p.From.Reg == REGZERO) && p.Reg == 0) {
		return
	}
	ttyp := c.aclass(p, &p.To)
	switch ttyp {
	case C_FREG: // MOVI $imm, Fd
		c.ctab[p] = op2A(p, MOVIdi)
	case C_ARNG:
		from3 := p.GetFrom3()
		if from3 != nil {
			if from3.Type != obj.TYPE_CONST {
				break
			}
			// VMOVI $<imm8>, $<amount>, <Vd>.<T> -> MOVI <Vd>.<T>, #<imm8>, MSL #<amount>
			c.ctab[p] = progToInst(p, MOVIvii_ts, []arg{addrToArg(p.To), addrToArg(p.From), addrToArg(*from3)})
			break
		}
		at := int((p.To.Reg >> 5) & 15)
		switch at {
		case ARNG_2D:
			c.ctab[p] = op2A(p, MOVIvi)
		case ARNG_2S, ARNG_4S:
			c.ctab[p] = op2A(p, MOVIvi_ts)
		case ARNG_4H, ARNG_8H:
			c.ctab[p] = op2A(p, MOVIvi_th)
		case ARNG_8B, ARNG_16B:
			c.ctab[p] = op2A(p, MOVIvi_tb)
		default:
			c.ctxt.Diag("illegal destination operand arrangement %v: %v", arrange(at), p)
		}
	}
}

type unfoldFunc func(c *ctxt7, p *obj.Prog)

// unfolding function table
var unfoldTab = []unfoldFunc{
	obj.ACALL:     unfoldCall,
	obj.ADUFFCOPY: unfoldCall,
	obj.ADUFFZERO: unfoldCall,
	obj.AJMP:      unfoldJMP,
	obj.ARET:      unfoldRET,
	obj.AUNDEF:    unfoldUNDEF,

	AADC - obj.ABaseARM64:   unfoldOpWithCarry,
	AADCW - obj.ABaseARM64:  unfoldOpWithCarry,
	AADCS - obj.ABaseARM64:  unfoldOpWithCarry,
	AADCSW - obj.ABaseARM64: unfoldOpWithCarry,
	ASBC - obj.ABaseARM64:   unfoldOpWithCarry,
	ASBCW - obj.ABaseARM64:  unfoldOpWithCarry,
	ASBCS - obj.ABaseARM64:  unfoldOpWithCarry,
	ASBCSW - obj.ABaseARM64: unfoldOpWithCarry,

	ANGC - obj.ABaseARM64:   unfoldNGCX,
	ANGCW - obj.ABaseARM64:  unfoldNGCX,
	ANGCS - obj.ABaseARM64:  unfoldNGCX,
	ANGCSW - obj.ABaseARM64: unfoldNGCX,

	AADD - obj.ABaseARM64:   unfoldADD,
	AADDW - obj.ABaseARM64:  unfoldADDW,
	AADDS - obj.ABaseARM64:  unfoldADDS,
	AADDSW - obj.ABaseARM64: unfoldADDSW,

	ASUB - obj.ABaseARM64:   unfoldSUB,
	ASUBW - obj.ABaseARM64:  unfoldSUBW,
	ASUBS - obj.ABaseARM64:  unfoldSUBS,
	ASUBSW - obj.ABaseARM64: unfoldSUBSW,

	AADR - obj.ABaseARM64:  unfoldADRX,
	AADRP - obj.ABaseARM64: unfoldADRX,

	AAND - obj.ABaseARM64:  unfoldAND,
	AANDW - obj.ABaseARM64: unfoldANDW,
	AEOR - obj.ABaseARM64:  unfoldEOR,
	AEORW - obj.ABaseARM64: unfoldEORW,
	AORR - obj.ABaseARM64:  unfoldORR,
	AORRW - obj.ABaseARM64: unfoldORRW,
	ABIC - obj.ABaseARM64:  unfoldBIC,
	ABICW - obj.ABaseARM64: unfoldBICW,
	AEON - obj.ABaseARM64:  unfoldEON,
	AEONW - obj.ABaseARM64: unfoldEONW,
	AORN - obj.ABaseARM64:  unfoldORN,
	AORNW - obj.ABaseARM64: unfoldORNW,

	AANDS - obj.ABaseARM64:  unfoldANDS,
	AANDSW - obj.ABaseARM64: unfoldANDSW,
	ABICS - obj.ABaseARM64:  unfoldBICS,
	ABICSW - obj.ABaseARM64: unfoldBICSW,

	ATST - obj.ABaseARM64:  unfoldTST,
	ATSTW - obj.ABaseARM64: unfoldTST,

	ABFM - obj.ABaseARM64:    unfoldBitFieldOps,
	ABFMW - obj.ABaseARM64:   unfoldBitFieldOps,
	ABFI - obj.ABaseARM64:    unfoldBitFieldOps,
	ABFIW - obj.ABaseARM64:   unfoldBitFieldOps,
	ABFXIL - obj.ABaseARM64:  unfoldBitFieldOps,
	ABFXILW - obj.ABaseARM64: unfoldBitFieldOps,
	ASBFM - obj.ABaseARM64:   unfoldBitFieldOps,
	ASBFMW - obj.ABaseARM64:  unfoldBitFieldOps,
	ASBFIZ - obj.ABaseARM64:  unfoldBitFieldOps,
	ASBFIZW - obj.ABaseARM64: unfoldBitFieldOps,
	ASBFX - obj.ABaseARM64:   unfoldBitFieldOps,
	ASBFXW - obj.ABaseARM64:  unfoldBitFieldOps,
	AUBFM - obj.ABaseARM64:   unfoldBitFieldOps,
	AUBFMW - obj.ABaseARM64:  unfoldBitFieldOps,
	AUBFIZ - obj.ABaseARM64:  unfoldBitFieldOps,
	AUBFIZW - obj.ABaseARM64: unfoldBitFieldOps,
	AUBFX - obj.ABaseARM64:   unfoldBitFieldOps,
	AUBFXW - obj.ABaseARM64:  unfoldBitFieldOps,

	AMOVD - obj.ABaseARM64:  unfoldMOVD,
	AMOVW - obj.ABaseARM64:  unfoldMOVW,
	AMOVWU - obj.ABaseARM64: unfoldMOVWU,
	AMOVH - obj.ABaseARM64:  unfoldMOVH,
	AMOVHU - obj.ABaseARM64: unfoldMOVHU,
	AMOVB - obj.ABaseARM64:  unfoldMOVB,
	AMOVBU - obj.ABaseARM64: unfoldMOVBU,

	AFMOVQ - obj.ABaseARM64: unfoldFMOVQ,
	AFMOVD - obj.ABaseARM64: unfoldFMOVD,
	AFMOVS - obj.ABaseARM64: unfoldFMOVS,

	ALDP - obj.ABaseARM64:   unfoldLDP,
	ALDPW - obj.ABaseARM64:  unfoldLDPW,
	ALDPSW - obj.ABaseARM64: unfoldLDPSW,
	AFLDPQ - obj.ABaseARM64: unfoldFLDPQ,
	AFLDPD - obj.ABaseARM64: unfoldFLDPD,
	AFLDPS - obj.ABaseARM64: unfoldFLDPS,
	ASTP - obj.ABaseARM64:   unfoldSTP,
	ASTPW - obj.ABaseARM64:  unfoldSTPW,
	AFSTPQ - obj.ABaseARM64: unfoldFSTPQ,
	AFSTPD - obj.ABaseARM64: unfoldFSTPD,
	AFSTPS - obj.ABaseARM64: unfoldFSTPS,

	ACBZ - obj.ABaseARM64:   unfoldCBZX,
	ACBZW - obj.ABaseARM64:  unfoldCBZX,
	ACBNZ - obj.ABaseARM64:  unfoldCBZX,
	ACBNZW - obj.ABaseARM64: unfoldCBZX,

	ACCMN - obj.ABaseARM64:  unfoldCondCMP,
	ACCMNW - obj.ABaseARM64: unfoldCondCMP,
	ACCMP - obj.ABaseARM64:  unfoldCondCMP,
	ACCMPW - obj.ABaseARM64: unfoldCondCMP,

	ACSEL - obj.ABaseARM64:   unfoldCSELX,
	ACSELW - obj.ABaseARM64:  unfoldCSELX,
	ACSINC - obj.ABaseARM64:  unfoldCSELX,
	ACSINCW - obj.ABaseARM64: unfoldCSELX,
	ACSINV - obj.ABaseARM64:  unfoldCSELX,
	ACSINVW - obj.ABaseARM64: unfoldCSELX,
	ACSNEG - obj.ABaseARM64:  unfoldCSELX,
	ACSNEGW - obj.ABaseARM64: unfoldCSELX,
	AFCSELD - obj.ABaseARM64: unfoldCSELX,
	AFCSELS - obj.ABaseARM64: unfoldCSELX,

	ACSET - obj.ABaseARM64:   unfoldCSETX,
	ACSETW - obj.ABaseARM64:  unfoldCSETX,
	ACSETM - obj.ABaseARM64:  unfoldCSETX,
	ACSETMW - obj.ABaseARM64: unfoldCSETX,

	ACINC - obj.ABaseARM64:  unfoldCINCX,
	ACINCW - obj.ABaseARM64: unfoldCINCX,
	ACINV - obj.ABaseARM64:  unfoldCINCX,
	ACINVW - obj.ABaseARM64: unfoldCINCX,
	ACNEG - obj.ABaseARM64:  unfoldCINCX,
	ACNEGW - obj.ABaseARM64: unfoldCINCX,

	AFCCMPD - obj.ABaseARM64:  unfoldFloatingCondCMP,
	AFCCMPS - obj.ABaseARM64:  unfoldFloatingCondCMP,
	AFCCMPED - obj.ABaseARM64: unfoldFloatingCondCMP,
	AFCCMPES - obj.ABaseARM64: unfoldFloatingCondCMP,

	ACLREX - obj.ABaseARM64: unfoldCLREX,

	ACLS - obj.ABaseARM64:    unfoldCLSX,
	ACLSW - obj.ABaseARM64:   unfoldCLSX,
	ACLZ - obj.ABaseARM64:    unfoldCLSX,
	ACLZW - obj.ABaseARM64:   unfoldCLSX,
	ARBIT - obj.ABaseARM64:   unfoldCLSX,
	ARBITW - obj.ABaseARM64:  unfoldCLSX,
	AREV - obj.ABaseARM64:    unfoldCLSX,
	AREVW - obj.ABaseARM64:   unfoldCLSX,
	AREV16 - obj.ABaseARM64:  unfoldCLSX,
	AREV16W - obj.ABaseARM64: unfoldCLSX,
	AREV32 - obj.ABaseARM64:  unfoldCLSX,

	ACMP - obj.ABaseARM64:  unfoldCMP,
	ACMPW - obj.ABaseARM64: unfoldCMPW,
	ACMN - obj.ABaseARM64:  unfoldCMN,
	ACMNW - obj.ABaseARM64: unfoldCMNW,

	ADMB - obj.ABaseARM64:  unfoldDMBX,
	ADSB - obj.ABaseARM64:  unfoldDMBX,
	AISB - obj.ABaseARM64:  unfoldDMBX,
	AHINT - obj.ABaseARM64: unfoldDMBX,

	AERET - obj.ABaseARM64:  unfoldOpwithoutArg,
	AWFE - obj.ABaseARM64:   unfoldOpwithoutArg,
	AWFI - obj.ABaseARM64:   unfoldOpwithoutArg,
	AYIELD - obj.ABaseARM64: unfoldOpwithoutArg,
	ASEV - obj.ABaseARM64:   unfoldOpwithoutArg,
	ASEVL - obj.ABaseARM64:  unfoldOpwithoutArg,
	ANOOP - obj.ABaseARM64:  unfoldOpwithoutArg,
	ADRPS - obj.ABaseARM64:  unfoldOpwithoutArg,

	AEXTR - obj.ABaseARM64:  unfoldEXTRX,
	AEXTRW - obj.ABaseARM64: unfoldEXTRX,

	ALDAR - obj.ABaseARM64:   unfoldLoadAcquire,
	ALDARW - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDARH - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDARB - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDAXR - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDAXRW - obj.ABaseARM64: unfoldLoadAcquire,
	ALDAXRH - obj.ABaseARM64: unfoldLoadAcquire,
	ALDAXRB - obj.ABaseARM64: unfoldLoadAcquire,
	ALDXR - obj.ABaseARM64:   unfoldLoadAcquire,
	ALDXRW - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDXRH - obj.ABaseARM64:  unfoldLoadAcquire,
	ALDXRB - obj.ABaseARM64:  unfoldLoadAcquire,

	ALDXP - obj.ABaseARM64:   unfoldLDXPX,
	ALDXPW - obj.ABaseARM64:  unfoldLDXPX,
	ALDAXP - obj.ABaseARM64:  unfoldLDXPX,
	ALDAXPW - obj.ABaseARM64: unfoldLDXPX,

	ASTLR - obj.ABaseARM64:  unfoldSTLRX,
	ASTLRW - obj.ABaseARM64: unfoldSTLRX,
	ASTLRH - obj.ABaseARM64: unfoldSTLRX,
	ASTLRB - obj.ABaseARM64: unfoldSTLRX,

	ASTLXR - obj.ABaseARM64:  unfoldSTXRX,
	ASTLXRW - obj.ABaseARM64: unfoldSTXRX,
	ASTLXRH - obj.ABaseARM64: unfoldSTXRX,
	ASTLXRB - obj.ABaseARM64: unfoldSTXRX,
	ASTXR - obj.ABaseARM64:   unfoldSTXRX,
	ASTXRW - obj.ABaseARM64:  unfoldSTXRX,
	ASTXRH - obj.ABaseARM64:  unfoldSTXRX,
	ASTXRB - obj.ABaseARM64:  unfoldSTXRX,

	ASTXP - obj.ABaseARM64:   unfoldSTXPX,
	ASTXPW - obj.ABaseARM64:  unfoldSTXPX,
	ASTLXP - obj.ABaseARM64:  unfoldSTXPX,
	ASTLXPW - obj.ABaseARM64: unfoldSTXPX,

	ALSL - obj.ABaseARM64:  unfoldShiftOp,
	ALSLW - obj.ABaseARM64: unfoldShiftOp,
	ALSR - obj.ABaseARM64:  unfoldShiftOp,
	ALSRW - obj.ABaseARM64: unfoldShiftOp,
	AASR - obj.ABaseARM64:  unfoldShiftOp,
	AASRW - obj.ABaseARM64: unfoldShiftOp,
	AROR - obj.ABaseARM64:  unfoldShiftOp,
	ARORW - obj.ABaseARM64: unfoldShiftOp,

	AMADD - obj.ABaseARM64:    unfoldFuseOp,
	AMADDW - obj.ABaseARM64:   unfoldFuseOp,
	AMSUB - obj.ABaseARM64:    unfoldFuseOp,
	AMSUBW - obj.ABaseARM64:   unfoldFuseOp,
	ASMADDL - obj.ABaseARM64:  unfoldFuseOp,
	ASMSUBL - obj.ABaseARM64:  unfoldFuseOp,
	AUMADDL - obj.ABaseARM64:  unfoldFuseOp,
	AUMSUBL - obj.ABaseARM64:  unfoldFuseOp,
	AFMADDD - obj.ABaseARM64:  unfoldFuseOp,
	AFMADDS - obj.ABaseARM64:  unfoldFuseOp,
	AFMSUBD - obj.ABaseARM64:  unfoldFuseOp,
	AFMSUBS - obj.ABaseARM64:  unfoldFuseOp,
	AFNMADDD - obj.ABaseARM64: unfoldFuseOp,
	AFNMADDS - obj.ABaseARM64: unfoldFuseOp,
	AFNMSUBD - obj.ABaseARM64: unfoldFuseOp,
	AFNMSUBS - obj.ABaseARM64: unfoldFuseOp,

	AMUL - obj.ABaseARM64:    unfoldMultipleX,
	AMULW - obj.ABaseARM64:   unfoldMultipleX,
	AMNEG - obj.ABaseARM64:   unfoldMultipleX,
	AMNEGW - obj.ABaseARM64:  unfoldMultipleX,
	ASMNEGL - obj.ABaseARM64: unfoldMultipleX,
	AUMNEGL - obj.ABaseARM64: unfoldMultipleX,
	ASMULH - obj.ABaseARM64:  unfoldMultipleX,
	ASMULL - obj.ABaseARM64:  unfoldMultipleX,
	AUMULH - obj.ABaseARM64:  unfoldMultipleX,
	AUMULL - obj.ABaseARM64:  unfoldMultipleX,

	AMOVK - obj.ABaseARM64:  unfoldMOVX,
	AMOVKW - obj.ABaseARM64: unfoldMOVX,
	AMOVN - obj.ABaseARM64:  unfoldMOVX,
	AMOVNW - obj.ABaseARM64: unfoldMOVX,
	AMOVZ - obj.ABaseARM64:  unfoldMOVX,
	AMOVZW - obj.ABaseARM64: unfoldMOVX,

	AMRS - obj.ABaseARM64: unfoldMRS,
	AMSR - obj.ABaseARM64: unfoldMSR,

	AMVN - obj.ABaseARM64:  unfoldMVN,
	AMVNW - obj.ABaseARM64: unfoldMVN,

	ANEG - obj.ABaseARM64:   unfoldNEGX,
	ANEGW - obj.ABaseARM64:  unfoldNEGX,
	ANEGS - obj.ABaseARM64:  unfoldNEGX,
	ANEGSW - obj.ABaseARM64: unfoldNEGX,

	APRFM - obj.ABaseARM64: unfoldPRFM,

	AREM - obj.ABaseARM64:   unfoldREMX,
	AREMW - obj.ABaseARM64:  unfoldREMX,
	AUREM - obj.ABaseARM64:  unfoldREMX,
	AUREMW - obj.ABaseARM64: unfoldREMX,

	ASDIV - obj.ABaseARM64:    unfoldDIVCRC,
	ASDIVW - obj.ABaseARM64:   unfoldDIVCRC,
	AUDIV - obj.ABaseARM64:    unfoldDIVCRC,
	AUDIVW - obj.ABaseARM64:   unfoldDIVCRC,
	ACRC32B - obj.ABaseARM64:  unfoldDIVCRC,
	ACRC32H - obj.ABaseARM64:  unfoldDIVCRC,
	ACRC32W - obj.ABaseARM64:  unfoldDIVCRC,
	ACRC32X - obj.ABaseARM64:  unfoldDIVCRC,
	ACRC32CB - obj.ABaseARM64: unfoldDIVCRC,
	ACRC32CH - obj.ABaseARM64: unfoldDIVCRC,
	ACRC32CW - obj.ABaseARM64: unfoldDIVCRC,
	ACRC32CX - obj.ABaseARM64: unfoldDIVCRC,

	ASVC - obj.ABaseARM64:   unfoldPE,
	AHVC - obj.ABaseARM64:   unfoldPE,
	AHLT - obj.ABaseARM64:   unfoldPE,
	ASMC - obj.ABaseARM64:   unfoldPE,
	ABRK - obj.ABaseARM64:   unfoldPE,
	ADCPS1 - obj.ABaseARM64: unfoldPE,
	ADCPS2 - obj.ABaseARM64: unfoldPE,
	ADCPS3 - obj.ABaseARM64: unfoldPE,

	ASXTB - obj.ABaseARM64:  unfoldExtend,
	ASXTBW - obj.ABaseARM64: unfoldExtend,
	ASXTH - obj.ABaseARM64:  unfoldExtend,
	ASXTHW - obj.ABaseARM64: unfoldExtend,
	ASXTW - obj.ABaseARM64:  unfoldExtend,
	AUXTBW - obj.ABaseARM64: unfoldExtend,
	AUXTHW - obj.ABaseARM64: unfoldExtend,

	AUXTB - obj.ABaseARM64: unfoldUnsignedExtend,
	AUXTH - obj.ABaseARM64: unfoldUnsignedExtend,
	AUXTW - obj.ABaseARM64: unfoldUnsignedExtend,

	ASYS - obj.ABaseARM64:  unfoldSYS,
	ASYSL - obj.ABaseARM64: unfoldSYSL,
	AAT - obj.ABaseARM64:   unfoldATIC,
	AIC - obj.ABaseARM64:   unfoldATIC,
	ADC - obj.ABaseARM64:   unfoldSYSAlias,
	ATLBI - obj.ABaseARM64: unfoldSYSAlias,

	ATBNZ - obj.ABaseARM64: unfoldTBZX,
	ATBZ - obj.ABaseARM64:  unfoldTBZX,

	ACASD - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ACASW - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ACASH - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ACASB - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ACASAD - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ACASAW - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ACASALD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ACASALW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ACASALH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ACASALB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ACASLD - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ACASLW - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ALDADDD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDADDW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDADDH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDADDB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDADDAD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDAW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDAH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDAB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDALD - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDADDALW - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDADDALH - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDADDALB - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDADDLD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDLW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDLH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDADDLB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDCLRW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDCLRH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDCLRB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDCLRAD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRAW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRAH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRAB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRALD - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDCLRALW - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDCLRALH - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDCLRALB - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDCLRLD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRLW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRLH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDCLRLB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDEORW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDEORH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDEORB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDEORAD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORAW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORAH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORAB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORALD - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDEORALW - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDEORALH - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDEORALB - obj.ABaseARM64: unfoldAtomicLoadOpStore,
	ALDEORLD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORLW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORLH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDEORLB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDORD - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ALDORW - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ALDORH - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ALDORB - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ALDORAD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORAW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORAH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORAB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORALD - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDORALW - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDORALH - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDORALB - obj.ABaseARM64:  unfoldAtomicLoadOpStore,
	ALDORLD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORLW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORLH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ALDORLB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ASWPD - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ASWPW - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ASWPH - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ASWPB - obj.ABaseARM64:     unfoldAtomicLoadOpStore,
	ASWPAD - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPAW - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPAH - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPAB - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPALD - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ASWPALW - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ASWPALH - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ASWPALB - obj.ABaseARM64:   unfoldAtomicLoadOpStore,
	ASWPLD - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPLW - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPLH - obj.ABaseARM64:    unfoldAtomicLoadOpStore,
	ASWPLB - obj.ABaseARM64:    unfoldAtomicLoadOpStore,

	ACASPD - obj.ABaseARM64: unfoldCASPX,
	ACASPW - obj.ABaseARM64: unfoldCASPX,

	ABEQ - obj.ABaseARM64: unfoldBCOND,
	ABNE - obj.ABaseARM64: unfoldBCOND,
	ABCS - obj.ABaseARM64: unfoldBCOND,
	ABHS - obj.ABaseARM64: unfoldBCOND,
	ABCC - obj.ABaseARM64: unfoldBCOND,
	ABLO - obj.ABaseARM64: unfoldBCOND,
	ABMI - obj.ABaseARM64: unfoldBCOND,
	ABPL - obj.ABaseARM64: unfoldBCOND,
	ABVS - obj.ABaseARM64: unfoldBCOND,
	ABVC - obj.ABaseARM64: unfoldBCOND,
	ABHI - obj.ABaseARM64: unfoldBCOND,
	ABLS - obj.ABaseARM64: unfoldBCOND,
	ABGE - obj.ABaseARM64: unfoldBCOND,
	ABLT - obj.ABaseARM64: unfoldBCOND,
	ABGT - obj.ABaseARM64: unfoldBCOND,
	ABLE - obj.ABaseARM64: unfoldBCOND,

	AFADDD - obj.ABaseARM64:   unfoldFloatingOp,
	AFADDS - obj.ABaseARM64:   unfoldFloatingOp,
	AFSUBD - obj.ABaseARM64:   unfoldFloatingOp,
	AFSUBS - obj.ABaseARM64:   unfoldFloatingOp,
	AFMULD - obj.ABaseARM64:   unfoldFloatingOp,
	AFMULS - obj.ABaseARM64:   unfoldFloatingOp,
	AFNMULD - obj.ABaseARM64:  unfoldFloatingOp,
	AFNMULS - obj.ABaseARM64:  unfoldFloatingOp,
	AFDIVD - obj.ABaseARM64:   unfoldFloatingOp,
	AFDIVS - obj.ABaseARM64:   unfoldFloatingOp,
	AFMAXD - obj.ABaseARM64:   unfoldFloatingOp,
	AFMAXS - obj.ABaseARM64:   unfoldFloatingOp,
	AFMIND - obj.ABaseARM64:   unfoldFloatingOp,
	AFMINS - obj.ABaseARM64:   unfoldFloatingOp,
	AFMAXNMD - obj.ABaseARM64: unfoldFloatingOp,
	AFMAXNMS - obj.ABaseARM64: unfoldFloatingOp,
	AFMINNMD - obj.ABaseARM64: unfoldFloatingOp,
	AFMINNMS - obj.ABaseARM64: unfoldFloatingOp,

	AFCMPD - obj.ABaseARM64:  unfoldFCMPX,
	AFCMPS - obj.ABaseARM64:  unfoldFCMPX,
	AFCMPED - obj.ABaseARM64: unfoldFCMPX,
	AFCMPES - obj.ABaseARM64: unfoldFCMPX,

	AFCVTDH - obj.ABaseARM64:  unfoldFConvertRounding,
	AFCVTDS - obj.ABaseARM64:  unfoldFConvertRounding,
	AFCVTHD - obj.ABaseARM64:  unfoldFConvertRounding,
	AFCVTHS - obj.ABaseARM64:  unfoldFConvertRounding,
	AFCVTSD - obj.ABaseARM64:  unfoldFConvertRounding,
	AFCVTSH - obj.ABaseARM64:  unfoldFConvertRounding,
	AFABSD - obj.ABaseARM64:   unfoldFConvertRounding,
	AFABSS - obj.ABaseARM64:   unfoldFConvertRounding,
	AFNEGD - obj.ABaseARM64:   unfoldFConvertRounding,
	AFNEGS - obj.ABaseARM64:   unfoldFConvertRounding,
	AFSQRTD - obj.ABaseARM64:  unfoldFConvertRounding,
	AFSQRTS - obj.ABaseARM64:  unfoldFConvertRounding,
	AFRINTAD - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTAS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTID - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTIS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTMD - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTMS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTND - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTNS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTPD - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTPS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTXD - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTXS - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTZD - obj.ABaseARM64: unfoldFConvertRounding,
	AFRINTZS - obj.ABaseARM64: unfoldFConvertRounding,

	AFCVTZSD - obj.ABaseARM64:  unfoldFConvertToFixed,
	AFCVTZSDW - obj.ABaseARM64: unfoldFConvertToFixed,
	AFCVTZSS - obj.ABaseARM64:  unfoldFConvertToFixed,
	AFCVTZSSW - obj.ABaseARM64: unfoldFConvertToFixed,
	AFCVTZUD - obj.ABaseARM64:  unfoldFConvertToFixed,
	AFCVTZUDW - obj.ABaseARM64: unfoldFConvertToFixed,
	AFCVTZUS - obj.ABaseARM64:  unfoldFConvertToFixed,
	AFCVTZUSW - obj.ABaseARM64: unfoldFConvertToFixed,

	ASCVTFD - obj.ABaseARM64:  unfoldFixedToFloating,
	ASCVTFS - obj.ABaseARM64:  unfoldFixedToFloating,
	ASCVTFWD - obj.ABaseARM64: unfoldFixedToFloating,
	ASCVTFWS - obj.ABaseARM64: unfoldFixedToFloating,
	AUCVTFD - obj.ABaseARM64:  unfoldFixedToFloating,
	AUCVTFS - obj.ABaseARM64:  unfoldFixedToFloating,
	AUCVTFWD - obj.ABaseARM64: unfoldFixedToFloating,
	AUCVTFWS - obj.ABaseARM64: unfoldFixedToFloating,

	AAESD - obj.ABaseARM64:      unfoldAESSHA,
	AAESE - obj.ABaseARM64:      unfoldAESSHA,
	AAESIMC - obj.ABaseARM64:    unfoldAESSHA,
	AAESMC - obj.ABaseARM64:     unfoldAESSHA,
	ASHA1SU1 - obj.ABaseARM64:   unfoldAESSHA,
	ASHA256SU0 - obj.ABaseARM64: unfoldAESSHA,
	ASHA512SU0 - obj.ABaseARM64: unfoldAESSHA,

	AVREV16 - obj.ABaseARM64: unfoldFBitwiseOp,
	AVREV32 - obj.ABaseARM64: unfoldFBitwiseOp,
	AVREV64 - obj.ABaseARM64: unfoldFBitwiseOp,
	AVCNT - obj.ABaseARM64:   unfoldFBitwiseOp,
	AVRBIT - obj.ABaseARM64:  unfoldFBitwiseOp,

	ASHA1H - obj.ABaseARM64: unfoldSHA1H,

	ASHA1C - obj.ABaseARM64:    unfoldSHAX,
	ASHA1P - obj.ABaseARM64:    unfoldSHAX,
	ASHA1M - obj.ABaseARM64:    unfoldSHAX,
	ASHA256H - obj.ABaseARM64:  unfoldSHAX,
	ASHA256H2 - obj.ABaseARM64: unfoldSHAX,
	ASHA512H - obj.ABaseARM64:  unfoldSHAX,
	ASHA512H2 - obj.ABaseARM64: unfoldSHAX,

	ASHA1SU0 - obj.ABaseARM64:   unfoldARNG3,
	ASHA256SU1 - obj.ABaseARM64: unfoldARNG3,
	ASHA512SU1 - obj.ABaseARM64: unfoldARNG3,
	AVRAX1 - obj.ABaseARM64:     unfoldARNG3,
	AVADDP - obj.ABaseARM64:     unfoldARNG3,
	AVAND - obj.ABaseARM64:      unfoldARNG3,
	AVORR - obj.ABaseARM64:      unfoldARNG3,
	AVEOR - obj.ABaseARM64:      unfoldARNG3,
	AVBIF - obj.ABaseARM64:      unfoldARNG3,
	AVBIT - obj.ABaseARM64:      unfoldARNG3,
	AVBSL - obj.ABaseARM64:      unfoldARNG3,
	AVUMAX - obj.ABaseARM64:     unfoldARNG3,
	AVUMIN - obj.ABaseARM64:     unfoldARNG3,
	AVUZP1 - obj.ABaseARM64:     unfoldARNG3,
	AVUZP2 - obj.ABaseARM64:     unfoldARNG3,
	AVFMLA - obj.ABaseARM64:     unfoldARNG3,
	AVFMLS - obj.ABaseARM64:     unfoldARNG3,
	AVZIP1 - obj.ABaseARM64:     unfoldARNG3,
	AVZIP2 - obj.ABaseARM64:     unfoldARNG3,
	AVTRN1 - obj.ABaseARM64:     unfoldARNG3,
	AVTRN2 - obj.ABaseARM64:     unfoldARNG3,

	AVPMULL - obj.ABaseARM64:  unfoldVPMULLX,
	AVPMULL2 - obj.ABaseARM64: unfoldVPMULLX,

	AVUADDW - obj.ABaseARM64:  unfoldVUADDWX,
	AVUADDW2 - obj.ABaseARM64: unfoldVUADDWX,

	AVADD - obj.ABaseARM64:   unfoldVReg3OrARNG3,
	AVSUB - obj.ABaseARM64:   unfoldVReg3OrARNG3,
	AVCMEQ - obj.ABaseARM64:  unfoldVReg3OrARNG3,
	AVCMTST - obj.ABaseARM64: unfoldVReg3OrARNG3,

	AVEOR3 - obj.ABaseARM64: unfoldARNG4,
	AVBCAX - obj.ABaseARM64: unfoldARNG4,

	AVEXT - obj.ABaseARM64: unfoldVEXT,

	AVXAR - obj.ABaseARM64: unfoldVXAR,

	AVMOV - obj.ABaseARM64: unfoldVMOV,

	AVMOVQ - obj.ABaseARM64: unfoldVMOVQ,
	AVMOVD - obj.ABaseARM64: unfoldVMOVD,
	AVMOVS - obj.ABaseARM64: unfoldVMOVS,

	AVLD1 - obj.ABaseARM64: unfoldVLD1,
	AVST1 - obj.ABaseARM64: unfoldVST1,
	AVLD2 - obj.ABaseARM64: unfoldVLD2,
	AVLD3 - obj.ABaseARM64: unfoldVLD3,
	AVLD4 - obj.ABaseARM64: unfoldVLD4,
	AVST2 - obj.ABaseARM64: unfoldVST2,
	AVST3 - obj.ABaseARM64: unfoldVST3,
	AVST4 - obj.ABaseARM64: unfoldVST4,

	AVLD1R - obj.ABaseARM64: unfoldVLD1R,
	AVLD2R - obj.ABaseARM64: unfoldVLD2R,
	AVLD3R - obj.ABaseARM64: unfoldVLD3R,
	AVLD4R - obj.ABaseARM64: unfoldVLD4R,

	AVDUP - obj.ABaseARM64: unfoldVDUP,

	AVUXTL - obj.ABaseARM64:  unfoldVUXTLX,
	AVUXTL2 - obj.ABaseARM64: unfoldVUXTLX,

	AVUSHLL - obj.ABaseARM64:  unfoldVUSHLLX,
	AVUSHLL2 - obj.ABaseARM64: unfoldVUSHLLX,

	AVSHL - obj.ABaseARM64:  unfoldVectorShift,
	AVSLI - obj.ABaseARM64:  unfoldVectorShift,
	AVSRI - obj.ABaseARM64:  unfoldVectorShift,
	AVUSRA - obj.ABaseARM64: unfoldVectorShift,
	AVUSHR - obj.ABaseARM64: unfoldVectorShift,

	AVTBL - obj.ABaseARM64: unfoldVTBL,

	AVADDV - obj.ABaseARM64:   unfoldVADDVX,
	AVUADDLV - obj.ABaseARM64: unfoldVADDVX,

	AVMOVI - obj.ABaseARM64: unfoldVMOVI,
}
