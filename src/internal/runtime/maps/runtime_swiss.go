// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swissmap

package maps

import (
	"internal/abi"
	"internal/asan"
	"internal/msan"
	"internal/race"
	"internal/runtime/sys"
	"unsafe"
)

// Functions below pushed from runtime.

//go:linkname mapKeyError
func mapKeyError(typ *abi.SwissMapType, p unsafe.Pointer) error

// Pull from runtime. It is important that is this the exact same copy as the
// runtime because runtime.mapaccess1_fat compares the returned pointer with
// &runtime.zeroVal[0].
// TODO: move zeroVal to internal/abi?
//
//go:linkname zeroVal runtime.zeroVal
var zeroVal [abi.ZeroValSize]byte

// mapaccess1 returns a pointer to h[key].  Never returns nil, instead
// it will return a reference to the zero object for the elem type if
// the key is not in the map.
// NOTE: The returned pointer may keep the whole map live, so don't
// hold onto it for very long.
//
//go:linkname runtime_mapaccess1 runtime.mapaccess1
func runtime_mapaccess1(typ *abi.SwissMapType, m *Map, key unsafe.Pointer) unsafe.Pointer {
	if race.Enabled && m != nil {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapaccess1)
		race.ReadObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
		race.ReadObjectPC(typ.Key, key, callerpc, pc)
	}
	if msan.Enabled && m != nil {
		msan.Read(key, typ.Key.Size_)
	}
	if asan.Enabled && m != nil {
		asan.Read(key, typ.Key.Size_)
	}

	if m == nil || m.Used() == 0 {
		if err := mapKeyError(typ, key); err != nil {
			panic(err) // see issue 23734
		}
		return unsafe.Pointer(&zeroVal[0])
	}

	if m.writing != 0 {
		fatal("concurrent map read and map write")
	}

	hash := typ.Hasher(key, m.seed)

	if m.dirLen <= 0 {
		_, elem, ok := m.getWithKeySmall(hash, key)
		if !ok {
			return unsafe.Pointer(&zeroVal[0])
		}
		return elem
	}

	// Select table.
	idx := m.directoryIndex(hash)
	t := m.directoryAt(idx)

	// Probe table.
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(typ, seq.offset)

		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()

			slotKey := g.key(typ, i)
			if typ.IndirectKey() {
				slotKey = *((*unsafe.Pointer)(slotKey))
			}
			if typ.Key.Equal(key, slotKey) {
				slotElem := g.elem(typ, i)
				if typ.IndirectElem() {
					slotElem = *((*unsafe.Pointer)(slotElem))
				}
				return slotElem
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return unsafe.Pointer(&zeroVal[0])
		}
	}
}

//go:linkname runtime_mapaccess2 runtime.mapaccess2
func runtime_mapaccess2(typ *abi.SwissMapType, m *Map, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if race.Enabled && m != nil {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapaccess1)
		race.ReadObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
		race.ReadObjectPC(typ.Key, key, callerpc, pc)
	}
	if msan.Enabled && m != nil {
		msan.Read(key, typ.Key.Size_)
	}
	if asan.Enabled && m != nil {
		asan.Read(key, typ.Key.Size_)
	}

	if m == nil || m.Used() == 0 {
		if err := mapKeyError(typ, key); err != nil {
			panic(err) // see issue 23734
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}

	if m.writing != 0 {
		fatal("concurrent map read and map write")
	}

	hash := typ.Hasher(key, m.seed)

	if m.dirLen <= 0 {
		_, elem, ok := m.getWithKeySmall(hash, key)
		if !ok {
			return unsafe.Pointer(&zeroVal[0]), false
		}
		return elem, true
	}

	// Select table.
	idx := m.directoryIndex(hash)
	t := m.directoryAt(idx)

	// Probe table.
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(typ, seq.offset)

		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()

			slotKey := g.key(typ, i)
			if typ.IndirectKey() {
				slotKey = *((*unsafe.Pointer)(slotKey))
			}
			if typ.Key.Equal(key, slotKey) {
				slotElem := g.elem(typ, i)
				if typ.IndirectElem() {
					slotElem = *((*unsafe.Pointer)(slotElem))
				}
				return slotElem, true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return unsafe.Pointer(&zeroVal[0]), false
		}
	}
}

