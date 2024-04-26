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
	capacity uint64
	// The number of filled slots (i.e. the number of elements in the bucket).
	used uint64
	// The number of slots we can still fill without needing to rehash.
	//
	// This is stored separately due to tombstones: we do not include
	// tombstones in the growth capacity because we'd like to rehash when the
	// table is filled with tombstones as otherwise probe sequences might get
	// unacceptably long without triggering a rehash.
	growthLeft uint64
}

func newTable(mt *abi.SwissMapType, capacity uint64) *table {
	t := &table{
		typ:        mt,
		//TODO
		//seed:       uintptr(fastrand64()),
	}

	if capacity == 0 {
		// No real reason to support zero capacity table, since an
		// empty Map simply won't have a table.
		panic("table must have positive capacity")
	}

	// N.B. group count must be a power of two for probeSeq to visit every
	// group.
	capacity = alignUpPow2(capacity)
	t.reset(capacity)

	return t
}

// reset resets the table with new, empty groups with the specified new total
// capacity.
func (t *table) reset(capacity uint64) {
	if capacity != alignUpPow2(capacity) {
		panic(fmt.Sprintf("capacity must be a power of two. got %#x", capacity))
	}

	groupCount := capacity/groupSlots
	t.groups = newGroups(t.typ, groupCount)
	t.capacity = capacity
	t.resetGrowthLeft()

	for i := uint64(0); i < t.groups.length; i++ {
		g := t.groups.group(i)
		g.ctrls().setEmpty()
	}
}

func (t *table) resetGrowthLeft() {
	var growthLeft uint64
	if t.capacity == 0 {
		growthLeft = 0
	} else if t.capacity <= groupSlots {
		// If the map fits in a single group then we're able to fill all of
		// the slots except 1 (an empty slot is needed to terminate find
		// operations).
		growthLeft = t.capacity - 1
	} else {
		if t.capacity * maxAvgGroupLoad < t.capacity {
			// TODO(prattmic): Do something cleaner.
			panic(fmt.Sprintf("overflow of %d * maxAvgGroupLoad", t.capacity))
		}
		growthLeft = (t.capacity * maxAvgGroupLoad) / groupSlots
	}
	t.growthLeft = growthLeft
}

// Get performs a lookup of the key that key points to. It returns a pointer to
// the element, or false if the key doesn't exist.
func (t *table) Get(key unsafe.Pointer) (unsafe.Pointer, bool) {
	hash := t.typ.Hasher(key, t.seed)

	if debugLog {
		fmt.Printf("Get hash %#x\n", hash)
	}

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
		g := t.groups.group(seq.offset)

		if debugLog {
			fmt.Printf("Get seq group %#x\n", seq.offset)
		}

		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()

			if debugLog {
				fmt.Printf("Get seq group %#x match %d\n", seq.offset, i)
			}

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
			if debugLog {
				fmt.Printf("Get seq group %#x match empty\n", seq.offset)
			}
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
		g := t.groups.group(seq.offset)
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
			// We likely want this optimization back. If we do add
			// it back, we could search for the first deleted slot
			// during the main search, but only use it if we don't
			// find an existing entry.

			if t.growthLeft != 0 {
				panic(fmt.Sprintf("invariant failed: growthLeft is unexpectedly non-zero: %d\n%#v", t.growthLeft, t))
			}

			t.rehash()

			// Note that we don't have to restart the entire Put process as we
			// know the key doesn't exist in the map.
			t.uncheckedPut(hash, key, elem)
			t.used++
			t.checkInvariants()
			return
		}
	}
}

// uncheckedPut inserts an entry known not to be in the table. Used by Put
// after it has failed to find an existing entry to overwrite duration
// insertion.
//
// Updates growthLeft if necessary, but does not update used.
func (t *table) uncheckedPut(hash uintptr, key, elem unsafe.Pointer) {
	if t.growthLeft == 0 {
		panic(fmt.Sprintf("invariant failed: growthLeft is unexpectedly 0\n%#v", t))
	}

	if debugLog {
		fmt.Printf("uncheckedPut hash %#x\n", hash)
	}

	// Given key and its hash hash(key), to insert it, we construct a
	// probeSeq, and use it to find the first group with an unoccupied (empty
	// or deleted) slot. We place the key/value into the first such slot in
	// the group and mark it as full with key's H2.
	seq := makeProbeSeq(h1(hash), t.groups.length-1)
	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)

		if debugLog {
			fmt.Printf("uncheckedPut seq group %#x\n", seq.offset)
		}

		match := g.ctrls().matchEmptyOrDeleted()
		if match != 0 {
			if debugLog {
				fmt.Printf("uncheckedPut seq group %#x match empty/deleted\n", seq.offset)
			}

			i := match.first()

			slotKey := g.key(i)
			typedmemmove(t.typ.Key, slotKey, key)
			slotElem := g.elem(i)
			typedmemmove(t.typ.Elem, slotElem, elem)

			if g.ctrls().get(i) == ctrlEmpty {
				t.growthLeft--
			}
			g.ctrls().set(i, ctrl(h2(hash)))
			return
		}
	}
}

