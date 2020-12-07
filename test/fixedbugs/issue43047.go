// run

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.
package main

var (
	x32 float32 = -32.0
	x64 float64 = -64.0
)

func main() {
	if uint32(x32) != 0xffffffe0 {
		panic("f32 -> u32 failed")
	}

	if uint64(x32) != 0xffffffffffffffe0 {
		panic("f32 -> u64 failed")
	}

	if uint32(x64) != 0xffffffc0 {
		panic("f64 -> u32 failed")
	}

	if uint64(x64) != 0xffffffffffffffc0 {
		panic("f64 -> u64 failed")
	}
}
