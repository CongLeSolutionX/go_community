// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mips32

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
	"fmt"
)

func defframe(ptxt *obj.Prog) {
	// fill in argument size, stack size
	ptxt.To.Type = obj.TYPE_TEXTSIZE

	ptxt.To.Val = int32(gc.Rnd(gc.Curfn.Type.ArgWidth(), int64(gc.Widthptr)))
	frame := uint32(gc.Rnd(gc.Stksize+gc.Maxarg, int64(gc.Widthreg)))
	ptxt.To.Offset = int64(frame)

	// insert code to zero ambiguously live variables
	// so that the garbage collector only sees initialized values
	// when it looks for pointers.
	p := ptxt

	hi := int64(0)
	lo := hi

	// iterate through declarations - they are sorted in decreasing xoffset order.
	for _, n := range gc.Curfn.Func.Dcl {
		if !n.Name.Needzero {
			continue
		}
		if n.Class != gc.PAUTO {
			gc.Fatalf("needzero class %d", n.Class)
		}
		if n.Type.Width%int64(gc.Widthptr) != 0 || n.Xoffset%int64(gc.Widthptr) != 0 || n.Type.Width == 0 {
			gc.Fatalf("var %v has size %d offset %d", gc.Nconv(n, gc.FmtLong), int(n.Type.Width), int(n.Xoffset))
		}

		if lo != hi && n.Xoffset+n.Type.Width >= lo-int64(2*gc.Widthreg) {
			// merge with range we already have
			lo = n.Xoffset

			continue
		}

		// zero old range
		p = zerorange(p, int64(frame), lo, hi)

		// set new range
		hi = n.Xoffset + n.Type.Width

		lo = n.Xoffset
	}

	// zero final range
	zerorange(p, int64(frame), lo, hi)
}

// TODO(mips32): implement DUFFZERO
func zerorange(p *obj.Prog, frame int64, lo int64, hi int64) *obj.Prog {

	cnt := hi - lo
	if cnt == 0 {
		return p
	}
	if cnt < int64(4*gc.Widthptr) {
		for i := int64(0); i < cnt; i += int64(gc.Widthptr) {
			p = appendpp(p, mips32.AMOVW, obj.TYPE_REG, mips32.REGZERO, 0, obj.TYPE_MEM, mips32.REGSP, gc.Ctxt.FixedFrameSize()+frame+lo+i)
		}
	} else {
		//fmt.Printf("zerorange frame:%v, lo: %v, hi:%v \n", frame ,lo, hi)
		//	ADD 	$(FIXED_FRAME+frame+lo-4), SP, r1
		//	ADD 	$cnt, r1, r2
		// loop:
		//	MOVW	R0, (Widthptr)r1
		//	ADD 	$Widthptr, r1
		//	BNE		r1, r2, loop
		p = appendpp(p, mips32.AADD, obj.TYPE_CONST, 0, gc.Ctxt.FixedFrameSize()+frame+lo-4, obj.TYPE_REG, mips32.REGRT1, 0)
		p.Reg = mips32.REGSP
		p = appendpp(p, mips32.AADD, obj.TYPE_CONST, 0, cnt, obj.TYPE_REG, mips32.REGRT2, 0)
		p.Reg = mips32.REGRT1
		p = appendpp(p, mips32.AMOVW, obj.TYPE_REG, mips32.REGZERO, 0, obj.TYPE_MEM, mips32.REGRT1, int64(gc.Widthptr))
		p1 := p
		p = appendpp(p, mips32.AADD, obj.TYPE_CONST, 0, int64(gc.Widthptr), obj.TYPE_REG, mips32.REGRT1, 0)
		p = appendpp(p, mips32.ABNE, obj.TYPE_REG, mips32.REGRT1, 0, obj.TYPE_BRANCH, 0, 0)
		p.Reg = mips32.REGRT2
		gc.Patch(p, p1)
	}

	return p
}

func appendpp(p *obj.Prog, as obj.As, ftype obj.AddrType, freg int, foffset int64, ttype obj.AddrType, treg int, toffset int64) *obj.Prog {
	q := gc.Ctxt.NewProg()
	gc.Clearp(q)
	q.As = as
	q.Lineno = p.Lineno
	q.From.Type = ftype
	q.From.Reg = int16(freg)
	q.From.Offset = foffset
	q.To.Type = ttype
	q.To.Reg = int16(treg)
	q.To.Offset = toffset
	q.Link = p.Link
	p.Link = q
	return q
}

