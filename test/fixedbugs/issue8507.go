// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 8507
// used to call algtype on invalid recursive type and get into infinite recursion

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
type T struct{ T } // ERROR "invalid recursive type T|cycle"
=======
type T struct{ T } // ERROR "invalid recursive type .*T"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

func f() {
	println(T{} == T{})
}
