// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/src"
)

type linepair struct {
	first, last uint32
}

type xposmap_old struct {
	maps map[int32]*biasedSparseMap
	// The next two fields provide a single-item cache for common case of repeated lines from same file.
	lastIndex int32            // -1 means no entry in cache
	lastMap   *biasedSparseMap // map found at maps[lastIndex]
}

func newXposmap_old(x map[int]linepair) *xposmap_old {
	maps := make(map[int32]*biasedSparseMap)
	for i, p := range x {
		maps[int32(i)] = newBiasedSparseMap(int(p.first), int(p.last))
	}
	return &xposmap_old{maps: maps, lastIndex: -1} // zero for the rest is okay
}

type xposmap struct {
	maps      []*biasedSparseMap
	densemaps []*biasedSparseMap
}

func newXposmap(x map[int]linepair) *xposmap {
	maxfile := int(0)
	for k, _ := range x {
		if maxfile < k {
			maxfile = k
		}
	}

	maps := make([]*biasedSparseMap, maxfile+2) //
	densemaps := make([]*biasedSparseMap, 0, len(x))
	for i, p := range x {
		maps[int32(i)] = newBiasedSparseMap(int(p.first), int(p.last))
	}
	for _, m := range maps { //
		if m != nil {
			densemaps = append(densemaps, m)
		}
	}
	return &xposmap{maps: maps, densemaps: densemaps}
}

func (m *xposmap_old) clear() {
	for _, l := range m.maps {
		if l != nil {
			l.clear()
		}
	}
	m.lastIndex = -1
	m.lastMap = nil
}

func (m *xposmap) clear() {
	for _, l := range m.densemaps {
		if l != nil {
			l.clear()
		}
	}
}

func (m *xposmap_old) mapFor(p src.XPos) *biasedSparseMap {
	if p == src.NoXPos {
		return nil
	}
	index := p.FileIndex()
	if index == m.lastIndex {
		return m.lastMap
	}
	mf := m.maps[index]
	m.lastIndex = index
	m.lastMap = mf
	return mf
}

func (m *xposmap) mapFor(p src.XPos) *biasedSparseMap {
	if p == src.NoXPos {
		return nil
	}
	index := p.FileIndex()
	//if int(index) >= len(m.maps) {
	//	println("Map for", index, "is out of bounds, len(m.maps)=", len(m.maps))
	//	return nil
	//}
	ma := m.maps[index]
	if ma == nil {
		println("Map for", index, "is nil")
	}
	return ma
}

func (m *xposmap_old) set(p src.XPos, v int32) {
	s := m.mapFor(p)
	if s == nil {
		return
	}
	s.set(p.Line(), v)
}

func (m *xposmap) set(p src.XPos, v int32) {
	s := m.mapFor(p)
	if s == nil {
		return
	}
	s.set(p.Line(), v)
}

func (m *xposmap_old) add(p src.XPos) {
	m.set(p, 0)
}

func (m *xposmap) add(p src.XPos) {
	m.set(p, 0)
}

func (m *xposmap_old) contains(p src.XPos) bool {
	s := m.maps[p.FileIndex()]
	if s == nil {
		return false
	}
	return s.contains(p.Line())
}

func (m *xposmap) contains(p src.XPos) bool {
	s := m.mapFor(p)
	if s == nil {
		return false
	}
	return s.contains(p.Line())
}

func (m *xposmap_old) get(p src.XPos) int32 {
	s := m.mapFor(p)
	if s == nil {
		return -1
	}
	return s.get(p.Line())
}

func (m *xposmap) get(p src.XPos) int32 {
	s := m.mapFor(p)
	if s == nil {
		return -1
	}
	return s.get(p.Line())
}

func (m *xposmap_old) remove(p src.XPos) {
	s := m.mapFor(p)
	if s == nil {
		return
	}
	s.remove(p.Line())
}

func (m *xposmap) remove(p src.XPos) {
	s := m.mapFor(p)
	if s == nil {
		return
	}
	s.remove(p.Line())
}

func (m *xposmap_old) foreachEntry(f func(j int32, l uint, v int32)) {
	for j, mm := range m.maps {
		s := mm.size()
		for i := 0; i < s; i++ {
			l, v := mm.getEntry(i)
			f(j, l, v)
		}
	}
}
func (m *xposmap) foreachEntry(f func(j int32, l uint, v int32)) {
	for j, mm := range m.maps {
		s := mm.size()
		for i := 0; i < s; i++ {
			l, v := mm.getEntry(i)
			f(int32(j), l, v)
		}
	}
}
