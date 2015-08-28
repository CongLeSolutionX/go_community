// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
// This code was written with an eye toward translating to Go
// in the future.  Methods have the form Type_Method(Type *t, ...).

package base

import (
	"unsafe"
)

const (
	DebugMalloc = false

	XFlagNoScan = FlagNoScan
	XFlagNoZero = FlagNoZero

	MaxTinySize    = TinySize
	XTinySizeClass = TinySizeClass
	XMaxSmallSize  = MaxSmallSize

	XPageShift = PageShift
	XPageSize  = PageSize
	XPageMask  = PageMask

	XMSpanInUse = MSpanInUse

	ConcurrentSweep = XConcurrentSweep
)

const (
	PageShift = 13
	PageSize  = 1 << PageShift
	PageMask  = PageSize - 1
)

const (
	// _64bit = 1 on 64-bit systems, 0 on 32-bit systems
	X64bit = 1 << (^uintptr(0) >> 63) / 2

	// Computed constant.  The definition of MaxSmallSize and the
	// algorithm in msize.go produces some number of different allocation
	// size classes.  NumSizeClasses is that number.  It's needed here
	// because there are static arrays of this length; when msize runs its
	// size choosing algorithm it double-checks that NumSizeClasses agrees.
	NumSizeClasses = 67

	// Tunable constants.
	MaxSmallSize = 32 << 10

	// Tiny allocator parameters, see "Tiny allocator" comment in malloc.go.
	TinySize      = 16
	TinySizeClass = 2

	FixAllocChunk  = 16 << 10              // Chunk size for FixAlloc
	MaxMHeapList   = 1 << (20 - PageShift) // Maximum page length for fixed-size list in MHeap.
	HeapAllocChunk = 1 << 20               // Chunk size for heap growth

	// Per-P, per order stack segment cache size.
	StackCacheSize = 32 * 1024

	// Number of orders that get caching.  Order 0 is FixedStack
	// and each successive order is twice as large.
	// We want to cache 2KB, 4KB, 8KB, and 16KB stacks.  Larger stacks
	// will be allocated directly.
	// Since FixedStack is different on different systems, we
	// must vary NumStackOrders to keep the same maximum cached size.
	//   OS               | FixedStack | NumStackOrders
	//   -----------------+------------+---------------
	//   linux/darwin/bsd | 2KB        | 4
	//   windows/32       | 4KB        | 3
	//   windows/64       | 8KB        | 2
	//   plan9            | 4KB        | 3
	NumStackOrders = 4 - PtrSize/4*Goos_windows - 1*goos_plan9

	// Number of bits in page to span calculations (4k pages).
	// On Windows 64-bit we limit the arena to 32GB or 35 bits.
	// Windows counts memory used by page table into committed memory
	// of the process, so we can't reserve too much memory.
	// See https://golang.org/issue/5402 and https://golang.org/issue/5236.
	// On other 64-bit platforms, we limit the arena to 512GB, or 39 bits.
	// On 32-bit, we don't bother limiting anything, so we use the full 32-bit address.
	// On Darwin/arm64, we cannot reserve more than ~5GB of virtual memory,
	// but as most devices have less than 4GB of physical memory anyway, we
	// try to be conservative here, and only ask for a 2GB heap.
	MHeapMap_TotalBits = (X64bit*Goos_windows)*35 + (X64bit*(1-Goos_windows)*(1-goos_darwin*Goarch_arm64))*39 + goos_darwin*Goarch_arm64*31 + (1-X64bit)*32
	MHeapMap_Bits      = MHeapMap_TotalBits - PageShift

	MaxMem = uintptr(1<<MHeapMap_TotalBits - 1)

	// Max number of threads to run garbage collection.
	// 2, 3, and 4 are all plausible maximums depending
	// on the hardware details of the machine.  The garbage
	// collector scales well to 32 cpus.
	MaxGcproc = 32
)

// Page number (address>>pageShift)
type pageID uintptr

const MaxArena32 = 2 << 30

