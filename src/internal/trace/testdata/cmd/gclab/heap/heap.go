// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package heap

import (
	"fmt"
	"iter"
	"log"
	"slices"
)

const (
	WordBytes  Bytes = 8
	PageBytes  Bytes = 8 << 10
	ArenaBytes Bytes = 64 << 20 // TODO Put in setup event?

	ArenaWords Words = Words(ArenaBytes / WordBytes)
)

type SizeClass struct {
	ObjectBytes Bytes
	SpanPages   int
}

type Heap struct {
	SizeClasses []SizeClass

	// arenaMap stores a mapping from VAddr.ArenaIndex to ArenaID+1. The +1
	// means that we can zero fill this array and subtracting 1 from a 0 entry
	// will result in NoArena.
	arenaMap []uint32 // VAddr.ArenaIndex -> ArenaID+1
	Arenas   []*Arena // ArenaID -> *Arena

	nextObject ObjectID
}

type ArenaID uint32

const NoArena = ^ArenaID(0)

func (a ArenaID) String() string {
	if a == NoArena {
		return "NoArena"
	}
	return fmt.Sprintf("ArenaID(%d)", a)
}

type Arena struct {
	ID    ArenaID
	Start VAddr

	SpanMap []*Span // Page index -> *Span
}

func (a Arena) Range() Range {
	return Range{a.Start, ArenaBytes}
}

func (a Arena) String() string {
	return fmt.Sprintf("Arena[%d]/%s", uint32(a.ID), Range{a.Start, ArenaBytes})
}

type Span struct {
	Start     VAddr
	End       VAddr
	NPages    int
	SizeClass *SizeClass
	NoScan    bool

	Objects []Object // By object index
}

func (s *Span) Range() Range {
	return Range{s.Start, s.End.Minus(s.Start)}
}

// ObjectBytes returns the byte size of objects in this Span.
func (s *Span) ObjectBytes() Bytes {
	if s.SizeClass != nil {
		return s.SizeClass.ObjectBytes
	}
	return PageBytes.Mul(s.NPages)
}

func (s *Span) NObjects() int {
	if s.SizeClass != nil {
		return PageBytes.Mul(s.SizeClass.SpanPages).Div(s.SizeClass.ObjectBytes)
	}
	return 1
}

// ObjectRange returns the Range of the i'th object in this Span.
func (s *Span) ObjectRange(i int) Range {
	if i < 0 || i >= s.NObjects() {
		panic(fmt.Sprintf("object %d out of range [0,%d)", i, s.NObjects()))
	}
	size := s.ObjectBytes()
	return Range{s.Start.Plus(size.Mul(i)), size}
}

type ObjectID uint64

type Object struct {
	ID    ObjectID
	Words []VAddr
}

func NewHeap(sc []SizeClass) *Heap {
	var h Heap
	h.SizeClasses = sc
	return &h
}

func (h *Heap) ArenaIndexToID(idx int) ArenaID {
	if idx >= len(h.arenaMap) {
		return NoArena
	}
	return ArenaID(h.arenaMap[idx] - 1)
}

func (h *Heap) EnsureArena(p VAddr) *Arena {
	// Ensure the arena map is large enough. In the real runtime this is
	// preallocated.
	idx := p.ArenaIndex()
	if idx >= len(h.arenaMap) {
		m := slices.Grow(h.arenaMap, idx-len(h.arenaMap)+1)
		h.arenaMap = m[:cap(m)]
	}

	aID := h.ArenaIndexToID(idx)
	if aID != NoArena {
		return h.Arenas[aID]
	}
	// Create a new arena.
	aID = ArenaID(len(h.Arenas))
	a := &Arena{
		ID:      aID,
		Start:   VAddr(ArenaBytes.Mul(idx)),
		SpanMap: make([]*Span, ArenaBytes/PageBytes),
	}
	h.arenaMap[idx] = uint32(aID + 1)
	h.Arenas = append(h.Arenas, a)
	return a
}

func (h *Heap) FindArena(p VAddr) *Arena {
	aID := h.ArenaIndexToID(p.ArenaIndex())
	if aID == NoArena {
		return nil
	}
	return h.Arenas[aID]
}

func (h *Heap) setSpan(span *Span) {
	if Bytes(span.Start)%PageBytes != 0 {
		panic("unaligned span base")
	}

	// Record page -> span map
	bp := span.Start.Page()
	for i := range span.NPages {
		pi := bp.Plus(i)
		arena := h.EnsureArena(pi.Start())
		ao := pi.ArenaOffset()

		if old := arena.SpanMap[ao]; old != nil {
			log.Fatalf("span overlap: new %+v, old %+v", span, old)
		}

		arena.SpanMap[ao] = span
	}

	// Allocate object metadata
	var nObj int
	if span.SizeClass == nil {
		nObj = 1
		span.End = span.Start.Plus(PageBytes.Mul(span.NPages))
	} else {
		// This intentionally rounds down
		nObj = PageBytes.Mul(span.NPages).Div(span.SizeClass.ObjectBytes)
		span.End = span.Start.Plus(span.SizeClass.ObjectBytes.Mul(nObj))
	}
	span.Objects = make([]Object, nObj)
	for i := range span.Objects {
		span.Objects[i].ID = h.nextObject
		h.nextObject++
	}
}

