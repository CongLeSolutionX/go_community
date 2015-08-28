// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package base

import (
	"unsafe"
)

// Needs to be in sync with ../cmd/internal/ld/decodesym.go:/^func.commonsize,
// ../cmd/internal/gc/reflect.go:/^func.dcommontype and
// ../reflect/type.go:/^type.rtype.
type Type struct {
	Size       uintptr
	Ptrdata    uintptr // size of memory prefix holding all pointers
	Hash       uint32
	_unused    uint8
	Align      uint8
	fieldalign uint8
	Kind       uint8
	Alg        *TypeAlg
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	Gcdata *byte
	String *string
	X      *Uncommontype
	Ptrto  *Type
	Zero   *byte // ptr to the zero value for this type
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
}
