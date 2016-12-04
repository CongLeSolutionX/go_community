// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 13264: integer to uinptr conversion in global variable
// initialization failed. Worked alright in a function.

package main

var (
	x uint
	y = x
	z = (uintptr)(y)
)

func f1() {
	var (
		x uint
		y = x
		z = (uintptr)(y)
	)

	var _ = z
}

func main() {
}
