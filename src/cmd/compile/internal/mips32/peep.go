// Derived from Inferno utils/6c/peep.c
// http://code.google.com/p/inferno-os/source/browse/utils/6c/peep.c
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
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
	"fmt"
)

var gactive uint32

func peep(firstp *obj.Prog) {

	g := gc.Flowstart(firstp, nil)
	if g == nil {
		return
	}
	gactive = 0

	var p *obj.Prog
	var r *gc.Flow
	var t int
loop1:
	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		gc.Dumpit("loop1", g.Start, 0)
	}

	// 	println(g.Start.Prog.From.Sym.Name)

	t = 0
	for r = g.Start; r != nil; r = r.Link {
		p = r.Prog
		if p.As == mips32.AMOVW || p.As == mips32.AMOVF || p.As == mips32.AMOVD {
			if regtyp(&p.To) {
				// Try to eliminate reg->reg moves
				if regtyp(&p.From) {
					if isfreg(&p.From) == isfreg(&p.To) {
						if copyprop(r) {
							excise(r)
							t++
						} else if subprop(r) && copyprop(r) {
							excise(r)
							t++
						}
					}
				}

				// Convert uses to $0 to uses of R0 and
				// propagate R0
				if regzer(&p.From) {
					if p.To.Type == obj.TYPE_REG && !isfreg(&p.To) {
						p.From.Type = obj.TYPE_REG
						p.From.Reg = mips32.REGZERO
						if copyprop(r) {
							excise(r)
							t++
						} else if subprop(r) && copyprop(r) {
							excise(r)
							t++
						}
					}
				}
			}
		} else if (p.As == mips32.AMOVBU || p.As == mips32.AMOVB || p.As == mips32.AMOVHU || p.As == mips32.AMOVH) && regtyp(&p.From) && regtyp(&p.To) {

			if smallconst(&p.From) {
				r.Prog.As = mips32.AMOVW
				t++
			} else {

				h := findpre(r, &p.From)

				if h != nil && (h.Prog.As == p.As || h.Prog.As == mips32.AMOVW && smallconst(&h.Prog.From)) {
					r.Prog.As = mips32.AMOVW
					t++
				}
			}
		}
	}

	if t != 0 {
		goto loop1
	}

	/*
	 * Conservative match of
	 * MOVB R1,R2;
	 * MOVB R2,$mem
	 * no use of R2
	 * --------------
	 * MOVB R1,$mem
	 */
	var p1 *obj.Prog
	var r1 *gc.Flow
	for r := g.Start; r != nil; r = r.Link {

		if gc.Uniqp(r) == nil {
			continue
		}

		p = r.Prog
		switch p.As {
		default:
			continue

		case mips32.AMOVH,
			mips32.AMOVHU,
			mips32.AMOVB,
			mips32.AMOVBU:
		}

		if p.To.Type != obj.TYPE_REG || p.From.Type != obj.TYPE_REG {
			continue
		}

		r1 = r.Link

		if r1 == nil || gc.Uniqp(r1) != r {
			continue
		}

		p1 = r1.Prog

		if p1.As != p.As {
			continue
		}
		if p1.From.Type != obj.TYPE_REG || p1.From.Reg != p.To.Reg {
			continue
		}
		if p1.To.Type != obj.TYPE_MEM {
			continue
		}

		if !canexcize(r1, &p.To) { // find if there are reads of To reg after r1
			continue
		}

		p1.From.Reg = p.From.Reg

		excise(r)

	}

	gc.Flowend(g)
}

func smallconst(a *obj.Addr) bool {
	return regzer(a) || a.Type == obj.TYPE_CONST && a.Offset >= 0 && a.Offset < 128
}

func canexcize(r *gc.Flow, v *obj.Addr) bool {

	for r1 := gc.Uniqs(r); r1 != nil; r1 = gc.Uniqs(r1) {
		if gc.Uniqp(r1) == nil {
			break
		}
		switch copyu(r1.Prog, v, nil) {
		default:
			break
		case 0:
			continue
		case 3:
			return true
		}

		break
	}
	return false
}

func findpre(r *gc.Flow, v *obj.Addr) *gc.Flow {
	var r1 *gc.Flow

	for r1 = gc.Uniqp(r); r1 != nil; r, r1 = r1, gc.Uniqp(r1) {
		if gc.Uniqs(r1) != r {
			return nil
		}
		switch copyu(r1.Prog, v, nil) {
		case 1, /* used */
			2: /* read-alter-rewrite */
			return nil

		case 3, /* set */
			4: /* set and used */
			return r1
		}
	}

	return nil
}

