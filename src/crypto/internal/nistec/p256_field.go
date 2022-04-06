// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !arm64 && !ppc64le

package nistec

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"math/bits"
)

// Field elements are represented as nine, unsigned 32-bit words.
//
// The value of a field element is:
//
//	x[0] + (x[1] * 2**29) + (x[2] * 2**57) + ... + (x[8] * 2**228)
//
// That is, each limb is alternately 29 or 28-bits wide in little-endian
// order.
//
// This means that a field element hits 2**257, rather than 2**256 as we would
// like. A 28, 29, ... pattern would cause us to hit 2**256, but that causes
// problems when multiplying as terms end up one bit short of a limb which
// would require much bit-shifting to correct.
//
// Finally, the values stored in a field element are in Montgomery form. So the
// value |y| is stored as (y*R) mod p, where p is the P-256 prime and R is
// 2**257.
type p256Element [9]uint32

const p256ElementLen = 32

// One sets out = 1, and returns out.
func (out *p256Element) One() *p256Element {
	// 1 in Montgomery representation (1*R mod p).
	*out = p256Element{2, 0, 0, 0xffff800, 0x1fffffff, 0xfffffff, 0x1fbfffff, 0x1ffffff, 0}
	return out
}

// Equal returns 1 if e == t, and zero otherwise.
func (e *p256Element) Equal(t *p256Element) int {
	eBytes := e.Bytes()
	tBytes := t.Bytes()
	return subtle.ConstantTimeCompare(eBytes, tBytes)
}

var p256ZeroEncoding = new(p256Element).Bytes()

// IsZero returns 1 if e == 0, and zero otherwise.
func (e *p256Element) IsZero() int {
	eBytes := e.Bytes()
	return subtle.ConstantTimeCompare(eBytes, p256ZeroEncoding)
}

// Set sets out = in, and returns out.
func (out *p256Element) Set(in *p256Element) *p256Element {
	*out = *in
	return out
}

const bottom29Bits = 0x1fffffff
const bottom28Bits = 0xfffffff

// nonZeroToAllOnes returns:
//
//	0xffffffff for 0 < x <= 2**31
//	0 for x == 0 or x > 2**31.
func nonZeroToAllOnes(x uint32) uint32 {
	return ((x - 1) >> 31) - 1
}

// p256ReduceCarry adds a multiple of p in order to cancel |carry|,
// which is a term at 2**257.
//
// On entry: carry < 2**3, inout[0,2,...] < 2**29, inout[1,3,...] < 2**28.
// On exit: inout[0,2,..] < 2**30, inout[1,3,...] < 2**29.
func p256ReduceCarry(inout *p256Element, carry uint32) {
	carry_mask := nonZeroToAllOnes(carry)

	inout[0] += carry << 1
	inout[3] += 0x10000000 & carry_mask
	// carry < 2**3 thus (carry << 11) < 2**14 and we added 2**28 in the
	// previous line therefore this doesn't underflow.
	inout[3] -= carry << 11
	inout[4] += (0x20000000 - 1) & carry_mask
	inout[5] += (0x10000000 - 1) & carry_mask
	inout[6] += (0x20000000 - 1) & carry_mask
	inout[6] -= carry << 22
	// This may underflow if carry is non-zero but, if so, we'll fix it in the
	// next line.
	inout[7] -= 1 & carry_mask
	inout[7] += carry << 25
}

// Add sets out = in+in2.
//
// On entry: in[i]+in2[i] must not overflow a 32-bit word.
// On exit: out[0,2,...] < 2**30, out[1,3,...] < 2**29.
func (out *p256Element) Add(in, in2 *p256Element) *p256Element {
	carry := uint32(0)
	for i := 0; ; i++ {
		out[i] = in[i] + in2[i]
		out[i] += carry
		carry = out[i] >> 29
		out[i] &= bottom29Bits

		i++
		if i == len(out) {
			break
		}

		out[i] = in[i] + in2[i]
		out[i] += carry
		carry = out[i] >> 28
		out[i] &= bottom28Bits
	}

	p256ReduceCarry(out, carry)

	return out
}

const (
	two30m2    = 1<<30 - 1<<2
	two30p13m2 = 1<<30 + 1<<13 - 1<<2
	two31m2    = 1<<31 - 1<<2
	two31m3    = 1<<31 - 1<<3
	two31p24m2 = 1<<31 + 1<<24 - 1<<2
	two30m27m2 = 1<<30 - 1<<27 - 1<<2
)

