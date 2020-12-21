// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure redeclaration errors report correct position.

package p

// 1
var f byte

<<<<<<< HEAD   (c45313 [dev.regabi] cmd/compile: remove prealloc map)
var f interface{} // ERROR "issue20415.go:12: previous declaration"
=======
var f interface{} // ERROR "previous declaration at issue20415.go:12|redefinition"
>>>>>>> BRANCH (89b44b cmd/compile: recognize reassignments involving receives)

func _(f int) {
}

// 2
var g byte

func _(g int) {
}

<<<<<<< HEAD   (c45313 [dev.regabi] cmd/compile: remove prealloc map)
var g interface{} // ERROR "issue20415.go:20: previous declaration"
=======
var g interface{} // ERROR "previous declaration at issue20415.go:20|redefinition"
>>>>>>> BRANCH (89b44b cmd/compile: recognize reassignments involving receives)

// 3
func _(h int) {
}

var h byte

<<<<<<< HEAD   (c45313 [dev.regabi] cmd/compile: remove prealloc map)
var h interface{} // ERROR "issue20415.go:31: previous declaration"
=======
var h interface{} // ERROR "previous declaration at issue20415.go:31|redefinition"
>>>>>>> BRANCH (89b44b cmd/compile: recognize reassignments involving receives)