func excise(r *gc.Flow) {
	p := r.Prog
	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		fmt.Printf("%v ===delete===\n", p)
	}
	obj.Nopout(p)
	gc.Ostats.Ndelmov++
}

// regzer returns true if a's value is 0 (a is R0 or $0)
func regzer(a *obj.Addr) bool {
	if a.Type == obj.TYPE_CONST || a.Type == obj.TYPE_ADDR {
		if a.Sym == nil && a.Reg == 0 {
			if a.Offset == 0 {
				return true
			}
		}
	}
	return a.Type == obj.TYPE_REG && a.Reg == mips32.REGZERO
}

func regtyp(a *obj.Addr) bool {
	return a.Type == obj.TYPE_REG && mips32.REG_R0 <= a.Reg && a.Reg <= mips32.REG_F31 && a.Reg != mips32.REGZERO
}

func isfreg(a *obj.Addr) bool {
	return mips32.REG_F0 <= a.Reg && a.Reg <= mips32.REG_F31
}

/*
 * the idea is to substitute
 * one register for another
 * from one MOV to another
 *	MOV	a, R1
 *	ADD	b, R1	/ no use of R2
 *	MOV	R1, R2
 * would be converted to
 *	MOV	a, R2
 *	ADD	b, R2
 *	MOV	R2, R1
 * hopefully, then the former or latter MOV
 * will be eliminated by copy propagation.
 *
 * r0 (the argument, not the register) is the MOV at the end of the
 * above sequences.  This returns 1 if it modified any instructions.
 */
func subprop(r0 *gc.Flow) bool {

	p := r0.Prog
	v1 := &p.From
	if !regtyp(v1) {
		return false
	}
	v2 := &p.To
	if !regtyp(v2) {
		return false
	}
	for r := gc.Uniqp(r0); r != nil; r = gc.Uniqp(r) {
		if gc.Uniqs(r) == nil {
			break
		}
		p = r.Prog
		if p.As == obj.AVARDEF || p.As == obj.AVARKILL {
			continue
		}
		if p.Info.Flags&gc.Call != 0 {
			return false
		}

		if p.Info.Flags&(gc.RightRead|gc.RightWrite) == gc.RightWrite {
			if p.To.Type == v1.Type {
				if p.To.Reg == v1.Reg {
					copysub(&p.To, v1, v2, true)
					if gc.Debug['P'] != 0 {
						fmt.Printf("gotit: %v->%v\n%v", gc.Ctxt.Dconv(v1), gc.Ctxt.Dconv(v2), r.Prog)
						if p.From.Type == v2.Type {
							fmt.Printf(" excise")
						}
						fmt.Printf("\n")
					}

					for r = gc.Uniqs(r); r != r0; r = gc.Uniqs(r) {
						p = r.Prog
						copysub(&p.From, v1, v2, true)
						copysub1(p, v1, v2, true)
						copysub(&p.To, v1, v2, true)
						if gc.Debug['P'] != 0 {
							fmt.Printf("%v\n", r.Prog)
						}
					}

					v1.Reg, v2.Reg = v2.Reg, v1.Reg
					if gc.Debug['P'] != 0 {
						fmt.Printf("%v last\n", r.Prog)
					}
					return true
				}
			}
		}

		if copyau(&p.From, v2) || copyau1(p, v2) || copyau(&p.To, v2) {
			break
		}
		if copysub(&p.From, v1, v2, false) || copysub1(p, v1, v2, false) || copysub(&p.To, v1, v2, false) {
			break
		}
	}

	return false
}

/*
 * The idea is to remove redundant copies.
 *	v1->v2	F=0
 *	(use v2	s/v2/v1/)*
 *	set v1	F=1
 *	use v2	return fail (v1->v2 move must remain)
 *	-----------------
 *	v1->v2	F=0
 *	(use v2	s/v2/v1/)*
 *	set v1	F=1
 *	set v2	return success (caller can remove v1->v2 move)
 */
func copyprop(r0 *gc.Flow) bool {

	p := r0.Prog
	v1 := &p.From
	v2 := &p.To
	if copyas(v1, v2) {
		if gc.Debug['P'] != 0 {
			fmt.Printf("eliminating self-move: %v\n", r0.Prog)
		}
		return true
	}

	gactive++
	if gc.Debug['P'] != 0 {
		fmt.Printf("trying to eliminate %v->%v move from:\n%v\n", gc.Ctxt.Dconv(v1), gc.Ctxt.Dconv(v2), r0.Prog)
	}
	return copy1(v1, v2, r0.S1, false)
}

