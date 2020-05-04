// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"cmd/internal/objabi"
	"cmd/link/internal/ld"
	"cmd/link/internal/sym"
	"log"
)

// Temporary dumping around for sym.Symbol version of helper
// functions in arm64/asm.go, still being used for some archs/os flavors.

func archreloc(target *ld.Target, syms *ld.ArchSyms, r *sym.Reloc, s *sym.Symbol, val int64) (int64, bool) {
	if target.IsExternal() {
		switch r.Type {
		default:
			return val, false
		case objabi.R_ARM64_GOTPCREL:
			var o1, o2 uint32
			if target.IsBigEndian() {
				o1 = uint32(val >> 32)
				o2 = uint32(val)
			} else {
				o1 = uint32(val)
				o2 = uint32(val >> 32)
			}
			// Any relocation against a function symbol is redirected to
			// be against a local symbol instead (see putelfsym in
			// symtab.go) but unfortunately the system linker was buggy
			// when confronted with a R_AARCH64_ADR_GOT_PAGE relocation
			// against a local symbol until May 2015
			// (https://sourceware.org/bugzilla/show_bug.cgi?id=18270). So
			// we convert the adrp; ld64 + R_ARM64_GOTPCREL into adrp;
			// add + R_ADDRARM64.
			if !(r.Sym.IsFileLocal() || r.Sym.Attr.VisibilityHidden() || r.Sym.Attr.Local()) && r.Sym.Type == sym.STEXT && target.IsDynlinkingGo() {
				if o2&0xffc00000 != 0xf9400000 {
					ld.Errorf(s, "R_ARM64_GOTPCREL against unexpected instruction %x", o2)
				}
				o2 = 0x91000000 | (o2 & 0x000003ff)
				r.Type = objabi.R_ADDRARM64
			}
			if target.IsBigEndian() {
				val = int64(o1)<<32 | int64(o2)
			} else {
				val = int64(o2)<<32 | int64(o1)
			}
			fallthrough
		case objabi.R_ADDRARM64:
			r.Done = false

			// set up addend for eventual relocation via outer symbol.
			rs := ld.ApplyOuterToXAdd(r)
			if rs.Type != sym.SHOSTOBJ && rs.Type != sym.SDYNIMPORT && rs.Sect == nil {
				ld.Errorf(s, "missing section for %s", rs.Name)
			}
			r.Xsym = rs

			// Note: ld64 currently has a bug that any non-zero addend for BR26 relocation
			// will make the linking fail because it thinks the code is not PIC even though
			// the BR26 relocation should be fully resolved at link time.
			// That is the reason why the next if block is disabled. When the bug in ld64
			// is fixed, we can enable this block and also enable duff's device in cmd/7g.
			if false && target.IsDarwin() {
				var o0, o1 uint32

				if target.IsBigEndian() {
					o0 = uint32(val >> 32)
					o1 = uint32(val)
				} else {
					o0 = uint32(val)
					o1 = uint32(val >> 32)
				}
				// Mach-O wants the addend to be encoded in the instruction
				// Note that although Mach-O supports ARM64_RELOC_ADDEND, it
				// can only encode 24-bit of signed addend, but the instructions
				// supports 33-bit of signed addend, so we always encode the
				// addend in place.
				o0 |= (uint32((r.Xadd>>12)&3) << 29) | (uint32((r.Xadd>>12>>2)&0x7ffff) << 5)
				o1 |= uint32(r.Xadd&0xfff) << 10
				r.Xadd = 0

				// when laid out, the instruction order must always be o1, o2.
				if target.IsBigEndian() {
					val = int64(o0)<<32 | int64(o1)
				} else {
					val = int64(o1)<<32 | int64(o0)
				}
			}

			return val, true
		case objabi.R_CALLARM64,
			objabi.R_ARM64_TLS_LE,
			objabi.R_ARM64_TLS_IE:
			r.Done = false
			r.Xsym = r.Sym
			r.Xadd = r.Add
			return val, true
		}
	}

	switch r.Type {
	case objabi.R_CONST:
		return r.Add, true

	case objabi.R_GOTOFF:
		return ld.Symaddr(r.Sym) + r.Add - ld.Symaddr(syms.GOT), true

	case objabi.R_ADDRARM64:
		t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
		if t >= 1<<32 || t < -1<<32 {
			ld.Errorf(s, "program too large, address relocation distance = %d", t)
		}

		var o0, o1 uint32

		if target.IsBigEndian() {
			o0 = uint32(val >> 32)
			o1 = uint32(val)
		} else {
			o0 = uint32(val)
			o1 = uint32(val >> 32)
		}

		o0 |= (uint32((t>>12)&3) << 29) | (uint32((t>>12>>2)&0x7ffff) << 5)
		o1 |= uint32(t&0xfff) << 10

		// when laid out, the instruction order must always be o1, o2.
		if target.IsBigEndian() {
			return int64(o0)<<32 | int64(o1), true
		}
		return int64(o1)<<32 | int64(o0), true

	case objabi.R_ARM64_TLS_LE:
		r.Done = false
		if target.IsDarwin() {
			ld.Errorf(s, "TLS reloc on unsupported OS %v", target.HeadType)
		}
		// The TCB is two pointers. This is not documented anywhere, but is
		// de facto part of the ABI.
		v := r.Sym.Value + int64(2*target.Arch.PtrSize)
		if v < 0 || v >= 32678 {
			ld.Errorf(s, "TLS offset out of range %d", v)
		}
		return val | (v << 5), true

	case objabi.R_ARM64_TLS_IE:
		if target.IsPIE() && target.IsElf() {
			// We are linking the final executable, so we
			// can optimize any TLS IE relocation to LE.
			r.Done = false
			if !target.IsLinux() {
				ld.Errorf(s, "TLS reloc on unsupported OS %v", target.HeadType)
			}

			// The TCB is two pointers. This is not documented anywhere, but is
			// de facto part of the ABI.
			v := ld.Symaddr(r.Sym) + int64(2*target.Arch.PtrSize) + r.Add
			if v < 0 || v >= 32678 {
				ld.Errorf(s, "TLS offset out of range %d", v)
			}

			var o0, o1 uint32
			if target.IsBigEndian() {
				o0 = uint32(val >> 32)
				o1 = uint32(val)
			} else {
				o0 = uint32(val)
				o1 = uint32(val >> 32)
			}

			// R_AARCH64_TLSIE_ADR_GOTTPREL_PAGE21
			// turn ADRP to MOVZ
			o0 = 0xd2a00000 | uint32(o0&0x1f) | (uint32((v>>16)&0xffff) << 5)
			// R_AARCH64_TLSIE_LD64_GOTTPREL_LO12_NC
			// turn LD64 to MOVK
			if v&3 != 0 {
				ld.Errorf(s, "invalid address: %x for relocation type: R_AARCH64_TLSIE_LD64_GOTTPREL_LO12_NC", v)
			}
			o1 = 0xf2800000 | uint32(o1&0x1f) | (uint32(v&0xffff) << 5)

			// when laid out, the instruction order must always be o0, o1.
			if target.IsBigEndian() {
				return int64(o0)<<32 | int64(o1), true
			}
			return int64(o1)<<32 | int64(o0), true
		} else {
			log.Fatalf("cannot handle R_ARM64_TLS_IE (sym %s) when linking internally", s.Name)
		}

	case objabi.R_CALLARM64:
		var t int64
		if r.Sym.Type == sym.SDYNIMPORT {
			t = (ld.Symaddr(syms.PLT) + r.Add) - (s.Value + int64(r.Off))
		} else {
			t = (ld.Symaddr(r.Sym) + r.Add) - (s.Value + int64(r.Off))
		}
		if t >= 1<<27 || t < -1<<27 {
			ld.Errorf(s, "program too large, call relocation distance = %d", t)
		}
		return val | ((t >> 2) & 0x03ffffff), true

	case objabi.R_ARM64_GOT:
		if s.P[r.Off+3]&0x9f == 0x90 {
			// R_AARCH64_ADR_GOT_PAGE
			// patch instruction: adrp
			t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
			if t >= 1<<32 || t < -1<<32 {
				ld.Errorf(s, "program too large, address relocation distance = %d", t)
			}
			var o0 uint32
			o0 |= (uint32((t>>12)&3) << 29) | (uint32((t>>12>>2)&0x7ffff) << 5)
			return val | int64(o0), true
		} else if s.P[r.Off+3] == 0xf9 {
			// R_AARCH64_LD64_GOT_LO12_NC
			// patch instruction: ldr
			t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
			if t&7 != 0 {
				ld.Errorf(s, "invalid address: %x for relocation type: R_AARCH64_LD64_GOT_LO12_NC", t)
			}
			var o1 uint32
			o1 |= uint32(t&0xfff) << (10 - 3)
			return val | int64(uint64(o1)), true
		} else {
			ld.Errorf(s, "unsupported instruction for %v R_GOTARM64", s.P[r.Off:r.Off+4])
		}

	case objabi.R_ARM64_PCREL:
		if s.P[r.Off+3]&0x9f == 0x90 {
			// R_AARCH64_ADR_PREL_PG_HI21
			// patch instruction: adrp
			t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
			if t >= 1<<32 || t < -1<<32 {
				ld.Errorf(s, "program too large, address relocation distance = %d", t)
			}
			o0 := (uint32((t>>12)&3) << 29) | (uint32((t>>12>>2)&0x7ffff) << 5)
			return val | int64(o0), true
		} else if s.P[r.Off+3]&0x91 == 0x91 {
			// R_AARCH64_ADD_ABS_LO12_NC
			// patch instruction: add
			t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
			o1 := uint32(t&0xfff) << 10
			return val | int64(o1), true
		} else {
			ld.Errorf(s, "unsupported instruction for %v R_PCRELARM64", s.P[r.Off:r.Off+4])
		}

	case objabi.R_ARM64_LDST8:
		t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
		o0 := uint32(t&0xfff) << 10
		return val | int64(o0), true

	case objabi.R_ARM64_LDST32:
		t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
		if t&3 != 0 {
			ld.Errorf(s, "invalid address: %x for relocation type: R_AARCH64_LDST32_ABS_LO12_NC", t)
		}
		o0 := (uint32(t&0xfff) >> 2) << 10
		return val | int64(o0), true

	case objabi.R_ARM64_LDST64:
		t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
		if t&7 != 0 {
			ld.Errorf(s, "invalid address: %x for relocation type: R_AARCH64_LDST64_ABS_LO12_NC", t)
		}
		o0 := (uint32(t&0xfff) >> 3) << 10
		return val | int64(o0), true

	case objabi.R_ARM64_LDST128:
		t := ld.Symaddr(r.Sym) + r.Add - ((s.Value + int64(r.Off)) &^ 0xfff)
		if t&15 != 0 {
			ld.Errorf(s, "invalid address: %x for relocation type: R_AARCH64_LDST128_ABS_LO12_NC", t)
		}
		o0 := (uint32(t&0xfff) >> 4) << 10
		return val | int64(o0), true
	}

	return val, false
}
