// cmd/7l/asm.c, cmd/7l/asmout.c, cmd/7l/optab.c, cmd/7l/span.c, cmd/ld/sub.c, cmd/ld/mod.c, from Vita Nuova.
// https://code.google.com/p/ken-cc/source/browse/
//
// 	Copyright © 1994-1999 Lucent Technologies Inc. All rights reserved.
// 	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
// 	Portions Copyright © 1997-1999 Vita Nuova Limited
// 	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
// 	Portions Copyright © 2004,2006 Bruce Ellis
// 	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
// 	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
// 	Portions Copyright © 2009 The Go Authors. All rights reserved.
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

package arm64

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"fmt"
	"log"
)

// ctxt7 holds state while assembling a single function.
// Each function gets a fresh ctxt7.
// This allows for multiple functions to be safely concurrently assembled.
type ctxt7 struct {
	ctxt       *obj.Link
	newprog    obj.ProgAlloc
	cursym     *obj.LSym
	blitrl     *obj.Prog
	elitrl     *obj.Prog
	autosize   int32
	extrasize  int32
	instoffset int64
	pc         int64
	pool       struct {
		start uint32
		size  uint32
	}
}

const (
	funcAlign = 16
)

// Atomic like instructions.
var atomicLDADD = map[obj.As]bool{
	ALDADDAD:  true,
	ALDADDAW:  true,
	ALDADDAH:  true,
	ALDADDAB:  true,
	ALDADDALD: true,
	ALDADDALW: true,
	ALDADDALH: true,
	ALDADDALB: true,
	ALDADDD:   true,
	ALDADDW:   true,
	ALDADDH:   true,
	ALDADDB:   true,
	ALDADDLD:  true,
	ALDADDLW:  true,
	ALDADDLH:  true,
	ALDADDLB:  true,
	ALDCLRAD:  true,
	ALDCLRAW:  true,
	ALDCLRAH:  true,
	ALDCLRAB:  true,
	ALDCLRALD: true,
	ALDCLRALW: true,
	ALDCLRALH: true,
	ALDCLRALB: true,
	ALDCLRD:   true,
	ALDCLRW:   true,
	ALDCLRH:   true,
	ALDCLRB:   true,
	ALDCLRLD:  true,
	ALDCLRLW:  true,
	ALDCLRLH:  true,
	ALDCLRLB:  true,
	ALDEORAD:  true,
	ALDEORAW:  true,
	ALDEORAH:  true,
	ALDEORAB:  true,
	ALDEORALD: true,
	ALDEORALW: true,
	ALDEORALH: true,
	ALDEORALB: true,
	ALDEORD:   true,
	ALDEORW:   true,
	ALDEORH:   true,
	ALDEORB:   true,
	ALDEORLD:  true,
	ALDEORLW:  true,
	ALDEORLH:  true,
	ALDEORLB:  true,
	ALDORAD:   true,
	ALDORAW:   true,
	ALDORAH:   true,
	ALDORAB:   true,
	ALDORALD:  true,
	ALDORALW:  true,
	ALDORALH:  true,
	ALDORALB:  true,
	ALDORD:    true,
	ALDORW:    true,
	ALDORH:    true,
	ALDORB:    true,
	ALDORLD:   true,
	ALDORLW:   true,
	ALDORLH:   true,
	ALDORLB:   true,
}

var atomicSWP = map[obj.As]bool{
	ASWPAD:  true,
	ASWPAW:  true,
	ASWPAH:  true,
	ASWPAB:  true,
	ASWPALD: true,
	ASWPALW: true,
	ASWPALH: true,
	ASWPALB: true,
	ASWPD:   true,
	ASWPW:   true,
	ASWPH:   true,
	ASWPB:   true,
	ASWPLD:  true,
	ASWPLW:  true,
	ASWPLH:  true,
	ASWPLB:  true,
	ACASD:   true,
	ACASW:   true,
	ACASH:   true,
	ACASB:   true,
	ACASAD:  true,
	ACASAW:  true,
	ACASLD:  true,
	ACASLW:  true,
	ACASALD: true,
	ACASALW: true,
	ACASALH: true,
	ACASALB: true,
}

func IsAtomicInstruction(as obj.As) bool {
	if atomicLDADD[as] || atomicSWP[as] {
		return true
	}
	return false
}

func buildop(ctxt *obj.Link) {
}

