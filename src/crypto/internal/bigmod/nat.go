// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bigmod

import (
	"encoding/binary"
	"errors"
	"math/big"
	"math/bits"
	_ "unsafe"
)

const (
	// _W is the size in bits of our limbs.
	_W = bits.UintSize
	// _S is the size in bytes of our limbs.
	_S = _W / 8
)

// choice represents a constant-time boolean. The value of choice is always
// either 1 or 0. We use an int instead of bool in order to make decisions in
// constant time by turning it into a mask.
type choice uint

func not(c choice) choice { return 1 ^ c }

const yes = choice(1)
const no = choice(0)

// ctMask is all 1s if on is yes, and all 0s otherwise.
func ctMask(on choice) uint { return -uint(on) }

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

// Nat represents an arbitrary natural number
//
// Each Nat has an announced length, which is the number of limbs it has stored.
// Operations on this number are allowed to leak this length, but will not leak
// any information about the values contained in those limbs.
type Nat struct {
	// limbs is little-endian in base 2^W with W = bits.UintSize.
	limbs []uint
}

// preallocTarget is the size in bits of the numbers used to implement the most
// common and most performant RSA key size. It's also enough to cover some of
// the operations of key sizes up to 4096.
const preallocTarget = 2048
const preallocLimbs = (preallocTarget + _W - 1) / _W

// NewNat returns a new nat with a size of zero, just like new(Nat), but with
// the preallocated capacity to hold a number of up to preallocTarget bits.
// NewNat inlines, so the allocation can live on the stack.
func NewNat() *Nat {
	limbs := make([]uint, 0, preallocLimbs)
	return &Nat{limbs}
}

