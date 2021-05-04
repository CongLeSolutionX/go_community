// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"sort"
	"strings"
)

// TODO(gri) This might be simpler and more efficient with a map.
//           (Map key is the type, and map value indicates exactness).

// A typeSet represents a set of exact and approximate (under) types.
// The zero value of a typeSet is a ready to use empty type set.
type typeSet struct {
	exact []Type
	under []Type
}

func (s *typeSet) String() string {
	if s.isTop() {
		return "⊤"
	}

	// sort entries for canonical output
	var list []string
	for _, e := range s.exact {
		list = append(list, e.String())
	}
	for _, e := range s.under {
		list = append(list, "~"+e.String())
	}
	sort.Strings(list)

	var b strings.Builder
	b.WriteByte('{')
	for i, s := range list {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(s)
	}
	b.WriteByte('}')
	return b.String()
}

// newTypeSet returns a new type set consisting of the exact and
// approximate (under) types provided.
func newTypeSet(exact, under []Type) *typeSet {
	s := new(typeSet)
	s.exact = exact
	s.under = under
	return s
}

// copy returns a copy of type set s.
func (s *typeSet) copy() *typeSet {
	if s.isTop() {
		return s
	}

	r := new(typeSet)
	if !s.isTop() {
		r.exact = make([]Type, len(s.exact))
		r.under = make([]Type, len(s.under))
		copy(r.exact, s.exact)
		copy(r.under, s.under)
	}
	return r
}

// addType adds type t to the type set s.
// The exact argument indicates whether t is an exact or approximate (~t) type.
func (s *typeSet) addType(t Type, exact bool) {
	if s.isTop() {
		panic("internal error: addType called on top type set")
	}

	if exact {
		s.exact = append(s.exact, t)
		return
	}
	s.under = append(s.under, t)
}

// isTop reports whether type set s is the set of all types.
func (s *typeSet) isTop() bool {
	return s == nil
}

// size returns the number of elements in the type set s.
// A ~T element counts as 1.
func (s *typeSet) size() int {
	if s.isTop() {
		panic("internal error: size called on top type set")
	}

	return len(s.exact) + len(s.under)
}

// hasType reports whether type set s includes type t.
func (s *typeSet) hasType(t Type, exact bool) bool {
	if s.isTop() {
		return true
	}

	if exact {
		for _, e := range s.exact {
			if t == e || Identical(t, e) {
				return true
			}
		}
	}
	u := under(t)
	for _, e := range s.under {
		if u == e || Identical(u, e) {
			return true
		}
	}
	return false
}

// intersect returns the intersection a ∩ b of the type sets a and b.
func (a *typeSet) intersect(b *typeSet) *typeSet {
	switch {
	case a.isTop():
		return b
	case b.isTop():
		return a
	}

	// TODO(gri) fix quadratic algorithm
	r := new(typeSet)
	a.walk(func(t Type, exact bool) bool {
		if b.hasType(t, exact) {
			r.addType(t, exact)
		}
		return true
	})
	return r
}

// union returns the union a ∪ b of the type sets a and b.
func (a *typeSet) union(b *typeSet) *typeSet {
	switch {
	case a.isTop():
		return a
	case b.isTop():
		return b
	}

	// TODO(gri) fix quadratic algorithm
	r := a.copy()
	b.walk(func(t Type, exact bool) bool {
		if !r.hasType(t, exact) {
			r.addType(t, exact)
		}
		return true
	})
	return r
}

// isSubsetOf reports whether type set a is (non-strict) subset of type set b: a ⊆ b.
func (a *typeSet) isSubsetOf(b *typeSet) bool {
	switch {
	case a.isTop():
		return b.isTop()
	case b.isTop():
		return true
		// We could check if a.size() > b.size() but that only
		// works if we are guaranteed that there are no duplicates.
		// Don't do this for now.
	}

	// TODO(gri) fix quadratic algorithm
	return a.walk(func(t Type, exact bool) bool {
		return b.hasType(t, exact)
	})
}

// isSameAs reports whether the type sets a and b contain the same types.
func (a *typeSet) isSameAs(b *typeSet) bool {
	return a.isSubsetOf(b) && b.isSubsetOf(a)
}

// walk calls f for each type in the type set of s. Iteration stops early
// if f returns false. The result of walk is false if f ever returned false;
// otherwise it is true.
func (s *typeSet) walk(f func(Type, bool) bool) bool {
	if s.isTop() {
		panic("internal error: walk called on top type set")
	}

	for _, e := range s.exact {
		if !f(e, true) {
			return false
		}
	}
	for _, e := range s.under {
		if !f(e, false) {
			return false
		}
	}
	return true
}
