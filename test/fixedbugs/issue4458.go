// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4458: gc accepts invalid method expressions
// like (**T).Method.

package main

type T struct{}

func (T) foo() {}

func main() {
	av := T{}
	pav := &av
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	(**T).foo(&pav) // ERROR "no method foo|requires named type or pointer to named|undefined"
=======
	(**T).foo(&pav) // ERROR "no method .*foo|requires named type or pointer to named"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
}
