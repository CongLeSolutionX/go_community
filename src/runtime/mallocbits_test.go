// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	. "runtime"
	"testing"
)

// Ensures chunk8 works as expected.
func TestChunk8(t *testing.T) {
	x := make([]uint8, 8)
	c := UnsafeChunkFromSlice(x)
	if v := c.Load(); v != 0 {
		t.Fatalf("first load: got %016x, want %016x", v, 0)
	}
	toStore := uint64(0xff00ff00ff00ff00)
	c.Store(toStore)
	if v := c.Load(); v != toStore {
		t.Fatalf("second load: got %016x, want %016x", v, 0)
	}
	for i := 0; i < len(x); i++ {
		expect := uint8((toStore >> uint(i*8)) & 0xff)
		if x[i] != expect {
			t.Fatalf("checking byte %d: got %02x, want %02x", i, x[i], expect)
		}
	}
}

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
			t.Logf("\t@ bit index %d", i*8)
			t.Logf("\t|  got: %08b", got[i])
			t.Logf("\t| want: %08b", want[i])
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
		want[PagesPerArena/8-1] = 0x80
		test(t, 1023, 1, want)
	})
	t.Run("Inner", func(t *testing.T) {
		want := new(MallocBits)
		want[2] = 0x3e
		test(t, 17, 5, want)
	})
	t.Run("Aligned", func(t *testing.T) {
		want := new(MallocBits)
		want[2] = 0xff
		want[3] = 0xff
		test(t, 16, 16, want)
	})
	t.Run("Begin", func(t *testing.T) {
		want := new(MallocBits)
		want[0] = 0xff
		want[1] = 0xff
		want[2] = 0xff
		want[3] = 0xff
		want[4] = 0xff
		want[5] = 0x1
		test(t, 0, 41, want)
	})
	t.Run("End", func(t *testing.T) {
		want := new(MallocBits)
		want[PagesPerArena/8-1] = 0xff
		want[PagesPerArena/8-2] = 0xff
		want[PagesPerArena/8-3] = 0xff
		want[PagesPerArena/8-4] = 0x80
		test(t, PagesPerArena-25, 25, want)
	})
	t.Run("All", func(t *testing.T) {
		want := new(MallocBits)
		for i := range want {
			want[i] = 0xff
		}
		test(t, 0, 1024, want)
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
	tests := map[string]struct {
		s               []BitRange
		start, max, end int
	}{
		"NoneFree": {},
		"OnlyStart": {
			s:     []BitRange{{0, 10}},
			start: 10,
			max:   10,
		},
		"OnlyEnd": {
			s:   []BitRange{{PagesPerArena - 40, 40}},
			max: 40,
			end: 40,
		},
		"StartAndEnd": {
			s:     []BitRange{{0, 11}, {PagesPerArena - 23, 23}},
			start: 11,
			max:   23,
			end:   23,
		},
		"StartMaxEnd": {
			s:     []BitRange{{0, 4}, {50, 100}, {PagesPerArena - 4, 4}},
			start: 4,
			max:   100,
			end:   4,
		},
		"OnlyMax": {
			s:   []BitRange{{1, 20}, {35, 241}, {PagesPerArena - 50, 30}},
			max: 241,
		},
		"MultiMax": {
			s:   []BitRange{{35, 2}, {40, 5}, {100, 5}},
			max: 5,
		},
		"One": {
			s:   []BitRange{{2, 1}},
			max: 1,
		},
		"AllFree": {
			s:     []BitRange{{0, 1024}},
			start: 1024,
			max:   1024,
			end:   1024,
		},
	}
	for name, v := range tests {
		v := v
		t.Run(name, func(t *testing.T) {
			b := makeMallocBits(v.s)
			// In the MallocBits we create 1's represent free spots, but in our actual
			// MallocBits 1 means not free, so invert.
			invertMallocBits(b)
			checkMallocSum(t, b.Summarize(), PackMallocSum(v.start, v.max, v.end))
		})
	}
}

// Benchmarks how quickly we can summarize a MallocBits.
func BenchmarkMallocBitsSummarize(b *testing.B) {
	buf0 := new(MallocBits)
	buf1 := new(MallocBits)
	for i := 0; i < len(buf1); i++ {
		buf1[i] = ^uint8(0)
	}
	bufa := new(MallocBits)
	for i := 0; i < len(bufa); i++ {
		bufa[i] = 0xaa
	}
	for _, buf := range []*MallocBits{buf0, buf1, bufa} {
		b.Run(fmt.Sprintf("Unpacked%02X", buf[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf.Summarize()
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
			before: []BitRange{{0, 1024}},
			npages: 1,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, 1024}},
		},
		"NoneFree2": {
			before: []BitRange{{0, 1024}},
			npages: 2,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, 1024}},
		},
		"NoneFree5": {
			before: []BitRange{{0, 1024}},
			npages: 5,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, 1024}},
		},
		"NoneFree65": {
			before: []BitRange{{0, 1024}},
			npages: 65,
			hits:   []int{-1, -1, -1, -1},
			after:  []BitRange{{0, 1024}},
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
		if i != result {
			t.Fatalf("case (%016x, %d): got %d, want %d", x, n, i, result)
		}
	}
	for i := 0; i <= 64; i++ {
		check(^uint64(0), i, 0)
	}
	check(0, 0, 0)
	for i := 1; i <= 64; i++ {
		check(0, i, 64)
	}
	check(0x8000000000000000, 1, 63)
	check(0xc000010001010000, 2, 62)
	check(0xc000010001030000, 2, 16)
	check(0xe000030001030000, 3, 61)
	check(0xe000030001070000, 3, 16)
	check(0xffff03ff01070000, 16, 48)
	check(0xffff03ff0107ffff, 16, 0)
	check(0x0fff03ff01079fff, 16, 64)
}
