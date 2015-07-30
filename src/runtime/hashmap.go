// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

const (
	// Maximum number of key/value pairs a bucket can hold.
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits

	// Maximum average load of a bucket that triggers growth.
	loadFactor = 6.5

	// Maximum key or value size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	// Fast versions cannot handle big values - the cutoff size for
	// fast versions in ../../cmd/internal/gc/walk.go must be at most this value.
	maxKeySize   = 128
	maxValueSize = 128

	// data offset should be the size of the bmap struct, but needs to be
	// aligned correctly.  For amd64p32 this means 64-bit alignment
	// even though pointers are 32 bit.
	dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// Possible tophash values.  We reserve a few possibilities for special marks.
	// Each bucket (including its overflow buckets, if any) will have either all or none of its
	// entries in the evacuated* states (except during the evacuate() method, which only happens
	// during map writes and thus no one else can observe the map during that time).
	empty          = 0 // cell is empty
	evacuatedEmpty = 1 // cell is empty, bucket is evacuated.
	evacuatedX     = 2 // key/value is valid.  Entry has been evacuated to first half of larger table.
	evacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	minTopHash     = 4 // minimum tophash for a normal filled cell.

	// flags
	iterator    = 1 // there may be an iterator using buckets
	oldIterator = 2 // there may be an iterator using oldbuckets

	// sentinel bucket ID for iterator checks
	noCheck = 1<<(8*_base.PtrSize) - 1
)

// A header for a Go map.
type hmap struct {
	// Note: the format of the Hmap is encoded in ../../cmd/internal/gc/reflect.go and
	// ../reflect/type.go.  Don't change this structure without also changing that code!
	count int // # live cells == size of map.  Must be first (used by len() builtin)
	flags uint8
	B     uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	hash0 uint32 // hash seed

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	// If both key and value do not contain pointers and are inline, then we mark bucket
	// type as containing no pointers. This avoids scanning such maps.
	// However, bmap.overflow is a pointer. In order to keep overflow buckets
	// alive, we store pointers to all overflow buckets in hmap.overflow.
	// Overflow is used only if key and value do not contain pointers.
	// overflow[0] contains overflow buckets for hmap.buckets.
	// overflow[1] contains overflow buckets for hmap.oldbuckets.
	// The first indirection allows us to reduce static size of hmap.
	// The second indirection allows to store a pointer to the slice in hiter.
	overflow *[2]*[]*bmap
}

