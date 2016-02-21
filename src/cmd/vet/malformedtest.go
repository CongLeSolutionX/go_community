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

func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == 0 {
		// "Test" is ok.
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}

// checkTest checks for Test-like functions that aren't run because
// they have malformed names.
func checkTest(f *File, node ast.Node) {
	if !strings.HasSuffix(f.name, "_test.go") {
		return
	}
	fn, ok := node.(*ast.FuncDecl)
	if !ok || fn.Recv != nil {
		// Ignore non-functions or functions with receivers.
		return
	}

	name := fn.Name.Name
	switch {
	case strings.HasPrefix(name, "Test") && !isTest(name, "Test"):
		f.Badf(node.Pos(), "%s has malformed test name: first letter after 'Test' must not be lowercase", name)
	case strings.HasPrefix(name, "Benchmark") && !isTest(name, "Benchmark"):
		f.Badf(node.Pos(), "%s has malformed benchmark name: first letter after 'Benchmark' must not be lowercase", name)
	}
}
