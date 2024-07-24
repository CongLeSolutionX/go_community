// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These functions are for testing only.
// (Assembly in _amd64_test.s files doesn't work.)

#include "go_asm.h"
#include "funcdata.h"

TEXT ·loadRegisters(SB),0,$0-8
	MOVQ	p+0(FP), AX

	MOVQ	(AX), R10
	MOVQ	(AX), R11
	MOVQ	(AX), R12
	MOVQ	(AX), R13

	MOVOU	(AX), X1
	MOVOU	(AX), X2
	MOVOU	(AX), X3
	MOVOU	(AX), X4

	CMPB	internal∕cpu·X86+const_offsetX86HasAVX(SB), $1
	JNE	return

	VMOVDQU	(AX), Y5
	VMOVDQU	(AX), Y6
	VMOVDQU	(AX), Y7
	VMOVDQU	(AX), Y8

	CMPB	internal∕cpu·X86+const_offsetX86HasAVX512F(SB), $1
	JNE	return

	// vmovupd (%rax), %zmm14
	BYTE $0x62; BYTE $0x71; BYTE $0xfd; BYTE $0x48; BYTE $0x10; BYTE $0x30
	// vmovupd (%rax), %zmm15
	BYTE $0x62; BYTE $0x71; BYTE $0xfd; BYTE $0x48; BYTE $0x10; BYTE $0x38
	// vmovupd (%rax), %zmm16
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x10; BYTE $0x00
	// vmovupd (%rax), %zmm17
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x10; BYTE $0x08

	// kmovq (%rax), %k2
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0x10
	// kmovq (%rax), %k3
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0x18
	// kmovq (%rax), %k4
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0x20
	// kmovq (%rax), %k5
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0x28

return:
	RET

TEXT ·spillRegisters(SB),0,$0-16
	MOVQ	p+0(FP), AX
	MOVQ	AX, BX

	MOVQ	R10, (AX)
	MOVQ	R11, 8(AX)
	MOVQ	R12, 16(AX)
	MOVQ	R13, 24(AX)
	ADDQ	$32, AX

	MOVOU	X1, (AX)
	MOVOU	X2, 16(AX)
	MOVOU	X3, 32(AX)
	MOVOU	X4, 48(AX)
	ADDQ	$64, AX

	CMPB	internal∕cpu·X86+const_offsetX86HasAVX(SB), $1
	JNE	return

	VMOVDQU	Y5, (AX)
	VMOVDQU	Y6, 32(AX)
	VMOVDQU	Y7, 64(AX)
	VMOVDQU	Y8, 96(AX)
	ADDQ	$128, AX

	CMPB	internal∕cpu·X86+const_offsetX86HasAVX512F(SB), $1
	JNE	return

	// vmovupd %zmm14, (AX)
	BYTE $0x62; BYTE $0x71; BYTE $0xfd; BYTE $0x48; BYTE $0x11; BYTE $0x30
	ADDQ	$64, AX
	// vmovupd %zmm15, (AX)
	BYTE $0x62; BYTE $0x71; BYTE $0xfd; BYTE $0x48; BYTE $0x11; BYTE $0x38
	ADDQ	$64, AX
	// vmovupd %zmm16, (AX)
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x11; BYTE $0x00
	ADDQ	$64, AX
	// vmovupd %zmm17, (AX)
	BYTE $0x62; BYTE $0xe1; BYTE $0xfd; BYTE $0x48; BYTE $0x11; BYTE $0x08
	ADDQ	$64, AX

	//kmovq %k2, (%rax)
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x91; BYTE $0x10
	ADDQ	$8, AX
	//kmovq %k3, (%rax)
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x91; BYTE $0x18
	ADDQ	$8, AX
	//kmovq %k4, (%rax)
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x91; BYTE $0x20
	ADDQ	$8, AX
	//kmovq %k5, (%rax)
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x91; BYTE $0x28
	ADDQ	$8, AX

return:
	SUBQ	BX, AX
	MOVQ	AX, ret+8(FP)
	RET

