// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This test case contains two static temps (the array literals)
// with same contents but different sizes.

package main

func main() {
	F([10]int{1, 2, 3, 4, 5, 6}, [20]int{1, 2, 3, 4, 5, 6})
}

//go:noinline
func F(x, y interface{}) {
	x1 := x.([10]int)
	y1 := y.([20]int)
	for i := range y1 {
		if i < 6 {
			if x1[i] != i+1 || y1[i] != i+1 {
				panic("FAIL")
			}
		} else {
			if (i < len(x1) && x1[i] != 0) || y1[i] != 0 {
				panic("FAIL")
			}
		}
	}
}
