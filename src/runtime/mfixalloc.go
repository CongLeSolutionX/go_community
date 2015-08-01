// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Fixed-size object allocator.  Returned memory is not zeroed.
//
// See malloc.go for overview.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

// Initialize f to allocate objects of the given size,
// using the allocator to obtain chunks of memory.
func fixAlloc_Init(f *_base.Fixalloc, size uintptr, first func(unsafe.Pointer, unsafe.Pointer), arg unsafe.Pointer, stat *uint64) {
	f.Size = size
	f.First = *(*unsafe.Pointer)(unsafe.Pointer(&first))
	f.Arg = arg
	f.List = nil
	f.Chunk = nil
	f.Nchunk = 0
	f.Inuse = 0
	f.Stat = stat
}
