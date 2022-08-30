// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

func mapaccess1_fast64(t *maptype, h *hmap, key uint64) unsafe.Pointer {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapaccess1_fast64))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}

	B := h.B
	if B == 0 {
		// We don't need hash if only one bucket.
		b := (*bmap)(h.buckets)
		for i, k := uintptr(0), add(unsafe.Pointer(b), dataOffset); i < bucketCnt; i, k = i+1, add(k, 8) {
			if *(*uint64)(k) == key && isFull(b.tophash[i]) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize))
			}
		}
		return unsafe.Pointer(&zeroVal[0])
	}

	hash := t.hasher(noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
	top := tophash(hash)
	mask := bucketMask(B)
	p := newProbe(hash, mask)

	if B <= loopaccessB {
		goto loopaccess
	}

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*8)
			if *(*uint64)(k) == key {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize))
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0])
		}
		p.Next()
	}
loopaccess:
	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		for i, k := uintptr(0), add(unsafe.Pointer(b), dataOffset); i < bucketCnt; i, k = i+1, add(k, 8) {
			if *(*uint64)(k) == key && isFull(b.tophash[i]) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize))
			}
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0])
		}
		p.Next()
	}
}

func mapaccess2_fast64(t *maptype, h *hmap, key uint64) (unsafe.Pointer, bool) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapaccess2_fast64))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map read and map write")
	}

	B := h.B
	if B == 0 {
		// We don't need hash if only one bucket.
		b := (*bmap)(h.buckets)
		for i, k := uintptr(0), add(unsafe.Pointer(b), dataOffset); i < bucketCnt; i, k = i+1, add(k, 8) {
			if *(*uint64)(k) == key && isFull(b.tophash[i]) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize)), true
			}
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}

	hash := t.hasher(noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
	top := tophash(hash)
	mask := bucketMask(B)
	p := newProbe(hash, mask)

	if B <= loopaccessB {
		goto loopaccess
	}

	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*8)
			if *(*uint64)(k) == key {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize)), true
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0]), false
		}
		p.Next()
	}
loopaccess:
	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		for i, k := uintptr(0), add(unsafe.Pointer(b), dataOffset); i < bucketCnt; i, k = i+1, add(k, 8) {
			if *(*uint64)(k) == key && isFull(b.tophash[i]) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize)), true
			}
		}
		if matchEmpty(b.tophash) != 0 {
			return unsafe.Pointer(&zeroVal[0]), false
		}
		p.Next()
	}
}

func mapassign_fast64(t *maptype, h *hmap, key uint64) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapassign_fast64))
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}
	hash := t.hasher(noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling t.hasher for consistency with mapassign.
	h.flags ^= hashWriting

	if h.buckets == nil {
		// Init an empty map.
		h.buckets = makeBucketArray(t, 0)
		h.growthLeft = bucketCnt
	}

	if h.needGrow() {
		grow_fast64(h, t)
	}

	p := newProbe(hash, bucketMask(h.B))

	var insertb *bmap
	var inserti uintptr
	var insertk unsafe.Pointer

	var (
		b      *bmap
		status bitmask64
		top    = tophash(hash)
	)
	// Check if the key in the map.
	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*8)
			if *(*uint64)(k) == key {
				insertb = b
				inserti = i
				goto done
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			goto insertnewkey
		}
		p.Next()
	}

insertnewkey:
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
			insertk = add(unsafe.Pointer(insertb), dataOffset+inserti*8)
			// Store new key at insert position.
			*(*uint64)(insertk) = key
			h.growthLeft -= 1
			h.count += 1
			goto done
		}
		p.Next()
	}
done:
	elem := add(unsafe.Pointer(insertb), dataOffset+bucketCnt*8+inserti*uintptr(t.elemsize))
	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
	return elem
}

