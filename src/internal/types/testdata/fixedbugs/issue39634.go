// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Examples adjusted to match new [T any] syntax for type parameters.
// Also, previously permitted empty type parameter lists and instantiations
// are now syntax errors.

package p

// crash 1
type nt1[_ any]interface{g /* ERR undefined */ }
type ph1[e nt1[e],g(d /* ERR undefined */ )]s /* ERR undefined */
func(*ph1[e,e /* ERR redeclared */ ])h(d /* ERR undefined */ )

// crash 2
// Disabled: empty []'s are now syntax errors. This example leads to too many follow-on errors.
// type Numeric2 interface{t2 /* ERR not a type */ }
// func t2[T Numeric2](s[]T){0 /* ERR not a type */ []{s /* ERR cannot index */ [0][0]}}

// crash 3
type t3 *interface{ t3.p /* ERR no field or method p */ }

// crash 4
type Numeric4 interface{t4 /* ERR not a type */ }
func t4[T Numeric4](s[]T){if( /* ERR non-boolean */ 0){*s /* ERR cannot indirect */ [0]}}

// crash 7
type foo7 interface { bar() }
type x7[A any] struct{ foo7 }
func main7() { var _ foo7 = x7[int]{} }

// crash 8
type foo8[A any] interface { ~A /* ERR cannot be a type parameter */ }
func bar8[A foo8[A]](a A) {}

// crash 9
type foo9[A any] interface { foo9 /* ERR invalid recursive type */ [A] }
func _() { var _ = new(foo9[int]) }

// crash 12
var u /* ERR cycle */ , i [func /* ERR used as value */ /* ERR used as value */ (u, c /* ERR undefined */ /* ERR undefined */ ) {}(0, len /* ERR must be called */ /* ERR must be called */ )]c /* ERR undefined */ /* ERR undefined */

// crash 15
func y15() { var a /* ERR declared and not used */ interface{ p() } = G15[string]{} }
type G15[X any] s /* ERR undefined */
func (G15 /* ERROR generic type .* without instantiation */ ) p()

// crash 16
type Foo16[T any] r16 /* ERR not a type */
func r16[T any]() Foo16[Foo16[T]] { panic(0) }

// crash 17
type Y17 interface{ c() }
type Z17 interface {
	c() Y17
	Y17 /* ERR duplicate method */
}
func F17[T Z17](T) {}

// crash 18
type o18[T any] []func(_ o18[[]_ /* ERR cannot use _ */ ])

// crash 19
type Z19 [][[]Z19{}[0][0]]c19 /* ERR undefined */

// crash 20
type Z20 /* ERR invalid recursive type */ interface{ Z20 }
func F20[t Z20]() { F20(t /* ERR invalid composite literal type */ {}) }

// crash 21
type Z21 /* ERR invalid recursive type */ interface{ Z21 }
func F21[T Z21]() { ( /* ERR not used */ F21[Z21]) }

// crash 24
type T24[P any] P // ERR cannot use a type parameter as RHS in type declaration
func (r T24[P]) m() { T24 /* ERR without instantiation */ .m() }

// crash 25
type T25[A any] int
func (t T25[A]) m1() {}
var x T25 /* ERR without instantiation */ .m1

// crash 26
type T26 = interface{ F26[ /* ERR interface method must have no type parameters */ Z any]() }
func F26[Z any]() T26 { return F26[] /* ERR operand */ }

// crash 27
func e27[T any]() interface{ x27 /* ERR not a type */ } { panic(0) }
func x27() { e27 /* ERR cannot infer T */ () }
