// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bigmod

import "math/bits"

const (
	// _W is the number of bits we use for our limbs.
	_W = bits.UintSize - 1
	// _MASK selects _W bits from a full machine word.
	_MASK = (1 << _W) - 1
)

// choice represents a constant-time boolean. The value of choice is always
// either 1 or 0. We use an int instead of bool in order to make decisions in
// constant time by turning it into a mask.
type choice uint

func not(c choice) choice { return 1 ^ c }

const yes = choice(1)
const no = choice(0)

// ctSelect returns x if on == 1, and y if on == 0. The execution time of this
// function does not depend on its inputs. If on is any value besides 1 or 0,
// the result is undefined.
func ctSelect(on choice, x, y uint) uint {
	// When on == 1, mask is 0b111..., otherwise mask is 0b000...
	mask := -uint(on)
	// When mask is all zeros, we just have y, otherwise, y cancels with itself.
	return y ^ (mask & (y ^ x))
}

// ctEq returns 1 if x == y, and 0 otherwise. The execution time of this
// function does not depend on its inputs.
func ctEq(x, y uint) choice {
	// If x != y, then either x - y or y - x will generate a carry.
	_, c1 := bits.Sub(x, y, 0)
	_, c2 := bits.Sub(y, x, 0)
	return not(choice(c1 | c2))
}

// ctGeq returns 1 if x >= y, and 0 otherwise. The execution time of this
// function does not depend on its inputs.
func ctGeq(x, y uint) choice {
	// If x < y, then x - y generates a carry.
	_, carry := bits.Sub(x, y, 0)
	return not(choice(carry))
}

// nat represents an arbitrary natural number
//
// Each nat has an announced length, which is the number of limbs it has stored.
// Operations on this number are allowed to leak this length, but will not leak
// any information about the values contained in those limbs.
type nat struct {
	// limbs is a little-endian representation in base 2^W with
	// W = bits.UintSize - 1. The top bit is always unset between operations.
	//
	// The top bit is left unset to optimize Montgomery multiplication, in the
	// inner loop of exponentiation. Using fully saturated limbs would leave us
	// working with 129-bit numbers on 64-bit platforms, wasting a lot of space,
	// and thus time.
	limbs []uint
}

// natFromBig creates a new natural number from a little-endian saturated uint
// slice, like those returned by (*big.Int).Bits
//
// The announced length of the resulting nat is based on the actual bit size of
// the input, ignoring leading zeroes.
func natFromBits(xLimbs []uint) *nat {
	bitSize := (len(xLimbs)-1)*bits.UintSize + bitLen(xLimbs[len(xLimbs)-1])
	requiredLimbs := (bitSize + _W - 1) / _W

	out := &nat{make([]uint, requiredLimbs)}
	outI := 0
	shift := 0
	for i := range xLimbs {
		xi := uint(xLimbs[i])
		out.limbs[outI] |= (xi << shift) & _MASK
		outI++
		if outI == requiredLimbs {
			return out
		}
		out.limbs[outI] = xi >> (_W - shift)
		shift++ // this assumes bits.UintSize - _W = 1
		if shift == _W {
			shift = 0
			outI++
		}
	}
	return out
}

// natFromBytes converts a slice of big-endian bytes into a nat.
//
// The announced length of the output depends on the length of bytes. Unlike
// big.Int, creating a nat will not remove leading zeros.
func natFromBytes(bytes []byte) *nat {
	bitSize := len(bytes) * 8
	requiredLimbs := (bitSize + _W - 1) / _W

	out := &nat{make([]uint, requiredLimbs)}
	outI := 0
	shift := 0
	for i := len(bytes) - 1; i >= 0; i-- {
		bi := bytes[i]
		out.limbs[outI] |= uint(bi) << shift
		shift += 8
		if shift >= _W {
			shift -= _W
			out.limbs[outI] &= _MASK
			outI++
			if shift > 0 {
				out.limbs[outI] = uint(bi) >> (8 - shift)
			}
		}
	}
	return out
}

// cmpEq returns 1 if x == y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *nat) cmpEq(y *nat) choice {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	equal := yes
	for i := 0; i < size; i++ {
		equal &= ctEq(xLimbs[i], yLimbs[i])
	}
	return equal
}

// cmpGeq returns 1 if x >= y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *nat) cmpGeq(y *nat) choice {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	var c uint
	for i := 0; i < size; i++ {
		c = (xLimbs[i] - yLimbs[i] - c) >> _W
	}
	// If there was a carry, then subtracting y underflowed, so
	// x is not greater than or equal to y.
	return not(choice(c))
}

