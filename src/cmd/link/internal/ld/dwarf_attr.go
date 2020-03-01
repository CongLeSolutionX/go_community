// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file defines the linker's 'attrTab' type, which provides
// a set up methods (notable 'lookup' and 'get') to help manage
// the creation of DWAttr objects.
//
// Motivation:
//
// In the original design of the linker (Go 1.4 up until 1.13 or so)
// DWARF dies and attributes were each distinct, separately allocated
// objects, with all relationships (containment, sibling, etc) expressed
// via pointers and linked lists. Consider the following two DWARF DIEs,
// an excerpt from a subprogram DIE, output from 'objdump --dwarf':
//
//  <2><85>: Abbrev Number: 17 (DW_TAG_formal_parameter)
//     <86>   DW_AT_name        : p1
//     <8a>   DW_AT_variable_parameter: 0
//     <8b>   DW_AT_type        : <0x3d2c1>
//  <2><8f>: Abbrev Number: 17 (DW_TAG_formal_parameter)
//     <90>   DW_AT_name        : p2
//     <96>   DW_AT_variable_parameter: 0
//     <97>   DW_AT_type        : <0x3d2c1>
//
// In the old implementation, we would construct the following forest of
// objects to represent these DIEs and attributes:
//
//   DWDie              DWAttr             DWAttr             DWAttr
//   ┌────────────┐     ┌────────────┐     ┌────────────┐     ┌─────────────┐
//   │ Abbrev: 17 │     │ Atr: 3     │     │ Atr: 75    │     │ Atr: 73     │
//   │ Sym:  ...  │     │ Cls:  ...  │     │ Cls:  ...  │     │ Cls:  ...   │
//   │ Child: nil │     │ Value: ... │     │ Value: 0   │     │ Value: ...  │
//   │ Attr:   ---│---> │ Data: "p1" │     │ Data: nil  │     │ Data: <tsym>│
//   │ Link: \    │     │ Link: -----│---> │ Link: -----│---> │ Link: nil   │
//   └───────|────┘     └────────────┘     └────────────┘     └─────────────┘
//           |
//   DWDie   v           DWAttr             DWAttr             DWAttr
//   ┌────────────┐     ┌────────────┐     ┌────────────┐     ┌─────────────┐
//   │ Abbrev: 17 │     │ Atr: 3     │     │ Atr: 75    │     │ Atr: 73     │
//   │ Sym:  ...  │     │ Cls:  ...  │     │ Cls:  ...  │     │ Cls:  ...   │
//   │ Child: nil │     │ Value: ... │     │ Value: 0   │     │ Value: ...  │
//   │ Attr:   ---│---> │ Data: "p3" │     │ Data: nil  │     │ Data: <tsym>│
//   │ Link: \    │     │ Link: -----│---> │ Link: -----│---> │ Link: nil   │
//   └───────|────┘     └────────────┘     └────────────┘     └─────────────┘
//           v
//          ...
//
// In addition to being very pointer-intensive, there is a fair amount of
// duplication. Each of these two dies has three attributes, however the
// last two DWAttr objects hanging off the attr list for the first die
// are identical to the last two DWAttr objects hanging off the second die.
//
// To reduce space overhead, at the point where a new attribute is
// created, the new scheme looks up the attribute in a table to see
// whether we've already seen an identical instance. Return value from
// the table lookup is a token or attribute "index", which is then
// stored in the DWDie. Thus things now look like
//
//   DWDie               dwAttrTab:
//   ┌────────────┐  ..........................................
//   │ Abbrev: 17 │  .    DWAttr index 0      DWAttr index 1  .
//   │ Sym:  ...  │  .    ┌────────────┐      ┌────────────┐  .
//   │ Child: nil │  .    │ Atr: 3     │      │ Atr: 75    │  .
//   │ Attrs:     │  .    │ Cls:  ...  │      │ Cls:  ...  │  .
//   │  [0 1 2]   │  .    │ Value: ... │      │ Value: ... │  .
//   │ Link: \    │  .    │ Data: "p1" │      │ Data: "p1" │  .
//   └───────|────┘  .    └────────────┘      └────────────┘  .
//           v       .                                        .
//   ┌────────────┐  .                                        .
//   │ Abbrev: 17 │  .    DWAttr index 2      DWAttr index 3  .
//   │ Sym:  ...  │  .    ┌──────────────┐    ┌────────────┐  .
//   │ Child: nil │  .    │ Atr: 73      │    │ Atr: 3     │  .
//   │ Attrs:     │  .    │ Cls:  ...    │    │ Cls:  ...  │  .
//   │  [3 1 2]   │  .    │ Value: ...   │    │ Value: ... │  .
//   │ Link: \    │  .    │ Data: <tsym> │    │ Data: "p2" │  .
//   └───────|────┘  .    └──────────────┘    └────────────┘  .
//           v       ..........................................
//          ...
//
// In terms of the effectiveness of the table when used in the Go
// linker: for kubernetes 'kubelet', about 70% of all attributes wind
// up being commoned, which is pretty decent.
//
// Worth noting: some attribute codes have better hit ratios than
// others. For example, DW_AT_go_runtime_type, DW_AT_go_package_name,
// DW_AT_stmt_list, and a couple of other attributes virtually all
// have unique values, meaning that the table provides no benefit. If
// this things change and the overall hit ratio goes down, it would
// not be hard to change the lookup method to skip the content check
// completely for certain attribute codes (just install a new attr
// without checking); this might speed thing up a bit.

