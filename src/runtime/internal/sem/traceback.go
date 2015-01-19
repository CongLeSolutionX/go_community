// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sem

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

func gcallers(gp *_core.G, skip int, pcbuf *uintptr, m int) int {
	return _lock.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, skip, pcbuf, m, nil, nil, 0)
}
