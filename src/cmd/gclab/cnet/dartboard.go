// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/bitmap"
	"cmd/gclab/heap"
	"cmd/gclab/invivo"
	"fmt"
	"log"
	"math/bits"
)

// TODO: If we were to segregate large objects and small objects, we could take
// advantage of the high alignment of large objects to dramatically reduce the
// dartboard size. If that reduction is enough, it might be reasonable to have a
// second mark bitmap that we use to avoid unnecessary region enqueuing.
//
// (TODO: Calculate how much this would save, without actually doing the
// implementation.)

const (
	// dartboardBitsPerRegion is the size of an atomic unit of the dartboard, in
	// either bits of dartboard bitmap, or words of heap. Dartboard regions are
	// the smallest unit of scanning.
	dartboardBitsPerRegion heap.Words = 32 << 10

	bytesPerRegion = heap.WordBytes * heap.Bytes(dartboardBitsPerRegion)

	// regionsPerArena is the number of dartboard regions in an arena.
	regionsPerArena = arenaRegionID(heap.ArenaWords / dartboardBitsPerRegion)
)

type anyLAddr interface {
	LAddr32 | LAddr64
	ArenaWord() heap.Words
}

// arenaRegionBitmap is a bitmap over regions within an arena.
type arenaRegionBitmap [regionsPerArena / 64]uint64

func (b *arenaRegionBitmap) set(arenaWord heap.Words) {
	region := arenaWord / dartboardBitsPerRegion
	b[region/64] |= 1 << (region % 64)
}

func copyToDartboard[T anyLAddr](addrs []T, arenaBitmap bitmap.Set[heap.Words], regionBitmap *arenaRegionBitmap, heapSpan heap.Bytes) {
	// TODO: Is it better to further sort these addresses to improve locality in
	// the bitmap? We could sort down to just the region, though it may make
	// sense to sort all the way.
	//
	// TODO: Is it worth checking if the bits in the arena are already set and
	// only queueing regions with newly set bits?
	//
	// TODO: Should this be excluding already marked objects? That would require
	// another bitmap, which is unfortunate, but may converge a lot faster.
	// Compressing the bitmap for large objects may make it practical to have a
	// second bitmap.
	var dupBits, newBits int
	for _, addr := range addrs {
		word := addr.ArenaWord()
		if traceDartboard || traceFlushAddrs {
			invivo.Invalidate()
			if arenaBitmap.Has(word) {
				log.Printf("  %s => dartboard bit %#08x already set", addr, word)
			} else {
				log.Printf("  %s => dartboard bit %#08x set", addr, word)
			}
		}
		if dartboardStats {
			if arenaBitmap.Has(word) {
				dupBits++
			} else {
				newBits++
			}
		}
		arenaBitmap.Add(word)
		regionBitmap.set(word)
	}
	if dartboardStats {
		maxBits := float64(heapSpan.Words())
		gStats.DartboardDupBits.Add(float64(dupBits) / maxBits)
		gStats.DartboardNewBits.Add(float64(newBits) / maxBits)
		gStats.DartboardAddrs.Add(len(addrs))
	}
}

func (c *CNet) enqueueRegions(arena *arena, regionBitmap *arenaRegionBitmap) {
	for i, word := range regionBitmap {
		if word == 0 {
			continue
		}
		// Check if all of the bits are already set.
		word &^= arena.regionQueued[i].Load()
		if word == 0 {
			continue
		}
		// Add the new queue bits. This atomic ensures exclusive ownership of
		// queueing these regions.
		word &^= arena.regionQueued[i].Or(word)

		// Enqueue regions.
		for word != 0 {
			lo := bits.TrailingZeros64(word)
			word &^= 1 << lo
			rid := makeRegionID(arena.ID, arenaRegionID(i*64+lo))
			c.regionQueue = append(c.regionQueue, rid)
			if traceEnqueue {
				invivo.Invalidate()
				log.Printf("  enqueue region %s", rid)
			}
		}
	}
}

type regionID uint

// Region index within an arena.
//
// TODO: This is only used for the work queue. It might make sense to generalize
// this for other kinds of work.
type arenaRegionID uint64

func makeRegionID(arena heap.ArenaID, arenaRegion arenaRegionID) regionID {
	if arenaRegion >= regionsPerArena {
		panic("arenaRegion out of bounds")
	}
	return regionID(arenaRegionID(arena)*regionsPerArena + arenaRegion)
}

func (id regionID) arenaID() heap.ArenaID {
	return heap.ArenaID(arenaRegionID(id) / regionsPerArena)
}

func (id regionID) arenaRegion() arenaRegionID {
	return arenaRegionID(id) % regionsPerArena
}

func (id regionID) toVAddr(h *heap.Heap) heap.VAddr {
	arenaStart := h.Arenas[id.arenaID()].Start
	arenaOff := bytesPerRegion.Mul(int(id.arenaRegion()))
	return arenaStart.Plus(arenaOff)
}

func (id regionID) Range(h *heap.Heap) heap.Range {
	rStart := h.Arenas[id.arenaID()].Start.Plus(bytesPerRegion.Mul(int(id.arenaRegion())))
	return heap.Range{Start: rStart, Len: bytesPerRegion}
}

func (id regionID) String() string {
	return fmt.Sprintf("%#x (%s/%#x)", uint(id), id.arenaID(), id.arenaRegion())
}
