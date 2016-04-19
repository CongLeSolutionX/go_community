// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program computes the value of rng_cooked in rng.go,
// which is used for seeding all instances of rand.Source.
// A 64bit and a 63bit version of the array is printed to
// the standard output.

package main

import "fmt"

const (
	LEN  = 607
	TAP  = 273
	MASK = (1 << 63) - 1
	A    = 48271
	M    = (1 << 31) - 1
	Q    = 44488
	R    = 3399
)

var (
	rng_vec           [LEN]int64
	rng_tap, rng_feed int
)

func seedrand(x int32) int32 {
	hi := x / Q
	lo := x % Q
	x = A*lo - R*hi
	if x < 0 {
		x += M
	}
	return x
}

func srand(seed int32) {
	rng_tap = 0
	rng_feed = LEN - TAP
	seed %= M
	if seed < 0 {
		seed += M
	} else if seed == 0 {
		seed = 89482311
	}
	x := seed
	for i := -20; i < LEN; i++ {
		x = seedrand(x)
		if i >= 0 {
			var u int64
			u = int64(x) << 20
			x = seedrand(x)
			u ^= int64(x) << 10
			x = seedrand(x)
			u ^= int64(x)
			rng_vec[i] = u
		}
	}
}

func vrand() int64 {
	rng_tap--
	if rng_tap < 0 {
		rng_tap += LEN
	}
	rng_feed--
	if rng_feed < 0 {
		rng_feed += LEN
	}
	x := (rng_vec[rng_feed] + rng_vec[rng_tap])
	rng_vec[rng_feed] = x
	return x
}

func main() {
	srand(1)
	for i := uint64(0); i < 7.8e12; i++ {
		vrand()
	}
	fmt.Printf("rng_vec after 7.8e12 calls to vrand:\n%#v\n", rng_vec)
	for i := range rng_vec {
		rng_vec[i] &= MASK
	}
	fmt.Printf("lower 63bit of rng_vec after 7.8e12 calls to vrand:\n%#v\n", rng_vec)
}
