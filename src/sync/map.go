// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync/atomic"
)

// An entry is a slot in the map corresponding to a particular key.
//
// TODO(bcmills): Is it possible/safe to scribble out the interface{} value
// with something else (perhaps a nil interface?) to allow stale values to be
// garbage-collected?
type entry struct {
	v           interface{}
	atomicState int32
}

func (e *entry) state() state {
	return state(atomic.LoadInt32(&e.atomicState))
}

func (e *entry) deleteIfStored() (isDeleted bool) {
	if atomic.CompareAndSwapInt32(&e.atomicState, int32(stored), int32(deleted)) {
		return true
	}
	return e.state() == deleted
}

func (e *entry) setState(st state) {
	atomic.StoreInt32(&e.atomicState, int32(st))
}

// A state is the state of an individual entry.
//
// Once published in the read map, state transitions are restricted:
// A stored entry may transition to deleted or dirty.
// A deleted entry may transition to dirty.
// A dirty entry may transition to deleted.
//
// A published entry which is no longer in the stored state cannot transition
// back to stored: the value stored in the entry cannot be changed without
// introducing a race with in-flight reads of that value.
type state int32

const (
	stored  = state(iota) // The entry contains the up-to-date value.
	deleted               // The entry has been deleted.
	dirty                 // The entry has been replaced with a new entry in the dirty map.
)

// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // true if the dirty map contains some key not in m.
}

// Map is a concurrent map with amortized-constant-time operations.
// It is safe for multiple goroutines to call a Map's methods concurrently.
//
// The zero Map is valid and empty.
//
// A Map must not be copied after first use.
type Map struct {
	mu Mutex

	// read contains the portion of the map's contents that can be read
	// concurrently (with or without mu held).
	//
	// read is always safe to load, but must only be stored with mu held.
	//
	// Entries in the read map may be in any state, but the values stored in those
	// entries cannot be changed.
	//
	// TODO(bcmills): Is it possible/safe to scribble out the entry values with
	// something else (perhaps a nil interface?) to allow stale values to be
	// garbage-collected?
	read atomic.Value // readOnly

	// dirty contains the portion of the map's contents that require mu
	// to be held, including all of the stored entries from the read map.
	//
	// Entries in the dirty map are always in the stored state.  The values
	// contained in dirty entries may only be changed if the entry has not been
	// published in the read map.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	dirty map[interface{}]*entry

	// misses counts the number of Load calls since the read map was last updated
	// for which either:
	// * the entry in the read map was in the dirty state, or
	// * there was no entry in the read map and the read map was amended.
	//
	// Once enough Load misses have occurred to cover the cost of copying the
	// dirty map, the dirty map will be promoted to the read map (in the unamended
	// state) and any subsequent stores to the map will make a new copy.
	misses int
}

func (m *Map) load(key interface{}) (value interface{}, ok bool, clean bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok {
		// If the map has not been amended, a missing entry implies that there is
		// definitely no value for the key.
		return nil, false, !read.amended
	}
	switch e.state() {
	case stored:
		return e.v, true, true
	case deleted:
		return nil, false, true
	case dirty:
		return nil, false, false
	default:
		panic("sync: entry in invalid state")
	}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	value, ok, clean := m.load(key)
	if clean {
		return value, ok
	}

	// Either the entry was dirty, or the key was not present and the map has been
	// amended.

	m.mu.Lock()
	// Avoid reporting a spurious miss if m.dirty got promoted while we were
	// blocked on m.mu.  (If further loads of the same key will not miss, it's not
	// worth copying the dirty map for this key.)
	value, ok, clean = m.load(key)
	if clean {
		m.mu.Unlock()
		return value, ok
	}

	if e, dirty := m.dirty[key]; dirty {
		value, ok = e.v, true
	} else {
		value, ok = nil, false
	}

	// Regardless of whether the entry was present, record a miss: this key will
	// take the slow path until either it is deleted or the dirty map is promoted
	// to the read map.
	m.missLocked()
	m.mu.Unlock()
	return value, ok
}

