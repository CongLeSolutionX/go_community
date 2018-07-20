// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define DST BX // dst address
#define A SI // a address
#define B CX // b address
#define LEN DX // len of slice

// func xorBytesSSE2(dst, a, b []byte, n int)
TEXT ·xorBytesSSE2(SB), NOSPLIT, $0
	MOVQ  dst+0(FP), DST
	MOVQ  a+24(FP), A
	MOVQ  b+48(FP), B
	MOVQ  len+72(FP), LEN
	TESTQ $15, LEN        // AND 15 & len
	JNZ   not_aligned     // if not zero jump to not_aligned

aligned:
	MOVQ $0, R8 // position in slice

loop16b:
	MOVOU (A)(R8*1), X0
	MOVOU (B)(R8*1), X1
	PXOR  X1, X0
	MOVOU X0, (DST)(R8*1)
	ADDQ  $16, R8
	CMPQ  LEN, R8
	JNE   loop16b
	RET

loop_1b:
	SUBQ  $1, LEN
	MOVB  (A)(LEN*1), R13
	MOVB  (B)(LEN*1), R14
	XORB  R14, R13
	MOVB  R13, (DST)(LEN*1)
	TESTQ $7, LEN
	JNZ   loop_1b
	CMPQ  LEN, $0
	JE    ret
	TESTQ $15, LEN
	JZ    aligned

not_aligned:
	TESTQ $7, LEN           // AND $7 & len
	JNE   loop_1b           // if not zero jump to loop_1b
	SUBQ  $8, LEN
	MOVQ  (A)(LEN*1), R13
	MOVQ  (B)(LEN*1), R14
	XORQ  R14, R13
	MOVQ  R13, (DST)(LEN*1)
	CMPQ  LEN, $16
	JGE   aligned

ret:
	RET
