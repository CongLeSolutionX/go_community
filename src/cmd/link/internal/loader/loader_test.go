// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

import (
	"bytes"
	"cmd/internal/objabi"
	"cmd/link/internal/sym"
	"testing"
)

func sameRelocSlice(s1 []Reloc, s2 []Reloc) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// dummyAddSym adds the named symbol to the loader as if it had been
// read from a Go object file. Note that it allocates a global
// index without creating an associated object reader, so one can't
// do anything interesting with this symbol (such as look at its
// data or relocations).
func addDummyObjSym(t *testing.T, ldr *Loader, or *oReader, name string) Sym {
	idx := ldr.max + 1
	ldr.max++
	if ok := ldr.AddSym(name, 0, idx, or, false, sym.SRODATA); !ok {
		t.Errorf("AddrSym failed for '" + name + "'")
	}
	return idx
}

func TestAddMaterializeSymbol(t *testing.T) {
	ldr := NewLoader(0)
	dummyOreader := oReader{version: -1}
	or := &dummyOreader

	// Create some syms from a dummy object file symbol to get things going.
	ts1 := addDummyObjSym(t, ldr, or, "type.uint8")
	ts2 := addDummyObjSym(t, ldr, or, "mumble")
	ts3 := addDummyObjSym(t, ldr, or, "type.string")
	ldr.InitReachable()

	// Create some external symbols.
	es1 := ldr.AddExtSym("extnew1", 0)
	if es1 == 0 {
		t.Fatalf("AddExtSym failed for extnew1")
	}
	es1x := ldr.AddExtSym("extnew1", 0)
	if es1x != 0 {
		t.Fatalf("AddExtSym lookup: expected 0 got %d for second lookup", es1x)
	}
	es2 := ldr.AddExtSym("go.info.type.uint8", 0)
	if es2 == 0 {
		t.Fatalf("AddExtSym failed for go.info.type.uint8")
	}
	// Create a nameless symbol
	es3 := ldr.CreateExtSym("")
	if es3 == 0 {
		t.Fatalf("CreateExtSym failed for nameless sym")
	}

	// Check get/set symbol type
	es3typ := ldr.SymType(es3)
	if es3typ != sym.Sxxx {
		t.Errorf("SymType(es3): expected %d, got %d", sym.Sxxx, es3typ)
	}
	ldr.SetType(es3, sym.SRODATA)
	es3typ = ldr.SymType(es3)
	if es3typ != sym.SRODATA {
		t.Errorf("SymType(es3): expected %d, got %d", sym.SRODATA, es3typ)
	}
	ldr.SetType(es3, sym.SRODATA)

	// New symbols should not be reachable
	if ldr.Reachable.Has(es1) || ldr.Reachable.Has(es2) ||
		ldr.Reachable.Has(es3) {
		t.Errorf("newly materialized symbols should not be reachable")
	}

	// Add some relocations to the new symbols.
	r1 := Reloc{0, 1, objabi.R_ADDR, 0, ts1}
	r2 := Reloc{3, 8, objabi.R_CALL, 0, ts2}
	r3 := Reloc{7, 1, objabi.R_USETYPE, 0, ts3}
	ldr.AddReloc(es1, r1)
	ldr.AddReloc(es1, r2)
	ldr.AddReloc(es2, r3)

	// Add some data to the symbols.
	d1 := []byte{1, 2, 3}
	d2 := []byte{4, 5, 6, 7}
	ldr.AddBytes(es1, d1)
	ldr.AddBytes(es2, d2)

	// Now invoke the usual loader interfaces to make sure
	// we're getting the right things back for these symbols.
	// First relocations...
	expRel := [][]Reloc{[]Reloc{r1, r2}, []Reloc{r3}}
	for k, idx := range []Sym{es1, es2} {
		relocs := ldr.Relocs(idx)
		exp := expRel[k]
		rsl := relocs.ReadAll(nil)
		if !sameRelocSlice(rsl, exp) {
			t.Errorf("expected relocs %v, got %v", exp, rsl)
		}
		r0 := relocs.At(0)
		if r0 != exp[0] {
			t.Errorf("expected reloc %v, got %v", exp[0], r0)
		}
	}

	// ... then data.
	dat := ldr.Data(es2)
	if bytes.Compare(dat, d2) != 0 {
		t.Errorf("expected es2 data %v, got %v", d2, dat)
	}

	// Nameless symbol should still be nameless.
	es3name := ldr.RawSymName(es3)
	if "" != es3name {
		t.Errorf("expected es3 name of '', got '%s'", es3name)
	}

	// Read value of materialized symbol.
	es1val := ldr.Value(es1)
	if 0 != es1val {
		t.Errorf("expected es1 value of 0, got %v", es1val)
	}

	// Test other misc methods
	irm := ldr.IsReflectMethod(es1)
	if 0 != es1val {
		t.Errorf("expected IsReflectMethod(es1) value of 0, got %v", irm)
	}

	// Writing data to a materialized symbol should mark it reachable.
	if !ldr.Reachable.Has(es1) || !ldr.Reachable.Has(es2) {
		t.Fatalf("written-to materialized symbols should be reachable")
	}

	// For materialized symbols we're allowed to reach into the relocations
	// slice and mutate the values there. Test to make sure we can do this.
	r0 := ldr.MutableReloc(es1, 0)
	r0.Type = objabi.R_ADDROFF
	relocs := ldr.Relocs(es1)
	rsl := relocs.ReadAll(nil)
	if rsl[0].Type != objabi.R_ADDROFF {
		t.Fatalf("mutating relocation 0: got %v wanted %v",
			rsl[0].Type, objabi.R_ADDROFF)
	}
}
