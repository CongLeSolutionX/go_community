// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package quat provides basic constants and mathematical functions
// for quaternions.
package quat

import "math"

// Abs returns the absolute value (also called the modulus) of x.
func Abs(x quaternion256) float64 {
	r, i, j, k := real(x), imag(x), jmag(x), kmag(x)

	// special cases
	switch {
	case math.IsInf(r, 0) || math.IsInf(i, 0) || math.IsInf(j, 0) || math.IsInf(k, 0):
		return math.Inf(1)
	case math.IsNaN(r) || math.IsNaN(i) || math.IsNaN(j) || math.IsNaN(k):
		return math.NaN()
	}

	if r < 0 {
		r = -r
	}
	if i < 0 {
		i = -i
	}
	if j < 0 {
		j = -j
	}
	if k < 0 {
		k = -k
	}

	if r < i {
		r, i = i, r
	}
	if r < j {
		r, j = j, r
	}
	if r < k {
		r, k = k, r
	}

	if r == 0 {
		return 0
	}

	i /= r
	j /= r
	k /= r

	return r * math.Sqrt(1+i*i+j*j+k*k)
}
