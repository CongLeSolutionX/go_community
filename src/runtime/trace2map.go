// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.exectracer2

// Simple append-only thread-safe hash map for tracing.
// Provides a mapping between variable-length data and a
// unique ID. Subsequent puts of the same data will return
// the same ID. The zero value is ready to use.
//
// Uses a region-based allocation scheme internally, and
// reset clears the whole map.
//
// It avoids doing any high-level Go operations so it's safe
// to use even in sensitive contexts.

package runtime

import (
	"internal/goarch"
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

type traceMap struct {
	root atomic.UnsafePointer // *traceMapNode (can't use generics because it's notinheap)
	seq  atomic.Uint64
	mem  traceRegionAlloc
}

// traceMapNode is an implementation of a hash-trie (a trie of the hash bits)
// from https://nullprogram.com/blog/2023/09/30/.
type traceMapNode struct {
	_ sys.NotInHeap

	children [4]atomic.UnsafePointer // *traceMapNode (can't use generics because it's notinheap)
	hash     uintptr
	id       uint64
	data     []byte
}

// stealID steals an ID from the table, ensuring that it will not
// appear in the table anymore.
func (tab *traceMap) stealID() uint64 {
	return tab.seq.Add(1)
}

// put inserts the data into the table.
//
// It's always safe to noescape data because its bytes are always copied.
//
// Returns a unique ID for the data and whether this is the first time
// the data has been added to the map.
func (tab *traceMap) put(data unsafe.Pointer, size uintptr) (uint64, bool) {
	if size == 0 {
		return 0, false
	}
	hash := memhash(data, 0, size)

	var newNode *traceMapNode
	m := &tab.root
	hashIter := hash
	for {
		n := (*traceMapNode)(m.Load())
		if n == nil {
			// Try to insert a new map node. We may end up discarding
			// this node if we fail to insert. This will only happen
			// if we race on inserting the same unique value, and because
			// we hold onto whatever node we create, we risk wasting at
			// most one additional node for a given value. Races should be
			// very rare once the map gets sufficiently large, but common
			// when the map is small. This just means smaller maps are
			// larger than they should be, but we still scale well.
			if newNode == nil {
				newNode = tab.newTraceMapNode(data, size, hash, tab.seq.Add(1))
			}
			if m.CompareAndSwapNoWB(nil, unsafe.Pointer(newNode)) {
				return newNode.id, true
			}
			// Reload n. Because pointers are only stored once,
			// we must have lost the race, and therefore n is not nil
			// anymore.
			n = (*traceMapNode)(m.Load())
		}
		if n.hash == hash && uintptr(len(n.data)) == size {
			if memequal(unsafe.Pointer(&n.data[0]), data, size) {
				return n.id, false
			}
		}
		m = &n.children[hashIter>>(8*goarch.PtrSize-2)]
		hashIter <<= 2
	}
}

func (tab *traceMap) newTraceMapNode(data unsafe.Pointer, size, hash uintptr, id uint64) *traceMapNode {
	// Create data array.
	sl := notInHeapSlice{
		array: tab.mem.alloc(size),
		len:   int(size),
		cap:   int(size),
	}
	memmove(unsafe.Pointer(sl.array), data, size)

	// Create metadata structure.
	meta := (*traceMapNode)(unsafe.Pointer(tab.mem.alloc(unsafe.Sizeof(traceMapNode{}))))
	*(*notInHeapSlice)(unsafe.Pointer(&meta.data)) = sl
	meta.id = id
	meta.hash = hash
	return meta
}

// reset drops all allocated memory from the table and resets it.
func (tab *traceMap) reset() {
	tab.root.Store(nil)
	tab.seq.Store(0)
	tab.mem.drop()
}
