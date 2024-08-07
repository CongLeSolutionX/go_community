// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps implements Go's builtin map type.
package maps

import (
	"internal/abi"
	"internal/goarch"
	"internal/runtime/sys"
	"unsafe"
)

// Extracts the H1 portion of a hash: the 57 upper bits.
// TODO(prattmic): what about 32-bit systems?
func h1(h uintptr) uintptr {
	return h >> 7
}

// Extracts the H2 portion of a hash: the 7 bits not used for h1.
//
// These are used as an occupied control byte.
func h2(h uintptr) uintptr {
	return h & 0x7f
}

type Map struct {
	// The number of filled slots (i.e. the number of elements in all
	// tables).
	used uint64

	// Type of this map.
	//
	// TODO(prattmic): Old maps pass this into every call instead of
	// keeping a reference in the map header. This is probably more
	// efficient and arguably more robust (crafty users can't reach into to
	// the map to change its type), but I leave it here for now for
	// simplicity.
	typ *abi.SwissMapType

	seed uintptr

	// The directory of tables. The length of this slice is
	// `1 << globalDepth`. Multiple entries may point to the same table.
	// See top-level comment for more details.
	directory []*table

	// The number of bits to use in table directory lookups.
	globalDepth uint32

	// clearSeq is a sequence counter of calls to Clear. It is used to
	// detect map clears during iteration.
	clearSeq uint64
}

func NewMap(mt *abi.SwissMapType, capacity uint64) *Map {
	if capacity < abi.SwissMapGroupSlots {
		// TODO: temporary to simplify initial implementation.
		capacity = abi.SwissMapGroupSlots
	}
	dirSize := (capacity + maxTableCapacity - 1) / maxTableCapacity
	dirSize, overflow := alignUpPow2(dirSize)
	if overflow {
		panic("rounded-up capacity overflows uint64")
	}
	globalDepth := uint32(sys.TrailingZeros64(dirSize))

	m := &Map{
		typ: mt,

		//TODO
		//seed: uintptr(rand()),

		directory: make([]*table, dirSize),

		globalDepth: globalDepth,
	}

	for i := range m.directory {
		// TODO: Think more about initial table capacity.
		m.directory[i] = newTable(mt, capacity/dirSize, i, globalDepth)
	}

	return m
}

func (m *Map) Type() *abi.SwissMapType {
	return m.typ
}

func (m *Map) directoryIndex(hash uintptr) uintptr {
	// TODO(prattmic): Store the shift as globalShift, as we need that more
	// often than globalDepth.
	if goarch.PtrSize == 4 {
		return hash >> (32 - m.globalDepth)
	}
	return hash >> (64 - m.globalDepth)
}

func (m *Map) replaceTable(nt *table) {
	// The number of entries that reference the same table doubles for each
	// time the globalDepth grows without the table splitting.
	entries := 1 << (m.globalDepth - nt.localDepth)
	for i := 0; i < entries; i++ {
		m.directory[nt.index+i] = nt
	}
}

func (m *Map) installTableSplit(old, left, right *table) {
	if old.localDepth == m.globalDepth {
		// No room for another level in the directory. Grow the
		// directory.
		newDir := make([]*table, len(m.directory)*2)
		for i, t := range m.directory {
			newDir[2*i] = t
			newDir[2*i+1] = t
			// t may already exist in multiple indicies. We should
			// only update t.index once. Since the index must
			// increase, seeing the original index means this must
			// be the first time we've encountered this table.
			// XXX: think about this more.
			if t.index == i {
				t.index = 2*i
			}
		}
		m.globalDepth++
		m.directory = newDir
	}

	// N.B. left and right may still consume multiple indicies if the
	// directory has grown multiple times since old was last split.
	left.index = old.index
	m.replaceTable(left)

	entries := 1 << (m.globalDepth - left.localDepth)
	right.index = left.index + entries
	m.replaceTable(right)
}

func (m *Map) Used() uint64 {
	return m.used
}

// Get performs a lookup of the key that key points to. It returns a pointer to
// the element, or false if the key doesn't exist.
func (m *Map) Get(key unsafe.Pointer) (unsafe.Pointer, bool) {
	_, elem, ok := m.getWithKey(key)
	return elem, ok
}

func (m *Map) getWithKey(key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
	hash := m.typ.Hasher(key, m.seed)

	idx := m.directoryIndex(hash)
	return m.directory[idx].getWithKey(key)
}

func (m *Map) Put(key, elem unsafe.Pointer) {
	slotElem := m.PutSlot(key)
	typedmemmove(m.typ.Elem, slotElem, elem)
}

// PutSlot returns a pointer to the element slot where an inserted element
// should be written.
//
// PutSlot never returns nil.
func (m *Map) PutSlot(key unsafe.Pointer) unsafe.Pointer {
	hash := m.typ.Hasher(key, m.seed)

	for {
		idx := m.directoryIndex(hash)
		elem, ok := m.directory[idx].PutSlot(m, key)
		if !ok {
			continue
		}
		return elem
	}
}

func (m *Map) Delete(key unsafe.Pointer) {
	hash := m.typ.Hasher(key, m.seed)

	idx := m.directoryIndex(hash)
	m.directory[idx].Delete(m, key)
}

// Clear deletes all entries from the map resulting in an empty map.
func (m *Map) Clear() {
	for _, t := range m.directory {
		t.Clear()
	}
	m.used = 0
	m.clearSeq++
	// TODO: shrink directory?
}