// p256Zero31 is 0 mod p.
var p256Zero31 = p256Element{two31m3, two30m2, two31m2, two30p13m2, two31m2, two30m2, two31p24m2, two30m27m2, two31m2}

// Sub sets out = in-in2.
//
// On entry: in[0,2,...] < 2**30, in[1,3,...] < 2**29 and
// in2[0,2,...] < 2**30, in2[1,3,...] < 2**29.
// On exit: out[0,2,...] < 2**30, out[1,3,...] < 2**29.
func (out *p256Element) Sub(in, in2 *p256Element) *p256Element {
	carry := uint32(0)
	for i := 0; ; i++ {
		out[i] = in[i] - in2[i]
		out[i] += p256Zero31[i]
		out[i] += carry
		carry = out[i] >> 29
		out[i] &= bottom29Bits

		i++
		if i == len(out) {
			break
		}

		out[i] = in[i] - in2[i]
		out[i] += p256Zero31[i]
		out[i] += carry
		carry = out[i] >> 28
		out[i] &= bottom28Bits
	}

	p256ReduceCarry(out, carry)
	return out
}

// p256ReduceDegree sets out = tmp/R mod p where tmp contains 64-bit words with
// the same 29,28,... bit positions as a field element.
//
// The values in field elements are in Montgomery form: x*R mod p where R =
// 2**257. Since we just multiplied two Montgomery values together, the result
// is x*y*R*R mod p. We wish to divide by R in order for the result also to be
// in Montgomery form.
//
// On entry: tmp[i] < 2**64.
// On exit: out[0,2,...] < 2**30, out[1,3,...] < 2**29.
func p256ReduceDegree(out *p256Element, tmp [17]uint64) {
	// The following table may be helpful when reading this code:
	//
	// Limb number:   0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10...
	// Width (bits):  29| 28| 29| 28| 29| 28| 29| 28| 29| 28| 29
	// Start bit:     0 | 29| 57| 86|114|143|171|200|228|257|285
	//   (odd phase): 0 | 28| 57| 85|114|142|171|199|228|256|285
	var tmp2 [18]uint32
	var carry, x, xMask uint32

	// tmp contains 64-bit words with the same 29,28,29-bit positions as a
	// field element. So the top of an element of tmp might overlap with
	// another element two positions down. The following loop eliminates
	// this overlap.
	tmp2[0] = uint32(tmp[0]) & bottom29Bits

	tmp2[1] = uint32(tmp[0]) >> 29
	tmp2[1] |= (uint32(tmp[0]>>32) << 3) & bottom28Bits
	tmp2[1] += uint32(tmp[1]) & bottom28Bits
	carry = tmp2[1] >> 28
	tmp2[1] &= bottom28Bits

	for i := 2; i < 17; i++ {
		tmp2[i] = (uint32(tmp[i-2] >> 32)) >> 25
		tmp2[i] += (uint32(tmp[i-1])) >> 28
		tmp2[i] += (uint32(tmp[i-1]>>32) << 4) & bottom29Bits
		tmp2[i] += uint32(tmp[i]) & bottom29Bits
		tmp2[i] += carry
		carry = tmp2[i] >> 29
		tmp2[i] &= bottom29Bits

		i++
		if i == 17 {
			break
		}
		tmp2[i] = uint32(tmp[i-2]>>32) >> 25
		tmp2[i] += uint32(tmp[i-1]) >> 29
		tmp2[i] += ((uint32(tmp[i-1] >> 32)) << 3) & bottom28Bits
		tmp2[i] += uint32(tmp[i]) & bottom28Bits
		tmp2[i] += carry
		carry = tmp2[i] >> 28
		tmp2[i] &= bottom28Bits
	}

	tmp2[17] = uint32(tmp[15]>>32) >> 25
	tmp2[17] += uint32(tmp[16]) >> 29
	tmp2[17] += uint32(tmp[16]>>32) << 3
	tmp2[17] += carry

	// Montgomery elimination of terms:
	//
	// Since R is 2**257, we can divide by R with a bitwise shift if we can
	// ensure that the right-most 257 bits are all zero. We can make that true
	// by adding multiplies of p without affecting the value.
	//
	// So we eliminate limbs from right to left. Since the bottom 29 bits of p
	// are all ones, then by adding tmp2[0]*p to tmp2 we'll make tmp2[0] == 0.
	// We can do that for 8 further limbs and then right shift to eliminate the
	// extra factor of R.
	for i := 0; ; i += 2 {
		tmp2[i+1] += tmp2[i] >> 29
		x = tmp2[i] & bottom29Bits
		xMask = nonZeroToAllOnes(x)
		tmp2[i] = 0

		// The bounds calculations for this loop are tricky. Each iteration of
		// the loop eliminates two words by adding values to words to their
		// right.
		//
		// The following table contains the amounts added to each word (as an
		// offset from the value of i at the top of the loop). The amounts are
		// accounted for from the first and second half of the loop separately
		// and are written as, for example, 28 to mean a value <2**28.
		//
		// Word:                   3   4   5   6   7   8   9   10
		// Added in top half:     28  11      29  21  29  28
		//                                        28  29
		//                                            29
		// Added in bottom half:      29  10      28  21  28   28
		//                                            29
		//
		// The value that is currently offset 7 will be offset 5 for the next
		// iteration and then offset 3 for the iteration after that. Therefore
		// the total value added will be the values added at 7, 5 and 3.
		//
		// The following table accumulates these values. The sums at the bottom
		// are written as, for example, 29+28, to mean a value < 2**29+2**28.
		//
		// Word:                   3   4   5   6   7   8   9  10  11  12  13
		//                        28  11  10  29  21  29  28  28  28  28  28
		//                            29  28  11  28  29  28  29  28  29  28
		//                                    29  28  21  21  29  21  29  21
		//                                        10  29  28  21  28  21  28
		//                                        28  29  28  29  28  29  28
		//                                            11  10  29  10  29  10
		//                                            29  28  11  28  11
		//                                                    29      29
		//                        --------------------------------------------
		//                                                30+ 31+ 30+ 31+ 30+
		//                                                28+ 29+ 28+ 29+ 21+
		//                                                21+ 28+ 21+ 28+ 10
		//                                                10  21+ 10  21+
		//                                                    11      11
		//
		// So the greatest amount is added to tmp2[10] and tmp2[12]. If
		// tmp2[10/12] has an initial value of <2**29, then the maximum value
		// will be < 2**31 + 2**30 + 2**28 + 2**21 + 2**11, which is < 2**32,
		// as required.
		tmp2[i+3] += (x << 10) & bottom28Bits
		tmp2[i+4] += (x >> 18)

		tmp2[i+6] += (x << 21) & bottom29Bits
		tmp2[i+7] += x >> 8

		// At position 200, which is the starting bit position for word 7, we
		// have a factor of 0xf000000 = 2**28 - 2**24.
		tmp2[i+7] += 0x10000000 & xMask
		tmp2[i+8] += (x - 1) & xMask
		tmp2[i+7] -= (x << 24) & bottom28Bits
		tmp2[i+8] -= x >> 4

		tmp2[i+8] += 0x20000000 & xMask
		tmp2[i+8] -= x
		tmp2[i+8] += (x << 28) & bottom29Bits
		tmp2[i+9] += ((x >> 1) - 1) & xMask

		if i+1 == len(out) {
			break
		}
		tmp2[i+2] += tmp2[i+1] >> 28
		x = tmp2[i+1] & bottom28Bits
		xMask = nonZeroToAllOnes(x)
		tmp2[i+1] = 0

		tmp2[i+4] += (x << 11) & bottom29Bits
		tmp2[i+5] += (x >> 18)

		tmp2[i+7] += (x << 21) & bottom28Bits
		tmp2[i+8] += x >> 7

		// At position 199, which is the starting bit of the 8th word when
		// dealing with a context starting on an odd word, we have a factor of
		// 0x1e000000 = 2**29 - 2**25. Since we have not updated i, the 8th
		// word from i+1 is i+8.
		tmp2[i+8] += 0x20000000 & xMask
		tmp2[i+9] += (x - 1) & xMask
		tmp2[i+8] -= (x << 25) & bottom29Bits
		tmp2[i+9] -= x >> 4

		tmp2[i+9] += 0x10000000 & xMask
		tmp2[i+9] -= x
		tmp2[i+10] += (x - 1) & xMask
	}

	// We merge the right shift with a carry chain. The words above 2**257 have
	// widths of 28,29,... which we need to correct when copying them down.
	carry = 0
	for i := 0; i < 8; i++ {
		// The maximum value of tmp2[i + 9] occurs on the first iteration and
		// is < 2**30+2**29+2**28. Adding 2**29 (from tmp2[i + 10]) is
		// therefore safe.
		out[i] = tmp2[i+9]
		out[i] += carry
		out[i] += (tmp2[i+10] << 28) & bottom29Bits
		carry = out[i] >> 29
		out[i] &= bottom29Bits

		i++
		out[i] = tmp2[i+9] >> 1
		out[i] += carry
		carry = out[i] >> 28
		out[i] &= bottom28Bits
	}

	out[8] = tmp2[17]
	out[8] += carry
	carry = out[8] >> 29
	out[8] &= bottom29Bits

	p256ReduceCarry(out, carry)
}

