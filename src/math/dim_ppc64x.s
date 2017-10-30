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
// return max of x-y, 0
TEXT ·Dim(SB),NOSPLIT,$0
	MOVD	$0x7FF, R5  // Gen +Inf
	SLD	$52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	CMPU	R3, R4      // Args equal? 
	BNE	dim3        // no
	CMPU	R3, R5      // yes: both +Inf?
	BEQ	retNaN      // return NaN
dim2:	// (-Inf, -Inf) special case
	MOVD	$0xFFF, R5  // Gen -Inf
	SLD	$52, R5
	CMPU	R3, R5      // both -Inf?
	BEQ	retNaN      // return NaN
dim3:	// Not Inf or -Inf
	MTVSRD	R3, F1
	MTVSRD	R4, F2
	FCMPU	F1, F2      // Check for NaN
	BVS	retNaN      // one is NaN
	BGT	sub	    // x>y: do sub
	MOVD	R0,ret+16(FP)  // ret 0
	RET
sub:	FSUB	F2, F1, F3  // substract
	FMOVD	F3, ret+16(FP)  //return diff
	RET
retNaN:	MOVD    $0x7FF8, R4 // Gen NaN
	SLD     $48, R4
	OR	$1, R4
	MOVD	R4, ret+16(FP) // return NaN
	RET

// func ·Max(x, y float64) float64
TEXT ·Max(SB),NOSPLIT,$0 
	MOVD    $0x7FF, R5  // Gen +Inf
	SLD     $52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	CMPU	R3, R5      // x = +Inf?
	BEQ	retInf      // return +Inf
	MTVSRD  R3, VS1     // vs1 overlaps f1
	CMP	R4, R5      // y = +Inf?
	BEQ	retInf      // return +Inf
	MTVSRD	R4, VS2     // vs2 overlaps f2
	FCMPU   F1, F2
	BVS	retNaN      // one is NaN
	BGT     retx        // x>y, return x
	BEQ	testsign0   // x==y, check -0
	FMOVD   F2,ret+16(FP) // return y
	RET
retx:	MOVD	R3, ret+16(FP) // return x
	RET
retNaN:	MOVD	$0x7FF8, R4 // Gen NaN
	SLD	$48, R4
	OR	$1, R4
	MOVD	R4, ret+16(FP) // return NaN
	RET
retInf: // return +Inf
	MOVD	R5, ret+16(FP) // return +Inf
	RET
	// If here then the arguments appear to be equal,
	// but could be +/- 0 so test for that.
testsign0:
	CMPU	R3,R0          // Check for +0
	BEQ	done           // first arg +0
	MOVD	R4, ret+16(FP)  // first arg -0
	RET
done:
	MOVD	R3, ret+16(FP)
	RET

// func Min(x, y float64) float64
TEXT ·Min(SB),NOSPLIT,$0
	MOVD    $0xFFF, R5    // Gen -Inf
	SLD     $52, R5
	MOVD	x+0(FP), R3
	MOVD	y+8(FP), R4
	MTVSRD	R3, VS1       // vs1 overlaps f1
	CMPU	R5, R3
	BEQ	retNegInf     // return -Inf
	MTVSRD	R4, VS2       // vs2 overlaps f2
	CMPU	R5, R4        
	BEQ	retNegInf     // return -Inf
	// normal case
	FCMPU	F1, F2
	BVS	retNaN        // one is NaN
	BLT	retx          // x<y, return x
	BEQ	testsign0     // x==y, check -0
	FMOVD	F2, ret+16(FP) // return y
	RET
retx:	FMOVD	F1, ret+16(FP) // return x
	RET
retNegInf: // return -Inf
	MOVD	R5, ret+16(FP) // return -Inf
	RET
retNaN:	MOVD    $0x7FF8, R4    // gen NaN
	SLD     $48, R4
	OR	$1, R4
        MOVD    R4, ret+16(FP) // return NaN
	RET
testsign0:
	CMPU    R3, R4         // Test equality
	BEQ     done    
	CMPU    R3,R0          // Check for -0
	BNE     done           // first arg -0
	MOVD	R4, ret+16(FP)  // return +0
	RET
done:
	MOVD	R3, ret+16(FP) // return -0
	RET
