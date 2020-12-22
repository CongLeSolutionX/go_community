// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for golang.org/issue/13471

package main

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	const _ int64 = 1e646456992 // ERROR "integer too large|truncated to .*"
	const _ int32 = 1e64645699  // ERROR "integer too large|truncated to .*"
	const _ int16 = 1e6464569   // ERROR "integer too large|truncated to .*"
	const _ int8 = 1e646456     // ERROR "integer too large|truncated to .*"
	const _ int = 1e64645       // ERROR "integer too large|truncated to .*"
=======
	const _ int64 = 1e646456992 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ int32 = 1e64645699  // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ int16 = 1e6464569   // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ int8 = 1e646456     // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ int = 1e64645       // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	const _ uint64 = 1e646456992 // ERROR "integer too large|truncated to .*"
	const _ uint32 = 1e64645699  // ERROR "integer too large|truncated to .*"
	const _ uint16 = 1e6464569   // ERROR "integer too large|truncated to .*"
	const _ uint8 = 1e646456     // ERROR "integer too large|truncated to .*"
	const _ uint = 1e64645       // ERROR "integer too large|truncated to .*"
=======
	const _ uint64 = 1e646456992 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ uint32 = 1e64645699  // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ uint16 = 1e6464569   // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ uint8 = 1e646456     // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
	const _ uint = 1e64645       // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	const _ rune = 1e64645 // ERROR "integer too large|truncated to .*"
=======
	const _ rune = 1e64645 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
