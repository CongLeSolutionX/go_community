// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	. "runtime"
	"testing"
)

// Given two MallocBits, returns a slice of indices where
// they differ.
func diffMallocBits(a, b *MallocBits) []int {
	var d []int
	for i := range a {
		if a[i] != b[i] {
			d = append(d, i)
		}
	}
	return d
}

// Ensures that got and want are the same, and if not, reports
// detailed diff information.
func checkMallocBits(t *testing.T, got, want *MallocBits) bool {
	d := diffMallocBits(got, want)
	if len(d) != 0 {
		t.Errorf("%d location(s) different", len(d))
		for _, i := range d {
			t.Logf("\t@ bit index %d", i*64)
			t.Logf("\t|  got: %064b", got[i])
			t.Logf("\t| want: %064b", want[i])
		}
		return false
	}
	return true
}

// makeMallocBits produces an initialized MallocBits by setting
// the ranges of described in s to 1 and the rest to zero.
func makeMallocBits(s []BitRange) *MallocBits {
	b := new(MallocBits)
	for _, v := range s {
		b.AllocRange(v.I, v.N)
	}
	return b
}

// Ensures that MallocBits.AllocRange works, which is a fundamental
// method used for testing and initialization since it's used by
// makeMallocBits.
func TestMallocBitsAllocRange(t *testing.T) {
	test := func(t *testing.T, i, n int, want *MallocBits) {
		checkMallocBits(t, makeMallocBits([]BitRange{{i, n}}), want)
	}
	t.Run("OneLow", func(t *testing.T) {
		want := new(MallocBits)
		want[0] = 0x1
		test(t, 0, 1, want)
	})
	t.Run("OneHigh", func(t *testing.T) {
		want := new(MallocBits)
		want[PagesPerArena/64-1] = 1 << 63
		test(t, PagesPerArena-1, 1, want)
	})
	t.Run("Inner", func(t *testing.T) {
		want := new(MallocBits)
		want[2] = 0x3e
		test(t, 129, 5, want)
	})
	t.Run("Aligned", func(t *testing.T) {
		want := new(MallocBits)
		want[2] = ^uint64(0)
		want[3] = ^uint64(0)
		test(t, 128, 128, want)
	})
	t.Run("Begin", func(t *testing.T) {
		want := new(MallocBits)
		want[0] = ^uint64(0)
		want[1] = ^uint64(0)
		want[2] = ^uint64(0)
		want[3] = ^uint64(0)
		want[4] = ^uint64(0)
		want[5] = 0x1
		test(t, 0, 321, want)
	})
	t.Run("End", func(t *testing.T) {
		want := new(MallocBits)
		want[PagesPerArena/64-1] = ^uint64(0)
		want[PagesPerArena/64-2] = ^uint64(0)
		want[PagesPerArena/64-3] = ^uint64(0)
		want[PagesPerArena/64-4] = 1 << 63
		test(t, PagesPerArena-(64*3+1), 64*3+1, want)
	})
	t.Run("All", func(t *testing.T) {
		want := new(MallocBits)
		for i := range want {
			want[i] = ^uint64(0)
		}
		test(t, 0, PagesPerArena, want)
	})
}

// Inverts every bit in the MallocBits.
func invertMallocBits(b *MallocBits) {
	for i := range b {
		b[i] = ^b[i]
	}
}

// Ensures two packed summaries are identical, and reports a detailed description
// of the difference if they're not.
func checkMallocSum(t *testing.T, got, want MallocSum) {
	if got.Start() != want.Start() {
		t.Errorf("inconsistent start: got %d, want %d", got.Start(), want.Start())
	}
	if got.Max() != want.Max() {
		t.Errorf("inconsistent max: got %d, want %d", got.Max(), want.Max())
	}
	if got.End() != want.End() {
		t.Errorf("inconsistent end: got %d, want %d", got.End(), want.End())
	}
}

