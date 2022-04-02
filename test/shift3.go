// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that the compiler's noder uses the correct type
// for RHS shift operands that are untyped. Must compile;
// run for good measure.

package main

import "fmt"

func f(x, y int) {
	if x != y {
		panic(fmt.Sprintf("%d != %d", x, y))
	}
}

func main() {
	var x int = 1
	f(x<<1, 2)
	f(x<<1., 2)
	f(x<<(1+0i), 2)
	f(x<<0i, 1)

	f(x<<(1<<x), 4)
	f(x<<(1.<<x), 4)
	f(x<<((1+0i)<<x), 4)
	f(x<<(0i<<x), 1)

	// corner cases
	const maxUint64 = 9223372036854775807
	f(x<<maxUint64, 0)
	f(x<<(maxUint64+1), 0)

	const fmaxUint64 = 9223372036854775807.0
	f(x<<fmaxUint64, 0)
	f(x<<(fmaxUint64+1), 0)

	const cmaxUint64 = 9223372036854775807.0 + 0i
	f(x<<cmaxUint64, 0)
	f(x<<(cmaxUint64+1), 0)
}
