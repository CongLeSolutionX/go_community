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
// created, the new scheme hashes the contents of the attribute and then
// does lookup to see whether we've already seen an identical instance.
// Return value from the table lookup is a token or attribute "index",
// which is then stored in the DWDie. Thus things now look like
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
//   │ Child: nil │  .    │ Atr: 73      │    │ Atr: 73    │  .
//   │ Attrs:     │  .    │ Cls:  ...    │    │ Cls:  ...  │  .
//   │  [3 1 2]   │  .    │ Value: ...   │    │ Value: ... │  .
//   │ Link: \    │  .    │ Data: <tsym> │    │ Data: "p2" │  .
//   └───────|────┘  .    └──────────────┘    └────────────┘  .
//           v       ..........................................
//          ...
//

package ld

import (
	"cmd/internal/dwarf"
	"cmd/link/internal/sym"
	"fmt"
)

const indexSlabSize = 4096
const bucketSlabSize = 1024
const bucketNumSlots = 3

type attrBucket struct {
	next  *attrBucket
	slots [bucketNumSlots]uint32
	count uint32
}

const attrSlabSzBits = 12
const attrSlabSize = 1 << attrSlabSzBits
const attrSlotMask = (1 << attrSlabSzBits) - 1

type attrTab struct {
	hm        map[uint64]*attrBucket
	attrSlabs [][]dwarf.DWAttr
	buckSlab  []attrBucket
	indexSlab []uint32
	lookups   uint64
	buckCount uint32
	attrCount uint32
}

// makeAttrTab returns a new 'attrTab' instance, prepopulated with a single
// dummy attr with index 0.
func makeAttrTab() *attrTab {
	return &attrTab{
		hm:        make(map[uint64]*attrBucket),
		attrCount: 1,
		attrSlabs: [][]dwarf.DWAttr{make([]dwarf.DWAttr, 1, attrSlabSize)},
		buckSlab:  make([]attrBucket, bucketSlabSize),
	}
}

func (at *attrTab) get(idx uint32) *dwarf.DWAttr {
	slot := idx & attrSlotMask
	slab := idx >> attrSlabSzBits
	return &at.attrSlabs[slab][slot]
}

func (at *attrTab) insert(attr uint16, cls int, value int64, data interface{}) uint32 {
	newIdx := at.attrCount
	newAt := dwarf.DWAttr{Atr: attr, Cls: uint8(cls), Value: value, Data: data}
	slot := at.attrCount & attrSlotMask
	slab := at.attrCount >> attrSlabSzBits
	if slot == attrSlotMask {
		// New slab needed
		at.attrSlabs = append(at.attrSlabs,
			make([]dwarf.DWAttr, 0, attrSlabSize))
	}
	at.attrSlabs[slab] = append(at.attrSlabs[slab], newAt)
	at.attrCount++
	return newIdx
}

func hashString(h uint64, s string) uint64 {
	for _, c := range s {
		h = (h << 4) + uint64(c)
		high := h & uint64(0xF0000000000000)
		if high != 0 {
			h ^= high >> 48
			h &^= high
		}
	}
	return h
}

func hashBytes(h uint64, bs []byte) uint64 {
	for _, b := range bs {
		h = (h << 4) + uint64(b)
		high := h & uint64(0xF0000000000000)
		if high != 0 {
			h ^= high >> 48
			h &^= high
		}
	}
	return h
}

// hashAttr hashes the contents of a given DWAttr object.
func hashAttr(at dwarf.DWAttr) uint64 {
	h := uint64(at.Cls)<<16 | uint64(at.Atr)
	h = h ^ uint64(at.Value)

	switch at.Cls {
	default:
		panic("unexpected attr class")

	case dwarf.DW_CLS_CONSTANT, dwarf.DW_CLS_FLAG:
		return h

	case dwarf.DW_CLS_STRING:
		// String.
		s := at.Data.(string)
		return hashString(h, s)

	case dwarf.DW_CLS_REFERENCE, dwarf.DW_CLS_ADDRESS,
		dwarf.DW_CLS_GO_TYPEREF, dwarf.DW_CLS_PTR:
		// Symbol-valued data. We still have to handle symbols of both
		// flavors (ugh).

		// First loader symbol.
		ds, ok := at.Data.(dwSym)
		if ok {
			return (h << 4) + uint64(ds)
		}
		ss := at.Data.(*sym.Symbol)

		// Next sym.Symbol.
		//
		// Here we have to be careful for the moment -- we may see a lot
		// of anonymous DWARF info symbols that no name and the same type,
		// which can lead to very poor hashing behavior (this will go away
		// once we can just hash the loader.Sym). For now, hash the
		// sym pointer itself.
		if ss.Name == "" {
			return hashString(h, fmt.Sprintf("%p", ss))
		} else {
			return hashString(h, ss.Name)
		}

	case dwarf.DW_CLS_BLOCK:
		return hashBytes(h, at.Data.([]byte))
	}
}

