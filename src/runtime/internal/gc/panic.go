// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	"unsafe"
)

// The arguments associated with a deferred call are stored
// immediately after the _defer header in memory.
//go:nosplit
func DeferArgs(d *_core.Defer) unsafe.Pointer {
	return _core.Add(unsafe.Pointer(d), unsafe.Sizeof(*d))
}
