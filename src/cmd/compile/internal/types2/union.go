// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "cmd/compile/internal/syntax"

// ----------------------------------------------------------------------------
// API

// A Union represents a union of terms.
// A term is a type with a ~ (tilde) flag.
type Union struct {
	types []Type // types are unique
	tilde []bool // if tilde[i] is set, terms[i] is of the form ~T
}

// NewUnion returns a new Union type with the given terms (types[i], tilde[i]).
// The lengths of both arguments must match. An empty union represents the set
// of no types.
func NewUnion(types []Type, tilde []bool) *Union { return newUnion(types, tilde) }

func (u *Union) IsEmpty() bool           { return len(u.types) == 0 }
func (u *Union) NumTerms() int           { return len(u.types) }
func (u *Union) Term(i int) (Type, bool) { return u.types[i], u.tilde[i] }

func (u *Union) Underlying() Type { return u }
func (u *Union) String() string   { return TypeString(u, nil) }

// ----------------------------------------------------------------------------
// Implementation

var emptyUnion = new(Union)

func newUnion(types []Type, tilde []bool) *Union {
	assert(len(types) == len(tilde))
	if len(types) == 0 {
		return emptyUnion
	}
	t := new(Union)
	t.types = types
	t.tilde = tilde
	return t
}

// is reports whether f returned true for all terms (type, tilde) of u.
func (u *Union) is(f func(Type, bool) bool) bool {
	if u.IsEmpty() {
		return false
	}
	for i, t := range u.types {
		if !f(t, u.tilde[i]) {
			return false
		}
	}
	return true
}

// is reports whether f returned true for the underlying types of all terms of u.
func (u *Union) underIs(f func(Type) bool) bool {
	if u.IsEmpty() {
		return false
	}
	for _, t := range u.types {
		if !f(under(t)) {
			return false
		}
	}
	return true
}

func parseUnion(check *Checker, tlist []syntax.Expr) Type {
	var types []Type
	var tilde []bool
	for _, x := range tlist {
		t, d := parseTilde(check, x)
		if len(tlist) == 1 && !d {
			return t // single type
		}
		types = append(types, t)
		tilde = append(tilde, d)
	}

	// Ensure that each type is only present once in the type list.
	// It's ok to do this check at the end because it's not a requirement
	// for correctness of the code.
	// Note: This is a quadratic algorithm, but unions tend to be short.
	check.later(func() {
		for i, t := range types {
			t := expand(t)
			if t == Typ[Invalid] {
				continue
			}

			x := tlist[i]
			pos := syntax.StartPos(x)
			// We may not know the position of x if it was a typechecker-
			// introduced ~T type of a type list entry T. Use the position
			// of T instead.
			// TODO(gri) remove this test once we don't support type lists anymore
			if !pos.IsKnown() {
				if op, _ := x.(*syntax.Operation); op != nil {
					pos = syntax.StartPos(op.X)
				}
			}

			u := under(t)
			if tilde[i] {
				// TODO(gri) enable this check once we have converted tests
				// if !Identical(u, t) {
				// 	check.errorf(x, "invalid use of ~ (underlying type of %s is %s)", t, u)
				// }
			}
			if _, ok := u.(*Interface); ok {
				check.errorf(pos, "cannot use interface %s with ~ or inside a union (implementation restriction)", t)
			}

			// Complain about duplicate entries a|a, but also a|~a, and ~a|~a.
			if includes(types[:i], t) {
				// TODO(gri) this currently doesn't print the ~ if present
				check.softErrorf(pos, "duplicate term %s in union element", t)
			}
		}
	})

	return newUnion(types, tilde)
}

func parseTilde(check *Checker, x syntax.Expr) (Type, bool) {
	tilde := false
	if op, _ := x.(*syntax.Operation); op != nil && op.Op == syntax.Tilde {
		x = op.X
		tilde = true
	}
	return check.anyType(x), tilde
}

// intersect computes the intersection of the types x and y,
// A nil type stands for the set of all types; an empty union
// stands for the set of no types.
func intersect(x, y Type) (r Type) {
	// If one of the types is nil (no restrictions)
	// the result is the other type.
	switch {
	case x == nil:
		return y
	case y == nil:
		return x
	}

	// Compute the terms which are in both x and y.
	xu, _ := x.(*Union)
	yu, _ := y.(*Union)
	switch {
	case xu != nil && yu != nil:
		// Quadratic algorithm, but good enough for now.
		var types []Type
		var tilde []bool
		for j, y := range yu.types {
			if i := xu.indexOf(y); i >= 0 {
				// Terms x[i] and y[j] match: Select the one that
				// is not a ~t because that is the intersection
				// type. If both are ~t, they are identical:
				//  T ∩  T =  T
				//  T ∩ ~t =  T
				// ~t ∩  T =  T
				// ~t ∩ ~t = ~t
				typ := y // = yu.types[j]
				if yu.tilde[j] {
					typ = xu.types[i]
				}
				types = append(types, typ)
				tilde = append(tilde, xu.tilde[i] && yu.tilde[j])
			}
		}
		return newUnion(types, tilde)

	case xu != nil:
		if xu.indexOf(y) >= 0 {
			return y
		}

	case yu != nil:
		if yu.indexOf(x) >= 0 {
			return x
		}

	default: // xu == nil && yu == nil
		if Identical(x, y) {
			return x
		}
	}

	return emptyUnion
}

// includes reports whether typ is in list.
func includes(list []Type, typ Type) bool {
	for _, e := range list {
		if Identical(typ, e) {
			return true
		}
	}
	return false
}

// indexOf reports the index of the term in the union u that matches typ.
func (u *Union) indexOf(typ Type) int {
	utyp := under(typ)
	for i, t := range u.types {
		tt := typ
		if u.tilde[i] {
			tt = utyp
		}
		if Identical(tt, t) {
			return i
		}
	}
	return -1
}
