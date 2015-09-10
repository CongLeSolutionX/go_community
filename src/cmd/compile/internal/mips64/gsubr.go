// Derived from Inferno utils/6c/txt.c
// http://code.google.com/p/inferno-os/source/browse/utils/6c/txt.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package mips64

import (
	"cmd/compile/internal/big"
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips"
	"fmt"
)

var resvd = []int{
	mips.REGZERO,
	mips.REGSP,   // reserved for SP
	mips.REGLINK, // reserved for link
	// We need to preserve the C ABI TLS pointer because sigtramp
	// may happen during C code and needs to access the g.  C
	// clobbers REGG, so if Go were to clobber REGTLS, sigtramp
	// won't know which convention to use.  By preserving REGTLS,
	// we can just retrieve g from TLS when we aren't sure.
	mips.REGTLS,

	mips.REGG,
	mips.REGTMP,
	mips.REG_R26, // kernel
	mips.REG_R27, // kernel
	mips.FREGZERO,
	mips.FREGHALF,
	mips.FREGONE,
	mips.FREGTWO,
}

/*
 * generate
 *	as $c, n
 */
func ginscon(as int, c int64, n2 *gc.Node) {
	var n1 gc.Node

	gc.Nodconst(&n1, gc.Types[gc.TINT64], c)

	if as != mips.AMOVV && (c < -mips.BIG || c > mips.BIG) || n2.Op != gc.OREGISTER || as == mips.AMUL || as == mips.AMULU || as == mips.AMULV || as == mips.AMULVU {
		// cannot have more than 16-bit of immediate in ADD, etc.
		// instead, MOV into register first.
		var ntmp gc.Node
		gc.Regalloc(&ntmp, gc.Types[gc.TINT64], nil)

		rawgins(mips.AMOVV, &n1, &ntmp)
		rawgins(as, &ntmp, n2)
		gc.Regfree(&ntmp)
		return
	}

	rawgins(as, &n1, n2)
}

// generate branch
// n1, n2 are registers
func ginsbranch(as int, t *gc.Type, n1, n2 *gc.Node, likely int) *obj.Prog {
	p := gc.Gbranch(as, t, likely)
	gc.Naddr(&p.From, n1)
	if n2 != nil {
		p.Reg = n2.Reg
	}
	return p
}

func ginscmp(op int, t *gc.Type, n1, n2 *gc.Node, likely int) *obj.Prog {
	if !gc.Isfloat[t.Etype] && (op == gc.OLT || op == gc.OGE) {
		// swap nodes to fit SGT instruction
		n1, n2 = n2, n1
	}
	if gc.Isfloat[t.Etype] && (op == gc.OLT || op == gc.OLE) {
		// swap nodes to fit CMPGT, CMPGE instructions and reverse relation
		n1, n2 = n2, n1
		if op == gc.OLT {
			op = gc.OGT
		} else {
			op = gc.OGE
		}
	}

	var r1, r2, g1, g2 gc.Node
	gc.Regalloc(&r1, t, n1)
	gc.Regalloc(&g1, n1.Type, &r1)
	gc.Cgen(n1, &g1)
	gmove(&g1, &r1)

	gc.Regalloc(&r2, t, n2)
	gc.Regalloc(&g2, n1.Type, &r2)
	gc.Cgen(n2, &g2)
	gmove(&g2, &r2)

	var p *obj.Prog
	var ntmp gc.Node
	gc.Nodreg(&ntmp, gc.Types[gc.TINT], mips.REGTMP)

	switch gc.Simtype[t.Etype] {
	case gc.TINT8,
		gc.TINT16,
		gc.TINT32,
		gc.TINT64:
		if op == gc.OEQ || op == gc.ONE {
			p = ginsbranch(optoas(op, t), nil, &r1, &r2, likely)
		} else {
			gins3(mips.ASGT, &r1, &r2, &ntmp)

			p = ginsbranch(optoas(op, t), nil, &ntmp, nil, likely)
		}

	case gc.TBOOL,
		gc.TUINT8,
		gc.TUINT16,
		gc.TUINT32,
		gc.TUINT64,
		gc.TPTR32,
		gc.TPTR64:
		if op == gc.OEQ || op == gc.ONE {
			p = ginsbranch(optoas(op, t), nil, &r1, &r2, likely)
		} else {
			gins3(mips.ASGTU, &r1, &r2, &ntmp)

			p = ginsbranch(optoas(op, t), nil, &ntmp, nil, likely)
		}

	case gc.TFLOAT32:
		switch op {
		default:
			gc.Fatalf("ginscmp: no entry for op=%v type=%v", gc.Oconv(int(op), 0), t)

		case gc.OEQ,
			gc.ONE:
			gins3(mips.ACMPEQF, &r1, &r2, nil)

		case gc.OGE:
			gins3(mips.ACMPGEF, &r1, &r2, nil)

		case gc.OGT:
			gins3(mips.ACMPGTF, &r1, &r2, nil)
		}
		p = gc.Gbranch(optoas(op, t), nil, likely)

	case gc.TFLOAT64:
		switch op {
		default:
			gc.Fatalf("ginscmp: no entry for op=%v type=%v", gc.Oconv(int(op), 0), t)

		case gc.OEQ,
			gc.ONE:
			gins3(mips.ACMPEQD, &r1, &r2, nil)

		case gc.OGE:
			gins3(mips.ACMPGED, &r1, &r2, nil)

		case gc.OGT:
			gins3(mips.ACMPGTD, &r1, &r2, nil)
		}
		p = gc.Gbranch(optoas(op, t), nil, likely)
	}

	gc.Regfree(&g2)
	gc.Regfree(&r2)
	gc.Regfree(&g1)
	gc.Regfree(&r1)

	return p
}

