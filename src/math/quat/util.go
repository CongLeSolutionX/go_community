// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

func lift(v float64) quaternion256 {
	return quaternion(v, 0, 0, 0)
}

func scale(s float64, q quaternion256) quaternion256 {
	return quaternion(s*real(q), s*imag(q), s*jmag(q), s*kmag(q))
}

func split(q quaternion256) (float64, quaternion256) {
	return real(q), quaternion(0, imag(q), jmag(q), kmag(q))
}

func unit(q quaternion256) quaternion256 {
	return scale(1/Abs(q), q)
}
