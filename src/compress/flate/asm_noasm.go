// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !amd64

package flate

var useSSE42 = false

// crc32sse should never be called.
func crc32sse(a []byte) uint32 {
	panic("no assembler")
}

// crc32sseAll should never be called.
func crc32sseAll(a []byte, dst []uint32) {
	panic("no assembler")
}

// matchLenSSE4 should never be called.
func matchLenSSE4(a, b []byte, max int) int {
	panic("no assembler")
}

// histogram accumulates a histogram of b in h.
// h must be at least 256 entries in length,
// and must be cleared before calling this function.
func histogram(b []byte, h []int32) {
	for _, t := range b {
		h[t]++
	}
}
