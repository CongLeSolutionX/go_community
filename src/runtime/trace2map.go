// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.exectracer2

// Simple hash table for tracing. Provides a mapping
// between variable-length data and a unique ID. Subsequent
// puts of the same data will return the same ID.
//
// Uses a region-based allocation scheme and assumes that the
// table doesn't ever grow very big.
//
// This is definitely not a general-purpose hash table! It avoids
// doing any high-level Go operations so it's safe to use even in
// sensitive contexts.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

type traceMap struct {
	lock mutex // Must be acquired on the system stack
	seq  atomic.Uint64
	mem  traceRegionAlloc
	tab  [1 << 13]*traceMapNode
}

type traceMapNode struct {
	_    sys.NotInHeap
	link *traceMapNode
	hash uintptr
	id   uint64
	data []byte
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
	// First, search the hashtable w/o the mutex.
	if id := tab.find(data, size, hash); id != 0 {
		return id, false
	}
	// Now, double check under the mutex.
	// Switch to the system stack so we can acquire tab.lock
	var id uint64
	var added bool
	systemstack(func() {
		lock(&tab.lock)
		if id = tab.find(data, size, hash); id != 0 {
			unlock(&tab.lock)
			return
		}
		// Create new record.
		id = tab.seq.Add(1)
		vd := tab.newTraceMapNode(data, size, hash, id)

		// Insert it into the table.
		part := int(hash % uintptr(len(tab.tab)))
		vd.link = tab.tab[part]
		atomicstorep(unsafe.Pointer(&tab.tab[part]), unsafe.Pointer(vd))
		unlock(&tab.lock)

		added = true
	})
	return id, added
}

// find looks up data in the table, assuming hash is a hash of data.
//
// Returns 0 if the data is not found, and the unique ID for it if it is.
func (tab *traceMap) find(data unsafe.Pointer, size, hash uintptr) uint64 {
	part := int(hash % uintptr(len(tab.tab)))
	for vd := tab.tab[part]; vd != nil; vd = vd.link {
		if vd.hash == hash && uintptr(len(vd.data)) == size {
			if memequal(unsafe.Pointer(&vd.data[0]), data, size) {
				return vd.id
			}
		}
	}
	return 0
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
//
// tab.lock must be held. Must run on the system stack because of this.
//
//go:systemstack
func (tab *traceMap) reset() {
	assertLockHeld(&tab.lock)
	tab.mem.drop()
	tab.seq.Store(0)
	tab.tab = [1 << 13]*traceMapNode{}
}
