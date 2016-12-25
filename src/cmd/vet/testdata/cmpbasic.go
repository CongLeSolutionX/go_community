// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdata

func UnsignedComparison() {
	var x uint
	if x < 0 { // ERROR "'x \(uint\) < 0' is always false"
		_ = 42
	}

	if 0 > x { // ERROR "'0 > x \(uint\)' is always false"
		_ = 42
	}

	var y uintptr
	if y < 0 { // ERROR "'y \(uintptr\) < 0' is always false"
		_ = 42
	}

	if 0 > y { // ERROR "'0 > y \(uintptr\)' is always false"
		_ = 42
	}
}
