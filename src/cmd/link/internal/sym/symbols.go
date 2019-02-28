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
	"fmt"
	"strings"
)

type Symbols struct {
	symbolBatch []Symbol

	// Symbol lookup based on name and indexed by version.
	hash []map[SymName]*Symbol

	// interned symbol strings
	names SymNameTable

	Allsym []*Symbol
}

func NewSymbols() *Symbols {
	hash := make([]map[SymName]*Symbol, SymVerStatic)
	// Preallocate about 2mb for hash of non static symbols
	hash[0] = make(map[SymName]*Symbol, 100000)
	// And another 1mb for internal ABI text symbols.
	hash[SymVerABIInternal] = make(map[SymName]*Symbol, 50000)
	nametable := NewSymNameTable(50000)
	return &Symbols{
		hash:   hash,
		names:  nametable,
		Allsym: make([]*Symbol, 0, 100000),
	}
}

func (syms *Symbols) Newsym(name string, v int) *Symbol {
	encoded := syms.names.Lookup(name)
	return syms.NewsymEncoded(encoded, v)
}

func (syms *Symbols) NewsymEncoded(encoded SymName, v int) *Symbol {
	batch := syms.symbolBatch
	if len(batch) == 0 {
		batch = make([]Symbol, 1000)
	}
	s := &batch[0]
	syms.symbolBatch = batch[1:]

	s.Dynid = -1
	s.SetName(encoded)
	s.Version = int16(v)
	syms.Allsym = append(syms.Allsym, s)

	return s
}

// Look up the symbol with the given name and version, creating the
// symbol if it is not found.
func (syms *Symbols) Lookup(name string, v int) *Symbol {
	encoded := syms.names.Lookup(name)
	return syms.LookupEncoded(encoded, v)
}

func (syms *Symbols) LookupEncoded(encoded SymName, v int) *Symbol {
	m := syms.hash[v]
	s := m[encoded]
	if s != nil {
		return s
	}
	s = syms.NewsymEncoded(encoded, v)
	m[encoded] = s
	return s
}

// Look up the symbol with the given name and version, returning nil
// if it is not found.
func (syms *Symbols) ROLookup(name string, v int) *Symbol {
	encoded := syms.names.Lookup(name)
	return syms.ROLookupEncoded(encoded, v)
}

func (syms *Symbols) ROLookupEncoded(encoded SymName, v int) *Symbol {
	return syms.hash[v][encoded]
}

// Allocate a new version (i.e. symbol namespace).
func (syms *Symbols) IncVersion() int {
	syms.hash = append(syms.hash, make(map[SymName]*Symbol))
	return len(syms.hash) - 1
}

// Rename renames a symbol.
func (syms *Symbols) Rename(old, new string, v int, reachparent map[*Symbol]*Symbol) {
	eold := syms.names.Lookup(old)
	enew := syms.names.Lookup(new)
	syms.RenameEncoded(eold, enew, v, reachparent)
}

func (syms *Symbols) RenameEncoded(old, new SymName, v int, reachparent map[*Symbol]*Symbol) {
	s := syms.hash[v][old]
	oldExtName := s.extname()
	s.SetName(new)
	if oldExtName == old {
		s.setExtname(new)
	}
	delete(syms.hash[v], old)

	dup := syms.hash[v][new]
	if dup == nil {
		syms.hash[v][new] = s
	} else {
		if s.Type == 0 {
			dup.Attr |= s.Attr
			if s.Attr.Reachable() && reachparent != nil {
				reachparent[dup] = reachparent[s]
			}
			*s = *dup
		} else if dup.Type == 0 {
			s.Attr |= dup.Attr
			if dup.Attr.Reachable() && reachparent != nil {
				reachparent[s] = reachparent[dup]
			}
			*dup = *s
			syms.hash[v][new] = s
		}
	}
}

func (syms *Symbols) SymName(s *Symbol) string {
	return syms.names.String(s.name)
}

func (syms *Symbols) SetSymName(s *Symbol, newname string) {
	encoded := syms.names.Lookup(newname)
	s.SetName(encoded)
}

func (syms *Symbols) SymNameHasPrefix(s *Symbol, pref string) bool {
	return strings.HasPrefix(syms.names.String(s.name), pref)
}

func (syms *Symbols) SymNameHasSuffix(s *Symbol, pref string) bool {
	return strings.HasSuffix(syms.names.String(s.name), pref)
}

func (syms *Symbols) SymExtname(s *Symbol) string {
	return syms.names.String(s.extname())
}

func (syms *Symbols) SetSymExtname(s *Symbol, newname string) {
	encoded := syms.names.Lookup(newname)
	s.setExtname(encoded)
}

func (syms *Symbols) SymString(s *Symbol) string {
	sn := syms.names.String(s.name)
	if s.Version == 0 {
		return sn
	}
	return fmt.Sprintf("%v<%d>", sn, s.Version)
}

func (syms *Symbols) Lock() {
	syms.names.lock()
}

func (syms *Symbols) Unlock() {
	syms.names.unlock()
}

func (syms *Symbols) NameStats() SNTStats {
	return syms.names.stats()
}
