// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/gclab/bitmap"
	"cmd/gclab/heap"
	"fmt"
	"internal/trace"
	"log"
)

// Heaper is a Scanner that constructs the heap graph.
type Heaper struct {
	heap   *Heap
	gcInfo *heap.GCInfo

	queue bitmap.DynSet[heap.ObjectID]
	marks bitmap.DynSet[heap.ObjectID]

	AtEnd []func(h *Heap, gcinfo *heap.GCInfo)
}

func (h *Heaper) GCStart(gomaxprocs int) {
	h.heap = heap.NewHeap(sizeClasses)
	h.gcInfo = new(heap.GCInfo)
	h.gcInfo.Ps = gomaxprocs
}

func (h *Heaper) GCEnd() {
	// Self-check.
	for addr, objID := range h.heap.Objects() {
		if h.queue.Has(objID) && !h.marks.Has(objID) {
			log.Printf("object %s is grey", addr)
		}
		if h.marks.Has(objID) && !h.queue.Has(objID) {
			log.Printf("object %s scanned without queuing", addr)
		}
	}

	log.Printf("GC marked %d objects", h.marks.Set().Len())

	h.gcInfo.Marks = h.marks.Set()
	for _, f := range h.AtEnd {
		f(h.heap, h.gcInfo)
	}

	h.heap = nil
	h.gcInfo = nil
	h.queue.Drop()
	h.marks.Drop()
}

func (h *Heaper) NewSpan(base VAddr, sc *heap.SizeClass, nPages int, noScan bool) {
	if sc == nil {
		h.heap.NewSpanLarge(base, nPages, noScan)
	} else {
		h.heap.NewSpan(base, sc, noScan)
	}
}

func (h *Heaper) SpanInfo(base VAddr, allocBits []uint64, heapBitsType heapBitsType, heapBits []uint64) {
	s := h.heap.FindSpan(base)
	if heap.HeapBitsType(heapBitsType) != s.HeapBitsType {
		log.Fatalf("span %s has heap bits type %d, but heap bits in trace have type %d", base, s.HeapBitsType, heapBitsType)
	}
	s.AllocBits = bitmap.FromWords[uint64](allocBits)
	s.HeapBits = heapBits
}

func (h *Heaper) NewType(id uint64, size Bytes, ptrWords heap.Words, ptrData []uint64) {
	h.heap.NewType(id, size, ptrWords, ptrData)
}

func (h *Heaper) Scan(addr VAddr, typ scanType, offs []uint64, ptrs []VAddr, found []bool) {
	if typ == scanTypeObject {
		base, span, objID := h.heap.FindObject(addr)
		if span == nil {
			log.Fatalf("object %s not found", addr)
		} else if h.marks.Has(objID) {
			// I've disabled this for now because it also triggers whenever a
			// scan overflows the P buffer and starts a new block in the trace's
			// experimental batch. I could add a "beginning of scan" flag to
			// overcome this.
			if false && base == addr {
				// If base != addr, this is an oblet.
				//
				// TODO: Do we need to record both old and new values
				// scanned so we get the whole picture of the reachable
				// heap?
				log.Printf("double scan of object %s", addr)
			}
		} else if span.HeapBitsType == heap.HeapBitsNone {
			log.Fatalf("scan of noscan object %s", addr)
		}
		// We don't check h.queue for two reasons:
		//
		// - The trace can be out of order, so we may see the scan event before
		// we see the address enqueued.
		//
		// - There are a couple cases involving finalizers where we scan white
		// objects, so they aren't expected to be in the queue.
		h.marks.Add(objID)

		mem := heap.CastSlice[heap.VAddr](h.heap.Mem(base, span.ObjectBytes()))
		for i, off := range offs {
			mem[Bytes(off).Words()] = ptrs[i]
			if found[i] {
				h.greyObject(ptrs[i])
			}
		}
	} else {
		for i, ptr := range ptrs {
			h.gcInfo.Roots = append(h.gcInfo.Roots, ptr)
			if found[i] {
				h.greyObject(ptr)
			}
		}
	}
}

func (h *Heaper) greyObject(addr VAddr) {
	_, span, objID := h.heap.FindObject(addr)
	if objID == 0 {
		panic(fmt.Sprintf("no object found at %s", addr))
	}
	h.queue.Add(objID)
	if span.HeapBitsType == heap.HeapBitsNone {
		// Direct-to-black optimization
		h.marks.Add(objID)
	}
	return
}

func (h *Heaper) ScanWB(ev trace.Event, value VAddr, found bool) {
	// Only add it to the roots if it's an "interesting" write barrier.
	//
	// TODO: Should we just do all of them for edge queuing purposes? When do we
	// inject?
	if found {
		h.greyObject(value)
		h.gcInfo.WBRoots = append(h.gcInfo.WBRoots, value)
	}
}

func (h *Heaper) AllocBlack(ev trace.Event, value VAddr) {
	base, _, objID := h.heap.FindObject(value)
	h.queue.Add(objID) // To keep final checks happy
	h.marks.Add(objID)
	h.gcInfo.AllocBlack = append(h.gcInfo.AllocBlack, base)
}
