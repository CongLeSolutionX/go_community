// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	. "runtime"
	"testing"
)

// makePallocData produces an initialized PallocData by setting
// the ranges of described in alloc and scavenge.
func makePallocData(alloc, scavenged []BitRange) *PallocData {
	b := new(PallocData)
	for _, v := range alloc {
		b.AllocRange(v.I, v.N)
	}
	for _, v := range scavenged {
		b.ScavengedSetRange(v.I, v.N)
	}
	return b
}

func TestPallocDataFindScavengeCandidate(t *testing.T) {
	type test struct {
		alloc, scavenged []BitRange
		max              uintptr
		want             BitRange
	}
	tests := map[string]test{
		"AllFree": {
			max:  PallocChunkPages,
			want: BitRange{0, PallocChunkPages},
		},
		"AllScavenged": {
			scavenged: []BitRange{{0, PallocChunkPages}},
			max:       PallocChunkPages,
			want:      BitRange{0, 0},
		},
		"NoneFree": {
			alloc:     []BitRange{{0, PallocChunkPages}},
			scavenged: []BitRange{{PallocChunkPages / 2, PallocChunkPages / 2}},
			max:       PallocChunkPages,
			want:      BitRange{0, 0},
		},
		"OneFree": {
			alloc: []BitRange{{0, 40}, {41, PallocChunkPages - 41}},
			max:   PallocChunkPages,
			want:  BitRange{40, 1},
		},
		"OneScavenged": {
			alloc:     []BitRange{{0, 40}, {41, PallocChunkPages - 41}},
			scavenged: []BitRange{{40, 1}},
			max:       PallocChunkPages,
			want:      BitRange{0, 0},
		},
		"Mixed": {
			alloc:     []BitRange{{0, 40}, {42, PallocChunkPages - 42}},
			scavenged: []BitRange{{0, 41}, {42, PallocChunkPages - 42}},
			max:       PallocChunkPages,
			want:      BitRange{41, 1},
		},
		"StartFree": {
			alloc: []BitRange{{1, PallocChunkPages - 1}},
			max:   PallocChunkPages,
			want:  BitRange{0, 1},
		},
		"EndFree": {
			alloc: []BitRange{{0, PallocChunkPages - 1}},
			max:   PallocChunkPages,
			want:  BitRange{PallocChunkPages - 1, 1},
		},
		"Straddle64": {
			alloc: []BitRange{{0, 63}, {65, PallocChunkPages - 65}},
			max:   2,
			want:  BitRange{63, 2},
		},
		"Multi": {
			alloc:     []BitRange{{0, 63}, {65, 20}, {87, PallocChunkPages - 87}},
			scavenged: []BitRange{{86, 1}},
			max:       PallocChunkPages,
			want:      BitRange{85, 1},
		},
		"BottomEdge64WithFull": {
			alloc:     []BitRange{{64, 64}, {131, PallocChunkPages - 131}},
			scavenged: []BitRange{{1, 10}},
			max:       3,
			want:      BitRange{128, 3},
		},
		"BottomEdge64WithPocket": {
			alloc:     []BitRange{{64, 62}, {127, 1}, {131, PallocChunkPages - 131}},
			scavenged: []BitRange{{1, 10}},
			max:       3,
			want:      BitRange{128, 3},
		},
	}
	if PhysHugePageSize > uintptr(PageSize) {
		// Check hugepage preserving behavior.
		bits := uint(PhysHugePageSize / uintptr(PageSize))
		tests["PreserveHugePageBottom"] = test{
			alloc: []BitRange{{bits + 2, PallocChunkPages - (bits + 2)}},
			max:   3, // Make it so that max would have us try to break the huge page.
			want:  BitRange{0, bits + 2},
		}
		if bits >= 3*PallocChunkPages {
			// We need at least 3 huge pages in an arena for this test to make sense.
			tests["PreserveHugePageMiddle"] = test{
				alloc: []BitRange{{0, bits - 10}, {2*bits + 10, PallocChunkPages - (2*bits + 10)}},
				max:   12, // Make it so that max would have us try to break the huge page.
				want:  BitRange{bits, bits + 10},
			}
		}
		tests["PreserveHugePageTop"] = test{
			alloc: []BitRange{{0, PallocChunkPages - bits}},
			max:   1, // Even one page would break a huge page in this case.
			want:  BitRange{PallocChunkPages - bits, bits},
		}
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makePallocData(v.alloc, v.scavenged)
			start, size := b.FindScavengeCandidate(PallocChunkPages-1, v.max)
			got := BitRange{start, size}
			if !(got.N == 0 && v.want.N == 0) && got != v.want {
				t.Fatalf("candidate mismatch: got %v, want %v", got, v.want)
			}
		})
	}
}