// arm64 assembly entry.
func span7(ctxt *obj.Link, cursym *obj.LSym, newprog obj.ProgAlloc) {
	if ctxt.Retpoline {
		ctxt.Diag("-spectre=ret not supported on arm64")
		ctxt.Retpoline = false // don't keep printing
	}

	p := cursym.Func().Text
	if p == nil || p.Link == nil { // handle external functions and ELF section symbols
		return
	}

	c := ctxt7{ctxt: ctxt, newprog: newprog, cursym: cursym, autosize: int32(p.To.Offset & 0xffffffff), extrasize: int32(p.To.Offset >> 32)}
	p.To.Offset &= 0xffffffff // extrasize is no longer needed

	c.unfold(p)            // convert Go macro assembly instructions to low level machine instructions
	pc := c.literalPool(p) // handle literal pool
	// if any procedure is large enough to generate a large SBRA branch, then
	// generate extra passes putting branches around jmps to fix. this is rare.
	pc = c.branchFixup(p)
	pc += -pc & (funcAlign - 1)
	c.cursym.Size = pc
	c.emitCode() // lay out the code, emitting code

	// Mark nonpreemptible instruction sequences.
	// We use REGTMP as a scratch register during call injection,
	// so instruction sequences that use REGTMP are unsafe to
	// preempt asynchronously.
	obj.MarkUnsafePoints(c.ctxt, c.cursym.Func().Text, c.newprog, c.isUnsafePoint, c.isRestartable)
}

// unfold converts each valid Prog of the text to the low level machine instruction representation,
// and set the relocation type, literal pool mark and branch mark, if necessary.
func (c *ctxt7) unfold(text *obj.Prog) {
	pre := text
	for p := pre.Link; p != nil; p = p.Link {
		// Special cases that don't need to unfold
		switch p.As {
		case obj.ATEXT, obj.ANOP, obj.AFUNCDATA, obj.APCDATA, obj.APCALIGN, ADWORD, AWORD:
			continue
		}
		index := p.As - obj.ABaseARM64
		if p.As < obj.A_ARCHSPECIFIC {
			index = p.As
		}
		unfoldTab[index](c, p)
	}
}

// validPcAlignLength checks if the PC alignment length is valid.
func validPcAlignLength(alignedValue int64) bool {
	return (alignedValue&(alignedValue-1) == 0) && 8 <= alignedValue && alignedValue <= 2048
}

// literalPool deals with literal pool and computes the PC value of each Prog.
func (c *ctxt7) literalPool(text *obj.Prog) int64 {
	pc := int64(0)
	text.Pc = pc
	for p := text.Link; p != nil; p = p.Link {
		p.Pc = pc
		switch p.As {
		case obj.ANOP, obj.AFUNCDATA, obj.APCDATA:
			continue
		case obj.APCALIGN:
			alignedValue := p.From.Offset
			if !validPcAlignLength(alignedValue) {
				c.ctxt.Diag("alignment value of an instruction must be a power of two and in the range [8, 2048], got %d\n", alignedValue)
				continue
			}
			pc += (-pc & (alignedValue - 1))
			// Update the current text symbol alignment value
			if int32(alignedValue) > c.cursym.Func().Align {
				c.cursym.Func().Align = int32(alignedValue)
			}
		case ADWORD:
			pc += 8 // DWORD holds 2 instructions
		case AWORD:
			pc += 4
		default:
			pc += int64(len(p.Insts) << 2)
		}
		if p.Mark&LFROM != 0 {
			c.addpool(p, &p.From)
		}
		if p.Mark&LFROM128 != 0 {
			c.addpool128(p, &p.From, p.GetFrom3())
		}
		if p.Mark&LTO != 0 {
			c.addpool(p, &p.To)
		}
		if c.blitrl != nil {
			if p.As == AB || p.As == obj.ARET || p.As == AERET { /* TODO: other unconditional operations */
				// Do not need to insert a JMP instruction after these instructions
				c.checkpool(p, 0)
			} else {
				c.checkpool(p, 1)
			}
		}
	}
	return pc
}

/*
 * when the first reference to the literal pool threatens
 * to go out of range of a 1Mb PC-relative offset
 * drop the pool now, and branch round it.
 */
