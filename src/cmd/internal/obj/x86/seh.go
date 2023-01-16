// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
)

const (
	sehOpSaveReg = iota
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

// populateSeh encodes the SEH unwind information for the symbol s.
func populateSeh(ctxt *obj.Link, s *obj.LSym) {
	fn := s.Func()
	ops := fn.SehUnwindOps
	if len(ops) == 0 {
		return
	}
	// We don't need to support all unwind operation,
	// only the necessary to save and set the frame pointer.
	// Stack allocation operations are not supported on purpose,
	// they take space and setting the frame pointer is enough.
	const (
		// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#unwind-operation-code
		UWOP_PUSH_NONVOL = 0
		UWOP_SET_FPREG   = 3
	)
	fn.SehUnwindInfoSym = &obj.LSym{
		Type: objabi.SSEHUNWINDINFO,
	}
	sctxt := sehctxt{ctxt: ctxt, s: fn.SehUnwindInfoSym}
	writecode := func(op, value uint8) {
		sctxt.write8(value<<4 | op)
	}
	var (
		fpoffset uint8
		fpreg    int16
		hasfp    bool
	)
	// Traverse ops looking for a sehOpSetFP
	// and validate the operations.
	for _, c := range ops {
		p := c.Prog
		switch c.Operation {
		case sehOpSaveReg:
			if p.As != APUSHQ {
				ctxt.Diag("unsupported registry save instruction\n%v", p)
			}
		case sehOpSetFP:
			fpoffset = uint8(p.From.Offset)
			if fpoffset >= 240 || fpoffset%16 != 0 {
				ctxt.Diag("unsupported frame pointer offset %d", fpoffset)
			}
			fpreg = p.To.Reg
			hasfp = true
		default:
			ctxt.Diag("unsupported SEH operation %d", c.Operation)
		}
	}
	if ctxt.Errors > 0 {
		return
	}
	s.GrowCap(int64(8 + len(ops)*2))                  // preallocate 8 bytes for the header and 2 bytes per operation
	sctxt.write8(1)                                   // Flags + version
	sctxt.write8(uint8(ops[len(ops)-1].Prog.Link.Pc)) // Size of prolog
	sctxt.write8(uint8(len(ops)))                     // Count of nodes
	if hasfp {
		sctxt.write8((fpoffset/16)<<4 | sehRegisters[fpreg])
	} else {
		sctxt.write8(0)
	}
	// Unwind operations are expected to appear in reverse order
	// of appearance, so traverse ops in reverse order.
	// The encoding of each operation is documented here:
	// https://learn.microsoft.com/en-us/cpp/build/exception-handling-x64#struct-unwind_code
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i].Operation
		p := ops[i].Prog
		// Offset in prolog of the end of the instruction
		// that performs this operation, plus 1.
		// This is equivalent to the offset of the next
		// instruction.
		sctxt.write8(uint8(p.Link.Pc))
		switch op {
		case sehOpSaveReg:
			writecode(UWOP_PUSH_NONVOL, sehRegisters[p.From.Reg])
		case sehOpSetFP:
			writecode(UWOP_SET_FPREG, 0)
		}
	}
	if len(ops)%2 != 0 {
		// For alignment purposes, this array always has an even number of entries.
		sctxt.write16(0)
	}
	// The following 4 bytes reference the RVA of the exception handler,
	// in case the function has one. We don't use it for now.
	sctxt.write32(0)
}
