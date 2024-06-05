// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"fmt"
	"os"
	"strings"
)

/*
Rewrite a composite interface literal to a composite struct literal.

# Theory of Operation

The basic idea is to rewrite:

	type I interface {
		Do()
	}

	_ = I{
	Do: func() {},
	}

into:

	_ = _S{
		Do: func() {},
	}

	type _S struct {
		_do func()
	}

	func (s *_S) Do() { s._do() }

*/

func DEBUG() bool {
	return os.Getenv("DEBUG25860") == "true"
}

// nopos is the zero syntax.Pos.
var nopos syntax.Pos

func funcToPriKey(s string) string {
	return "_" + strings.ToLower(s)
}

func unique(s string) string {
	return s + "Unique"
}

type rewriter struct {
	pkg  *types2.Package
	info *types2.Info
	file *syntax.File
}

func (r *rewriter) inspect(n syntax.Node) bool {
	if n == nil {
		return false
	}
	switch n := n.(type) {
	case *syntax.CompositeLit:
		syntax.Fdump(os.Stderr, n)
		// ignore the case where this is an struct literal.
		if n.Type == nil {
			return false
		}
		if _, ok := n.Type.GetTypeInfo().Type.Underlying().(*types2.Interface); !ok {
			return false
		}
		r.rewriteCompIface(n)
		return false

	default:
		return true
	}
}

func rewrite(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	for _, file := range files {
		r := &rewriter{
			pkg:  pkg,
			info: info,
			file: file,
		}
		syntax.Inspect(file, r.inspect)
	}
}

// useObj returns syntax for a reference to decl, which should be its declaration.
func (r *rewriter) useObj(obj types2.Object) *syntax.Name {
	n := syntax.NewName(nopos, obj.Name())
	tv := syntax.TypeAndValue{Type: obj.Type()}
	tv.SetIsValue()
	n.SetTypeInfo(tv)
	r.info.Uses[n] = obj
	return n
}

func (r *rewriter) makeTypeDecl(structName string, fields []*syntax.Field) *syntax.TypeDecl {
	s, ts := r.makeStruct(fields, nil)
	td := &syntax.TypeDecl{
		Name: syntax.NewName(nopos, structName),
		Type: s,
	}
	obj := types2.NewTypeName(td.Name.Pos(), r.pkg, td.Name.Value, ts)
	r.info.Defs[td.Name] = obj
	r.useObj(obj)
	return td
}

func (r *rewriter) fieldToVar(fields ...*syntax.Field) []*types2.Var {
	var vars []*types2.Var
	for _, field := range fields {
		vars = append(vars, types2.NewField(nopos, r.pkg, field.Name.Value, field.Name.GetTypeInfo().Type, false))
	}
	return vars
}

func (r *rewriter) makeFuncDecl(funcName, structName string, params, results []*syntax.Field) *syntax.FuncDecl {
	recv := makeFieldGeneric("i", makeOperator(structName))
	fd := &syntax.FuncDecl{
		Pragma:     nil,
		Recv:       recv,
		Name:       syntax.NewName(nopos, funcName),
		TParamList: nil,
		Type:       makeFuncType(params, results),
		Body:       makeBlockStmt("i", funcToPriKey(funcName)),
	}

	obj := types2.NewFunc(fd.Name.Pos(), r.pkg, fd.Name.Value, nil)
	paramVars := r.fieldToVar(params...)
	resultVars := r.fieldToVar(results...)
	recvVars := r.fieldToVar(recv) // suboptimal
	tv := syntax.TypeAndValue{
		Type: types2.NewSignatureType(recvVars[0], nil, nil,
			types2.NewTuple(paramVars...),
			types2.NewTuple(resultVars...),
			false),
	}
	tv.SetIsValue()
	r.info.Defs[fd.Name] = obj
	r.useObj(obj)
	return fd
}

func makeOperator(structName string) *syntax.Operation {
	return &syntax.Operation{
		Op: syntax.Mul,
		X:  syntax.NewName(nopos, structName),
		Y:  nil,
	}
}

func makeBlockStmt(selectorName, varName string) *syntax.BlockStmt {
	return &syntax.BlockStmt{
		List: []syntax.Stmt{
			&syntax.ExprStmt{
				X: &syntax.CallExpr{
					Fun: &syntax.SelectorExpr{
						X:   syntax.NewName(nopos, selectorName),
						Sel: syntax.NewName(nopos, varName),
					},
					ArgList: nil,
					HasDots: false,
				},
			},
		},
		Rbrace: nopos,
	}
}

