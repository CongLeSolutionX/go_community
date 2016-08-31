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
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
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

package mips32

import (
	"cmd/compile/internal/big"
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
	"encoding/binary"
	"fmt"
)

var resvd = []int{
	mips32.REGZERO,
	mips32.REGSP,   // reserved for SP
	mips32.REGLINK, // reserved for link
	mips32.REGG,
	mips32.REGTMP,
	mips32.REG_R26, // kernel
	mips32.REG_R27, // kernel

	mips32.REG_F1, // odd fp registers hold hi parts of float64
	mips32.REG_F3,
	mips32.REG_F5,
	mips32.REG_F7,
	mips32.REG_F9,
	mips32.REG_F11,
	mips32.REG_F13,
	mips32.REG_F15,
	mips32.REG_F17,
	mips32.REG_F19,
	mips32.REG_F21,
	mips32.REG_F23,
	mips32.REG_F25,
	mips32.REG_F27,
	mips32.REG_F29,
	mips32.REG_F31,
}

/*
 * generate
 *	as $c, n
 */
func ginscon(as obj.As, c int64, n2 *gc.Node) {
	var n1 gc.Node

	gc.Nodconst(&n1, gc.Types[gc.TINT32], c)

	if (c < -mips32.BIG || c > mips32.BIG) || n2.Op != gc.OREGISTER || as == mips32.AMUL || as == mips32.AMULU {
		// cannot have more than 16-bit of immediate in ADD, etc.
		// instead, MOV into register first.
		var ntmp gc.Node
		gc.Regalloc(&ntmp, gc.Types[gc.TINT32], nil)

		rawgins(mips32.AMOVW, &n1, &ntmp)
		rawgins(as, &ntmp, n2)
		gc.Regfree(&ntmp)
		return
	}

	rawgins(as, &n1, n2)
}

// generate branch
// n1, n2 are registers
func ginsbranch(as obj.As, t *gc.Type, n1, n2 *gc.Node, likely int) *obj.Prog {
	p := gc.Gbranch(as, t, likely)
	gc.Naddr(&p.From, n1)
	if n2 != nil {
		p.Reg = n2.Reg
	}
	return p
}

