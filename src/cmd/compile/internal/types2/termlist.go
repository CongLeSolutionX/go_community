// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

// A termlist represents the type set represented by the union
// t1 ∪ y2 ∪ ... tn of the type sets of the terms t1 to tn.
// A termlist is in normal form if all terms are disjoint.
type termlist []*term

func (xl termlist) String() string {
	// use a fake union for printing
	return TypeString(&Union{([]*term)(xl), nil}, nil)
}

// isEmpty reports whether the termlist xl represents the empty set of types.
func (xl termlist) isEmpty() bool {
	return len(xl) == 0
}

// isTop reports whether the termlist xl represents the set of all types.
func (xl termlist) isTop() bool {
	// If there's an unrestricted term, the entire list is unrestricted.
	for _, x := range xl {
		if x.typ == nil {
			return true
		}
	}
	return false
}

// If the type set represented by xl is specified by a single type,
// structuralType returns that type. Otherwise it returns nil.
func (xl termlist) structuralType() Type {
	if len(xl) == 1 {
		return xl[0].typ
	}
	return nil
}

// equal reports whether xl and yl represent the same type set.
func (xl termlist) equal(yl termlist) bool {
	// TODO(gri) this should be more efficient
	return xl.subsetOf(yl) && yl.subsetOf(xl)
}

// union returns the union xl ∪ yl.
func (xl termlist) union(yl termlist) termlist {
	return append(xl, yl...)
}

// intersect returns the intersection xl ∩ yl.
func (xl termlist) intersect(yl termlist) termlist {
	if xl.isEmpty() || yl.isEmpty() {
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

// includes reports whether t ∈ xl.
func (xl termlist) includes(t Type) bool {
	for _, x := range xl {
		if x.includes(t) {
			return true
		}
	}
	return false
}

// subsetOf reports whether xl ⊆ yl.
func (xl termlist) subsetOf(yl termlist) bool {
	if yl.isEmpty() {
		return xl.isEmpty()
	}

	// each term x of xl must be a subset of yl
	for _, x := range xl {
		if !yl.supersetOf(x) {
			return false // x is not a subset yl
		}
	}
	return true
}

// supersetOf reports whether y ⊆ xl.
func (xl termlist) supersetOf(y *term) bool {
	for _, x := range xl {
		if y.subsetOf(x) {
			return true
		}
	}
	return false
}
