// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector (GC)

package sched

const (
	// Four bits per word (see #defines below).
	GcBits             = 4
	WordsPerBitmapByte = 8 / GcBits
)

const (
	// Pointer map
	BitsPerPointer  = 2
	BitsMask        = (1 << BitsPerPointer) - 1
	PointersPerByte = 8 / BitsPerPointer

	// If you change these, also change scanblock.
	// scanblock does "if(bits == BitsScalar || bits == BitsDead)" as "if(bits <= BitsScalar)".
	BitsDead          = 0
	BitsScalar        = 1                              // 01
	BitsPointer       = 2                              // 10
	BitsCheckMarkXor  = 1                              // 10
	BitsScalarMarked  = BitsScalar ^ BitsCheckMarkXor  // 00
	BitsPointerMarked = BitsPointer ^ BitsCheckMarkXor // 11

	// 64 bytes cover objects of size 1024/512 on 64/32 bits, respectively.
	MaxGCMask = 65536 // TODO(rsc): change back to 64
)

// Bits in per-word bitmap.
// #defines because we shift the values beyond 32 bits.
//
// Each word in the bitmap describes wordsPerBitmapWord words
// of heap memory.  There are 4 bitmap bits dedicated to each heap word,
// so on a 64-bit system there is one bitmap word per 16 heap words.
//
// The bitmap starts at mheap.arena_start and extends *backward* from
// there.  On a 64-bit system the off'th word in the arena is tracked by
// the off/16+1'th word before mheap.arena_start.  (On a 32-bit system,
// the only difference is that the divisor is 8.)
const (
	BitBoundary = 1 // boundary of an object
	BitMarked   = 2 // marked object
	BitMask     = BitBoundary | BitMarked
	BitPtrMask  = BitsMask << 2
)
