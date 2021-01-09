// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
)

// match reports whether types t1 and t2 are consistent
// representations for a given expression's type.
func (g *irgen) match(t1 *types.Type, t2 types2.Type, hasOK bool) bool {
	tuple, ok := t2.(*types2.Tuple)
	if !ok {
		// Not a tuple; can use simple type identity comparison.
		return types.Identical(t1, g.typ(t2))
	}

	if hasOK {
		// For has-ok values, types2 represents the expression's type as
		// a 2-element tuple, whereas ir just uses the first type and
		// infers that the second type is boolean.
		return tuple.Len() == 2 && types.Identical(t1, g.typ(tuple.At(0).Type()))
	}

	if t1 == nil || tuple == nil {
		return t1 == nil && tuple == nil
	}
	if !t1.IsFuncArgStruct() {
		return false
	}
	if t1.NumFields() != tuple.Len() {
		return false
	}
	for i, result := range t1.FieldSlice() {
		if !types.Identical(result.Type, g.typ(tuple.At(i).Type())) {
			return false
		}
	}
	return true
}

func (g *irgen) validate(n syntax.Node) {
	switch n := n.(type) {
	case *syntax.CallExpr:
		tv := g.info.Types[n.Fun]
		if tv.IsBuiltin() {
			switch builtin := n.Fun.(type) {
			case *syntax.Name:
				g.validateBuiltin(builtin.Value, n)
			case *syntax.SelectorExpr:
				g.validateBuiltin(builtin.Sel.Value, n)
			default:
				g.unhandled("builtin", n)
			}
		}
	}
}

func (g *irgen) validateBuiltin(name string, call *syntax.CallExpr) {
	switch name {
	case "Alignof", "Sizeof":
		got, ok := constant.Int64Val(g.info.Types[call].Value)
		if !ok {
			base.FatalfAt(g.pos(call), "expected int64 constant value")
		}

		typ := g.typ(g.info.Types[call.ArgList[0]].Type)
		want := typ.Size()
		if name == "Alignof" {
			want = typ.Alignment()
		}

		if got != want {
			base.FatalfAt(g.pos(call), "%s(%v): got %v from types2, but want %v", name, typ, got, want)
		}
	}
}