// Square sets out=in*in.
//
// On entry: in[0,2,...] < 2**30, in[1,3,...] < 2**29.
// On exit: out[0,2,...] < 2**30, out[1,3,...] < 2**29.
func (out *p256Element) Square(in *p256Element) *p256Element {
	var tmp [17]uint64

	tmp[0] = uint64(in[0]) * uint64(in[0])
	tmp[1] = uint64(in[0]) * (uint64(in[1]) << 1)
	tmp[2] = uint64(in[0])*(uint64(in[2])<<1) +
		uint64(in[1])*(uint64(in[1])<<1)
	tmp[3] = uint64(in[0])*(uint64(in[3])<<1) +
		uint64(in[1])*(uint64(in[2])<<1)
	tmp[4] = uint64(in[0])*(uint64(in[4])<<1) +
		uint64(in[1])*(uint64(in[3])<<2) +
		uint64(in[2])*uint64(in[2])
	tmp[5] = uint64(in[0])*(uint64(in[5])<<1) +
		uint64(in[1])*(uint64(in[4])<<1) +
		uint64(in[2])*(uint64(in[3])<<1)
	tmp[6] = uint64(in[0])*(uint64(in[6])<<1) +
		uint64(in[1])*(uint64(in[5])<<2) +
		uint64(in[2])*(uint64(in[4])<<1) +
		uint64(in[3])*(uint64(in[3])<<1)
	tmp[7] = uint64(in[0])*(uint64(in[7])<<1) +
		uint64(in[1])*(uint64(in[6])<<1) +
		uint64(in[2])*(uint64(in[5])<<1) +
		uint64(in[3])*(uint64(in[4])<<1)
	// tmp[8] has the greatest value of 2**61 + 2**60 + 2**61 + 2**60 + 2**60,
	// which is < 2**64 as required.
	tmp[8] = uint64(in[0])*(uint64(in[8])<<1) +
		uint64(in[1])*(uint64(in[7])<<2) +
		uint64(in[2])*(uint64(in[6])<<1) +
		uint64(in[3])*(uint64(in[5])<<2) +
		uint64(in[4])*uint64(in[4])
	tmp[9] = uint64(in[1])*(uint64(in[8])<<1) +
		uint64(in[2])*(uint64(in[7])<<1) +
		uint64(in[3])*(uint64(in[6])<<1) +
		uint64(in[4])*(uint64(in[5])<<1)
	tmp[10] = uint64(in[2])*(uint64(in[8])<<1) +
		uint64(in[3])*(uint64(in[7])<<2) +
		uint64(in[4])*(uint64(in[6])<<1) +
		uint64(in[5])*(uint64(in[5])<<1)
	tmp[11] = uint64(in[3])*(uint64(in[8])<<1) +
		uint64(in[4])*(uint64(in[7])<<1) +
		uint64(in[5])*(uint64(in[6])<<1)
	tmp[12] = uint64(in[4])*(uint64(in[8])<<1) +
		uint64(in[5])*(uint64(in[7])<<2) +
		uint64(in[6])*uint64(in[6])
	tmp[13] = uint64(in[5])*(uint64(in[8])<<1) +
		uint64(in[6])*(uint64(in[7])<<1)
	tmp[14] = uint64(in[6])*(uint64(in[8])<<1) +
		uint64(in[7])*(uint64(in[7])<<1)
	tmp[15] = uint64(in[7]) * (uint64(in[8]) << 1)
	tmp[16] = uint64(in[8]) * uint64(in[8])

	p256ReduceDegree(out, tmp)
	return out
}

