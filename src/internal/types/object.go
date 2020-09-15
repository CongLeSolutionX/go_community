// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "go/token"

type Object interface {
	RemoteReference

	Name() string
	Pos() Pos
	Type() Type
	ID() string

	Parent() Scope
	SetParent(scope Scope)
	Pkg() Package
	SetPackage(pkg Package)

	color() color
	setOrder(uint32)
	setColor(color color)
}

// color encodes the color of an object (see Checker.objDecl for details).
type color uint32

// An object may be painted in one of three colors.
// Color values other than white or black are considered grey.
const (
	white color = iota
	black
	grey // must be > white and black
)

type object struct {
	pkg    Package
	parent Scope

	name string
	pos  Pos

	typ    Type
	order_ uint32
	color_ color

	// TODO: can we eliminate this somehow?
	// scopePos_ itypes.Pos
}

func (obj object) Name() string           { return obj.name }
func (obj object) Pos() Pos               { return obj.pos }
func (obj object) Type() Type             { return obj.typ }
func (obj object) Parent() Scope          { return obj.parent }
func (obj object) Pkg() Package           { return obj.pkg }
func (obj object) SetParent(scope Scope)  { obj.parent = scope }
func (obj object) SetPackage(pkg Package) { obj.pkg = pkg }
func (obj object) color() color           { return obj.color_ }
func (obj *object) setOrder(n uint32)     { obj.order_ = n }
func (obj *object) setColor(color color)  { assert(color != white); obj.color_ = color }
func (obj object) ID() string             { return ID(obj.pkg, obj.name) }

func ID(pkg Package, name string) string {
	if token.IsExported(name) {
		return name
	}
	// unexported names need the package path for differentiation
	// (if there's no package, make sure we don't start with '.'
	// as that may change the order of methods between a setup
	// inside a package and outside a package - which breaks some
	// tests)
	path := "_"
	// pkg is nil for objects in Universe scope and possibly types
	// introduced via Eval (see also comment in object.sameId)
	if pkg != nil && pkg.Path() != "" {
		path = pkg.Path()
	}
	return path + "." + name
}

type Var struct {
	reference
	object
	embedded bool // if set, the variable is an embedded struct field, and name is the type name
	isField  bool // var is struct field
	used     bool // set if the variable was used
}

func NewField(pos Pos, pkg Package, name string, typ Type, embedded bool) *Var {
	return &Var{
		object: object{
			pkg:    pkg,
			name:   name,
			pos:    pos,
			typ:    typ,
			order_: 0,
			color_: colorFor(typ),
		},
		embedded: embedded,
		isField:  true,
	}
}

type TypeName struct {
	reference
	object
}

func NewTypeName(pos Pos, pkg Package, name string, typ Type) *TypeName {
	return &TypeName{
		object: object{
			pkg:    pkg,
			name:   name,
			pos:    pos,
			typ:    typ,
			order_: 0,
			color_: colorFor(typ),
		},
	}
}
