// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

func inlinedMul(x float32) {
	// stack frame setup for "test" call should happen after call to runtime.fmul32
	// right now the test is using the fact that first available register is allocated
	// to hold runtime.fmul32 result, so constant $1 is loaded into second allocatable register

	// arm/5:-"MOVW\t[$]1, R0", "MOVW\t[$]1, R1"
	// mips/softfloat:-"MOVW\t[$]1, R1", "MOVW\t[$]1, R2"
	// mips64/softfloat:-"MOVV\t[$]1, R1", "MOVV\t[$]1, R2"
	test(1, x*x)
}

//go:noinline
func test(id int32, a float32) {}
