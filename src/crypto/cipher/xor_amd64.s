// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


#include "textflag.h"

#define DST BX // dst address
#define SRC0 SI // a address
#define SRC1 CX // b address
#define LEN DX // len of vect
#define POS R8 // loop16b position in vect

#define SRC0_VAL R13
#define SRC1_VAL R14

// func xorBytesSSE2(dst, a, b []byte, n int)
TEXT ·xorBytesSSE2(SB), NOSPLIT, $0
	MOVQ  d+0(FP), DST
	MOVQ  s0+24(FP), SRC0
	MOVQ  s1+48(FP), SRC1
	MOVQ  l+72(FP), LEN
	TESTQ $15, LEN        // AND 15 & len
	JNZ   not_aligned     // if not zero jump to not_aligned

aligned:
	MOVQ $0, POS

loop16b:
	MOVOU (SRC0)(POS*1), X0
	MOVOU (SRC1)(POS*1), X1
	PXOR  X1, X0
	MOVOU X0, (DST)(POS*1)
	ADDQ  $16, POS
	CMPQ  LEN, POS
	JNE   loop16b
	RET

loop_1b:
	SUBQ  $1, LEN
	MOVB  (SRC0)(LEN*1), SRC0_VAL
	MOVB  (SRC1)(LEN*1), SRC1_VAL
	XORB  SRC1_VAL, SRC0_VAL
	MOVB  SRC0_VAL, (DST)(LEN*1)
	TESTQ $7, LEN
	JNZ   loop_1b
	CMPQ  LEN, $0
	JE    ret
	TESTQ $15, LEN
	JZ    aligned

not_aligned:
	TESTQ $7, LEN                 // AND $7 & len
	JNE   loop_1b                 // if not zero jump to loop_1b
	SUBQ  $8, LEN
	MOVQ  (SRC0)(LEN*1), SRC0_VAL
	MOVQ  (SRC1)(LEN*1), SRC1_VAL
	XORQ  SRC1_VAL, SRC0_VAL
	MOVQ  SRC0_VAL, (DST)(LEN*1)
	CMPQ  LEN, $16
	JGE   aligned

ret:
	RET
