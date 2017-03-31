// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

import "math"

// IsInf returns true if any of real(q), imag(q), jmag(q), or kmag(q) is an infinity.
func IsInf(q quaternion256) bool {
	return math.IsInf(real(q), 0) || math.IsInf(imag(q), 0) || math.IsInf(jmag(q), 0) || math.IsInf(kmag(q), 0)
}

// Inf returns a quaternion infinity, quaternion(+Inf, +Inf, +Inf, +Inf).
func Inf() quaternion256 {
	inf := math.Inf(1)
	return quaternion(inf, inf, inf, inf)
}