// A bucket for a Go map.
type bmap struct {
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt values.
	// NOTE: packing all the keys together and then all the values together makes the
	// code a bit more complicated than alternating key/value/key/value/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}

// A hash iteration structure.
// If you modify hiter, also change cmd/internal/gc/reflect.go to indicate
// the layout of this structure.
type hiter struct {
	key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/internal/gc/range.go).
	value       unsafe.Pointer // Must be in second position (see cmd/internal/gc/range.go).
	t           *maptype
	h           *hmap
	buckets     unsafe.Pointer // bucket ptr at hash_iter initialization time
	bptr        *bmap          // current bucket
	overflow    [2]*[]*bmap    // keeps overflow buckets alive
	startBucket uintptr        // bucket iteration started at
	offset      uint8          // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	wrapped     bool           // already wrapped around from end of bucket array to beginning
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

func evacuated(b *bmap) bool {
	h := b.tophash[0]
	return h > empty && h < minTopHash
}

func (b *bmap) overflow(t *maptype) *bmap {
	return *(**bmap)(_base.Add(unsafe.Pointer(b), uintptr(t.bucketsize)-_base.RegSize))
}

func (h *hmap) setoverflow(t *maptype, b, ovf *bmap) {
	if t.bucket.Kind&_iface.KindNoPointers != 0 {
		h.createOverflow()
		*h.overflow[0] = append(*h.overflow[0], ovf)
	}
	*(**bmap)(_base.Add(unsafe.Pointer(b), uintptr(t.bucketsize)-_base.RegSize)) = ovf
}

func (h *hmap) createOverflow() {
	if h.overflow == nil {
		h.overflow = new([2]*[]*bmap)
	}
	if h.overflow[0] == nil {
		h.overflow[0] = new([]*bmap)
	}
}

// makemap implements a Go map creation make(map[k]v, hint)
// If the compiler has determined that the map or the first bucket
// can be created on the stack, h and/or bucket may be non-nil.
// If h != nil, the map can be created directly in h.
// If bucket != nil, bucket can be used as the first bucket.
func makemap(t *maptype, hint int64, h *hmap, bucket unsafe.Pointer) *hmap {
	if sz := unsafe.Sizeof(hmap{}); sz > 48 || sz != uintptr(t.hmap.Size) {
		println("runtime: sizeof(hmap) =", sz, ", t.hmap.size =", t.hmap.Size)
		_base.Throw("bad hmap size")
	}

	if hint < 0 || int64(int32(hint)) != hint {
		panic("makemap: size out of range")
		// TODO: make hint an int, then none of this nonsense
	}

	if !ismapkey(t.key) {
		_base.Throw("runtime.makemap: unsupported map key type")
	}

	// check compiler's and reflect's math
	if t.key.Size > maxKeySize && (!t.indirectkey || t.keysize != uint8(_base.PtrSize)) ||
		t.key.Size <= maxKeySize && (t.indirectkey || t.keysize != uint8(t.key.Size)) {
		_base.Throw("key size wrong")
	}
	if t.elem.Size > maxValueSize && (!t.indirectvalue || t.valuesize != uint8(_base.PtrSize)) ||
		t.elem.Size <= maxValueSize && (t.indirectvalue || t.valuesize != uint8(t.elem.Size)) {
		_base.Throw("value size wrong")
	}

	// invariants we depend on.  We should probably check these at compile time
	// somewhere, but for now we'll do it here.
	if t.key.Align > bucketCnt {
		_base.Throw("key align too big")
	}
	if t.elem.Align > bucketCnt {
		_base.Throw("value align too big")
	}
	if uintptr(t.key.Size)%uintptr(t.key.Align) != 0 {
		_base.Throw("key size not a multiple of key align")
	}
	if uintptr(t.elem.Size)%uintptr(t.elem.Align) != 0 {
		_base.Throw("value size not a multiple of value align")
	}
	if bucketCnt < 8 {
		_base.Throw("bucketsize too small for proper alignment")
	}
	if dataOffset%uintptr(t.key.Align) != 0 {
		_base.Throw("need padding in bucket (key)")
	}
	if dataOffset%uintptr(t.elem.Align) != 0 {
		_base.Throw("need padding in bucket (value)")
	}

	// make sure zero of element type is available.
	mapzero(t.elem)

	// find size parameter which will hold the requested # of elements
	B := uint8(0)
	for ; hint > bucketCnt && float32(hint) > loadFactor*float32(uintptr(1)<<B); B++ {
	}

	// allocate initial hash table
	// if B == 0, the buckets field is allocated lazily later (in mapassign)
	// If hint is large zeroing this memory could take a while.
	buckets := bucket
	if B != 0 {
		buckets = newarray(t.bucket, uintptr(1)<<B)
	}

	// initialize Hmap
	if h == nil {
		h = (*hmap)(_iface.Newobject(t.hmap))
	}
	h.count = 0
	h.B = B
	h.flags = 0
	h.hash0 = _base.Fastrand1()
	h.buckets = buckets
	h.oldbuckets = nil
	h.nevacuate = 0

	return h
}

// mapaccess1 returns a pointer to h[key].  Never returns nil, instead
// it will return a reference to the zero object for the value type if
// the key is not in the map.
// NOTE: The returned pointer may keep the whole map live, so don't
// hold onto it for very long.
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		pc := _base.FuncPC(mapaccess1)
		_race.Racereadpc(unsafe.Pointer(h), callerpc, pc)
		_race.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
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
			if b.tophash[i] != top {
				continue
			}
			k := _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
				if t.indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return v
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero)
		}
	}
}

func mapaccess2(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		pc := _base.FuncPC(mapaccess2)
		_race.Racereadpc(unsafe.Pointer(h), callerpc, pc)
		_race.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(t.elem.Zero), false
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(unsafe.Pointer(uintptr(c) + (hash&(m>>1))*uintptr(t.bucketsize)))
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
			if b.tophash[i] != top {
				continue
			}
			k := _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
				if t.indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return v, true
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(t.elem.Zero), false
		}
	}
}

