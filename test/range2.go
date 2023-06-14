// errorcheck -goexperiment range

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type MyInt int32
type MyBool bool
type MyString string

type T struct{}

func (*T) PM() {}
func (T) M()   {}

func f1()                             {}
func f2(func())                       {}
func f4(func(int) bool)               {}
func f5(func(int, string) bool)       {}
func f7(func(int) MyBool)             {}
func f8(func(MyInt, MyString) MyBool) {}

func test() {
	for range T.M { // ERROR "cannot range over T.M \(value of type func\(T\)\): func must be func\(yield func\(...\)bool\)"
	}
	for range (*T).PM { // ERROR "cannot range over \(\*T\).PM \(value of type func\(\*T\)\): func must be func\(yield func\(...\)bool\)"
	}
	for range f1 { // ERROR "cannot range over f1 \(value of type func\(\)\): func must be func\(yield func\(...\)bool\)"
	}
	for range f2 { // ERROR "cannot range over f2 \(value of type func\(func\(\)\)\): func must be func\(yield func\(...\)bool\): yield func does not return bool"
	}
	for range f4 { // ERROR "range over f4 \(value of type func\(func\(int\) bool\)\) must have one iteration variable"
	}
	for _ = range f4 {
	}
	for _, _ = range f5 {
	}
	for _ = range f7 {
	}
	for _, _ = range f8 {
	}
	for range 1 {
	}
	for range uint8(1) {
	}
	for range int64(1) {
	}
	for range MyInt(1) {
	}
	for range 'x' {
	}
	for range 1.0 { // ERROR "cannot range over 1.0 \(untyped float constant 1\)"
	}

	var i int
	var s string
	var mi MyInt
	var ms MyString
	for i := range f4 {
		_ = i
	}
	for i = range f4 {
		_ = i
	}
	for i, s := range f5 {
		_, _ = i, s
	}
	for i, s = range f5 {
		_, _ = i, s
	}
	for i, _ := range f5 {
		_ = i
	}
	for i, _ = range f5 {
		_ = i
	}
	for i := range f7 {
		_ = i
	}
	for i = range f7 {
		_ = i
	}
	for mi, _ := range f8 {
		_ = mi
	}
	for mi, _ = range f8 {
		_ = mi
	}
	for mi, ms := range f8 {
		_, _ = mi, ms
	}
	for i, s = range f8 { // ERROR "cannot use i \(value of type MyInt\) as int value in assignment" "cannot use s \(value of type MyString\) as string value in assignment"
		_, _ = mi, ms
	}
	for mi, ms := range f8 {
		i, s = mi, ms // ERROR "cannot use mi \(variable of type MyInt\) as int value in assignment" "cannot use ms \(variable of type MyString\) as string value in assignment"
	}
	for mi, ms = range f8 {
		_, _ = mi, ms
	}

	for i := range 10 {
		_ = i
	}
	for i = range 10 {
		_ = i
	}
	for i, j := range 10 { // ERROR "range over 10 \(untyped int constant\) permits only one iteration variable"
		_, _ = i, j
	}
	for mi := range MyInt(10) {
		_ = mi
	}
	for mi = range MyInt(10) {
		_ = mi
	}
}
