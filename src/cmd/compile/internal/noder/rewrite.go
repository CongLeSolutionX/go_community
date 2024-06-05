// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
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
	return s + "struct"
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
		if n.Type == nil {
			return false
		}
		typ, ok := n.Type.GetTypeInfo().Type.Underlying().(*types2.Interface)
		if !ok {
			return false
		}
		r.rewriteCompIface(n, typ)
		return false

	default:
		return true
	}
}

// rewrite changes...
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

func (r *rewriter) makeTypeDecl(structName string, fields []*syntax.Field, sigs []*types2.Signature) *syntax.TypeDecl {
	s, ts := r.makeStruct(fields, nil, sigs)
	td := &syntax.TypeDecl{
		Name: syntax.NewName(nopos, structName),
		Type: s,
	}
	obj := types2.NewTypeName(td.Name.Pos(), r.pkg, td.Name.Value, ts)
	r.pkg.Scope().Insert(obj)
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
	recv := makeFieldGeneric("_i", makeOperator(structName))
	type methodInfo struct {
		obj  *types2.Func // method
		ptr  bool         // true if pointer receiver
		recv *syntax.Name // receiver type name
	}
	fd := &syntax.FuncDecl{
		Pragma:     nil,
		Recv:       recv,
		Name:       syntax.NewName(nopos, funcName),
		TParamList: nil,
		Type:       makeFuncType(params, results),
		Body:       makeBlockStmt("_i", funcToPriKey(funcName)),
	}
	paramVars := r.fieldToVar(params...)
	resultVars := r.fieldToVar(results...)
	recvVars := r.fieldToVar(recv) // suboptimal
	sig := types2.NewSignatureType(recvVars[0], nil, nil,
		types2.NewTuple(paramVars...),
		types2.NewTuple(resultVars...),
		false)
	tv := syntax.TypeAndValue{Type: sig}
	obj := types2.NewFunc(fd.Name.Pos(), r.pkg, fd.Name.Value, sig)
	r.pkg.Scope().Insert(obj)
	tv.SetIsValue()
	r.info.Defs[fd.Name] = obj
	r.useObj(obj)
	return fd
}

func makeOperator(structName string) *syntax.Operation {
	op := &syntax.Operation{
		Op: syntax.Mul,
		X:  syntax.NewName(nopos, structName),
		Y:  nil,
	}
	tv := syntax.TypeAndValue{Type: op.GetTypeInfo().Type}
	tv.SetIsValue()
	op.SetTypeInfo(tv)
	return op
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

func (r *rewriter) makeStruct(fields []*syntax.Field, tags []*syntax.BasicLit, sigs []*types2.Signature) (*syntax.StructType, *types2.Struct) {
	var fieldVars []*types2.Var
	for idx, field := range fields {
		obj := types2.NewField(nopos, r.pkg, field.Name.Value, sigs[idx], false)
		n := syntax.NewName(nopos, field.Name.Value)
		tv := syntax.TypeAndValue{Type: field.Name.GetTypeInfo().Type}
		tv.SetIsValue()
		n.SetTypeInfo(tv)
		r.info.Defs[n] = obj
		fieldVars = append(fieldVars, obj)
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

// ========================================================================================
// ENTRY POINT
//
// ========================================================================================
// TODO(amedee) consider renaming this func
func (r *rewriter) rewriteCompIface(lit *syntax.CompositeLit, typ *types2.Interface) {
	methods := map[string]*syntax.FuncType{}

	sn, _ := lit.Type.(*syntax.Name)
	structName := unique(sn.Value)

	for _, e := range lit.ElemList {
		if kv, ok := e.(*syntax.KeyValueExpr); ok {
			n, _ := kv.Key.(*syntax.Name)
			if !ok {
				panic("expected *syntax.Name not found")
			}
			fl, ok := kv.Value.(*syntax.FuncLit)
			if !ok {
				panic("expected *syntax.FuncLit not found")
			}
			methods[n.Value] = fl.Type
		}
	}

	// TODO(amedee) Fix this document
	// TODO(amedee) Handle the case where this is the second interface literal for an interface

	// Declare struct type matching the interface literal.
	// Given
	//
	// T{name1: funclit1, name2: funclit2, ...}
	//
	// declare a struct of the form
	//
	// type Tstruct struct {
	// _name1: signature1
	// _name2: signature2
	// ...
	// }
	//
	// where each signature is derived from the corresponding function literal.

	// create a methodField for each method in the Interface Literal
	var methodFields []*syntax.Field
	for methodName, values := range methods {
		methodFields = append(methodFields, r.makeField(funcToPriKey(methodName), values.ParamList, values.ResultList))
	}

	var sigs []*types2.Signature
	for idx := 0; idx < typ.NumExplicitMethods(); idx += 1 {
		sigs = append(sigs, typ.Method(idx).Signature())
	}
	td := r.makeTypeDecl(structName, methodFields, sigs)
	r.file.DeclList = append(r.file.DeclList, td)
	// ========================================================================================

	// Declare func type method matching the interface literal.
	// Given
	//
	// T{Name1: funclit1, Name2: funclit2, ...}
	//
	// type Tstruct struct {
	// _name1: signature1
	// _name2: signature2
	// ...
	// }
	//
	// declare methods of the form
	//
	// func (_i *Tstruct) Name1() { return _i._name1() }
	// func (_i *Tstruct) Name2() { return _i._name2() }
	//
	// where each signature is derived from the corresponding function literal.
	var funcdecls []*syntax.FuncDecl
	for k, v := range methods {
		funcdecls = append(funcdecls, r.makeFuncDecl(k, structName, v.ParamList, v.ResultList))
	}
	for _, fd := range funcdecls {
		r.file.DeclList = append(r.file.DeclList, fd)
	}

	// ========================================================================================

	// Convert the existing interface literal into a struct literal.
	// Given
	//
	// T{Name1: funclit1, Name2: funclit2, ...}
	//
	// type Tstruct struct {
	// _name1: signature1
	// _name2: signature2
	// ...
	// }
	//
	// func (_i *Tstruct) Name1() { return _i._name1() }
	// func (_i *Tstruct) Name2() { return _i._name2() }
	//
	// Convert T into Tstruct
	//
	// &Tstruct{Name1: funclit1, Name2: funclit2, ...}
	//
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
