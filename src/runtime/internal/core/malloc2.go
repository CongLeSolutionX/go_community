// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
)

// Memory allocator, based on tcmalloc.
// http://goog-perftools.sourceforge.net/doc/tcmalloc.html

// The main allocator works in runs of pages.
// Small allocation sizes (up to and including 32 kB) are
// rounded to one of about 100 size classes, each of which
// has its own free list of objects of exactly that size.
// Any free page of memory can be split into a set of objects
// of one size class, which are then managed using free list
// allocators.
//
// The allocator's data structures are:
//
//	FixAlloc: a free-list allocator for fixed-size objects,
//		used to manage storage used by the allocator.
//	MHeap: the malloc heap, managed at page (4096-byte) granularity.
//	MSpan: a run of pages managed by the MHeap.
//	MCentral: a shared free list for a given size class.
//	MCache: a per-thread (in Go, per-P) cache for small objects.
//	MStats: allocation statistics.
//
// Allocating a small object proceeds up a hierarchy of caches:
//
//	1. Round the size up to one of the small size classes
//	   and look in the corresponding MCache free list.
//	   If the list is not empty, allocate an object from it.
//	   This can all be done without acquiring a lock.
//
//	2. If the MCache free list is empty, replenish it by
//	   taking a bunch of objects from the MCentral free list.
//	   Moving a bunch amortizes the cost of acquiring the MCentral lock.
//
//	3. If the MCentral free list is empty, replenish it by
//	   allocating a run of pages from the MHeap and then
//	   chopping that memory into objects of the given size.
//	   Allocating many objects amortizes the cost of locking
//	   the heap.
//
//	4. If the MHeap is empty or has no page runs large enough,
//	   allocate a new group of pages (at least 1MB) from the
//	   operating system.  Allocating a large run of pages
//	   amortizes the cost of talking to the operating system.
//
// Freeing a small object proceeds up the same hierarchy:
//
//	1. Look up the size class for the object and add it to
//	   the MCache free list.
//
//	2. If the MCache free list is too long or the MCache has
//	   too much memory, return some to the MCentral free lists.
//
//	3. If all the objects in a given span have returned to
//	   the MCentral list, return that span to the page heap.
//
//	4. If the heap has too much memory, return some to the
//	   operating system.
//
//	TODO(rsc): Step 4 is not implemented.
//
// Allocating and freeing a large object uses the page heap
// directly, bypassing the MCache and MCentral free lists.
//
// The small objects on the MCache and MCentral free lists
// may or may not be zeroed.  They are zeroed if and only if
// the second word of the object is zero.  A span in the
// page heap is zeroed unless s->needzero is set. When a span
// is allocated to break into small objects, it is zeroed if needed
// and s->needzero is set. There are two main benefits to delaying the
// zeroing this way:
//
//	1. stack frames allocated from the small object lists
//	   or the page heap can avoid zeroing altogether.
//	2. the cost of zeroing when reusing a small object is
//	   charged to the mutator, not the garbage collector.
//
// This C code was written with an eye toward translating to Go
// in the future.  Methods have the form Type_Method(Type *t, ...).

const (
	PageShift = 13
	PageSize  = 1 << PageShift
	PageMask  = PageSize - 1
)

