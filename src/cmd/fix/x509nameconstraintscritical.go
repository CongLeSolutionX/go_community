// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "go/ast"

func init() {
	register(x509nameconstraintscriticalFix)
}

var x509nameconstraintscriticalFix = fix{
	"x509nameconstraintscritical",
	"2015-03-11",
	x509nameconstraintscritical,
	`Adapt element PermittedDNSDomainsCritical to more generic NameConstraintsCritical.

https://go-review.googlesource.com/#/c/3230/
`,
}

func x509nameconstraintscritical(f *ast.File) bool {
	if !imports(f, "crypto/x509") {
		return false
	}

	fixed := false
	walk(f, func(n interface{}) {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return
		}
		se, ok := cl.Type.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if !isTopName(se.X, "x509") || se.Sel == nil {
			return
		}
		switch ss := se.Sel.String(); ss {
		case "Certificate":
			for _, e := range cl.Elts {
				if _, ok := e.(*ast.KeyValueExpr); !ok {
					break
				}

				if e.(*ast.KeyValueExpr).Key.(*ast.Ident).String() == "PermittedDNSDomainsCritical" {
					e.(*ast.KeyValueExpr).Key = ast.NewIdent("NameConstraintsCritical")
					fixed = true
				}
			}
		}
	})
	return fixed
}