// returns both key and value.  Used by map iterator
func mapaccessK(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer) {
	if h == nil || h.count == 0 {
		return nil, nil
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		oldb := (*bmap)(unsafe.Pointer(uintptr(c) + (hash&(m>>1))*uintptr(t.bucketsize)))
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
			if b.tophash[i] != top {
				continue
			}
			k := _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
				if t.indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return k, v
			}
		}
		b = b.overflow(t)
		if b == nil {
			return nil, nil
		}
	}
}

func mapassign1(t *maptype, h *hmap, key unsafe.Pointer, val unsafe.Pointer) {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		pc := _base.FuncPC(mapassign1)
		_race.Racewritepc(unsafe.Pointer(h), callerpc, pc)
		_race.RaceReadObjectPC(t.key, key, callerpc, pc)
		_race.RaceReadObjectPC(t.elem, val, callerpc, pc)
	}

	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))

	if h.buckets == nil {
		h.buckets = newarray(t.bucket, 1)
	}

again:
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.oldbuckets != nil {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := uint8(hash >> (_base.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}

	var inserti *uint8
	var insertk unsafe.Pointer
	var insertv unsafe.Pointer
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				if b.tophash[i] == empty && inserti == nil {
					inserti = &b.tophash[i]
					insertk = _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
					insertv = _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
				}
				continue
			}
			k := _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			k2 := k
			if t.indirectkey {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if !alg.Equal(key, k2) {
				continue
			}
			// already have a mapping for key.  Update it.
			_iface.Typedmemmove(t.key, k2, key)
			v := _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
			v2 := v
			if t.indirectvalue {
				v2 = *((*unsafe.Pointer)(v2))
			}
			_iface.Typedmemmove(t.elem, v2, val)
			return
		}
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// did not find mapping for key.  Allocate new cell & add entry.
	if float32(h.count) >= loadFactor*float32((uintptr(1)<<h.B)) && h.count >= bucketCnt {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if inserti == nil {
		// all current buckets are full, allocate a new one.
		newb := (*bmap)(_iface.Newobject(t.bucket))
		h.setoverflow(t, b, newb)
		inserti = &newb.tophash[0]
		insertk = _base.Add(unsafe.Pointer(newb), dataOffset)
		insertv = _base.Add(insertk, bucketCnt*uintptr(t.keysize))
	}

	// store new key/value at insert position
	if t.indirectkey {
		kmem := _iface.Newobject(t.key)
		*(*unsafe.Pointer)(insertk) = kmem
		insertk = kmem
	}
	if t.indirectvalue {
		vmem := _iface.Newobject(t.elem)
		*(*unsafe.Pointer)(insertv) = vmem
		insertv = vmem
	}
	_iface.Typedmemmove(t.key, insertk, key)
	_iface.Typedmemmove(t.elem, insertv, val)
	*inserti = top
	h.count++
}

func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		pc := _base.FuncPC(mapdelete)
		_race.Racewritepc(unsafe.Pointer(h), callerpc, pc)
		_race.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.count == 0 {
		return
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.oldbuckets != nil {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := uint8(hash >> (_base.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := _base.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			k2 := k
			if t.indirectkey {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if !alg.Equal(key, k2) {
				continue
			}
			_base.Memclr(k, uintptr(t.keysize))
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + dataOffset + bucketCnt*uintptr(t.keysize) + i*uintptr(t.valuesize))
			_base.Memclr(v, uintptr(t.valuesize))
			b.tophash[i] = empty
			h.count--
			return
		}
		b = b.overflow(t)
		if b == nil {
			return
		}
	}
}

func mapiterinit(t *maptype, h *hmap, it *hiter) {
	// Clear pointer fields so garbage collector does not complain.
	it.key = nil
	it.value = nil
	it.t = nil
	it.h = nil
	it.buckets = nil
	it.bptr = nil
	it.overflow[0] = nil
	it.overflow[1] = nil

	if _base.Raceenabled && h != nil {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapiterinit))
	}

	if h == nil || h.count == 0 {
		it.key = nil
		it.value = nil
		return
	}

	if unsafe.Sizeof(hiter{})/_base.PtrSize != 12 {
		_base.Throw("hash_iter size incorrect") // see ../../cmd/internal/gc/reflect.go
	}
	it.t = t
	it.h = h

	// grab snapshot of bucket state
	it.B = h.B
	it.buckets = h.buckets
	if t.bucket.Kind&_iface.KindNoPointers != 0 {
		// Allocate the current slice and remember pointers to both current and old.
		// This preserves all relevant overflow buckets alive even if
		// the table grows and/or overflow buckets are added to the table
		// while we are iterating.
		h.createOverflow()
		it.overflow = *h.overflow
	}

	// decide where to start
	r := uintptr(_base.Fastrand1())
	if h.B > 31-bucketCntBits {
		r += uintptr(_base.Fastrand1()) << 31
	}
	it.startBucket = r & (uintptr(1)<<h.B - 1)
	it.offset = uint8(r >> h.B & (bucketCnt - 1))

	// iterator state
	it.bucket = it.startBucket
	it.wrapped = false
	it.bptr = nil

	// Remember we have an iterator.
	// Can run concurrently with another hash_iter_init().
	if old := h.flags; old&(iterator|oldIterator) != iterator|oldIterator {
		_base.Atomicor8(&h.flags, iterator|oldIterator)
	}

	mapiternext(it)
}

