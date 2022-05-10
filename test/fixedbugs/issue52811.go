// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 52811: gofrontend loaded x twice

package main

func main() {
	one := map[int]string{0: "one"}
	two := map[int]string{0: "two"}

	x := one
	x[0] += func() string {
		x = two
		return ""
	}()

	if one[0] != "one" || two[0] != "two" {
		panic("double eval of x")
	}
}