func (t *table) Delete(key unsafe.Pointer) {
	hash := t.typ.Hasher(key, t.seed)

	seq := makeProbeSeq(h1(hash), t.groups.length-1)
	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)
		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()
			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				// TODO(prattmic): Zero the slot? Important for GC!
				t.used--

				// Only a full group can appear in the middle
				// of a probe sequence (a group with at least
				// one empty slot terminates probing). Once a
				// group becomes full, it stays full until
				// rehashing/resizing. So if the group isn't
				// full now, we can simply remove the element.
				// Otherwise, we create a tombstone to mark the
				// slot as deleted.
				if g.ctrls().matchEmpty() != 0 {
					g.ctrls().set(i, ctrlEmpty)
					t.growthLeft++
				} else {
					g.ctrls().set(i, ctrlDeleted)
				}

				t.checkInvariants()
				return
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return
		}
	}
}

// tombstones returns the number of deleted (tombstone) entries in the table. A
// tombstone is a slot that has been deleted but is still considered occupied
// so as not to violate the probing invariant.
func (t *table) tombstones() uint64 {
	return (t.capacity*maxAvgGroupLoad)/groupSlots - t.used - t.growthLeft
}

// Clear deletes all entries from the map resulting in an empty map.
func (t *table) Clear() {
	for i := uint64(0); i < t.groups.length; i++ {
		g := t.groups.group(i)
		g.ctrls().setEmpty()
		for j := uint32(0); j < groupSlots; j++ {
			// TODO(prattmic): Zero the slot? Important for GC!
		}
	}

	t.used = 0
	t.resetGrowthLeft()

	// Reset the hash seed to make it more difficult for attackers to
	// repeatedly trigger hash collisions. See issue
	// https://github.com/golang/go/issues/25237.
	// TODO
	//t.seed = uintptr(fastrand64())
}

func (t *table) rehash() {
	// Rehash in place if we can recover >= 1/3 of the capacity. Note that
	// this heuristic differs from Abseil's and was experimentally determined
	// to balance performance on the PutDelete benchmark vs achieving a
	// reasonable load-factor.
	//
	// Abseil notes that in the worst case it takes ~4 Put/Delete pairs to
	// create a single tombstone. Rehashing in place is significantly faster
	// than resizing because the common case is that elements remain in their
	// current location. The performance of rehashInPlace is dominated by
	// recomputing the hash of every key. We know how much space we're going
	// to reclaim because every tombstone will be dropped and we're only
	// called if we've reached the thresold of capacity/8 empty slots. So the
	// number of tomstones is capacity*7/8 - used.
	if t.capacity > groupSlots && t.tombstones() >= t.capacity/3 {
		t.rehashInPlace()
		return
	}

	// TODO(prattmic): split table

	newCapacity := 2 * t.capacity
	t.resize(newCapacity)
}

// resize the capacity of the table by allocating a bigger array and
// uncheckedPutting each element of the table into the new array (we know that
// no insertion here will Put an already-present value), and discard the old
// backing array.
func (t *table) resize(newCapacity uint64) {
	if debugLog {
		fmt.Printf("Before resize: %s\n", t)
	}

	oldGroups := t.groups
	oldCapacity := t.capacity
	t.reset(newCapacity)

	if oldCapacity > 0 {
		for i := uint64(0); i < oldGroups.length; i++ {
			g := oldGroups.group(i)
			for j := uint32(0); j < groupSlots; j++ {
				if (g.ctrls().get(j) & ctrlEmpty) == ctrlEmpty {
					// Empty or deleted
					continue
				}
				key := g.key(j)
				elem := g.elem(j)
				hash := t.typ.Hasher(key, t.seed)
				t.uncheckedPut(hash, key, elem)
			}
		}
	}

	if debugLog {
		fmt.Printf("After resize: %s\n", t)
	}

	t.checkInvariants()
}

