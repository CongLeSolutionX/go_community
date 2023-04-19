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
	// TFlagUncommon means that there is a pointer, *UncommonType,
	// just beyond the outer type structure.
	//
	// For example, if t.Kind() == Struct and t.tflag&tflagUncommon != 0,
	// then t has UncommonType data and it can be accessed as:
	//
	//	type tUncommon struct {
	//		structType
	//		u UncommonType
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

// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
type Kind uint

const (
	Invalid Kind = iota
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
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Slice
	String
	Struct
	UnsafePointer
)

const (
	KindDirectIface = 1 << 5
	KindGCProg      = 1 << 6 // Type.gc points to GC program
	KindMask        = (1 << 5) - 1
)

// String returns the name of k.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return kindNames[0]
}

var kindNames = []string{
	Invalid:       "invalid",
	Bool:          "bool",
	Int:           "int",
	Int8:          "int8",
	Int16:         "int16",
	Int32:         "int32",
	Int64:         "int64",
	Uint:          "uint",
	Uint8:         "uint8",
	Uint16:        "uint16",
	Uint32:        "uint32",
	Uint64:        "uint64",
	Uintptr:       "uintptr",
	Float32:       "float32",
	Float64:       "float64",
	Complex64:     "complex64",
	Complex128:    "complex128",
	Array:         "array",
	Chan:          "chan",
	Func:          "func",
	Interface:     "interface",
	Map:           "map",
	Pointer:       "ptr",
	Slice:         "slice",
	String:        "string",
	Struct:        "struct",
	UnsafePointer: "unsafe.Pointer",
}

func (t *Type) Kind() Kind { return Kind(t.Kind_ & KindMask) }

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

// UncommonType is present only for defined types or types with methods
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

type ChanType struct {
	Type
	Elem *Type
	Dir  uintptr
}

type StructTypeUncommon struct {
	StructType
	u UncommonType
}

func (t *Type) Uncommon() *UncommonType {
	if t.TFlag&TFlagUncommon == 0 {
		return nil
	}
	switch t.Kind() {
	case Struct:
		return &(*StructTypeUncommon)(unsafe.Pointer(t)).u
	case Pointer:
		type u struct {
			PtrType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Func:
		type u struct {
			FuncType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Slice:
		type u struct {
			SliceType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Array:
		type u struct {
			ArrayType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Chan:
		type u struct {
			ChanType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Map:
		type u struct {
			MapType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Interface:
		type u struct {
			InterfaceType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	default:
		type u struct {
			Type
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	}
}

func (t *Type) Size() uintptr { return t.Size_ }

func (t *Type) Align() int { return int(t.Align_) }

func (t *Type) FieldAlign() int { return int(t.FieldAlign_) }

type InterfaceType struct {
	Type
	PkgPath Name
	Mhdr    []Imethod
}

type MapType struct {
	Type
	Key    *Type
	Elem   *Type
	Bucket *Type // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	KeySize    uint8  // size of key slot
	ElemSize   uint8  // size of elem slot
	BucketSize uint16 // size of bucket
	Flags      uint32
}

type SliceType struct {
	Type
	Elem *Type
}

type FuncType struct {
	Type
	InCount  uint16
	OutCount uint16
}

type PtrType struct {
	Type
	Elem *Type
}

type StructField struct {
	Name   Name
	Typ    *Type
	Offset uintptr
}

type StructType struct {
	Type
	PkgPath Name
	Fields  []StructField
}

// Name is an encoded type Name with optional extra data.
//
// The first byte is a bit field containing:
//
//	1<<0 the name is exported
//	1<<1 tag data follows the name
//	1<<2 pkgPath nameOff follows the name and tag
//	1<<3 the name is of an embedded (a.k.a. anonymous) field
//
// Following that, there is a varint-encoded length of the name,
// followed by the name itself.
//
// If tag data is present, it also has a varint-encoded length
// followed by the tag itself.
//
// If the import path follows, then 4 bytes at the end of
// the data form a nameOff. The import path is only set for concrete
// methods that are defined in a different package than their type.
//
// If a name starts with "*", then the exported bit represents
// whether the pointed to type is exported.
//
// Note: this encoding must match here and in:
//   cmd/compile/internal/reflectdata/reflect.go
//   runtime/type.go
//   internal/reflectlite/type.go
//   cmd/link/internal/ld/decodesym.go

type Name struct {
	Bytes *byte
}

func (n Name) Data(off int, whySafe string) *byte {
	return (*byte)(add(unsafe.Pointer(n.Bytes), uintptr(off), whySafe))
}

func (n Name) IsExported() bool {
	return (*n.Bytes)&(1<<0) != 0
}

func (n Name) HasTag() bool {
	return (*n.Bytes)&(1<<1) != 0
}

func (n Name) Embedded() bool {
	return (*n.Bytes)&(1<<3) != 0
}

// ReadVarint parses a varint as encoded by encoding/binary.
// It returns the number of encoded bytes and the encoded value.
func (n Name) ReadVarint(off int) (int, int) {
	v := 0
	for i := 0; ; i++ {
		x := *n.Data(off+i, "read varint")
		v += int(x&0x7f) << (7 * i)
		if x&0x80 == 0 {
			return i + 1, v
		}
	}
}

// writeVarint writes n to buf in varint form. Returns the
// number of bytes written. n must be nonnegative.
// Writes at most 10 bytes.
func writeVarint(buf []byte, n int) int {
	for i := 0; ; i++ {
		b := byte(n & 0x7f)
		n >>= 7
		if n == 0 {
			buf[i] = b
			return i + 1
		}
		buf[i] = b | 0x80
	}
}

func (n Name) Name() string {
	if n.Bytes == nil {
		return ""
	}
	i, l := n.ReadVarint(1)
	return unsafeStringFor(n.Data(1+i, "non-empty string"), l)
}

func (n Name) Tag() string {
	if !n.HasTag() {
		return ""
	}
	i, l := n.ReadVarint(1)
	i2, l2 := n.ReadVarint(1 + i + l)
	return unsafeStringFor(n.Data(1+i+l+i2, "non-empty string"), l2)
}

func NewName(n, tag string, exported, embedded bool) Name {
	if len(n) >= 1<<29 {
		panic("reflect.nameFrom: name too long: " + n[:1024] + "...")
	}
	if len(tag) >= 1<<29 {
		panic("reflect.nameFrom: tag too long: " + tag[:1024] + "...")
	}
	var nameLen [10]byte
	var tagLen [10]byte
	nameLenLen := writeVarint(nameLen[:], len(n))
	tagLenLen := writeVarint(tagLen[:], len(tag))

	var bits byte
	l := 1 + nameLenLen + len(n)
	if exported {
		bits |= 1 << 0
	}
	if len(tag) > 0 {
		l += tagLenLen + len(tag)
		bits |= 1 << 1
	}
	if embedded {
		bits |= 1 << 3
	}

	b := make([]byte, l)
	b[0] = bits
	copy(b[1:], nameLen[:nameLenLen])
	copy(b[1+nameLenLen:], n)
	if len(tag) > 0 {
		tb := b[1+nameLenLen+len(n):]
		copy(tb, tagLen[:tagLenLen])
		copy(tb[tagLenLen:], tag)
	}

	return Name{Bytes: &b[0]}
}
