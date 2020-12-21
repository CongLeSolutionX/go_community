// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 7153: array invalid index error duplicated on successive bad values

package p

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = []int{a: true, true} // ERROR "undefined: a" "cannot use true \(type untyped bool\) as type int in slice literal|cannot convert true"
=======
var _ = []int{a: true, true} // ERROR "undefined: a" "cannot use true \(type untyped bool\) as type int in slice literal|undefined name .*a|incompatible type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
