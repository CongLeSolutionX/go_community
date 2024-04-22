// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps implements Go's builtin map type.
package maps

import (
	"fmt"
	"internal/abi"
	"unsafe"
)

// table is a Swiss table hash table structure.
//
// Each table is a complete hash table implementation.
//
// Map uses one or more tables to store entries. Extendible hashing (hash
// prefix) is used to select the table to use for a specific key. Using
// multiple tables enables incremental growth by growing only one table at a
// time.
type table struct {
	typ *abi.SwissMapType

	seed uintptr

	// groups is an array of slot groups. Each group holds groupSlots
	// key/elem slots and their control bytes.
	//
	// TODO(prattmic): keys and elements are interleaved to maximize
	// locality, but it comes at the expense of wasted space for some types
	// (consider uint8 key, uint64 element). Consider placing all keys
	// together in these cases to save space.
	//
	// TODO(prattmic): Support indirect keys/values?
	groups groups

	// The total number of slots (always 2^N). Equal to
	// `groups.length*groupSlots`.
	capacity uint32
	// The number of filled slots (i.e. the number of elements in the bucket).
	used uint32
	// The number of slots we can still fill without needing to rehash.
	//
	// This is stored separately due to tombstones: we do not include
	// tombstones in the growth capacity because we'd like to rehash when the
	// table is filled with tombstones as otherwise probe sequences might get
	// unacceptably long without triggering a rehash.
	growthLeft uint32
}

func newTable(mt *abi.SwissMapType, capacity uint32) *table {
	// N.B. group count must be a power of two for probeSeq to visit every
	// group.
	capacity = alignUpPow2(capacity)
	groupCount := capacity/groupSlots

	t := &table{
		typ:        mt,
		seed:       uintptr(fastrand64()),
		groups:     newGroups(mt, groupCount),
		capacity:   capacity,
		growthLeft: capacity-1,
	}

	for i := uint32(0); i < t.groups.length; i++ {
		g := t.groups.group(uint64(i))
		g.ctrls().setEmpty()
	}

	return t
}

// Get performs a lookup of the key that key points to. It returns a pointer to
// the element, or false if the key doesn't exist.
func (t *table) Get(key unsafe.Pointer) (unsafe.Pointer, bool)  {
	hash := t.typ.Hasher(key, t.seed)

	// To find the location of a key in the table, we compute hash(key). From
	// h1(hash(key)) and the capacity, we construct a probeSeq that visits
	// every group of slots in some interesting order.
	//
	// We walk through these indices. At each index, we select the entire group
	// starting with that index and extract potential candidates: occupied slots
	// with a control byte equal to h2(hash(key)). If we find an empty slot in the
	// group, we stop and return an error. The key at candidate slot y is compared
	// with key; if key == m.slots[y].key we are done and return y; otherwise we
	// continue to the next probe index. Tombstones (ctrlDeleted) effectively
	// behave like full slots that never match the value we're looking for.
	//
	// The h2 bits ensure when we compare a key we are likely to have actually
	// found the object. That is, the chance is low that keys compare false. Thus,
	// when we search for an object, we are unlikely to call == many times. This
	// likelyhood can be analyzed as follows (assuming that h2 is a random enough
	// hash function).
	//
	// Let's assume that there are k "wrong" objects that must be examined in a
	// probe sequence. For example, when doing a find on an object that is in the
	// table, k is the number of objects between the start of the probe sequence
	// and the final found object (not including the final found object). The
	// expected number of objects with an h2 match is then k/128. Measurements and
	// analysis indicate that even at high load factors, k is less than 32,
	// meaning that the number of false positive comparisons we must perform is
	// less than 1/8 per find.
	seq := makeProbeSeq(h1(hash), t.groups.length-1)
	for ; ; seq = seq.next() {
		g := t.groups.group(uint64(seq.offset))
		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()
			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				return g.elem(i), true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return nil, false
		}
	}
}

func (t *table) Put(key, elem unsafe.Pointer) {
	hash := t.typ.Hasher(key, t.seed)

	if debugLog {
		fmt.Printf("Put hash %#x\n", hash)
	}

	seq := makeProbeSeq(h1(hash), t.groups.length-1)
	//startOffset := seq.offset

	for ; ; seq = seq.next() {
		g := t.groups.group(uint64(seq.offset))
		match := g.ctrls().matchH2(h2(hash))

		if debugLog {
			fmt.Printf("Put seq group %#x\n", seq.offset)
		}

		// Look for an existing slot containing this key.
		for match != 0 {
			i := match.first()
			if debugLog {
				fmt.Printf("Put seq group %#x match %d\n", seq.offset, i)
			}

			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				slotElem := g.elem(i)
				typedmemmove(t.typ.Elem, slotElem, elem)

				t.checkInvariants()
				return
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			if debugLog {
				fmt.Printf("Put seq group %#x match empty\n", seq.offset)
			}

			// If there is room left to grow, just insert the new entry.
			if t.growthLeft > 0 {
				i := match.first()

				slotKey := g.key(i)
				typedmemmove(t.typ.Key, slotKey, key)
				slotElem := g.elem(i)
				typedmemmove(t.typ.Elem, slotElem, elem)

				g.ctrls().set(i, ctrl(h2(hash)))
				t.growthLeft--
				t.used++

				t.checkInvariants()
				return
			}

			// TODO(prattmic): While searching the probe sequence,
			// we may have passed deleted slots which we could use
			// for this entry.
			//
			// At the moment, we leave this behind for
			// rehashInPlace to free up.
			//
			// cockroachlabs/swiss restarts search of the probe
			// sequence for a deleted slot.
			//
			// We don't do that, instead leaving it for
			// rehashInPlace. We likely want this optimization
			// back. If we do add it back, we could search for the
			// first deleted slot during the main search, but only
			// use it if we don't find an existing entry.

			if t.growthLeft != 0 {
				panic(fmt.Sprintf("invariant failed: growthLeft is unexpectedly non-zero: %d\n%#v", t.growthLeft, t))
			}

			panic("grow unimplemented")
		}
	}
}

// probeSeq maintains the state for a probe sequence that iterates through the
// groups in a table. The sequence is a triangular progression of the form
//
//	p(i) := (i^2 + i)/2 + hash (mod mask+1)
//
// The sequence effectively outputs the indexes of *groups*. The group
// machinery allows us to check an entire group with minimal branching.
//
// It turns out that this probe sequence visits every group exactly once if
// the number of groups is a power of two, since (i^2+i)/2 is a bijection in
// Z/(2^m). See https://en.wikipedia.org/wiki/Quadratic_probing
type probeSeq struct {
	mask   uint32
	offset uint32
	index  uint32
}

func makeProbeSeq(hash uintptr, mask uint32) probeSeq {
	return probeSeq{
		mask:   mask,
		offset: uint32(hash) & mask,
		index:  0,
	}
}

func (s probeSeq) next() probeSeq {
	s.index++
	s.offset = (s.offset + s.index) & s.mask
	return s
}