// Mul sets out=in*in2.
//
// On entry: in[0,2,...] < 2**30, in[1,3,...] < 2**29 and
// in2[0,2,...] < 2**30, in2[1,3,...] < 2**29.
// On exit: out[0,2,...] < 2**30, out[1,3,...] < 2**29.
func (out *p256Element) Mul(in, in2 *p256Element) *p256Element {
	var tmp [17]uint64

	tmp[0] = uint64(in[0]) * uint64(in2[0])
	tmp[1] = uint64(in[0])*(uint64(in2[1])<<0) +
		uint64(in[1])*(uint64(in2[0])<<0)
	tmp[2] = uint64(in[0])*(uint64(in2[2])<<0) +
		uint64(in[1])*(uint64(in2[1])<<1) +
		uint64(in[2])*(uint64(in2[0])<<0)
	tmp[3] = uint64(in[0])*(uint64(in2[3])<<0) +
		uint64(in[1])*(uint64(in2[2])<<0) +
		uint64(in[2])*(uint64(in2[1])<<0) +
		uint64(in[3])*(uint64(in2[0])<<0)
	tmp[4] = uint64(in[0])*(uint64(in2[4])<<0) +
		uint64(in[1])*(uint64(in2[3])<<1) +
		uint64(in[2])*(uint64(in2[2])<<0) +
		uint64(in[3])*(uint64(in2[1])<<1) +
		uint64(in[4])*(uint64(in2[0])<<0)
	tmp[5] = uint64(in[0])*(uint64(in2[5])<<0) +
		uint64(in[1])*(uint64(in2[4])<<0) +
		uint64(in[2])*(uint64(in2[3])<<0) +
		uint64(in[3])*(uint64(in2[2])<<0) +
		uint64(in[4])*(uint64(in2[1])<<0) +
		uint64(in[5])*(uint64(in2[0])<<0)
	tmp[6] = uint64(in[0])*(uint64(in2[6])<<0) +
		uint64(in[1])*(uint64(in2[5])<<1) +
		uint64(in[2])*(uint64(in2[4])<<0) +
		uint64(in[3])*(uint64(in2[3])<<1) +
		uint64(in[4])*(uint64(in2[2])<<0) +
		uint64(in[5])*(uint64(in2[1])<<1) +
		uint64(in[6])*(uint64(in2[0])<<0)
	tmp[7] = uint64(in[0])*(uint64(in2[7])<<0) +
		uint64(in[1])*(uint64(in2[6])<<0) +
		uint64(in[2])*(uint64(in2[5])<<0) +
		uint64(in[3])*(uint64(in2[4])<<0) +
		uint64(in[4])*(uint64(in2[3])<<0) +
		uint64(in[5])*(uint64(in2[2])<<0) +
		uint64(in[6])*(uint64(in2[1])<<0) +
		uint64(in[7])*(uint64(in2[0])<<0)
	// tmp[8] has the greatest value but doesn't overflow. See logic in
	// p256Square.
	tmp[8] = uint64(in[0])*(uint64(in2[8])<<0) +
		uint64(in[1])*(uint64(in2[7])<<1) +
		uint64(in[2])*(uint64(in2[6])<<0) +
		uint64(in[3])*(uint64(in2[5])<<1) +
		uint64(in[4])*(uint64(in2[4])<<0) +
		uint64(in[5])*(uint64(in2[3])<<1) +
		uint64(in[6])*(uint64(in2[2])<<0) +
		uint64(in[7])*(uint64(in2[1])<<1) +
		uint64(in[8])*(uint64(in2[0])<<0)
	tmp[9] = uint64(in[1])*(uint64(in2[8])<<0) +
		uint64(in[2])*(uint64(in2[7])<<0) +
		uint64(in[3])*(uint64(in2[6])<<0) +
		uint64(in[4])*(uint64(in2[5])<<0) +
		uint64(in[5])*(uint64(in2[4])<<0) +
		uint64(in[6])*(uint64(in2[3])<<0) +
		uint64(in[7])*(uint64(in2[2])<<0) +
		uint64(in[8])*(uint64(in2[1])<<0)
	tmp[10] = uint64(in[2])*(uint64(in2[8])<<0) +
		uint64(in[3])*(uint64(in2[7])<<1) +
		uint64(in[4])*(uint64(in2[6])<<0) +
		uint64(in[5])*(uint64(in2[5])<<1) +
		uint64(in[6])*(uint64(in2[4])<<0) +
		uint64(in[7])*(uint64(in2[3])<<1) +
		uint64(in[8])*(uint64(in2[2])<<0)
	tmp[11] = uint64(in[3])*(uint64(in2[8])<<0) +
		uint64(in[4])*(uint64(in2[7])<<0) +
		uint64(in[5])*(uint64(in2[6])<<0) +
		uint64(in[6])*(uint64(in2[5])<<0) +
		uint64(in[7])*(uint64(in2[4])<<0) +
		uint64(in[8])*(uint64(in2[3])<<0)
	tmp[12] = uint64(in[4])*(uint64(in2[8])<<0) +
		uint64(in[5])*(uint64(in2[7])<<1) +
		uint64(in[6])*(uint64(in2[6])<<0) +
		uint64(in[7])*(uint64(in2[5])<<1) +
		uint64(in[8])*(uint64(in2[4])<<0)
	tmp[13] = uint64(in[5])*(uint64(in2[8])<<0) +
		uint64(in[6])*(uint64(in2[7])<<0) +
		uint64(in[7])*(uint64(in2[6])<<0) +
		uint64(in[8])*(uint64(in2[5])<<0)
	tmp[14] = uint64(in[6])*(uint64(in2[8])<<0) +
		uint64(in[7])*(uint64(in2[7])<<1) +
		uint64(in[8])*(uint64(in2[6])<<0)
	tmp[15] = uint64(in[7])*(uint64(in2[8])<<0) +
		uint64(in[8])*(uint64(in2[7])<<0)
	tmp[16] = uint64(in[8]) * (uint64(in2[8]) << 0)

	p256ReduceDegree(out, tmp)
	return out
}

