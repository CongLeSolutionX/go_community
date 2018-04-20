// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT errors(SB),$0
	// TODO: fix invalid printing of "foo<>(SB)(AX)". It's currently printed as "foo<>(AX)".
	MOVL	foo<>(SB)(AX), AX	// ERROR "global vars can't use index registers in foo<>(AX)"
	MOVL	(AX)(SP*1), AX		// ERROR "can't use SP as the index register in (AX)(SP*1)"
	EXTRACTPS $4, X2, (BX)          // ERROR "illegal arguments combination"
	EXTRACTPS $-1, X2, (BX)         // ERROR "illegal arguments combination"
	// VSIB addressing does not permit non-vector (X/Y)
	// scaled index register.
	VPGATHERDQ X12,(R13)(AX*2), X11 // ERROR "illegal arguments combination"
	VPGATHERDQ X2, 664(BX*1), X1    // ERROR "illegal arguments combination"
	VPGATHERDQ Y2, (BP)(AX*2), Y1   // ERROR "illegal arguments combination"
	VPGATHERDQ Y5, 664(DX*8), Y6    // ERROR "illegal arguments combination"
	VPGATHERDQ Y5, (DX), Y0         // ERROR "illegal arguments combination"
	// VM/X rejects Y index register.
	VPGATHERDQ Y5, 664(Y14*8), Y6   // ERROR "illegal arguments combination"
	VPGATHERQQ X2, (BP)(Y7*2), X1   // ERROR "illegal arguments combination"
	// VM/Y rejects X index register.
	VPGATHERQQ Y2, (BP)(X7*2), Y1   // ERROR "illegal arguments combination"
	VPGATHERDD Y5, -8(X14*8), Y6    // ERROR "illegal arguments combination"
	// No VSIB for legacy instructions.
	MOVL (AX)(X0*1), AX             // ERROR "illegal arguments combination"
	MOVL (AX)(Y0*1), AX             // ERROR "illegal arguments combination"
	// AVX2GATHER mask/index/dest #UD cases.
	VPGATHERQQ Y2, (BP)(X2*2), Y2   // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERQQ Y2, (BP)(X2*2), Y7   // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERQQ Y2, (BP)(X7*2), Y2   // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERQQ Y7, (BP)(X2*2), Y2   // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERDQ X2, 664(X2*8), X2    // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERDQ X2, 664(X2*8), X7    // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERDQ X2, 664(X7*8), X2    // ERROR "mask, index, and destination registers should be distinct"
	VPGATHERDQ X7, 664(X2*8), X2    // ERROR "mask, index, and destination registers should be distinct"
	// Non-X0 for Yxr0 should produce an error
	BLENDVPD X1, (BX), X2           // ERROR "illegal arguments combination"
	// Check offset overflow. Must fit in int32.
	MOVQ 2147483647+1(AX), AX       // ERROR "offset too large in 2147483648(AX)"
	MOVQ 3395469782(R10), R8        // ERROR "offset too large in 3395469782(R10)"
	LEAQ 3395469782(AX), AX         // ERROR "offset too large in 3395469782(AX)"
	ADDQ 3395469782(AX), AX         // ERROR "offset too large in 3395469782(AX)"
	ADDL 3395469782(AX), AX         // ERROR "offset too large in 3395469782(AX)"
	ADDW 3395469782(AX), AX         // ERROR "offset too large in 3395469782(AX)"
	LEAQ 433954697820(AX), AX       // ERROR "offset too large in 433954697820(AX)"
	ADDQ 433954697820(AX), AX       // ERROR "offset too large in 433954697820(AX)"
	ADDL 433954697820(AX), AX       // ERROR "offset too large in 433954697820(AX)"
	ADDW 433954697820(AX), AX       // ERROR "offset too large in 433954697820(AX)"
	// Check updated error messages to avoid regressions.
	MOVL AX, (AX)(DX)               // ERROR "scaling must be explicit and equal to 1/2/4/8 in (AX)(DX)"
	MOVQ DX, 2147483647+1(AX)       // ERROR "offset too large in 2147483648(AX)"
	VPEXTRW $2, X0, 2147483648(AX)  // ERROR "offset too large in 2147483648(AX)"
	MOVL (AX)(F0*1), DX             // ERROR "invalid index register in (AX)(F0*1)"
	MOVL $1, (AX)(F0*1)             // ERROR "invalid index register in (AX)(F0*1)"
	MOVL (F0)(AX*1), DX             // ERROR "invalid base register in (F0)(AX*1)"
	MOVL $1, (F0)(AX*1)             // ERROR "invalid base register in (F0)(AX*1)"
	PUSHL AX                        // ERROR "can't encode in 64-bit mode"
	POPL AX                         // ERROR "can't encode in 64-bit mode"
	RET
