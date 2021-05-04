// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
)

// A Union represents a union of types.
type Union struct {
	types []Type // types are unique
	tilde []bool // tilde[i] means types[i] is of the form ~T
}

func (t *Union) Underlying() Type { return t }
func (t *Union) String() string   { return TypeString(t, nil) }

func NewUnion(types []Type, tilde []bool) *Union {
	t := new(Union)
	t.types = types
	t.tilde = tilde
	return t
}

func parseUnion(check *Checker, tlist []syntax.Expr) Type {
	// single types
	if len(tlist) == 1 && !isTilde(tlist[0]) {
		return check.anyType(tlist[0])
	}

	utyp := new(Union)
	for _, t := range tlist {
		tilde := false
		if isTilde(t) {
			t = t.(*syntax.Operation).X
			tilde = true
		}
		utyp.types = append(utyp.types, check.anyType(t))
		utyp.tilde = append(utyp.tilde, tilde)
	}

	// Ensure that each type is only present once in the type list.
	// It's ok to do this check at the end because it's not a requirement
	// for correctness of the code.
	// Note: This is a quadratic algorithm, but type lists tend to be short.
	// TODO(gri) move this into a separate verify method
	check.later(func() {
		for i, t := range utyp.types {
			if t == Typ[Invalid] {
				continue
			}
			if includes(utyp.types[:i], t) {
				check.softErrorf(tlist[i], "duplicate type %s in union element", t)
			}
		}
	})

	return utyp
}

// isTilde reports whether x is of the form ~T.
func isTilde(x syntax.Expr) bool {
	o, _ := x.(*syntax.Operation)
	return o != nil && o.Op == syntax.Tilde
}

type iser interface {
	is(func(Type) bool) bool
}

// is reports whether all types in the type set represented by union u satisfy predicate pred.
func (u *Union) is(pred func(Type) bool) bool {
	if len(u.types) == 0 {
		return false
	}
	for _, e := range u.types {
		if e, _ := e.(iser); e != nil {
			if !e.is(pred) {
				return false
			}
			continue
		}
		if !pred(e) {
			return false
		}
	}
	return true
}

type hasTyper interface {
	hasType(Type) bool
}

// has reports whether type T is an element of the union.
func (u *Union) hasType(T Type) bool {
	if len(u.types) == 0 {
		return true
	}
	Tu := under(T)
	for i, e := range u.types {
		t := T
		if u.tilde[i] {
			t = Tu
		}
		if e, _ := e.(hasTyper); e != nil {
			if e.hasType(t) {
				return true
			}
			continue
		}
		if Identical(e, t) {
			return true
		}
	}
	return false
}

func (u *Union) lookupMethod(pkg *Package, name string) (int, *Func) {
	if len(u.types) == 0 {
		return 0, nil
	}
	var m0 *Func
	for _, e := range u.types {
		if e, _ := e.(*Interface2); e != nil {
			_, m := e.lookupMethod(pkg, name)
			if m == nil || m0 != nil && !Identical(m0.typ, m.typ) {
				return 0, nil
			}
			m0 = m
		}
	}
	return 0, m0
}
