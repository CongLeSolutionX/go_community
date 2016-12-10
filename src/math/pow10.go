// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

var pow10tab = [...]float64{
	1.0e00, 1.0e01, 1.0e02, 1.0e03, 1.0e04, 1.0e05, 1.0e06, 1.0e07, 1.0e08, 1.0e09,
	1.0e10, 1.0e11, 1.0e12, 1.0e13, 1.0e14, 1.0e15, 1.0e16, 1.0e17, 1.0e18, 1.0e19,
	1.0e20, 1.0e21, 1.0e22, 1.0e23, 1.0e24, 1.0e25, 1.0e26, 1.0e27, 1.0e28, 1.0e29,
	1.0e30, 1.0e31,
}

var pow10postab32 = [...]float64{
	1.0e00, 1.0e32, 1.0e64, 1.0e96, 1.0e128, 1.0e160, 1.0e192, 1.0e224, 1.0e256, 1.0e288,
}

var pow10negtab32 = [...]float64{
	1.0e-00, 1.0e-32, 1.0e-64, 1.0e-96, 1.0e-128, 1.0e-160, 1.0e-192, 1.0e-224, 1.0e-256, 1.0e-288,
	1.0e-320,
}

// Pow10 returns 10**n, the base-10 exponential of n.
//
// Special cases are:
//	Pow10(n) =    0 for n < -323
//	Pow10(n) = +Inf for n > 308
func Pow10(n int) float64 {
	if 0 <= n && n <= 308 {
		return pow10tab[n%32] * pow10postab32[n/32]
	}

	if -323 <= n && n < 0 {
		return pow10negtab32[-n/32] / pow10tab[-n%32]
	}

	if n > 308 {
		return Inf(1)
	}

	// n < -323
	return 0
}
