// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import "unsafe"

// type algorithms - known to compiler
// a copy of runtime/alg.go - keep in sync
const (
	alg_MEM = iota
	alg_MEM0
	alg_MEM8
	alg_MEM16
	alg_MEM32
	alg_MEM64
	alg_MEM128
	alg_NOEQ
	alg_NOEQ0
	alg_NOEQ8
	alg_NOEQ16
	alg_NOEQ32
	alg_NOEQ64
	alg_NOEQ128
	alg_STRING
	alg_INTER
	alg_NILINTER
	alg_SLICE
	alg_FLOAT32
	alg_FLOAT64
	alg_CPLX64
	alg_CPLX128
	alg_max
)

// a copy of runtime.typeAlg
type typeAlg struct {
	// function for hashing objects of this type
	// (ptr to object, size, seed) -> hash
	hash func(unsafe.Pointer, uintptr, uintptr) uintptr
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B, size) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer, uintptr) bool
}

// implemented in runtime as reflect_algarray
func algarray(i int) *typeAlg

// algtype1 returns the index into algarray or -1 and the associated typeAlg (or nil)
//
// algtype1 is modeled after cmd/gc/subr.c's algtype1
func algtype1(t Type) (int, *typeAlg) {
	switch t.Kind() {
	case Int8, Int16, Int32, Int64,
		Uint8, Uint16, Uint32, Uint64,
		Int, Uint, Uintptr,
		Bool,
		Ptr,
		Chan, UnsafePointer:
		return alg_MEM, algarray(alg_MEM)

	case Func, Map:
		return alg_NOEQ, algarray(alg_NOEQ)

	case Float32:
		return alg_FLOAT32, algarray(alg_FLOAT32)

	case Float64:
		return alg_FLOAT64, algarray(alg_FLOAT64)

	case Complex64:
		return alg_CPLX64, algarray(alg_CPLX64)

	case Complex128:
		return alg_CPLX128, algarray(alg_CPLX128)

	case String:
		return alg_STRING, algarray(alg_STRING)

	case Interface:
		if t.NumMethod() <= 0 {
			return alg_NILINTER, algarray(alg_NILINTER)
		}
		return alg_INTER, algarray(alg_INTER)

	case Array:
		et := t.Elem()
		alg := et.common().alg
		if alg == nil {
			panic("reflect: type " + et.Name() + " has nil typeAlg")
		}
		return -1, alg

	case Slice:
		return alg_SLICE, algarray(alg_SLICE)

	case Struct:
		panic("not implemented")

	default:
		panic("unknown kind: " + t.Kind().String())
	}

	return 0, nil
}

// algtype returns the correct typeAlg for the given Type.
//
// algtype is modeled after cmd/gc/subr.c's algtype
func algtype(t Type) *typeAlg {
	i, alg := algtype1(t)
	if i == alg_MEM || i == alg_NOEQ {
		switch t.Size() {
		case 0:
			i += alg_MEM0 - alg_MEM
		case 1:
			i += alg_MEM8 - alg_MEM
		case 2:
			i += alg_MEM16 - alg_MEM
		case 4:
			i += alg_MEM32 - alg_MEM
		case 8:
			i += alg_MEM64 - alg_MEM
		case 16:
			i += alg_MEM128 - alg_MEM
		}
		alg = algarray(i)
	}

	if alg == nil {
		panic("reflect: nil typeAlg for type " + t.Name())
	}

	return alg
}
