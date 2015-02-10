// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
)

// isDirectIface reports whether t is stored directly in an interface value.
func IsDirectIface(t *_core.Type) bool {
	return t.Kind&_channels.KindDirectIface != 0
}
