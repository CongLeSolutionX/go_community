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
)

type sehUnwindCode struct {
	Prog      *Prog
	Operation uint8
}

func (ctxt *Link) HasSeh() bool {
	return ctxt.Headtype == objabi.Hwindows && ctxt.Arch.Family == sys.AMD64
}

func (fi *FuncInfo) AddSehUnwindCode(p *Prog, op uint8) {
	fi.sehUnwindInfo = append(fi.sehUnwindInfo, sehUnwindCode{p, op})
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

func (ctxt *Link) populateSehamd64(fn *FuncInfo) {
	if len(fn.sehUnwindInfo) == 0 {
		return
	}
	const (
		UWOP_ALLOC_LARGE = 1
		UWOP_ALLOC_SMALL = 2
	)
	fn.sehUnwindInfoSym = &LSym{
		Type: objabi.SSEHUNWINDINFO,
	}
	var off int64
	write8 := func(v uint8) {
		fn.sehUnwindInfoSym.WriteInt(ctxt, off, 1, int64(v))
		off += 1
	}
	write16 := func(v uint16) {
		fn.sehUnwindInfoSym.WriteInt(ctxt, off, 2, int64(v))
		off += 2
	}
	write32 := func(v uint32) {
		fn.sehUnwindInfoSym.WriteInt(ctxt, off, 4, int64(v))
		off += 4
	}
	writecode := func(op, value uint8) {
		write8(value<<4 | op)
	}
	write8(0b00000001)                                                    // Flags + version
	write8(uint8(fn.sehUnwindInfo[len(fn.sehUnwindInfo)-1].Prog.Link.Pc)) // Size of prolog
	write8(uint8(len(fn.sehUnwindInfo) / 2))                              // Count of unwind codes
	write8(0)                                                             // FP
	for i := len(fn.sehUnwindInfo) - 1; i >= 0; i-- {
		p := fn.sehUnwindInfo[i].Prog
		write8(uint8(p.Link.Pc))
		switch fn.sehUnwindInfo[i].Operation {
		case SehOpStackGrow:
			size := p.From.Offset
			if size < 136 {
				writecode(UWOP_ALLOC_SMALL, uint8(size/8)-1)
			} else if size < 512*1000 {
				writecode(UWOP_ALLOC_LARGE, 0)
				write16(uint16(size / 8))
			} else {
				writecode(UWOP_ALLOC_LARGE, 1)
				write16(uint16(size))
				write16(uint16(size << 2))
			}
		default:
			ctxt.Diag("unsupported SEH operation %d", fn.sehUnwindInfo[i].Operation)
		}
	}
	if len(fn.sehUnwindInfo)%2 != 0 {
		// For alignment purposes, this array always has an even number of entries.
		write16(0)
	}
	write32(0) // Exception handler
}
