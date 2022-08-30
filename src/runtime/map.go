// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// This file contains the implementation of Go's map type.
//
// The implementation is mainly based on SwissTable
// (https://abseil.io/about/design/swisstables).
//
// A map is just a hash table. The data is arranged
// into an array of buckets. Each bucket contains up to
// 8 key/elem pairs. The hash are used to select a bucket.
// Each bucket contains high-order 7 bits of each hash to
// distinguish the entries within a single bucket.
//
// When the hashtable grows, we allocate a new array
// of buckets twice as big. Buckets are incrementally
// copied from the old bucket array to the new bucket array.
//
// Map iterators walk through the array of buckets and
// return the keys in walk order (bucket #, then the next
// bucket index).  To maintain iteration  semantics, we only
// move keys within their bucket if without any iterators.
// When growing the table, iterators remain iterating through
// the old table and must check the new table if the bucket
// they are iterating through has beenmoved to the new table.

import (
	"internal/abi"
	"internal/goarch"
	"runtime/internal/atomic"
	"runtime/internal/math"
	"unsafe"
)

const (
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits

	// Maximum key or elem size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	// Fast versions cannot handle big elems - the cutoff size for
	// fast versions in cmd/compile/internal/gc/walk.go must be at most this elem.
	maxKeySize  = 128
	maxElemSize = 128

	// data offset should be the size of the bmap struct, but needs to be
	// aligned correctly. For amd64p32 this means 64-bit alignment
	// even though pointers are 32 bit.
	dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// Possible tophash values.
	// fullSlot uint8 = 0b0xxx_xxxx, the xxx_xxxx is the top 7 bits of the hash.
	emptySlot          uint8  = 0b1111_1111
	deletedSlot        uint8  = 0b1000_0000
	emptyOrDeletedMask uint64 = 0x8080_8080_8080_8080

	// flags
	iterator    = 1 // there may be an iterator using buckets
	hashWriting = 2 // a goroutine is writing to the map

	// LoadFactor.
	// 0 -> invalid
	// 1 -> 0.5(min)
	// 2 -> 0.75
	// 3 -> 0.875(max)
	// 4 -> 0.9375, invalid if bucketCnt=8, we need to make sure at least one empty slot.
	loadFactorShift = 3

	// Maximum number of key/elem pairs a bucket can hold for maximally loaded tables.
	maxItemInBucket   = bucketCnt - (bucketCnt >> loadFactorShift)
	emptyItemInBucket = bucketCnt >> loadFactorShift // used for optimazation

	// Use for loop to traverse values in the buckets if B <= loopaccessB.
	loopaccessB = 1
)

// isFull returns true if the slot is full.
func isFull(v uint8) bool {
	return v < 128
}

// A header for a Go map.
type hmap struct {
	// Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
	// Make sure this stays in sync with the compiler's definition.
	count      int // live cells == size of map.  Must be first (used by len() builtin)
	flags      uint8
	B          uint8 // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	_pad       uint16
	hash0      uint32         // hash seed
	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	growthLeft int
}

// A bucket for a Go map.
type bmap struct {
	// tophash generally contains the top byte of the hash value
	// for each key in this bucket.
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt elems.
	// NOTE: packing all the keys together and then all the elems together makes the
	// code a bit more complicated than alternating key/elem/key/elem/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
}

func (b *bmap) keys() unsafe.Pointer {
	return add(unsafe.Pointer(b), dataOffset)
}

// A hash iteration structure.
// If you modify hiter, also change cmd/compile/internal/reflectdata/reflect.go
// and reflect/value.go to match the layout of this structure.
type hiter struct {
	key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go).
	elem        unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go).
	t           *maptype
	h           *hmap
	buckets     unsafe.Pointer // bucket ptr at hash_iter initialization time
	bptr        *bmap          // current bucket
	startBucket uintptr        // bucket iteration started at
	offset      uint8          // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	wrapped     bool           // already wrapped around from end of bucket array to beginning
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

