// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_fp "runtime/internal/fp"
)

func isposinf(f float64) bool { return f > _fp.MaxFloat64 }
func isneginf(f float64) bool { return f < -_fp.MaxFloat64 }

func neginf() float64 {
	var f float64 = _fp.MaxFloat64
	return -f * f
}

func complex128div(n complex128, d complex128) complex128 {
	// Special cases as in C99.
	ninf := isposinf(real(n)) || isneginf(real(n)) ||
		isposinf(imag(n)) || isneginf(imag(n))
	dinf := isposinf(real(d)) || isneginf(real(d)) ||
		isposinf(imag(d)) || isneginf(imag(d))

	nnan := !ninf && (_fp.Isnan(real(n)) || _fp.Isnan(imag(n)))
	dnan := !dinf && (_fp.Isnan(real(d)) || _fp.Isnan(imag(d)))

	switch {
	case nnan || dnan:
		return complex(_fp.Nan(), _fp.Nan())
	case ninf && !dinf:
		return complex(_fp.Posinf(), _fp.Posinf())
	case !ninf && dinf:
		return complex(0, 0)
	case real(d) == 0 && imag(d) == 0:
		if real(n) == 0 && imag(n) == 0 {
			return complex(_fp.Nan(), _fp.Nan())
		} else {
			return complex(_fp.Posinf(), _fp.Posinf())
		}
	default:
		// Standard complex arithmetic, factored to avoid unnecessary overflow.
		a := real(d)
		if a < 0 {
			a = -a
		}
		b := imag(d)
		if b < 0 {
			b = -b
		}
		if a <= b {
			ratio := real(d) / imag(d)
			denom := real(d)*ratio + imag(d)
			return complex((real(n)*ratio+imag(n))/denom,
				(imag(n)*ratio-real(n))/denom)
		} else {
			ratio := imag(d) / real(d)
			denom := imag(d)*ratio + real(d)
			return complex((imag(n)*ratio+real(n))/denom,
				(imag(n)-real(n)*ratio)/denom)
		}
	}
}