func mapiternext(it *hiter) {
	h := it.h
	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&it))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(mapiternext))
	}
	t := it.t
	bucket := it.bucket
	b := it.bptr
	i := it.i
	checkBucket := it.checkBucket
	alg := t.key.Alg

next:
	if b == nil {
		if bucket == it.startBucket && it.wrapped {
			// end of iteration
			it.key = nil
			it.value = nil
			return
		}
		if h.oldbuckets != nil && it.B == h.B {
			// Iterator was started in the middle of a grow, and the grow isn't done yet.
			// If the bucket we're looking at hasn't been filled in yet (i.e. the old
			// bucket hasn't been evacuated) then we need to iterate through the old
			// bucket and only return the ones that will be migrated to this bucket.
			oldbucket := bucket & (uintptr(1)<<(it.B-1) - 1)
			b = (*bmap)(_base.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
			if !evacuated(b) {
				checkBucket = bucket
			} else {
				b = (*bmap)(_base.Add(it.buckets, bucket*uintptr(t.bucketsize)))
				checkBucket = noCheck
			}
		} else {
			b = (*bmap)(_base.Add(it.buckets, bucket*uintptr(t.bucketsize)))
			checkBucket = noCheck
		}
		bucket++
		if bucket == uintptr(1)<<it.B {
			bucket = 0
			it.wrapped = true
		}
		i = 0
	}
	for ; i < bucketCnt; i++ {
		offi := (i + it.offset) & (bucketCnt - 1)
		k := _base.Add(unsafe.Pointer(b), dataOffset+uintptr(offi)*uintptr(t.keysize))
		v := _base.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+uintptr(offi)*uintptr(t.valuesize))
		if b.tophash[offi] != empty && b.tophash[offi] != evacuatedEmpty {
			if checkBucket != noCheck {
				// Special case: iterator was started during a grow and the
				// grow is not done yet.  We're working on a bucket whose
				// oldbucket has not been evacuated yet.  Or at least, it wasn't
				// evacuated when we started the bucket.  So we're iterating
				// through the oldbucket, skipping any keys that will go
				// to the other new bucket (each oldbucket expands to two
				// buckets during a grow).
				k2 := k
				if t.indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				if t.reflexivekey || alg.Equal(k2, k2) {
					// If the item in the oldbucket is not destined for
					// the current new bucket in the iteration, skip it.
					hash := alg.Hash(k2, uintptr(h.hash0))
					if hash&(uintptr(1)<<it.B-1) != checkBucket {
						continue
					}
				} else {
					// Hash isn't repeatable if k != k (NaNs).  We need a
					// repeatable and randomish choice of which direction
					// to send NaNs during evacuation.  We'll use the low
					// bit of tophash to decide which way NaNs go.
					// NOTE: this case is why we need two evacuate tophash
					// values, evacuatedX and evacuatedY, that differ in
					// their low bit.
					if checkBucket>>(it.B-1) != uintptr(b.tophash[offi]&1) {
						continue
					}
				}
			}
			if b.tophash[offi] != evacuatedX && b.tophash[offi] != evacuatedY {
				// this is the golden data, we can return it.
				if t.indirectkey {
					k = *((*unsafe.Pointer)(k))
				}
				it.key = k
				if t.indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				it.value = v
			} else {
				// The hash table has grown since the iterator was started.
				// The golden data for this key is now somewhere else.
				k2 := k
				if t.indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				if t.reflexivekey || alg.Equal(k2, k2) {
					// Check the current hash table for the data.
					// This code handles the case where the key
					// has been deleted, updated, or deleted and reinserted.
					// NOTE: we need to regrab the key as it has potentially been
					// updated to an equal() but not identical key (e.g. +0.0 vs -0.0).
					rk, rv := mapaccessK(t, h, k2)
					if rk == nil {
						continue // key has been deleted
					}
					it.key = rk
					it.value = rv
				} else {
					// if key!=key then the entry can't be deleted or
					// updated, so we can just return it.  That's lucky for
					// us because when key!=key we can't look it up
					// successfully in the current table.
					it.key = k2
					if t.indirectvalue {
						v = *((*unsafe.Pointer)(v))
					}
					it.value = v
				}
			}
			it.bucket = bucket
			it.bptr = b
			it.i = i + 1
			it.checkBucket = checkBucket
			return
		}
	}
	b = b.overflow(t)
	i = 0
	goto next
}

