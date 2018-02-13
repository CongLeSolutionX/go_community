// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	if foo(1) != int32(0x3edae8) {
		panic(42)
	}
}

func foo(_a int32) (r int32) {
	//defer func() { println("r ", r) }() // <- Uncommenting this defer statement makes the panic go away.

	return shr1(int32(shr2(int64(0x14ff6e2207db5d1f), int(_a))), 4)
}

func shr1(n int32, m int) int32 { return n >> uint(m) }

func shr2(n int64, m int) int64 {
	if m < 0 { // <- Commenting out this if statement makes the panic go away.
		m = -m
	}
	if m >= 64 { // <- Commenting out this if statement makes the panic go away.
		return n
	}

	return n >> uint(m)
}
