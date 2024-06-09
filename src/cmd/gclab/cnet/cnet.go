// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"fmt"
	"log"
	"sync"
	"unsafe"
)

// TODO: Does increasing dartboard density decrease the number of region visits?

// TODO: We spent a lot of time shuffling pointers around. Would it be better to
// just have two layers, where the per-P buffers on the top layer are sized
// proportionally to the heap to amortize the number of addresses written to
// each heap buffer on each flush, and the per-heap region buffers on the bottom
// layer are sized to achieve a target dartboard density? This would involve a
// deeper sort to flush a per-P buffer, but we could do that in a
// cache-conscious way, and potentially more contention on the heap buffers, but
// maybe Ps or groups of Ps reserve blocks, or we have more than one buffer per
// heap region.

const (
	// radixBase is the base used for the radix sort. Making this a constant
	// simplifies a bunch of code and makes it compile much better, but does make it
	// a bit harder to experiment with different network configurations.
	radixBase = 1 << radixBits

	// radixBits is the log_2 of radixBase.
	radixBits = 4
)

// CNet describes a concentrator network of work buffers.
//
// At the top of the network is one buffer per P, and at the bottom of the
// network is one buffer per "queue region" of the address space. The network in
// between acts as 1) an adaptor from the network span at the top to the network
// span at the bottom, 2) a distributed radix sort that shuffles unsorted
// pointers at the top of the network to sorted pointers at the bottom, and 3) a
// contention-limiting mechanism as parallel threads move work down the network.
type CNet struct {
	heap heapExtra

	tmpBufs sync.Pool // of *buffer

	layers []layer

	// regionQueue is the queue of dartboard regions that have set bits to
	// process.
	//
	// TODO: This will have to be concurrent. Or maybe the region bitmap is
	// simply compact enough that we could store it contiguously and scan the
	// whole thing?
	regionQueue []regionID
}

func NewCNet(config CNetConfig, h heapExtra, Ps int) *CNet {
	heapSize := heap.ArenaBytes.Mul(len(h.Arenas))
	layers := config.makeLayers(Ps, heapSize)

	// The top layer must be the number of Ps.
	if len(layers[0].buffers) != Ps {
		panic(fmt.Sprintf("top layer has %d buffers, want %d (== Ps)", len(layers[0].buffers), Ps))
	}
	// TODO: Check that the bottom layer has exactly one buffer per queue region.
	// TODO: Each bottom buffer must cover a span that fits in an arena.
	// TODO: Each bottom buffer must cover 1 or more exact bitmap regions.

	// Find the largest buffer in the network. We'll size the temporary buffers
	// to match.
	var maxBufSize int
	for _, l := range layers {
		for _, buf := range l.buffers {
			maxBufSize = max(maxBufSize, len(buf.data))
		}
	}

	return &CNet{
		heap: h,
		tmpBufs: sync.Pool{New: func() any {
			return &buffer{
				data: make([]byte, maxBufSize),
			}
		}},
		layers: layers,
	}
}

type CNetConfig interface {
	makeLayers(Ps int, heapSize heap.Bytes) []layer
}

type layer struct {
	buffers []buffer

	// shift is the bit shift of the LAddr32/LAddr64 digit this layer is sorted
	// by. Note that because LAddr32/64 are in terms of words, not bytes, so is
	// this shift.
	shift uint

	// topo gives the topology of the network at this layer. topo[i] corresponds
	// to buffer i. buffers[i] should be flushed into buffers topo[i] to
	// topo[i]+fanOut in the next layer.
	topo []int
	// fanOut is this layer's local fan out. This may be less than the general
	// fan out if that would exceed the heap size.
	fanOut int

	// heapSpan is the number of bytes of heap covered by each buffer. Only for
	// debugging.
	heapSpan heap.Bytes
}

