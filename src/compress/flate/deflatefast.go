// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flate

const maxOffset = 1 << logMaxOffsetSize // Maximum deflate offset
const tableBits = 16                    // Bits used in the table
const tableSize = 1 << tableBits        // Size of the table

func newEncState() *encState {
	return &encState{cur: 1}
}

// encState maintains the table for matches,
// and the previous byte block for level 2.
// This is the generic implementation.
type encState struct {
	table [tableSize]uint32
	block [maxStoreBlockSize]byte
	prev  []byte
	cur   int
}

func (e *encState) Encode(dst []token, src []byte) []token {
	// Return early if src is short.
	if len(src) <= 4 {
		if len(src) != 0 {
			dst = emitLiteral(dst, src)
		}
		e.cur += len(src)
		e.prev = nil
		return dst
	}

	// Ensure that e.cur doesn't wrap, mainly an issue on 32 bits.
	if e.cur > 1<<30 {
		e.cur = 1
	}

	var s int   // The iterator position.
	var t int   // The last position with the same hash as s.
	var lit int // The start position of any pending literal bytes.

	for s+3 < len(src) {
		srcS := src[s : s+4]
		// Update the hash table.
		h := uint32(srcS[0]) | uint32(srcS[1])<<8 | uint32(srcS[2])<<16 | uint32(srcS[3])<<24
		p := &e.table[(h*0x1e35a7bd)>>(32-tableBits)]
		// We need to to store values in [-1, inf) in table.
		// To save some initialization time, we make sure that
		// e.cur is always > 0.
		t, *p = int(*p)-e.cur, uint32(s+e.cur)

		// If t is positive, the match starts in the current block
		if t >= 0 {
			offset := uint(s - t - 1)
			// Check that the offset is valid and that we match at least 4 bytes
			if offset >= (maxOffset - 1) {
				s += 1 + ((s - lit) >> 5)
				continue
			}
			srcT := src[t : t+4]
			if srcS[0] != srcT[0] || srcS[1] != srcT[1] || srcS[2] != srcT[2] || srcS[3] != srcT[3] {
				s += 1 + ((s - lit) >> 5)
				continue
			}
			// Otherwise, we have a match. First, emit any pending literal bytes.
			if lit != s {
				dst = emitLiteral(dst, src[lit:s])
			}
			// Extend the match to be as long as possible.
			s0 := s
			s1 := s + maxMatchLength
			if s1 > len(src) {
				s1 = len(src)
			}
			s, t = s+4, t+4
			for s < s1 && src[s] == src[t] {
				s++
				t++
			}
			// Emit the matched bytes.
			dst = append(dst, matchToken(uint32(s-s0-3), uint32(s-t-minOffsetSize)))
			lit = s
			continue
		}
		// We found a match in the previous block.
		tp := len(e.prev) + t
		if tp < 0 || t > -5 || s-t >= maxOffset {
			s += 1 + ((s - lit) >> 5)
			continue
		}
		ep := e.prev[tp : tp+4]
		if srcS[0] != ep[0] || srcS[1] != ep[1] || srcS[2] != ep[2] || srcS[3] != ep[3] {
			s += 1 + ((s - lit) >> 5)
			continue
		}
		// Otherwise, we have a match. First, emit any pending literal bytes.
		if lit != s {
			dst = emitLiteral(dst, src[lit:s])
		}
		// Extend the match to be as long as possible.
		s0 := s
		s1 := s + maxMatchLength
		if s1 > len(src) {
			s1 = len(src)
		}
		s, tp = s+4, tp+4
	loop:
		for s < s1 && src[s] == e.prev[tp] {
			s++
			tp++
			if tp == len(e.prev) {
				t = 0
				// continue in current buffer
				for s < s1 && src[s] == src[t] {
					s++
					t++
				}
				break loop
			}
		}
		// Emit the copied bytes.
		if t < 0 {
			t = tp - len(e.prev)
		}
		dst = append(dst, matchToken(uint32(s-s0-3), uint32(s-t-minOffsetSize)))
		lit = s
	}

	// Emit any final pending literal bytes and return.
	if lit != len(src) {
		dst = emitLiteral(dst, src[lit:])
	}
	e.cur += len(src)

	// Store this block, if it was full length.
	if len(src) == maxStoreBlockSize {
		copy(e.block[:], src)
		e.prev = e.block[:]
	} else {
		e.prev = nil
	}

	return dst
}

// emitLiteral writes a literal chunk and returns the number of bytes written.
func emitLiteral(dst []token, lit []byte) []token {
	for _, v := range lit {
		dst = append(dst, token(v))
	}
	return dst
}

// Reset the encoding state.
func (e *encState) Reset() {
	e.prev = nil
}