func hashGrow(t *maptype, h *hmap) {
	if h.oldbuckets != nil {
		_base.Throw("evacuation not done in time")
	}
	oldbuckets := h.buckets
	newbuckets := newarray(t.bucket, uintptr(1)<<(h.B+1))
	flags := h.flags &^ (iterator | oldIterator)
	if h.flags&iterator != 0 {
		flags |= oldIterator
	}
	// commit the grow (atomic wrt gc)
	h.B++
	h.flags = flags
	h.oldbuckets = oldbuckets
	h.buckets = newbuckets
	h.nevacuate = 0

	if h.overflow != nil {
		// Promote current overflow buckets to the old generation.
		if h.overflow[1] != nil {
			_base.Throw("overflow is not nil")
		}
		h.overflow[1] = h.overflow[0]
		h.overflow[0] = nil
	}

	// the actual copying of the hash table data is done incrementally
	// by growWork() and evacuate().
}

func growWork(t *maptype, h *hmap, bucket uintptr) {
	noldbuckets := uintptr(1) << (h.B - 1)

	// make sure we evacuate the oldbucket corresponding
	// to the bucket we're about to use
	evacuate(t, h, bucket&(noldbuckets-1))

	// evacuate one more oldbucket to make progress on growing
	if h.oldbuckets != nil {
		evacuate(t, h, h.nevacuate)
	}
}

