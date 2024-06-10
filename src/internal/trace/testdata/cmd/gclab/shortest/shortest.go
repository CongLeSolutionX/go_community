// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shortest

import (
	"internal/trace/testdata/cmd/gclab/bitmap"
	"internal/trace/testdata/cmd/gclab/heap"
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
		objLen := h.FindSpan(base).ObjectBytes()
		mem := heap.CastSlice[heap.VAddr](h.Mem(base, objLen))
		for _, word := range mem {
			if word != 0 {
				enqueue(word)
			}
		}
	}

	log.Printf("shortest marked %d objects", queued.Set().Len())
	gcinfo.CompareMarks(h, queued.Set())
}
