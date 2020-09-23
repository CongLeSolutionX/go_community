// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T0 T0 // ERROR "(?s)(?s)invalid recursive type: T0.\tT0 refers to.\tT0$"
type _ map[T0]int

type T1 struct{ T1 } // ERROR "(?s)invalid recursive type: T1.\tT1 refers to.\tT1$"
type _ map[T1]int

func f() {
	type T2 T2 // ERROR "(?s)invalid recursive type: T2.\tT2 refers to.\tT2$"
	type _ map[T2]int
}

func g() {
	type T3 struct{ T3 } // ERROR "(?s)invalid recursive type: T3.\tT3 refers to.\tT3$"
	type _ map[T3]int
}

func h() {
	type T4 struct{ m map[T4]int } // ERROR "invalid map key"
	type _ map[T4]int              // ERROR "invalid map key"
}
