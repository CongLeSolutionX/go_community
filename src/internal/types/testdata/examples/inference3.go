// -EnableInterfaceInference

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file shows examples of type inference across interfaces.

package p

// Examples where type parameters used in methods are used in type inference.
type I1 interface {
	m1(int)
	m2()
}

type S1 struct{}

func (S1) m1(int)

type S1p[P any] struct{}

func (S1p[P]) m1(P) {}

type I1p[T any] interface {
	m1(T)
}

func g1[P any](I1p[P]) {}

func g2[P ~int | ~string](I1p[P]) {}

func _() {
	// Type I1 of x implements I1p[int].
	var x I1
	g1[int](x) // we can pass the type argument explicitly
	g1(x)      // or we can use inference to infer the type argument from the method m1

	// Type S1 of s implements I1p[string]
	var s S1
	g1[int](s) // we can pass the type argument explicitly
	g1(s)      // or we can use inference to infer the type argument from the method m1

	// And for different instantiated types S1p
	g1[int](S1p[int]{})
	g1(S1p[int]{})
	g1(S1p[string]{})
	g1(S1p[any]{})

	// A type parameter may be inferred correctly, but instantiation may fail.
	g2 /* ERROR "float64 does not satisfy ~int | ~string" */ (S1p[float64]{})

	// A type parameter may be inferred correctly, but assignment may fail.
	var _ func(S1) = g1 /* ERROR "cannot use g1 (value of type func(I1p[int])) as func(S1) value in variable declaration" */
}

// If a type parameter is not used, the interface types must originate
// in the same declaration and type argument lists must match for inference
// to work.
type Tu1[_ any] any
type Tu2[_ any] any
type Tu3[_, _ any] any

func f[P any](Tu1[P]) {}

func _() {
	var x1 Tu1[string]
	f(x1)

	// In this case, even though Tu2[int] and Tu1[int] are assignment-compatible,
	// the type parameter for Tu1 cannot be inferred because it is not used anywhere.
	var x2 Tu2[int]
	f /* ERROR "cannot infer P" */ (x2)

	// In this case the type parameter lists don't match.
	// This is no different from the Tu2 case.
	var x3 Tu3[int, string]
	f /* ERROR "cannot infer P" */ (x3)

	// The Tu2 case is no different from the any case.
	var a any
	f /* ERROR "cannot infer P" */ (a)
}
