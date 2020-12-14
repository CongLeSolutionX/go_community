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
<<<<<<< HEAD   (a20021 [dev.typeparams] cmd/compile/internal/types2: bring over sub)
	runtime.init() // ERROR "undefined.*runtime\.init|undefined: runtime"
=======
	runtime.init() // ERROR "undefined.*runtime\.init|reference to undefined name"
>>>>>>> BRANCH (89f383 [dev.regabi] cmd/compile: add register ABI analysis utilitie)
	var _ = init   // ERROR "undefined.*init"
}
