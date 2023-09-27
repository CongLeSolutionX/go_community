// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.allocheaders

// Malloc small size classes.
//
// See malloc.go for overview.
// See also mksizeclasses.go for how we decide what size classes to use.

package runtime

import "unsafe"

// Returns size of the memory block that mallocgc will allocate if you ask for the size,
// minus any inline space for metadata.
func roundupsize(size uintptr, noscan bool) (reqSize uintptr) {
	reqSize = size
	if reqSize <= maxSmallSize-mallocHeaderSize {
		// Small object.
		if !noscan && reqSize > minSizeForMallocHeader { // !noscan && !heapBitsInSpan(reqSize)
			reqSize += mallocHeaderSize
		}
		// (reqSize - size) is either mallocHeaderSize or 0. We need to subtract mallocHeaderSize
		// from the result if we have one, since mallocgc will add it back in.
		if reqSize <= smallSizeMax-8 {
			return uintptr(class_to_size[size_to_class8[divRoundUp(reqSize, smallSizeDiv)]]) - (reqSize - size)
		}
		return uintptr(class_to_size[size_to_class128[divRoundUp(reqSize-smallSizeMax, largeSizeDiv)]]) - (reqSize - size)
	}
	// Large object. Align reqSize up to the next page. Check for overflow.
	reqSize += pageSize - 1
	if reqSize < size {
		return size
	}
	return reqSize &^ (pageSize - 1)
}

// Size of heap memory blocks that are allowed to point to the stack,
// including any inline space for metadata.
var (
	deferSize uintptr
	gSize     uintptr
	sudogSize uintptr
)

func init() {
	// Compute the size class in bytes for _defer, g, and sudog.
	// We currently assume they are all small objects.
	sizeClassSize := func(reqSize uintptr) uintptr {
		if reqSize > maxSmallSize-mallocHeaderSize {
			throw("unexpectedly large")
		}
		if reqSize < maxTinySize {
			throw("unexpectedly small")
		}
		size := roundupsize(reqSize, false)
		if reqSize > minSizeForMallocHeader {
			// Add back in the mallocHeaderSize that roundupsize subtracted.
			size += mallocHeaderSize
		}
		return size
	}
	deferSize = sizeClassSize(unsafe.Sizeof(_defer{}))
	gSize = sizeClassSize(unsafe.Sizeof(g{}))
	sudogSize = sizeClassSize(unsafe.Sizeof(sudog{}))
}
