// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides Go implementations of elementary multi-precision
// arithmetic operations on word vectors. These have the suffix _g.
// These are needed for platforms without assembly implementations of these routines.
// This file also contains elementary operations that can be implemented
// sufficiently efficiently in Go.

package big

import "math/bits"

// A Word represents a single digit of a multi-precision unsigned integer.
type Word uint

const (
	_S = _W / 8 // word size in bytes

	_W = bits.UintSize // word size in bits
	_B = 1 << _W       // digit base
	_M = _B - 1        // digit mask
)

// Many of the loops in this file are of the form
//   for i := 0; i < len(z) && i < len(x) && i < len(y); i++
// i < len(z) is the real condition.
// However, checking i < len(x) && i < len(y) as well is faster than
// having the compiler do a bounds check in the body of the loop;
// remarkably it is even faster than hoisting the bounds check
// out of the loop, by doing something like
//   _, _ = x[len(z)-1], y[len(z)-1]
// There are other ways to hoist the bounds check out of the loop,
// but the compiler's BCE isn't powerful enough for them (yet?).
// See the discussion in CL 164966.

// ----------------------------------------------------------------------------
// Elementary operations on words
//
// These operations are used by the vector operations below.

// z1<<_W + z0 = x*y
func mulWW_g(x, y Word) (z1, z0 Word) {
	hi, lo := bits.Mul(uint(x), uint(y))
	return Word(hi), Word(lo)
}

// z1<<_W + z0 = x*y + c
func mulAddWWW_g(x, y, c Word) (z1, z0 Word) {
	hi, lo := bits.Mul(uint(x), uint(y))
	var cc uint
	lo, cc = bits.Add(lo, uint(c), 0)
	return Word(hi + cc), Word(lo)
}

// nlz returns the number of leading zeros in x.
// Wraps bits.LeadingZeros call for convenience.
func nlz(x Word) uint {
	return uint(bits.LeadingZeros(uint(x)))
}

