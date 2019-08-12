// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
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
func checkMallocBits(t *testing.T, got, want *MallocBits) {
	d := diffMallocBits(got, want)
	if len(d) != 0 {
		t.Errorf("%d location(s) different", len(d))
		for _, i := range d {
			t.Logf("\t@ bit index %d", i*64)
			t.Logf("\t|  got: %064b", got[i])
			t.Logf("\t| want: %064b", want[i])
		}
	}
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

// Ensures page allocation works.
func TestMallocBitsAlloc(t *testing.T) {
	tests := map[string]struct {
		before []BitRange
		after  []BitRange
		npages uintptr
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
			hits:   []int{-1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree2": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 2,
			hits:   []int{-1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree5": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 5,
			hits:   []int{-1, -1},
			after:  []BitRange{{0, PagesPerArena}},
		},
		"NoneFree65": {
			before: []BitRange{{0, PagesPerArena}},
			npages: 65,
			hits:   []int{-1, -1},
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
		"SomeFree161": {
			before: []BitRange{{0, 185}, {331, 1}},
			npages: 161,
			hits:   []int{332},
			after:  []BitRange{{0, 185}, {331, 162}},
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
		npages    uintptr
	}{
		"SomeFree": {
			npages:    1,
			beforeInv: []BitRange{{0, 32}, {64, 32}, {100, 1}},
			frees:     []int{32},
			afterInv:  []BitRange{{0, 33}, {64, 32}, {100, 1}},
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
			for _, i := range v.frees {
				b.Free(i, int(v.npages))
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
			t.Errorf("got %016x, want %016x", y, result)
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
			t.Errorf("got %016x, want %016x", y, result)
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
			t.Errorf("case (%016x, %d): got %d, want failure", x, n, i)
		} else if result >= 0 && i != result {
			t.Errorf("case (%016x, %d): got %d, want %d", x, n, i, result)
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
