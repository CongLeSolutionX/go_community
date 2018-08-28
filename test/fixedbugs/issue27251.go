// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 26495: gccgo produces incorrect order of evaluation
// for expressions involving &&, || subexpressions.

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