// assign sets x <- y if on == 1, and does nothing otherwise.
//
// Both operands must have the same announced length.
func (x *nat) assign(on choice, y *nat) *nat {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	for i := 0; i < size; i++ {
		xLimbs[i] = ctSelect(on, yLimbs[i], xLimbs[i])
	}
	return x
}

// add computes x += y if on == 1, and does nothing otherwise. It returns the
// carry of the addition regardless of on.
//
// Both operands must have the same announced length.
func (x *nat) add(on choice, y *nat) (c uint) {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	for i := 0; i < size; i++ {
		res := xLimbs[i] + yLimbs[i] + c
		xLimbs[i] = ctSelect(on, res&_MASK, xLimbs[i])
		c = res >> _W
	}
	return
}

// sub computes x -= y if on == 1, and does nothing otherwise. It returns the
// borrow of the subtraction regardless of on.
//
// Both operands must have the same announced length.
func (x *nat) sub(on choice, y *nat) (c uint) {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	for i := 0; i < size; i++ {
		res := xLimbs[i] - yLimbs[i] - c
		xLimbs[i] = ctSelect(on, res&_MASK, xLimbs[i])
		c = res >> _W
	}
	return
}

// shiftIn calculates x = x << _W + y mod m.
//
// This assumes that x is already reduced mod m, and that y < 2^_W.
func (x *nat) shiftIn(y uint, m *Modulus) *nat {
	d := &nat{make([]uint, len(m.nat.limbs))}

	// Eliminate bounds checks in the loop.
	size := len(m.nat.limbs)
	xLimbs := x.limbs[:size]
	dLimbs := d.limbs[:size]
	mLimbs := m.nat.limbs[:size]

	// Each iteration of this loop computes x = 2x + b mod m, where b is a bit
	// from y. Effectively, it left-shifts x and adds y one bit at a time,
	// reducing it every time.
	//
	// To do the reduction, each iteration computes both 2x + b and 2x + b - m.
	// The next iteration (and finally the return line) will use either result
	// based on whether the subtraction underflowed.
	needSubtraction := no
	for i := _W - 1; i >= 0; i-- {
		carry := (y >> i) & 1
		var borrow uint
		for i := 0; i < size; i++ {
			l := ctSelect(needSubtraction, dLimbs[i], xLimbs[i])

			res := l<<1 + carry
			xLimbs[i] = res & _MASK
			carry = res >> _W

			res = xLimbs[i] - mLimbs[i] - borrow
			dLimbs[i] = res & _MASK
			borrow = res >> _W
		}
		// See modAdd for how carry (aka overflow), borrow (aka underflow), and
		// needSubtraction relate.
		needSubtraction = ctEq(carry, borrow)
	}
	return x.assign(needSubtraction, d)
}

// mod calculates out = x mod m.
//
// This works regardless how large the value of x is.
//
// The output will be resized to the size of m and overwritten.
func (out *nat) mod(x *nat, m *Modulus) *nat {
	zeroLimbs(out.limbs)
	// Working our way from the most significant to the least significant limb,
	// we can insert each limb at the least significant position, shifting all
	// previous limbs left by _W. This way each limb will get shifted by the
	// correct number of bits. We can insert at least N - 1 limbs without
	// overflowing m. After that, we need to reduce every time we shift.
	i := len(x.limbs) - 1
	// For the first N - 1 limbs we can skip the actual shifting and position
	// them at the shifted position, which starts at min(N - 2, i).
	start := len(m.nat.limbs) - 2
	if i < start {
		start = i
	}
	for j := start; j >= 0; j-- {
		out.limbs[j] = x.limbs[i]
		i--
	}
	// We shift in the remaining limbs, reducing modulo m each time.
	for i >= 0 {
		out.shiftIn(x.limbs[i], m)
		i--
	}
	return out
}

// montgomeryRepresentation calculates x = x * R mod m, with R = 2^(_W * n) and
// n = len(m.nat.limbs).
//
// Faster Montgomery multiplication replaces standard modular multiplication for
// numbers in this representation.
//
// This assumes that x is already reduced mod m.
func (x *nat) montgomeryRepresentation(m *Modulus) *nat {
	for i := 0; i < len(m.nat.limbs); i++ {
		x.shiftIn(0, m) // x = x * 2^_W mod m
	}
	return x
}
