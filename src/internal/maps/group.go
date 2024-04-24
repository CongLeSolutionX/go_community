// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
	"math/bits"
	"unsafe"
)

const (
	// groupSlots is the number of slots in a group.
	groupSlots = 8

	ctrlEmpty   ctrl = 0b10000000
	ctrlDeleted ctrl = 0b11111110

	bitsetLSB     = 0x0101010101010101
	bitsetMSB     = 0x8080808080808080
	bitsetEmpty   = bitsetLSB * uint64(ctrlEmpty)
	bitsetDeleted = bitsetLSB * uint64(ctrlDeleted)
)

// bitset represents a set of slots within a group.
//
// The underlying representation uses one byte per slot, where each byte is
// either 0x80 if the slot is part of the set or 0x00 otherwise. This makes it
// convenient to calculate for an entire group at once (e.g. see matchEmpty).
type bitset uint64

// first assumes that only the MSB of each control byte can be set (e.g. bitset
// is the result of matchEmpty or similar) and returns the relative index of the
// first control byte in the group that has the MSB set.
//
// Returns 8 if the bitset is 0.
// Returns groupSize if the bitset is empty.
func (b bitset) first() uint32 {
	return uint32(bits.TrailingZeros64(uint64(b))) >> 3
}

// removeFirst removes the first set bit (that is, resets the least significant set bit to 0).
func (b bitset) removeFirst() bitset {
	return b & (b - 1)
}

// Each slot in the hash table has a control byte which can have one of three
// states: empty, deleted, and full. They have the following bit patterns:
//
//	  empty: 1 0 0 0 0 0 0 0
//	deleted: 1 1 1 1 1 1 1 0
//	   full: 0 h h h h h h h  // h represents the H1 hash bits
type ctrl uint8

// ctrlGroup is a fixed size array of groupSize control bytes stored in a
// uint64.
type ctrlGroup uint64

// get returns the i-th control byte.
func (g *ctrlGroup) get(i uint32) ctrl {
	return *(*ctrl)(unsafe.Add(unsafe.Pointer(g), i))
}

// set sets the i-th control byte.
func (g *ctrlGroup) set(i uint32, c ctrl) {
	*(*ctrl)(unsafe.Add(unsafe.Pointer(g), i)) = c
}

// setEmpty sets all the control bytes to empty.
func (g *ctrlGroup) setEmpty() {
	*g = ctrlGroup(bitsetEmpty)
}

// matchH2 returns the set of slots which are full and for which the 7-bit hash
// matches the given value. May return false positives.
func (g ctrlGroup) matchH2(h uintptr) bitset {
	// NB: This generic matching routine produces false positive matches when
	// h is 2^N and the control bytes have a seq of 2^N followed by 2^N+1. For
	// example: if ctrls==0x0302 and h=02, we'll compute v as 0x0100. When we
	// subtract off 0x0101 the first 2 bytes we'll become 0xffff and both be
	// considered matches of h. The false positive matches are not a problem,
	// just a rare inefficiency. Note that they only occur if there is a real
	// match and never occur on ctrlEmpty, or ctrlDeleted. The subsequent key
	// comparisons ensure that there is no correctness issue.
	v := uint64(g) ^ (bitsetLSB * uint64(h))
	return bitset(((v - bitsetLSB) &^ v) & bitsetMSB)
}

// matchEmpty returns the set of slots in the group that are empty.
func (g ctrlGroup) matchEmpty() bitset {
	// An empty slot is   1000 0000
	// A deleted slot is  1111 1110
	// A full slot is     0??? ????
	//
	// A slot is empty iff bit 7 is set and bit 1 is not. We could select any
	// of the other bits here (e.g. v << 1 would also work).
	v := uint64(g)
	return bitset((v &^ (v << 6)) & bitsetMSB)
}

