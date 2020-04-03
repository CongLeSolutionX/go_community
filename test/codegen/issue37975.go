// asmcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// Make sure 32-bit integer does not extend to 64-bit in comparison
// See golang.go/cl/226737
func f(x int) {
	// 386:-"SARL"
	_ = make([]byte, x, 10)
}
