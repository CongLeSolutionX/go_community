// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mips32

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
)

func cgen64(n *gc.Node, res *gc.Node) {

	if res.Op != gc.OINDREG && res.Op != gc.ONAME {
		gc.Dump("n", n)
		gc.Dump("res", res)
		gc.Fatalf("cgen64 %v of %v", n.Op, res.Op)
	}

	l := n.Left
	var t1 gc.Node
	if !l.Addable {
		gc.Tempname(&t1, l.Type)
		gc.Cgen(l, &t1)
		l = &t1
	}

	var hi1 gc.Node
	var lo1 gc.Node
	split64(l, &lo1, &hi1)
	switch n.Op {
	default:
		gc.Fatalf("mips32/cgen64 (1) %v", n.Op)
		return

	case gc.OMINUS:
		var lo2 gc.Node
		var hi2 gc.Node
		split64(res, &lo2, &hi2)

		var al gc.Node
		gc.Regalloc(&al, lo1.Type, nil)
		gins(mips32.AMOVW, &lo1, &al)

		var bit31 gc.Node
		gc.Regalloc(&bit31, lo1.Type, nil)

		gmove(ncon(0), &bit31)
		gins(mips32.ASUBU, &al, &bit31)
		gmove(&bit31, &lo2)

		var third gc.Node
		gc.Regalloc(&third, lo1.Type, nil)
		gmove(ncon(0), &third)
		gins(mips32.ASGTU, &bit31, &third)

		var ah gc.Node
		gc.Regalloc(&ah, hi1.Type, nil)
		gins(mips32.AMOVW, &hi1, &ah)

		var bit31_hi gc.Node
		gc.Regalloc(&bit31_hi, hi1.Type, nil)

		gmove(ncon(0), &bit31_hi)
		gins(mips32.ASUBU, &ah, &bit31_hi)

		gins(mips32.ASUBU, &third, &bit31_hi)

		gmove(&bit31_hi, &hi2)

		gc.Regfree(&al)
		gc.Regfree(&ah)
		gc.Regfree(&bit31)
		gc.Regfree(&bit31_hi)
		gc.Regfree(&third)
		splitclean()
		splitclean()
		return

	case gc.OCOM:
		var al gc.Node
		gc.Regalloc(&al, gc.Types[gc.TPTR32], nil)
		var ah gc.Node
		gc.Regalloc(&ah, gc.Types[gc.TPTR32], nil)
		gins(mips32.AMOVW, &hi1, &ah)
		gins(mips32.AMOVW, &lo1, &al)

		var tn gc.Node
		gc.Nodconst(&tn, gc.Types[gc.TINT32], int64(-1))
		gins(mips32.AXOR, &tn, &al)
		gins(mips32.AXOR, &tn, &ah)

		split64(res, &lo1, &hi1)
		gins(mips32.AMOVW, &al, &lo1)
		gins(mips32.AMOVW, &ah, &hi1)
		splitclean()

		gc.Regfree(&al)
		gc.Regfree(&ah)
		splitclean()
		return

	// binary operators.
	// common setup below.
	case gc.OADD,
		gc.OSUB,
		gc.OMUL,
		gc.OLSH,
		gc.ORSH,
		gc.OAND,
		gc.OOR,
		gc.OXOR,
		gc.OLROT:
		break
	}

	// setup for binary operators
	r := n.Right

	/*if n.Bounded && r.Op != gc.OLITERAL {

		fmt.Printf("!!!BOUNDED 64 op:%v l:%v r:%v \n",
			n.Op, gc.FmtSharp,
			gc.Nconv(n.Left, gc.FmtLong),
			gc.Nconv(r, gc.FmtLong))
	}*/

	if r != nil && !r.Addable {
		var t2 gc.Node
		gc.Tempname(&t2, r.Type)
		gc.Cgen(r, &t2)
		r = &t2
	}

	var hi2 gc.Node
	var lo2 gc.Node
	if gc.Is64(r.Type) {
		split64(r, &lo2, &hi2)
	}

	var al gc.Node
	gc.Regalloc(&al, lo1.Type, nil)
	var ah gc.Node
	gc.Regalloc(&ah, hi1.Type, nil)

	// Do op.  Leave result in ah:al.
	switch n.Op {
	default:
		gc.Fatalf("mips32/cgen64 (2): not implemented: %v\n", n.Op)
		return

	case gc.OMUL:
		var bl gc.Node
		gc.Regalloc(&bl, gc.Types[gc.TPTR32], nil)

		var bh gc.Node
		gc.Regalloc(&bh, gc.Types[gc.TPTR32], nil)
		gins(mips32.AMOVW, &hi1, &ah)
		gins(mips32.AMOVW, &lo1, &al)
		gins(mips32.AMOVW, &hi2, &bh)
		gins(mips32.AMOVW, &lo2, &bl)

		// TODO(mips32): generate mul instead of mult+mflo

		var tmp1 gc.Node
		gc.Regalloc(&tmp1, gc.Types[gc.TPTR32], nil)
		var tmp2 gc.Node
		gc.Regalloc(&tmp2, gc.Types[gc.TPTR32], nil)
		var reglo gc.Node
		gc.Nodreg(&reglo, gc.Types[gc.TINT], mips32.REG_LO)
		var reghi gc.Node
		gc.Nodreg(&reghi, gc.Types[gc.TINT], mips32.REG_HI)

		gins3(mips32.AMUL, &bl, &ah, nil)
		gins(mips32.AMOVW, &reglo, &tmp1)
		gins3(mips32.AMUL, &al, &bh, nil)
		gins(mips32.AMOVW, &reglo, &tmp2)
		gins(mips32.AADDU, &tmp1, &tmp2)
		gins3(mips32.AMULU, &al, &bl, nil)
		gins(mips32.AMOVW, &reglo, &al)
		gins(mips32.AMOVW, &reghi, &ah)
		gins(mips32.AADDU, &tmp2, &ah)

		gc.Regfree(&tmp1)
		gc.Regfree(&tmp2)
		gc.Regfree(&bl)
		gc.Regfree(&bh)

	case gc.OLSH:
		if r.Op == gc.OLITERAL {
			v := uint64(r.Int64())
			if v < 32 {

				var n1 gc.Node
				gc.Regalloc(&n1, lo1.Type, nil)

				gmove(&hi1, &ah)
				gmove(&lo1, &al)

				if v != 0 {
					gins(mips32.ASLL, ncon(uint32(v)), &ah)
					gmove(&al, &n1)
					gins(mips32.ASLL, ncon(uint32(v)), &al)
					gins(mips32.ASRL, ncon(32-uint32(v)), &n1)
					gins(mips32.AOR, &n1, &ah)
				}

				gc.Regfree(&n1)

			} else if v < 64 {

				gins(mips32.AMOVW, ncon(0), &al)
				gmove(&lo1, &ah)

				if v != 32 {
					gins(mips32.ASLL, ncon(uint32(v)-32), &ah)
				}

			} else {
				gins(mips32.AMOVW, ncon(0), &al) // res_lo = 0
				gins(mips32.AMOVW, ncon(0), &ah) // res_hi = 0

			}
		} else {

			// srl $2, $lo,1
			// nor $tmp,$0,$r
			// sll  $hi,$hi,$r
			// srl $tmp,$2, $tmp
			// sll $2,$lo,$r
			// or $hi,$tmp,$hi
			// sltiu $tmp,$r, 64 //
			// sltiu $r,$r, 32   //
			// move $lo,$2
			// movz $hi,$2,$r
			// movz $lo,$0,$r
			// movz $hi,$0,$tmp //

			var n1 gc.Node
			var n2 gc.Node
			var nr gc.Node
			var nz gc.Node

			gc.Regalloc(&n1, gc.Types[gc.TPTR32], nil)
			gc.Regalloc(&n2, gc.Types[gc.TPTR32], nil)
			gc.Regalloc(&nr, gc.Types[gc.TPTR32], nil)
			gc.Nodreg(&nz, gc.Types[gc.TINT], mips32.REGZERO)

			if gc.Is64(r.Type) {
				gmove(&lo2, &nr)
			} else {
				gmove(r, &nr)
			}

			gmove(&lo1, &al)
			gmove(&hi1, &ah)

			gins3(mips32.ASRL, ncon(1), &al, &n1)
			gins3(mips32.ANOR, &nr, &nz, &n2)
			gins(mips32.ASLL, &nr, &ah)
			gins3(mips32.ASRL, &n2, &n1, &n2)
			gins3(mips32.ASLL, &nr, &al, &n1)
			gins(mips32.AOR, &n2, &ah)

			// TODO(mips32): maybe remove  checks if shift is bounded
			gins3(mips32.ASGTU, ncon(64), &nr, &n2)
			gins(mips32.ASGTU, ncon(32), &nr)

			gins(mips32.AMOVW, &n1, &al)
			gins3(mips32.AMOVZ, &nr, &n1, &ah)
			gins3(mips32.AMOVZ, &nr, &nz, &al)

			if gc.Is64(r.Type) {
				gmove(&hi2, &nr)
			}

			gins3(mips32.AMOVZ, &n2, &nz, &ah)

			if gc.Is64(r.Type) {
				gins3(mips32.AMOVN, &nr, &nz, &al)
				gins3(mips32.AMOVN, &nr, &nz, &ah)
			}

			gc.Regfree(&n1)
			gc.Regfree(&n2)
			gc.Regfree(&nr)
		}

	case gc.ORSH:
		if r.Op == gc.OLITERAL {
			v := uint64(r.Int64())
			if v < 32 {

				// sll     $2,$5,15
				// srl     $4,$4,17
				// sra     $3,$5,17
				// j       $31
				// or      $2,$2,$4

				var n1 gc.Node
				gc.Regalloc(&n1, hi1.Type, nil)

				gmove(&lo1, &al)
				gmove(&hi1, &ah)
				if v != 0 {
					gins(mips32.ASRL, ncon(uint32(v)), &al)
					gmove(&ah, &n1)
					gins(optoas(gc.ORSH, ah.Type), ncon(uint32(v)), &ah)
					gins(mips32.ASLL, ncon(32-uint32(v)), &n1)
					gins(mips32.AOR, &n1, &al)
				}

				gc.Regfree(&n1)

			} else if v < 64 {
				gmove(&hi1, &al)

				if hi1.Type.IsSigned() {
					gins3(mips32.ASRA, ncon(31), &al, &ah)
				} else {
					gmove(ncon(0), &ah)
				}

				if v != 32 {
					gins(optoas(gc.ORSH, ah.Type), ncon(uint32(v)-32), &al)
				}

			} else {

				if !hi1.Type.IsSigned() {
					gmove(ncon(0), &al)
					gmove(ncon(0), &ah)
				} else {
					gmove(&hi1, &ah)
					gins(mips32.ASRA, ncon(31), &ah)
					gmove(&ah, &al)
				}
			}
		} else {

			// sll     $2,$hi,1
			// nor     $tmp,$0,$r
			// srl     $lo,$lo,$r
			// sll     $2,$2,$tmp
			// sra     $tmp,$hi,$r
			// sra     $hi,$hi,31
			// andi    $r,$r,0x20
			// or      $lo,$2,$lo
			// movz    $hi,$tmp,$r
			// movn    $lo,$tmp,$r

			var n1 gc.Node
			var n2 gc.Node
			var nr gc.Node
			var nz gc.Node
			var ntmp gc.Node

			gc.Regalloc(&n1, gc.Types[gc.TPTR32], nil)
			gc.Regalloc(&n2, gc.Types[gc.TPTR32], nil)
			gc.Regalloc(&nr, gc.Types[gc.TPTR32], nil)
			gc.Nodreg(&nz, gc.Types[gc.TPTR32], mips32.REGZERO)
			gc.Nodreg(&ntmp, gc.Types[gc.TPTR32], mips32.REGTMP)

			gmove(&hi1, &ah)
			gmove(&lo1, &al)

			if gc.Is64(r.Type) {
				gmove(&lo2, &nr)
			} else {
				gmove(r, &nr)
			}

			if hi1.Type.IsSigned() {

				if gc.Is64(r.Type) {
					gmove(&hi2, &n1)
					gmove(ncon(63), &n2)
					gins3(mips32.AMOVN, &n1, &n2, &nr)
				}

				gins3(mips32.ASLL, ncon(1), &ah, &n1)
				gins3(mips32.ANOR, &nr, &nz, &n2)
				gins(mips32.ASRL, &nr, &al)
				gins(mips32.ASLL, &n2, &n1)
				gins3(mips32.ASRA, &nr, &ah, &n2)
				gins(mips32.ASRA, ncon(31), &ah)
				gins(mips32.AOR, &n1, &al)
				gins3(mips32.ASGTU, ncon(32), &nr, &ntmp)
				gins3(mips32.ASGTU, ncon(64), &nr, &n1)
				gins3(mips32.AMOVZ, &ntmp, &n2, &al)
				gins3(mips32.AMOVN, &ntmp, &n2, &ah)
				gins3(mips32.AMOVZ, &n1, &ah, &al)
			} else {

				// sll     $2,$5,1
				// nor     $3,$0,$6
				// srl     $4,$4,$6

				// sll     $3,$2,$3
				// srl     $2,$5,$6

				// andi    $6,$6,0x20

				// or      $4,$3,$4

				// move    $5,$2
				// movn    $4,$2,$6
				// movn    $5,$0,$6

				gins3(mips32.ASLL, ncon(1), &ah, &n1)
				gins3(mips32.ANOR, &nr, &nz, &n2)
				gins(mips32.ASRL, &nr, &al)

				gins3(mips32.ASLL, &n2, &n1, &n2)

				gins3(mips32.ASRL, &nr, &ah, &n1)

				gins(mips32.AOR, &n2, &al)
				gins3(mips32.ASGTU, ncon(64), &nr, &n2)
				gins(mips32.ASGTU, ncon(32), &nr)

				gins(mips32.AMOVW, &n1, &ah)

				gins3(mips32.AMOVZ, &nr, &n1, &al)
				gins3(mips32.AMOVZ, &nr, &nz, &ah)

				if gc.Is64(r.Type) {
					gmove(&hi2, &nr)
				}

				gins3(mips32.AMOVZ, &n2, &nz, &al)

				if gc.Is64(r.Type) {
					gins3(mips32.AMOVN, &nr, &nz, &ah)
					gins3(mips32.AMOVN, &nr, &nz, &al)
				}

			}

			gc.Regfree(&n1)
			gc.Regfree(&n2)
			gc.Regfree(&nr)
		}

		// TODO: Constants.
	case gc.OSUB:
		var bl gc.Node
		gc.Regalloc(&bl, gc.Types[gc.TPTR32], nil)

		var bh gc.Node
		gc.Regalloc(&bh, gc.Types[gc.TPTR32], nil)

		var tn gc.Node
		gc.Regalloc(&tn, gc.Types[gc.TPTR32], nil)

		gins(mips32.AMOVW, &lo1, &al)
		gins(mips32.AMOVW, &hi1, &ah)
		gins(mips32.AMOVW, &lo2, &bl)
		gins(mips32.AMOVW, &hi2, &bh)

		gins(mips32.ASUBU, &bl, &al)

		gins(mips32.AMOVW, &lo1, &tn)
		gins(mips32.ASGTU, &al, &tn)

		gins(mips32.ASUBU, &bh, &ah)

		gins(mips32.ASUBU, &tn, &ah)

		gc.Regfree(&bl)
		gc.Regfree(&bh)
		gc.Regfree(&tn)

		// TODO: Constants
	case gc.OADD:
		var bl gc.Node
		gc.Regalloc(&bl, gc.Types[gc.TPTR32], nil)

		var bh gc.Node
		gc.Regalloc(&bh, gc.Types[gc.TPTR32], nil)
		gins(mips32.AMOVW, &hi1, &ah)
		gins(mips32.AMOVW, &lo1, &al)
		gins(mips32.AMOVW, &hi2, &bh)
		gins(mips32.AMOVW, &lo2, &bl)

		gins(mips32.AADDU, &al, &bl)       // bl = al +bl
		gins3(mips32.ASGTU, &al, &bl, &al) // al = bl < al
		gins(mips32.AADDU, &bh, &ah)       // ah = ah +bh
		gins(mips32.AADDU, &al, &ah)       // ah = al +ah
		gins(mips32.AMOVW, &bl, &al)       // al = bl

		gc.Regfree(&bl)
		gc.Regfree(&bh)

	case gc.OOR,
		gc.OAND,
		gc.OXOR:
		var bl gc.Node
		gc.Regalloc(&bl, gc.Types[gc.TPTR32], nil)

		var bh gc.Node
		gc.Regalloc(&bh, gc.Types[gc.TPTR32], nil)
		gins(mips32.AMOVW, &hi1, &ah)
		gins(mips32.AMOVW, &lo1, &al)

		gins(mips32.AMOVW, &hi2, &bh)
		gins(mips32.AMOVW, &lo2, &bl)

		gins(optoas(n.Op, lo1.Type), &bl, &al)
		gins(optoas(n.Op, hi1.Type), &bh, &ah)
		gc.Regfree(&bl)
		gc.Regfree(&bh)

		// TODO: Constants.
	}

	// .....

	if gc.Is64(r.Type) {
		splitclean()
	}
	splitclean()

	split64(res, &lo1, &hi1)
	gins(mips32.AMOVW, &al, &lo1)
	gins(mips32.AMOVW, &ah, &hi1)
	splitclean()

	// .....

	gc.Regfree(&al)
	gc.Regfree(&ah)
}