func mHeap_SysAlloc(h *Mheap, n uintptr) unsafe.Pointer {
	if n > uintptr(h.Arena_end)-uintptr(h.Arena_used) {
		// We are in 32-bit mode, maybe we didn't use all possible address space yet.
		// Reserve some more space.
		p_size := Round(n+PageSize, 256<<20)
		new_end := h.Arena_end + p_size
		if new_end <= h.Arena_start+MaxArena32 {
			// TODO: It would be bad if part of the arena
			// is reserved and part is not.
			var reserved bool
			p := uintptr(SysReserve((unsafe.Pointer)(h.Arena_end), p_size, &reserved))
			if p == h.Arena_end {
				h.Arena_end = new_end
				h.Arena_reserved = reserved
			} else if p+p_size <= h.Arena_start+MaxArena32 {
				// Keep everything page-aligned.
				// Our pages are bigger than hardware pages.
				h.Arena_end = p + p_size
				used := p + (-uintptr(p) & (PageSize - 1))
				mHeap_MapBits(h, used)
				mHeap_MapSpans(h, used)
				h.Arena_used = used
				h.Arena_reserved = reserved
			} else {
				var stat uint64
				SysFree((unsafe.Pointer)(p), p_size, &stat)
			}
		}
	}

	if n <= uintptr(h.Arena_end)-uintptr(h.Arena_used) {
		// Keep taking from our reservation.
		p := h.Arena_used
		sysMap((unsafe.Pointer)(p), n, h.Arena_reserved, &Memstats.Heap_sys)
		mHeap_MapBits(h, p+n)
		mHeap_MapSpans(h, p+n)
		h.Arena_used = p + n
		if Raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}

		if uintptr(p)&(PageSize-1) != 0 {
			Throw("misrounded allocation in MHeap_SysAlloc")
		}
		return (unsafe.Pointer)(p)
	}

	// If using 64-bit, our reservation is all we have.
	if uintptr(h.Arena_end)-uintptr(h.Arena_start) >= MaxArena32 {
		return nil
	}

	// On 32-bit, once the reservation is gone we can
	// try to get memory at a location chosen by the OS
	// and hope that it is in the range we allocated bitmap for.
	p_size := Round(n, PageSize) + PageSize
	p := uintptr(SysAlloc(p_size, &Memstats.Heap_sys))
	if p == 0 {
		return nil
	}

	if p < h.Arena_start || uintptr(p)+p_size-uintptr(h.Arena_start) >= MaxArena32 {
		print("runtime: memory allocated by OS (", p, ") not in usable range [", Hex(h.Arena_start), ",", Hex(h.Arena_start+MaxArena32), ")\n")
		SysFree((unsafe.Pointer)(p), p_size, &Memstats.Heap_sys)
		return nil
	}

	p_end := p + p_size
	p += -p & (PageSize - 1)
	if uintptr(p)+n > uintptr(h.Arena_used) {
		mHeap_MapBits(h, p+n)
		mHeap_MapSpans(h, p+n)
		h.Arena_used = p + n
		if p_end > h.Arena_end {
			h.Arena_end = p_end
		}
		if Raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}
	}

	if uintptr(p)&(PageSize-1) != 0 {
		Throw("misrounded allocation in MHeap_SysAlloc")
	}
	return (unsafe.Pointer)(p)
}

const (
	// flags to malloc
	FlagNoScan = 1 << 0 // GC doesn't have to scan object
	FlagNoZero = 1 << 1 // don't zero memory
)

type persistentAlloc struct {
	base unsafe.Pointer
	off  uintptr
}

var globalAlloc struct {
	Mutex
	persistentAlloc
}

// Wrapper around sysAlloc that can allocate small chunks.
// There is no associated free operation.
// Intended for things like function/type/debug-related persistent data.
// If align is 0, uses default align (currently 8).
func Persistentalloc(size, align uintptr, sysStat *uint64) unsafe.Pointer {
	var p unsafe.Pointer
	Systemstack(func() {
		p = persistentalloc1(size, align, sysStat)
	})
	return p
}

// Must run on system stack because stack growth can (re)invoke it.
// See issue 9174.
//go:systemstack
func persistentalloc1(size, align uintptr, sysStat *uint64) unsafe.Pointer {
	const (
		chunk    = 256 << 10
		maxBlock = 64 << 10 // VM reservation granularity is 64K on windows
	)

	if size == 0 {
		Throw("persistentalloc: size == 0")
	}
	if align != 0 {
		if align&(align-1) != 0 {
			Throw("persistentalloc: align is not a power of 2")
		}
		if align > PageSize {
			Throw("persistentalloc: align is too large")
		}
	} else {
		align = 8
	}

	if size >= maxBlock {
		return SysAlloc(size, sysStat)
	}

	mp := Acquirem()
	var persistent *persistentAlloc
	if mp != nil && mp.P != 0 {
		persistent = &mp.P.Ptr().palloc
	} else {
		Lock(&globalAlloc.Mutex)
		persistent = &globalAlloc.persistentAlloc
	}
	persistent.off = Round(persistent.off, align)
	if persistent.off+size > chunk || persistent.base == nil {
		persistent.base = SysAlloc(chunk, &Memstats.Other_sys)
		if persistent.base == nil {
			if persistent == &globalAlloc.persistentAlloc {
				Unlock(&globalAlloc.Mutex)
			}
			Throw("runtime: cannot allocate memory")
		}
		persistent.off = 0
	}
	p := Add(persistent.base, persistent.off)
	persistent.off += size
	Releasem(mp)
	if persistent == &globalAlloc.persistentAlloc {
		Unlock(&globalAlloc.Mutex)
	}

	if sysStat != &Memstats.Other_sys {
		mSysStatInc(sysStat, size)
		mSysStatDec(&Memstats.Other_sys, size)
	}
	return p
}
