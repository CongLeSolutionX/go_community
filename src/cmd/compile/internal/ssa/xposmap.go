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

type xposmap struct {
	maps map[int32]*biasedSparseMap
	// The next two fields provide a single-item cache for common case of repeated lines from same file.
	lastIndex int32            // -1 means no entry in cache
	lastMap   *biasedSparseMap // map found at maps[lastIndex]
}

func newXposmap(x map[int]linepair) *xposmap {
	maps := make(map[int32]*biasedSparseMap)
	for i, p := range x {
		maps[int32(i)] = newBiasedSparseMap(int(p.first), int(p.last))
	}
	return &xposmap{maps: maps, lastIndex: -1} // zero for the rest is okay
}

func (m *xposmap) clear() {
	for _, l := range m.maps {
		if l != nil {
			l.clear()
		}
	}
	m.lastIndex = -1
	m.lastMap = nil
}

func (m *xposmap) mapFor(p src.XPos) *biasedSparseMap {
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

func (m *xposmap) set(p src.XPos, v int32) {
	s := m.mapFor(p)
	if s == nil {
		return
	}
	s.set(p.Line(), v)
}

func (m *xposmap) add(p src.XPos) {
	m.set(p, 0)
}

func (m *xposmap) contains(p src.XPos) bool {
	s := m.mapFor(p)
	if s == nil {
		return false
	}
	return s.contains(p.Line())
}

func (m *xposmap) get(p src.XPos) int32 {
	s := m.mapFor(p)
	if s == nil {
		return -1
	}
	return s.get(p.Line())
}

func (m *xposmap) remove(p src.XPos) {
	s := m.mapFor(p)
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