func mapassign_fast64ptr(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapassign_fast64))
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}
	hash := t.hasher(noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling t.hasher for consistency with mapassign.
	h.flags ^= hashWriting

	if h.buckets == nil {
		// Init an empty map.
		h.buckets = makeBucketArray(t, 0)
		h.growthLeft = bucketCnt
	}

	if h.needGrow() {
		grow_fast64ptr(h, t)
	}

	p := newProbe(hash, bucketMask(h.B))

	var insertb *bmap
	var inserti uintptr
	var insertk unsafe.Pointer

	var (
		b      *bmap
		status bitmask64
		top    = tophash(hash)
	)
	// Check if the key in the map.
	for {
		b := (*bmap)(add(h.buckets, p.Bucket()*uintptr(t.bucketsize)))
		status := matchTopHash(b.tophash, top)
		for {
			i := status.NextMatch()
			if i >= bucketCnt {
				break
			}
			k := add(unsafe.Pointer(b), dataOffset+i*8)
			if *(*unsafe.Pointer)(k) == key {
				insertb = b
				inserti = i
				goto done
			}
			status.RemoveNextMatch()
		}
		if matchEmpty(b.tophash) != 0 {
			goto insertnewkey
		}
		p.Next()
	}
insertnewkey:
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
			insertk = add(unsafe.Pointer(insertb), dataOffset+inserti*8)
			// Store new key at insert position.
			*(*unsafe.Pointer)(insertk) = key
			h.growthLeft -= 1
			h.count += 1
			goto done
		}
		p.Next()
	}
done:
	elem := add(unsafe.Pointer(insertb), dataOffset+bucketCnt*8+inserti*uintptr(t.elemsize))
	if h.flags&hashWriting == 0 {
		fatal("concurrent map writes")
	}
	h.flags &^= hashWriting
	return elem
}

func mapdelete_fast64(t *maptype, h *hmap, key uint64) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapdelete_fast64))
	}
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		fatal("concurrent map writes")
	}

	hash := t.hasher(noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

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
			k := add(unsafe.Pointer(b), dataOffset+i*8)
			if key == *(*uint64)(k) {
				// Found this key.
				h.count -= 1
				// Only clear key if there are pointers in it.
				if t.key.ptrdata != 0 {
					if goarch.PtrSize == 8 {
						*(*unsafe.Pointer)(k) = nil
					} else {
						// There are three ways to squeeze at one ore more 32 bit pointers into 64 bits.
						// Just call memclrHasPointers instead of trying to handle all cases here.
						memclrHasPointers(k, 8)
					}
				}
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.elemsize))
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

func grow_fast64(h *hmap, t *maptype) {
	cap := bucketShift(h.B) * bucketCnt
	if uintptr(h.count*32) <= cap*25 && (h.flags&iterator != iterator) {
		// Rehash in place if the current size is <= 25/32 of capacity.
		// If there may be an iterator using buckets, we disable growsamesize.
		// Because it may move data to different buckets, this behavior will break the iterator.
		growsamesize_fast64(h, t)
	} else {
		growbig_fast64(h, t)
	}
}

func growsamesize_fast64(h *hmap, t *maptype) {
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
			k := add(base, i*8)
			e := add(base, bucketCnt*8+i*uintptr(t.elemsize))
			hash := t.hasher(noescape(unsafe.Pointer((*uint64)(k))), uintptr(h.hash0))
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
			dstk := add(unsafe.Pointer(dstbase), dsti*8)
			dste := add(unsafe.Pointer(dstbase), bucketCnt*8+dsti*uintptr(t.elemsize))

			// Target is in another bucket.
			switch dstb.tophash[dsti] {
			case emptySlot:
				// 1. Transfer element to target
				// 2. Mark target as FULL
				// 3. Mark slot as EMPTY

				// Store new key and value at insert position.
				*(*uint64)(dstk) = *(*uint64)(k)
				typedmemmove(t.elem, dste, e)

				// Clear key and value.
				if t.key.ptrdata != 0 {
					if goarch.PtrSize == 8 {
						*(*unsafe.Pointer)(k) = nil
					} else {
						// There are three ways to squeeze at one ore more 32 bit pointers into 64 bits.
						// Just call memclrHasPointers instead of trying to handle all cases here.
						memclrHasPointers(k, 8)
					}
				}
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
				*(*uint64)(tmpk) = *(*uint64)(dstk)
				typedmemmove(t.elem, tmpe, dste)

				// dstk,dste = k,e
				*(*uint64)(dstk) = *(*uint64)(k)
				typedmemmove(t.elem, dste, e)

				// k,e = tmpk,tmpe
				*(*uint64)(k) = *(*uint64)(tmpk)
				typedmemmove(t.elem, e, tmpe)

				dstb.tophash[dsti] = top
			}
		}
	}
	h.growthLeft = int(bucketNum*bucketCnt) - h.count
}

