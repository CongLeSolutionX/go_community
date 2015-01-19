// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime _type representation.

package finalize

import (
	_core "runtime/internal/core"
)

type functype struct {
	typ       _core.Type
	dotdotdot bool
	in        _core.Slice
	out       _core.Slice
}
