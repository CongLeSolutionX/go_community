// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

import (
	"unsafe"
)

// TFlag is used by an Type to signal what extra type information is
// available in the memory directly following the Type value.
type TFlag uint8

const (
	// TFlagUncommon means that there is a pointer, *uncommonType,
	// just beyond the outer type structure.
	//
	// For example, if t.Kind() == Struct and t.tflag&tflagUncommon != 0,
	// then t has uncommonType data and it can be accessed as:
	//
	//	type tUncommon struct {
	//		structType
	//		u uncommonType
	//	}
	//	u := &(*tUncommon)(unsafe.Pointer(t)).u
	TFlagUncommon TFlag = 1 << 0

	// TFlagExtraStar means the name in the str field has an
	// extraneous '*' prefix. This is because for most types T in
	// a program, the type *T also exists and reusing the str data
	// saves binary size.
	TFlagExtraStar TFlag = 1 << 1

	// TFlagNamed means the type has a name.
	TFlagNamed TFlag = 1 << 2

	// TFlagRegularMemory means that equal and hash functions can treat
	// this type as a single region of t.size bytes.
	TFlagRegularMemory TFlag = 1 << 3
)

// NameOff is the offset to a name from moduledata.types.  See resolveNameOff in runtime.
type NameOff int32

// TypeOff is the offset to a type from moduledata.types.  See resolveTypeOff in runtime.
type TypeOff int32

// TextOff is an offset from the top of a text section.  See (*_type).textOff in runtime.
type TextOff int32

// Type is the runtime representation of a Go type.
//
// Type is also referenced implicitly
// (in the form of expressions involving constants and arch.PtrSize)
// in cmd/compile/internal/reflectdata/reflect.go
// and cmd/link/internal/ld/decodesym.go
// (e.g. data[2*arch.PtrSize+4] references the TFlag field)
// It cannot be used directly because it varies with
// cross compilation and experiments.
type Type struct {
	Size_       uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       TFlag   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       NameOff // string form
	PtrToThis TypeOff // type for pointer to this type, may be zero
}

func CommonSize(ptrSize int) int      { return 4*ptrSize + 8 + 8 } // sizeof(Type) for a given ptrSize
func StructFieldSize(ptrSize int) int { return 3 * ptrSize }       // sizeof(StructField) for a given ptrSize
func UncommonSize(ptrSize int) int    { return 4 + 2 + 2 + 4 + 4 } // sizeof(UncommonType) for a given ptrSize
func IMethodSize(ptrSize int) int     { return 4 + 4 }             // sizeof(IMethod) for a given ptrSize

func KindOff(ptrSize int) int     { return 2*ptrSize + 7 }
func SizeOff(ptrSize int) int     { return 0 }
func PtrBytesOff(ptrSize int) int { return ptrSize }
func TFlagOff(ptrSize int) int    { return 2*ptrSize + 4 }

// Method on non-interface type
type Method struct {
	Name NameOff // name of method
	Mtyp TypeOff // method type (without receiver)
	Ifn  TextOff // fn used in interface call (one-word receiver)
	Tfn  TextOff // fn used for normal method call
}

// uncommonType is present only for defined types or types with methods
// (if T is a defined type, the uncommonTypes for T and *T have methods).
// Using a pointer to this struct reduces the overall size required
// to describe a non-defined type with no methods.
type UncommonType struct {
	PkgPath NameOff // import path; empty for built-in types like int, string
	Mcount  uint16  // number of methods
	Xcount  uint16  // number of exported methods
	Moff    uint32  // offset from this uncommontype to [mcount]method
	_       uint32  // unused
}

func (t *UncommonType) Methods() []Method {
	if t.Mcount == 0 {
		return nil
	}
	return (*[1 << 16]Method)(add(unsafe.Pointer(t), uintptr(t.Moff), "t.mcount > 0"))[:t.Mcount:t.Mcount]
}

func (t *UncommonType) ExportedMethods() []Method {
	if t.Xcount == 0 {
		return nil
	}
	return (*[1 << 16]Method)(add(unsafe.Pointer(t), uintptr(t.Moff), "t.xcount > 0"))[:t.Xcount:t.Xcount]
}

// add returns p+x.
//
// The whySafe string is ignored, so that the function still inlines
// as efficiently as p+x, but all call sites should use the string to
// record why the addition is safe, which is to say why the addition
// does not cause x to advance to the very end of p's allocation
// and therefore point incorrectly at the next block in memory.
func add(p unsafe.Pointer, x uintptr, whySafe string) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

// imethod represents a method on an interface type
type Imethod struct {
	Name NameOff // name of method
	Typ  TypeOff // .(*FuncType) underneath
}

// arrayType represents a fixed array type.
type ArrayType struct {
	Type
	Elem  *Type // array element type
	Slice *Type // slice type
	Len   uintptr
}
