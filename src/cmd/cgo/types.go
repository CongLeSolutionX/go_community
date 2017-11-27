// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Load types.Info for type calculations.

package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

func (p *Package) TypeCheck(files []*File) {
	cpkg := p.newCgoPackage()

	cfg := types.Config{
		Error: func(err error) {
			error_(token.NoPos, err.Error())
		},
		Importer: cgoImporter{
			Default: importer.For("source", nil).(types.ImporterFrom),
			C:       cpkg,
		},
		Sizes: &types.StdSizes{WordSize: p.PtrSize, MaxAlign: p.PtrSize},
	}

	fs := make([]*ast.File, len(files))
	for i, f := range files {
		fs[i] = f.AST
	}

	pkg, err := cfg.Check("cmd/cgo", fset, fs, nil)
	if err != nil {
		error_(token.NoPos, "error occurred during type checking")
		return
	}

	// for types.EvalExpr
	for _, imp := range pkg.Imports() {
		pkg.Scope().Insert(types.NewPkgName(token.NoPos, pkg, imp.Name(), imp))
	}

	p.Type = pkg
}

type cgoImporter struct {
	Default types.ImporterFrom
	C       *cgoPackage
}

func (imp cgoImporter) Import(path string) (*types.Package, error) {
	if path == "C" {
		return imp.C.Package, nil
	}
	return imp.Default.Import(path)
}

func (imp cgoImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	if path == "C" {
		return imp.C.Package, nil
	}
	return imp.Default.ImportFrom(path, dir, mode)
}

type cgoPackage struct {
	*types.Package
}

func (p *Package) newCgoPackage() *cgoPackage {
	cpkg := &cgoPackage{
		Package: types.NewPackage("C", "C"),
	}

	for mname, t := range typedef {
		if strings.HasPrefix(mname, "_Ctype_") {
			_, err := cpkg.insertTypedef(mname[7:], t)
			if err != nil {
				error_(token.NoPos, err.Error())
			}
		}
	}

L:
	for _, goname := range nameKeys(p.Name) {
		var (
			typ types.Type
			obj types.Object
			err error
		)

		n := p.Name[goname]

		goname = fixGo(goname)

		switch n.Kind {
		case "iconst":
			obj = types.NewConst(token.NoPos, cpkg.Package, goname, types.Typ[types.UntypedInt], constant.MakeFromLiteral(n.Const, token.INT, 0))
		case "fconst":
			obj = types.NewConst(token.NoPos, cpkg.Package, goname, types.Typ[types.UntypedFloat], constant.MakeFromLiteral(n.Const, token.FLOAT, 0))
		case "sconst":
			obj = types.NewConst(token.NoPos, cpkg.Package, goname, types.Typ[types.UntypedString], constant.MakeFromLiteral(n.Const, token.STRING, 0))
		case "type":
			if cpkg.lookupType(goname) == nil {
				error_(token.NoPos, "cannot find definitions for C.%s", goname)
			}
			continue L
		case "var", "fpvar":
			typ, err = cpkg.guessType(n.Type.Go)
			obj = types.NewVar(token.NoPos, cpkg.Package, goname, typ)
		case "func", "vfunc":
			obj, err = cpkg.guessFunc(goname, n.FuncType)
		case "macro":
			typ, err = cpkg.guessType(n.FuncType.Result.Go)
			obj = types.NewVar(token.NoPos, cpkg.Package, goname, typ) // handle as var.
		}

		if obj == nil {
			if err != nil {
				error_(token.NoPos, "cannot create object for C.%s: %v", goname, err)
			} else {
				error_(token.NoPos, "cannot create object for C.%s", goname)
			}
		} else {
			cpkg.Scope().Insert(obj)
		}
	}

	cpkg.MarkComplete()

	return cpkg
}

func (cpkg *cgoPackage) lookupType(goname string) types.Type {
	obj := cpkg.Scope().Lookup(goname)
	if obj == nil {
		return nil
	}
	return obj.Type()
}

