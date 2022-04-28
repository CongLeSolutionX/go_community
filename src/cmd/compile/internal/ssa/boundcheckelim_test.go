// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test for bound check elimination

package ssa

import (
	"testing"
)

var B []byte

//go:noinline
func test_geq(b []byte) {
	_ = len(b) >= 3 && b[len(b)-3] == '#' // OPT (IN BOUND)
	_ = len(b) >= 3 && b[len(b)-2] == '#' // OPT (IN BOUND)
	_ = len(b) >= 3 && b[2] == '#'        // SIMPLIFIED ON PROVE

	_ = len(b) >= 3 && b[3] == '#'        // NO OPT (UNKNOWN)
	_ = len(b) >= 3 && b[4] == '#'        // NO OPT (UNKNOWN)
	_ = len(b) >= 3 && b[len(b)-4] == '#' // NO OPT (UNKNOWN)
	_ = len(b) >= 3 && b[len(b)+1] == '#' // NO OPT (UNKNOWN)
	_ = len(b) >= 3 && b[10] == '#'       // NO OPT (UNKNOWN)
}

//go:noinline
func test_ge(b []byte) {
	_ = len(b) > 3 && b[len(b)-3] == '#' // OPT (IN BOUND)
	_ = len(b) > 3 && b[len(b)-2] == '#' // OPT (IN BOUND)
	_ = len(b) > 3 && b[3] == '#'        // SIMPLIFIED ON PROVE

	_ = len(b) > 3 && b[4] == '#'        // NO OPT (UNKNOWN)
	_ = len(b) > 3 && b[len(b)-4] == '#' // NO OPT (UNKNOWN)
	_ = len(b) > 3 && b[len(b)+1] == '#' // NO OPT (UNKNOWN)
	_ = len(b) > 3 && b[10] == '#'       // NO OPT (UNKNOWN)
}

//go:noinline
func test_le(b []byte) {
	_ = len(b) < 3 && b[len(b)-3] == '#' // NO OPT (UNKNOWN)
	_ = len(b) < 3 && b[len(b)-2] == '#' // NO OPT (UNKNOWN)
	_ = len(b) < 3 && b[3] == '#'        // SIMPLIFIED ON PROVE
}

// NOTE these tests lead to out-of-bound exception diring testing
//go:noinline
/*func test_leq(b []byte) {
	_ = len(b) <= 3 && b[4] == '#'         // SIMPLIFIED ON PROVE
	_ = len(b) <= 3 && b[len(b)-2] == '#'  // NO OPT (UNKNOWN)
	_ = len(b) <= 3 && b[2] == '#'         // NO OPT (UNKNOWN)
}*/

// NOTE in golang in function (*builder).tagGroupLabel there is an
// example of OpEq operation that is no simplified with PROVE but
// it is not reproduced in short example
//
//go:noinline
func test_eq(b []byte) {
	_ = len(b) == 3 && b[0] == '#' // SIMPLIFIED ON PROVE
	_ = len(b) == 3 && b[2] == '#' // SIMPLIFIED ON PROVE
	_ = len(b) == 3 && b[3] == '#' // SIMPLIFIED ON PROVE
	_ = len(b) == 3 && b[4] == '#' // SIMPLIFIED ON PROVE
}

func TestBoundCheckElimination(t *testing.T) {
	test_geq(B)
	test_ge(B)
	//	test_leq(B)
	test_eq(B)
}
