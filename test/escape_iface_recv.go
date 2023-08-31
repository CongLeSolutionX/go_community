// errorcheck -0 -m -l

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for interface receivers.
package example

type Stringer interface{ String() string }

var sink any // a global sink used to force escapes to the heap.

// P1 is a print-ish function.

func P1(arg any) { // ERROR "might leak param: arg$"
	if v, ok := arg.(Stringer); ok {
		_ = v.String()
	}
}

// Define a named type with a method where the pointer receiver escapes.

type IntEscaping int

func (i *IntEscaping) String() string { // ERROR "leaking param: i$"
	sink = i
	return ""
}

// Use P1 with IntEscaping.
// (We then repeat this pattern with a few different types).

func f1(val IntEscaping) {
	P1(val) // ERROR "val escapes to heap$"
}

func f2(val IntEscaping) { // ERROR "moved to heap: val$"
	P1(&val)
}

func f3(ptr *IntEscaping) { // ERROR "leaking param: ptr$"
	P1(ptr)
}

func f4(val IntEscaping) {
	arg := []IntEscaping{val} // ERROR "\[\]IntEscaping{...} escapes to heap$"
	P1(arg)                   // ERROR "arg escapes to heap$"
}

func f5() {
	val := IntEscaping(1000)
	P1(val) // ERROR "val escapes to heap$"
}

func f6() {
	val := IntEscaping(1000) // ERROR "moved to heap: val$"
	P1(&val)
}

func f7() {
	val := IntEscaping(1000)
	arg := []IntEscaping{val} // ERROR "\[\]IntEscaping{...} escapes to heap$"
	P1(arg)                   // ERROR "arg escapes to heap$"
}

func f8() {
	val := IntEscaping(1000)    // ERROR "moved to heap: val$"
	arg := []*IntEscaping{&val} // ERROR "\[\]\*IntEscaping{...} escapes to heap$"
	P1(arg)                     // ERROR "arg escapes to heap$"
}

func f9() {
	val := IntEscaping(1000)   // ERROR "moved to heap: val$"
	s := []*IntEscaping{&val}  // ERROR "\[\]\*IntEscaping{...} escapes to heap$"
	arg := [][]*IntEscaping{s} // ERROR "\[\]\[\]\*IntEscaping{...} escapes to heap$"
	P1(arg)                    // ERROR "arg escapes to heap$"
}

// Define a struct type with a method where the pointer receiver escapes.

type FooEscaping1 struct{ a, b int }

func (f *FooEscaping1) String() string { // ERROR "leaking param: f$"
	sink = f
	return ""
}

// Use P1 with FooEscaping1.
// (These tests follow the pattern of using P1 with IntEscaping above).

func f10(val FooEscaping1) {
	P1(val) // ERROR "val escapes to heap$"
}

func f11(val FooEscaping1) { // ERROR "moved to heap: val$"
	P1(&val)
}

func f12(ptr *FooEscaping1) { // ERROR "leaking param: ptr$"
	P1(ptr)
}

func f13(val FooEscaping1) {
	arg := []FooEscaping1{val} // ERROR "\[\]FooEscaping1{...} escapes to heap$"
	P1(arg)                    // ERROR "arg escapes to heap$"
}

func f14() {
	val := FooEscaping1{1, 2}
	P1(val) // ERROR "val escapes to heap$"
}

func f15() {
	val := FooEscaping1{1, 2} // ERROR "moved to heap: val$"
	P1(&val)
}

func f16() {
	val := FooEscaping1{1, 2}
	arg := []FooEscaping1{val} // ERROR "\[\]FooEscaping1{...} escapes to heap$"
	P1(arg)                    // ERROR "arg escapes to heap$"
}

func f17() {
	val := FooEscaping1{1, 2}    // ERROR "moved to heap: val$"
	arg := []*FooEscaping1{&val} // ERROR "\[\]\*FooEscaping1{...} escapes to heap$"
	P1(arg)                      // ERROR "arg escapes to heap$"
}