// set up nodes representing 2^63
var (
	bigi         gc.Node
	bigf         gc.Node
	bignodes_did bool
)

func bignodes() {
	if bignodes_did {
		return
	}
	bignodes_did = true

	var i big.Int
	i.SetInt64(1)
	i.Lsh(&i, 63)

	gc.Nodconst(&bigi, gc.Types[gc.TUINT64], 0)
	bigi.SetBigInt(&i)

	bigi.Convconst(&bigf, gc.Types[gc.TFLOAT64])
}

/*
 * generate move:
 *	t = f
 * hard part is conversions.
 */
func gmove(f *gc.Node, t *gc.Node) {
	if gc.Debug['M'] != 0 {
		fmt.Printf("gmove %v -> %v\n", gc.Nconv(f, obj.FmtLong), gc.Nconv(t, obj.FmtLong))
	}

	ft := int(gc.Simsimtype(f.Type))
	tt := int(gc.Simsimtype(t.Type))
	cvt := (*gc.Type)(t.Type)

	if gc.Iscomplex[ft] || gc.Iscomplex[tt] {
		gc.Complexmove(f, t)
		return
	}

	// cannot have two memory operands
	var r2 gc.Node
	var r1 gc.Node
	var a int
	if gc.Ismem(f) && gc.Ismem(t) {
		goto hard
	}

	// convert constant to desired type
	if f.Op == gc.OLITERAL {
		var con gc.Node
		switch tt {
		default:
			f.Convconst(&con, t.Type)

		case gc.TINT32,
			gc.TINT16,
			gc.TINT8:
			var con gc.Node
			f.Convconst(&con, gc.Types[gc.TINT64])
			var r1 gc.Node
			gc.Regalloc(&r1, con.Type, t)
			gins(mips.AMOVV, &con, &r1)
			gmove(&r1, t)
			gc.Regfree(&r1)
			return

		case gc.TUINT32,
			gc.TUINT16,
			gc.TUINT8:
			var con gc.Node
			f.Convconst(&con, gc.Types[gc.TUINT64])
			var r1 gc.Node
			gc.Regalloc(&r1, con.Type, t)
			gins(mips.AMOVV, &con, &r1)
			gmove(&r1, t)
			gc.Regfree(&r1)
			return
		}

		f = &con
		ft = tt // so big switch will choose a simple mov

		// constants can't move directly to memory.
		if gc.Ismem(t) {
			goto hard
		}
	}

	// value -> value copy, first operand in memory.
	// any floating point operand requires register
	// src, so goto hard to copy to register first.
	if gc.Ismem(f) && ft != tt && (gc.Isfloat[ft] || gc.Isfloat[tt]) {
		cvt = gc.Types[ft]
		goto hard
	}

	// value -> value copy, only one memory operand.
	// figure out the instruction to use.
	// break out of switch for one-instruction gins.
	// goto rdst for "destination must be register".
	// goto hard for "convert to cvt type first".
	// otherwise handle and return.

	switch uint32(ft)<<16 | uint32(tt) {
	default:
		gc.Fatalf("gmove %v -> %v", gc.Tconv(f.Type, obj.FmtLong), gc.Tconv(t.Type, obj.FmtLong))

		/*
		 * integer copy and truncate
		 */
	case gc.TINT8<<16 | gc.TINT8, // same size
		gc.TUINT8<<16 | gc.TINT8,
		gc.TINT16<<16 | gc.TINT8, // truncate
		gc.TUINT16<<16 | gc.TINT8,
		gc.TINT32<<16 | gc.TINT8,
		gc.TUINT32<<16 | gc.TINT8,
		gc.TINT64<<16 | gc.TINT8,
		gc.TUINT64<<16 | gc.TINT8:
		a = mips.AMOVB

	case gc.TINT8<<16 | gc.TUINT8, // same size
		gc.TUINT8<<16 | gc.TUINT8,
		gc.TINT16<<16 | gc.TUINT8, // truncate
		gc.TUINT16<<16 | gc.TUINT8,
		gc.TINT32<<16 | gc.TUINT8,
		gc.TUINT32<<16 | gc.TUINT8,
		gc.TINT64<<16 | gc.TUINT8,
		gc.TUINT64<<16 | gc.TUINT8:
		a = mips.AMOVBU

	case gc.TINT16<<16 | gc.TINT16, // same size
		gc.TUINT16<<16 | gc.TINT16,
		gc.TINT32<<16 | gc.TINT16, // truncate
		gc.TUINT32<<16 | gc.TINT16,
		gc.TINT64<<16 | gc.TINT16,
		gc.TUINT64<<16 | gc.TINT16:
		a = mips.AMOVH

	case gc.TINT16<<16 | gc.TUINT16, // same size
		gc.TUINT16<<16 | gc.TUINT16,
		gc.TINT32<<16 | gc.TUINT16, // truncate
		gc.TUINT32<<16 | gc.TUINT16,
		gc.TINT64<<16 | gc.TUINT16,
		gc.TUINT64<<16 | gc.TUINT16:
		a = mips.AMOVHU

	case gc.TINT32<<16 | gc.TINT32, // same size
		gc.TUINT32<<16 | gc.TINT32,
		gc.TINT64<<16 | gc.TINT32, // truncate
		gc.TUINT64<<16 | gc.TINT32:
		a = mips.AMOVW

	case gc.TINT32<<16 | gc.TUINT32, // same size
		gc.TUINT32<<16 | gc.TUINT32,
		gc.TINT64<<16 | gc.TUINT32, // truncate
		gc.TUINT64<<16 | gc.TUINT32:
		a = mips.AMOVWU

	case gc.TINT64<<16 | gc.TINT64, // same size
		gc.TINT64<<16 | gc.TUINT64,
		gc.TUINT64<<16 | gc.TINT64,
		gc.TUINT64<<16 | gc.TUINT64:
		a = mips.AMOVV

		/*
		 * integer up-conversions
		 */
	case gc.TINT8<<16 | gc.TINT16, // sign extend int8
		gc.TINT8<<16 | gc.TUINT16,
		gc.TINT8<<16 | gc.TINT32,
		gc.TINT8<<16 | gc.TUINT32,
		gc.TINT8<<16 | gc.TINT64,
		gc.TINT8<<16 | gc.TUINT64:
		a = mips.AMOVB

		goto rdst

	case gc.TUINT8<<16 | gc.TINT16, // zero extend uint8
		gc.TUINT8<<16 | gc.TUINT16,
		gc.TUINT8<<16 | gc.TINT32,
		gc.TUINT8<<16 | gc.TUINT32,
		gc.TUINT8<<16 | gc.TINT64,
		gc.TUINT8<<16 | gc.TUINT64:
		a = mips.AMOVBU

		goto rdst

	case gc.TINT16<<16 | gc.TINT32, // sign extend int16
		gc.TINT16<<16 | gc.TUINT32,
		gc.TINT16<<16 | gc.TINT64,
		gc.TINT16<<16 | gc.TUINT64:
		a = mips.AMOVH

		goto rdst

	case gc.TUINT16<<16 | gc.TINT32, // zero extend uint16
		gc.TUINT16<<16 | gc.TUINT32,
		gc.TUINT16<<16 | gc.TINT64,
		gc.TUINT16<<16 | gc.TUINT64:
		a = mips.AMOVHU

		goto rdst

	case gc.TINT32<<16 | gc.TINT64, // sign extend int32
		gc.TINT32<<16 | gc.TUINT64:
		a = mips.AMOVW

		goto rdst

	case gc.TUINT32<<16 | gc.TINT64, // zero extend uint32
		gc.TUINT32<<16 | gc.TUINT64:
		a = mips.AMOVWU

		goto rdst

		//warn("gmove: convert float to int not implemented: %N -> %N\n", f, t);
	//return;
	// algorithm is:
	//	if small enough, use native float64 -> int64 conversion.
	//	otherwise, subtract 2^63, convert, and add it back.
	/*
	* float to integer
	 */
	case gc.TFLOAT32<<16 | gc.TINT32,
		gc.TFLOAT64<<16 | gc.TINT32,
		gc.TFLOAT32<<16 | gc.TINT64,
		gc.TFLOAT64<<16 | gc.TINT64,
		gc.TFLOAT32<<16 | gc.TINT16,
		gc.TFLOAT32<<16 | gc.TINT8,
		gc.TFLOAT32<<16 | gc.TUINT16,
		gc.TFLOAT32<<16 | gc.TUINT8,
		gc.TFLOAT64<<16 | gc.TINT16,
		gc.TFLOAT64<<16 | gc.TINT8,
		gc.TFLOAT64<<16 | gc.TUINT16,
		gc.TFLOAT64<<16 | gc.TUINT8,
		gc.TFLOAT32<<16 | gc.TUINT32,
		gc.TFLOAT64<<16 | gc.TUINT32,
		gc.TFLOAT32<<16 | gc.TUINT64,
		gc.TFLOAT64<<16 | gc.TUINT64:
		bignodes()

		gc.Regalloc(&r1, gc.Types[gc.TFLOAT64], nil)
		gmove(f, &r1)
		if tt == gc.TUINT64 {
			gc.Regalloc(&r2, gc.Types[gc.TFLOAT64], nil)
			gmove(&bigf, &r2)
			gins3(mips.ACMPGED, &r1, &r2, nil)
			p1 := gc.Gbranch(mips.ABFPF, nil, 0)
			gins(mips.ASUBD, &r2, &r1)
			gc.Patch(p1, gc.Pc)
			gc.Regfree(&r2)
		}

		gc.Regalloc(&r2, gc.Types[gc.TINT64], t)
		gins(mips.ATRUNCDV, &r1, &r1)
		gins(mips.AMOVV, &r1, &r2)
		gc.Regfree(&r1)

		if tt == gc.TUINT64 {
			p1 := gc.Gbranch(mips.ABFPF, nil, 0) // use FCR0 here again
			gc.Nodreg(&r1, gc.Types[gc.TINT64], mips.REGTMP)
			gmove(&bigi, &r1)
			gins(mips.AADDVU, &r1, &r2)
			gc.Patch(p1, gc.Pc)
		}

		gmove(&r2, t)
		gc.Regfree(&r2)
		return

		//warn("gmove: convert int to float not implemented: %N -> %N\n", f, t);
	//return;
	// algorithm is:
	//	if small enough, use native int64 -> float64 conversion.
	//	otherwise, halve (rounding to odd?), convert, and double.
	/*
	 * integer to float
	 */
	case gc.TINT32<<16 | gc.TFLOAT32,
		gc.TINT32<<16 | gc.TFLOAT64,
		gc.TINT64<<16 | gc.TFLOAT32,
		gc.TINT64<<16 | gc.TFLOAT64,
		gc.TINT16<<16 | gc.TFLOAT32,
		gc.TINT16<<16 | gc.TFLOAT64,
		gc.TINT8<<16 | gc.TFLOAT32,
		gc.TINT8<<16 | gc.TFLOAT64,
		gc.TUINT16<<16 | gc.TFLOAT32,
		gc.TUINT16<<16 | gc.TFLOAT64,
		gc.TUINT8<<16 | gc.TFLOAT32,
		gc.TUINT8<<16 | gc.TFLOAT64,
		gc.TUINT32<<16 | gc.TFLOAT32,
		gc.TUINT32<<16 | gc.TFLOAT64,
		gc.TUINT64<<16 | gc.TFLOAT32,
		gc.TUINT64<<16 | gc.TFLOAT64:
		bignodes()

		var rtmp gc.Node
		gc.Regalloc(&r1, gc.Types[gc.TINT64], nil)
		gmove(f, &r1)
		if ft == gc.TUINT64 {
			gc.Nodreg(&rtmp, gc.Types[gc.TUINT64], mips.REGTMP)
			gmove(&bigi, &rtmp)
			gins(mips.AAND, &r1, &rtmp)
			p1 := ginsbranch(mips.ABEQ, nil, &rtmp, nil, 0)
			p2 := gins(mips.ASRLV, nil, &r1)
			p2.From.Type = obj.TYPE_CONST
			p2.From.Offset = 1
			gc.Patch(p1, gc.Pc)
		}

		gc.Regalloc(&r2, gc.Types[gc.TFLOAT64], t)
		gins(mips.AMOVV, &r1, &r2)
		gins(mips.AMOVVD, &r2, &r2)
		gc.Regfree(&r1)

		if ft == gc.TUINT64 {
			p1 := ginsbranch(mips.ABEQ, nil, &rtmp, nil, 0)
			gc.Nodreg(&r1, gc.Types[gc.TFLOAT64], mips.FREGTWO)
			gins(mips.AMULD, &r1, &r2)
			gc.Patch(p1, gc.Pc)
		}

		gmove(&r2, t)
		gc.Regfree(&r2)
		return

		/*
		 * float to float
		 */
	case gc.TFLOAT32<<16 | gc.TFLOAT32:
		a = mips.AMOVF

	case gc.TFLOAT64<<16 | gc.TFLOAT64:
		a = mips.AMOVD

	case gc.TFLOAT32<<16 | gc.TFLOAT64:
		a = mips.AMOVFD
		goto rdst

	case gc.TFLOAT64<<16 | gc.TFLOAT32:
		a = mips.AMOVDF
		goto rdst
	}

	gins(a, f, t)
	return

	// requires register destination
rdst:
	{
		gc.Regalloc(&r1, t.Type, t)

		gins(a, f, &r1)
		gmove(&r1, t)
		gc.Regfree(&r1)
		return
	}

	// requires register intermediate
hard:
	gc.Regalloc(&r1, cvt, t)

	gmove(f, &r1)
	gmove(&r1, t)
	gc.Regfree(&r1)
	return
}

