// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parsing of Mach-O executables (OS X).

package objfile

import (
	"cmd/internal/goobj"
	"debug/dwarf"
	"debug/gosym"
	"debug/macho"
	"fmt"
	"os"
	"sort"
)

const stabTypeMask = 0xe0

type machoFile struct {
	macho *macho.File
	pcln  *gosym.Table
}

func openMacho(r *os.File) (rawFile, error) {
	f, err := macho.NewFile(r)
	if err != nil {
		return nil, err
	}
	return &machoFile{macho: f, pcln: nil}, nil
}

func (f *machoFile) symbols() ([]Sym, error) {
	if f.macho.Symtab == nil {
		return nil, fmt.Errorf("missing symbol table")
	}

	// Build sorted list of addresses of all symbols.
	// We infer the size of a symbol by looking at where the next symbol begins.
	var addrs []uint64
	for _, s := range f.macho.Symtab.Syms {
		// Skip stab debug info.
		if s.Type&stabTypeMask == 0 {
			addrs = append(addrs, s.Value)
		}
	}
	sort.Sort(uint64s(addrs))

	var syms []Sym
	for _, s := range f.macho.Symtab.Syms {
		if s.Type&stabTypeMask != 0 {
			// Skip stab debug info.
			continue
		}
		sym := Sym{Name: s.Name, Addr: s.Value, Code: '?'}
		i := sort.Search(len(addrs), func(x int) bool { return addrs[x] > s.Value })
		if i < len(addrs) {
			sym.Size = int64(addrs[i] - s.Value)
		}
		if s.Sect == 0 {
			sym.Code = 'U'
		} else if int(s.Sect) <= len(f.macho.Sections) {
			sect := f.macho.Sections[s.Sect-1]
			switch sect.Seg {
			case "__TEXT":
				sym.Code = 'R'
			case "__DATA":
				sym.Code = 'D'
			}
			switch sect.Seg + " " + sect.Name {
			case "__TEXT __text":
				sym.Code = 'T'
			case "__DATA __bss", "__DATA __noptrbss":
				sym.Code = 'B'
			}
		}
		syms = append(syms, sym)
	}

	return syms, nil
}

// getText returns the text (instruction bytes) for s.
func (f *machoFile) getText(s *Sym) ([]byte, error) {
	sect := f.macho.Section("__text")
	if sect == nil {
		return nil, fmt.Errorf("text section not found")
	}
	text := make([]byte, s.Size)
	_, err := sect.ReadAt(text, int64(s.Addr-sect.Addr))
	return text, err
}

func (f *machoFile) goarch() string {
	switch f.macho.Cpu {
	case macho.Cpu386:
		return "386"
	case macho.CpuAmd64:
		return "amd64"
	case macho.CpuArm:
		return "arm"
	case macho.CpuPpc64:
		return "ppc64"
	}
	return ""
}

type uint64s []uint64

func (x uint64s) Len() int           { return len(x) }
func (x uint64s) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x uint64s) Less(i, j int) bool { return x[i] < x[j] }

func (f *machoFile) loadAddress() (uint64, error) {
	return 0, fmt.Errorf("unknown load address")
}

func (f *machoFile) dwarf() (*dwarf.Data, error) {
	return f.macho.DWARF()
}

func (f *machoFile) pc2line(s *Sym, pc uint64) (string, int) {
	if f.pcln == nil {
		var textStart uint64
		var symtab []byte
		var pclntab []byte
		var err error
		if sect := f.macho.Section("__text"); sect != nil {
			textStart = sect.Addr
		}
		if sect := f.macho.Section("__gosymtab"); sect != nil {
			if symtab, err = sect.Data(); err != nil {
				return "unknown", 0
			}
		}
		if sect := f.macho.Section("__gopclntab"); sect != nil {
			if pclntab, err = sect.Data(); err != nil {
				return "unknown", 0
			}
		}
		f.pcln, err = gosym.NewTable(symtab, gosym.NewLineTable(pclntab, textStart))
		if err != nil {
			return "unknown", 0
		}
	}
	file, line, _ := f.pcln.PCToLine(pc)
	return file, line
}

func (f *machoFile) relocs(s *Sym) []goobj.Reloc {
	return nil
}
