// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swissmap

package maps

import (
	"internal/abi"
	"internal/race"
	"internal/runtime/sys"
	"unsafe"
)

// TODO: more string-specific optimizations possible.

func (m *Map) getWithoutKeySmallFastStr(hash uintptr, key string) (unsafe.Pointer, bool) {
	g := groupReference{
		data: m.dirPtr,
	}

	h2 := uint8(h2(hash))
	ctrls := *g.ctrls()

	for i := uintptr(0); i < 8; i++ {
		c := uint8(ctrls)
		ctrls >>= 8
		if c != h2 {
			continue
		}

		slotKey := g.key(m.typ, i)

		if key == *(*string)(slotKey) {
			slotElem := g.elem(m.typ, i)
			return slotElem, true
		}
	}

	return nil, false
}

//go:linkname runtime_mapaccess1_faststr runtime.mapaccess1_faststr
func runtime_mapaccess1_faststr(typ *abi.SwissMapType, m *Map, key string) unsafe.Pointer {
	if race.Enabled && m != nil {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapaccess1)
		race.ReadObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
	}

	if m == nil || m.Used() == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}

	if m.writing != 0 {
		fatal("concurrent map read and map write")
	}

	hash := typ.Hasher(noescape(unsafe.Pointer(&key)), m.seed)

	if m.dirLen <= 0 {
		elem, ok := m.getWithoutKeySmallFastStr(hash, key)
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
			if key == *(*string)(slotKey) {
				slotElem := g.elem(typ, i)
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

//go:linkname runtime_mapaccess2_faststr runtime.mapaccess2_faststr
func runtime_mapaccess2_faststr(typ *abi.SwissMapType, m *Map, key string) (unsafe.Pointer, bool) {
	if race.Enabled && m != nil {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapaccess1)
		race.ReadObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
	}

	if m == nil || m.Used() == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}

	if m.writing != 0 {
		fatal("concurrent map read and map write")
	}

	hash := typ.Hasher(noescape(unsafe.Pointer(&key)), m.seed)

	if m.dirLen <= 0 {
		elem, ok := m.getWithoutKeySmallFastStr(hash, key)
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
			if key == *(*string)(slotKey) {
				slotElem := g.elem(typ, i)
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

func (m *Map) putSlotSmallFastStr(hash uintptr, key string) unsafe.Pointer {
	g := groupReference{
		data: m.dirPtr,
	}

	match := g.ctrls().matchH2(h2(hash))

	// Look for an existing slot containing this key.
	for match != 0 {
		i := match.first()

		slotKey := g.key(m.typ, i)
		if key == *(*string)(slotKey) {
			// Key needs update, as the backing storage may differ.
			*(*string)(slotKey) = key
			slotElem := g.elem(m.typ, i)
			return slotElem
		}
		match = match.removeFirst()
	}

	match = g.ctrls().matchEmpty()
	if match != 0 {
		i := match.first()

		slotKey := g.key(m.typ, i)
		*(*string)(slotKey) = key

		slotElem := g.elem(m.typ, i)

		g.ctrls().set(i, ctrl(h2(hash)))
		m.dirLen++
		m.used++

		return slotElem
	}

	panic("small map with negative dirLen has no empty slot")
}

//go:linkname runtime_mapassign_faststr runtime.mapassign_faststr
func runtime_mapassign_faststr(typ *abi.SwissMapType, m *Map, key string) unsafe.Pointer {
	if m == nil {
		panic(errNilAssign)
	}
	if race.Enabled {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapassign)
		race.WriteObjectPC(&typ.Type, unsafe.Pointer(m), callerpc, pc)
	}
	if m.writing != 0 {
		fatal("concurrent map writes")
	}

	hash := typ.Hasher(noescape(unsafe.Pointer(&key)), m.seed)

	// Set writing after calling Hasher, since Hasher may panic, in which
	// case we have not actually done a write.
	m.writing ^= 1 // toggle, see comment on writing

	if m.dirPtr == nil {
		m.growToSmall()
	}

	if m.dirLen < 0 {
		slotElem := m.putSlotSmallFastStr(hash, key)

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
				if key == *(*string)(slotKey) {
					// Key needs update, as the backing
					// storage may differ.
					*(*string)(slotKey) = key
					slotElem = g.elem(typ, i)

					t.checkInvariants(m)
					break outer
				}
				match = match.removeFirst()
			}

			// No existing slot for this key in this group. Is this the end
			// of the probe sequence?
			match = g.ctrls().matchEmpty()
			if match != 0 {
				// Finding an empty slot means we've reached the end of
				// the probe sequence.

				var i uintptr

				// If we found a deleted slot along the way, we
				// can replace it without consuming growthLeft.
				if firstDeletedGroup.data != nil {
					g = firstDeletedGroup
					i = firstDeletedSlot
					t.growthLeft++ // will be decremented below to become a no-op.
				} else {
					// Otherwise, use the empty slot.
					i = match.first()
				}

				// If there is room left to grow, just insert the new entry.
				if t.growthLeft > 0 {
					slotKey := g.key(m.typ, i)
					*(*string)(slotKey) = key

					slotElem = g.elem(m.typ, i)

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

			// No empty slots in this group. Check for a deleted
			// slot, which we'll use if we don't find a match later
			// in the probe sequence.
			//
			// We only need to remember a single deleted slot.
			if firstDeletedGroup.data == nil {
				// Since we already checked for empty slots
				// above, matches here must be deleted slots.
				match = g.ctrls().matchEmptyOrDeleted()
				if match != 0 {
					firstDeletedGroup = g
					firstDeletedSlot = match.first()
				}
			}
		}
	}

	if m.writing == 0 {
		fatal("concurrent map writes")
	}
	m.writing ^= 1

	return slotElem
}

//go:linkname runtime_mapdelete_faststr runtime.mapdelete_faststr
func runtime_mapdelete_faststr(t *abi.SwissMapType, m *Map, key string) {
	if race.Enabled {
		callerpc := sys.GetCallerPC()
		pc := abi.FuncPCABIInternal(runtime_mapassign)
		race.WriteObjectPC(&t.Type, unsafe.Pointer(m), callerpc, pc)
	}

	if m == nil || m.Used() == 0 {
		return
	}

	m.Delete(noescape(unsafe.Pointer(&key)))
}
