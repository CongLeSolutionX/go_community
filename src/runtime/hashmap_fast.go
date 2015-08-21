// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_race "runtime/internal/race"
	"unsafe"
)

func mapaccess1_fast32(t *maptype, h *hmap, key uint32) unsafe.Pointer {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess1_fast32))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint32)(_base.Add(unsafe.Pointer(b), dataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_fast32(t *maptype, h *hmap, key uint32) (unsafe.Pointer, bool) {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess2_fast32))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint32)(_base.Add(unsafe.Pointer(b), dataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}

func mapaccess1_fast64(t *maptype, h *hmap, key uint64) unsafe.Pointer {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess1_fast64))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint64)(_base.Add(unsafe.Pointer(b), dataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_fast64(t *maptype, h *hmap, key uint64) (unsafe.Pointer, bool) {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess2_fast64))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint64)(_base.Add(unsafe.Pointer(b), dataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}

func mapaccess1_faststr(t *maptype, h *hmap, ky string) unsafe.Pointer {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess1_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	key := (*_base.StringStruct)(unsafe.Pointer(&ky))
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < bucketCnt; i++ {
				x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
				if x == empty {
					continue
				}
				k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
				if k.Len != key.Len {
					continue
				}
				if k.Str == key.Str || memeq(k.Str, key.Str, uintptr(key.Len)) {
					return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize))
				}
			}
			return unsafe.Pointer(t.elem.Zero)
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize))
			}
			// check first 4 bytes
			// TODO: on amd64/386 at least, make this compile to one 4-byte comparison instead of
			// four 1-byte comparisons.
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(_base.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(_base.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches.  Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+keymaybe*2*_base.PtrSize))
			if memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+keymaybe*uintptr(t.valuesize))
			}
		}
		return unsafe.Pointer(t.elem.Zero)
	}
dohash:
	hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (_base.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x != top {
				continue
			}
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize))
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_faststr(t *maptype, h *hmap, ky string) (unsafe.Pointer, bool) {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapaccess2_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	key := (*_base.StringStruct)(unsafe.Pointer(&ky))
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < bucketCnt; i++ {
				x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
				if x == empty {
					continue
				}
				k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
				if k.Len != key.Len {
					continue
				}
				if k.Str == key.Str || memeq(k.Str, key.Str, uintptr(key.Len)) {
					return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize)), true
				}
			}
			return unsafe.Pointer(t.elem.Zero), false
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == empty {
				continue
			}
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize)), true
			}
			// check first 4 bytes
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(_base.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(_base.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches.  Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+keymaybe*2*_base.PtrSize))
			if memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+keymaybe*uintptr(t.valuesize)), true
			}
		}
		return unsafe.Pointer(t.elem.Zero), false
	}
dohash:
	hash := t.key.Alg.Hash(_base.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(_base.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(_base.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (_base.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(_base.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x != top {
				continue
			}
			k := (*_base.StringStruct)(_base.Add(unsafe.Pointer(b), dataOffset+i*2*_base.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*_base.PtrSize+i*uintptr(t.valuesize)), true
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}
