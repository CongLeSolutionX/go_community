// run -gcflags=-G=4

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"unsafe"
)

func main() {
	type x[U any] int

	fmt.Printf("%T\n", F[float64]())

	// BAD: The output for this should have `main.x·1`, not `"".x·1`.
	// See reader.mangle.
	fmt.Printf("%T\n", F[x[float64]]())

	println(F[float64]() == F[float64]())
	println(F[float64]() == F[float32]())

	var xs []interface{}
	xs = append(xs, fn()...)
	xs = append(xs, G[float32]()...)
	xs = append(xs, G[float64]()...)
	xs = append(xs, G[float64]()...)

	for i, x := range xs {
		fmt.Printf("%d: %T\t", i, x)
		for _, y := range xs {
			ch := '.'
			if x == y {
				ch = 'X'
			}
			fmt.Printf(" %c", ch)
		}
		fmt.Println()
	}
}

var fn func() []interface{} = G[float32]

func F[T any]() interface{} {
	_ = func() {
	}

	type x[U any] []struct {
		t T
		u U

		x [unsafe.Sizeof(func() (_ [0]T) {
			type uwu[T any] int
			var _ uwu[uwu[T]]
			var _ uwu[uwu[U]]
			return
		}())]int
	}

	var _ x[int]

	type t[U any] struct{}

	return t[int]{}
}

func G[T any]() []interface{} {

foo:
	for range [0]int{} {
		break foo
	}

	type x[U any] struct{}

	foo := func() interface{} {
		return x[int32]{}
	}

	bar := func() interface{} {
		type y = x[int32]
		return y{}
	}

	return []interface{}{foo(), bar(), x[uint32]{}}
}
