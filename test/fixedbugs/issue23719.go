// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	v1 := [2]int32{-102, -102}
	v2 := [2]int32{-102, 1126}
	if v1 == v2 {
		panic("bad comparison")
	}
}
