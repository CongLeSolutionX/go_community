// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bits

import (
	"testing"
	"unsafe"
)

func TestUintSize(t *testing.T) {
	var x uint
	if want := unsafe.Sizeof(x) * 8; UintSize != want {
		t.Fatalf("UintSize = %d; want %d", UintSize, want)
	}
}

func TestLeadingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		nlz := tab[i].nlz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			if x <= 1<<8-1 {
				got := LeadingZeros8(uint8(x))
				want := nlz - k + (8 - 8)
				if x == 0 {
					want = 8
				}
				if got != want {
					t.Fatalf("LeadingZeros8(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := LeadingZeros16(uint16(x))
				want := nlz - k + (16 - 8)
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("LeadingZeros16(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := LeadingZeros32(uint32(x))
				want := nlz - k + (32 - 8)
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("LeadingZeros32(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := LeadingZeros64(uint64(x))
				want := nlz - k + (64 - 8)
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("LeadingZeros64(%#016x) == %d; want %d", x, got, want)
				}
			}
		}
	}
}

func TestTrailingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		ntz := tab[i].ntz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := ntz + k
			if x <= 1<<8-1 {
				got := TrailingZeros8(uint8(x))
				if x == 0 {
					want = 8
				}
				if got != want {
					t.Fatalf("TrailingZeros8(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := TrailingZeros16(uint16(x))
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("TrailingZeros16(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := TrailingZeros32(uint32(x))
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("TrailingZeros32(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := TrailingZeros64(uint64(x))
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("TrailingZeros64(%#016x) == %d; want %d", x, got, want)
				}
			}
		}
	}
}

