// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Non-Go-constant but constant values aren't ok for shifts.

package p

import "unsafe"

func f() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = complex(1<<uintptr(unsafe.Pointer(nil)), 0) // ERROR "invalid operation: .*shift of type float64.*|shifted operand .* must be integer"
=======
	_ = complex(1<<uintptr(unsafe.Pointer(nil)), 0) // ERROR "invalid operation: .*shift of type float64.*|non-integer type for left operand of shift"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
