// errorcheck -0 -m -l

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type t [10000]*int

func (t) f() {
}

func main() {
	x := t{}.f // ERROR "t literal.f does not escape"
	x()

	var i int       // ERROR "moved to heap: i"
	y := (&t{&i}).f // ERROR "&t literal escapes to heap" "\(&t literal\).f does not escape"
	y()

	z := t{&i}.f // ERROR "t literal.f does not escape"
	z()
}
