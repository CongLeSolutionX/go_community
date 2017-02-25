// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test integer literal syntax.

package main

import "os"

func main() {
	s := 	0 +
		123 +
		0123 +
		0000 +
		0x0 +
		0x123 +
		0X0 +
		0X123 +
		0b0 +
		0b101 +
		0B0 +
		0B101
	if s != 798 {
		print("s is ", s, "; should be 798\n")
		os.Exit(1)
	}
}
