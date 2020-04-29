// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc64

import (
	"cmd/internal/objabi"
	"cmd/link/internal/ld"
	"cmd/link/internal/sym"
	"strings"
)

// Return the value of .TOC. for symbol s
func symtoc(syms *ld.ArchSyms, s *sym.Symbol) int64 {
	v := s.Version
	if s.Outer != nil {
		v = s.Outer.Version
	}

	toc := syms.DotTOC[v]
	if toc == nil {
		ld.Errorf(s, "TOC-relative relocation in object without .TOC.")
		return 0
	}

	return toc.Value
}

// archreloctoc relocates a TOC relative symbol.
// If the symbol pointed by this TOC relative symbol is in .data or .bss, the
// default load instruction can be changed to an addi instruction and the
// symbol address can be used directly.
// This code is for AIX only.
func archreloctoc(target *ld.Target, syms *ld.ArchSyms, r *sym.Reloc, s *sym.Symbol, val int64) int64 {
	if target.IsLinux() {
		ld.Errorf(s, "archrelocaddr called for %s relocation\n", r.Sym.Name)
	}
	var o1, o2 uint32

	o1 = uint32(val >> 32)
	o2 = uint32(val)

	var t int64
	useAddi := false
	const prefix = "TOC."
	var tarSym *sym.Symbol
	if strings.HasPrefix(r.Sym.Name, prefix) {
		tarSym = r.Sym.R[0].Sym
	} else {
		ld.Errorf(s, "archreloctoc called for a symbol without TOC anchor")
	}

	if target.IsInternal() && tarSym != nil && tarSym.Attr.Reachable() && (tarSym.Sect.Seg == &ld.Segdata) {
		t = ld.Symaddr(tarSym) + r.Add - syms.TOC.Value
		// change ld to addi in the second instruction
		o2 = (o2 & 0x03FF0000) | 0xE<<26
		useAddi = true
	} else {
		t = ld.Symaddr(r.Sym) + r.Add - syms.TOC.Value
	}

	if t != int64(int32(t)) {
		ld.Errorf(s, "TOC relocation for %s is too big to relocate %s: 0x%x", s.Name, r.Sym, t)
	}

	if t&0x8000 != 0 {
		t += 0x10000
	}

	o1 |= uint32((t >> 16) & 0xFFFF)

	switch r.Type {
	case objabi.R_ADDRPOWER_TOCREL_DS:
		if useAddi {
			o2 |= uint32(t) & 0xFFFF
		} else {
			if t&3 != 0 {
				ld.Errorf(s, "bad DS reloc for %s: %d", s.Name, ld.Symaddr(r.Sym))
			}
			o2 |= uint32(t) & 0xFFFC
		}
	default:
		return -1
	}

	return int64(o1)<<32 | int64(o2)
}

// archrelocaddr relocates a symbol address.
// This code is for AIX only.
func archrelocaddr(target *ld.Target, syms *ld.ArchSyms, r *sym.Reloc, s *sym.Symbol, val int64) int64 {
	if target.IsAIX() {
		ld.Errorf(s, "archrelocaddr called for %s relocation\n", r.Sym.Name)
	}
	var o1, o2 uint32
	if target.IsBigEndian() {
		o1 = uint32(val >> 32)
		o2 = uint32(val)
	} else {
		o1 = uint32(val)
		o2 = uint32(val >> 32)
	}

	// We are spreading a 31-bit address across two instructions, putting the
	// high (adjusted) part in the low 16 bits of the first instruction and the
	// low part in the low 16 bits of the second instruction, or, in the DS case,
	// bits 15-2 (inclusive) of the address into bits 15-2 of the second
	// instruction (it is an error in this case if the low 2 bits of the address
	// are non-zero).

	t := ld.Symaddr(r.Sym) + r.Add
	if t < 0 || t >= 1<<31 {
		ld.Errorf(s, "relocation for %s is too big (>=2G): 0x%x", s.Name, ld.Symaddr(r.Sym))
	}
	if t&0x8000 != 0 {
		t += 0x10000
	}

	switch r.Type {
	case objabi.R_ADDRPOWER:
		o1 |= (uint32(t) >> 16) & 0xffff
		o2 |= uint32(t) & 0xffff
	case objabi.R_ADDRPOWER_DS:
		o1 |= (uint32(t) >> 16) & 0xffff
		if t&3 != 0 {
			ld.Errorf(s, "bad DS reloc for %s: %d", s.Name, ld.Symaddr(r.Sym))
		}
		o2 |= uint32(t) & 0xfffc
	default:
		return -1
	}

	if target.IsBigEndian() {
		return int64(o1)<<32 | int64(o2)
	}
	return int64(o2)<<32 | int64(o1)
}