// Invert calculates |out| = |in|^{-1}
//
// Based on Fermat's Little Theorem:
//
//	a^p = a (mod p)
//	a^{p-1} = 1 (mod p)
//	a^{p-2} = a^{-1} (mod p)
func (out *p256Element) Invert(in *p256Element) *p256Element {
	var ftmp, ftmp2 p256Element

	// each e_I will hold |in|^{2^I - 1}
	var e2, e4, e8, e16, e32, e64 p256Element

	ftmp.Square(in)     // 2^1
	ftmp.Mul(in, &ftmp) // 2^2 - 2^0
	e2.Set(&ftmp)
	ftmp.Square(&ftmp)   // 2^3 - 2^1
	ftmp.Square(&ftmp)   // 2^4 - 2^2
	ftmp.Mul(&ftmp, &e2) // 2^4 - 2^0
	e4.Set(&ftmp)
	ftmp.Square(&ftmp)   // 2^5 - 2^1
	ftmp.Square(&ftmp)   // 2^6 - 2^2
	ftmp.Square(&ftmp)   // 2^7 - 2^3
	ftmp.Square(&ftmp)   // 2^8 - 2^4
	ftmp.Mul(&ftmp, &e4) // 2^8 - 2^0
	e8.Set(&ftmp)
	for i := 0; i < 8; i++ {
		ftmp.Square(&ftmp)
	} // 2^16 - 2^8
	ftmp.Mul(&ftmp, &e8) // 2^16 - 2^0
	e16.Set(&ftmp)
	for i := 0; i < 16; i++ {
		ftmp.Square(&ftmp)
	} // 2^32 - 2^16
	ftmp.Mul(&ftmp, &e16) // 2^32 - 2^0
	e32.Set(&ftmp)
	for i := 0; i < 32; i++ {
		ftmp.Square(&ftmp)
	} // 2^64 - 2^32
	e64.Set(&ftmp)
	ftmp.Mul(&ftmp, in) // 2^64 - 2^32 + 2^0
	for i := 0; i < 192; i++ {
		ftmp.Square(&ftmp)
	} // 2^256 - 2^224 + 2^192

	ftmp2.Mul(&e64, &e32) // 2^64 - 2^0
	for i := 0; i < 16; i++ {
		ftmp2.Square(&ftmp2)
	} // 2^80 - 2^16
	ftmp2.Mul(&ftmp2, &e16) // 2^80 - 2^0
	for i := 0; i < 8; i++ {
		ftmp2.Square(&ftmp2)
	} // 2^88 - 2^8
	ftmp2.Mul(&ftmp2, &e8) // 2^88 - 2^0
	for i := 0; i < 4; i++ {
		ftmp2.Square(&ftmp2)
	} // 2^92 - 2^4
	ftmp2.Mul(&ftmp2, &e4) // 2^92 - 2^0
	ftmp2.Square(&ftmp2)   // 2^93 - 2^1
	ftmp2.Square(&ftmp2)   // 2^94 - 2^2
	ftmp2.Mul(&ftmp2, &e2) // 2^94 - 2^0
	ftmp2.Square(&ftmp2)   // 2^95 - 2^1
	ftmp2.Square(&ftmp2)   // 2^96 - 2^2
	ftmp2.Mul(&ftmp2, in)  // 2^96 - 3

	return out.Mul(&ftmp2, &ftmp) // 2^256 - 2^224 + 2^192 + 2^96 - 3
}

