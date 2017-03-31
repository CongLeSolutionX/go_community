// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

import "math"

// IsNaN returns true if any of real(q), imag(q), jmag(q), or kmag(q) is NaN
// and none are an infinity.
func IsNaN(q quaternion256) bool {
	switch {
	case math.IsInf(real(q), 0) || math.IsInf(imag(q), 0) || math.IsInf(jmag(q), 0) || math.IsInf(kmag(q), 0):
		return false
	case math.IsNaN(real(q)) || math.IsNaN(imag(q)) || math.IsNaN(jmag(q)) || math.IsNaN(kmag(q)):
		return true
	}
	return false
}

// NaN returns a quaternion ``not-a-number'' value.
func NaN() quaternion256 {
	nan := math.NaN()
	return quaternion(nan, nan, nan, nan)
}
