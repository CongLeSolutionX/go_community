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
	drainLIFO
	drainSparsest
	drainDensest
	drainRandom
	drainAddress
)

const useAVX = false

const (
	traceFlush      = false
	traceFlushAddrs = false
	traceEnqueue    = false
	traceDartboard  = false
	traceScan       = false
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
		perP:   make([]perP, gcInfo.Ps),
	}
	for i := range gc.perP {
		gc.perP[i].id = i
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
		gc.enqueue(&gc.perP[0], root)
	}
	// Gray write barrier roots
	//
	// TODO: This is clearly the wrong time to do this, but I'm not sure when
	// would be better. Toward the end? Spread it out?
	for _, root := range gcInfo.WBRoots {
		gc.enqueue(&gc.perP[0], root)
	}

	gc.drain(&gc.perP[0], drainFIFO)

	log.Printf("cnet marked %d objects", gc.marked.Len())

	gcInfo.CompareMarks(h, gc.marked)

	invivo.Report()
	return

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

	perP []perP
}

type perP struct {
	id int

	dartboardCopy [dartboardBitsPerRegion / 64]uint64

	stats statsPerP
}

func (gc *scanner) enqueue(p *perP, addr heap.VAddr) {
	// TODO: Exclude other obviously not-heap addresses.
	if addr < 4096 {
		return
	}
	// TODO: Redoing this over and over is silly.
	buf := gc.cNet.pBuffer(p.id)
	vs := buf.asVAddr()
	if int(buf.n) == len(vs) {
		gc.cNet.flush(p, 0, p.id)
	}
	if traceEnqueue {
		invivo.Invalidate()
		log.Printf("enqueue %s", addr)
	}
	vs[buf.n] = addr
	buf.n++
}

func (gc *scanner) drain(p *perP, policy drainPolicy) {
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
		metricMarkedMBPerSec.Set(bench, float64(marked)/1e6, bench.Elapsed().Seconds())

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
		metricSpanMBPerSec.Set(bench, float64(spanBytes)/1e6, bench.Elapsed().Seconds())

		bench.Done()
	}()

	// TODO: Structurally, I could better separate the "work manager" from the
	// "work processor", where the former consists of the concentrator network,
	// dartboard, and region queue, and the latter consists of scanning logic.
	// The interface between these is quite small. The work manager needs to
	// attach things to heap structures but otherwise needs to know very little
	// about the heap.

	var lastRID regionID
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
					bitStart := dartboardBitsPerRegion.Mul(int(arenaRegion))
					nBits := a.dartboard.LenRange(bitStart, bitStart+dartboardBitsPerRegion)
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
					bitStart := dartboardBitsPerRegion.Mul(int(arenaRegion))
					nBits := a.dartboard.LenRange(bitStart, bitStart+dartboardBitsPerRegion)
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

			case drainLIFO:
				rid = c.regionQueue[len(c.regionQueue)-1]
				c.regionQueue = c.regionQueue[:len(c.regionQueue)-1]

			case drainRandom:
				invivo.Invalidate()
				n := rand.IntN(len(c.regionQueue))
				rid = c.regionQueue[n]
				c.regionQueue = slices.Delete(c.regionQueue, n, n+1)

			case drainAddress:
				invivo.Invalidate()
				var minI int
				var minRID regionID
				for i, rid := range c.regionQueue {
					if i == 0 || rid-lastRID < minRID-lastRID {
						minI, minRID = i, rid
					}
				}
				rid = c.regionQueue[minI]
				c.regionQueue = slices.Delete(c.regionQueue, minI, minI+1)
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