// probe represents the state of a probe sequence.
//
// The sequence is a triangular progression.
type probe struct {
	bucket uintptr
	mask   uintptr
	stride uintptr
}

func newProbe(hash, mask uintptr) *probe {
	return &probe{
		bucket: hash & mask,
		mask:   mask,
	}
}

func (p *probe) Bucket() uintptr {
	return p.bucket
}

func (p *probe) Next() {
	p.stride += 1
	p.bucket += p.stride
	p.bucket &= p.mask
}

func (p *probe) Reset(hash uintptr) {
	p.stride = 0
	p.bucket = hash & p.mask
}

// bucketShift returns 1<<b, optimized for code generation.
func bucketShift(b uint8) uintptr {
	// Masking the shift amount allows overflow checks to be elided.
	return uintptr(1) << (b & (goarch.PtrSize*8 - 1))
}

// bucketMask returns 1<<b - 1, optimized for code generation.
func bucketMask(b uint8) uintptr {
	return bucketShift(b) - 1
}

// tophash calculates the tophash value for hash.
func tophash(hash uintptr) uint8 {
	return uint8(hash >> (goarch.PtrSize*8 - 7))
}

func overLoadFactor(count int, B uint8) bool {
	return count > bucketCnt && uintptr(count) > bucketShift(B)*maxItemInBucket
}

// makemap_small implements Go map creation for make(map[k]v) and
// make(map[k]v, hint) when hint is known to be at most bucketCnt
// at compile time and the map needs to be allocated on the heap.
func makemap_small() *hmap {
	h := new(hmap)
	h.hash0 = fastrand()
	h.growthLeft = bucketCnt
	return h
}

func makemap64(t *maptype, hint int64, h *hmap) *hmap {
	if int64(int(hint)) != hint {
		hint = 0
	}
	return makemap(t, int(hint), h)
}

// makemap implements Go map creation for make(map[k]v, hint).
// If the compiler has determined that the map or the first bucket
// can be created on the stack, h and/or bucket may be non-nil.
// If h != nil, the map can be created directly in h.
// If h.buckets != nil, bucket pointed to can be used as the first bucket.
func makemap(t *maptype, hint int, h *hmap) *hmap {
	mem, overflow := math.MulUintptr(uintptr(hint), t.bucket.size)
	if overflow || mem > maxAlloc {
		hint = 0
	}

	// initialize Hmap
	if h == nil {
		h = new(hmap)
	}
	h.hash0 = fastrand()

	// Find the size parameter B which will hold the requested # of elements.
	// For hint < 0 overLoadFactor returns false since hint < bucketCnt.
	B := uint8(0)
	for overLoadFactor(hint, B) {
		B++
	}
	h.B = B
	h.growthLeft = bucketCnt * int(bucketShift(B))

	// allocate initial hash table
	// if B == 0, the buckets field is allocated lazily later (in mapassign)
	// If hint is large zeroing this memory could take a while.
	if h.B != 0 {
		h.buckets = makeBucketArray(t, h.B)
	}

	return h
}

// makeBucketArray initializes a backing array for map buckets.
// 1<<b is the number of buckets to allocate.
func makeBucketArray(t *maptype, b uint8) unsafe.Pointer {
	nbuckets := bucketShift(b)

	buckets := newarray(t.bucket, int(nbuckets))

	base := buckets
	for i := uintptr(0); i < nbuckets; i++ {
		*(*uint64)(base) = 0xffff_ffff_ffff_ffff // all empty
		base = add(base, uintptr(t.bucketsize))
	}

	return buckets
}

