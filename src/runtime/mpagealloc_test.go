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

	for i := gotStart; i <= gotEnd; i++ {
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
				0xc00: {},
			},
			hits: []hit{
				{1, PageBase(0xc00, 0)},
				{1, PageBase(0xc00, 1)},
				{1, PageBase(0xc00, 2)},
				{1, PageBase(0xc00, 3)},
				{1, PageBase(0xc00, 4)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 5}},
			},
		},
		"ManyArena1": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena - 1}},
			},
			hits: []hit{
				{1, PageBase(0xc02, PagesPerArena-1)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
			},
		},
		"NotContiguous1": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xcff: {{0, 0}},
			},
			hits: []hit{
				{1, PageBase(0xcff, 0)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xcff: {{0, 1}},
			},
		},
		"AllFree2": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{2, PageBase(0xc00, 0)},
				{2, PageBase(0xc00, 2)},
				{2, PageBase(0xc00, 4)},
				{2, PageBase(0xc00, 6)},
				{2, PageBase(0xc00, 8)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 10}},
			},
		},
		"Straddle2": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena - 1}},
				0xc01: {{1, PagesPerArena - 1}},
			},
			hits: []hit{
				{2, PageBase(0xc00, PagesPerArena-1)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
			},
		},
		"AllFree5": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{5, PageBase(0xc00, 0)},
				{5, PageBase(0xc00, 5)},
				{5, PageBase(0xc00, 10)},
				{5, PageBase(0xc00, 15)},
				{5, PageBase(0xc00, 20)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 25}},
			},
		},
		"AllFree64": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{64, PageBase(0xc00, 0)},
				{64, PageBase(0xc00, 64)},
				{64, PageBase(0xc00, 128)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 192}},
			},
		},
		"AllFree65": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{65, PageBase(0xc00, 0)},
				{65, PageBase(0xc00, 65)},
				{65, PageBase(0xc00, 130)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 195}},
			},
		},
		"ExhaustPagesPerArena-3": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{PagesPerArena - 3, PageBase(0xc00, 0)},
				{PagesPerArena - 3, 0},
				{1, PageBase(0xc00, PagesPerArena-3)},
				{2, PageBase(0xc00, PagesPerArena-2)},
				{1, 0},
				{PagesPerArena - 3, 0},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
		},
		"AllFreePagesPerArena": {
			before: map[int][]BitRange{
				0xc00: {},
			},
			hits: []hit{
				{PagesPerArena, PageBase(0xc00, 0)},
				{PagesPerArena, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}},
				0xc01: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
			hits: []hit{
				{PagesPerArena, PageBase(0xc00, PagesPerArena/2)},
				{PagesPerArena, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena+1": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}},
				0xc01: {},
			},
			hits: []hit{
				{PagesPerArena + 1, PageBase(0xc00, PagesPerArena/2)},
				{PagesPerArena, 0},
				{1, PageBase(0xc01, PagesPerArena/2+1)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena/2 + 2}},
			},
		},
		"AllFreePagesPerArena*2": {
			before: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(0xc00, 0)},
				{PagesPerArena * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
			},
		},
		"NotContiguousPagesPerArena*2": {
			before: map[int][]BitRange{
				0xc00: {},
				0xd00: {},
				0xd01: {},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(0xd00, 0)},
				{21, PageBase(0xc00, 0)},
				{1, PageBase(0xc00, 21)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, 22}},
				0xd00: {{0, PagesPerArena}},
				0xd01: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena*2": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}},
				0xc01: {},
				0xc02: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
			hits: []hit{
				{PagesPerArena * 2, PageBase(0xc00, PagesPerArena/2)},
				{PagesPerArena * 2, 0},
				{1, 0},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
			},
		},
		"StraddlePagesPerArena*5/4": {
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena * 3 / 4}},
				0xc02: {{0, PagesPerArena * 3 / 4}},
				0xc03: {{0, 0}},
			},
			hits: []hit{
				{PagesPerArena * 5 / 4, PageBase(0xc02, PagesPerArena*3/4)},
				{PagesPerArena * 5 / 4, 0},
				{1, PageBase(0xc01, PagesPerArena*3/4)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena*3/4 + 1}},
				0xc02: {{0, PagesPerArena}},
				0xc03: {{0, PagesPerArena}},
			},
		},
		"AllFreePagesPerArena*7+5": {
			before: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
				0xc02: {},
				0xc03: {},
				0xc04: {},
				0xc05: {},
				0xc06: {},
				0xc07: {},
			},
			hits: []hit{
				{PagesPerArena*7 + 5, PageBase(0xc00, 0)},
				{PagesPerArena*7 + 5, 0},
				{1, PageBase(0xc07, 5)},
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
				0xc03: {{0, PagesPerArena}},
				0xc04: {{0, PagesPerArena}},
				0xc05: {{0, PagesPerArena}},
				0xc06: {{0, PagesPerArena}},
				0xc07: {{0, 6}},
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
				bDesc[0xc00+i] = []BitRange{}
			}
			b := GetTestPageAlloc(bDesc)
			defer PutTestPageAlloc(b)

			// Allocate into b with npages until we've exhausted the heap.
			nAlloc := (PagesPerArena * 4) / int(npages)
			for i := 0; i < nAlloc; i++ {
				addr := PageBase(0xc00, i*int(npages))
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
					wantDesc[0xc00+i] = []BitRange{{0, 1024}}
					allocPages -= PagesPerArena
				} else if allocPages > 0 {
					wantDesc[0xc00+i] = []BitRange{{0, allocPages}}
					allocPages = 0
				} else {
					wantDesc[0xc00+i] = []BitRange{}
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
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
				PageBase(0xc00, 1),
				PageBase(0xc00, 2),
				PageBase(0xc00, 3),
				PageBase(0xc00, 4),
			},
			after: map[int][]BitRange{
				0xc00: {{5, PagesPerArena - 5}},
			},
		},
		"ManyArena1": {
			npages: 1,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, PagesPerArena/2),
				PageBase(0xc01, 0),
				PageBase(0xc02, PagesPerArena-1),
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}, {PagesPerArena/2 + 1, PagesPerArena/2 - 1}},
				0xc01: {{1, PagesPerArena - 1}},
				0xc02: {{0, PagesPerArena - 1}},
			},
		},
		"Free2": {
			npages: 2,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
				PageBase(0xc00, 2),
				PageBase(0xc00, 4),
				PageBase(0xc00, 6),
				PageBase(0xc00, 8),
			},
			after: map[int][]BitRange{
				0xc00: {{10, PagesPerArena - 10}},
			},
		},
		"Straddle2": {
			npages: 2,
			before: map[int][]BitRange{
				0xc00: {{PagesPerArena - 1, 1}},
				0xc01: {{0, 1}},
			},
			frees: []uintptr{
				PageBase(0xc00, PagesPerArena-1),
			},
			after: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
			},
		},
		"Free5": {
			npages: 5,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
				PageBase(0xc00, 5),
				PageBase(0xc00, 10),
				PageBase(0xc00, 15),
				PageBase(0xc00, 20),
			},
			after: map[int][]BitRange{
				0xc00: {{25, PagesPerArena - 25}},
			},
		},
		"Free64": {
			npages: 64,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
				PageBase(0xc00, 64),
				PageBase(0xc00, 128),
			},
			after: map[int][]BitRange{
				0xc00: {{192, PagesPerArena - 192}},
			},
		},
		"Free65": {
			npages: 65,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
				PageBase(0xc00, 65),
				PageBase(0xc00, 130),
			},
			after: map[int][]BitRange{
				0xc00: {{195, PagesPerArena - 195}},
			},
		},
		"FreePagesPerArena": {
			npages: PagesPerArena,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
			},
			after: map[int][]BitRange{
				0xc00: {},
			},
		},
		"StraddlePagesPerArena": {
			npages: PagesPerArena,
			before: map[int][]BitRange{
				0xc00: {{PagesPerArena / 2, PagesPerArena / 2}},
				0xc01: {{0, PagesPerArena / 2}},
			},
			frees: []uintptr{
				PageBase(0xc00, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
			},
		},
		"StraddlePagesPerArena+1": {
			npages: PagesPerArena + 1,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}},
				0xc01: {{PagesPerArena/2 + 1, PagesPerArena/2 - 1}},
			},
		},
		"FreePagesPerArena*2": {
			npages: PagesPerArena * 2,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
			},
			after: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
			},
		},
		"StraddlePagesPerArena*2": {
			npages: PagesPerArena * 2,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, PagesPerArena/2),
			},
			after: map[int][]BitRange{
				0xc00: {{0, PagesPerArena / 2}},
				0xc01: {},
				0xc02: {{PagesPerArena / 2, PagesPerArena / 2}},
			},
		},
		"AllFreePagesPerArena*7+5": {
			npages: PagesPerArena*7 + 5,
			before: map[int][]BitRange{
				0xc00: {{0, PagesPerArena}},
				0xc01: {{0, PagesPerArena}},
				0xc02: {{0, PagesPerArena}},
				0xc03: {{0, PagesPerArena}},
				0xc04: {{0, PagesPerArena}},
				0xc05: {{0, PagesPerArena}},
				0xc06: {{0, PagesPerArena}},
				0xc07: {{0, PagesPerArena}},
			},
			frees: []uintptr{
				PageBase(0xc00, 0),
			},
			after: map[int][]BitRange{
				0xc00: {},
				0xc01: {},
				0xc02: {},
				0xc03: {},
				0xc04: {},
				0xc05: {},
				0xc06: {},
				0xc07: {{5, PagesPerArena - 5}},
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
				0xc00: {},
				0xc01: {},
				0xc02: {},
				0xc03: {},
				0xc04: {},
				0xc05: {},
				0xc06: {},
				0xc07: {},
			},
			hits: []hit{
				{true, PagesPerArena * 8, PageBase(0xc00, 0)},
				{false, PagesPerArena * 8, PageBase(0xc00, 0)},
				{true, PagesPerArena * 8, PageBase(0xc00, 0)},
				{false, PagesPerArena * 8, PageBase(0xc00, 0)},
				{true, PagesPerArena * 8, PageBase(0xc00, 0)},
				{false, PagesPerArena * 8, PageBase(0xc00, 0)},
				{true, 1, PageBase(0xc00, 0)},
				{false, 1, PageBase(0xc00, 0)},
				{true, PagesPerArena * 8, PageBase(0xc00, 0)},
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