// copy1 replaces uses of v2 with v1 starting at r and returns true if
// all uses were rewritten.
func copy1(v1 *obj.Addr, v2 *obj.Addr, r *gc.Flow, f bool) bool {
	if uint32(r.Active) == gactive {
		if gc.Debug['P'] != 0 {
			fmt.Printf("act set; return 1\n")
		}
		return true
	}

	r.Active = int32(gactive)
	if gc.Debug['P'] != 0 {
		fmt.Printf("copy1 replace %v with %v f=%v\n", gc.Ctxt.Dconv(v2), gc.Ctxt.Dconv(v1), f)
	}
	for ; r != nil; r = r.S1 {
		p := r.Prog
		if gc.Debug['P'] != 0 {
			fmt.Printf("%v", p)
		}
		if !f && gc.Uniqp(r) == nil {
			// Multiple predecessors; conservatively
			// assume v1 was set on other path
			f = true

			if gc.Debug['P'] != 0 {
				fmt.Printf("; merge; f=%v", f)
			}
		}

		switch t := copyu(p, v2, nil); t {
		case 2: /* rar, can't split */
			if gc.Debug['P'] != 0 {
				fmt.Printf("; %v rar; return 0\n", gc.Ctxt.Dconv(v2))
			}
			return false

		case 3: /* set */
			if gc.Debug['P'] != 0 {
				fmt.Printf("; %v set; return 1\n", gc.Ctxt.Dconv(v2))
			}
			return true

		case 1, /* used, substitute */
			4: /* use and set */
			if f {
				if gc.Debug['P'] == 0 {
					return false
				}
				if t == 4 {
					fmt.Printf("; %v used+set and f=%v; return 0\n", gc.Ctxt.Dconv(v2), f)
				} else {
					fmt.Printf("; %v used and f=%v; return 0\n", gc.Ctxt.Dconv(v2), f)
				}
				return false
			}

			if copyu(p, v2, v1) != 0 {
				if gc.Debug['P'] != 0 {
					fmt.Printf("; sub fail; return 0\n")
				}
				return false
			}

			if gc.Debug['P'] != 0 {
				fmt.Printf("; sub %v->%v\n => %v", gc.Ctxt.Dconv(v2), gc.Ctxt.Dconv(v1), p)
			}
			if t == 4 {
				if gc.Debug['P'] != 0 {
					fmt.Printf("; %v used+set; return 1\n", gc.Ctxt.Dconv(v2))
				}
				return true
			}
		}

		if !f {
			t := copyu(p, v1, nil)
			if t == 2 || t == 3 || t == 4 {
				f = true
				if gc.Debug['P'] != 0 {
					fmt.Printf("; %v set and !f; f=%v", gc.Ctxt.Dconv(v1), f)
				}
			}
		}

		if gc.Debug['P'] != 0 {
			fmt.Printf("\n")
		}
		if r.S2 != nil {
			if !copy1(v1, v2, r.S2, f) {
				return false
			}
		}
	}

	return true
}