// matchEmptyOrDeleted returns the set of slots in the group that are empty or
// deleted.
func (g ctrlGroup) matchEmptyOrDeleted() bitset {
	// An empty slot is  1000 0000
	// A deleted slot is 1111 1110
	// A full slot is    0??? ????
	//
	// A slot is empty or deleted iff bit 7 is set and bit 0 is not.
	v := uint64(g)
	return bitset((v &^ (v << 7)) & bitsetMSB)
}

// group is a wrapper type representing a single slot group stored at data.
//
// A group holds groupSlots slots (key/elem pairs) plus their control word.
type group struct {
	typ *abi.SwissMapType

	// data points to the group, which has layout:
	//
	// type realGroup struct {
	// 	ctrls ctrlGroup
	// 	slots [groupSize]slot
	// }
	//
	// type slot struct {
	// 	key  typ.Key
	// 	elem typ.Elem
	// }
	data unsafe.Pointer // data *realGroup
}

const (
	ctrlGroupsSize   = unsafe.Sizeof(ctrlGroup(0))
	groupSlotsOffset = ctrlGroupsSize
)

// alignUp rounds n up to a multiple of a. a must be a power of 2.
func alignUp(n, a uintptr) uintptr {
	return (n + a - 1) &^ (a - 1)
}

// alignUpPow2 rounds n up to the next power of 2.
func alignUpPow2(n uint64) uint64 {
	v := (uint64(1) << bits.Len64(n-1))
	if v != 0 {
		return v
	}
	return uint64(1) << 63
}

// slotSize returns the size of a slot struct and the offset of elem in a slot.
func slotSize(typ *abi.SwissMapType) (uintptr, uintptr) {
	// Align key size up to elem alignment to account for padding before
	// next field.
	keySize := alignUp(typ.Key.Size_, uintptr(typ.Elem.Align_))
	// Align elem size up to key alignment to account for padding before
	// next slot.
	elemSize := alignUp(typ.Elem.Size_, uintptr(typ.Key.Align_))
	slotSize := keySize + elemSize

	return slotSize, keySize
}

// groupSize returns the size of a group struct.
func groupSize(typ *abi.SwissMapType) uintptr {
	slotSize, _ := slotSize(typ)
	return ctrlGroupsSize + groupSlots*slotSize
}

// ctrls returns the group control word.
func (g *group) ctrls() *ctrlGroup {
	return (*ctrlGroup)(g.data)
}

// key returns a pointer to the key at index i.
func (g *group) key(i uint32) unsafe.Pointer {
	slotSize, _ := slotSize(g.typ)
	offset := groupSlotsOffset + uintptr(i)*slotSize

	return unsafe.Pointer(uintptr(g.data) + offset)
}

// elem returns a pointer to the element at index i.
func (g *group) elem(i uint32) unsafe.Pointer {
	slotSize, elemOff := slotSize(g.typ)
	offset := groupSlotsOffset + uintptr(i)*slotSize + elemOff

	return unsafe.Pointer(uintptr(g.data) + offset)
}

// groups is a wrapper type describing an array of groups stored at data.
type groups struct {
	typ *abi.SwissMapType

	// data points to an array of realGroup. See group above for the
	// definition of realGroup.
	data unsafe.Pointer // data *[length]realGroup

	// length is the number of groups in data. Must be a power of two.
	length uint64
}

// newGroups allocates a new array of length groups.
func newGroups(typ *abi.SwissMapType, length uint64) groups {
	// TODO(prattmic): this is only GC safe as long as key/elem don't
	// contain pointers.
	data := make([]byte, length*uint64(groupSize(typ)))
	return groups{
		typ:    typ,
		data:   unsafe.Pointer(&data[0]),
		length: length,
	}
}

// group returns the group at index i.
func (g *groups) group(i uint64) group {
	// TODO(prattmic): Do something here about truncation on cast to
	// uintptr on 32-bit systems?
	offset := uintptr(i)*groupSize(g.typ)

	return group{
		typ:  g.typ,
		data: unsafe.Pointer(uintptr(g.data) + offset),
	}
}
