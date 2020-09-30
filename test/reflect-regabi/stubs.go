package main

import "unsafe"

func addInt(a, b int) int
func addFloat64(a, b float64) float64

func sumSpillInt(a, b, c, d, e, f, g, h, i, j int) int
func sumSpillFloat64(a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p float64) float64
func sumSpillMix(a, b, c, d, e, f, g, h, i, j int, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z float64) (int, float64)

func splitSpillInt(a int) (b, c, d, e, f, g, h, i, j, k int)

// Struct1 is a simple integer-only aggregate struct.
type Struct1 struct {
	a, b, c uint
}

// Struct2 is Struct1 but with an array-typed field that will
// force it to get passed on the stack.
type Struct2 struct {
	a, b, c uint
	d       [2]uint32
}

// Struct3 is Struct2 but with an anonymous array-typed field.
// This should act identically to Struct2.
type Struct3 struct {
	a, b, c uint
	_       [2]uint32
}

// Struct4 has byte-length fields that should
// each use up a whole registers.
type Struct4 struct {
	a, b int8
	c, d uint8
	e    bool
}

// Struct5 is a relatively large struct
// with both integer and floating point values.
type Struct5 struct {
	a             uint16
	b             int16
	c, d          uint32
	e             int32
	f, g, h, i, j float32
}

// Struct6 has a nested struct.
type Struct6 struct {
	Struct1
}

// Struct7 is a struct with a nested array-typed field
// that cannot be passed in registers as a result.
type Struct7 struct {
	Struct1
	Struct2
}

// Struct8 is large aggregate struct type that may be
// passed in registers.
type Struct8 struct {
	Struct5
	Struct1
}

// Struct9 is a type that has an array type nested
// 2 layers deep, and as a result needs to be passed
// on the stack.
type Struct9 struct {
	Struct1
	Struct7
}

// Struct10 is a struct type that is too large to be
// passed in registers.
type Struct10 struct {
	Struct5
	Struct8
}

// Struct11 is a struct type that has several reference
// types in it.
type Struct11 struct {
	w unsafe.Pointer
	x map[string]int
	y chan int
	z func() int
}

// Struct12 has Struct11 embedded into it to test more
// paths.
type Struct12 struct {
	a int
	Struct11
}

// Struct13 tests an empty field.
type Struct13 struct {
	a int
	x struct{}
	b int
}

func passArray1(a [1]uint32) [1]uint32
func passArray(a [2]uintptr) [2]uintptr
func passArray1Mix(f int, a [1]uint32, g float64) (int, [1]uint32, float64)
func passString(a string) string
func passInterface(a interface{}) interface{}
func passSlice(a []byte) []byte
func setPointer(a *byte) *byte
func passStruct1(a Struct1) Struct1
func passStruct2(a Struct2) Struct2
func passStruct3(a Struct3) Struct3
func passStruct4(a Struct4) Struct4
func passStruct5(a Struct5) Struct5
func passStruct6(a Struct6) Struct6
func passStruct7(a Struct7) Struct7
func passStruct8(a Struct8) Struct8
func passStruct9(a Struct9) Struct9
func passStruct10(a Struct10) Struct10
func passStruct11(a Struct11) Struct11
func passStruct12(a Struct12) Struct12
func incStruct13(a Struct13) Struct13
func pass2Struct1(a, b Struct1) (x, y Struct1)
func passEmptyStruct(a int, b struct{}, c float64) (int, struct{}, float64)
