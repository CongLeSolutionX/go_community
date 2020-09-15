// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package minitypes is a pared-down copy of go/types that can only typecheck a
// subset of the go spec. It is used for experimenting with different internal APIs.
package minitypes

import (
	"fmt"
	"go/ast"
	"go/token"

	itypes "internal/types"
)

func CheckFiles(files []*ast.File, fset *token.FileSet) (*Package, *driver, error) {
	var astFiles []itypes.File
	for _, file := range files {
		astFiles = append(astFiles, fileWrapper{file})
	}
	pkg := NewPackage("", "")
	d := newDriver()
	return pkg, d, itypes.CheckFiles(astFiles, d, packageWrapper{pkg})
}

type packageWrapper struct {
	*Package
}

func (w packageWrapper) Scope() itypes.Scope { return scopeWrapper{w.Package.Scope()} }
func (w packageWrapper) SetName(name string) { w.Package.name = name }

type scopeWrapper struct {
	*Scope
}

func wrapScope(scope *Scope) itypes.Scope {
	return scopeWrapper{scope}
}

func unwrapScope(scope itypes.Scope) *Scope {
	if scope == nil {
		return nil
	}
	return scope.(scopeWrapper).Scope
}

func (w scopeWrapper) Insert(inner itypes.Object) itypes.Object {
	obj := mapObj(inner)
	alt := w.Scope.Insert(obj)
	if alt == nil {
		return nil
	}
	return alt.internal()
}

func mapObj(inner itypes.Object) (outer Object) {
	if inner == nil {
		return nil
	}
	if ref := inner.Reference(); ref != nil {
		return ref.(Object)
	}
	defer inner.SetReference(outer)
	switch x := inner.(type) {
	case *itypes.TypeName:
		return &TypeName{object{x}, x}
	case *itypes.Var:
		return &Var{object{x}, x}
	}
	panic(fmt.Sprintf("unhandled internal object %T", inner))
}

func mapType(inner itypes.Type) (outer Type) {
	if inner == nil {
		return nil
	}
	if ref := inner.Reference(); ref != nil {
		return ref.(Type)
	}
	defer inner.SetReference(outer)
	switch t := inner.(type) {
	case *itypes.Basic:
		return &Basic{t}
	case *itypes.Named:
		return &Named{t}
	case *itypes.Struct:
		return &Struct{t}
	}
	panic(fmt.Sprintf("unhandled internal type %T", inner))
}

func (w scopeWrapper) Lookup(name string) itypes.Object {
	obj := w.Scope.Lookup(name)
	if obj == nil {
		return nil
	}
	return obj.internal()
}

type driver struct {
	universe itypes.Scope
	defs     map[*ast.Ident]Object
}

func newDriver() *driver {
	return &driver{
		universe: wrapScope(Universe),
		defs:     make(map[*ast.Ident]Object),
	}
}

func (d *driver) Universe() itypes.Scope {
	return d.universe
}

func (d *driver) Errorf(pos itypes.Pos, format string, args ...interface{}) {
	// TODO: implement, maybe?
}

func (d *driver) RecordDef(iid itypes.Ident, inner itypes.Object) {
	id := iid.(unwrapper).Unwrap().(*ast.Ident)
	if id == nil {
		panic("nil id")
	}
	if m := d.defs; m != nil {
		m[id] = mapObj(inner)
	}
}

func (d *driver) NewScope(parent itypes.Scope, start, end itypes.Pos) itypes.Scope {
	return wrapScope(NewScope(unwrapScope(parent), unwrapPos(start), unwrapPos(end), ""))
}

func (d *driver) NewIdent(name string, pos itypes.Pos) itypes.Ident {
	return wrapIdent(&ast.Ident{Name: name, NamePos: unwrapPos(pos)})
}
