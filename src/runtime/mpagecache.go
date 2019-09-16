// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"math/bits"
)

const pageCacheSize = 64

type pageCache struct {
	base  uintptr
	cache uint64
	scav  uint64
}

func (c *pageCache) alloc(npages uintptr) uintptr {
	if c.cache == 0 {
		return 0
	}
	if npages == 1 {
		i := uintptr(bits.TrailingZeros64(c.cache))
		c.cache &^= 1 << i // clear bit
		return c.base + i*pageSize
	}
	return c.allocN(npages)
}

func (c *pageCache) allocN(npages uintptr) uintptr {
	i := findConsecN64(c.cache, int(npages))
	if i >= 64 {
		return 0
	}
	c.cache = clearConsecBits64(c.cache, i, int(npages))
	return c.base + uintptr(i*pageSize)
}
