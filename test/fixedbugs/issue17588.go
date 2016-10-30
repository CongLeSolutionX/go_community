// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 17588: internal compiler error in typecheckclosure()
// because in case of Func.Nname.Type == nil, Decldepth
// is not initialized in typecheckfunc(). This test
// produces that case.

package haha

type Haha func(b Hehe)  // ERROR "Hehe is not a type"

func Hehe(fn Haha) {
    func() {
        fn(nil)  // If Decldepth is not initialized properly, typecheckclosure() Fatals here.
    }()
}
