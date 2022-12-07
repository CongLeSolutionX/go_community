// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// variable declarations

package decls1

import (
	"math"
)

// Global variables without initialization
var (
	a, b bool
	c byte
	d uint8
	r rune
	i int
	j, k, l int
	x, y float32
	xx, yy float64
	u, v complex64
	uu, vv complex128
	s, t string
	array []byte
	iface interface{}

	blank _ /* ERR cannot use _ */
)

// Global variables with initialization
var (
	s1 = i + j
	s2 = i /* ERR mismatched types */ + x
	s3 = c + d
	s4 = s + t
	s5 = s /* ERR invalid operation */ / t
	s6 = array[t1]
	s7 = array[x /* ERR integer */]
	s8 = &a
	s10 = &42 /* ERR cannot take address */
	s11 = &v
	s12 = -(u + *t11) / *&v
	s13 = a /* ERR shifted operand */ << d
	s14 = i << j
	s18 = math.Pi * 10.0
	s19 = s1 /* ERR cannot call */ ()
 	s20 = f0 /* ERR no value */ ()
	s21 = f6(1, s1, i)
	s22 = f6(1, s1, uu /* ERROR cannot use .* in argument */ )

	t1 int = i + j
	t2 int = i /* ERR mismatched types */ + x
	t3 int = c /* ERROR cannot use .* variable declaration */ + d
	t4 string = s + t
	t5 string = s /* ERR invalid operation */ / t
	t6 byte = array[t1]
	t7 byte = array[x /* ERR must be integer */]
	t8 *int = & /* ERROR cannot use .* variable declaration */ a
	t10 *int = &42 /* ERR cannot take address */
	t11 *complex64 = &v
	t12 complex64 = -(u + *t11) / *&v
	t13 int = a /* ERR shifted operand */ << d
	t14 int = i << j
	t15 math /* ERR not in selector */
	t16 math.xxx /* ERR undefined */
	t17 math /* ERR not a type */ .Pi
	t18 float64 = math.Pi * 10.0
	t19 int = t1 /* ERR cannot call */ ()
	t20 int = f0 /* ERR no value */ ()
	t21 int = a /* ERROR cannot use .* variable declaration */
)

// Various more complex expressions
var (
	u1 = x /* ERR not an interface */ .(int)
	u2 = iface.([]int)
	u3 = iface.(a /* ERR not a type */ )
	u4, ok = iface.(int)
	u5, ok2, ok3 = iface /* ERR assignment mismatch */ .(int)
)

// Constant expression initializations
var (
	v1 = 1 /* ERR mismatched types untyped int and untyped string */ + "foo"
	v2 = c + 255
	v3 = c + 256 /* ERR overflows */
	v4 = r + 2147483647
	v5 = r + 2147483648 /* ERR overflows */
	v6 = 42
	v7 = v6 + 9223372036854775807
	v8 = v6 + 9223372036854775808 /* ERR overflows */
	v9 = i + 1 << 10
	v10 byte = 1024 /* ERR overflows */
	v11 = xx/yy*yy - xx
	v12 = true && false
	v13 = nil /* ERR use of untyped nil */
	v14 string = 257 // ERROR cannot use 257 .* as string value in variable declaration$
	v15 int8 = 257 // ERROR cannot use 257 .* as int8 value in variable declaration .*overflows
)

// Multiple assignment expressions
var (
	m1a, m1b = 1, 2
	m2a, m2b, m2c /* ERR missing init expr for m2c */ = 1, 2
	m3a, m3b = 1, 2, 3 /* ERR extra init expr 3 */
)

func _() {
	var (
		m1a, m1b = 1, 2
		m2a, m2b, m2c /* ERR missing init expr for m2c */ = 1, 2
		m3a, m3b = 1, 2, 3 /* ERR extra init expr 3 */
	)

	_, _ = m1a, m1b
	_, _, _ = m2a, m2b, m2c
	_, _ = m3a, m3b
}

// Declaration of parameters and results
func f0() {}
func f1(a /* ERR not a type */) {}
func f2(a, b, c d /* ERR not a type */) {}

func f3() int { return 0 }
func f4() a /* ERR not a type */ { return 0 }
func f5() (a, b, c d /* ERR not a type */) { return }

func f6(a, b, c int) complex128 { return 0 }

// Declaration of receivers
type T struct{}

func (T) m0() {}
func (*T) m1() {}
func (x T) m2() {}
func (x *T) m3() {}

// Initialization functions
func init() {}
func init /* ERR no arguments and no return values */ (int) {}
func init /* ERR no arguments and no return values */ () int { return 0 }
func init /* ERR no arguments and no return values */ (int) int { return 0 }
func (T) init(int) int { return 0 }
