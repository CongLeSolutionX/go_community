// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
func f(a, b, c, d ...int)       {} // ERROR "non-final parameter a|can only use ... with final parameter"
func g(a ...int, b ...int)      {} // ERROR "non-final parameter a|can only use ... with final parameter"
func h(...int, ...int, float32) {} // ERROR "non-final parameter|can only use ... with final parameter"
=======
func f(a, b, c, d ...int)       {} // ERROR "non-final parameter a|only permits one name"
func g(a ...int, b ...int)      {} // ERROR "non-final parameter a|must be last parameter"
func h(...int, ...int, float32) {} // ERROR "non-final parameter|must be last parameter"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
type a func(...float32, ...interface{}) // ERROR "non-final parameter|can only use ... with final parameter"
=======
type a func(...float32, ...interface{}) // ERROR "non-final parameter|must be last parameter"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
type b interface {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	f(...int, ...int)                // ERROR "non-final parameter|can only use ... with final parameter"
	g(a ...int, b ...int, c float32) // ERROR "non-final parameter a|can only use ... with final parameter"
=======
	f(...int, ...int)                // ERROR "non-final parameter|must be last parameter"
	g(a ...int, b ...int, c float32) // ERROR "non-final parameter a|must be last parameter"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	valid(...int)
}
