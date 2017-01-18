// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 11371 (cmd/compile: meaningless error message "truncated to
// integer")

package issue11371

const a int = 1.1        // ERROR "1.1 truncated to integer"
const b int = 1e20       // ERROR "overflows int"
const c int = 1 + 1e-100 // ERROR "invalid floating-point constant expression in integer context"
const d int = 1 - 1e-100 // ERROR "invalid floating-point constant expression in integer context"
const e int = 1.00000001 // ERROR "invalid floating-point constant expression in integer context"
const f int = 0.00000001 // ERROR "1e-08 truncated to integer"