func (c *ctxt7) checkpool(p *obj.Prog, skip int) {
	if c.pool.size >= 0xffff0 || !ispcdisp(int32(p.Pc+4+int64(c.pool.size)-int64(c.pool.start)+8)) {
		// Don't break relocation types whose Siz > 4, such as R_ADDRARM64
		if !(p.RelocIdx > 0 && c.cursym.R[p.RelocIdx-1].Siz > 4 && skip != 0) {
			c.flushpool(p, skip)
		}
	} else if p.Link == nil {
		// Don't need to insert a JMP instruction at the end of the function
		c.flushpool(p, 0)
	}
}

func (c *ctxt7) flushpool(p *obj.Prog, skip int) {
	if skip != 0 {
		if c.ctxt.Debugvlog && skip == 1 {
			fmt.Printf("note: flush literal pool at %#x: len=%d ref=%x\n", uint64(p.Pc+4), c.pool.size, c.pool.start)
		}
		q := c.newprog()
		q.As = AB
		q.Pos = p.Pos
		q.Mark |= BRANCH26BITS
		q.To.Type = obj.TYPE_BRANCH
		q.To.SetTarget(p.Link)
		q.Insts = progToInst(q, Bl, []obj.Oprd{p.To.ToOprd()})
		q.Link = c.blitrl
		c.blitrl = q
	} else if p.Pc+int64(c.pool.size)-int64(c.pool.start) < maxPCDisp && p.Link != nil {
		return
	}

	// The line number for constant pool entries doesn't really matter.
	// We set it to the line number of the preceding instruction so that
	// there are no deltas to encode in the pc-line tables.
	for q := c.blitrl; q != nil; q = q.Link {
		q.Pos = p.Pos
	}

	c.elitrl.Link = p.Link
	p.Link = c.blitrl

	c.blitrl = nil /* BUG: should refer back to values until out-of-range */
	c.elitrl = nil
	c.pool.size = 0
	c.pool.start = 0
}

// addpool128 adds a 128-bit constant to literal pool by two consecutive DWORD
// instructions, the 128-bit constant is formed by ah.Offset<<64+al.Offset.
func (c *ctxt7) addpool128(p *obj.Prog, al, ah *obj.Addr) {
	q := c.newprog()
	q.As = ADWORD
	q.To.Type = obj.TYPE_CONST
	q.To.Offset = al.Offset // q.Pc is lower than t.Pc, so al.Offset is stored in q.

	t := c.newprog()
	t.As = ADWORD
	t.To.Type = obj.TYPE_CONST
	t.To.Offset = ah.Offset

	q.Link = t

	if c.blitrl == nil {
		c.blitrl = q
		c.pool.start = uint32(p.Pc)
	} else {
		c.elitrl.Link = q
	}

	c.elitrl = t
	c.pool.size = roundUp(c.pool.size, 16)
	c.pool.size += 16
	p.Pool = q
}

func (c *ctxt7) addpool(p *obj.Prog, a *obj.Addr) {
	cls := c.aclass(p, a)
	lit := a.Offset
	t := c.newprog()
	t.As = AWORD
	sz := 4

	if a.Type == obj.TYPE_CONST {
		if (lit != int64(int32(lit)) && uint64(lit) != uint64(uint32(lit))) || p.As == AVMOVQ || p.As == AVMOVD {
			// out of range -0x80000000 ~ 0xffffffff or VMOVQ or VMOVD operand, must store 64-bit.
			t.As = ADWORD
			sz = 8
		} // else store 32-bit
	} else if p.As == AMOVD && a.Type != obj.TYPE_MEM || cls == C_ADDR || cls == C_VCON || lit != int64(int32(lit)) || uint64(lit) != uint64(uint32(lit)) {
		// conservative: don't know if we want signed or unsigned extension.
		// in case of ambiguity, store 64-bit
		t.As = ADWORD
		sz = 8
	}

	t.To.Type = obj.TYPE_CONST
	t.To.Offset = lit

	for q := c.blitrl; q != nil; q = q.Link { /* could hash on t.t0.offset */
		if q.To == t.To {
			p.Pool = q
			return
		}
	}

	if c.blitrl == nil {
		c.blitrl = t
		c.pool.start = uint32(p.Pc)
	} else {
		c.elitrl.Link = t
	}
	c.elitrl = t
	if t.As == ADWORD {
		// make DWORD 8-byte aligned, this is not required by ISA,
		// just to avoid performance penalties when loading from
		// the constant pool across a cache line.
		c.pool.size = roundUp(c.pool.size, 8)
	}
	c.pool.size += uint32(sz)
	p.Pool = t
}

