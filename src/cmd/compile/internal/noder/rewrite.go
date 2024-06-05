// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"fmt"
	"os"
	"slices"
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

// nopos is the zero syntax.Pos.
var nopos syntax.Pos

// funcToPriField converts a function name to a field name.
func funcToPriField(s string) string {
	return "_" + strings.ToLower(s)
}

// ifaceToStructName converts an interface name to a struct name.
func ifaceToStructName(s string) string {
	return s + "struct"
}

type rewriter struct {
	pkg  *types2.Package
	info *types2.Info
	file *syntax.File

	recv        types2.Type
	structField map[string]types2.Object
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
		r.convertIfaceLit(n, typ)
		return false

	default:
		return true
	}
}

func rewrite(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	// printInfo(info)
	for _, file := range files {
		r := &rewriter{
			pkg:         pkg,
			info:        info,
			file:        file,
			structField: make(map[string]types2.Object),
		}
		syntax.Inspect(file, r.inspect)
	}
	// printInfo(info)
	os.Exit(2)
}

func (r *rewriter) convertIfaceLit(lit *syntax.CompositeLit, typ *types2.Interface) {
	methods := map[string]*syntax.FuncType{}

	sn, _ := lit.Type.(*syntax.Name)
	structName := ifaceToStructName(sn.Value)

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

	// for the prototype it's ok to just number the structs.

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
		methodFields = append(methodFields, r.makeField(funcToPriField(methodName), values.ParamList, values.ResultList))
	}

	var sigs []*types2.Signature
	for idx := 0; idx < typ.NumExplicitMethods(); idx += 1 {
		sigs = append(sigs, typ.Method(idx).Signature())
	}
	td, ts, tn := r.makeTypeDecl(structName, methodFields, sigs)
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
		funcdecls = append(funcdecls, r.makeFuncDecl(k, structName, v.ParamList, v.ResultList, ts, tn))
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
	r.convertInterfaceLitToStructLit(lit, structName)
}

func (r *rewriter) makeTypeDecl(structName string, fields []*syntax.Field, sigs []*types2.Signature) (*syntax.TypeDecl, *types2.Named, *types2.TypeName) {
	s, ts := r.makeStruct(fields, nil, sigs)
	td := &syntax.TypeDecl{
		Name: syntax.NewName(nopos, structName),
		Type: s,
	}
	obj := types2.NewTypeName(td.Name.Pos(), r.pkg, td.Name.Value, nil)
	nt := types2.NewNamed(obj, ts, nil)
	// new named will set the object type to itself
	// todo go back and add the methods

	r.pkg.Scope().Insert(obj)
	r.info.Defs[td.Name] = obj
	r.useObj(obj)
	return td, nt, obj
}

func (r *rewriter) fieldToVar(fields ...*syntax.Field) []*types2.Var {
	var vars []*types2.Var
	for _, field := range fields {
		vars = append(vars, types2.NewField(nopos, r.pkg, field.Name.Value, field.Name.GetTypeInfo().Type, false))
	}
	return vars
}

func (r *rewriter) makeFuncDecl(funcName, structName string, params, results []*syntax.Field, baseType *types2.Named, baseTypeName *types2.TypeName) *syntax.FuncDecl {
	// make the operator
	op := &syntax.Operation{
		Op: syntax.Mul,
		X:  syntax.NewName(nopos, structName),
		Y:  nil,
	}
	setValueType(op, op.GetTypeInfo().Type)

	// make the reciever
	recv := &syntax.Field{
		Name: syntax.NewName(nopos, "_i"),
		Type: op,
	}

	ptr := types2.NewPointer(baseType)
	recv.Name.SetTypeInfo(syntax.TypeAndValue{Type: ptr})

	// needed for selections and composite literal type
	r.recv = ptr

	fd := &syntax.FuncDecl{
		Pragma:     nil,
		Recv:       recv,
		Name:       syntax.NewName(nopos, funcName),
		TParamList: nil,
		Type:       makeFuncType(params, results),
		Body:       r.makeBlockStmt("_i", funcToPriField(funcName)),
	}
	paramVars := r.fieldToVar(params...)
	resultVars := r.fieldToVar(results...)

	// TODO should this be a field and not a var?
	recvVar2, recvName := r.makeVarName(nopos, recv.Name.Value, recv.Name.GetTypeInfo().Type)
	fd.Recv.Name = recvName

	sig := types2.NewSignatureType(recvVar2, nil, nil,
		types2.NewTuple(paramVars...),
		types2.NewTuple(resultVars...),
		false)

	tv := syntax.TypeAndValue{Type: sig}
	obj := types2.NewFunc(fd.Name.Pos(), r.pkg, fd.Name.Value, sig)
	r.pkg.Scope().Insert(obj)
	tv.SetIsValue()
	r.info.Defs[fd.Name] = obj
	setTypeInfo(fd.Body, &tv)
	r.useObj(recvVar2)
	return fd
}

// TODO(amedee) convert this to work on multiple sets
func setTypeInfo(bs *syntax.BlockStmt, tv *syntax.TypeAndValue) {
	se := bs.List[0].(*syntax.ExprStmt)
	ce := se.X.(*syntax.CallExpr)
	// TODO(amedee) double check this
	ce.Fun.SetTypeInfo(*tv)
}

