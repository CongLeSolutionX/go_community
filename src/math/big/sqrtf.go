// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import "math"

// Sqrt sets z to the rounded square root of x, and returns it.
//
// If z's precision is 0, it is changed to x's precision before the
// operation. Rounding is performed according to z's precision and
// rounding mode.
//
// The function panics if z < 0. The value of z is undefined in that
// case.
func (z *Float) Sqrt(x *Float) *Float {

	if x.Sign() == -1 {
		// following IEEE754-2008 (section 7.2)
		panic(ErrNaN{"square root with negative operand"})
	}

	// IEEE754-2008 requires √±0 = ±0
	if x.Sign() == 0 {
		z.form = zero
		z.neg = x.neg
		z.acc = Exact
		return z
	}

	// √+∞ = +∞
	if x.IsInf() {
		z.form = inf
		z.neg = false
		z.acc = Exact
		return z
	}

	prec := z.Prec()
	if prec == 0 {
		prec = x.Prec()
	}

	b := x.MantExp(z)
	z.SetPrec(prec)

	// Compute √(z·2ᵇ) as
	//   √( z)·2**(½b)      if b is even
	//   √(2z)·2**(½b)      if b > 0 is odd
	//   √(½z)·2**(½b)      if b < 0 is odd
	switch b % 2 {
	case 0:
		// nothing to do
	case 1:
		z.Mul(NewFloat(2.0), z)
	case -1:
		z.Mul(NewFloat(0.5), z)
	}
	// 0.25 <= z < 2.0

	// Solving x² - z = 0 directly requires a Quo call, but it's
	// faster for small precisions.
	//
	// Solving 1/x² - z = 0 avoids the Quo call and is much faster for
	// high precisions.
	//
	if prec <= 128 {
		z.sqrtDirect(z)
	} else {
		z.sqrtInverse(z)
	}

	// re-attach halved exponent
	return z.SetMantExp(z, b/2)
}

// Compute √x (to z.prec precision) by solving
//   1/t² - x = 0
// for t (using Newton's method), and then inverting.
func (z *Float) sqrtInverse(x *Float) *Float {
	one, nh := NewFloat(1), NewFloat(-0.5)

	// let
	//  f(t) = 1/t² - x
	// then
	//   g(t) = f(t)/f'(t) = -½t(1 - xt²)
	u := new(Float)
	g := func(t *Float) *Float {
		u.prec = t.prec
		u.Mul(t, t)   // u = t²
		u.Mul(x, u)   //   = xt²
		u.Sub(one, u) //   = 1 - xt²
		u.Mul(nh, u)  //   = -½(1 - xt²)
		u.Mul(t, u)   //   = -½t(1 - xt²)
		return u
	} // g(t) = -½t(1 - xt²)

	xf, _ := x.Float64()
	sqi := NewFloat(1 / math.Sqrt(xf))
	for p := sqi.prec; p < 2*(z.prec+32); p *= 2 {
		sqi.prec = p
		sqi.Sub(sqi, g(sqi))
	}
	// sqi = 1/√x

	return z.Mul(x, sqi) // x/√x = √x
}

// Compute √x (up to prec 128) by solving
//   t² - x = 0
// for t, using one or two iterations of Newton's method.
func (z *Float) sqrtDirect(x *Float) *Float {
	half := NewFloat(0.5)

	// let
	//  f(t) = t² - x
	// then
	//   g(t) = f(t)/f'(t) = ½(t² - x)/t
	u := new(Float)
	g := func(t *Float) *Float {
		u.prec = t.prec
		u.Mul(t, t)    // u = t²
		u.Sub(u, x)    //   = t² - x
		u.Mul(half, u) //   = ½(t² - x)
		u.Quo(u, t)    //   = ½(t² - x)/t
		return u
	} // g(t) = ½(t² - x)/t

	xf, _ := x.Float64()
	sq := NewFloat(math.Sqrt(xf))
	sq.prec *= 2

	switch {
	case z.prec >= 64:
		sq.Sub(sq, g(sq))
		fallthrough
	default:
		sq.Sub(sq, g(sq))
	}

	return z.Set(sq)
}
