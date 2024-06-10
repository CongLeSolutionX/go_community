// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shortest

import (
	"cmd/gclab/bitmap"
	"cmd/gclab/heap"
	"log"
)

func Scanner(h *heap.Heap, gcinfo *heap.GCInfo) {
	var q []heap.VAddr
	var queued bitmap.DynSet[heap.ObjectID]

	enqueue := func(addr heap.VAddr) {
		base, _, objID := h.FindObject(addr)
		if objID != 0 && !queued.Has(objID) {
			queued.Add(objID)
			q = append(q, base)
		}
	}

	for _, addr := range gcinfo.AllocBlack {
		enqueue(addr)
	}
	q = nil // Alloc-black objects skip gray.
	for _, addr := range gcinfo.Roots {
		enqueue(addr)
	}
	for _, addr := range gcinfo.WBRoots {
		enqueue(addr)
	}

	for len(q) > 0 {
		base := q[0]
		q = q[1:]
		span := h.FindSpan(base)
		objLen := span.ObjectBytes()
		mem := heap.CastSlice[heap.VAddr](h.Mem(base, objLen))
		if false {
			for _, word := range mem {
				if word != 0 {
					enqueue(word)
				}
			}
		} else {
			switch span.HeapBitsType {
			case heap.HeapBitsNone:
			case heap.HeapBitsPacked:
				bitBase := base.Minus(span.Start).Words()
				for i, word := range mem {
					bitI := bitBase + heap.Words(i)
					if span.HeapBits[bitI/64]&(1<<(bitI%64)) != 0 {
						enqueue(word)
					} else if word != 0 {
						log.Printf("object %v word %v is non-zero but not marked as a pointer", base, i)
					}
				}
			case heap.HeapBitsHeader, heap.HeapBitsOOB:
				if span.HeapBitsType == heap.HeapBitsHeader {
					// Skip header
					mem = mem[1:]
				}
				_, _, objID := h.FindObject(base)
				typeID := span.HeapBits[objID-span.FirstObject]
				typ := h.Types[typeID]
				for i, word := range mem {
					bitI := heap.Words(i) % typ.Size.Words()
					if bitI.Div(64) < len(typ.PtrMask) && typ.PtrMask[bitI/64]&(1<<(bitI%64)) != 0 {
						enqueue(word)
					} else if word != 0 {
						log.Printf("object %v word %v is non-zero but not marked as a pointer", base, i)
					}
				}
			}
		}
	}

	log.Printf("shortest marked %d objects", queued.Set().Len())
	gcinfo.CompareMarks(h, queued.Set())
}
