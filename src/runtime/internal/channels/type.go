// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package channels

import (
	_core "runtime/internal/core"
)

type Chantype struct {
	typ  _core.Type
	Elem *_core.Type
	dir  uintptr
}
