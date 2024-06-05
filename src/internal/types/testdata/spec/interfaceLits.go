// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I0 any

var (
	_ = I0{}
	_ = I0{func /* ERROR "non-empty interface literal for empty interface any" */ () {}}
	_ = I0{m /* ERROR "non-empty interface literal" */ : func() {}}
	_ = I0{nil /* ERROR "non-empty interface literal" */}
)

type I1 interface{ m(int) }

var (
	_ = I1{}
	_ = I1{nil}
	_ = I1{func(int) {}}
	_ = I1{func /* ERROR "wrong type for method" */ () int { return 0 }}
	// TODO(gri) consider reporting only the first error
	_ = I1{0 /* ERROR "interface literal element is not a function" */, 1 /* ERROR "interface literal element is not a function" */}
	_ = I1{m: func(int) {}}
	_ = I1{m: func /* ERROR "wrong type for method" */ () int { return 0 }}
	_ = I1{n /* ERROR "method n not found in interface I1" */ : func(int) {}}
	_ = I1{n /* ERROR "method n not found in interface I1" */ : func() int {}}
	_ = I1{m: func(int) {}, m /* ERROR "duplicate method name m in interface literal" */ : nil}
)

type I2 interface {
	m(int)
	n() float64
}

func m(int)      {}
func n() float64 { return 0 }

var (
	_ = I2{}
	_ = I2{nil /* ERROR "missing method name" */}
	_ = I2{m: nil}
	_ = I2{n: nil}
	_ = I2{m: nil, n: nil, m /* ERROR "duplicate method name" */ : nil}
	_ = I2{m: m, n: n}
)

type R interface {
	m([R /* ERROR "invalid recursive type" */ {}]byte)
}
