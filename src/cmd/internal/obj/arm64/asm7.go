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
	"cmd/internal/src"
	"fmt"
	"log"
	"math"
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
	ctab map[*obj.Prog][]inst // Prog to insts conversion table
	rtab map[*obj.Prog]uint32 // record the relocation index of each Prog
}

// inst describes a single arch-related machine instruction.
// Inst only contains fields related to encoding.
type inst struct {
	As    obj.As   // assembler opcode
	Optab uint16   // arch-specific opcode index
	Pos   src.XPos // source position of this instruction
	Args  []arg    // operands, in the same order as the machine instruction operands
}

// arg represent a machine instruction operand. It's a subset of the
// Addr data structure, and the meaning of each field is the same as
// that of the field with the same name in Addr.
type arg struct {
	Reg    int16
	Index  int16
	Type   obj.AddrType
	Offset int64

	// argument value:
	//	for TYPE_SCONST, a string
	//	for TYPE_FCONST, a float64
	//	for TYPE_BRANCH, a *Prog (optional)
	//	for TYPE_TEXTSIZE, an int32 (optional)
	Val interface{}
}

func (a *arg) target() *obj.Prog {
	if a.Type == obj.TYPE_BRANCH && a.Val != nil {
		return a.Val.(*obj.Prog)
	}
	return nil
}

// addrToArg returns an arg created based on an obj.Addr
func addrToArg(a obj.Addr) arg {
	return arg{a.Reg, a.Index, a.Type, a.Offset, a.Val}
}

const (
	funcAlign = 16
)

func IsAtomicInstruction(as obj.As) bool {
	return atomicLDADD[as] || atomicSWP[as]
}

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

// valid pstate field values, and value to use in instruction
var pstatefield = []struct {
	opd SpecialOperand
	enc uint32
}{
	{SPOP_DAIFSet, 3<<16 | 6<<5},
	{SPOP_DAIFClr, 3<<16 | 7<<5},
}

// valid prfop field values
var prfopfield = map[SpecialOperand]uint32{
	SPOP_PLDL1KEEP: 0,
	SPOP_PLDL1STRM: 1,
	SPOP_PLDL2KEEP: 2,
	SPOP_PLDL2STRM: 3,
	SPOP_PLDL3KEEP: 4,
	SPOP_PLDL3STRM: 5,
	SPOP_PLIL1KEEP: 8,
	SPOP_PLIL1STRM: 9,
	SPOP_PLIL2KEEP: 10,
	SPOP_PLIL2STRM: 11,
	SPOP_PLIL3KEEP: 12,
	SPOP_PLIL3STRM: 13,
	SPOP_PSTL1KEEP: 16,
	SPOP_PSTL1STRM: 17,
	SPOP_PSTL2KEEP: 18,
	SPOP_PSTL2STRM: 19,
	SPOP_PSTL3KEEP: 20,
	SPOP_PSTL3STRM: 21,
}