/*
 * generate comparison of nl, nr, both 64-bit.
 * nl is memory; nr is constant or memory.
 */
func cmp64(nl *gc.Node, nr *gc.Node, op gc.Op, likely int, to *obj.Prog) {
	var lo1 gc.Node
	var hi1 gc.Node
	var lo2 gc.Node
	var hi2 gc.Node
	var r1 gc.Node
	var r2 gc.Node

	split64(nl, &lo1, &hi1)
	split64(nr, &lo2, &hi2)

	// compare most significant word;
	// if they differ, we're done.
	t := hi1.Type

	/////
	tr := hi2.Type
	sgt := mips32.ASGTU
	if t.IsSigned() && tr.IsSigned() {
		sgt = mips32.ASGT
	}
	////

	gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
	gc.Regalloc(&r2, gc.Types[gc.TINT32], nil)
	gins(mips32.AMOVW, &hi1, &r1)
	gins(mips32.AMOVW, &hi2, &r2)
	gc.Regfree(&r1)
	gc.Regfree(&r2)

	var br *obj.Prog
	switch op {
	default:
		gc.Fatalf("cmp64 %v %v", op, t)

	case gc.OEQ:

		br = ginsbranch(mips32.ABNE, nil, &r1, &r2, 0)

	case gc.ONE:

		gc.Patch(ginsbranch(mips32.ABNE, nil, &r1, &r2, 0), to)

	case gc.OGE:
		var ntmp gc.Node
		gc.Nodreg(&ntmp, gc.Types[gc.TINT], mips32.REGTMP)
		gins3(sgt, &r2, &r1, &ntmp)
		br = ginsbranch(mips32.ABNE, nil, &ntmp, nil, likely) // jmp false

		gc.Patch(ginsbranch(mips32.ABNE, nil, &r1, &r2, likely), to) // jmp true

		t = lo1.Type

		gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
		gc.Regalloc(&r2, gc.Types[gc.TINT32], nil)
		gins(mips32.AMOVW, &lo1, &r1)
		gins(mips32.AMOVW, &lo2, &r2)
		gc.Regfree(&r1)
		gc.Regfree(&r2)

		var ntmp2 gc.Node
		gc.Nodreg(&ntmp2, gc.Types[gc.TINT], mips32.REGTMP)
		gins3(mips32.ASGTU, &r2, &r1, &ntmp2) // mips32.ASGT

		br2 := ginsbranch(mips32.ABNE, nil, &ntmp2, nil, likely) // jmp false

		// TODO(mips32): simplify
		var rz gc.Node
		gc.Nodreg(&rz, gc.Types[gc.TINT], mips32.REGZERO)
		gc.Patch(ginsbranch(mips32.ABEQ, nil, &rz, nil, likely), to) // unconditionally, jmp true

		// point first branch down here if appropriate
		if br != nil {
			gc.Patch(br, gc.Pc)
		}
		if br2 != nil {
			gc.Patch(br2, gc.Pc)
		}
		splitclean()
		splitclean()
		return

	case
		gc.OGT:
		var ntmp gc.Node
		gc.Nodreg(&ntmp, gc.Types[gc.TINT], mips32.REGTMP)
		gins3(sgt, &r1, &r2, &ntmp)

		gc.Patch(ginsbranch(optoas(gc.OLT, t), nil, &ntmp, nil, likely), to) // jmp true

		br = ginsbranch(mips32.ABNE, nil, &r1, &r2, likely) // jmp false

		// cmp lo

		// compare least significant word
		t = lo1.Type

		gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
		gc.Regalloc(&r2, gc.Types[gc.TINT32], nil)
		gins(mips32.AMOVW, &lo1, &r1)
		gins(mips32.AMOVW, &lo2, &r2)
		gc.Regfree(&r1)
		gc.Regfree(&r2)

		var ntmp2 gc.Node
		gc.Nodreg(&ntmp2, gc.Types[gc.TINT], mips32.REGTMP)
		gins3(mips32.ASGTU, &r1, &r2, &ntmp2) // mips32.ASGT

		bre := ginsbranch(mips32.ABEQ, nil, &ntmp2, nil, likely) // jmp false

		var rz gc.Node
		gc.Nodreg(&rz, gc.Types[gc.TINT], mips32.REGZERO)
		gc.Patch(ginsbranch(mips32.ABEQ, nil, &rz, nil, likely), to) // unconditionally, jmp true

		// point first branch down here if appropriate
		if br != nil {
			gc.Patch(br, gc.Pc)
		}
		if bre != nil {
			gc.Patch(bre, gc.Pc)
		}

		splitclean()
		splitclean()
		return

	case gc.OLE,
		gc.OLT:

		var ntmp gc.Node
		gc.Nodreg(&ntmp, gc.Types[gc.TINT], mips32.REGTMP)
		gins3(sgt, &r1, &r2, &ntmp) // mips32.ASGT

		br = ginsbranch(optoas(gc.OGT, t), nil, &ntmp, nil, likely) // jmp true

		gc.Patch(ginsbranch(mips32.ABNE, nil, &r1, &r2, likely), to) // jmp false

		// compare least significant word
		t = lo1.Type

		gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
		gc.Regalloc(&r2, gc.Types[gc.TINT32], nil)
		gins(mips32.AMOVW, &lo1, &r1)
		gins(mips32.AMOVW, &lo2, &r2)

		gc.Regfree(&r1)
		gc.Regfree(&r2)

		var ntmp2 gc.Node
		gc.Nodreg(&ntmp2, gc.Types[gc.TINT], mips32.REGTMP)

		if op == gc.OLT {
			gins3(mips32.ASGTU, &r2, &r1, &ntmp2)
			gc.Patch(ginsbranch(optoas(op, t), nil, &ntmp2, nil, likely), to) // jmp false // true
		} else { // gc.OLE
			gins3(mips32.ASGTU, &r1, &r2, &ntmp2)
			br2 := ginsbranch(mips32.ABNE, nil, &ntmp2, nil, likely) // jmp false

			var rz gc.Node
			gc.Nodreg(&rz, gc.Types[gc.TINT], mips32.REGZERO)
			gc.Patch(ginsbranch(mips32.ABEQ, nil, &rz, nil, likely), to) // unconditionally, jmp true

			if br != nil {
				gc.Patch(br2, gc.Pc)
			}
		}

		// point first branch down here if appropriate
		if br != nil {
			gc.Patch(br, gc.Pc)
		}

		splitclean()
		splitclean()
		return
	}

	// compare least significant word
	t = lo1.Type

	gc.Regalloc(&r1, gc.Types[gc.TINT32], nil)
	gc.Regalloc(&r2, gc.Types[gc.TINT32], nil)
	gins(mips32.AMOVW, &lo1, &r1)
	gins(mips32.AMOVW, &lo2, &r2)
	gc.Regfree(&r1)
	gc.Regfree(&r2)

	// jump again
	gc.Patch(ginsbranch(optoas(op, t), nil, &r1, &r2, 0), to)

	// point first branch down here if appropriate
	if br != nil {
		gc.Patch(br, gc.Pc)
	}

	splitclean()
	splitclean()
}
