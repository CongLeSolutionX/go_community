// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

#include "textflag.h"

// Special constants generated in registers
// rather than from memory:
// PosInf 0x7FF0000000000000
// NaN    0x7FF8000000000001
// NegInf 0xFFF0000000000000

// func Dim(x, y float64) float64
TEXT ·Dim(SB),NOSPLIT,$0
	MOVD	$0x7FF, R5  // Set up Inf
	SLD	$52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	CMPU	R3, R4
	BNE	dim3	// Not the same
	CMPU	R3, R5	// Same: Inf?
	BEQ	retNaN  // yes
dim2:	// (-Inf, -Inf) special case
	MOVD	$0xFFF, R5
	SLD	$52, R5
	CMPU	R3, R5	// Same: -Inf?
	BEQ	retNaN
dim3:	// Not Inf or -Inf
	MTVSRD	R3, F1
	MTVSRD	R4, F2
	FCMPU	F1, F2
	BVS	retNaN  // one is NaN
	BGT	sub	// subtract values
	MOVD	R0,ret+16(FP)  // ret 0
	RET
sub:	FSUB	F2, F1, F3
	FMOVD	F3, ret+16(FP)
	RET
retNaN:	MOVD    $0x7FF8, R4
	SLD     $48, R4
	OR	$1, R4
	MOVD	R4, ret+16(FP)
	RET

// func ·Max(x, y float64) float64
TEXT ·Max(SB),NOSPLIT,$0
	MOVD    $0x7FF, R5
	SLD     $52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	CMPU	R3, R5
	BEQ	retInf
	CMP	R4, R5
	BEQ	retInf
	MTVSRD	R3, VS1
	MTVSRD	R4, VS2
	FCMPU   F1, F2
	BVS	retNaN  // unordered?
	BGT     retx
	BEQ	testsign0
	FMOVD   F2,ret+16(FP)
	RET
retx:	MOVD	R3, ret+16(FP)
	RET
retNaN:	MOVD	$0x7FF8, R4
	SLD	$48, R4
	OR	$1, R4
	MOVD	R4, ret+16(FP)
	RET
retInf: // return +Inf
	MOVD	R5, ret+16(FP)
	RET
	// If here then the arguments appear to be equal,
	// but could be +/- 0 so test for that.
testsign0:
	CMPU	R3,R0  // Check for +0
	BEQ	done  // if the first argument is +0 return that
	MOVD	R4, ret+16(FP)  // first arg must be -0
	RET
done:
	MOVD	R3, ret+16(FP)
	RET

// func Min(x, y float64) float64
TEXT ·Min(SB),NOSPLIT,$0
	MOVD    $0xFFF, R5
	SLD     $52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	MTVSRD	R3, VS1
	CMPU	R5, R3
	BEQ	retNegInf
	MTVSRD	R4, VS2
	CMPU	R5, R4
	BEQ	retNegInf
	// normal case
	FCMPU	F1, F2
	BVS	retNaN
	BLT	retx
	BEQ	testsign0
	FMOVD	F2, ret+16(FP)
	RET
retx:	FMOVD	F1, ret+16(FP)
	RET
retNegInf: // return -Inf
	MOVD	R5, ret+16(FP)
	RET
retNaN:	MOVD    $0x7FF8, R4
	SLD     $48, R4
	OR	$1, R4
        MOVD    R4, ret+16(FP)
	RET
testsign0:
	CMPU    R3, R4 // Are they really equal?
	BEQ     done
	CMPU    R3,R0  // Check for -0
	BNE     done  // if the first argument is -0 return that
	MOVD	R4, ret+16(FP)  // first arg must be +0
	RET
done:
	MOVD	R3, ret+16(FP)
	RET
