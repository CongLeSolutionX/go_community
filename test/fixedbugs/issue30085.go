// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	var c, d = 1, 2, 3 // ERROR "assignment mismatch: 2 variables but 3 values|extra init expr"
	var e, f, g = 1, 2 // ERROR "assignment mismatch: 3 variables but 2 values|missing init expr"
	_, _, _, _ = c, d, e, f
=======
	var c, d = 1, 2, 3 // ERROR "assignment mismatch: 2 variables but 3 values|wrong number of initializations"
	var e, f, g = 1, 2 // ERROR "assignment mismatch: 3 variables but 2 values|wrong number of initializations"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
