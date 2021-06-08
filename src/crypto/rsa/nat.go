// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rsa

import (
	"math/big"
	"math/bits"
)

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

// div calculates (hi:lo / d, hi:lo % d), like bits.Div. Unlike bits.Div, the
// execution time of this function does not depend on its inputs.
//
// Furthermore, this function does not panic in exceptional situations, to avoid
// leaking information. If d = 0 or d <= hi, both return values are undefined.
// Constant time selection should be used to handle these edge cases.
func div(hi, lo, d uint) (quo uint, rem uint) {
	// The rough idea is to iterate from high to low bits b,
	// and check if we can remove d << b from hi:lo.
	// If so, mark that bit of the quotient as set. Whatever value we're left
	// with after all of these subtractions is then our remainder.
	// This is similar to pen-and-paper division, one bit at a time.
	for i := bits.UintSize - 1; i >= 0; i-- {
		quo <<= 1
		j := bits.UintSize - i
		w := (hi << j) | (lo >> i)
		// If w >= d, then we can remove d. hi >> i is the bit right above the
		// MSB of w. If it's set, we should also remove d.
		sel := ctGeq(w, d) | choice(hi>>i)
		hi2 := (w - d) >> j
		lo2 := lo - (d << i)
		hi = ctSelect(sel, hi2, hi)
		lo = ctSelect(sel, lo2, lo)
		quo |= uint(sel)
	}
	rem = lo
	return
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

// expand expands x to n limbs, leaving its value unchanged.
func (x *nat) expand(n int) *nat {
	if n < len(x.limbs) {
		panic("rsa: internal error: shrinking nat")
	}
	if cap(x.limbs) < n {
		newLimbs := make([]uint, n)
		copy(newLimbs, x.limbs)
		x.limbs = newLimbs
		return x
	}
	extraLimbs := x.limbs[len(x.limbs):n]
	for i := range extraLimbs {
		extraLimbs[i] = 0
	}
	x.limbs = x.limbs[:n]
	return x
}

// reset returns a zero nat of n limbs, optionally reusing x's storage.
func (x *nat) reset(n int) *nat {
	if cap(x.limbs) < n {
		x.limbs = make([]uint, n)
		return x
	}
	for i := range x.limbs {
		x.limbs[i] = 0
	}
	x.limbs = x.limbs[:n]
	return x
}

// clone returns a new nat, with the same value and announced length as x.
func (x *nat) clone() *nat {
	out := &nat{make([]uint, len(x.limbs))}
	copy(out.limbs, x.limbs)
	return out
}

// natFromBig creates a new natural number from a big.Int.
//
// The announced length of the resulting nat is based on the actual bit size of
// the input, ignoring leading zeroes.
func natFromBig(x *big.Int) *nat {
	xLimbs := x.Bits()
	bitSize := x.BitLen()
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

// fillBytes sets bytes to x as a zero-extended big-endian byte slice.
//
// If bytes is not long enough to contain the number or at least len(x.limbs)-1
// limbs, or has zero length, fillBytes will panic.
func (x *nat) fillBytes(bytes []byte) []byte {
	if len(bytes) == 0 {
		panic("nat: fillBytes invoked with too small buffer")
	}
	for i := range bytes {
		bytes[i] = 0
	}
	shift := 0
	outI := len(bytes) - 1
	for i, limb := range x.limbs {
		remainingBits := _W
		for remainingBits >= 8 {
			bytes[outI] |= byte(limb) << shift
			consumed := 8 - shift
			limb >>= consumed
			remainingBits -= consumed
			shift = 0
			outI--
			if outI < 0 {
				if limb != 0 || i < len(x.limbs)-1 {
					panic("nat: fillBytes invoked with too small buffer")
				}
				return bytes
			}
		}
		bytes[outI] = byte(limb)
		shift = remainingBits
	}
	return bytes
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
			out.limbs[outI] = uint(bi) >> (8 - shift)
		}
	}
	return out
}