// If s==nil, copyu returns the set/use of v in p; otherwise, it
// modifies p to replace reads of v with reads of s and returns 0 for
// success or non-zero for failure.
//
// If s==nil, copy returns one of the following values:
// 	1 if v only used
//	2 if v is set and used in one address (read-alter-rewrite;
// 	  can't substitute)
//	3 if v is only set
//	4 if v is set in one address and used in another (so addresses
// 	  can be rewritten independently)
//	0 otherwise (not touched)
func copyu(p *obj.Prog, v *obj.Addr, s *obj.Addr) int {
	if p.From3Type() != obj.TYPE_NONE {
		// never generates a from3
		fmt.Printf("copyu: from3 (%v) not implemented\n", gc.Ctxt.Dconv(p.From3))
	}

	switch p.As {
	default:
		fmt.Printf("copyu: can't find %v . %v %v %v\n", p, gc.Ctxt.Dconv(&p.From), p.Reg, gc.Ctxt.Dconv(&p.To))
		return 2

	case obj.ANOP, /* read p->from, write p->to */
		mips32.AMOVF,
		mips32.AMOVD,
		mips32.AMOVH,
		mips32.AMOVHU,
		mips32.AMOVB,
		mips32.AMOVBU,
		mips32.AMOVW,
		mips32.AMOVFD,
		mips32.AMOVDF,
		mips32.AMOVDW,
		mips32.AMOVWD,
		mips32.AMOVFW,
		mips32.AMOVWF,
		mips32.ATRUNCFW,
		mips32.ATRUNCDW,
		mips32.ASQRTF,
		mips32.ASQRTD:
		if s != nil {
			if copysub(&p.From, v, s, true) {
				return 1
			}

			// Update only indirect uses of v in p->to
			if !copyas(&p.To, v) {
				if copysub(&p.To, v, s, true) {
					return 1
				}
			}
			return 0
		}

		if copyas(&p.To, v) {
			// Fix up implicit from
			if p.From.Type == obj.TYPE_NONE {
				p.From = p.To
			}
			if copyau(&p.From, v) {
				return 4
			}
			return 3
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau(&p.To, v) {
			// p->to only indirectly uses v
			return 1
		}

		return 0
	case mips32.ATEQ,
		mips32.ATNE: //read p->from, read p->to

		if s != nil {
			if copysub1(p, v, s, true) {
				return 1
			}
			if copysub(&p.To, v, s, true) {
				return 1
			}
			return 0
		}

		if copyau1(p, v) {
			return 1
		}
		if copyau(&p.To, v) {
			return 1
		}
		return 0

	case mips32.AMOVN, /* read p->from, read p->reg, rmw p->to */
		mips32.AMOVZ:

		if s != nil {
			if copysub(&p.From, v, s, true) {
				return 1
			}
			if copysub1(p, v, s, true) {
				return 1
			}

			return 0
		}

		if copyau(&p.From, v) {
			return 1
		}

		if copyau1(p, v) {
			return 1
		}

		return 2

	case mips32.ASGT, /* read p->from, read p->reg, write p->to */
		mips32.ASGTU,

		mips32.AADD,
		mips32.AADDU,
		mips32.ASUB,
		mips32.ASUBU,
		mips32.ASLL,
		mips32.ASRL,
		mips32.ASRA,
		mips32.AOR,
		mips32.ANOR,
		mips32.AAND,
		mips32.AXOR,

		mips32.AADDF,
		mips32.AADDD,
		mips32.ASUBF,
		mips32.ASUBD,
		mips32.AMULF,
		mips32.AMULD,
		mips32.ADIVF,
		mips32.ADIVD,
		mips32.AMUL,
		mips32.AMULU:
		if s != nil {
			if copysub(&p.From, v, s, true) {
				return 1
			}
			if copysub1(p, v, s, true) {
				return 1
			}

			// Update only indirect uses of v in p->to
			if p.To.Reg != 0 && !copyas(&p.To, v) {
				if copysub(&p.To, v, s, true) {
					return 1
				}
			}
			return 0
		}

		if p.To.Reg != 0 && copyas(&p.To, v) {
			if p.Reg == 0 {
				// Fix up implicit reg (e.g., ADD
				// R3,R4 -> ADD R3,R4,R4) so we can
				// update reg and to separately.
				p.Reg = p.To.Reg
			}

			if copyau(&p.From, v) {
				return 4
			}
			if copyau1(p, v) {
				return 4
			}
			return 3
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau1(p, v) {
			return 1
		}
		if p.To.Reg != 0 && copyau(&p.To, v) {
			return 1
		}
		return 0

	case obj.ACHECKNIL, /* read p->from */
		mips32.ABEQ, /* read p->from, read p->reg */
		mips32.ABNE,
		mips32.ABGTZ,
		mips32.ABGEZ,
		mips32.ABLTZ,
		mips32.ABLEZ,

		mips32.ACMPEQD,
		mips32.ACMPEQF,
		mips32.ACMPGED,
		mips32.ACMPGEF,
		mips32.ACMPGTD,
		mips32.ACMPGTF,
		mips32.ABFPF,
		mips32.ABFPT,
		mips32.ADIV,
		mips32.ADIVU:
		if s != nil {
			if copysub(&p.From, v, s, true) {
				return 1
			}
			if copysub1(p, v, s, true) {
				return 1
			}
			return 0
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau1(p, v) {
			return 1
		}
		return 0

	case mips32.AJMP: /* read p->to */
		if s != nil {
			if copysub(&p.To, v, s, true) {
				return 1
			}
			return 0
		}

		if copyau(&p.To, v) {
			return 1
		}
		return 0

	case mips32.ARET: /* funny */
		if s != nil {
			return 0
		}

		// All registers die at this point, so claim
		// everything is set (and not used).
		return 3

	case mips32.AJAL: /* funny */

		if v.Type == obj.TYPE_REG && v.Reg == mips32.REGCTXT {
			return 2
		}

		if p.From.Type == obj.TYPE_REG && v.Type == obj.TYPE_REG && p.From.Reg == v.Reg {
			return 2
		}

		if s != nil {
			if copysub(&p.To, v, s, true) {
				return 1
			}
			return 0
		}

		if copyau(&p.To, v) {
			return 4
		}
		return 3

		// TODO(mips32): revise
		// R0 is zero, used by DUFFZERO, cannot be substituted.
		// R1 is ptr to memory, used and set, cannot be substituted.
		// case obj.ADUFFZERO:
		// 	if v.Type == obj.TYPE_REG {
		// 		if v.Reg == mips32.REG_R0 {
		// 			return 1
		// 		}
		// 		if v.Reg == mips32.REG_R0+1 {
		// 			return 2
		// 		}
		// 	}

		// 	return 0

		// // R0 is scratch, set by DUFFCOPY, cannot be substituted.
		// // R1, R2 areptr to src, dst, used and set, cannot be substituted.
		// case obj.ADUFFCOPY:
		// 	if v.Type == obj.TYPE_REG {
		// 		if v.Reg == mips32.REG_R0 {
		// 			return 3
		// 		}
		// 		if v.Reg == mips32.REG_R0+1 || v.Reg == mips32.REG_R0+2 {
		// 			return 2
		// 		}
		// 	}

		return 0

	case obj.ATEXT,
		obj.APCDATA,
		obj.AFUNCDATA,
		obj.AVARDEF,
		obj.AVARKILL,
		obj.AVARLIVE,
		obj.AUSEFIELD:
		return 0
	}
}

// copyas returns 1 if a and v address the same register.
//
// If a is the from operand, this means this operation reads the
// register in v. If a is the to operand, this means this operation
// writes the register in v.
func copyas(a *obj.Addr, v *obj.Addr) bool {
	if regtyp(v) {
		if a.Type == v.Type {
			if a.Reg == v.Reg {
				return true
			}
		}
	}
	return false
}

// copyau returns 1 if a either directly or indirectly addresses the
// same register as v.
//
// If a is the from operand, this means this operation reads the
// register in v. If a is the to operand, this means the operation
// either reads or writes the register in v (if !copyas(a, v), then
// the operation reads the register in v).
func copyau(a *obj.Addr, v *obj.Addr) bool {
	if copyas(a, v) {
		return true
	}
	if v.Type == obj.TYPE_REG {
		if a.Type == obj.TYPE_MEM || (a.Type == obj.TYPE_ADDR && a.Reg != 0) {
			if v.Reg == a.Reg {
				return true
			}
		}
	}
	return false
}

// copyau1 returns true if p->reg references the same register as v and v
// is a direct reference.
func copyau1(p *obj.Prog, v *obj.Addr) bool {
	return regtyp(v) && v.Reg != 0 && p.Reg == v.Reg
}

// copysub replaces v with s in a if f==true or indicates it if could if f==false.
// Returns true on failure to substitute (it always succeeds on mips).
// TODO(dfc) remove unused return value, remove calls with f=false as they do nothing.
func copysub(a *obj.Addr, v *obj.Addr, s *obj.Addr, f bool) bool {
	if f && copyau(a, v) {
		a.Reg = s.Reg
	}
	return false
}

// copysub1 replaces v with s in p1->reg if f==true or indicates if it could if f==false.
// Returns true on failure to substitute (it always succeeds on mips).
// TODO(dfc) remove unused return value, remove calls with f=false as they do nothing.
func copysub1(p1 *obj.Prog, v *obj.Addr, s *obj.Addr, f bool) bool {
	if f && copyau1(p1, v) {
		p1.Reg = s.Reg
	}
	return false
}

func sameaddr(a *obj.Addr, v *obj.Addr) bool {
	if a.Type != v.Type {
		return false
	}
	if regtyp(v) && a.Reg == v.Reg {
		return true
	}
	if v.Type == obj.NAME_AUTO || v.Type == obj.NAME_PARAM {
		if v.Offset == a.Offset {
			return true
		}
	}
	return false
}

func smallindir(a *obj.Addr, reg *obj.Addr) bool {
	return reg.Type == obj.TYPE_REG && a.Type == obj.TYPE_MEM && a.Reg == reg.Reg && 0 <= a.Offset && a.Offset < 4096
}

func stackaddr(a *obj.Addr) bool {
	return a.Type == obj.TYPE_REG && a.Reg == mips32.REGSP
}
