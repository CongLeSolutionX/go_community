// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that > 10 non-syntax errors on the same line
// don't lead to early exit. Specifically, here test
// that we see the initialization error for variable
// s.

package main

type T struct{}

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	t := T{X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1} // ERROR "unknown field 'X' in struct literal of type T|unknown field X"
	_ = t
	var s string = 1 // ERROR "cannot use 1|cannot convert"
	_ = s
=======
	t := T{X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1, X: 1} // ERROR "unknown field 'X' in struct literal of type T|unknown field .*X.* in .*T.*"
	var s string = 1 // ERROR "cannot use 1|incompatible type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
