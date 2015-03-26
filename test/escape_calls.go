// errorcheck -0 -m -l

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for function parameters.

// In this test almost everything is BAD except the simplest cases
// where input directly flows to output.

package foo

func f(buf []byte) []byte { // ERROR "leaking param: buf to result ~r1 level=0$"
	return buf
}

func g(*byte) string

func h(e int) {
	var x [32]byte // ERROR "moved to heap: x$"
	g(&f(x[:])[0]) // ERROR "&f\(x\[:\]\)\[0\] escapes to heap$" "x escapes to heap$"
}
