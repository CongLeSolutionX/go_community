// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program is run via "go generate" (via a directive in sort.go)
// to generate zfuncversion.go.
//
// It copies sort.go to zfuncversion.go, only retaining funcs which
// take a "data Interface" parameter, and renaming each to have a
// "_func" suffix and taking a "data lessSwap" instead. It then rewrites
// each internal function call to the appropriate _func variants.

package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"regexp"
)

var fset = token.NewFileSet()

func main() {
	af, err := parser.ParseFile(fset, "sort.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	af.Doc = nil
	af.Imports = nil
	af.Comments = nil

	// Functions whose declarations we rewrite.
	functions := make(map[string]bool, 50)

	var newDecl []ast.Decl
	for _, d := range af.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Recv != nil || fd.Name.IsExported() {
			continue
		}
		typ := fd.Type
		if len(typ.Params.List) < 1 {
			continue
		}
		arg0 := typ.Params.List[0]
		arg0Name := arg0.Names[0].Name
		arg0Type := arg0.Type.(*ast.Ident)
		if arg0Name != "data" || arg0Type.Name != "Interface" {
			continue
		}
		arg0Type.Name = "lessSwap"

		functions[fd.Name.Name] = true

		newDecl = append(newDecl, fd)
	}
	af.Decls = newDecl
	ast.Walk(visitMap(functions), af)

	var out bytes.Buffer
	if err := format.Node(&out, fset, af); err != nil {
		log.Fatalf("format.Node: %v", err)
	}

	// Get rid of blank lines after removal of comments.
	src := regexp.MustCompile(`\n{2,}`).ReplaceAll(out.Bytes(), []byte("\n"))

	// Add comments to each func, for the lost reader.
	// This is so much easier than adding comments via the AST
	// and trying to get position info correct.
	src = regexp.MustCompile(`(?m)^func (\w+)`).ReplaceAll(src, []byte("\n// Auto-generated variant of sort.go:$1\nfunc ${1}_func"))

	// Final gofmt.
	src, err = format.Source(src)
	if err != nil {
		log.Fatalf("format.Source: %v on\n%s", err, src)
	}

	out.Reset()
	out.WriteString(`// Code generated from sort.go using genzfunc.go; DO NOT EDIT.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

`)
	out.Write(src)

	const target = "zfuncversion.go"
	if err := ioutil.WriteFile(target, out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

type visitMap map[string]bool

func (m visitMap) Visit(n ast.Node) ast.Visitor {
	ce, ok := n.(*ast.CallExpr)
	if ok {
		rewriteCall(ce, map[string]bool(m))
	}
	return m
}

func rewriteCall(ce *ast.CallExpr, m map[string]bool) {
	ident, ok := ce.Fun.(*ast.Ident)
	if !ok {
		// e.g. skip SelectorExpr (data.Less(..) calls)
		return
	}
	// Rewrite call if we rewrote callee's declaration.
	if _, in := m[ident.Name]; in {
		ident.Name += "_func"
	}
}
