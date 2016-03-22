// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flate

// Maximum deflate offset
const maxOffset = 1 << 15

// emitLiteral writes a literal chunk and returns the number of bytes written.
func emitLiteral(dst []token, lit []byte) []token {
	for _, v := range lit {
		dst = append(dst, token(v))
	}
	return dst
}

type fastEnc interface {
	Encode([]token, []byte) []token
	Reset()
}

func newFastEnc(level int) fastEnc {
	e := &encState{}
	switch level {
	case 1:
		e.enc = e.encodeL1
	case 2:
		e.enc = e.encodeL2
	case 3:
		e.enc = e.encodeL3
	default:
		panic("invalid level specified")
	}
	return e
}

const tableBits = 14             // Bits used in the table
const tableSize = 1 << tableBits // Size of the table

// encState maintains the table for matches,
// and the previous byte block for level 2.
// This is the generic implementation.
type encState struct {
	table [tableSize]int64
	block [maxStoreBlockSize]byte
	prev  []byte
	cur   int
	enc   func(dst []token, src []byte) []token
}

func (e *encState) Encode(dst []token, src []byte) []token {
	return e.enc(dst, src)
}

// encodeL1 uses simple LZ77 matching similar to Snappy,
// but also does Huffman entropy encoding.
// Level 1 does not attempt to match across block boundaries.
func (e *encState) encodeL1(dst []token, src []byte) []token {
	// Return early if src is short.
	if len(src) <= 4 {
		if len(src) != 0 {
			dst = emitLiteral(dst, src)
		}
		e.cur += 4
		return dst
	}

	// Ensure that e.cur doesn't wrap, mainly an issue on 32 bits.
	if e.cur > 1<<30 {
		e.cur = 0
	}

	// Iterate over the source bytes.
	var (
		s   int // The iterator position.
		t   int // The last position with the same hash as s.
		lit int // The start position of any pending literal bytes.
	)

	for s+3 < len(src) {
		// Update the hash table.
		b0, b1, b2, b3 := src[s], src[s+1], src[s+2], src[s+3]
		h := uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
		p := &e.table[(h*0x1e35a7bd)>>(32-tableBits)]
		// We need to to store values in [-1, inf) in table. To save
		// some initialization time, (re)use the table's zero value
		// and shift the values against this zero: add 1 on writes,
		// subtract 1 on reads.
		t, *p = int(*p)-1-e.cur, int64(s+1+e.cur)

		offset := uint(s - t - 1)

		// If t is invalid or src[s:s+4] differs from src[t:t+4], accumulate a literal byte.
		if t < 0 || offset >= (maxOffset-1) || b0 != src[t] || b1 != src[t+1] || b2 != src[t+2] || b3 != src[t+3] {
			// Skip 1 byte for 16 consecutive missed.
			s += 1 + ((s - lit) >> 4)
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
		// Emit the copied bytes.
		dst = append(dst, matchToken(uint32(s-s0-3), uint32(s-t-minOffsetSize)))
		lit = s
	}

	// Emit any final pending literal bytes and return.
	if lit != len(src) {
		dst = emitLiteral(dst, src[lit:])
	}
	e.cur += len(src)
	return dst
}

// encodeL2 uses a similar algorithm to level 1, but is capable
// of matching across blocks giving better compression at a small slowdown.
func (e *encState) encodeL2(dst []token, src []byte) []token {
	// Return early if src is short.
	if len(src) <= 4 {
		if len(src) != 0 {
			dst = emitLiteral(dst, src)
		}
		e.prev = nil
		e.cur += len(src)
		return dst
	}

	// Ensure that e.cur doesn't wrap, mainly an issue if int is 32 bits.
	if e.cur > 1<<30 {
		e.cur = 0
	}

	// Iterate over the source bytes.
	var (
		s   int // The iterator position.
		t   int // The last position with the same hash as s.
		lit int // The start position of any pending literal bytes.
	)

	for s+3 < len(src) {
		// Update the hash table.
		b0, b1, b2, b3 := src[s], src[s+1], src[s+2], src[s+3]
		h := uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
		p := &e.table[(h*0x1e35a7bd)>>(32-tableBits)]
		// We need to to store values in [-1, inf) in table. To save
		// some initialization time, (re)use the table's zero value
		// and shift the values against this zero: add 1 on writes,
		// subtract 1 on reads.
		t, *p = int(*p)-1-e.cur, int64(s+1+e.cur)

		// If t is positive, the match starts in the current block
		if t >= 0 {

			offset := uint(s - t - 1)
			// Check that the offset is valid and that we match at least 4 bytes
			if offset >= (maxOffset-1) || b0 != src[t] || b1 != src[t+1] || b2 != src[t+2] || b3 != src[t+3] {
				// Skip 1 byte for 32 consecutive missed.
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
		if tp < 0 || t > -5 || s-t >= maxOffset || b0 != e.prev[tp] || b1 != e.prev[tp+1] || b2 != e.prev[tp+2] || b3 != e.prev[tp+3] {
			// Skip 1 byte for 32 consecutive missed.
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
				goto l
			}
		}
	l:
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
		e.prev = e.block[:len(src)]
	} else {
		e.prev = nil
	}
	return dst
}

// encodeL3 uses a similar algorithm to level 2, but it
// will keep two matches per hash.
// Both hashes are checked if the first isn't ok, and the longest is selected.
func (e *encState) encodeL3(dst []token, src []byte) []token {
	// Return early if src is short.
	if len(src) <= 4 {
		if len(src) != 0 {
			dst = emitLiteral(dst, src)
		}
		e.prev = nil
		e.cur += len(src)
		return dst
	}

	// Ensure that e.cur doesn't wrap, mainly an issue on 32 bits.
	if e.cur > 1<<30 {
		e.cur = 0
	}

	// Iterate over the source bytes.
	var (
		s   int // The iterator position.
		lit int // The start position of any pending literal bytes.
	)

	for s+3 < len(src) {
		// Update the hash table.
		h := uint32(src[s]) | uint32(src[s+1])<<8 | uint32(src[s+2])<<16 | uint32(src[s+3])<<24
		p := &e.table[(h*0x1e35a7bd)>>(32-tableBits)]
		tmp := *p
		p1 := int(tmp & 0xffffffff) // Closest match position
		p2 := int(tmp >> 32)        // Furthest match position

		// We need to to store values in [-1, inf) in table. To save
		// some initialization time, (re)use the table's zero value
		// and shift the values against this zero: add 1 on writes,
		// subtract 1 on reads.
		t1 := p1 - 1 - e.cur

		var l2 int
		var t2 int
		l1 := e.matchlen(s, t1, src)
		// If fist match was ok, don't do the second.
		if l1 < 16 {
			t2 = p2 - 1 - e.cur
			l2 = e.matchlen(s, t2, src)

			// If both are short, continue
			if l1 < 4 && l2 < 4 {
				// Update hash table
				*p = int64(s+1+e.cur) | (int64(p1) << 32)
				// Skip 1 byte for 32 consecutive missed.
				s += 1 + ((s - lit) >> 5)
				continue
			}
		}

		// Otherwise, we have a match. First, emit any pending literal bytes.
		if lit != s {
			dst = emitLiteral(dst, src[lit:s])
		}
		// Update hash table
		*p = int64(s+1+e.cur) | (int64(p1) << 32)

		// Store the longest match l1 will be closest, so we prefer that if equal length
		if l1 >= l2 {
			dst = append(dst, matchToken(uint32(l1-3), uint32(s-t1-minOffsetSize)))
			s += l1
		} else {
			dst = append(dst, matchToken(uint32(l2-3), uint32(s-t2-minOffsetSize)))
			s += l2
		}
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
		e.prev = e.block[:len(src)]
	} else {
		e.prev = nil
	}
	return dst
}

// matchlen returns the match length of src[s:] and src[t:].
// If t is negative, the match may be in the previous block,
// and that will be checked if available.
func (e *encState) matchlen(s, t int, src []byte) int {
	offset := uint(s - t - 1)

	// Check if we are inside the current block.
	if t >= 0 {
		if offset >= (maxOffset-1) ||
			src[s] != src[t] || src[s+1] != src[t+1] ||
			src[s+2] != src[t+2] || src[s+3] != src[t+3] {
			return 0
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
		return s - s0
	}

	// We have a potential match in the previous block.
	tp := len(e.prev) + t
	// We check in the order most likely to least likely to be true.
	if tp < 0 || offset >= (maxOffset-1) || t > -5 ||
		src[s] != e.prev[tp] || src[s+1] != e.prev[tp+1] ||
		src[s+2] != e.prev[tp+2] || src[s+3] != e.prev[tp+3] {
		return 0
	}

	// Extend the match to be as long as possible.
	s0 := s
	s1 := s + maxMatchLength
	if s1 > len(src) {
		s1 = len(src)
	}
	s, tp = s+4, tp+4
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
			return s - s0
		}
	}
	return s - s0
}

// Reset the encoding state.
func (e *encState) Reset() {
	e.prev = nil
}
