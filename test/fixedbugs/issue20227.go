// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 20227: panic while constructing constant "1i/1e-600000000"

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = 1 / 1e-600000000i  // ERROR "(complex )?division by zero"
var _ = 1i / 1e-600000000  // ERROR "(complex )?division by zero"
var _ = 1i / 1e-600000000i // ERROR "(complex )?division by zero"
=======
var _ = 1 / 1e-600000000i  // ERROR "division by zero"
var _ = 1i / 1e-600000000  // ERROR "division by zero"
var _ = 1i / 1e-600000000i // ERROR "division by zero"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = 1 / (1e-600000000 + 1e-600000000i)  // ERROR "(complex )?division by zero"
var _ = 1i / (1e-600000000 + 1e-600000000i) // ERROR "(complex )?division by zero"
=======
var _ = 1 / (1e-600000000 + 1e-600000000i)  // ERROR "division by zero"
var _ = 1i / (1e-600000000 + 1e-600000000i) // ERROR "division by zero"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
