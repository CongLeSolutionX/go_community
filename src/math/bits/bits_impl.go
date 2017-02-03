// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides basic and slow implementations of the bits functions.
// TODO(gri) Replace them with faster implementations.

package bits

func ntz(x uint64, size int) (n int) {
	if x == 0 {
		return size
	}
	// If popcount is fast, return popcount(^x & (x - 1))
	// instead of the code below.
	x = ^x & (x - 1)
	for x != 0 {
		n++
		x >>= 1
	}
	return
}

func pop(x uint64) (n int) {
	for x != 0 {
		n++
		x &= x - 1
	}
	return
}

func rot(x uint64, size, k uint) uint64 {
	return x<<k | x>>(size-k)&(1<<k-1)
}

func rev(x uint64, size uint) (r uint64) {
	for i := size; i > 0; i-- {
		r = r<<1 | x&1
		x >>= 1
	}
	return
}

func swap(x uint64, size uint) (r uint64) {
	for i := size / 8; i > 0; i-- {
		r = r<<8 | x&0xff
		x >>= 8
	}
	return
}

func blen(x uint64) (i int) {
	for ; x >= 1<<(16-1); x >>= 16 {
		i += 16
	}
	if x >= 1<<(8-1) {
		x >>= 8
		i += 8
	}
	if x >= 1<<(4-1) {
		x >>= 4
		i += 4
	}
	if x >= 1<<(2-1) {
		x >>= 2
		i += 2
	}
	if x >= 1<<(1-1) {
		i++
	}
	return
}
