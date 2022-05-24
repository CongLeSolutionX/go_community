// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import "unsafe"

// CommonPrefixLen returns the length of the common prefix of two strings.
func CommonPrefixLen(a, b string) int {
	commonLen := min(len(a), len(b))
	i := 0

	// Optimization: load and compare word-sized chunks at a time.
	// This is about 6x faster than the naive approach when len > 64.
	//
	// TODO(adonovan): further optimizations are possible,
	// at the cost of portability, for example by:
	// - better elimination of bounds checks;
	// - use of uint64 instead of an array may result in better
	//   registerization, and allows computing the final portion
	//   from the bitmask:
	//
	// 	cmp := load64le(a, i) ^ load64le(b, i)
	// 	if cmp != 0 {
	// 		return i + bits.LeadingZeros64(cmp)/8
	// 	}
	//
	// - use of vector instructions in the manner of
	//   runtime.cmpstring, which is expected to achieve 3x
	//   further improvement when len > 32.
	//
	const wordsize = int(unsafe.Sizeof(uint(0)))
	var aword, bword [wordsize]byte
	for i+wordsize <= commonLen {
		copy(aword[:], a[i:i+wordsize])
		copy(bword[:], b[i:i+wordsize])
		if aword != bword {
			break
		}
		i += wordsize
	}

	// naive implementation
	for i < commonLen {
		if a[i] != b[i] {
			return i
		}
		i++
	}

	return i
}

func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

// CommonPrefix returns the common prefix of two strings.
//
// REVIEWERS: this API function is redundant wrt CommonPrefixLen, and
// the latter one has the merit that it can exist in the same form in
// the bytes package, whereas bytes.CommonPrefix would have to choose
// which of a or b should be aliased by the result. My preference
// would be to delete it.
func CommonPrefix(a, b string) string {
	return a[:CommonPrefixLen(a, b)]
}
