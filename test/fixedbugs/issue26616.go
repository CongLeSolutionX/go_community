// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var x int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|3\-valued"
=======
var x int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

func f() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	var _ int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|3\-valued"
	var a int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|3\-valued"
	a = three()         // ERROR "assignment mismatch: 1 variable but three returns 3 values|cannot assign"
	b := three()        // ERROR "assignment mismatch: 1 variable but three returns 3 values|cannot initialize"
=======
	var _ int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context"
	var a int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context"
	a = three()         // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context"
	b := three()        // ERROR "assignment mismatch: 1 variable but three returns 3 values|single variable set to multiple-value|multiple-value function call in single-value context"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

	_, _ = a, b
}

func three() (int, int, int)
