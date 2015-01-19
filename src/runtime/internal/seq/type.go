// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime _type representation.

package seq

import (
	_core "runtime/internal/core"
)

type slicetype struct {
	typ  _core.Type
	elem *_core.Type
}
