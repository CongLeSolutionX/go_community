// run

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
)

// check confirms that a zeroing range loop has the
// correct side-effects on the index variable
// for slice size n.
func check(n int) {
	// When n == 0, i is untouched by the range loop.
	// Picking an initial value of -1 for i makes the
	// "want" calculation below correct in all cases.
	i := -1
	s := make([]byte, n)
	for i = range s {
		s[i] = 0
	}
	if want := n - 1; i != want {
		fmt.Printf("index after range with side-effect = %d want %d\n", i, want)
		os.Exit(1)
	}

	i = n + 1
	for i := range s {
		s[i] = 0
	}
	if want := n + 1; i != want {
		fmt.Printf("index after range without side-effect = %d want %d\n", i, want)
		os.Exit(1)
	}
}

func main() {
	check(0)
	check(1)
	check(15)
}