// p256MinusOneEncoding is the encoding of -1 mod p, so p - 1, the
// highest canonical encoding. It is used by SetBytes to check for non-canonical
// encodings such as p + k, 2p + k, etc.
var p256MinusOneEncoding = new(p256Element).Sub(
	new(p256Element), new(p256Element).One()).Bytes()

// SetBytes sets e = v, where v is a big-endian 32-byte encoding, and returns e.
// If v is not 32 bytes or it encodes a value higher than 2^256 - 2^224 + 2^192 + 2^96 - 1,
// SetBytes returns nil and an error, and e is unchanged.
func (e *p256Element) SetBytes(v []byte) (*p256Element, error) {
	if len(v) != p256ElementLen {
		return nil, errors.New("invalid p256Element encoding")
	}
	for i := range v {
		if v[i] < p256MinusOneEncoding[i] {
			break
		}
		if v[i] > p256MinusOneEncoding[i] {
			return nil, errors.New("invalid p256Element encoding")
		}
	}

	offset := 0
	for i := range e {
		byteOff := offset / 8
		bitOff := offset % 8

		// Copy five bytes (we might need up to 29 + 7 = 36 bits = 4.5 bytes)
		// into a big-endian uint64 buffer from the right offset.
		buf := make([]byte, 8)
		for i := 0; i < 5; i++ {
			vIdx := len(v) - 1 - byteOff - i
			bufIdx := len(buf) - 1 - i
			if vIdx < 0 {
				break
			}
			buf[bufIdx] = v[vIdx]
		}

		e[i] = uint32(binary.BigEndian.Uint64(buf[:]) >> bitOff)
		if i%2 == 0 {
			e[i] &= bottom29Bits
			offset += 29
		} else {
			e[i] &= bottom28Bits
			offset += 28
		}
	}

	// This implementation operates in the Montgomery domain where a is
	// represented as a×R mod p. Mul calculates (a × b × R⁻¹) mod p. RR is R in
	// the domain, or R×R mod p, thus Mul(x, RR) gives x×R, i.e. converts
	// x into the Montgomery domain.
	RR := &p256Element{0xc, 0, 0x1ffffe00, 0x0fffbfff, 0x1ffeffff, 0x0fffffff, 0x1effffff, 0x03ffffff, 1}
	return e.Mul(e, RR), nil
}

