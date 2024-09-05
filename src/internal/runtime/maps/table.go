// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps implements Go's builtin map type.
package maps

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

// Maximum size of a table before it is split at the directory level.
//
// TODO: Completely made up value. This should be tuned for performance vs grow
// latency.
// TODO: This should likely be based on byte size, as copying costs will
// dominate grow latency for large objects.
const maxTableCapacity = 1024

// table is a Swiss table hash table structure.
//
// Each table is a complete hash table implementation.
//
// Map uses one or more tables to store entries. Extendible hashing (hash
// prefix) is used to select the table to use for a specific key. Using
// multiple tables enables incremental growth by growing only one table at a
// time.
type table struct {
	// The number of filled slots (i.e. the number of elements in the table).
	used uint64

	typ *abi.SwissMapType

	seed uintptr

	// The number of bits used by directory lookups above this table. Note
	// that this may be less then globalDepth, if the directory has grown
	// but this table has not yet been split.
	localDepth uint32

	// Index of this table in the Map directory. This is the index of the
	// _first_ location in the directory. The table may occur in multiple
	// sequential indicies.
	index int

	// groups is an array of slot groups. Each group holds abi.SwissMapGroupSlots
	// key/elem slots and their control bytes. A table has a fixed size
	// groups array. The table is replaced (in rehash) when more space is
	// required.
	//
	// TODO(prattmic): keys and elements are interleaved to maximize
	// locality, but it comes at the expense of wasted space for some types
	// (consider uint8 key, uint64 element). Consider placing all keys
	// together in these cases to save space.
	//
	// TODO(prattmic): Support indirect keys/values?
	groups groupsReference

	// The total number of slots (always 2^N). Equal to
	// `(groups.lengthMask+1)*abi.SwissMapGroupSlots`.
	capacity uint64
	// The number of slots we can still fill without needing to rehash.
	//
	// This is stored separately due to tombstones: we do not include
	// tombstones in the growth capacity because we'd like to rehash when the
	// table is filled with tombstones as otherwise probe sequences might get
	// unacceptably long without triggering a rehash.
	growthLeft uint64
}

func newTable(mt *abi.SwissMapType, capacity uint64, index int, localDepth uint32) *table {
	if capacity < abi.SwissMapGroupSlots {
		// TODO: temporary until we have a real map type.
		capacity = abi.SwissMapGroupSlots
	}

	t := &table{
		typ: mt,
		//TODO
		//seed:       uintptr(rand()),

		index:      index,
		localDepth: localDepth,
	}

	if capacity == 0 {
		// No real reason to support zero capacity table, since an
		// empty Map simply won't have a table.
		panic("table must have positive capacity")
	}

	if capacity > maxTableCapacity {
		panic("initial table capacity too large")
	}

	// N.B. group count must be a power of two for probeSeq to visit every
	// group.
	capacity, overflow := alignUpPow2(capacity)
	if overflow {
		panic("rounded-up capacity overflows uint64")
	}

	t.reset(capacity)

	return t
}

// reset resets the table with new, empty groups with the specified new total
// capacity.
func (t *table) reset(capacity uint64) {
	ac, overflow := alignUpPow2(capacity)
	if capacity != ac || overflow {
		panic("capacity must be a power of two")
	}

	groupCount := capacity / abi.SwissMapGroupSlots
	t.groups = newGroups(t.typ, groupCount)
	t.capacity = capacity
	t.resetGrowthLeft()

	for i := uint64(0); i <= t.groups.lengthMask; i++ {
		g := t.groups.group(i)
		g.ctrls().setEmpty()
	}
}

