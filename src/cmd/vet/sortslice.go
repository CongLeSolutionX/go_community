// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the check for http.Response values being used before
// checking for errors.

package main

import (
	"go/ast"
	"go/types"
)

func init() {
	register("sortslice",
		"check errors on calls to sort.Slice",
		checkSortSlice, callExpr)
}

func checkSortSlice(f *File, node ast.Node) {
	call := node.(*ast.CallExpr)
	if !checkCallToSortSlice(call) {
		return
	}

	arg := call.Args[0]
	typ := f.pkg.types[arg].Type
	if _, ok := typ.(*types.Slice); !ok {
		f.Badf(arg.Pos(), "first argument to sort.Slice should be a slice")
	}
}

func checkCallToSortSlice(call *ast.CallExpr) bool {
	sel, _ := call.Fun.(*ast.SelectorExpr)
	if sel == nil {
		return false
	}
	pkg, _ := sel.X.(*ast.Ident)
	return pkg != nil && pkg.Obj == nil &&
		pkg.Name == "sort" &&
		sel.Sel.Name == "Slice"
}
