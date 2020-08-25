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

// q = ( x1 << _W + x0 - r)/y. Ensure x1<y to avoid overflow.
// inv = floor(( _B^2 - 1 ) / d - _B). d = y << nlz(y).
// From now on we equate _B with B, X = x << nlz(y), realQ = x / y = X / d, x1=X/B, x0=X mod B
//
// Q1: Why inv needs to subtract B?
// A1: Because B/2 <= d <= B-1, so B + 1 <= ( B^2 - 1) / d <= 2 * (B^2 - 1) / B ==> 1 <= ( B^2 - 1) / d - B <= ( B^2 - 2) / B
// 		==> 1 <= inv <= B-1. We can just put inv into an uint.
//
// Q2: How to get quotient use inv?
// A2: We calculate (qq,q0) = floor(X/B)*inv+X.
//	  Let realQ = X / d = (B*x1 + x0)/d. We want floor(realQ).
//	  So  (qq,q0) <= X/B*(B^2-1)/d-B*X/B+X = realQ*(B-1/B) <= realQ*B. So qq <= realQ.
//	  Meanwhile (qq,q0) >= (X/B-1)*((B^2-1)/d-B-1) = realQ*((B^2-1)/B)-X/B-(B^2-1)/d+B+1 = realQ*B-realQ/B-X/B-(B^2-1)/d+B+1
//		because reaqQ/B<1, X/B<B, (B^2-1)/d <= 2*(B^2-1)/B <= 2*B
//		So (qq,q0) >= realQ*B-1-B-2*B+B+1 = (realQ-2)*B. So qq >= realQ-2
//	  Then we run qq++. Let realQ-1 <= qq <= realQ+1
//
// Q3: Why and how should we adjust qq to make it the quotient we want?
// A3: First calculate rr = x0<< - d*qq = (B*x1+x0)-d*qq mod B.(qq has plus one). Note that rr doesn't necessarily represent the remainder.
//	  Because floor(realQ)-1 = floor(realQ-1), floor(realQ)+1 = floor(realQ+1)
//	  So floor(realQ)-1 <= qq <= floor(realQ)+1
//	  (1) if qq=floor(realQ)+1, real Remainder -d<= r < 0. Next we prove rr >= q0:
//			Prove1: because -d<= r <0
//					so rr = ( B*x1 + x0) - d*qq +B
//					because q0 = inv*x1 + B*x1 +x0 +B- B*qq, and qq=(B*x1+x0-r)/d
//					so rr-q0=B*x1 - d*qq - inv*x1 + B*qq >= (B-d)*(B*x1+x0-r)/d+(Bd-B^2-d)/d*x1
//							=[x1+(B-d)(x0-r)]/d.
//					because B-d>=0,r<0, so rr>=q0.
//		  so we adjust qq--, rr+=d. as B> rr >=B-d, rr+d mod B <d, we get the floor(realQ)
//	  (2) if qq=floor(realQ), real Remainder 0<= r < d. Next we prove when rr=r >= q0, rr+d <B:
//			prove2: rr = B*x1 +x0 -qq*d, q0=inv*x1 + B*x1 +x0 +B- B*qq,rr>=q0
//					let B^2-1 = qd*d + rd, so we can get qq*(B-d)-(qd-B)*x1-B>=0
//					because qq=(B*x1+x0-r)/d
//					so (B*x1+x0-r)(B-d)-(B^2-1-rd-Bd)*x1-Bd>=0 ==> (1+rd)*x1+(B-d)(x0-r)>=Bd
//						==> r<= x0-[Bd-(1+rd)*x1]/(B-d)
//					so r+d<=B <== (x0+d)(B-d)+(1+rd)*x1<B^2
//					because 1+rd<=d,x1<d, so (1+rd)*x1<d^2
//					so <== (B-d)x0+Bd<B^2 <== B^2-B+d <B^2 <== d < B
//			so when r<q0, we adjust nothing. When r>=q0, first adjust q--, then adjust q++, we get the floor(realQ).
//	  (3) if qq=floor(realQ)-1, we prove real Remainder can only be [d,B), can't be [B, 2b):
//			prove3: r = B*x1+x0-qq*d < B ==> qq > (B*x1+x0-B)/d
//					let B^2-1 = qd*d +rd, then qq = (x1*qd +x0 +B -q0)/B
//					so qq > (B*x1+x0-B)/d <==  x1*(B^2-1-rd)+x0d+Bd-q0d-B^2*x1-B*x0+B^2>0
//					<== B^2-(B-d)x0-(1+rd)*x1+Bd-q0*d>0
//					because x0<B, (1+rd)*x1<d^2
//					so <== 2B*d -d^2 -q0*d>0 <== 2B > d+q0. prove over.
//			so when qq=floor(realQ)-1, we just adjust qq++, then get the floor(realQ).
func divWW(x1, x0, y Word, inv uint) (q, r Word) {
	shift := nlz(y)
	if shift != 0 {
		x1 = (x1<<shift | x0>>(_W-shift))
		x0 <<= shift
		y <<= shift
	}
	d := uint(y)
	// refers to Q2
	qq, q0 := bits.Mul(uint(x1), inv)
	q0, cc := bits.Add(q0, uint(x0), 0)
	qq, _ = bits.Add(qq, uint(x1), cc)
	qq++

	//  refers to Q3.
	rr := uint(x0) - d*qq

	if rr >= q0 {
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

	// conditionally adjust ql, similar with proves in divWW.
	r = (r << halfSize) + mask - ql*u1

	if r >= (p << halfSize) {
		ql--
		r += u1
	}
	if r >= u1 {
		ql++
		r -= u1
	}

	inv = (qh << halfSize) + ql
	return
}