// cmpEq returns 1 if x == y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *nat) cmpEq(y *nat) choice {
	equal := choice(1)
	for i := 0; i < len(x.limbs) && i < len(y.limbs); i++ {
		equal &= ctEq(x.limbs[i], y.limbs[i])
	}
	return equal
}

// cmpGeq returns 1 if x >= y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *nat) cmpGeq(y *nat) choice {
	var c uint
	for i := 0; i < len(x.limbs) && i < len(y.limbs); i++ {
		c = (x.limbs[i] - y.limbs[i] - c) >> _W
	}
	// If there was a carry, then subtracting y underflowed, so
	// x is not greater than or equal to y.
	return not(choice(c))
}

// assign sets x <- y if on == 1, and does nothing otherwise.
//
// Both operands must have the same announced length.
func (x *nat) assign(on choice, y *nat) *nat {
	for i := 0; i < len(x.limbs) && i < len(y.limbs); i++ {
		x.limbs[i] = ctSelect(on, y.limbs[i], x.limbs[i])
	}
	return x
}

// add computes x += y if on == 1, and does nothing otherwise. It returns the
// carry of the addition regardless of on.
//
// Both operands must have the same announced length.
func (x *nat) add(on choice, y *nat) (c uint) {
	for i := 0; i < len(x.limbs) && i < len(y.limbs); i++ {
		res := x.limbs[i] + y.limbs[i] + c
		x.limbs[i] = ctSelect(on, res&_MASK, x.limbs[i])
		c = res >> _W
	}
	return
}

// sub computes x -= y if on == 1, and does nothing otherwise. It returns the
// borrow of the subtraction regardless of on.
//
// Both operands must have the same announced length.
func (x *nat) sub(on choice, y *nat) (c uint) {
	for i := 0; i < len(x.limbs) && i < len(y.limbs); i++ {
		res := x.limbs[i] - y.limbs[i] - c
		x.limbs[i] = ctSelect(on, res&_MASK, x.limbs[i])
		c = res >> _W
	}
	return
}

// mulSub calculates x -= q * m, producing a uint-sized borrow value.
//
// Both nat operands must have the same length. q may be longer than _W.
func (x *nat) mulSub(q uint, m *nat) (cc uint) {
	for i := range x.limbs {
		hi, lo := bits.Mul(q, m.limbs[i])
		// hi:lo <= (2^k - 1) * (2^w - 1) = 2^(k+w) - 2^k - 2^w + 1
		lo, cc = bits.Add(lo, cc, 0)
		hi += cc
		// hi:lo <= 2^(k+w) - 2^w
		cc = (hi << 1) | (lo >> _W)
		// cc <= 2^k - 1
		res := x.limbs[i] - (lo & _MASK)
		// If cc = 2^k - 1, then hi:lo = 2^(k+w) - 2^w, and lo & _MASK = 0,
		// so the subtraction doesn't underflow, and res >> _W = 0.
		// Otherwise, cc < 2^k - 1 and there is space for res >> _W.
		cc += res >> _W
		x.limbs[i] = res & _MASK
	}
	return
}

// modulus is used for modular arithmetic, precomputing relevant constants.
//
// Moduli are assumed to be odd numbers. Moduli can also leak the exact
// number of bits needed to store their value, and are stored without padding.
//
// Their actual value is still kept secret.
type modulus struct {
	// The underlying natural number for this modulus.
	//
	// This will be stored without any padding, and shouldn't alias with any
	// other natural number being used.
	nat     *nat
	leading int  // number of leading zeros in the modulus
	m0inv   uint // -nat.limbs[0]⁻¹ mod _W
}

