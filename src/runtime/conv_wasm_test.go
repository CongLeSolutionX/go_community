// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"testing"
)

var res int64
var ures uint64

func TestFloatTruncation(t *testing.T) {
	for _, n := range []float64{
		9223372036854775807, // max +- 1
		9223372036854775808,
		9223372036854775806,

		9223372036854775295, // trunc point +- 1
		9223372036854775296,
		9223372036854775294,

		18446744073709551615, // umax +- 1
		18446744073709551616,
		18446744073709551614,

		18446744073709550591, // umax trunc +- 1
		18446744073709550593,
		18446744073709550590,
	} {
		res = int64(n)
		ures = uint64(n)
	}
}
