// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minitypes

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const src = `
package p

type T struct {
	i int
}

type T2 T
`

func TestCheck(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	_, driver, err := CheckFiles([]*ast.File{file}, fset)
	if err != nil {
		t.Error(err)
	}
	for id, obj := range driver.defs {
		var typ, under Type
		if obj != nil {
			typ = obj.Type()
			if typ != nil {
				under = typ.Underlying()
			}
		}
		fmt.Printf("%s: %T (%T->%T)\n", id, obj, typ, under)
	}
}
