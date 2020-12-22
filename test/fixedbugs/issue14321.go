// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that error message reports _ambiguous_ method.

package p

type A struct{
	H int
}

func (A) F() {}
func (A) G() {}

type B struct{
	G int
	H int
}

func (B) F() {}

type C struct {
	A
	B
}

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = C.F // ERROR "ambiguous selector"
var _ = C.G // ERROR "ambiguous selector"
var _ = C.H // ERROR "ambiguous selector"
var _ = C.I // ERROR "no method I|C.I undefined"
=======
var _ = C.F // ERROR "ambiguous"
var _ = C.G // ERROR "ambiguous"
var _ = C.H // ERROR "ambiguous"
var _ = C.I // ERROR "no method .*I.*"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
