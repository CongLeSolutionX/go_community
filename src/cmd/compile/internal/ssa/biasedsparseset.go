// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	// "fmt"
	"math"
)

type biasedSparseSet struct {
	s     *sparseSet
	first int
}

// returns a new biasedSparseSet for values between first and last, inclusive.
func newBiasedSparseSet(first, last int) *biasedSparseSet {
	if first > last {
		return &biasedSparseSet{first: math.MaxInt32, s: nil}
	}
	return &biasedSparseSet{first: first, s: newSparseSet(1 + last - first)}
}

func (s *biasedSparseSet) cap() int {
	return s.s.cap() + int(s.first)
}

// func (s *biasedSparseSet) size() int {
// 	return s.s.size()
// }

func (s *biasedSparseSet) contains(x uint) bool {
	if s.s == nil {
		return false
	}
	if int(x) < s.first {
		return false
	}
	if int(x) >= s.cap() {
		return false
	}
	return s.s.contains(ID(int(x) - s.first))
}

func (s *biasedSparseSet) add(x uint) {
	if int(x)-s.first < 0 || int(x) >= s.cap() {
		return
		// panic(fmt.Sprintf("biasedSparseSet.add(%d), first=%d, cap=%d", x, s.first, s.cap()))
	}
	s.s.add(ID(int(x) - s.first))
}

func (s *biasedSparseSet) remove(x uint) {
	s.s.remove(ID(int(x) - s.first))
}

func (s *biasedSparseSet) clear() {
	if s.s != nil {
		s.s.clear()
	}
}
