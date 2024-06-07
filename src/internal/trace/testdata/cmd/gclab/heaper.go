// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"internal/trace"
	"internal/trace/testdata/cmd/gclab/bitmap"
	"internal/trace/testdata/cmd/gclab/heap"
	"log"
)

// Heaper is a Scanner that constructs the heap graph.
type Heaper struct {
	heap   *Heap
	gcInfo *heap.GCInfo

	sizeClasses []heap.SizeClass

	queue bitmap.DynSet[heap.ObjectID]
	marks bitmap.DynSet[heap.ObjectID]

	AtEnd []func(h *Heap, gcinfo *heap.GCInfo)

	ps map[trace.ProcID]*heaperP
}

type heaperP struct {
	scanning bool
	scanB    VAddr
	scanTyp  scanType
	scanObj  *Object
}

func (h *Heaper) getP(id trace.ProcID) *heaperP {
	p, ok := h.ps[id]
	if !ok {
		p = new(heaperP)
		h.ps[id] = p
	}
	return p
}

func (h *Heaper) GCStart() {
	h.heap = heap.NewHeap(h.sizeClasses)
	h.gcInfo = new(heap.GCInfo)
	h.ps = make(map[trace.ProcID]*heaperP)
}

func (h *Heaper) GCEnd() {
	// Self-check.
	for addr, obj := range h.heap.Objects() {
		if h.queue.Has(obj.ID) && !h.marks.Has(obj.ID) {
			log.Printf("object %s is grey", addr)
		}
		if h.marks.Has(obj.ID) && !h.queue.Has(obj.ID) {
			log.Printf("object %s scanned without queuing", addr)
		}
	}

	log.Printf("GC marked %d objects", h.marks.Set().Len())

	h.gcInfo.Ps = len(h.ps)
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

func (h *Heaper) Scan(ev trace.Event, addr VAddr, typ scanType) {
	p := h.getP(ev.Proc())
	if p.scanning {
		log.Fatalf("P %d already scanning %#x", ev.Proc(), p.scanB)
	}
	p.scanning = true

	p.scanB = addr
	p.scanTyp = typ
	if typ == scanTypeObject {
		base, obj, span := h.heap.Find(addr)
		if obj == nil {
			log.Fatalf("object %s not found", addr)
		} else if obj.Words != nil {
			if base == addr {
				// If base != addr, this is an oblet.
				//
				// TODO: Do we need to record both old and new values
				// scanned so we get the whole picture of the reachable
				// heap?
				log.Printf("double scan of object %s", addr)
			}
		} else if span.NoScan {
			log.Fatalf("scan of noscan object %s", addr)
		} else {
			obj.Words = make([]VAddr, span.ObjectBytes().Div(heap.WordBytes))
		}
		// Note that we do scan white objects in a few cases related
		// to finalizers.
		h.marks.Add(obj.ID)
		p.scanObj = obj
	}
}

func (h *Heaper) ScanEnd(ev trace.Event) {
	p := h.getP(ev.Proc())
	if !p.scanning {
		log.Fatalf("P %d not scanning", ev.Proc())
	}
	p.scanning = false
}

func (h *Heaper) greyObject(addr VAddr) {
	_, obj, span := h.heap.Find(addr)
	h.queue.Add(obj.ID)
	if span.NoScan {
		// Direct-to-black optimization
		h.marks.Add(obj.ID)
	}
	return
}

func (h *Heaper) ScanPointer(ev trace.Event, value VAddr, found bool, offset Bytes) {
	p := h.getP(ev.Proc())
	if !p.scanning {
		log.Fatalf("P %d not scanning", ev.Proc())
	}
	if p.scanTyp == scanTypeObject {
		p.scanObj.Words[heap.Bytes(offset).Div(heap.WordBytes)] = value
	} else {
		h.gcInfo.Roots = append(h.gcInfo.Roots, value)
	}

	if found {
		h.greyObject(value)
	}
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
	base, obj, _ := h.heap.Find(value)
	h.queue.Add(obj.ID) // To keep final checks happy
	h.marks.Add(obj.ID)
	h.gcInfo.AllocBlack = append(h.gcInfo.AllocBlack, base)
}
