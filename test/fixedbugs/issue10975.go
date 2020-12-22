// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 10975: Returning an invalid interface would cause
// `internal compiler error: getinarg: not a func`.

package main

type I interface {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	int // ERROR "interface contains embedded non-interface int|not an interface"
=======
	int // ERROR "interface contains embedded non-interface"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func New() I {
	return struct{}{}
}