func f18() {
	val := FooEscaping1{1, 2}   // ERROR "moved to heap: val$"
	s := []*FooEscaping1{&val}  // ERROR "\[\]\*FooEscaping1{...} escapes to heap$"
	arg := [][]*FooEscaping1{s} // ERROR "\[\]\[\]\*FooEscaping1{...} escapes to heap$"
	P1(arg)                     // ERROR "arg escapes to heap$"
}

// Define a struct type with a method where the pointer receiver escapes and the struct has pointer fields.

type FooEscaping2 struct{ a, b *int }

func (f *FooEscaping2) String() string { // ERROR "leaking param: f$"
	sink = f
	return ""
}

// Use P1 with FooEscaping2.
// (These tests also follow the pattern of using P1 with IntEscaping above).

func f19(val FooEscaping2) { // ERROR "leaking param: val$"
	P1(val) // ERROR "val escapes to heap$"
}

func f20(val FooEscaping2) { // ERROR "moved to heap: val$"
	P1(&val)
}

func f21(ptr *FooEscaping2) { // ERROR "leaking param: ptr$"
	P1(ptr)
}

func f22(val FooEscaping2) { // ERROR "leaking param: val$"
	arg := []FooEscaping2{val} // ERROR "\[\]FooEscaping2{...} escapes to heap$"
	P1(arg)                    // ERROR "arg escapes to heap$"
}

func f23() {
	i := 1000 // ERROR "moved to heap: i$"
	val := FooEscaping2{&i, nil}
	P1(val) // ERROR "val escapes to heap$"
}

func f24() {
	i := 1000                    // ERROR "moved to heap: i$"
	val := FooEscaping2{&i, nil} // ERROR "moved to heap: val$"
	P1(&val)
}

func f25() {
	i := 1000 // ERROR "moved to heap: i$"
	val := FooEscaping2{&i, nil}
	arg := []FooEscaping2{val} // ERROR "\[\]FooEscaping2{...} escapes to heap$"
	P1(arg)                    // ERROR "arg escapes to heap$"
}

func f26() {
	i := 1000                    // ERROR "moved to heap: i$"
	val := FooEscaping2{&i, nil} // ERROR "moved to heap: val$"
	arg := []*FooEscaping2{&val} // ERROR "\[\]\*FooEscaping2{...} escapes to heap$"
	P1(arg)                      // ERROR "arg escapes to heap$"
}

func f27() {
	i := 1000                    // ERROR "moved to heap: i$"
	val := FooEscaping2{&i, nil} // ERROR "moved to heap: val$"
	s := []*FooEscaping2{&val}   // ERROR "\[\]\*FooEscaping2{...} escapes to heap$"
	arg := [][]*FooEscaping2{s}  // ERROR "\[\]\[\]\*FooEscaping2{...} escapes to heap$"
	P1(arg)                      // ERROR "arg escapes to heap$"
}

// Define a struct type without any methods that only has scalar fields.

type Point1 struct{ x, y int }

// Use P1 with Point1.
// (These tests also follow the pattern of using P1 with IntEscaping above).

func f28(val Point1) {
	P1(val) // ERROR "val does not escape$"
}

func f29(val Point1) {
	P1(&val)
}

func f30(ptr *Point1) { // ERROR "ptr does not escape$"
	P1(ptr)
}

func f31(val Point1) {
	arg := []Point1{val} // ERROR "\[\]Point1{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f32() {
	val := Point1{1, 2}
	P1(val) // ERROR "val does not escape$"
}

func f33() {
	val := Point1{1, 2}
	P1(&val)
}

