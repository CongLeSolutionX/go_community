// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that an incorrect use of the blank identifier is caught.
// Does not compile.

package main

func f() (_, _ []int)         { return }
func g() (x []int, y float64) { return }

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = append(f()) // ERROR "cannot use \[\]int value as type int in append|cannot use .* as int value"
	_ = append(g()) // ERROR "cannot use float64 value as type int in append|cannot use .* as int value"
=======
	_ = append(f()) // ERROR "cannot use \[\]int value as type int in append|incompatible type"
	_ = append(g()) // ERROR "cannot use float64 value as type int in append|incompatible type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
