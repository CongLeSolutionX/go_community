// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compare_test

import "testing"

type StructMem3 struct {
	A uint8
	B uint8
	C uint8
}

func BenchmarkCompareStructMem3(b *testing.B) {
	var a StructMem3
	var c StructMem3

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

type StructMem5 struct {
	A uint8
	B uint8
	C uint8
	D uint16
}

func BenchmarkCompareStructMem5(b *testing.B) {
	var a StructMem5
	var c StructMem5

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}
