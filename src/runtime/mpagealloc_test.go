// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	. "runtime"
	"testing"
)

func checkPageAlloc(t *testing.T, want, got *PageAlloc) {
	// Ensure start and end are correct.
	wantStart, wantEnd := want.Bounds()
	gotStart, gotEnd := got.Bounds()
	if gotStart != wantStart {
		t.Fatalf("start values not equal: got %d, want %d", gotStart, wantStart)
	}
	if gotEnd != wantEnd {
		t.Fatalf("end values not equal: got %d, want %d", gotEnd, wantEnd)
	}

	for i := gotStart; i < gotEnd; i++ {
		// Sanity check mheap_.arenas.
		if got.HasChunk(i) != want.HasChunk(i) {
			t.Fatalf("unexpected nilness mismatch for arenas at %d; bad test?", i)
		} else if !got.HasChunk(i) {
			continue
		}

		// Check the bitmaps.
		if !checkMallocBits(t, got.MallocBits(i), want.MallocBits(i)) {
			t.Logf("in chunk %d", i)
		}
	}
	// TODO(mknyszek): Verify summaries too?
}

func TestPageAllocAlloc(t *testing.T) {
	type hit struct {
		npages, base uintptr
	}
	tests := map[string]struct {
		before map[int][]BitRange
		after  map[int][]BitRange
		hits   []hit
	}{
		"AllFree1": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{1, PageBase(BaseChunkIdx, 0)},
				{1, PageBase(BaseChunkIdx, 1)},
				{1, PageBase(BaseChunkIdx, 2)},
				{1, PageBase(BaseChunkIdx, 3)},
				{1, PageBase(BaseChunkIdx, 4)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, 5}},
			},
		},
		"ManyArena1": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages - 1}},
			},
			hits: []hit{
				{1, PageBase(BaseChunkIdx+2, MallocChunkPages-1)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
		},
		"NotContiguous1": {
			before: map[int][]BitRange{
				BaseChunkIdx:        {{0, MallocChunkPages}},
				BaseChunkIdx + 0xff: {{0, 0}},
			},
			hits: []hit{
				{1, PageBase(BaseChunkIdx+0xff, 0)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:        {{0, MallocChunkPages}},
				BaseChunkIdx + 0xff: {{0, 1}},
			},
		},
		"AllFree2": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{2, PageBase(BaseChunkIdx, 0)},
				{2, PageBase(BaseChunkIdx, 2)},
				{2, PageBase(BaseChunkIdx, 4)},
				{2, PageBase(BaseChunkIdx, 6)},
				{2, PageBase(BaseChunkIdx, 8)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, 10}},
			},
		},
		"Straddle2": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages - 1}},
				BaseChunkIdx + 1: {{1, MallocChunkPages - 1}},
			},
			hits: []hit{
				{2, PageBase(BaseChunkIdx, MallocChunkPages-1)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
		},
		"AllFree5": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{5, PageBase(BaseChunkIdx, 0)},
				{5, PageBase(BaseChunkIdx, 5)},
				{5, PageBase(BaseChunkIdx, 10)},
				{5, PageBase(BaseChunkIdx, 15)},
				{5, PageBase(BaseChunkIdx, 20)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, 25}},
			},
		},
		"AllFree64": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{64, PageBase(BaseChunkIdx, 0)},
				{64, PageBase(BaseChunkIdx, 64)},
				{64, PageBase(BaseChunkIdx, 128)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, 192}},
			},
		},
		"AllFree65": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{65, PageBase(BaseChunkIdx, 0)},
				{65, PageBase(BaseChunkIdx, 65)},
				{65, PageBase(BaseChunkIdx, 130)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, 195}},
			},
		},
		// TODO(mknyszek): Add tests close to the chunk size.
		"ExhaustMallocChunkPages-3": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{MallocChunkPages - 3, PageBase(BaseChunkIdx, 0)},
				{MallocChunkPages - 3, 0},
				{1, PageBase(BaseChunkIdx, MallocChunkPages-3)},
				{2, PageBase(BaseChunkIdx, MallocChunkPages-2)},
				{1, 0},
				{MallocChunkPages - 3, 0},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
		},
		"AllFreeMallocChunkPages": {
			before: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			hits: []hit{
				{MallocChunkPages, PageBase(BaseChunkIdx, 0)},
				{MallocChunkPages, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
		},
		"StraddleMallocChunkPages": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {{MallocChunkPages / 2, MallocChunkPages / 2}},
			},
			hits: []hit{
				{MallocChunkPages, PageBase(BaseChunkIdx, MallocChunkPages/2)},
				{MallocChunkPages, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
		},
		"StraddleMallocChunkPages+1": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {},
			},
			hits: []hit{
				{MallocChunkPages + 1, PageBase(BaseChunkIdx, MallocChunkPages/2)},
				{MallocChunkPages, 0},
				{1, PageBase(BaseChunkIdx+1, MallocChunkPages/2+1)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages/2 + 2}},
			},
		},
		"AllFreeMallocChunkPages*2": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
			hits: []hit{
				{MallocChunkPages * 2, PageBase(BaseChunkIdx, 0)},
				{MallocChunkPages * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
		},
		"NotContiguousMallocChunkPages*2": {
			before: map[int][]BitRange{
				BaseChunkIdx:         {},
				BaseChunkIdx + 0x100: {},
				BaseChunkIdx + 0x101: {},
			},
			hits: []hit{
				{MallocChunkPages * 2, PageBase(BaseChunkIdx+0x100, 0)},
				{21, PageBase(BaseChunkIdx, 0)},
				{1, PageBase(BaseChunkIdx, 21)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:         {{0, 22}},
				BaseChunkIdx + 0x100: {{0, MallocChunkPages}},
				BaseChunkIdx + 0x101: {{0, MallocChunkPages}},
			},
		},
		"StraddleMallocChunkPages*2": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {{MallocChunkPages / 2, MallocChunkPages / 2}},
			},
			hits: []hit{
				{MallocChunkPages * 2, PageBase(BaseChunkIdx, MallocChunkPages/2)},
				{MallocChunkPages * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
		},
		"StraddleMallocChunkPages*5/4": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages * 3 / 4}},
				BaseChunkIdx + 2: {{0, MallocChunkPages * 3 / 4}},
				BaseChunkIdx + 3: {{0, 0}},
			},
			hits: []hit{
				{MallocChunkPages * 5 / 4, PageBase(BaseChunkIdx+2, MallocChunkPages*3/4)},
				{MallocChunkPages * 5 / 4, 0},
				{1, PageBase(BaseChunkIdx+1, MallocChunkPages*3/4)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages*3/4 + 1}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
				BaseChunkIdx + 3: {{0, MallocChunkPages}},
			},
		},
		"AllFreeMallocChunkPages*7+5": {
			before: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
				BaseChunkIdx + 3: {},
				BaseChunkIdx + 4: {},
				BaseChunkIdx + 5: {},
				BaseChunkIdx + 6: {},
				BaseChunkIdx + 7: {},
			},
			hits: []hit{
				{MallocChunkPages*7 + 5, PageBase(BaseChunkIdx, 0)},
				{MallocChunkPages*7 + 5, 0},
				{1, PageBase(BaseChunkIdx+7, 5)},
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
				BaseChunkIdx + 3: {{0, MallocChunkPages}},
				BaseChunkIdx + 4: {{0, MallocChunkPages}},
				BaseChunkIdx + 5: {{0, MallocChunkPages}},
				BaseChunkIdx + 6: {{0, MallocChunkPages}},
				BaseChunkIdx + 7: {{0, 6}},
			},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := GetTestPageAlloc(v.before)
			defer PutTestPageAlloc(b)

			for iter, i := range v.hits {
				if a := b.Alloc(i.npages); a != i.base {
					t.Fatalf("bad alloc #%d: want 0x%x, got 0x%x", iter+1, i.base, a)
				}
			}
			want := GetTestPageAlloc(v.after)
			defer PutTestPageAlloc(want)

			checkPageAlloc(t, want, b)
		})
	}
}

