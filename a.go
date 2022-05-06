// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func rot64nc(x uint64, z uint) uint64 {
	var a uint64

	z &= 63

	// amd64:"ROLQ"
	// ppc64:"ROTL"
	// ppc64le:"ROTL"
	a += x<<z | x>>(64-z)

	// amd64:"RORQ"
	a += x>>z | x<<(64-z)

	return a
}

func main() {

}
