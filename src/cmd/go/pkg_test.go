// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"strings"
	"testing"
)

var foldDupTests = []struct {
	list   []string
	f1, f2 string
}{
	{stringList("math/rand", "math/big"), "", ""},
	{stringList("math", "strings"), "", ""},
	{stringList("strings"), "", ""},
	{stringList("strings", "strings"), "strings", "strings"},
	{stringList("Rand", "rand", "math", "math/rand", "math/Rand"), "Rand", "rand"},
}

func TestFoldDup(t *testing.T) {
	for _, tt := range foldDupTests {
		f1, f2 := foldDup(tt.list)
		if f1 != tt.f1 || f2 != tt.f2 {
			t.Errorf("foldDup(%q) = %q, %q, want %q, %q", tt.list, f1, f2, tt.f1, tt.f2)
		}
	}
}

var parseMetaGoImportsTests = []struct {
	in  string
	out []metaImport
}{
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">`,
		[]metaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		<meta name="go-import" content="baz/quux git http://github.com/rsc/baz/quux">`,
		[]metaImport{
			{"foo/bar", "git", "https://github.com/rsc/foo/bar"},
			{"baz/quux", "git", "http://github.com/rsc/baz/quux"},
		},
	},
	{
		`<head>
		<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		</head>`,
		[]metaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		<body>`,
		[]metaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
}

func TestParseMetaGoImports(t *testing.T) {
	for i, tt := range parseMetaGoImportsTests {
		out, err := parseMetaGoImports(strings.NewReader(tt.in))
		if err != nil {
			t.Errorf("test#%d: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("test#%d:\n\thave %q\n\twant %q", i, out, tt.out)
		}
	}
}

func TestSharedLibName(t *testing.T) {
	// TODO(avdva) - make these values platform-specific
	prefix := "lib"
	suffix := ".so"
	testData := []struct {
		args     []string
		pkgs     []*Package
		expected string
	}{
		{
			[]string{"std"},
			[]*Package{},
			"std",
		},
		{
			[]string{"std", "cmd"},
			[]*Package{},
			"std,cmd",
		},
		{
			[]string{},
			[]*Package{&Package{ImportPath: "gopkg.in/somelib"}},
			"gopkg.in-somelib",
		},
		{
			[]string{"./..."},
			[]*Package{&Package{ImportPath: "somelib"}},
			"somelib",
		},
		{
			[]string{"../somelib", "../somelib"},
			[]*Package{&Package{ImportPath: "somelib"}},
			"somelib",
		},
		{
			[]string{"../lib1", "../lib2"},
			[]*Package{&Package{ImportPath: "gopkg.in/lib1"}, &Package{ImportPath: "gopkg.in/lib2"}},
			"gopkg.in-lib1,gopkg.in-lib2",
		},
		{
			[]string{"./..."},
			[]*Package{
				&Package{ImportPath: "gopkg.in/dir/lib1"},
				&Package{ImportPath: "gopkg.in/lib2"},
				&Package{ImportPath: "gopkg.in/lib3"},
			},
			"gopkg.in-dir-lib1,gopkg.in-lib2", // the 3rd component makes the name too long
		},
	}
	for _, data := range testData {
		expected := prefix + data.expected + suffix
		computed := libname(data.args, data.pkgs)
		if expected != computed {
			t.Errorf("shared library name must be '%s', not '%s'", expected, computed)
		}
	}
}
