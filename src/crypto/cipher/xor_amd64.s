#include "textflag.h"

#define dst BX // dst address
#define src0 SI // a address
#define src1 CX // b address
#define len DX // len of vect
#define pos R8 // job position in vect

#define not_aligned_len R12
#define src_val0 R13
#define src_val1 R14

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
	MOVB  (src0)(len*1), src_val0
	MOVB  (src1)(len*1), src_val1
	XORB  src_val1, src_val0
	MOVB  src_val0, (dst)(len*1)
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
	MOVQ  (src0)(len*1), src_val0
	MOVQ  (src1)(len*1), src_val1
	XORQ  src_val1, src_val0
	MOVQ  src_val0, (dst)(len*1)
	CMPQ  len, $16
	JGE   aligned

ret:
	RET