func ginsnop() {
	var reg gc.Node
	gc.Nodreg(&reg, gc.Types[gc.TINT], mips32.REG_R0)
	gins(mips32.ANOR, &reg, &reg)
}

var panicdiv *gc.Node

/*
 * generate division.
 * generates one of:
 *	res = nl / nr
 *	res = nl % nr
 * according to op.
 */
func dodiv(op gc.Op, nl *gc.Node, nr *gc.Node, res *gc.Node) {
	t := nl.Type

	t0 := t

	if t.Width < 4 {
		if t.IsSigned() {
			t = gc.Types[gc.TINT32]
		} else {
			t = gc.Types[gc.TUINT32]
		}
	}

	a := optoas(gc.ODIV, t)

	var tl gc.Node
	gc.Regalloc(&tl, t0, nil)
	var tr gc.Node
	gc.Regalloc(&tr, t0, nil)
	if nl.Ullman >= nr.Ullman {
		gc.Cgen(nl, &tl)
		gc.Cgen(nr, &tr)
	} else {
		gc.Cgen(nr, &tr)
		gc.Cgen(nl, &tl)
	}

	if t != t0 {
		// Convert
		tl2 := tl

		tr2 := tr
		tl.Type = t
		tr.Type = t
		gmove(&tl2, &tl)
		gmove(&tr2, &tr)
	}

	gins3(a, &tr, &tl, nil)

	if !(gc.Isconst(nr, gc.CTINT) && nr.Int64() != 0) {

		var rz gc.Node
		gc.Nodreg(&rz, gc.Types[gc.TUINT32], mips32.REGZERO)
		gins3(mips32.ATEQ, ncon(0x7), &tr, &rz) // BRK_DIVZERO

	}

	gc.Regfree(&tr)
	if op == gc.ODIV {
		var lo gc.Node
		gc.Nodreg(&lo, gc.Types[gc.TUINT32], mips32.REG_LO)
		gins(mips32.AMOVW, &lo, &tl)
	} else { // remainder in REG_HI
		var hi gc.Node
		gc.Nodreg(&hi, gc.Types[gc.TUINT32], mips32.REG_HI)
		gins(mips32.AMOVW, &hi, &tl)
	}
	gmove(&tl, res)
	gc.Regfree(&tl)
}

/*
 * generate high multiply:
 *   res = (nl*nr) >> width
 */
func cgen_hmul(nl *gc.Node, nr *gc.Node, res *gc.Node) {
	// largest ullman on left.
	if nl.Ullman < nr.Ullman {
		nl, nr = nr, nl
	}

	t := nl.Type
	w := t.Width * 8
	var n1 gc.Node
	gc.Cgenr(nl, &n1, res)
	var n2 gc.Node
	gc.Cgenr(nr, &n2, nil)
	switch gc.Simtype[t.Etype] {
	case gc.TINT8,
		gc.TINT16,
		gc.TINT32:
		gins3(optoas(gc.OMUL, t), &n2, &n1, nil)
		var lo gc.Node
		gc.Nodreg(&lo, gc.Types[gc.TUINT32], mips32.REG_LO)
		gins(mips32.AMOVW, &lo, &n1)
		p := (*obj.Prog)(gins(mips32.ASRA, nil, &n1))
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(w)

	case gc.TUINT8,
		gc.TUINT16,
		gc.TUINT32:
		gins3(optoas(gc.OMUL, t), &n2, &n1, nil)
		var lo gc.Node
		gc.Nodreg(&lo, gc.Types[gc.TUINT32], mips32.REG_LO)
		gins(mips32.AMOVW, &lo, &n1)
		p := (*obj.Prog)(gins(mips32.ASRL, nil, &n1))
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(w)

	default:
		gc.Fatalf("cgen_hmul %v", t)
	}

	gc.Cgen(&n1, res)
	gc.Regfree(&n1)
	gc.Regfree(&n2)
}