// Store sets the value for a key.
func (m *Map) Store(key, value interface{}) {
	m.mu.Lock()

	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		if e.state() == dirty {
			// This entry is dirty, so there is a corresponding new entry in m.dirty
			// which has not been published.
			m.dirty[key].v = value
		} else {
			m.dirtyLocked()
			m.dirty[key] = &entry{v: value}
			e.setState(dirty)
		}
		m.mu.Unlock()
		return
	}

	if e, ok := m.dirty[key]; ok {
		e.v = value
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = &entry{v: value}
	}

	m.mu.Unlock()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	// Avoid locking if it's a clean hit.
	actual, loaded, clean := m.load(key)
	if clean && loaded {
		return actual, true
	}

	m.mu.Lock()
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if ok {
		switch e.state() {
		case stored:
			m.mu.Unlock()
			return e.v, true
		case dirty:
			// This entry is dirty, so there is a corresponding new entry in m.dirty
			// which has not been published.
			actual = m.dirty[key].v
			m.missLocked()
			m.mu.Unlock()
			return actual, true
		case deleted:
			m.dirtyLocked()
			m.dirty[key] = &entry{v: value}
			e.setState(dirty)
			m.mu.Unlock()
			return value, false
		default:
			m.mu.Unlock()
			panic("sync: entry in invalid state")
		}
	}

	if d, ok := m.dirty[key]; ok {
		actual = d.v
		m.missLocked()
		m.mu.Unlock()
		return actual, true
	}

	// There is no existing entry: add a new one.

	if !read.amended {
		m.dirtyLocked()
		m.read.Store(readOnly{m: read.m, amended: true})
	}
	m.dirty[key] = &entry{v: value}
	m.mu.Unlock()
	return value, false
}

// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) {
	m.mu.Lock()
	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		e.setState(deleted)
	}
	delete(m.dirty, key)
	m.mu.Unlock()
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (m *Map) Range(f func(key, value interface{}) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read, _ := m.read.Load().(readOnly)
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N),
		// so a call to Range amortizes an entire copy of the map: we can promote
		// the dirty copy immediately!
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		if read.amended {
			read = readOnly{m: m.dirty}
			m.read.Store(read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	// read gives us a consistent set of keys, but some of them might be dirty.
	// cur stores a fresh copy of m.read every time we acquire m.mu, which allows
	// us to do as many clean lookups as possible without taking the lock.
	cur := read
	for k, old := range read.m {
		switch old.state() {
		case stored:
			// old is still fresh!
			if f(k, old.v) {
				continue
			} else {
				break
			}
		case deleted:
			continue
		}
		e, ok := cur.m[k]
		if ok {
			if e != old {
				switch e.state() {
				case stored:
					if f(k, e.v) {
						continue
					} else {
						break
					}
				case deleted:
					continue
				}
			}
		} else if !cur.amended {
			// The entry for the key was deleted as of cur.
			continue
		}

		m.mu.Lock()
		cur = m.read.Load().(readOnly)
		if m.dirty == nil {
			e, ok = cur.m[k]
			if ok && e.state() == deleted {
				ok = false
			}
		} else {
			e, ok = m.dirty[k]
			if ok {
				m.missLocked()
			}
		}
		m.mu.Unlock()
		if !ok {
			continue
		}
		if !f(k, e.v) {
			break
		}
	}
}

func (m *Map) missLocked() {
	if m.misses++; m.misses >= len(m.dirty) {
		m.read.Store(readOnly{m: m.dirty})
		m.dirty = nil
		m.misses = 0
	}
}

func (m *Map) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, e := range read.m {
		switch e.state() {
		case stored:
			m.dirty[k] = e
		case deleted:
			continue
		case dirty:
			panic("sync: dirty key without dirty map")
		default:
			panic("sync: entry in invalid state")
		}
	}
}
