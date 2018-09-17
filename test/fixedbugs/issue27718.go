// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// (-0)+0 should be 0, not -0.

package main

//go:noinline
func f(x float64) float64 {
	return x + 0
}

func main() {
	var zero float64
	var inf = 1.0 / zero
	var negZero = -1 / inf
	if 1/f(negZero) != inf {
		panic("negZero+0 != posZero")
	}
}
