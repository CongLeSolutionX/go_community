// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/aclements/go-perfevent/events"
	"github.com/aclements/go-perfevent/perf"
)

func TestCount32(t *testing.T) {
	const iters = 32
	const bufLen = 4 << 10

	// TODO: Test for counter overflow.
	// TODO: Test buffers that aren't nice lengths.

	src := make([]LAddr32, bufLen/4)

	rnd := rand.New(rand.NewPCG(0, 0))

	for range iters {
		// Generate random source data.
		for i := range src {
			src[i] = LAddr32(rnd.Uint32())
		}
		// Generate a random shift.
		shift := rnd.UintN(31)

		// Do the counting.
		var counts [16]uint16
		count32AVX2(src, shift, &counts)

		// Reference implementation.
		var wantCounts [16]uint16
		for _, val := range src {
			wantCounts[(val>>shift)&0xF]++
		}
		if counts != wantCounts {
			t.Fatalf("want %v, got %v", wantCounts, counts)
		}
	}
}

func TestBucketSort32(t *testing.T) {
	const iters = 32
	const bufLen = 4 << 10

	src := make([]LAddr32, bufLen)
	dst := make([]LAddr32, bufLen)

	rnd := rand.New(rand.NewPCG(0, 0))

	for range iters {
		// Generate random source data.
		for i := range src {
			src[i] = LAddr32(rnd.Uint32())
		}
		// Generate a random shift.
		shift := rnd.UintN(31)

		// Do the bucketing.
		//
		// TODO: Test different implementations
		var counts [radixBase]uint16
		bucketSort(src, dst, &counts, shift)

		// Check the result.
		t.Logf("counts: %v", counts)
		pos := 0
		for digit, count := range counts {
			sub := dst[pos : pos+int(count)]
			for i, val := range sub {
				got := (val >> shift) % radixBase
				if got != LAddr32(digit) {
					t.Fatalf("dst[%d]=%d, got digit %d, want %d", pos+i, val, got, digit)
				}
			}
			pos += int(count)
		}
	}
}

var bufBytes = makeBufBytes()

func makeBufBytes() []int {
	var bufBytes []int
	for i := 64; i <= 32<<10; i *= 2 {
		bufBytes = append(bufBytes, i)
	}
	return bufBytes
}

func BenchmarkBucketSort(b *testing.B) {
	// I've experimented with some ways to speed this up.
	//
	// Computing counts as 4 uint64s split into 4 16 bit lanes using lots of
	// shifts and branch-free code. This wound up being about 3x slower.
	//
	// Computing two histograms in parallel from even/odd elements is closer,
	// but still slower than the obvious approach. To eliminate bounds and shift
	// checks, I had to unsafe cast the []LAddr32 source to []uint64, precompute
	// the larger shift, and check the two shift values before the loop.

	// We can't really change the radix because it's a const, but we can do an
	// okay job of simulating it in how we generate the input data.
	var radixes []int
	for i := radixBase; i >= 4; i /= 2 {
		radixes = append(radixes, i)
	}

	b.Run("type=uint32/impl=base", func(b *testing.B) {
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
							var counts [radixBase]uint16
							bucketSort(src, dst, &counts, 0)
						}
					})
				}
			})
		}
	})
	b.Run("type=uint32/impl=avx", func(b *testing.B) {
		for _, radix := range radixes {
			b.Run(fmt.Sprintf("radix=%d", radix), func(b *testing.B) {
				for _, n := range bufBytes {
					// Generate data.
					src := make([]LAddr32, n/4)
					dst := make([]LAddr32, n/4)
					rnd := rand.New(rand.NewPCG(0, 0))
					for i := range src {
						src[i] = LAddr32(rnd.Uint32N(uint32(radix)))
					}
					b.Run(fmt.Sprintf("bytes=%d", n), func(b *testing.B) {
						b.SetBytes(int64(n))
						for range b.N {
							var counts [radixBase]uint16
							bucketSort32AVX(src, dst, &counts, 0)
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
							var counts [radixBase]uint16
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

func BenchmarkCount32(b *testing.B) {
	gen := func(b *testing.B, run func(b *testing.B, src []LAddr32)) {
		for _, n := range bufBytes {
			// Generate data.
			src := make([]LAddr32, n/4)
			rnd := rand.New(rand.NewPCG(0, 0))
			for i := range src {
				src[i] = LAddr32(rnd.Uint32())
			}
			b.Run(fmt.Sprintf("bytes=%d", n), func(b *testing.B) {
				b.SetBytes(int64(n))

				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				counters, err := perf.OpenCounter(perf.TargetThisGoroutine, events.EventCPUCycles, events.EventTaskClock)
				if err != nil {
					b.Logf("error opening perf counters: %s", err)
				}
				defer counters.Close()
				b.ResetTimer()
				counters.Start()
				var start [2]perf.Count
				if err := counters.ReadGroup(start[:]); err != nil {
					b.Fatalf("error reading perf event: %s", err)
				}

				run(b, src)

				b.StopTimer()
				var end [2]perf.Count
				if err := counters.ReadGroup(end[:]); err != nil {
					b.Fatalf("error reading perf event: %s", err)
				}

				nPointers := len(src) * b.N
				sc, _ := start[0].Value()
				ec, _ := end[0].Value()
				cycles := ec - sc
				b.ReportMetric(float64(nPointers)/float64(cycles), "pointers/cpu-cycle")

				stc, _ := start[1].Value()
				etc, _ := end[1].Value()
				ghz := float64(cycles) / ((time.Duration(etc-stc) * time.Nanosecond).Seconds() * 1000 * 1000 * 1000)
				b.ReportMetric(ghz, "avg-GHz")
			})
		}
	}
	b.Run("impl=go", func(b *testing.B) {
		gen(b, func(b *testing.B, src []LAddr32) {
			var counts [16]uint16
			for range b.N {
				count32Go(src, 0, &counts)
			}
		})
	})
	b.Run("impl=avx2", func(b *testing.B) {
		gen(b, func(b *testing.B, src []LAddr32) {
			var counts [16]uint16
			for range b.N {
				count32AVX2(src, 0, &counts)
			}
		})
	})
}
