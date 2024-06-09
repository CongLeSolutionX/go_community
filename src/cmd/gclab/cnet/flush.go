// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"cmd/gclab/invivo"
	"log"
	"slices"
	"unsafe"
)

const xxxDebug = false

const dedupAddrs = false

// TODO: Have a debug mode that makes slow paths run more frequently. For
// example, make flushes happen more frequently and unpredictably (especially to
// the dartboard).

func (c *CNet) flush(p *perP, layer int, bufI int) {
	// TODO: There are a lot of ways to do this and this may not be the most
	// efficient. E.g., we could copy directly out to the destination buffers
	// and avoid the scratch buffer, though I didn't do that because it seemed
	// like it would require hugely more bounds checks and flush checks.
	//
	// TODO: At the cost of some additional complexity, we could do the whole
	// flush cascade with only one temp buffer. We would do the counts first,
	// then check if that would overflow any buffers and flush those buffers,
	// then we would use those counts to sort into the temporary buffer and then
	// copy out to the destination buffers (which we would now know have enough
	// space). This might be worse for concurrency, though, since we'd have to
	// keep the whole chain of buffers locked.

	var bench invivo.Run
	if layer == 0 {
		bench = benchmarkFlush.Start()
	}
	benchLayer := p.stats.pushBenchmark(benchmarkFlushLayer[layer])
	var statNAddrs int
	defer func() {
		if layer == 0 {
			bench.StopTimer()
		}
		benchLayer.StopTimer()
		if layer == 0 {
			metricAddrsPerSec.Set(bench, float64(statNAddrs), bench.Elapsed().Seconds())
			metricAddrsPerOp.Set(bench, float64(statNAddrs))
			bench.Done()
		}
		metricAddrsPerSec.Set(benchLayer, float64(statNAddrs), benchLayer.Elapsed().Seconds())
		metricAddrsPerOp.Set(benchLayer, float64(statNAddrs))
		p.stats.popBenchmark()
	}()

	if traceFlush {
		invivo.Invalidate()
	}

	incStat(&p.stats.scanStats.flushes, 1)

	l := &c.layers[layer]
	src := &l.buffers[bufI]
	statNAddrs = int(src.n)

	if layer == 0 {
		// Map virtual addresses to linear addresses. We do this in place.
		//
		// TODO: It might be worth having a direct VAddr -> LAddr32 path.
		//
		// TODO: It might make sense to have a small VAddr buffer that we flush
		// to the per-P LAddr{64,32} buffer. Then the whole concentrator network
		// is in terms of LAddrs and we don't have this weird in-place rewrite.
		src.mapToLAddr64(c.heap.Heap)
		defer func() { src.typ = bufferVAddr }()
		// Now we can treat it as an LAddr buffer.
	}

	if dedupAddrs {
		// TODO: Ideally we would keep using the buffer if we free up a lot of
		// space. We can't do that on layer 0 because we just transformed it.
		buf := src
		if buf.typ == bufferLAddr32 {
			slices.Sort(buf.asLAddr32()[:buf.n])
			prev := buf.n
			buf.n = int32(nub(buf.asLAddr32()[:buf.n]))
			if traceFlush {
				log.Printf("dedup reduced %d -> %d", prev, buf.n)
			}
			// if buf.n < int32(len(buf.asLAddr32())-32) {
			// 	return
			// }
		} else {
			slices.Sort(buf.asLAddr64()[:buf.n])
			prev := buf.n
			buf.n = int32(nub(buf.asLAddr64()[:buf.n]))
			if traceFlush {
				log.Printf("dedup reduced %d -> %d", prev, buf.n)
			}
			// if buf.n < int32(len(buf.asLAddr64())-32) {
			// 	return
			// }
		}
	}

	if layer == len(c.layers)-1 {
		// Final layer. Flush to the dartboard. This takes advantage of the fact
		// that the bottom layer buffers always cover a region contained within
		// a single arena.
		aID := src.start.ArenaID()
		arena := c.heap.arenas[aID]
		if traceFlush || traceFlushAddrs {
			log.Printf("flush %d/%d (start %s), %d addrs -> dartboard arena %s ID %d", layer, bufI, src.start, src.n, arena.Range(), aID)
		}
		var regionBitmap arenaRegionBitmap
		if src.typ == bufferLAddr32 {
			if xxxDebug { // XXX
				slices.Sort(src.asLAddr32()[:src.n])
				for _, laddr32 := range src.asLAddr32()[:src.n] {
					lAddr := src.start | heap.LAddr(laddr32*LAddr32(heap.WordBytes))
					vAddr := c.heap.LAddrToVAddr(lAddr)
					if traceFlush && traceEnqueue {
						log.Printf("  flush bit %#08x => %s, VAddr %s", laddr32.ArenaWord(), lAddr, vAddr)
					}
					c.heap.FindObject(vAddr)
				}
			}
			copyToDartboard(src.asLAddr32()[:src.n], arena.dartboard, &regionBitmap, l.heapSpan)
		} else {
			copyToDartboard(src.asLAddr64()[:src.n], arena.dartboard, &regionBitmap, l.heapSpan)
		}
		src.n = 0
		c.enqueueRegions(arena, &regionBitmap)
		return
	}

	ln := &c.layers[layer+1]
	dstI := l.topo[bufI]
	dsts := ln.buffers[dstI:]
	// At the end of the heap, we may have fewer than l.fanOut buffers, but
	// should never see an address that would need to go past the last buffer.
	dsts = dsts[:min(len(dsts), l.fanOut)]
	shift := ln.shift

	if traceFlush || traceFlushAddrs {
		log.Printf("flush %d/%d, %d addrs -> %d/[%d,%d)", layer, bufI, src.n, layer+1, dstI, dstI+l.fanOut)
	}

	// Bucket sort the digit into the temporary buffer, then block copy into the
	// destination buffers.
	var counts [radixBase]uint16
	tmpBuf := c.tmpBufs.Get().(*buffer)
	tmpBuf.typ = src.typ
	full := func(i int) {
		c.flush(p, layer+1, dstI+i)
	}
	if src.typ == bufferLAddr32 {
		gStats.LAddr32s.Add(int(src.n))

		tmp := tmpBuf.asLAddr32()
		if useAVX {
			bucketSort32AVX(src.asLAddr32()[:src.n], tmp, &counts, shift)
		} else {
			bucketSort(src.asLAddr32()[:src.n], tmp, &counts, shift)
		}
		copy32To32(tmp[:src.n], dsts, &counts, full)
		if xxxDebug {
			invivo.Invalidate()
			// Check the destination buffers.
			// XXX Should check in full, too.
			for _, dst := range dsts {
				lo := dst.start
				hi := dst.start.Plus(ln.heapSpan)
				for _, lAddr32 := range dst.asLAddr32()[:dst.n] {
					lAddr := dst.start | heap.LAddr(lAddr32*LAddr32(heap.WordBytes))
					if !(lo <= lAddr && lAddr < hi) {
						log.Printf("address %s out of range [%s,%s)", lAddr, lo, hi)
						log.Printf("shift %d", shift)
						log.Println(tmp[:src.n])
						log.Println(counts)
						panic("out of range")
					}
				}
			}
		}
	} else {
		gStats.LAddr64s.Add(int(src.n))
		tmp := tmpBuf.asLAddr64()
		bucketSort(src.asLAddr64()[:src.n], tmp, &counts, shift)
		if dsts[0].typ == bufferLAddr32 {
			copy64To32(tmp[:src.n], dsts, &counts, full)
		} else {
			copy64To64(tmp[:src.n], dsts, &counts, full)
		}
	}
	c.tmpBufs.Put(tmpBuf)

	// Done with the source buffer. Reset it.
	src.n = 0
}

