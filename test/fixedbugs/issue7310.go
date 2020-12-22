// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Internal compiler crash used to stop errors during second copy.

package main

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = copy(nil, []int{}) // ERROR "use of untyped nil|untyped nil"
	_ = copy([]int{}, nil) // ERROR "use of untyped nil|untyped nil"
	_ = 1 + true           // ERROR "mismatched types untyped int and untyped bool|untyped int .* untyped bool"
=======
	_ = copy(nil, []int{}) // ERROR "use of untyped nil|left argument must be a slice"
	_ = copy([]int{}, nil) // ERROR "use of untyped nil|second argument must be slice or string"
	_ = 1 + true           // ERROR "mismatched types untyped int and untyped bool|incompatible types"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