func (gc *scanner) scanRegion(p *perP, rid regionID) {
	defer p.stats.startScanRegion()()

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
	// because that could cause a lock cycle, unless we detect that cycle and
	// prevent it somehow.)
	regionInArena := rid.arenaRegion()
	a := gc.h.arenas[rid.arenaID()]
	a.regionQueued[regionInArena/64].And(^(1 << (regionInArena % 64)))

	// TODO: We really need a naming convention (or type system support?) for
	// keeping track of what all of these indexes are relative to. E.g., is the
	// word offset relative to address 0, the arena, the region, the span, the
	// page?

	bitInArena := dartboardBitsPerRegion.Mul(int(regionInArena))
	if traceScanOnce {
		invivo.Invalidate()
		log.Printf("scan region %s %s bit start %#8x", rid, rid.Range(gc.h.Heap), bitInArena)
	}

	if scanStats {
		invivo.Invalidate()
		// Update stats.
		nBits := a.dartboard.LenRange(bitInArena, bitInArena+dartboardBitsPerRegion)
		gStats.RegionBitDensity.Add(float64(nBits) / float64(dartboardBitsPerRegion))
	}

	// Make a copy of the dartboard and clear the real dartboard. We need to do
	// this in case there are more writes to the dartboard while we're scanning
	// this region.
	//
	// TODO: This is overly pessimistic. Writes to this dartboard region while
	// we're scanning are very rare. It would be nice to do this lazily. It's
	// tempting to make flushing to this region just leave behind pointers in
	// the buffer, but that has the potential to back up and deadlock flushing.
	// We could have a per-P side dartboard that gets used lazily for flushing
	// in this situation.
	//
	// TODO: Measure how often this happens.
	//
	// TODO: There are a few things we could fast-path for immediate rescanning.
	// If we scan a pointer back into our own region, we could queue that up for
	// rescanning. In that case, we may want to make a local decision about
	// whether to do dense or sparse scanning, which might make it not even
	// worth it. If we flush into our own region (possible though I think much
	// less likely), we could queue that up for rescanning.
	{
		srcDartboardInRegion := a.dartboard.Words(bitInArena, uint(dartboardBitsPerRegion))
		copy(p.dartboardCopy[:], srcDartboardInRegion)
		clear(srcDartboardInRegion)
	}
	dartboardInRegion := bitmap.FromWords[heap.Words](p.dartboardCopy[:])

	// Iterate over each span in the region.
	addr := rid.toVAddr(gc.h.Heap)
	endAddr := addr.Plus(bytesPerRegion)
	for addr < endAddr {
		span := gc.h.FindSpan(addr) // XXX All spans are in the same arena, so we could skip the arena lookup
		if span == nil {
			incStat(&p.stats.scanStats.pagesSkipped, 1)
			addr = addr.Plus(heap.PageBytes)
			continue
		}
		incStat(&p.stats.scanStats.spans, 1)
		spanEnd := span.Start.Plus(heap.PageBytes.Mul(span.NPages))
		if span.Start < addr || spanEnd > endAddr {
			// This span is only partially in this region. We scan only the
			// objects that overlap with this region in this span, not the whole
			// overlapping span. As a result, we "round out" from the region to
			// object boundaries on either side. While it's fine for multiple
			// object scans to overlap, we can only safely access the dartboard
			// strictly within our region, so we *don't* expand dartboard access
			// to the whole object. That's fine, because if there are darts on
			// these objects outside our region, they'll get picked up by a scan
			// of the neighboring region.
			//
			// TODO: Figure out how to split large objects into oblets.
			// Presumably those oblets should be aligned with regions. I think
			// having a per-region "scanned" bit is unavoidable. Unlike grey
			// queuing, we may enter the object at any oblet. To avoid n^2
			// queuing behavior, I could have a rule like "if I'm scanning the
			// head, enqueue all of the other oblets; if I'm scanning layer,
			// enqueue the head".
			incStat(&p.stats.scanStats.partialSpans, 1)
			startAddr := max(span.Start, addr)
			endAddr := min(spanEnd, endAddr)
			incStat(&p.stats.scanStats.pagesScanned, endAddr.Minus(startAddr).Div(heap.PageBytes))
			if scanStats {
				spanRange := heap.Range{Start: startAddr, Len: endAddr.Minus(startAddr)}
				spanObjects := 0
				for i := range span.NObjects() {
					if span.ObjectRange(i).Overlaps(spanRange) {
						spanObjects++
					}
				}
				incStat(&p.stats.scanStats.objects, spanObjects)
			}

			startWordInRegion := (startAddr.Minus(0) % bytesPerRegion).Words()
			nWords := endAddr.Minus(startAddr).Words()
			if traceScanOnce {
				spanRange := heap.Range{Start: startAddr, Len: endAddr.Minus(startAddr)}
				log.Printf("  partial span %s => %s, region dartboard bits [%#x,%#x)", span.Range(), spanRange, startWordInRegion, startWordInRegion+nWords)
			}
			dartboard := dartboardInRegion.Words(startWordInRegion, uint(nWords))
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
		} else {
			// Full span
			incStat(&p.stats.scanStats.fullSpans, 1)
			incStat(&p.stats.scanStats.pagesScanned, span.NPages)
			incStat(&p.stats.scanStats.objects, span.NObjects())

			startWordInRegion := (span.Start.Minus(0) % bytesPerRegion).Words()
			nWords := heap.PageWords.Mul(span.NPages)
			dartboard := dartboardInRegion.Words(startWordInRegion, uint(nWords))
			if traceScanOnce {
				log.Printf("  full span %s, region dartboard bits [%#x,%#x)", span.Range(), startWordInRegion, startWordInRegion+nWords)
			}
			if span.SizeClass == nil {
				gc.scanLargeSpan(p, span, dartboard)
			} else {
				cd := gc.h.condensers[span.SizeClass.ID]
				gc.scanSpan(p, span, dartboard, cd)
			}
		}

		addr = spanEnd
	}

	if scanStats {
		gStats.RegionObjectMarkedDensity.Add(float64(p.stats.scanStats.objectsScanned) / float64(p.stats.scanStats.objects))
		gc.regionScanCount[rid]++
	}
}

