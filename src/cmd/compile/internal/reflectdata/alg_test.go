// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectdata_test

import "testing"

func BenchmarkEqArrayOfStrings5(b *testing.B) {
	var a [5]string
	var c [5]string

	for i := 0; i < 5; i++ {
		a[i] = "aaaa"
		c[i] = "cccc"
	}

	for j := 0; j < b.N; j++ {
		_ = a == c
	}
}

func BenchmarkEqArrayOfStrings64(b *testing.B) {
	var a [64]string
	var c [64]string

	for i := 0; i < 64; i++ {
		a[i] = "aaaa"
		c[i] = "cccc"
	}

	for j := 0; j < b.N; j++ {
		_ = a == c
	}
}

func BenchmarkEqArrayOfStrings1024(b *testing.B) {
	var a [1024]string
	var c [1024]string

	for i := 0; i < 1024; i++ {
		a[i] = "aaaa"
		c[i] = "cccc"
	}

	for j := 0; j < b.N; j++ {
		_ = a == c
	}
}

func BenchmarkEqArrayOfFloats5(b *testing.B) {
	var a [5]float32
	var c [5]float32

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

func BenchmarkEqArrayOfFloats64(b *testing.B) {
	var a [64]float32
	var c [64]float32

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

func BenchmarkEqArrayOfFloats1024(b *testing.B) {
	var a [1024]float32
	var c [1024]float32

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

type StructMem7 struct {
	A uint16
	B uint16
	C uint16
	D uint8
	E float32
}

func BenchmarkEqStructMem7(b *testing.B) {
	var a StructMem7
	var c StructMem7

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

type StructMem8 struct {
	A uint16
	B uint16
	C uint16
	D uint16
	E float32
}

func BenchmarkEqStructMem8(b *testing.B) {
	var a StructMem8
	var c StructMem8

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

type StructMem16 struct {
	B uint32
	C uint32
	D uint32
	E uint32
	F float32
}

func BenchmarkEqStructMem16(b *testing.B) {
	var a StructMem16
	var c StructMem16

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}
