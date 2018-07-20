#include "textflag.h"

#define dst BX	// parity's address
#define src0 SI	// two-dimension src_slice's address
#define src1 CX	// cnt of src
#define len DX	// len of vect
#define pos R8	// job position in vect

#define not_aligned_len R12
#define src_val0 R13
#define src_val1 R14

// func xorAVX512(dst, a, b []byte, n int)
TEXT ·xorAVX512(SB), NOSPLIT, $0
    MOVQ  d+0(FP), dst
	MOVQ  s0+24(FP), src0
	MOVQ  s1+48(FP), src1
	MOVQ  l+72(FP), len
	TESTQ $63, len
	JNZ   not_aligned

aligned:
	MOVQ $0, pos

loop64b:
	VMOVDQU8  (src0)(pos*1), Z0
	VPXORQ   (src1)(pos*1), Z0, Z0
	VMOVDQU8 Z0, (dst)(pos*1)
	ADDQ     $64, pos
	CMPQ     len, pos
	JNE      loop64b
	VZEROUPPER
	RET

loop_1b:
	MOVB  -1(src0)(len*1), src_val0
	MOVB  -1(src1)(len*1), src_val1
	XORB  src_val1, src_val0
	MOVB  src_val0, -1(dst)(len*1)
	SUBQ  $1, len
	TESTQ $7, len
	JNZ   loop_1b
	CMPQ  len, $0
	JE    ret
	TESTQ $63, len
	JZ    aligned

not_aligned:
	TESTQ $7, len
	JNE   loop_1b
	MOVQ  len, not_aligned_len
	ANDQ  $63, not_aligned_len

loop_8b:
    MOVQ -8(src0)(len*1), src_val0
	MOVQ -8(src1)(len*1), src_val1
	XORQ src_val1, src_val0
	MOVQ src_val0, -8(dst)(len*1)
	SUBQ $8, len
	SUBQ $8, not_aligned_len
	JG   loop_8b

	CMPQ len, $64
	JGE  aligned
	RET

ret:
	RET

// func xorAVX2(dst, a, b []byte, n int)
TEXT ·xorAVX2(SB), NOSPLIT, $0
    MOVQ  d+0(FP), dst
	MOVQ  s0+24(FP), src0
	MOVQ  s1+48(FP), src1
	MOVQ  l+72(FP), len
	TESTQ $31, len
	JNZ   not_aligned

aligned:
	MOVQ $0, pos

loop32b:
	VMOVDQU  (src0)(pos*1), Y0
	VPXOR    (src1)(pos*1), Y0, Y0
	VMOVDQU  Y0, (dst)(pos*1)
	ADDQ     $32, pos
	CMPQ     len, pos
	JNE      loop32b
	VZEROUPPER
	RET

loop_1b:
	MOVB  -1(src0)(len*1), src_val0
	MOVB  -1(src1)(len*1), src_val1
	XORB  src_val1, src_val0
	MOVB  src_val0, -1(dst)(len*1)
	SUBQ  $1, len
	TESTQ $7, len
	JNZ   loop_1b
	CMPQ  len, $0
	JE    ret
	TESTQ $31, len
	JZ    aligned

not_aligned:
	TESTQ $7, len
	JNE   loop_1b
	MOVQ  len, not_aligned_len
	ANDQ  $31, not_aligned_len

loop_8b:
    MOVQ -8(src0)(len*1), src_val0
	MOVQ -8(src1)(len*1), src_val1
	XORQ src_val1, src_val0
	MOVQ src_val0, -8(dst)(len*1)
	SUBQ $8, len
	SUBQ $8, not_aligned_len
	JG   loop_8b

	CMPQ len, $32
	JGE  aligned
	RET

ret:
	RET

// func xorSSE2(dst, a, b []byte, n int)
TEXT ·xorSSE2(SB), NOSPLIT, $0
    MOVQ  d+0(FP), dst
	MOVQ  s0+24(FP), src0
	MOVQ  s1+48(FP), src1
	MOVQ  l+72(FP), len
	TESTQ $15, len
	JNZ   not_aligned

aligned:
	MOVQ $0, pos

loop16b:
	MOVOU    (src0)(pos*1), X0
	XORPD    (src1)(pos*1), X0
	MOVOU    X0, (dst)(pos*1)
	ADDQ     $16, pos
	CMPQ     len, pos
	JNE      loop16b
	VZEROUPPER
	RET

loop_1b:
	MOVB  -1(src0)(len*1), src_val0
	MOVB  -1(src1)(len*1), src_val1
	XORB  src_val1, src_val0
	MOVB  src_val0, -1(dst)(len*1)
	SUBQ  $1, len
	TESTQ $7, len
	JNZ   loop_1b
	CMPQ  len, $0
	JE    ret
	TESTQ $15, len
	JZ    aligned

not_aligned:
	TESTQ $7, len
	JNE   loop_1b
	MOVQ  len, not_aligned_len
	ANDQ  $15, not_aligned_len

loop_8b:
    MOVQ -8(src0)(len*1), src_val0
	MOVQ -8(src1)(len*1), src_val1
	XORQ src_val1, src_val0
	MOVQ src_val0, -8(dst)(len*1)
	SUBQ $8, len
	SUBQ $8, not_aligned_len
	JG   loop_8b

	CMPQ len, $16
	JGE  aligned
	RET

ret:
	RET