// mapaccess1 returns a pointer to h[key].  Never returns nil, instead
// it will return a reference to the zero object for the elem type if
// the key is not in the map.
// NOTE: The returned pointer may keep the whole map live, so don't
// hold onto it for very long.
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		pc := abi.FuncPCABIInternal(mapaccess1)
		racereadpc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	if msanenabled && h != nil {
		msanread(key, t.key.size)
	}
	if asanenabled && h != nil {
		asanread(key, t.key.size)
	}
	if h == nil || h.count == 0 {
		if t.hashMightPanic() {
			t.hasher(key, 0) // see issue 23734
		}
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}

	hash := t.hasher(key, uintptr(h.hash0))
	top := tophash(hash)

	p := newProbe(hash, bucketMask(h.B))

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			if t.key.equal(key, k) {
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					e = *((*unsafe.Pointer)(e))
				}
				return e
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0])
		}
		p.Next()
	}
}

func mapaccess2(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		pc := abi.FuncPCABIInternal(mapaccess2)
		racereadpc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	if msanenabled && h != nil {
		msanread(key, t.key.size)
	}
	if asanenabled && h != nil {
		asanread(key, t.key.size)
	}
	if h == nil || h.count == 0 {
		if t.hashMightPanic() {
			t.hasher(key, 0) // see issue 23734
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}

	hash := t.hasher(key, uintptr(h.hash0))
	top := tophash(hash)

	p := newProbe(hash, bucketMask(h.B))

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			if t.key.equal(key, k) {
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					e = *((*unsafe.Pointer)(e))
				}
				return e, true
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0]), false
		}
		p.Next()
	}
}

// returns both key and elem. Used by map iterator
func mapaccessK(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer) {
	if h == nil || h.count == 0 {
		return nil, nil
	}
	hash := t.hasher(key, uintptr(h.hash0))
	top := tophash(hash)

	p := newProbe(hash, bucketMask(h.B))

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			if t.key.equal(key, k) {
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					e = *((*unsafe.Pointer)(e))
				}
				return k, e
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return nil, nil
		}
		p.Next()
	}
}

func mapaccess1_fat(t *maptype, h *hmap, key, zero unsafe.Pointer) unsafe.Pointer {
	e := mapaccess1(t, h, key)
	if e == unsafe.Pointer(&zeroVal[0]) {
		return zero
	}
	return e
}

func mapaccess2_fat(t *maptype, h *hmap, key, zero unsafe.Pointer) (unsafe.Pointer, bool) {
	e := mapaccess1(t, h, key)
	if e == unsafe.Pointer(&zeroVal[0]) {
		return zero, false
	}
	return e, true
}

