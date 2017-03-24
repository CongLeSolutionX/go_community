// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
This file contains the code to check for missing error assignments in if
statements. For example

	if f(); err != nil {
		...
	}

will compile if err is declared in the outer scope. But it is likely that
the programmer meant

	if err := f(); err != nil {
		...
	}
*/

package main

import "go/ast"

func init() {
	register("missingerr",
		"check for possibly missing error assignments in if statements",
		checkMissingErr,
		ifStmt)
}

// checkMissingErr looks for a function call without assignment in the init
// statement and an error comparison in the condition expression.
// It's likely that the init statement is missing an assignment to the
// error variable.
func checkMissingErr(f *File, node ast.Node) {
	is := node.(*ast.IfStmt)
	es, ok := is.Init.(*ast.ExprStmt)
	if !ok {
		return
	}
	_, ok = es.X.(*ast.CallExpr)
	if !ok {
		return
	}
	be, ok := is.Cond.(*ast.BinaryExpr)
	if !ok {
		return
	}
	lhs, ok := be.X.(*ast.Ident)
	if !ok {
		return
	}
	rhs, ok := be.Y.(*ast.Ident)
	if !ok {
		return
	}
	if lhs.Name == "err" && rhs.Name == "nil" {
		f.Badf(lhs.Pos(), "possibly missing error assignment")
	}
	return
}
