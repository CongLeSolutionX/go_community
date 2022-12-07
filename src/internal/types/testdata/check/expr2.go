// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// comparisons

package expr2

func _bool() {
	const t = true == true
	const f = true == false
	_ = t /* ERROR operator .* not defined */ < f
	_ = 0 == t /* ERR mismatched types untyped int and untyped bool */
	var b bool
	var x, y float32
	b = x < y
	_ = b
	_ = struct{b bool}{x < y}
}

// corner cases
var (
	v0 = nil == nil // ERR operator == not defined on untyped nil
)

func arrays() {
	// basics
	var a, b [10]int
	_ = a == b
	_ = a != b
	_ = a /* ERR < not defined */ < b
	_ = a == nil /* ERR mismatched types */

	type C [10]int
	var c C
	_ = a == c

	type D [10]int
	var d D
	_ = c == d /* ERR mismatched types */

	var e [10]func() int
	_ = e /* ERR [10]func() int cannot be compared */ == e
}

func structs() {
	// basics
	var s, t struct {
		x int
		a [10]float32
		_ bool
	}
	_ = s == t
	_ = s != t
	_ = s /* ERR < not defined */ < t
	_ = s == nil /* ERR mismatched types */

	type S struct {
		x int
		a [10]float32
		_ bool
	}
	type T struct {
		x int
		a [10]float32
		_ bool
	}
	var ss S
	var tt T
	_ = s == ss
	_ = ss == tt /* ERR mismatched types */

	var u struct {
		x int
		a [10]map[string]int
	}
	_ = u /* ERR cannot be compared */ == u
}

func pointers() {
	// nil
	_ = nil == nil // ERR operator == not defined on untyped nil
	_ = nil != nil // ERR operator != not defined on untyped nil
	_ = nil /* ERR < not defined */ < nil
	_ = nil /* ERR <= not defined */ <= nil
	_ = nil /* ERR > not defined */ > nil
	_ = nil /* ERR >= not defined */ >= nil

	// basics
	var p, q *int
	_ = p == q
	_ = p != q

	_ = p == nil
	_ = p != nil
	_ = nil == q
	_ = nil != q

	_ = p /* ERR < not defined */ < q
	_ = p /* ERR <= not defined */ <= q
	_ = p /* ERR > not defined */ > q
	_ = p /* ERR >= not defined */ >= q

	// various element types
	type (
		S1 struct{}
		S2 struct{}
		P1 *S1
		P2 *S2
	)
	var (
		ps1 *S1
		ps2 *S2
		p1 P1
		p2 P2
	)
	_ = ps1 == ps1
	_ = ps1 == ps2 /* ERR mismatched types */
	_ = ps2 == ps1 /* ERR mismatched types */

	_ = p1 == p1
	_ = p1 == p2 /* ERR mismatched types */

	_ = p1 == ps1
}

func channels() {
	// basics
	var c, d chan int
	_ = c == d
	_ = c != d
	_ = c == nil
	_ = c /* ERR < not defined */ < d

	// various element types (named types)
	type (
		C1 chan int
		C1r <-chan int
		C1s chan<- int
		C2 chan float32
	)
	var (
		c1 C1
		c1r C1r
		c1s C1s
		c1a chan int
		c2 C2
	)
	_ = c1 == c1
	_ = c1 == c1r /* ERR mismatched types */
	_ = c1 == c1s /* ERR mismatched types */
	_ = c1r == c1s /* ERR mismatched types */
	_ = c1 == c1a
	_ = c1a == c1
	_ = c1 == c2 /* ERR mismatched types */
	_ = c1a == c2 /* ERR mismatched types */

	// various element types (unnamed types)
	var (
		d1 chan int
		d1r <-chan int
		d1s chan<- int
		d1a chan<- int
		d2 chan float32
	)
	_ = d1 == d1
	_ = d1 == d1r
	_ = d1 == d1s
	_ = d1r == d1s /* ERR mismatched types */
	_ = d1 == d1a
	_ = d1a == d1
	_ = d1 == d2 /* ERR mismatched types */
	_ = d1a == d2 /* ERR mismatched types */
}

// for interfaces test
type S1 struct{}
type S11 struct{}
type S2 struct{}
func (*S1) m() int
func (*S11) m() int
func (*S11) n()
func (*S2) m() float32

func interfaces() {
	// basics
	var i, j interface{ m() int }
	_ = i == j
	_ = i != j
	_ = i == nil
	_ = i /* ERR < not defined */ < j

	// various interfaces
	var ii interface { m() int; n() }
	var k interface { m() float32 }
	_ = i == ii
	_ = i == k /* ERR mismatched types */

	// interfaces vs values
	var s1 S1
	var s11 S11
	var s2 S2

	_ = i == 0 /* ERR cannot convert */
	_ = i == s1 /* ERR mismatched types */
	_ = i == &s1
	_ = i == &s11

	_ = i == s2 /* ERR mismatched types */
	_ = i == & /* ERR mismatched types */ s2

	// issue #28164
	// testcase from issue
	_ = interface{}(nil) == [ /* ERR slice can only be compared to nil */ ]int(nil)

	// related cases
	var e interface{}
	var s []int
	var x int
	_ = e == s // ERR slice can only be compared to nil
	_ = s /* ERR slice can only be compared to nil */ == e
	_ = e /* ERR operator < not defined on interface */ < x
	_ = x < e // ERR operator < not defined on interface
}

func slices() {
	// basics
	var s []int
	_ = s == nil
	_ = s != nil
	_ = s /* ERR < not defined */ < nil

	// slices are not otherwise comparable
	_ = s /* ERR slice can only be compared to nil */ == s
	_ = s /* ERR < not defined */ < s
}

func maps() {
	// basics
	var m map[string]int
	_ = m == nil
	_ = m != nil
	_ = m /* ERR < not defined */ < nil

	// maps are not otherwise comparable
	_ = m /* ERR map can only be compared to nil */ == m
	_ = m /* ERR < not defined */ < m
}

func funcs() {
	// basics
	var f func(int) float32
	_ = f == nil
	_ = f != nil
	_ = f /* ERR < not defined */ < nil

	// funcs are not otherwise comparable
	_ = f /* ERR func can only be compared to nil */ == f
	_ = f /* ERR < not defined */ < f
}
