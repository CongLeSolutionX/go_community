// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T T // ERROR "invalid recursive type"
type _ map[T]int

func f() {
	type T1 T1 // ERROR "invalid recursive type"
	type _ map[T1]int
}

func g() {
	type T2 struct{ T2 } // ERROR "invalid recursive type"
	type _ map[T2]int
}

func h() {
	type T3 struct{ m map[T3]int }
	type _ map[T3]int // ERROR "invalid map key type"
}