func (gc *scanner) scanSpan(p *perP, span *heap.Span, dartboard []uint64, cd *condenser) {
	// TODO: Between 25% and 50% of the time, there are NO DARTS in a whole
	// span. We should summarize at the page level to skip whole spans. With 256
	// KiB regions and 8 KiB pages, that's 32 bits per region. So we could
	// easily use that instead of the existing region queued bits.

	// Get mark bits for this span.
	nObj := span.NObjects()
	marks := gc.marked.Words(span.FirstObject, uint(nObj))

	// Condense the dartboard into one bit per object. Use a fixed cap so this
	// is stack allocated.
	objDarts := make([]uint64, len(marks), (heap.MaxObjsPerSpan+63)/64)
	cd.do(dartboard, objDarts)

	if traceScanOnce {
		invivo.Invalidate()
		log.Printf("    darts        %016x", dartboard)
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

	if span.HeapBitsType == heap.HeapBitsNone {
		if traceScanOnce {
			log.Printf("    noscan span")
		}
		return
	}

	// Prepare for scanning.
	buf := gc.cNet.pBuffer(p.id)
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
		gc.scanObject(p, vs, bufPos, span, objIndex, base, objBytes)
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
func (gc *scanner) scanLargeSpan(p *perP, span *heap.Span, dartboard []uint64) {
	incStat(&p.stats.scanStats.largeSpans, 1)

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

	if span.HeapBitsType == heap.HeapBitsNone {
		// No need to scan.
		return
	}

	// Prepare for scanning.
	buf := gc.cNet.pBuffer(p.id)
	bufPos := &buf.n
	vs := buf.asVAddr()

	// Scan the object.
	gc.scanObject(p, vs, bufPos, span, 0, span.Start, heap.PageBytes.Mul(span.NPages))
}

func (gc *scanner) scanBuf(p *perP, buf *buffer) {
	defer p.stats.startScanBuf()()

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

func (gc *scanner) mark(p *perP, base heap.VAddr, span *heap.Span, objID heap.ObjectID) bool {
	if gc.marked.Has(objID) {
		return false
	}
	gc.marked.Add(objID)
	gc.scanObjectSlow(p, base, span.ObjectBytes())
	return true
}

func (gc *scanner) scanObjectSlow(p *perP, base heap.VAddr, length heap.Bytes) {
	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, length))
	incStat(&p.stats.scanStats.objectsScanned, 1)
	incStat(&p.stats.scanStats.wordsScanned, len(mem))
	for _, val := range mem {
		if val != 0 {
			// Individual enqueue is very slow
			gc.enqueue(p, val)
		}
	}
}

func (gc *scanner) scanObject(p *perP, buf []heap.VAddr, bufPos *int32, span *heap.Span, objIndex int, base heap.VAddr, length heap.Bytes) {
	// TODO: The way we do buffer management here is weird and error-prone

	// TODO: We can vectorize this in various ways.

	mem := heap.CastSlice[heap.VAddr](gc.h.Mem(base, length))
	incStat(&p.stats.scanStats.objectsScanned, 1)
	incStat(&p.stats.scanStats.wordsScanned, len(mem))

	// TODO: Lift this logic up to the scanSpan level as much as possible.
	var ptrMask []uint64
	var bitBase, bitWidth, bitStride heap.Words
	switch span.HeapBitsType {
	case heap.HeapBitsPacked:
		ptrMask = span.HeapBits
		bitBase = base.Minus(span.Start).Words()
		bitStride = length.Words()
		bitWidth = bitBase + bitStride
	case heap.HeapBitsHeader:
		// Skip header
		mem = mem[1:]
		fallthrough
	case heap.HeapBitsOOB:
		typeID := span.HeapBits[objIndex]
		typ := gc.h.Types[typeID]
		ptrMask = typ.PtrMask
		bitWidth = typ.PtrWords
		bitStride = typ.Size.Words()
	}

	pos := *bufPos
	for wordI, val := range mem {
		wordI := heap.Words(wordI)

		bitI := bitBase + wordI%bitStride
		if bitI >= bitWidth || ptrMask[bitI/64]&(1<<(bitI%64)) == 0 {
			continue
		}

		if int(pos) == len(buf) {
			*bufPos = pos
			gc.cNet.flush(p, 0, p.id)
			pos = *bufPos
		}

		buf[pos] = val
		pos++
		if traceEnqueue {
			invivo.Invalidate()
			log.Printf("enqueue %s", val)
		}
	}
	*bufPos = pos
}
