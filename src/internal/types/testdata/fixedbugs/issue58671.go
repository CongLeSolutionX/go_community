// -inferMaxDefaultType

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func g[P any](...P) P { var x P; return x }

func _() {
	var _ float64 = g(1, 2.3)
	var _ rune = g(1, 'a')
	var _ complex128 = g(1, 2.3, 1i)
	// g(1, 'a')     // invalid
	// g(1, "foo")   // invalid
}
