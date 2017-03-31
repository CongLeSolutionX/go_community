// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

import "math"

// Sin returns the sine of q.
func Sin(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Sin(w))
	}
	v := Abs(uv)
	s, c := math.Sincos(w)
	sh, ch := sinhcosh(v)
	return lift(s*ch) + scale(c*sh/v, uv)
}

// Sinh returns the hyperbolic sine of q.
func Sinh(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Sinh(w))
	}
	v := Abs(uv)
	s, c := math.Sincos(v)
	sh, ch := sinhcosh(w)
	return lift(c*sh) + scale(s*ch/v, uv)
}

// Cos returns the cosine of q.
func Cos(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Cos(w))
	}
	v := Abs(uv)
	s, c := math.Sincos(w)
	sh, ch := sinhcosh(v)
	return lift(c*ch) - scale(s*sh/v, uv)
}

// Cosh returns the hyperbolic cosine of q.
func Cosh(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Cosh(w))
	}
	v := Abs(uv)
	s, c := math.Sincos(v)
	sh, ch := sinhcosh(w)
	return lift(c*ch) + scale(s*sh/v, uv)
}

// Tan returns the tangent of q.
func Tan(q quaternion256) quaternion256 {
	d := Cos(q)
	if d == 0 {
		return Inf()
	}
	return Sin(q) * Inv(d)
}

// Tanh returns the hyperbolic tangent of q.
func Tanh(q quaternion256) quaternion256 {
	d := Cosh(q)
	if d == 0 {
		return Inf()
	}
	return Sinh(q) * Inv(d)
}

// Asin returns the inverse sine of q.
func Asin(q quaternion256) quaternion256 {
	_, uv := split(q)
	if uv == 0 {
		return lift(math.Asin(real(q)))
	}
	u := unit(uv)
	return -u * Log(u*q+Sqrt(1-q*q))
}

// Asinh returns the inverse hyperbolic sine of q.
func Asinh(q quaternion256) quaternion256 {
	return Log(q + Sqrt(1+q*q))
}

// Acos returns the inverse cosine of q.
func Acos(q quaternion256) quaternion256 {
	w, uv := split(Asin(q))
	return lift(math.Pi/2-w) - uv
}

// Acosh returns the inverse hyperbolic cosine of q.
func Acosh(q quaternion256) quaternion256 {
	w := Acos(q)
	_, uv := split(w)
	if uv == 0 {
		return w
	}
	w *= unit(uv)
	if real(w) < 0 {
		w = -w
	}
	return w
}

// Atan returns the inverse targent of q.
func Atan(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Atan(w))
	}
	u := unit(uv)
	return lift(0.5) * u * Log((u+q)*Inv(u-q))
}

// Atanh returns the inverse hyperbolic tangent of q.
func Atanh(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Atanh(w))
	}
	u := unit(uv)
	return -u * Atan(u*q)
}

// calculate sinh and cosh
func sinhcosh(x float64) (sh, ch float64) {
	if math.Abs(x) <= 0.5 {
		return math.Sinh(x), math.Cosh(x)
	}
	e := math.Exp(x)
	ei := 0.5 / e
	e *= 0.5
	return e - ei, e + ei
}
