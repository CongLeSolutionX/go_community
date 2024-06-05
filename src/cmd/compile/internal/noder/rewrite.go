// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"fmt"
)

// perhaps declare the stuct and then refrence it in th composite literal
// TODO: I think what I need to dump is the compositeliteral

func dumpStruct(s *syntax.StructType) {
	fmt.Println("--> dumpStruct")
	for i, f := range s.FieldList {
		fmt.Printf("\t%d field=%v\n", i, f)
	}
	for i, t := range s.TagList {
		fmt.Printf("\t%d tag=%v\n", i, t)
	}
	//fmt.Printf("\texpr=%v\n", s.expr)
	// for s.FieldList = append(s.FieldList, )

	// StructType struct {
	// 	FieldList []*Field
	// 	TagList   []*BasicLit // i >= len(TagList) || TagList[i] == nil means no tag for field i
	// 	expr
	// }
}

// func RewriteInterfaceLiteral(pkg *types2.Package, info *types2.Info, files []*syntax.File) map[*syntax.FuncLit]bool {
func RewriteInterfaceLiteral(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	fmt.Println("<-- RewriteInterfaceLiteral -->")
	defer fmt.Println("</- RewriteInterfaceLiteral -->")
	fmt.Printf("files=%+v\n\n", files)

	for _, file := range files {
		syntax.Inspect(file, func(n syntax.Node) bool {
			if n == nil {
				return false
			}
			fmt.Printf("--\nNodeType=%T\nnode=%+v\n", n, n)
			switch n := n.(type) {
			case *syntax.FuncDecl:
				sig, _ := info.Defs[n.Name].Type().(*types2.Signature)
				//rewriteFunc(pkg, info, n.Type, n.Body, sig, ri)
				println(sig)
				return true
			case *syntax.StructType:
				fmt.Printf("struct=%+v\n", n)
				//sig, _ := info.Defs[n.Name].Type().(*types2.Signature)
				//rewriteFunc(pkg, info, n.Type, n.Body, sig, ri)
				//println(sig)
				dumpStruct(n)
				return false
			case *syntax.FuncLit:
				sig, _ := info.Types[n].Type.(*types2.Signature)
				// if sig == nil {
				// 	tv := n.GetTypeInfo()
				// 	sig = tv.Type.(*types2.Signature)
				// }
				//rewriteFunc(pkg, info, n.Type, n.Body, sig, ri)
				println(sig)
				return true
			default:
				fmt.Printf("n=%T n=%v\n", n, n)
				//println(n.Name)
				return true
			}
			return true
		})
	}
}
