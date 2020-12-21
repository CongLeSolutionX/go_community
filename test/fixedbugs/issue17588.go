// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 17588: internal compiler error in typecheckclosure()
// because in case of Func.Nname.Type == nil, Decldepth
// is not initialized in typecheckfunc(). This test
// produces that case.

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
type F func(b T)  // ERROR "T .*is not a type"
=======
type F func(b T)  // ERROR "T is not a type|expected type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

func T(fn F) {
    func() {
        fn(nil)  // If Decldepth is not initialized properly, typecheckclosure() Fatals here.
    }()
}
