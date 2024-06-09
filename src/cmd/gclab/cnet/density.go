// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"log"
	"math/bits"
	"slices"
)

type DensityNetworkConfig struct {
	// DartboardInvDensity is the inverse of the expected density of bits set in
	// the dartboard when flushing a full buffer to the dartboard. That is, we
	// expect 1 bit per DartboardInvDensity bits to be set, modulo duplicate
	// pointers.
	//
	// The dartboard density directly correlates with the efficiency of both
	// flushing to the dartboard and of dense scans.
	//
	// Almost all work buffer memory overhead is from the bottom layer, and the
	// size of that is directly proportional to the density we want to achieve
	// (and the heap size), so we can build a network that dials in a specific
	// overhead by setting the desired density.
	//
	// The amount of memory in work buffers in turn directly correlates with how
	// much work can be held in buffers (requiring sparse scans) before making
	// it to the dartboard (for dense scans).
	//
	// Thus, lower values of this (higher density) increase the efficiency of
	// flushes and dense scans, but increase the amount of work held in work
	// buffers.
	DartboardInvDensity int

	// QueueRegion is the bytes of heap covered by a single buffer at the bottom
	// of the network. This must be a multiple of the dartboard scan region.
	//
	// The size of the mark region (and the heap size) determines the number of
	// buffers at the bottom level. Thus, lower values of this reduce the
	// overhead of tracking buffers (by making each buffer larger), but increase
	// contention.
	QueueRegion heap.Bytes

	// FanOut is the number of destination buffers that each buffer flushes to
	// in the next layer.
	//
	// This is also the degree of the radix sort performed by the network, and
	// thus must be a power of two.
	//
	// Higher values of this reduce the height of the network, but increase the
	// cost of flushing from one layer to the next.
	//
	// TODO: Depending on how we flush buffers, the effect of this may be fairly
	// negligible up to a fairly high value. Like, if we first sort the buffer
	// and then copy out blocks, we might be mostly limited by how many counts
	// we can keep at once.
	FanOut int

	// FanIn is the number of source buffers that can be flushed to a buffer in
	// the next layer. It controls the degree of thread contention during
	// flushing.
	//
	// Higher values of this reduce the height of the network, but increase the
	// contention during flushing.
	FanIn int

	// Overall network height is max(log_fanOut(heapSize / C), log_fanIn(Ps)).
}

var DefaultDensityNetworkConfig = DensityNetworkConfig{
	// 1 bit per 32 bytes. With a 64 byte cache line, this is 2 bits per cache
	// line on average, which strikes a pretty good balance.
	DartboardInvDensity: 32 * 8,

	QueueRegion: 32 * heap.MiB,

	FanOut: 16,

	FanIn: 4,
}

func (config DensityNetworkConfig) makeLayers(Ps int, heapSize heap.Bytes) []layer {
	// TODO: This will have to adapt as the heap grows during a GC. That's rare
	// enough and messy enough that we may just want to STW to swap in the new
	// network. We can do the vast majority of the work before STW, and then
	// just STW as a global barrier to swap in the new topology.

	if _, ok := log2(int64(config.FanOut)); !ok {
		panic("fan out must be a power of 2")
	}
	if _, ok := log2(int64(config.QueueRegion)); !ok {
		panic("queue region size must be a power of 2")
	}

	if config.FanOut != radixBase {
		panic("fan out and radix base must be ==")
	}

	// How big do the bottom buffers need to be to achieve the given density?
	// We'll make all buffers this same size, since doing otherwise introduces
	// complexity without obvious value.
	bottomAddrBytes := addrBytes(config.QueueRegion) // Bytes to represent an address
	bottomEntries := config.QueueRegion.Div(heap.WordBytes.Mul(config.DartboardInvDensity))
	bufSize := bottomAddrBytes.Mul(bottomEntries)

	log.Printf("buffer size: %s", bufSize)

	// Go from the top down to figure out P fan-in.
	pSpans := make([]int, 0, 16) // Indexed by layer
	pSpans = append(pSpans, 1)
	for last(pSpans) < Ps {
		pSpans = append(pSpans, last(pSpans)*config.FanIn)
	}

	// Go from the bottom up to figure out the heap segmentation.
	heapSpans := make([]heap.Bytes, 0, 16) // Indexed by layer
	heapSpans = append(heapSpans, config.QueueRegion)
	for last(heapSpans) < heapSize {
		heapSpans = append(heapSpans, last(heapSpans).Mul(config.FanOut))
	}

	// Align the two factors.
	for len(heapSpans) < len(pSpans) {
		heapSpans = append(heapSpans, last(heapSpans))
	}
	for len(pSpans) < len(heapSpans) {
		pSpans = append(pSpans, last(pSpans))
	}
	slices.Reverse(heapSpans)

	// bufIndex returns the index of the buffer in the given layer for a given p
	// and queue region index.
	bufIndex := func(layer int, p int, heapSeg heap.LAddr) int {
		pIdx := p / pSpans[layer]
		heapIdx := heapSeg.FloorDiv(heapSpans[layer])
		heapSegs := heapSize.CeilDiv(heapSpans[layer])
		return pIdx*heapSegs + heapIdx
	}

	// Build the layers.
	layers := make([]layer, len(heapSpans))
	for layer, heapSpan := range heapSpans {
		pSpan := pSpans[layer]

		var buffers []buffer
		var topo []int
		for p := 0; p < Ps; p += pSpan {
			for lAddr := heap.LAddr(0); lAddr < heap.LAddr(heapSize); lAddr = lAddr.Plus(heapSpan) {
				var buf buffer
				if layer == 0 {
					// The top layer is always VAddrs since this is what Ps
					// write to directly.
					buf.typ = bufferVAddr
				} else if addrBytes(heapSpan) == 8 {
					buf.typ = bufferLAddr64
				} else if addrBytes(heapSpan) == 4 {
					// Other layers are just big enough to hold addresses within
					// their region of the heap. Often 4 bytes is plenty.
					buf.typ = bufferLAddr32
				} else {
					panic("bad addrBytes")
				}
				buf.data = make([]byte, bufSize)
				buf.start = lAddr
				buffers = append(buffers, buf)
				if layer+1 < len(heapSpans) {
					topo = append(topo, bufIndex(layer+1, p, lAddr))
				}
			}
		}
		// Make sure we don't slice past the len.
		buffers = buffers[:len(buffers):len(buffers)]

		layers[layer].buffers = buffers
		shift, ok := log2(int64(heapSpan.Words()))
		if !ok {
			panic("heapSpan not a power of 2")
		}
		layers[layer].shift = uint(shift)
		layers[layer].topo = topo
		if layer > 0 {
			layers[layer-1].fanOut = min(heapSize.CeilDiv(heapSpan), config.FanOut)
		}
		layers[layer].heapSpan = heapSpan
	}

	return layers
}

func log2(x int64) (int, bool) {
	if x <= 0 || (x&(x-1)) != 0 {
		return 0, false
	}
	return bits.TrailingZeros64(uint64(x)), true
}

func last[T any](x []T) T {
	return x[len(x)-1]
}

// addrBytes returns the number of bytes necessary to represent addresses within
// a region of the given size. This is 1, 2, 4, or 8.
func addrBytes(regionSize heap.Bytes) heap.Bytes {
	words := regionSize.Words()
	switch {
	case words <= 1<<8:
		return 1
	case words <= 1<<16:
		return 2
	case words <= 1<<32:
		return 4
	default:
		return 8
	}
}
