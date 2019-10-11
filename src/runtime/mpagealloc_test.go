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

	for i := gotStart / MallocChunksPerArena; i <= gotEnd/MallocChunksPerArena; i++ {
		// Sanity check mheap_.arenas.
		if got.HasArena(i) != want.HasArena(i) {
			t.Fatalf("unexpected nilness mismatch for arenas at %d; bad test?", i)
		} else if !got.HasArena(i) {
			continue
		}

		// Check mheap_.arenas' bitmaps.
		if !checkMallocBits(t, got.MallocBits(i), want.MallocBits(i)) {
			t.Logf("in arena %d", i)
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
				BaseArenaIdx: {},
			},
			hits: []hit{
				{1, PageBase(BaseArenaIdx, 0)},
				{1, PageBase(BaseArenaIdx, 1)},
				{1, PageBase(BaseArenaIdx, 2)},
				{1, PageBase(BaseArenaIdx, 3)},
				{1, PageBase(BaseArenaIdx, 4)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, 5}},
			},
		},
		"ManyArena1": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena - 1}},
			},
			hits: []hit{
				{1, PageBase(BaseArenaIdx+2, PagesPerArena-1)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
			},
		},
		"NotContiguous1": {
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
				0xcff:        {{0, 0}},
			},
			hits: []hit{
				{1, PageBase(0xcff, 0)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
				0xcff:        {{0, 1}},
			},
		},
		"AllFree2": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{2, PageBase(BaseArenaIdx, 0)},
				{2, PageBase(BaseArenaIdx, 2)},
				{2, PageBase(BaseArenaIdx, 4)},
				{2, PageBase(BaseArenaIdx, 6)},
				{2, PageBase(BaseArenaIdx, 8)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, 10}},
			},
		},
		"Straddle2": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena - 1}},
				BaseArenaIdx + 1: {{1, PagesPerArena - 1}},
			},
			hits: []hit{
				{2, PageBase(BaseArenaIdx, PagesPerArena-1)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
			},
		},
		"AllFree5": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{5, PageBase(BaseArenaIdx, 0)},
				{5, PageBase(BaseArenaIdx, 5)},
				{5, PageBase(BaseArenaIdx, 10)},
				{5, PageBase(BaseArenaIdx, 15)},
				{5, PageBase(BaseArenaIdx, 20)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, 25}},
			},
		},
		"AllFree64": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{64, PageBase(BaseArenaIdx, 0)},
				{64, PageBase(BaseArenaIdx, 64)},
				{64, PageBase(BaseArenaIdx, 128)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, 192}},
			},
		},
		"AllFree65": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{65, PageBase(BaseArenaIdx, 0)},
				{65, PageBase(BaseArenaIdx, 65)},
				{65, PageBase(BaseArenaIdx, 130)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, 195}},
			},
		},
		// TODO(mknyszek): Add tests close to the chunk size.
		"ExhaustPagesPerArena-3": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{PagesPerArena - 3, PageBase(BaseArenaIdx, 0)},
				{PagesPerArena - 3, 0},
				{1, PageBase(BaseArenaIdx, PagesPerArena-3)},
				{2, PageBase(BaseArenaIdx, PagesPerArena-2)},
				{1, 0},
				{PagesPerArena - 3, 0},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
		},
		"AllFreePagesPerArena": {
			before: map[int][]BitRange{
				BaseArenaIdx: {},
			},
			hits: []hit{
				{PagesPerArena, PageBase(BaseArenaIdx, 0)},
				{PagesPerArena, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}},
				BaseArenaIdx + 1: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
			hits: []hit{
				{PagesPerArena, PageBase(BaseArenaIdx, PagesPerArena/2)},
				{PagesPerArena, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena+1": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}},
				BaseArenaIdx + 1: {},
			},
			hits: []hit{
				{PagesPerArena + 1, PageBase(BaseArenaIdx, PagesPerArena/2)},
				{PagesPerArena, 0},
				{1, PageBase(BaseArenaIdx+1, PagesPerArena/2+1)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena/2 + 2}},
			},
		},
		"AllFreePagesPerArena*2": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(BaseArenaIdx, 0)},
				{PagesPerArena * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
			},
		},
		"NotContiguousPagesPerArena*2": {
			before: map[int][]BitRange{
				BaseArenaIdx:         {},
				BaseArenaIdx + 0x100: {},
				BaseArenaIdx + 0x101: {},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(BaseArenaIdx+0x100, 0)},
				{21, PageBase(BaseArenaIdx, 0)},
				{1, PageBase(BaseArenaIdx, 21)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:         {{0, 22}},
				BaseArenaIdx + 0x100: {{0, PagesPerArena}},
				BaseArenaIdx + 0x101: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena*2": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}},
				BaseArenaIdx + 1: {},
				BaseArenaIdx + 2: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(BaseArenaIdx, PagesPerArena/2)},
				{PagesPerArena * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena*5/4": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena * 3 / 4}},
				BaseArenaIdx + 2: {{0, PagesPerArena * 3 / 4}},
				BaseArenaIdx + 3: {{0, 0}},
			},
			hits: []hit{
				{PagesPerArena * 5 / 4, PageBase(BaseArenaIdx+2, PagesPerArena*3/4)},
				{PagesPerArena * 5 / 4, 0},
				{1, PageBase(BaseArenaIdx+1, PagesPerArena*3/4)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena*3/4 + 1}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
				BaseArenaIdx + 3: {{0, PagesPerArena}},
			},
		},
		"AllFreePagesPerArena*7+5": {
			before: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
				BaseArenaIdx + 2: {},
				BaseArenaIdx + 3: {},
				BaseArenaIdx + 4: {},
				BaseArenaIdx + 5: {},
				BaseArenaIdx + 6: {},
				BaseArenaIdx + 7: {},
			},
			hits: []hit{
				{PagesPerArena*7 + 5, PageBase(BaseArenaIdx, 0)},
				{PagesPerArena*7 + 5, 0},
				{1, PageBase(BaseArenaIdx+7, 5)},
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
				BaseArenaIdx + 3: {{0, PagesPerArena}},
				BaseArenaIdx + 4: {{0, PagesPerArena}},
				BaseArenaIdx + 5: {{0, PagesPerArena}},
				BaseArenaIdx + 6: {{0, PagesPerArena}},
				BaseArenaIdx + 7: {{0, 6}},
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

// TODO(mknyszek): Devise a test which ensures that the page allocator never
// accesses summaries which don't represent real areans. The trouble right now
// is that because summary memory is mapped in as needed, there may be
// "out-of-bounds" accesses that don't cause a fault. For the most part these
// are harmless since these summaries indicate that there's nothing to be found,
// and the page allocator is implemented defensively to avoid making assumptions,
// but in real code bugs have already been found where this defensiveness was
// not enough. One way to do this might be to maintain a separate slice for
// each summary level which represents only what you're allowed to acccess, such
// that bounds checks would cause a fault and existing tests should be sufficient
// to catch these cases. The trouble with this is now the summaries cannot be
// indexed directly, we must maintain a bottom index for each level and subtract
// the index from that before access, which is a pain to do in every location,
// and further obscures the code. The alternative is to write a set of tests
// such that the arena boundaries line up with physical page boundaries in the
// summary reservation, but this makes for obscure tests.

func TestPageAllocExhaust(t *testing.T) {
	for _, npages := range []uintptr{1, 2, 3, 4, 5, 8, 16, 64, 1024, 1025, 2048, 2049} {
		npages := npages
		t.Run(fmt.Sprintf("%d", npages), func(t *testing.T) {
			// Construct b.
			bDesc := make(map[int][]BitRange)
			for i := 0; i < 4; i++ {
				bDesc[BaseArenaIdx+i] = []BitRange{}
			}
			b := GetTestPageAlloc(bDesc)
			defer PutTestPageAlloc(b)

			// Allocate into b with npages until we've exhausted the heap.
			nAlloc := (PagesPerArena * 4) / int(npages)
			for i := 0; i < nAlloc; i++ {
				addr := PageBase(BaseArenaIdx, i*int(npages))
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
				if allocPages >= PagesPerArena {
					wantDesc[BaseArenaIdx+i] = []BitRange{{0, PagesPerArena}}
					allocPages -= PagesPerArena
				} else if allocPages > 0 {
					wantDesc[BaseArenaIdx+i] = []BitRange{{0, allocPages}}
					allocPages = 0
				} else {
					wantDesc[BaseArenaIdx+i] = []BitRange{}
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
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
				PageBase(BaseArenaIdx, 1),
				PageBase(BaseArenaIdx, 2),
				PageBase(BaseArenaIdx, 3),
				PageBase(BaseArenaIdx, 4),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{5, PagesPerArena - 5}},
			},
		},
		"ManyArena1": {
			npages: 1,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, PagesPerArena/2),
				PageBase(BaseArenaIdx+1, 0),
				PageBase(BaseArenaIdx+2, PagesPerArena-1),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}, {PagesPerArena/2 + 1, PagesPerArena/2 - 1}},
				BaseArenaIdx + 1: {{1, PagesPerArena - 1}},
				BaseArenaIdx + 2: {{0, PagesPerArena - 1}},
			},
		},
		"Free2": {
			npages: 2,
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
				PageBase(BaseArenaIdx, 2),
				PageBase(BaseArenaIdx, 4),
				PageBase(BaseArenaIdx, 6),
				PageBase(BaseArenaIdx, 8),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{10, PagesPerArena - 10}},
			},
		},
		"Straddle2": {
			npages: 2,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{PagesPerArena - 1, 1}},
				BaseArenaIdx + 1: {{0, 1}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, PagesPerArena-1),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
			},
		},
		"Free5": {
			npages: 5,
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
				PageBase(BaseArenaIdx, 5),
				PageBase(BaseArenaIdx, 10),
				PageBase(BaseArenaIdx, 15),
				PageBase(BaseArenaIdx, 20),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{25, PagesPerArena - 25}},
			},
		},
		"Free64": {
			npages: 64,
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
				PageBase(BaseArenaIdx, 64),
				PageBase(BaseArenaIdx, 128),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{192, PagesPerArena - 192}},
			},
		},
		"Free65": {
			npages: 65,
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
				PageBase(BaseArenaIdx, 65),
				PageBase(BaseArenaIdx, 130),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {{195, PagesPerArena - 195}},
			},
		},
		"FreePagesPerArena": {
			npages: PagesPerArena,
			before: map[int][]BitRange{
				BaseArenaIdx: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
			},
			after: map[int][]BitRange{
				BaseArenaIdx: {},
			},
		},
		"StraddlePagesPerArena": {
			npages: PagesPerArena,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{PagesPerArena / 2, PagesPerArena / 2}},
				BaseArenaIdx + 1: {{0, PagesPerArena / 2}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
			},
		},
		"StraddlePagesPerArena+1": {
			npages: PagesPerArena + 1,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}},
				BaseArenaIdx + 1: {{PagesPerArena/2 + 1, PagesPerArena/2 - 1}},
			},
		},
		"FreePagesPerArena*2": {
			npages: PagesPerArena * 2,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
			},
		},
		"StraddlePagesPerArena*2": {
			npages: PagesPerArena * 2,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena / 2}},
				BaseArenaIdx + 1: {},
				BaseArenaIdx + 2: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
		},
		"AllFreePagesPerArena*7+5": {
			npages: PagesPerArena*7 + 5,
			before: map[int][]BitRange{
				BaseArenaIdx:     {{0, PagesPerArena}},
				BaseArenaIdx + 1: {{0, PagesPerArena}},
				BaseArenaIdx + 2: {{0, PagesPerArena}},
				BaseArenaIdx + 3: {{0, PagesPerArena}},
				BaseArenaIdx + 4: {{0, PagesPerArena}},
				BaseArenaIdx + 5: {{0, PagesPerArena}},
				BaseArenaIdx + 6: {{0, PagesPerArena}},
				BaseArenaIdx + 7: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(BaseArenaIdx, 0),
			},
			after: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
				BaseArenaIdx + 2: {},
				BaseArenaIdx + 3: {},
				BaseArenaIdx + 4: {},
				BaseArenaIdx + 5: {},
				BaseArenaIdx + 6: {},
				BaseArenaIdx + 7: {{5, PagesPerArena - 5}},
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
		"Arenas8": {
			init: map[int][]BitRange{
				BaseArenaIdx:     {},
				BaseArenaIdx + 1: {},
				BaseArenaIdx + 2: {},
				BaseArenaIdx + 3: {},
				BaseArenaIdx + 4: {},
				BaseArenaIdx + 5: {},
				BaseArenaIdx + 6: {},
				BaseArenaIdx + 7: {},
			},
			hits: []hit{
				{true, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{false, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{true, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{false, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{true, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{false, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
				{true, 1, PageBase(BaseArenaIdx, 0)},
				{false, 1, PageBase(BaseArenaIdx, 0)},
				{true, PagesPerArena * 8, PageBase(BaseArenaIdx, 0)},
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
