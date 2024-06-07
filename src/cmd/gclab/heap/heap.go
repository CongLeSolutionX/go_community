// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package heap

import (
	"cmd/gclab/bitmap"
	"fmt"
	"iter"
	"log"
	"slices"
)

const (
	WordBytes  Bytes = 8
	PageBytes  Bytes = 8 << 10
	ArenaBytes Bytes = 64 << 20 // TODO Put in setup event?

	PageWords  Words = Words(PageBytes / WordBytes)
	ArenaWords Words = Words(ArenaBytes / WordBytes)

	MaxObjsPerSpan  = 1024
	MaxPagesPerSpan = 10 // Excluding large object spans

	// Align first object ID in each span to objectCountAlign.
	objectCountAlign = 64

	NumSizeClasses = 68
)

type SizeClass struct {
	ID           int
	ObjectBytes  Bytes
	SpanPages    int
	HeapBitsType HeapBitsType
}

type HeapBitsType int

const (
	HeapBitsNone   HeapBitsType = iota // No heap bits (no pointers)
	HeapBitsPacked                     // Packed at end of span
	HeapBitsHeader                     // Pointed to by each object header
	HeapBitsOOB                        // Pointed to by span struct (one object per span)
)

type Heap struct {
	SizeClasses []SizeClass

	Types map[uint64]*Type

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

	// mem is the memory backing this arena, starting at Start and continuing to
	// the end of the contiguous memory starting at Start. This is generally
	// aliased between Arenas. Copies of this slice may go stale when a new
	// Arena is created.
	mem []byte
}

func (a Arena) Range() Range {
	return Range{a.Start, ArenaBytes}
}

func (a Arena) String() string {
	return fmt.Sprintf("Arena[%d]/%s", uint32(a.ID), Range{a.Start, ArenaBytes})
}

