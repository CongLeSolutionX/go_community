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

func NewTestMap[K comparable, V any](length uint64) *Map {
	mt := newTestMapType[K, V]()
	return NewMap(mt, length)
}

// Return a key from a group containing no empty slots, or nil if there are no
// full groups.
//
// Also returns nil if a group is full but contains entirely deleted slots.
func (m *Map) KeyFromFullGroup() unsafe.Pointer {
	var lastTab *table
	for i := range m.dirLen {
		t := m.directoryAt(uintptr(i))
		if t == lastTab {
			continue
		}
		lastTab = t

		for i := uint64(0); i <= t.groups.lengthMask; i++ {
			g := t.groups.group(i)
			match := g.ctrls().matchEmpty()
			if match != 0 {
				continue
			}

			// All full or deleted slots.
			for j := uint32(0); j < abi.SwissMapGroupSlots; j++ {
				if g.ctrls().get(j) == ctrlDeleted {
					continue
				}
				return g.key(j)
			}
		}
	}

	return nil
}

func (m *Map) TableFor(key unsafe.Pointer) *table {
	hash := m.typ.Hasher(key, m.seed)
	idx := m.directoryIndex(hash)
	return m.directoryAt(idx)
}

func (t *table) GrowthLeft() uint64 {
	return t.growthLeft
}
