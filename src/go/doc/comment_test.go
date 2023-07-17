// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package doc

import (
	"bytes"
	"go/parser"
	"go/token"
	"internal/diff"
	"testing"
)

func TestComment(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata/pkgdoc", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if pkgs["pkgdoc"] == nil {
		t.Fatal("missing package pkgdoc")
	}
	pkg := New(pkgs["pkgdoc"], "testdata/pkgdoc", 0)

	var (
		input           = "[T] and [U] are types, [T.M] is a method, and [T.F] is a field, but [V] is a broken link. [rand.Int], [crand.Reader], and [go/doc.Package.Doc] are things. [G.M1] and [G.M2] are generic methods.\n"
		wantHTML        = `<p><a href="#T">T</a> and <a href="#U">U</a> are types, <a href="#T.M">T.M</a> is a method, and <a href="#T.F">T.F</a> is a field, but [V] is a broken link. <a href="/math/rand#Int">rand.Int</a>, <a href="/crypto/rand#Reader">crand.Reader</a>, and <a href="/go/doc#Package.Doc">go/doc.Package.Doc</a> are things. <a href="#G.M1">G.M1</a> and <a href="#G.M2">G.M2</a> are generic methods.` + "\n"
		wantOldHTML     = `<p>[T] and [U] are <i>types</i>, [T.M] is a method, and [T.F] is a field, but [V] is a broken link. [rand.Int], [crand.Reader], and <a href="/go/doc#Package.Doc">go/doc.Package.Doc</a> are things. [G.M1] and [G.M2] are generic methods.` + "\n"
		wantMarkdown    = "[T](#T) and [U](#U) are types, [T.M](#T.M) is a method, and [T.F](#T.F) is a field, but \\[V] is a broken link. [rand.Int](/math/rand#Int), [crand.Reader](/crypto/rand#Reader), and [go/doc.Package.Doc](/go/doc#Package.Doc) are things. [G.M1](#G.M1) and [G.M2](#G.M2) are generic methods.\n"
		wantText        = "T and U are types, T.M is a method, and T.F is a field, but [V] is a broken\nlink. rand.Int, crand.Reader, and go/doc.Package.Doc are things. G.M1 and G.M2\nare generic methods.\n"
		wantOldText     = "[T] and [U] are types, [T.M] is a method, and [T.F] is a field, but [V] is a\nbroken link. [rand.Int], [crand.Reader], and go/doc.Package.Doc are things.\n[G.M1] and [G.M2] are generic methods.\n"
		wantSynopsis    = "T and U are types, T.M is a method, and T.F is a field, but [V] is a broken link."
		wantOldSynopsis = "[T] and [U] are types, [T.M] is a method, and [T.F] is a field, but [V] is a broken link."
	)

	if b := pkg.HTML(input); string(b) != wantHTML {
		t.Errorf("%s", diff.Diff("pkg.HTML", b, "want", []byte(wantHTML)))
	}
	if b := pkg.Markdown(input); string(b) != wantMarkdown {
		t.Errorf("%s", diff.Diff("pkg.Markdown", b, "want", []byte(wantMarkdown)))
	}
	if b := pkg.Text(input); string(b) != wantText {
		t.Errorf("%s", diff.Diff("pkg.Text", b, "want", []byte(wantText)))
	}
	if b := pkg.Synopsis(input); b != wantSynopsis {
		t.Errorf("%s", diff.Diff("pkg.Synopsis", []byte(b), "want", []byte(wantText)))
	}

	var buf bytes.Buffer

	buf.Reset()
	ToHTML(&buf, input, map[string]string{"types": ""})
	if b := buf.Bytes(); string(b) != wantOldHTML {
		t.Errorf("%s", diff.Diff("ToHTML", b, "want", []byte(wantOldHTML)))
	}

	buf.Reset()
	ToText(&buf, input, "", "\t", 80)
	if b := buf.Bytes(); string(b) != wantOldText {
		t.Errorf("%s", diff.Diff("ToText", b, "want", []byte(wantOldText)))
	}

	if b := Synopsis(input); b != wantOldSynopsis {
		t.Errorf("%s", diff.Diff("Synopsis", []byte(b), "want", []byte(wantOldText)))
	}
}
