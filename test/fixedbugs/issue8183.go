// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests correct reporting of line numbers for errors involving iota,
// Issue #8183.
package foo

const (
	ok = byte(iota + 253)
	bad
	barn
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	bard // ERROR "constant 256 overflows byte|cannot convert"
=======
	bard // ERROR "constant 256 overflows byte|integer constant overflow"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
)

const (
	c = len([1 - iota]int{})
	d
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	e // ERROR "array bound must be non-negative|invalid array length"
	f // ERROR "array bound must be non-negative|invalid array length"
=======
	e // ERROR "array bound must be non-negative|negative array bound"
	f // ERROR "array bound must be non-negative|negative array bound"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
)