// Bytes returns the 32-byte big-endian encoding of e.
func (e *p256Element) Bytes() []byte {
	// This function is outlined to make the allocations inline in the caller
	// rather than happen on the heap.
	var out [p256ElementLen]byte
	return e.bytes(&out)
}

func (e *p256Element) bytes(out *[p256ElementLen]byte) []byte {
	// We have e = a×R (an element in the Montgomery domain) and want a mod P.
	// First, we do a Montgomery multiplication (that calculates a × b × R⁻¹) by R⁻¹,
	// which is 1 outside the domain as R⁻¹×R = 1, getting a×R × 1 × R⁻¹ = a.
	// Then, we need to carry the limb overflows, and reduce the result modulo P.

	t := new(p256Element).Mul(e, &p256Element{1, 0, 0, 0, 0, 0, 0, 0, 0})
	p256TightReduce(t)

	var window uint64
	var consumedBits, nextLimb int
	for i := range out {
		if (i+1)*8 > consumedBits {
			window |= uint64(t[nextLimb]) << (consumedBits - i*8)
			if nextLimb%2 == 0 {
				consumedBits += 29
			} else {
				consumedBits += 28
			}
			nextLimb += 1
		}
		out[len(out)-1-i] = byte(window)
		window >>= 8
	}

	return out[:]
}

