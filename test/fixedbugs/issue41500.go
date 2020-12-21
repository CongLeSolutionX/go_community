// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type s struct {
	slice []int
}

func f() {
	var x *s

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = x == nil || len(x.slice) // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|cannot convert x == nil"
	_ = len(x.slice) || x == nil // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|cannot convert x == nil"
	_ = x == nil && len(x.slice) // ERROR "invalid operation: .+ \(operator && not defined on int\)|cannot convert x == nil"
	_ = len(x.slice) && x == nil // ERROR "invalid operation: .+ \(operator && not defined on int\)|cannot convert x == nil"
=======
	_ = x == nil || len(x.slice) // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types"
	_ = len(x.slice) || x == nil // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types"
	_ = x == nil && len(x.slice) // ERROR "invalid operation: .+ \(operator && not defined on int\)|incompatible types"
	_ = len(x.slice) && x == nil // ERROR "invalid operation: .+ \(operator && not defined on int\)|incompatible types"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
