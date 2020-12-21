// errorcheck

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var a []int = []int{1: 1}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var b []int = []int{-1: 1} // ERROR "must be non-negative integer constant|must not be negative"
=======
var b []int = []int{-1: 1} // ERROR "must be non-negative integer constant|index expression is negative"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

var c []int = []int{2.0: 2}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var d []int = []int{-2.0: 2} // ERROR "must be non-negative integer constant|must not be negative"
=======
var d []int = []int{-2.0: 2} // ERROR "must be non-negative integer constant|index expression is negative"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

var e []int = []int{3 + 0i: 3}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var f []int = []int{3i: 3} // ERROR "truncated to integer|truncated to int"
=======
var f []int = []int{3i: 3} // ERROR "truncated to integer|index expression is not integer constant"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var g []int = []int{"a": 4} // ERROR "must be non-negative integer constant|cannot convert"
=======
var g []int = []int{"a": 4} // ERROR "must be non-negative integer constant|index expression is not integer constant"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
