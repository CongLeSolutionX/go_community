// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// unary expressions

package expr0 

type mybool bool

var (
	// bool
	b0 = true
	b1 bool = b0
	b2 = !true
	b3 = !b1
	b4 bool = !true
	b5 bool = !b4
	b6 = +b0 /* ERR not defined */
	b7 = -b0 /* ERR not defined */
	b8 = ^b0 /* ERR not defined */
	b9 = *b0 /* ERR cannot indirect */
	b10 = &true /* ERR cannot take address */
	b11 = &b0
	b12 = <-b0 /* ERR cannot receive */
	b13 = & & /* ERR cannot take address */ b0

	// byte
	_ = byte(0)
	_ = byte(- /* ERR cannot convert */ 1)
	_ = - /* ERR -byte(1) (constant -1 of type byte) overflows byte */ byte(1) // test for issue 11367
	_ = byte /* ERR overflows byte */ (0) - byte(1)

	// int
	i0 = 1
	i1 int = i0
	i2 = +1
	i3 = +i0
	i4 int = +1
	i5 int = +i4
	i6 = -1
	i7 = -i0
	i8 int = -1
	i9 int = -i4
	i10 = !i0 /* ERR not defined */
	i11 = ^1
	i12 = ^i0
	i13 int = ^1
	i14 int = ^i4
	i15 = *i0 /* ERR cannot indirect */
	i16 = &i0
	i17 = *i16
	i18 = <-i16 /* ERR cannot receive */

	// uint
	u0 = uint(1)
	u1 uint = u0
	u2 = +1
	u3 = +u0
	u4 uint = +1
	u5 uint = +u4
	u6 = -1
	u7 = -u0
	u8 uint = - /* ERR overflows */ 1
	u9 uint = -u4
	u10 = !u0 /* ERR not defined */
	u11 = ^1
	u12 = ^i0
	u13 uint = ^ /* ERR overflows */ 1
	u14 uint = ^u4
	u15 = *u0 /* ERR cannot indirect */
	u16 = &u0
	u17 = *u16
	u18 = <-u16 /* ERR cannot receive */
	u19 = ^uint(0)

	// float64
	f0 = float64(1)
	f1 float64 = f0
	f2 = +1
	f3 = +f0
	f4 float64 = +1
	f5 float64 = +f4
	f6 = -1
	f7 = -f0
	f8 float64 = -1
	f9 float64 = -f4
	f10 = !f0 /* ERR not defined */
	f11 = ^1
	f12 = ^i0
	f13 float64 = ^1
	f14 float64 = ^f4 /* ERR not defined */
	f15 = *f0 /* ERR cannot indirect */
	f16 = &f0
	f17 = *u16
	f18 = <-u16 /* ERR cannot receive */

	// complex128
	c0 = complex128(1)
	c1 complex128 = c0
	c2 = +1
	c3 = +c0
	c4 complex128 = +1
	c5 complex128 = +c4
	c6 = -1
	c7 = -c0
	c8 complex128 = -1
	c9 complex128 = -c4
	c10 = !c0 /* ERR not defined */
	c11 = ^1
	c12 = ^i0
	c13 complex128 = ^1
	c14 complex128 = ^c4 /* ERR not defined */
	c15 = *c0 /* ERR cannot indirect */
	c16 = &c0
	c17 = *u16
	c18 = <-u16 /* ERR cannot receive */

	// string
	s0 = "foo"
	s1 = +"foo" /* ERR not defined */
	s2 = -s0 /* ERR not defined */
	s3 = !s0 /* ERR not defined */
	s4 = ^s0 /* ERR not defined */
	s5 = *s4
	s6 = &s4
	s7 = *s6
	s8 = <-s7

	// channel
	ch chan int
	rc <-chan float64
	sc chan <- string
	ch0 = +ch /* ERR not defined */
	ch1 = -ch /* ERR not defined */
	ch2 = !ch /* ERR not defined */
	ch3 = ^ch /* ERR not defined */
	ch4 = *ch /* ERR cannot indirect */
	ch5 = &ch
	ch6 = *ch5
	ch7 = <-ch
	ch8 = <-rc
	ch9 = <-sc /* ERR cannot receive */
	ch10, ok = <-ch
	// ok is of type bool
	ch11, myok = <-ch
	_ mybool = myok /* ERROR cannot use .* in variable declaration */
)

// address of composite literals
type T struct{x, y int}

func f() T { return T{} }

var (
	_ = &T{1, 2}
	_ = &[...]int{}
	_ = &[]int{}
	_ = &[]int{}
	_ = &map[string]T{}
	_ = &(T{1, 2})
	_ = &((((T{1, 2}))))
	_ = &f /* ERR cannot take address */ ()
)

// recursive pointer types
type P *P

var (
	p1 P = new(P)
	p2 P = *p1
	p3 P = &p2
)

func g() (a, b int) { return }

func _() {
	_ = -g /* ERR multiple-value g */ ()
	_ = <-g /* ERR multiple-value g */ ()
}

// ~ is accepted as unary operator only permitted in interface type elements
var (
	_ = ~ /* ERR cannot use ~ outside of interface or type constraint */ 0
	_ = ~ /* ERR cannot use ~ outside of interface or type constraint */ "foo"
	_ = ~ /* ERR cannot use ~ outside of interface or type constraint */ i0
)
