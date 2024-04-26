// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"fmt"
	"internal/abi"
	"math/rand/v2"
	"slices"
	"testing"
	"unsafe"
)

func newTestTable[K comparable, V any](length uint64) *table {
	var m map[K]V
	mTyp := abi.TypeOf(m)
	omt := (*abi.OldMapType)(unsafe.Pointer(mTyp))

	mt := &abi.SwissMapType{
		Key:    omt.Key,
		Elem:   omt.Elem,
		Hasher: omt.Hasher,
	}
	return newTable(mt, length)
}

func TestTablePut(t *testing.T) {
	tab := newTestTable[uint32, uint64](8)

	key := uint32(0)
	elem := uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&elem))

		if debugLog {
			fmt.Printf("After put %d: %v\n", key, tab)
		}
	}

	key = uint32(0)
	elem = uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		got, ok := tab.Get(unsafe.Pointer(&key))
		if !ok {
			t.Errorf("Get(%d) got ok false want true", key)
		}
		gotElem := *(*uint64)(got)
		if gotElem != elem {
			t.Errorf("Get(%d) got elem %d want %d", key, gotElem, elem)
		}
	}
}

func TestTableDelete(t *testing.T) {
	tab := newTestTable[uint32, uint64](32)

	key := uint32(0)
	elem := uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&elem))

		if debugLog {
			fmt.Printf("After put %d: %v\n", key, tab)
		}
	}

	key = uint32(0)
	elem = uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		tab.Delete(unsafe.Pointer(&key))
	}

	key = uint32(0)
	elem = uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		_, ok := tab.Get(unsafe.Pointer(&key))
		if ok {
			t.Errorf("Get(%d) got ok true want false", key)
		}
	}
}

func TestTableClear(t *testing.T) {
	tab := newTestTable[uint32, uint64](32)

	key := uint32(0)
	elem := uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&elem))

		if debugLog {
			fmt.Printf("After put %d: %v\n", key, tab)
		}
	}

	tab.Clear()

	if tab.used != 0 {
		t.Errorf("Clear() used got %d want 0", tab.used)
	}

	key = uint32(0)
	elem = uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		_, ok := tab.Get(unsafe.Pointer(&key))
		if ok {
			t.Errorf("Get(%d) got ok true want false", key)
		}
	}
}

// Returns true if tab contains a full group (no empty slots).
//
// This also ignores groups that consist only of tombstones, as
// TestTableRehashInPlace intentionally leaves those behind.
func containsFullNonTombstoneGroup(tab *table) bool {
	const ctrlsAllDeleted =
		(ctrlGroup(ctrlDeleted) << 56) |
		(ctrlGroup(ctrlDeleted) << 48) |
		(ctrlGroup(ctrlDeleted) << 40) |
		(ctrlGroup(ctrlDeleted) << 32) |
		(ctrlGroup(ctrlDeleted) << 24) |
		(ctrlGroup(ctrlDeleted) << 16) |
		(ctrlGroup(ctrlDeleted) << 8) |
		ctrlGroup(ctrlDeleted)

	for i := uint64(0); i < tab.groups.length; i++ {
		g := tab.groups.group(i)
		if *g.ctrls() == ctrlsAllDeleted {
			continue
		}
		match := g.ctrls().matchEmpty()
		if match == 0 {
			return true
		}
	}
	return false
}

func countTombstones(tab *table) uint64 {
	var tombstones uint64
	for i := uint64(0); i < tab.groups.length; i++ {
		g := tab.groups.group(i)
		for j := uint32(0); j < groupSlots; j++ {
			c := g.ctrls().get(j)
			if c == ctrlDeleted {
				tombstones++
			}
		}
	}
	return tombstones
}

// TestTableRehashInPlace execises rehashInPlace by creating lots of tombstones
// that rehashInPlace eliminates.
func TestTableRehashInPlace(t *testing.T) {
	// Rehash triggers at a load factor of 7/8 (0.875).
	//
	// Rehash in place trigger if 1/3 (0.333) of slots are tombstones at
	// that point.
	//
	// Use at most 50% of slots for live keys. That leaves >1/3 available
	// for tombstones before hitting the load factor.
	const capacity = 128
	const usedLimit = capacity / 2
	const wantTombstones = (capacity / 3) + 1

	// Attempt to create tombstones.
	//
	// We can only create tombstones when deleting from a completely full
	// group. This is somewhat unlikely given the 7/8th load factor keeps 1
	// free slot per group in the average case. But if we Put/Delete enough
	// keys we ought to find some that create full groups and thus
	// tombstones.
	//
	// The most obvious way to create lots of tombstones:
	//
	// 1. Add to the map until at lease one group is completely full. This
	//    is likely to occur at load factor ~60%+.
	// 2. Manually clear the map (delete entries one at a time). Every slot
	//    in the full group becomes a tombstone.
	// 3. Repeat to create more tombstones.

	tab := newTestTable[uint32, uint32](capacity)
	usedKeys := make([]uint32, 0, usedLimit)

	r := rand.New(rand.NewPCG(1, 2))

	var tombstones uint64
	for tombstones < wantTombstones {
		// Add to the table until we manage to get a full group. We only use up
		// to half the capacity (to avoid growing), so if we are at the limit,
		// we need to delete a key to try something else.
		//
		// Once we get a full group, we delete one of its keys to create a
		// tombstone.
		count := 0
		for !containsFullNonTombstoneGroup(tab) {
			count++
			if len(usedKeys) >= usedLimit {
				i := r.IntN(usedLimit)
				key := usedKeys[i]
				t.Logf("Delete %#x", key)
				tab.Delete(unsafe.Pointer(&key))
				usedKeys = slices.Delete(usedKeys, i, i+1)
			}

			key := r.Uint32()
			t.Logf("Add %#x", key)
			tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&key))
			usedKeys = append(usedKeys, key)
		}

		t.Logf("Created full group in %d iterations: %v", count, tab)

		// Delete all keys. This will convert every slot in the full group to a
		// tombstone.
		for _, key := range usedKeys {
			t.Logf("Delete %#x", key)
			tab.Delete(unsafe.Pointer(&key))
		}
		usedKeys = usedKeys[0:0:cap(usedKeys)]

		tombstones = countTombstones(tab)
		t.Logf("Created %d tombstones: %v", tombstones, tab)

		if tombstones != tab.tombstones() {
			t.Errorf("tombstones() got %d want %d", tab.tombstones(), tombstones)
		}
	}

	// Now that >1/3rd of the slots are tombstones, grow beyond the load
	// factor to trigger rehashInPlace,
	add := tab.growthLeft + 1
	for i := uint64(0); i < add; i++ {
		key := r.Uint32()
		t.Logf("Add %#x", key)
		tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&key))
		usedKeys = append(usedKeys, key)
	}

	t.Logf("Rehashed: %v", tab)

	if tab.capacity != capacity {
		t.Errorf("capacity after rehash got %d want 0 (rehash in place)", tab.capacity)
	}
	if tab.tombstones() != 0 {
		t.Errorf("tombstones after rehash got %d want 0", tab.tombstones())
	}

	// Verfy values are still correct.
	for _, key := range usedKeys {
		elemPtr, ok := tab.Get(unsafe.Pointer(&key))
		if !ok {
			t.Errorf("tab.Get(%#x) ok got false want true", key)
		}
		elem := *(*uint32)(elemPtr)
		if elem != key {
			t.Errorf("tab.Get(%#x) got %#x want %#x", key, elem, key)
		}
	}
}
