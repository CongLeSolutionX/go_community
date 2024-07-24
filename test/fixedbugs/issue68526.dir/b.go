// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import "./a"

// The following code must not crash the compiler,
// but instead report an error message (even though
// the error message is incorrect).
type _ = a.A[int] // ERROR "not a generic type"
