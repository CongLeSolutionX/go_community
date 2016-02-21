// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"regexp"
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

func isTestSuffix(s string) bool {
	if len(s) == 0 {
		// "Test" is ok.
		return true
	}
	r, _ := utf8.DecodeRuneInString(s)
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

	var testRE = regexp.MustCompile(`(Test|Benchmark)(.*)`)
	match := testRE.FindStringSubmatch(fn.Name.Name)
	if match == nil || len(match) != 3 {
		// Ignore functions not starting with Test|Benchmark: they are
		// "clearly" not meant to be tests.
		return
	}

	prefix := match[1]
	tstName := match[2]
	if !isTestSuffix(tstName) {
		f.Badf(node.Pos(), "%s has malformed test name: first letter after '%s' must not be lowercase: %s", fn.Name.Name, prefix, tstName)
	}
}
