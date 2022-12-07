// -lang=go1.17

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parser accepts type parameters but the type checker
// needs to report any operations that are not permitted
// before Go 1.18.

package p

type T[P /* ERR type parameter requires go1.18 or later */ any /* ERR predeclared any requires go1.18 or later */] struct{}

// for init (and main, but we're not in package main) we should only get one error
func init[P /* ERR func init must have no type parameters */ any /* ERR predeclared any requires go1.18 or later */]() {
}
func main[P /* ERR type parameter requires go1.18 or later */ any /* ERR predeclared any requires go1.18 or later */]() {
}

func f[P /* ERR type parameter requires go1.18 or later */ any /* ERR predeclared any requires go1.18 or later */](x P) {
	var _ T[ /* ERR type instantiation requires go1.18 or later */ int]
	var _ (T[ /* ERR type instantiation requires go1.18 or later */ int])
	_ = T[ /* ERR type instantiation requires go1.18 or later */ int]{}
	_ = T[ /* ERR type instantiation requires go1.18 or later */ int](struct{}{})
}

func (T[ /* ERR type instantiation requires go1.18 or later */ P]) g(x int) {
	f[ /* ERR function instantiation requires go1.18 or later */ int](0)     // explicit instantiation
	(f[ /* ERR function instantiation requires go1.18 or later */ int])(0)   // parentheses (different code path)
	f( /* ERR implicit function instantiation requires go1.18 or later */ x) // implicit instantiation
}

type C1 interface {
	comparable // ERR predeclared comparable requires go1.18 or later
}

type C2 interface {
	comparable // ERR predeclared comparable requires go1.18 or later
	int        // ERR embedding non-interface type int requires go1.18 or later
	~ /* ERR embedding interface element ~int requires go1.18 or later */ int
	int /* ERR embedding interface element int | ~string requires go1.18 or later */ | ~string
}

type _ interface {
	// errors for these were reported with their declaration
	C1
	C2
}

type (
	_ comparable // ERR predeclared comparable requires go1.18 or later
	// errors for these were reported with their declaration
	_ C1
	_ C2

	_ = comparable // ERR predeclared comparable requires go1.18 or later
	// errors for these were reported with their declaration
	_ = C1
	_ = C2
)
