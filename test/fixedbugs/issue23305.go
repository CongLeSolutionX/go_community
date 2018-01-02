// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var mask1 = func(a, b uint64) uint64 {
	op1 := int32(a)
	op2 := int32(b)
	return uint64(uint32(op1 / op2))
}

func mask2(a, b uint64) uint64 {
	op1 := int32(a)
	op2 := int32(b)
	return uint64(uint32(op1 / op2))
}

func main() {
	res1 := mask1(0x1, 0xfffffffeffffffff)
	res2 := mask2(0x1, 0xfffffffeffffffff)
	if res1 != res2 {
		println("got", res1, "want", res2)
		panic("FAIL")
	}
}