/*
 * generate shift according to op, one of:
 *	res = nl << nr
 *	res = nl >> nr
 */
func cgen_shift(op gc.Op, bounded bool, nl *gc.Node, nr *gc.Node, res *gc.Node) {
	if nl.Type.Width > 4 {
		gc.Fatalf("cgen_shift %v", nl.Type)
	}

	switch op {
	default:
		gc.Fatalf("cgen_shift %v", op, gc.FmtSharp)
	case gc.OLSH, gc.ORSH:
		break
	}

	w := uint32(nl.Type.Width * 8) // could be replaced by word size
	a := optoas(op, nl.Type)
	flushToZero := !(op == gc.ORSH && nl.Type.IsSigned())

	if nr.Op == gc.OLITERAL {
		var n1 gc.Node
		gc.Regalloc(&n1, nl.Type, res)
		gc.Cgen(nl, &n1)
		sc := uint64(nr.Int64())

		if sc >= uint64(w) {
			if flushToZero {
				gins(mips32.AMOVW, ncon(0), &n1)
			} else {
				gins(a, ncon(w-1), &n1)
			}
		} else if sc != 0 {
			gins(a, nr, &n1)
		}

		gmove(&n1, res)
		gc.Regfree(&n1)
		return
	}

	var n1 gc.Node
	var n2 gc.Node
	var n3 gc.Node
	var n4 gc.Node

	var nz gc.Node
	gc.Nodreg(&nz, gc.Types[gc.TUINT], mips32.REGZERO)
	var ntmp gc.Node
	gc.Nodreg(&ntmp, gc.Types[gc.TUINT], mips32.REGTMP)

	if gc.Is64(nr.Type) {
		var nt gc.Node
		gc.Tempname(&nt, nr.Type)

		if nl.Ullman >= nr.Ullman {
			gc.Regalloc(&n1, nl.Type, res)
			gc.Cgen(nl, &n1)
			gc.Cgen(nr, &nt)
		} else {
			gc.Cgen(nr, &nt)
			gc.Regalloc(&n1, nl.Type, res)
			gc.Cgen(nl, &n1)
		}

		var hi gc.Node
		var lo gc.Node
		split64(&nt, &lo, &hi)

		gc.Regalloc(&n2, gc.Types[gc.TUINT32], nil)
		gmove(&lo, &n2)

		if !bounded {
			gc.Regalloc(&n3, gc.Types[gc.TUINT32], nil)
			gmove(&hi, &n3)
		}

		splitclean()
	} else {
		if nl.Ullman >= nr.Ullman {
			gc.Regalloc(&n1, nl.Type, res)
			gc.Cgen(nl, &n1)
			gc.Regalloc(&n2, nr.Type, nil)
			gc.Cgen(nr, &n2)
		} else {
			gc.Regalloc(&n2, nr.Type, nil)
			gc.Cgen(nr, &n2)
			gc.Regalloc(&n1, nl.Type, res)
			gc.Cgen(nl, &n1)
		}
	}

	if bounded {
		gins(a, &n2, &n1)
		goto finish
	}

	// $op     $res,$nl,$nr
	// sltu    $tmp,$nr,32
	// movz    $res,$0,$tmp

	if flushToZero {
		gins(a, &n2, &n1)
		gins3(mips32.ASGTU, ncon(w), &n2, &ntmp)
		gins3(mips32.AMOVZ, &ntmp, &nz, &n1)
		if n3.Op != 0 {
			gins3(mips32.AMOVN, &n3, &nz, &n1)
		}

		goto finish
	}

	// li      $n4,31
	// sltu    $tmp,$nr,32
	// movz    $nr,$n4,$tmp
	// $op     $res,$nl,$nr

	gc.Regalloc(&n4, gc.Types[gc.TUINT32], nil)

	gmove(ncon(w-1), &n4)

	gins3(mips32.ASGTU, ncon(w), &n2, &ntmp)
	gins3(mips32.AMOVZ, &ntmp, &n4, &n2)

	if n3.Op != 0 {
		gins3(mips32.AMOVN, &n3, &n4, &n2)
	}

	gc.Regfree(&n4)

	gins(a, &n2, &n1)

finish:

	gmove(&n1, res)

	gc.Regfree(&n1)
	gc.Regfree(&n2)

	if n3.Op != 0 {
		gc.Regfree(&n3)
	}
}

