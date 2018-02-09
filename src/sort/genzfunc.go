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
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"regexp"
)

const (
	sourceFile    = "sort.go"
	sourcePackage = "sort"

	targetFile = "zfuncversion.go"

	method1Name       = "Less"
	method2Name       = "Swap"
	methodParamType   = "int"
	method1ResultType = "bool"

	paramIdent = "data"
	paramType  = "Interface" // an interface type
	newType    = "lessSwap"  // a struct

	panicMsg = "Sanity check failed, " + sourceFile + " or this program need modification."
)

// Copies fields without position information and without comments.
func fieldListCopy(fields []*ast.Field) []*ast.Field {
	retFields := make([]*ast.Field, len(fields))
	for i := range fields {
		retFields[i] = &ast.Field{Names: make([]*ast.Ident, len(fields[i].Names))}
		for j := range fields[i].Names {
			retFields[i].Names[j] = &ast.Ident{Name: fields[i].Names[j].Name}
		}
		retFields[i].Type = &ast.Ident{Name: fields[i].Type.(*ast.Ident).Name}
	}
	return retFields
}

func exprListCopy(exprList []ast.Expr, ident, replacementIdent string) []ast.Expr {
	ret := make([]ast.Expr, len(exprList))
	for i := range exprList {
		ret[i] = copyNode(exprList[i], ident, replacementIdent).(ast.Expr)
	}
	return ret
}

// Copies node without position information and without comments while renaming
// identifers named ident to replacementIdent.
func copyNode(node ast.Node, ident, replacementIdent string) ast.Node {
	// TODO: for more generality/correctness, handle all possible node types
	// like in src/go/ast/walk.go

	switch n := node.(type) {
	case *ast.Ident:
		retIdent := &ast.Ident{Name: n.Name}
		if n.Name == ident {
			retIdent.Name = replacementIdent
		}
		return retIdent
	case *ast.IndexExpr:
		return &ast.IndexExpr{X: copyNode(n.X, ident, replacementIdent).(ast.Expr), Index: copyNode(n.Index, ident, replacementIdent).(ast.Expr)}
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{X: copyNode(n.X, ident, replacementIdent).(ast.Expr), Op: n.Op, Y: copyNode(n.Y, ident, replacementIdent).(ast.Expr)}
	case *ast.AssignStmt:
		return &ast.AssignStmt{Lhs: exprListCopy(n.Lhs, ident, replacementIdent), Tok: n.Tok, Rhs: exprListCopy(n.Rhs, ident, replacementIdent)}
	case *ast.ReturnStmt:
		return &ast.ReturnStmt{Results: exprListCopy(n.Results, ident, replacementIdent)}
	default:
		panic(panicMsg)
	}
}

// Copies a stmts without position information and without comments while
// renaming identifers named ident to replacementIdent.
func copyStatements(stmts []ast.Stmt, ident, replacementIdent string) []ast.Stmt {
	retStmts := make([]ast.Stmt, len(stmts))
	for i := range stmts {
		retStmts[i] = copyNode(stmts[i], ident, replacementIdent).(ast.Stmt)
	}
	return retStmts
}

func funcsFromMethods(aFDecls []ast.Decl, t *types.Named, newName string) *ast.CompositeLit {
	retStruct := &ast.CompositeLit{Type: &ast.Ident{Name: newType}, Elts: []ast.Expr{
		&ast.FuncLit{Body: &ast.BlockStmt{}, Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		}},
		&ast.FuncLit{Body: &ast.BlockStmt{}, Type: &ast.FuncType{
			Params: &ast.FieldList{},
		}},
	}}

	var methodPos [2]token.Pos
	n := t.NumMethods()
	for i := 0; i < n; i++ {
		f := t.Method(i)
		switch f.Name() {
		case method1Name:
			methodPos[0] = f.Pos()
		case method2Name:
			methodPos[1] = f.Pos()
		}
	}
	if methodPos[0] == token.NoPos || methodPos[1] == token.NoPos {
		panic(panicMsg)
	}

	for i := range methodPos {
		declI := 0
		for j, decl := range aFDecls {
			if methodPos[i] < decl.Pos() {
				break
			}
			declI = j
		}
		f := aFDecls[declI].(*ast.FuncDecl)
		retStruct.Elts[i].(*ast.FuncLit).Type.Params.List = fieldListCopy(f.Type.Params.List)
		if i == 0 {
			retStruct.Elts[i].(*ast.FuncLit).Type.Results.List = fieldListCopy(f.Type.Results.List)
		}
		retStruct.Elts[i].(*ast.FuncLit).Body.List = copyStatements(f.Body.List, f.Recv.List[0].Names[0].Name, newName)
	}

	return retStruct
}

type visitStruct struct {
	// For finding the method declarations.
	aFDecls []ast.Decl

	// paramType
	i *types.Interface

	// For finding types that implement method1Name and method2Name, a
	// subset of the paramType method set.
	checkInfoUses map[*ast.Ident]types.Object
}

