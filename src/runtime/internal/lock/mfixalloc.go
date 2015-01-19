// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Fixed-size object allocator.  Returned memory is not zeroed.
//
// See malloc.h for overview.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

func FixAlloc_Alloc(f *Fixalloc) unsafe.Pointer {
	if f.Size == 0 {
		print("runtime: use of FixAlloc_Alloc before FixAlloc_Init\n")
		Throw("runtime: internal error")
	}

	if f.List != nil {
		v := unsafe.Pointer(f.List)
		f.List = f.List.Next
		f.Inuse += f.Size
		return v
	}
	if uintptr(f.Nchunk) < f.Size {
		f.Chunk = (*uint8)(Persistentalloc(_core.FixAllocChunk, 0, f.Stat))
		f.Nchunk = _core.FixAllocChunk
	}

	v := (unsafe.Pointer)(f.Chunk)
	if f.First != nil {
		fn := *(*func(unsafe.Pointer, unsafe.Pointer))(unsafe.Pointer(&f.First))
		fn(f.Arg, v)
	}
	f.Chunk = (*byte)(_core.Add(unsafe.Pointer(f.Chunk), f.Size))
	f.Nchunk -= uint32(f.Size)
	f.Inuse += f.Size
	return v
}