func ginscmp(op gc.Op, t *gc.Type, n1, n2 *gc.Node, likely int) *obj.Prog {
	if !t.IsFloat() && (op == gc.OLT || op == gc.OGE) {
		// swap nodes to fit SGT instruction
		n1, n2 = n2, n1
	}
	if t.IsFloat() && (op == gc.OLT || op == gc.OLE) {
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
	gc.Nodreg(&ntmp, gc.Types[gc.TINT], mips32.REGTMP)

	switch gc.Simtype[t.Etype] {
	default:
		gc.Fatalf("ginscmp: no entry for op=%v type=%v", op, t)

	case gc.TINT8,
		gc.TINT16,
		gc.TINT32:
		if op == gc.OEQ || op == gc.ONE {
			p = ginsbranch(optoas(op, t), nil, &r1, &r2, likely)
		} else {
			gins3(mips32.ASGT, &r1, &r2, &ntmp)

			p = ginsbranch(optoas(op, t), nil, &ntmp, nil, likely)
		}

	case gc.TBOOL,
		gc.TUINT8,
		gc.TUINT16,
		gc.TUINT32,
		gc.TPTR32:
		if op == gc.OEQ || op == gc.ONE {
			p = ginsbranch(optoas(op, t), nil, &r1, &r2, likely)
		} else {
			gins3(mips32.ASGTU, &r1, &r2, &ntmp)

			p = ginsbranch(optoas(op, t), nil, &ntmp, nil, likely)
		}
	case gc.TFLOAT32:
		switch op {
		default:
			gc.Fatalf("ginscmp: no entry for op=%v type=%v", op, t)

		case gc.OEQ,
			gc.ONE:
			gins3(mips32.ACMPEQF, &r1, &r2, nil)

		case gc.OGE:
			gins3(mips32.ACMPGEF, &r1, &r2, nil)

		case gc.OGT:
			gins3(mips32.ACMPGTF, &r1, &r2, nil)
		}
		p = gc.Gbranch(optoas(op, t), nil, likely)

	case gc.TFLOAT64:
		switch op {
		default:
			gc.Fatalf("ginscmp: no entry for op=%v type=%v", op, t)

		case gc.OEQ,
			gc.ONE:
			gins3(mips32.ACMPEQD, &r1, &r2, nil)

		case gc.OGE:
			gins3(mips32.ACMPGED, &r1, &r2, nil)

		case gc.OGT:
			gins3(mips32.ACMPGTD, &r1, &r2, nil)
		}
		p = gc.Gbranch(optoas(op, t), nil, likely)
	}

	gc.Regfree(&g2)
	gc.Regfree(&r2)
	gc.Regfree(&g1)
	gc.Regfree(&r1)

	return p
}

// set up nodes
var (
	two31i       gc.Node
	two31f       gc.Node
	two32f       gc.Node
	bignodes_did bool
)

func bignodes() {
	if bignodes_did {
		return
	}
	bignodes_did = true

	var i big.Int
	i.SetInt64(1)
	i.Lsh(&i, 31)

	gc.Nodconst(&two31i, gc.Types[gc.TUINT32], 0)
	two31i.SetBigInt(&i)

	two31i.Convconst(&two31f, gc.Types[gc.TFLOAT64])

	i.Lsh(&i, 1)
	var big gc.Node
	gc.Nodconst(&big, gc.Types[gc.TUINT64], 0)
	big.SetBigInt(&i)

	big.Convconst(&two32f, gc.Types[gc.TFLOAT64])
}

/*
 * return constant i node.
 * overwritten by next call, but useful in calls to gins.
 */

var ncon_n gc.Node

func ncon(i uint32) *gc.Node {
	if ncon_n.Type == nil {
		gc.Nodconst(&ncon_n, gc.Types[gc.TUINT32], 0)
	}
	ncon_n.SetInt(int64(i))
	return &ncon_n
}

var sclean [10]gc.Node

var nsclean int

/*
 * n is a 64-bit value.  fill in lo and hi to refer to its 32-bit halves.
 */
func split64(n *gc.Node, lo *gc.Node, hi *gc.Node) {
	if !gc.Is64(n.Type) {
		gc.Fatalf("split64 %v", n.Type)
	}

	if nsclean >= len(sclean) {
		gc.Fatalf("split64 clean")
		return
	}
	sclean[nsclean].Op = gc.OEMPTY
	nsclean++
	switch n.Op {
	default:
		switch n.Op {
		default:
			var n1 gc.Node
			if !dotaddable(n, &n1) {
				gc.Igen(n, &n1, nil)
				sclean[nsclean-1] = n1
			}

			n = &n1

		case gc.ONAME:
		case gc.OINDREG:
			break
		}

		*lo = *n
		*hi = *n
		lo.Type = gc.Types[gc.TUINT32]
		if n.Type.Etype == gc.TINT64 {
			hi.Type = gc.Types[gc.TINT32]
		} else {
			hi.Type = gc.Types[gc.TUINT32]
		}

		if gc.Ctxt.Arch.ByteOrder == binary.BigEndian {
			lo.Xoffset += 4
		} else {
			hi.Xoffset += 4
		}

	case gc.OLITERAL:
		var n1 gc.Node
		n.Convconst(&n1, n.Type)
		i := n1.Int64()
		gc.Nodconst(lo, gc.Types[gc.TUINT32], int64(uint32(i)))
		i >>= 32
		if n.Type.Etype == gc.TINT64 {
			gc.Nodconst(hi, gc.Types[gc.TINT32], int64(int32(i)))
		} else {
			gc.Nodconst(hi, gc.Types[gc.TUINT32], int64(uint32(i)))
		}
	}
}

func splitclean() {
	if nsclean <= 0 {
		gc.Fatalf("splitclean")
		return
	}
	nsclean--
	if sclean[nsclean].Op != gc.OEMPTY {
		gc.Regfree(&sclean[nsclean])
	}
}

func dotaddable(n *gc.Node, n1 *gc.Node) bool {
	if n.Op != gc.ODOT {
		return false
	}

	var oary [10]int64
	var nn *gc.Node
	o := gc.Dotoffset(n, oary[:], &nn)
	if nn != nil && nn.Addable && o == 1 && oary[0] >= 0 {
		*n1 = *nn
		n1.Type = n.Type
		n1.Xoffset += oary[0]
		return true
	}

	return false
}

///////////////////////////

/*
 * generate move:
 *	t = f
 * hard part is conversions.
 */
func gmove(f *gc.Node, t *gc.Node) {
	if gc.Debug['M'] != 0 {
		fmt.Printf("gmove %v -> %v\n", gc.Nconv(f, gc.FmtLong), gc.Nconv(t, gc.FmtLong))
	}

	ft := int(gc.Simsimtype(f.Type))
	tt := int(gc.Simsimtype(t.Type))
	cvt := t.Type

	if gc.Iscomplex[ft] || gc.Iscomplex[tt] {
		gc.Complexmove(f, t)
		return
	}

	// cannot have two memory operands
	var r2 gc.Node
	var r1 gc.Node
	var a obj.As
	if !gc.Is64(f.Type) && !gc.Is64(t.Type) && gc.Ismem(f) && gc.Ismem(t) {
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
			f.Convconst(&con, gc.Types[gc.TINT32])
			var r1 gc.Node
			gc.Regalloc(&r1, con.Type, t)
			gins(mips32.AMOVW, &con, &r1)
			gmove(&r1, t)
			gc.Regfree(&r1)
			return

		case gc.TUINT32,
			gc.TUINT16,
			gc.TUINT8:
			var con gc.Node
			f.Convconst(&con, gc.Types[gc.TUINT32])
			var r1 gc.Node
			gc.Regalloc(&r1, con.Type, t)
			gins(mips32.AMOVW, &con, &r1)
			gmove(&r1, t)
			gc.Regfree(&r1)
			return
		}

		f = &con
		ft = tt // so big switch will choose a simple mov

		// constants can't move directly to memory.
		if gc.Ismem(t) && !gc.Is64(t.Type) {
			goto hard
		}
	}

	// value -> value copy, first operand in memory.
	// any floating point operand requires register
	// src, so goto hard to copy to register first.
	//	if !gc.Is64(f.Type) && !gc.Is64(t.Type) && gc.Ismem(f) && gc.Ismem(t) && (gc.Isfloat[ft] || gc.Isfloat[tt]) { // +arm conditions
	if gc.Ismem(f) && ft != tt && (gc.Isfloat[ft] || gc.Isfloat[tt]) {
		if !gc.Is64(f.Type) && !gc.Is64(t.Type) { // && gc.Ismem(f) && gc.Ismem(t) {//arm condition
			cvt = gc.Types[ft]
			goto hard
		}
	}

	// value -> value copy, only one memory operand.
	// figure out the instruction to use.
	// break out of switch for one-instruction gins.
	// goto rdst for "destination must be register".
	// goto hard for "convert to cvt type first".
	// otherwise handle and return.

	switch uint32(ft)<<16 | uint32(tt) {
	default:
		gc.Fatalf("gmove %v -> %v", gc.Tconv(f.Type, gc.FmtLong), gc.Tconv(t.Type, gc.FmtLong))

		/*
		 * integer copy and truncate
		 */
	case gc.TINT8<<16 | gc.TINT8, // same size
		gc.TUINT8<<16 | gc.TINT8,
		gc.TINT16<<16 | gc.TINT8, // truncate
		gc.TUINT16<<16 | gc.TINT8,
		gc.TINT32<<16 | gc.TINT8,
		gc.TUINT32<<16 | gc.TINT8:
		a = mips32.AMOVB

	case gc.TINT64<<16 | gc.TINT8,
		gc.TUINT64<<16 | gc.TINT8:

		if gc.Ctxt.Arch.ByteOrder == binary.BigEndian {
			f.Xoffset += 7
		}
		a = mips32.AMOVB

	case gc.TINT8<<16 | gc.TUINT8, // same size
		gc.TUINT8<<16 | gc.TUINT8,
		gc.TINT16<<16 | gc.TUINT8, // truncate
		gc.TUINT16<<16 | gc.TUINT8,
		gc.TINT32<<16 | gc.TUINT8,
		gc.TUINT32<<16 | gc.TUINT8:
		a = mips32.AMOVBU

	case gc.TINT64<<16 | gc.TUINT8,
		gc.TUINT64<<16 | gc.TUINT8:

		if gc.Ctxt.Arch.ByteOrder == binary.BigEndian {
			f.Xoffset += 7
		}
		a = mips32.AMOVBU

	case gc.TINT16<<16 | gc.TINT16, // same size
		gc.TUINT16<<16 | gc.TINT16,
		gc.TINT32<<16 | gc.TINT16, // truncate
		gc.TUINT32<<16 | gc.TINT16:
		a = mips32.AMOVH

	case gc.TINT64<<16 | gc.TINT16,
		gc.TUINT64<<16 | gc.TINT16:

		if gc.Ctxt.Arch.ByteOrder == binary.BigEndian {
			f.Xoffset += 6
		}
		a = mips32.AMOVH

	case gc.TINT16<<16 | gc.TUINT16, // same size
		gc.TUINT16<<16 | gc.TUINT16,
		gc.TINT32<<16 | gc.TUINT16, // truncate
		gc.TUINT32<<16 | gc.TUINT16:
		a = mips32.AMOVHU

	case gc.TINT64<<16 | gc.TUINT16,
		gc.TUINT64<<16 | gc.TUINT16:

		if gc.Ctxt.Arch.ByteOrder == binary.BigEndian {
			f.Xoffset += 6
		}
		a = mips32.AMOVHU

	case gc.TINT32<<16 | gc.TINT32, // same size
		gc.TUINT32<<16 | gc.TINT32:

		a = mips32.AMOVW

	case gc.TINT32<<16 | gc.TUINT32, // same size
		gc.TUINT32<<16 | gc.TUINT32:
		a = mips32.AMOVW
	case gc.TINT64<<16 | gc.TINT32, // truncate
		gc.TUINT64<<16 | gc.TINT32,
		gc.TINT64<<16 | gc.TUINT32,
		gc.TUINT64<<16 | gc.TUINT32:
		var flo gc.Node
		var fhi gc.Node
		split64(f, &flo, &fhi)

		var r1 gc.Node
		gc.Regalloc(&r1, t.Type, nil)
		gins(mips32.AMOVW, &flo, &r1)
		gins(mips32.AMOVW, &r1, t)
		gc.Regfree(&r1)
		splitclean()
		return

	case gc.TINT64<<16 | gc.TINT64, // same size
		gc.TINT64<<16 | gc.TUINT64,
		gc.TUINT64<<16 | gc.TINT64,
		gc.TUINT64<<16 | gc.TUINT64:
		var fhi gc.Node
		var flo gc.Node
		split64(f, &flo, &fhi)

		var tlo gc.Node
		var thi gc.Node
		split64(t, &tlo, &thi)
		var r1 gc.Node
		gc.Regalloc(&r1, flo.Type, nil)
		var r2 gc.Node
		gc.Regalloc(&r2, fhi.Type, nil)
		gins(mips32.AMOVW, &flo, &r1)
		gins(mips32.AMOVW, &fhi, &r2)
		gins(mips32.AMOVW, &r1, &tlo)
		gins(mips32.AMOVW, &r2, &thi)
		gc.Regfree(&r1)
		gc.Regfree(&r2)
		splitclean()
		splitclean()
		return
		//////////////////////////////////////////////////

		/*
		 * integer up-conversions
		 */
	case gc.TINT8<<16 | gc.TINT16, // sign extend int8
		gc.TINT8<<16 | gc.TUINT16,
		gc.TINT8<<16 | gc.TINT32,
		gc.TINT8<<16 | gc.TUINT32:

		a = mips32.AMOVB

		goto rdst

	case gc.TINT8<<16 | gc.TINT64, // convert via int32
		gc.TINT8<<16 | gc.TUINT64:
		cvt = gc.Types[gc.TINT32]

		goto hard

	case gc.TUINT8<<16 | gc.TINT16, // zero extend uint8
		gc.TUINT8<<16 | gc.TUINT16,
		gc.TUINT8<<16 | gc.TINT32,
		gc.TUINT8<<16 | gc.TUINT32:
		a = mips32.AMOVBU

		goto rdst
	case gc.TUINT8<<16 | gc.TINT64, // convert via uint32
		gc.TUINT8<<16 | gc.TUINT64:
		cvt = gc.Types[gc.TUINT32]

		goto hard

	case gc.TINT16<<16 | gc.TINT32, // sign extend int16
		gc.TINT16<<16 | gc.TUINT32:
		a = mips32.AMOVH
		goto rdst

	case gc.TINT16<<16 | gc.TINT64,
		gc.TINT16<<16 | gc.TUINT64:
		cvt = gc.Types[gc.TINT32]
		goto hard

	case gc.TUINT16<<16 | gc.TINT32, // zero extend uint16
		gc.TUINT16<<16 | gc.TUINT32:

		a = mips32.AMOVHU
		goto rdst

	case gc.TUINT16<<16 | gc.TINT64,
		gc.TUINT16<<16 | gc.TUINT64:
		cvt = gc.Types[gc.TUINT32]

		goto hard

	case gc.TINT32<<16 | gc.TINT64, // sign extend int32
		gc.TINT32<<16 | gc.TUINT64:
		var tlo gc.Node
		var thi gc.Node
		split64(t, &tlo, &thi)

		var r1 gc.Node
		gc.Regalloc(&r1, tlo.Type, nil)
		var r2 gc.Node
		gc.Regalloc(&r2, thi.Type, nil)
		gmove(f, &r1)
		gins3(mips32.ASRA, ncon(31), &r1, &r2)
		gins(mips32.AMOVW, &r1, &tlo)

		gins(mips32.AMOVW, &r2, &thi)
		gc.Regfree(&r1)
		gc.Regfree(&r2)
		splitclean()
		return

	case gc.TUINT32<<16 | gc.TINT64, // zero extend uint32
		gc.TUINT32<<16 | gc.TUINT64:
		var thi gc.Node
		var tlo gc.Node
		split64(t, &tlo, &thi)

		gmove(f, &tlo)
		var r1 gc.Node
		gc.Regalloc(&r1, thi.Type, nil)
		gins(mips32.AMOVW, ncon(0), &r1)
		gins(mips32.AMOVW, &r1, &thi)
		gc.Regfree(&r1)
		splitclean()
		return

	// algorithm is:
	//	if small enough, use native float64 -> int32 conversion.
	//	otherwise, subtract 2^31, convert, and add it back.
	/*
	* float to integer
	 */
	case gc.TFLOAT32<<16 | gc.TINT32,
		gc.TFLOAT64<<16 | gc.TINT32,
		gc.TFLOAT32<<16 | gc.TINT16,
		gc.TFLOAT32<<16 | gc.TINT8,
		gc.TFLOAT32<<16 | gc.TUINT16,
		gc.TFLOAT32<<16 | gc.TUINT8,
		gc.TFLOAT64<<16 | gc.TINT16,
		gc.TFLOAT64<<16 | gc.TINT8,
		gc.TFLOAT64<<16 | gc.TUINT16,
		gc.TFLOAT64<<16 | gc.TUINT8,
		gc.TFLOAT32<<16 | gc.TUINT32,
		gc.TFLOAT64<<16 | gc.TUINT32:
		bignodes()

		gc.Regalloc(&r1, gc.Types[gc.TFLOAT64], nil)
		gmove(f, &r1)
		if tt == gc.TUINT32 {
			gc.Regalloc(&r2, gc.Types[gc.TFLOAT64], nil)
			gmove(&two31f, &r2)
			gins3(mips32.ACMPGED, &r1, &r2, nil)
			p1 := gc.Gbranch(mips32.ABFPF, nil, 0)
			gins(mips32.ASUBD, &r2, &r1)
			gc.Patch(p1, gc.Pc)
			gc.Regfree(&r2)
		}

		gc.Regalloc(&r2, gc.Types[gc.TINT32], t)
		gins(mips32.ATRUNCDW, &r1, &r1)
		gins(mips32.AMOVW, &r1, &r2)
		gc.Regfree(&r1)

		if tt == gc.TUINT32 {
			p1 := gc.Gbranch(mips32.ABFPF, nil, 0) // use FCR0 here again
			gc.Nodreg(&r1, gc.Types[gc.TINT32], mips32.REGTMP)
			gmove(&two31i, &r1)
			gins(mips32.AADDU, &r1, &r2)
			gc.Patch(p1, gc.Pc)
		}

		gmove(&r2, t)
		gc.Regfree(&r2)
		return

		// algorithm is:
		//	use native int32 -> float64 conversion.
		//	in case of f:uint32 if ((int32)f) < 0 { t += 2^32 }
		/*
		 * integer to float
		 */
	case gc.TINT32<<16 | gc.TFLOAT32,
		gc.TINT32<<16 | gc.TFLOAT64,
		gc.TINT16<<16 | gc.TFLOAT32,
		gc.TINT16<<16 | gc.TFLOAT64,
		gc.TINT8<<16 | gc.TFLOAT32,
		gc.TINT8<<16 | gc.TFLOAT64,
		gc.TUINT16<<16 | gc.TFLOAT32,
		gc.TUINT16<<16 | gc.TFLOAT64,
		gc.TUINT8<<16 | gc.TFLOAT32,
		gc.TUINT8<<16 | gc.TFLOAT64,
		gc.TUINT32<<16 | gc.TFLOAT32,
		gc.TUINT32<<16 | gc.TFLOAT64:

		bignodes()
		gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
		gmove(f, &r1)

		gc.Regalloc(&r2, gc.Types[gc.TFLOAT64], t)
		gins(mips32.AMOVW, &r1, &r2)
		gins(mips32.AMOVWD, &r2, &r2)

		if ft == gc.TUINT32 {
			var rtmp gc.Node
			gc.Nodreg(&rtmp, gc.Types[gc.TUINT32], mips32.REGTMP)
			gins3(mips32.ASGT, ncon(0), &r1, &rtmp)

			p1 := ginsbranch(mips32.ABEQ, nil, &rtmp, nil, 0)

			var r3 gc.Node
			gc.Regalloc(&r3, gc.Types[gc.TFLOAT64], nil)
			gmove(&two32f, &r3)
			gins(mips32.AADDD, &r3, &r2)
			gc.Patch(p1, gc.Pc)
			gc.Regfree(&r3)
		}

		gmove(&r2, t)
		gc.Regfree(&r1)
		gc.Regfree(&r2)
		return

		/*
		 * float to float
		 */
	case gc.TFLOAT32<<16 | gc.TFLOAT32:
		a = mips32.AMOVF

	case gc.TFLOAT64<<16 | gc.TFLOAT64:
		a = mips32.AMOVD

	case gc.TFLOAT32<<16 | gc.TFLOAT64:
		a = mips32.AMOVFD
		goto rdst

	case gc.TFLOAT64<<16 | gc.TFLOAT32:
		a = mips32.AMOVDF
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
func gins(as obj.As, f, t *gc.Node) *obj.Prog {
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
func gins3(as obj.As, f, r, t *gc.Node) *obj.Prog {
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
func rawgins(as obj.As, f *gc.Node, t *gc.Node) *obj.Prog {
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
	// Special cases
	case mips32.AMUL, mips32.AMULU:
		if p.From.Type == obj.TYPE_CONST {
			gc.Debug['h'] = 1
			gc.Fatalf("bad inst: %v", p)
		}
		if t != nil {
			pp := gc.Prog(mips32.AMOVW)
			pp.From.Type = obj.TYPE_REG
			pp.From.Reg = mips32.REG_LO
			pp.To = p.To

			p.Reg = p.To.Reg
			p.To = obj.Addr{}
		}

	case mips32.ASUBU:
		// unary
		if f == nil {
			p.From = p.To
			p.Reg = mips32.REGZERO
		}
	}

	if gc.Debug['g'] != 0 {
		fmt.Printf("%v\n", p)
	}

	return p
}

/*
 * return Axxx for Oxxx on type t.
 */
func optoas(op gc.Op, t *gc.Type) obj.As {
	if t == nil {
		gc.Fatalf("optoas: t is nil")
	}

	// avoid constant conversions in switches below
	const (
		OMINUS_ = uint32(gc.OMINUS) << 16
		OLSH_   = uint32(gc.OLSH) << 16
		ORSH_   = uint32(gc.ORSH) << 16
		OADD_   = uint32(gc.OADD) << 16
		OSUB_   = uint32(gc.OSUB) << 16
		OMUL_   = uint32(gc.OMUL) << 16
		ODIV_   = uint32(gc.ODIV) << 16
		OOR_    = uint32(gc.OOR) << 16
		OAND_   = uint32(gc.OAND) << 16
		OXOR_   = uint32(gc.OXOR) << 16
		OEQ_    = uint32(gc.OEQ) << 16
		ONE_    = uint32(gc.ONE) << 16
		OLT_    = uint32(gc.OLT) << 16
		OLE_    = uint32(gc.OLE) << 16
		OGE_    = uint32(gc.OGE) << 16
		OGT_    = uint32(gc.OGT) << 16
		OCMP_   = uint32(gc.OCMP) << 16
		OAS_    = uint32(gc.OAS) << 16
		OHMUL_  = uint32(gc.OHMUL) << 16
		OSQRT_  = uint32(gc.OSQRT) << 16
	)

	a := obj.AXXX
	switch uint32(op)<<16 | uint32(gc.Simtype[t.Etype]) {
	default:
		gc.Fatalf("optoas: no entry for op=%v type=%v", op, t)

	case OEQ_ | gc.TBOOL,
		OEQ_ | gc.TINT8,
		OEQ_ | gc.TUINT8,
		OEQ_ | gc.TINT16,
		OEQ_ | gc.TUINT16,
		OEQ_ | gc.TINT32,
		OEQ_ | gc.TUINT32,
		OEQ_ | gc.TINT64,
		OEQ_ | gc.TUINT64,
		OEQ_ | gc.TPTR32:
		a = mips32.ABEQ

	case OEQ_ | gc.TFLOAT32, // ACMPEQF
		OEQ_ | gc.TFLOAT64: // ACMPEQD
		a = mips32.ABFPT

	case ONE_ | gc.TBOOL,
		ONE_ | gc.TINT8,
		ONE_ | gc.TUINT8,
		ONE_ | gc.TINT16,
		ONE_ | gc.TUINT16,
		ONE_ | gc.TINT32,
		ONE_ | gc.TUINT32,
		ONE_ | gc.TINT64,
		ONE_ | gc.TUINT64,
		ONE_ | gc.TPTR32,
		ONE_ | gc.TPTR64:
		a = mips32.ABNE

	case ONE_ | gc.TFLOAT32, // ACMPEQF
		ONE_ | gc.TFLOAT64: // ACMPEQD
		a = mips32.ABFPF

	case OLT_ | gc.TINT8, // ASGT
		OLT_ | gc.TINT16,
		OLT_ | gc.TINT32,
		OLT_ | gc.TINT64,
		OLT_ | gc.TUINT8, // ASGTU
		OLT_ | gc.TUINT16,
		OLT_ | gc.TUINT32,
		OLT_ | gc.TUINT64:
		a = mips32.ABNE

	case OLT_ | gc.TFLOAT32, // ACMPGEF
		OLT_ | gc.TFLOAT64: // ACMPGED
		a = mips32.ABFPT

	case OLE_ | gc.TINT8, // ASGT
		OLE_ | gc.TINT16,
		OLE_ | gc.TINT32,
		OLE_ | gc.TINT64,
		OLE_ | gc.TUINT8, // ASGTU
		OLE_ | gc.TUINT16,
		OLE_ | gc.TUINT32,
		OLE_ | gc.TUINT64:
		a = mips32.ABEQ

	case OLE_ | gc.TFLOAT32, // ACMPGTF
		OLE_ | gc.TFLOAT64: // ACMPGTD
		a = mips32.ABFPT

	case OGT_ | gc.TINT8, // ASGT
		OGT_ | gc.TINT16,
		OGT_ | gc.TINT32,
		OGT_ | gc.TINT64,
		OGT_ | gc.TUINT8, // ASGTU
		OGT_ | gc.TUINT16,
		OGT_ | gc.TUINT32,
		OGT_ | gc.TUINT64:
		a = mips32.ABNE

	case OGT_ | gc.TFLOAT32, // ACMPGTF
		OGT_ | gc.TFLOAT64: // ACMPGTD
		a = mips32.ABFPT

	case OGE_ | gc.TINT8, // ASGT
		OGE_ | gc.TINT16,
		OGE_ | gc.TINT32,
		OGE_ | gc.TINT64,
		OGE_ | gc.TUINT8, // ASGTU
		OGE_ | gc.TUINT16,
		OGE_ | gc.TUINT32,
		OGE_ | gc.TUINT64:
		a = mips32.ABEQ

	case OGE_ | gc.TFLOAT32, // ACMPGEF
		OGE_ | gc.TFLOAT64: // ACMPGED
		a = mips32.ABFPT

	case OAS_ | gc.TBOOL,
		OAS_ | gc.TINT8:
		a = mips32.AMOVB

	case OAS_ | gc.TUINT8:
		a = mips32.AMOVBU

	case OAS_ | gc.TINT16:
		a = mips32.AMOVH

	case OAS_ | gc.TUINT16:
		a = mips32.AMOVHU

	case OAS_ | gc.TINT32:
		a = mips32.AMOVW

	case OAS_ | gc.TUINT32,
		OAS_ | gc.TPTR32:
		a = mips32.AMOVW //U

	case OAS_ | gc.TINT64,
		OAS_ | gc.TUINT64,
		OAS_ | gc.TPTR64:
		a = mips32.AMOVW

	case OAS_ | gc.TFLOAT32:
		a = mips32.AMOVF

	case OAS_ | gc.TFLOAT64:
		a = mips32.AMOVD

	case OADD_ | gc.TINT8,
		OADD_ | gc.TUINT8,
		OADD_ | gc.TINT16,
		OADD_ | gc.TUINT16,
		OADD_ | gc.TINT32,
		OADD_ | gc.TUINT32,
		OADD_ | gc.TPTR32:
		a = mips32.AADDU

	case OADD_ | gc.TFLOAT32:
		a = mips32.AADDF

	case OADD_ | gc.TFLOAT64:
		a = mips32.AADDD

	case OSUB_ | gc.TINT8,
		OSUB_ | gc.TUINT8,
		OSUB_ | gc.TINT16,
		OSUB_ | gc.TUINT16,
		OSUB_ | gc.TINT32,
		OSUB_ | gc.TUINT32,
		OSUB_ | gc.TPTR32:
		a = mips32.ASUBU

	case OSUB_ | gc.TFLOAT32:
		a = mips32.ASUBF

	case OSUB_ | gc.TFLOAT64:
		a = mips32.ASUBD

	case OMINUS_ | gc.TINT8,
		OMINUS_ | gc.TUINT8,
		OMINUS_ | gc.TINT16,
		OMINUS_ | gc.TUINT16,
		OMINUS_ | gc.TINT32,
		OMINUS_ | gc.TUINT32,
		OMINUS_ | gc.TPTR32:
		a = mips32.ASUBU

	case OAND_ | gc.TINT8,
		OAND_ | gc.TUINT8,
		OAND_ | gc.TINT16,
		OAND_ | gc.TUINT16,
		OAND_ | gc.TINT32,
		OAND_ | gc.TUINT32,
		OAND_ | gc.TPTR32,
		OAND_ | gc.TINT64,
		OAND_ | gc.TUINT64,
		OAND_ | gc.TPTR64:
		a = mips32.AAND

	case OOR_ | gc.TINT8,
		OOR_ | gc.TUINT8,
		OOR_ | gc.TINT16,
		OOR_ | gc.TUINT16,
		OOR_ | gc.TINT32,
		OOR_ | gc.TUINT32,
		OOR_ | gc.TPTR32,
		OOR_ | gc.TINT64,
		OOR_ | gc.TUINT64,
		OOR_ | gc.TPTR64:
		a = mips32.AOR

	case OXOR_ | gc.TINT8,
		OXOR_ | gc.TUINT8,
		OXOR_ | gc.TINT16,
		OXOR_ | gc.TUINT16,
		OXOR_ | gc.TINT32,
		OXOR_ | gc.TUINT32,
		OXOR_ | gc.TPTR32,
		OXOR_ | gc.TINT64,
		OXOR_ | gc.TUINT64,
		OXOR_ | gc.TPTR64:
		a = mips32.AXOR
	case OLSH_ | gc.TINT8,
		OLSH_ | gc.TUINT8,
		OLSH_ | gc.TINT16,
		OLSH_ | gc.TUINT16,
		OLSH_ | gc.TINT32,
		OLSH_ | gc.TUINT32,
		OLSH_ | gc.TPTR32:
		a = mips32.ASLL

	case ORSH_ | gc.TUINT8,
		ORSH_ | gc.TUINT16,
		ORSH_ | gc.TUINT32,
		ORSH_ | gc.TPTR32:
		a = mips32.ASRL

	case ORSH_ | gc.TINT8,
		ORSH_ | gc.TINT16,
		ORSH_ | gc.TINT32:
		a = mips32.ASRA

	case OMUL_ | gc.TINT8,
		OMUL_ | gc.TINT16,
		OMUL_ | gc.TINT32:
		a = mips32.AMUL

	case OMUL_ | gc.TUINT8,
		OMUL_ | gc.TUINT16,
		OMUL_ | gc.TUINT32,
		OMUL_ | gc.TPTR32:
		a = mips32.AMULU

	case OMUL_ | gc.TFLOAT32:
		a = mips32.AMULF

	case OMUL_ | gc.TFLOAT64:
		a = mips32.AMULD

	case ODIV_ | gc.TINT8,
		ODIV_ | gc.TINT16,
		ODIV_ | gc.TINT32:
		a = mips32.ADIV

	case ODIV_ | gc.TUINT8,
		ODIV_ | gc.TUINT16,
		ODIV_ | gc.TUINT32,
		ODIV_ | gc.TPTR32:
		a = mips32.ADIVU

	case ODIV_ | gc.TFLOAT32:
		a = mips32.ADIVF

	case ODIV_ | gc.TFLOAT64:
		a = mips32.ADIVD

	case OSQRT_ | gc.TFLOAT32:
		a = mips32.ASQRTF

	case OSQRT_ | gc.TFLOAT64:
		a = mips32.ASQRTD
	}

	return a
}

const (
	ODynam   = 1 << 0
	OAddable = 1 << 1
)

func xgen(n *gc.Node, a *gc.Node, o int) bool {
	return true
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
func sudoaddable(as obj.As, n *gc.Node, a *obj.Addr) bool {
	// TODO(minux)

	*a = obj.Addr{}
	return false
}