func (vS *visitStruct) Visit(n ast.Node) ast.Visitor {
	// TODO: for more generality we could handle call expressions (when
	// func takes vS.i and argument type is not vS.i) and token.DEFINEs
	// (when some variables in the LHS tuple are already declared and of
	// type vS.i), too.

	aS, ok := n.(*ast.AssignStmt)
	if ok && aS.Tok == token.ASSIGN {
		for t := range aS.Lhs {
			rHS, ok := aS.Rhs[t].(*ast.UnaryExpr)
			lHS, ok1 := aS.Lhs[t].(*ast.Ident)
			if !ok || !ok1 || rHS.Op != token.AND {
				continue
			}
			lHSType := vS.checkInfoUses[lHS].Type()
			if !types.IdenticalIgnoreTags(vS.i, lHSType.Underlying()) {
				continue
			}
			ident := rHS.X.(*ast.Ident)
			aS.Rhs[t] = funcsFromMethods(vS.aFDecls, vS.checkInfoUses[ident].Type().(*types.Named), ident.Name)
		}
	}
	return vS
}

func main() {
	fset := token.NewFileSet()
	var err error
	aFiles := make([]*ast.File, 2)
	for i, fName := range []string{sourceFile, "search.go"} {
		aFiles[i], err = parser.ParseFile(fset, fName, nil, 0)
		if err != nil {
			log.Fatalf("parser: %v", err)
		}
	}
	for i := range aFiles {
		aFiles[i].Doc = nil
		aFiles[i].Imports = nil
		aFiles[i].Comments = nil
	}

	typesConfInfo := &types.Info{
		// For getting paramType type.
		Defs: make(map[*ast.Ident]types.Object),

		// For comparing types of assignment LHS identifiers with paramType.
		Uses: make(map[*ast.Ident]types.Object),
	}
	_, _ = (&types.Config{Importer: importer.Default()}).Check(sourcePackage, fset, aFiles, typesConfInfo)

	//mapPrint(fset, typesConfInfo.Uses)

	// Get paramType type.
	var paramTypeType *types.Interface
	for id := range typesConfInfo.Defs {
		if id.Name == paramType {
			paramTypeType = typesConfInfo.Defs[id].Type().Underlying().(*types.Interface)

			// Sanity check the interface.
			n := paramTypeType.NumMethods()
			methodFound := false
			for i := 0; i < n; i++ {
				fun := paramTypeType.Method(i)
				if fun.Name() == method1Name {
					s := fun.Type().(*types.Signature)
					if t := s.Params(); t.Len() != 2 || t.At(0).Type().String() != "int" || t.At(1).Type().String() != "int" {
						panic(panicMsg)
					}
					if t := s.Results(); t.Len() != 1 || t.At(0).Type().String() != "bool" {
						panic(panicMsg)
					}
					if s.Variadic() != false {
						panic(panicMsg)
					}

					methodFound = true
					break
				}
			}
			if !methodFound {
				panic(panicMsg)
			}
			methodFound = false
			for i := 0; i < n; i++ {
				fun := paramTypeType.Method(i)
				if fun.Name() == method2Name {
					s := fun.Type().(*types.Signature)
					if t := s.Params(); t.Len() != 2 || t.At(0).Type().String() != "int" || t.At(1).Type().String() != "int" {
						panic(panicMsg)
					}
					if t := s.Results(); t.Len() != 0 {
						panic(panicMsg)
					}
					if s.Variadic() != false {
						panic(panicMsg)
					}

					methodFound = true
					break
				}
			}
			if !methodFound {
				panic(panicMsg)
			}

			break
		}
	}
	if paramTypeType == nil {
		panic(panicMsg)
	}

	// Rewrite assignments.
	ast.Walk(&visitStruct{aFiles[0].Decls, paramTypeType, typesConfInfo.Uses}, aFiles[0])

	funcsToRewrite := make(map[string]bool)
	var newDecl []ast.Decl
	for _, d := range aFiles[0].Decls {
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
		if arg0Name != paramIdent || arg0Type.Name != paramType {
			continue
		}
		arg0Type.Name = newType

		funcsToRewrite[fd.Name.Name] = true

		newDecl = append(newDecl, fd)
	}
	aFiles[0].Decls = newDecl
	ast.Walk(visitMap(funcsToRewrite), aFiles[0])

	var out bytes.Buffer
	if err = format.Node(&out, fset, aFiles[0]); err != nil {
		log.Fatalf("format.Node: %v", err)
	}

	// Get rid of blank lines after removal of comments.
	src := regexp.MustCompile(`\n{2,}`).ReplaceAll(out.Bytes(), []byte("\n"))

	// Add comments to each func, for the lost reader.
	// This is so much easier than adding comments via the AST
	// and trying to get position info correct.
	src = regexp.MustCompile(`(?m)^func (\w+)`).ReplaceAll(src, []byte("\n// Auto-generated variant of "+sourceFile+":$1\nfunc ${1}_func"))

	// Final gofmt.
	src, err = format.Source(src)
	if err != nil {
		log.Fatalf("format.Source: %v on\n%s", err, src)
	}

	out.Reset()
	out.WriteString(`// Code generated from ` + sourceFile + ` using genzfunc.go; DO NOT EDIT.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

`)
	out.Write(src)

	if err := ioutil.WriteFile(targetFile, out.Bytes(), 0644); err != nil {
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
