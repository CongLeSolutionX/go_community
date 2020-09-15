// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file sets up the universe scope and the unsafe package.

package types

import (
	"strings"
)

// Typ contains the predeclared *Basic types indexed by their
// corresponding BasicKind.
//
// The *Basic type for Typ[Byte] will have the name "uint8".
// Use Universe.Lookup("byte").Type() to obtain the specific
// alias basic type named "byte" (and analogous for "rune").
var Typ = []*Basic{
	Invalid: {reference{}, Invalid, 0, "invalid type"},

	Bool:          {reference{}, Bool, IsBoolean, "bool"},
	Int:           {reference{}, Int, IsInteger, "int"},
	Int8:          {reference{}, Int8, IsInteger, "int8"},
	Int16:         {reference{}, Int16, IsInteger, "int16"},
	Int32:         {reference{}, Int32, IsInteger, "int32"},
	Int64:         {reference{}, Int64, IsInteger, "int64"},
	Uint:          {reference{}, Uint, IsInteger | IsUnsigned, "uint"},
	Uint8:         {reference{}, Uint8, IsInteger | IsUnsigned, "uint8"},
	Uint16:        {reference{}, Uint16, IsInteger | IsUnsigned, "uint16"},
	Uint32:        {reference{}, Uint32, IsInteger | IsUnsigned, "uint32"},
	Uint64:        {reference{}, Uint64, IsInteger | IsUnsigned, "uint64"},
	Uintptr:       {reference{}, Uintptr, IsInteger | IsUnsigned, "uintptr"},
	Float32:       {reference{}, Float32, IsFloat, "float32"},
	Float64:       {reference{}, Float64, IsFloat, "float64"},
	Complex64:     {reference{}, Complex64, IsComplex, "complex64"},
	Complex128:    {reference{}, Complex128, IsComplex, "complex128"},
	String:        {reference{}, String, IsString, "string"},
	UnsafePointer: {reference{}, UnsafePointer, 0, "Pointer"},

	UntypedBool:    {reference{}, UntypedBool, IsBoolean | IsUntyped, "untyped bool"},
	UntypedInt:     {reference{}, UntypedInt, IsInteger | IsUntyped, "untyped int"},
	UntypedRune:    {reference{}, UntypedRune, IsInteger | IsUntyped, "untyped rune"},
	UntypedFloat:   {reference{}, UntypedFloat, IsFloat | IsUntyped, "untyped float"},
	UntypedComplex: {reference{}, UntypedComplex, IsComplex | IsUntyped, "untyped complex"},
	UntypedString:  {reference{}, UntypedString, IsString | IsUntyped, "untyped string"},
	UntypedNil:     {reference{}, UntypedNil, IsUntyped, "untyped nil"},
}

var aliases = [...]*Basic{
	{reference{}, Byte, IsInteger | IsUnsigned, "byte"},
	{reference{}, Rune, IsInteger, "rune"},
}

func defPredeclaredTypes(scope Scope) {
	for _, t := range Typ {
		def(NewTypeName(nil, nil, t.name, t), scope)
	}
	for _, t := range aliases {
		def(NewTypeName(nil, nil, t.name, t), scope)
	}
}

func BuildUniverse(universe Scope) {
	defPredeclaredTypes(universe)
}

// Objects with names containing blanks are internal and not entered into
// a scope. Objects with exported names are inserted in the unsafe package
// scope; other objects are inserted in the universe scope.
//
func def(obj Object, scope Scope) {
	assert(obj.color() == black)
	name := obj.Name()
	if strings.Contains(name, " ") {
		return // nothing to do
	}
	// fix Obj link for named types
	if typ, ok := obj.Type().(*Named); ok {
		typ.obj = obj.(*TypeName)
	}
	if scope.Insert(obj) != nil {
		panic("internal error: double declaration")
	}
}
