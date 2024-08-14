// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast_test

import (
	. "go/ast"
	"go/parser"
	"go/token"
	"testing"
)

var comments = []struct {
	list []string
	text string
}{
	{[]string{"//"}, ""},
	{[]string{"//   "}, ""},
	{[]string{"//", "//", "//   "}, ""},
	{[]string{"// foo   "}, "foo\n"},
	{[]string{"//", "//", "// foo"}, "foo\n"},
	{[]string{"// foo  bar  "}, "foo  bar\n"},
	{[]string{"// foo", "// bar"}, "foo\nbar\n"},
	{[]string{"// foo", "//", "//", "//", "// bar"}, "foo\n\nbar\n"},
	{[]string{"// foo", "/* bar */"}, "foo\n bar\n"},
	{[]string{"//", "//", "//", "// foo", "//", "//", "//"}, "foo\n"},

	{[]string{"/**/"}, ""},
	{[]string{"/*   */"}, ""},
	{[]string{"/**/", "/**/", "/*   */"}, ""},
	{[]string{"/* Foo   */"}, " Foo\n"},
	{[]string{"/* Foo  Bar  */"}, " Foo  Bar\n"},
	{[]string{"/* Foo*/", "/* Bar*/"}, " Foo\n Bar\n"},
	{[]string{"/* Foo*/", "/**/", "/**/", "/**/", "// Bar"}, " Foo\n\nBar\n"},
	{[]string{"/* Foo*/", "/*\n*/", "//", "/*\n*/", "// Bar"}, " Foo\n\nBar\n"},
	{[]string{"/* Foo*/", "// Bar"}, " Foo\nBar\n"},
	{[]string{"/* Foo\n Bar*/"}, " Foo\n Bar\n"},

	{[]string{"// foo", "//go:noinline", "// bar", "//:baz"}, "foo\nbar\n:baz\n"},
	{[]string{"// foo", "//lint123:ignore", "// bar"}, "foo\nbar\n"},
}

func TestCommentText(t *testing.T) {
	for i, c := range comments {
		list := make([]*Comment, len(c.list))
		for i, s := range c.list {
			list[i] = &Comment{Text: s}
		}

		text := (&CommentGroup{list}).Text()
		if text != c.text {
			t.Errorf("case %d: got %q; expected %q", i, text, c.text)
		}
	}
}

var directiveTests = []struct {
	in               string
	tool, name, args string // name="" => not a directive
}{
	{"abc",
		"", "", ""},
	{"go:inline",
		"go", "inline", ""},
	{"Go:inline",
		"", "", ""},
	{"go:Inline",
		"", "", ""},
	{":inline",
		"", "", ""},
	{"lint:ignore",
		"lint", "ignore", ""},
	{"lint:1234",
		"lint", "1234", ""},
	{"1234:lint",
		"1234", "lint", ""},
	{"go: inline",
		"", "", ""},
	{"go:",
		"", "", ""},
	{"go:*",
		"", "", ""},
	{"go:x*",
		"go", "x*", ""},
	{"export foo",
		"", "export", "foo"},
	{"extern foo",
		"", "extern", "foo"},
	{"expert foo",
		"", "", ""},
	{"tool:name     args with  spaces ",
		"tool", "name", "args with  spaces"},
	{"foo file.go:1234",
		"", "", ""},
	// //line directives get swallowed by the scanner
	// and are not part of the comment.
	// {"line file.go:1234",
	// 	"", "line", "file.go:1234"},
}

// TestIsDirective exercises the internal isDirective helper function.
func TestIsDirective(t *testing.T) {
	for _, test := range directiveTests {
		want := test.name != ""
		if got := IsDirective(test.in); got != want {
			t.Errorf("isDirective(%q) = %v, want %v", test.in, got, want)
		}
	}
}

func TestCommentGroup_Directives(t *testing.T) {
	fset := token.NewFileSet()
	for _, test := range directiveTests {
		src := "package p\n\n//" + test.in + "\nfunc f()"
		f, err := parser.ParseFile(fset, "a.go", src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		doc := f.Decls[0].(*FuncDecl).Doc
		for dir := range doc.Directives() {
			// The column should always be 1 in these tests.
			col := fset.Position(dir.Pos).Column
			if col != 1 {
				t.Errorf("%q: column was %d, want 1", test.in, col)
			}
			if dir.Tool != test.tool ||
				dir.Name != test.name ||
				dir.Args != test.args {
				t.Errorf("%q: got %#v, want tool=%q name=%q args=%q",
					test.in, dir, test.tool, test.name, test.args)
			}
		}
	}
}
