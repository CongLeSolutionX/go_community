// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 20185: type switching on untyped values (e.g. nil or consts)
// caused an internal compiler error.

package p

func F() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	switch t := nil.(type) { // ERROR "cannot type switch on non-interface value nil|not an interface"
=======
	switch t := nil.(type) { // ERROR "cannot type switch on non-interface value"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	default:
		_ = t
	}
}

const x = 1

func G() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	switch t := x.(type) { // ERROR "cannot type switch on non-interface value x \(type untyped int\)|not an interface"
=======
	switch t := x.(type) { // ERROR "cannot type switch on non-interface value|declared but not used"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	default:
	}
}