// roundUp rounds up x to "to".
func roundUp(x, to uint32) uint32 {
	if to == 0 || to&(to-1) != 0 {
		log.Fatalf("rounded up to a value that is not a power of 2: %d\n", to)
	}
	return (x + to - 1) &^ (to - 1)
}

// Maximum PC-relative displacement.
// The actual limit is ±2²⁰, but we are conservative
// to avoid needing to recompute the literal pool flush points
// as span-dependent jumps are enlarged.
const maxPCDisp = 512 * 1024

// ispcdisp reports whether v is a valid PC-relative displacement.
func ispcdisp(v int32) bool {
	return -maxPCDisp < v && v < maxPCDisp && v&3 == 0
}

// branchFixup handles large branches that exceed the jump range of
// a specific branch instruction, and computes the PC value of each Prog.
func (c *ctxt7) branchFixup(text *obj.Prog) int64 {
	bflag := 1
	pc := int64(0)
	text.Pc = pc
	for bflag != 0 {
		bflag = 0
		pc = 0
		for p := text.Link; p != nil; p = p.Link {
			p.Pc = pc
			switch p.As {
			case obj.ANOP, obj.AFUNCDATA, obj.APCDATA:
				continue
			case obj.APCALIGN:
				alignedValue := p.From.Offset
				if !validPcAlignLength(alignedValue) {
					c.ctxt.Diag("alignment value of an instruction must be a power of two and in the range [8, 2048], got %d\n", alignedValue)
					continue
				}
				pc += (-pc & (alignedValue - 1))
				// Update the current text symbol alignment value
				if int32(alignedValue) > c.cursym.Func().Align {
					c.cursym.Func().Align = int32(alignedValue)
				}
			case ADWORD:
				pc += 8 // DWORD holds 2 instructions
			case AWORD:
				pc += 4
			default:
				pc += int64(len(p.Insts) << 2)
			}
			// very large branches
			if (p.Mark&BRANCH14BITS != 0 || p.Mark&BRANCH19BITS != 0 || p.Mark&BRANCH26BITS != 0) && p.To.Target() != nil {
				/* BRANCH19BITS: BEQ, CBZ and like, BRANCH14BITS: TBZ and like, BRANCH26BITS: B, BL and like */
				otxt := p.To.Target().Pc - pc
				var toofar bool
				if p.Mark&BRANCH14BITS != 0 { // branch instruction encodes 14 bits
					toofar = otxt <= -(1<<15)+10 || otxt >= (1<<15)-10
				} else if p.Mark&BRANCH19BITS != 0 { // branch instruction encodes 19 bits
					toofar = otxt <= -(1<<20)+10 || otxt >= (1<<20)-10
				} else if p.Mark&BRANCH26BITS != 0 { // branch instruction encodes 26 bits
					toofar = otxt <= -(1<<27)+10 || otxt >= (1<<27)-10
				}
				if toofar {
					q := c.newprog()
					q.Link = p.Link
					p.Link = q
					q.As = AB
					q.Pos = p.Pos
					q.Mark |= BRANCH26BITS
					q.To.Type = obj.TYPE_BRANCH
					q.To.SetTarget(p.To.Target())
					q.Insts = progToInst(q, Bl, []obj.Oprd{q.To.ToOprd()})
					// Update p.To.Val and p.Insts[0].Args[len(p.Insts[0].Args)-1].Val.
					// Branch instructions usually correspond to only one machine instruction,
					// and the last operand of the machine instruction is a label.
					p.To.SetTarget(q)
					p.Insts[0].Args[len(p.Insts[0].Args)-1].Val = q

					q = c.newprog()
					q.Link = p.Link
					p.Link = q
					q.As = AB
					q.Pos = p.Pos
					q.Mark |= BRANCH26BITS
					q.To.Type = obj.TYPE_BRANCH
					q.To.SetTarget(q.Link.Link)
					q.Insts = progToInst(q, Bl, []obj.Oprd{q.To.ToOprd()})

					bflag = 1
				}
			}
		}
	}
	return pc
}

