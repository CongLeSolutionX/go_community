// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"math/bits"
)

const pageCacheSize = 64

// pageCache represents a per-p cache of pages the allocator can
// allocate from without a lock. More specifically, it represents
// a pageCacheSize*pageSize chunk of memory with 0 or more free
// pages in it.
type pageCache struct {
	base  uintptr // base address of the chunk
	cache uint64  // 64-bit bitmap representing free pages (1 means free)
	scav  uint64  // 64-bit bitmap representing scavenged pages (1 means scavenged)
}

// alloc allocates npages from the page cache and is the main entry
// point for allocation.
//
// Returns a base address and the amount of scavenged memory in the
// allocated region in bytes.
func (c *pageCache) alloc(npages uintptr) (uintptr, uintptr) {
	if c.cache == 0 {
		return 0, 0
	}
	if npages == 1 {
		i := uintptr(bits.TrailingZeros64(c.cache))
		c.cache &^= 1 << i // clear bit
		return c.base + i*pageSize, uintptr((c.scav>>i)&1) * pageSize
	}
	return c.allocN(npages)
}

// allocN is a helper which attempts to allocate npages worth of pages
// from the cache. It represents the general case for allocating from
// the page cache.
//
// Returns a base address and the amount of scavenged memory in the
// allocated region in bytes.
func (c *pageCache) allocN(npages uintptr) (uintptr, uintptr) {
	i := findBitRange64(c.cache, uint(npages))
	if i >= 64 {
		return 0, 0
	}
	mask := ((uint64(1) << npages) - 1) << i
	c.cache &^= mask
	scav := bits.OnesCount64(c.scav & mask)
	return c.base + uintptr(i*pageSize), uintptr(scav) * pageSize
}

// flush empties out unallocated free pages in the given cache
// into s. Then, it clears the cache.
func (c *pageCache) flush(s *pageAlloc) {
	if c.cache == 0 {
		// Empty cache, nothing to do.
		return
	}
	ci := chunkIndex(c.base)
	pi := chunkPageIndex(c.base)
	for i := uint(0); i < 64; i++ {
		if c.cache&(1<<i) != 0 {
			s.chunks[ci].free1(pi + i)
			if c.scav&(1<<i) != 0 {
				s.chunks[ci].scavenged.setRange(pi+i, 1)
			}
		}
	}
	// Since this is a lot like a free, we need to make sure
	// we update the searchAddr just like free does.
	if s.compareSearchAddrTo(c.base) < 0 {
		s.searchAddr = c.base
	}
	s.update(c.base, pageCacheSize, false, false)
	*c = pageCache{}
}

// allocToCache acquires a pageCacheSize-aligned chunk of free pages which
// may not be contiguous, and returns a pageCache structure which owns the
// chunk.
//
// s.mheapLock must be held.
func (s *pageAlloc) allocToCache() pageCache {
	// If the searchAddr refers to a region which has a higher address than
	// any known chunk, then we know we're out of memory.
	if chunkIndex(s.searchAddr) >= s.end {
		return pageCache{}
	}
	c := pageCache{}
	ci := chunkIndex(s.searchAddr) // chunk index
	if s.summary[len(s.summary)-1][ci] != 0 {
		// Fast path: there's free pages at or near the searchAddr address.
		j, _ := s.chunks[ci].find(1, chunkPageIndex(s.searchAddr))
		if j < 0 {
			throw("bad summary data")
		}
		c = pageCache{
			base:  chunkBase(ci) + alignDown(uintptr(j), 64)*pageSize,
			cache: ^s.chunks[ci].block64(j),
			scav:  s.chunks[ci].scavenged.block64(j),
		}
	} else {
		// Slow path: the searchAddr address had nothing there, so go find
		// the first free page the slow way.
		addr, _ := s.find(1)
		if addr == 0 {
			// We failed to find adequate free space, so mark the searchAddr as OoM
			// and return an empty pageCache.
			s.searchAddr = maxSearchAddr
			return pageCache{}
		}
		ci := chunkIndex(addr)
		c = pageCache{
			base:  alignDown(addr, 64*pageSize),
			cache: ^s.chunks[ci].block64(chunkPageIndex(addr)),
			scav:  s.chunks[ci].scavenged.block64(chunkPageIndex(addr)),
		}
	}

	// Set the bits as allocated and clear the scavenged bits.
	s.allocRange(c.base, pageCacheSize)

	// Update as an allocation, but note that it's not contiguous.
	s.update(c.base, pageCacheSize, false, true)

	// We're always searching for the first free page, and we always know the
	// up to pageCache size bits will be allocated, so we can always move the
	// searchAddr past the cache.
	s.searchAddr = c.base + pageSize*pageCacheSize
	return c
}