func (t *table) resetGrowthLeft() {
	var growthLeft uint64
	if t.capacity == 0 {
		growthLeft = 0
	} else if t.capacity <= abi.SwissMapGroupSlots {
		// If the map fits in a single group then we're able to fill all of
		// the slots except 1 (an empty slot is needed to terminate find
		// operations).
		//
		// TODO(go.dev/issue/54766): With a special case in probing for
		// single-group tables, we could fill all slots.
		growthLeft = t.capacity - 1
	} else {
		if t.capacity*maxAvgGroupLoad < t.capacity {
			// TODO(prattmic): Do something cleaner.
			panic("overflow")
		}
		growthLeft = (t.capacity * maxAvgGroupLoad) / abi.SwissMapGroupSlots
	}
	t.growthLeft = growthLeft
}

func (t *table) Used() uint64 {
	return t.used
}

// Get performs a lookup of the key that key points to. It returns a pointer to
// the element, or false if the key doesn't exist.
func (t *table) Get(key unsafe.Pointer) (unsafe.Pointer, bool) {
	hash := t.typ.Hasher(key, t.seed)
	_, elem, ok := t.getWithKey(hash, key)
	return elem, ok
}

// getWithKey performs a lookup of key, returning a pointer to the version of
// the key in the map in addition to the element.
//
// This is relevant when multiple different key values compare equal (e.g.,
// +0.0 and -0.0). When a grow occurs during iteration, iteration perform a
// lookup of keys from the old group in the new group in order to correctly
// expose updated elements. For NeedsKeyUpdate keys, iteration also must return
// the new key value, not the old key value.
//
// hash must be the hash of the key.
func (t *table) getWithKey(hash uintptr, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
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
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)

		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()

			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				return slotKey, g.elem(i), true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return nil, nil, false
		}
	}
}

// PutSlot returns a pointer to the element slot where an inserted element
// should be written, and ok if it returned a valid slot.
//
// PutSlot returns ok false if the table was split and the Map needs to find
// the new table.
//
// hash must be the hash of key.
func (t *table) PutSlot(m *Map, hash uintptr, key unsafe.Pointer) (unsafe.Pointer, bool) {
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	//startOffset := seq.offset

	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)
		match := g.ctrls().matchH2(h2(hash))

		// Look for an existing slot containing this key.
		for match != 0 {
			i := match.first()

			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				if t.typ.NeedKeyUpdate() {
					typedmemmove(t.typ.Key, slotKey, key)
				}

				slotElem := g.elem(i)

				t.checkInvariants()
				return slotElem, true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.

			// If there is room left to grow, just insert the new entry.
			if t.growthLeft > 0 {
				i := match.first()

				slotKey := g.key(i)
				typedmemmove(t.typ.Key, slotKey, key)
				slotElem := g.elem(i)

				g.ctrls().set(i, ctrl(h2(hash)))
				t.growthLeft--
				t.used++
				m.used++

				t.checkInvariants()
				return slotElem, true
			}

			// TODO(prattmic): While searching the probe sequence,
			// we may have passed deleted slots which we could use
			// for this entry.
			//
			// At the moment, we leave this behind for
			// rehash to free up.
			//
			// cockroachlabs/swiss restarts search of the probe
			// sequence for a deleted slot.
			//
			// TODO(go.dev/issue/54766): We want this optimization
			// back. We could search for the first deleted slot
			// during the main search, but only use it if we don't
			// find an existing entry.

			t.rehash(m)
			return nil, false
		}
	}
}

// uncheckedPutSlot inserts an entry known not to be in the table, returning an
// entry to the element slot where the element should be written. Used by
// PutSlot after it has failed to find an existing entry to overwrite duration
// insertion.
//
// Updates growthLeft if necessary, but does not update used.
//
// Never returns nil.
func (t *table) uncheckedPutSlot(hash uintptr, key unsafe.Pointer) unsafe.Pointer {
	if t.growthLeft == 0 {
		panic("invariant failed: growthLeft is unexpectedly 0")
	}

	// Given key and its hash hash(key), to insert it, we construct a
	// probeSeq, and use it to find the first group with an unoccupied (empty
	// or deleted) slot. We place the key/value into the first such slot in
	// the group and mark it as full with key's H2.
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)

		match := g.ctrls().matchEmptyOrDeleted()
		if match != 0 {
			i := match.first()

			slotKey := g.key(i)
			typedmemmove(t.typ.Key, slotKey, key)
			slotElem := g.elem(i)

			if g.ctrls().get(i) == ctrlEmpty {
				t.growthLeft--
			}
			g.ctrls().set(i, ctrl(h2(hash)))
			return slotElem
		}
	}
}

