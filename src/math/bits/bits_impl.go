// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides basic and slow implementations of the bits functions.
// TODO(gri) Replace them with fast implementations.

package bits

func nlz(x uint64) int {
	n := 0
	for m := uint64(1 << 63); m != 0 && x&m == 0; m >>= 1 {
		n++
	}
	return n
}

func ntz(x uint64) int {
	n := 0
	for m := uint64(1); m != 0 && x&m == 0; m <<= 1 {
		n++
	}
	return n
}

func pop(x uint64) int {
	var n int
	for x != 0 {
		n++
		x &= x - 1
	}
	return n
}

// TODO(gri) rot is untested
func rot(x uint64, size, k uint) uint64 {
	return x<<k | x>>(size-k)&(1<<k-1)
}

// TODO(gri) rev is untested
func rev(x uint64) uint64 {
	var r uint64
	for i := 0; i < 64; i++ {
		r <<= 1 | x&1
		x >>= 1
	}
	return r
}

// TODO(gri) swap is untested
func swap(x uint64) uint64 {
	var r uint64
	for i := 0; i < 8; i++ {
		r <<= 8 | x&0xff
		x >>= 8
	}
	return r
}

func log(x uint64) int {
	return 63 - nlz(x)
}
