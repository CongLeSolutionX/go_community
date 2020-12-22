// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that calling a function shadowing a built-in provides a good
// error message.

package main

func F() {
	slice := []int{1, 2, 3}
	_ = slice
	len := int(2)
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at LINE-1|cannot call non-function len"
=======
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at LINE-1|expected function"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	const iota = 1
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	println(iota(slice)) // ERROR "cannot call non-function iota .type int., declared at LINE-1|cannot call non-function iota"
=======
	println(iota(slice)) // ERROR "cannot call non-function iota .type int., declared at LINE-1|expected function"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