func bool2int(x bool) int {
	// This particular pattern gets optimized by gc.
	var b int
	if x {
		b = 1
	}
	return b
}

// bucketSort bucket sorts data from src into dst by the digit at the given bit
// shift. It stores the count if each digit value in *counts. The caller must
// zero counts before calling.
func bucketSort[T ~uint32 | ~uint64](src, dst []T, counts *[radixBase]uint16, shift uint) {
	// TODO: The main cost of this function is dealing with the super
	// unpredictable writes and reads to counts and offs (the writes to dst
	// aren't as bad, presumably because nothing depends on them.) For a radix
	// base of 16, I could use a 16x16, 256 bit AVX register to keep counts and
	// offsets. That could handle buffers up to 64K entries (258 KiB for uint32,
	// 512 KiB for uint64). E.g., see stackoverflow.com/questions/61122144

	// Count the digit.
	//
	// TODO: Use separate count32* functions.
	mask := T(len(counts) - 1)
	for _, val := range src {
		counts[(val>>shift)&mask]++
	}

	// Turn the counts into offsets.
	var offs [len(counts)]uint16
	var pos uint16
	for i, count := range counts {
		offs[i] = pos
		pos += count
	}

	// Sort into output buffer.
	//
	// TODO: It may be just as well to write directly into the next layer,
	// though it does mean we'd have to interleave a lot of overflow checking.
	//
	// TODO: Prefetch?
	for _, val := range src {
		digit := (val >> shift) & mask
		dst[offs[digit]] = val
		offs[digit]++
	}
}

func count32Go(src []LAddr32, shift uint, counts *[radixBase]uint16) {
	mask := LAddr32(len(counts) - 1)
	for _, val := range src {
		counts[(val>>shift)&mask]++
	}
}

