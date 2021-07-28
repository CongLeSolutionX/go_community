// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"strings"
	"testing"
)

var testTerms = map[string]*term{
	"∅":       nil,
	"⊤":       &term{},
	"int":     &term{false, Typ[Int]},
	"~int":    &term{true, Typ[Int]},
	"string":  &term{false, Typ[String]},
	"~string": &term{true, Typ[String]},
}

func TestTermString(t *testing.T) {
	for r, x := range testTerms {
		if rr := x.String(); rr != r {
			t.Errorf("%v.String() == %v; want %v", x, rr, r)
		}
	}
}

func split(s string, n int) []string {
	r := strings.Split(s, " ")
	if len(r) != n {
		panic("invalid test case: " + s)
	}
	return r
}

func testTerm(name string) *term {
	r, ok := testTerms[name]
	if !ok {
		panic("invalid test argument: " + name)
	}
	return r
}

func TestTermEqual(t *testing.T) {
	for _, test := range []string{
		"∅ ∅ T",
		"⊤ ⊤ T",
		"int int T",
		"~int ~int T",
		"∅ ⊤ F",
		"∅ int F",
		"∅ ~int F",
		"⊤ int F",
		"⊤ ~int F",
		"int ~int F",
	} {
		args := split(test, 3)
		x := testTerm(args[0])
		y := testTerm(args[1])
		r := args[2] == "T"
		if rr := x.equal(y); rr != r {
			t.Errorf("%v.equal(%v) = %v; want %v", x, y, rr, r)
		}
		// equal is symmetric
		x, y = y, x
		if rr := x.equal(y); rr != r {
			t.Errorf("%v.equal(%v) = %v; want %v", x, y, rr, r)
		}
	}
}

func TestTermUnion(t *testing.T) {
	for _, test := range []string{
		"∅ ∅ ∅ ∅",
		"∅ ⊤ ⊤ ∅",
		"∅ int int ∅",
		"∅ ~int ~int ∅",
		"⊤ ⊤ ⊤ ∅",
		"⊤ int ⊤ ∅",
		"⊤ ~int ⊤ ∅",
		"int int int ∅",
		"int ~int ~int ∅",
		"int string int string",
		"int ~string int ~string",
		"~int ~string ~int ~string",

		// union is symmetric, but the result order isn't - repeat symmetric cases explictly
		"⊤ ∅ ⊤ ∅",
		"int ∅ int ∅",
		"~int ∅ ~int ∅",
		"int ⊤ ⊤ ∅",
		"~int ⊤ ⊤ ∅",
		"~int int ~int ∅",
		"string int string int",
		"~string int ~string int",
		"~string ~int ~string ~int",
	} {
		args := split(test, 4)
		x := testTerm(args[0])
		y := testTerm(args[1])
		r := testTerm(args[2])
		s := testTerm(args[3])
		if rr, ss := x.union(y); !rr.equal(r) || !ss.equal(s) {
			t.Errorf("%v.union(%v) = %v, %v; want %v, %v", x, y, rr, ss, r, s)
		}
	}
}

func TestTermIntersection(t *testing.T) {
	for _, test := range []string{
		"∅ ∅ ∅",
		"∅ ⊤ ∅",
		"∅ int ∅",
		"∅ ~int ∅",
		"⊤ ⊤ ⊤",
		"⊤ int int",
		"⊤ ~int ~int",
		"int int int",
		"int ~int int",
		"int string ∅",
		"int ~string ∅",
		"~int ~string ∅",
	} {
		args := split(test, 3)
		x := testTerm(args[0])
		y := testTerm(args[1])
		r := testTerm(args[2])
		if rr := x.intersect(y); !rr.equal(r) {
			t.Errorf("%v.intersect(%v) = %v; want %v", x, y, rr, r)
		}
		// intersect is symmetric
		x, y = y, x
		if rr := x.intersect(y); !rr.equal(r) {
			t.Errorf("%v.intersect(%v) = %v; want %v", x, y, rr, r)
		}
	}
}

func TestTermIncludes(t *testing.T) {
	for _, test := range []string{
		"∅ int F",
		"⊤ int T",
		"int int T",
		"~int int T",
		"string int F",
		"~string int F",
	} {
		args := split(test, 3)
		x := testTerm(args[0])
		y := testTerm(args[1]).typ
		r := args[2] == "T"
		if rr := x.includes(y); rr != r {
			t.Errorf("%v.includes(%v) = %v; want %v", x, y, rr, r)
		}
	}
}

func TestTermSubsetOf(t *testing.T) {
	for _, test := range []string{
		"∅ ∅ T",
		"⊤ ⊤ T",
		"int int T",
		"~int ~int T",
		"∅ ⊤ T",
		"∅ int T",
		"∅ ~int T",
		"⊤ int F",
		"⊤ ~int F",
		"int ~int T",
	} {
		args := split(test, 3)
		x := testTerm(args[0])
		y := testTerm(args[1])
		r := args[2] == "T"
		if rr := x.subsetOf(y); rr != r {
			t.Errorf("%v.subsetOf(%v) = %v; want %v", x, y, rr, r)
		}
	}
}

func TestTermDisjoint(t *testing.T) {
	for _, test := range []string{
		"int int F",
		"~int ~int F",
		"int ~int F",
		"int string T",
		"int ~string T",
		"~int ~string T",
	} {
		args := split(test, 3)
		x := testTerm(args[0])
		y := testTerm(args[1])
		r := args[2] == "T"
		if rr := x.disjoint(y); rr != r {
			t.Errorf("%v.disjoint(%v) = %v; want %v", x, y, rr, r)
		}
		// disjoint is symmetric
		x, y = y, x
		if rr := x.disjoint(y); rr != r {
			t.Errorf("%v.disjoint(%v) = %v; want %v", x, y, rr, r)
		}
	}
}
