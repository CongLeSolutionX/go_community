// run -gcflags=all=-d=unified=2

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"reflect"
)

type test struct {
	TArgs    [2]reflect.Type
	Instance reflect.Type
}

var tests []test

type intish interface{ ~int }

type Int int
type GlobalInt = Int // allow access to global Int, even when shadowed

func F[A intish]() {
	add := func(B, T interface{}) {
		tests = append(tests, test{
			TArgs: [2]reflect.Type{
				reflect.TypeOf(A(0)),
				reflect.TypeOf(B),
			},
			Instance: reflect.TypeOf(T),
		})
	}

	type Int int

	type T[B intish] struct{}

	add(int(0), T[int]{})
	add(Int(0), T[Int]{})
	add(GlobalInt(0), T[GlobalInt]{})
	add(A(0), T[A]{}) // NOTE: intentionally dups with int and GlobalInt

	type U[_ any] int
	type V U[int]
	type W V

	add(U[int](0), T[U[int]]{})
	add(U[Int](0), T[U[Int]]{})
	add(U[GlobalInt](0), T[U[GlobalInt]]{})
	add(U[A](0), T[U[A]]{}) // NOTE: intentionally dups with U[int] and U[GlobalInt]
	add(V(0), T[V]{})
	add(W(0), T[W]{})
}

func main() {
	type Int int

	F[int]()
	F[Int]()
	F[GlobalInt]()

	type U[_ any] int
	type V U[int]
	type W V

	F[U[int]]()
	F[U[Int]]()
	F[U[GlobalInt]]()
	F[V]()
	F[W]()

	type X[A any] U[X[A]]

	F[X[int]]()
	F[X[Int]]()
	F[X[GlobalInt]]()

	for j, tj := range tests {
		for i, ti := range tests[:j+1] {
			if (ti.TArgs == tj.TArgs) != (ti.Instance == tj.Instance) {
				fmt.Printf("FAIL: %d,%d: %s, but %s\n", i, j, eq(ti.TArgs, tj.TArgs), eq(ti.Instance, tj.Instance))
			}

			// The test is constructed so we should see a few identical types.
			// See "NOTE" comments above.
			if i != j && ti.Instance == tj.Instance {
				fmt.Printf("%d,%d: %v\n", i, j, ti.Instance)
			}
		}
	}
}

func eq(a, b interface{}) string {
	op := "=="
	if a != b {
		op = "!="
	}
	return fmt.Sprintf("%v %s %v", a, op, b)
}
