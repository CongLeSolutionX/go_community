// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

func sysFault(v unsafe.Pointer, n uintptr) {
	_base.Mmap(v, n, _base.PROT_NONE, _base.MAP_ANON|_base.MAP_PRIVATE|_base.MAP_FIXED, -1, 0)
}
