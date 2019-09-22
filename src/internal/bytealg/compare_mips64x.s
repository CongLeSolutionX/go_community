// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips64 mips64le

#include "go_asm.h"
#include "textflag.h"

TEXT ·Compare(SB),NOSPLIT,$0-56
	MOVV	a_base+0(FP), R3
	MOVV	b_base+24(FP), R4
	MOVV	a_len+8(FP), R1
	MOVV	b_len+32(FP), R2
	MOVV	$ret+48(FP), R9
	JMP	cmpbody<>(SB)

TEXT runtime·cmpstring(SB),NOSPLIT,$0-40
	MOVV	a_base+0(FP), R3
	MOVV	b_base+16(FP), R4
	MOVV	a_len+8(FP), R1
	MOVV	b_len+24(FP), R2
	MOVV	$ret+32(FP), R9
	JMP	cmpbody<>(SB)

// On entry:
// R1 length of a
// R2 length of b
// R3 points to the start of a
// R4 points to the start of b
// R9 points to the return value (-1/0/1)
//
// On exit:
// R6, R7, R8, R10, R11, R12 clobbered
TEXT cmpbody<>(SB),NOSPLIT|NOFRAME,$0
	BEQ	R3, R4, samebytes // same start of a and b

	SGTU	R1, R2, R7
	MOVV	R1, R10
	CMOVN	R7, R2, R10	// R10 is min(R1, R2)
	ADDV	R3, R10, R8	// R3 start of a, R8 end of a
	BEQ	R3, R8, samebytes // length is 0

	SRLV	$4, R10		// R10 is number of chunks
	BEQ	$0, R10, byte_loop

	// make sure both a and b are aligned.
	OR	R3, R4, R11
	AND	$0x7, R11
	BNE	$0, R11, byte_loop

chunk16_loop:
	BEQ	$0, R10, byte_loop
	MOVV	(R3), R6
	MOVV	(R4), R7
	BNE	R6, R7, chunk_diff
	MOVV	8(R3), R13
	MOVV	8(R4), R14
	ADDV	$16, R3
	ADDV	$16, R4
	SUBVU	$1, R10
	BEQ	R13, R14, chunk16_loop
	MOVV	R13, R6
	MOVV	R14, R7
chunk_diff:
	//find diff
	XOR	R12, R12
	XOR	R6, R7, R10
chunk_diff_loop:
	SRLV	R12, R10, R11
	AND	$0xff, R11
	ADDV	$8, R12
	BEQ	$0, R11, chunk_diff_loop
	SUBV	$8, R12
	SRLV	R12, R6
	SRLV	R12, R7
	AND	$0xff, R6 // mask off
	AND	$0xff, R7
	JMP	byte_cmp

byte_loop:
	BEQ	R3, R8, samebytes
	MOVBU	(R3), R6
	ADDVU	$1, R3
	MOVBU	(R4), R7
	ADDVU	$1, R4
	BEQ	R6, R7, byte_loop

byte_cmp:
	SGTU	R6, R7, R8
	MOVV	$-1, R6
	CMOVZ	R8, R6, R8
	JMP	ret

samebytes:
	SGTU	R1, R2, R6
	SGTU	R2, R1, R7
	SUBV	R7, R6, R8

ret:
	MOVV	R8, (R9)
	RET