// minusInverseModW computes -x⁻¹ mod _W with x odd.
//
// This operation is used to precompute a constant involved in Montgomery
// multiplication.
func minusInverseModW(x uint) uint {
	// Every iteration of this loop doubles the least-significant bits of
	// correct inverse in y. The first three bits are already correct (1⁻¹ = 1,
	// 3⁻¹ = 3, 5⁻¹ = 5, and 7⁻¹ = 7 mod 8), so doubling five times is enough
	// for 61 bits (and wastes only one iteration for 31 bits).
	//
	// See https://crypto.stackexchange.com/a/47496.
	y := x
	for i := 0; i < 5; i++ {
		y = y * (2 - x*y)
	}
	return (1 << _W) - (y & _MASK)
}

// modulusFromNat creates a new modulus from a nat.
//
// The nat should be odd, nonzero, and the number of significant bits in the
// number should be leakable. The nat shouldn't be reused.
func modulusFromNat(nat *nat) *modulus {
	m := &modulus{}
	m.nat = nat
	size := len(m.nat.limbs)
	for m.nat.limbs[size-1] == 0 {
		size--
	}
	m.nat.limbs = m.nat.limbs[:size]
	m.leading = bits.LeadingZeros(m.nat.limbs[size-1]) - (bits.UintSize - _W)
	m.m0inv = minusInverseModW(m.nat.limbs[0])
	return m
}

// modulusSize returns the size of m in bytes.
func modulusSize(m *modulus) int {
	bits := len(m.nat.limbs)*_W - int(m.leading)
	return (bits + 7) / 8
}

// shiftIn calculates x = x << _W + y mod m.
//
// This assumes that x is already reduced mod m, and that y < 2^_W.
func (x *nat) shiftIn(y uint, m *modulus) *nat { // TODO
	checkReduced(m, x)
	if y > _MASK {
		panic("rsa: internal error: shiftIn input out of bounds")
	}

	size := len(m.nat.limbs)
	if size == 1 {
		// In this case, we need to calculate x:y mod m which is what div
		// returns. div expects fully saturated limbs, though. We know d != 0
		// and d > hi because d = m.
		_, r := div(x.limbs[0]>>1, (x.limbs[0]<<_W)|y, m.nat.limbs[0])
		x.limbs[0] = r
		return x
	}

	// We want to shift y into x, and then divide by m to get the remainder. We
	// start with a good estimate, using the top 2*_W bits of x (a1:a0), and the
	// top _W bits of m (b0).

	// The actual shift: move the limbs of x up, then insert y.
	hi := x.limbs[size-1] // the top limb of x, pre-shifts
	for i := size - 1; i > 0; i-- {
		x.limbs[i] = x.limbs[i-1]
	}
	x.limbs[0] = y

	a1 := ((hi << m.leading) | (x.limbs[size-1] >> (_W - m.leading))) & _MASK
	a0 := ((x.limbs[size-1] << m.leading) | (x.limbs[size-2] >> (_W - m.leading))) & _MASK
	b0 := ((m.nat.limbs[size-1] << m.leading) | (m.nat.limbs[size-2] >> (_W - m.leading))) & _MASK

	// We want to use a1:a0 / b0 - 1 as our estimate. If that subtraction would
	// underflow, we use 0. The result can't overflow: a1 > b0 is impossible
	// because x < m, and if a1 = b0 the quotient will be 1<<_W and the
	// subtraction will bring it back in range.
	q, _ := div(a1>>1, (a1<<_W)|a0, b0)
	q = ctSelect(ctEq(q, 0), 0, q-1)

	// q is off by +- 1, so we subtract q * m, and then either add or subtract
	// m, based on the result.
	cc := x.mulSub(q, m.nat)
	// If the carry from subtraction is greater than the limb of x we've shifted out,
	// then we've underflowed, and need to add in m
	under := not(ctGeq(hi, cc))
	// For us to be too large, we first need to not be too low, as per the previous flag.
	// Then, if the lower limbs of x are still larger, or the top limb of x is equal to the carry,
	// we can conclude that we're too large, and need to subtract m
	stillBigger := x.cmpGeq(m.nat)
	over := not(under) & (stillBigger | not(ctEq(cc, hi)))
	x.add(under, m.nat)
	x.sub(over, m.nat)
	return x
}

