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

func f1()                                    {}
func f2(func())                              {}
func f3(func()) bool                         { return false }
func f4(func(int) bool) bool                 { return false }
func f5(func(int, string) bool) bool         { return false }
func f6(func(int) bool)                      {}
func f7(func(int) MyBool) MyBool             { return false }
func f8(func(MyInt, MyString) MyBool) MyBool { return false }

func test() {
	for range T.M { // ERROR "cannot range over T.M \(value of type func\(T\)\): func must have one func argument"
	}
	for range (*T).PM { // ERROR "cannot range over \(\*T\).PM \(value of type func\(\*T\)\): func must have one func argument"
	}
	for range f1 { // ERROR "cannot range over f1 \(value of type func\(\)\): func must have one func argument"
	}
	for range f2 { // ERROR "cannot range over f2 \(value of type func\(func\(\)\)\): func does not return bool"
	}
	for range f3 { // ERROR "cannot range over f3 \(value of type func\(func\(\)\) bool\): callback does not return bool"
	}
	for range f4 {
	}
	for range f5 {
	}
	for range f6 { // ERROR "cannot range over f6 \(value of type func\(func\(int\) bool\)\): func does not return bool"
	}
	for range f7 {
	}
	for range f8 {
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
	for i := range f5 {
		_ = i
	}
	for i = range f5 {
		_ = i
	}
	for i := range f7 {
		_ = i
	}
	for i = range f7 {
		_ = i
	}
	for mi := range f8 {
		_ = mi
	}
	for mi = range f8 {
		_ = mi
	}
	for mi, ms := range f8 {
		_, _ = mi, ms
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