func (t *table) Delete(m *Map, key unsafe.Pointer) {
	hash := t.typ.Hasher(key, t.seed)

	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(seq.offset)
		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()
			slotKey := g.key(i)
			if t.typ.Key.Equal(key, slotKey) {
				t.used--
				m.used--

				typedmemclr(t.typ.Key, slotKey)
				typedmemclr(t.typ.Elem, g.elem(i))

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
	return (t.capacity*maxAvgGroupLoad)/abi.SwissMapGroupSlots - t.used - t.growthLeft
}

// Clear deletes all entries from the map resulting in an empty map.
func (t *table) Clear() {
	for i := uint64(0); i <= t.groups.lengthMask; i++ {
		g := t.groups.group(i)
		typedmemclr(t.typ.Group, g.data)
		g.ctrls().setEmpty()
	}

	t.used = 0
	t.resetGrowthLeft()

	// Reset the hash seed to make it more difficult for attackers to
	// repeatedly trigger hash collisions. See issue
	// https://github.com/golang/go/issues/25237.
	// TODO
	//t.seed = uintptr(rand())
}

type Iter struct {
	key  unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go).
	elem unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go).
	typ  *abi.SwissMapType
	m    *Map

	// Randomize iteration order by starting iteration at a random slot
	// offset. The same offset is used to offset the table, group, and slot
	// selection.
	offset uint64

	// Snapshot of Map.clearSeq at iteration initialization time. Used to
	// detect clear during iteration.
	clearSeq uint64

	// Value of Map.globalDepth during the last call to Next. Used to
	// detect directory grow during iteration.
	globalDepth uint32

	// dirIdx is the current directory index, prior to adjustment by offset.
	dirIdx int

	// tab is the table at dirIdx during the previous call to Next.
	tab *table

	// TODO: these could be merged into a single counter (and pre-offset
	// with offset).
	groupIdx uint64
	slotIdx  uint32

	//// Snapshot of the groups at iteration initialization time. If the
	//// table resizes during iteration, we continue to iterate over the old
	//// groups.
	////
	//// If the table grows we must consult the updated table to observe
	//// changes, though we continue to use the snapshot to determine order
	//// and avoid duplicating results.
	//groups groupsReference

	// XXX: compiler wants this a multiple of word size.
	XXPad uint32
}

// Init initializes Iter for iteration.
func (it *Iter) Init(typ *abi.SwissMapType, m *Map) {
	it.typ = typ
	if m == nil || m.used == 0 {
		return
	}

	// N.B. it may be allocated on the stack and not zeroed. Be sure to
	// zero all fields.
	// XXX: double check if this is true.
	*it = Iter{
		typ:         m.typ,
		m:           m,
		offset:      rand(),
		globalDepth: m.globalDepth,
		clearSeq:    m.clearSeq,
	}
}

func (it *Iter) Initialized() bool {
	return it.typ != nil
}

// Map returns the map this iterator is iterating over.
func (it *Iter) Map() *Map {
	return it.m
}

// Key returns a pointer to the current key. nil indicates end of iteration.
//
// Must not be called prior to Next.
func (it *Iter) Key() unsafe.Pointer {
	return it.key
}

// Key returns a pointer to the current element. nil indicates end of
// iteration.
//
// Must not be called prior to Next.
func (it *Iter) Elem() unsafe.Pointer {
	return it.elem
}

