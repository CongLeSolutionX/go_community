// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var (
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	a, b = f() // ERROR "initialization loop|depends upon itself|initialization cycle"
	c    = b
=======
	a, b = f() // ERROR "initialization loop|depends upon itself|depend upon each other"
	c    = b   // GCCGO_ERROR "depends upon itself|depend upon each other"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
)

func f() (int, int) {
	return c, c
}

func main() {}
