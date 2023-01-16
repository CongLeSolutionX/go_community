// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
)

const (
	sehOpStackGrow = iota
	sehOpSaveReg
	sehOpSetFP
)

type sehctxt struct {
	ctxt *obj.Link
	s    *obj.LSym
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

var sehRegisters = map[int16]uint8{
	REG_AX:  0,
	REG_CX:  1,
	REG_DX:  2,
	REG_BX:  3,
	REG_SP:  4,
	REG_BP:  5,
	REG_SI:  6,
	REG_DI:  7,
	REG_R8:  8,
	REG_R9:  9,
	REG_R10: 10,
	REG_R11: 11,
	REG_R12: 12,
	REG_R13: 13,
	REG_R14: 14,
	REG_R15: 15,
}

// populateSeh encodes the SEH unwind operations of the symbol s.
func populateSeh(ctxt *obj.Link, s *obj.LSym) {
	fn := s.Func()
	ops := fn.SehUnwindOps
	if len(ops) == 0 {
		return
	}
	const (
		// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#unwind-operation-code
		UWOP_ALLOC_LARGE     = 1
		UWOP_ALLOC_SMALL     = 2
		UWOP_SET_FPREG       = 3
		UWOP_SAVE_NONVOL     = 4
		UWOP_SAVE_NONVOL_FAR = 5
	)
	const (
		stackGrowSmall = 136
		stackGrowLarge = 512 * 1000
	)
	fn.SehUnwindInfoSym = &obj.LSym{
		Type: objabi.SSEHUNWINDINFO,
	}
	sctxt := sehctxt{ctxt: ctxt, s: fn.SehUnwindInfoSym}
	writecode := func(op, value uint8) {
		sctxt.write8(value<<4 | op)
	}
	// We need to write the number of codes before encoding each one.
	// Each code can be encoded as 1, 2 or 3 entries.
	var ncodes, fpoffset uint8
	var hasfp bool
	for _, c := range ops {
		p := c.Prog
		switch c.Operation {
		case sehOpStackGrow:
			size := p.From.Offset
			if size < stackGrowSmall {
				ncodes += 1
			} else if size < stackGrowLarge {
				ncodes += 2
			} else {
				ncodes += 3
			}
		case sehOpSaveReg:
			size := p.To.Offset
			if size < stackGrowLarge {
				ncodes += 2
			} else if size < stackGrowLarge {
				ncodes += 3
			}
		case sehOpSetFP:
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
	s.GrowCap(int64(ncodes*2) + 9)
	sctxt.write8(1)                                   // Flags + version
	sctxt.write8(uint8(ops[len(ops)-1].Prog.Link.Pc)) // Size of prolog
	sctxt.write8(uint8(ncodes))                       // Count of unwind codes
	if hasfp {
		sctxt.write8((fpoffset/16)<<4 | sehRegisters[REG_BP])
	} else {
		sctxt.write8(0)
	}
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i].Operation
		if op == sehOpSetFP && !hasfp {
			continue
		}
		p := ops[i].Prog
		sctxt.write8(uint8(p.Link.Pc))
		switch op {
		case sehOpStackGrow:
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
		case sehOpSaveReg:
			size := p.To.Offset
			reg := sehRegisters[p.From.Reg]
			if size < stackGrowLarge {
				writecode(UWOP_SAVE_NONVOL, reg)
				sctxt.write16(uint16(size / 8))
			} else {
				writecode(UWOP_SAVE_NONVOL_FAR, reg)
				sctxt.write16(uint16(size))
				sctxt.write16(uint16(size) << 2)
			}
		case sehOpSetFP:
			writecode(UWOP_SET_FPREG, 0)
		}
	}
	if ncodes%2 != 0 {
		// For alignment purposes, this array always has an even number of entries.
		sctxt.write16(0)
	}
	sctxt.write32(0) // Exception handler
}
