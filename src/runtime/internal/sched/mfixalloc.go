// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Fixed-size object allocator.  Returned memory is not zeroed.
//
// See malloc.h for overview.

package sched

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func FixAlloc_Free(f *_lock.Fixalloc, p unsafe.Pointer) {
	f.Inuse -= f.Size
	v := (*_lock.Mlink)(p)
	v.Next = f.List
	f.List = v
}
