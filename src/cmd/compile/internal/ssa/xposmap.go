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
	maps      []*biasedSparseMap
	densemaps []int
	lastIndex int32            // -1 means no entry in cache
	lastMap   *biasedSparseMap // map found at maps[lastIndex]
}

func newXposmap(x map[int]linepair) *xposmap {
	maxfile := int(0)
	for k, _ := range x {
		if maxfile < k {
			maxfile = k
		}
	}

	maps := make([]*biasedSparseMap, maxfile+2) //
	densemaps := make([]int, 0, len(x))
	for i, p := range x {
		maps[int32(i)] = newBiasedSparseMap(int(p.first), int(p.last))
	}
	for i, m := range maps {
		if m != nil {
			densemaps = append(densemaps, i)
		}
	}
	return &xposmap{maps: maps, densemaps: densemaps, lastIndex: -1}
}

func (m *xposmap) clear() {
	for _, i := range m.densemaps {
		sm := m.maps[i]
		if sm != nil {
			sm.clear()
		}
	}
	m.lastIndex = -1
	m.lastMap = nil
}

func (m *xposmap) mapFor(p src.XPos) *biasedSparseMap {
	index := p.FileIndex()
	if index == m.lastIndex {
		return m.lastMap
	}
	if p == src.NoXPos {
		return nil
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
	for _, j := range m.densemaps {
		mm := m.maps[j]
		s := mm.size()
		for i := 0; i < s; i++ {
			l, v := mm.getEntry(i)
			f(int32(j), l, v)
		}
	}
}
