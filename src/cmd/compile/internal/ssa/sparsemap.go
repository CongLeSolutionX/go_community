// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// from http://research.swtch.com/sparse
// in turn, from Briggs and Torczon

type sparseEntry struct {
	key ID
	val int32
}

type sparseMap struct {
	dense  []sparseEntry
	sparse []int
}

// newSparseMap returns a sparseMap that can map
// integers between 0 and n-1 to int32s.
func newSparseMap(n int) *sparseMap {
	return &sparseMap{nil, make([]int, n)}
}

func (s *sparseMap) size() int {
	return len(s.dense)
}

func (s *sparseMap) contains(k ID) bool {
	i := s.sparse[k]
	return i < len(s.dense) && s.dense[i].key == k
}

// get returns the value for key k, or -1 if k does
// not appear in the map.
func (s *sparseMap) get(k ID) int32 {
	i := s.sparse[k]
	if i < len(s.dense) && s.dense[i].key == k {
		return s.dense[i].val
	}
	return -1
}

func (s *sparseMap) set(k ID, v int32) {
	i := s.sparse[k]
	if i < len(s.dense) && s.dense[i].key == k {
		s.dense[i].val = v
		return
	}
	s.dense = append(s.dense, sparseEntry{k, v})
	s.sparse[k] = len(s.dense) - 1
}

// setOrIfDifferent sets value v0 for key k, unless there is
// already a *different* value present in which case it sets v.
func (s *sparseMap) setOrIfDifferent(k ID, v0, v int32) {
	i := s.sparse[k]
	if i < len(s.dense) && s.dense[i].key == k {
		if s.dense[i].val != v0 {
			s.dense[i].val = v
		}
		return
	}
	s.dense = append(s.dense, sparseEntry{k, v0})
	s.sparse[k] = len(s.dense) - 1
}

func (s *sparseMap) remove(k ID) {
	i := s.sparse[k]
	if i < len(s.dense) && s.dense[i].key == k {
		y := s.dense[len(s.dense)-1]
		s.dense[i] = y
		s.sparse[y.key] = i
		s.dense = s.dense[:len(s.dense)-1]
	}
}

func (s *sparseMap) clear() {
	s.dense = s.dense[:0]
}

func (s *sparseMap) contents() []sparseEntry {
	return s.dense
}
