// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 11371 (cmd/compile: meaningless error message "truncated to
// integer")

package issue11371

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
const a int = 1.1        // ERROR "constant 1.1 truncated to integer|truncated to int"
const b int = 1e20       // ERROR "overflows int|truncated to int"
const c int = 1 + 1e-100 // ERROR "constant truncated to integer|truncated to int"
const d int = 1 - 1e-100 // ERROR "constant truncated to integer|truncated to int"
const e int = 1.00000001 // ERROR "constant truncated to integer|truncated to int"
const f int = 0.00000001 // ERROR "constant 1e-08 truncated to integer|truncated to int"
=======
const a int = 1.1        // ERROR "constant 1.1 truncated to integer|floating-point constant truncated to integer"
const b int = 1e20       // ERROR "overflows int|integer constant overflow"
const c int = 1 + 1e-70  // ERROR "constant truncated to integer"
const d int = 1 - 1e-70  // ERROR "constant truncated to integer"
const e int = 1.00000001 // ERROR "constant truncated to integer"
const f int = 0.00000001 // ERROR "constant 1e-08 truncated to integer|floating-point constant truncated to integer"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
