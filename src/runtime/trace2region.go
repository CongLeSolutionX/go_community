// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.exectracer2

// Simple not-in-heap bump-pointer traceRegion allocator.

package runtime

import (
	"internal/goarch"
	"internal/runtime/atomic"
	"runtime/internal/sys"
	"unsafe"
)

// traceRegionAlloc is a thread-safe region allocator.
// It holds a linked list of traceRegionAllocBlock.
type traceRegionAlloc struct {
	lock    mutex
	current atomic.UnsafePointer // *traceRegionAllocBlock
	full    *traceRegionAllocBlock
}

// traceRegionAllocBlock is a block in traceRegionAlloc.
//
// traceRegionAllocBlock is allocated from non-GC'd memory, so it must not
// contain heap pointers. Writes to pointers to traceRegionAllocBlocks do
// not need write barriers.
type traceRegionAllocBlock struct {
	_ sys.NotInHeap
	traceRegionAllocBlockHeader
	data [traceRegionAllocBlockData]byte
}

type traceRegionAllocBlockHeader struct {
	next *traceRegionAllocBlock
	off  atomic.Uintptr
}

const traceRegionAllocBlockData = 64<<10 - unsafe.Sizeof(traceRegionAllocBlockHeader{})

// alloc allocates n-byte block.
func (a *traceRegionAlloc) alloc(n uintptr) *notInHeap {
	n = alignUp(n, goarch.PtrSize)
	if n > traceRegionAllocBlockData {
		throw("traceRegion: alloc too large")
	}

	// Try to bump-pointer allocate into the current block.
	block := (*traceRegionAllocBlock)(a.current.Load())
	if block != nil {
		r := block.off.Add(n)
		if r <= uintptr(len(block.data)) {
			return (*notInHeap)(unsafe.Pointer(&block.data[r-n]))
		}
	}

	// Try to install a new block.
	lock(&a.lock)

	// Check block again under the lock. Someone may
	// have gotten here first.
	block = (*traceRegionAllocBlock)(a.current.Load())
	if block != nil {
		r := block.off.Add(n)
		if r <= uintptr(len(block.data)) {
			unlock(&a.lock)
			return (*notInHeap)(unsafe.Pointer(&block.data[r-n]))
		}

		// Add the existing block to the full list.
		block.next = a.full
		a.full = block
	}

	// Allocate a new block.
	block = (*traceRegionAllocBlock)(sysAlloc(unsafe.Sizeof(traceRegionAllocBlock{}), &memstats.other_sys))
	if block == nil {
		throw("traceRegion: out of memory")
	}

	// Allocate space for our current request, so we always make
	// progress.
	block.off.Store(n)
	x := (*notInHeap)(unsafe.Pointer(&block.data[0]))

	// Publish the new block.
	a.current.Store(unsafe.Pointer(block))
	unlock(&a.lock)
	return x
}

// drop frees all previously allocated memory and resets the allocator.
func (a *traceRegionAlloc) drop() {
	for a.full != nil {
		block := a.full
		a.full = block.next
		sysFree(unsafe.Pointer(block), unsafe.Sizeof(traceRegionAllocBlock{}), &memstats.other_sys)
	}
	sysFree(a.current.Load(), unsafe.Sizeof(traceRegionAllocBlock{}), &memstats.other_sys)
	a.current.Store(nil)
}