// sysInstFields helps convert SYS alias instructions to SYS instructions.
// For example, the format of TLBI is: TLBI <tlbi_op>{, <Xt>}.
// It's equivalent to: SYS #<op1>, C8, <Cm>, #<op2>{, <Xt>}.
// The field hasOperand2 indicates whether Xt is required. It helps to check
// some combinations that may be undefined, such as TLBI VMALLE1IS, R0.
var sysInstFields = map[SpecialOperand]struct {
	op1         uint8
	cn          uint8
	cm          uint8
	op2         uint8
	hasOperand2 bool
}{
	// TLBI
	SPOP_VMALLE1IS:    {0, 8, 3, 0, false},
	SPOP_VAE1IS:       {0, 8, 3, 1, true},
	SPOP_ASIDE1IS:     {0, 8, 3, 2, true},
	SPOP_VAAE1IS:      {0, 8, 3, 3, true},
	SPOP_VALE1IS:      {0, 8, 3, 5, true},
	SPOP_VAALE1IS:     {0, 8, 3, 7, true},
	SPOP_VMALLE1:      {0, 8, 7, 0, false},
	SPOP_VAE1:         {0, 8, 7, 1, true},
	SPOP_ASIDE1:       {0, 8, 7, 2, true},
	SPOP_VAAE1:        {0, 8, 7, 3, true},
	SPOP_VALE1:        {0, 8, 7, 5, true},
	SPOP_VAALE1:       {0, 8, 7, 7, true},
	SPOP_IPAS2E1IS:    {4, 8, 0, 1, true},
	SPOP_IPAS2LE1IS:   {4, 8, 0, 5, true},
	SPOP_ALLE2IS:      {4, 8, 3, 0, false},
	SPOP_VAE2IS:       {4, 8, 3, 1, true},
	SPOP_ALLE1IS:      {4, 8, 3, 4, false},
	SPOP_VALE2IS:      {4, 8, 3, 5, true},
	SPOP_VMALLS12E1IS: {4, 8, 3, 6, false},
	SPOP_IPAS2E1:      {4, 8, 4, 1, true},
	SPOP_IPAS2LE1:     {4, 8, 4, 5, true},
	SPOP_ALLE2:        {4, 8, 7, 0, false},
	SPOP_VAE2:         {4, 8, 7, 1, true},
	SPOP_ALLE1:        {4, 8, 7, 4, false},
	SPOP_VALE2:        {4, 8, 7, 5, true},
	SPOP_VMALLS12E1:   {4, 8, 7, 6, false},
	SPOP_ALLE3IS:      {6, 8, 3, 0, false},
	SPOP_VAE3IS:       {6, 8, 3, 1, true},
	SPOP_VALE3IS:      {6, 8, 3, 5, true},
	SPOP_ALLE3:        {6, 8, 7, 0, false},
	SPOP_VAE3:         {6, 8, 7, 1, true},
	SPOP_VALE3:        {6, 8, 7, 5, true},
	SPOP_VMALLE1OS:    {0, 8, 1, 0, false},
	SPOP_VAE1OS:       {0, 8, 1, 1, true},
	SPOP_ASIDE1OS:     {0, 8, 1, 2, true},
	SPOP_VAAE1OS:      {0, 8, 1, 3, true},
	SPOP_VALE1OS:      {0, 8, 1, 5, true},
	SPOP_VAALE1OS:     {0, 8, 1, 7, true},
	SPOP_RVAE1IS:      {0, 8, 2, 1, true},
	SPOP_RVAAE1IS:     {0, 8, 2, 3, true},
	SPOP_RVALE1IS:     {0, 8, 2, 5, true},
	SPOP_RVAALE1IS:    {0, 8, 2, 7, true},
	SPOP_RVAE1OS:      {0, 8, 5, 1, true},
	SPOP_RVAAE1OS:     {0, 8, 5, 3, true},
	SPOP_RVALE1OS:     {0, 8, 5, 5, true},
	SPOP_RVAALE1OS:    {0, 8, 5, 7, true},
	SPOP_RVAE1:        {0, 8, 6, 1, true},
	SPOP_RVAAE1:       {0, 8, 6, 3, true},
	SPOP_RVALE1:       {0, 8, 6, 5, true},
	SPOP_RVAALE1:      {0, 8, 6, 7, true},
	SPOP_RIPAS2E1IS:   {4, 8, 0, 2, true},
	SPOP_RIPAS2LE1IS:  {4, 8, 0, 6, true},
	SPOP_ALLE2OS:      {4, 8, 1, 0, false},
	SPOP_VAE2OS:       {4, 8, 1, 1, true},
	SPOP_ALLE1OS:      {4, 8, 1, 4, false},
	SPOP_VALE2OS:      {4, 8, 1, 5, true},
	SPOP_VMALLS12E1OS: {4, 8, 1, 6, false},
	SPOP_RVAE2IS:      {4, 8, 2, 1, true},
	SPOP_RVALE2IS:     {4, 8, 2, 5, true},
	SPOP_IPAS2E1OS:    {4, 8, 4, 0, true},
	SPOP_RIPAS2E1:     {4, 8, 4, 2, true},
	SPOP_RIPAS2E1OS:   {4, 8, 4, 3, true},
	SPOP_IPAS2LE1OS:   {4, 8, 4, 4, true},
	SPOP_RIPAS2LE1:    {4, 8, 4, 6, true},
	SPOP_RIPAS2LE1OS:  {4, 8, 4, 7, true},
	SPOP_RVAE2OS:      {4, 8, 5, 1, true},
	SPOP_RVALE2OS:     {4, 8, 5, 5, true},
	SPOP_RVAE2:        {4, 8, 6, 1, true},
	SPOP_RVALE2:       {4, 8, 6, 5, true},
	SPOP_ALLE3OS:      {6, 8, 1, 0, false},
	SPOP_VAE3OS:       {6, 8, 1, 1, true},
	SPOP_VALE3OS:      {6, 8, 1, 5, true},
	SPOP_RVAE3IS:      {6, 8, 2, 1, true},
	SPOP_RVALE3IS:     {6, 8, 2, 5, true},
	SPOP_RVAE3OS:      {6, 8, 5, 1, true},
	SPOP_RVALE3OS:     {6, 8, 5, 5, true},
	SPOP_RVAE3:        {6, 8, 6, 1, true},
	SPOP_RVALE3:       {6, 8, 6, 5, true},
	// DC
	SPOP_IVAC:    {0, 7, 6, 1, true},
	SPOP_ISW:     {0, 7, 6, 2, true},
	SPOP_CSW:     {0, 7, 10, 2, true},
	SPOP_CISW:    {0, 7, 14, 2, true},
	SPOP_ZVA:     {3, 7, 4, 1, true},
	SPOP_CVAC:    {3, 7, 10, 1, true},
	SPOP_CVAU:    {3, 7, 11, 1, true},
	SPOP_CIVAC:   {3, 7, 14, 1, true},
	SPOP_IGVAC:   {0, 7, 6, 3, true},
	SPOP_IGSW:    {0, 7, 6, 4, true},
	SPOP_IGDVAC:  {0, 7, 6, 5, true},
	SPOP_IGDSW:   {0, 7, 6, 6, true},
	SPOP_CGSW:    {0, 7, 10, 4, true},
	SPOP_CGDSW:   {0, 7, 10, 6, true},
	SPOP_CIGSW:   {0, 7, 14, 4, true},
	SPOP_CIGDSW:  {0, 7, 14, 6, true},
	SPOP_GVA:     {3, 7, 4, 3, true},
	SPOP_GZVA:    {3, 7, 4, 4, true},
	SPOP_CGVAC:   {3, 7, 10, 3, true},
	SPOP_CGDVAC:  {3, 7, 10, 5, true},
	SPOP_CGVAP:   {3, 7, 12, 3, true},
	SPOP_CGDVAP:  {3, 7, 12, 5, true},
	SPOP_CGVADP:  {3, 7, 13, 3, true},
	SPOP_CGDVADP: {3, 7, 13, 5, true},
	SPOP_CIGVAC:  {3, 7, 14, 3, true},
	SPOP_CIGDVAC: {3, 7, 14, 5, true},
	SPOP_CVAP:    {3, 7, 12, 1, true},
	SPOP_CVADP:   {3, 7, 13, 1, true},
}

