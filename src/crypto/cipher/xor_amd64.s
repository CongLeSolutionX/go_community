// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


#include "textflag.h"

#define dst BX // dst address
#define src0 SI // a address
#define src1 CX // b address
#define len DX // len of vect
#define pos R8 // loop16b position in vect

#define src0_val R13
#define src1_val R14

// func xorBytesSSE2(dst, a, b []byte, n int)
TEXT ·xorBytesSSE2(SB), NOSPLIT, $0
	MOVQ  d+0(FP), dst
	MOVQ  s0+24(FP), src0
	MOVQ  s1+48(FP), src1
	MOVQ  l+72(FP), len
	TESTQ $15, len        // AND 15 & len
	JNZ   not_aligned     // if not zero jump to not_aligned

aligned:
	MOVQ $0, pos

loop16b:
	MOVOU (src0)(pos*1), X0
	MOVOU (src1)(pos*1), X1
	PXOR  X1, X0
	MOVOU X0, (dst)(pos*1)
	ADDQ  $16, pos
	CMPQ  len, pos
	JNE   loop16b
	RET

loop_1b:
	SUBQ  $1, len
	MOVB  (src0)(len*1), src0_val
	MOVB  (src1)(len*1), src1_val
	XORB  src1_val, src0_val
	MOVB  src0_val, (dst)(len*1)
	TESTQ $7, len
	JNZ   loop_1b
	CMPQ  len, $0
	JE    ret
	TESTQ $15, len
	JZ    aligned

not_aligned:
	TESTQ $7, len                 // AND $7 & len
	JNE   loop_1b                 // if not zero jump to loop_1b
	SUBQ  $8, len
	MOVQ  (src0)(len*1), src0_val
	MOVQ  (src1)(len*1), src1_val
	XORQ  src1_val, src0_val
	MOVQ  src0_val, (dst)(len*1)
	CMPQ  len, $16
	JGE   aligned

ret:
	RET
