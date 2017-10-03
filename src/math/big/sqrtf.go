// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import "math"

// Sqrt sets z to the rounded square root of x, and returns z.
//
// If z's precision is 0, it is changed to x's precision before the
// operation. Rounding is performed according to z's precision and
// rounding mode.
//
// The function panics if z < 0. The value of z is undefined in that
// case.
func (z *Float) Sqrt(x *Float) *Float {

	if x.Sign() == -1 {
		// following IEEE754-2008 rules (section 7.2)
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

	// Solving x² - z = 0 directly requires a Quo call, but it's
	// faster for small precisions.
	//
	// Solving 1/x² - z = 0 avoids the Quo call and is much faster for
	// high precisions.
	//
	// TODO: add fast path with direct method for low prec
	// TODO: find optimal direct/inverse threshold

	z.sqrtInverse(z) // always inverse for now

	acc := z.acc
	// re-attach the exponent and return
	z.SetMantExp(z, b/2)
	z.acc = acc // SetMantExp destroys acc
	return z
}

// Compute √x (to z.Prec() precision) by solving
//   1/t² - x = 0
// for t (using Newton's method), and then inverting.
func (z *Float) sqrtInverse(x *Float) *Float {
	// let
	//  f(t) = 1/t² - x
	// then
	//   g(t) = f(t)/f'(t) = -½t(1 - xt²)
	one, nh := NewFloat(1), NewFloat(-0.5)
	g := func(t *Float) *Float {
		u := new(Float)
		u.Mul(t, t)        // u = t²
		u.Mul(x, u)        //   = xt²
		u.Sub(one, u)      //   = 1 - xt²
		u.Mul(nh, u)       //   = -½(1 - xt²)
		return u.Mul(t, u) //   = -½t(1 - xt²)
	} // g(t) = -½t(1 - xt²)

	// use float64 for the initial guess
	// TODO: protect from bad guess
	xf, _ := x.Float64()
	sqi := NewFloat(1 / math.Sqrt(xf))

	// There's another operation after newton, so we need to request
	// higher prec to have a few guard digits.
	sqi.newton(g, x.Prec()+32) // sqi = 1/√x
	z.Mul(x, sqi)              // z = x/√x = √x
	return z
}

// Given a function g(t) = f(t)/f'(t), newton returns a solution to
//    f(t) = 0
// to precision prec, using the Newton Method with the receiver as the
// initial point.
//
// g must not mutate its argument.
func (z *Float) newton(g func(x *Float) *Float, prec uint) *Float {
	for p := z.Prec(); p < 2*prec; p *= 2 {
		z.SetPrec(p)
		z.Sub(z, g(z))
	}

	return z.SetPrec(prec)
}

// func (z *Float) sqrtDirect(x *Float) *Float {
//
// 	xc := new(Float).Copy(z)
// 	prec := z.Prec()
//
// 	half := NewFloat(0.5)
// 	g := func(t *Float) *Float {
// 		u := new(Float).Mul(t, t) // u = t²
// 		u.Sub(u, xc)              //   = t² - x
// 		u.Mul(half, u)            //   = ½(t² - x)
// 		return u.Quo(u, t)        //   = ½(t² - x)/t
// 	}
//
// 	// initial guess
// 	zf, _ := z.Float64()
// 	z.SetFloat64(math.Sqrt(zf)).SetPrec(53 + 64)
// 	switch {
// 	case prec >= 64:
// 		z.Sub(z, g(z))
// 		fallthrough
// 	default:
// 		z.Sub(z, g(z))
// 	}
//
// 	return z.SetPrec(prec)
// }
