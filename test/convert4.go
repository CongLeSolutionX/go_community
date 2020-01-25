// run

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test conversion from slice to array pointer.

package main

func main() {
	s := make([]byte, 1, 8)
	a8 := (*[8]byte)(s)
	if &a8[0] != &s[0] {
		panic("a8 conversion failed")
	}
	a9 := (*[9]byte)(s)
	if a9 != nil {
		panic("a9 conversion succeeded")
	}

	var n []byte
	a0 := (*[0]byte)(n)
	if a0 != nil {
		panic("a0 should be nil")
	}

	z := make([]byte, 0, 0)
	a00 := (*[0]byte)(z)
	if a00 == nil {
		panic("a00 should be non-nil")
	}

	// Test with named types
	type Slice []int
	type Int4 [4]int
	type PInt4 *[4]int
	ii := make(Slice, 4)
	a4 := (*Int4)(ii)
	if &a4[0] != &ii[0] {
		panic("a4 conversion failed")
	}
	pa4 := PInt4(ii)
	if &pa4[0] != &ii[0] {
		panic("pa4 conversion failed")
	}
}

// test static variable conversion

var (
	ss  = make([]string, 10)
	s5  = (*[5]string)(ss)
	s15 = (*[15]string)(ss)

	ns  []string
	ns0 = (*[0]string)(ns)

	zs  = make([]string, 0)
	zs0 = (*[0]string)(zs)
)

func init() {
	if &ss[0] != &s5[0] {
		panic("s5 conversion failed")
	}
	if s15 != nil {
		panic("s15 should be nil")
	}
	if ns0 != nil {
		panic("ns0 should be nil")
	}
	if zs0 == nil {
		panic("zs0 should not be nil")
	}
}
