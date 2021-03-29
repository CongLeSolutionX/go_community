// compile -G=3

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains regression tests for bugs found.

package p

// Assignability of an unnamed pointer type to a type parameter that
// has a matching underlying type.
func _[T interface{}, PT interface{type *T}] (x T) PT {
    return &x
}

// Indexing of generic types containing type parameters in their type list:
func at[T interface{ type []E }, E interface{}](x T, i int) E {
        return x[i]
}

// A generic type inside a function acts like a named type. Its underlying
// type is itself, its "operational type" is defined by the type list in
// the tybe bound, if any.
func _[T interface{type int}](x T) {
	type myint int
	var _ int = int(x)
	var _ T = 42
	var _ T = T(myint(42))
}

// Indexing a generic type with an array type bound checks length.
// (Example by mdempsky@.)
func _[T interface { type [10]int }](x T) {
	_ = x[9] // ok
}

// Pointer indirection of a generic type.
func _[T interface{ type *int }](p T) int {
	return *p
}

// Channel sends and receives on generic types.
func _[T interface{ type chan int }](ch T) int {
	ch <- 0
	return <- ch
}

// Calling of a generic variable.
func _[T interface{ type func() }](f T) {
	f()
	go f()
}