func evacuate(t *maptype, h *hmap, oldbucket uintptr) {
	b := (*bmap)(_base.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
	newbit := uintptr(1) << (h.B - 1)
	alg := t.key.Alg
	if !evacuated(b) {
		// TODO: reuse overflow buckets instead of using new ones, if there
		// is no iterator using the old buckets.  (If !oldIterator.)

		x := (*bmap)(_base.Add(h.buckets, oldbucket*uintptr(t.bucketsize)))
		y := (*bmap)(_base.Add(h.buckets, (oldbucket+newbit)*uintptr(t.bucketsize)))
		xi := 0
		yi := 0
		xk := _base.Add(unsafe.Pointer(x), dataOffset)
		yk := _base.Add(unsafe.Pointer(y), dataOffset)
		xv := _base.Add(xk, bucketCnt*uintptr(t.keysize))
		yv := _base.Add(yk, bucketCnt*uintptr(t.keysize))
		for ; b != nil; b = b.overflow(t) {
			k := _base.Add(unsafe.Pointer(b), dataOffset)
			v := _base.Add(k, bucketCnt*uintptr(t.keysize))
			for i := 0; i < bucketCnt; i, k, v = i+1, _base.Add(k, uintptr(t.keysize)), _base.Add(v, uintptr(t.valuesize)) {
				top := b.tophash[i]
				if top == empty {
					b.tophash[i] = evacuatedEmpty
					continue
				}
				if top < minTopHash {
					_base.Throw("bad map state")
				}
				k2 := k
				if t.indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				// Compute hash to make our evacuation decision (whether we need
				// to send this key/value to bucket x or bucket y).
				hash := alg.Hash(k2, uintptr(h.hash0))
				if h.flags&iterator != 0 {
					if !t.reflexivekey && !alg.Equal(k2, k2) {
						// If key != key (NaNs), then the hash could be (and probably
						// will be) entirely different from the old hash.  Moreover,
						// it isn't reproducible.  Reproducibility is required in the
						// presence of iterators, as our evacuation decision must
						// match whatever decision the iterator made.
						// Fortunately, we have the freedom to send these keys either
						// way.  Also, tophash is meaningless for these kinds of keys.
						// We let the low bit of tophash drive the evacuation decision.
						// We recompute a new random tophash for the next level so
						// these keys will get evenly distributed across all buckets
						// after multiple grows.
						if (top & 1) != 0 {
							hash |= newbit
						} else {
							hash &^= newbit
						}
						top = uint8(hash >> (_base.PtrSize*8 - 8))
						if top < minTopHash {
							top += minTopHash
						}
					}
				}
				if (hash & newbit) == 0 {
					b.tophash[i] = evacuatedX
					if xi == bucketCnt {
						newx := (*bmap)(_iface.Newobject(t.bucket))
						h.setoverflow(t, x, newx)
						x = newx
						xi = 0
						xk = _base.Add(unsafe.Pointer(x), dataOffset)
						xv = _base.Add(xk, bucketCnt*uintptr(t.keysize))
					}
					x.tophash[xi] = top
					if t.indirectkey {
						*(*unsafe.Pointer)(xk) = k2 // copy pointer
					} else {
						_iface.Typedmemmove(t.key, xk, k) // copy value
					}
					if t.indirectvalue {
						*(*unsafe.Pointer)(xv) = *(*unsafe.Pointer)(v)
					} else {
						_iface.Typedmemmove(t.elem, xv, v)
					}
					xi++
					xk = _base.Add(xk, uintptr(t.keysize))
					xv = _base.Add(xv, uintptr(t.valuesize))
				} else {
					b.tophash[i] = evacuatedY
					if yi == bucketCnt {
						newy := (*bmap)(_iface.Newobject(t.bucket))
						h.setoverflow(t, y, newy)
						y = newy
						yi = 0
						yk = _base.Add(unsafe.Pointer(y), dataOffset)
						yv = _base.Add(yk, bucketCnt*uintptr(t.keysize))
					}
					y.tophash[yi] = top
					if t.indirectkey {
						*(*unsafe.Pointer)(yk) = k2
					} else {
						_iface.Typedmemmove(t.key, yk, k)
					}
					if t.indirectvalue {
						*(*unsafe.Pointer)(yv) = *(*unsafe.Pointer)(v)
					} else {
						_iface.Typedmemmove(t.elem, yv, v)
					}
					yi++
					yk = _base.Add(yk, uintptr(t.keysize))
					yv = _base.Add(yv, uintptr(t.valuesize))
				}
			}
		}
		// Unlink the overflow buckets & clear key/value to help GC.
		if h.flags&oldIterator == 0 {
			b = (*bmap)(_base.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
			_base.Memclr(_base.Add(unsafe.Pointer(b), dataOffset), uintptr(t.bucketsize)-dataOffset)
		}
	}

	// Advance evacuation mark
	if oldbucket == h.nevacuate {
		h.nevacuate = oldbucket + 1
		if oldbucket+1 == newbit { // newbit == # of oldbuckets
			// Growing is all done.  Free old main bucket array.
			h.oldbuckets = nil
			// Can discard old overflow buckets as well.
			// If they are still referenced by an iterator,
			// then the iterator holds a pointers to the slice.
			if h.overflow != nil {
				h.overflow[1] = nil
			}
		}
	}
}

func ismapkey(t *_base.Type) bool {
	return t.Alg.Hash != nil
}

// Reflect stubs.  Called from ../reflect/asm_*.s

//go:linkname reflect_makemap reflect.makemap
func reflect_makemap(t *maptype) *hmap {
	return makemap(t, 0, nil, nil)
}

//go:linkname reflect_mapaccess reflect.mapaccess
func reflect_mapaccess(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	val, ok := mapaccess2(t, h, key)
	if !ok {
		// reflect wants nil for a missing element
		val = nil
	}
	return val
}

//go:linkname reflect_mapassign reflect.mapassign
func reflect_mapassign(t *maptype, h *hmap, key unsafe.Pointer, val unsafe.Pointer) {
	mapassign1(t, h, key, val)
}

//go:linkname reflect_mapdelete reflect.mapdelete
func reflect_mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
	mapdelete(t, h, key)
}

//go:linkname reflect_mapiterinit reflect.mapiterinit
func reflect_mapiterinit(t *maptype, h *hmap) *hiter {
	it := new(hiter)
	mapiterinit(t, h, it)
	return it
}

//go:linkname reflect_mapiternext reflect.mapiternext
func reflect_mapiternext(it *hiter) {
	mapiternext(it)
}

//go:linkname reflect_mapiterkey reflect.mapiterkey
func reflect_mapiterkey(it *hiter) unsafe.Pointer {
	return it.key
}

//go:linkname reflect_maplen reflect.maplen
func reflect_maplen(h *hmap) int {
	if h == nil {
		return 0
	}
	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&h))
		_race.Racereadpc(unsafe.Pointer(h), callerpc, _base.FuncPC(reflect_maplen))
	}
	return h.count
}

