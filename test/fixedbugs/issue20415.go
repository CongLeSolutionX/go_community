// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure redeclaration errors report correct position.

package p

// 1
var f byte

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var f interface{} // ERROR "previous declaration at issue20415.go:12|f redeclared"
=======
var f interface{} // ERROR "issue20415.go:12: previous declaration|redefinition"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

func _(f int) {
}

// 2
var g byte

func _(g int) {
}

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var g interface{} // ERROR "previous declaration at issue20415.go:20|g redeclared"
=======
var g interface{} // ERROR "issue20415.go:20: previous declaration|redefinition"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

// 3
func _(h int) {
}

var h byte

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var h interface{} // ERROR "previous declaration at issue20415.go:31|h redeclared"
=======
var h interface{} // ERROR "issue20415.go:31: previous declaration|redefinition"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
