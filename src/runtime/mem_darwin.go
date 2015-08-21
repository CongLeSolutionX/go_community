// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

func sysUnused(v unsafe.Pointer, n uintptr) {
	// Linux's MADV_DONTNEED is like BSD's MADV_FREE.
	madvise(v, n, _base.MADV_FREE)
}
