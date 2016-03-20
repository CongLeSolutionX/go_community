// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
This file contains the code to check for bad usage of append.
*/

package main

import (
	"go/ast"
	"go/token"
	"reflect"
)

func init() {
	register("append",
		"check for bad usages of append",
		checkAppendStmt,
		assignStmt)
}

// checkAppendStmt checks for bad usage of append like b = append(a, sth)
func checkAppendStmt(f *File, node ast.Node) {
	stmt := node.(*ast.AssignStmt)
	if stmt.Tok != token.ASSIGN {
		return // only check =
	}
	if len(stmt.Lhs) != len(stmt.Rhs) {
		// If LHS and RHS have different cardinality, it should not be a legal assignment statement.
		return
	}
	for i, lhs := range stmt.Lhs {
		rhs := stmt.Rhs[i]
		call, ok := rhs.(*ast.CallExpr)
		if !ok {
			continue
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok {
			continue
		}
		if id.Name != "append" {
			continue
		}
		if len(call.Args) < 2 {
			continue
		}
		arg0 := call.Args[0]
		if reflect.TypeOf(lhs) != reflect.TypeOf(arg0) {
			continue // short-circuit the heavy-weight gofmt check
		}
		le := f.gofmt(lhs)
		re := f.gofmt(arg0)
		if le != re {
			f.Badf(stmt.Pos(), "bad usage of append: append element to %s but assign to %s", re, le)
		}
	}
}
