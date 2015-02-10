// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// A generic linked list of blocks.  (Typically the block is bigger than sizeof(MLink).)
// Since assignments to mlink.next will result in a write barrier being preformed
// this can not be used by some of the internal GC structures. For example when
// the sweeper is placing an unmarked object on the free list it does not want the
// write barrier to be called since that could result in the object being reachable.
type Mlink struct {
	Next *Mlink
}

// sysAlloc obtains a large chunk of zeroed memory from the
// operating system, typically on the order of a hundred kilobytes
// or a megabyte.
// NOTE: sysAlloc returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysUnused notifies the operating system that the contents
// of the memory region are no longer needed and can be reused
// for other purposes.
// SysUsed notifies the operating system that the contents
// of the memory region are needed again.
//
// SysFree returns it unconditionally; this is only used if
// an out-of-memory error has been detected midway through
// an allocation.  It is okay if SysFree is a no-op.
//
// SysReserve reserves address space without allocating memory.
// If the pointer passed to it is non-nil, the caller wants the
// reservation there, but SysReserve can still choose another
// location if that one is unavailable.  On some systems and in some
// cases SysReserve will simply check that the address space is
// available and not actually reserve it.  If SysReserve returns
// non-nil, it sets *reserved to true if the address space is
// reserved, false if it has merely been checked.
// NOTE: SysReserve returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysMap maps previously reserved address space for use.
// The reserved argument is true if the address space was really
// reserved, not merely checked.
//
// SysFault marks a (already sysAlloc'd) region to fault
// if accessed.  Used only for debugging the runtime.

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
	List   *Mlink
	Chunk  *byte
	Nchunk uint32
	Inuse  uintptr // in-use bytes now
	Stat   *uint64
}

// Statistics.
// Shared with Go: if you edit this structure, also edit type MemStats in mem.go.
type Mstats struct {
	// General statistics.
	Alloc       uint64 // bytes allocated and still in use
	Total_alloc uint64 // bytes allocated (even if freed)
	Sys         uint64 // bytes obtained from system (should be sum of xxx_sys below, no locking, approximate)
	Nlookup     uint64 // number of pointer lookups
	Nmalloc     uint64 // number of mallocs
	Nfree       uint64 // number of frees

	// Statistics about malloc heap.
	// protected by mheap.lock
	Heap_alloc    uint64 // bytes allocated and still in use
	Heap_sys      uint64 // bytes obtained from system
	Heap_idle     uint64 // bytes in idle spans
	Heap_inuse    uint64 // bytes in non-idle spans
	Heap_released uint64 // bytes released to the os
	Heap_objects  uint64 // total number of allocated objects

	// Statistics about allocation of low-level fixed-size structures.
	// Protected by FixAlloc locks.
	Stacks_inuse uint64 // this number is included in heap_inuse above
	Stacks_sys   uint64 // always 0 in mstats
	Mspan_inuse  uint64 // mspan structures
	Mspan_sys    uint64
	Mcache_inuse uint64 // mcache structures
	Mcache_sys   uint64
	Buckhash_sys uint64 // profiling bucket hash table
	Gc_sys       uint64
	Other_sys    uint64

	// Statistics about garbage collector.
	// Protected by mheap or stopping the world during GC.
	Next_gc        uint64 // next gc (in heap_alloc time)
	Last_gc        uint64 // last gc (in absolute time)
	Pause_total_ns uint64
	Pause_ns       [256]uint64 // circular buffer of recent gc pause lengths
	Pause_end      [256]uint64 // circular buffer of recent gc end times (nanoseconds since 1970)
	Numgc          uint32
	Enablegc       bool
	Debuggc        bool

	// Statistics about allocation size classes.

	By_size [_core.NumSizeClasses]struct {
		Size    uint32
		Nmalloc uint64
		Nfree   uint64
	}

	Tinyallocs uint64 // number of tiny allocations that didn't cause actual allocation; not exported to go directly
}

var Memstats Mstats

// Every MSpan is in one doubly-linked list,
// either one of the MHeap's free lists or one of the
// MCentral's span lists.  We use empty MSpan structures as list heads.

// Central list of free objects of a given size.
type Mcentral struct {
	Lock      _core.Mutex
	Sizeclass int32
	Nonempty  _core.Mspan // list of spans with a free object
	Empty     _core.Mspan // list of spans with no free objects (or cached in an mcache)
}

// Main malloc heap.
// The heap itself is the "free[]" and "large" arrays,
// but all the other global data is here too.
type Mheap struct {
	Lock      _core.Mutex
	Free      [_core.MaxMHeapList]_core.Mspan // free lists of given length
	Freelarge _core.Mspan                     // free lists length >= _MaxMHeapList
	Busy      [_core.MaxMHeapList]_core.Mspan // busy lists of large objects of given length
	Busylarge _core.Mspan                     // busy lists of large objects length >= _MaxMHeapList
	Allspans  **_core.Mspan                   // all spans out there
	Gcspans   **_core.Mspan                   // copy of allspans referenced by gc marker or sweeper
	Nspan     uint32
	Sweepgen  uint32 // sweep generation, see comment in mspan
	Sweepdone uint32 // all spans are swept

	// span lookup
	Spans        **_core.Mspan
	Spans_mapped uintptr

	// range of addresses we might see in the heap
	Bitmap         uintptr
	Bitmap_mapped  uintptr
	Arena_start    uintptr
	Arena_used     uintptr
	Arena_end      uintptr
	Arena_reserved bool

	// write barrier shadow data+heap.
	// 64-bit systems only, enabled by GODEBUG=wbshadow=1.
	Shadow_enabled  bool    // shadow should be updated and checked
	Shadow_reserved bool    // shadow memory is reserved
	Shadow_heap     uintptr // heap-addr + shadow_heap = shadow heap addr
	Shadow_data     uintptr // data-addr + shadow_data = shadow data addr
	Data_start      uintptr // start of shadowed data addresses
	Data_end        uintptr // end of shadowed data addresses

	// central free lists for small size classes.
	// the padding makes sure that the MCentrals are
	// spaced CacheLineSize bytes apart, so that each MCentral.lock
	// gets its own cache line.
	Central [_core.NumSizeClasses]struct {
		Mcentral Mcentral
		pad      [CacheLineSize]byte
	}

	Spanalloc             Fixalloc    // allocator for span*
	Cachealloc            Fixalloc    // allocator for mcache*
	Specialfinalizeralloc Fixalloc    // allocator for specialfinalizer*
	Specialprofilealloc   Fixalloc    // allocator for specialprofile*
	Speciallock           _core.Mutex // lock for sepcial record allocators.

	// Malloc stats.
	Largefree  uint64                       // bytes freed for large objects (>maxsmallsize)
	Nlargefree uint64                       // number of frees for large objects (>maxsmallsize)
	Nsmallfree [_core.NumSizeClasses]uint64 // number of frees for small objects (<=maxsmallsize)
}

var Mheap_ Mheap

// Information from the compiler about the layout of stack frames.
type Bitvector struct {
	N        int32 // # of bits
	Bytedata *uint8
}
