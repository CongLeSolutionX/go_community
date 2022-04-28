// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compare_test

import "testing"

type StructNested struct {
	A string
}

type StructTop struct {
	A StructNested
	B StructNested
	C StructNested
	D StructNested
}

func BenchmarkCompareStructTop(b *testing.B) {
	var nested StructNested
	nested.A = "aaaa"

	var a StructTop
	var c StructTop

	a.A = nested
	a.B = nested
	a.C = nested
	a.D = nested
	c = a

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}

func BenchmarkCompareStructTopLeafDiffers(b *testing.B) {
	var nested StructNested
	nested.A = "aaaa"

	var a StructTop
	var c StructTop

	a.A = nested
	a.B = nested
	a.C = nested
	a.D = nested
	c = a
	c.D.A = "a"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = a == c
	}
}
