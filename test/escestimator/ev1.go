// errorcheck -0 -z=1

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Demonstrate/verify operation of escape analysis experiments

var v = 11
var letters = "abcdefghijklmnopqrstuvwxyz"

func main() {
	a := make([]int, v)    // ERRORANY "alloc,.test/escestimator/ev1.go.,15,11,1,.makeslice."
	s := make([]string, v) // ERRORANY "alloc,.test/escestimator/ev1.go.,16,11,1,.makeslice."
	for i := range a {
		a[i] = i*i + 1
		s[i] = letters[i : i+3]
	}
	for i := range a {
		println("a[", i, "]=", a[i], ", s[", i, "]=", s[i])
	}
}