func clearfat(nl *gc.Node) {
	/* clear a fat object */
	if gc.Debug['g'] != 0 {
		fmt.Printf("clearfat %v (%v, size: %d)\n", nl, nl.Type, nl.Type.Width)
	}

	w := uint64(uint64(nl.Type.Width))

	// Avoid taking the address for simple enough types.
	if gc.Componentgen(nil, nl) {
		return
	}

	c := uint64(w % 4) // bytes
	q := uint64(w / 4) // dwords

	if gc.Reginuse(mips32.REGRT1) {
		gc.Fatalf("%v in use during clearfat", obj.Rconv(mips32.REGRT1))
	}

	var r0 gc.Node
	gc.Nodreg(&r0, gc.Types[gc.TUINT32], mips32.REGZERO)
	var dst gc.Node
	gc.Nodreg(&dst, gc.Types[gc.Tptr], mips32.REGRT1)
	gc.Regrealloc(&dst)
	gc.Agen(nl, &dst)

	var boff uint64
	if q > 128 {
		p := gins(mips32.ASUB, nil, &dst)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = 4

		var end gc.Node
		gc.Regalloc(&end, gc.Types[gc.Tptr], nil)
		p = gins(mips32.AMOVW, &dst, &end)
		p.From.Type = obj.TYPE_ADDR
		p.From.Offset = int64(q * 4)

		p = gins(mips32.AMOVW, &r0, &dst)
		p.To.Type = obj.TYPE_MEM
		p.To.Offset = 4
		pl := (*obj.Prog)(p)

		p = gins(mips32.AADD, nil, &dst)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = 4

		gc.Patch(ginsbranch(mips32.ABNE, nil, &dst, &end, 0), pl)

		gc.Regfree(&end)

		// The loop leaves R1 on the last zeroed dword
		boff = 4
		// TODO(mips32): DUFFZERO
	} else {
		var p *obj.Prog
		for t := uint64(0); t < q; t++ {
			p = gins(mips32.AMOVW, &r0, &dst)
			p.To.Type = obj.TYPE_MEM
			p.To.Offset = int64(4 * t)
		}

		boff = 4 * q
	}

	var p *obj.Prog
	for t := uint64(0); t < c; t++ {
		p = gins(mips32.AMOVB, &r0, &dst)
		p.To.Type = obj.TYPE_MEM
		p.To.Offset = int64(t + boff)
	}

	gc.Regfree(&dst)
}

// Called after regopt and peep have run.
// Expand CHECKNIL pseudo-op into actual nil pointer check.
func expandchecks(firstp *obj.Prog) {
	var p1 *obj.Prog

	for p := (*obj.Prog)(firstp); p != nil; p = p.Link {
		if gc.Debug_checknil != 0 && gc.Ctxt.Debugvlog != 0 {
			fmt.Printf("expandchecks: %v\n", p)
		}
		if p.As != obj.ACHECKNIL {
			continue
		}
		if gc.Debug_checknil != 0 && p.Lineno > 1 { // p->lineno==1 in generated wrappers
			gc.Warnl(p.Lineno, "generated nil check")
		}

		//TODO maybe use TEQ or MOWB instruction
		// check is
		//	BNE arg, 2(PC)
		//	MOVW R0, 0(R0)
		p1 = gc.Ctxt.NewProg()
		gc.Clearp(p1)
		p1.Link = p.Link
		p.Link = p1
		p1.Lineno = p.Lineno
		p1.Pc = 9999
		//
		p.As = mips32.ABNE
		p.To.Type = obj.TYPE_BRANCH
		p.To.Val = p1.Link

		// crash by write to memory address 0.
		p1.As = mips32.AMOVW
		p1.From.Type = obj.TYPE_REG
		p1.From.Reg = mips32.REGZERO
		p1.To.Type = obj.TYPE_MEM
		p1.To.Reg = mips32.REGZERO
		p1.To.Offset = 0
	}
}

// res = runtime.getg()
func getg(res *gc.Node) {
	var n1 gc.Node
	gc.Nodreg(&n1, res.Type, mips32.REGG)
	gmove(&n1, res)
}
