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
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	bard // ERROR "constant 256 overflows byte|cannot convert"
=======
	bard // ERROR "constant 256 overflows byte|integer constant overflow"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
)

const (
	c = len([1 - iota]int{})
	d
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	e // ERROR "array bound must be non-negative|invalid array length"
	f // ERROR "array bound must be non-negative|invalid array length"
=======
	e // ERROR "array bound must be non-negative|negative array bound"
	f // ERROR "array bound must be non-negative|negative array bound"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
)
