// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that erroneous use of init is detected.
// Does not compile.

package main

func init() {
}

func main() {
	init()         // ERROR "undefined.*init"
<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
	runtime.init() // ERROR "undefined.*runtime\.init|undefined: runtime"
=======
	runtime.init() // ERROR "undefined.*runtime\.init|reference to undefined name"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
	var _ = init   // ERROR "undefined.*init"
}
