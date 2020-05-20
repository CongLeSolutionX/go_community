// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "testing"

// We generate memmove for copy(x[1:], x[:]), however we may change it to OpMove,
// because size is known. Check that OpMove is alias-safe, or we did call memmove.
func TestMove(t *testing.T) {
	x := [...]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}
	copy(x[1:], x[:])
	for i := 1; i < len(x); i++ {
		if int(x[i]) != i {
			t.Errorf("Memmove got converted to OpMove in alias-unsafe way. Got %d insted of %d in position %d", int(x[i]), i, i+1)
		}
	}
}

func TestMoveSmall(t *testing.T) {
	x := [...]byte{1, 2, 3, 4, 5, 6, 7}
	copy(x[1:], x[:])
	for i := 1; i < len(x); i++ {
		if int(x[i]) != i {
			t.Errorf("Memmove got converted to OpMove in alias-unsafe way. Got %d instead of %d in position %d", int(x[i]), i, i+1)
		}
	}
}

func TestSubFlags(t *testing.T) {
	if !subFlags32(0, 1).lt() {
		t.Errorf("subFlags32(0,1).lt() returned false")
	}
	if !subFlags32(0, 1).ult() {
		t.Errorf("subFlags32(0,1).ult() returned false")
	}
}

func genREV16D(c uint64) uint64 {
	var b uint64
	b = ((c & 0xff00ff00ff00ff00) >> 8) | ((c & 0x00ff00ff00ff00ff) << 8)
	return b
}

func genREV16W(c uint32) uint32 {
	var b uint32
	b = ((c & 0xff00ff00) >> 8) | ((c & 0x00ff00ff) << 8)
	return b
}

func TestREV16D(t *testing.T) {
	var ref1 uint64 = 0x8f7f6f5f4f3f2f1f
	var ref2 uint64 = 0x7f8f5f6f3f4f1f2f
	var test1 uint64

	test1 = genREV16D(ref2)
	if test1 != ref1 {
		t.Errorf("want %08x given %08x", ref1, test1)
	}
}

func TestREV16W(t *testing.T) {
	var ref1 uint32 = 0x4f3f2f1f
	var ref2 uint32 = 0x3f4f1f2f
	var test1 uint32

	test1 = genREV16W(ref2)
	if test1 != ref1 {
		t.Errorf("want %#x given %#x", ref1, test1)
	}
}

func storeREV16(u uint64, p []uint8, i int) {
	s := p[i : i+8 : i+8]
	s[0] = uint8(u >> 8)
	s[1] = uint8(u)
	s[2] = uint8(u >> 24)
	s[3] = uint8(u >> 16)
	s[4] = uint8(u >> 40)
	s[5] = uint8(u >> 32)
	s[6] = uint8(u >> 56)
	s[7] = uint8(u >> 48)
}

func TestStoreREV16(t *testing.T) {
	p := make([]uint8, 32)
	want := [8]uint8{0x2f, 0x1f, 0x4f, 0x3f, 0x6f, 0x5f, 0x8f, 0x7f}
	var ref uint64 = 0x8f7f6f5f4f3f2f1f

	for i := 0; i <= 24; i = i + 8 {
		storeREV16(ref, p, i)
		for j, v := range want {
			if v != p[i+j] {
				t.Errorf("want %#x given %#x", v, p[i+j])
			}
		}
	}
}

func storeSmall(p []uint8, ref uint64, count int) {
	for i := 0; i < count; i = i + 8 {
		storeREV16(ref, p, i)
	}
}

func BenchmarkStoreRev16(b *testing.B) {
	p := make([]uint8, 65)
	var ref uint64 = 0x8f7f6f5f4f3f2f1f
	for n := 0; n < b.N; n++ {
		storeSmall(p, ref, 64)
	}
}

func BenchmarkGenRev16D(b *testing.B) {
	p := make([]uint64, 65)
	var ref uint64 = 0x8f7f6f5f4f3f2f1f
	for n := 0; n < b.N; n++ {
		for i := 0; i < 65; i++ {
			p[i] = genREV16D(ref)
		}
	}
}

func BenchmarkGenRev16W(b *testing.B) {
	p := make([]uint32, 65)
	var ref uint32 = 0x4f3f2f1f
	for n := 0; n < b.N; n++ {
		for i := 0; i < 65; i++ {
			p[i] = genREV16W(ref)
		}
	}
}