func TestPageAllocExhaust(t *testing.T) {
	for _, npages := range []uintptr{1, 2, 3, 4, 5, 8, 16, 64, 1024, 1025, 2048, 2049} {
		npages := npages
		t.Run(fmt.Sprintf("%d", npages), func(t *testing.T) {
			// Construct b.
			bDesc := make(map[int][]BitRange)
			for i := 0; i < 4; i++ {
				bDesc[BaseChunkIdx+i] = []BitRange{}
			}
			b := GetTestPageAlloc(bDesc)
			defer PutTestPageAlloc(b)

			// Allocate into b with npages until we've exhausted the heap.
			nAlloc := (MallocChunkPages * 4) / int(npages)
			for i := 0; i < nAlloc; i++ {
				addr := PageBase(BaseChunkIdx, i*int(npages))
				if a := b.Alloc(npages); a != addr {
					t.Fatalf("bad alloc #%d: want 0x%x, got 0x%x", i+1, addr, a)
				}
			}

			// Check to make sure the next allocation fails.
			if a := b.Alloc(npages); a != 0 {
				t.Fatalf("bad alloc #%d: want 0, got 0x%x", nAlloc, a)
			}

			// Construct what we want the heap to look like now.
			allocPages := nAlloc * int(npages)
			wantDesc := make(map[int][]BitRange)
			for i := 0; i < 4; i++ {
				if allocPages >= MallocChunkPages {
					wantDesc[BaseChunkIdx+i] = []BitRange{{0, MallocChunkPages}}
					allocPages -= MallocChunkPages
				} else if allocPages > 0 {
					wantDesc[BaseChunkIdx+i] = []BitRange{{0, allocPages}}
					allocPages = 0
				} else {
					wantDesc[BaseChunkIdx+i] = []BitRange{}
				}
			}
			want := GetTestPageAlloc(wantDesc)
			defer PutTestPageAlloc(want)

			// Check to make sure the heap b matches what we want.
			checkPageAlloc(t, want, b)
		})
	}
}

