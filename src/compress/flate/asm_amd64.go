// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flate

// haveSSE42 returns true if SSE 4.2 is available on the CPU.
func haveSSE42() bool

var useSSE42 = haveSSE42()

// crc32SSE returns a hash for the first 4 bytes of the slice.
//
// len(a) must be >= 4.
//
//go:noescape
func crc32SSE(a []byte) uint32

// crc32SSEAll calculates hashes for each 4-byte set in a.
//
// len(dst) must be >= len(a)-3.
//
//go:noescape
func crc32SSEAll(a []byte, dst []uint32)

// matchLenSSE4 returns the number of matching bytes in a and b up to length
// max. Both slices must be at least max bytes in size.
//
// It uses the PCMPESTRI SSE 4.2 instruction.
//
//go:noescape
func matchLenSSE4(a, b []byte, max int) int

// histogram accumulates a histogram of b in h.
//
// len(h) must be >= 256, and h's elements must be all zeroes.
//
//go:noescape
func histogram(b []byte, h []int32)
