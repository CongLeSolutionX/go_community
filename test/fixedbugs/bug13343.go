// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var (
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	a, b = f() // ERROR "initialization loop|depends upon itself|initialization cycle"
	c    = b
=======
	a, b = f() // ERROR "initialization loop|depends upon itself|depend upon each other"
	c    = b   // GCCGO_ERROR "depends upon itself|depend upon each other"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
)

func f() (int, int) {
	return c, c
}

func main() {}
