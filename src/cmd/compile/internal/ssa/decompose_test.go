// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "testing"

func TestDecomposeUserStruct(t *testing.T) {
	c := testConfig(t)
	typ2 := &TypeImpl{Size_: 32, Name: "U", struct_: true, fields: []StructField{{Type: TypeInt64, Name: "y"}}}
	typ := &TypeImpl{Size_: 32, Name: "T", struct_: true, fields: []StructField{{Type: typ2, Name: "x"}}}

	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, TypeMem, 0, nil),
			Valu("arg", OpArg, typ, 0, nil),
			Valu("v1", OpStructSelect, typ2, 0, nil, "arg"),
			Valu("v2", OpStructSelect, TypeInt64, 0, nil, "v1"),
			Goto("exit")),
		Bloc("exit",
			Exit("v2")))

	for k, v := range fun.values {
		if k == "arg" {
			loc := LocalSlot{Off: 0, Type: typ, N: &DummyAuto{s: "arg"}}
			values, ok := fun.f.NamedValues[loc]
			if !ok {
				fun.f.Names = append(fun.f.Names, loc)
			}
			fun.f.NamedValues[loc] = append(values, v)
		}
	}

	if got := len(fun.f.Names); got != 1 {
		t.Errorf("expected 1 name, got %d", got)
	}
	decomposeUser(fun.f)

	// Ideally we would also check the names here as well as the number, however to set
	// the names we need to pass a gc.Node instead of the DummyAuto which would
	// create a cyclic dependency.
	if got := len(fun.f.Names); got != 2 {
		t.Errorf("expected 2 names, got %d", got)
	}
}