func TestPageAllocFree(t *testing.T) {
	tests := map[string]struct {
		before map[int][]BitRange
		after  map[int][]BitRange
		npages uintptr
		frees  []uintptr // es to free
	}{
		"Free1": {
			npages: 1,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
				PageBase(BaseChunkIdx, 1),
				PageBase(BaseChunkIdx, 2),
				PageBase(BaseChunkIdx, 3),
				PageBase(BaseChunkIdx, 4),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{5, MallocChunkPages - 5}},
			},
		},
		"ManyArena1": {
			npages: 1,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, MallocChunkPages/2),
				PageBase(BaseChunkIdx+1, 0),
				PageBase(BaseChunkIdx+2, MallocChunkPages-1),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}, {MallocChunkPages/2 + 1, MallocChunkPages/2 - 1}},
				BaseChunkIdx + 1: {{1, MallocChunkPages - 1}},
				BaseChunkIdx + 2: {{0, MallocChunkPages - 1}},
			},
		},
		"Free2": {
			npages: 2,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
				PageBase(BaseChunkIdx, 2),
				PageBase(BaseChunkIdx, 4),
				PageBase(BaseChunkIdx, 6),
				PageBase(BaseChunkIdx, 8),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{10, MallocChunkPages - 10}},
			},
		},
		"Straddle2": {
			npages: 2,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{MallocChunkPages - 1, 1}},
				BaseChunkIdx + 1: {{0, 1}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, MallocChunkPages-1),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
		},
		"Free5": {
			npages: 5,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
				PageBase(BaseChunkIdx, 5),
				PageBase(BaseChunkIdx, 10),
				PageBase(BaseChunkIdx, 15),
				PageBase(BaseChunkIdx, 20),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{25, MallocChunkPages - 25}},
			},
		},
		"Free64": {
			npages: 64,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
				PageBase(BaseChunkIdx, 64),
				PageBase(BaseChunkIdx, 128),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{192, MallocChunkPages - 192}},
			},
		},
		"Free65": {
			npages: 65,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
				PageBase(BaseChunkIdx, 65),
				PageBase(BaseChunkIdx, 130),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {{195, MallocChunkPages - 195}},
			},
		},
		"FreeMallocChunkPages": {
			npages: MallocChunkPages,
			before: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
			},
			after: map[int][]BitRange{
				BaseChunkIdx: {},
			},
		},
		"StraddleMallocChunkPages": {
			npages: MallocChunkPages,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{MallocChunkPages / 2, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {{0, MallocChunkPages / 2}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, MallocChunkPages/2),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
		},
		"StraddleMallocChunkPages+1": {
			npages: MallocChunkPages + 1,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, MallocChunkPages/2),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {{MallocChunkPages/2 + 1, MallocChunkPages/2 - 1}},
			},
		},
		"FreeMallocChunkPages*2": {
			npages: MallocChunkPages * 2,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
		},
		"StraddleMallocChunkPages*2": {
			npages: MallocChunkPages * 2,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, MallocChunkPages/2),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages / 2}},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {{MallocChunkPages / 2, MallocChunkPages / 2}},
			},
		},
		"AllFreeMallocChunkPages*7+5": {
			npages: MallocChunkPages*7 + 5,
			before: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
				BaseChunkIdx + 3: {{0, MallocChunkPages}},
				BaseChunkIdx + 4: {{0, MallocChunkPages}},
				BaseChunkIdx + 5: {{0, MallocChunkPages}},
				BaseChunkIdx + 6: {{0, MallocChunkPages}},
				BaseChunkIdx + 7: {{0, MallocChunkPages}},
			},
			frees: []uintptr{
				PageBase(BaseChunkIdx, 0),
			},
			after: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
				BaseChunkIdx + 3: {},
				BaseChunkIdx + 4: {},
				BaseChunkIdx + 5: {},
				BaseChunkIdx + 6: {},
				BaseChunkIdx + 7: {{5, MallocChunkPages - 5}},
			},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := GetTestPageAlloc(v.before)
			defer PutTestPageAlloc(b)

			for _, addr := range v.frees {
				b.Free(addr, v.npages)
			}

			want := GetTestPageAlloc(v.after)
			defer PutTestPageAlloc(want)

			checkPageAlloc(t, want, b)
		})
	}
}

func TestPageAllocAllocAndFree(t *testing.T) {
	type hit struct {
		alloc  bool
		npages uintptr
		base   uintptr
	}
	tests := map[string]struct {
		init map[int][]BitRange
		hits []hit
	}{
		// TODO(mknyszek): Write more tests here.
		"Chunks8": {
			init: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
				BaseChunkIdx + 3: {},
				BaseChunkIdx + 4: {},
				BaseChunkIdx + 5: {},
				BaseChunkIdx + 6: {},
				BaseChunkIdx + 7: {},
			},
			hits: []hit{
				{true, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{false, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{true, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{false, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{true, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{false, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
				{true, 1, PageBase(BaseChunkIdx, 0)},
				{false, 1, PageBase(BaseChunkIdx, 0)},
				{true, MallocChunkPages * 8, PageBase(BaseChunkIdx, 0)},
			},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := GetTestPageAlloc(v.init)
			defer PutTestPageAlloc(b)

			for iter, i := range v.hits {
				if i.alloc {
					if a := b.Alloc(i.npages); a != i.base {
						t.Fatalf("bad alloc #%d: want 0x%x, got 0x%x", iter+1, i.base, a)
					}
				} else {
					b.Free(i.base, i.npages)
				}
			}
		})
	}
}
