// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"testing"
)

func TestIsAlias(t *testing.T) {
	check := func(obj *types2.TypeName, want bool) {
		if got := obj.IsAlias(); got != want {
			t.Errorf("%v: got IsAlias = %v; want %v", obj, got, want)
		}
	}

	// predeclared types
	check(types2.Unsafe.Scope().Lookup("Pointer").(*types2.TypeName), false)
	for _, name := range types2.Universe.Names() {
		if obj, _ := types2.Universe.Lookup(name).(*types2.TypeName); obj != nil {
			check(obj, name == "any" || name == "byte" || name == "rune")
		}
	}

	// for convenience
	newTN := types2.NewTypeName

	// various other types
	pkg := types2.NewPackage("p", "p")
	t1 := newTN(nopos, pkg, "t1", nil)
	n1 := types2.NewNamed(t1, new(types2.Struct), nil)
	t5 := newTN(nopos, pkg, "t5", nil)
	types2.NewTypeParam(t5, nil)
	for _, test := range []struct {
		name  *types2.TypeName
		alias bool
	}{
		{newTN(nopos, nil, "t0", nil), false}, // no type yet
		{newTN(nopos, pkg, "t0", nil), false}, // no type yet
		{t1, false},                           // type name refers to named type and vice versa
		{newTN(nopos, nil, "t2", types2.NewInterfaceType(nil, nil)), true}, // type name refers to unnamed type
		{newTN(nopos, pkg, "t3", n1), true},                                // type name refers to named type with different type name
		{newTN(nopos, nil, "t4", types2.Typ[types2.Int32]), true},          // type name refers to basic type with different name
		{newTN(nopos, nil, "int32", types2.Typ[types2.Int32]), false},      // type name refers to basic type with same name
		{newTN(nopos, pkg, "int32", types2.Typ[types2.Int32]), true},       // type name is declared in user-defined package (outside Universe)
		{newTN(nopos, nil, "rune", types2.Typ[types2.Rune]), true},         // type name refers to basic type rune which is an alias already
		{t5, false}, // type name refers to type parameter and vice versa
	} {
		check(test.name, test.alias)
	}
}

// TestEmbeddedMethod checks that an embedded method is represented by
// the same Func Object as the original method. See also issue #34421.
func TestEmbeddedMethod(t *testing.T) {
	const src = `package p; type I interface { error }`

	// type-check src
	f, err := parseSrc("", src)
	if err != nil {
		t.Fatalf("parse failed: %s", err)
	}
	var conf types2.Config
	pkg, err := conf.Check(f.PkgName.Value, []*syntax.File{f}, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %s", err)
	}

	// get original error.Error method
	eface := types2.Universe.Lookup("error")
	orig, _, _ := types2.LookupFieldOrMethod(eface.Type(), false, nil, "Error")
	if orig == nil {
		t.Fatalf("original error.Error not found")
	}

	// get embedded error.Error method
	iface := pkg.Scope().Lookup("I")
	embed, _, _ := types2.LookupFieldOrMethod(iface.Type(), false, nil, "Error")
	if embed == nil {
		t.Fatalf("embedded error.Error not found")
	}

	// original and embedded Error object should be identical
	if orig != embed {
		t.Fatalf("%s (%p) != %s (%p)", orig, orig, embed, embed)
	}
}