// emitCode encodes each Prog and writes the encoding into Lsym.P.
func (c *ctxt7) emitCode() {
	c.cursym.Grow(c.cursym.Size)
	bp := c.cursym.P
	for p := c.cursym.Func().Text.Link; p != nil; p = p.Link {
		if p.RelocIdx > 0 {
			c.cursym.R[p.RelocIdx-1].Off = int32(p.Pc)
		}
		c.pc = p.Pc
		switch p.As {
		case obj.AFUNCDATA, obj.APCDATA, obj.ANOP:
			continue
		case ADWORD:
			if !(p.To.Type == obj.TYPE_CONST || p.To.Type == obj.TYPE_MEM) {
				c.ctxt.Diag("illegal combination: %v\n", p)
			}
			// the large constant is saved in p.To.Offset
			o1 := uint32(p.To.Offset)
			o2 := uint32(p.To.Offset >> 32)
			if p.To.Sym != nil {
				idx := newRelocation(c.cursym, 8, p.To.Sym, p.To.Offset, objabi.R_ADDR)
				c.cursym.R[idx-1].Off = int32(c.pc)
				o2 = 0
				o1 = o2
			}
			// Mark DWORD is not restartable, don't need to set p.Isize.
			p.Mark |= NOTUSETMP
			c.ctxt.Arch.ByteOrder.PutUint32(bp, o1)
			bp = bp[4:]
			c.ctxt.Arch.ByteOrder.PutUint32(bp, o2)
			bp = bp[4:]
		case AWORD:
			if !(p.To.Type == obj.TYPE_CONST && uint64(p.To.Offset) == uint64(uint32(p.To.Offset))) {
				c.ctxt.Diag("invalid constant for WORD: %v\n", p)
			}
			o1 := uint32(p.To.Offset)
			c.ctxt.Arch.ByteOrder.PutUint32(bp, o1)
			bp = bp[4:]
		case obj.APCALIGN:
			v := (-p.Pc & (p.From.Offset - 1)) // p.From.Offset is the alignment value
			var OP_NOOP = optab[NOP].skeleton  // nop is always the same as its skeleton
			for i := 0; i < int(v/4); i++ {
				// emit ANOOP instruction by the padding size
				c.ctxt.Arch.ByteOrder.PutUint32(bp, OP_NOOP)
				bp = bp[4:]
			}
		default:
			if p.Insts == nil {
				c.ctxt.Diag("illegal combination: %v", p)
				continue
			}
			for _, q := range p.Insts {
				var out uint32 = c.asmins(p, q)
				c.ctxt.Arch.ByteOrder.PutUint32(bp, out)
				bp = bp[4:]
			}
		}
	}
}

// asmins encodes an Inst q of p and returns the binary.
func (c *ctxt7) asmins(p *obj.Prog, q obj.Inst) uint32 {
	if q.Optab < 1 {
		// This is supposed to be something that stops execution.
		// It's not supposed to be reached, ever, but if it is, we'd
		// like to be able to tell how we got there. Assemble as
		// 0x0000ffff which is guaranteed to raise undefined instruction
		// exception.
		return 0x0000ffff
	}
	enc := optab[int(q.Optab)].skeleton
	enc |= c.encodeOpcode(q.As)
	for i, a := range q.Args {
		enc |= c.encodeArg(p, &a, optab[int(q.Optab)].args[i])
	}
	return enc
}

// isUnsafePoint returns whether p is an unsafe point.
func (c *ctxt7) isUnsafePoint(p *obj.Prog) bool {
	// If p explicitly uses REGTMP, it's unsafe to preempt, because the
	// preemption sequence clobbers REGTMP.
	for _, a := range p.RestArgs {
		if a.Reg == REGTMP {
			return true
		}
	}
	return p.From.Reg == REGTMP || p.To.Reg == REGTMP || p.Reg == REGTMP
}

// isRestartable returns whether p is a multi-instruction sequence that,
// if preempted, can be restarted.
func (c *ctxt7) isRestartable(p *obj.Prog) bool {
	if c.isUnsafePoint(p) {
		return false
	}
	// If p is a multi-instruction sequence with uses REGTMP inserted by
	// the assembler in order to materialize a large constant/offset, we
	// can restart p (at the start of the instruction sequence), recompute
	// the content of REGTMP, upon async preemption. Currently, all cases
	// of assembler-inserted REGTMP fall into this category.
	// If p doesn't use REGTMP, it can be simply preempted, so we don't
	// mark it.
	return len(p.Insts) > 1 && p.Mark&NOTUSETMP == 0
}