//go:linkname reflect_ismapkey reflect.ismapkey
func reflect_ismapkey(t *_base.Type) bool {
	return ismapkey(t)
}

var zerobuf struct {
	lock _base.Mutex
	p    *byte
	size uintptr
}

var zerotiny [1024]byte

// mapzero ensures that t.zero points at a zero value for type t.
// Types known to the compiler are in read-only memory and all point
// to a single zero in the bss of a large enough size.
// Types allocated by package reflect are in writable memory and
// start out with zero set to nil; we initialize those on demand.
func mapzero(t *_base.Type) {
	// On ARM, atomicloadp is implemented as xadd(p, 0),
	// so we cannot use atomicloadp on read-only memory.
	// Check whether the pointer is in the heap; if not, it's not writable
	// so the zero value must already be set.
	if _base.GOARCH == "arm" && !_base.Inheap(uintptr(unsafe.Pointer(t))) {
		if t.Zero == nil {
			print("runtime: map element ", *t.String, " missing zero value\n")
			_base.Throw("mapzero")
		}
		return
	}

	// Already done?
	// Check without lock, so must use atomicload to sync with atomicstore in allocation case below.
	if _iface.Atomicloadp(unsafe.Pointer(&t.Zero)) != nil {
		return
	}

	// Small enough for static buffer?
	if t.Size <= uintptr(len(zerotiny)) {
		_base.Atomicstorep(unsafe.Pointer(&t.Zero), unsafe.Pointer(&zerotiny[0]))
		return
	}

	// Use allocated buffer.
	_base.Lock(&zerobuf.lock)
	if zerobuf.size < t.Size {
		if zerobuf.size == 0 {
			zerobuf.size = 4 * 1024
		}
		for zerobuf.size < t.Size {
			zerobuf.size *= 2
			if zerobuf.size == 0 {
				// need >2GB zero on 32-bit machine
				_base.Throw("map element too large")
			}
		}
		zerobuf.p = (*byte)(_base.Persistentalloc(zerobuf.size, 64, &_base.Memstats.Other_sys))
	}
	_base.Atomicstorep(unsafe.Pointer(&t.Zero), unsafe.Pointer(zerobuf.p))
	_base.Unlock(&zerobuf.lock)
}
