// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var zero int

var sink any

func main() {
	type T [576/8 - 1]*byte // pointer-ful, has type header, maxes out 576-byte size class
	t := &T{}
	sink = t // force heap allocation

	// Here we're relying on the fact that the next object in the 576-byte
	// objects span hasn't been allocated yet, so its type field is nil.
	// Normally this will be true at startup.

	// Bug will happen as soon as the write barrier turns on.
	for range 10000 {
		sink = make([]*byte, 1024)
		s := t[:]
		s = append(s, make([]*byte, zero)...)
	}
}