// Next proceeds to the next element in iteration, which can be accessed via
// the Key and Elem methods.
//
// The table can be mutated during iteration, though there is no guarantee that
// the mutations will be visible to the iteration.
//
// Init must be called prior to Next.
func (it *Iter) Next() {
	if it.m == nil {
		// Map was empty at Iter.Init.
		it.key = nil
		it.elem = nil
		return
	}

	if it.globalDepth != it.m.globalDepth {
		// Directory has grown since the last call to Next. Adjust our
		// directory index.
		it.dirIdx *= 1 << (it.m.globalDepth - it.globalDepth)

		it.globalDepth = it.m.globalDepth
	}

	// Continue iteration until we find a full slot.
	for ; it.dirIdx < len(it.m.directory); {
		// N.B. This computation would be inconsistent when the
		// directory grows, but if the directory grows we will always
		// skip the new entries, so it doesn't matter.
		dirIdx := int((uint64(it.dirIdx)+it.offset) % uint64(len(it.m.directory)))
		newTab := it.m.directory[dirIdx]
		if newTab.index != dirIdx {
			// Skip duplicate entries to the same table.
			it.dirIdx++
			continue
		}
		if it.tab == nil {
			it.tab = newTab
		}

		// N.B. Use it.tab, not newTab. It is important to use the old
		// table for key selection if the table has grown. See comment
		// on grown below.
		for ; it.groupIdx <= it.tab.groups.lengthMask; it.groupIdx++ {
			g := it.tab.groups.group((it.groupIdx + it.offset) & it.tab.groups.lengthMask)

			// TODO(prattmic): Skip over groups that are composed of only empty
			// or deleted slots using matchEmptyOrDeleted() and counting the
			// number of bits set.
			for ; it.slotIdx < abi.SwissMapGroupSlots; it.slotIdx++ {
				k := (it.slotIdx + uint32(it.offset)) % abi.SwissMapGroupSlots

				if (g.ctrls().get(k) & ctrlEmpty) == ctrlEmpty {
					// Empty or deleted.
					continue
				}

				key := g.key(k)

				// If the table has changed since the last
				// call, then it has grown or split. In this
				// case, further mutations (changes to
				// key->elem or deletions) will not be visible
				// in our snapshot table. Instead we must
				// consult the new table by doing a full
				// lookup.
				//
				// We still use our old table to decide which
				// keys to lookup in order to avoid returning
				// the same key twice.
				//
				// TODO(prattmic): Rather than growing t.groups
				// directly, a cleaner design may be to always
				// create a new table on grow or split, leaving
				// behind 1 or 2 forwarding pointers. This lets
				// us handle this update after grow problem the
				// same way both within a single table and
				// across split.
				// TODO: remove TODO
				//grown := it.groups.data != it.tab.groups.data
				grown := it.tab != newTab
				var elem unsafe.Pointer
				if grown {
					var ok bool
					newKey, newElem, ok := it.m.getWithKey(key)
					if !ok {
						// Key has likely been deleted, and
						// should be skipped.
						//
						// One exception is keys that don't
						// compare equal to themselves (e.g.,
						// NaN). These keys cannot be looked
						// up, so getWithKey will fail even if
						// the key exists.
						//
						// However, we are in luck because such
						// keys cannot be updated and they
						// cannot be deleted except with clear.
						// Thus if no clear has occurted, the
						// key/elem must still exist exactly as
						// in the old groups, so we can return
						// them from there.
						//
						// TODO(prattmic): Consider checking
						// clearSeq early. If a clear occurred,
						// Next could always return
						// immediately, as iteration doesn't
						// need to return anything added after
						// clear.
						if it.clearSeq == it.m.clearSeq && !it.m.typ.Key.Equal(key, key) {
							elem = g.elem(k)
						} else {
							continue
						}
					} else {
						key = newKey
						elem = newElem
					}
				} else {
					elem = g.elem(k)
				}

				it.slotIdx++
				if it.slotIdx >= abi.SwissMapGroupSlots {
					it.groupIdx++
					it.slotIdx = 0
				}
				it.key = key
				it.elem = elem
				return
			}
			it.slotIdx = 0
		}
		it.groupIdx = 0
		if it.tab != newTab {
			// Table grew or split since we starting iteration over
			// it. If it split, we must skip the right split as
			// well.
			entries := 1 << (it.m.globalDepth - it.tab.localDepth) // entries in the pre-split table.
			it.dirIdx += entries
		} else {
			// TODO: do entries skip here too rather than continue
			// at the top of the loop.
			it.dirIdx++
		}
		it.tab = nil
	}

	it.key = nil
	it.elem = nil
	return
}