// expand expands x to n limbs, leaving its value unchanged.
func (x *Nat) expand(n int) *Nat {
	if len(x.limbs) > n {
		panic("bigmod: internal error: shrinking nat")
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

// reset returns a zero nat of n limbs, reusing x's storage if n <= cap(x.limbs).
func (x *Nat) reset(n int) *Nat {
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

// set assigns x = y, optionally resizing x to the appropriate size.
func (x *Nat) set(y *Nat) *Nat {
	x.reset(len(y.limbs))
	copy(x.limbs, y.limbs)
	return x
}

// setBig assigns x = n, optionally resizing n to the appropriate size.
//
// The announced length of x is set based on the actual bit size of the input,
// ignoring leading zeroes.
func (x *Nat) setBig(n *big.Int) *Nat {
	limbs := n.Bits()
	x.reset(len(limbs))
	for i := range limbs {
		x.limbs[i] = uint(limbs[i])
	}
	return x
}

// Bytes returns x as a zero-extended big-endian byte slice. The size of the
// slice will match the size of m.
//
// x must have the same size as m and it must be reduced modulo m.
func (x *Nat) Bytes(m *Modulus) []byte {
	i := m.Size()
	bytes := make([]byte, i)
	for _, limb := range x.limbs {
		for j := 0; j < _S; j++ {
			i--
			if i < 0 {
				if limb == 0 {
					break
				}
				panic("bigmod: modulus is smaller than nat")
			}
			bytes[i] = byte(limb)
			limb >>= 8
		}
	}
	return bytes
}

// SetBytes assigns x = b, where b is a slice of big-endian bytes.
// SetBytes returns an error if b >= m.
//
// The output will be resized to the size of m and overwritten.
func (x *Nat) SetBytes(b []byte, m *Modulus) (*Nat, error) {
	if err := x.setBytes(b, m); err != nil {
		return nil, err
	}
	if x.cmpGeq(m.nat) == yes {
		return nil, errors.New("input overflows the modulus")
	}
	return x, nil
}

// SetOverflowingBytes assigns x = b, where b is a slice of big-endian bytes.
// SetOverflowingBytes returns an error if b has a longer bit length than m, but
// reduces overflowing values up to 2^⌈log2(m)⌉ - 1.
//
// The output will be resized to the size of m and overwritten.
func (x *Nat) SetOverflowingBytes(b []byte, m *Modulus) (*Nat, error) {
	if err := x.setBytes(b, m); err != nil {
		return nil, err
	}
	leading := _W - bitLen(x.limbs[len(x.limbs)-1])
	if leading < m.leading {
		return nil, errors.New("input overflows the modulus size")
	}
	x.sub(x.cmpGeq(m.nat), m.nat)
	return x, nil
}

// bigEndianUint returns the contents of buf interpreted as a
// big-endian encoded uint value.
func bigEndianUint(buf []byte) uint {
	if _W == 64 {
		return uint(binary.BigEndian.Uint64(buf))
	}
	return uint(binary.BigEndian.Uint32(buf))
}

func (x *Nat) setBytes(b []byte, m *Modulus) error {
	x.resetFor(m)
	i, k := len(b), 0
	for k < len(x.limbs) && i >= _S {
		x.limbs[k] = bigEndianUint(b[i-_S : i])
		i -= _S
		k++
	}
	for s := 0; s < _W && k < len(x.limbs) && i > 0; s += 8 {
		x.limbs[k] |= uint(b[i-1]) << s
		i--
	}
	if i > 0 {
		return errors.New("input overflows the modulus size")
	}
	return nil
}

// Equal returns 1 if x == y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *Nat) Equal(y *Nat) choice {
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

// IsZero returns 1 if x == 0, and 0 otherwise.
func (x *Nat) IsZero() choice {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]

	zero := yes
	for i := 0; i < size; i++ {
		zero &= ctEq(xLimbs[i], 0)
	}
	return zero
}

// cmpGeq returns 1 if x >= y, and 0 otherwise.
//
// Both operands must have the same announced length.
func (x *Nat) cmpGeq(y *Nat) choice {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	var c uint
	for i := 0; i < size; i++ {
		_, c = bits.Sub(xLimbs[i], yLimbs[i], c)
	}
	// If there was a carry, then subtracting y underflowed, so
	// x is not greater than or equal to y.
	return not(choice(c))
}

// assign sets x <- y if on == 1, and does nothing otherwise.
//
// Both operands must have the same announced length.
func (x *Nat) assign(on choice, y *Nat) *Nat {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	mask := ctMask(on)
	for i := 0; i < size; i++ {
		xLimbs[i] ^= mask & (xLimbs[i] ^ yLimbs[i])
	}
	return x
}

// add computes x += y if on == 1, and does nothing otherwise. It returns the
// carry of the addition regardless of on.
//
// Both operands must have the same announced length.
func (x *Nat) add(on choice, y *Nat) (c uint) {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	mask := ctMask(on)
	for i := 0; i < size; i++ {
		var res uint
		res, c = bits.Add(xLimbs[i], yLimbs[i], c)
		xLimbs[i] ^= mask & (xLimbs[i] ^ res)
	}
	return
}

// sub computes x -= y if on == 1, and does nothing otherwise. It returns the
// borrow of the subtraction regardless of on.
//
// Both operands must have the same announced length.
func (x *Nat) sub(on choice, y *Nat) (c uint) {
	// Eliminate bounds checks in the loop.
	size := len(x.limbs)
	xLimbs := x.limbs[:size]
	yLimbs := y.limbs[:size]

	mask := ctMask(on)
	for i := 0; i < size; i++ {
		var res uint
		res, c = bits.Sub(xLimbs[i], yLimbs[i], c)
		xLimbs[i] ^= mask & (xLimbs[i] ^ res)
	}
	return
}

// Modulus is used for modular arithmetic, precomputing relevant constants.
//
// Moduli are assumed to be odd numbers. Moduli can also leak the exact
// number of bits needed to store their value, and are stored without padding.
//
// Their actual value is still kept secret.
type Modulus struct {
	// The underlying natural number for this modulus.
	//
	// This will be stored without any padding, and shouldn't alias with any
	// other natural number being used.
	nat     *Nat
	leading int  // number of leading zeros in the modulus
	m0inv   uint // -nat.limbs[0]⁻¹ mod _W
	rr      *Nat // R*R for montgomeryRepresentation
}

// rr returns R*R with R = 2^(_W * n) and n = len(m.nat.limbs).
func rr(m *Modulus) *Nat {
	rr := NewNat().ExpandFor(m)
	// R*R is 2^(2 * _W * n). We can safely get 2^(_W * (n - 1)) by setting the
	// most significant limb to 1. We then get to R*R by shifting left by _W
	// n + 1 times.
	n := len(rr.limbs)
	rr.limbs[n-1] = 1
	for i := n - 1; i < 2*n; i++ {
		rr.shiftIn(0, m) // x = x * 2^_W mod m
	}
	return rr
}

// minusInverseModW computes -x⁻¹ mod _W with x odd.
//
// This operation is used to precompute a constant involved in Montgomery
// multiplication.
func minusInverseModW(x uint) uint {
	// Every iteration of this loop doubles the least-significant bits of
	// correct inverse in y. The first three bits are already correct (1⁻¹ = 1,
	// 3⁻¹ = 3, 5⁻¹ = 5, and 7⁻¹ = 7 mod 8), so doubling five times is enough
	// for 64 bits (and wastes only one iteration for 32 bits).
	//
	// See https://crypto.stackexchange.com/a/47496.
	y := x
	for i := 0; i < 5; i++ {
		y = y * (2 - x*y)
	}
	return -y
}

// NewModulusFromBig creates a new Modulus from a [big.Int].
//
// The Int must be odd. The number of significant bits (and nothing else) is
// leaked through timing side-channels.
func NewModulusFromBig(n *big.Int) *Modulus {
	m := &Modulus{}
	m.nat = NewNat().setBig(n)
	m.leading = _W - bitLen(m.nat.limbs[len(m.nat.limbs)-1])
	m.m0inv = minusInverseModW(m.nat.limbs[0])
	m.rr = rr(m)
	return m
}

// bitLen is a version of bits.Len that only leaks the bit length of n, but not
// its value. bits.Len and bits.LeadingZeros use a lookup table for the
// low-order bits on some architectures.
func bitLen(n uint) int {
	var len int
	// We assume, here and elsewhere, that comparison to zero is constant time
	// with respect to different non-zero values.
	for n != 0 {
		len++
		n >>= 1
	}
	return len
}

// Size returns the size of m in bytes.
func (m *Modulus) Size() int {
	return (m.BitLen() + 7) / 8
}

// BitLen returns the size of m in bits.
func (m *Modulus) BitLen() int {
	return len(m.nat.limbs)*_W - int(m.leading)
}

// Nat returns m as a Nat. The return value must not be written to.
func (m *Modulus) Nat() *Nat {
	return m.nat
}

// shiftIn calculates x = x << _W + y mod m.
//
// This assumes that x is already reduced mod m.
func (x *Nat) shiftIn(y uint, m *Modulus) *Nat {
	d := NewNat().resetFor(m)

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
	needSubtraction, mask := no, ctMask(no)
	for i := _W - 1; i >= 0; i-- {
		carry := (y >> i) & 1
		var borrow uint
		for i := 0; i < size; i++ {
			l := xLimbs[i] ^ (mask & (xLimbs[i] ^ dLimbs[i]))
			xLimbs[i], carry = bits.Add(l, l, carry)
			dLimbs[i], borrow = bits.Sub(xLimbs[i], mLimbs[i], borrow)
		}
		// See Add for how carry (aka overflow), borrow (aka underflow), and
		// needSubtraction relate.
		needSubtraction = ctEq(carry, borrow)
		mask = ctMask(needSubtraction)
	}
	return x.assign(needSubtraction, d)
}

// Mod calculates out = x mod m.
//
// This works regardless how large the value of x is.
//
// The output will be resized to the size of m and overwritten.
func (out *Nat) Mod(x *Nat, m *Modulus) *Nat {
	out.resetFor(m)
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

// ExpandFor ensures out has the right size to work with operations modulo m.
//
// The announced size of out must be smaller than or equal to that of m.
func (out *Nat) ExpandFor(m *Modulus) *Nat {
	return out.expand(len(m.nat.limbs))
}

// resetFor ensures out has the right size to work with operations modulo m.
//
// out is zeroed and may start at any size.
func (out *Nat) resetFor(m *Modulus) *Nat {
	return out.reset(len(m.nat.limbs))
}

// Sub computes x = x - y mod m.
//
// The length of both operands must be the same as the modulus. Both operands
// must already be reduced modulo m.
func (x *Nat) Sub(y *Nat, m *Modulus) *Nat {
	underflow := x.sub(yes, y)
	// If the subtraction underflowed, add m.
	x.add(choice(underflow), m.nat)
	return x
}

// Add computes x = x + y mod m.
//
// The length of both operands must be the same as the modulus. Both operands
// must already be reduced modulo m.
func (x *Nat) Add(y *Nat, m *Modulus) *Nat {
	overflow := x.add(yes, y)
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
func (x *Nat) montgomeryRepresentation(m *Modulus) *Nat {
	// A Montgomery multiplication (which computes a * b / R) by R * R works out
	// to a multiplication by R, which takes the value out of the Montgomery domain.
	return x.montgomeryMul(x, m.rr, m)
}

// montgomeryReduction calculates x = x / R mod m, with R = 2^(_W * n) and
// n = len(m.nat.limbs).
//
// This assumes that x is already reduced mod m.
func (x *Nat) montgomeryReduction(m *Modulus) *Nat {
	// By Montgomery multiplying with 1 not in Montgomery representation, we
	// convert out back from Montgomery representation, because it works out to
	// dividing by R.
	one := NewNat().ExpandFor(m)
	one.limbs[0] = 1
	return x.montgomeryMul(x, one, m)
}

// montgomeryMul calculates d = a * b / R mod m, with R = 2^(_W * n) and
// n = len(m.nat.limbs), using the Montgomery Multiplication technique.
//
// All inputs should be the same length and already reduced modulo m.
// d will be resized to the size of m and overwritten.
func (d *Nat) montgomeryMul(a *Nat, b *Nat, m *Modulus) *Nat {
	n := len(m.nat.limbs)
	mLimbs := m.nat.limbs[:n]
	aLimbs := a.limbs[:n]
	bLimbs := b.limbs[:n]

	switch n {
	default:
		z := make([]uint, 0, preallocLimbs*2)
		if cap(z) < n*2 {
			z = make([]uint, 0, n*2)
		}
		z = z[:n*2]

		// TODO(filippo): document this loop.
		var c choice
		for i := 0; i < n; i++ {
			_ = z[n+i]
			d := bLimbs[i]
			c2 := addMulVVW(z[i:n+i], aLimbs, d)
			t := z[i] * m.m0inv
			c3 := addMulVVW(z[i:n+i], mLimbs, t)
			cx := uint(c) + c2
			cy := cx + c3
			z[n+i] = cy
			c = not(ctGeq(cx, c2)) | not(ctGeq(cy, c3))
		}
		copy(d.reset(n).limbs, z[n:])
		d.sub(c, m.nat)

	// The following specialized cases follow the exact same algorithm, but
	// optimized for the sizes most used in RSA. addMulVVW is implemented in
	// assembly with loop unrolling depending on the architecture and bounds
	// checks are removed by the compiler thanks to the constant size.
	case 1024 / _W:
		const n = 1024 / _W // compiler hint
		z := make([]uint, n*2)
		var c choice
		for i := 0; i < n; i++ {
			d := bLimbs[i]
			c2 := addMulVVW1024(&z[i], &aLimbs[0], d)
			t := z[i] * m.m0inv
			c3 := addMulVVW1024(&z[i], &mLimbs[0], t)
			cx := uint(c) + c2
			cy := cx + c3
			z[n+i] = cy
			c = not(ctGeq(cx, c2)) | not(ctGeq(cy, c3))
		}
		copy(d.reset(n).limbs, z[n:])
		d.sub(c, m.nat)

	case 1536 / _W:
		const n = 1536 / _W // compiler hint
		z := make([]uint, n*2)
		var c choice
		for i := 0; i < n; i++ {
			d := bLimbs[i]
			c2 := addMulVVW1536(&z[i], &aLimbs[0], d)
			t := z[i] * m.m0inv
			c3 := addMulVVW1536(&z[i], &mLimbs[0], t)
			cx := uint(c) + c2
			cy := cx + c3
			z[n+i] = cy
			c = not(ctGeq(cx, c2)) | not(ctGeq(cy, c3))
		}
		copy(d.reset(n).limbs, z[n:])
		d.sub(c, m.nat)

	case 2048 / _W:
		const n = 2048 / _W // compiler hint
		z := make([]uint, n*2)
		var c choice
		for i := 0; i < n; i++ {
			d := bLimbs[i]
			c2 := addMulVVW2048(&z[i], &aLimbs[0], d)
			t := z[i] * m.m0inv
			c3 := addMulVVW2048(&z[i], &mLimbs[0], t)
			cx := uint(c) + c2
			cy := cx + c3
			z[n+i] = cy
			c = not(ctGeq(cx, c2)) | not(ctGeq(cy, c3))
		}
		copy(d.reset(n).limbs, z[n:])
		d.sub(c, m.nat)
	}

	return d
}

// addMulVVW multiplies each element of x by y, adding the result to the
// corresponding element of z and carrying the overflows up z and eventually
// returning the final carry as c.
func addMulVVW(z, x []uint, y uint) (carry uint) {
	_ = x[len(z)-1] // bounds check elimination hint
	for i := range z {
		hi, lo := bits.Mul(x[i], y)
		lo, c := bits.Add(lo, z[i], 0)
		// We use bits.Add with zero to get an add-with-carry instruction that
		// absorbs the carry from the previous bits.Add.
		hi, _ = bits.Add(hi, 0, c)
		lo, c = bits.Add(lo, carry, 0)
		hi, _ = bits.Add(hi, 0, c)
		carry = hi
		z[i] = lo
	}
	return carry
}

// Mul calculates x = x * y mod m.
//
// The length of both operands must be the same as the modulus. Both operands
// must already be reduced modulo m.
func (x *Nat) Mul(y *Nat, m *Modulus) *Nat {
	// A Montgomery multiplication by a value out of the Montgomery domain
	// takes the result out of Montgomery representation.
	xR := NewNat().set(x).montgomeryRepresentation(m) // xR = x * R mod m
	return x.montgomeryMul(xR, y, m)                  // x = xR * y / R mod m
}

// Exp calculates out = x^e mod m.
//
// The exponent e is represented in big-endian order. The output will be resized
// to the size of m and overwritten. x must already be reduced modulo m.
func (out *Nat) Exp(x *Nat, e []byte, m *Modulus) *Nat {
	// We use a 4 bit window. For our RSA workload, 4 bit windows are faster
	// than 2 bit windows, but use an extra 12 nats worth of scratch space.
	// Using bit sizes that don't divide 8 are more complex to implement.

	table := [(1 << 4) - 1]*Nat{ // table[i] = x ^ (i+1)
		// newNat calls are unrolled so they are allocated on the stack.
		NewNat(), NewNat(), NewNat(), NewNat(), NewNat(),
		NewNat(), NewNat(), NewNat(), NewNat(), NewNat(),
		NewNat(), NewNat(), NewNat(), NewNat(), NewNat(),
	}
	table[0].set(x).montgomeryRepresentation(m)
	for i := 1; i < len(table); i++ {
		table[i].montgomeryMul(table[i-1], table[0], m)
	}

	out.resetFor(m)
	out.limbs[0] = 1
	out.montgomeryRepresentation(m)
	tmp := NewNat().ExpandFor(m)
	for _, b := range e {
		for _, j := range []int{4, 0} {
			// Square four times.
			out.montgomeryMul(out, out, m)
			out.montgomeryMul(out, out, m)
			out.montgomeryMul(out, out, m)
			out.montgomeryMul(out, out, m)

			// Select x^k in constant time from the table.
			k := uint((b >> j) & 0b1111)
			for i := range table {
				tmp.assign(ctEq(k, uint(i+1)), table[i])
			}

			// Multiply by x^k, discarding the result if k = 0.
			tmp.montgomeryMul(out, tmp, m)
			out.assign(not(ctEq(k, 0)), tmp)
		}
	}

	return out.montgomeryReduction(m)
}

// ExpShort calculates out = x^e mod m.
//
// The output will be resized to the size of m and overwritten. x must already
// be reduced modulo m. This leaks the exact bit size of the exponent.
func (out *Nat) ExpShort(x *Nat, e uint, m *Modulus) *Nat {
	xR := NewNat().set(x).montgomeryRepresentation(m)

	out.resetFor(m)
	out.limbs[0] = 1
	out.montgomeryRepresentation(m)

	// For short exponents, precomputing a table and using a window like in Exp
	// doesn't pay off. Instead, we do a simple constant-time conditional
	// double-and-add, skipping the initial run of zeroes.
	tmp := NewNat().ExpandFor(m)
	for i := bits.UintSize - bitLen(e); i < bits.UintSize; i++ {
		out.montgomeryMul(out, out, m)
		k := (e >> (bits.UintSize - i - 1)) & 1
		tmp.montgomeryMul(out, xR, m)
		out.assign(ctEq(k, 1), tmp)
	}
	return out.montgomeryReduction(m)
}
