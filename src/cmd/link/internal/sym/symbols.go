// Derived from Inferno utils/6l/l.h and related files.
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/6l/l.h
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sym

import (
	"runtime"
	"sync"
)

type LookupCounts struct {
	Lookup   uint
	ROLookup uint
}

type Stringver struct {
	Name string
	Ver  int
}

type LookupStack struct {
	Pcs   []uintptr
	Count uint
}

type Symbols struct {
	symbolBatch []Symbol

	// Symbol lookup based on name and indexed by version.
	hash []map[string]*Symbol

	Allsym []*Symbol

	Unique      map[Stringver]uint
	LookupSites map[uintptr]LookupStack
	LSCollide   uint
	umu         sync.Mutex
	ustats      bool

	Globs  LookupCounts
	ABIIs  LookupCounts
	Locals LookupCounts
}

func NewSymbols() *Symbols {
	hash := make([]map[string]*Symbol, SymVerStatic)
	// Preallocate about 2mb for hash of non static symbols
	hash[0] = make(map[string]*Symbol, 100000)
	// And another 1mb for internal ABI text symbols.
	hash[SymVerABIInternal] = make(map[string]*Symbol, 50000)
	unique := make(map[Stringver]uint)
	lookupSites := make(map[uintptr]LookupStack)
	return &Symbols{
		hash:        hash,
		Unique:      unique,
		LookupSites: lookupSites,
		Allsym:      make([]*Symbol, 0, 100000),
	}
}

func (syms *Symbols) Newsym(name string, v int) *Symbol {
	batch := syms.symbolBatch
	if len(batch) == 0 {
		batch = make([]Symbol, 1000)
	}
	s := &batch[0]
	syms.symbolBatch = batch[1:]

	s.Dynid = -1
	s.Name = name
	s.Version = int16(v)
	syms.Allsym = append(syms.Allsym, s)

	return s
}

func (syms *Symbols) CollectLookupStats(val bool) {
	syms.ustats = val
}

func sameSlice(sl1 []uintptr, sl2 []uintptr) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i := 0; i < len(sl1); i++ {
		if sl1[i] != sl2[i] {
			return false
		}
	}
	return true
}

func (syms *Symbols) recordLookupTrace(pcs []uintptr) {
	// Hash
	var h uintptr
	for _, pc := range pcs {
		h += pc
		h += h << 10
		h ^= h >> 6
	}
	// Install in map
	var v LookupStack
	var ok bool
	if v, ok = syms.LookupSites[h]; ok {
		// Record collision for sanity check
		if !sameSlice(v.Pcs, pcs) {
			syms.LSCollide++
		}
		v.Count++
	} else {
		v.Pcs = pcs
		v.Count = 1
	}
	syms.LookupSites[h] = v
}

func (syms *Symbols) stats(ro bool, v int, n string) {
	if !syms.ustats {
		return
	}
	var lp *LookupCounts
	if v == 0 {
		lp = &syms.Globs
	} else if v == SymVerABIInternal {
		lp = &syms.ABIIs
	} else {
		lp = &syms.Locals
	}
	if ro {
		lp.ROLookup++
	} else {
		lp.Lookup++
	}
	sv := Stringver{Name: n, Ver: v}
	syms.umu.Lock()
	defer syms.umu.Unlock()
	count := syms.Unique[sv]
	count += 1
	syms.Unique[sv] = count
	var stk [32]uintptr
	nstk := runtime.Callers(2, stk[:])
	syms.recordLookupTrace(stk[:nstk])
}

// Look up the symbol with the given name and version, creating the
// symbol if it is not found.
func (syms *Symbols) Lookup(name string, v int) *Symbol {
	syms.stats(false, v, name)
	m := syms.hash[v]
	s := m[name]
	if s != nil {
		return s
	}
	s = syms.Newsym(name, v)
	m[name] = s
	return s
}

// Look up the symbol with the given name and version, returning nil
// if it is not found.
func (syms *Symbols) ROLookup(name string, v int) *Symbol {
	syms.stats(true, v, name)
	return syms.hash[v][name]
}

// Add an existing symbol to the symbol table.
func (syms *Symbols) Add(s *Symbol) {
	name := s.Name
	v := int(s.Version)
	m := syms.hash[v]
	if _, ok := m[name]; ok {
		panic(name + " already added")
	}
	m[name] = s
	syms.Allsym = append(syms.Allsym, s)
}

// Allocate a new version (i.e. symbol namespace).
func (syms *Symbols) IncVersion() int {
	syms.hash = append(syms.hash, make(map[string]*Symbol))
	return len(syms.hash) - 1
}

// Rename renames a symbol.
func (syms *Symbols) Rename(old, new string, v int) {
	s := syms.hash[v][old]
	oldExtName := s.Extname()
	s.Name = new
	if oldExtName == old {
		s.SetExtname(new)
	}
	delete(syms.hash[v], old)

	dup := syms.hash[v][new]
	if dup == nil {
		syms.hash[v][new] = s
	} else {
		if s.Type == 0 {
			dup.Attr |= s.Attr
			*s = *dup
		} else if dup.Type == 0 {
			s.Attr |= dup.Attr
			*dup = *s
			syms.hash[v][new] = s
		}
	}
}
