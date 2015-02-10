// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package finalize

import (
	_core "runtime/internal/core"
)

type Functype struct {
	typ       _core.Type
	Dotdotdot bool
	In        _core.Slice
	Out       _core.Slice
}