// validPcAlignLength checks if the PC alignment length is valid.
func validPcAlignLength(alignedValue int64) bool {
	return (alignedValue&(alignedValue-1) == 0) && 8 <= alignedValue && alignedValue <= 2048
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
	c.ctab = make(map[*obj.Prog][]inst)
	c.rtab = make(map[*obj.Prog]uint32)
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

	// Now that we know byte offsets, we can generate jump table entries.
	for _, jt := range cursym.Func().JumpTables {
		for i, p := range jt.Targets {
			// The ith jumptable entry points to the p.Pc'th
			// byte in the function symbol s.
			// TODO: try using relative PCs.
			jt.Sym.WriteAddr(ctxt, int64(i)*8, 8, cursym, p.Pc)
		}
	}
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
			pc += int64(len(c.ctab[p]) << 2)
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
				pc += int64(len(c.ctab[p]) << 2)
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
					c.ctab[p] = progToInst(q, Bl, []arg{addrToArg(q.To)})
					// Update p.To.Val and c.ctab[p][0].Args[len(c.ctab[p][0].Args)-1].Val.
					// Branch instructions usually correspond to only one machine instruction,
					// and the last operand of the machine instruction is a label.
					p.To.SetTarget(q)
					c.ctab[p][0].Args[len(c.ctab[p][0].Args)-1].Val = q

					q = c.newprog()
					q.Link = p.Link
					p.Link = q
					q.As = AB
					q.Pos = p.Pos
					q.Mark |= BRANCH26BITS
					q.To.Type = obj.TYPE_BRANCH
					q.To.SetTarget(q.Link.Link)
					c.ctab[p] = progToInst(q, Bl, []arg{addrToArg(q.To)})

					bflag = 1
				}
			}
		}
	}
	return pc
}

