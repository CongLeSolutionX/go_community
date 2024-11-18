// -gotypesalias=1

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T[P any] struct{}

// A1
type A1[P any] = T[P]
type B1[P any] = *T[P]

func (A1 /* ERROR "cannot define new methods on generic alias type A1[P any]" */ [P]) m() {}
func (B1 /* ERROR "cannot define new methods on generic alias type B1[P any]" */ [P]) m() {}

// A2
type A2[P any] = T[int]
type B2[P any] = *T[int]

func (A2 /* ERROR "cannot define new methods on generic alias type A2[P any]" */ [P]) m() {}
func (B2 /* ERROR "cannot define new methods on generic alias type B2[P any]" */ [P]) m() {}

// A3
type A3 = T[int]
type B3 = *T[int]

func (A3 /* ERROR "cannot define new methods on instantiated type A3" */) m1()     {} // base type is A3 (== T[int])
func (B3 /* ERROR "cannot define new methods on instantiated type T[int]" */) m2() {} // base type is T[int] (!= B3)

// A4
type A4 = T  // ERROR "cannot use generic type T[P any] without instantiation"
type B4 = *T // ERROR "cannot use generic type T[P any] without instantiation"

func (A4[P]) m1() {} // don't report a follow-on error on A4
func (B4[P]) m2() {} // don't report a follow-on error on B4
