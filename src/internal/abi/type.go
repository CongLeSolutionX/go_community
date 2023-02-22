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

// Offset is for computing offsets of type data structures at compile/link time;
// the target platform may not be the host platform.  Its state includes the
// current offset, necessary alignment for the sequence of types, and the size
// of pointers and alignment of slices.
type Offset struct {
	off        uint32
	align      uint8
	ptrSize    uint8
	sliceAlign uint8
}

// NewOffset returns a new Offset with offset 0 and alignment 1.
func NewOffset(ptrSize uint8, twoWordAlignSlices bool) Offset {
	if twoWordAlignSlices {
		return Offset{off: 0, align: 1, ptrSize: ptrSize, sliceAlign: 2 * ptrSize}
	}
	return Offset{off: 0, align: 1, ptrSize: ptrSize, sliceAlign: ptrSize}

}

func assertIsAPowerOfTwo(x uint8) {
	if x == 0 {
		panic("Zero is not a power of two")
	}
	y := int(x)
	if y&-y == y {
		return
	}
	panic("Not a power of two")
}

// NewOffset returns a new Offset with specified offset and alignment.
func InitializedOffset(off int, align uint8, ptrSize uint8, twoWordAlignSlices bool) Offset {
	assertIsAPowerOfTwo(align)
	o0 := NewOffset(ptrSize, twoWordAlignSlices)
	o0.off = uint32(off)
	o0.align = align
	return o0
}

func (o Offset) align_(a uint8) Offset {
	o.off = (o.off + uint32(a) - 1) & ^(uint32(a) - 1)
	if o.align < a {
		o.align = a
	}
	return o
}

// Align advances the offset as necessary to obtain an alignment.
// a must be a power of two
func (o Offset) Align(a uint8) Offset {
	assertIsAPowerOfTwo(a)
	return o.align_(a)
}

func (o Offset) plus(x uint32) Offset {
	o = o.align_(uint8(x))
	o.off += x
	return o
}

// D8 appends an 8-bit field to o.
func (o Offset) D8() Offset {
	return o.plus(1)
}

// D16 appends an 16-bit field to o.
func (o Offset) D16() Offset {
	return o.plus(2)
}

// D32 appends an 32-bit field to o.
func (o Offset) D32() Offset {
	return o.plus(4)
}

// D64 appends an 64-bit field to o.
func (o Offset) D64() Offset {
	return o.plus(8)
}

// D64 appends an pointer field to o.
func (o Offset) P() Offset {
	if o.ptrSize == 0 {
		panic("This offset has no defined pointer size")
	}
	return o.plus(uint32(o.ptrSize))
}

// Slice appends a slice field to o.
func (o Offset) Slice() Offset {
	o = o.align_(o.sliceAlign)
	o.off += 3 * uint32(o.ptrSize)
	return o.Align(o.sliceAlign)
}

// String appends a string field to o.
func (o Offset) String() Offset {
	o = o.align_(o.ptrSize)
	o.off += 2 * uint32(o.ptrSize)
	return o
}

// Interface appends an interface field to o.
func (o Offset) Interface() Offset {
	o = o.align_(o.ptrSize)
	o.off += 2 * uint32(o.ptrSize)
	return o
}

// Offset returns the struct-aligned offset (size) of o.
// This is at least as large as the current internal offset; it may be larger.
func (o Offset) Offset() int {
	return int(o.Align(o.align).off)
}

func (o Offset) PlusUncommon() Offset {
	o.off += uint32(UncommonSize())
	return o
}

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

func CommonOffset(ptrSize int, twoWordAlignSlices bool) Offset {
	return InitializedOffset(CommonSize(ptrSize), uint8(ptrSize), uint8(ptrSize), twoWordAlignSlices)
}

func CommonSize(ptrSize int) int      { return 4*ptrSize + 8 + 8 } // sizeof(Type) for a given ptrSize
func StructFieldSize(ptrSize int) int { return 3 * ptrSize }       // sizeof(StructField) for a given ptrSize
func UncommonSize() int               { return 4 + 2 + 2 + 4 + 4 } // sizeof(UncommonType) for a given ptrSize
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
