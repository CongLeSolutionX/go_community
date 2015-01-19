// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sem

import (
	_core "runtime/internal/core"
)

//go:nosplit
func Gomcache() *_core.Mcache {
	return _core.Getg().M.Mcache
}
