// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime _type representation.

package core

import (
	"unsafe"
)

// Needs to be in sync with ../../cmd/ld/decodesym.c:/^commonsize and pkg/reflect/type.go:/type.
type Type struct {
	Size       uintptr
	Hash       uint32
	_unused    uint8
	Align      uint8
	fieldalign uint8
	Kind       uint8
	Alg        unsafe.Pointer
	// gc stores _type info required for garbage collector.
	// If (kind&KindGCProg)==0, then gc[0] points at sparse GC bitmap
	// (no indirection), 4 bits per word.
	// If (kind&KindGCProg)!=0, then gc[1] points to a compiler-generated
	// read-only GC program; and gc[0] points to BSS space for sparse GC bitmap.
	// For huge _types (>MaxGCMask), runtime unrolls the program directly into
	// GC bitmap and gc[0] is not used. For moderately-sized _types, runtime
	// unrolls the program into gc[0] space on first use. The first byte of gc[0]
	// (gc[0][0]) contains 'unroll' flag saying whether the program is already
	// unrolled into gc[0] or not.
	Gc     [2]uintptr
	String *string
	X      *Uncommontype
	Ptrto  *Type
	Zero   *byte // ptr to the zero value for this _type
}

type Method struct {
	Name    *string
	Pkgpath *string
	Mtyp    *Type
	typ     *Type
	Ifn     unsafe.Pointer
	tfn     unsafe.Pointer
}

type Uncommontype struct {
	Name    *string
	Pkgpath *string
	Mhdr    []Method
	m       [0]Method
}

type Imethod struct {
	Name    *string
	Pkgpath *string
	Type    *Type
}

type Interfacetype struct {
	Typ  Type
	Mhdr []Imethod
	m    [0]Imethod
}
