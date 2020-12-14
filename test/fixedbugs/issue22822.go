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
<<<<<<< HEAD   (a20021 [dev.typeparams] cmd/compile/internal/types2: bring over sub)
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at|cannot call non-function len"
	_ = slice
=======
	println(len(slice)) // ERROR "cannot call non-function len .type int., declared at LINE-1"
	const iota = 1
	println(iota(slice)) // ERROR "cannot call non-function iota .type int., declared at LINE-1"
>>>>>>> BRANCH (89f383 [dev.regabi] cmd/compile: add register ABI analysis utilitie)
}
