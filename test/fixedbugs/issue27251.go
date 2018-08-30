// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A tricky case for bounds check elimination.
// x[len(x)-1] is not safe if len(x) is zero.

package main

import (
	"fmt"
	"os"
)

var glob, size int

func test(a []int) int {
	l := len(a)
	glob = a[l-1]
	fmt.Println("FAIL, a bounds check error should have prevented this")
	return glob
}

func main() {
	defer func() {
		os.Exit(0)
	}()
	slc := make([]int, size)
	fmt.Println(test(slc))
}
