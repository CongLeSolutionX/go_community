// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// Per-thread (in Go, per-P) cache for small objects.
// No locking needed because it is per-thread (per-P).
type Mcache struct {
	// The following members are accessed on every malloc,
	// so they are grouped here for better caching.
	Next_sample      int32   // trigger heap sample after allocating this many bytes
	Local_cachealloc uintptr // bytes allocated from cache since last lock of heap
	Local_scan       uintptr // bytes of scannable heap allocated
	// Allocator cache for tiny objects w/o pointers.
	// See "Tiny allocator" comment in malloc.go.
	Tiny             unsafe.Pointer
	Tinyoffset       uintptr
	Local_tinyallocs uintptr // number of tiny allocs not counted in other stats

	// The rest is not accessed on every malloc.
	Alloc [NumSizeClasses]*Mspan // spans to allocate from

	Stackcache [NumStackOrders]Stackfreelist

	// Local allocator stats, flushed during GC.
	Local_nlookup    uintptr                 // number of pointer lookups
	Local_largefree  uintptr                 // bytes freed for large objects (>maxsmallsize)
	Local_nlargefree uintptr                 // number of frees for large objects (>maxsmallsize)
	Local_nsmallfree [NumSizeClasses]uintptr // number of frees for small objects (<=maxsmallsize)
}

// A gclink is a node in a linked list of blocks, like mlink,
// but it is opaque to the garbage collector.
// The GC does not trace the pointers during collection,
// and the compiler does not emit write barriers for assignments
// of gclinkptr values. Code should store references to gclinks
// as gclinkptr, not as *gclink.
type Gclink struct {
	Next Gclinkptr
}

// A gclinkptr is a pointer to a gclink, but it is opaque
// to the garbage collector.
type Gclinkptr uintptr

// ptr returns the *gclink form of p.
// The result should be used for accessing fields, not stored
// in other data structures.
func (p Gclinkptr) Ptr() *Gclink {
	return (*Gclink)(unsafe.Pointer(p))
}

type Stackfreelist struct {
	List Gclinkptr // linked list of free stacks
	Size uintptr   // total size of stacks in list
}

// dummy MSpan that contains no free objects.
var Emptymspan Mspan

func Allocmcache() *Mcache {
	Lock(&Mheap_.Lock)
	c := (*Mcache)(FixAlloc_Alloc(&Mheap_.Cachealloc))
	Unlock(&Mheap_.Lock)
	Memclr(unsafe.Pointer(c), unsafe.Sizeof(*c))
	for i := 0; i < NumSizeClasses; i++ {
		c.Alloc[i] = &Emptymspan
	}

	// Set first allocation sample size.
	rate := MemProfileRate
	if rate > 0x3fffffff { // make 2*rate not overflow
		rate = 0x3fffffff
	}
	if rate != 0 {
		c.Next_sample = int32(int(Fastrand1()) % (2 * rate))
	}

	return c
}
