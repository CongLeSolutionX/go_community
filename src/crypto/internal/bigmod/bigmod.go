// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bigmod

import (
	"errors"
	"math/bits"
)

// Modulus is a reusable modulus for Int values. Moduli can only be odd numbers.
// Moduli leak their exact bit length but not their actual value.
type Modulus struct {
	// The underlying natural number for this modulus.
	//
	// This will be stored without any padding, and can't alias with any
	// other natural number being used.
	nat     *nat
	leading int  // number of leading zeros in the modulus
	m0inv   uint // -nat.limbs[0]⁻¹ mod _W
}

// NewModulusFromBytes produces a new Modulus from a big-endian byte slice.
// The input must be an odd number.
func NewModulusFromBytes(b []byte) (*Modulus, error) {
	if len(b) == 0 {
		return nil, errors.New("bigmod: Modulus can't be zero")
	}
	if b[len(b)-1]&1 == 0 {
		return nil, errors.New("bigmod: Modulus can't be even")
	}
	return modulusFromNat(natFromBytes(b)), nil
}

// NewModulusFromBits creates a new Modulus from a little-endian saturated uint
// slice, like those returned by (*big.Int).Bits. The input must be an odd
// number.
func NewModulusFromBits(b []uint) (*Modulus, error) {
	if len(b) == 0 {
		return nil, errors.New("bigmod: Modulus can't be zero")
	}
	if b[0]&1 == 0 {
		return nil, errors.New("bigmod: Modulus can't be even")
	}
	return modulusFromNat(natFromBits(b)), nil
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
func modulusFromNat(nat *nat) *Modulus {
	m := &Modulus{}
	m.nat = nat
	size := len(m.nat.limbs)
	for m.nat.limbs[size-1] == 0 {
		size--
	}
	m.nat.limbs = m.nat.limbs[:size]
	m.leading = _W - bitLen(m.nat.limbs[size-1])
	m.m0inv = minusInverseModW(m.nat.limbs[0])
	return m
}

// An Int is a natural number modulo an odd integer. The modulus is fixed for
// the lifetime of the Int. All operations on the Int are performed in constant
// time and leak only the bit size of the modulus.
type Int struct {
	// nat is always less than m, of the same announced size as m, and in
	// Montgomery representation with R = 2^(_W * len(m.nat.limbs)).
	nat *nat
	m   *Modulus
}

// NewInt produces a new zero Int with a given modulus.
func NewInt(m *Modulus) *Int {
	return &Int{nat: &nat{make([]uint, len(m.nat.limbs))}, m: m}
}

// Size returns the size of i's modulus in bytes.
func (i *Int) Size() int {
	bits := len(i.m.nat.limbs)*_W - int(i.m.leading)
	return (bits + 7) / 8
}

// SetBytes sets i to the big-endian value in b. b must be the exact byte length
// of i's modulus, and its value must be reduced by i's modulus.
func (i *Int) SetBytes(b []byte) (*Int, error) {
	if len(b) != i.Size() {
		return nil, errors.New("bigmod: SetBytes input has the wrong length")
	}
	n := natFromBytes(b)
	n.limbs = n.limbs[:len(i.nat.limbs)]
	if n.cmpGeq(i.m.nat) == 1 {
		return nil, errors.New("bigmod: SetBytes input is not reduced")
	}
	copy(i.nat.limbs, n.limbs)
	i.nat.montgomeryRepresentation(i.m)
	return i, nil
}

func (i *Int) SetBits(b []uint) (*Int, error) {
	n := natFromBits(b)
	if len(i.nat.limbs) < len(n.limbs) {
		n.limbs = n.limbs[:len(i.nat.limbs)] // TODO
	} else if len(i.nat.limbs) > len(n.limbs) {
		limbs := make([]uint, len(i.nat.limbs)) // TODO
		copy(limbs, n.limbs)
		n.limbs = limbs
	}
	if n.cmpGeq(i.m.nat) == 1 {
		return nil, errors.New("bigmod: SetBits input is not reduced")
	}
	copy(i.nat.limbs, n.limbs)
	i.nat.montgomeryRepresentation(i.m)
	return i, nil
}

// SetInt sets i to k. If the modulus of i is equal or larger than the modulus
// of k, a simple copy is performed. Otherwise, k is reduced mod i's modulus.
func (i *Int) SetInt(k *Int) *Int {
	switch {
	// Same modulus, a copy is enough.
	case i.m == k.m:
		copy(i.nat.limbs, k.nat.limbs)

	// New module has fewer limbs, so the Montgomery R value is smaller. Get out
	// of the Montgomery domain (see FillBytes), do a reduction, and get into
	// the new Montgomery domain. TODO(filippo): is there a more efficient way?
	case len(i.m.nat.limbs) < len(k.m.nat.limbs):
		x := NewInt(k.m)
		x.nat.limbs[0] = 1
		x.Mul(k)
		i.nat.mod(x.nat, i.m)
		i.nat.montgomeryRepresentation(i.m)

	// New module is smaller but has the same number of limbs. The Montgomery
	// domain is the same, we just need to do a reduction.
	case len(i.m.nat.limbs) == len(k.m.nat.limbs) && i.m.nat.cmpGeq(k.m.nat) == 0:
		i.nat.mod(k.nat, i.m)

	// New module is bigger. No need to reduce, but increase the Montgomery R
	// value by as many extra limbs as needed.
	default:
		zeroLimbs(i.nat.limbs)
		copy(i.nat.limbs, k.nat.limbs)
		for k := len(k.nat.limbs); k < len(i.nat.limbs); k++ {
			i.nat.shiftIn(0, i.m)
		}
	}

	return i
}

// FillBytes sets bytes to i as a zero-extended big-endian byte slice.
//
// If bytes is not long enough to contain i'd modulus, FillBytes will panic.
func (i *Int) FillBytes(bytes []byte) []byte {
	if len(bytes) < i.Size() {
		panic("bigmod: FillBytes invoked with too small buffer")
	}
	for i := range bytes {
		bytes[i] = 0
	}

	// By Montgomery multiplying with 1 not in Montgomery representation, we
	// convert out back from Montgomery representation, because it works out to
	// dividing by R.
	x := NewInt(i.m)
	x.nat.limbs[0] = 1
	x.Mul(i)

	shift := 0
	outI := len(bytes) - 1
	for i, limb := range x.nat.limbs {
		remainingBits := _W
		for remainingBits >= 8 {
			bytes[outI] |= byte(limb) << shift
			consumed := 8 - shift
			limb >>= consumed
			remainingBits -= consumed
			shift = 0
			outI--
			if outI < 0 {
				if limb != 0 || i < len(x.nat.limbs)-1 {
					panic("bigmod: internal error: FillBytes size calculation was incorrect")
				}
				return bytes
			}
		}
		bytes[outI] = byte(limb)
		shift = remainingBits
	}
	return bytes
}

// Mul sets i = i * k. i and k must share the same modulus.
func (i *Int) Mul(k *Int) *Int {
	a, b, m, m0inv := i.nat, k.nat, i.m.nat, i.m.m0inv
	d := &nat{make([]uint, len(m.limbs))}

	// This function calculates d = a * b / R mod m, with R = 2^(_W * n) and
	// n = len(m.nat.limbs), using the Montgomery Multiplication technique.
	//
	// See https://bearssl.org/bigint.html#montgomery-reduction-and-multiplication
	// for a description of the algorithm.

	// Eliminate bounds checks in the loop.
	size := len(m.limbs)
	aLimbs := a.limbs[:size]
	bLimbs := b.limbs[:size]
	dLimbs := d.limbs[:size]
	mLimbs := m.limbs[:size]

	var overflow uint
	for i := 0; i < size; i++ {
		f := ((dLimbs[0] + aLimbs[i]*bLimbs[0]) * m0inv) & _MASK
		carry := uint(0)
		for j := 0; j < size; j++ {
			// z = d[j] + a[i] * b[j] + f * m[j] + carry <= 2^(2W+1) - 2^(W+1) + 2^W
			hi, lo := bits.Mul(aLimbs[i], bLimbs[j])
			z_lo, c := bits.Add(dLimbs[j], lo, 0)
			z_hi, _ := bits.Add(0, hi, c)
			hi, lo = bits.Mul(f, mLimbs[j])
			z_lo, c = bits.Add(z_lo, lo, 0)
			z_hi, _ = bits.Add(z_hi, hi, c)
			z_lo, c = bits.Add(z_lo, carry, 0)
			z_hi, _ = bits.Add(z_hi, 0, c)
			if j > 0 {
				dLimbs[j-1] = z_lo & _MASK
			}
			carry = z_hi<<1 | z_lo>>_W // carry <= 2^(W+1) - 2
		}
		z := overflow + carry // z <= 2^(W+1) - 1
		dLimbs[size-1] = z & _MASK
		overflow = z >> _W // overflow <= 1
	}
	// See modAdd for how overflow, underflow, and needSubtraction relate.
	underflow := not(d.cmpGeq(m)) // d < m
	needSubtraction := ctEq(overflow, uint(underflow))
	d.sub(needSubtraction, m)

	copy(i.nat.limbs, d.limbs)
	return i
}

// Add sets i = i + k. i and k must share the same modulus.
func (i *Int) Add(k *Int) *Int {
	x, y, m := i.nat, k.nat, i.m.nat

	overflow := x.add(yes, y)
	underflow := not(x.cmpGeq(m)) // x < m

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
	x.sub(needSubtraction, m)

	return i
}

// Sub sets i = i - k. i and k must share the same modulus.
func (i *Int) Sub(k *Int) *Int {
	x, y, m := i.nat, k.nat, i.m.nat
	underflow := x.sub(yes, y)
	// If the subtraction underflowed, add back m.
	x.add(choice(underflow), m)
	return i
}

// exp calculates i = i^e mod m.
//
// The exponent e is represented in big-endian order.
func (i *Int) Exp(e []byte) *Int {
	// We use a 4 bit window. For our RSA workload, 4 bit windows are faster
	// than 2 bit windows, but use an extra 12 nats worth of scratch space.
	// Using bit sizes that don't divide 8 are more complex to implement.
	table := make([]*Int, (1<<4)-1) // table[i] = x ^ (i+1)
	table[0] = NewInt(i.m).SetInt(i)
	for k := 1; k < len(table); k++ {
		table[k] = NewInt(i.m).SetInt(i)
		table[k].Mul(table[k-1])
	}

	zeroLimbs(i.nat.limbs)
	i.nat.limbs[0] = 1
	i.nat.montgomeryRepresentation(i.m)
	t0 := NewInt(i.m)
	t1 := NewInt(i.m)
	for _, b := range e {
		for _, j := range []int{4, 0} {
			// Square four times.
			i.Mul(i).Mul(i).Mul(i).Mul(i)

			// Select x^k in constant time from the table.
			k := uint((b >> j) & 0b1111)
			for i := range table {
				t0.nat.assign(ctEq(k, uint(i+1)), table[i].nat)
			}

			// Multiply by x^k, discarding the result if k = 0.
			t1.SetInt(i)
			i.Mul(t0)
			i.nat.assign(ctEq(k, 0), t1.nat)
		}
	}

	return i
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

func zeroLimbs(l []uint) {
	for i := range l {
		l[i] = 0
	}
}