// gins is called by the front end.
// It synthesizes some multiple-instruction sequences
// so the front end can stay simpler.
func gins(as int, f, t *gc.Node) *obj.Prog {
	if as >= obj.A_ARCHSPECIFIC {
		if x, ok := f.IntLiteral(); ok {
			ginscon(as, x, t)
			return nil // caller must not use
		}
	}
	return rawgins(as, f, t)
}

/*
 * generate one instruction:
 *	as f, r, t
 * r must be register, if not nil
 */
func gins3(as int, f, r, t *gc.Node) *obj.Prog {
	p := rawgins(as, f, t)
	if r != nil {
		p.Reg = r.Reg
	}
	return p
}

/*
 * generate one instruction:
 *	as f, t
 */
func rawgins(as int, f *gc.Node, t *gc.Node) *obj.Prog {
	// TODO(austin): Add self-move test like in 6g (but be careful
	// of truncation moves)

	p := gc.Prog(as)
	gc.Naddr(&p.From, f)
	gc.Naddr(&p.To, t)

	switch as {
	case obj.ACALL:
		if p.To.Type == obj.TYPE_REG {
			// Allow front end to emit CALL REG, and rewrite into CALL (REG).
			p.From = obj.Addr{}
			p.To.Type = obj.TYPE_MEM
			p.To.Offset = 0

			if gc.Debug['g'] != 0 {
				fmt.Printf("%v\n", p)
			}

			return p
		}

	// Bad things the front end has done to us. Crash to find call stack.
	case mips.AAND:
		if p.From.Type == obj.TYPE_CONST {
			gc.Debug['h'] = 1
			gc.Fatalf("bad inst: %v", p)
		}
	case mips.ASGT, mips.ASGTU:
		if p.From.Type == obj.TYPE_MEM || p.To.Type == obj.TYPE_MEM {
			gc.Debug['h'] = 1
			gc.Fatalf("bad inst: %v", p)
		}

	// Special cases
	case mips.AMUL, mips.AMULU, mips.AMULV, mips.AMULVU:
		if p.From.Type == obj.TYPE_CONST {
			gc.Debug['h'] = 1
			gc.Fatalf("bad inst: %v", p)
		}

		pp := gc.Prog(mips.AMOVV)
		pp.From.Type = obj.TYPE_REG
		pp.From.Reg = mips.REG_LO
		pp.To = p.To

		p.Reg = p.To.Reg
		p.To = obj.Addr{}

	case mips.ASUBVU:
		// unary
		if f == nil {
			p.From = p.To
			p.Reg = mips.REGZERO
		}
	}

	if gc.Debug['g'] != 0 {
		fmt.Printf("%v\n", p)
	}

	w := int32(0)
	switch as {
	case mips.AMOVB,
		mips.AMOVBU:
		w = 1

	case mips.AMOVH,
		mips.AMOVHU:
		w = 2

	case mips.AMOVW,
		mips.AMOVWU:
		w = 4

	case mips.AMOVV:
		if p.From.Type == obj.TYPE_CONST || p.From.Type == obj.TYPE_ADDR {
			break
		}
		w = 8
	}

	if w != 0 && ((f != nil && p.From.Width < int64(w)) || (t != nil && p.To.Type != obj.TYPE_REG && p.To.Width > int64(w))) {
		gc.Dump("f", f)
		gc.Dump("t", t)
		gc.Fatalf("bad width: %v (%d, %d)\n", p, p.From.Width, p.To.Width)
	}

	return p
}

