// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// Verify that the compiler complains even if the array
// has length 0.
var a [0]int
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = a[2:] // ERROR "invalid slice index 2|index 2 out of bounds"
=======
var _ = a[2:] // ERROR "invalid slice index 2|array index out of bounds"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

var b [1]int
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = b[2:] // ERROR "invalid slice index 2|index 2 out of bounds"
=======
var _ = b[2:] // ERROR "invalid slice index 2|array index out of bounds"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
