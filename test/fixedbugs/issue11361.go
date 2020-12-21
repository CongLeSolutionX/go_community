// errorcheck

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
import "fmt"  // ERROR "imported and not used|imported but not used"
=======
import "fmt"  // GC_ERROR "imported and not used"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
const n = fmt // ERROR "fmt without selector|not in selector"
=======
const n = fmt // ERROR "fmt without selector|unexpected reference to package"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
