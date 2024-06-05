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

// nopos is the zero syntax.Pos.
var nopos syntax.Pos

func funcToPriKey(s string) string {
	return "_" + strings.ToLower(s)
}

func unique(s string) string {
	return s + "Unique"
}

func pline() {
	fmt.Println("========================================")
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
		r.rewriteCompIface(r.pkg, r.info, r.file, n)
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

func (r *rewriter) makeFuncDecl(funcName, structName string, params, results []*syntax.Field) *syntax.FuncDecl {
	fd := &syntax.FuncDecl{
		Pragma:     nil,
		Recv:       makeFieldGeneric("i", makeOperator(structName)),
		Name:       syntax.NewName(nopos, funcName),
		TParamList: nil,
		Type:       makeFuncType(params, results),
		Body:       makeBlockStmt("i", funcToPriKey(funcName)),
	}
	//recv := types2.NewField(nopos, r.pkg, "i", typ types2.Type, embedded bool)
	//_ = types2.NewSignatureType(recv *types2.Var, recvTypeParams []*types2.TypeParam, typeParams []*types2.TypeParam, params *types2.Tuple, results *types2.Tuple, variadic bool)
	//_ = types2.NewFunc(pos syntax.Pos, pkg *types2.Package, name string, sig *types2.Signature)

	obj := types2.NewFunc(fd.Name.Pos(), r.pkg, fd.Name.Value, nil)
	// the reciever will never be nil
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

// TODO Collapse new fields
func makeFieldGeneric(name string, t syntax.Expr) *syntax.Field {
	return &syntax.Field{
		Name: syntax.NewName(nopos, name),
		Type: t,
	}
}

func (r *rewriter) makeField(name string, params, results []*syntax.Field) *syntax.Field {
	//func (r *rewriter) makeField(name string, params, results []*syntax.Field) (*syntax.Field, *types2.Var) {
	s := &syntax.Field{
		Name: syntax.NewName(nopos, name),
		Type: makeFuncType(params, results),
	}
	//t := types2.NewField(nopos, r.pkg, name, s.Type(), false)
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

func (r *rewriter) rewriteCompIface(pkg *types2.Package, info *types2.Info, file *syntax.File, typ *syntax.CompositeLit) {
	// TODO move this to the rewiter
	type method struct {
		params  *types2.Tuple
		results *types2.Tuple
	}
	methods := map[string]method{}
	fldsVars := make([]*types2.Var, 0, len(typ.ElemList))

	type rawMethod struct {
		params  []*syntax.Field
		results []*syntax.Field
	}
	rawMethods := map[string]rawMethod{}

	// add struct type with private variables that are the appropriate function types
	sn, _ := typ.Type.(*syntax.Name)
	structName := unique(sn.Value)

	for _, e := range typ.ElemList {
		if kve, ok := e.(*syntax.KeyValueExpr); ok {
			n, _ := kve.Key.(*syntax.Name)
			fl, ok := kve.Value.(*syntax.FuncLit)
			if !ok {
				panic("not a FuncLit")
			}

			rawMethods[n.Value] = rawMethod{
				params:  fl.Type.ParamList,
				results: fl.Type.ResultList,
			}

			fldtyp := &syntax.FuncType{
				ParamList:  fl.Type.ParamList,
				ResultList: fl.Type.ResultList,
			}

			paramVars := make([]*types2.Var, 0, len(fl.Type.ParamList))
			for _, p := range fl.Type.ParamList {
				paramVars = append(paramVars, types2.NewParam(nopos, pkg, p.Name.Value, p.Type.GetTypeInfo().Type))
			}
			params := types2.NewTuple(paramVars...)

			resultVars := make([]*types2.Var, 0, len(fl.Type.ResultList))
			for _, r := range fl.Type.ParamList {
				resultVars = append(resultVars, types2.NewParam(nopos, pkg, r.Name.Value, r.Type.GetTypeInfo().Type))
			}
			results := types2.NewTuple(resultVars...)

			// TODO Handle variadic
			sig := types2.NewSignatureType(nil, nil, nil, params, results, false)
			_ = types2.NewFunc(nopos, pkg, funcToPriKey(n.Value), sig)

			fldsVars = append(fldsVars, types2.NewField(nopos, pkg, funcToPriKey(n.Value), fldtyp.GetTypeInfo().Type, false))
			// TODO should the fldtype be a funttype?

			//flds = append(flds, types2.NewField(nopos, pkg, funcToPriKey(n.Value), ft, false))
			methods[n.Value] = method{
				params:  params,
				results: results,
			}
		} else {
			panic("error with interface literal")
		}
	}

	// typedecl for struct  -------------------------------------------------------

	var methodFields []*syntax.Field
	for methodName, values := range rawMethods {
		methodFields = append(methodFields, r.makeField(funcToPriKey(methodName), values.params, values.results))
	}
	// bad wording maybe? Method names need to be converted to field names
	td := r.makeTypeDecl(structName, methodFields)
	file.DeclList = append(file.DeclList, td)

	// funcdecl for funcs  -------------------------------------------------------

	var funcdecls []*syntax.FuncDecl
	for k, v := range rawMethods {
		funcdecls = append(funcdecls, r.makeFuncDecl(k, structName, v.params, v.results))
	}
	pline()
	fmt.Println("--> dump funcdecls")
	for _, fd := range funcdecls {
		file.DeclList = append(file.DeclList, fd)
	}

	// down the iface literal  -------------------------------------------------------

	r.downCompLit(typ, structName)

	// lower interface literal to struct literal
	// TODO underlying must be changed to struct
	// change name to struct name
	name, _ := typ.Type.(*syntax.Name)
	name.Value = structName

	// change method assignment to variable assignment
	for _, e := range typ.ElemList {
		if kve, ok := e.(*syntax.KeyValueExpr); ok {
			name, ok := kve.Key.(*syntax.Name)
			if !ok {
				panic("not a name")
			}
			name.Value = funcToPriKey(name.Value)
		}
	}

}

func makeVarName(pos syntax.Pos, name string, typ types2.Type, pkg *types2.Package, info *types2.Info) (*types2.Var, *syntax.Name) {
	obj := types2.NewVar(pos, pkg, name, typ)
	n := syntax.NewName(pos, name)
	tv := syntax.TypeAndValue{Type: typ}
	tv.SetIsValue()
	n.SetTypeInfo(tv)
	info.Defs[n] = obj
	return obj, n
}

// declSingleVar declares a variable with a given name, type, and initializer value,
// and returns both the declaration and variable, so that the declaration can be placed
// in a specific scope.
func declSingleVar(name string, typ types2.Type, init syntax.Expr, pkg *types2.Package, info *types2.Info) (*syntax.DeclStmt, *types2.Var) {
	stmt := &syntax.DeclStmt{}
	obj, n := makeVarName(stmt.Pos(), name, typ, pkg, info)
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
