// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bits

import "testing"

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
					t.Fatalf("LeadingZeros8(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := LeadingZeros16(uint16(x))
				want := nlz - k + (16 - 8)
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("LeadingZeros16(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := LeadingZeros32(uint32(x))
				want := nlz - k + (32 - 8)
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("LeadingZeros32(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := LeadingZeros64(uint64(x))
				want := nlz - k + (64 - 8)
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("LeadingZeros64(%d) == %d; want %d", x, got, want)
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
					t.Fatalf("TrailingZeros8(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := TrailingZeros16(uint16(x))
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("TrailingZeros16(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := TrailingZeros32(uint32(x))
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("TrailingZeros32(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := TrailingZeros64(uint64(x))
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("TrailingZeros64(%d) == %d; want %d", x, got, want)
				}
			}
		}
	}
}

func TestOnes(t *testing.T) {
	for i := 0; i < 256; i++ {
		want := tab[i].pop
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			if x <= 1<<8-1 {
				got := Ones8(uint8(x))
				if got != want {
					t.Fatalf("Ones8(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := Ones16(uint16(x))
				if got != want {
					t.Fatalf("Ones16(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := Ones32(uint32(x))
				if got != want {
					t.Fatalf("Ones32(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := Ones64(uint64(x))
				if got != want {
					t.Fatalf("Ones64(%d) == %d; want %d", x, got, want)
				}
			}
		}
	}

	// TODO(gri) this could use some more tests
}

func TestRotateLeft(t *testing.T) {
	// TODO(gri) implement this
}

func TestRotateRight(t *testing.T) {
	// TODO(gri) implement this
}

func TestReverse(t *testing.T) {
	// TODO(gri) implement this
}

func TestSwapBytes(t *testing.T) {
	// TODO(gri) implement this
}

func TestLog(t *testing.T) {
	for i := 0; i < 256; i++ {
		log := tab[i].log
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := log + k
			if x <= 1<<8-1 {
				got := Log8(uint8(x))
				if x == 0 {
					want = -1
				}
				if got != want {
					t.Fatalf("Log8(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := Log16(uint16(x))
				if x == 0 {
					want = -1
				}
				if got != want {
					t.Fatalf("Log16(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := Log32(uint32(x))
				if x == 0 {
					want = -1
				}
				if got != want {
					t.Fatalf("Log32(%d) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<64-1 {
				got := Log64(uint64(x))
				if x == 0 {
					want = -1
				}
				if got != want {
					t.Fatalf("Log64(%d) == %d; want %d", x, got, want)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Testing support

type entry = struct {
	nlz, ntz, pop, log int
}

// tab contains results for all uint8 values
var tab [256]entry

func init() {
	tab[0] = entry{8, 8, 0, -1}
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

		// log
		x = i // x != 0
		n = 0
		for x >= 1<<uint(n) {
			n++
		}
		// 2**(n-1) <= x < 2**n
		tab[i].log = n - 1

	}
}
