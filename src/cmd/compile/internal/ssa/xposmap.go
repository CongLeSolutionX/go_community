// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/src"
	"fmt"
)

type lineRange struct {
	first, last uint32
}

// An xposmap is a map from src.XPos to int32, implemented
// sparsely to save space.  The sparse skeleton is constructed
// once, and then reused by ssa phases that (re)move values
// with statements attached.
type xposmap struct {
	// A map from file index to maps from line range to integers (block numbers)
	maps map[int32]*biasedSparseMap
	// The next two fields provide a single-item cache for common case of repeated lines from same file.
	lastIndex int32            // -1 means no entry in cache
	lastMap   *biasedSparseMap // map found at maps[lastIndex]
}

// newXposmap uses a map from integers (file indices) to line ranges to
// construct an empty, skeletal Xposmap.
func newXposmap(x map[int]lineRange) *xposmap {
	maps := make(map[int32]*biasedSparseMap)
	for i, p := range x {
		maps[int32(i)] = newBiasedSparseMap(int(p.first), int(p.last))
	}
	return &xposmap{maps: maps, lastIndex: -1} // zero for the rest is okay
}

// clear removes data from the map but leaves the sparse skeleton.
func (m *xposmap) clear() {
	for _, l := range m.maps {
		if l != nil {
			l.clear()
		}
	}
	m.lastIndex = -1
	m.lastMap = nil
}

// mapFor returns the line range map for a given file, with
// appropriate use and update of the single-entry cache.
func (m *xposmap) mapFor(index int32) *biasedSparseMap {
	if index == m.lastIndex {
		return m.lastMap
	}
	mf := m.maps[index]
	m.lastIndex = index
	m.lastMap = mf
	return mf
}

// set inserts p->v into the map.
// If p does not fall within the set of file->lineRange used to
// construct the map, this will panic.
func (m *xposmap) set(p src.XPos, v int32) {
	s := m.mapFor(p.FileIndex())
	if s == nil {
		panic(fmt.Sprintf("xposmap.set(%d), file index not found in map\n", p.FileIndex()))
	}
	s.set(p.Line(), v)
}

func (m *xposmap) add(p src.XPos) {
	m.set(p, 0)
}

func (m *xposmap) contains(p src.XPos) bool {
	s := m.mapFor(p.FileIndex())
	if s == nil {
		return false
	}
	return s.contains(p.Line())
}

func (m *xposmap) get(p src.XPos) int32 {
	s := m.mapFor(p.FileIndex())
	if s == nil {
		return -1
	}
	return s.get(p.Line())
}

func (m *xposmap) remove(p src.XPos) {
	s := m.mapFor(p.FileIndex())
	if s == nil {
		return
	}
	s.remove(p.Line())
}

func (m *xposmap) foreachEntry(f func(j int32, l uint, v int32)) {
	for j, mm := range m.maps {
		s := mm.size()
		for i := 0; i < s; i++ {
			l, v := mm.getEntry(i)
			f(j, l, v)
		}
	}
}
