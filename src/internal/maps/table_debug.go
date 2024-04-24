// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps implements Go's builtin map type.
package maps

import (
	"fmt"
	"unsafe"
)

const debugLog = false

func (t *table) checkInvariants() {
	// For every non-empty slot, verify we can retrieve the key using Get.
	// Count the number of used and deleted slots.
	var used uint64
	var deleted uint64
	var empty uint64
	for i := uint64(0); i < t.groups.length; i++ {
		g := t.groups.group(i)
		for j := uint32(0); j < groupSlots; j++ {
			c := g.ctrls().get(j)
			switch {
			case c == ctrlDeleted:
				deleted++
			case c == ctrlEmpty:
				empty++
			default:
				key := g.key(j)
				if _, ok := t.Get(key); !ok {
					hash := t.typ.Hasher(key, t.seed)
					panic(fmt.Sprintf("invariant failed: slot(%d/%d): key %v not found [h2=%02x h1=%07x]\n%#v",
					i, j, hexdump(key, t.typ.Key.Size_), h2(hash), h1(hash), t))
				}
				used++
			}
		}
	}

	if used != t.used {
		panic(fmt.Sprintf("invariant failed: found %d used slots, but used count is %d\n%#v",
		used, t.used, t))
	}

	//growthLeft := (b.capacity*maxAvgGroupLoad)/groupSize - b.used - deleted
	//if growthLeft != b.growthLeft {
	//	panic(fmt.Sprintf("invariant failed: found %d growthLeft, but expected %d\n%#v",
	//	b.growthLeft, growthLeft, b))
	//}
	//if deleted != b.tombstones() {
	//	panic(fmt.Sprintf("invariant failed: found %d tombstones, but expected %d\n%#v",
	//	deleted, b.tombstones(), b))
	//}

	if empty == 0 {
		panic(fmt.Sprintf("invariant failed: found no empty slots (violates probe invariant)\n%#v", t))
	}
}

func (t *table) String() string {
	s := fmt.Sprintf(`table{
	seed: %#x
	capacity: %d
	growthLeft: %d
	groups:
`, t.seed, t.capacity, t.growthLeft)

	for i := uint64(0); i < t.groups.length; i++ {
		s += fmt.Sprintf("\t\tgroup %#x\n", i)

		g := t.groups.group(i)
		ctrls := g.ctrls()
		for j := uint32(0); j < groupSlots; j++ {
			s += fmt.Sprintf("\t\t\tslot %d\n", j)

			c := ctrls.get(j)
			s += fmt.Sprintf("\t\t\t\tctrl %#x", c)
			switch c {
			case ctrlEmpty:
				s += fmt.Sprintf(" (empty)\n")
			case ctrlDeleted:
				s += fmt.Sprintf(" (deleted)\n")
			default:
				s += fmt.Sprintf("\n")
			}

			s += fmt.Sprintf("\t\t\t\tkey  %s\n", hexdump(g.key(j), t.typ.Key.Size_))
			s += fmt.Sprintf("\t\t\t\telem %s\n", hexdump(g.elem(j), t.typ.Elem.Size_))
		}
	}

	return s
}

func hexdump(ptr unsafe.Pointer, size uintptr) string {
	var s string
	for size > 0 {
		s += fmt.Sprintf("%#x ", *(*byte)(ptr))
		ptr = unsafe.Pointer(uintptr(ptr) + 1)
		size--
	}
	return s
}
