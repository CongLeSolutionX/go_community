// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 6402: spurious 'use of untyped nil' error

package p

func f() uintptr {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return nil // ERROR "cannot use nil as type uintptr in return argument|cannot convert nil"
=======
	return nil // ERROR "cannot use nil as type uintptr in return argument|incompatible type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
