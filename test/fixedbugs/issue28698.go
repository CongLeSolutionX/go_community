// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

// When using non-soft-float, ODIV and OMOD will not
// panic on non-soft-float platforms when working with
// floats.

var x int64 = 1

func main() {
	var y float32 = 1.0
	test(x, y/y)
}

//go:noinline
func test(id int64, a float32) {

	if id != x {
		fmt.Printf("got: %d, want: %d\n", id, x)
		panic("FAIL")
	}
}
