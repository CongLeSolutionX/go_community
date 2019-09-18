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
	i := findConsecN64(c.cache, int(npages))
	if i >= 64 {
		return 0, 0
	}
	c.cache = clearConsecBits64(c.cache, i, int(npages))
	mask := ((uint64(1) << npages) - 1) << i
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
	a := s.arenas(arenaIndex(c.base))
	pi := arenaPageIndex(c.base)
	for i := 0; i < 64; i++ {
		if c.cache&(1<<i) != 0 {
			a.pageAlloc.free1(pi + i)
			if c.scav&(1<<i) != 0 {
				a.pageAlloc.scavengeRange(pi+i, 1)
			}
		}
	}
	// Since this is a lot like a free, we need to make sure
	// we update the searchAddr.
	if c.base+arenaBaseOffset < s.searchAddr+arenaBaseOffset {
		s.searchAddr = c.base
	}
	s.update(c.base, pageCacheSize, false, false)
	*c = pageCache{}
}

// allocToCache finds an aligned 64-bit chunk in b which has some free space
// and takes all the zero bitmap it can.
//
// Returns the index where the chunk was taken from and a 64-bit bitmap where
// each 1 represents a free page in that 64-bit chunk.
func (b *mallocBits) allocToCache(searchAddr int) (int, uint64) {
	for i := searchAddr / 64; i < len(b); i++ {
		if x := b[i]; x != ^uint64(0) {
			b[i] = ^uint64(0)
			return i * 64, ^x
		}
	}
	return -1, 0
}

// allocToCache wraps mallocBits.allocToCache and additionally manages
// the scavenged bits appropriately.
//
// Returns the index where the chunk was taken from and a 64-bit bitmap where
// each 1 represents a free page in that 64-bit chunk. Also returns the number
// of pages that were scavenged in the newly-allocated chunk.
func (m *mallocData) allocToCache(searchAddr int) (int, uint64, uint64) {
	base, cache := m.mallocBits.allocToCache(searchAddr)
	if base < 0 {
		// Failed to allocate, so skip all the other stuff.
		return -1, 0, 0
	}
	// Load the scavenge bits and mask them with the free memory
	// in the cache that we actually care about.
	if base%64 != 0 {
		print("runtime: base = ", base, "\n")
		throw("base must be 64-bit aligned")
	}
	scav := m.scavenged[base/64] & cache
	// Clear the scavenged bits when we alloc.
	m.scavenged.clearRange(base, pageCacheSize)
	return base, cache, scav
}

// Slow path for allocToCache.
func (s *pageAlloc) allocToCacheSlow() pageCache {
	// Search algorithm
	//
	// Iterate over each level, looking for the first non-zero summary.
	// Any non-zero summary means there's at least one free page. Once
	// the bottom is reached, we have the index of the first arena which
	// has some free pages.
	i := 0
	lastsum := mallocSum(0)
	lastidx := -1
nextLevel:
	for l := 0; l < len(s.summary); l++ {
		b := levelBits[l]
		e := 1 << b
		i <<= b
		level := s.summary[l][i : i+e]
		// Determine j0, the first index we should start iterating from.
		// The searchAddr may help us eliminate iterations if we followed the
		// searchAddr on the previous level, in which case the top bits of the
		// searchAddr address should be the same as i, after levelShift.
		j0 := 0
		if searchAddrIdx := int((s.searchAddr + arenaBaseOffset) >> levelShift[l]); searchAddrIdx&^(e-1) == i {
			j0 = searchAddrIdx & (e - 1)
		}
		for j := j0; j < len(level); j++ {
			sum := level[j]
			if sum != 0 {
				lastidx = i + j
				lastsum = sum
				i += j
				continue nextLevel
			}
		}
		if l != 0 {
			print("runtime: summary[", l-1, "][", lastidx, "] = ", lastsum.start(), ", ", lastsum.max(), ", ", lastsum.end(), "\n")
			print("runtime: level = ", l, ", j0 = ", j0, "\n")
			print("runtime: s.searchAddr = ", hex(s.searchAddr), ", i = ", i, ", levelShift[level] = ", levelShift[l], ", e = ", e, "\n")
			for j := 0; j < len(level); j++ {
				sum := level[j]
				print("runtime: summary[", l, "][", i+j, "] = (", sum.start(), ", ", sum.max(), ", ", sum.end(), ")\n")
				if l == len(s.summary)-1 {
					ci := chunkIdx(i + j)
					ai := arenaIdx(ci / mallocChunksPerArena)
					aci := arenaChunkIndex(chunkBase(ci))
					if a := s.arenas(ai); a != nil {
						for z, b := range a.pageAlloc.mallocBits[aci*mallocChunkPages/64 : (aci+1)*mallocChunkPages/64] {
							print("runtime: chunk[", z, "] = ", hex(uintptr(b)), "\n")
						}
					}
				}
			}
			throw("bad summary data")
		}
		return pageCache{}
	}
	ai := arenaIdx(i / mallocChunksPerArena)
	j, c, scav := s.arenas(ai).pageAlloc.allocToCache((i % mallocChunksPerArena) * mallocChunkPages)
	if j < 0 {
		throw("bad summary data")
	}
	addr := arenaBase(ai) + uintptr(j)*pageSize
	return pageCache{addr, c, scav}
}

// allocToCache acquires a pageCacheSize-aligned chunk of free pages which
// may not be contiguous, and returns a pageCache structure which owns the
// chunk.
func (s *pageAlloc) allocToCache() pageCache {
	// If the searchAddr refers to a region which has a higher address than
	// any known arena, then we know we're out of memory.
	if chunkIndex(s.searchAddr) >= s.end {
		return pageCache{}
	}
	c := pageCache{}
	ci := chunkIndex(s.searchAddr) // chunk index
	if s.summary[len(s.summary)-1][ci] != 0 {
		// Fast path: there's free pages at or near the searchAddr address.
		ai := arenaIndex(s.searchAddr)
		j, cv, scav := s.arenas(ai).pageAlloc.allocToCache(arenaPageIndex(s.searchAddr))
		if j < 0 {
			throw("bad summary data")
		}
		addr := arenaBase(ai) + uintptr(j)*pageSize
		c = pageCache{addr, cv, scav}
	} else {
		// Slow path: the searchAddr address had nothing there, so go find
		// the first free page the slow way.
		c = s.allocToCacheSlow()
		if c.base == 0 {
			// We failed to find adequate free space, so mark the searchAddr as OoM
			// and return an empty pageCache.
			s.searchAddr = maxSearchAddr
			return pageCache{}
		}
	}
	s.update(c.base, pageCacheSize, false, true)
	s.searchAddr = c.base + pageSize*pageCacheSize
	return c
}
