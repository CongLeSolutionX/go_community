// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT errors(SB), $0
	// Check updated error messages to avoid regressions.
	MOVL (AX)(R10*1), DX             // ERROR "invalid index register for 32-bit mode in (AX)(R10*1)"
	MOVL $1, (AX)(R10*1)             // ERROR "invalid index register for 32-bit mode in (AX)(R10*1)"
	MOVL $1, (R10)(AX*1)             // ERROR "invalid base register for 32-bit mode in (R10)(AX*1)"
	VPSLLDQ $1, X0, X11              // ERROR "invalid register for 32-bit mode in X11"
	VPGATHERQQ Y2, (BP)(Y8*2), Y1    // ERROR "invalid index register for 32-bit mode in (BP)(Y8*2)"
	MOVQ $1, AX                      // ERROR "can't encode in 32-bit mode"
	PADDSB M0, M1                    // ERROR "can't encode in 32-bit mode"
	PSLLL (AX), M0                   // ERROR "can't encode in 32-bit mode"
	RET