func (cpkg *cgoPackage) guessType(expr ast.Expr) (typ types.Type, err error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		switch mname := expr.Name; mname {
		case "bool":
			typ = types.Typ[types.Bool]
		case "byte":
			typ = types.Typ[types.Byte]
		case "int8":
			typ = types.Typ[types.Int8]
		case "int16":
			typ = types.Typ[types.Int16]
		case "int32":
			typ = types.Typ[types.Int32]
		case "int64":
			typ = types.Typ[types.Int64]
		case "uint8":
			typ = types.Typ[types.Uint8]
		case "uint16":
			typ = types.Typ[types.Uint16]
		case "uint32":
			typ = types.Typ[types.Uint32]
		case "uint64":
			typ = types.Typ[types.Uint64]
		case "uintptr":
			typ = types.Typ[types.Uintptr]
		case "float32":
			typ = types.Typ[types.Float32]
		case "float64":
			typ = types.Typ[types.Float64]
		case "complex64":
			typ = types.Typ[types.Complex64]
		case "complex128":
			typ = types.Typ[types.Complex128]
		case "string":
			typ = types.Typ[types.String]
		case "unsafe.Pointer":
			typ = types.Typ[types.UnsafePointer]
		case "struct{}":
			typ = types.NewStruct(nil, nil)
		case "[]byte":
			typ = types.NewSlice(types.Typ[types.Byte])
		case "...":
			typ = types.NewSlice(types.NewInterface(nil, nil)) // just for avoiding errors
		default:
			if strings.HasPrefix(mname, "_Ctype_") {
				typ, err = cpkg.insertTypedef(mname[7:], typedef[mname])
				if err != nil {
					return nil, err
				}
			}
		}
	case *ast.SelectorExpr:
		if x, ok := expr.X.(*ast.Ident); ok && x.Name == "unsafe" && expr.Sel.Name == "Pointer" {
			typ = types.Typ[types.UnsafePointer]
		}
	case *ast.ArrayType:
		etyp, err := cpkg.guessType(expr.Elt)
		if err != nil {
			return nil, err
		}
		switch l := expr.Len.(type) {
		case nil:
			typ = types.NewSlice(etyp)
		case *ast.BasicLit:
			len, err := strconv.ParseInt(l.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			typ = types.NewArray(etyp, len)
		}
	case *ast.StarExpr:
		xtyp, err := cpkg.guessType(expr.X)
		if err != nil {
			return nil, err
		}
		typ = types.NewPointer(xtyp)
	case *ast.StructType:
		var fields []*types.Var
		if expr.Fields.List != nil {
			fields = make([]*types.Var, 0, len(expr.Fields.List)*3/2)
			for _, f := range expr.Fields.List {
				t, err := cpkg.guessType(f.Type)
				if err != nil {
					return nil, err
				}
				if f.Names == nil {
					fields = append(fields, types.NewVar(f.Pos(), cpkg.Package, "", t))
				} else {
					for _, name := range f.Names {
						fields = append(fields, types.NewVar(f.Pos(), cpkg.Package, name.Name, t))
					}
				}
			}
		}
		typ = types.NewStruct(fields, nil)
	}

	if typ == nil {
		return nil, fmt.Errorf("cannot guess type of %s", types.ExprString(expr))
	}

	return typ, nil
}

func (cpkg *cgoPackage) guessFunc(goname string, fntype *FuncType) (types.Object, error) {
	var params *types.Tuple
	if len(fntype.Params) != 0 {
		ps := make([]*types.Var, len(fntype.Params))
		for i, p := range fntype.Params {
			t, err := cpkg.guessType(p.Go)
			if err != nil {
				return nil, err
			}
			ps[i] = types.NewVar(token.NoPos, cpkg.Package, "", t)
		}
		params = types.NewTuple(ps...)
	}

	var results *types.Tuple
	if fntype.Result == nil {
		// void is assignable due to backward compatibility
		t := types.NewArray(types.Typ[types.Byte], 0)
		results = types.NewTuple(types.NewVar(token.NoPos, cpkg.Package, "", t))
	} else {
		t, err := cpkg.guessType(fntype.Result.Go)
		if err != nil {
			return nil, err
		}
		results = types.NewTuple(types.NewVar(token.NoPos, cpkg.Package, "", t))
	}

	var isVariadic bool
	if len(fntype.Params) != 0 {
		if e, ok := fntype.Params[len(fntype.Params)-1].Go.(*ast.Ident); ok && e.Name == "..." {
			isVariadic = true
		}
	}

	return types.NewFunc(
		token.NoPos,
		cpkg.Package,
		goname,
		types.NewSignature(
			nil,
			params,
			results,
			isVariadic,
		),
	), nil
}

func (cpkg *cgoPackage) insertTypedef(goname string, ctyp *Type) (types.Type, error) {
	if ctyp == nil {
		return nil, fmt.Errorf("cannot find definitions for C.%s", fixGo(goname))
	}

	// already inserted
	if typ := cpkg.lookupType(goname); typ != nil {
		return typ, nil
	}

	// type alias
	if ident, ok := ctyp.Go.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "_Ctype_") {
		typ, err := cpkg.guessType(ctyp.Go)
		if err != nil {
			return nil, err
		}

		obj := types.NewTypeName(token.NoPos, cpkg.Package, goname, typ)

		cpkg.Scope().Insert(obj)

		return typ, nil
	}

	obj := types.NewTypeName(token.NoPos, cpkg.Package, goname, nil)

	typ := types.NewNamed(obj, nil, nil)

	cpkg.Scope().Insert(obj) // insert before recursive call

	utyp, err := cpkg.guessType(ctyp.Go)
	if err != nil {
		return nil, err
	}
	if ntyp, ok := utyp.(*types.Named); ok {
		utyp = ntyp.Underlying()
	}

	typ.SetUnderlying(utyp)

	return typ, nil
}
