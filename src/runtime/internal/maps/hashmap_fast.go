// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_hash "runtime/internal/hash"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func mapaccess1_fast32(t *Maptype, h *Hmap, key uint32) unsafe.Pointer {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess1_fast32))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&key)), 4, uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			k := *((*uint32)(_core.Add(unsafe.Pointer(b), DataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*4+i*uintptr(t.valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_fast32(t *Maptype, h *Hmap, key uint32) (unsafe.Pointer, bool) {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess2_fast32))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&key)), 4, uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			k := *((*uint32)(_core.Add(unsafe.Pointer(b), DataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*4+i*uintptr(t.valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}

func mapaccess1_fast64(t *Maptype, h *Hmap, key uint64) unsafe.Pointer {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess1_fast64))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&key)), 8, uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			k := *((*uint64)(_core.Add(unsafe.Pointer(b), DataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*8+i*uintptr(t.valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_fast64(t *Maptype, h *Hmap, key uint64) (unsafe.Pointer, bool) {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess2_fast64))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table.  No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&key)), 8, uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
		if c := h.oldbuckets; c != nil {
			oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			k := *((*uint64)(_core.Add(unsafe.Pointer(b), DataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*8+i*uintptr(t.valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}

func mapaccess1_faststr(t *Maptype, h *Hmap, ky string) unsafe.Pointer {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess1_faststr))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	key := (*_lock.StringStruct)(unsafe.Pointer(&ky))
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < BucketCnt; i++ {
				x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
				if x == Empty {
					continue
				}
				k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
				if k.Len != key.Len {
					continue
				}
				if k.Str == key.Str || _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
					return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize))
				}
			}
			return unsafe.Pointer(t.elem.Zero)
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(BucketCnt)
		for i := uintptr(0); i < BucketCnt; i++ {
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize))
			}
			// check first 4 bytes
			// TODO: on amd64/386 at least, make this compile to one 4-byte comparison instead of
			// four 1-byte comparisons.
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(_core.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(_core.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != BucketCnt {
				// Two keys are potential matches.  Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != BucketCnt {
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+keymaybe*2*_core.PtrSize))
			if _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+keymaybe*uintptr(t.valuesize))
			}
		}
		return unsafe.Pointer(t.elem.Zero)
	}
dohash:
	hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&ky)), 2*_core.PtrSize, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x != top {
				continue
			}
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize))
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2_faststr(t *Maptype, h *Hmap, ky string) (unsafe.Pointer, bool) {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapaccess2_faststr))
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	key := (*_lock.StringStruct)(unsafe.Pointer(&ky))
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < BucketCnt; i++ {
				x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
				if x == Empty {
					continue
				}
				k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
				if k.Len != key.Len {
					continue
				}
				if k.Str == key.Str || _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
					return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize)), true
				}
			}
			return unsafe.Pointer(t.elem.Zero), false
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(BucketCnt)
		for i := uintptr(0); i < BucketCnt; i++ {
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x == Empty {
				continue
			}
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize)), true
			}
			// check first 4 bytes
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(_core.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(_core.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != BucketCnt {
				// Two keys are potential matches.  Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != BucketCnt {
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+keymaybe*2*_core.PtrSize))
			if _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+keymaybe*uintptr(t.valuesize)), true
			}
		}
		return unsafe.Pointer(t.elem.Zero), false
	}
dohash:
	hash := t.key.Alg.Hash(_core.Noescape(unsafe.Pointer(&ky)), 2*_core.PtrSize, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(_core.Add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(_core.Add(c, (hash&(m>>1))*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			x := *((*uint8)(_core.Add(unsafe.Pointer(b), i))) // b.topbits[i] without the bounds check
			if x != top {
				continue
			}
			k := (*_lock.StringStruct)(_core.Add(unsafe.Pointer(b), DataOffset+i*2*_core.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || _hash.Memeq(k.Str, key.Str, uintptr(key.Len)) {
				return _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*2*_core.PtrSize+i*uintptr(t.valuesize)), true
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}
