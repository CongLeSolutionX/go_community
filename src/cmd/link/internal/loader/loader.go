// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package loader abstracts the object files from the linker, allowing the
// linker to just use global indices without needing to know where/how symbols
// have been defined.
package loader

import (
	"cmd/internal/bio"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type NameVer struct {
	Name string
	V    int
}

type objIdx struct {
	r *obj.Reader
	i int // start index
}

// A Loader loads new object files and resolves indexed symbol references.
//
// TODO: describe local-global index mapping.
type Loader struct {
	start map[*obj.Reader]int // map from object file to its start index
	objs  []objIdx            // sorted by start index (i.e. objIdx.i)
	max   int                 // current max index

	SymsByName map[NameVer]int // map symbol name to index

	Syms []*sym.Symbol // indexed symbols. XXX we still make sym.Symbol for now.
}

func NewLoader() *Loader {
	return &Loader{
		start:      make(map[*obj.Reader]int),
		objs:       []objIdx{{nil, 0}},
		SymsByName: make(map[NameVer]int),
		Syms:       []*sym.Symbol{nil},
	}
}

// Return the start index in the global index space for a given object file.
func (l *Loader) StartIndex(r *obj.Reader) int {
	return l.start[r]
}

// Add object file r, return the start index.
func (l *Loader) AddObj(r *obj.Reader) int {
	if _, ok := l.start[r]; ok {
		panic("already added")
	}
	n := r.NSym() + r.NNonpkgdef()
	i := l.max + 1
	l.start[r] = i
	l.objs = append(l.objs, objIdx{r, i})
	l.max += n
	return i
}

// Add a symbol with a given index, return if it is added.
func (l *Loader) AddSym(name string, ver int, i int, dupok bool) bool {
	nv := NameVer{name, ver}
	if _, ok := l.SymsByName[nv]; ok {
		if dupok || true { // TODO: "true" isn't quite right. need to implement "overwrite" logic.
			return false
		}
		panic("duplicated definition of symbol " + name)
	}
	l.SymsByName[nv] = i
	return true
}

// Add an external symbol (without index). Return the index of newly added
// symbol, or 0 if not added.
func (l *Loader) AddExtSym(name string, ver int) int {
	nv := NameVer{name, ver}
	if _, ok := l.SymsByName[nv]; ok {
		return 0
	}
	i := l.max + 1
	l.SymsByName[nv] = i
	l.max++
	return i
}

// Convert a local index to a global index.
func (l *Loader) ToGlobal(r *obj.Reader, i int) int {
	return l.StartIndex(r) + i
}

// Convert a global index to a global index. Is it useful?
func (l *Loader) ToLocal(i int) (*obj.Reader, int) {
	k := sort.Search(i, func(k int) bool {
		return l.objs[k].i >= i
	})
	if k == len(l.objs) {
		return nil, 0
	}
	return l.objs[k].r, i - l.objs[k].i
}

// Look up a symbol by name, return global index, or 0 if not found.
// This is more like Syms.ROLookup than Lookup -- it doesn't create
// new symbol.
func (l *Loader) Lookup(name string, ver int) int {
	nv := NameVer{name, ver}
	return l.SymsByName[nv]
}

// Preload a package: add autolibs, add symbols to the symbol table.
// Does not read symbol data yet.
func (l *Loader) LoadNew(arch *sys.Arch, syms *sym.Symbols, f *bio.Reader, lib *sym.Library, unit *sym.CompilationUnit, length int64, pn string, flags int) {
	//start := f.Offset()
	roObject, readonly, err := f.Slice(uint64(length))
	if err != nil {
		log.Fatal("cannot read object file:", err)
	}
	r := obj.NewReaderFromBytes(roObject, readonly)
	if r == nil {
		panic("cannot read object file")
	}
	localSymVersion := syms.IncVersion()
	lib.Readers = append(lib.Readers, struct {
		Reader  *obj.Reader
		Version int
	}{r, localSymVersion})

	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."

	// Autolib
	lib.ImportStrings = append(lib.ImportStrings, r.Pkglist()...)

	istart := l.AddObj(r)

	ndef := r.NSym()
	nnonpkgdef := r.NNonpkgdef()

	// XXX add all symbols for now
	l.Syms = append(l.Syms, make([]*sym.Symbol, ndef+nnonpkgdef)...)
	for i, n := 0, ndef+nnonpkgdef; i < n; i++ {
		osym := obj.OSym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		if name == "" {
			continue // don't add unnamed aux symbol
		}
		v := AbiToVer(osym.ABI, localSymVersion)
		dupok := osym.Flag&obj.SymFlagDupok != 0
		if l.AddSym(name, v, istart+i, dupok) {
			s := syms.Newsym(name, v)
			preprocess(arch, s) // TODO: put this at a better place
			l.Syms[istart+i] = s
		}
	}

	// The caller expects us consuming all the data
	f.MustSeek(length, os.SEEK_CUR)
}

// Make sure referenced symbols are added. Most of them should already be added.
// This should only be needed for referenced external symbols.
func (l *Loader) LoadRefs(r *obj.Reader, lib *sym.Library, arch *sys.Arch, syms *sym.Symbols, localSymVersion int) {
	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."
	ndef := r.NSym() + r.NNonpkgdef()
	for i, n := 0, r.NNonpkgref(); i < n; i++ {
		osym := obj.OSym{}
		osym.Read(r, r.SymOff(ndef+i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		v := AbiToVer(osym.ABI, localSymVersion)
		if ii := l.AddExtSym(name, v); ii != 0 {
			s := syms.Newsym(name, v)
			preprocess(arch, s) // TODO: put this at a better place
			if ii != len(l.Syms) {
				panic("AddExtSym returned bad index")
			}
			l.Syms = append(l.Syms, s)
		}
	}
}

func AbiToVer(abi uint16, localSymVersion int) int {
	var v int
	if abi == obj.SymABIstatic {
		// Static
		v = localSymVersion
	} else if abiver := sym.ABIToVersion(obj.ABI(abi)); abiver != -1 {
		// Note that data symbols are "ABI0", which maps to version 0.
		v = abiver
	} else {
		log.Fatalf("invalid symbol ABI: %d", abi)
	}
	return v
}

func preprocess(arch *sys.Arch, s *sym.Symbol) {
	if s.Name != "" && s.Name[0] == '$' && len(s.Name) > 5 && s.Type == 0 && len(s.P) == 0 {
		x, err := strconv.ParseUint(s.Name[5:], 16, 64)
		if err != nil {
			log.Panicf("failed to parse $-symbol %s: %v", s.Name, err)
		}
		s.Type = sym.SRODATA
		s.Attr |= sym.AttrLocal
		switch s.Name[:5] {
		case "$f32.":
			if uint64(uint32(x)) != x {
				log.Panicf("$-symbol %s too large: %d", s.Name, x)
			}
			s.AddUint32(arch, uint32(x))
		case "$f64.", "$i64.":
			s.AddUint64(arch, x)
		default:
			log.Panicf("unrecognized $-symbol: %s", s.Name)
		}
		s.Attr.Set(sym.AttrReachable, false)
	}
	if strings.HasPrefix(s.Name, "runtime.gcbits.") {
		s.Attr |= sym.AttrLocal
	}
}
