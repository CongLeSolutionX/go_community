// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

func f(...int) {}

func g() {
	var x []int
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	f(x, x...) // ERROR "have \(\[\]int, \.\.\.int\)|too many arguments in call to f"
=======
	f(x, x...) // ERROR "have \(\[\]int, \.\.\.int\)|too many arguments"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
