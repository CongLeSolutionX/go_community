// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "testing"

func TestDecomposeUserStruct(t *testing.T) {
	c := testConfig(t)
	typ3 := &TypeImpl{Size_: 32, Name: "V", struct_: true, fields: []StructField{{Type: TypeInt64, Name: "y"}}}
	typ2 := &TypeImpl{Size_: 32, Name: "U", struct_: false, array: true, numElem: 1, Elem_: typ3}
	typ := &TypeImpl{Size_: 32, Name: "T", struct_: true, fields: []StructField{{Type: typ2, Name: "x"}}}

	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, TypeMem, 0, nil),
			Valu("arg", OpArg, typ, 0, nil),
			Valu("v1", OpStructSelect, typ2, 0, nil, "arg"),
			Valu("v2", OpArraySelect, typ3, 0, nil, "v1"),
			Valu("v3", OpStructSelect, TypeInt64, 0, nil, "v2"),
			Goto("exit")),
		Bloc("exit",
			Exit("v3")))

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

	if got := len(fun.f.Names); got != 1 {
		t.Errorf("expected 2 names, got %d", got)
	} else {
		exp := "arg.x[0].y"
		if got := fun.f.Names[0].N.String(); got != exp {
			t.Errorf("expected name %s, got %s", exp, got)
		}
	}
}
