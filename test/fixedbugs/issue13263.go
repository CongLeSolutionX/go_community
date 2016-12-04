// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 13264: integer to uinptr conversion in global variable
// initialization failed. Also uint to uint64 failed.
// They all compiled alright in a function.

package b

var (
	a uint
	b = a
	c = (uintptr)(b)
	d = (uint64)(b)
)

func f1() {
	var (
		a uint
		b = a
		c = (uintptr)(b)
		d = (uint64)(b)
	)

	var _ = c
	var _ = d
}
