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

<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
var _ = C.F // ERROR "ambiguous selector"
var _ = C.G // ERROR "ambiguous selector"
var _ = C.H // ERROR "ambiguous selector"
var _ = C.I // ERROR "no method I|C.I undefined"
=======
var _ = C.F // ERROR "ambiguous"
var _ = C.G // ERROR "ambiguous"
var _ = C.H // ERROR "ambiguous"
var _ = C.I // ERROR "no method .*I.*"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
