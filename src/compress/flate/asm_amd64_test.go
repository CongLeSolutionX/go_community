// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build amd64

package flate

import (
	"math/rand"
	"testing"
)

func TestCRC(t *testing.T) {
	if !useSSE42 {
		t.Skip("Skipping CRC test, no SSE 4.2 available")
	}
	for _, x := range deflateTests {
		y := x.out
		if len(y) >= 4 {
			t.Logf("In: %v, Out:0x%08x", y[:4], crc32SSE(y[:4]))
		}
	}
}

func TestCRCAll(t *testing.T) {
	if !useSSE42 {
		t.Skip("Skipping CRC test, no SSE 4.2 available")
	}
	for _, x := range deflateTests {
		y := x.out
		y = append(y, y...)
		y = append(y, y...)
		y = append(y, y...)
		y = append(y, y...)
		y = append(y, y...)
		y = append(y, y...)
		if !testing.Short() {
			y = append(y, y...)
			y = append(y, y...)
		}
		y = append(y, 1)

		for j := len(y) - 1; j >= 4; j-- {
			// Create copy, so we detect out-of-bound reads more easily.
			test1 := make([]byte, j)
			test2 := make([]byte, j)
			copy(test1, y[:j])
			copy(test2, y[:j])

			// We allocate one more than we need to test for unintentional overwrites.
			dst := make([]uint32, j-3+1)
			ref := make([]uint32, j-3+1)
			for i := range dst {
				dst[i] = uint32(i + 100)
				ref[i] = uint32(i + 101)
			}
			// Last entry must NOT be overwritten.
			dst[j-3] = 0x1234
			ref[j-3] = 0x1234

			// Do two encodes we can compare.
			crc32SSEAll(test1, dst)
			crc32SSEAll(test2, ref)

			// Check all values.
			for i, got := range dst {
				if i == j-3 {
					if dst[i] != 0x1234 {
						t.Fatalf("end of dst overwritten, was 0x%08x", dst[i])
					}
					continue
				}
				if want := crc32SSE(y[i : i+4]); got != want {
					notModified := ""
					if got == uint32(i)+100 {
						notModified = " (not modified)"
					}
					t.Errorf("len:%d index:%d, vs crc32SSE, got 0x%08x%s, want 0x%08x", len(y), i, got, notModified, want)
				}
				if want := ref[i]; got != want {
					t.Errorf("len:%d index:%d, vs ref, got 0x%08x, want 0x%08x", len(y), i, got, want)
				}
			}
		}
	}
}

func TestMatchLen(t *testing.T) {
	if !useSSE42 {
		t.Skip("Skipping Matchlen test, no SSE 4.2 available")
	}
	// Maximum length tested.
	const maxLen = 512

	// Skips per iteration.
	is, js, ks := 3, 2, 1
	if testing.Short() {
		is, js, ks = 7, 5, 3
	}

	rng := rand.New(rand.NewSource(1))
	a := make([]byte, maxLen)
	b := make([]byte, maxLen)
	for i := range a {
		a[i] = byte(rng.Int63())
		b[i] = byte(rng.Int63())
	}
	bb := make([]byte, maxLen)

	// Test different lengths.
	for i := 0; i < maxLen; i += is {
		// Test different dst offsets.
		for j := 0; j < maxLen-1; j += js {
			copy(bb, b)
			// Test different src offsets.
			for k := i - 1; k >= 0; k -= ks {
				copy(bb[j:], a[k:i])
				maxTest := maxLen - j
				if maxTest > maxLen-k {
					maxTest = maxLen - k
				}
				got := matchLenSSE4(a[k:], bb[j:], maxTest)
				want := matchLenReference(a[k:], bb[j:], maxTest)
				if got != want {
					t.Fatalf("i:%d, src offset:%d, dst offset:%d, maxTest:%d, got %d, want %d",
						i, k, j, maxTest, got, want)
				}
			}
		}
	}
}

func matchLenReference(a, b []byte, max int) int {
	for i := 0; i < max; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}

func TestHistogram(t *testing.T) {
	if !useSSE42 {
		t.Skip("Skipping Matchlen test, no SSE 4.2 available")
	}
	// Maximum length tested.
	const maxLen = 65536
	var maxOff = 8

	// Skips per iteration.
	is, js := 5, 3
	if testing.Short() {
		is, js = 9, 1
		maxOff = 1
	}

	rng := rand.New(rand.NewSource(1))
	a := make([]byte, maxLen+maxOff)
	for i := range a {
		a[i] = byte(rng.Int63())
	}

	// Test different lengths.
	for i := 0; i <= maxLen; i += is {
		// Test different offsets.
		for j := 0; j < maxOff; j += js {
			var got, want [256]int32

			histogram(a[j:i+j], got[:])
			histogramReference(a[j:i+j], want[:])
			for k := range got {
				if got[k] != want[k] {
					t.Fatalf("mismatch at len:%d, offset:%d, value %d: got %d, want %d", i, j, k, got[k], want[k])
				}
			}
		}
	}
}

func histogramReference(b []byte, h []int32) {
	if len(h) < 256 {
		panic("Histogram too small")
	}
	for _, t := range b {
		h[t]++
	}
}
