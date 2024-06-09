// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"fmt"
	"internal/trace/testdata/cmd/gclab/bitmap"
	"internal/trace/testdata/cmd/gclab/heap"
	"internal/trace/testdata/cmd/gclab/stats"
	"log"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync/atomic"
)

type drainPolicy int

const (
	drainFIFO drainPolicy = iota
	drainSparsest
	drainDensest
	drainRandom
)

const (
	traceFlush   = false
	traceEnqueue = false
	traceScan    = false
)

const (
	scanStats      = true
	dartboardStats = true
)

func Scanner(h *heap.Heap, gcInfo *heap.GCInfo) {
	gStats = Stats{}

	//heap.FindDebug = true
	//defer func() { heap.FindDebug = false }()

	cNet := NewCNet(DefaultDensityNetworkConfig, newHeap(h), gcInfo.Ps)

	if xxxDebug {
		fmt.Println(cNet.ToDot())
	}

	gc := scanner{
		h:    h,
		cNet: cNet,
	}
	gc.regionScanCount = make(map[regionID]int)

	// Mark allocate-black objects. This way if we find a pointer to them, we
	// won't try scanning into them.
	for _, allocBlack := range gcInfo.AllocBlack {
		_, _, objID := h.FindObject(allocBlack)
		gc.marked.Add(objID)
	}
	// Gray roots
	//
	// TODO: Do we need to "distribute" this?
	for _, root := range gcInfo.Roots {
		gc.enqueue(0, root)
	}
	// Gray write barrier roots
	//
	// TODO: This is clearly the wrong time to do this, but I'm not sure when
	// would be better. Toward the end? Spread it out?
	for _, root := range gcInfo.WBRoots {
		gc.enqueue(0, root)
	}

	gc.drain(0, drainFIFO)

	log.Printf("marked %d objects", gc.marked.Set().Len())

	gcInfo.CompareMarks(h, gc.marked.Set())

	for _, count := range gc.regionScanCount {
		gStats.RegionScanCount.Add(count)
	}
	stats.ForEachDist(&gStats, func(dist stats.DistCommon, tag reflect.StructTag) {
		log.Printf("%s\n%s", tag, dist)
	})
}

// heapExtra wraps heap.Heap with extra data specific to this experiment.
type heapExtra struct {
	*heap.Heap
	arenas []*arena
}

type arena struct {
	*heap.Arena

	bitmap       bitmap.Set[heap.Words]              // One bit per word
	regionQueued [regionsPerArena / 64]atomic.Uint64 // TODO: This is only 32 bytes per arena. Should this just be separate?
}

func newHeap(h *heap.Heap) heapExtra {
	h2 := heapExtra{
		Heap: h,
		// TODO: This will have to grow with the heap.
		arenas: make([]*arena, len(h.Arenas)),
	}
	// For each arena, add our own arena metadata
	for i, a := range h.Arenas {
		if a == nil {
			continue
		}
		h2.arenas[i] = &arena{
			Arena:  a,
			bitmap: bitmap.NewSet(heap.ArenaWords),
		}
	}
	return h2
}

type scanner struct {
	h    *heap.Heap
	cNet *CNet

	marked bitmap.DynSet[heap.ObjectID]

	regionScanCount map[regionID]int
}

type Stats struct {
	RegionBitDensity          stats.Dist[float64] `region bit density`
	RegionObjectDensity       stats.Dist[float64] `region object density`
	RegionObjectMarkedDensity stats.Dist[float64] `fraction of newly marked objected per region scan`
	DartboardDupBits          stats.Dist[float64] `fraction of dartboard region already set per dartboard flush`
	DartboardNewBits          stats.Dist[float64] `fraction of dartboard region newly set per dartboard flush`
	DartboardAddrs            stats.Dist[int]     `count of addresses per flush to dartboard`

	RegionScanCount stats.Dist[int] `number of times each mark region is scanned`
}

var gStats Stats

func (gc *scanner) enqueue(p int, addr heap.VAddr) {
	if addr < 4096 {
		return
	}
	// TODO: Redoing this over and over is silly.
	buf := gc.cNet.pBuffer(p)
	vs := buf.asVAddr()
	if int(buf.n) == len(vs) {
		gc.cNet.flush(0, p)
	}
	vs[buf.n] = addr
	buf.n++
}

func (gc *scanner) drain(p int, policy drainPolicy) {
	c := gc.cNet
	for {
		// Try to scan a dense region.
		if len(c.regionQueue) > 0 {
			var rid regionID
			switch policy {
			default:
				panic("unimplemented drain policy")
			case drainFIFO:
				rid = c.regionQueue[0]
				c.regionQueue = c.regionQueue[1:]
			case drainRandom:
				n := rand.IntN(len(c.regionQueue))
				rid = c.regionQueue[n]
				c.regionQueue = slices.Delete(c.regionQueue, n, n+1)
			}

			gc.scanRegion(p, rid)
		} else if buf := c.getScanBuf(); buf != nil {
			gc.scanBuf(p, buf)
		} else {
			// No more work
			return
		}
	}
}

