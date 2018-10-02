// errorcheck -0 -z=9

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Demonstrate/verify operation of escape analysis experiments

var v = 11
var letters = "abcdefghijklmnopqrstuvwxyz"

func main() {
	a := make([]int, v)
	s := make([]string, v)
	for i := range a {
		a[i] = i*i + 1
		s[i] = letters[i : i+3]
	}
	for i := range a {
		println("a[", i, "]=", a[i], ", s[", i, "]=", s[i])
	}
}
