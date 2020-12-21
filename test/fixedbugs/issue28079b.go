// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Non-Go-constant but constant values aren't ok for array sizes.

package p

import "unsafe"

type T [uintptr(unsafe.Pointer(nil))]int // ERROR "non-constant array bound|array bound is not constant"

func f() {
<<<<<<< HEAD   (c45313 [dev.regabi] cmd/compile: remove prealloc map)
	_ = complex(1<<uintptr(unsafe.Pointer(nil)), 0) // ERROR "shift of type float64"
=======
	_ = complex(1<<uintptr(unsafe.Pointer(nil)), 0) // GCCGO_ERROR "non-integer type for left operand of shift"
>>>>>>> BRANCH (89b44b cmd/compile: recognize reassignments involving receives)
}