package ld

import (
	"cmd/internal/dwarf"
	"cmd/link/internal/loader"
)

const indexSlabSize = 4096

const attrSlabSzBits = 12
const attrSlabSize = 1 << attrSlabSzBits
const attrSlotMask = (1 << attrSlabSzBits) - 1

// attrTab is a repository for DWAttr structs indended to allow
// detection and commoning of duplicate DWARF attributes. Each time a
// new attribute is created, the attrTab.lookup method looks up the
// attr to see if we've already hit an attr with the same content. If
// so, the index of the previously created attr is returned. If not, a
// new DWAttr is added to 'attrSlabs' and an index describing the slot
// is returned.
type attrTab struct {
	hm        map[dwarf.DWAttr]uint32
	attrSlabs [][]dwarf.DWAttr
	indexSlab []uint32
	lookups   uint64
	attrCount uint32
}

// makeAttrTab returns a new 'attrTab' instance, prepopulated with a single
// dummy attr with index 0.
func makeAttrTab() *attrTab {
	return &attrTab{
		hm:        make(map[dwarf.DWAttr]uint32),
		attrCount: 1,
		attrSlabs: [][]dwarf.DWAttr{make([]dwarf.DWAttr, 1, attrSlabSize)},
	}
}

// get method returns a pointer to the DWAttr struct for the specified
// index. The index space is segmented into slabs of size 4096, so
// we first pick the right slab and then the slot within the slab.
func (at *attrTab) get(idx uint32) *dwarf.DWAttr {
	slot := idx & attrSlotMask
	slab := idx >> attrSlabSzBits
	return &at.attrSlabs[slab][slot]
}

// insert adds a new DWAttr struct to the table (the assumption being that
// lookup has failed to find an existing identical DWAttr in the table).
func (at *attrTab) insert(newAt dwarf.DWAttr) uint32 {
	newIdx := at.attrCount
	slot := newIdx & attrSlotMask
	slab := newIdx >> attrSlabSzBits
	if slot == attrSlotMask {
		// New slab needed
		at.attrSlabs = append(at.attrSlabs,
			make([]dwarf.DWAttr, 0, attrSlabSize))
	}
	at.attrSlabs[slab] = append(at.attrSlabs[slab], newAt)
	at.attrCount++
	return newIdx
}

// allocIndexSlice returns a subslice from an internally allocated
// uint32 slab. The intent here is to provide more efficient allocation
// of attribute index slices in the dwarf.DWDie struct.
func (at *attrTab) allocIndexSlice(sz int) []uint32 {
	if len(at.indexSlab) < sz {
		if sz > indexSlabSize {
			panic("unexpectedly large index allocation request")
		}
		at.indexSlab = make([]uint32, indexSlabSize)
	}
	rval := at.indexSlab[:sz:sz]
	at.indexSlab = at.indexSlab[sz:]
	return rval
}

// lookup looks up a DWARF attribute by content, returning an index or token
// for that attribute. If an attribute with the specified content has been
// seen already, the existing index will be returned, otherwise a new
// attr will be added to the table.
func (at *attrTab) lookup(attr uint16, cls int, value int64, data interface{}) uint32 {
	at.lookups++
	cand := dwarf.DWAttr{Atr: attr, Cls: uint8(cls), Value: value, Data: data}

	// Some attributes are almost always unique -- don't bother trying to
	// look them up, just create new instances right away.
	nocache := (attr == dwarf.DW_AT_go_runtime_type ||
		attr == dwarf.DW_AT_location || attr == dwarf.DW_AT_go_elem)

	// See if we've encountered this already.
	if !nocache {
		if idx, ok := at.hm[cand]; ok {
			return idx
		}
	}

	// This is a new, not-yet encountered attr. Add it to the table.
	newIdx := at.insert(cand)
	at.hm[cand] = newIdx
	return newIdx
}

// Convert loader.Sym to sym.Symbol in all hashed attributes.
// Temporary only needed until DWARF phase 2 is always on.
func (at *attrTab) convertSymbols(l *loader.Loader) {
	for _, slab := range at.attrSlabs {
		for i := range slab {
			if attrSym, ok := slab[i].Data.(dwSym); ok {
				slab[i].Data = l.Syms[loader.Sym(attrSym)]
			}
		}
	}
}

type attrTabStats struct {
	lookups   uint64
	created   uint64
	missratio float64
}

// stats returns a few statistics on what's happened with lookup table
// performance (total lookups, total buckets, total attrs created, etc)
// for debugging purposes.
func (at *attrTab) stats() attrTabStats {
	missratio := float64(0)
	if at.attrCount != 0 {
		missratio = float64(at.attrCount) / float64(at.lookups) * float64(100)
	}
	return attrTabStats{
		lookups:   at.lookups,
		created:   uint64(at.attrCount),
		missratio: missratio,
	}
}