// On entry: in[0,2,..] < 2**30, in[1,3,...] < 2**29.
// On exit: inout[0,2,4,6] < 2**29, inout[1,3,5,7,8] < 2**28, inout < p.
func p256TightReduce(inout *p256Element) {
	// Start with a carry chain up to 256 bits (meaning the last limb is cut at
	// 28 instead of 29 bits). At the end carry < 2³.
	carry := uint32(0)
	for i := 0; i < len(inout)-1; i++ {
		inout[i] += carry
		carry = inout[i] >> 29
		inout[i] &= bottom29Bits

		i++

		inout[i] += carry
		carry = inout[i] >> 28
		inout[i] &= bottom28Bits
	}
	inout[8] += carry
	carry = inout[8] >> 28
	inout[8] &= bottom28Bits

	// Now use the reduction identity to add back the carry from 2²⁵⁶.
	//
	//   c * 2²⁵⁶ + a = c * 2²⁵⁶ + a - c * p = c * (2²⁵⁶ - p) + a  (mod p)
	//   c * (2²⁵⁶ - p) + a = c * (2²²⁴ - 2¹⁹² - 2⁹⁶ + 1) + a      (mod p)
	//
	// Note that c * (2²²⁴ - 2¹⁹² - 2⁹⁶ + 1) + a < 2p because c < 2³ and a < 2²⁵⁶.
	//
	// The carry gets added/subtracted at the following offsets:
	//
	// 		+     1       [0] + 0 bits
	//      -    2⁹⁶      [3] + 10 bits
	//      -    2¹⁹²     [6] + 21 bits
	//      +    2²²⁴     [7] + 24 bits
	//
	inout[0] += carry

	inout[1] += inout[0] >> 29
	inout[0] &= bottom29Bits

	inout[2] += inout[1] >> 28
	inout[1] &= bottom28Bits

	inout[3] += inout[2] >> 29
	inout[2] &= bottom29Bits

	// If the addition above made [3] overflow, then it will go back with this
	// subtraction. From here the addition carry chain turns into a subtraction
	// borrow chain.
	var b uint32
	inout[3], b = bits.Sub32(inout[3], carry<<10, 0)
	inout[3] += b << 28

	inout[4], b = bits.Sub32(inout[4], 0, b)
	inout[4] += b << 29

	inout[5], b = bits.Sub32(inout[5], 0, 0)
	inout[5] += b << 28

	inout[6], b = bits.Sub32(inout[6], carry<<21, b)
	inout[6] += b << 29

	// Back to an addition chain. If b is one, then carry is not zero, so this
	// can't underflow.
	inout[7] = inout[7] + carry<<24 - b

	inout[8] += inout[7] >> 28
	inout[7] &= bottom28Bits

	// Finally, a conditional subtraction to bring the term from < 2p to < p.
	var t p256Element
	p := p256Element{0x1fffffff, 0xfffffff, 0x1fffffff, 0x3ff, 0, 0, 0x200000, 0xf000000, 0xfffffff}
	t[0], b = bits.Sub32(inout[0], p[0], b)
	t[0] += b << 29
	t[1], b = bits.Sub32(inout[1], p[1], b)
	t[1] += b << 28
	t[2], b = bits.Sub32(inout[2], p[2], b)
	t[2] += b << 29
	t[3], b = bits.Sub32(inout[3], p[3], b)
	t[3] += b << 28
	t[4], b = bits.Sub32(inout[4], p[4], b)
	t[4] += b << 29
	t[5], b = bits.Sub32(inout[5], p[5], b)
	t[5] += b << 28
	t[6], b = bits.Sub32(inout[6], p[6], b)
	t[6] += b << 29
	t[7], b = bits.Sub32(inout[7], p[7], b)
	t[7] += b << 28
	t[8], b = bits.Sub32(inout[8], p[8], b)
	t[8] += b << 29
	inout.Select(inout, &t, int(b))
}

// mask32Bits returns 0xffffffff if cond is 1, and 0 otherwise.
func mask32Bits(cond int) uint32 { return ^(uint32(cond) - 1) }

// Select sets out to a if cond == 1, and to b if cond == 0.
func (out *p256Element) Select(a, b *p256Element, cond int) *p256Element {
	m := mask32Bits(cond)
	out[0] = (m & a[0]) | (^m & b[0])
	out[1] = (m & a[1]) | (^m & b[1])
	out[2] = (m & a[2]) | (^m & b[2])
	out[3] = (m & a[3]) | (^m & b[3])
	out[4] = (m & a[4]) | (^m & b[4])
	out[5] = (m & a[5]) | (^m & b[5])
	out[6] = (m & a[6]) | (^m & b[6])
	out[7] = (m & a[7]) | (^m & b[7])
	out[8] = (m & a[8]) | (^m & b[8])
	return out
}