TEXT ·useSecret(SB),0,$64-24
	NO_LOCAL_POINTERS

	// Load secret into AX
	MOVQ	secret_base+0(FP), AX
	MOVQ	(AX), AX

	// Scatter secret all across registers.
	// Increment low byte so we can tell which register
	// a leaking secret came from.
	ADDQ	$2, AX // add 2 so Rn has secret #n.
	MOVQ	AX, BX
	INCQ	AX
	MOVQ	AX, CX
	INCQ	AX
	MOVQ	AX, DX
	INCQ	AX
	MOVQ	AX, SI
	INCQ	AX
	MOVQ	AX, DI
	INCQ	AX
	MOVQ	AX, BP
	INCQ	AX
	MOVQ	AX, R8
	INCQ	AX
	MOVQ	AX, R9
	INCQ	AX
	MOVQ	AX, R10
	INCQ	AX
	MOVQ	AX, R11
	INCQ	AX
	MOVQ	AX, R12
	INCQ	AX
	MOVQ	AX, R13
	INCQ	AX
	MOVQ	AX, R14
	INCQ	AX
	MOVQ	AX, R15

	// Fill in float registers. Assume avx512 here (TODO).
	// Used gcc to assemble these instructions.
	// vmovupd (%rsp), %zmm0..31
#define AVX512(r0, r1) \
	INCQ	AX \
	MOVQ	AX, (SP) \
	MOVQ	AX, 8(SP) \
	MOVQ	AX, 16(SP) \
	MOVQ	AX, 24(SP) \
	MOVQ	AX, 32(SP) \
	MOVQ	AX, 40(SP) \
	MOVQ	AX, 48(SP) \
	MOVQ	AX, 56(SP) \
	BYTE	$0x62; BYTE $r0; BYTE $0xfd; BYTE $0x48; BYTE $0x10; BYTE $r1; BYTE $0x24

	// TODO: enable
	/*
	AVX512(0xf1, 0x04)
	AVX512(0xf1, 0x0c)
	AVX512(0xf1, 0x14)
	AVX512(0xf1, 0x1c)
	AVX512(0xf1, 0x24)
	AVX512(0xf1, 0x2c)
	AVX512(0xf1, 0x34)
	AVX512(0xf1, 0x3c)

	AVX512(0x71, 0x04)
	AVX512(0x71, 0x0c)
	AVX512(0x71, 0x14)
	AVX512(0x71, 0x1c)
	AVX512(0x71, 0x24)
	AVX512(0x71, 0x2c)
	AVX512(0x71, 0x34)
	AVX512(0x71, 0x3c)

	AVX512(0xe1, 0x04)
	AVX512(0xe1, 0x0c)
	AVX512(0xe1, 0x14)
	AVX512(0xe1, 0x1c)
	AVX512(0xe1, 0x24)
	AVX512(0xe1, 0x2c)
	AVX512(0xe1, 0x34)
	AVX512(0xe1, 0x3c)

	AVX512(0x61, 0x04)
	AVX512(0x61, 0x0c)
	AVX512(0x61, 0x14)
	AVX512(0x61, 0x1c)
	AVX512(0x61, 0x24)
	AVX512(0x61, 0x2c)
	AVX512(0x61, 0x34)
	AVX512(0x61, 0x3c)
	*/

	/*
	MOVOU	(SP), X0
	MOVOU	(SP), X1
	MOVOU	(SP), X2
	MOVOU	(SP), X3
	MOVOU	(SP), X4
	MOVOU	(SP), X5
	MOVOU	(SP), X6
	MOVOU	(SP), X7
	MOVOU	(SP), X8
	MOVOU	(SP), X9
	MOVOU	(SP), X10
	MOVOU	(SP), X11
	MOVOU	(SP), X12
	MOVOU	(SP), X13
	MOVOU	(SP), X14
	MOVOU	(SP), X15
	*/

	// Put secret on the stack.
	INCQ	AX
	MOVQ	AX, (SP)
	MOVQ	AX, 8(SP)
	MOVQ	AX, 16(SP)
	MOVQ	AX, 24(SP)
	MOVQ	AX, 32(SP)
	MOVQ	AX, 40(SP)
	MOVQ	AX, 48(SP)
	MOVQ	AX, 56(SP)

	// Delay a bit.  This makes it more likely that
	// we will be the target of a signal while
	// registers contain secrets.
	// It also tests the path from G stack to M stack
	// to scheduler and back.
	CALL	·delay(SB)

	RET
