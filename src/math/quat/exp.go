// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

import "math"

// Exp returns e**q, the base-e exponential of q.
func Exp(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Exp(w))
	}
	v := Abs(uv)
	e := math.Exp(w)
	s, c := math.Sincos(v)
	return lift(e*c) + scale(e*s/v, uv)
}

// Log returns the natural logarithm of q.
func Log(q quaternion256) quaternion256 {
	w, uv := split(q)
	if uv == 0 {
		return lift(math.Log(w))
	}

	v := Abs(uv)
	return lift(math.Log(Abs(q))) + scale(math.Atan2(v, w)/v, uv)
}

// Pow return q**r, the base-q exponential of r.
// For generalized compatibility with math.Pow:
//      Pow(0, ±0) returns 1+0i+0j+0k
//      Pow(0, c) for real(c)<0 returns Inf+0i+0j+0k if imag(c), jmag(c), kmag(c) are zero,
//          otherwise Inf+Inf i+Inf j+Inf k.
func Pow(q, r quaternion256) quaternion256 {
	if q == 0 {
		w, uv := split(r)
		switch {
		case w == 0:
			return 1
		case w < 0:
			if uv == 0 {
				return quaternion(math.Inf(1), 0, 0, 0)
			}
			return Inf()
		case w > 0:
			return 0
		}
		panic("not reached")
	}
	return Exp(Log(q) * r)
}

// Sqrt returns the square root of q.
func Sqrt(q quaternion256) quaternion256 {
	if q == 0 {
		return 0
	}
	return Pow(q, 0.5)
}
