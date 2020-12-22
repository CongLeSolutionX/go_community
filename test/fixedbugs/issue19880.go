// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T struct {
	f [1]int
}

func a() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = T // ERROR "type T is not an expression|T \(type\) is not an expression"
=======
	_ = T // ERROR "type T is not an expression|invalid use of type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func b() {
	var v [len(T{}.f)]int // ok
	_ = v
}