// Like mapaccess, but allocates a slot for the key if it is not present in the map.
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		pc := abi.FuncPCABIInternal(mapassign)
		racewritepc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	if msanenabled {
		msanread(key, t.key.size)
	}
	if asanenabled {
		asanread(key, t.key.size)
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}
	hash := t.hasher(key, uintptr(h.hash0))

	// Set hashWriting after calling t.hasher, since t.hasher may panic,
	// in which case we have not actually done a write.
	h.flags ^= hashWriting

	if h.buckets == nil {
		// Init an empty map.
		h.buckets = makeBucketArray(t, 0)
		h.growthLeft = bucketCnt
	}

	top := tophash(hash)

	if h.needGrow() {
		grow(h, t)
	}

	p := newProbe(hash, bucketMask(h.B))

	var inserti *uint8
	var insertk unsafe.Pointer
	var elem unsafe.Pointer

	var (
		b      *bmap
		status bitmask64
	)
	// Check if the key in the map.
	for {
		b = (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status = matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			if t.key.equal(key, k) {
				// Found a key.
				// already have a mapping for key. Update it.
				if t.needkeyupdate() {
					typedmemmove(t.key, k, key)
				}
				elem = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				goto done
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			break
		}
		p.Next()
	}

	// The key is not in the map.
	p.Reset(hash)
	for {
		b = (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		// Can't find the key in this bucket.
		// Check empty slot or deleted slot.
		status = matchEmptyOrDeleted(b.tophash)
		i := status.NextMatch()
		if i < bucketCnt {
			inserti = &b.tophash[i]
			insertk = add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			elem = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
			// Insert key and value.
			if t.indirectkey() {
				kmem := newobject(t.key)
				*(*unsafe.Pointer)(insertk) = kmem
				insertk = kmem
			}
			if t.indirectelem() {
				vmem := newobject(t.elem)
				*(*unsafe.Pointer)(elem) = vmem
			}
			typedmemmove(t.key, insertk, key)
			*inserti = top
			h.growthLeft -= 1
			h.count += 1
			goto done
		}
		p.Next()
	}
done:
	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
	if t.indirectelem() {
		elem = *((*unsafe.Pointer)(elem))
	}
	return elem
}

func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		pc := abi.FuncPCABIInternal(mapdelete)
		racewritepc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	if msanenabled && h != nil {
		msanread(key, t.key.size)
	}
	if asanenabled && h != nil {
		asanread(key, t.key.size)
	}
	if h == nil || h.count == 0 {
		if t.hashMightPanic() {
			t.hasher(key, 0) // see issue 23734
		}
		return
	}

	if h.buckets == nil {
		return
	}

	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}

	hash := t.hasher(key, uintptr(h.hash0))

	// Set hashWriting after calling t.hasher, since t.hasher may panic,
	// in which case we have not actually done a write (delete).
	h.flags ^= hashWriting

	p := newProbe(hash, bucketMask(h.B))
	top := tophash(hash)

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			k2 := k
			if t.indirectkey() {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if t.key.equal(key, k2) {
				// Found this key.
				h.count -= 1
				// Only clear key if there are pointers in it.
				if t.indirectkey() {
					*(*unsafe.Pointer)(k) = nil
				} else if t.key.ptrdata != 0 {
					memclrHasPointers(k, t.key.size)
				}
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					*(*unsafe.Pointer)(e) = nil
				} else if t.elem.ptrdata != 0 {
					memclrHasPointers(e, t.elem.size)
				} else {
					memclrNoHeapPointers(e, t.elem.size)
				}
				// Update tophash.
				if matchEmpty(b.tophash) == 0 {
					// We only ever mark the slot as deleted if the entry we want to delete
					// is in a pack of bucketCnt non-EMPTY buckets.
					b.tophash[i] = deletedSlot
				} else {
					h.growthLeft += 1
					b.tophash[i] = emptySlot
				}
				goto done
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			// The key is not in this map.
			goto done
		}
		p.Next()
	}
done:
	if h.count == 0 {
		// Reset the hash seed to make it more difficult for attackers to
		// repeatedly trigger hash collisions. See issue 25237.
		h.hash0 = fastrand()
	}
	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
}

// mapiterinit initializes the hiter struct used for ranging over maps.
// The hiter struct pointed to by 'it' is allocated on the stack
// by the compilers order pass or on the heap by reflect_mapiterinit.
// Both need to have zeroed hiter since the struct contains pointers.
func mapiterinit(t *maptype, h *hmap, it *hiter) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapiterinit))
	}

	it.t = t
	if h == nil || h.count == 0 {
		return
	}

	if unsafe.Sizeof(hiter{})/goarch.PtrSize != 10 {
		throw("hash_iter size incorrect") // see cmd/compile/internal/reflectdata/reflect.go
	}
	it.h = h

	// grab snapshot of bucket state
	it.B = h.B
	it.buckets = h.buckets

	// decide where to start
	var r uintptr
	if h.B > 31-bucketCntBits {
		r = uintptr(fastrand64())
	} else {
		r = uintptr(fastrand())
	}
	it.startBucket = r & bucketMask(h.B)
	it.offset = uint8(r >> h.B & (bucketCnt - 1))

	// iterator state
	it.bucket = it.startBucket

	// Remember we have an iterator.
	// Can run concurrently with another mapiterinit().
	if old := h.flags; old&(iterator) != iterator {
		atomic.Or8(&h.flags, iterator)
	}

	mapiternext(it)
}

