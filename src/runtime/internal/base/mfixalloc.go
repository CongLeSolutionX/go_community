// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Fixed-size object allocator.  Returned memory is not zeroed.
//
// See malloc.go for overview.

package base

import (
	"unsafe"
)

// FixAlloc is a simple free-list allocator for fixed size objects.
// Malloc uses a FixAlloc wrapped around sysAlloc to manages its
// MCache and MSpan objects.
//
// Memory returned by FixAlloc_Alloc is not zeroed.
// The caller is responsible for locking around FixAlloc calls.
// Callers can keep state in the object but the first word is
// smashed by freeing and reallocating.
type Fixalloc struct {
	Size   uintptr
	First  unsafe.Pointer // go func(unsafe.pointer, unsafe.pointer); f(arg, p) called first time p is returned
	Arg    unsafe.Pointer
	List   *mlink
	Chunk  *byte
	Nchunk uint32
	Inuse  uintptr // in-use bytes now
	Stat   *uint64
}

// A generic linked list of blocks.  (Typically the block is bigger than sizeof(MLink).)
// Since assignments to mlink.next will result in a write barrier being preformed
// this can not be used by some of the internal GC structures. For example when
// the sweeper is placing an unmarked object on the free list it does not want the
// write barrier to be called since that could result in the object being reachable.
type mlink struct {
	next *mlink
}

func FixAlloc_Alloc(f *Fixalloc) unsafe.Pointer {
	if f.Size == 0 {
		print("runtime: use of FixAlloc_Alloc before FixAlloc_Init\n")
		Throw("runtime: internal error")
	}

	if f.List != nil {
		v := unsafe.Pointer(f.List)
		f.List = f.List.next
		f.Inuse += f.Size
		return v
	}
	if uintptr(f.Nchunk) < f.Size {
		f.Chunk = (*uint8)(Persistentalloc(FixAllocChunk, 0, f.Stat))
		f.Nchunk = FixAllocChunk
	}

	v := (unsafe.Pointer)(f.Chunk)
	if f.First != nil {
		fn := *(*func(unsafe.Pointer, unsafe.Pointer))(unsafe.Pointer(&f.First))
		fn(f.Arg, v)
	}
	f.Chunk = (*byte)(Add(unsafe.Pointer(f.Chunk), f.Size))
	f.Nchunk -= uint32(f.Size)
	f.Inuse += f.Size
	return v
}

func FixAlloc_Free(f *Fixalloc, p unsafe.Pointer) {
	f.Inuse -= f.Size
	v := (*mlink)(p)
	v.next = f.List
	f.List = v
}
