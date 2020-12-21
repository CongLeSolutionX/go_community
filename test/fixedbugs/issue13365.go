// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 13365: confusing error message (array vs slice)

package main

var t struct{}

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = []int{-1: 0}    // ERROR "index must be non\-negative integer constant|must not be negative"
	_ = [10]int{-1: 0}  // ERROR "index must be non\-negative integer constant|must not be negative"
	_ = [...]int{-1: 0} // ERROR "index must be non\-negative integer constant|must not be negative"
=======
	_ = []int{-1: 0}    // ERROR "index must be non\-negative integer constant|index expression is negative"
	_ = [10]int{-1: 0}  // ERROR "index must be non\-negative integer constant|index expression is negative"
	_ = [...]int{-1: 0} // ERROR "index must be non\-negative integer constant|index expression is negative"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

	_ = []int{100: 0}
	_ = [10]int{100: 0} // ERROR "array index 100 out of bounds|out of range"
	_ = [...]int{100: 0}

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = []int{t}    // ERROR "cannot use .* as (type )?int( in slice literal)?"
	_ = [10]int{t}  // ERROR "cannot use .* as (type )?int( in array literal)?"
	_ = [...]int{t} // ERROR "cannot use .* as (type )?int( in array literal)?"
=======
	_ = []int{t}    // ERROR "cannot use .* as type int in slice literal|incompatible type"
	_ = [10]int{t}  // ERROR "cannot use .* as type int in array literal|incompatible type"
	_ = [...]int{t} // ERROR "cannot use .* as type int in array literal|incompatible type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