func archreloc(target *ld.Target, syms *ld.ArchSyms, r *sym.Reloc, s *sym.Symbol, val int64) (int64, bool) {
	if target.IsExternal() {
		// On AIX, relocations (except TLS ones) must be also done to the
		// value with the current addresses.
		switch r.Type {
		default:
			if target.IsAIX() {
				return val, false
			}
		case objabi.R_POWER_TLS, objabi.R_POWER_TLS_LE, objabi.R_POWER_TLS_IE:
			r.Done = false
			// check Outer is nil, Type is TLSBSS?
			r.Xadd = r.Add
			r.Xsym = r.Sym
			return val, true
		case objabi.R_ADDRPOWER,
			objabi.R_ADDRPOWER_DS,
			objabi.R_ADDRPOWER_TOCREL,
			objabi.R_ADDRPOWER_TOCREL_DS,
			objabi.R_ADDRPOWER_GOT,
			objabi.R_ADDRPOWER_PCREL:
			r.Done = false

			// set up addend for eventual relocation via outer symbol.
			rs := r.Sym
			r.Xadd = r.Add
			for rs.Outer != nil {
				r.Xadd += ld.Symaddr(rs) - ld.Symaddr(rs.Outer)
				rs = rs.Outer
			}

			if rs.Type != sym.SHOSTOBJ && rs.Type != sym.SDYNIMPORT && rs.Type != sym.SUNDEFEXT && rs.Sect == nil {
				ld.Errorf(s, "missing section for %s", rs.Name)
			}
			r.Xsym = rs

			if !target.IsAIX() {
				return val, true
			}
		case objabi.R_CALLPOWER:
			r.Done = false
			r.Xsym = r.Sym
			r.Xadd = r.Add
			if !target.IsAIX() {
				return val, true
			}
		}
	}

	switch r.Type {
	case objabi.R_CONST:
		return r.Add, true
	case objabi.R_GOTOFF:
		return ld.Symaddr(r.Sym) + r.Add - ld.Symaddr(syms.GOT), true
	case objabi.R_ADDRPOWER_TOCREL, objabi.R_ADDRPOWER_TOCREL_DS:
		return archreloctoc(target, syms, r, s, val), true
	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS:
		return archrelocaddr(target, syms, r, s, val), true
	case objabi.R_CALLPOWER:
		// Bits 6 through 29 = (S + A - P) >> 2

		t := ld.Symaddr(r.Sym) + r.Add - (s.Value + int64(r.Off))

		if t&3 != 0 {
			ld.Errorf(s, "relocation for %s+%d is not aligned: %d", r.Sym.Name, r.Off, t)
		}
		// If branch offset is too far then create a trampoline.

		if int64(int32(t<<6)>>6) != t {
			ld.Errorf(s, "direct call too far: %s %x", r.Sym.Name, t)
		}
		return val | int64(uint32(t)&^0xfc000003), true
	case objabi.R_POWER_TOC: // S + A - .TOC.
		return ld.Symaddr(r.Sym) + r.Add - symtoc(syms, s), true

	case objabi.R_POWER_TLS_LE:
		// The thread pointer points 0x7000 bytes after the start of the
		// thread local storage area as documented in section "3.7.2 TLS
		// Runtime Handling" of "Power Architecture 64-Bit ELF V2 ABI
		// Specification".
		v := r.Sym.Value - 0x7000
		if target.IsAIX() {
			// On AIX, the thread pointer points 0x7800 bytes after
			// the TLS.
			v -= 0x800
		}
		if int64(int16(v)) != v {
			ld.Errorf(s, "TLS offset out of range %d", v)
		}
		return (val &^ 0xffff) | (v & 0xffff), true
	}

	return val, false
}
