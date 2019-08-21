// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	. "runtime"
	"testing"
)

// makeMallocData produces an initialized MallocData by setting
// the ranges of described in alloc and scavenge.
func makeMallocData(alloc, scavenged []BitRange) *MallocData {
	b := new(MallocData)
	for _, v := range alloc {
		b.AllocRange(v.I, v.N)
	}
	for _, v := range scavenged {
		b.ScavengeRange(v.I, v.N)
	}
	return b
}

func TestMallocDataFindScavengeCandidate(t *testing.T) {
	type test struct {
		alloc, scavenged []BitRange
		max              uint
		hit              BitRange
	}
	tests := map[string]test{
		"AllFree": {
			max: MallocChunkPages,
			hit: BitRange{0, MallocChunkPages},
		},
		"AllScavenged": {
			scavenged: []BitRange{{0, MallocChunkPages}},
			max:       MallocChunkPages,
			hit:       BitRange{0, 0},
		},
		"NoneFree": {
			alloc:     []BitRange{{0, MallocChunkPages}},
			scavenged: []BitRange{{MallocChunkPages / 2, MallocChunkPages / 2}},
			max:       MallocChunkPages,
			hit:       BitRange{0, 0},
		},
		"OneFree": {
			alloc: []BitRange{{0, 40}, {41, MallocChunkPages - 41}},
			max:   MallocChunkPages,
			hit:   BitRange{40, 1},
		},
		"OneScavenged": {
			alloc:     []BitRange{{0, 40}, {41, MallocChunkPages - 41}},
			scavenged: []BitRange{{40, 1}},
			max:       MallocChunkPages,
			hit:       BitRange{0, 0},
		},
		"Mixed": {
			alloc:     []BitRange{{0, 40}, {42, MallocChunkPages - 42}},
			scavenged: []BitRange{{0, 41}, {42, MallocChunkPages - 42}},
			max:       MallocChunkPages,
			hit:       BitRange{41, 1},
		},
		"StartFree": {
			alloc: []BitRange{{1, MallocChunkPages - 1}},
			max:   MallocChunkPages,
			hit:   BitRange{0, 1},
		},
		"EndFree": {
			alloc: []BitRange{{0, MallocChunkPages - 1}},
			max:   MallocChunkPages,
			hit:   BitRange{MallocChunkPages - 1, 1},
		},
		"Straddle64": {
			alloc: []BitRange{{0, 63}, {65, MallocChunkPages - 65}},
			max:   2,
			hit:   BitRange{63, 2},
		},
		"Multi": {
			alloc:     []BitRange{{0, 63}, {65, 20}, {87, MallocChunkPages - 87}},
			scavenged: []BitRange{{86, 1}},
			max:       MallocChunkPages,
			hit:       BitRange{85, 1},
		},
		"BottomEdge64WithFull": {
			alloc:     []BitRange{{64, 64}, {131, MallocChunkPages - 131}},
			scavenged: []BitRange{{1, 10}},
			max:       3,
			hit:       BitRange{128, 3},
		},
		"BottomEdge64WithPocket": {
			alloc:     []BitRange{{64, 62}, {127, 1}, {131, MallocChunkPages - 131}},
			scavenged: []BitRange{{1, 10}},
			max:       3,
			hit:       BitRange{128, 3},
		},
	}
	if PhysHugePageSize > uintptr(PageSize) {
		// Check hugepage preserving behavior.
		bits := uint(PhysHugePageSize / uintptr(PageSize))
		tests["PreserveHugePageBottom"] = test{
			alloc: []BitRange{{bits + 2, MallocChunkPages - (bits + 2)}},
			max:   3, // Make it so that max would have us try to break the huge page.
			hit:   BitRange{0, bits + 2},
		}
		if bits >= 3*MallocChunkPages {
			// We need at least 3 huge pages in an arena for this test to make sense.
			tests["PreserveHugePageMiddle"] = test{
				alloc: []BitRange{{0, bits - 10}, {2*bits + 10, MallocChunkPages - (2*bits + 10)}},
				max:   12, // Make it so that max would have us try to break the huge page.
				hit:   BitRange{bits, bits + 10},
			}
		}
		tests["PreserveHugePageTop"] = test{
			alloc: []BitRange{{0, MallocChunkPages - bits}},
			max:   1, // Even one page would break a huge page in this case.
			hit:   BitRange{MallocChunkPages - bits, bits},
		}
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makeMallocData(v.alloc, v.scavenged)
			start, size := b.FindScavengeCandidate(MallocChunkPages-1, v.max)
			got := BitRange{start, size}
			if !(got.N == 0 && v.hit.N == 0) && got != v.hit {
				t.Fatalf("candidate mismatch: got %v, want %v", got, v.hit)
			}
		})
	}
}

