// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
 
#include "../../../../../runtime/textflag.h"
 
TEXT errors(SB),$0
start:
	// RISCV Bitmanip Extension
	// Zba
	SLLIUW	$64, X10, X11				// ERROR "shift amount out of range 0 to 63"

	// Zbb
	RORIW	$32, X16, X17				// ERROR "shift amount out of range 0 to 31"
	RORI    $64, X11, X12               // ERROR "shift amount out of range 0 to 63"

	// Zbs
	BCLRI   $64, X24, X25               // ERROR "shift amount out of range 0 to 63"
	BEXTI   $64, X28, X29               // ERROR "shift amount out of range 0 to 63"
	BINVI   $64, X6, X7                 // ERROR "shift amount out of range 0 to 63"
	BSETI   $64, X9, X10                // ERROR "shift amount out of range 0 to 63"

	RET
