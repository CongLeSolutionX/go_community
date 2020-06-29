// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"strings"
)

func init() {
	register(buildtagFix)
}

const buildtagGoVersionCutoff = 1_99 // TODO fix

var buildtagFix = fix{
	name: "buildtag",
	date: "2020-06-29", // TODO fix
	f:    buildtag,
	desc: `Remove +build comments from modules using Go 1.(N+1)`, // TODO fix
}

func buildtag(f *ast.File) bool {
	if goVersion < buildtagGoVersionCutoff {
		return false
	}

	// File is already gofmt-ed, so we know that if there are +build lines,
	// they are in a comment group that starts with a //go:build line followed
	// by a blank line. While we cannot delete comments from an AST and
	// expect consistent output in general, this specific case - deleting only
	// some lines from a comment block - does format correctly.
	fixed := false
	for _, g := range f.Comments {
		sawGoBuild := false
		for i, c := range g.List {
			if strings.HasPrefix(c.Text, "//go:build ") {
				sawGoBuild = true
			}
			if sawGoBuild && strings.HasPrefix(c.Text, "// +build ") {
				g.List = g.List[:i]
				fixed = true
				break
			}
		}
	}

	return fixed
}