func mapiternext(it *hiter) {
	h := it.h
	if raceenabled {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapiternext))
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map iteration and map write")
	}
	t := it.t
	bucket := it.bucket
	b := it.bptr
	i := it.i
	if b != nil {
		goto bucketloop
	}
nextbucket:
	if it.wrapped && bucket == it.startBucket {
		// end of iteration
		it.key = nil
		it.elem = nil
		return
	}
	b = (*bmap)(add(it.buckets, bucket*uintptr(t.bucketsize)))
	bucket++
	if bucket == bucketShift(it.B) {
		bucket = 0
		it.wrapped = true
	}
bucketloop:
	for ; i < bucketCnt; i++ {
		offi := (i + it.offset) & (bucketCnt - 1)
		if !isFull(b.tophash[offi]) {
			continue
		}
		k := add(unsafe.Pointer(b), dataOffset+uintptr(offi)*uintptr(t.keysize))
		if t.indirectkey() {
			k = *((*unsafe.Pointer)(k))
		}
		e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+uintptr(offi)*uintptr(t.elemsize))

		if h.B == it.B || !(t.reflexivekey() || t.key.equal(k, k)) {
			// This is the golden data, we can return it.
			// OR
			// key!=key, so the entry can't be deleted or updated, so we can just return it.
			// That's lucky for us because when key!=key we can't look it up successfully.
			it.key = k
			if t.indirectelem() {
				e = *((*unsafe.Pointer)(e))
			}
			it.elem = e
		} else {
			// The hash table has grown since the iterator was started.
			// The golden data for this key is now somewhere else.
			// Check the current hash table for the data.
			// This code handles the case where the key
			// has been deleted, updated, or deleted and reinserted.
			// NOTE: we need to regrab the key as it has potentially been
			// updated to an equal() but not identical key (e.g. +0.0 vs -0.0).
			rk, re := mapaccessK(t, h, k)
			if rk == nil {
				continue // key has been deleted
			}
			it.key = rk
			it.elem = re
		}
		// Got a data, save info into it.
		it.bucket = bucket
		if it.bptr != b { // avoid unnecessary write barrier; see issue 14921
			it.bptr = b
		}
		it.i = i + 1
		return
	}
	// No valid item in this bucket.
	i = 0
	goto nextbucket
}

// mapclear deletes all keys from a map.
func mapclear(t *maptype, h *hmap) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		pc := abi.FuncPCABIInternal(mapclear)
		racewritepc(unsafe.Pointer(h), callerpc, pc)
	}

	if h == nil || h.count == 0 {
		return
	}

	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}

	h.flags ^= hashWriting

	h.count = 0

	// Clear bucket.
	nbuckets := bucketShift(h.B)
	h.growthLeft = int(nbuckets * bucketCnt)

	buckets := h.buckets
	size := uintptr(t.bucketsize) * nbuckets

	// Only clear buckets if there are pointers in it.
	if t.bucket.ptrdata != 0 {
		memclrHasPointers(buckets, size)
	}

	for bptr := h.buckets; uintptr(bptr) < uintptr(add(h.buckets, size)); bptr = add(bptr, uintptr(t.bucketsize)) {
		*(*uint64)(bptr) = 0xffff_ffff_ffff_ffff
	}

	// Reset the hash seed to make it more difficult for attackers to
	// repeatedly trigger hash collisions. See issue 25237.
	h.hash0 = fastrand()

	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
}

func (h *hmap) needGrow() bool {
	threshold := int(bucketShift(h.B) * emptyItemInBucket) // (capcity - capcity*loadFactor)
	return h.growthLeft <= threshold
}