// Ensures computing bit summaries works as expected.
func TestMallocBitsSummarize(t *testing.T) {
	var emptySum = PackMallocSum(MallocChunkPages, MallocChunkPages, MallocChunkPages)
	type hit struct {
		n    int
		want MallocSum
	}
	tests := map[string]struct {
		free []BitRange // Ranges of free (zero) bits.
		hits []hit
	}{
		"NoneFree": {
			free: []BitRange{},
			hits: []hit{
				{0, PackMallocSum(0, 0, 0)},
				{1, PackMallocSum(0, 0, 0)},
				{2, PackMallocSum(0, 0, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"OnlyStart": {
			free: []BitRange{{0, 10}},
			hits: []hit{
				{0, PackMallocSum(10, 10, 0)},
				{1, PackMallocSum(0, 0, 0)},
				{2, PackMallocSum(0, 0, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"OnlyEnd": {
			free: []BitRange{{PagesPerArena - 40, 40}},
			hits: []hit{
				{MallocChunksPerArena - 1, PackMallocSum(0, 40, 40)},
				{MallocChunksPerArena - 2, PackMallocSum(0, 0, 0)},
				{0, PackMallocSum(0, 0, 0)},
			},
		},
		"StartAndEnd": {
			free: []BitRange{{0, 11}, {MallocChunkPages - 23, 23}},
			hits: []hit{
				{0, PackMallocSum(11, 23, 23)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"StartMaxEnd": {
			free: []BitRange{{0, 4}, {50, 100}, {MallocChunkPages - 4, 4}},
			hits: []hit{
				{0, PackMallocSum(4, 100, 4)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"OnlyMax": {
			free: []BitRange{{1, 20}, {35, 241}, {MallocChunkPages - 50, 30}},
			hits: []hit{
				{0, PackMallocSum(0, 241, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"MultiMax": {
			free: []BitRange{{35, 2}, {40, 5}, {100, 5}},
			hits: []hit{
				{0, PackMallocSum(0, 5, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"One": {
			free: []BitRange{{2, 1}},
			hits: []hit{
				{0, PackMallocSum(0, 1, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"AllFree": {
			free: []BitRange{{0, PagesPerArena}},
			hits: []hit{
				{0, emptySum},
				{1, emptySum},
				{2, emptySum},
				{MallocChunksPerArena / 2, emptySum},
				{MallocChunksPerArena - 1, emptySum},
			},
		},
		"Chunk5Free": {
			free: []BitRange{{5 * MallocChunkPages, MallocChunkPages}},
			hits: []hit{
				{0, PackMallocSum(0, 0, 0)},
				{1, PackMallocSum(0, 0, 0)},
				{2, PackMallocSum(0, 0, 0)},
				{5, emptySum},
				{MallocChunksPerArena / 2, PackMallocSum(0, 0, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
		"ChunkOverlap": {
			free: []BitRange{{MallocChunkPages/2 + 1, MallocChunkPages}},
			hits: []hit{
				{0, PackMallocSum(0, MallocChunkPages/2-1, MallocChunkPages/2-1)},
				{1, PackMallocSum(MallocChunkPages/2+1, MallocChunkPages/2+1, 0)},
				{2, PackMallocSum(0, 0, 0)},
				{5, PackMallocSum(0, 0, 0)},
				{MallocChunksPerArena / 2, PackMallocSum(0, 0, 0)},
				{MallocChunksPerArena - 1, PackMallocSum(0, 0, 0)},
			},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makeMallocBits(v.free)
			// In the MallocBits we create 1's represent free spots, but in our actual
			// MallocBits 1 means not free, so invert.
			invertMallocBits(b)
			for _, h := range v.hits {
				checkMallocSum(t, b.Summarize(h.n), h.want)
			}
		})
	}
}

// Benchmarks how quickly we can summarize a MallocBits.
func BenchmarkMallocBitsSummarize(b *testing.B) {
	buf0 := new(MallocBits)
	buf1 := new(MallocBits)
	for i := 0; i < len(buf1); i++ {
		buf1[i] = ^uint64(0)
	}
	bufa := new(MallocBits)
	for i := 0; i < len(bufa); i++ {
		bufa[i] = 0xaa
	}
	for _, buf := range []*MallocBits{buf0, buf1, bufa} {
		b.Run(fmt.Sprintf("Unpacked%02X", buf[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf.Summarize(0)
			}
		})
	}
}

// Ensures page allocation works.
func TestMallocBitsAlloc(t *testing.T) {
	tests := map[string]struct {
		before []BitRange
		after  []BitRange
		npages uintptr // if == 0, do fast 1-page alloc
		hits   []int
	}{
		"AllFree1": {
			npages: 1,
			hits:   []int{0, 1, 2, 3, 4, 5},
			after:  []BitRange{{0, 6}},
		},
		"AllFree2": {
			npages: 2,
			hits:   []int{0, 2, 4, 6, 8, 10},
			after:  []BitRange{{0, 12}},
		},
		"AllFree5": {
			npages: 5,
			hits:   []int{0, 5, 10, 15, 20},
			after:  []BitRange{{0, 25}},
		},
		"AllFree64": {
			npages: 64,
			hits:   []int{0, 64, 128},
			after:  []BitRange{{0, 192}},
		},
		"AllFree65": {
			npages: 65,
			hits:   []int{0, 65, 130},
			after:  []BitRange{{0, 195}},
		},
		"SomeFree64": {
			before: []BitRange{{0, 32}, {64, 32}, {100, PagesPerArena - 100}},
			npages: 64,
			hits:   []int{-1},
			after:  []BitRange{{0, 32}, {64, 32}, {100, PagesPerArena - 100}},
		},
		"NoneFree1": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 1,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree2": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 2,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree5": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 5,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree65": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 65,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"ExactFit1": {
			before: []BitRange{{0, PagesPerArena/2 - 3}, {PagesPerArena/2 - 2, PagesPerArena/2 + 2}},
			npages: 1,
			hits:   []int{PagesPerArena/2 - 3, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"ExactFit2": {
			before: []BitRange{{0, PagesPerArena/2 - 3}, {PagesPerArena/2 - 1, PagesPerArena/2 + 1}},
			npages: 2,
			hits:   []int{PagesPerArena/2 - 3, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"ExactFit5": {
			before: []BitRange{{0, PagesPerArena/2 - 3}, {PagesPerArena/2 + 2, PagesPerArena/2 - 2}},
			npages: 5,
			hits:   []int{PagesPerArena/2 - 3, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"ExactFit65": {
			before: []BitRange{{0, PagesPerArena/2 - 31}, {PagesPerArena/2 + 34, PagesPerArena/2 - 34}},
			npages: 65,
			hits:   []int{PagesPerArena/2 - 31, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"SomeFree680": {
			before: []BitRange{{0, 185}, {331, 1}},
			npages: 680,
			hits:   []int{332},
			after:  []BitRange{{0, 185}, {331, 681}},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makeMallocBits(v.before)
			for iter, i := range v.hits {
				if a, _ := b.Alloc(v.npages, 0); i != a {
					t.Fatalf("alloc #%d picked wrong index: want %d, got %d", iter+1, i, a)
				}
			}
			want := makeMallocBits(v.after)
			checkMallocBits(t, b, want)
		})
	}
}

// Ensures page freeing works.
func TestMallocBitsFree(t *testing.T) {
	tests := map[string]struct {
		beforeInv []BitRange
		afterInv  []BitRange
		frees     []int
		npages    uintptr // if == 0, do fast 1-page free
	}{
		"SomeFreeFast1": {
			beforeInv: []BitRange{{0, 32}, {64, 32}, {100, 1}},
			frees:     []int{32},
			afterInv:  []BitRange{{0, 33}, {64, 32}, {100, 1}},
		},
		"NoneFreeFast1": {
			frees:    []int{0, 1, 2, 3, 4, 5},
			afterInv: []BitRange{{0, 6}},
		},
		"NoneFree1": {
			npages:   1,
			frees:    []int{0, 1, 2, 3, 4, 5},
			afterInv: []BitRange{{0, 6}},
		},
		"NoneFree2": {
			npages:   2,
			frees:    []int{0, 2, 4, 6, 8, 10},
			afterInv: []BitRange{{0, 12}},
		},
		"NoneFree5": {
			npages:   5,
			frees:    []int{0, 5, 10, 15, 20},
			afterInv: []BitRange{{0, 25}},
		},
		"NoneFree64": {
			npages:   64,
			frees:    []int{0, 64, 128},
			afterInv: []BitRange{{0, 192}},
		},
		"NoneFree65": {
			npages:   65,
			frees:    []int{0, 65, 130},
			afterInv: []BitRange{{0, 195}},
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makeMallocBits(v.beforeInv)
			invertMallocBits(b)
			if v.npages == 0 {
				for _, i := range v.frees {
					b.Free1(i)
				}
			} else {
				for _, i := range v.frees {
					b.Free(i, int(v.npages))
				}
			}
			want := makeMallocBits(v.afterInv)
			invertMallocBits(want)
			checkMallocBits(t, b, want)
		})
	}
}

func TestSetConsecBits64(t *testing.T) {
	check := func(x uint64, i, n int, result uint64) {
		y := SetConsecBits64(x, i, n)
		if y != result {
			t.Fatalf("got %016x, want %016x", y, result)
		}
	}
	check(0, 0, 5, 0x1f)
	check(0, 0, 64, 0xffffffffffffffff)
	check(0, 48, 3, 0x0007000000000000)
	check(0, 48, 16, 0xffff000000000000)
	check(0, 48, 20, 0xffff000000000000)
	check(0x3, 0, 2, 0x3)
	check(0x3, 0, 3, 0x7)
}

func TestClearConsecBits64(t *testing.T) {
	check := func(x uint64, i, n int, result uint64) {
		y := ClearConsecBits64(x, i, n)
		if y != result {
			t.Fatalf("got %016x, want %016x", y, result)
		}
	}
	check(^uint64(0), 0, 5, 0xffffffffffffffe0)
	check(^uint64(0), 0, 64, 0)
	check(0, 0, 64, 0)
	check(^uint64(0), 48, 3, 0xfff8ffffffffffff)
	check(^uint64(0), 48, 16, 0x0000ffffffffffff)
	check(^uint64(0), 48, 20, 0x0000ffffffffffff)
	check(0xfffffffffffffc, 0, 2, 0xfffffffffffffc)
	check(0xfffffffffffffc, 0, 3, 0xfffffffffffff8)
}

func TestFindConsecN64(t *testing.T) {
	check := func(x uint64, n int, result int) {
		i := FindConsecN64(x, n)
		if result < 0 && i < 64 {
			t.Fatalf("case (%016x, %d): got %d, want failure", x, n, i)
		} else if result >= 0 && i != result {
			t.Fatalf("case (%016x, %d): got %d, want %d", x, n, i, result)
		}
	}
	for i := 0; i <= 64; i++ {
		check(^uint64(0), i, 0)
	}
	check(0, 0, 0)
	for i := 1; i <= 64; i++ {
		check(0, i, -1)
	}
	check(0x8000000000000000, 1, 63)
	check(0xc000010001010000, 2, 62)
	check(0xc000010001030000, 2, 16)
	check(0xe000030001030000, 3, 61)
	check(0xe000030001070000, 3, 16)
	check(0xffff03ff01070000, 16, 48)
	check(0xffff03ff0107ffff, 16, 0)
	check(0x0fff03ff01079fff, 16, -1)
}
