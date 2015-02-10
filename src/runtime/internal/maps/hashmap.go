// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

const (
	// Maximum number of key/value pairs a bucket can hold.
	BucketCntBits = 3
	BucketCnt     = 1 << BucketCntBits

	// Maximum average load of a bucket that triggers growth.
	LoadFactor = 6.5

	// Maximum key or value size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	// Fast versions cannot handle big values - the cutoff size for
	// fast versions in ../../cmd/gc/walk.c must be at most this value.
	MaxKeySize   = 128
	MaxValueSize = 128

	// data offset should be the size of the bmap struct, but needs to be
	// aligned correctly.  For amd64p32 this means 64-bit alignment
	// even though pointers are 32 bit.
	DataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// Possible tophash values.  We reserve a few possibilities for special marks.
	// Each bucket (including its overflow buckets, if any) will have either all or none of its
	// entries in the evacuated* states (except during the evacuate() method, which only happens
	// during map writes and thus no one else can observe the map during that time).
	Empty          = 0 // cell is empty
	EvacuatedEmpty = 1 // cell is empty, bucket is evacuated.
	EvacuatedX     = 2 // key/value is valid.  Entry has been evacuated to first half of larger table.
	EvacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	MinTopHash     = 4 // minimum tophash for a normal filled cell.

	// flags
	Iterator    = 1 // there may be an iterator using buckets
	OldIterator = 2 // there may be an iterator using oldbuckets

	// sentinel bucket ID for iterator checks
	NoCheck = 1<<(8*_core.PtrSize) - 1

	// trigger a garbage collection at every alloc called from this code
	Checkgc = false
)

