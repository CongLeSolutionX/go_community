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
//		because realQ/B<1, X/B<B, (B^2-1)/d <= 2*(B^2-1)/B <= 2*B
//		So (qq,q0) >= realQ*B-1-B-2*B+B+1 = (realQ-2)*B. So qq >= realQ-2
//	   Then we run qq++, and we get realQ-1 <= qq <= realQ+1.
//
// Q3: Why and how should we adjust qq to make it the quotient we want?
// A3: First calculate rr = x0-d*qq = (B*x1+x0)-d*qq mod B. Note that rr doesn't necessarily represent the real remainder.
//	  Because floor(realQ)-1 = floor(realQ-1), floor(realQ)+1 = floor(realQ+1)
//	  So floor(realQ)-1 <= qq <= floor(realQ)+1
//	  Let B^2-1 = qd*d +rd, then B*qq+q0 = inv*x1+B*x1+x0+B = x1*qd+x0+B.
//	  (1) if qq=floor(realQ)-1, then real remainder d <= r < 2b. We first prove r can only be [d,B), can't be [B, 2b).
//			prove1-1: r = B*x1+x0-qq*d < B <==> qq > (B*x1+x0-B)/d
//					because qq = (x1*qd+x0+B-q0)/B
//					so qq > (B*x1+x0-B)/d <==>  x1*(B^2-1-rd)+x0d+B*d-q0d-B^2*x1-B*x0+B^2>0
//					<==> B^2-(B-d)x0-(1+rd)*x1+B*d-q0*d>0
//					because x0<B, (1+rd)*x1<d^2
//					so B^2-(B-d)x0-(1+rd)*x1+B*d-q0*d>0 <== 2B*d -d^2 -q0*d>0 <==> 2B > d+q0. Prove over.
//			Then we prove rr < q0.
//			prove1-2: because d <= r < B, so rr = r = Bx1+x0-qq*d, q0=x1*qd+x0+B-qq*B
//					so q0 - rr = x1*(qd-B)+B-(B-d)*qq
//					because qq=(B*x1+x0-r)/d, qd*d=B^2-1-rd
//					so q0 > rr <==> x1*(B^2-1-rd-B*d)+B*d-(B-d)(Bx1+x0-r)>0 <==> B*d-(1+rd)*x1-(B-d)(x0-r)>0
//					because (1+rd)*x1<d^2
//					so B*d-(1+rd)*x1-(B-d)(x0-r)>0 <== (B-d)(d+r-x0)>0 <==> d+r>x0. And we know r >= d >= B/2. Prove over.
//			So when qq=floor(realQ)-1, because prove1-2, we won't enter branch1.
//			Because of prove1-1, we enter branch2 and adjust qq++, then get the floor(realQ).
//	  (2) if qq=floor(realQ)+1, real remainder -d<= r < 0. Next we prove rr >= q0:
//			Prove2: because -d<= r <0
//					so rr = B+r = (B*x1+x0)-d*qq+B
//					because q0=inv*x1+B*x1+x0+B-B*qq, and qq=(B*x1+x0-r)/d
//					so rr-q0=B*x1-d*qq-inv*x1+B*qq >= (B-d)*(B*x1+x0-r)/d+(B*d-B^2-d)/d*x1
//							=[x1+(B-d)(x0-r)]/d.
//					because B-d>=0,r<0, so rr>=q0.
//			So when qq=floor(realQ)+1, we enter branch1 and adjust qq--.
//			And because 0<= rr+d <d, we won't enter branch2. Then we get the floor(realQ).
//	  (3) if qq=floor(realQ), real remainder 0<= r < d. rr = r. We may enter branch1 or not.
//	 		If r<q0, We won't enter branch1, of course won't enter branch2. Then we get floor(realQ) without adjusting.
//	 		But when r>=q0, we will enter branch1, which means we must also enter branch2. So next we prove if rr >= q0, then d< rr+d <B.
//			prove3: rr=B*x1+x0-qq*d, q0=x1*qd+x0+B-B*qq
//					because	rr>=q0
//					so qq*(B-d)-(qd-B)*x1-B>=0
//					because qq=(B*x1+x0-r)/d
//					so qq*(B-d)-(qd-B)*x1-B>=0 <==> (B*x1+x0-r)(B-d)-(B^2-1-rd-B*d)*x1-B*d>=0 <==> (1+rd)*x1+(B-d)(x0-r)>=B*d
//						<==> r <= x0-[B*d-(1+rd)*x1]/(B-d)
//					so rr+d<B <==> r+d<B <== (x0+d)(B-d)+(1+rd)*x1<B^2
//					because 1+rd<=d, x1<d, x0<B
//					so rr+d<B <== (B-d)*B+B*d<=B^2 <==> 0 <= 0. Prove over.
//			So when r<q0, we adjust nothing. When r>=q0, we first adjust q--, then adjust q++. Anyway we get the floor(realQ).
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

	if rr >= q0 { // branch1
		qq--
		rr += d
	}

	if rr >= d { // branch2
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

// getInvert return the inverse of the divisor. inv = floor(( _B^2 - 1 ) / u - _B). u = d1 << nlz(d1).
func getInvert(d1 Word) (inv uint) {
	u := uint(d1 << nlz(d1))
	x1 := ^u
	x0 := uint(_M)
	inv, _ = bits.Div(x1, x0, u) // (_B^2-1)/U-_B = (_B*(_M-C)+_M)/U
	return
}
