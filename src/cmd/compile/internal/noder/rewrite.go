// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"strings"
)

func funcToPriKey(s string) string {
	return "_" + strings.ToLower(s)
}

func unique(s string) string {
	return s + "_unique"
}

func rewriteCompIface(cl *syntax.CompositeLit) {
	if cl.Type == nil {
		println("    ** no type: nil means no literal type")
		return
	}
	// TODO(amedee) there must be a better way
	if _, ok := cl.Type.GetTypeInfo().Type.Underlying().(*types2.Interface); !ok {
		println("    ** not an interface literal")
		return
	}

	sn, _ := cl.Type.(*syntax.Name)
	structFieldList := make([]*syntax.Field, 0, len(cl.ElemList))

	for _, e := range cl.ElemList {
		if kve, ok := e.(*syntax.KeyValueExpr); ok {
			n, _ := kve.Key.(*syntax.Name)
			fl, _ := e.(*syntax.FuncLit)

			structFieldList = append(structFieldList, &syntax.Field{
				Name: &syntax.Name{
					Value: funcToPriKey(n.Value),
				},
				Type: &syntax.FuncType{
					ParamList:  fl.Type.ParamList,
					ResultList: fl.Type.ResultList,
				},
			})
		} else {
			panic("invalid interface literal")
		}
	}

	// add struct type with unexported fields that are named like the functions whey will serve.
	// TODO(amedee) handle pos
	s := &syntax.TypeDecl{
		Name: &syntax.Name{
			Value: unique(sn.Value),
		},
		TParamList: make([]*syntax.Field, 0),
		Alias:      false,
		Type: &syntax.StructType{
			FieldList: structFieldList,
			TagList:   []*syntax.BasicLit{}, // i >= len(TagList) || TagList[i] == nil means no tag for field i
		},
	}

	println(s) // only here while debugging

	// add methods to the struct type that implement the interface and call the corresponding field functions.

	funcLits := make([]*syntax.FuncLit, 0, len(structFieldList))

	for _, sfl := range structFieldList {
		ft, _ := sfl.Type.(*syntax.FuncType)

		funcLits = append(funcLits, &syntax.FuncLit{
			Type: &syntax.FuncType{
				ParamList:  ft.ParamList,
				ResultList: ft.ResultList,
			},
			Body: &syntax.BlockStmt{
				List:   []syntax.Stmt{},
				Rbrace: syntax.Pos{},
			},
		})
	}

	// lower interface literal to struct literal

}

func RewriteInterfaceLiteral(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	for _, file := range files {
		syntax.Inspect(file, func(n syntax.Node) bool {
			if n == nil {
				return false
			}
			switch n := n.(type) {
			case *syntax.CompositeLit:
				rewriteCompIface(n)
				return true
			default:
				return true
			}
			return true
		})
	}
}
