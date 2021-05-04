// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "testing"

var testSets = map[string]*typeSet{
	"⊤":            nil,
	"{}":           newTypeSet([]Type{}, nil),
	"{int}":        newTypeSet([]Type{Typ[Int]}, nil),
	"{~int}":       newTypeSet(nil, []Type{Typ[Int]}),
	"{bool}":       newTypeSet([]Type{Typ[Bool]}, nil),
	"{~bool}":      newTypeSet(nil, []Type{Typ[Bool]}),
	"{bool, int}":  newTypeSet([]Type{Typ[Int], Typ[Bool]}, nil),
	"{int, ~bool}": newTypeSet([]Type{Typ[Int]}, []Type{Typ[Bool]}),
}

func TestTypeSetString(t *testing.T) {
	for want, s := range testSets {
		got := s.String()
		if got != want {
			t.Errorf("got %s; want %s", got, want)
		}
	}
}

func TestTypeSetIntersect(t *testing.T) {
	for _, test := range []struct {
		a, b, want string
	}{
		{"⊤", "⊤", "⊤"},
		{"⊤", "{int}", "{int}"},
		{"{int}", "⊤", "{int}"},
		{"{int}", "{bool}", "{}"},
		{"{bool}", "{int}", "{}"},
		{"{bool}", "{bool, int}", "{bool}"},
		{"{bool}", "{int, ~bool}", "{bool}"},
		{"{~bool}", "{int, ~bool}", "{~bool}"},
	} {
		a := testSets[test.a]
		b := testSets[test.b]
		got := a.intersect(b)
		want := testSets[test.want]
		if !got.isSameAs(want) {
			t.Errorf("%s ∩ %s: got %s; want %s", a, b, got, want)
		}
	}
}

func TestTypeSetUnion(t *testing.T) {
	for _, test := range []struct {
		a, b, want string
	}{
		{"⊤", "⊤", "⊤"},
		{"⊤", "{int}", "⊤"},
		{"{int}", "⊤", "⊤"},
		{"{int}", "{bool}", "{bool, int}"},
		{"{bool}", "{int}", "{bool, int}"},
		{"{bool}", "{bool, int}", "{bool, int}"},
		{"{bool}", "{int, ~bool}", "{int, ~bool}"},
		{"{~bool}", "{int, ~bool}", "{int, ~bool}"},
	} {
		a := testSets[test.a]
		b := testSets[test.b]
		got := a.union(b)
		want := testSets[test.want]
		if !got.isSameAs(want) {
			t.Errorf("%s ∪ %s: got %s; want %s", a, b, got, want)
		}
	}
}

func TestTypeSetIsSubsetOf(t *testing.T) {
	for _, test := range []struct {
		a, b string
		want bool
	}{
		{"⊤", "⊤", true},
		{"⊤", "{int}", false},
		{"{int}", "⊤", true},
		{"{int}", "{bool}", false},
		{"{bool}", "{int}", false},
		{"{bool}", "{bool, int}", true},
		{"{bool}", "{int, ~bool}", true},
		{"{~bool}", "{int, ~bool}", true},
		{"{~bool}", "{bool, int}", false},
	} {
		a := testSets[test.a]
		b := testSets[test.b]
		got := a.isSubsetOf(b)
		if got != test.want {
			t.Errorf("%s ⊆ %s: got %v; want %v", a, b, got, test.want)
		}
	}
}
