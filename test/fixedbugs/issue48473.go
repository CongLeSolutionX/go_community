// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(x uint64) uint64 {
	s := "\x04"
	c := s[0]
	return x - x<<c<<4
}

func g(x uint32) uint32 {
	s := "\x04"
	c := s[0]
	return x - x<<c<<4
}

func main() {
	if f(1) != 0xffffffffffffff01 {
		panic("bad")
	}
	if g(1) != 0xffffff01 {
		panic("bad")
	}
}
