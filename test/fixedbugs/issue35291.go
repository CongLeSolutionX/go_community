// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check error message for duplicated index in slice literal

package p

var s = []string{
	1: "dup",
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	1: "dup", // ERROR "duplicate index in slice literal: 1|duplicate index 1 in array or slice literal"
=======
	1: "dup", // ERROR "duplicate index in slice literal: 1|duplicate value for index 1"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