func (gc *scanner) scanRegion(p int, rid regionID) {
	// XXX How to deal with objects that span regions? I think just, if there
	// are any bits in my region, I consider scanning the whole object, and if
	// it's already marked, there's nothing to do.

	c := gc.cNet
	a := c.heap.arenas[rid.arenaID()]
	arenaRegion := rid.arenaRegion()
	// TODO: If this is happening concurrently, we need a better answer here. If
	// we clear this bit now, the region could get re-enqueued and then a
	// concurrent scan of it could start. If we wait until we're done scanning
	// the region, then more pointers could have been enqueued (by us or other
	// threads). Given that we want some way to lock the region anyway, maybe we
	// lock it, copy the bitmap out, clear the bitmap, clear the region bit, and
	// then unlock it. (We can't keep it locked when we might enqueue anything
	// because that could cause a lock cycle.)
	a.regionQueued[arenaRegion/64].And(^(1 << (arenaRegion % 64)))

	bitStart := dartboardRegion.Mul(int(arenaRegion))
	if traceScan {
		log.Printf("scan region %s %s bit start %#8x", rid, c.regionRange(rid), bitStart)
	}

	// TODO: This is silly but expedient. We should walk the spans, objects, and
	// bitmap in parallel.
	nBits := 0
	nObjects := 0
	nEnqueued := 0
	var prevObject heap.VAddr
	for arenaWord := range a.bitmap.Range(bitStart, bitStart+dartboardRegion) {
		nBits++
		a.bitmap.Remove(arenaWord)
		if traceScan {
			log.Printf("  bit %#08x", arenaWord)
		}
		addr := a.Start.Plus(arenaWord.Bytes())
		base, span, objID := gc.h.FindObject(addr)
		if base != 0 && base != prevObject {
			prevObject = base
			nObjects++
			// TODO: Skip forward in the bitmap to the end of the object.
			if gc.mark(p, base, span, objID) {
				nEnqueued++
			}
		}
	}

	if scanStats {
		// Update stats.
		gStats.RegionBitDensity.Add(float64(nBits) / float64(dartboardRegion))

		// Figure out how much objects were in this region
		totalObjects := 0
		regionRange := heap.Range{a.Start.Plus(bitStart.Bytes()), dartboardRegion.Bytes()}
		for span := range gc.h.SpansIn(regionRange) {
			for i := range span.NObjects() {
				if span.ObjectRange(i).Overlaps(regionRange) {
					totalObjects++
				}
			}
		}
		gStats.RegionObjectDensity.Add(float64(nObjects) / float64(totalObjects))
		gStats.RegionObjectMarkedDensity.Add(float64(nEnqueued) / float64(totalObjects))

		gc.regionScanCount[rid]++
	}
}

func (gc *scanner) scanBuf(p int, buf *buffer) {
	// Swap out the buffer so it doesn't get flushed out from under us.
	tmpBuf := gc.cNet.tmpBufs.Get().(*buffer)
	defer gc.cNet.tmpBufs.Put(tmpBuf)
	tmpBuf.typ, tmpBuf.start = buf.typ, buf.start
	tmpBuf.n, buf.n = buf.n, 0
	tmpBuf.data, buf.data = buf.data, tmpBuf.data
	buf = nil

	h := gc.cNet.heap
	if tmpBuf.typ == bufferVAddr {
		tmpBuf.mapToLAddr64(h.Heap)
	}

	// Mark objects from the buffer.
	if tmpBuf.typ == bufferLAddr32 {
		lAddr32s := tmpBuf.asLAddr32()[:tmpBuf.n]
		for _, lAddr32 := range lAddr32s {
			lAddr := lAddr32.ToLAddr(tmpBuf.start)
			vAddr := h.LAddrToVAddr(lAddr)
			if traceScan {
				log.Printf("scan32 VAddr %s", vAddr)
			}
			base, span, objID := h.FindObject(vAddr)
			if base != 0 {
				gc.mark(p, base, span, objID)
			}
		}
	} else {
		lAddr64s := tmpBuf.asLAddr64()[:tmpBuf.n]
		for _, lAddr64 := range lAddr64s {
			lAddr := lAddr64.ToLAddr()
			vAddr := h.LAddrToVAddr(lAddr)
			if traceScan {
				log.Printf("scan64 VAddr %s", vAddr)
			}
			base, span, objID := h.FindObject(vAddr)
			if base != 0 {
				gc.mark(p, base, span, objID)
			}
		}
	}
}

func (gc *scanner) mark(p int, base heap.VAddr, span *heap.Span, objID heap.ObjectID) bool {
	if gc.marked.Has(objID) {
		return false
	}
	gc.marked.Add(objID)
	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, span.ObjectBytes()))
	for _, val := range mem {
		if val != 0 {
			gc.enqueue(p, val)
		}
	}
	return true
}