// asmins encodes an Inst q of p and returns the binary.
func (c *ctxt7) asmins(p *obj.Prog, q inst) uint32 {
	if q.Optab < 1 {
		// This is supposed to be something that stops execution.
		// It's not supposed to be reached, ever, but if it is, we'd
		// like to be able to tell how we got there. Assemble as
		// 0x0000ffff which is guaranteed to raise undefined instruction
		// exception.
		return 0x0000ffff
	}
	enc := optab[int(q.Optab)].skeleton
	// TODO enable this when necessary
	// enc |= c.encodeOpcode(q.As)
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
	return len(c.ctab[p]) > 1 && p.Mark&NOTUSETMP == 0
}

/*
 * when the first reference to the literal pool threatens
 * to go out of range of a 1Mb PC-relative offset
 * drop the pool now, and branch round it.
 */
func (c *ctxt7) checkpool(p *obj.Prog, skip int) {
	if c.pool.size >= 0xffff0 || !ispcdisp(int32(p.Pc+4+int64(c.pool.size)-int64(c.pool.start)+8)) {
		// Don't break relocation types whose Siz > 4, such as R_ADDRARM64
		if !(c.rtab[p] > 0 && c.cursym.R[c.rtab[p]-1].Siz > 4 && skip != 0) {
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
		c.ctab[p] = progToInst(q, Bl, []arg{addrToArg(p.To)})
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

func isANDWop(op obj.As) bool {
	switch op {
	case AANDW, AORRW, AEORW, AANDSW, ATSTW,
		ABICW, AEONW, AORNW, ABICSW:
		return true
	}
	return false
}

func isADDWop(op obj.As) bool {
	switch op {
	case AADDW, AADDSW, ASUBW, ASUBSW, ACMNW, ACMPW:
		return true
	}
	return false
}

// idx is obj.Addr.Index field.
func isRegShiftOrExt(idx int16) bool {
	return (idx-obj.RBaseARM64)&REG_EXT != 0 || (idx-obj.RBaseARM64)&REG_LSL != 0
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

// bitconEncode returns the encoding of a bitcon used in logical instructions
// x is known to be a bitcon
// a bitcon is a sequence of n ones at low bits (i.e. 1<<n-1), right rotated
// by R bits, and repeated with period of 64, 32, 16, 8, 4, or 2.
// it is encoded in logical instructions with 3 bitfields
// N (1 bit) : R (6 bits) : S (6 bits), where
// N=1           -- period=64
// N=0, S=0xxxxx -- period=32
// N=0, S=10xxxx -- period=16
// N=0, S=110xxx -- period=8
// N=0, S=1110xx -- period=4
// N=0, S=11110x -- period=2
// R is the shift amount, low bits of S = n-1
func bitconEncode(x uint64, mode int) uint32 {
	var period uint32
	// determine the period and sign-extend a unit to 64 bits
	switch {
	case x != x>>32|x<<32:
		period = 64
	case x != x>>16|x<<48:
		period = 32
		x = uint64(int64(int32(x)))
	case x != x>>8|x<<56:
		period = 16
		x = uint64(int64(int16(x)))
	case x != x>>4|x<<60:
		period = 8
		x = uint64(int64(int8(x)))
	case x != x>>2|x<<62:
		period = 4
		x = uint64(int64(x<<60) >> 60)
	default:
		period = 2
		x = uint64(int64(x<<62) >> 62)
	}
	neg := false
	if int64(x) < 0 {
		x = ^x
		neg = true
	}
	y := x & -x // lowest set bit of x.
	s := log2(y)
	n := log2(x+y) - s // x (or ^x) is a sequence of n ones left shifted by s bits
	if neg {
		// ^x is a sequence of n ones left shifted by s bits
		// adjust n, s for x
		s = n + s
		n = period - n
	}

	N := uint32(0)
	if mode == 64 && period == 64 {
		N = 1
	}
	R := (period - s) & (period - 1) & uint32(mode-1) // shift amount of right rotate
	S := (n - 1) | 63&^(period<<1-1)                  // low bits = #ones - 1, high bits encodes period
	return N<<22 | R<<16 | S<<10
}

func log2(x uint64) uint32 {
	if x == 0 {
		panic("log2 of 0")
	}
	n := uint32(0)
	for i := uint32(32); i > 0; i >>= 1 {
		if x >= 1<<i {
			x >>= i
			n += i
		}
	}
	return n
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

// rclass classifies register r.
func rclass(r int16) argtype {
	switch {
	case REG_R0 <= r && r <= REG_R31, r == REG_RSP:
		return C_REG
	case REG_F0 <= r && r <= REG_F31:
		return C_FREG
	case REG_V0 <= r && r <= REG_V31:
		return C_VREG
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

	case obj.TYPE_SPECIAL:
		opd := SpecialOperand(a.Offset)
		if SPOP_EQ <= opd && opd <= SPOP_NV {
			return C_COND
		}
		return C_SPOP
	}

	return C_NONE
}

func buildop(ctxt *obj.Link) {
}

// chipfloat7() checks if the immediate constants available in  FMOVS/FMOVD instructions.
// For details of the range of constants available, see
// http://infocenter.arm.com/help/topic/com.arm.doc.dui0473m/dom1359731199385.html.
func (c *ctxt7) chipfloat7(e float64) int {
	ei := math.Float64bits(e)
	l := uint32(int32(ei))
	h := uint32(int32(ei >> 32))

	if l != 0 || h&0xffff != 0 {
		return -1
	}
	h1 := h & 0x7fc00000
	if h1 != 0x40000000 && h1 != 0x3fc00000 {
		return -1
	}
	n := 0

	// sign bit (a)
	if h&0x80000000 != 0 {
		n |= 1 << 7
	}

	// exp sign bit (b)
	if h1 == 0x3fc00000 {
		n |= 1 << 6
	}

	// rest of exp and mantissa (cd-efgh)
	n |= int((h >> 16) & 0x3f)

	//print("match %.8lux %.8lux %d\n", l, h, n);
	return n
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

// checkindex checks if index >= 0 && index <= maxindex.
func (c *ctxt7) checkindex(p *obj.Prog, index, maxindex int16) {
	if index < 0 || index > maxindex {
		c.ctxt.Diag("register element index out of range 0 to %d: %v", maxindex, p)
	}
}

// checkShiftAmount checks whether the index shift amount is valid
// for load with register offset instructions.
func (c *ctxt7) checkShiftAmount(p *obj.Prog, a *arg) {
	var amount int16
	amount = (a.Index >> 5) & 7
	switch p.As {
	case AMOVB, AMOVBU:
		if amount != 0 {
			c.ctxt.Diag("invalid index shift amount: %v", p)
		}
	case AMOVH, AMOVHU:
		if amount != 1 && amount != 0 {
			c.ctxt.Diag("invalid index shift amount: %v", p)
		}
	case AMOVW, AMOVWU, AFMOVS:
		if amount != 2 && amount != 0 {
			c.ctxt.Diag("invalid index shift amount: %v", p)
		}
	case AMOVD, AFMOVD, APRFM:
		if amount != 3 && amount != 0 {
			c.ctxt.Diag("invalid index shift amount: %v", p)
		}
	default:
		panic("invalid operation")
	}
}

// emitCode encodes each Prog and writes the encoding into Lsym.P.
func (c *ctxt7) emitCode() {
	c.cursym.Grow(c.cursym.Size)
	bp := c.cursym.P
	for p := c.cursym.Func().Text.Link; p != nil; p = p.Link {
		if c.rtab[p] > 0 {
			c.cursym.R[c.rtab[p]-1].Off = int32(p.Pc)
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
			if insts, ok := c.ctab[p]; !ok || insts == nil {
				c.ctxt.Diag("illegal combination: %v", p)
				continue
			}
			for _, q := range c.ctab[p] {
				var out uint32 = c.asmins(p, q)
				c.ctxt.Arch.ByteOrder.PutUint32(bp, out)
				bp = bp[4:]
			}
		}
	}
}

// brdist computes branch distance.
func (c *ctxt7) brdist(p *obj.Prog, arg *arg, preshift int, flen int, shift int) int64 {
	v := int64(0)
	t := int64(0)
	q := arg.target()
	if q == nil {
		q = p.Pool
	}
	if q != nil {
		v = (q.Pc >> uint(preshift)) - (c.pc >> uint(preshift))
		if (v & ((1 << uint(shift)) - 1)) != 0 {
			c.ctxt.Diag("misaligned label, func %s: %v", c.cursym.Name, p)
		}
		v >>= uint(shift)
		t = int64(1) << uint(flen-1)
		if v < -t || v >= t {
			c.ctxt.Diag("branch too far %#x vs %#x [%p]\n%v\n%v", v, t, c.blitrl, p, q)
			panic("branch too far")
		}
	}
	return v & ((t << 1) - 1)
}

func (c *ctxt7) opbfm(p *obj.Prog, rs, shift uint, max uint32) uint32 {
	if rs < 0 || uint32(rs) > max {
		c.ctxt.Diag("illegal bit number: %v", p)
	}
	return (uint32(rs) & 0x3F) << shift
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

// rm is the Rm register value, o is the extension, amount is the left shift value.
func roff(rm int16, o uint32, amount int16) uint32 {
	return uint32(rm&31)<<16 | o<<13 | uint32(amount)<<10
}

// encRegShiftOrExt returns the encoding of shifted/extended register, Rx<<n and Rx.UXTW<<n, etc.
func (c *ctxt7) encRegShiftOrExt(typ obj.AddrType, r int16, is32bit bool) uint32 {
	var num, rm int16
	num = (r >> 5) & 7
	rm = r & 31
	switch {
	case REG_UXTB <= r && r < REG_UXTH:
		return roff(rm, 0, num)
	case REG_UXTH <= r && r < REG_UXTW:
		return roff(rm, 1, num)
	case REG_UXTW <= r && r < REG_UXTX:
		if typ == obj.TYPE_MEM {
			if num == 0 {
				return roff(rm, 2, 2)
			} else {
				return roff(rm, 2, 6)
			}
		} else {
			return roff(rm, 2, num)
		}
	case REG_UXTX <= r && r < REG_SXTB:
		return roff(rm, 3, num)
	case REG_SXTB <= r && r < REG_SXTH:
		return roff(rm, 4, num)
	case REG_SXTH <= r && r < REG_SXTW:
		return roff(rm, 5, num)
	case REG_SXTW <= r && r < REG_SXTX:
		if typ == obj.TYPE_MEM {
			if num == 0 {
				return roff(rm, 6, 2)
			} else {
				return roff(rm, 6, 6)
			}
		} else {
			return roff(rm, 6, num)
		}
	case REG_SXTX <= r && r < REG_SPECIAL:
		if typ == obj.TYPE_MEM {
			if num == 0 {
				return roff(rm, 7, 2)
			} else {
				return roff(rm, 7, 6)
			}
		} else {
			return roff(rm, 7, num)
		}
	case REG_LSL <= r && r < REG_ARNG:
		if typ == obj.TYPE_MEM { // (R1)(R2<<1)
			return roff(rm, 3, 6)
		} else if is32bit {
			// For 32-bit arithmetic operation instructions, such as ADD (extended register),
			// the encoding of LSL is "010".
			return roff(rm, 2, num)
		}
		return roff(rm, 3, num)
	default:
		c.ctxt.Diag("unsupported register extension type.")
	}
	return 0
}
