// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the check for http.Response values being used before
// checking for errors.

package main

import (
	"go/ast"
	"go/types"
)

var (
	httpResponseType types.Type
	httpClientType   types.Type
)

func init() {
	if typ := importType("net/http", "Response"); typ != nil {
		httpResponseType = typ.Underlying()
	}
	if typ := importType("net/http", "Client"); typ != nil {
		httpClientType = typ.Underlying()
	}
	// if http.Response is not defined don't register this check.
	if httpResponseType == nil || httpClientType == nil {
		return
	}

	register("httpresponse",
		"check errors are checked before using a http Response",
		checkHTTPResponse, callExpr)
}

func checkHTTPResponse(f *File, node ast.Node) {
	call := node.(*ast.CallExpr)
	if !isHTTPFuncOrMethodOnClient(f, call) {
		return
	}

	finder := &blockStmtFinder{node: call}
	ast.Walk(finder, f.file)
	stmts := finder.stmts()
	if len(stmts) < 2 {
		return
	}

	// the next statement is defer calling something on the http.Response.
	def, ok := stmts[1].(*ast.DeferStmt)
	if !ok {
		return
	}

	root := rootIdent(def.Call.Fun)
	resp := rootIdent(stmts[0].(*ast.AssignStmt).Lhs[0])
	if root == nil || resp == nil || root.Obj != resp.Obj {
		return
	}
	f.Badf(root.Pos(), "using %s before checking for errors", resp.Name)
}

// isHTTPFuncOrMethodOnClient checks whether the given call expression is on
// either a function of the net/http package or a method of http.Client that
// returns (*http.Response, error).
func isHTTPFuncOrMethodOnClient(f *File, expr *ast.CallExpr) bool {
	// The expression calls a function of the good type.
	fun, _ := expr.Fun.(*ast.SelectorExpr)
	sig, ok := f.pkg.types[fun].Type.(*types.Signature)
	if !ok || sig == nil {
		return false
	}
	res := sig.Results()
	if res.Len() != 2 || res.At(1).Type().Underlying() != errorType {
		return false
	}
	if ptr, ok := res.At(0).Type().(*types.Pointer); !ok || ptr.Elem().Underlying() != httpResponseType {
		return false
	}

	id, ok := fun.X.(*ast.Ident)
	if !ok {
		return false
	}

	typ := f.pkg.types[id]
	if typ.Type == nil {
		return id.Name == "http" // function in net/http package.
	}

	rcv := typ.Type.Underlying()
	if rcv == httpClientType {
		return true //  method on http.Client.
	}
	ptr, ok := rcv.(*types.Pointer)
	return ok && ptr.Elem().Underlying() == httpClientType // it is a method on *http.Client.
}

// blockStmtFinder is an ast.Visitor that given any ast node can find the
// statement containing it and its succeeding statements in the same block.
type blockStmtFinder struct {
	node  ast.Node
	stmt  ast.Stmt
	block *ast.BlockStmt
}

// Visit finds f.node performing a search down a the ast tree.
// It keeps the last block statement and statement seen for later use.
func (f *blockStmtFinder) Visit(node ast.Node) ast.Visitor {
	if node == nil || f.node.Pos() < node.Pos() || f.node.End() > node.End() {
		return nil // not here
	}
	switch n := node.(type) {
	case *ast.BlockStmt:
		f.block = n
	case ast.Stmt:
		f.stmt = n
	}
	if f.node.Pos() == node.Pos() && f.node.End() == node.End() {
		return nil // found
	}
	return f // keep looking
}

// stmts returns a list of statements starting from the one including f.node.
func (f *blockStmtFinder) stmts() []ast.Stmt {
	for i, v := range f.block.List {
		if f.stmt == v {
			return f.block.List[i:]
		}
	}
	return nil
}

// rootIdent finds the root identifier in a chain of selector expressions.
func rootIdent(n ast.Node) *ast.Ident {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return rootIdent(n.X)
	case *ast.Ident:
		return n
	default:
		return nil
	}
}