func bucketSort32x16(src, dst []LAddr32, counts *[16]int, shift uint) {
	if shift >= 32 {
		panic("bad shift")
	}
	var counts0, counts1 [16]int
	const mask = uint64(len(counts) - 1)
	const mask32 = LAddr32(len(counts) - 1)
	src64 := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(src))), len(src)/2)
	shift2 := shift + 32
	if shift2 >= 64 {
		panic("bad shift")
	}
	for _, vv := range src64 {
		counts0[(vv>>shift)&mask]++
		counts1[(vv>>shift2)&mask]++
	}
	if len(src)%2 == 1 {
		val0 := src[len(src)-1]
		counts0[(val0>>shift)&mask32]++
	}
	for i := range counts {
		counts[i] = counts0[i] + counts1[i]
	}

	// Turn the counts into offsets.
	var offs [len(counts)]int
	pos := 0
	for i, count := range counts {
		offs[i] = pos
		pos += count
	}

	// Sort into output buffer.
	//
	// TODO: Prefetch?
	for _, val := range src {
		digit := (val >> shift) & mask32
		dst[offs[digit]] = val
		offs[digit]++
	}
}

func count32AVX2(src []LAddr32, shift uint, counts *[16]uint16) {
	// The assembly core works in multiples of 32 bytes, or 8 entries.
	const blockSize = 32 / 4
	nBlock := len(src) / blockSize * blockSize
	count32AVX2Asm(src[:nBlock], shift, counts)
	// Handle the tail.
	if nBlock > 0 {
		const mask = LAddr32(len(counts) - 1)
		for _, val := range src[nBlock:] {
			counts[(val>>shift)&mask]++
		}
	}
}

func count32AVX2Asm(src []LAddr32, shift uint, counts *[16]uint16)

func bucketSort32AVX(src, dst []LAddr32, counts *[16]uint16, shift uint) {
	// Count each digit.
	mask := LAddr32(len(counts) - 1)
	count32AVX2(src, shift, counts)

	// Turn the counts into offsets.
	var offs [len(counts)]uint16
	var pos uint16
	for i, count := range counts {
		offs[i] = pos
		pos += count
	}

	// Sort into output buffer.
	//
	// TODO: There are a ton of bounds checks in this. Consider overriding those
	// or perhaps doing this whole thing in assembly.
	//
	// TODO: Prefetch?
	for _, val := range src {
		digit := (val >> shift) & mask
		dst[offs[digit]] = val
		offs[digit]++
	}
}

func copy32To32(src []LAddr32, dsts []buffer, counts *[radixBase]uint16, full func(dstI int)) {
	// TODO: Consider rounding chunks up to cache line boundaries to reduce
	// false sharing.

	// Block copy into target buffers.
	for i := range dsts {
		buf := &dsts[i]
		srcN := counts[i]
	more:
		dst := buf.asLAddr32()
		n := copy(dst[buf.n:], src[:srcN])
		buf.n += int32(n)
		src = src[n:]
		srcN -= uint16(n)
		if srcN > 0 {
			// TODO: This will be rare. It might be better to trigger a slow
			// path after the loop to avoid the conditional in each iteration.
			full(i)
			goto more
		}
	}
	if len(src) != 0 {
		panic("source buffer not fully drained")
	}
}

func copy64To32(src []LAddr64, dsts []buffer, counts *[radixBase]uint16, full func(dstI int)) {
	for i := range dsts {
		buf := &dsts[i]
		srcN := int(counts[i])
	more:
		dst := buf.asLAddr32()[buf.n:]
		n := min(srcN, len(dst))
		for i, srcVal := range src[:n] {
			// TODO: I bet there are a ton of bounds checks here and that we can
			// eliminate them, but we might have to teach prove about min.
			dst[i] = LAddr32(srcVal)
			if traceFlushAddrs {
				log.Printf("  %s -> %s", srcVal, LAddr32(srcVal))
			}
		}
		buf.n += int32(n)
		src = src[n:]
		srcN -= n
		if srcN > 0 {
			if traceFlushAddrs {
				log.Printf("  buffer full")
			}
			full(i)
			goto more
		}
	}
	if len(src) != 0 {
		panic("source buffer not fully drained")
	}
}

func copy64To64(src []LAddr64, dsts []buffer, counts *[radixBase]uint16, full func(dstI int)) {
	panic("not implemented")
}

func nub[T comparable](s []T) int {
	if len(s) == 0 {
		return 0
	}
	o := 0
	for i := 1; i < len(s); i++ {
		if s[i] != s[o] {
			o++
			s[o] = s[i]
		}
	}
	return o + 1
}
