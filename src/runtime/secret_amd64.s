// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"

TEXT ·secretEraseRegisters(SB),0,$0
	// integer registers
	XORQ	AX, AX
	XORQ	BX, BX
	XORQ	CX, CX
	XORQ	DX, DX
	XORQ	DI, DI
	XORQ	SI, SI
	// BP = frame pointer
	// SP = stack pointer
	XORQ	R8, R8
	XORQ	R9, R9
	XORQ	R10, R10
	XORQ	R11, R11
	XORQ	R12, R12
	XORQ	R13, R13
	XORQ	R14, R14
	XORQ	R15, R15

	// floating-point registers
	CMPB	internal∕cpu·X86+const_offsetX86HasAVX(SB), $1
	JEQ	avx

	PXOR	X0, X0
	PXOR	X1, X1
	PXOR	X2, X2
	PXOR	X3, X3
	PXOR	X4, X4
	PXOR	X5, X5
	PXOR	X6, X6
	PXOR	X7, X7
	PXOR	X8, X8
	PXOR	X9, X9
	PXOR	X10, X10
	PXOR	X11, X11
	PXOR	X12, X12
	PXOR	X13, X13
	PXOR	X14, X14
	PXOR	X15, X15
	JMP	noavx512

avx:
	// VZEROALL zeroes all of the X0-X15 registers, no matter how wide.
	// That includes Y0-Y15 (256-bit avx) and Z0-Z15 (512-bit avx512).
	VZEROALL

	// Clear all the avx512 state.
	CMPB	internal∕cpu·X86+const_offsetX86HasAVX512F(SB), $1
	JNE	noavx512

	// Zero X16-X31
	// Note that VZEROALL above already cleared Z0-Z15.
	// (Assembled with gcc. The Go assembler doesn't know these instructions.)
	//   	62 e1 fd 48 28 c0    	vmovapd %zmm0,%zmm16
	//   	62 e1 fd 48 28 c8    	vmovapd %zmm0,%zmm17
	//   	62 e1 fd 48 28 d0    	vmovapd %zmm0,%zmm18
	//   	62 e1 fd 48 28 d8    	vmovapd %zmm0,%zmm19
	//   	62 e1 fd 48 28 e0    	vmovapd %zmm0,%zmm20
	//   	62 e1 fd 48 28 e8    	vmovapd %zmm0,%zmm21
	//   	62 e1 fd 48 28 f0    	vmovapd %zmm0,%zmm22
	//   	62 e1 fd 48 28 f8    	vmovapd %zmm0,%zmm23
	//   	62 61 fd 48 28 c0    	vmovapd %zmm0,%zmm24
	//   	62 61 fd 48 28 c8    	vmovapd %zmm0,%zmm25
	//   	62 61 fd 48 28 d0    	vmovapd %zmm0,%zmm26
	//   	62 61 fd 48 28 d8    	vmovapd %zmm0,%zmm27
	//   	62 61 fd 48 28 e0    	vmovapd %zmm0,%zmm28
	//   	62 61 fd 48 28 e8    	vmovapd %zmm0,%zmm29
	//   	62 61 fd 48 28 f0    	vmovapd %zmm0,%zmm30
	//   	62 61 fd 48 28 f8    	vmovapd %zmm0,%zmm31
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc0
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc8
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xd0
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xd8
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xe0
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xe8
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xf0
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xf8
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc0
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc8
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xd0
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xd8
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xe0
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xe8
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xf0
	BYTE $0x62; BYTE $0x61; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xf8

	// Zero k0-k7
	// (Assembled with gcc. The Go assembler doesn't know these instructions.)
	//  	c4 e1 fc 47 c0       	kxorq  %k0,%k0,%k0
	//  	c4 e1 fc 47 c8       	kxorq  %k0,%k0,%k1
	//  	c4 e1 fc 47 d0       	kxorq  %k0,%k0,%k2
	//  	c4 e1 fc 47 d8       	kxorq  %k0,%k0,%k3
	//  	c4 e1 fc 47 e0       	kxorq  %k0,%k0,%k4
	//  	c4 e1 fc 47 e8       	kxorq  %k0,%k0,%k5
	//  	c4 e1 fc 47 f0       	kxorq  %k0,%k0,%k6
	//  	c4 e1 fc 47 f8       	kxorq  %k0,%k0,%k7
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xc0
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xc8
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xd0
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xd8
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xe0
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xe8
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xf0
	BYTE $0xc4; BYTE $0xe1; BYTE $0xfc; BYTE $0x47; BYTE $0xf8

noavx512:
	// misc registers
	CMPQ	AX, AX	//eflags
	// segment registers? Direction flag? Both seem overkill.

	RET
