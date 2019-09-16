// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"testing"
)

func TestPageCacheAlloc(t *testing.T) {
	base := uintptr(0xc000000)
	tryAlloc := func(t *testing.T, c *cache, n uintptr, addr uintptr, cache uint64) {
		a := c.alloc(n)
		if a != addr {
			t.Fatalf("got addr %x, want %x", a, addr)
		}
		if c.cache != cache {
			t.Fatalf("got cache state %016x, want %016x", c.cache, cache)
		}
	}
	t.Run("Empty", func(t *testing.T) {
		c := cache{base: base, cache: 0}
		tryAlloc(t, &c, 1, 0, 0)
	})
	t.Run("OneLo", func(t *testing.T) {
		c := cache{base: base, cache: 1}
		tryAlloc(t, &c, 1, base, 0)
	})
	t.Run("OneHi", func(t *testing.T) {
		c := cache{base: base, cache: 1 << 63}
		tryAlloc(t, &c, 1, base+63*pageSize, 0)
	})
	t.Run("OneSwiss", func(t *testing.T) {
		c := cache{base: base, cache: 0xaaaaaaaaaaaaaaaa}
		tryAlloc(t, &c, 1, base+pageSize, 0xaaaaaaaaaaaaaaa8)
	})
	t.Run("TwoLo", func(t *testing.T) {
		c := cache{base: base, cache: 3}
		tryAlloc(t, &c, 2, base, 0)
	})
	t.Run("TwoHi", func(t *testing.T) {
		c := cache{base: base, cache: 3 << 62}
		tryAlloc(t, &c, 2, base+62*pageSize, 0)
	})
	t.Run("TwoSwiss", func(t *testing.T) {
		c := cache{base: base, cache: 0xaaaaaaaaaaaaaaaa}
		tryAlloc(t, &c, 2, 0, 0xaaaaaaaaaaaaaaaa)
	})
	t.Run("TwoMiddle", func(t *testing.T) {
		c := cache{base: base, cache: 0x3000300030003000}
		tryAlloc(t, &c, 2, base+12*pageSize, 0x3000300030000000)
	})
}
