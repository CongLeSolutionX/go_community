// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// This file contains the check that durations are not unit-free constants.

import (
	"go/ast"
	"go/constant"
)

func init() {
	register("duration",
		"check that durations are specified with units",
		checkDurationUnits,
		callExpr)
}

// checkDurationUnits checks whether node
// contains unit-free time durations.
func checkDurationUnits(f *File, node ast.Node) {
	call := node.(*ast.CallExpr)

	// Don't complain about time.Duration(30).
	// It's usually unnecessary, but it's clearly intentional.
	if len(call.Args) == 1 {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "Duration" {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "time" {
				return
			}
		}
		ident, ok := call.Fun.(*ast.Ident)
		if ok && ident.Name == "Duration" { // time.Duration from within the time package
			return
		}
	}

	for _, arg := range call.Args {
		tv := f.pkg.types[arg]
		if tv.Type == nil || tv.Value == nil || tv.Type.String() != "time.Duration" {
			continue
		}
		x, ok := constant.Int64Val(tv.Value)
		if !ok {
			continue
		}
		if x == 0 || x > 100 {
			continue
		}
		if _, ok := arg.(*ast.BasicLit); ok {
			f.Badf(arg.Pos(), "time.Duration without unit: %v", f.gofmt(arg))
		}
	}
}
