// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

import (
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"fmt"
	"testing"
)

func TestRelocAllocator(t *testing.T) {
	edummy := func(s *sym.Symbol, str string, off int) {}
	ldr := NewLoader(0, edummy)
	dummyOreader := oReader{version: -1, syms: make([]Sym, 100)}
	or := &dummyOreader

	// Populate loader with some symbols...
	addDummyObjSym(t, ldr, or, "type.uint8")
	syms := []Sym{}
	for i := 0; i < 100; i++ {
		syms = append(syms, ldr.AddExtSym(fmt.Sprintf("a%d", i), 0))
	}

	// ... then add some relocations.
	arch := sys.ArchAMD64
	for i := 0; i < 100; i++ {
		sb := ldr.MakeSymbolUpdater(syms[i])
		for j := i + 1; j < 100; j++ {
			sb.AddAddrPlus4(arch, syms[j], int64(j+i))
		}
	}

	// Create a reloc allocator. Use a smaller chunk size so that
	// the allocator's chunk handling will be exercised.
	ra := MakeRelocAllocator(ldr)
	ra.csize = 27

	// Starting point should be sane.
	ra.SanityCheck()

	// Exercise the allocator.
	for i := 0; i < 100; i++ {
		sl := []Relocs{}
		for j := i; j < 100; j++ {
			relocs := ldr.Relocs(syms[j])
			rsl := ra.Alloc(relocs.Count)
			relocs.ReadAll(rsl)
			if len(rsl) > 0 && rsl[0].Size != 4 {
				t.Errorf("bad")
			}
			sl = append(sl, relocs)
		}
		for j := i; j < 100; j++ {
			relocs := sl[len(sl)-1]
			sl = sl[0 : len(sl)-1]
			ra.Release(relocs.Count)
		}
	}

	// Check sanity again.
	ra.SanityCheck()
}
