// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/bitmap"
	"cmd/gclab/heap"
	"cmd/gclab/stats"
	"fmt"
	"log"
	"math/bits"
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
	traceFlush     = false
	traceEnqueue   = false
	traceDartboard = false
	traceScan      = false
)

const (
	scanStats      = false
	dartboardStats = false
)

func Scanner(h *heap.Heap, gcInfo *heap.GCInfo) {
	gStats = Stats{}

	//heap.FindDebug = true
	//defer func() { heap.FindDebug = false }()

	h2 := newHeap(h)
	cNet := NewCNet(DefaultDensityNetworkConfig, h2, gcInfo.Ps)

	if xxxDebug {
		fmt.Println(cNet.ToDot())
	}

	gc := scanner{
		h:      h2,
		cNet:   cNet,
		marked: bitmap.NewSet(h.ObjectIDs()),
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

	log.Printf("marked %d objects", gc.marked.Len())

	gcInfo.CompareMarks(h, gc.marked)

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

	condensers []*condenser // Indexed by size class

	// For partial spans at the beginning of a region, we need to skip some
	// number of pages, for which we use tailCondensers (we're condensing the
	// tail of the span). For partial spans at the end of a region, we need to
	// stop scanning after some number of pages, for which we use
	// headCondensers. No sized span is large enough to need both.
	tailCondensers [][heap.MaxPagesPerSpan - 1]*condenser // Indexed by size class, start page offset - 1
	headCondensers [][heap.MaxPagesPerSpan - 1]*condenser // Indexed by size class, pages count - 1
}

type arena struct {
	*heap.Arena

	dartboard    bitmap.Set[heap.Words]              // One bit per word
	regionQueued [regionsPerArena / 64]atomic.Uint64 // TODO: This is only 32 bytes per arena. Should this just be separate?
}

func newHeap(h *heap.Heap) heapExtra {
	h2 := heapExtra{
		Heap: h,
		// TODO: This will have to grow with the heap.
		arenas: make([]*arena, len(h.Arenas)),

		condensers:     make([]*condenser, len(h.SizeClasses)),
		tailCondensers: make([][heap.MaxPagesPerSpan - 1]*condenser, len(h.SizeClasses)),
		headCondensers: make([][heap.MaxPagesPerSpan - 1]*condenser, len(h.SizeClasses)),
	}
	// For each arena, add our own arena metadata
	for i, a := range h.Arenas {
		if a == nil {
			continue
		}
		h2.arenas[i] = &arena{
			Arena:     a,
			dartboard: bitmap.NewSet(heap.ArenaWords),
		}
	}
	// Create condensers
	for i, sc := range h.SizeClasses {
		if i == 0 {
			continue
		}
		cd := newCondenser(sc.ObjectBytes.Words(), heap.PageBytes.Mul(sc.SpanPages).Words())
		h2.condensers[i] = cd
		for page := 1; page < sc.SpanPages; page++ {
			h2.tailCondensers[i][page-1] = cd.slice(heap.PageWords.Mul(page), ^heap.Words(0))
			h2.headCondensers[i][page-1] = cd.slice(0, heap.PageWords.Mul(page))
		}
	}
	return h2
}

type scanner struct {
	h    heapExtra
	cNet *CNet

	marked bitmap.Set[heap.ObjectID]

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

	LAddr32s stats.Dist[int] `LAddr32 count per buffer->buffer flush`
	LAddr64s stats.Dist[int] `LAddr64 count per buffer->buffer flush`
}

var gStats Stats

func (gc *scanner) enqueue(p int, addr heap.VAddr) {
	// TODO: Exclude other obviously not-heap addresses.
	if addr < 4096 {
		return
	}
	// TODO: Redoing this over and over is silly.
	buf := gc.cNet.pBuffer(p)
	vs := buf.asVAddr()
	if int(buf.n) == len(vs) {
		gc.cNet.flush(0, p)
	}
	if traceEnqueue {
		log.Printf("enqueue %s", addr)
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

func (gc *scanner) scanRegionOld(p int, rid regionID) {
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
		log.Printf("scan region %s %s bit start %#8x", rid, rid.Range(gc.h.Heap), bitStart)
	}

	// TODO: This is silly but expedient. We should walk the spans, objects, and
	// bitmap in parallel.
	nBits := 0
	nObjects := 0
	nEnqueued := 0
	var prevObject heap.VAddr
	for arenaWord := range a.dartboard.Range(bitStart, bitStart+dartboardRegion) {
		nBits++
		a.dartboard.Remove(arenaWord)
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
		regionRange := heap.Range{Start: a.Start.Plus(bitStart.Bytes()), Len: dartboardRegion.Bytes()}
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

func (gc *scanner) scanRegion(p int, rid regionID) {
	// For each span in the region:
	//
	// Condense bitmap to 1 bit per object (see khr@'s CL) based on size class
	//
	// - 64 bits (8 bytes) of bitmap covers 512 bytes of heap.
	//
	// - All one-page spans have 1024 bits = 128 bytes of bitmap.
	//
	// - If we were to segregate large spans and use their 128 byte alignment,
	// the largest spans would have 576 bits = 72 bytes. This is a mere 64 bits
	// (8 bytes) per page.
	//
	// - This is tricky if we start part way into the a previous span. This
	// can't happen with any single-page span classes (another reason to
	// segregate them). But for larger spans I think we need special condensers
	// starting at each whole-page offset into the span. I don't think we can
	// just shift it enough to use the normal condenser because that's still
	// going to expect a whole span's worth of bitmap.
	//
	// - Perhaps we have a condenser per span class per page offset. We'd also a
	// little more metadata, like how many bytes of the first object we skipped,
	// and maybe the object index (though we could calculate either of these on
	// the fly easily enough).
	//
	// Subtract mark bits from condensed bitmap. Atomic OR bits into mark bits.
	//
	// If noscan, move on to the next span.
	//
	// Turn set bits into object offsets.
	//
	// Scan each object. (Maybe scan each object in parallel? Tricky with
	// different pointer/scalar bitmaps.)

	// TODO: If this is happening concurrently, we need a better answer here. If
	// we clear this bit now, the region could get re-enqueued and then a
	// concurrent scan of it could start. If we wait until we're done scanning
	// the region, then more pointers could have been enqueued (by us or other
	// threads). Given that we want some way to lock the region anyway, maybe we
	// lock it, copy the bitmap out, clear the bitmap, clear the region bit, and
	// then unlock it. (We can't keep it locked when we might enqueue anything
	// because that could cause a lock cycle.)
	arenaRegion := rid.arenaRegion()
	a := gc.h.arenas[rid.arenaID()]
	a.regionQueued[arenaRegion/64].And(^(1 << (arenaRegion % 64)))

	bitStart := dartboardRegion.Mul(int(arenaRegion))
	if traceScan {
		log.Printf("scan region %s %s bit start %#8x", rid, rid.Range(gc.h.Heap), bitStart)
	}

	// Iterate over each span in the region.
	addr := rid.toVAddr(gc.h.Heap)
	endAddr := addr.Plus(regionHeapBytes)
	for addr < endAddr {
		span := gc.h.FindSpan(addr) // XXX All spans are in the same arena, so we could skip the arena lookup
		if span == nil {
			addr = addr.Plus(heap.PageBytes)
			continue
		}
		spanEnd := span.Start.Plus(heap.PageBytes.Mul(span.NPages))
		if span.Start < addr || spanEnd > endAddr {
			// This span is only partially in this region.
			startAddr := max(span.Start, addr)
			endAddr := min(spanEnd, endAddr)
			startWord := startAddr.ArenaOffset().Words()
			nWords := endAddr.Minus(startAddr).Words()
			if traceScan {
				spanRange := heap.Range{Start: startAddr, Len: endAddr.Minus(startAddr)}
				log.Printf("  partial span %s => %s, dartboard bits [%d,%d)", span.Range(), spanRange, startWord, startWord+nWords)
			}
			dartboard := a.dartboard.Words(startWord, uint(nWords))
			if span.SizeClass == nil {
				gc.scanLargeSpan(p, span, dartboard)
			} else if endAddr == spanEnd {
				// We're condensing just the tail of this span.
				cd := gc.h.tailCondensers[span.SizeClass.ID][startAddr.Minus(span.Start).Div(heap.PageBytes)-1]
				gc.scanSpan(p, span, dartboard, cd)
			} else if startAddr == span.Start {
				// We're condensing just the head of this span.
				cd := gc.h.headCondensers[span.SizeClass.ID][endAddr.Minus(span.Start).Div(heap.PageBytes)-1]
				gc.scanSpan(p, span, dartboard, cd)
			} else {
				log.Fatalf("span %s extends beyond beginning and end of region %s", span.Range(), rid.Range(gc.h.Heap))
			}
			clear(dartboard)
		} else {
			// Full span
			startWord := span.Start.ArenaOffset().Words()
			nWords := heap.PageWords.Mul(span.NPages)
			dartboard := a.dartboard.Words(startWord, uint(nWords))
			if traceScan {
				log.Printf("  full span %s, dartboard bits [%d,%d)", span.Range(), startWord, startWord+nWords)
			}
			if span.SizeClass == nil {
				gc.scanLargeSpan(p, span, dartboard)
			} else {
				cd := gc.h.condensers[span.SizeClass.ID]
				gc.scanSpan(p, span, dartboard, cd)
			}
			clear(dartboard)
		}

		addr = spanEnd
	}
}

func (gc *scanner) scanSpan(p int, span *heap.Span, dartboard []uint64, cd *condenser) {
	// Get mark bits for this span.
	nObj := span.NObjects()
	marks := gc.marked.Words(span.FirstObject, uint(nObj))

	// Condense the dartboard into one bit per object.
	objDarts := make([]uint64, len(marks), (heap.MaxObjsPerSpan+63)/64)
	cd.do(dartboard, objDarts)

	if traceScan {
		log.Printf("    marks        %016x", marks)
		log.Printf("    object darts %016x", objDarts)
	}

	// Subtract marks from object darts and set new mark bits.
	anyGrey := false
	for i := range marks {
		// Clear marked objects from the darts. After this objDarts is
		// exactly the grey objects.
		//
		// Reading from marks is intentionally racy because the worst that
		// can happen is that we mark an object twice.
		//
		// TODO: How often do we mark twice?
		toGrey := objDarts[i] &^ marks[i]
		objDarts[i] = toGrey
		// Set marks for objects with darts.
		//
		// This must be atomic because 1. there can be races between
		// scanning for a span that crosses mark regions, 2. there can be
		// races with allocate-black.
		//
		// We could use the result for clearing darts, but don't because on
		// some architectures (notably x86), using the result requires a
		// CAS-based atomic.
		//
		// TODO: If this atomic is expensive, there are a lot of cases where
		// it *doesn't* have to be atomic.
		if toGrey != 0 {
			atomic.OrUint64(&marks[i], toGrey)
			anyGrey = true
		}
	}
	if !anyGrey {
		// Either there were no darts in this span, or they were all already
		// marked.
		if traceScan {
			log.Printf("    no objects to grey")
		}
		return
	}

	if traceScan {
		log.Printf("    grey darts   %016x", objDarts)
	}

	if span.NoScan {
		if traceScan {
			log.Printf("    noscan span")
		}
		return
	}

	// Scan grey objects.
	spanBase := span.Start
	objBytes := span.SizeClass.ObjectBytes
	for i, darts := range objDarts {
		for range bits.OnesCount64(darts) {
			bitI := bits.TrailingZeros64(darts)
			base := spanBase.Plus(objBytes.Mul(i*64 + bitI))
			if traceScan {
				log.Printf("    scan %s %s", base, objBytes)
			}
			gc.scanObject(p, base, objBytes)
			darts &^= 1 << bitI
		}
	}
}

// scanLargeSpan scans a span that contains only a single (large) object.
func (gc *scanner) scanLargeSpan(p int, span *heap.Span, dartboard []uint64) {
	if gc.marked.Has(span.FirstObject) {
		// Nothing to do.
		return
	}

	// Check if there are any darts in this span.
	var anyDarts uint64
	for _, d := range dartboard {
		anyDarts |= d
	}
	if anyDarts == 0 {
		return
	}

	// Mark the object.
	gc.marked.Add(span.FirstObject)

	// Scan the object.
	gc.scanObject(p, span.Start, heap.PageBytes.Mul(span.NPages))
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
	gc.scanObject(p, base, span.ObjectBytes())
	return true
}

func (gc *scanner) scanObject(p int, base heap.VAddr, length heap.Bytes) {
	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, length))
	for _, val := range mem {
		if val != 0 {
			// XXX Individual enqueue is very slow
			gc.enqueue(p, val)
		}
	}
}
