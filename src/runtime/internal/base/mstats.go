// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory statistics

package base

import (
	"unsafe"
)

// Statistics.
// Shared with Go: if you edit this structure, also edit type MemStats in mem.go.
type Mstats struct {
	// General statistics.
	Alloc       uint64 // bytes allocated and not yet freed
	Total_alloc uint64 // bytes allocated (even if freed)
	Sys         uint64 // bytes obtained from system (should be sum of xxx_sys below, no locking, approximate)
	Nlookup     uint64 // number of pointer lookups
	Nmalloc     uint64 // number of mallocs
	Nfree       uint64 // number of frees

	// Statistics about malloc heap.
	// protected by mheap.lock
	Heap_alloc    uint64 // bytes allocated and not yet freed (same as alloc above)
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
	debuggc        bool

	// Statistics about allocation size classes.

	By_size [NumSizeClasses]struct {
		Size    uint32
		Nmalloc uint64
		Nfree   uint64
	}

	// Statistics below here are not exported to Go directly.

	Tinyallocs uint64 // number of tiny allocations that didn't cause actual allocation; not exported to go directly

	// heap_live is the number of bytes considered live by the GC.
	// That is: retained by the most recent GC plus allocated
	// since then. heap_live <= heap_alloc, since heap_live
	// excludes unmarked objects that have not yet been swept.
	Heap_live uint64

	// heap_scan is the number of bytes of "scannable" heap. This
	// is the live heap (as counted by heap_live), but omitting
	// no-scan objects and no-scan tails of objects.
	Heap_scan uint64

	// heap_marked is the number of bytes marked by the previous
	// GC. After mark termination, heap_live == heap_marked, but
	// unlike heap_live, heap_marked does not change until the
	// next mark termination.
	Heap_marked uint64

	// heap_reachable is an estimate of the reachable heap bytes
	// at the end of the previous GC.
	Heap_reachable uint64
}

var Memstats Mstats

// Atomically increases a given *system* memory stat.  We are counting on this
// stat never overflowing a uintptr, so this function must only be used for
// system memory stats.
//
// The current implementation for little endian architectures is based on
// xadduintptr(), which is less than ideal: xadd64() should really be used.
// Using xadduintptr() is a stop-gap solution until arm supports xadd64() that
// doesn't use locks.  (Locks are a problem as they require a valid G, which
// restricts their useability.)
//
// A side-effect of using xadduintptr() is that we need to check for
// overflow errors.
//go:nosplit
func mSysStatInc(sysStat *uint64, n uintptr) {
	if BigEndian != 0 {
		Xadd64(sysStat, int64(n))
		return
	}
	if val := xadduintptr((*uintptr)(unsafe.Pointer(sysStat)), n); val < n {
		print("runtime: stat overflow: val ", val, ", n ", n, "\n")
		Exit(2)
	}
}

// Atomically decreases a given *system* memory stat.  Same comments as
// mSysStatInc apply.
//go:nosplit
func mSysStatDec(sysStat *uint64, n uintptr) {
	if BigEndian != 0 {
		Xadd64(sysStat, -int64(n))
		return
	}
	if val := xadduintptr((*uintptr)(unsafe.Pointer(sysStat)), uintptr(-int64(n))); val+n < n {
		print("runtime: stat underflow: val ", val, ", n ", n, "\n")
		Exit(2)
	}
}