// equalAttr compares two DWAttr objects for equality.
//
// This helper takes advantage of the attr and class info to perform
// more specialized/customized equality testing of attribute data (a
// previous more general version did checked interface conversions to
// discover the data flavor, and that was a bit slow).
func equalAttr(at1 *dwarf.DWAttr, at2 *dwarf.DWAttr) bool {
	if at1.Atr != at2.Atr {
		return false
	}
	if at1.Cls != at2.Cls {
		return false
	}
	if at1.Value != at2.Value {
		return false
	}

	switch at1.Cls {

	case dwarf.DW_CLS_CONSTANT, dwarf.DW_CLS_FLAG:
		// Constant (no data to compare, only Value field).
		return true

	case dwarf.DW_CLS_STRING:
		// String.
		s1 := at1.Data.(string)
		s2 := at2.Data.(string)
		return s1 == s2

	case dwarf.DW_CLS_REFERENCE, dwarf.DW_CLS_ADDRESS, dwarf.DW_CLS_GO_TYPEREF:
		// Symbol-valued data.
		// Currently we have to handle symbols of both flavors (ugh)
		ds1, ok1d := at1.Data.(dwSym)
		ds2, ok2d := at2.Data.(dwSym)
		if ok1d != ok2d {
			return false
		}
		if ok1d {
			return ds1 == ds2
		}
		ss1 := at1.Data.(*sym.Symbol)
		ss2 := at2.Data.(*sym.Symbol)
		return ss1 == ss2

	case dwarf.DW_CLS_BLOCK:
		b1 := at1.Data.([]byte)
		b2 := at2.Data.([]byte)
		if len(b1) != len(b2) {
			return false
		}
		for i := range b1 {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	}

	panic("unhandled attr")

	//panic(fmt.Sprintf("unhandled attr %d cls %d data contents: t:%s %+v t:%s %+v ", at1.Atr, at2.Cls, reflect.TypeOf(at1.Data).String(), at1.Data, reflect.TypeOf(at2.Data).String(), at2.Data))
}

func (at *attrTab) newBucket() *attrBucket {
	if len(at.buckSlab) == 0 {
		at.buckSlab = make([]attrBucket, bucketSlabSize)
	}
	ab := &at.buckSlab[len(at.buckSlab)-1]
	at.buckSlab = at.buckSlab[:len(at.buckSlab)-1]
	return ab
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

	// Hash the contents of the attribute and look it up in our table to
	// see if we have an instance already.
	hashCode := hashAttr(cand)
	buck := at.hm[hashCode]
	if buck != nil {
		for {
			for i := range buck.slots {
				if equalAttr(&cand, at.get(buck.slots[i])) {
					return buck.slots[i]
				}
			}
			if buck.next == nil {
				break
			}
			buck = buck.next
		}
	} else {
		buck = at.newBucket()
		at.hm[hashCode] = buck
	}

	newIdx := at.insert(attr, cls, value, data)
	if buck.count < bucketNumSlots {
		buck.slots[buck.count] = newIdx
	} else {
		nbp := at.newBucket()
		buck.next = nbp
		buck = nbp
	}
	buck.count++
	return newIdx
}

type attrTabStats struct {
	lookups     uint64
	created     uint64
	buckets     uint32
	longestbuck uint32
	avgbucklen  float64
}

// stats returns a few statistics on what's happened with hash table
// performance (total lookups, total buckets, total attrs created, etc)
// for debugging purposes.
func (at *attrTab) stats() attrTabStats {
	buckets := uint32(len(at.hm))
	totbucklen := uint32(0)
	longestbucklen := uint32(0)
	for _, buck := range at.hm {
		bl := uint32(len(buck.slots))
		totbucklen += bl
		if bl > longestbucklen {
			longestbucklen = bl
		}
	}
	avgbucklen := float64(0)
	if buckets != 0 {
		avgbucklen = float64(totbucklen) / float64(buckets)
	}
	return attrTabStats{
		lookups:     at.lookups,
		created:     uint64(at.attrCount),
		buckets:     buckets,
		longestbuck: longestbucklen,
		avgbucklen:  avgbucklen,
	}
}
