// errorcheck

// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 9083: map/chan error messages show non-explicit capacity.

package main

// untyped constant
const zero = 0

func main() {
	var x int
	x = make(map[int]int) // GC_ERROR "cannot use make\(map\[int\]int\)"
	x = make(map[int]int, 0) // GC_ERROR "cannot use make\(map\[int\]int, 0\)"
	x = make(map[int]int, zero) // GC_ERROR "cannot use make\(map\[int\]int, zero\)"
	x = make(chan int) // GC_ERROR "cannot use make\(chan int\)"
	x = make(chan int, 0) // GC_ERROR "cannot use make\(chan int, 0\)"
	x = make(chan int, zero) // GC_ERROR "cannot use make\(chan int, zero\)"
}
