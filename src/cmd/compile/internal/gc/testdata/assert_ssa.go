// run

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests type assertion expressions and statements

package main

import "fmt"

type (
	S struct{}
	T struct{}

	I interface {
		F()
	}
)

var (
	s *S
	t *T
)

func (s *S) F() {}
func (t *T) F() {}

// TODO: implement panicking when type assertion fails,
// enable the code and tests below, and add i2t tests.

// func e2t_ssa(e interface{}) *T {
// 	return e.(*T)
// }

// func i2t_ssa(i I) *T {
// 	return i.(*T)
// }

// func testAssertE2TOk() {
// 	if got := e2t_ssa(t); got != t {
// 		fmt.Printf("e2t_ssa(t)=%v want %v", got, t)
// 		failed = true
// 	}
// }

// func testAssertE2TPanic() {
// 	var got *T
// 	defer func() {
// 		if got != nil {
// 			fmt.Printf("e2i_ssa(s)=%v want nil", got)
// 			failed = true
// 		}
// 		err := recover()
// 		if _, ok := err.(*runtime.TypeAssertionError); !ok {
// 			fmt.Printf("e2i_ssa(s) panic type %T", err)
// 		}
// 	}()
// 	got = e2t_ssa(s)
// 	fmt.Printf("e2i_ssa(s) should panic")
// 	failed = true
// }

func e2t2_ssa(e interface{}) (*T, bool) {
	t, ok := e.(*T)
	return t, ok
}

func i2t2_ssa(i I) (*T, bool) {
	t, ok := i.(*T)
	return t, ok
}

func testAssertE2T2() {
	if got, ok := e2t2_ssa(t); !ok || got != t {
		fmt.Printf("e2t2_ssa(t)=(%v, %v) want (%v, %v)", got, ok, t, true)
		failed = true
	}
	if got, ok := e2t2_ssa(s); ok || got != nil {
		fmt.Printf("e2t2_ssa(s)=(%v, %v) want (%v, %v)", got, ok, nil, false)
		failed = true
	}
}

func testAssertI2T2() {
	if got, ok := i2t2_ssa(t); !ok || got != t {
		fmt.Printf("i2t2_ssa(t)=(%v, %v) want (%v, %v)", got, ok, t, true)
		failed = true
	}
	if got, ok := i2t2_ssa(s); ok || got != nil {
		fmt.Printf("i2t2_ssa(s)=(%v, %v) want (%v, %v)", got, ok, nil, false)
		failed = true
	}
}

var failed = false

func main() {
	// testAssertE2TOk()
	// testAssertE2TPanic()
	testAssertE2T2()
	testAssertI2T2()
	if failed {
		panic("failed")
	}
}