func TestPopCount(t *testing.T) {
	for i := 0; i < 256; i++ {
		want := tab[i].pop
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			if x <= 1<<8-1 {
				got := PopCount8(uint8(x))
				if got != want {
					t.Fatalf("PopCount8(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := PopCount16(uint16(x))
				if got != want {
					t.Fatalf("PopCount16(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := PopCount32(uint32(x))
				if got != want {
					t.Fatalf("PopCount32(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := PopCount64(uint64(x))
				if got != want {
					t.Fatalf("PopCount64(%#016x) == %d; want %d", x, got, want)
				}
			}
		}
	}

	// TODO(gri) this could use some more tests
}

func TestRotateLeft(t *testing.T) {
	for i := 0; i < 256; i++ {
		x := uint8(i)
		for k := uint(0); k < 64; k++ {
			got := RotateLeft8(x, k)
			want := x<<(k&0x7) | x>>(8-k&0x7)
			if got != want {
				t.Fatalf("RotateLeft8(%#02x, %d) == %#02x; want %#02x", x, k, got, want)
			}
		}
	}

	// TODO(gri) complete this
}

func TestRotateRight(t *testing.T) {
	for i := 0; i < 256; i++ {
		x := uint8(i)
		for k := uint(0); k < 64; k++ {
			got := RotateRight8(x, k)
			want := x>>(k&0x7) | x<<(8-k&0x7)
			if got != want {
				t.Fatalf("RotateRight8(%#02x, %d) == %#02x; want %#02x", x, k, got, want)
			}
		}
	}

	// TODO(gri) complete this
}

func TestReverse(t *testing.T) {
	// test each bit
	for i := uint(0); i < 64; i++ {
		testReverse(t, uint64(1)<<i, uint64(1)<<(63-i))
	}

	// test a few patterns
	for _, test := range []struct {
		x, r uint64
	}{
		{0, 0},
		{0x1, 0x8 << 60},
		{0x2, 0x4 << 60},
		{0x3, 0xc << 60},
		{0x4, 0x2 << 60},
		{0x5, 0xa << 60},
		{0x6, 0x6 << 60},
		{0x7, 0xe << 60},
		{0x8, 0x1 << 60},
		{0x9, 0x9 << 60},
		{0xa, 0x5 << 60},
		{0xb, 0xd << 60},
		{0xc, 0x3 << 60},
		{0xd, 0xb << 60},
		{0xe, 0x7 << 60},
		{0xf, 0xf << 60},
		{0x5686487, 0xe12616a000000000},
		{0x0123456789abcdef, 0xf7b3d591e6a2c480},
	} {
		testReverse(t, test.x, test.r)
		testReverse(t, test.r, test.x)
	}
}

func testReverse(t *testing.T, x64, want64 uint64) {
	x8 := uint8(x64)
	got8 := Reverse8(x8)
	want8 := uint8(want64 >> (64 - 8))
	if got8 != want8 {
		t.Fatalf("Reverse8(%#02x) == %#02x; want %#02x", x8, got8, want8)
	}

	x16 := uint16(x64)
	got16 := Reverse16(x16)
	want16 := uint16(want64 >> (64 - 16))
	if got16 != want16 {
		t.Fatalf("Reverse16(%#04x) == %#04x; want %#04x", x16, got16, want16)
	}

	x32 := uint32(x64)
	got32 := Reverse32(x32)
	want32 := uint32(want64 >> (64 - 32))
	if got32 != want32 {
		t.Fatalf("Reverse32(%#08x) == %#08x; want %#08x", x32, got32, want32)
	}

	got64 := Reverse64(x64)
	if got64 != want64 {
		t.Fatalf("Reverse64(%#016x) == %#016x; want %#016x", x64, got64, want64)
	}
}

func TestReverseBytes(t *testing.T) {
	for _, test := range []struct {
		x, r uint64
	}{
		{0, 0},
		{0x01, 0x01 << 56},
		{0x0123, 0x2301 << 48},
		{0x012345, 0x452301 << 40},
		{0x01234567, 0x67452301 << 32},
		{0x0123456789, 0x8967452301 << 24},
		{0x0123456789ab, 0xab8967452301 << 16},
		{0x0123456789abcd, 0xcdab8967452301 << 8},
		{0x0123456789abcdef, 0xefcdab8967452301 << 0},
	} {
		testReverseBytes(t, test.x, test.r)
		testReverseBytes(t, test.r, test.x)
	}
}

func testReverseBytes(t *testing.T, x64, want64 uint64) {
	x16 := uint16(x64)
	got16 := ReverseBytes16(x16)
	want16 := uint16(want64 >> (64 - 16))
	if got16 != want16 {
		t.Fatalf("ReverseBytes16(%#04x) == %#04x; want %#04x", x16, got16, want16)
	}

	x32 := uint32(x64)
	got32 := ReverseBytes32(x32)
	want32 := uint32(want64 >> (64 - 32))
	if got32 != want32 {
		t.Fatalf("ReverseBytes32(%#08x) == %#08x; want %#08x", x32, got32, want32)
	}

	got64 := ReverseBytes64(x64)
	if got64 != want64 {
		t.Fatalf("ReverseBytes64(%#016x) == %#016x; want %#016x", x64, got64, want64)
	}
}

func TestLen(t *testing.T) {
	for i := 0; i < 256; i++ {
		len := tab[i].len
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := 0
			if x != 0 {
				want = len + k
			}
			if x <= 1<<8-1 {
				got := Len8(uint8(x))
				if got != want {
					t.Fatalf("Len8(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := Len16(uint16(x))
				if got != want {
					t.Fatalf("Len16(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := Len32(uint32(x))
				if got != want {
					t.Fatalf("Len32(%#016x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := Len64(uint64(x))
				if got != want {
					t.Fatalf("Len64(%#016x) == %d; want %d", x, got, want)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Testing support

type entry = struct {
	nlz, ntz, pop, len int
}

// tab contains results for all uint8 values
var tab [256]entry

func init() {
	tab[0] = entry{8, 8, 0, 0}
	for i := 1; i < len(tab); i++ {
		// nlz
		x := i // x != 0
		n := 0
		for x&0x80 == 0 {
			n++
			x <<= 1
		}
		tab[i].nlz = n

		// ntz
		x = i // x != 0
		n = 0
		for x&1 == 0 {
			n++
			x >>= 1
		}
		tab[i].ntz = n

		// pop
		x = i // x != 0
		n = 0
		for x != 0 {
			n += int(x & 1)
			x >>= 1
		}
		tab[i].pop = n

		// len
		x = i // x != 0
		n = 0
		for x >= 1<<uint(n) {
			n++
		}
		// 2**(n-1) <= x < 2**n
		tab[i].len = n

	}
}
