// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// On riscv64, the float32 becomes a float64
// when loaded in a register, different from
// other platforms. These functions are
// needed to ensure correct conversions on riscv64

// Convert float32->uint64
TEXT ·archFloat32ToReg(SB),NOSPLIT,$0-16
	MOVF	val+0(FP), F1
	MOVD	F1, ret+8(FP)
	RET

// Convert uint64->float32
TEXT ·archFloat32FromReg(SB),NOSPLIT,$0-12
	MOVD	reg+0(FP), F1
	// Normally a float64->float32 conversion
	// would need rounding, but that is not needed
	// here since the uint64 was originally converted
	// from float32, and should be avoided to
	// preserve SNaN values.
	MOVF	F1, ret+8(FP)
	RET

