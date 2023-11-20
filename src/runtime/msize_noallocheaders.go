// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.allocheaders

// Malloc small size classes.
//
// See malloc.go for overview.
// See also mksizeclasses.go for how we decide what size classes to use.

package runtime

import "unsafe"

// Returns size of the memory block that mallocgc will allocate if you ask for the size.
//
// The noscan argument is purely for compatibility with goexperiment.AllocHeaders.
func roundupsize(size uintptr, noscan bool) uintptr {
	if size < _MaxSmallSize {
		if size <= smallSizeMax-8 {
			return uintptr(class_to_size[size_to_class8[divRoundUp(size, smallSizeDiv)]])
		} else {
			return uintptr(class_to_size[size_to_class128[divRoundUp(size-smallSizeMax, largeSizeDiv)]])
		}
	}
	if size+_PageSize < size {
		return size
	}
	return alignUp(size, _PageSize)
}

var deferSize = roundupsize(unsafe.Sizeof(_defer{}), false)
var gSize = roundupsize(unsafe.Sizeof(g{}), false)
var sudogSize = roundupsize(unsafe.Sizeof(sudog{}), false)

func init() {
	// TODO: TEMPORARY
	println("msize_noallocheaders:")
	println("  deferSize =", deferSize)
	println("  gSize =", gSize)
	println("  sudogSize =", sudogSize)
}