// rehashInPlace reclaimed every deleted slot.
func (t *table) rehashInPlace() {
	if t.capacity == 0 {
		return
	}

	// We want to drop all of the deletes in place. We first walk over the
	// control bytes and mark every DELETED slot as EMPTY and every FULL slot
	// as DELETED. Marking the DELETED slots as EMPTY has effectively dropped
	// the tombstones, but we fouled up the probe invariant. Marking the FULL
	// slots as DELETED gives us a marker to locate the previously FULL slots.

	// Mark all DELETED slots as EMPTY and all FULL slots as DELETED.
	for i := uint64(0); i < t.groups.length; i++ {
		g := t.groups.group(i)
		g.ctrls().convertNonFullToEmptyAndFullToDeleted()
	}

	// Now we walk over all of the DELETED slots (a.k.a. the previously FULL
	// slots). For each slot we find the first probe group we can place the
	// element in, which reestablishes the probe invariant. Note that as this
	// loop proceeds we have the invariant that there are no DELETED slots in
	// the range [0, i). We may move the element at i to the range [0, i) if
	// that is where the first group with an empty slot in its probe chain
	// resides, but we never set a slot in [0, i) to DELETED.
	for i := uint64(0); i < t.groups.length; i++ {
		g := t.groups.group(i)
		for j := uint32(0); j < groupSlots; j++ {
			if g.ctrls().get(j) != ctrlDeleted {
				continue
			}

			key := g.key(j)
			elem := g.elem(j)
			hash := t.typ.Hasher(key, t.seed)
			seq := makeProbeSeq(h1(hash), t.groups.length-1)
			desiredOffset := seq.offset

			var targetGroup group
			var target uint32
			for ; ; seq = seq.next() {
				targetGroup = t.groups.group(seq.offset)
				if match := targetGroup.ctrls().matchEmptyOrDeleted(); match != 0 {
					target = match.first()
					break
				}
			}

			switch {
			case i == desiredOffset:
				// If the target index falls within the first probe group
				// then we don't need to move the element as it already
				// falls in the best probe position.
				g.ctrls().set(j, ctrl(h2(hash)))

			case targetGroup.ctrls().get(target) == ctrlEmpty:
				// The target slot is empty. Transfer the element to the
				// empty slot and mark the slot at index i as empty.
				targetGroup.ctrls().set(target, ctrl(h2(hash)))

				targetKey := targetGroup.key(target)
				typedmemmove(t.typ.Key, targetKey, key)
				targetElem := targetGroup.elem(target)
				typedmemmove(t.typ.Elem, targetElem, elem)

				// Clear old slot.
				// TODO(prattmic): zero old key/elem.
				g.ctrls().set(j, ctrlEmpty)

			case targetGroup.ctrls().get(target) == ctrlDeleted:
				// The slot at target has an element (i.e. it was FULL).
				// We're going to swap our current element with that
				// element and then repeat processing of index j which now
				// holds the element which was at target.
				targetGroup.ctrls().set(target, ctrl(h2(hash)))

				// TODO(prattmic): Put a scratch slot somewhere
				// to avoid allocation here.
				scratchKey := make([]byte, t.typ.Key.Size_)
				scratchElem := make([]byte, t.typ.Elem.Size_)

				targetKey := targetGroup.key(target)
				targetElem := targetGroup.elem(target)

				typedmemmove(t.typ.Key, unsafe.Pointer(&scratchKey[0]), key)
				typedmemmove(t.typ.Elem, unsafe.Pointer(&scratchElem[0]), elem)

				typedmemmove(t.typ.Key, key, targetKey)
				typedmemmove(t.typ.Elem, elem, targetElem)

				typedmemmove(t.typ.Key, targetKey, unsafe.Pointer(&scratchKey[0]))
				typedmemmove(t.typ.Elem, targetElem, unsafe.Pointer(&scratchElem[0]))

				// Repeat processing of the j'th slot which now holds a
				// new key/value.
				j--

			default:
				panic(fmt.Sprintf("ctrl at position %d (%02x) should be empty or deleted",
					target, targetGroup.ctrls().get(target)))
			}
		}
	}

	t.resetGrowthLeft()
	t.growthLeft -= t.used

	t.checkInvariants()
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
	mask   uint64
	offset uint64
	index  uint64
}

func makeProbeSeq(hash uintptr, mask uint64) probeSeq {
	return probeSeq{
		mask:   mask,
		offset: uint64(hash) & mask,
		index:  0,
	}
}

func (s probeSeq) next() probeSeq {
	s.index++
	s.offset = (s.offset + s.index) & s.mask
	return s
}
