// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"slices"
)

func Bug(s1, s2, s3 []string) string {
	var c1 string
	for v1 := range slices.Values(s1) {
		var c2 string
		for v2 := range slices.Values(s2) {
			var c3 string
			for v3 := range slices.Values(s3) {
				c3 = c3 + v3
			}
			c2 = c2 + v2 + c3
		}
		c1 = c1 + v1 + c2
	}
	return c1
}

func main() {
	got := Bug([]string{"1", "2", "3"}, []string{"a", "b", "c"}, []string{"A", "B", "C"})
	want := "1aABCbABCcABC2aABCbABCcABC3aABCbABCcABC"
	if got != want {
		panic("got: " + got + ", want: " + want)
	}
}
