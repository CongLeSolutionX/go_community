// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parsing of Go intermediate object files and archives.

package objfile

import (
	"cmd/internal/goobj"
	"debug/dwarf"
	"debug/gosym"
	"errors"
	"fmt"
	"os"
)

type goobjFile struct {
	goobj *goobj.Package
	f     *os.File
	m     map[uint64]*goobj.Sym // map from addr to goobj symbol

	// cached pcln tables
	s     *Sym
	line  *gosym.LineTable
	fidx  *gosym.LineTable
	fname []string
}

func openGoobj(r *os.File) (rawFile, error) {
	f, err := goobj.Parse(r, `""`)
	if err != nil {
		return nil, err
	}
	return &goobjFile{goobj: f, f: r, m: map[uint64]*goobj.Sym{}}, nil
}

func goobjName(id goobj.SymID) string {
	if id.Version == 0 {
		return id.Name
	}
	return fmt.Sprintf("%s<%d>", id.Name, id.Version)
}

func (f *goobjFile) symbols() ([]Sym, error) {
	seen := make(map[goobj.SymID]bool)

	var syms []Sym
	for _, s := range f.goobj.Syms {
		seen[s.SymID] = true
		sym := Sym{Addr: uint64(s.Data.Offset), Name: goobjName(s.SymID), Size: int64(s.Size), Type: s.Type.Name, Code: '?'}
		switch s.Kind {
		case goobj.STEXT, goobj.SELFRXSECT:
			sym.Code = 'T'
			f.m[sym.Addr] = s
		case goobj.STYPE, goobj.SSTRING, goobj.SGOSTRING, goobj.SGOFUNC, goobj.SRODATA, goobj.SFUNCTAB, goobj.STYPELINK, goobj.SITABLINK, goobj.SSYMTAB, goobj.SPCLNTAB, goobj.SELFROSECT:
			sym.Code = 'R'
		case goobj.SMACHOPLT, goobj.SELFSECT, goobj.SMACHO, goobj.SMACHOGOT, goobj.SNOPTRDATA, goobj.SINITARR, goobj.SDATA, goobj.SWINDOWS:
			sym.Code = 'D'
		case goobj.SBSS, goobj.SNOPTRBSS, goobj.STLSBSS:
			sym.Code = 'B'
		case goobj.SXREF, goobj.SMACHOSYMSTR, goobj.SMACHOSYMTAB, goobj.SMACHOINDIRECTPLT, goobj.SMACHOINDIRECTGOT, goobj.SFILE, goobj.SFILEPATH, goobj.SCONST, goobj.SDYNIMPORT, goobj.SHOSTOBJ:
			sym.Code = 'X' // should not see
		}
		if s.Version != 0 {
			sym.Code += 'a' - 'A'
		}
		syms = append(syms, sym)
	}

	for _, s := range f.goobj.Syms {
		for _, r := range s.Reloc {
			if !seen[r.Sym] {
				seen[r.Sym] = true
				sym := Sym{Name: goobjName(r.Sym), Code: 'U'}
				if s.Version != 0 {
					// should not happen but handle anyway
					sym.Code = 'u'
				}
				syms = append(syms, sym)
			}
		}
	}

	return syms, nil
}

func (f *goobjFile) goarch() string {
	return f.goobj.Arch
}

func (f *goobjFile) getText(s *Sym) ([]byte, error) {
	text := make([]byte, s.Size)
	_, err := f.f.ReadAt(text, int64(s.Addr))
	return text, err
}

func (f *goobjFile) loadAddress() (uint64, error) {
	return 0, fmt.Errorf("unknown load address")
}

func (f *goobjFile) dwarf() (*dwarf.Data, error) {
	return nil, errors.New("no DWARF data in go object file")
}

func (f *goobjFile) pc2line(s *Sym, addr uint64) (string, int) {
	r := f.m[s.Addr]
	if r == nil || r.Func == nil {
		return "", 0
	}
	lineData := make([]byte, r.Func.PCLine.Size)
	f.f.ReadAt(lineData, r.Func.PCLine.Offset)
	nameIdxData := make([]byte, r.Func.PCFile.Size)
	f.f.ReadAt(nameIdxData, r.Func.PCFile.Offset)
	// TODO: cache these reads

	return r.Func.File[f.pcvalue(nameIdxData, 0, s.Addr, addr)], int(f.pcvalue(lineData, 0, s.Addr, addr))
}

// pcvalue reports the value associated with the target pc.
// off is the offset to the beginning of the pc-value table,
// and entry is the start PC for the corresponding function.
func (f *goobjFile) pcvalue(t []byte, off uint32, entry, targetpc uint64) int32 {
	val := int32(-1)
	pc := entry
	for f.step(&t, &pc, &val, pc == entry) {
		if targetpc < pc {
			return val
		}
	}
	return -1
}

// step advances to the next pc, value pair in the encoded table.
func (f *goobjFile) step(p *[]byte, pc *uint64, val *int32, first bool) bool {
	uvdelta := f.readvarint(p)
	if uvdelta == 0 && !first {
		return false
	}
	if uvdelta&1 != 0 {
		uvdelta = ^(uvdelta >> 1)
	} else {
		uvdelta >>= 1
	}
	vdelta := int32(uvdelta)
	var quantum uint32
	switch f.goobj.Arch {
	case "amd64", "386":
		quantum = 1
	default:
		quantum = 4
	}
	pcdelta := f.readvarint(p) * quantum
	*pc += uint64(pcdelta)
	*val += vdelta
	return true
}

// readvarint reads, removes, and returns a varint from *pp.
func (f *goobjFile) readvarint(pp *[]byte) uint32 {
	var v, shift uint32
	p := *pp
	for shift = 0; ; shift += 7 {
		b := p[0]
		p = p[1:]
		v |= (uint32(b) & 0x7F) << shift
		if b&0x80 == 0 {
			break
		}
	}
	*pp = p
	return v
}

func (f *goobjFile) relocs(s *Sym) []goobj.Reloc {
	r := f.m[s.Addr]
	if r == nil {
		return nil
	}
	return r.Reloc
}
