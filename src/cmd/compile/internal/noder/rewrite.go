// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"fmt"
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

func rewrite(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	for _, file := range files {
		syntax.Inspect(file, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.CompositeLit:
				// ignore the case where this is an struct literal.
				if n.Type == nil {
					return false
				}
				if _, ok := n.Type.GetTypeInfo().Type.Underlying().(*types2.Interface); !ok {
					return false
				}
				rewriteCompIface(pkg, info, file, n)
				return false
			default:
				return true
			}
		})
	}
}

func rewriteCompositeLit(pkg *types2.Package, info *types2.Info, file *syntax.File, typ *syntax.CompositeLit) {
	r := &rewriter{
		pkg:  pkg,
		info: info,
		file: file,
		cl:   typ,
	}
	syntax.Inspect(typ, r.inspect)
}

// OLD ONE BEING REFACTORED
func rewriteCompIface(pkg *types2.Package, info *types2.Info, file *syntax.File, typ *syntax.CompositeLit) {
	type method struct {
		params  *types2.Tuple
		results *types2.Tuple
	}
	methods := map[string]method{}
	flds := make([]*types2.Var, 0, len(typ.ElemList))

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

			flds = append(flds, types2.NewField(nopos, pkg, funcToPriKey(n.Value), fldtyp.GetTypeInfo().Type, false))
			//flds = append(flds, types2.NewField(nopos, pkg, funcToPriKey(n.Value), ft, false))
			methods[n.Value] = method{
				params:  params,
				results: results,
			}
		} else {
			panic("error with interface literal")
		}
	}

	sobj := types2.NewStruct(flds, nil)
	//pkg.Scope().Insert(sobj)

	tnobj := types2.NewTypeName(nopos, pkg, structName, sobj)
	//pkg.Scope().Insert(tnobj)

	nobj := types2.NewNamed(tnobj, sobj, nil)
	// TODO set tv and add to info.defs
	//nobj := types2.NewNamed(tnobj, sobj, nil)
	// pkg.Scope().Insert(nobj)
	// info.Defs[structName] = nobj

	for name, m := range methods {
		// TODO this doesn't seem right. Should this be a named sobj?
		obj := types2.NewFunc(nopos, pkg, name, types2.NewSignatureType(tnobj, nil, nil, m.params, m.results, false))
		pkg.Scope().Insert(obj)
	}

	// funcDecls := make([]*syntax.FuncDecl, 0, len(methods))
	// for _, ifo := range methods {
	// 	funcDecls = append(funcDecls, &syntax.FuncDecl{
	// 		Pragma: nil,
	// 		Recv: &syntax.Field{
	// 			Name: &syntax.Name{
	// 				Value: "i",
	// 			},
	// 			Type: &syntax.Operation{
	// 				Op: syntax.Mul,
	// 				X: &syntax.Name{
	// 					Value: structName, // has to match struct name
	// 				},
	// 				Y: nil,
	// 			},
	// 		},
	// 		Name: &syntax.Name{
	// 			Value: ifo.funcName, // set to function name
	// 		},
	// 		TParamList: nil,
	// 		Type:       &syntax.FuncType{},
	// 		Body: &syntax.BlockStmt{ // TODO(amedee) what if there are return values?
	// 			List: []syntax.Stmt{
	// 				&syntax.ExprStmt{
	// 					X: &syntax.CallExpr{
	// 						Fun: &syntax.SelectorExpr{
	// 							X: &syntax.Name{
	// 								Value: "i",
	// 							},
	// 							Sel: &syntax.Name{
	// 								Value: ifo.priVarName, // unexproted variable
	// 							},
	// 						},
	// 						ArgList: nil,
	// 						HasDots: false,
	// 					},
	// 				},
	// 			},
	// 			Rbrace: syntax.Pos{},
	// 		},
	// 	})
	// }

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

	// append items to the file

	// add items to info.def (and possibly others)
}

// A rewriter implements rewriting the range-over-funcs in a given function.
type rewriter struct {
	pkg  *types2.Package
	info *types2.Info
	file *syntax.File
	cl   *syntax.CompositeLit
}

func (r *rewriter) inspect(n syntax.Node) bool {
	fmt.Printf("--> inspect=%T\n", n)

	switch n := n.(type) {
	case *syntax.CompositeLit:
		println(n)
		return false
	default:
	case nil:
	}
	return true
}

func makeStruct(pos syntax.Pos, name string, fields []*types2.Var, tags []string, pkg *types2.Package, info *types2.Info) (*types2.Struct, *syntax.Name) {
	obj := types2.NewStruct(fields, tags)
	n := syntax.NewName(pos, name)
	// tv := syntax.TypeAndValue{Type: typ}
	// tv.SetIsValue()
	// n.SetTypeInfo(tv)
	// info.Defs[n] = *obj
	return obj, n
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
