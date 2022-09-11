// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.allocheaders

// Malloc small size classes.
//
// See malloc.go for overview.
// See also mksizeclasses.go for how we decide what size classes to use.

package runtime

// Returns size of the memory block that mallocgc will allocate if you ask for the size,
// minus any inline space for metadata.
func roundupsize(size uintptr, noscan bool) uintptr {
	if size <= maxSmallSize-mallocHeaderSize {
		hasHeader := !noscan && !heapBitsInSpan(size)
		if hasHeader {
			size += mallocHeaderSize
		}
		if size <= smallSizeMax-8 {
			size = uintptr(class_to_size[size_to_class8[divRoundUp(size, smallSizeDiv)]])
		} else {
			size = uintptr(class_to_size[size_to_class128[divRoundUp(size-smallSizeMax, largeSizeDiv)]])
		}
		if hasHeader {
			size -= mallocHeaderSize
		}
		return size
	}
	if size+pageSize < size {
		return size
	}
	return alignUp(size, pageSize)
}
