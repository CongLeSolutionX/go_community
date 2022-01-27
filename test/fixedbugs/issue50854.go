// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:noinline
func f(p int64, x, y int64, b bool) bool { return -x <= p && p <= y }

//go:noinline
func g(p int32, x, y int32, b bool) bool { return -x <= p && p <= y }

//go:noinline
func check(b bool) {
	if b {
		return
	}
	panic("FAILURE")
}

func main() {
	check(f(1, -1<<63, 1<<63-1, true))
	check(g(1, -1<<31, 1<<31-1, true))
}