type Span struct {
	Start     VAddr
	End       VAddr // Byte just past last object in the span
	NPages    int
	SizeClass *SizeClass

	HeapBitsType HeapBitsType
	// HeapBits specifies the pointer/scalar information for objects in this
	// span. The representation depends on HeapBitsType:
	//
	// HeapBitsNone: This is nil.
	//
	// HeapBitsPacked: Bit i indicates that word i of the span is a pointer.
	//
	// HeapBitsHeader: Entry i is the type ID for object i. In the real runtime,
	// these IDs are stored in allocation headers immediately preceding each
	// object.
	//
	// HeapBitsOOB: Like HeapBitsHeader, but the values are stored in the span
	// metadata.
	HeapBits []uint64

	AllocBits bitmap.Set[uint64]

	FirstObject ObjectID
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

type Type struct {
	ID       uint64
	Size     Bytes
	PtrWords Words
	PtrMask  []uint64

	// TODO: dgcptrmask in the compiler writes out PtrMask as bytes, but already
	// padded out to uintptr. The tracer rewrites it to uint64s for us, but in
	// the real system, we should just make dgcptrmask write uintptrs. That
	// would simplify a bunch of stuff in the runtime anyway.
}

func NewHeap(sc []SizeClass) *Heap {
	if len(sc) != NumSizeClasses {
		log.Fatalf("expected %d size classes, got %d", NumSizeClasses, len(sc))
	}

	for i, s := range sc {
		if s.ID != i {
			panic("bad SizeClass ID")
		}
		if i == 0 {
			if s != (SizeClass{HeapBitsType: HeapBitsOOB}) {
				log.Fatalf("bad 0th size class: %+v", s)
			}
			continue
		}
		if s.SpanPages > MaxPagesPerSpan {
			log.Fatalf("too many pages in size class %+v: %d > %d", s, s.SpanPages, MaxPagesPerSpan)
		}
		if PageBytes.Mul(s.SpanPages).Div(s.ObjectBytes) > MaxObjsPerSpan {
			panic("too many objects in size class")
		}
	}

	var h Heap
	h.SizeClasses = sc
	h.Types = make(map[uint64]*Type)
	h.nextObject = objectCountAlign
	return &h
}

func (h *Heap) ArenaIndexToID(idx int) ArenaID {
	if idx < 0 || idx >= len(h.arenaMap) {
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

	// Allocate objects.
	var nObj int
	if span.SizeClass == nil {
		nObj = 1
		span.End = span.Start.Plus(PageBytes.Mul(span.NPages))
	} else {
		// This intentionally rounds down
		nObj = PageBytes.Mul(span.NPages).Div(span.SizeClass.ObjectBytes)
		span.End = span.Start.Plus(span.SizeClass.ObjectBytes.Mul(nObj))
	}
	span.FirstObject = h.nextObject
	h.nextObject += ObjectID(nObj)
	h.nextObject = (h.nextObject + objectCountAlign - 1) &^ (objectCountAlign - 1)
}

func (h *Heap) NewSpan(base VAddr, sc *SizeClass, noScan bool) *Span {
	span := &Span{
		Start:        base,
		NPages:       sc.SpanPages,
		SizeClass:    sc,
		HeapBitsType: sc.HeapBitsType,
	}
	if noScan {
		span.HeapBitsType = HeapBitsNone
	}
	h.setSpan(span)
	return span
}

func (h *Heap) NewSpanLarge(base VAddr, pages int, noScan bool) *Span {
	span := &Span{
		Start:        base,
		NPages:       pages,
		SizeClass:    nil,
		HeapBitsType: HeapBitsOOB,
	}
	if noScan {
		span.HeapBitsType = HeapBitsNone
	}
	h.setSpan(span)
	return span
}

func (h *Heap) NewType(id uint64, size Bytes, ptrWords Words, ptrData []uint64) {
	if h.Types[id] != nil {
		log.Fatalf("duplicate type ID %d", id)
	}
	h.Types[id] = &Type{id, size, ptrWords, ptrData}
}

var FindDebug = false

func (h *Heap) FindSpan(addr VAddr) (span *Span) {
	arena := h.FindArena(addr)
	if arena == nil {
		return nil
	}
	ao := addr.Page().ArenaOffset()
	return arena.SpanMap[ao]
}

func (h *Heap) FindArenaAndSpan(addr VAddr) (arena *Arena, span *Span) {
	arena = h.FindArena(addr)
	if arena == nil {
		return nil, nil
	}
	ao := addr.Page().ArenaOffset()
	return arena, arena.SpanMap[ao]
}

func (h *Heap) FindObject(addr VAddr) (objBase VAddr, span *Span, objID ObjectID) {
	span = h.FindSpan(addr)
	if span == nil {
		return 0, nil, 0
	}
	if addr >= span.End {
		log.Printf("FindObject(%s) after end of span [%s,%s)", addr, span.Start, span.End)
		panic("bad pointer")
	}
	offset := addr.Minus(span.Start)
	objBase = span.Start
	objID = span.FirstObject
	if span.SizeClass != nil {
		// This intentionally rounds down.
		objIndex := int(offset / span.SizeClass.ObjectBytes)
		objBase = span.Start.Plus(Bytes(objIndex) * span.SizeClass.ObjectBytes)
		objID += ObjectID(objIndex)

		if FindDebug {
			log.Printf("FindObject(%s) => span [%s,%s), offset %#x, oidx %#x, objBase %s", addr, span.Start, span.End, offset, objIndex, objBase)
		}
	}
	return objBase, span, objID
}

func (h *Heap) Mem(base VAddr, length Bytes) []byte {
	aIdx := base.ArenaIndex()
	aID := h.ArenaIndexToID(aIdx)
	if aID == NoArena {
		return nil
	}
	a := h.Arenas[aID]

	if a.mem == nil {
		// Lazily initialize the memory.
		//
		// Find the contiguous run of arenas containing a.
		start := aIdx
		for start > 0 && h.ArenaIndexToID(start-1) != NoArena {
			start--
		}
		end := aIdx + 1
		for h.ArenaIndexToID(end) != NoArena {
			end++
		}
		// Build the new mem.
		var mem []byte
		for idx1 := start; idx1 < end; idx1++ {
			a1 := h.Arenas[h.ArenaIndexToID(idx1)]
			if a1.mem == nil {
				mem = append(mem, make([]byte, ArenaBytes)...)
			} else {
				mem = append(mem, a1.mem...)
			}
		}
		// Install the new mme.
		for idx1 := start; idx1 < end; idx1++ {
			a1 := h.Arenas[h.ArenaIndexToID(idx1)]
			a1.mem = mem
			mem = mem[ArenaBytes:]
		}
		if len(mem) != 0 {
			panic("mem wrong length")
		}
	}

	return a.mem[base.Minus(a.Start):][:length]
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

func (h *Heap) ObjectIDs() ObjectID {
	return h.nextObject
}

func (h *Heap) Objects() iter.Seq2[VAddr, ObjectID] {
	return func(yield func(VAddr, ObjectID) bool) {
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
				objID := span.FirstObject
				for range span.NObjects() {
					yield(addr, objID)
					objID++
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
	endPage := r.End().Page()
	if Bytes(r.End())%PageBytes != 0 {
		endPage++
	}
	return func(yield func(*Span) bool) {
		for page := startPage; page < endPage; {
			aIdx := page.Start().ArenaIndex()
			aID := h.ArenaIndexToID(aIdx)
			if aID == NoArena {
				if aIdx >= len(h.arenaMap) {
					break
				}
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
