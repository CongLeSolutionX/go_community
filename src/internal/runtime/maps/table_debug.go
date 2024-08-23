// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps implements Go's builtin map type.
package maps

import (
	"internal/abi"
	"unsafe"
)

const debugLog = false

func (t *table) checkInvariants() {
	if !debugLog {
		return
	}

	// For every non-empty slot, verify we can retrieve the key using Get.
	// Count the number of used and deleted slots.
	var used uint64
	var deleted uint64
	var empty uint64
	for i := uint64(0); i <= t.groups.lengthMask; i++ {
		g := t.groups.group(t.typ, i)
		for j := uint32(0); j < abi.SwissMapGroupSlots; j++ {
			c := g.ctrls().get(j)
			switch {
			case c == ctrlDeleted:
				deleted++
			case c == ctrlEmpty:
				empty++
			default:
				used++

				key := g.key(t.typ, j)

				// Can't lookup NaN.
				if isNaN(t.typ.Key, key) {
					continue
				}

				if _, ok := t.Get(key); !ok {
					//hash := t.typ.Hasher(key, t.seed)
					//panic(fmt.Sprintf("invariant failed: slot(%d/%d): key %v not found [hash=%#x, h2=%#02x h1=%#07x]\n%v",
					//i, j, hexdump(key, t.typ.Key.Size_), hash, h2(hash), h1(hash), t))
					panic("invariant failed: slot: key not found")
				}
			}
		}
	}

	if used != t.used {
		//panic(fmt.Sprintf("invariant failed: found %d used slots, but used count is %d\n%v",
		//used, t.used, t))
		panic("invariant failed: found mismatched used slot count")
	}

	growthLeft := (t.capacity*maxAvgGroupLoad)/abi.SwissMapGroupSlots - t.used - deleted
	if growthLeft != t.growthLeft {
		//panic(fmt.Sprintf("invariant failed: found %d growthLeft, but expected %d\n%v",
		//t.growthLeft, growthLeft, t))
		panic("invariant failed: found mismatched growthLeft")
	}
	if deleted != t.tombstones() {
		//panic(fmt.Sprintf("invariant failed: found %d tombstones, but expected %d\n%v",
		//deleted, t.tombstones(), t))
		panic("invariant failed: found mismatched tombstones")
	}

	if empty == 0 {
		//panic(fmt.Sprintf("invariant failed: found no empty slots (violates probe invariant)\n%v", t))
		panic("invariant failed: found no empty slots (violates probe invariant)")
	}
}

func isNaN(typ *abi.Type, ptr unsafe.Pointer) bool {
	var val float64

	switch typ.Kind() {
	case abi.Float32:
		val32 := *(*float32)(ptr)
		val = float64(val32)
	case abi.Float64:
		val = *(*float64)(ptr)
	default:
		// TODO(prattmic): handle aggregates containing floats. e.g.,
		// struct{float64;int} with a NaN.
		return false
	}

	// From math.IsNaN: IEEE 754 says that only NaNs satisfy f != f.
	return val != val
}

//func (t *table) String() string {
//	s := fmt.Sprintf(`table{
//	seed: %#x
//	capacity: %d
//	used: %d
//	growthLeft: %d
//	groups:
//`, t.seed, t.capacity, t.used, t.growthLeft)
//
//	for i := uint64(0); i <= t.groups.lengthMask; i++ {
//		s += fmt.Sprintf("\t\tgroup %#x\n", i)
//
//		g := t.groups.group(i)
//		ctrls := g.ctrls()
//		for j := uint32(0); j < abi.SwissMapGroupSlots; j++ {
//			s += fmt.Sprintf("\t\t\tslot %d\n", j)
//
//			c := ctrls.get(j)
//			s += fmt.Sprintf("\t\t\t\tctrl %#x", c)
//			switch c {
//			case ctrlEmpty:
//				s += fmt.Sprintf(" (empty)\n")
//			case ctrlDeleted:
//				s += fmt.Sprintf(" (deleted)\n")
//			default:
//				s += fmt.Sprintf("\n")
//			}
//
//			s += fmt.Sprintf("\t\t\t\tkey  %s\n", hexdump(g.key(j), t.typ.Key.Size_))
//			s += fmt.Sprintf("\t\t\t\telem %s\n", hexdump(g.elem(j), t.typ.Elem.Size_))
//		}
//	}
//
//	return s
//}
//
//func hexdump(ptr unsafe.Pointer, size uintptr) string {
//	var s string
//	for size > 0 {
//		s += fmt.Sprintf("%#x ", *(*byte)(ptr))
//		ptr = unsafe.Pointer(uintptr(ptr) + 1)
//		size--
//	}
//	return s
//}