const (
	// _64bit = 1 on 64-bit systems, 0 on 32-bit systems
	X64bit = 1 << (^uintptr(0) >> 63) / 2

	// Computed constant.  The definition of MaxSmallSize and the
	// algorithm in msize.c produce some number of different allocation
	// size classes.  NumSizeClasses is that number.  It's needed here
	// because there are static arrays of this length; when msize runs its
	// size choosing algorithm it double-checks that NumSizeClasses agrees.
	NumSizeClasses = 67

	// Tunable constants.
	MaxSmallSize = 32 << 10

	// Tiny allocator parameters, see "Tiny allocator" comment in malloc.goc.
	TinySize      = 16
	TinySizeClass = 2

	FixAllocChunk  = 16 << 10              // Chunk size for FixAlloc
	MaxMHeapList   = 1 << (20 - PageShift) // Maximum page length for fixed-size list in MHeap.
	HeapAllocChunk = 1 << 20               // Chunk size for heap growth

	// Per-P, per order stack segment cache size.
	StackCacheSize = 32 * 1024

	// Number of orders that get caching.  Order 0 is FixedStack
	// and each successive order is twice as large.
	NumStackOrders = 3

	// Number of bits in page to span calculations (4k pages).
	// On Windows 64-bit we limit the arena to 32GB or 35 bits.
	// Windows counts memory used by page table into committed memory
	// of the process, so we can't reserve too much memory.
	// See http://golang.org/issue/5402 and http://golang.org/issue/5236.
	// On other 64-bit platforms, we limit the arena to 128GB, or 37 bits.
	// On 32-bit, we don't bother limiting anything, so we use the full 32-bit address.
	MHeapMap_TotalBits = (X64bit*Goos_windows)*35 + (X64bit*(1-Goos_windows))*37 + (1-X64bit)*32
	MHeapMap_Bits      = MHeapMap_TotalBits - PageShift

	MaxMem = uintptr(1<<MHeapMap_TotalBits - 1)

	// Max number of threads to run garbage collection.
	// 2, 3, and 4 are all plausible maximums depending
	// on the hardware details of the machine.  The garbage
	// collector scales well to 32 cpus.
	MaxGcproc = 32
)

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

// Per-thread (in Go, per-P) cache for small objects.
// No locking needed because it is per-thread (per-P).
type Mcache struct {
	// The following members are accessed on every malloc,
	// so they are grouped here for better caching.
	Next_sample      int32  // trigger heap sample after allocating this many bytes
	Local_cachealloc Intptr // bytes allocated (or freed) from cache since last lock of heap
	// Allocator cache for tiny objects w/o pointers.
	// See "Tiny allocator" comment in malloc.goc.
	Tiny             *byte
	Tinysize         uintptr
	Local_tinyallocs uintptr // number of tiny allocs not counted in other stats

	// The rest is not accessed on every malloc.
	Alloc [NumSizeClasses]*Mspan // spans to allocate from

	Stackcache [NumStackOrders]Stackfreelist

	Sudogcache *Sudog

	// Local allocator stats, flushed during GC.
	Local_nlookup    uintptr                 // number of pointer lookups
	Local_largefree  uintptr                 // bytes freed for large objects (>maxsmallsize)
	Local_nlargefree uintptr                 // number of frees for large objects (>maxsmallsize)
	Local_nsmallfree [NumSizeClasses]uintptr // number of frees for small objects (<=maxsmallsize)
}

type Special struct {
	Next   *Special // linked list in span
	Offset uint16   // span offset of object
	Kind   byte     // kind of special
}

type Mspan struct {
	Next     *Mspan    // in a span linked list
	Prev     *Mspan    // in a span linked list
	Start    PageID    // starting page number
	Npages   uintptr   // number of pages in span
	Freelist Gclinkptr // list of free objects
	// sweep generation:
	// if sweepgen == h->sweepgen - 2, the span needs sweeping
	// if sweepgen == h->sweepgen - 1, the span is currently being swept
	// if sweepgen == h->sweepgen, the span is swept and ready to use
	// h->sweepgen is incremented by 2 after every GC
	Sweepgen    uint32
	Ref         uint16   // capacity - number of objects in freelist
	Sizeclass   uint8    // size class
	Incache     bool     // being used by an mcache
	State       uint8    // mspaninuse etc
	Needzero    uint8    // needs to be zeroed before allocation
	Elemsize    uintptr  // computed from sizeclass or from npages
	Unusedsince int64    // first time spotted by gc in mspanfree state
	Npreleased  uintptr  // number of pages released to the os
	Limit       uintptr  // end of data in span
	Speciallock Mutex    // guards specials list
	Specials    *Special // linked list of special records sorted by offset.
}