//go:linkname runtime_mapassign runtime.mapassign
func runtime_mapassign(typ *abi.SwissMapType, m *Map, key unsafe.Pointer) unsafe.Pointer {
	if m == nil {
		// XXX
		//panic(plainError("assignment to entry in nil map"))
		panic("assignment to entry in nil map")
	}
	if race.Enabled {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapassign)
		race.WriteObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
		race.ReadObjectPC(typ.Key, key, callerpc, pc)
	}
	if msan.Enabled {
		msan.Read(key, typ.Key.Size_)
	}
	if asan.Enabled {
		asan.Read(key, typ.Key.Size_)
	}
	if m.writing != 0 {
		fatal("concurrent map writes")
	}

	hash := typ.Hasher(key, m.seed)

	// Set writing after calling Hasher, since Hasher may panic, in which
	// case we have not actually done a write.
	m.writing ^= 1 // toggle, see comment on writing

	if m.dirPtr == nil {
		m.growToSmall()
	}

	if m.dirLen < 0 {
		slotElem := m.putSlotSmall(hash, key)

		if m.writing == 0 {
			fatal("concurrent map writes")
		}
		m.writing ^= 1

		return slotElem
	} else if m.dirLen == 0 {
		m.growToTable()
	}

	var slotElem unsafe.Pointer
outer:
	for {
		// Select table.
		idx := m.directoryIndex(hash)
		t := m.directoryAt(idx)

		seq := makeProbeSeq(h1(hash), t.groups.lengthMask)

		// As we look for a match, keep track of the first deleted slot we
		// find, which we'll use to insert the new entry if necessary.
		var firstDeletedGroup groupReference
		var firstDeletedSlot uintptr

		for ; ; seq = seq.next() {
			g := t.groups.group(typ, seq.offset)
			match := g.ctrls().matchH2(h2(hash))

			// Look for an existing slot containing this key.
			for match != 0 {
				i := match.first()

				slotKey := g.key(typ, i)
				if typ.IndirectKey() {
					slotKey = *((*unsafe.Pointer)(slotKey))
				}
				if t.typ.Key.Equal(key, slotKey) {
					if t.typ.NeedKeyUpdate() {
						typedmemmove(t.typ.Key, slotKey, key)
					}

					slotElem = g.elem(typ, i)
					if typ.IndirectElem() {
						slotElem = *((*unsafe.Pointer)(slotElem))
					}

					t.checkInvariants(m)
					break outer
				}
				match = match.removeFirst()
			}

			// No existing slot for this key in this group. Is this the end
			// of the probe sequence?
			//
			// In parallel, we check for a deleted slot, which we may use,
			// but only once we get to the end of the probe sequence and
			// know there is no existing match.
			if firstDeletedGroup.data == nil {
				match = g.ctrls().matchEmptyOrDeleted()
			} else {
				match = g.ctrls().matchEmpty()
			}
			if match != 0 {
				i := match.first()
				if firstDeletedGroup.data == nil && g.ctrls().get(i) == ctrlDeleted {
					firstDeletedGroup = g
					firstDeletedSlot = i
					// It is impossible to have both empty and
					// deleted slots in a group, so this group must
					// be full. Continue to next group to keep
					// looking for an existing matching slot.
					continue
				}

				// Finding an empty slot means we've reached the end of
				// the probe sequence.

				// If we found a deleted slot along the way, we can
				// replace it without consuming growthLeft.
				if firstDeletedGroup.data != nil {
					g = firstDeletedGroup
					i = firstDeletedSlot
					t.growthLeft++ // will be decremented below to become a no-op.
				}

				// If there is room left to grow, just insert the new entry.
				if t.growthLeft > 0 {
					slotKey := g.key(m.typ, i)
					if t.typ.IndirectKey() {
						kmem := newobject(t.typ.Key)
						*(*unsafe.Pointer)(slotKey) = kmem
						slotKey = kmem
					}
					typedmemmove(t.typ.Key, slotKey, key)

					slotElem = g.elem(m.typ, i)
					if t.typ.IndirectElem() {
						emem := newobject(t.typ.Elem)
						*(*unsafe.Pointer)(slotElem) = emem
						slotElem = emem
					}

					g.ctrls().set(i, ctrl(h2(hash)))
					t.growthLeft--
					t.used++
					m.used++

					t.checkInvariants(m)
					break outer
				}

				t.rehash(m)
				continue outer
			}
		}
	}

	if m.writing == 0 {
		fatal("concurrent map writes")
	}
	m.writing ^= 1

	return slotElem
}
