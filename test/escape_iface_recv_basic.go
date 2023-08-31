// errorcheck -0 -m -l

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test basic escape analysis for interface receivers.
//
// This file has some overview examples that broadly cover
// the core behavior with some commentary.
//
// Starting in Go 1.22, escape analysis examines method calls on interfaces
// to see if it can prove that the interface receiver cannot be leaked
// by the called method. (Prior to Go 1.22, almost all of these examples
// would show escaping variables and parameters).
package example

var sink any // a global sink used to force escapes to the heap.

// In general, methods can leak their receivers to the heap.
type FooEscaping struct{ a, b int }

func (f *FooEscaping) String() string { // ERROR "leaking param: f$"
	sink = f
	return ""
}

// If a value is used as the receiver in an interface method call,
// that value must always escape if the called method leaks its receiver to the heap.
func f1() {
	var eface any
	val := FooEscaping{1, 2}
	eface = val // ERROR "val escapes to heap$"
	if v, ok := eface.(Stringer); ok {
		_ = v.String()
	}
}

// Types without methods.

// In contrast, we understand that an int stored in an interface
// does not have any methods capable of leaking a receiver,
// and hence the int does not escape due to use as an interface receiver.
func f2() {
	var eface any
	val := 1000
	eface = val // ERROR "val does not escape$"
	if v, ok := eface.(Stringer); ok {
		_ = v.String()
	}
}

type Point struct{ x, y int }

// We similarly understand that a struct without any methods that is stored in an interface
// does not escape due to use as an interface receiver.
func f3() {
	var eface any
	val := Point{}
	eface = val // ERROR "val does not escape$"
	if v, ok := eface.(Stringer); ok {
		_ = v.String()
	}
}

// Propagating uncertainty.

// If a function has an interface parameter, and the function calls
// a method on the interface, we do not know if the value stored
// in the interface parameter will leak the receiver.
// However, we understand that it might, and we propagate that possibility.
func Print1(arg any) { // ERROR "might leak param: arg$"
	switch v := arg.(type) {
	case Stringer:
		println(v.String())
	default:
		println(arg)
	}
}

// We also propagate that possibility across multiple calling layers.
func print2(arg any) { // ERROR "might leak param: arg$"
	switch v := arg.(type) {
	case Stringer:
		println(v.String())
	default:
		println(arg)
	}
}

func Print2(arg any) { // ERROR "might leak param: arg$"
	print2(arg)
}

// Resolving uncertainty.

// If our print functions are called with the FooEscaping type,
// we are able to settle the uncertainty and conclude the value escapes.
func f4() {
	val := FooEscaping{1, 2}
	Print1(val) // ERROR "val escapes to heap$"
	Print2(val) // ERROR "val escapes to heap$"
}

// In contrast, when calling our print functions on an int type,
// we settle the uncertainty in the opposite direction and conclude the value
// does not have any methods capable of leaking a receiver,
// and hence we can conclude the int value does not escape.
func f5() {
	val := 1000
	Print1(val) // ERROR "val does not escape$"
	Print2(val) // ERROR "val does not escape$"
}

// Similarly, when calling the print functions on a Point type that has no methods,
// we can conclude the value does not escape.
func f6() {
	val := Point{1, 2}
	Print1(val) // ERROR "val does not escape$"
	Print2(val) // ERROR "val does not escape$"
}

// Methodless types with fields.

// In this example, a function parameter has no methods but has an interface field,
// and the function calls a method on that interface.
//
// We do not know if the value stored in the interface can leak the method receiver.
// We propagate the possibility that the interface might or might not leak the receiver
// until we can analyze the type stored in the interface field.
type Bar struct {
	eface any
	a, b  int
}

func print3(bar Bar) { // ERROR "might leak param: bar$"
	switch v := bar.eface.(type) {
	case Stringer:
		println(v.String())
	default:
		println(bar.eface)
	}
}

func Print3(arg any) { // ERROR "might leak param: arg$"
	bar := Bar{eface: arg}
	print3(bar)
}

func f7() {
	val := FooEscaping{}
	Print3(val) // ERROR "val escapes to heap$"
}

func f8() {
	val := 1000
	Print3(val) // ERROR "val does not escape$"
}

func f9() {
	val := Point{1, 2}
	Print3(val) // ERROR "val does not escape$"
}

// Types with methods and fields.

// Even if a type's methods do not have pointer receivers,
// a method can still leak one of the type's fields to the heap.
type FooEscapingField1 struct {
	ptr *int
	a   int
}

type FooEscapingField2 struct {
	eface any
	a     int
}

func (f FooEscapingField1) String() string { // ERROR "leaking param: f$"
	sink = f.ptr
	return ""
}

func (f FooEscapingField2) String() string { // ERROR "leaking param: f$"
	switch v := f.eface.(type) {
	case *int:
		sink = v
	}
	return ""
}

// The next two examples make use of these FooEscapingFieldN types.
//
// Backing up, when a value is used as the receiver in an interface method call,
// a part of the analysis is we examine the value's type recursively to
// see if we can prove its fields are not retained.
//
// In some of the examples above, we do prove that.
//
// In these next two examples, that cannot be proved -- correctness requires
// val to be placed on the heap. The foo value stores a pointer to val.
// In theory, foo could be on the stack pointing to val on the heap,
// but currently foo is placed on the heap as well (our analysis of data flow
// and types is currently conservative).
// TODO: improve explanation.
//
// TODO: this would be better using types imported from an external package.
func f10() {
	val := 1000 // ERROR "moved to heap: val$"
	foo := FooEscapingField1{}
	foo.ptr = &val
	Print1(foo) // ERROR "foo escapes to heap$"
}

func f11() {
	val := 1000
	foo := FooEscapingField2{}
	foo.eface = val // ERROR "val escapes to heap$"
	Print1(foo)     // ERROR "foo escapes to heap$"
}

// Using a type's escape analysis results.

// If escape analysis has completed on a type, such as if the type is from an external package,
// we understand when no method on the type has any escaping receivers, and
// hence can conclude that using the type in an interface method call cannot leak the receiver.
//
// First, a simple example, then two examples showing we check this recursively.
type FooNonEscaping struct{ a, b int }

func (f *FooNonEscaping) String() string { return "" } // ERROR "f does not escape$"

func f12() {
	val := FooNonEscaping{1, 2}
	// Force escape analysis to be complete on val's type before f12 is analyzed.
	// TODO: better to use an imported package here, probably. (Using escape analysis results for
	// types in the same package is a somewhat aggressive optimization in some cases, and we have
	// a TODO in the code to consider being more conservative).
	Print1(val) // ERROR "val does not escape$"
}

type BarEscaping struct{ foo FooEscaping }
type BarNonEscaping struct{ foo FooNonEscaping }

func f12a() {
	val := BarEscaping{foo: FooEscaping{1, 2}}
	Print1(val) // ERROR "val escapes to heap$"
}

func f12b() {
	val := BarNonEscaping{foo: FooNonEscaping{1, 2}}
	Print1(val) // ERROR "val does not escape$"
}

type Stringer interface{ String() string }
