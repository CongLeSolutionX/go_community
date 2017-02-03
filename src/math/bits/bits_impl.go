// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides basic and slow implementations of the bits functions.
// TODO(gri) Replace them with fast implementations.

package bits

func nlz(x uint64, size uint) int {
	n := 0
	for b := uint64(1) << (size - 1); b != 0 && x&b == 0; b >>= 1 {
		n++
	}
	return n
}

func ntz(x uint64, size uint) int {
	m := uint64(1)<<size - 1
	n := 0
	for b := uint64(1); b&m != 0 && x&b == 0; b <<= 1 {
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

func rot(x uint64, size, k uint) uint64 {
	return x<<k | x>>(size-k)&(1<<k-1)
}

func rev(x uint64, size uint) uint64 {
	var r uint64
	for i := size; i > 0; i-- {
		r = r<<1 | x&1
		x >>= 1
	}
	return r
}

func swap(x uint64, size uint) uint64 {
	var r uint64
	for i := size / 8; i > 0; i-- {
		r = r<<8 | x&0xff
		x >>= 8
	}
	return r
}

func log(x uint64, size uint) int {
	return int(size) - nlz(x, size) - 1
}
