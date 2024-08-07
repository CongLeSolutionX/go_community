// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
	"unsafe"
)

type CtrlGroup = ctrlGroup

const DebugLog = debugLog

var AlignUpPow2 = alignUpPow2

const MaxTableCapacity = maxTableCapacity

func NewTestMap[K comparable, V any](length uint64) (*Map, *abi.SwissMapType) {
	mt := newTestMapType[K, V]()
	return NewMap(mt, length), mt
}

func (m *Map) TableFor(key unsafe.Pointer) *table {
	hash := m.typ.Hasher(key, m.seed)
	idx := m.directoryIndex(hash)
	return m.directory[idx]
}

// Returns the start address of the groups array.
func (t *table) GroupsStart() unsafe.Pointer {
	return t.groups.data
}

// Returns the length of the groups array.
func (t *table) GroupsLength() uintptr {
	return uintptr(t.groups.lengthMask + 1)
}
