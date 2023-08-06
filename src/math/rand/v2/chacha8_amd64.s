// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define QR(A, B, C, D, T) \
	PADDD B, A; PXOR A, D; PSHUFB ·rol16<>(SB), D \
	PADDD D, C; PXOR C, B; MOVO B, T; PSLLL $12, T; PSRLL $20, B; PXOR T, B; \
	PADDD B, A; PXOR A, D; PSHUFB ·rol8<>(SB), D \
	PADDD D, C; PXOR C, B; MOVO B, T; PSLLL $7, T; PSRLL $25, B; PXOR T, B

#define SEED(off, XR) \
	MOVL (4*off)(DX), CX; \
	MOVQ CX, XR; \
	PSHUFD $0, XR, XR

// func chacha8block(counter uint64, seed *[8]uint32, blocks *[16][4]uint32)
TEXT ·chacha8block(SB), NOSPLIT, $0
	MOVQ counter+0(FP), BX
	MOVQ seed+8(FP), DX
	MOVQ blocks+16(FP), AX

	MOVOU ·chachaConst0<>(SB), X0
	MOVOU ·chachaConst1<>(SB), X1
	MOVOU ·chachaConst2<>(SB), X2
	MOVOU ·chachaConst3<>(SB), X3

	SEED(0, X4)
	SEED(1, X5)
	SEED(2, X6)
	SEED(3, X7)
	SEED(4, X8)
	SEED(5, X9)
	SEED(6, X10)
	SEED(7, X11)

	MOVQ counter+0(FP), R8
	LEAQ 1(R8), R9
	LEAQ 2(R8), R10
	LEAQ 3(R8), R11

	MOVL R8, X12
	PINSRD $1, R9, X12
	PINSRD $2, R10, X12
	PINSRD $3, R11, X12

	SHRQ $32, R8
	SHRQ $32, R9
	SHRQ $32, R10
	SHRQ $32, R11

	MOVL R8, X13
	PINSRD $1, R9, X13
	PINSRD $2, R10, X13
	PINSRD $3, R11, X13

	MOVL $0, CX
	MOVQ CX, X14
	MOVOU X14, (15*16)(AX)

	MOVL $4, CX

loop:
	QR(X0, X4, X8, X12, X15)
	MOVOU X4, (4*16)(AX)
	QR(X1, X5, X9, X13, X15)
	MOVOU (15*16)(AX), X15
	QR(X2, X6, X10, X14, X4)
	QR(X3, X7, X11, X15, X4)

	QR(X0, X5, X10, X15, X4)
	MOVOU X15, (15*16)(AX)
	QR(X1, X6, X11, X12, X4)
	MOVOU (4*16)(AX), X4
	QR(X2, X7, X8, X13, X15)
	QR(X3, X4, X9, X14, X15)

	DECL CX
	JNZ loop

	MOVOU X0, (0*16)(AX)
	MOVOU X1, (1*16)(AX)
	MOVOU X2, (2*16)(AX)
	MOVOU X3, (3*16)(AX)
	MOVOU X4, (4*16)(AX)
	MOVOU X5, (5*16)(AX)
	MOVOU X6, (6*16)(AX)
	MOVOU X7, (7*16)(AX)
	MOVOU X8, (8*16)(AX)
	MOVOU X9, (9*16)(AX)
	MOVOU X10, (10*16)(AX)
	MOVOU X11, (11*16)(AX)
	MOVOU X12, (12*16)(AX)
	MOVOU X13, (13*16)(AX)
	MOVOU X14, (14*16)(AX)
	RET

// <<< 16 with PSHUFB
GLOBL ·rol16<>(SB), NOPTR|RODATA, $16
DATA ·rol16<>+0(SB)/8, $0x0504070601000302
DATA ·rol16<>+8(SB)/8, $0x0D0C0F0E09080B0A

// <<< 8 with PSHUFB
GLOBL ·rol8<>(SB), NOPTR|RODATA, $16
DATA ·rol8<>+0(SB)/8, $0x0605040702010003
DATA ·rol8<>+8(SB)/8, $0x0E0D0C0F0A09080B

GLOBL ·chachaConst0<>(SB), NOPTR|RODATA, $16
DATA ·chachaConst0<>+0(SB)/8, $0x61707865_61707865
DATA ·chachaConst0<>+8(SB)/8, $0x61707865_61707865

GLOBL ·chachaConst1<>(SB), NOPTR|RODATA, $16
DATA ·chachaConst1<>+0(SB)/8, $0x3320646e_3320646e
DATA ·chachaConst1<>+8(SB)/8, $0x3320646e_3320646e

GLOBL ·chachaConst2<>(SB), NOPTR|RODATA, $16
DATA ·chachaConst2<>+0(SB)/8, $0x79622d32_79622d32
DATA ·chachaConst2<>+8(SB)/8, $0x79622d32_79622d32

GLOBL ·chachaConst3<>(SB), NOPTR|RODATA, $16
DATA ·chachaConst3<>+0(SB)/8, $0x6b206574_6b206574
DATA ·chachaConst3<>+8(SB)/8, $0x6b206574_6b206574

