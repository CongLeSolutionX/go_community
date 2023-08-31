// errorcheck -0 -m -l

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test basic escape analysis.
//
// We start with some simple examples, which we also use as examples in an
// intro comment in cmd/internal/escape/escape.go.
//
// We then show some variations on simple pointer and interface examples that
// mostly can be compared to each other. For example, ptr3Local can be
// compared to eface3Local.

package example

// Intro examples.
//
// If these first ~4 tests starts to fail due to some improvement in output or behavior,
// consider updating the comments in escape.go, though probably also fine
// for those comments to go a bit stale.

func f1(inptr *int) (outptr *int) { // ERROR "leaking param: inptr to result outptr level=0$"
	localptr := inptr
	return localptr
}

func f2() *Foo {
	foo := Foo{} // ERROR "moved to heap: foo$"
	return &foo
}

func f3(inptr *int) *Foo { // ERROR "leaking param: inptr$"
	foo := Foo{} // ERROR "moved to heap: foo$"
	foo.ptr = inptr
	return &foo
}

func f4(inptr *int) { // ERROR "inptr does not escape$"
	var in1 int
	var in3 int // ERROR "moved to heap: in3$"
	_ = f1(&in1)
	_ = f3(&in3)
}

type Foo struct {
	ptr   *int
	eface any
}

// Pointers.
//
// Now we have some basic examples involving pointers,
// followed by very similar examples involving interfaces.

var (
	sinkIntPtr *int
	sinkFoo    Foo
	sinkFooPtr *Foo
)

// Some variations of storing locals in globals involving pointers.

func ptr1Local() {
	val := 1000 // ERROR "moved to heap: val$"
	sinkIntPtr = &val
}

func ptr2Local() {
	foo := Foo{}
	sinkFoo = foo
}

func ptr3Local() {
	foo := Foo{} // ERROR "moved to heap: foo$"
	sinkFooPtr = &foo
}

func ptr4Local() {
	foo := Foo{}
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	sinkFoo = foo
}

func ptr5Local() {
	foo := Foo{} // ERROR "moved to heap: foo$"
	val := 1000  // ERROR "moved to heap: val$"
	foo.ptr = &val
	sinkFooPtr = &foo
}

// Some variations of storing parameters in globals involving pointers.

func ptr1Param(intptr *int) { // ERROR "leaking param: intptr$"
	sinkIntPtr = intptr
}

func ptr2Param(foo Foo) { // ERROR "leaking param: foo$"
	sinkFoo = foo
}

func ptr3Param(foo Foo) { // ERROR "moved to heap: foo$"
	sinkFooPtr = &foo
}

func ptr4aParam(foo Foo) { // ERROR "leaking param: foo$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	sinkFoo = foo
}

func ptr4bParam(foo Foo, intptr *int) { // ERROR "leaking param: foo$" "leaking param: intptr$"
	foo.ptr = intptr
	sinkFoo = foo
}

func ptr5aParam(foo Foo) { // ERROR "moved to heap: foo$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	sinkFooPtr = &foo
}

func ptr5bParam(foo Foo, intptr *int) { // ERROR "leaking param: intptr$" "moved to heap: foo$"
	foo.ptr = intptr
	sinkFooPtr = &foo
}

func ptr6Param(fooptr *Foo) { // ERROR "leaking param: fooptr$"
	sinkFooPtr = fooptr
}

func ptr7aParam(fooptr *Foo) { // ERROR "leaking param: fooptr$"
	val := 1000 // ERROR "moved to heap: val$"
	fooptr.ptr = &val
	sinkFooPtr = fooptr
}

func ptr7bParam(fooptr *Foo, intptr *int) { // ERROR "leaking param: fooptr$" "leaking param: intptr$"
	fooptr.ptr = intptr
	sinkFooPtr = fooptr
}

// Some variations of returning locals involving pointers.

func ptr1Return() *int {
	val := 1000 // ERROR "moved to heap: val$"
	return &val
}

func ptr2Return() Foo {
	foo := Foo{}
	return foo
}

func ptr3Return() *Foo {
	foo := Foo{} // ERROR "moved to heap: foo$"
	return &foo
}

func ptr4Return() Foo {
	foo := Foo{}
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return foo
}

func ptr5Return() *Foo {
	foo := Foo{} // ERROR "moved to heap: foo$"
	val := 1000  // ERROR "moved to heap: val$"
	foo.ptr = &val
	return &foo
}

// Some variations of returning parameters involving pointers.

func ptr1ReturnParam(intptr *int) *int { // ERROR "leaking param: intptr to result ~r0 level=0$"
	return intptr
}

func ptr2ReturnParam(foo Foo) Foo { // ERROR "leaking param: foo to result ~r0 level=0$"
	return foo
}

func ptr3ReturnParam(fooptr *Foo) *Foo { // ERROR "leaking param: fooptr to result ~r0 level=0$"
	return fooptr
}

func ptr4aReturnParam(foo Foo) Foo { // ERROR "leaking param: foo to result ~r0 level=0$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return foo
}

func ptr4bReturnParam(foo Foo, intptr *int) Foo { // ERROR "leaking param: foo to result ~r0 level=0$" "leaking param: intptr to result ~r0 level=0$"
	foo.ptr = intptr
	return foo
}

func ptr5aReturnParam(foo Foo) *Foo { // ERROR "moved to heap: foo$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return &foo
}

func ptr5bReturnParam(foo Foo, intptr *int) *Foo { // ERROR "leaking param: intptr$" "moved to heap: foo$"
	foo.ptr = intptr
	return &foo
}