func grow(h *hmap, t *maptype) {
	cap := bucketShift(h.B) * bucketCnt
	if uintptr(h.count*32) <= cap*25 && (h.flags&iterator != iterator) {
		// Rehash in place if the current size is <= 25/32 of capacity.
		// If there may be an iterator using buckets, we disable growsamesize.
		// Because it may move data to different buckets, this behavior will
		// break the iterator(keys might be returned 0 or 2 times).
		// TODO(zyh): We can use a field to indicate the number of iterators
		// instead of just setting a flag. If that, we can still do growsamesize
		// after a traversal.
		growsamesize(h, t)
	} else {
		growbig(h, t)
	}
}

func growsamesize(h *hmap, t *maptype) {
	bucketNum := bucketShift(h.B)
	mask := bucketNum - 1
	indirectk := t.indirectkey()
	indirecte := t.indirectelem()
	// For all buckets:
	// - mark all DELETED slots as EMPTY
	// - mark all FULL slots as DELETED
	for bucket := uintptr(0); bucket < bucketNum; bucket++ {
		b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
		prepareSameSizeGrow(&b.tophash)
	}
	// Temporary key and value used to swap.
	var (
		inittmp    bool
		tmpk, tmpe unsafe.Pointer
	)

	for bucket := uintptr(0); bucket < bucketNum; bucket++ {
		b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
		for i := uintptr(0); i < bucketCnt; {
			if b.tophash[i] != deletedSlot {
				i++
				continue
			}
			base := add(unsafe.Pointer(b), dataOffset)
			k := add(base, i*uintptr(t.keysize))
			e := add(base, bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
			k2 := k
			if indirectk {
				k2 = *(*unsafe.Pointer)(k)
			}
			hash := t.hasher(k2, uintptr(h.hash0))
			top := tophash(hash)
			// Find the first non-null slot.
			var (
				dstb *bmap
				dsti uintptr
				dstp = newProbe(hash, mask)
			)
			for {
				dstb = (*bmap)(add(h.buckets, dstp.Bucket()*uintptr(t.bucketsize)))
				status := matchEmptyOrDeleted(dstb.tophash)
				dsti = status.NextMatch()
				if dsti < bucketCnt {
					break
				}
				dstp.Next()
			}

			// The target bucket is the same.
			if dstp.Bucket() == bucket {
				// Just mark slot as FULL.
				b.tophash[i] = top
				i += 1
				continue
			}

			dstbase := add(unsafe.Pointer(dstb), dataOffset) // key and value start
			dstk := add(unsafe.Pointer(dstbase), dsti*uintptr(t.keysize))
			dste := add(unsafe.Pointer(dstbase), bucketCnt*uintptr(t.keysize)+dsti*uintptr(t.elemsize))

			// Target is in another bucket.
			switch dstb.tophash[dsti] {
			case emptySlot:
				// 1. Transfer element to target
				// 2. Mark target as FULL
				// 3. Mark slot as EMPTY

				// Store new key and value at insert position.
				if indirectk {
					*(*unsafe.Pointer)(dstk) = k2
				} else {
					typedmemmove(t.key, dstk, k)
				}
				if indirecte {
					*(*unsafe.Pointer)(dste) = *(*unsafe.Pointer)(e)
				} else {
					typedmemmove(t.elem, dste, e)
				}

				// Clear key and value.
				if indirectk {
					*(*unsafe.Pointer)(k) = nil
				} else if t.key.ptrdata != 0 {
					memclrHasPointers(k, t.key.size)
				}
				if indirecte {
					*(*unsafe.Pointer)(e) = nil
				} else if t.elem.ptrdata != 0 {
					memclrHasPointers(e, t.elem.size)
				} else {
					memclrNoHeapPointers(e, t.elem.size)
				}
				dstb.tophash[dsti] = top
				b.tophash[i] = emptySlot
				i++
			case deletedSlot:
				// 1. Swap current element with target element
				// 2. Mark target as FULL
				// 3. Repeat procedure for current slot with moved from element (target)

				// This path is not commonly executed, just lazy initializing temporary variables.
				if !inittmp {
					if !indirectk {
						tmpk = newobject(t.key)
					}
					if !indirecte {
						tmpe = newobject(t.elem)
					}
					inittmp = true
				}

				if indirectk {
					*(*unsafe.Pointer)(dstk), *(*unsafe.Pointer)(k) = *(*unsafe.Pointer)(k), *(*unsafe.Pointer)(dstk)
				} else {
					typedmemmove(t.key, tmpk, dstk)
					typedmemmove(t.key, dstk, k)
					typedmemmove(t.key, k, tmpk)
				}
				if indirecte {
					*(*unsafe.Pointer)(dste), *(*unsafe.Pointer)(e) = *(*unsafe.Pointer)(e), *(*unsafe.Pointer)(dste)
				} else {
					typedmemmove(t.elem, tmpe, dste)
					typedmemmove(t.elem, dste, e)
					typedmemmove(t.elem, e, tmpe)
				}
				dstb.tophash[dsti] = top
			}
		}
	}
	h.growthLeft = int(bucketNum*bucketCnt) - h.count
}

func growbig(h *hmap, t *maptype) {
	oldB := h.B
	newB := h.B + 1
	oldBucketnum := bucketShift(oldB)
	newBucketnum := bucketShift(newB)
	newCap := newBucketnum * bucketCnt

	newBucketArray := makeBucketArray(t, newB)
	newMask := newBucketnum - 1
	hash0 := uintptr(h.hash0)

	for bucket := uintptr(0); bucket < oldBucketnum; bucket++ {
		b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
		base := add(unsafe.Pointer(b), dataOffset) // key and value start
		status := matchFull(b.tophash)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(base, i*uintptr(t.keysize))
			e := add(base, bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
			mapassignwithoutgrow(t, hash0, newMask, newBucketArray, k, e)
			status.RemoveNextMatch()
		}
	}

	h.B = newB
	h.flags &^= iterator
	h.buckets = newBucketArray
	h.growthLeft = int(newCap) - h.count
}

func mapassignwithoutgrow(t *maptype, hash0, mask uintptr, buckets, key, elem unsafe.Pointer) {
	var hash uintptr

	if t.indirectkey() {
		hash = t.hasher(*(*unsafe.Pointer)(key), hash0)
	} else {
		hash = t.hasher(key, hash0)
	}

	top := tophash(hash)
	p := newProbe(hash, mask)

	// The key is not in the map.
	for {
		b := (*bmap)(add(buckets, p.Bucket()*uintptr(t.bucketsize)))
		// Check empty slot or deleted slot.
		status := matchEmptyOrDeleted(b.tophash)
		i := status.NextMatch()
		if i < bucketCnt {
			// Insert key and value.
			b.tophash[i] = top
			base := add(unsafe.Pointer(b), dataOffset)
			k := add(unsafe.Pointer(base), i*uintptr(t.keysize))
			if t.indirectkey() {
				*(*unsafe.Pointer)(k) = *(*unsafe.Pointer)(key)
			} else {
				typedmemmove(t.key, k, key)
			}
			e := add(unsafe.Pointer(base), bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
			if t.indirectelem() {
				*(*unsafe.Pointer)(e) = *(*unsafe.Pointer)(elem)
			} else {
				typedmemmove(t.elem, e, elem)
			}
			return
		}
		p.Next()
	}
}

// Reflect stubs. Called from ../reflect/asm_*.s

//go:linkname reflect_makemap reflect.makemap
func reflect_makemap(t *maptype, cap int) *hmap {
	// Check invariants and reflects math.
	if t.key.equal == nil {
		throw("runtime.reflect_makemap: unsupported map key type")
	}
	if t.key.size > maxKeySize && (!t.indirectkey() || t.keysize != uint8(goarch.PtrSize)) ||
		t.key.size <= maxKeySize && (t.indirectkey() || t.keysize != uint8(t.key.size)) {
		throw("key size wrong")
	}
	if t.elem.size > maxElemSize && (!t.indirectelem() || t.elemsize != uint8(goarch.PtrSize)) ||
		t.elem.size <= maxElemSize && (t.indirectelem() || t.elemsize != uint8(t.elem.size)) {
		throw("elem size wrong")
	}
	if t.key.align > bucketCnt {
		throw("key align too big")
	}
	if t.elem.align > bucketCnt {
		throw("elem align too big")
	}
	if t.key.size%uintptr(t.key.align) != 0 {
		throw("key size not a multiple of key align")
	}
	if t.elem.size%uintptr(t.elem.align) != 0 {
		throw("elem size not a multiple of elem align")
	}
	if bucketCnt < 8 {
		throw("bucketsize too small for proper alignment")
	}
	if dataOffset%uintptr(t.key.align) != 0 {
		throw("need padding in bucket (key)")
	}
	if dataOffset%uintptr(t.elem.align) != 0 {
		throw("need padding in bucket (elem)")
	}

	return makemap(t, cap, nil)
}

//go:linkname reflect_mapaccess reflect.mapaccess
func reflect_mapaccess(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	elem, ok := mapaccess2(t, h, key)
	if !ok {
		// reflect wants nil for a missing element
		elem = nil
	}
	return elem
}

//go:linkname reflect_mapaccess_faststr reflect.mapaccess_faststr
func reflect_mapaccess_faststr(t *maptype, h *hmap, key string) unsafe.Pointer {
	elem, ok := mapaccess2_faststr(t, h, key)
	if !ok {
		// reflect wants nil for a missing element
		elem = nil
	}
	return elem
}

//go:linkname reflect_mapassign reflect.mapassign
func reflect_mapassign(t *maptype, h *hmap, key unsafe.Pointer, elem unsafe.Pointer) {
	p := mapassign(t, h, key)
	typedmemmove(t.elem, p, elem)
}

//go:linkname reflect_mapassign_faststr reflect.mapassign_faststr
func reflect_mapassign_faststr(t *maptype, h *hmap, key string, elem unsafe.Pointer) {
	p := mapassign_faststr(t, h, key)
	typedmemmove(t.elem, p, elem)
}

//go:linkname reflect_mapdelete reflect.mapdelete
func reflect_mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
	mapdelete(t, h, key)
}

//go:linkname reflect_mapdelete_faststr reflect.mapdelete_faststr
func reflect_mapdelete_faststr(t *maptype, h *hmap, key string) {
	mapdelete_faststr(t, h, key)
}

//go:linkname reflect_mapiterinit reflect.mapiterinit
func reflect_mapiterinit(t *maptype, h *hmap, it *hiter) {
	mapiterinit(t, h, it)
}

//go:linkname reflect_mapiternext reflect.mapiternext
func reflect_mapiternext(it *hiter) {
	mapiternext(it)
}

//go:linkname reflect_mapiterkey reflect.mapiterkey
func reflect_mapiterkey(it *hiter) unsafe.Pointer {
	return it.key
}

//go:linkname reflect_mapiterelem reflect.mapiterelem
func reflect_mapiterelem(it *hiter) unsafe.Pointer {
	return it.elem
}

//go:linkname reflect_maplen reflect.maplen
func reflect_maplen(h *hmap) int {
	if h == nil {
		return 0
	}
	if raceenabled {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(reflect_maplen))
	}
	return h.count
}

//go:linkname reflectlite_maplen internal/reflectlite.maplen
func reflectlite_maplen(h *hmap) int {
	if h == nil {
		return 0
	}
	if raceenabled {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(reflect_maplen))
	}
	return h.count
}

const maxZero = 1024 // must match value in reflect/value.go:maxZero cmd/compile/internal/gc/walk.go:zeroValSize
var zeroVal [maxZero]byte
