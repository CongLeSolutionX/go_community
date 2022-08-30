// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

func mapaccess1_faststr(t *maptype, h *hmap, ky string) unsafe.Pointer {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapaccess1_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}
	key := stringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.len < 32 {
			// short key, doing lots of comparisons is ok
			for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*goarch.PtrSize) {
				k := (*stringStruct)(kptr)
				if k.len != key.len || !isFull(b.tophash[i]) {
					continue
				}
				if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
					return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
				}
			}
			return unsafe.Pointer(&zeroVal[0])
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*goarch.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || !isFull(b.tophash[i]) {
				continue
			}
			if k.str == key.str {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
			}
			// check first 4 bytes
			if *((*[4]byte)(key.str)) != *((*[4]byte)(k.str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(add(key.str, uintptr(key.len)-4))) != *((*[4]byte)(add(k.str, uintptr(key.len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+keymaybe*2*goarch.PtrSize))
			if memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+keymaybe*uintptr(t.elemsize))
			}
		}
		return unsafe.Pointer(&zeroVal[0])
	}
dohash:
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
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
			kptr := add(unsafe.Pointer(b), dataOffset+i*2*goarch.PtrSize)
			k := (*stringStruct)(kptr)
			if k.len == key.len && (k.str == key.str || memequal(k.str, key.str, uintptr(key.len))) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0])
		}
		p.Next()
	}
}

func mapaccess2_faststr(t *maptype, h *hmap, ky string) (unsafe.Pointer, bool) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapaccess2_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}
	key := stringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.len < 32 {
			// short key, doing lots of comparisons is ok
			for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*goarch.PtrSize) {
				k := (*stringStruct)(kptr)
				if k.len != key.len || !isFull(b.tophash[i]) {
					continue
				}
				if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
					return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize)), true
				}
			}
			return unsafe.Pointer(&zeroVal[0]), false
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*goarch.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || !isFull(b.tophash[i]) {
				continue
			}
			if k.str == key.str {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize)), true
			}
			// check first 4 bytes
			if *((*[4]byte)(key.str)) != *((*[4]byte)(k.str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(add(key.str, uintptr(key.len)-4))) != *((*[4]byte)(add(k.str, uintptr(key.len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+keymaybe*2*goarch.PtrSize))
			if memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+keymaybe*uintptr(t.elemsize)), true
			}
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}
dohash:
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
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
			kptr := add(unsafe.Pointer(b), dataOffset+i*2*goarch.PtrSize)
			k := (*stringStruct)(kptr)
			if k.len == key.len && (k.str == key.str || memequal(k.str, key.str, uintptr(key.len))) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize)), true
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0]), false
		}
		p.Next()
	}
}

func mapassign_faststr(t *maptype, h *hmap, ky string) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapassign_faststr))
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}
	key := stringStructOf(&ky)
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))

	// Set hashWriting after calling t.hasher for consistency with mapassign.
	h.flags ^= hashWriting

	if h.buckets == nil {
		// Init an empty map.
		h.buckets = makeBucketArray(t, 0)
		h.growthLeft = bucketCnt
	}

	top := tophash(hash)

	if h.needGrow() {
		grow_faststr(h, t)
	}

	p := newProbe(hash, bucketMask(h.B))

	var insertb *bmap
	var inserti uintptr
	var insertk unsafe.Pointer

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
			kptr := add(unsafe.Pointer(b), dataOffset+i*2*goarch.PtrSize)
			k := (*stringStruct)(kptr)
			if k.len == key.len && (k.str == key.str || memequal(k.str, key.str, uintptr(key.len))) {
				insertb = b
				inserti = i
				// Overwrite existing key, so it can be garbage collected.
				// The size is already guaranteed to be set correctly.
				k.str = key.str
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
			// Insert key and value.
			insertb = b
			inserti = i
			insertb.tophash[i] = top
			insertk = add(unsafe.Pointer(insertb), dataOffset+inserti*2*goarch.PtrSize)
			// Store new key at insert position.
			*((*stringStruct)(insertk)) = *key

			h.growthLeft -= 1
			h.count += 1
			goto done
		}
		p.Next()
	}
done:
	elem := add(unsafe.Pointer(insertb), dataOffset+bucketCnt*2*goarch.PtrSize+inserti*uintptr(t.elemsize))
	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
	return elem
}

