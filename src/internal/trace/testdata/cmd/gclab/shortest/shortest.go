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
	var q []*heap.Object
	var queued bitmap.DynSet[heap.ObjectID]

	enqueue := func(addr heap.VAddr) {
		_, o, _ := h.Find(addr)
		if o != nil && !queued.Has(o.ID) {
			queued.Add(o.ID)
			q = append(q, o)
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
		obj := q[0]
		q = q[1:]
		for _, word := range obj.Words {
			enqueue(word)
		}
	}

	log.Printf("shortest marked %d objects", queued.Set().Len())
	gcinfo.CompareMarks(h, queued.Set())
}
