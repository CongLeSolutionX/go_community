// Copyright 2018 The Go Authors. All rights reserved.
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
	maps              []*biasedSparseMap
	cachedsize        int
	geIndex, geOffset int // Used to make getEntry run quickly
}

func newXposmap(x []linepair) *xposmap {
	maps := make([]*biasedSparseMap, len(x))
	for i, p := range x {
		if p.last > 0 {
			maps[i] = newBiasedSparseMap(int(p.first), int(p.last))
		}
	}
	return &xposmap{maps: maps} // zero for the rest is okay
}

func (m *xposmap) clear() {
	for _, m := range m.maps {
		if m != nil {
			m.clear()
		}
	}
	m.cachedsize = 0
	m.geIndex = -1
}

func (m *xposmap) size() int {
	return m.cachedsize
}

func (m *xposmap) set(p src.XPos, v int32) {
	i := p.Index()
	if int(i) >= len(m.maps) || i < 0 || m.maps[i] == nil {
		return
	}
	s := m.maps[i]
	o := s.size()
	s.set(p.Line(), v)
	m.cachedsize += s.size() - o
	m.geIndex = -1
}

func (m *xposmap) add(p src.XPos) {
	m.set(p, 0)
}

func (m *xposmap) contains(p src.XPos) bool {
	i := p.Index()
	if int(i) >= len(m.maps) || i < 0 || m.maps[i] == nil {
		return false
	}
	return m.maps[i].contains(p.Line())
}

func (m *xposmap) get(p src.XPos) int32 {
	i := p.Index()
	if int(i) >= len(m.maps) || i < 0 || m.maps[i] == nil {
		return -1
	}
	return m.maps[i].get(p.Line())
}

func (m *xposmap) remove(p src.XPos) {
	i := p.Index()
	if int(i) >= len(m.maps) || i < 0 || m.maps[i] == nil {
		return
	}
	s := m.maps[i]
	o := s.size()
	s.remove(p.Line())
	m.cachedsize += s.size() - o
	m.geIndex = -1
}

// getEntry returns the i'th key and value stored in s,
// where 0 <= i < s.size().  "Key" is XPos index and line number.
func (m *xposmap) getEntry(i int) (j int32, l uint, v int32) {
	geI := m.geIndex
	geO := m.geOffset
	if geI < 0 { // known-good values
		geO = 0
		geI = 0
	}
	for {
		next := geO + m.maps[geI].size()
		if next > i {
			break
		}
		geO = next
		geI++
	}
	j = int32(geI)
	l, v = m.maps[geI].getEntry(i - geO)
	m.geIndex = geI
	m.geOffset = geO
	return
}
