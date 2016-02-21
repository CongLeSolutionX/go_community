// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"
)

func init() {
	register("malformedtest",
		"check for test and benchmark functions with malformed names",
		checkTest,
		funcDecl)
}

func isTestSuffix(name string) bool {
	if len(name) == 0 {
		// "Test" is ok.
		return true
	}
	r, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsLower(r)
}

func isTestParam(typ ast.Expr, wantType string) bool {
	ptr, ok := typ.(*ast.StarExpr)
	if !ok {
		return false
	}
	// No easy way of making sure it's a *testing.T or *testing.B:
	// make sure the name of the type matches.
	if name, ok := ptr.X.(*ast.Ident); ok {
		return name.Name == wantType
	}
	if sel, ok := ptr.X.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == wantType
	}
	return false
}

// checkTest checks for Test-like functions that aren't run because
// they have malformed names.
func checkTest(f *File, node ast.Node) {
	if !strings.HasSuffix(f.name, "_test.go") {
		return
	}
	fn, ok := node.(*ast.FuncDecl)
	if !ok {
		// Ignore non-functions.
		return
	}

	var wantType, prefix string
	switch {
	case strings.HasPrefix(fn.Name.Name, "Test"):
		wantType = "T"
		prefix = "Test"
	case strings.HasPrefix(fn.Name.Name, "Benchmark"):
		wantType = "B"
		prefix = "Benchmark"
	default:
		return
	}

	// Want functions with no receivers, 0 results and 1 parameter.
	if fn.Recv != nil ||
		fn.Type.Results != nil && len(fn.Type.Results.List) > 0 ||
		fn.Type.Params == nil ||
		len(fn.Type.Params.List) != 1 ||
		len(fn.Type.Params.List[0].Names) > 1 {
		return
	}

	// The param must look like a *testing.T or *testing.B.
	if !isTestParam(fn.Type.Params.List[0].Type, wantType) {
		return
	}

	if !isTestSuffix(fn.Name.Name[len(prefix):]) {
		f.Badf(node.Pos(), "%s has malformed name: first letter after '%s' must not be lowercase", fn.Name.Name, prefix)
	}
}
