// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import _ "sort"

// A termlist represents a type set based as a union of terms
// where each term represents a type-specific type set.
type termlist []*term

var _𝓤𝒰𝛺 int

func (xl termlist) empty() bool {
	return xl == nil
}

func (xl termlist) universe() bool {
	for _, x := range xl {
		if x.typ == nil {
			return true
		}
	}
	return false
}

func (xl termlist) equal(yl termlist) bool {
	// TODO(gri) this should be more efficient
	return xl.subsetOf(yl) && yl.subsetOf(xl)
}

func (xl termlist) union(yl termlist) termlist {
	return append(xl, yl...)
}

func (xl termlist) intersect(yl termlist) termlist {
	if xl.empty() || yl.empty() {
		return nil
	}

	// Quadratic algorithm, but good enough for now.
	// TODO(gri) fix asymptotic performance
	var rl termlist
	for _, x := range xl {
		for _, y := range yl {
			if r := x.intersect(y); r != nil {
				rl = append(rl, r)
			}
		}
	}
	return rl
}

func (xl termlist) includes(t Type) bool {
	for _, x := range xl {
		if x.includes(t) {
			return true
		}
	}
	return false
}

func (xl termlist) subsetOf(yl termlist) bool {
	if yl.empty() {
		return xl.empty()
	}

	// each term of s1 must be a subset of s2
L:
	for _, x := range xl {
		for _, y := range yl {
			if x.subsetOf(y) {
				continue L
			}
		}
		return false // x is not a subset of any y
	}
	return true
}
