// errorcheck

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
import "fmt"  // ERROR "imported and not used|imported but not used"
=======
import "fmt"  // GC_ERROR "imported and not used"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)

<<<<<<< HEAD   (ddf449 [dev.typeparams] test: exclude 32bit-specific test that fail)
const n = fmt // ERROR "fmt without selector|not in selector"
=======
const n = fmt // ERROR "fmt without selector|unexpected reference to package"
>>>>>>> BRANCH (2a1cf9 [dev.regabi] merge: get recent changes from 1.16dev into reg)