// Interfaces.
//
// Now we have some basic examples involving interfaces that are
// generally similar to the preceding pointer-based examples.

var sinkEface any

// Some variations of storing locals in globals involving interfaces.

func eface1aLocal() {
	val := 1000
	sinkEface = val // ERROR "val escapes to heap$"
}

func eface1bLocal() {
	val := 1000
	var eface any = val // ERROR "val escapes to heap$"
	sinkEface = eface
}

func eface2Local() {
	foo := Foo{}
	sinkFoo = foo // this is a boring example, but leaving for contrast
}

func eface3Local() {
	foo := Foo{}
	sinkEface = foo // ERROR "foo escapes to heap$"
}

func eface4Local() {
	foo := Foo{}
	val := 1000
	foo.eface = val // ERROR "val escapes to heap$"
	sinkFoo = foo
}

func eface5Local() {
	foo := Foo{}
	val := 1000
	foo.eface = val // ERROR "val escapes to heap$"
	sinkEface = foo // ERROR "foo escapes to heap$"
}

// Some variations of storing parameters in globals involving interfaces.

func eface1aParam(in int) {
	sinkEface = in // ERROR "in escapes to heap$"
}

func eface1bParam(in any) { // ERROR "leaking param: in$"
	sinkEface = in
}

func eface1cParam(intptr *int) { // ERROR "leaking param: intptr$"
	sinkEface = intptr
}

func eface2Param(foo Foo) { // ERROR "leaking param: foo$"
	sinkEface = foo // ERROR "foo escapes to heap$"
}

func eface3Param(foo Foo) { // ERROR "moved to heap: foo$"
	sinkEface = &foo
}

func eface4aParam(foo Foo) { // ERROR "leaking param: foo$"
	val := 1000
	foo.eface = val // ERROR "val escapes to heap$"
	sinkFoo = foo
}

func eface4bParam(foo Foo, intptr *int) { // ERROR "leaking param: foo$" "leaking param: intptr$"
	foo.eface = intptr
	sinkFoo = foo
}

func eface4cParam(foo Foo, in any) { // ERROR "leaking param: foo$" "leaking param: in$"
	foo.eface = in
	sinkFoo = foo
}

func eface5aParam(foo Foo) { // ERROR "moved to heap: foo$"
	val := 1000
	foo.eface = val // ERROR "val escapes to heap$"
	sinkFooPtr = &foo
}

func eface5bParam(foo Foo, intptr *int) { // ERROR "leaking param: intptr$" "moved to heap: foo$"
	foo.eface = intptr
	sinkFooPtr = &foo
}

func eface5cParam(foo Foo, in any) { // ERROR "leaking param: in$" "moved to heap: foo$"
	foo.eface = in
	sinkFooPtr = &foo
}

func eface6Param(fooptr *Foo) { // ERROR "leaking param: fooptr$"
	sinkEface = fooptr
}

func eface7aParam(fooptr *Foo) { // ERROR "leaking param: fooptr$"
	val := 1000
	fooptr.eface = val // ERROR "val escapes to heap$"
	sinkFooPtr = fooptr
}

func eface7bParam(fooptr *Foo, intptr *int) { // ERROR "leaking param: fooptr$" "leaking param: intptr$"
	fooptr.eface = intptr
	sinkFooPtr = fooptr
}

func eface7cParam(fooptr *Foo, in any) { // ERROR "leaking param: fooptr$" "leaking param: in$"
	fooptr.eface = in
	sinkFooPtr = fooptr
}

// Some variations of returning locals involving interfaces.

func eface1Return() any {
	val := 1000 // ERROR "moved to heap: val$"
	return &val
}

func eface2Return() any {
	foo := Foo{}
	return foo // ERROR "foo escapes to heap$"
}

func eface3Return() any {
	foo := Foo{} // ERROR "moved to heap: foo$"
	return &foo
}

func eface4Return() any {
	foo := Foo{}
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return foo // ERROR "foo escapes to heap$"
}

func eface5Return() any {
	foo := Foo{} // ERROR "moved to heap: foo$"
	val := 1000  // ERROR "moved to heap: val$"
	foo.ptr = &val
	return &foo
}

// Some variations of returning parameters involving interfaces.

func eface1aReturnParam(intptr *int) any { // ERROR "leaking param: intptr to result ~r0 level=0$"
	return intptr
}

func eface1bReturnParam(in any) any { // ERROR "leaking param: in to result ~r0 level=0$"
	return in
}

func eface2ReturnParam(foo Foo) any { // ERROR "leaking param: foo$"
	return foo // ERROR "foo escapes to heap$"
}

func eface3ReturnParam(fooptr *Foo) any { // ERROR "leaking param: fooptr to result ~r0 level=0$"
	return fooptr
}

func eface4aReturnParam(foo Foo) any { // ERROR "leaking param: foo$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return foo // ERROR "foo escapes to heap$"
}

func eface4bReturnParam(foo Foo, intptr *int) any { // ERROR "leaking param: foo$" "leaking param: intptr$"
	foo.ptr = intptr
	return foo // ERROR "foo escapes to heap$"
}

func eface5aReturnParam(foo Foo) any { // ERROR "moved to heap: foo$"
	val := 1000 // ERROR "moved to heap: val$"
	foo.ptr = &val
	return &foo
}

func eface5bReturnParam(foo Foo, intptr *int) any { // ERROR "leaking param: intptr$" "moved to heap: foo$"
	foo.ptr = intptr
	return &foo
}