func mapdelete_faststr(t *maptype, h *hmap, ky string) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapdelete_faststr))
	}
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}

	key := stringStructOf(&ky)
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))

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
			kptr := add(unsafe.Pointer(b), dataOffset+i*2*goarch.PtrSize)
			k := (*stringStruct)(kptr)
			if k.len == key.len && (k.str == key.str || memequal(k.str, key.str, uintptr(key.len))) {
				// Found this key.
				h.count -= 1
				k.str = nil
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
				if t.elem.ptrdata != 0 {
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

func grow_faststr(h *hmap, t *maptype) {
	cap := bucketShift(h.B) * bucketCnt
	if uintptr(h.count*32) <= cap*25 && (h.flags&iterator != iterator) {
		// Rehash in place if the current size is <= 25/32 of capacity.
		// If there may be an iterator using buckets, we disable growsamesize.
		// Because it may move data to different buckets, this behavior will break the iterator.
		growsamesize_faststr(h, t)
	} else {
		growbig_faststr(h, t)
	}
}

func growbig_faststr(h *hmap, t *maptype) {
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
			kptr := add(base, i*2*goarch.PtrSize)
			k := (*stringStruct)(kptr)
			e := add(base, bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
			mapassignwithoutgrow_faststr(t, hash0, newMask, newBucketArray, k, e)
			status.RemoveNextMatch()
		}
	}

	h.B = newB
	h.flags &^= iterator
	h.buckets = newBucketArray
	h.growthLeft = int(newCap) - h.count
}

func growsamesize_faststr(h *hmap, t *maptype) {
	bucketNum := bucketShift(h.B)
	mask := bucketNum - 1
	// For all buckets:
	// - mark all DELETED slots as EMPTY
	// - mark all FULL slots as DELETED
	for bucket := uintptr(0); bucket < bucketNum; bucket++ {
		b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
		prepareSameSizeGrow(&b.tophash)
	}
	// Temporary key and value used to swap.
	tmpk := newobject(t.key)
	tmpe := newobject(t.elem)

	for bucket := uintptr(0); bucket < bucketNum; bucket++ {
		b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
		for i := uintptr(0); i < bucketCnt; {
			if b.tophash[i] != deletedSlot {
				i++
				continue
			}
			base := add(unsafe.Pointer(b), dataOffset)
			k := add(base, i*2*goarch.PtrSize)
			e := add(base, bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
			hash := t.hasher(noescape(unsafe.Pointer(k)), uintptr(h.hash0))
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
			dstk := add(unsafe.Pointer(dstbase), dsti*2*goarch.PtrSize)
			dste := add(unsafe.Pointer(dstbase), bucketCnt*2*goarch.PtrSize+dsti*uintptr(t.elemsize))

			// Target is in another bucket.
			switch dstb.tophash[dsti] {
			case emptySlot:
				// 1. Transfer element to target
				// 2. Mark target as FULL
				// 3. Mark slot as EMPTY

				// Store new key and value at insert position.
				*(*stringStruct)(dstk) = *(*stringStruct)(k)
				typedmemmove(t.elem, dste, e)

				// Clear key and value.
				(*stringStruct)(k).str = nil

				if t.elem.ptrdata != 0 {
					memclrHasPointers(e, uintptr(t.elemsize))
				} else {
					memclrNoHeapPointers(e, uintptr(t.elemsize))
				}

				dstb.tophash[dsti] = top
				b.tophash[i] = emptySlot
				i++
			case deletedSlot:
				// 1. Swap current element with target element
				// 2. Mark target as FULL
				// 3. Repeat procedure for current slot with moved from element (target)

				// tmpk,tmpe = dstk,dste
				*(*stringStruct)(tmpk) = *(*stringStruct)(dstk)
				typedmemmove(t.elem, tmpe, dste)

				// dstk,dste = k,e
				*(*stringStruct)(dstk) = *(*stringStruct)(k)
				typedmemmove(t.elem, dste, e)

				// k,e = tmpk,tmpe
				*(*stringStruct)(k) = *(*stringStruct)(tmpk)
				typedmemmove(t.elem, e, tmpe)

				dstb.tophash[dsti] = top
			}
		}
	}
	h.growthLeft = int(bucketNum*bucketCnt) - h.count
}

func mapassignwithoutgrow_faststr(t *maptype, hash0, mask uintptr, buckets unsafe.Pointer, key *stringStruct, elem unsafe.Pointer) {
	hash := t.hasher(noescape(unsafe.Pointer(key)), hash0)
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
			k := add(unsafe.Pointer(base), i*2*goarch.PtrSize)
			*((*stringStruct)(k)) = *key
			e := add(unsafe.Pointer(base), bucketCnt*2*goarch.PtrSize+i*uintptr(t.elemsize))
			typedmemmove(t.elem, e, elem)
			return
		}
		p.Next()
	}
}
