// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
 
#include "../../../../../runtime/textflag.h"
 
TEXT asmtest(SB),DUPOK|NOSPLIT,$0
start:
    // RISCV Bitmanip Extension
	// Zba
	ADDUW       X10, X11, X12           // 3b86a508
	ADDUW       X10, X11				// bb85a508
	SH1ADD      X11, X12, X13           // b326b620
	SH1ADD      X11, X12				// 3326b620
	SH1ADDUW    X12, X13, X14           // 3ba7c620
	SH1ADDUW    X12, X13				// bba6c620
	SH2ADD      X13, X14, X15           // b347d720
	SH2ADD      X13, X14				// 3347d720
	SH2ADDUW    X14, X15, X16           // 3bc8e720
	SH2ADDUW    X14, X15				// bbc7e720
	SH3ADD      X15, X16, X17           // b368f820
	SH3ADD      X15, X16				// 3368f820
	SH3ADDUW    X16, X17, X18           // 3be90821
	SH3ADDUW    X16, X17				// bbe80821
	SLLIUW      $31, X17, X18           // 1b99f809
	SLLIUW      $63, X17				// 9b98f80b
	SLLIUW      $63, X17, X18           // 1b99f80b
	SLLIUW      $1, X18, X19            // 9b191908
	// pseudo-instructions
	MOVWU	X19, X20					// 3b8a0908

	// Zbb
	ANDN    X19, X20, X21               // b37a3a41
	ANDN    X19, X20					// 337a3a41
	CLZ     X20, X21                    // 931a0a60
	CLZW    X21, X22                    // 1b9b0a60
	CPOP    X22, X23                    // 931b2b60
	CPOPW   X23, X24                    // 1b9c2b60
	CTZ     X24, X25                    // 931c1c60
	CTZW    X25, X26                    // 1b9d1c60
	MAX     X26, X28, X29               // b36eae0b
	MAX     X26, X28					// 336eae0b
	MAXU    X28, X29, X30               // 33ffce0b
	MAXU    X28, X29					// b3fece0b
	MIN     X29, X30, X5                // b342df0b
	MIN     X29, X30					// 334fdf0b
	MINU    X30, X5, X6                 // 33d3e20b
	MINU    X30, X5						// b3d2e20b
	ORCB    X5, X6                      // 13d37228
	ORN     X6, X7, X8                  // 33e46340
	ORN     X6, X7						// b3e36340
	REV8    X7, X8                      // 13d4836b
	ROL     X8, X9, X10                 // 33958460
	ROL     X8, X9						// b3948460
	ROLW    X9, X10, X11                // bb159560
	ROLW    X9, X10						// 3b159560
	ROR     X10, X11, X12               // 33d6a560
	ROR     X10, X11					// b3d5a560
	ROR     $63, X11					// 93d5f563
	RORI    $63, X11, X12               // 13d6f563
	RORI    $1, X12, X13                // 93561660
	RORIW   $31, X13, X14               // 1bd7f661
	RORIW   $1, X14, X15                // 9b571760
	RORW    X15, X16, X17               // bb58f860
	RORW    X15, X16					// 3b58f860
	RORW    $31, X13					// 9bd6f661
	SEXTB   X16, X17                    // 93184860
	SEXTH   X17, X18                    // 13995860
	XNOR    X18, X19, X20               // 33ca2941
	XNOR    X18, X19					// b3c92941
	ZEXTH   X19, X20                    // 3bca0908
	// pseudo-instructions
	MOVB	X5, X6						// 13934260
	MOVH	X6, X7						// 93135360
	MOVHU	X7, X8						// 3bc40308

	// Zbs
	BCLR    X23, X24, X25               // b31c7c49
	BCLR    $63, X24					// 131cfc4b
	BCLRI   $1, X25, X26                // 139d1c48
	BEXT    X26, X28, X29               // b35eae49
	BEXT    $63, X28					// 135efe4b
	BEXTI   $1, X29, X30                // 13df1e48
	BINV    X30, X5, X6                 // 3393e269
	BINV    $63, X6						// 1313f36b
	BINVI   $1, X7, X8                  // 13941368
	BSET    X8, X9, X10                 // 33958428
	BSET    $63, X9						// 9394f42b
	BSETI   $1, X10, X11                // 93151528
