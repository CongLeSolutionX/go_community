// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 10975: Returning an invalid interface would cause
// `internal compiler error: getinarg: not a func`.

package main

type I interface {
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	int // ERROR "interface contains embedded non-interface int|not an interface"
=======
	int // ERROR "interface contains embedded non-interface"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
}

func New() I {
	return struct{}{}
}
