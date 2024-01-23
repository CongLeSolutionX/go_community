// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && (ppc64 || ppc64le)

#include "textflag.h"

// func addMulVVW1024(z, x *uint, y uint) (c uint)
TEXT ·addMulVVW1024(SB), $0-32
	MOVD	$16, R6 // R6 = z_len
	JMP		addMulVVWx(SB)

// func addMulVVW1536(z, x *uint, y uint) (c uint)
TEXT ·addMulVVW1536(SB), $0-32
	MOVD	$24, R6 // R6 = z_len
	JMP		addMulVVWx(SB)

// func addMulVVW2048(z, x *uint, y uint) (c uint)
TEXT ·addMulVVW2048(SB), $0-32
	MOVD	$32, R6 // R6 = z_len
	JMP		addMulVVWx(SB)

// func addMulVVWx(z, x []Word, y Word) (c Word)
TEXT addMulVVWx(SB), NOSPLIT, $0
	MOVD z+0(FP), R3        // R10 = z[]
	MOVD x+8(FP), R4       // R8 = x[]
	MOVD y+16(FP), R5       // R9 = y

	CMP  R6, $4		// R22 = z_len
	MOVD R0, R9             // R9 = c = 0
	BLT     tail
	SRD  $2, R6, R7
	MOVD R7, CTR            // Initialize loop counter
	PCALIGN $16

loop:
	MOVD  0(R4), R14        // x[i]
	MOVD  8(R4), R16        // x[i+1]
	MOVD 16(R4), R18        // x[i+2]
	MOVD 24(R4), R20        // x[i+3]
	MOVD  0(R3), R15        // z[i]
	MOVD  8(R3), R17        // z[i+1]
	MOVD 16(R3), R19        // z[i+2]
	MOVD 24(R3), R21        // z[i+3]
	MULLD R5, R14, R10      // low x[i]*y
	MULHDU R5, R14, R11     // high x[i]*y
	ADDC    R15, R10
	ADDZE   R11
	ADDC    R9, R10
	ADDZE   R11, R9
	MULLD  R5, R16, R14     // R14 = Low-order(x[i]*y)
	MULHDU R5, R16, R15     // R15 = High-order(x[i]*y)
	ADDC    R17, R14
	ADDZE   R15
	ADDC    R9, R14
	ADDZE   R15, R9
	MULLD  R5, R18, R16
	MULHDU R5, R18, R17
	ADDC    R19, R16
	ADDZE   R17
	ADDC    R9, R16
	ADDZE   R17, R9
	MULLD  R5, R20, R18
	MULHDU R5, R20, R19
	ADDC    R21, R18
	ADDZE  R19
	ADDC    R9, R18
	ADDZE   R19, R9
	MOVD    R10, 0(R3)
	MOVD    R14, 8(R3)
	MOVD    R16, 16(R3)
	MOVD    R18, 24(R3)
	ADD     $32, R3
	ADD     $32, R4
	BDNZ    loop

	ANDCC   $3, R6
tail:
	CMP     R0, R6
	BEQ     done
	MOVD    R6, CTR
	PCALIGN $16
tailloop:
	MOVD    0(R4), R14
	MOVD    0(R3), R15
	MULLD   R5, R14, R10
	MULHDU  R5, R14, R11
	ADDC    R15, R10
	ADDZE   R11
	ADDC    R9, R10
	ADDZE   R11, R9
	MOVD    R10, 0(R3)
	ADD     $8, R3
	ADD     $8, R4
	BDNZ    tailloop

done:
	MOVD R9, c+24(FP)
	RET

