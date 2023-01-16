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

// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#operation-info
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
	var ncodes uint8
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
		default:
			ctxt.Diag("unsupported SEH operation %d", c.Operation)
		}
	}
	s.GrowCap(int64(ncodes*2) + 9)
	sctxt.write8(1)                                   // Flags + version
	sctxt.write8(uint8(ops[len(ops)-1].Prog.Link.Pc)) // Size of prolog
	sctxt.write8(uint8(ncodes))                       // Count of unwind codes
	sctxt.write8(0)                                   // FP
	// Unwind codes are expected to appear in reverse order
	// of appearance, so traverse ops in reverse order.
	// The encoding of each operation is documented here:
	// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#struct-unwind_code
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i].Operation
		p := ops[i].Prog
		// Offset in prolog of the end of the instruction
		// that performs this operation, plus 1.
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
		}
	}
	if ncodes%2 != 0 {
		// For alignment purposes, this array always has an even number of entries.
		sctxt.write16(0)
	}
	sctxt.write32(0) // Exception handler
}
