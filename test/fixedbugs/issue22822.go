// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that calling a function shadowing a built-in provides a good
// error message.

package main

func F() {
	slice := []int{1, 2, 3}
	len := int(2)
<<<<<<< HEAD   (c45313 [dev.regabi] cmd/compile: remove prealloc map)
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at LINE-1"
	const iota = 1
	println(iota(slice)) // ERROR "cannot call non-function iota .type int., declared at LINE-1"
=======
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at|expected function"
>>>>>>> BRANCH (89b44b cmd/compile: recognize reassignments involving receives)
}