// A header for a Go map.
type Hmap struct {
	// Note: the format of the Hmap is encoded in ../../cmd/gc/reflect.c and
	// ../reflect/type.go.  Don't change this structure without also changing that code!
	Count int // # live cells == size of map.  Must be first (used by len() builtin)
	flags uint8
	B     uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	hash0 uint32 // hash seed

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	// If both key and value do not contain pointers, then we mark bucket
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
	tophash [BucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt values.
	// NOTE: packing all the keys together and then all the values together makes the
	// code a bit more complicated than alternating key/value/key/value/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}

// A hash iteration structure.
// If you modify hiter, also change cmd/gc/reflect.c to indicate
// the layout of this structure.
type Hiter struct {
	Key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/gc/range.c).
	value       unsafe.Pointer // Must be in second position (see cmd/gc/range.c).
	t           *Maptype
	h           *Hmap
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
	return h > Empty && h < MinTopHash
}

func (b *bmap) overflow(t *Maptype) *bmap {
	return *(**bmap)(_core.Add(unsafe.Pointer(b), uintptr(t.bucketsize)-_lock.RegSize))
}

func (h *Hmap) setoverflow(t *Maptype, b, ovf *bmap) {
	if t.bucket.Kind&_channels.KindNoPointers != 0 {
		h.createOverflow()
		*h.overflow[0] = append(*h.overflow[0], ovf)
	}
	*(**bmap)(_core.Add(unsafe.Pointer(b), uintptr(t.bucketsize)-_lock.RegSize)) = ovf
}

func (h *Hmap) createOverflow() {
	if h.overflow == nil {
		h.overflow = new([2]*[]*bmap)
	}
	if h.overflow[0] == nil {
		h.overflow[0] = new([]*bmap)
	}
}

func Makemap(t *Maptype, hint int64) *Hmap {
	if sz := unsafe.Sizeof(Hmap{}); sz > 48 || sz != uintptr(t.hmap.Size) {
		_lock.Throw("bad hmap size")
	}

	if hint < 0 || int64(int32(hint)) != hint {
		panic("makemap: size out of range")
		// TODO: make hint an int, then none of this nonsense
	}

	if !Ismapkey(t.key) {
		_lock.Throw("runtime.makemap: unsupported map key type")
	}

	// check compiler's and reflect's math
	if t.key.Size > MaxKeySize && (!t.indirectkey || t.keysize != uint8(_core.PtrSize)) ||
		t.key.Size <= MaxKeySize && (t.indirectkey || t.keysize != uint8(t.key.Size)) {
		_lock.Throw("key size wrong")
	}
	if t.elem.Size > MaxValueSize && (!t.indirectvalue || t.valuesize != uint8(_core.PtrSize)) ||
		t.elem.Size <= MaxValueSize && (t.indirectvalue || t.valuesize != uint8(t.elem.Size)) {
		_lock.Throw("value size wrong")
	}

	// invariants we depend on.  We should probably check these at compile time
	// somewhere, but for now we'll do it here.
	if t.key.Align > BucketCnt {
		_lock.Throw("key align too big")
	}
	if t.elem.Align > BucketCnt {
		_lock.Throw("value align too big")
	}
	if uintptr(t.key.Size)%uintptr(t.key.Align) != 0 {
		_lock.Throw("key size not a multiple of key align")
	}
	if uintptr(t.elem.Size)%uintptr(t.elem.Align) != 0 {
		_lock.Throw("value size not a multiple of value align")
	}
	if BucketCnt < 8 {
		_lock.Throw("bucketsize too small for proper alignment")
	}
	if DataOffset%uintptr(t.key.Align) != 0 {
		_lock.Throw("need padding in bucket (key)")
	}
	if DataOffset%uintptr(t.elem.Align) != 0 {
		_lock.Throw("need padding in bucket (value)")
	}

	// find size parameter which will hold the requested # of elements
	B := uint8(0)
	for ; hint > BucketCnt && float32(hint) > LoadFactor*float32(uintptr(1)<<B); B++ {
	}

	// allocate initial hash table
	// if B == 0, the buckets field is allocated lazily later (in mapassign)
	// If hint is large zeroing this memory could take a while.
	var buckets unsafe.Pointer
	if B != 0 {
		if Checkgc {
			_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
		}
		buckets = Newarray(t.bucket, uintptr(1)<<B)
	}

	// initialize Hmap
	if Checkgc {
		_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
	}
	h := (*Hmap)(Newobject(t.hmap))
	h.Count = 0
	h.B = B
	h.flags = 0
	h.hash0 = _lock.Fastrand1()
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
func mapaccess1(t *Maptype, h *Hmap, key unsafe.Pointer) unsafe.Pointer {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		pc := _lock.FuncPC(mapaccess1)
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, pc)
		_channels.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.Count == 0 {
		return unsafe.Pointer(t.elem.Zero)
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
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
			if b.tophash[i] != top {
				continue
			}
			k := _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
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

func Mapaccess2(t *Maptype, h *Hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		pc := _lock.FuncPC(Mapaccess2)
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, pc)
		_channels.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.Count == 0 {
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
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
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
func mapaccessK(t *Maptype, h *Hmap, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer) {
	if h == nil || h.Count == 0 {
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
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
			if t.indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
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

func Mapassign1(t *Maptype, h *Hmap, key unsafe.Pointer, val unsafe.Pointer) {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		pc := _lock.FuncPC(Mapassign1)
		Racewritepc(unsafe.Pointer(h), callerpc, pc)
		_channels.RaceReadObjectPC(t.key, key, callerpc, pc)
		_channels.RaceReadObjectPC(t.elem, val, callerpc, pc)
	}

	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))

	if h.buckets == nil {
		if Checkgc {
			_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
		}
		h.buckets = Newarray(t.bucket, 1)
	}

again:
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.oldbuckets != nil {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}

	var inserti *uint8
	var insertk unsafe.Pointer
	var insertv unsafe.Pointer
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			if b.tophash[i] != top {
				if b.tophash[i] == Empty && inserti == nil {
					inserti = &b.tophash[i]
					insertk = _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
					insertv = _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
				}
				continue
			}
			k := _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
			k2 := k
			if t.indirectkey {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if !alg.Equal(key, k2) {
				continue
			}
			// already have a mapping for key.  Update it.
			_channels.Typedmemmove(t.key, k2, key)
			v := _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
			v2 := v
			if t.indirectvalue {
				v2 = *((*unsafe.Pointer)(v2))
			}
			_channels.Typedmemmove(t.elem, v2, val)
			return
		}
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// did not find mapping for key.  Allocate new cell & add entry.
	if float32(h.Count) >= LoadFactor*float32((uintptr(1)<<h.B)) && h.Count >= BucketCnt {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if inserti == nil {
		// all current buckets are full, allocate a new one.
		if Checkgc {
			_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
		}
		newb := (*bmap)(Newobject(t.bucket))
		h.setoverflow(t, b, newb)
		inserti = &newb.tophash[0]
		insertk = _core.Add(unsafe.Pointer(newb), DataOffset)
		insertv = _core.Add(insertk, BucketCnt*uintptr(t.keysize))
	}

	// store new key/value at insert position
	if t.indirectkey {
		if Checkgc {
			_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
		}
		kmem := Newobject(t.key)
		*(*unsafe.Pointer)(insertk) = kmem
		insertk = kmem
	}
	if t.indirectvalue {
		if Checkgc {
			_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
		}
		vmem := Newobject(t.elem)
		*(*unsafe.Pointer)(insertv) = vmem
		insertv = vmem
	}
	_channels.Typedmemmove(t.key, insertk, key)
	_channels.Typedmemmove(t.elem, insertv, val)
	*inserti = top
	h.Count++
}

func Mapdelete(t *Maptype, h *Hmap, key unsafe.Pointer) {
	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		pc := _lock.FuncPC(Mapdelete)
		Racewritepc(unsafe.Pointer(h), callerpc, pc)
		_channels.RaceReadObjectPC(t.key, key, callerpc, pc)
	}
	if h == nil || h.Count == 0 {
		return
	}
	alg := t.key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.oldbuckets != nil {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := uint8(hash >> (_core.PtrSize*8 - 8))
	if top < MinTopHash {
		top += MinTopHash
	}
	for {
		for i := uintptr(0); i < BucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := _core.Add(unsafe.Pointer(b), DataOffset+i*uintptr(t.keysize))
			k2 := k
			if t.indirectkey {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if !alg.Equal(key, k2) {
				continue
			}
			_core.Memclr(k, uintptr(t.keysize))
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + DataOffset + BucketCnt*uintptr(t.keysize) + i*uintptr(t.valuesize))
			_core.Memclr(v, uintptr(t.valuesize))
			b.tophash[i] = Empty
			h.Count--
			return
		}
		b = b.overflow(t)
		if b == nil {
			return
		}
	}
}

func Mapiterinit(t *Maptype, h *Hmap, it *Hiter) {
	// Clear pointer fields so garbage collector does not complain.
	it.Key = nil
	it.value = nil
	it.t = nil
	it.h = nil
	it.buckets = nil
	it.bptr = nil
	it.overflow[0] = nil
	it.overflow[1] = nil

	if _sched.Raceenabled && h != nil {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(Mapiterinit))
	}

	if h == nil || h.Count == 0 {
		it.Key = nil
		it.value = nil
		return
	}

	if unsafe.Sizeof(Hiter{})/_core.PtrSize != 12 {
		_lock.Throw("hash_iter size incorrect") // see ../../cmd/gc/reflect.c
	}
	it.t = t
	it.h = h

	// grab snapshot of bucket state
	it.B = h.B
	it.buckets = h.buckets
	if t.bucket.Kind&_channels.KindNoPointers != 0 {
		// Allocate the current slice and remember pointers to both current and old.
		// This preserves all relevant overflow buckets alive even if
		// the table grows and/or overflow buckets are added to the table
		// while we are iterating.
		h.createOverflow()
		it.overflow = *h.overflow
	}

	// decide where to start
	r := uintptr(_lock.Fastrand1())
	if h.B > 31-BucketCntBits {
		r += uintptr(_lock.Fastrand1()) << 31
	}
	it.startBucket = r & (uintptr(1)<<h.B - 1)
	it.offset = uint8(r >> h.B & (BucketCnt - 1))

	// iterator state
	it.bucket = it.startBucket
	it.wrapped = false
	it.bptr = nil

	// Remember we have an iterator.
	// Can run concurrently with another hash_iter_init().
	if old := h.flags; old&(Iterator|OldIterator) != Iterator|OldIterator {
		_sched.Atomicor8(&h.flags, Iterator|OldIterator)
	}

	mapiternext(it)
}

func mapiternext(it *Hiter) {
	h := it.h
	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&it))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(mapiternext))
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
			it.Key = nil
			it.value = nil
			return
		}
		if h.oldbuckets != nil && it.B == h.B {
			// Iterator was started in the middle of a grow, and the grow isn't done yet.
			// If the bucket we're looking at hasn't been filled in yet (i.e. the old
			// bucket hasn't been evacuated) then we need to iterate through the old
			// bucket and only return the ones that will be migrated to this bucket.
			oldbucket := bucket & (uintptr(1)<<(it.B-1) - 1)
			b = (*bmap)(_core.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
			if !evacuated(b) {
				checkBucket = bucket
			} else {
				b = (*bmap)(_core.Add(it.buckets, bucket*uintptr(t.bucketsize)))
				checkBucket = NoCheck
			}
		} else {
			b = (*bmap)(_core.Add(it.buckets, bucket*uintptr(t.bucketsize)))
			checkBucket = NoCheck
		}
		bucket++
		if bucket == uintptr(1)<<it.B {
			bucket = 0
			it.wrapped = true
		}
		i = 0
	}
	for ; i < BucketCnt; i++ {
		offi := (i + it.offset) & (BucketCnt - 1)
		k := _core.Add(unsafe.Pointer(b), DataOffset+uintptr(offi)*uintptr(t.keysize))
		v := _core.Add(unsafe.Pointer(b), DataOffset+BucketCnt*uintptr(t.keysize)+uintptr(offi)*uintptr(t.valuesize))
		if b.tophash[offi] != Empty && b.tophash[offi] != EvacuatedEmpty {
			if checkBucket != NoCheck {
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
			if b.tophash[offi] != EvacuatedX && b.tophash[offi] != EvacuatedY {
				// this is the golden data, we can return it.
				if t.indirectkey {
					k = *((*unsafe.Pointer)(k))
				}
				it.Key = k
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
					it.Key = rk
					it.value = rv
				} else {
					// if key!=key then the entry can't be deleted or
					// updated, so we can just return it.  That's lucky for
					// us because when key!=key we can't look it up
					// successfully in the current table.
					it.Key = k2
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

func hashGrow(t *Maptype, h *Hmap) {
	if h.oldbuckets != nil {
		_lock.Throw("evacuation not done in time")
	}
	oldbuckets := h.buckets
	if Checkgc {
		_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
	}
	newbuckets := Newarray(t.bucket, uintptr(1)<<(h.B+1))
	flags := h.flags &^ (Iterator | OldIterator)
	if h.flags&Iterator != 0 {
		flags |= OldIterator
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
			_lock.Throw("overflow is not nil")
		}
		h.overflow[1] = h.overflow[0]
		h.overflow[0] = nil
	}

	// the actual copying of the hash table data is done incrementally
	// by growWork() and evacuate().
}

func growWork(t *Maptype, h *Hmap, bucket uintptr) {
	noldbuckets := uintptr(1) << (h.B - 1)

	// make sure we evacuate the oldbucket corresponding
	// to the bucket we're about to use
	evacuate(t, h, bucket&(noldbuckets-1))

	// evacuate one more oldbucket to make progress on growing
	if h.oldbuckets != nil {
		evacuate(t, h, h.nevacuate)
	}
}

func evacuate(t *Maptype, h *Hmap, oldbucket uintptr) {
	b := (*bmap)(_core.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
	newbit := uintptr(1) << (h.B - 1)
	alg := t.key.Alg
	if !evacuated(b) {
		// TODO: reuse overflow buckets instead of using new ones, if there
		// is no iterator using the old buckets.  (If !oldIterator.)

		x := (*bmap)(_core.Add(h.buckets, oldbucket*uintptr(t.bucketsize)))
		y := (*bmap)(_core.Add(h.buckets, (oldbucket+newbit)*uintptr(t.bucketsize)))
		xi := 0
		yi := 0
		xk := _core.Add(unsafe.Pointer(x), DataOffset)
		yk := _core.Add(unsafe.Pointer(y), DataOffset)
		xv := _core.Add(xk, BucketCnt*uintptr(t.keysize))
		yv := _core.Add(yk, BucketCnt*uintptr(t.keysize))
		for ; b != nil; b = b.overflow(t) {
			k := _core.Add(unsafe.Pointer(b), DataOffset)
			v := _core.Add(k, BucketCnt*uintptr(t.keysize))
			for i := 0; i < BucketCnt; i, k, v = i+1, _core.Add(k, uintptr(t.keysize)), _core.Add(v, uintptr(t.valuesize)) {
				top := b.tophash[i]
				if top == Empty {
					b.tophash[i] = EvacuatedEmpty
					continue
				}
				if top < MinTopHash {
					_lock.Throw("bad map state")
				}
				k2 := k
				if t.indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				// Compute hash to make our evacuation decision (whether we need
				// to send this key/value to bucket x or bucket y).
				hash := alg.Hash(k2, uintptr(h.hash0))
				if h.flags&Iterator != 0 {
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
						top = uint8(hash >> (_core.PtrSize*8 - 8))
						if top < MinTopHash {
							top += MinTopHash
						}
					}
				}
				if (hash & newbit) == 0 {
					b.tophash[i] = EvacuatedX
					if xi == BucketCnt {
						if Checkgc {
							_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
						}
						newx := (*bmap)(Newobject(t.bucket))
						h.setoverflow(t, x, newx)
						x = newx
						xi = 0
						xk = _core.Add(unsafe.Pointer(x), DataOffset)
						xv = _core.Add(xk, BucketCnt*uintptr(t.keysize))
					}
					x.tophash[xi] = top
					if t.indirectkey {
						*(*unsafe.Pointer)(xk) = k2 // copy pointer
					} else {
						_channels.Typedmemmove(t.key, xk, k) // copy value
					}
					if t.indirectvalue {
						*(*unsafe.Pointer)(xv) = *(*unsafe.Pointer)(v)
					} else {
						_channels.Typedmemmove(t.elem, xv, v)
					}
					xi++
					xk = _core.Add(xk, uintptr(t.keysize))
					xv = _core.Add(xv, uintptr(t.valuesize))
				} else {
					b.tophash[i] = EvacuatedY
					if yi == BucketCnt {
						if Checkgc {
							_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc
						}
						newy := (*bmap)(Newobject(t.bucket))
						h.setoverflow(t, y, newy)
						y = newy
						yi = 0
						yk = _core.Add(unsafe.Pointer(y), DataOffset)
						yv = _core.Add(yk, BucketCnt*uintptr(t.keysize))
					}
					y.tophash[yi] = top
					if t.indirectkey {
						*(*unsafe.Pointer)(yk) = k2
					} else {
						_channels.Typedmemmove(t.key, yk, k)
					}
					if t.indirectvalue {
						*(*unsafe.Pointer)(yv) = *(*unsafe.Pointer)(v)
					} else {
						_channels.Typedmemmove(t.elem, yv, v)
					}
					yi++
					yk = _core.Add(yk, uintptr(t.keysize))
					yv = _core.Add(yv, uintptr(t.valuesize))
				}
			}
		}
		// Unlink the overflow buckets & clear key/value to help GC.
		if h.flags&OldIterator == 0 {
			b = (*bmap)(_core.Add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
			_core.Memclr(_core.Add(unsafe.Pointer(b), DataOffset), uintptr(t.bucketsize)-DataOffset)
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

func Ismapkey(t *_core.Type) bool {
	return t.Alg.Hash != nil
}

//go:linkname reflect_mapiternext reflect.mapiternext
func reflect_mapiternext(it *Hiter) {
	mapiternext(it)
}
