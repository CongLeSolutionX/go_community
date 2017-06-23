// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

func countByte(s []byte, c byte) int // ../runtime/asm_arm64.s

// primeRK is the prime base used in Rabin-Karp algorithm.
const primeRK = 16777619

// 8 bytes can be completely loaded into 1 register.
const shortStringLen = 8

// hashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func hashStr(sep []byte) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

//go:noescape
func indexShortStr(s, sep []byte) int

// Index returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func Index(s, sep []byte) int {
	n := len(sep)
	last := 0
	switch {
	case n == 0:
		return 0
	case n == 1:
		return IndexByte(s, sep[0])
	case n == len(s):
		if Equal(sep, s) {
			return 0
		}
		return -1
	case n > len(s):
		return -1

	case n <= shortStringLen:
		// Use brute force when both s and sep are small
		if len(s) < 32 {
			return indexShortStr(s, sep)
		}
		c := sep[0]
		i := 0
		t := s[:len(s)-n+1]
		tLen := len(t)
		fails := 0
		for i < tLen {
			if t[i] != c {
				// IndexByte 32 bytes per iteration,
				// so it's faster than indexShortStr.
				o := IndexByte(t[i:], c)
				if o < 0 {
					return -1
				}
				i += o
			}
			if Equal(s[i:i+n], sep) {
				return i
			}
			fails++
			i++
			// Switch to Rabin-Karp search when IndexByte produces too many false positives.
			// Too many means more that 1 error per 8 characters.
			// Allow some errors in the beginning.
			if fails > (i+16)/8 {
				if i == tLen {
					return -1
				}
				last = i
				goto rabin_karp_search
			}
		}
		return -1

	}
rabin_karp_search:
	// Rabin-Karp search
	hashsep, pow := hashStr(sep)
	var h uint32
	for i := last; i < n+last; i++ {
		h = h*primeRK + uint32(s[i])
	}
	if h == hashsep && Equal(s[:n], sep) {
		return last
	}
	for i := n + last; i < len(s); {
		h *= primeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashsep && Equal(s[i-n:i], sep) {
			return i - n
		}
	}
	return -1
}

// Count counts the number of non-overlapping instances of sep in s.
// If sep is an empty slice, Count returns 1 + the number of Unicode code points in s.
func Count(s, sep []byte) int {
	if len(sep) == 1 {
		return countByte(s, sep[0])
	}
	return countGeneric(s, sep)
}