// A buffer stores linear addresses as either []uint64 or []uint32. In the case
// of []uint32, the upper 32 bits of all of the linear addresses are the same,
// so a buffer can cover a given 32 GiB frame of the address space. (We don't
// bother with uint16s because that would only cover 512 KiB of the address
// space and we never go down that far).
//
// TODO: Mark this "no copy".
type buffer struct {
	n   int32
	typ bufferType
	// start is the address of the region of the heap covered by this buffer.
	// For uint32 buffers, this provides the shared upper bits of all addresses.
	start heap.LAddr
	// data is the raw data backing this buffer. It gets cast to different types
	// based on typ.
	data []byte // TODO: This would be an array with a const length.
}

type bufferType byte

const (
	bufferInvalid bufferType = iota
	bufferVAddr
	bufferLAddr64
	bufferLAddr32
)

func (b *buffer) asLAddr32() []LAddr32 {
	if b.typ != bufferLAddr32 {
		b.wrongType(bufferLAddr32)
	}
	d := unsafe.SliceData(b.data)
	return unsafe.Slice((*LAddr32)(unsafe.Pointer(d)), len(b.data)/4)
}

func (b *buffer) asLAddr64() []LAddr64 {
	if b.typ != bufferLAddr64 {
		b.wrongType(bufferLAddr64)
	}
	d := unsafe.SliceData(b.data)
	return unsafe.Slice((*LAddr64)(unsafe.Pointer(d)), len(b.data)/8)
}

func (b *buffer) asVAddr() []heap.VAddr {
	if b.typ != bufferVAddr {
		b.wrongType(bufferVAddr)
	}
	d := unsafe.SliceData(b.data)
	return unsafe.Slice((*heap.VAddr)(unsafe.Pointer(d)), len(b.data)/8)
}

func (b *buffer) wrongType(want bufferType) {
	panic(fmt.Sprintf("buffer is %s, not %s", b.typ, want))
}

func (t bufferType) String() string {
	switch t {
	case bufferInvalid:
		return "invalid"
	case bufferVAddr:
		return "VAddr"
	case bufferLAddr32:
		return "LAddr32"
	case bufferLAddr64:
		return "LAddr64"
	}
	return fmt.Sprintf("bufferType(%d)", t)
}

func (c *CNet) pBuffer(p int) *buffer {
	return &c.layers[0].buffers[p]
}

func (b *buffer) mapToLAddr64(h *heap.Heap) {
	vaddrs := b.asVAddr()[:b.n]
	b.typ = bufferLAddr64
	b.n = int32(mapVAddrToLAddr64(h, vaddrs, b.asLAddr64()))
}

// mapVAddrToLAddr converts a slice of VAddrs to LAddrs and returns the number
// of successfully mapped addresses. src and dst may alias.
func mapVAddrToLAddr64(h *heap.Heap, src []heap.VAddr, dst []LAddr64) int {
	// TODO: Sometimes this can dramatically compact the buffer. If this is
	// common, maybe we want to keep filling the P buffer before flushing down.
	o := 0
	for _, addr := range src {
		lAddr := h.VAddrToLAddr(addr)
		lAddr64 := LAddr64(lAddr) / LAddr64(heap.WordBytes)
		dst[o] = lAddr64
		o += bool2int(lAddr != heap.NoLAddr)
		if traceFlushAddrs {
			if lAddr == heap.NoLAddr {
				log.Printf("%s -> no LAddr", addr)
			} else {
				log.Printf("%s -> %s", addr, lAddr64)
			}
		}
	}
	if traceFlush && len(src) != o {
		log.Printf("mapVAddrToLAddr64 %d reduced to %d", len(src), o)
	}
	return o
}

// getScanBuf returns a non-empty buffer in the network, with preference for
// less dense buffers.
func (c *CNet) getScanBuf() *buffer {
	for i := range c.layers {
		layer := &c.layers[i]
		for j := range layer.buffers {
			buf := &layer.buffers[j]
			if buf.n != 0 {
				return buf
			}
		}
	}
	return nil
}
