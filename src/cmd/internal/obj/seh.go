// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
)

const (
	SehOpStackGrow = iota
	SehOpSaveReg
	SehOpSetFP
)

type sehUnwindCode struct {
	Prog      *Prog
	Operation uint8
}

func (ctxt *Link) HasSeh() bool {
	return ctxt.Headtype == objabi.Hwindows && ctxt.Arch.Family == sys.AMD64
}

func (fi *FuncInfo) AddSehUnwindCode(p *Prog, op uint8) {
	fi.sehUnwindCodes = append(fi.sehUnwindCodes, sehUnwindCode{p, op})
}

// populateSeh fills in the SEH unwind information symbol for TEXT symbol 's'.
func (ctxt *Link) populateSeh(s *LSym) {
	if s.Type != objabi.STEXT {
		ctxt.Diag("populateSeh of non-TEXT %v", s)
	}
	switch ctxt.Arch.Family {
	case sys.AMD64:
		ctxt.populateSehamd64(s.Func())
	}
}

type sehctxt struct {
	ctxt *Link
	s    *LSym
	off  int64
}

func (ctxt *sehctxt) write8(v uint8) {
	ctxt.s.WriteInt(ctxt.ctxt, ctxt.off, 1, int64(v))
	ctxt.off += 1
}

func (ctxt *sehctxt) write16(v uint16) {
	ctxt.s.WriteInt(ctxt.ctxt, ctxt.off, 2, int64(v))
	ctxt.off += 2
}

func (ctxt *sehctxt) write32(v uint32) {
	ctxt.s.WriteInt(ctxt.ctxt, ctxt.off, 4, int64(v))
	ctxt.off += 4
}

func (ctxt *Link) populateSehamd64(fn *FuncInfo) {
	if len(fn.sehUnwindCodes) == 0 {
		return
	}
	const (
		// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#unwind-operation-code
		UWOP_ALLOC_LARGE     = 1
		UWOP_ALLOC_SMALL     = 2
		UWOP_SET_FPREG       = 3
		UWOP_SAVE_NONVOL     = 4
		UWOP_SAVE_NONVOL_FAR = 5

		REG_RBP = 5
	)
	const (
		stackGrowSmall = 136
		stackGrowLarge = 512 * 1000
	)
	fn.sehUnwindInfoSym = &LSym{
		Type: objabi.SSEHUNWINDINFO,
	}
	sctxt := sehctxt{ctxt: ctxt, s: fn.sehUnwindInfoSym}
	writecode := func(op, value uint8) {
		sctxt.write8(value<<4 | op)
	}
	// We need to write the number of codes before encoding each one.
	// Each code can be encoded as 1, 2 or 3 entries.
	var ncodes, fpoffset uint8
	var hasfp bool
	for _, c := range fn.sehUnwindCodes {
		p := c.Prog
		switch c.Operation {
		case SehOpStackGrow:
			size := p.From.Offset
			if size < stackGrowSmall {
				ncodes += 1
			} else if size < stackGrowLarge {
				ncodes += 2
			} else {
				ncodes += 3
			}
		case SehOpSaveReg:
			size := p.To.Offset
			if size < stackGrowLarge {
				ncodes += 2
			} else if size < stackGrowLarge {
				ncodes += 3
			}
		case SehOpSetFP:
			// Windows ABI defines a 16-byte aligned stack, so
			// SEH only allows to define 16-byte FP offsets.
			// Unfortunately, Go ABI defines a 8-byte aligned stack,
			// which means that FP registers can have an offset from SP
			// which is not compatible with SEH.
			// Additionally, SEH only allows offsets lower than 240 bytes.
			// We can only skip the SEH frame pointer in both situations,
			// which means that some functions won't be unwindable if
			// they change the stack pointer outside the prologue,
			// e.g., when using systemstack.
			fpoffset = uint8(p.From.Offset)
			if fpoffset < 240 && fpoffset%16 == 0 {
				ncodes += 1
				hasfp = true
			}
		default:
			ctxt.Diag("unsupported SEH operation %d", c.Operation)
		}
	}
	fn.sehUnwindInfoSym.GrowCap(int64(ncodes*2) + 9)
	sctxt.write8(1)                                                               // Flags + version
	sctxt.write8(uint8(fn.sehUnwindCodes[len(fn.sehUnwindCodes)-1].Prog.Link.Pc)) // Size of prolog
	sctxt.write8(uint8(ncodes))                                                   // Count of unwind codes
	if hasfp {
		sctxt.write8((fpoffset/16)<<4 | REG_RBP)
	} else {
		sctxt.write8(0)
	}
	for i := len(fn.sehUnwindCodes) - 1; i >= 0; i-- {
		op := fn.sehUnwindCodes[i].Operation
		if op == SehOpSetFP && !hasfp {
			continue
		}
		p := fn.sehUnwindCodes[i].Prog
		sctxt.write8(uint8(p.Link.Pc))
		switch op {
		case SehOpStackGrow:
			if p.Spadj < stackGrowSmall {
				writecode(UWOP_ALLOC_SMALL, uint8(p.Spadj/8)-1)
			} else if p.Spadj < stackGrowLarge {
				writecode(UWOP_ALLOC_LARGE, 0)
				sctxt.write16(uint16(p.Spadj / 8))
			} else {
				writecode(UWOP_ALLOC_LARGE, 1)
				sctxt.write16(uint16(p.Spadj))
				sctxt.write16(uint16(p.Spadj) << 2)
			}
		case SehOpSaveReg:
			size := p.To.Offset
			if size < stackGrowLarge {
				writecode(UWOP_SAVE_NONVOL, REG_RBP)
				sctxt.write16(uint16(size / 8))
			} else {
				writecode(UWOP_SAVE_NONVOL_FAR, REG_RBP)
				sctxt.write16(uint16(size))
				sctxt.write16(uint16(size) << 2)
			}
		case SehOpSetFP:
			writecode(UWOP_SET_FPREG, 0)
		}
	}
	if ncodes%2 != 0 {
		// For alignment purposes, this array always has an even number of entries.
		sctxt.write16(0)
	}
	sctxt.write32(0) // Exception handler
}
