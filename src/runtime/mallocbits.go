// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"math/bits"
)

// pageBits is a bitmap representing one bit per page in an arena.
type pageBits [pagesPerArena / 64]uint64

// set1 clears a single bit in the pageBits at i.
func (b *pageBits) set1(i int) {
	b[i/64] |= 1 << (i % 64)
}

// setRange sets bits in the range [i, i+n).
func (b *pageBits) setRange(i, n int) {
	_ = b[i/64]
	if n == 1 {
		// Add a fast path for the n == 1 case.
		b.set1(i)
		return
	}
	j := i + n - 1
	if i/64 == j/64 {
		b[i/64] = setConsecBits64(b[i/64], i%64, n)
		return
	}
	_ = b[j/64]
	b[i/64] = setConsecBits64(b[i/64], i%64, 64-i%64)
	for k := i/64 + 1; k < j/64; k++ {
		b[k] = ^uint64(0)
	}
	b[j/64] = setConsecBits64(b[j/64], 0, j%64+1)
}

// setAll sets all the bits of b.
func (b *pageBits) setAll() {
	for i := 0; i < len(b); i++ {
		b[i] = ^uint64(0)
	}
}

// clear1 clears a single bit in the pageBits at i.
func (b *pageBits) clear1(i int) {
	b[i/64] &^= 1 << (i % 64)
}

// clearRange clears bits in the range [i, i+n).
func (b *pageBits) clearRange(i, n int) {
	_ = b[i/64]
	if n == 1 {
		// Add a fast path for the n == 1 case.
		b.clear1(i)
		return
	}
	j := i + n - 1
	if i/64 == j/64 {
		b[i/64] = clearConsecBits64(b[i/64], i%64, n)
		return
	}
	_ = b[j/64]
	b[i/64] = clearConsecBits64(b[i/64], i%64, 64-i%64)
	for k := i/64 + 1; k < j/64; k++ {
		b[k] = 0
	}
	b[j/64] = clearConsecBits64(b[j/64], 0, j%64+1)
}

// clearAll frees all the bits of b.
func (b *pageBits) clearAll() {
	for i := range b {
		b[i] = 0
	}
}

// mallocBits is a page-per-bit bitmap.
//
// It wraps a pageBits with different names and additional features,
// such as summarizing and searching.
//
// TODO(mknyszek): Consider making all of mallocBits' methods accept
// a chunk index and have the searchAddr be relative to the chunk. This way,
// we avoid situations where the page allocator succeeds where it should
// have failed, making bugs more difficult to identify.
type mallocBits pageBits

// alloc allocates npages mallocBits from this mallocBits and returns
// the index where that run of contiguous mallocBits starts as well as a
// new searchAddr.
//
// If alloc fails to find any free space, it returns an index of -1 and
// the new searchAddr should be ignored.
//
// The returned searchAddr is always the index of the first free page found
// in this bitmap during the search, except if npages == 1, in which
// case it will be the index after the first free page, because that
// index is assumed to be allocated and so represents a minor
// optimization for that case.
//
// searchAddr represents the first known index and where to begin
// the search from.
func (b *mallocBits) alloc(npages uintptr, searchAddr int) (int, int) {
	if npages == 1 {
		addr := b.alloc1(searchAddr)
		// Return a searchAddr of addr + 1 since we assume addr will be
		// allocated.
		return addr, addr + 1
	} else if npages <= 64 {
		return b.allocSmallN(npages, searchAddr)
	}
	return b.allocLargeN(npages, searchAddr)
}

// alloc1 is a helper for alloc which allocates a single page from the mallocBits
// and returns the index.
//
// See alloc for an explanation of the searchAddr parameter.
func (b *mallocBits) alloc1(searchAddr int) int {
	for i := searchAddr / 64; i < len(b); i++ {
		x := b[i]
		if x == ^uint64(0) {
			continue
		}
		z := bits.TrailingZeros64(^x)
		// z will always be in the range [0, 63) since we
		// skipped the one case where z == 64 earlier.
		b[i] |= 1 << z
		return i*64 + z
	}
	return -1
}