func growbig_fast64(h *hmap, t *maptype) {
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
			k := *(*uint64)(add(base, i*8))
			e := add(base, bucketCnt*8+i*uintptr(t.elemsize))
			mapassignwithoutgrow_fast64(t, hash0, newMask, newBucketArray, k, e)
			status.RemoveNextMatch()
		}
	}

	h.B = newB
	h.flags &^= iterator
	h.buckets = newBucketArray
	h.growthLeft = int(newCap) - h.count
}

func mapassignwithoutgrow_fast64(t *maptype, hash0, mask uintptr, buckets unsafe.Pointer, key uint64, elem unsafe.Pointer) {
	hash := t.hasher(noescape(unsafe.Pointer(&key)), hash0)
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
			k := add(unsafe.Pointer(base), i*8)
			*(*uint64)(k) = key
			e := add(unsafe.Pointer(base), bucketCnt*8+i*uintptr(t.elemsize))
			typedmemmove(t.elem, e, elem)
			return
		}
		p.Next()
	}
}

func grow_fast64ptr(h *hmap, t *maptype) {
	cap := bucketShift(h.B) * bucketCnt
	if uintptr(h.count*32) <= cap*25 && (h.flags&iterator != iterator) {
		// Rehash in place if the current size is <= 25/32 of capacity.
		// If there may be an iterator using buckets, we disable growsamesize.
		// Because it may move data to different buckets, this behavior will break the iterator.
		growsamesize_fast64ptr(h, t)
	} else {
		growbig_fast64ptr(h, t)
	}
}

func growsamesize_fast64ptr(h *hmap, t *maptype) {
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
			k := add(base, i*8)
			e := add(base, bucketCnt*8+i*uintptr(t.elemsize))
			hash := t.hasher(noescape(unsafe.Pointer((*unsafe.Pointer)(k))), uintptr(h.hash0))
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
			dstk := add(unsafe.Pointer(dstbase), dsti*8)
			dste := add(unsafe.Pointer(dstbase), bucketCnt*8+dsti*uintptr(t.elemsize))

			// Target is in another bucket.
			switch dstb.tophash[dsti] {
			case emptySlot:
				// 1. Transfer element to target
				// 2. Mark target as FULL
				// 3. Mark slot as EMPTY

				// Store new key and value at insert position.
				*(*unsafe.Pointer)(dstk) = *(*unsafe.Pointer)(k)
				typedmemmove(t.elem, dste, e)

				// Clear key and value.
				if t.key.ptrdata != 0 {
					if goarch.PtrSize == 8 {
						*(*unsafe.Pointer)(k) = nil
					} else {
						// There are three ways to squeeze at one ore more 32 bit pointers into 64 bits.
						// Just call memclrHasPointers instead of trying to handle all cases here.
						memclrHasPointers(k, 8)
					}
				}
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
				*(*unsafe.Pointer)(tmpk) = *(*unsafe.Pointer)(dstk)
				typedmemmove(t.elem, tmpe, dste)

				// dstk,dste = k,e
				*(*unsafe.Pointer)(dstk) = *(*unsafe.Pointer)(k)
				typedmemmove(t.elem, dste, e)

				// k,e = tmpk,tmpe
				*(*unsafe.Pointer)(k) = *(*unsafe.Pointer)(tmpk)
				typedmemmove(t.elem, e, tmpe)

				dstb.tophash[dsti] = top
			}
		}
	}
	h.growthLeft = int(bucketNum*bucketCnt) - h.count
}

func growbig_fast64ptr(h *hmap, t *maptype) {
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
			k := *(*unsafe.Pointer)(add(base, i*8))
			e := add(base, bucketCnt*8+i*uintptr(t.elemsize))
			mapassignwithoutgrow_fast64ptr(t, hash0, newMask, newBucketArray, k, e)
			status.RemoveNextMatch()
		}
	}

	h.B = newB
	h.flags &^= iterator
	h.buckets = newBucketArray
	h.growthLeft = int(newCap) - h.count
}

func mapassignwithoutgrow_fast64ptr(t *maptype, hash0, mask uintptr, buckets unsafe.Pointer, key unsafe.Pointer, elem unsafe.Pointer) {
	hash := t.hasher(noescape(unsafe.Pointer(&key)), hash0)
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
			k := add(unsafe.Pointer(base), i*8)
			*(*unsafe.Pointer)(k) = key
			e := add(unsafe.Pointer(base), bucketCnt*8+i*uintptr(t.elemsize))
			typedmemmove(t.elem, e, elem)
			return
		}
		p.Next()
	}
}
