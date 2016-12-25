// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
This file contains the code to check for basic type comparison.
*/

package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

func init() {
	register("cmpbasic",
		"check for basic types comparisons",
		cmpBasic,
		binaryExpr)
}

func cmpBasic(f *File, node ast.Node) {
	e := node.(*ast.BinaryExpr)
	cmpUnsigned(f, e)
}

func cmpUnsigned(f *File, e *ast.BinaryExpr) {
	if rval, ok := ival(e.Y); ok && rval == 0 && e.Op == token.LSS {
		ltype := f.pkg.types[e.X].Type
		if isUnsigned(ltype) {
			f.Badf(e.Pos(), "'%v (%v) < 0' is always false", e.X, ltype)
		}
	} else if lval, ok := ival(e.X); ok && lval == 0 && e.Op == token.GTR {
		rtype := f.pkg.types[e.Y].Type
		if isUnsigned(rtype) {
			f.Badf(e.Pos(), "'0 > %v (%v)' is always false", e.Y, rtype)
		}
	}
}

// isUnsigned reports whether type is basic unsigned.
func isUnsigned(typ types.Type) bool {
	t, ok := typ.(*types.Basic)
	if !ok {
		return false
	}

	return types.Uint <= t.Kind() && t.Kind() <= types.Uintptr
}

// ival returns int value from given basic expr.
func ival(e ast.Expr) (int, bool) {
	switch n := e.(type) {
	case *ast.BasicLit:
		if n.Kind != token.INT {
			return 0, false
		}
		rval, err := strconv.Atoi(n.Value)
		if err != nil {
			return 0, false
		}
		return rval, true
	case *ast.UnaryExpr:
		if n.Op == token.SUB {
			if v, ok := ival(n.X); ok {
				return -1 * v, false
			}
		}
	}

	return 0, false
}
