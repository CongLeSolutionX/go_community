// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minitypes

import (
	"go/token"
	itypes "internal/types"
)

// An Object describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
// All objects implement the Object interface.
//
type Object interface {
	Parent() *Scope // scope in which this object is declared; nil for methods and struct fields
	Pos() token.Pos // position of object identifier in declaration
	Pkg() *Package  // package to which this object belongs; nil for labels and objects in the Universe scope
	Name() string   // package local object name
	Type() Type     // object type
	Exported() bool // reports whether the name starts with a capital letter

	setParent(parent *Scope)

	internal() itypes.Object
}

// An object implements the common parts of an Object.
type object struct {
	itypes.Object
}

func (obj *object) internal() itypes.Object {
	return obj.Object
}

// Parent returns the scope in which the object is declared.
// The result is nil for methods and struct fields.
func (obj *object) Parent() *Scope {
	if obj.Object == nil {
		return nil
	}
	return unwrapScope(obj.Object.Parent())
}

func (obj *object) setParent(parent *Scope) {
	obj.Object.SetParent(wrapScope(parent))
}

func (obj *object) Pos() token.Pos {
	if obj.Object == nil {
		return token.NoPos
	}
	return unwrapPos(obj.Object.Pos())
}

func (obj *object) Pkg() *Package {
	if obj.Object == nil {
		return nil
	}
	return obj.Object.Pkg().(packageWrapper).Package
}

func (obj *object) Type() Type { return mapType(obj.Object.Type()) }

func (obj *object) Exported() bool { return token.IsExported(obj.Name()) }

type TypeName struct {
	object
	inner *itypes.TypeName
}

func NewTypeName(pos token.Pos, pkg *Package, name string, typ Type) *TypeName {
	inner := itypes.NewTypeName(wrapPos(pos), packageWrapper{pkg}, name, typ.internal())
	return &TypeName{object{inner}, inner}
}

type Var struct {
	object
	inner *itypes.Var
}

func NewField(pos token.Pos, pkg *Package, name string, typ Type, embedded bool) *Var {
	inner := itypes.NewField(wrapPos(pos), packageWrapper{pkg}, name, typ.internal(), embedded)
	return &Var{object{inner}, inner}
}
