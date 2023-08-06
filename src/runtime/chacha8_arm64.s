// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define QR(A, B, C, D) \
	VADD A.S4, B.S4, A.S4; \
	VEOR D.B16, A.B16, D.B16; \
	VREV32 D.H8, D.H8; \
	\
	VADD C.S4, D.S4, C.S4; \
	VEOR B.B16, C.B16, V16.B16; \
	VSHL $12, V16.S4, B.S4; \
	VSRI $20, V16.S4, B.S4 \
	\
	VADD A.S4, B.S4, A.S4; \
	VEOR D.B16, A.B16, D.B16; \
	VTBL V31.B16, [D.B16], D.B16; \
	\
	VADD C.S4, D.S4, C.S4; \
	VEOR B.B16, C.B16, V16.B16; \
	VSHL $7, V16.S4, B.S4; \
	VSRI $25, V16.S4, B.S4

// func chacha8block(counter uint64, seed *[8]uint32, blocks *[4][16]uint32)
TEXT ·chacha8block<ABIInternal>(SB), NOSPLIT, $16
	// counter in R0
	// seed in R1
	// blocks in R2

	MOVD $·chachaConst(SB), R10
	VLD4R (R10), [V0.S4, V1.S4, V2.S4, V3.S4]

	MOVD $·chachaIncRot(SB), R11
	VLD1 (R11), [V30.S4, V31.S4]

	// seed
	VLD4R.P 16(R1), [V4.S4, V5.S4, V6.S4, V7.S4]
	VLD4R.P 16(R1), [V8.S4, V9.S4, V10.S4, V11.S4]

	// store counter to memory to replicate its uint32 halfs back out
	MOVD R0, 0(RSP)
	MOVD RSP, R1
	VLD1R.P 4(R1), [V12.S4]
	VLD1R.P 4(R1), [V13.S4]
	VADD V30.S4, V12.S4, V12.S4

	VEOR V14.B16, V14.B16, V14.B16
	VEOR V15.B16, V15.B16, V15.B16

	MOVD $4, R1

loop:
	QR(V0, V4, V8, V12)
	QR(V1, V5, V9, V13)
	QR(V2, V6, V10, V14)
	QR(V3, V7, V11, V15)

	QR(V0, V5, V10, V15)
	QR(V1, V6, V11, V12)
	QR(V2, V7, V8, V13)
	QR(V3, V4, V9, V14)

	SUB $1, R1
	CBNZ R1, loop

	VST1.P [ V0.B16,  V1.B16,  V2.B16,  V3.B16], 64(R2)
	VST1.P [ V4.B16,  V5.B16,  V6.B16,  V7.B16], 64(R2)
	VST1.P [ V8.B16,  V9.B16, V10.B16, V11.B16], 64(R2)
	VST1.P [V12.B16, V13.B16, V14.B16, V15.B16], 64(R2)
	RET

GLOBL	·chachaConst(SB), NOPTR|RODATA, $32
DATA	·chachaConst+0x00(SB)/4, $0x61707865
DATA	·chachaConst+0x04(SB)/4, $0x3320646e
DATA	·chachaConst+0x08(SB)/4, $0x79622d32
DATA	·chachaConst+0x0c(SB)/4, $0x6b206574

GLOBL	·chachaIncRot(SB), NOPTR|RODATA, $32
DATA	·chachaIncRot+0x00(SB)/4, $0x00000000
DATA	·chachaIncRot+0x04(SB)/4, $0x00000001
DATA	·chachaIncRot+0x08(SB)/4, $0x00000002
DATA	·chachaIncRot+0x0c(SB)/4, $0x00000003
DATA	·chachaIncRot+0x10(SB)/4, $0x02010003
DATA	·chachaIncRot+0x14(SB)/4, $0x06050407
DATA	·chachaIncRot+0x18(SB)/4, $0x0A09080B
DATA	·chachaIncRot+0x1c(SB)/4, $0x0E0D0C0F
