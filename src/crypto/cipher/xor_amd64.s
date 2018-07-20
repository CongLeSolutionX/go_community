// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define DST BX // dst address
#define A SI // a address
#define B CX // b address
#define LEN DX // len of job

// func xorBytesSSE2(dst, a, b []byte, n int)
TEXT ·xorBytesSSE2(SB), NOSPLIT, $0
	MOVQ  dst+0(FP), DST
	MOVQ  a+24(FP), A
	MOVQ  b+48(FP), B
	MOVQ  len+72(FP), LEN
	TESTQ $15, LEN        // AND 15 & len, if not zero jump to not_aligned.
	JNZ   not_aligned

aligned:
	MOVQ $0, R8 // position in slices

loop16b:
	MOVOU (A)(R8*1), X0   // XOR 16byte forwards.
	MOVOU (B)(R8*1), X1
	PXOR  X1, X0
	MOVOU X0, (DST)(R8*1)
	ADDQ  $16, R8
	CMPQ  LEN, R8
	JNE   loop16b
	RET

loop_1b:
	SUBQ  $1, LEN           // XOR 1byte backwards.
	MOVB  (A)(LEN*1), R13
	MOVB  (B)(LEN*1), R14
	XORB  R14, R13
	MOVB  R13, (DST)(LEN*1)
	TESTQ $7, LEN           // AND 7 & len, if not zero jump to loop_1b.
	JNZ   loop_1b
	CMPQ  LEN, $0           // if len is 0, ret.
	JE    ret
	TESTQ $15, LEN          // AND 15 & len, if zero jump to aligned.
	JZ    aligned

not_aligned:
	TESTQ $7, LEN           // AND $7 & len, if not zero jump to loop_1b.
	JNE   loop_1b
	SUBQ  $8, LEN           // XOR 8bytes backwards.
	MOVQ  (A)(LEN*1), R13
	MOVQ  (B)(LEN*1), R14
	XORQ  R14, R13
	MOVQ  R13, (DST)(LEN*1)
	CMPQ  LEN, $16          // if len is greater or equal 16 here, it must be aligned.
	JGE   aligned

ret:
	RET
