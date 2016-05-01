// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package objfile implements portable access to OS-specific executable files.
package objfile

import (
	"cmd/internal/goobj"
	"debug/dwarf"
	"fmt"
	"os"
	"sort"
)

type rawFile interface {
	symbols() (syms []Sym, err error)
	goarch() string
	loadAddress() (uint64, error)
	dwarf() (*dwarf.Data, error)

	// getText returns the text (instruction bytes) for s.
	getText(s *Sym) ([]byte, error)

	// pc2line returns file/line information for a pc in s.
	// Returns "unknown",0 if unknown.
	pc2line(s *Sym, pc uint64) (string, int)

	// relocs returns relocations for s.
	// Returned relocations are sorted in increasing Offset order.
	relocs(s *Sym) []goobj.Reloc
}

// A File is an opened executable file.
type File struct {
	r    *os.File
	raw  rawFile
	syms []Sym
}

// A Sym is a symbol defined in an executable file.
type Sym struct {
	Name string // symbol name
	Addr uint64 // virtual address of symbol
	Size int64  // size in bytes
	Code rune   // nm code (T for text, D for data, and so on)
	Type string // XXX?
}

var openers = []func(*os.File) (rawFile, error){
	openElf,
	openGoobj,
	openMacho,
	openPE,
	openPlan9,
}

// Open opens the named file.
// The caller must call f.Close when the file is no longer needed.
func Open(name string) (*File, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	for _, try := range openers {
		if raw, err := try(r); err == nil {
			syms, err := raw.symbols()
			if err != nil {
				return nil, err
			}
			sort.Sort(byAddr(syms))
			return &File{r, raw, syms}, nil
		}
	}
	r.Close()
	return nil, fmt.Errorf("open %s: unrecognized object file", name)
}

func (f *File) Close() error {
	return f.r.Close()
}

func (f *File) Symbols() ([]Sym, error) {
	return f.syms, nil
}

type byAddr []Sym

func (x byAddr) Less(i, j int) bool { return x[i].Addr < x[j].Addr }
func (x byAddr) Len() int           { return len(x) }
func (x byAddr) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func (f *File) PC2Sym(pc uint64) *Sym {
	// Binary search on sorted syms.
	i := sort.Search(len(f.syms), func(i int) bool {
		return pc < f.syms[i].Addr+uint64(f.syms[i].Size)
	})
	if i == len(f.syms) {
		return nil
	}
	s := &f.syms[i]
	if pc >= s.Addr {
		return s
	}
	return nil
}

func (f *File) PC2Line(s *Sym, pc uint64) (string, int) {
	return f.raw.pc2line(s, pc)
}

func (f *File) GetText(s *Sym) ([]byte, error) {
	return f.raw.getText(s)
}

func (f *File) Relocs(s *Sym) []goobj.Reloc {
	return f.raw.relocs(s)
}

func (f *File) GOARCH() string {
	return f.raw.goarch()
}

// LoadAddress returns the expected load address of the file.
// This differs from the actual load address for a position-independent
// executable.
func (f *File) LoadAddress() (uint64, error) {
	return f.raw.loadAddress()
}

// DWARF returns DWARF debug data for the file, if any.
// This is for cmd/pprof to locate cgo functions.
func (f *File) DWARF() (*dwarf.Data, error) {
	return f.raw.dwarf()
}