// allocSmallN is a helper for alloc which allocates npages mallocBits from
// this mallocBits and returns the index where that run of contiguous mallocBits
// starts as well as a new searchAddr. See alloc for an explanation of the searchAddr parameter.
//
// Returns a -1 index on failure and the new searchAddr should be ignored.
//
// allocSmallN assumes npages <= 64, where any such allocation
// crosses at most one aligned 64-bit chunk boundary in the bits.
func (b *mallocBits) allocSmallN(npages uintptr, searchAddr int) (int, int) {
	end, nSearchAddr := int(0), -1
	for i := searchAddr / 64; i < len(b); i++ {
		bi := b[i]
		if bi == ^uint64(0) {
			end = 0
			continue
		}
		// First see if we can pack our allocation in the trailing
		// zeros plus the end of the last 64 bits.
		start := bits.TrailingZeros64(bi)
		if nSearchAddr == -1 {
			// The new hint is going to be at these 64 bits after any
			// 1s we file, so count trailing 1s.
			nSearchAddr = i*64 + bits.TrailingZeros64(^bi)
		}
		if end+start >= int(npages) {
			if end != 0 {
				// Set the end highest mallocBits and store.
				b[i-1] = setConsecBits64(b[i-1], 64-end, end)
			}
			// Set the npages-end lowest mallocBits and store.
			b[i] = setConsecBits64(bi, 0, int(npages)-end)
			return i*64 - end, nSearchAddr
		}
		// Next, check the interior of the 64-bit chunk.
		j := findConsecN64(^bi, int(npages))
		if j < 64 {
			// Set mallocBits [j, j+npages) and store.
			b[i] = setConsecBits64(bi, j, int(npages))
			return i*64 + j, nSearchAddr
		}
		end = bits.LeadingZeros64(bi)
	}
	return -1, nSearchAddr
}

// allocLargeN is a helper for alloc which allocates npages mallocBits from
// this mallocBits and returns the index where that run of contiguous mallocBits
// starts as well as a new searchAddr. See alloc for an explanation of the searchAddr parameter.
//
// Returns a -1 index on failure and the new searchAddr should be ignored.
//
// allocLargeN assumes npages > 64, where any such allocation
// crosses at least one aligned 64-bit chunk boundary in the bits.
func (b *mallocBits) allocLargeN(npages uintptr, searchAddr int) (int, int) {
	start, size, nSearchAddr := -1, int(0), -1
	for i := searchAddr / 64; i < len(b); i++ {
		x := b[i]
		if x == ^uint64(0) {
			size = 0
			continue
		}
		if nSearchAddr == -1 {
			// The new hint is going to be at these 64 bits after any
			// 1s we file, so count trailing 1s.
			nSearchAddr = i*64 + bits.TrailingZeros64(^x)
		}
		if size == 0 {
			size = bits.LeadingZeros64(x)
			start = i*64 + 64 - size
			continue
		}
		s := bits.TrailingZeros64(x)
		if s+size >= int(npages) {
			size += s
			break
		}
		if s < 64 {
			size = bits.LeadingZeros64(x)
			start = i*64 + 64 - size
			continue
		}
		size += 64
	}
	if size < int(npages) {
		return -1, nSearchAddr
	}
	b.allocRange(start, int(npages))
	return start, nSearchAddr
}

// allocRange allocates the range [i, i+n).
func (b *mallocBits) allocRange(i, n int) {
	(*pageBits)(b).setRange(i, n)
}

// allocAll allocates all the bits of b.
func (b *mallocBits) allocAll() {
	(*pageBits)(b).setAll()
}

// free1 frees a single page in the mallocBits at i.
func (b *mallocBits) free1(i int) {
	(*pageBits)(b).clear1(i)
}

// free frees the range [i, i+n) of pages in the mallocBits.
func (b *mallocBits) free(i, n int) {
	(*pageBits)(b).clearRange(i, n)
}

// freeAll frees all the bits of b.
func (b *mallocBits) freeAll() {
	(*pageBits)(b).clearAll()
}

// setConsecBits64 sets n consecutive bits to 1 in x starting
// at bit index i.
func setConsecBits64(x uint64, i, n int) uint64 {
	return x | ((uint64(1<<n) - 1) << i)
}

// clearConsecBits64 sets n consecutive bits to 0 in x starting
// at bit index i.
func clearConsecBits64(x uint64, i, n int) uint64 {
	return x &^ ((uint64(1<<n) - 1) << i)
}

// findConsecN64 returns the bit index of the first set of
// n consecutive 1 bits. If no consecutive set of 1 bits of
// size n may be found in c, then it returns an integer > 64.
func findConsecN64(c uint64, n int) int {
	i := 0
	cont := bits.TrailingZeros64(^c)
	for cont < n && i < 64 {
		i += cont
		i += bits.TrailingZeros64(c >> i)
		cont = bits.TrailingZeros64(^(c >> i))
	}
	return i
}