func (r *rewriter) makeBlockStmt(selectorName, varName string) *syntax.BlockStmt {
	fun := &syntax.SelectorExpr{
		X:   syntax.NewName(nopos, selectorName),
		Sel: syntax.NewName(nopos, varName),
	}

	obj, ok := r.structField[varName]
	if !ok {
		panic("obj missing: " + selectorName)
	}
	types2.RecordSelection(r.info, fun, types2.FieldVal, r.recv, obj, []int{0}, true)
	return &syntax.BlockStmt{
		List: []syntax.Stmt{
			&syntax.ExprStmt{
				X: &syntax.CallExpr{
					Fun:     fun,
					ArgList: nil,
					HasDots: false,
				},
			},
		},
		Rbrace: nopos,
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

		// TODO needed for selections
		r.structField[field.Name.Value] = obj

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

// convertInterfaceLitToStructLit converts the iface literal to a struct literal.
func (r *rewriter) convertInterfaceLitToStructLit(cl *syntax.CompositeLit, structName string) {
	// change the type name from Interface to &structName
	clt, ok := cl.Type.(*syntax.Name)
	if !ok {
		panic("cl.Type not name")
	}
	clt = syntax.NewName(clt.Pos(), "&"+structName)

	// convert function names to field names:
	// Do() -> _do
	for idx, _ := range cl.ElemList {
		kve, ok := cl.ElemList[idx].(*syntax.KeyValueExpr)
		if !ok {
			panic("kve not key value expression")
		}
		key, ok := kve.Key.(*syntax.Name)
		if !ok {
			panic("key not name")
		}
		kve.Key = syntax.NewName(key.Pos(), funcToPriField(key.Value))
	}

	// change composite lit value to *struct
	tv := syntax.TypeAndValue{Type: r.recv}
	cl.Type = syntax.NewName(clt.Pos(), structName)
	cl.Type.SetTypeInfo(tv)
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

// setValueType marks x as a value with type typ.
func setValueType(x syntax.Expr, typ syntax.Type) {
	tv := syntax.TypeAndValue{Type: typ}
	tv.SetIsValue()
	x.SetTypeInfo(tv)
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

func (r *rewriter) makeVarName(pos syntax.Pos, name string, typ types2.Type) (*types2.Var, *syntax.Name) {
	obj := types2.NewVar(pos, r.pkg, name, typ)
	n := syntax.NewName(pos, name)
	tv := syntax.TypeAndValue{Type: typ}
	tv.SetIsValue()
	n.SetTypeInfo(tv)
	r.info.Defs[n] = obj
	return obj, n
}

// -------------------------
// Helpers
// -------------------------
func DEBUG() bool {
	return os.Getenv("DEBUG25860") == "true"
}

func PrintDebug() bool {
	return os.Getenv("DEBUGOUTPUT") == "true"
}

func del() {
	println(">> ----------------------------------")
}

func header(s string) {
	del()
	println(">> ", s)
	del()
}

func printInfo(info *types2.Info) {
	printDefs(info)
	printImplicits(info)
	printUses(info)
	printTypes(info)
	printScopes(info)
	printSelections(info)
}

func printDefs(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Defs")
	type pair struct {
		s string
		o types2.Object
		n *syntax.Name
	}
	var pairs []pair
	for k, v := range info.Defs {
		pairs = append(pairs, pair{k.Value, v, k})
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		return strings.Compare(a.s, b.s)
	})
	for _, p := range pairs {
		fmt.Printf("%-20s%+v\n", p.s, p.o)
	}
	del()
}

func printImplicits(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Implicits")
	type pair struct {
		n syntax.Node
		o types2.Object
	}
	var pairs []pair
	for k, v := range info.Implicits {
		pairs = append(pairs, pair{k, v})
	}
	for _, p := range pairs {
		fmt.Printf("%-20s%+v\n", p.n, p.o)
	}
	del()
}

func printSelections(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Selections")
	type pair struct {
		s *syntax.SelectorExpr
		o *types2.Selection
	}
	var pairs []pair
	for k, v := range info.Selections {
		pairs = append(pairs, pair{k, v})
	}
	for _, p := range pairs {
		fmt.Printf("selExpr=%-20v selection=%+v\n", p.s, p.o)
	}
	del()
}

func printScopes(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Scopes")
	type pair struct {
		s syntax.Node
		o *types2.Scope
	}
	var pairs []pair
	for k, v := range info.Scopes {
		pairs = append(pairs, pair{k, v})
	}
	for _, p := range pairs {
		del()
		fmt.Printf("%-60s\t%+v\n", p.s.Pos(), p.o.Names())
		switch t := p.s.(type) {
		case *syntax.File:
			fmt.Printf("file=%+v", t.DeclList)
			for idx, d := range t.DeclList {
				fmt.Printf("idx=%d decl=%+v\n", idx, d)
			}
		case *syntax.FuncType:
		default:
			println("missed a case")
			fmt.Printf("t=%+v\n", t)
		}
	}
	del()
}

func printUses(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Uses")
	type pair struct {
		s string
		o types2.Object
	}
	var pairs []pair
	for k, v := range info.Uses {
		pairs = append(pairs, pair{k.Value, v})
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		return strings.Compare(a.s, b.s)
	})
	for _, p := range pairs {
		fmt.Printf("%-20s%+v\t%+v\n", p.s, p.o, p.o.Pos())
	}
	del()
}

func printTypes(info *types2.Info) {
	if !PrintDebug() {
		return
	}
	header("Types")
	for k, v := range info.Types {
		fmt.Printf("%-20s%+v\n", k, v)
	}
	del()
}