// Tests end-to-end scavenging on a pageAlloc.
func TestPageAllocScavenge(t *testing.T) {
	type test struct {
		request, expect uintptr
	}
	tests := map[string]struct {
		beforeAlloc map[ChunkIdx][]BitRange
		beforeScav  map[ChunkIdx][]BitRange
		expect      []test
		afterScav   map[ChunkIdx][]BitRange
	}{
		"AllFreeUnscavExhaust": {
			beforeAlloc: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
			},
			beforeScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
			},
			expect: []test{
				{^uintptr(0), 3 * PallocChunkPages * PageSize},
			},
			afterScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {{0, PallocChunkPages}},
				BaseChunkIdx + 1: {{0, PallocChunkPages}},
				BaseChunkIdx + 2: {{0, PallocChunkPages}},
			},
		},
		"NoneFreeUnscavExhaust": {
			beforeAlloc: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {{0, PallocChunkPages}},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {{0, PallocChunkPages}},
			},
			beforeScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {{0, PallocChunkPages}},
				BaseChunkIdx + 2: {},
			},
			expect: []test{
				{^uintptr(0), 0},
			},
			afterScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {{0, PallocChunkPages}},
				BaseChunkIdx + 2: {},
			},
		},
		"ScavHighestPageFirst": {
			beforeAlloc: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {},
			},
			beforeScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {{1, PallocChunkPages - 2}},
			},
			expect: []test{
				{1, PageSize},
			},
			afterScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {{1, PallocChunkPages - 1}},
			},
		},
		"ScavMultiple": {
			beforeAlloc: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {},
			},
			beforeScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {{1, PallocChunkPages - 2}},
			},
			expect: []test{
				{1, PageSize},
				{1, PageSize},
			},
			afterScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx: {{0, PallocChunkPages}},
			},
		},
		"ScavMultiple2": {
			beforeAlloc: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
			beforeScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {{1, PallocChunkPages - 2}},
				BaseChunkIdx + 1: {{0, PallocChunkPages - 2}},
			},
			expect: []test{
				{2 * PageSize, 2 * PageSize},
				{1, PageSize},
				{1, PageSize},
			},
			afterScav: map[ChunkIdx][]BitRange{
				BaseChunkIdx:     {{0, PallocChunkPages}},
				BaseChunkIdx + 1: {{0, PallocChunkPages}},
			},
		},
	}
	for name, v := range tests {
		v := v
		runTest := func(t *testing.T, locked bool) {
			b := NewPageAlloc(v.beforeAlloc, v.beforeScav)
			defer FreePageAlloc(b)

			for iter, h := range v.expect {
				if got := b.Scavenge(h.request, locked); got != h.expect {
					t.Fatalf("bad scavenge #%d: want %d, got %d", iter+1, h.expect, got)
				}
			}
			want := NewPageAlloc(v.beforeAlloc, v.afterScav)
			defer FreePageAlloc(want)

			checkPageAlloc(t, want, b)
		}
		t.Run(name, func(t *testing.T) {
			runTest(t, false)
		})
		t.Run(name+"Locked", func(t *testing.T) {
			runTest(t, true)
		})
	}
}