func f34() {
	val := Point1{1, 2}
	arg := []Point1{val} // ERROR "\[\]Point1{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f35() {
	val := Point1{1, 2}
	arg := []*Point1{&val} // ERROR "\[\]\*Point1{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

func f36() {
	val := Point1{1, 2}
	s := []*Point1{&val}  // ERROR "\[\]\*Point1{...} does not escape$"
	arg := [][]*Point1{s} // ERROR "\[\]\[\]\*Point1{...} does not escape$"
	P1(arg)               // ERROR "arg does not escape$"
}

// Define a struct type without any methods that has pointer fields.

type Point2 struct{ x, y *int }

// Use P1 with Point2.
// (These tests also follow the pattern of using P1 with IntEscaping above).

func f37(val Point2) { // ERROR "val does not escape$"
	P1(val) // ERROR "val does not escape$"
}

func f38(val Point2) { // ERROR "val does not escape$"
	P1(&val)
}

func f39(ptr *Point2) { // ERROR "ptr does not escape$"
	P1(ptr)
}

func f40(val Point2) { // ERROR "val does not escape$"
	arg := []Point2{val} // ERROR "\[\]Point2{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f41() {
	val := Point2{}
	P1(val) // ERROR "val does not escape$"
}

func f42() {
	val := Point2{}
	P1(&val)
}

func f43() {
	val := Point2{}
	arg := []Point2{val} // ERROR "\[\]Point2{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f44() {
	val := Point2{}
	arg := []*Point2{&val} // ERROR "\[\]\*Point2{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

func f45() {
	val := Point2{}
	s := []*Point2{&val}  // ERROR "\[\]\*Point2{...} does not escape$"
	arg := [][]*Point2{s} // ERROR "\[\]\[\]\*Point2{...} does not escape$"
	P1(arg)               // ERROR "arg does not escape$"
}

// Use P1 with int.
// (These tests also follow the pattern of using P1 with IntEscaping above).

func f46(val int) {
	P1(val) // ERROR "val does not escape$"
}

func f47(val int) {
	P1(&val)
}

func f48(ptr *int) { // ERROR "ptr does not escape$"
	P1(ptr)
}

func f49(val int) {
	arg := []int{val} // ERROR "\[\]int{...} does not escape$"
	P1(arg)           // ERROR "arg does not escape$"
}

func f50() {
	// TODO: consider stop using 1000 in all our examples. (Escape analysis doesn't currently care,
	// but walk has optimizations to avoid allocations for readonly globals and some direct
	// use of constants and whatnot. Currently, using the local variable val I think defeats
	// those optimizations in this example and likely similar examples.
	// Maybe it doesn't matter if escape analysis doesn't care, but might be more future proof
	// to use a different pattern).
	val := 1000
	P1(val) // ERROR "val does not escape$"
}

func f51() {
	val := 1000
	P1(&val)
}

func f52() {
	val := 1000
	arg := []int{val} // ERROR "\[\]int{...} does not escape$"
	P1(arg)           // ERROR "arg does not escape$"
}

func f53() {
	val := 1000
	arg := []*int{&val} // ERROR "\[\]\*int{...} does not escape$"
	P1(arg)             // ERROR "arg does not escape$"
}

func f54() {
	val := 1000
	s := []*int{&val}  // ERROR "\[\]\*int{...} does not escape$"
	arg := [][]*int{s} // ERROR "\[\]\[\]\*int{...} does not escape$"
	P1(arg)            // ERROR "arg does not escape$"
}

// Use P1 with string.
// (These tests also follow the pattern of using P1 with IntEscaping above).

func f55(val string) { // ERROR "val does not escape$"
	P1(val) // ERROR "val does not escape$"
}

func f56(val string) { // ERROR "val does not escape$"
	P1(&val)
}

func f57(ptr *string) { // ERROR "ptr does not escape$"
	P1(ptr)
}

func f58(val string) { // ERROR "val does not escape$"
	arg := []string{val} // ERROR "\[\]string{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f59() {
	val := "abc"
	P1(val) // ERROR "val does not escape$"
}

func f60() {
	val := "abc"
	P1(&val)
}

func f61() {
	val := "abc"
	arg := []string{val} // ERROR "\[\]string{...} does not escape$"
	P1(arg)              // ERROR "arg does not escape$"
}

func f62() {
	val := "abc"
	arg := []*string{&val} // ERROR "\[\]\*string{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

func f63() {
	val := "abc"
	s := []*string{&val}  // ERROR "\[\]\*string{...} does not escape$"
	arg := [][]*string{s} // ERROR "\[\]\[\]\*string{...} does not escape$"
	P1(arg)               // ERROR "arg does not escape$"
}

// Define a named type without any methods using int as the underlying type.

type MyInt1 int

// Use P1 with MyInt1.
// (This is a shorter set of tests).

func f64() {
	val := MyInt1(1000)
	P1(val) // ERROR "val does not escape$"
}

func f65() {
	val := MyInt1(1000)
	P1(&val)
}

func f66() {
	val := MyInt1(1000)
	arg := []*MyInt1{&val} // ERROR "\[\]\*MyInt1{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

// Define a named type with a non-escaping String method with a non-pointer receier.

type MyInt2 int

func (i MyInt2) String() string { return "" }

// Use P1 with MyInt2.
// (These tests follow the pattern of using P1 with MyInt1 above).

func f67() {
	val := MyInt2(1000)
	P1(val) // ERROR "val does not escape$"
}

func f68() {
	val := MyInt2(1000)
	P1(&val)
}

func f69() {
	val := MyInt2(1000)
	arg := []*MyInt2{&val} // ERROR "\[\]\*MyInt2{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

// Define a named type with a non-escaping String method with a pointer receiver.

type MyInt3 int

func (i *MyInt3) String() string { return "" } // ERROR "i does not escape$"

// Use P1 with MyInt3.
// (These tests also follow the pattern of using P1 with MyInt1 above).

func f70() {
	val := MyInt3(1000)
	P1(val) // ERROR "val does not escape$"
}

func f71() {
	val := MyInt3(1000)
	P1(&val)
}

func f72() {
	val := MyInt3(1000)
	arg := []*MyInt3{&val} // ERROR "\[\]\*MyInt3{...} does not escape$"
	P1(arg)                // ERROR "arg does not escape$"
}

// P2 is a print-ish variation that takes a slice.

func P2(args []any) { // ERROR "might leak param content: args$"
	for _, a := range args {
		if v, ok := a.(Stringer); ok {
			_ = v.String()
		}
	}
}

// Use P2 with int.
// (We then repeat this pattern with another type with P2, then
// repeat this pattern with variations on the print-ish function).

func f73(val int) {
	arg := []any{val} // ERROR "\[\]any{...} does not escape$" "val does not escape$"
	P2(arg)
}

func f74(val int) {
	arg := []any{&val} // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f75(ptr *int) { // ERROR "ptr does not escape$"
	arg := []any{ptr} // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f76() {
	val := 1000
	arg := []any{val} // ERROR "\[\]any{...} does not escape$" "val does not escape$"
	P2(arg)
}

func f77() {
	val := 1000
	arg := []any{&val} // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f78() {
	val := 1000
	arg := []any{[]*int{&val}} // ERROR "\[\]\*int{...} does not escape$" "\[\]any{...} does not escape$"
	P2(arg)
}

// Use P2 with IntEscaping.
// (These tests follow the pattern of using P2 with int above).

func f79(val IntEscaping) {
	arg := []any{val} // ERROR "\[\]any{...} does not escape$" "val escapes to heap$"
	P2(arg)
}

func f80(val IntEscaping) { // ERROR "moved to heap: val$"
	arg := []any{&val} // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f81(ptr *IntEscaping) { // ERROR "leaking param: ptr$"
	arg := []any{ptr} // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f82() {
	val := IntEscaping(1000)
	arg := []any{val} // ERROR "\[\]any{...} does not escape$" "val escapes to heap$"
	P2(arg)
}

func f83() {
	val := IntEscaping(1000) // ERROR "moved to heap: val$"
	arg := []any{&val}       // ERROR "\[\]any{...} does not escape$"
	P2(arg)
}

func f84() {
	// TODO: we likely could avoid the []*IntEscaping escaping (the slice itself).
	val := IntEscaping(1000)           // ERROR "moved to heap: val$"
	arg := []any{[]*IntEscaping{&val}} // ERROR "\[\]\*IntEscaping{...} escapes to heap$" "\[\]any{...} does not escape$"
	P2(arg)
}

// P3 is a print-ish variation that takes a struct.

type inputArg struct {
	eface any
	a, b  int
}

func P3(a inputArg) { // ERROR "might leak param: a$"
	if v, ok := a.eface.(Stringer); ok {
		_ = v.String()
	}
}

// Use P3 with int.
// (These tests also follow the pattern of using P2 with int above).

func f85(val int) {
	arg := inputArg{eface: val} // ERROR "val does not escape$"
	P3(arg)
}

func f86(val int) {
	arg := inputArg{eface: &val}
	P3(arg)
}

func f87(ptr *int) { // ERROR "ptr does not escape$"
	arg := inputArg{eface: ptr}
	P3(arg)
}

func f88() {
	val := 1000
	arg := inputArg{eface: val} // ERROR "val does not escape$"
	P3(arg)
}

func f89() {
	val := 1000
	arg := inputArg{eface: &val}
	P3(arg)
}

func f90() {
	val := 1000
	arg := inputArg{eface: []*int{&val}} // ERROR "\[\]\*int{...} does not escape$"
	P3(arg)
}

// Use P3 with IntEscaping.
// (These tests also follow the pattern of using P2 with int above).

func f91(val IntEscaping) {
	arg := inputArg{eface: val} // ERROR "val escapes to heap$"
	P3(arg)
}

func f92(val IntEscaping) { // ERROR "moved to heap: val$"
	arg := inputArg{eface: &val}
	P3(arg)
}

func f93(ptr *IntEscaping) { // ERROR "leaking param: ptr$"
	arg := inputArg{eface: ptr}
	P3(arg)
}

func f94() {
	val := IntEscaping(1000)
	arg := inputArg{eface: val} // ERROR "val escapes to heap$"
	P3(arg)
}

func f95() {
	val := IntEscaping(1000) // ERROR "moved to heap: val$"
	arg := inputArg{eface: &val}
	P3(arg)
}

func f96() {
	// TODO: we likely could avoid the []*IntEscaping escaping (the slice itself).
	val := IntEscaping(1000)                     // ERROR "moved to heap: val$"
	arg := inputArg{eface: []*IntEscaping{&val}} // ERROR "\[\]\*IntEscaping{...} escapes to heap$"
	P3(arg)
}

// P4 is like P3, but takes a pointer to a struct.

func P4(a *inputArg) { // ERROR "might leak param content: a$"
	if v, ok := a.eface.(Stringer); ok {
		_ = v.String()
	}
}

// Use P4 with int.
// (These tests also follow the pattern of using P2 with int above).

func f97(val int) {
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val does not escape$"
	P4(arg)
}

func f98(val int) {
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f99(ptr *int) { // ERROR "ptr does not escape$"
	arg := &inputArg{eface: ptr} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f100() {
	val := 1000
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val does not escape$"
	P4(arg)
}

func f101() {
	val := 1000
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f102() {
	val := 1000
	arg := &inputArg{eface: []*int{&val}} // ERROR "&inputArg{...} does not escape$" "\[\]\*int{...} does not escape$"
	P4(arg)
}

// Use P4 with IntEscaping.
// (These tests also follow the pattern of using P2 with int above).

func f103(val IntEscaping) {
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val escapes to heap$"
	P4(arg)
}

func f104(val IntEscaping) { // ERROR "moved to heap: val$"
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f105(ptr *IntEscaping) { // ERROR "leaking param: ptr$"
	arg := &inputArg{eface: ptr} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f106() {
	val := IntEscaping(1000)
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val escapes to heap$"
	P4(arg)
}

func f107() {
	val := IntEscaping(1000)      // ERROR "moved to heap: val$"
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P4(arg)
}

func f108() {
	// TODO: we likely could avoid the []*IntEscaping escaping (the slice itself).
	val := IntEscaping(1000)                      // ERROR "moved to heap: val$"
	arg := &inputArg{eface: []*IntEscaping{&val}} // ERROR "&inputArg{...} does not escape$" "\[\]\*IntEscaping{...} escapes to heap$"
	P4(arg)
}

// P5 is like P4, but for contrast, it just leaks its argument's interface field.

func P5(a *inputArg) { // ERROR "leaking param content: a$"
	sink = a.eface
}

// Use P5 with int.
// (These tests also follow the pattern of using P2 with int above).

func f109(val int) {
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val escapes to heap$"
	P5(arg)
}

func f110(val int) { // ERROR "moved to heap: val$"
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P5(arg)
}

func f111(ptr *int) { // ERROR "leaking param: ptr$"
	arg := &inputArg{eface: ptr} // ERROR "&inputArg{...} does not escape$"
	P5(arg)
}

func f112() {
	val := 1000
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val escapes to heap$"
	P5(arg)
}

func f113() {
	val := 1000                   // ERROR "moved to heap: val$"
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P5(arg)
}

func f114() {
	val := 1000                           // ERROR "moved to heap: val$"
	arg := &inputArg{eface: []*int{&val}} // ERROR "&inputArg{...} does not escape$" "\[\]\*int{...} escapes to heap$"
	P5(arg)
}

// P6 is like P5, but for contrast, it is a noop.

func P6(a *inputArg) {} // ERROR "a does not escape$"

// Use P6 with int.
// (These tests also follow the pattern of using P2 with int above).

func f115(val int) {
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val does not escape$"
	P6(arg)
}

func f116(val int) {
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P6(arg)
}

func f117(ptr *int) { // ERROR "ptr does not escape$"
	arg := &inputArg{eface: ptr} // ERROR "&inputArg{...} does not escape$"
	P6(arg)
}

func f118() {
	val := 1000
	arg := &inputArg{eface: val} // ERROR "&inputArg{...} does not escape$" "val does not escape$"
	P6(arg)
}

func f119() {
	val := 1000
	arg := &inputArg{eface: &val} // ERROR "&inputArg{...} does not escape$"
	P6(arg)
}

func f120() {
	val := 1000
	arg := &inputArg{eface: []*int{&val}} // ERROR "&inputArg{...} does not escape$" "\[\]\*int{...} does not escape$"
	P6(arg)
}

// Use P1 with some types that either are understood that they must escape,
// or not understood by interface receiver aware escape analysis, and hence escape.
//
// For interface receiver escape analysis, we give up on:

// Recursive types.
type Node struct {
	next *Node
	a    int
}

func f121() {
	val := Node{}
	P1(val) // ERROR "val escapes to heap$"
}

// Types containing recursive types.
type Graph struct {
	node *Node
	a    int
}

func f122() {
	val := Graph{}
	P1(val) // ERROR "val escapes to heap$"
}

func f123() {
	node := Node{} // ERROR "moved to heap: node$"
	val := Graph{&node, 0}
	P1(val) // ERROR "val escapes to heap$"
}

// Channels and other non-basic types.
func f124() {
	var val chan int
	P1(val)
}

// Types containing channels and other non-basic types.
type C struct {
	c chan int
	a int
}

func f125() {
	val := C{}
	P1(val) // ERROR "val escapes to heap$"
}

type F struct {
	f func()
	a int
}

func f126() {
	val := F{}
	P1(val) // ERROR "val escapes to heap$"
}

// Map pointer content.
func f127() {
	val := 1000                    // ERROR "moved to heap: val$"
	arg := map[int]*int{val: &val} // ERROR "map\[int\]\*int{...} escapes to heap$"
	P1(arg)
}

func f128() {
	val := 1000                    // ERROR "moved to heap: val$"
	arg := map[*int]int{&val: val} // ERROR "map\[\*int\]int{...} escapes to heap$"
	P1(arg)
}

func f129() {
	val := Point1{1, 2}                  // ERROR "moved to heap: val$"
	arg := map[Point1]*Point1{val: &val} // ERROR "map\[Point1\]\*Point1{...} escapes to heap$"
	P1(arg)
}

func f130() {
	val := Point1{1, 2}                  // ERROR "moved to heap: val$"
	arg := map[*Point1]Point1{&val: val} // ERROR "map\[\*Point1\]Point1{...} escapes to heap$"
	P1(arg)
}

// Map value content does not escape.
func f131() {
	val := 1000
	arg := map[int]int{val: val} // ERROR "map\[int\]int{...} escapes to heap$"
	P1(arg)
}

// Some mutual recursion examples.
// Most of these would not make sense to run due to infinite recursion.
// TODO: clean up these older examples.

type FooRecurse1 struct {
	a, b *int
}

func (f FooRecurse1) String() string { // ERROR "leaking param: f$"
	r1(f)
	return ""
}

func r1(foo FooRecurse1) { // ERROR "leaking param: foo$"
	foo.String()
	P1(foo) // ERROR "foo escapes to heap$"
}

// Another mutual recursion example.
// Like prior example, but without pointer fields.

type FooRecurse2 struct {
	a, b int
}

func (f FooRecurse2) String() string {
	r2(f)
	return ""
}

func r2(foo FooRecurse2) {
	foo.String()
	P1(foo) // ERROR "foo does not escape$"
}

// Another mutual recursion example.
// Like first example, but with pointer receiver.

type FooRecurse3 struct {
	a, b *int
}

func (f *FooRecurse3) String() string { // ERROR "leaking param: f$"
	r3(f)
	return ""
}

func r3(foo *FooRecurse3) { // ERROR "leaking param: foo$"
	foo.String()
	P1(foo)
}

// Another mutual recursion example.

type FooRecurse4 struct {
	a, b *int
}

func (f FooRecurse4) String() string { return f.S() }      // ERROR "f does not escape$"
func (f FooRecurse4) S() string      { return f.String() } // ERROR "f does not escape$"

func r4() {
	foo := FooRecurse4{}
	// TODO: here, the methods on foo complete escape analysis before we start escape analysis
	// on this function, so we are able to use those results here to conclude foo does not
	// escape, which is probably too aggressive. (See related TODO in ifaceRecvPath in solve.go
	// regarding using escape analysis results for types in the same package).
	// Details visible via:
	//  go build -gcflags='-m=1 -l -d=escrecvdebug=3'
	P1(foo) // ERROR "foo does not escape$"
}

// Another mutual recursion example.
// Like prior example, but with pointer receiver for S.

type FooRecurse5 struct {
	a, b *int
}

func (f FooRecurse5) String() string { return f.S() }      // ERROR "f does not escape$"
func (f *FooRecurse5) S() string     { return f.String() } // ERROR "f does not escape$"

func r5() {
	foo := FooRecurse5{}
	// TODO: see prior TODO for FooRecurse4.
	P1(foo) // ERROR "foo does not escape$"
}

// Another mutual recursion example.
// Like prior example, but with pointer receiver for String and S.

type FooRecurse6 struct {
	a, b *int
}

func (f *FooRecurse6) String() string { return f.S() }      // ERROR "f does not escape$"
func (f *FooRecurse6) S() string      { return f.String() } // ERROR "f does not escape$"

func r6() {
	foo := FooRecurse6{}
	// TODO: see prior TODO for FooRecurse4.
	P1(foo) // ERROR "foo does not escape$"
}

// TODO: delete these TODOs listing variations of tests.

// TODO: maybe step through more examples based on what triggers no escape?
//   [done] simple values like int with no methods (no escape)
//   [done] struct with no methods (no escape)
//   [done] struct with methods but all scalar fields (no escape)
//   [done] struct with methods with pointer fields but no methods (no escape)
//   [done] struct with methods with all scalar fields and methods escape (escapes)
//   [done] struct with methods with pointer fields and no methods escape (no escape)
//   ...

// TODO: some more leaking param content examples? (e.g., see escape2.go, escape_iface_recv_extracted.go)

// TODO: mark up which of these variations we've done (and later delete):

// A pointer receiver:
//   is leaked to the heap .
//   is leaked to a return value.
//   is leaked to another parameter.
//   has a pointer field set to a local pointer.
//   has an interface field set to a local pointer.

// A pointer receiver has pointer content that:
//   is leaked to the heap.
//   is leaked to a return value.
//   is leaked to another parameter.
//   has a pointer field set to a local pointer.
//   has an interface field set to a local pointer.

// A non-pointer receiver has pointer content that:
//   is leaked to the heap.
//   is leaked to a return value.
//   is leaked to another parameter.
//   has a pointer field set to a local pointer.
//   has an interface field set to a local pointer.

// Interface methods with pointer arguments (some with escaping non-receiver args, some not).
