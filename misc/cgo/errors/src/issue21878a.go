// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

/*
static void return_void() { return; }
*/
import "C"

func F() {
	// If a C void function can return a value at all, it should return errno,
	// which is not assignable to [0]byte.
	var x [0]byte
	x = C.return_void() // ERROR HERE
	_ = x
}
