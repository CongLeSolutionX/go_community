// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func g() {}

func f() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	g()[:] // ERROR "g.* used as value"
=======
	g()[:] // ERROR "g.. used as value|attempt to slice object that is not"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func g2() ([]byte, []byte) { return nil, nil }

func f2() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	g2()[:] // ERROR "multiple-value g2.. in single-value context|2-valued g"
=======
	g2()[:] // ERROR "multiple-value g2.. in single-value context|attempt to slice object that is not"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