// The resulting carry c is either 0 or 1.
func addVV_g(z, x, y []Word) (c Word) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Add(uint(x[i]), uint(y[i]), uint(c))
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// The resulting carry c is either 0 or 1.
func subVV_g(z, x, y []Word) (c Word) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Sub(uint(x[i]), uint(y[i]), uint(c))
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// The resulting carry c is either 0 or 1.
func addVW_g(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		zi, cc := bits.Add(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// addVWlarge is addVW, but intended for large z.
// The only difference is that we check on every iteration
// whether we are done with carries,
// and if so, switch to a much faster copy instead.
// This is only a good idea for large z,
// because the overhead of the check and the function call
// outweigh the benefits when z is small.
func addVWlarge(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		if c == 0 {
			copy(z[i:], x[i:])
			return
		}
		zi, cc := bits.Add(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

func subVW_g(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		zi, cc := bits.Sub(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// subVWlarge is to subVW as addVWlarge is to addVW.
func subVWlarge(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		if c == 0 {
			copy(z[i:], x[i:])
			return
		}
		zi, cc := bits.Sub(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

func shlVU_g(z, x []Word, s uint) (c Word) {
	if s == 0 {
		copy(z, x)
		return
	}
	if len(z) == 0 {
		return
	}
	s &= _W - 1 // hint to the compiler that shifts by s don't need guard code
	ŝ := _W - s
	ŝ &= _W - 1 // ditto
	c = x[len(z)-1] >> ŝ
	for i := len(z) - 1; i > 0; i-- {
		z[i] = x[i]<<s | x[i-1]>>ŝ
	}
	z[0] = x[0] << s
	return
}

func shrVU_g(z, x []Word, s uint) (c Word) {
	if s == 0 {
		copy(z, x)
		return
	}
	if len(z) == 0 {
		return
	}
	s &= _W - 1 // hint to the compiler that shifts by s don't need guard code
	ŝ := _W - s
	ŝ &= _W - 1 // ditto
	c = x[0] << ŝ
	for i := 0; i < len(z)-1; i++ {
		z[i] = x[i]>>s | x[i+1]<<ŝ
	}
	z[len(z)-1] = x[len(z)-1] >> s
	return
}

func mulAddVWW_g(z, x []Word, y, r Word) (c Word) {
	c = r
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		c, z[i] = mulAddWWW_g(x[i], y, c)
	}
	return
}

func addMulVVW_g(z, x []Word, y Word) (c Word) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		z1, z0 := mulAddWWW_g(x[i], y, z[i])
		lo, cc := bits.Add(uint(z0), uint(c), 0)
		c, z[i] = Word(cc), Word(lo)
		c += z1
	}
	return
}

// q = (x1<<_W + x0 - r)/v. inverse = ( _B^2 - 1 ) / d - _B. d = y << nlz(y).
func divWW(x1, x0, y Word, inv uint) (q, r Word) {
	shift := nlz(y)
	if shift != 0 {
		x1 = (x1<<shift | x0>>(_W-shift))
		x0 <<= shift
		y <<= shift
	}
	d := uint(y)
	qq, q0 := bits.Mul(uint(x1), inv)   // multipy inverse instead of dividing
	q0, cc := bits.Add(q0, uint(x0), 0) // add dividend once to correct quotient
	qq, cc = bits.Add(qq, uint(x1), cc)

	rr := uint(x0) - d*qq
	rr -= d
	qq++ // plus one to ensure quotient not less than real quotient

	if rr >= q0 { // conditionally adjust quotient
		qq--
		rr += d
	}

	if rr >= d {
		qq++
		rr -= d
	}
	rr >>= shift
	return Word(qq), Word(rr)
}

func divWVW(z []Word, xn Word, x []Word, y Word) (r Word) {
	r = xn
	if len(x) == 1 {
		qq, rr := bits.Div(uint(r), uint(x[0]), uint(y))
		z[0] = Word(qq)
		return Word(rr)
	}
	inv := getInvert(y)
	for i := len(z) - 1; i >= 0; i-- {
		z[i], r = divWW(r, x[i], y, inv)
	}
	return r
}

// getInvert return the inverse of the divisor. inv = floor(( _B^2 - 1 ) / u1 - _B). u1 = d1 << nlz(d1).
func getInvert(d1 Word) (inv uint) {
	// ( _B^2 - 1) / u1 - _B = ( _B * ^u1 + _M) / u1.
	const halfSize = _W >> 1
	const mask = uint((1 << halfSize) - 1)
	shift := nlz(d1)
	u1 := uint(d1) << shift
	ul := u1 & mask      // low bits of u1
	uh := u1 >> halfSize // high bits of u1

	qh := ^u1 / uh // high bits of the quotient, which is floor((( _B ^ 1/2) * ^u1 + mask) / u1). similar with bits.go:533

	r := ((^u1 - qh*uh) << halfSize) | mask // let b = ( _B ^ 1/2) , the remainder r =  b * (^u) + b-1 - qh * (b * uh + ul) = b (^u - qh * uh) + b-1 - qh * ul

	// conditionally adjust qh to ensure that the remainder is positive.
	// As u1 = d1 << nlz(d1), 2 * u1 >= _B > qh * ul, we need to subtract one from qh at most twice.
	p := qh * ul
	if r < p {
		qh--
		r += u1
		if r >= u1 { // if no overflow occurred
			if r < p {
				qh--
				r += u1
			}
		}
	}
	r -= p

	// low half of quotient, which is floor((( _B ^ 1/2) * r +mask) / u1).
	p = (r>>halfSize)*qh + r  // the trick that qh is a suitable inverse of uh now. same as p = ( r / uh) << halfSize (bits.go:544)
	ql := (p >> halfSize) + 1 // only need the high bits of p, plus one to ensure ql not less than real ql.

	r = (r << halfSize) + mask - ql*u1 // real remainder now. we don't need the high halfSize bits of remainder

	if r >= (p << halfSize) { // conditionally adjust ql
		ql--
		r += u1
	}
	if r >= u1 { // if no overflow occurred
		ql++
		r -= u1
	}

	inv = (qh << halfSize) + ql
	return
}