func (h *Heap) NewSpan(base VAddr, sc *SizeClass, noScan bool) *Span {
	span := &Span{
		Start:     base,
		NPages:    sc.SpanPages,
		SizeClass: sc,
		NoScan:    noScan,
	}
	h.setSpan(span)
	return span
}

func (h *Heap) NewSpanLarge(base VAddr, pages int, noScan bool) *Span {
	span := &Span{
		Start:     base,
		NPages:    pages,
		SizeClass: nil,
		NoScan:    noScan,
	}
	h.setSpan(span)
	return span
}

var FindDebug = false

func (h *Heap) Find(addr VAddr) (VAddr, *Object, *Span) {
	arena := h.FindArena(addr)
	if arena == nil {
		return 0, nil, nil
	}
	ao := addr.Page().ArenaOffset()
	span := arena.SpanMap[ao]
	if span == nil {
		return 0, nil, nil
	}
	if addr >= span.End {
		log.Printf("Find(%s) after end of span [%s,%s)", addr, span.Start, span.End)
		panic("bad pointer")
	}
	offset := addr.Minus(span.Start)
	var oidx int
	objBase := span.Start
	if span.SizeClass != nil {
		// This intentionally rounds down.
		oidx = int(offset / span.SizeClass.ObjectBytes)
		objBase = span.Start.Plus(Bytes(oidx) * span.SizeClass.ObjectBytes)

		if FindDebug {
			log.Printf("Find(%s) => span [%s,%s), offset %#x, oidx %#x, objBase %s", addr, span.Start, span.End, offset, oidx, objBase)
		}
	}
	return objBase, &span.Objects[oidx], span
}

func (h *Heap) VAddrToLAddr(addr VAddr) LAddr {
	idx, off := addr.Arena()
	aID := h.ArenaIndexToID(idx)
	if aID == NoArena {
		return NoLAddr
	}
	return LAddr(aID)*LAddr(ArenaBytes) + LAddr(off)
}

func (h *Heap) LAddrToVAddr(addr LAddr) VAddr {
	aID, off := addr.Arena()
	arena := h.Arenas[aID]
	return arena.Start.Plus(off)
}

func (h *Heap) Objects() iter.Seq2[VAddr, *Object] {
	return func(yield func(VAddr, *Object) bool) {
		for _, x := range h.arenaMap {
			aid := ArenaID(x - 1)
			if aid == NoArena {
				continue
			}
			var prev *Span
			for _, span := range h.Arenas[aid].SpanMap {
				if span == nil || span == prev {
					continue
				}
				prev = span
				addr := span.Start
				for i := range span.Objects {
					yield(addr, &span.Objects[i])
					if span.SizeClass != nil {
						addr = addr.Plus(span.SizeClass.ObjectBytes)
					}
				}
			}
		}
	}
}

func (h *Heap) ArenasIn(start, end VAddr) iter.Seq[*Arena] {
	r := Range{start, end.Minus(start)}
	startIdx := start.ArenaIndex()
	endIdx := (end.Plus(ArenaBytes - 1)).ArenaIndex()
	return func(yield func(*Arena) bool) {
		for _, x := range h.arenaMap[startIdx:endIdx] {
			aid := ArenaID(x - 1)
			if aid == NoArena {
				continue
			}
			arena := h.Arenas[aid]
			if !arena.Range().Overlaps(r) {
				panic("arena doesn't overlap range")
			}
			yield(arena)
		}
	}
}

func (h *Heap) SpansIn(r Range) iter.Seq[*Span] {
	startPage := r.Start.Page()
	endPage := r.End().Plus(PageBytes - 1).Page()
	return func(yield func(*Span) bool) {
		for page := startPage; page < endPage; {
			aIdx := page.Start().ArenaIndex()
			aID := h.ArenaIndexToID(aIdx)
			if aID == NoArena {
				// Skip to the next arena start.
				page = VPage((aIdx + 1) * ArenaBytes.Div(PageBytes))
				continue
			}
			arena := h.Arenas[aID]
			// Loop over spans in the arena or until end.
			loopEnd := min(endPage, arena.Start.Plus(ArenaBytes).Page())
			for page < loopEnd {
				span := arena.SpanMap[page.ArenaOffset()]
				if span == nil {
					page++
					continue
				}
				if !span.Range().Overlaps(r) {
					panic("span doesn't overlap range")
				}
				if !yield(span) {
					return
				}
				page = span.Start.Page().Plus(span.NPages)
			}
		}
	}
}
