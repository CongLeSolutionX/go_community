// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

type Type interface {
	RemoteReference
	Underlying() Type
}

type BasicKind int

const (
	Invalid BasicKind = iota // type is invalid

	// predeclared types
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer

	// types for untyped values
	UntypedBool
	UntypedInt
	UntypedRune
	UntypedFloat
	UntypedComplex
	UntypedString
	UntypedNil

	// aliases
	Byte = Uint8
	Rune = Int32
)

// BasicInfo is a set of flags describing properties of a basic type.
type BasicInfo int

// Properties of basic types.
const (
	IsBoolean BasicInfo = 1 << iota
	IsInteger
	IsUnsigned
	IsFloat
	IsComplex
	IsString
	IsUntyped

	IsOrdered   = IsInteger | IsFloat | IsString
	IsNumeric   = IsInteger | IsFloat | IsComplex
	IsConstType = IsBoolean | IsNumeric | IsString
)

// A Basic represents a basic type.
type Basic struct {
	reference

	kind BasicKind
	info BasicInfo
	name string
}

func (b *Basic) Underlying() Type { return b }

// Kind returns the kind of basic type b.
func (b *Basic) Kind() BasicKind { return b.kind }

// Info returns information about properties of basic type b.
func (b *Basic) Info() BasicInfo { return b.info }

// Name returns the name of basic type b.
func (b *Basic) Name() string { return b.name }

type Struct struct {
	reference
	fields []*Var
	tags   []string // field tags; nil if there are no tags
}

func NewStruct(fields []*Var, tags []string) *Struct {
	var fset objset
	for _, f := range fields {
		if f.name != "_" && fset.insert(f) != nil {
			panic("multiple fields with the same name")
		}
	}
	if len(tags) > len(fields) {
		panic("more tags than fields")
	}
	return &Struct{fields: fields, tags: tags}
}

// NumFields returns the number of fields in the struct (including blank and embedded fields).
func (s *Struct) NumFields() int { return len(s.fields) }

// Field returns the i'th field for 0 <= i < NumFields().
func (s *Struct) Field(i int) *Var { return s.fields[i] }

// Tag returns the i'th field tag for 0 <= i < NumFields().
func (s *Struct) Tag(i int) string {
	if i < len(s.tags) {
		return s.tags[i]
	}
	return ""
}

func (s *Struct) Underlying() Type { return s }

type typeInfo uint

// A Named represents a named type.
type Named struct {
	reference
	info       typeInfo  // for cycle detection
	obj        *TypeName // corresponding declared object
	orig       Type      // type (on RHS of declaration) this *Named type is derived of (for cycle reporting)
	underlying Type      // possibly a *Named during setup; never a *Named once set up completely
}

func NewNamed(obj *TypeName, underlying Type) *Named {
	if _, ok := underlying.(*Named); ok {
		panic("types.NewNamed: underlying type must not be *Named")
	}
	typ := &Named{obj: obj, orig: underlying, underlying: underlying}
	if obj.typ == nil {
		obj.typ = typ
	}
	return typ
}

func (n *Named) Obj() Object { return n.obj }

func (n *Named) Underlying() Type { return n.underlying }

func (n *Named) SetUnderlying(under Type) {
	if n != nil {
		n.underlying = under
	}
}
