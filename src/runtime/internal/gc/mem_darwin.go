// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func sysFault(v unsafe.Pointer, n uintptr) {
	_lock.Mmap(v, n, _lock.PROT_NONE, _lock.MAP_ANON|_lock.MAP_PRIVATE|_lock.MAP_FIXED, -1, 0)
}