// mod calculates out = x mod m.
//
// This works regardless how large the value of x is.
//
// The output will be resized to the size of m and overwritten.
func (out *nat) mod(x *nat, m *modulus) *nat {
	out.reset(len(m.nat.limbs))
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

// expandFor ensures out has the right size to work with operations modulo m.
//
// This assumes that out has as many or fewer limbs than m. Since modular
// operations assume that operands are exactly the right size, this allows us to
// expand a natural number to meet that expectation.
func (out *nat) expandFor(m *modulus) *nat {
	return out.expand(len(m.nat.limbs))
}

// modSub computes x = x - y mod m.
//
// The length of both operands must be the same as the modulus. Both operands
// must already be reduced modulo m.
func (x *nat) modSub(y *nat, m *modulus) *nat {
	checkReduced(m, x, y)
	underflow := x.sub(1, y)
	// If the subtraction underflowed, add m.
	x.add(choice(underflow), m.nat)
	return x
}

// modAdd computes x = x + y mod m.
//
// The length of both operands must be the same as the modulus. Both operands
// must already be reduced modulo m.
func (x *nat) modAdd(y *nat, m *modulus) *nat {
	checkReduced(m, x, y)
	overflow := x.add(1, y)
	underflow := not(x.cmpGeq(m.nat)) // x < m

	// Three cases are possible:
	//
	//   - overflow = 0, underflow = 0
	//
	// In this case, addition fits in our limbs, but we can still subtract away
	// m without an underflow, so we need to perform the subtraction to reduce
	// our result.
	//
	//   - overflow = 0, underflow = 1
	//
	// The addition fits in our limbs, but we can't subtract m without
	// underflowing. The result is already reduced.
	//
	//   - overflow = 1, underflow = 1
	//
	// The addition does not fit in our limbs, and the subtraction's borrow
	// would cancel out with the addition's carry. We need to subtract m to
	// reduce our result.
	//
	// The overflow = 1, underflow = 0 case is not possible, because y is at
	// most m - 1, and if adding m - 1 overflows, then subtracting m must
	// necessarily underflow.
	needSubtraction := ctEq(overflow, uint(underflow))

	x.sub(needSubtraction, m.nat)
	return x
}

// montgomeryRepresentation calculates x = x * R mod m, with R = 2^(_W * n) and
// n = len(m.nat.limbs).
//
// Faster Montgomery multiplication replaces standard modular multiplication for
// numbers in this representation.
//
// This assumes that x is already reduced mod m.
func (x *nat) montgomeryRepresentation(m *modulus) *nat {
	checkReduced(m, x)
	for i := 0; i < len(m.nat.limbs); i++ {
		x.shiftIn(0, m) // x = x * 2^_W mod m
	}
	return x
}

// montgomeryMul calculates out = x * y / R mod m, with R = 2^(_W * n) and
// n = len(m.nat.limbs).
//
// This is faster than your standard modular multiplication.
//
// All inputs should be the same length, and not alias each other. TODO: reduced?
func (out *nat) montgomeryMul(x *nat, y *nat, m *modulus) *nat { // TODO
	for i := 0; i < len(out.limbs); i++ {
		out.limbs[i] = 0
	}

	overflow := uint(0)
	// The different loops are over the same size, but we use different conditions
	// to try and make the compiler elide bounds checking.
	for i := 0; i < len(x.limbs); i++ {
		f := ((out.limbs[0] + x.limbs[i]*y.limbs[0]) * m.m0inv) & _MASK
		// Carry fits on 64 bits
		var carry uint
		for j := 0; j < len(y.limbs) && j < len(m.nat.limbs) && j < len(out.limbs); j++ {
			hi, lo := bits.Mul(x.limbs[i], y.limbs[j])
			z_lo, c := bits.Add(out.limbs[j], lo, 0)
			z_hi, _ := bits.Add(0, hi, c)
			hi, lo = bits.Mul(f, m.nat.limbs[j])
			z_lo, c = bits.Add(z_lo, lo, 0)
			z_hi, _ = bits.Add(z_hi, hi, c)
			z_lo, c = bits.Add(z_lo, carry, 0)
			z_hi, _ = bits.Add(z_hi, 0, c)
			if j > 0 {
				out.limbs[j-1] = z_lo & _MASK
			}
			carry = (z_lo >> _W) | (z_hi << 1)
		}
		z := overflow + carry
		out.limbs[len(out.limbs)-1] = z & _MASK
		overflow = z >> _W
	}
	underflow := not(out.cmpGeq(m.nat))
	// See modAdd
	needSubtraction := ctEq(overflow, uint(underflow))
	out.sub(needSubtraction, m.nat)
	return out
}

// modMul calculates x *= y mod m.
//
// x and y must already be reduced modulo m, they must share its announced
// length, and they may not alias.
func (x *nat) modMul(y *nat, m *modulus) *nat {
	checkReduced(m, x, y)
	// A Montgomery multiplication by a value out of the Montgomery domain
	// takes the result out of Montgomery representation.
	xR := x.clone().montgomeryRepresentation(m) // xR = x * R mod m
	return x.montgomeryMul(xR, y, m)            // x = xR * y / R mod m
}

// exp calculates out <- x^e modulo m
//
// The exponent, e, is presented as bytes in big endian order.
//
// The output will be resized to the size of m and overwritten. x must already
// be reduced modulo m.
func (out *nat) exp(x *nat, e []byte, m *modulus) *nat { // TODO
	checkReduced(m, x)
	size := len(m.nat.limbs)
	out.reset(size)

	// We use 4 bit windows. For our RSA workload, 4 bit windows are
	// faster than 2 bit windows, but use an extra 12 nats worth of scratch space.
	// Using bit sizes that don't divide 8 are a bit awkward to implement.
	xs := make([]*nat, 15)
	xs[0] = x.clone().montgomeryRepresentation(m)
	for i := 1; i < len(xs); i++ {
		xs[i] = &nat{make([]uint, size)}
		xs[i].montgomeryMul(xs[i-1], xs[0], m)
	}

	selectedX := &nat{make([]uint, size)}
	out.limbs[0] = 1
	out.montgomeryRepresentation(m)
	scratch := &nat{make([]uint, size)}
	for _, b := range e {
		for j := 4; j >= 0; j -= 4 {
			scratch.montgomeryMul(out, out, m)
			out.montgomeryMul(scratch, scratch, m)
			scratch.montgomeryMul(out, out, m)
			out.montgomeryMul(scratch, scratch, m)

			window := uint((b >> j) & 0b1111)
			for i := 0; i < len(xs); i++ {
				selectedX.assign(ctEq(window, uint(i+1)), xs[i])
			}
			scratch.montgomeryMul(out, selectedX, m)
			out.assign(not(ctEq(window, 0)), scratch)
		}
	}
	for i := 0; i < len(scratch.limbs); i++ {
		scratch.limbs[i] = 0
	}
	scratch.limbs[0] = 1
	// By montgomery multiplying with 1, we convert back from montgomery representation
	outC := out.clone()
	out.montgomeryMul(outC, scratch, m)
	return out
}

func checkReduced(m *modulus, xs ...*nat) {
	for _, x := range xs {
		if len(x.limbs) != len(m.nat.limbs) {
			panic("rsa: internal error: checkReduced called on unevenly sized nat")
		}
		if x.cmpGeq(m.nat) == 1 {
			panic("rsa: internal error: out of range nat")
		}
	}
}
