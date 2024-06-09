// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestBucketSort32(t *testing.T) {
	const iters = 32
	const bufLen = 4 << 10

	src := make([]uint32, bufLen)
	dst := make([]uint32, bufLen)

	rnd := rand.New(rand.NewPCG(0, 0))

	for range iters {
		// Generate random source data.
		for i := range src {
			src[i] = rnd.Uint32()
		}
		// Generate a random shift.
		shift := rnd.UintN(31)

		// Do the bucketing.
		var counts [radixBase]int
		bucketSort(src, dst, &counts, shift)

		// Check the result.
		t.Logf("counts: %v", counts)
		pos := 0
		for digit, count := range counts {
			sub := dst[pos : pos+count]
			for i, val := range sub {
				got := (val >> shift) % radixBase
				if got != uint32(digit) {
					t.Fatalf("dst[%d]=%d, got digit %d, want %d", pos+i, val, got, digit)
				}
			}
			pos += count
		}
	}
}

func BenchmarkBucketSort(b *testing.B) {
	var bufBytes []int
	for i := 64; i <= 32<<10; i *= 2 {
		bufBytes = append(bufBytes, i)
	}
	// We can't really change the radix because it's a const, but we can do an
	// okay job of simulating it in how we generate the input data.
	var radixes []int
	for i := radixBase; i >= 4; i /= 2 {
		radixes = append(radixes, i)
	}

	b.Run("type=uint32", func(b *testing.B) {
		for _, radix := range radixes {
			b.Run(fmt.Sprintf("radix=%d", radix), func(b *testing.B) {
				for _, n := range bufBytes {
					// Generate data.
					src := make([]uint32, n/4)
					dst := make([]uint32, n/4)
					rnd := rand.New(rand.NewPCG(0, 0))
					for i := range src {
						src[i] = rnd.Uint32N(uint32(radix))
					}
					b.Run(fmt.Sprintf("bytes=%d", n), func(b *testing.B) {
						b.SetBytes(int64(n))
						for range b.N {
							var counts [radixBase]int
							bucketSort(src, dst, &counts, 0)
						}
					})
				}
			})
		}
	})
	b.Run("type=uint64", func(b *testing.B) {
		for _, radix := range radixes {
			b.Run(fmt.Sprintf("radix=%d", radix), func(b *testing.B) {
				for _, n := range bufBytes {
					// Generate data.
					src := make([]uint64, n/8)
					dst := make([]uint64, n/8)
					rnd := rand.New(rand.NewPCG(0, 0))
					for i := range src {
						src[i] = rnd.Uint64N(uint64(radix))
					}
					b.Run(fmt.Sprintf("bytes=%d", n), func(b *testing.B) {
						b.SetBytes(int64(n))
						for range b.N {
							var counts [radixBase]int
							bucketSort(src, dst, &counts, 0)
						}
					})
				}
			})
		}
	})
}

func BenchmarkSliceSort(b *testing.B) {
	var bufBytes []int
	for i := 64; i <= 32<<10; i *= 2 {
		bufBytes = append(bufBytes, i)
	}

	b.Run("type=uint32", func(b *testing.B) {
		for _, n := range bufBytes {
			// Generate data.
			src := make([]uint32, n/4)
			dst := make([]uint32, n/4)
			rnd := rand.New(rand.NewPCG(0, 0))
			for i := range src {
				src[i] = rnd.Uint32()
			}
			b.Run(fmt.Sprintf("bytes=%d", n), func(b *testing.B) {
				b.SetBytes(int64(n))
				for range b.N {
					copy(dst, src)
					slices.Sort(dst)
				}
			})
		}
	})
}