// Tests end-to-end scavenging on a pageAlloc.
func TestPageAllocScavenge(t *testing.T) {
	type hit struct {
		request, expect uintptr
	}
	tests := map[string]struct {
		beforeAlloc map[int][]BitRange
		beforeScav  map[int][]BitRange
		hits        []hit
		afterScav   map[int][]BitRange
	}{
		"AllFreeUnscavExhaust": {
			beforeAlloc: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
			},
			beforeScav: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {},
			},
			hits: []hit{
				{^uintptr(0), 3 * MallocChunkPages * PageSize},
			},
			afterScav: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
		},
		"NoneFreeUnscavExhaust": {
			beforeAlloc: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {},
				BaseChunkIdx + 2: {{0, MallocChunkPages}},
			},
			beforeScav: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {},
			},
			hits: []hit{
				{^uintptr(0), 0},
			},
			afterScav: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
				BaseChunkIdx + 2: {},
			},
		},
		"ScavHighestPageFirst": {
			beforeAlloc: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			beforeScav: map[int][]BitRange{
				BaseChunkIdx: {{1, MallocChunkPages - 2}},
			},
			hits: []hit{
				{1, PageSize},
			},
			afterScav: map[int][]BitRange{
				BaseChunkIdx: {{1, MallocChunkPages - 1}},
			},
		},
		"ScavMultiple": {
			beforeAlloc: map[int][]BitRange{
				BaseChunkIdx: {},
			},
			beforeScav: map[int][]BitRange{
				BaseChunkIdx: {{1, MallocChunkPages - 2}},
			},
			hits: []hit{
				{1, PageSize},
				{1, PageSize},
			},
			afterScav: map[int][]BitRange{
				BaseChunkIdx: {{0, MallocChunkPages}},
			},
		},
		"ScavMultiple2": {
			beforeAlloc: map[int][]BitRange{
				BaseChunkIdx:     {},
				BaseChunkIdx + 1: {},
			},
			beforeScav: map[int][]BitRange{
				BaseChunkIdx:     {{1, MallocChunkPages - 2}},
				BaseChunkIdx + 1: {{0, MallocChunkPages - 2}},
			},
			hits: []hit{
				{2 * PageSize, 2 * PageSize},
				{1, PageSize},
				{1, PageSize},
			},
			afterScav: map[int][]BitRange{
				BaseChunkIdx:     {{0, MallocChunkPages}},
				BaseChunkIdx + 1: {{0, MallocChunkPages}},
			},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := GetTestPageAlloc(v.beforeAlloc)
			b.InitScavState(v.beforeScav)
			defer PutTestPageAlloc(b)

			for iter, h := range v.hits {
				if got := b.Scavenge(h.request); got != h.expect {
					t.Fatalf("bad scavenge #%d: want %d, got %d", iter+1, h.expect, got)
				}
			}
			want := GetTestPageAlloc(v.beforeAlloc)
			want.InitScavState(v.afterScav)
			defer PutTestPageAlloc(want)

			checkPageAlloc(t, want, b)
		})
	}
}
