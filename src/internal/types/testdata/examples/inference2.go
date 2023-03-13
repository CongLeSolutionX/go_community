// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file shows some examples of "reverse" type inference
// where the type arguments for generic functions are determined
// from assigning the functions.

package p

func f1[P any](P)      {}
func f2[P any]() P     { var x P; return x }
func f3[P, Q any](P) Q { var x Q; return x }
func f4[P any](P, P)   {}
func f5[P any](P) []P  { return nil }

var (
	v1           = f1 // ERROR "cannot use generic function f1 without instantiation"
	v2 func(int) = f2 // ERROR "cannot infer P"

	v3 func(int)     = f1
	v4 func() int    = f2
	v5 func(int) int = f3
	_  func(int) int = f3[int]

	v6 func(int, int)     = f4
	v7 func(int, string)  = f4 // ERROR "type string of 2nd parameter does not match inferred type int for P"
	v8 func(int) []int    = f5
	v9 func(string) []int = f5 // ERROR "type []int of 1st result parameter does not match inferred type []string for []P"

	_, _ func(int) = f1, f1
	_, _ func(int) = f1, f2 // ERROR "cannot infer P"
)

func _() {
	v1 = f1 // ERROR "cannot use generic function f1 without instantiation"
	v2 = f2 // ERROR "cannot infer P"

	v3 = f1
	v4 = f2
	v5 = f3
	v5 = f3[int]

	v6 = f4
	v7 = f4 // ERROR "type string of 2nd parameter does not match inferred type int for P"
	v8 = f5
	v9 = f5 // ERROR "type []int of 1st result parameter does not match inferred type []string for []P"
}