// Replaces the table with one larger table or two split tables to fit more
// entries. table is no longer usable.
func (t *table) rehash(m *Map) {
	// TODO(prattmic): SwissTables typically perform a "rehash in place"
	// operation which recovers capacity consumed by tombstones without growing
	// the table by reordering slots as necessary to maintain the probe
	// invariant while eliminating all tombstones.
	//
	// However, it is unclear how to make rehash in place work with
	// iteration. Since iteration simply walks through all slots in order
	// (with random start offset), reordering the slots would break
	// iteration.
	//
	// As an alternative, we could do a "resize" to new groups allocation
	// of the same size. This would eliminate the tombstones, but using a
	// new allocation, so the existing grow support in iteration would
	// continue to work.

	// TODO(prattmic): split table
	// TODO(prattmic): Avoid overflow (splitting the table will achieve this)

	newCapacity := 2 * t.capacity
	if newCapacity <= maxTableCapacity {
		t.grow(m, newCapacity)
		return
	}

	t.split(m)
}

// Bitmask for the last selection bit at this depth.
func localDepthMask(localDepth uint32) uintptr {
	if goarch.PtrSize == 4 {
		return uintptr(1) << (32 - localDepth)
	}
	return uintptr(1) << (64 - localDepth)
}

// split the table into two, installing the new tables in the map directory.
func (t *table) split(m *Map) {
	localDepth := t.localDepth
	localDepth++

	// TODO: is this the best capacity?
	left := newTable(t.typ, maxTableCapacity, -1, localDepth)
	right := newTable(t.typ, maxTableCapacity, -1, localDepth)

	// Split in half at the localDepth bit from the top.
	mask := localDepthMask(localDepth)

	for i := uint64(0); i <= t.groups.lengthMask; i++ {
		g := t.groups.group(i)
		for j := uint32(0); j < abi.SwissMapGroupSlots; j++ {
			if (g.ctrls().get(j) & ctrlEmpty) == ctrlEmpty {
				// Empty or deleted
				continue
			}
			key := g.key(j)
			elem := g.elem(j)
			hash := t.typ.Hasher(key, t.seed)
			var newTable *table
			if hash&mask == 0 {
				newTable = left
			} else {
				newTable = right
			}
			slotElem := newTable.uncheckedPutSlot(hash, key)
			typedmemmove(newTable.typ.Elem, slotElem, elem)
			newTable.used++
		}
	}

	m.installTableSplit(t, left, right)
}

// grow the capacity of the table by allocating a new table with a bigger array
// and uncheckedPutting each element of the table into the new table (we know
// that no insertion here will Put an already-present value), and discard the
// old table.
func (t *table) grow(m *Map, newCapacity uint64) {
	newTable := newTable(t.typ, newCapacity, t.index, t.localDepth)

	if t.capacity > 0 {
		for i := uint64(0); i <= t.groups.lengthMask; i++ {
			g := t.groups.group(i)
			for j := uint32(0); j < abi.SwissMapGroupSlots; j++ {
				if (g.ctrls().get(j) & ctrlEmpty) == ctrlEmpty {
					// Empty or deleted
					continue
				}
				key := g.key(j)
				elem := g.elem(j)
				hash := newTable.typ.Hasher(key, t.seed)
				slotElem := newTable.uncheckedPutSlot(hash, key)
				typedmemmove(newTable.typ.Elem, slotElem, elem)
				newTable.used++
			}
		}
	}

	newTable.checkInvariants()
	m.replaceTable(newTable)
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