func makeFieldGeneric(name string, t syntax.Expr) *syntax.Field {
	return &syntax.Field{
		Name: syntax.NewName(nopos, name),
		Type: t,
	}
}

func (r *rewriter) makeField(name string, params, results []*syntax.Field) *syntax.Field {
	s := &syntax.Field{
		Name: syntax.NewName(nopos, name),
		Type: makeFuncType(params, results),
	}
	return s
}

func makeFuncType(params, results []*syntax.Field) *syntax.FuncType {
	return &syntax.FuncType{
		ParamList:  params,
		ResultList: results,
	}
}

func (r *rewriter) makeStruct(fields []*syntax.Field, tags []*syntax.BasicLit) (*syntax.StructType, *types2.Struct) {
	var fieldVars []*types2.Var
	for _, field := range fields {
		fieldVars = append(fieldVars, types2.NewField(nopos, r.pkg, field.Name.Value, field.Name.GetTypeInfo().Type, false))
	}
	ts := types2.NewStruct(fieldVars, nil)
	return &syntax.StructType{
		FieldList: fields,
		TagList:   tags,
	}, ts
}

func (r *rewriter) downCompLit(cl *syntax.CompositeLit, structName string) {
	cl.Type = syntax.NewName(nopos, structName)
}

func (r *rewriter) rewriteCompIface(lit *syntax.CompositeLit) {
	type method struct {
		params  []*syntax.Field
		results []*syntax.Field
	}
	methods := map[string]method{}
	sn, _ := lit.Type.(*syntax.Name)
	structName := unique(sn.Value)

	for _, e := range lit.ElemList {
		if kv, ok := e.(*syntax.KeyValueExpr); ok {
			n, _ := kv.Key.(*syntax.Name)
			fl, _ := kv.Value.(*syntax.FuncLit)

			methods[n.Value] = method{
				params:  fl.Type.ParamList,
				results: fl.Type.ResultList,
			}
		}
	}

	// typedecl
	var methodFields []*syntax.Field
	for methodName, values := range methods {
		fmt.Printf("rawMethods methodName=%s values=%+v\n", methodName, values)
		methodFields = append(methodFields, r.makeField(funcToPriKey(methodName), values.params, values.results))
	}
	td := r.makeTypeDecl(structName, methodFields)
	r.file.DeclList = append(r.file.DeclList, td)

	// funcdecl
	var funcdecls []*syntax.FuncDecl
	for k, v := range methods {
		funcdecls = append(funcdecls, r.makeFuncDecl(k, structName, v.params, v.results))
	}
	for _, fd := range funcdecls {
		r.file.DeclList = append(r.file.DeclList, fd)
	}

	// lower the iface literal
	// TODO underlying must be changed to struct
	// change name to struct name
	r.downCompLit(lit, structName)
}

func makeVar(pos syntax.Pos, pkg *types2.Package, name string, typ types2.Type, info *types2.Info) (*types2.Var, *syntax.Name) {
	obj := types2.NewVar(pos, pkg, name, typ)
	n := syntax.NewName(pos, name)
	tv := syntax.TypeAndValue{Type: typ}
	tv.SetIsValue()
	n.SetTypeInfo(tv)
	info.Defs[n] = obj
	return obj, n
}

// declSingleVar declares a variable with a given name, type, and initializer expression,
// and returns both the declaration and variable, so that the declaration can be placed
// in a specific scope.
func declSingleVar(name string, typ types2.Type, init syntax.Expr, pkg *types2.Package, info *types2.Info) (*syntax.DeclStmt, *types2.Var) {
	stmt := &syntax.DeclStmt{}
	obj, n := makeVar(stmt.Pos(), pkg, name, typ, info)
	stmt.DeclList = append(stmt.DeclList, &syntax.VarDecl{
		NameList: []*syntax.Name{n},
		// Note: Type is ignored
		Values: init,
	})
	return stmt, obj
}

// setValueType marks x as a value with type typ.
func setValueType(x syntax.Expr, typ syntax.Type) {
	tv := syntax.TypeAndValue{Type: typ}
	tv.SetIsValue()
	x.SetTypeInfo(tv)
}
