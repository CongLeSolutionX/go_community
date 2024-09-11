// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swissmap

package maps

import (
	"internal/abi"
	"internal/asan"
	"internal/msan"
	//"internal/runtime/sys"
	"unsafe"
)

// Functions below pushed from runtime.

//go:linkname mapKeyError
func mapKeyError(typ *abi.SwissMapType, p unsafe.Pointer) error

// Pushed from runtime in order to use runtime.plainError
//
//go:linkname errNilAssign
var errNilAssign error

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
	// TODO: concurrent checks.
	//if raceenabled && m != nil {
	//	callerpc := sys.GetCallerPC()
	//	pc := abi.FuncPCABIInternal(mapaccess1)
	//	racereadpc(unsafe.Pointer(m), callerpc, pc)
	//	raceReadObjectPC(t.Key, key, callerpc, pc)
	//}
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
			if typ.Key.Kind() == abi.Int32 {
				if *(*int32)(key) == *(*int32)(slotKey) {
					return g.elem(typ, i)
				}
			} else if typ.Key.Equal(key, slotKey) {
				return g.elem(typ, i)
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

//go:linkname runtime_mapassign runtime.mapassign
func runtime_mapassign(typ *abi.SwissMapType, m *Map, key unsafe.Pointer) unsafe.Pointer {
	// TODO: concurrent checks.
	if m == nil {
		panic(errNilAssign)
	}
	//if raceenabled {
	//	callerpc := sys.GetCallerPC()
	//	pc := abi.FuncPCABIInternal(mapassign)
	//	racewritepc(unsafe.Pointer(m), callerpc, pc)
	//	raceReadObjectPC(t.Key, key, callerpc, pc)
	//}
	if msan.Enabled {
		msan.Read(key, typ.Key.Size_)
	}
	if asan.Enabled {
		asan.Read(key, typ.Key.Size_)
	}

	hash := typ.Hasher(key, m.seed)

	if m.dirLen < 0 {
		return m.putSlotSmall(hash, key)
	} else if m.dirLen == 0 {
		m.growToTable()
	}

outer:
	for {
		// Select table.
		idx := m.directoryIndex(hash)
		t := m.directoryAt(idx)

		seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
		//startOffset := seq.offset

		for ; ; seq = seq.next() {
			g := t.groups.group(typ, seq.offset)
			match := g.ctrls().matchH2(h2(hash))

			// Look for an existing slot containing this key.
			for match != 0 {
				i := match.first()

				slotKey := g.key(typ, i)
				if t.typ.Key.Equal(key, slotKey) {
					if t.typ.NeedKeyUpdate() {
						typedmemmove(t.typ.Key, slotKey, key)
					}

					slotElem := g.elem(typ, i)

					t.checkInvariants()
					return slotElem
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

					slotKey := g.key(typ, i)
					typedmemmove(t.typ.Key, slotKey, key)
					slotElem := g.elem(typ, i)

					g.ctrls().set(i, ctrl(h2(hash)))
					t.growthLeft--
					t.used++
					m.used++

					t.checkInvariants()
					return slotElem
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
				continue outer
			}
		}
	}
}
