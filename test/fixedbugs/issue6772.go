// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f1() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	for a, a := range []int{1, 2, 3} { // ERROR "a repeated on left side of :=|a redeclared"
=======
	for a, a := range []int{1, 2, 3} { // ERROR "a.* repeated on left side of :="
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
		println(a)
	}
}

func f2() {
	var a int
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	for a, a := range []int{1, 2, 3} { // ERROR "a repeated on left side of :=|a redeclared"
=======
	for a, a := range []int{1, 2, 3} { // ERROR "a.* repeated on left side of :="
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
		println(a)
	}
	println(a)
}
