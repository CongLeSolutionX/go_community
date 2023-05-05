// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Windows UTF-16 strings can contain unpaired surrogates, which can't be
// decoded into a valid UTF-8 string. This file defines a set of functions
// that can be used to encode and decode potentially ill-formed UTF-16 strings
// by using the [the WTF-8 encoding](https://simonsapin.github.io/wtf-8/).
//
// WTF-8 is a strict superset of UTF-8 superset, i.e. any string that is
// well-formed in UTF-8 is also well-formed in WTF-8 and the content
// is unchanged. Also, the conversion never fails and is lossless.
//
// The benefit of using WTF-8 instead of UTF-8 when decoding a UTF-16 is
// that the conversion is lossless even for ill-formed UTF-16 strings.
// This property allows to read an ill-formed UTF-16 string, convert it
// to a Go string, and convert it back to the same original UTF-16 string.
//
// See go.dev/issues/59971 for more info.

package syscall

import (
	"unicode/utf16"
	"unicode/utf8"
)

// encodeWTF16 returns the potentially ill-formed
// UTF-16 encoding of s.
func encodeWTF16(s string, buf []uint16) []uint16 {
	// Go's range loop over a string decodes and expects
	// UTF-8 runes. Use a for loop instead to support
	// invalid UTF-8 runes.
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			// WTF-18 fallback.
			r, size = decodeRuneInString(s[i:])
			buf = append(buf, uint16(r))
			i += size
			continue
		}
		i += size
		buf = utf16.AppendRune(buf, r)
	}
	return buf
}

// decodeWTF16 returns the WTF-8 encoding of
// the potentially ill-formed UTF-16 s.
func decodeWTF16(s []uint16, buf []byte) []byte {
	const (
		surr1    = 0xd800
		surr2    = 0xdc00
		surr3    = 0xe000
		surrSelf = 0x10000
	)
	for i := 0; i < len(s); i++ {
		var ar rune
		switch r := s[i]; {
		case r < surr1, surr3 <= r:
			// normal rune
			ar = rune(r)
		case surr1 <= r && r < surr2 && i+1 < len(s) &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			ar = utf16.DecodeRune(rune(r), rune(s[i+1]))
			i++
		default:
			// WTF-18 fallback.
			buf = appendRune(buf, rune(r))
			continue
		}
		buf = utf8.AppendRune(buf, ar)

	}
	return buf
}

// decodeRuneInString is like utf8.DecodeRuneInString but
// also handles runes encoded in the surrogate range.
func decodeRuneInString(s string) (r rune, size int) {
	const (
		maskx = 0b00111111
		mask2 = 0b00011111
		mask3 = 0b00001111
		mask4 = 0b00000111
	)
	n := len(s)
	if n < 1 {
		return utf8.RuneError, 0
	}
	switch c := s[0]; {
	case c <= 0x7f:
		return rune(c), 1
	case n >= 2 && 0xC2 <= c && c <= 0xDF:
		return rune(c&mask2)<<6 + rune(s[1]&maskx), 2
	case n >= 3 && 0xE0 <= c && c <= 0xEF:
		return rune(c&mask3)<<12 + rune(s[1]&maskx)<<6 + rune(s[2]&maskx), 3
	case n >= 4 && 0xF0 <= c && c <= 0xF4:
		return rune(c&mask4)<<18 | rune(s[1]&maskx)<<12 | rune(s[2]&maskx)<<6 | rune(s[3]&maskx), 4
	}
	return utf8.RuneError, 1
}

// appendRune is like utf8.AppendRune but
// also handles runes encoded in the surrogate range.
func appendRune(p []byte, r rune) []byte {
	const (
		tx = 0b10000000
		t2 = 0b11000000
		t3 = 0b11100000
		t4 = 0b11110000

		maskx = 0b00111111

		rune1Max = 1<<7 - 1
		rune2Max = 1<<11 - 1
		rune3Max = 1<<16 - 1
	)
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); {
	case i <= rune1Max:
		return append(p, byte(r))
	case i <= rune2Max:
		return append(p, t2|byte(r>>6), tx|byte(r)&maskx)
	case i <= rune3Max:
		return append(p, t3|byte(r>>12), tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	default:
		return append(p, t4|byte(r>>18), tx|byte(r>>12)&maskx, tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	}
}
