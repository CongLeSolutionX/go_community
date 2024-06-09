// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/bitmap"
	"cmd/gclab/heap"
	"cmd/gclab/invivo"
	"cmd/gclab/stats"
	"fmt"
	"iter"
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
		invivo.Invalidate()
		log.Printf("enqueue %s", addr)
	}
	vs[buf.n] = addr
	buf.n++
}

func (gc *scanner) drain(p int, policy drainPolicy) {
	bench := benchmarkDrain.Start()
	defer func() {
		bench.StopTimer()

		// This computation below is kind of expensive, so invalidate any
		// benchmarks we have running.
		invivo.Invalidate()

		// How many bytes were marked?
		var marked heap.Bytes
		for va, oid := range gc.h.Heap.Objects() {
			if gc.marked.Has(oid) {
				span := gc.h.Heap.FindSpan(va)
				marked += span.ObjectBytes()
			}
		}
		bench.SetMetric(float64(marked)/1e6/bench.Elapsed().Seconds(), metricMarkedMBPerSec)

		// How much span memory is there?
		var spanBytes heap.Bytes
		for _, arena := range gc.h.Heap.Arenas {
			var prev *heap.Span
			for _, span := range arena.SpanMap {
				if span != nil && span != prev {
					prev = span
					spanBytes += heap.PageBytes.Mul(span.NPages)
				}
			}
		}
		bench.SetMetric(float64(spanBytes)/1e6/bench.Elapsed().Seconds(), metricSpanMBPerSec)

		bench.Done()
	}()

	// TODO: Structurally, I could better separate the "work manager" from the
	// "work processor", where the former consists of the concentrator network,
	// dartboard, and region queue, and the latter consists of scanning logic.
	// The interface between these is quite small. The work manager needs to
	// attach things to heap structures but otherwise needs to know very little
	// about the heap.

	c := gc.cNet
	for {
		// Try to scan a dense region.
		if len(c.regionQueue) > 0 {
			var rid regionID
			switch policy {
			default:
				panic("unimplemented drain policy")

			case drainDensest:
				// TODO: This is obviously a terrible way to do this.
				invivo.Invalidate()
				var maxI, maxBits int
				for i, rid := range c.regionQueue {
					arenaRegion := rid.arenaRegion()
					a := gc.h.arenas[rid.arenaID()]
					bitStart := dartboardRegion.Mul(int(arenaRegion))
					nBits := a.dartboard.LenRange(bitStart, bitStart+dartboardRegion)
					if int(nBits) > maxBits {
						maxBits = int(nBits)
						maxI = i
					}
				}
				rid = c.regionQueue[maxI]
				c.regionQueue = slices.Delete(c.regionQueue, maxI, maxI+1)

			case drainSparsest:
				// TODO: This is obviously a terrible way to do this.
				invivo.Invalidate()
				var minI, minBits int
				for i, rid := range c.regionQueue {
					arenaRegion := rid.arenaRegion()
					a := gc.h.arenas[rid.arenaID()]
					bitStart := dartboardRegion.Mul(int(arenaRegion))
					nBits := a.dartboard.LenRange(bitStart, bitStart+dartboardRegion)
					if minBits == 0 || int(nBits) < minBits {
						minBits = int(nBits)
						minI = i
					}
				}
				rid = c.regionQueue[minI]
				c.regionQueue = slices.Delete(c.regionQueue, minI, minI+1)

			case drainFIFO:
				rid = c.regionQueue[0]
				c.regionQueue = c.regionQueue[1:]

			case drainRandom:
				invivo.Invalidate()
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
	// it's already marked, there's nothing to do. Maybe we can have a better
	// policy that also supports oblets.

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
	defer statsScanRegion()()

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
	if traceScanOnce {
		invivo.Invalidate()
		log.Printf("scan region %s %s bit start %#8x", rid, rid.Range(gc.h.Heap), bitStart)
	}

	if scanStats {
		invivo.Invalidate()
		// Update stats.
		nBits := a.dartboard.LenRange(bitStart, bitStart+dartboardRegion)
		gStats.RegionBitDensity.Add(float64(nBits) / float64(dartboardRegion))
	}

	// Iterate over each span in the region.
	addr := rid.toVAddr(gc.h.Heap)
	endAddr := addr.Plus(regionHeapBytes)
	for addr < endAddr {
		span := gc.h.FindSpan(addr) // XXX All spans are in the same arena, so we could skip the arena lookup
		if span == nil {
			incStat(&scanRegionStatsOne.pagesSkipped, 1)
			addr = addr.Plus(heap.PageBytes)
			continue
		}
		incStat(&scanRegionStatsOne.spans, 1)
		spanEnd := span.Start.Plus(heap.PageBytes.Mul(span.NPages))
		if span.Start < addr || spanEnd > endAddr {
			// This span is only partially in this region. We scan only the
			// objects that overlap with this region in this span. That does
			// mean we sort of "round out" from the region to object boundaries
			// on either side.
			//
			// TODO: Figure out how to split large objects into oblets.
			// Presumably those oblets should be aligned with regions. I think
			// having a per-region "scanned" bit is unavoidable. To avoid n^2, I
			// could have a rule like "if I'm scanning the head, enqueue all of
			// the other oblets; if I'm scanning layer, enqueue the head".
			incStat(&scanRegionStatsOne.partialSpans, 1)
			startAddr := max(span.Start, addr)
			endAddr := min(spanEnd, endAddr)
			incStat(&scanRegionStatsOne.pagesScanned, endAddr.Minus(startAddr).Div(heap.PageBytes))
			if scanStats {
				spanRange := heap.Range{Start: startAddr, Len: endAddr.Minus(startAddr)}
				spanObjects := 0
				for i := range span.NObjects() {
					if span.ObjectRange(i).Overlaps(spanRange) {
						spanObjects++
					}
				}
				incStat(&scanRegionStatsOne.objects, spanObjects)
			}

			startWord := startAddr.ArenaOffset().Words()
			nWords := endAddr.Minus(startAddr).Words()
			if traceScanOnce {
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
			incStat(&scanRegionStatsOne.fullSpans, 1)
			incStat(&scanRegionStatsOne.pagesScanned, span.NPages)
			incStat(&scanRegionStatsOne.objects, span.NObjects())

			startWord := span.Start.ArenaOffset().Words()
			nWords := heap.PageWords.Mul(span.NPages)
			dartboard := a.dartboard.Words(startWord, uint(nWords))
			if traceScanOnce {
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

	if scanStats {
		gStats.RegionObjectMarkedDensity.Add(float64(scanRegionStatsOne.objectsScanned) / float64(scanRegionStatsOne.objects))
		gc.regionScanCount[rid]++
	}
}

func (gc *scanner) scanSpan(p int, span *heap.Span, dartboard []uint64, cd *condenser) {
	// TODO: Between 25% and 50% of the time, there are NO DARTS in a whole
	// span. We should summarize at the page level to skip whole spans. With 256
	// KiB regions and 8 KiB pages, that's 32 bits per region. So we could
	// easily use that instead of the existing region queued bits.

	// Get mark bits for this span.
	nObj := span.NObjects()
	marks := gc.marked.Words(span.FirstObject, uint(nObj))

	// Condense the dartboard into one bit per object.
	objDarts := make([]uint64, len(marks), (heap.MaxObjsPerSpan+63)/64)
	cd.do(dartboard, objDarts)

	if traceScanOnce {
		invivo.Invalidate()
		log.Printf("    marks        %016x", marks)
		log.Printf("    object darts %016x", objDarts)
	}
	if scanStats {
		invivo.Invalidate()
		queuedWords := bitmap.FromWords[uint64](dartboard).Len()
		queuedObjs := bitmap.FromWords[uint64](objDarts).Len()
		// XXX Wrong for partial spans
		nWords := heap.PageWords.Mul(span.NPages)
		gStats.SpanQueuedWordDensity.Add(float64(queuedWords) / float64(nWords))
		gStats.SpanQueuedObjectDensity.Add(float64(queuedObjs) / float64(nObj))
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
	if scanStats {
		greyObjs := bitmap.FromWords[uint64](objDarts).Len()
		gStats.SpanGreyObjectDensity.Add(float64(greyObjs) / float64(nObj))
	}
	if !anyGrey {
		// Either there were no darts in this span, or they were all already
		// marked.
		if traceScanOnce {
			log.Printf("    no objects to grey")
		}
		return
	}

	if traceScanOnce {
		log.Printf("    grey darts   %016x", objDarts)
	}

	if span.NoScan {
		if traceScanOnce {
			log.Printf("    noscan span")
		}
		return
	}

	// Prepare for scanning.
	buf := gc.cNet.pBuffer(p)
	bufPos := &buf.n
	vs := buf.asVAddr()

	// Scan grey objects.
	spanBase := span.Start
	objBytes := span.SizeClass.ObjectBytes
	for objIndex := range uint64Bits(objDarts) {
		base := spanBase.Plus(objBytes.Mul(objIndex))
		if traceScanOnce {
			log.Printf("    scan %s %s", base, objBytes)
		}
		gc.scanObject(p, vs, bufPos, base, objBytes)
	}
}

// uint64Bits iterates over the bit indexes of all set bits in b.
func uint64Bits(b []uint64) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i, val := range b {
			for range bits.OnesCount64(val) {
				bitI := bits.TrailingZeros64(val)
				yield(i*64 + bitI)
				val &^= 1 << bitI
			}
		}
	}
}

// scanLargeSpan scans a span that contains only a single (large) object.
func (gc *scanner) scanLargeSpan(p int, span *heap.Span, dartboard []uint64) {
	incStat(&scanRegionStatsOne.largeSpans, 1)

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

	// Prepare for scanning.
	buf := gc.cNet.pBuffer(p)
	bufPos := &buf.n
	vs := buf.asVAddr()

	// Scan the object.
	gc.scanObject(p, vs, bufPos, span.Start, heap.PageBytes.Mul(span.NPages))
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
			if traceScanOnce {
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
			if traceScanOnce {
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
	gc.scanObjectSlow(p, base, span.ObjectBytes())
	return true
}

func (gc *scanner) scanObjectSlow(p int, base heap.VAddr, length heap.Bytes) {
	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, length))
	incStat(&scanRegionStatsOne.objectsScanned, 1)
	incStat(&scanRegionStatsOne.wordsScanned, len(mem))
	for _, val := range mem {
		if val != 0 {
			// XXX Individual enqueue is very slow
			gc.enqueue(p, val)
		}
	}
}

func (gc *scanner) scanObject(p int, buf []heap.VAddr, bufPos *int32, base heap.VAddr, length heap.Bytes) {
	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, length))
	incStat(&scanRegionStatsOne.objectsScanned, 1)
	incStat(&scanRegionStatsOne.wordsScanned, len(mem))
	pos := *bufPos
	for _, val := range mem {
		if val >= 4096 {
			if int(pos) == len(buf) {
				*bufPos = pos
				gc.cNet.flush(0, p)
				pos = *bufPos
			}
			buf[pos] = val
			pos++
		}
	}
	*bufPos = pos
}
