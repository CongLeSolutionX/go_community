// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quat

// Conj returns the quaternion conjugate of q.
func Conj(q quaternion256) quaternion256 {
	return quaternion(real(q), -imag(q), -jmag(q), -kmag(q))
}

// Inv returns the quaternion inverse of q.
func Inv(q quaternion256) quaternion256 {
	x := Abs(q)
	return scale(1/(x*x), Conj(q))
}
