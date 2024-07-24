// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	// VZEROALL zeroes all of the X0-X15 registers, no matter how wide.
	VZEROALL

	// Test if we need to zero avx512 registers.
	// First, ask for how many "leaves" of cpuid bits we support.
	MOVL	$0, AX
	CPUID
	CMPL	AX, $7
	JB	noavx512 // can't even ask for leaf 7 of the cpuid bits

	// Then, ask about avx512.
	MOVL	$7, AX  // leaf 7
	MOVL	$0, CX  // sub-leaf 0 (= "extended features")
	CPUID
	TESTL	$(1<<16), BX // bit 16 = avx512 Foundation
	JEQ	noavx512

	// Note that we don't check OS support. We're happy to clear
	// the registers and then have the OS drop those zeroes on the
	// floor across a context switch.

	// Clear all the avx512 state.
	// Note that VZEROALL above already cleared bits 256-511 of X0-X15. (TODO: check)

	// Zero X16-X31
	// (Assembled with gcc. The Go assembler doesn't know these instructions.)
	//   	62 b1 fd 48 28 c0    	vmovapd %zmm16,%zmm0
	//   	62 b1 fd 48 28 c1    	vmovapd %zmm17,%zmm0
	//   	62 b1 fd 48 28 c2    	vmovapd %zmm18,%zmm0
	//  	62 b1 fd 48 28 c3    	vmovapd %zmm19,%zmm0
	//  	62 b1 fd 48 28 c4    	vmovapd %zmm20,%zmm0
	//  	62 b1 fd 48 28 c5    	vmovapd %zmm21,%zmm0
	//  	62 b1 fd 48 28 c6    	vmovapd %zmm22,%zmm0
	//  	62 b1 fd 48 28 c7    	vmovapd %zmm23,%zmm0
	//  	62 91 fd 48 28 c0    	vmovapd %zmm24,%zmm0
	//  	62 91 fd 48 28 c1    	vmovapd %zmm25,%zmm0
	//  	62 91 fd 48 28 c2    	vmovapd %zmm26,%zmm0
	//  	62 91 fd 48 28 c3    	vmovapd %zmm27,%zmm0
	//  	62 91 fd 48 28 c4    	vmovapd %zmm28,%zmm0
	//  	62 91 fd 48 28 c5    	vmovapd %zmm29,%zmm0
	//  	62 91 fd 48 28 c6    	vmovapd %zmm30,%zmm0
	//  	62 91 fd 48 28 c7    	vmovapd %zmm31,%zmm0
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc0
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc1
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc2
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc3
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc4
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc5
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc6
	BYTE $0x62; BYTE $0xb1; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc7
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc0
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc1
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc2
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc3
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc4
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc5
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc6
	BYTE $0x62; BYTE $0x91; BYTE $0xfd; BYTE $0x48; BYTE $0x28; BYTE $0xc7

	// Zero k0-k7
	// (Assembled with gcc. The Go assembler doesn't know these instructions.)
	//  	c4 e1 f8 90 c1       	kmovq  %k1,%k0
	//  	c4 e1 f8 90 c2       	kmovq  %k2,%k0
	//  	c4 e1 f8 90 c3       	kmovq  %k3,%k0
	//  	c4 e1 f8 90 c4       	kmovq  %k4,%k0
	//  	c4 e1 f8 90 c5       	kmovq  %k5,%k0
	//  	c4 e1 f8 90 c6       	kmovq  %k6,%k0
	//  	c4 e1 f8 90 c7       	kmovq  %k7,%k0
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc1
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc2
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc3
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc4
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc5
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc6
	BYTE $0xc4; BYTE $0xe1; BYTE $0xf8; BYTE $0x90; BYTE $0xc7

	// Note: On Darwin, if we haven't issued an avx512 instruction
	// yet the CPUID instruction claims not to support
	// avx512. Fortunately for this code, that's not a
	// deal-breaker. If an avx512 instruction has not been issued
	// yet, no need to zero anything. See issue 43089.

	// Note: we use CPUID here instead of using internal/cpu, because the
	// flags in internal/cpu can be toggled, whereas we really want this
	// clearing to happen regardless of flags.

noavx512:
	// misc registers
	CMPQ	AX, AX	//eflags
	// segment registers? Seems overkill.

	RET
