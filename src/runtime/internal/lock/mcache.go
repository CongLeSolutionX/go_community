// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Per-P malloc cache for small objects.
//
// See malloc.h for an overview.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// dummy MSpan that contains no free objects.
var Emptymspan _core.Mspan

func Allocmcache() *_core.Mcache {
	Lock(&Mheap_.Lock)
	c := (*_core.Mcache)(FixAlloc_Alloc(&Mheap_.Cachealloc))
	Unlock(&Mheap_.Lock)
	_core.Memclr(unsafe.Pointer(c), unsafe.Sizeof(*c))
	for i := 0; i < _core.NumSizeClasses; i++ {
		c.Alloc[i] = &Emptymspan
	}

	// Set first allocation sample size.
	rate := MemProfileRate
	if rate > 0x3fffffff { // make 2*rate not overflow
		rate = 0x3fffffff
	}
	if rate != 0 {
		c.Next_sample = int32(int(Fastrand1()) % (2 * rate))
	}

	return c
}
