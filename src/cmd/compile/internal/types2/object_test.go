// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"internal/testenv"
	"strings"
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

var testObjects = []struct {
	src  string
	obj  string
	want string
}{
	{"import \"io\"; var r io.Reader", "r", "var p.r io.Reader"},

	{"const c = 1.2", "c", "const p.c untyped float"},
	{"const c float64 = 3.14", "c", "const p.c float64"},

	{"type t struct{f int}", "t", "type p.t struct{f int}"},
	{"type t func(int)", "t", "type p.t func(int)"},
	{"type t[P any] struct{f P}", "t", "type p.t[P interface{}] struct{f P}"},
	{"type t[P any] struct{f P}", "t.P", "type parameter P interface{}"},
	{"type C interface{m()}; type t[P C] struct{}", "t.P", "type parameter P p.C"},

	{"type t = struct{f int}", "t", "type p.t = struct{f int}"},
	{"type t = func(int)", "t", "type p.t = func(int)"},

	{"var v int", "v", "var p.v int"},

	{"func f(int) string", "f", "func p.f(int) string"},
	{"func g[P any](x P){}", "g", "func p.g[P interface{}](x P)"},
	{"func g[P interface{~int}](x P){}", "g.P", "type parameter P interface{~int}"},
}

func TestObjectString(t *testing.T) {
	testenv.MustHaveGoBuild(t)

	for _, test := range testObjects {
		src := "package p; " + test.src
		pkg, err := makePkg(src)
		if err != nil {
			t.Errorf("%s: %s", src, err)
			continue
		}

		names := strings.Split(test.obj, ".")
		if len(names) != 1 && len(names) != 2 {
			t.Errorf("%s: invalid object path %s", test.src, test.obj)
			continue
		}
		obj := pkg.Scope().Lookup(names[0])
		if obj == nil {
			t.Errorf("%s: %s not found", test.src, names[0])
			continue
		}
		if len(names) == 2 {
			if typ, ok := obj.Type().(interface{ TypeParams() *types2.TypeParamList }); ok {
				obj = lookupTypeParamObj(typ.TypeParams(), names[1])
				if obj == nil {
					t.Errorf("%s: %s not found", test.src, test.obj)
					continue
				}
			} else {
				t.Errorf("%s: %s has no type parameters", test.src, names[0])
				continue
			}
		}

		if got := obj.String(); got != test.want {
			t.Errorf("%s: got %s, want %s", test.src, got, test.want)
		}
	}
}

func lookupTypeParamObj(list *types2.TypeParamList, name string) types2.Object {
	for i := 0; i < list.Len(); i++ {
		tpar := list.At(i)
		if tpar.Obj().Name() == name {
			return tpar.Obj()
		}
	}
	return nil
}