/*
 * return Axxx for Oxxx on type t.
 */
func optoas(op int, t *gc.Type) int {
	if t == nil {
		gc.Fatalf("optoas: t is nil")
	}

	a := int(obj.AXXX)
	switch uint32(op)<<16 | uint32(gc.Simtype[t.Etype]) {
	default:
		gc.Fatalf("optoas: no entry for op=%v type=%v", gc.Oconv(int(op), 0), t)

	case gc.OEQ<<16 | gc.TBOOL,
		gc.OEQ<<16 | gc.TINT8,
		gc.OEQ<<16 | gc.TUINT8,
		gc.OEQ<<16 | gc.TINT16,
		gc.OEQ<<16 | gc.TUINT16,
		gc.OEQ<<16 | gc.TINT32,
		gc.OEQ<<16 | gc.TUINT32,
		gc.OEQ<<16 | gc.TINT64,
		gc.OEQ<<16 | gc.TUINT64,
		gc.OEQ<<16 | gc.TPTR32,
		gc.OEQ<<16 | gc.TPTR64:
		a = mips.ABEQ

	case gc.OEQ<<16 | gc.TFLOAT32, // ACMPEQF
		gc.OEQ<<16 | gc.TFLOAT64: // ACMPEQD
		a = mips.ABFPT

	case gc.ONE<<16 | gc.TBOOL,
		gc.ONE<<16 | gc.TINT8,
		gc.ONE<<16 | gc.TUINT8,
		gc.ONE<<16 | gc.TINT16,
		gc.ONE<<16 | gc.TUINT16,
		gc.ONE<<16 | gc.TINT32,
		gc.ONE<<16 | gc.TUINT32,
		gc.ONE<<16 | gc.TINT64,
		gc.ONE<<16 | gc.TUINT64,
		gc.ONE<<16 | gc.TPTR32,
		gc.ONE<<16 | gc.TPTR64:
		a = mips.ABNE

	case gc.ONE<<16 | gc.TFLOAT32, // ACMPEQF
		gc.ONE<<16 | gc.TFLOAT64: // ACMPEQD
		a = mips.ABFPF

	case gc.OLT<<16 | gc.TINT8, // ASGT
		gc.OLT<<16 | gc.TINT16,
		gc.OLT<<16 | gc.TINT32,
		gc.OLT<<16 | gc.TINT64,
		gc.OLT<<16 | gc.TUINT8, // ASGTU
		gc.OLT<<16 | gc.TUINT16,
		gc.OLT<<16 | gc.TUINT32,
		gc.OLT<<16 | gc.TUINT64:
		a = mips.ABNE

	case gc.OLT<<16 | gc.TFLOAT32, // ACMPGEF
		gc.OLT<<16 | gc.TFLOAT64: // ACMPGED
		a = mips.ABFPT

	case gc.OLE<<16 | gc.TINT8, // ASGT
		gc.OLE<<16 | gc.TINT16,
		gc.OLE<<16 | gc.TINT32,
		gc.OLE<<16 | gc.TINT64,
		gc.OLE<<16 | gc.TUINT8, // ASGTU
		gc.OLE<<16 | gc.TUINT16,
		gc.OLE<<16 | gc.TUINT32,
		gc.OLE<<16 | gc.TUINT64:
		a = mips.ABEQ

	case gc.OLE<<16 | gc.TFLOAT32, // ACMPGTF
		gc.OLE<<16 | gc.TFLOAT64: // ACMPGTD
		a = mips.ABFPT

	case gc.OGT<<16 | gc.TINT8, // ASGT
		gc.OGT<<16 | gc.TINT16,
		gc.OGT<<16 | gc.TINT32,
		gc.OGT<<16 | gc.TINT64,
		gc.OGT<<16 | gc.TUINT8, // ASGTU
		gc.OGT<<16 | gc.TUINT16,
		gc.OGT<<16 | gc.TUINT32,
		gc.OGT<<16 | gc.TUINT64:
		a = mips.ABNE

	case gc.OGT<<16 | gc.TFLOAT32, // ACMPGTF
		gc.OGT<<16 | gc.TFLOAT64: // ACMPGTD
		a = mips.ABFPT

	case gc.OGE<<16 | gc.TINT8, // ASGT
		gc.OGE<<16 | gc.TINT16,
		gc.OGE<<16 | gc.TINT32,
		gc.OGE<<16 | gc.TINT64,
		gc.OGE<<16 | gc.TUINT8, // ASGTU
		gc.OGE<<16 | gc.TUINT16,
		gc.OGE<<16 | gc.TUINT32,
		gc.OGE<<16 | gc.TUINT64:
		a = mips.ABEQ

	case gc.OGE<<16 | gc.TFLOAT32, // ACMPGEF
		gc.OGE<<16 | gc.TFLOAT64: // ACMPGED
		a = mips.ABFPT

	case gc.OAS<<16 | gc.TBOOL,
		gc.OAS<<16 | gc.TINT8:
		a = mips.AMOVB

	case gc.OAS<<16 | gc.TUINT8:
		a = mips.AMOVBU

	case gc.OAS<<16 | gc.TINT16:
		a = mips.AMOVH

	case gc.OAS<<16 | gc.TUINT16:
		a = mips.AMOVHU

	case gc.OAS<<16 | gc.TINT32:
		a = mips.AMOVW

	case gc.OAS<<16 | gc.TUINT32,
		gc.OAS<<16 | gc.TPTR32:
		a = mips.AMOVWU

	case gc.OAS<<16 | gc.TINT64,
		gc.OAS<<16 | gc.TUINT64,
		gc.OAS<<16 | gc.TPTR64:
		a = mips.AMOVV

	case gc.OAS<<16 | gc.TFLOAT32:
		a = mips.AMOVF

	case gc.OAS<<16 | gc.TFLOAT64:
		a = mips.AMOVD

	case gc.OADD<<16 | gc.TINT8,
		gc.OADD<<16 | gc.TUINT8,
		gc.OADD<<16 | gc.TINT16,
		gc.OADD<<16 | gc.TUINT16,
		gc.OADD<<16 | gc.TINT32,
		gc.OADD<<16 | gc.TUINT32,
		gc.OADD<<16 | gc.TPTR32:
		a = mips.AADDU

	case gc.OADD<<16 | gc.TINT64,
		gc.OADD<<16 | gc.TUINT64,
		gc.OADD<<16 | gc.TPTR64:
		a = mips.AADDVU

	case gc.OADD<<16 | gc.TFLOAT32:
		a = mips.AADDF

	case gc.OADD<<16 | gc.TFLOAT64:
		a = mips.AADDD

	case gc.OSUB<<16 | gc.TINT8,
		gc.OSUB<<16 | gc.TUINT8,
		gc.OSUB<<16 | gc.TINT16,
		gc.OSUB<<16 | gc.TUINT16,
		gc.OSUB<<16 | gc.TINT32,
		gc.OSUB<<16 | gc.TUINT32,
		gc.OSUB<<16 | gc.TPTR32:
		a = mips.ASUBU

	case gc.OSUB<<16 | gc.TINT64,
		gc.OSUB<<16 | gc.TUINT64,
		gc.OSUB<<16 | gc.TPTR64:
		a = mips.ASUBVU

	case gc.OSUB<<16 | gc.TFLOAT32:
		a = mips.ASUBF

	case gc.OSUB<<16 | gc.TFLOAT64:
		a = mips.ASUBD

	case gc.OMINUS<<16 | gc.TINT8,
		gc.OMINUS<<16 | gc.TUINT8,
		gc.OMINUS<<16 | gc.TINT16,
		gc.OMINUS<<16 | gc.TUINT16,
		gc.OMINUS<<16 | gc.TINT32,
		gc.OMINUS<<16 | gc.TUINT32,
		gc.OMINUS<<16 | gc.TPTR32,
		gc.OMINUS<<16 | gc.TINT64,
		gc.OMINUS<<16 | gc.TUINT64,
		gc.OMINUS<<16 | gc.TPTR64:
		a = mips.ASUBVU

	case gc.OAND<<16 | gc.TINT8,
		gc.OAND<<16 | gc.TUINT8,
		gc.OAND<<16 | gc.TINT16,
		gc.OAND<<16 | gc.TUINT16,
		gc.OAND<<16 | gc.TINT32,
		gc.OAND<<16 | gc.TUINT32,
		gc.OAND<<16 | gc.TPTR32,
		gc.OAND<<16 | gc.TINT64,
		gc.OAND<<16 | gc.TUINT64,
		gc.OAND<<16 | gc.TPTR64:
		a = mips.AAND

	case gc.OOR<<16 | gc.TINT8,
		gc.OOR<<16 | gc.TUINT8,
		gc.OOR<<16 | gc.TINT16,
		gc.OOR<<16 | gc.TUINT16,
		gc.OOR<<16 | gc.TINT32,
		gc.OOR<<16 | gc.TUINT32,
		gc.OOR<<16 | gc.TPTR32,
		gc.OOR<<16 | gc.TINT64,
		gc.OOR<<16 | gc.TUINT64,
		gc.OOR<<16 | gc.TPTR64:
		a = mips.AOR

	case gc.OXOR<<16 | gc.TINT8,
		gc.OXOR<<16 | gc.TUINT8,
		gc.OXOR<<16 | gc.TINT16,
		gc.OXOR<<16 | gc.TUINT16,
		gc.OXOR<<16 | gc.TINT32,
		gc.OXOR<<16 | gc.TUINT32,
		gc.OXOR<<16 | gc.TPTR32,
		gc.OXOR<<16 | gc.TINT64,
		gc.OXOR<<16 | gc.TUINT64,
		gc.OXOR<<16 | gc.TPTR64:
		a = mips.AXOR

		// TODO(minux): handle rotates
	//case CASE(OLROT, TINT8):
	//case CASE(OLROT, TUINT8):
	//case CASE(OLROT, TINT16):
	//case CASE(OLROT, TUINT16):
	//case CASE(OLROT, TINT32):
	//case CASE(OLROT, TUINT32):
	//case CASE(OLROT, TPTR32):
	//case CASE(OLROT, TINT64):
	//case CASE(OLROT, TUINT64):
	//case CASE(OLROT, TPTR64):
	//	a = 0//???; RLDC?
	//	break;

	case gc.OLSH<<16 | gc.TINT8,
		gc.OLSH<<16 | gc.TUINT8,
		gc.OLSH<<16 | gc.TINT16,
		gc.OLSH<<16 | gc.TUINT16,
		gc.OLSH<<16 | gc.TINT32,
		gc.OLSH<<16 | gc.TUINT32,
		gc.OLSH<<16 | gc.TPTR32,
		gc.OLSH<<16 | gc.TINT64,
		gc.OLSH<<16 | gc.TUINT64,
		gc.OLSH<<16 | gc.TPTR64:
		a = mips.ASLLV

	case gc.ORSH<<16 | gc.TUINT8,
		gc.ORSH<<16 | gc.TUINT16,
		gc.ORSH<<16 | gc.TUINT32,
		gc.ORSH<<16 | gc.TPTR32,
		gc.ORSH<<16 | gc.TUINT64,
		gc.ORSH<<16 | gc.TPTR64:
		a = mips.ASRLV

	case gc.ORSH<<16 | gc.TINT8,
		gc.ORSH<<16 | gc.TINT16,
		gc.ORSH<<16 | gc.TINT32,
		gc.ORSH<<16 | gc.TINT64:
		a = mips.ASRAV

		// TODO(minux): handle rotates
	//case CASE(ORROTC, TINT8):
	//case CASE(ORROTC, TUINT8):
	//case CASE(ORROTC, TINT16):
	//case CASE(ORROTC, TUINT16):
	//case CASE(ORROTC, TINT32):
	//case CASE(ORROTC, TUINT32):
	//case CASE(ORROTC, TINT64):
	//case CASE(ORROTC, TUINT64):
	//	a = 0//??? RLDC??
	//	break;

	case gc.OHMUL<<16 | gc.TINT64:
		a = mips.AMULV

	case gc.OHMUL<<16 | gc.TUINT64,
		gc.OHMUL<<16 | gc.TPTR64:
		a = mips.AMULVU

	case gc.OMUL<<16 | gc.TINT8,
		gc.OMUL<<16 | gc.TINT16,
		gc.OMUL<<16 | gc.TINT32,
		gc.OMUL<<16 | gc.TINT64:
		a = mips.AMULV

	case gc.OMUL<<16 | gc.TUINT8,
		gc.OMUL<<16 | gc.TUINT16,
		gc.OMUL<<16 | gc.TUINT32,
		gc.OMUL<<16 | gc.TPTR32,
		gc.OMUL<<16 | gc.TUINT64,
		gc.OMUL<<16 | gc.TPTR64:
		a = mips.AMULVU

	case gc.OMUL<<16 | gc.TFLOAT32:
		a = mips.AMULF

	case gc.OMUL<<16 | gc.TFLOAT64:
		a = mips.AMULD

	case gc.ODIV<<16 | gc.TINT8,
		gc.ODIV<<16 | gc.TINT16,
		gc.ODIV<<16 | gc.TINT32,
		gc.ODIV<<16 | gc.TINT64:
		a = mips.ADIVV

	case gc.ODIV<<16 | gc.TUINT8,
		gc.ODIV<<16 | gc.TUINT16,
		gc.ODIV<<16 | gc.TUINT32,
		gc.ODIV<<16 | gc.TPTR32,
		gc.ODIV<<16 | gc.TUINT64,
		gc.ODIV<<16 | gc.TPTR64:
		a = mips.ADIVVU

	case gc.ODIV<<16 | gc.TFLOAT32:
		a = mips.ADIVF

	case gc.ODIV<<16 | gc.TFLOAT64:
		a = mips.ADIVD
	}

	return a
}

const (
	ODynam   = 1 << 0
	OAddable = 1 << 1
)

func xgen(n *gc.Node, a *gc.Node, o int) bool {
	// TODO(minux)

	return -1 != 0 /*TypeKind(100016)*/
}

func sudoclean() {
	return
}

/*
 * generate code to compute address of n,
 * a reference to a (perhaps nested) field inside
 * an array or struct.
 * return 0 on failure, 1 on success.
 * on success, leaves usable address in a.
 *
 * caller is responsible for calling sudoclean
 * after successful sudoaddable,
 * to release the register used for a.
 */
func sudoaddable(as int, n *gc.Node, a *obj.Addr) bool {
	// TODO(minux)

	*a = obj.Addr{}
	return false
}
