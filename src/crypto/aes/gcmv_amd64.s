// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is an optimized implementation of AES-GCM using AES-NI, CLMUL-NI, VAES, and VPCLMULQDQ
// The implementation uses some optimization as described in:
// [1] Gueron, S., Kounavis, M.E.: Intel® Carry-Less Multiplication
//     Instruction and its Usage for Computing the GCM Mode rev. 2.02
// [2] Gueron, S., Krasnov, V.: Speeding up Counter Mode in Software and
//     Hardware

#include "textflag.h"

#define B0 X0
#define B1 X1
#define B2 X2
#define B3 X3
#define B4 X4
#define B5 X5
#define B6 X6
#define B7 X7

#define ACC0 X8
#define ACC1 X9
#define ACCM X10

#define T0 X11
#define T1 X12
#define T2 X13
#define POLY X14
#define BSWAP X15

DATA bswapMask<>+0x00(SB)/8, $0x08090a0b0c0d0e0f
DATA bswapMask<>+0x08(SB)/8, $0x0001020304050607

DATA gcmPoly<>+0x00(SB)/8, $0x0000000000000001
DATA gcmPoly<>+0x08(SB)/8, $0xc200000000000000

DATA andMask<>+0x00(SB)/8, $0x00000000000000ff
DATA andMask<>+0x08(SB)/8, $0x0000000000000000
DATA andMask<>+0x10(SB)/8, $0x000000000000ffff
DATA andMask<>+0x18(SB)/8, $0x0000000000000000
DATA andMask<>+0x20(SB)/8, $0x0000000000ffffff
DATA andMask<>+0x28(SB)/8, $0x0000000000000000
DATA andMask<>+0x30(SB)/8, $0x00000000ffffffff
DATA andMask<>+0x38(SB)/8, $0x0000000000000000
DATA andMask<>+0x40(SB)/8, $0x000000ffffffffff
DATA andMask<>+0x48(SB)/8, $0x0000000000000000
DATA andMask<>+0x50(SB)/8, $0x0000ffffffffffff
DATA andMask<>+0x58(SB)/8, $0x0000000000000000
DATA andMask<>+0x60(SB)/8, $0x00ffffffffffffff
DATA andMask<>+0x68(SB)/8, $0x0000000000000000
DATA andMask<>+0x70(SB)/8, $0xffffffffffffffff
DATA andMask<>+0x78(SB)/8, $0x0000000000000000
DATA andMask<>+0x80(SB)/8, $0xffffffffffffffff
DATA andMask<>+0x88(SB)/8, $0x00000000000000ff
DATA andMask<>+0x90(SB)/8, $0xffffffffffffffff
DATA andMask<>+0x98(SB)/8, $0x000000000000ffff
DATA andMask<>+0xa0(SB)/8, $0xffffffffffffffff
DATA andMask<>+0xa8(SB)/8, $0x0000000000ffffff
DATA andMask<>+0xb0(SB)/8, $0xffffffffffffffff
DATA andMask<>+0xb8(SB)/8, $0x00000000ffffffff
DATA andMask<>+0xc0(SB)/8, $0xffffffffffffffff
DATA andMask<>+0xc8(SB)/8, $0x000000ffffffffff
DATA andMask<>+0xd0(SB)/8, $0xffffffffffffffff
DATA andMask<>+0xd8(SB)/8, $0x0000ffffffffffff
DATA andMask<>+0xe0(SB)/8, $0xffffffffffffffff
DATA andMask<>+0xe8(SB)/8, $0x00ffffffffffffff

GLOBL bswapMask<>(SB), (NOPTR+RODATA), $16
GLOBL gcmPoly<>(SB), (NOPTR+RODATA), $16
GLOBL andMask<>(SB), (NOPTR+RODATA), $240

DATA ·ctr_init<>+0x00(SB)/8, $0x0000000000000000 
DATA ·ctr_init<>+0x08(SB)/8, $0x0000000100000000
DATA ·ctr_init<>+0x10(SB)/8, $0x0000000000000000 
DATA ·ctr_init<>+0x18(SB)/8, $0x0000000200000000
DATA ·ctr_init<>+0x20(SB)/8, $0x0000000000000000 
DATA ·ctr_init<>+0x28(SB)/8, $0x0000000300000000
DATA ·ctr_init<>+0x30(SB)/8, $0x0000000000000000 
DATA ·ctr_init<>+0x38(SB)/8, $0x0000000400000000
GLOBL ·ctr_init<>(SB), (NOPTR+RODATA), $64

DATA ·ctr_incr<>+0x00(SB)/8, $0x0000000000000000 
DATA ·ctr_incr<>+0x08(SB)/8, $0x0000000400000000
DATA ·ctr_incr<>+0x10(SB)/8, $0x0000000000000000 
DATA ·ctr_incr<>+0x18(SB)/8, $0x0000000400000000
DATA ·ctr_incr<>+0x20(SB)/8, $0x0000000000000000 
DATA ·ctr_incr<>+0x28(SB)/8, $0x0000000400000000
DATA ·ctr_incr<>+0x30(SB)/8, $0x0000000000000000 
DATA ·ctr_incr<>+0x38(SB)/8, $0x0000000400000000
GLOBL ·ctr_incr<>(SB), (NOPTR+RODATA), $64

DATA ·vbswapMask<>+0x00(SB)/8, $0x08090a0b0c0d0e0f
DATA ·vbswapMask<>+0x08(SB)/8, $0x0001020304050607
DATA ·vbswapMask<>+0x10(SB)/8, $0x18191a1b1c1d1e1f
DATA ·vbswapMask<>+0x18(SB)/8, $0x1011121314151617
DATA ·vbswapMask<>+0x20(SB)/8, $0x28292a2b2c2d2e2f
DATA ·vbswapMask<>+0x28(SB)/8, $0x2021222324252627
DATA ·vbswapMask<>+0x30(SB)/8, $0x38393a3b3c3d3e3f
DATA ·vbswapMask<>+0x38(SB)/8, $0x3031323334353637
GLOBL ·vbswapMask<>(SB), (NOPTR+RODATA), $64

DATA ·ivBlendMask<>+0x00(SB)/2, $0x7777
GLOBL ·ivBlendMask<>(SB), (NOPTR+RODATA), $2

DATA ·vmsswapMask<>+0x00(SB)/8, $0x0706050403020100
DATA ·vmsswapMask<>+0x08(SB)/8, $0x0c0d0e0f0b0a0908
DATA ·vmsswapMask<>+0x10(SB)/8, $0x1716151413121110
DATA ·vmsswapMask<>+0x18(SB)/8, $0x1c1d1e1f1b1a1918
DATA ·vmsswapMask<>+0x20(SB)/8, $0x2726252423222120
DATA ·vmsswapMask<>+0x28(SB)/8, $0x2c2d2e2f2b2a2928
DATA ·vmsswapMask<>+0x30(SB)/8, $0x3736353433323130
DATA ·vmsswapMask<>+0x38(SB)/8, $0x3c3d3e3f3b3a3938
GLOBL ·vmsswapMask<>(SB), (NOPTR+RODATA), $64

// func gcmAesInitv(productTable *[1024]byte, ks []uint32)
TEXT ·gcmAesInitv(SB),NOSPLIT,$0
#define dst DI
#define KS SI
#define NR DX

	MOVQ productTable+0(FP), dst
	MOVQ ks_base+8(FP), KS
	MOVQ ks_len+16(FP), NR

	SHRQ $2, NR
	DECQ NR

	MOVOU bswapMask<>(SB), BSWAP
	MOVOU gcmPoly<>(SB), POLY

	// Encrypt block 0, with the AES key to generate the hash key H
	MOVOU (16*0)(KS), B0
	MOVOU (16*1)(KS), T0
	AESENC T0, B0
	MOVOU (16*2)(KS), T0
	AESENC T0, B0
	MOVOU (16*3)(KS), T0
	AESENC T0, B0
	MOVOU (16*4)(KS), T0
	AESENC T0, B0
	MOVOU (16*5)(KS), T0
	AESENC T0, B0
	MOVOU (16*6)(KS), T0
	AESENC T0, B0
	MOVOU (16*7)(KS), T0
	AESENC T0, B0
	MOVOU (16*8)(KS), T0
	AESENC T0, B0
	MOVOU (16*9)(KS), T0
	AESENC T0, B0
	MOVOU (16*10)(KS), T0
	CMPQ NR, $12
	JB initEncLast
	AESENC T0, B0
	MOVOU (16*11)(KS), T0
	AESENC T0, B0
	MOVOU (16*12)(KS), T0
	JE initEncLast
	AESENC T0, B0
	MOVOU (16*13)(KS), T0
	AESENC T0, B0
	MOVOU (16*14)(KS), T0
initEncLast:
	AESENCLAST T0, B0

	PSHUFB BSWAP, B0
	// H * 2
	PSHUFD $0xff, B0, T0
	MOVOU B0, T1
	PSRAL $31, T0
	PAND POLY, T0
	PSRLL $31, T1
	PSLLDQ $4, T1
	PSLLL $1, B0
	PXOR T0, B0
	PXOR T1, B0
	// Karatsuba pre-computations
	MOVOU B0, (16*31)(dst)
	PSHUFD $78, B0, B1
	PXOR B0, B1
	MOVOU B1, (16*63)(dst)

	MOVOU B0, B2
	MOVOU B1, B3
	// Now prepare powers of H and pre-computations for them
	MOVQ $31, AX

initLoop:
		MOVOU B2, T0
		MOVOU B2, T1
		MOVOU B3, T2
		PCLMULQDQ $0x00, B0, T0
		PCLMULQDQ $0x11, B0, T1
		PCLMULQDQ $0x00, B1, T2

		PXOR T0, T2
		PXOR T1, T2
		MOVOU T2, B4
		PSLLDQ $8, B4
		PSRLDQ $8, T2
		PXOR B4, T0
		PXOR T2, T1

		MOVOU POLY, B2
		PCLMULQDQ $0x01, T0, B2
		PSHUFD $78, T0, T0
		PXOR B2, T0
		MOVOU POLY, B2
		PCLMULQDQ $0x01, T0, B2
		PSHUFD $78, T0, T0
		PXOR T0, B2
		PXOR T1, B2

	  MOVOU B2, (16*30)(dst)
		PSHUFD $78, B2, B3
		PXOR B2, B3
		MOVOU B3, (16*62)(dst)

		DECQ AX
		LEAQ (-16)(dst), dst
	JNE initLoop
	RET
#undef NR
#undef KS
#undef dst

// func gcmAesEncv(productTable *[1024]byte, dst, src []byte, ctr, T *[16]byte, ks []uint32)
TEXT ·gcmAesEncv(SB),0,$256-96
#define pTbl DI
#define ctx DX
#define ctrPtr CX
#define ptx SI
#define ks AX
#define tPtr R8
#define ptxLen R9
#define aluCTR R10
#define aluTMP R11
#define aluK R12
#define NR R13
#define ACC0z Z8
#define ACC0y Y8
#define ACC1z Z4
#define ACC1x X4
#define ACCMz Z5
#define ACCMx X5
#define T0z Z9
#define T0x X9
#define T1v Z6
#define T1x X6
#define T1y Y6
#define T2v Z7
#define T2x X7
#define T2y Y7
#define POLYx X30
#define POLYz Z30

#define vaesRound(k) \
  VAESENC k, Z10, Z10; \
  VAESENC k, Z11, Z11; \ 
  VAESENC k, Z12, Z12; \
  VAESENC k, Z13, Z13; \
  VAESENC k, Z14, Z14; \
  VAESENC k, Z15, Z15; \
  VAESENC k, Z16, Z16; \
  VAESENC k, Z17, Z17;

#define vaesLastRound(k) \
  VAESENCLAST k, Z10, Z10; \
  VAESENCLAST k, Z11, Z11; \ 
  VAESENCLAST k, Z12, Z12; \
  VAESENCLAST k, Z13, Z13; \
  VAESENCLAST k, Z14, Z14; \
  VAESENCLAST k, Z15, Z15; \
  VAESENCLAST k, Z16, Z16; \
  VAESENCLAST k, Z17, Z17;

#define vcombinedRound(z,i) \
  VMOVUPS (64*i)(pTbl), T1v; \
	VBROADCASTI64X2 ((i+1)*16)(ks), K2, Z29; \   
	VMOVUPS T1v, T2v; \
  VAESENC Z29, Z10, Z10; \
  VAESENC Z29, Z11, Z11; \ 
  VAESENC Z29, Z12, Z12; \
  VAESENC Z29, Z13, Z13; \
	VPCLMULQDQ $0x00, z, T1v, T1v; \
  VAESENC Z29, Z14, Z14; \
  VAESENC Z29, Z15, Z15; \
  VAESENC Z29, Z16, Z16; \
  VAESENC Z29, Z17, Z17; \
  VPCLMULQDQ $0x11, z, T2v, T2v; \
	VPXORQ T1v, ACC0z, K1, ACC0z; \
	VPXORQ T2v, ACC1z, K1, ACC1z; \
	VPSHUFD $78, z, T1v; \
	VPXORQ T1v, z, z; \
  VMOVUPS (512+64*i)(pTbl), T1v; \
	VPCLMULQDQ $0x00, z, T1v, T1v; \
	VPXORQ T1v, ACCMz, K1, ACCMz

#define vmulRound(z,i) \
	VMOVUPS z, T0z;\
	VMOVUPS (64*i)(pTbl), T1v;\
	VMOVUPS T1v, T2v;\
	VPCLMULQDQ $0x00, T0z, T1v, T1v;\
	VPXORQ T1v, ACC0z, K1, ACC0z;\
	VPCLMULQDQ $0x11, T0z, T2v, T2v;\
	VPXORQ T2v, ACC1z, K1, ACC1z;\
	VPSHUFD $78, T0z, T1v;\
	VPXORQ T1v, T0z, T0z;\
	VMOVUPS (512+64*i)(pTbl), T1v;\
	VPCLMULQDQ $0x00, T0z, T1v, T1v;\
	VPXORQ T1v, ACCMz, K1, ACCMz

#define reduceRound(a) 	MOVOU POLY, T0;	PCLMULQDQ $0x01, a, T0; PSHUFD $78, a, a; PXOR T0, a
#define increment(i) ADDL $1, aluCTR; MOVL aluCTR, aluTMP; XORL aluK, aluTMP; BSWAPL aluTMP; MOVL aluTMP, (3*4 + 8*16 + i*16)(SP)
#define vreduceRound VPCLMULQDQ $0x01, ACC0z, POLYz, T1v; PSHUFD $78, ACC0, ACC0; PXOR T1x, ACC0
#define vincrement(k) VPADDD Z2, Z0, Z0; VPXORQ Z0, Z1, k; VPSHUFB Z3, k, k;        
#define xorh512to128(z) VEXTRACTI64X4 $1, z, T1y; VPXORQ z, T1v, T1v; VEXTRACTI64X2 $1, T1y, T2x; VPXORQ T2v, T1v, z

	MOVQ productTable+0(FP), pTbl
	MOVQ dst+8(FP), ctx
	MOVQ src_base+32(FP), ptx
	MOVQ src_len+40(FP), ptxLen
	MOVQ ctr+56(FP), ctrPtr
	MOVQ T+64(FP), tPtr
	MOVQ ks_base+72(FP), ks
	MOVQ ks_len+80(FP), NR

	SHRQ $2, NR
	DECQ NR

	MOVOU gcmPoly<>(SB), T1x
	VMOVUPS T1v, POLYz
	VPXORQ ACC0z, ACC0z, ACC0z
	MOVOU (tPtr), ACC0
	VPXORQ ACC1z, ACC1z, ACC1z
	VPXORQ ACCMz, ACCMz, ACCMz
 
	CMPQ ptxLen, $512
	JB gcmAesEncSingles
	SUBQ $512, ptxLen

	// load round keys
	KXNORQ K1, K1, K1                  
  VBROADCASTI64X2 0(ks), K1, Z18     
	VBROADCASTI64X2 16(ks), K1, Z19    
	VBROADCASTI64X2 32(ks), K1, Z20    
	VBROADCASTI64X2 48(ks), K1, Z21    
	VBROADCASTI64X2 64(ks), K1, Z22    
	VBROADCASTI64X2 80(ks), K1, Z23    
	VBROADCASTI64X2 96(ks), K1, Z24    
	VBROADCASTI64X2 112(ks), K1, Z25   
	VBROADCASTI64X2 128(ks), K1, Z26   
	VBROADCASTI64X2 144(ks), K1, Z27   
	VBROADCASTI64X2 160(ks), K1, Z28   

  // init ctr in Z10..Z17
	VMOVUPS ·vmsswapMask<>(SB), Z3         
	VBROADCASTI64X2 0(ctrPtr), K1, Z0  // [ iv iv iv iv ] 
	VMOVUPS Z0, Z2                     
	VPXORQ Z18, Z0, Z0                 
	VMOVUPS Z18, Z1                    
	VPSHUFB Z3, Z1, Z1                 
	KMOVW ·ivBlendMask<>(SB), K1       
	VPBLENDMD Z0, Z1, K1, Z1           // [ k0[96..127] iv^k0[95..0] ], 4 wide
	VPXORQ Z0, Z0, Z0                  
	VPSHUFB Z3, Z2, Z2                  
	VPBLENDMD Z0, Z2, K1, Z0           // [ iv[96..127]) iv[95..0] ], 4 wide               
	VMOVUPS ·ctr_init<>(SB), Z2        
	VPADDD Z2, Z0, Z0                   
	VPXORQ Z0, Z1, Z10                  
  VPSHUFB Z3, Z10, Z10                
	VMOVUPS ·ctr_incr<>(SB), Z2        
  vincrement(Z11)
  vincrement(Z12)
  vincrement(Z13)
  vincrement(Z14)
  vincrement(Z15)
  vincrement(Z16)
  vincrement(Z17)
	KXNORQ K2, K2, K2
  VMOVUPS ·vbswapMask<>(SB), Z31     

  // do aes rounds
  vaesRound(Z19)
  vaesRound(Z20)
  vaesRound(Z21)
  vaesRound(Z22)
  vaesRound(Z23)
  vaesRound(Z24)
  vaesRound(Z25)
  vaesRound(Z26)
  vaesRound(Z27)
	VBROADCASTI64X2 (10*16)(ks), K2, Z28   
	VBROADCASTI64X2 (11*16)(ks), K2, Z29   
	CMPQ NR, $12
	JB encLast1
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (12*16)(ks), K2, Z28   
	VBROADCASTI64X2 (13*16)(ks), K2, Z29   
	JE encLast1
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (14*16)(ks), K2, Z28   
encLast1:
  vaesLastRound(Z28)

	// generate and store ciphertext
	VMOVUPS (64*0)(ptx), T1v
  VPXORQ Z10, T1v, Z10
	VMOVUPS (64*1)(ptx), T1v
  VPXORQ Z11, T1v, Z11
	VMOVUPS (64*2)(ptx), T1v
  VPXORQ Z12, T1v, Z12
	VMOVUPS (64*3)(ptx), T1v
  VPXORQ Z13, T1v, Z13
	VMOVUPS (64*4)(ptx), T1v
  VPXORQ Z14, T1v, Z14
	VMOVUPS (64*5)(ptx), T1v
  VPXORQ Z15, T1v, Z15
	VMOVUPS (64*6)(ptx), T1v
  VPXORQ Z16, T1v, Z16
	VMOVUPS (64*7)(ptx), T1v
  VPXORQ Z17, T1v, Z17
	VMOVUPS Z10, (64*0)(ctx)
	VPSHUFB Z31, Z10, Z10
	VPXORQ ACC0z, Z10, Z18
	VMOVUPS Z11, (64*1)(ctx)
	VPSHUFB Z31, Z11, Z19
	VMOVUPS Z12, (64*2)(ctx)
	VPSHUFB Z31, Z12, Z20
	VMOVUPS Z13, (64*3)(ctx)
	VPSHUFB Z31, Z13, Z21
	VMOVUPS Z14, (64*4)(ctx)
	VPSHUFB Z31, Z14, Z22
	VMOVUPS Z15, (64*5)(ctx)
	VPSHUFB Z31, Z15, Z23
	VMOVUPS Z16, (64*6)(ctx)
	VPSHUFB Z31, Z16, Z24
	VMOVUPS Z17, (64*7)(ctx)
	VPSHUFB Z31, Z17, Z25

	LEAQ 512(ptx), ptx
	LEAQ 512(ctx), ctx

gcmAesEncOctetsLoop:
	CMPQ ptxLen, $512
	JB gcmAesEncOctetsEnd
	SUBQ $512, ptxLen

	vincrement(Z10)
  VPXORQ ACC0z, ACC0z, ACC0z
	VPXORQ ACCMz, ACCMz, ACCMz
	VEXTRACTI64X2 $0, Y18, T0x
  MOVOU (16*0)(pTbl), ACC0
	MOVOU (512+16*0)(pTbl), ACCMx
  vincrement(Z11)
  VMOVUPS ACC0z, ACC1z
  PSHUFD $78, T0x, T1x
  PXOR T0x, T1x
  vincrement(Z12)
	PCLMULQDQ $0x00, T0x, ACC0
  vincrement(Z13)
	PCLMULQDQ $0x11, T0x, ACC1x
  vincrement(Z14)
 	PCLMULQDQ $0x00, T1x, ACCMx
  vincrement(Z15)
  vincrement(Z16)
  vincrement(Z17)

	// mask lane 0 
	MOVW $0xfc, R12
  KMOVW R12, K1
  vcombinedRound(Z18,0)
	// unmask lane 0
  KXNORQ K1, K1, K1                 
  vcombinedRound(Z19,1)
  vcombinedRound(Z20,2)
  vcombinedRound(Z21,3)
  vcombinedRound(Z22,4)
  vcombinedRound(Z23,5)
  vcombinedRound(Z24,6)
  vcombinedRound(Z25,7)
	VBROADCASTI64X2 (9*16)(ks), K2, Z29   
  vaesRound(Z29)
  VBROADCASTI64X2 (10*16)(ks), K2, Z28 
	VBROADCASTI64X2 (11*16)(ks), K2, Z29   
	CMPQ NR, $12
	JB encLast2
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (12*16)(ks), K2, Z28   
	VBROADCASTI64X2 (13*16)(ks), K2, Z29   
	JE encLast2
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (14*16)(ks), K2, Z28   

encLast2:
  vaesLastRound(Z28)

	VPXORQ ACC0z, ACCMz, ACCMz
	VPXORQ ACC1z, ACCMz, ACCMz
  xorh512to128(ACCMz)
	MOVOU ACCMx, T0x
	PSRLDQ $8, ACCMx
	PSLLDQ $8, T0x
	// xorh uses T1
  xorh512to128(ACC1z)
	PXOR ACCMx, ACC1x
  xorh512to128(ACC0z)
	PXOR T0x, ACC0
	vreduceRound
	vreduceRound
  PXOR ACC1x, ACC0

  // generate and store cipher text
  VMOVUPS (64*0)(ptx), T1v
	VPXORQ Z10, T1v, Z10
  VMOVUPS (64*1)(ptx), T1v
	VPXORQ Z11, T1v, Z11
  VMOVUPS (64*2)(ptx), T1v
	VPXORQ Z12, T1v, Z12
  VMOVUPS (64*3)(ptx), T1v
	VPXORQ Z13, T1v, Z13
  VMOVUPS (64*4)(ptx), T1v
	VPXORQ Z14, T1v, Z14
  VMOVUPS (64*5)(ptx), T1v
	VPXORQ Z15, T1v, Z15
  VMOVUPS (64*6)(ptx), T1v
	VPXORQ Z16, T1v, Z16
  VMOVUPS (64*7)(ptx), T1v
	VPXORQ Z17, T1v, Z17
	VMOVUPS Z10, (64*0)(ctx)
	VPSHUFB Z31, Z10, Z18
	MOVW $0x3, R12
  KMOVW R12, K1
	VPXORQ ACC0z, Z18, K1, Z18
	VMOVUPS Z11, (64*1)(ctx)
 	VPSHUFB Z31, Z11, Z19
	VMOVUPS Z12, (64*2)(ctx)
 	VPSHUFB Z31, Z12, Z20
	VMOVUPS Z13, (64*3)(ctx)
 	VPSHUFB Z31, Z13, Z21
	VMOVUPS Z14, (64*4)(ctx)
 	VPSHUFB Z31, Z14, Z22
	VMOVUPS Z15, (64*5)(ctx)
 	VPSHUFB Z31, Z15, Z23
	VMOVUPS Z16, (64*6)(ctx)
 	VPSHUFB Z31, Z16, Z24
	VMOVUPS Z17, (64*7)(ctx)
 	VPSHUFB Z31, Z17, Z25

	LEAQ 512(ptx), ptx
	LEAQ 512(ctx), ctx
  JMP gcmAesEncOctetsLoop

gcmAesEncOctetsEnd:
	VPXORQ ACC0z, ACC0z, ACC0z
	VPXORQ ACC1z, ACC1z, ACC1z
	VPXORQ ACCMz, ACCMz, ACCMz
	VEXTRACTI64X2 $0, Y18, T0x;
	MOVOU (16*0)(pTbl), ACC0
	MOVOU (512+16*0)(pTbl), ACCMx
	MOVOU ACC0, ACC1x
	PSHUFD $78, T0x, T1x
  PXOR T0x, T1x
  PCLMULQDQ $0x00, T0x, ACC0
  PCLMULQDQ $0x11, T0x, ACC1x
  PCLMULQDQ $0x00, T1x, ACCMx

	// mask [0:127] for first round
	MOVW $0xfc, R12
  KMOVW R12, K1
	vmulRound(Z18,0)
  // 4-wide on remaining rounds
  KXNORQ K1, K1, K1                 
	vmulRound(Z19,1)
	vmulRound(Z20,2)
	vmulRound(Z21,3)
	vmulRound(Z22,4)
	vmulRound(Z23,5)
	vmulRound(Z24,6)
	vmulRound(Z25,7)

	VPXORQ ACC0z, ACCMz, ACCMz
  VPXORQ ACC1z, ACCMz, ACCMz
	xorh512to128(ACCMz)
	MOVOU ACCMx, T0x
	PSRLDQ $8, ACCMx
  PSLLDQ $8, T0x
	xorh512to128(ACC1z)
  PXOR ACCMx, ACC1x
	xorh512to128(ACC0z)
	PXOR T0x, ACC0
  vreduceRound
  vreduceRound
  PXOR ACC1x, ACC0

	TESTQ ptxLen, ptxLen
	JE gcmAesEncDone

	SUBQ $7, aluCTR

	// singles uses scalar ctr
	vincrement(Z7)
	VEXTRACTPS $3, X0, aluCTR
	MOVL (3*4)(ks), aluK
	BSWAPL aluK
	VEXTRACTPS $3, X1, aluTMP
	XORL aluCTR, aluTMP
	BSWAPL aluTMP
	MOVOU X1, (128 + 0*16)(SP)
	MOVL aluTMP, (128 + 0*16 + 12)(SP)
	JMP gcmAesEncSingles1

gcmAesEncSingles:
  MOVOU (ctrPtr), B0
  MOVL (3*4)(ctrPtr), aluCTR
  MOVOU (ks), T0
  MOVL (3*4)(ks), aluK
  BSWAPL aluCTR
  BSWAPL aluK
	PXOR B0, T0
  MOVOU T0, (128 + 0*16)(SP)
  increment(0)

gcmAesEncSingles1:
	MOVOU bswapMask<>(SB), BSWAP
	MOVOU gcmPoly<>(SB), POLY
	MOVOU (16*1)(ks), B1
	MOVOU (16*2)(ks), B2
	MOVOU (16*3)(ks), B3
	MOVOU (16*4)(ks), B4
	MOVOU (16*5)(ks), B5
	MOVOU (16*6)(ks), B6
	MOVOU (16*7)(ks), B7
	MOVOU (16*31)(pTbl), T2

gcmAesEncSinglesLoop:
		CMPQ ptxLen, $16
		JB gcmAesEncTail
		SUBQ $16, ptxLen

		MOVOU (128 + 0*16)(SP), B0
		increment(0)

		AESENC B1, B0
		AESENC B2, B0
		AESENC B3, B0
		AESENC B4, B0
		AESENC B5, B0
		AESENC B6, B0
		AESENC B7, B0
		MOVOU (16*8)(ks), T0
		AESENC T0, B0
		MOVOU (16*9)(ks), T0
		AESENC T0, B0
		MOVOU (16*10)(ks), T0
		CMPQ NR, $12
		JB encLast3
		AESENC T0, B0
		MOVOU (16*11)(ks), T0
		AESENC T0, B0
		MOVOU (16*12)(ks), T0
		JE encLast3
		AESENC T0, B0
		MOVOU (16*13)(ks), T0
		AESENC T0, B0
		MOVOU (16*14)(ks), T0
encLast3:
		AESENCLAST T0, B0

		MOVOU (ptx), T0
		PXOR T0, B0
		MOVOU B0, (ctx)

		PSHUFB BSWAP, B0
		PXOR ACC0, B0

		MOVOU T2, ACC0
		MOVOU T2, ACC1
		MOVOU (16*63)(pTbl), ACCM

		PSHUFD $78, B0, T0
		PXOR B0, T0
		PCLMULQDQ $0x00, B0, ACC0
		PCLMULQDQ $0x11, B0, ACC1
		PCLMULQDQ $0x00, T0, ACCM

		PXOR ACC0, ACCM
		PXOR ACC1, ACCM
		MOVOU ACCM, T0
		PSRLDQ $8, ACCM
		PSLLDQ $8, T0
		PXOR ACCM, ACC1
		PXOR T0, ACC0

		reduceRound(ACC0)
		reduceRound(ACC0)
		PXOR ACC1, ACC0

		LEAQ (16*1)(ptx), ptx
		LEAQ (16*1)(ctx), ctx

	JMP gcmAesEncSinglesLoop

gcmAesEncTail:
	TESTQ ptxLen, ptxLen
	JE gcmAesEncDone

	MOVOU (8*16 + 0*16)(SP), B0

	AESENC B1, B0
	AESENC B2, B0
	AESENC B3, B0
	AESENC B4, B0
	AESENC B5, B0
	AESENC B6, B0
	AESENC B7, B0
	MOVOU (16*8)(ks), T0
	AESENC T0, B0
	MOVOU (16*9)(ks), T0
	AESENC T0, B0
	MOVOU (16*10)(ks), T0
	CMPQ NR, $12
	JB encLast4
	AESENC T0, B0
	MOVOU (16*11)(ks), T0
	AESENC T0, B0
	MOVOU (16*12)(ks), T0
	JE encLast4
	AESENC T0, B0
	MOVOU (16*13)(ks), T0
	AESENC T0, B0
	MOVOU (16*14)(ks), T0
encLast4:
	AESENCLAST T0, B0
	MOVOU B0, T0

	LEAQ -1(ptx)(ptxLen*1), ptx

	MOVQ ptxLen, aluTMP
	SHLQ $4, aluTMP

	LEAQ andMask<>(SB), aluCTR
	MOVOU -16(aluCTR)(aluTMP*1), T1

	PXOR B0, B0
ptxLoadLoop:
		PSLLDQ $1, B0
		PINSRB $0, (ptx), B0
		LEAQ -1(ptx), ptx
		DECQ ptxLen
	JNE ptxLoadLoop

	PXOR T0, B0
	PAND T1, B0
	MOVOU B0, (ctx)	// I assume there is always space, due to TAG in the end of the CT

	PSHUFB BSWAP, B0
	PXOR ACC0, B0

	MOVOU T2, ACC0
	MOVOU T2, ACC1
	MOVOU (16*63)(pTbl), ACCM

	PSHUFD $78, B0, T0
	PXOR B0, T0
	PCLMULQDQ $0x00, B0, ACC0
	PCLMULQDQ $0x11, B0, ACC1
	PCLMULQDQ $0x00, T0, ACCM

	PXOR ACC0, ACCM
	PXOR ACC1, ACCM
	MOVOU ACCM, T0
	PSRLDQ $8, ACCM
	PSLLDQ $8, T0
	PXOR ACCM, ACC1
	PXOR T0, ACC0

	reduceRound(ACC0)
	reduceRound(ACC0)
	PXOR ACC1, ACC0

gcmAesEncDone:
	MOVOU ACC0, (tPtr)
	RET

#undef increment
#undef POLY

// func gcmAesDecv(productTable *[1024]byte, dst, src []byte, ctr, T *[16]byte, ks []uint32)
TEXT ·gcmAesDecv(SB),0,$128-96
#define POLY X14
#define VBSWAP Z31
#define increment(i) ADDL $1, aluCTR; MOVL aluCTR, aluTMP; XORL aluK, aluTMP; BSWAPL aluTMP; MOVL aluTMP, (3*4 + i*16)(SP)
#define vcombinedDecRound(z,i) \
	VBROADCASTI64X2 ((i+1)*16)(ks), K2, T0z; \   
  VAESENC T0z, Z10, Z10; \
  VAESENC T0z, Z11, Z11; \ 
  VAESENC T0z, Z12, Z12; \
  VAESENC T0z, Z13, Z13; \
  VMOVUPS (64*i)(pTbl), T1v; \
	VMOVUPS T1v, T2v; \
  VAESENC T0z, Z14, Z14; \
  VAESENC T0z, Z15, Z15; \
  VAESENC T0z, Z16, Z16; \
  VAESENC T0z, Z17, Z17; \
	VMOVUPS (64*i)(ctx), z; \
	VPSHUFB VBSWAP, z, T0z; \
	VPCLMULQDQ $0x00, T0z, T1v, T1v; \
  VPCLMULQDQ $0x11, T0z, T2v, T2v; \
	VPXORQ T1v, ACC0z, K1, ACC0z; \
	VPSHUFD $78, T0z, T1v; \
	VPXORQ T1v, T0z, T0z; \
	VPXORQ T2v, ACC1z, K1, ACC1z; \
  VMOVUPS (512+64*i)(pTbl), T1v; \
	VPCLMULQDQ $0x00, T0z, T1v, T0z; \
	VPXORQ T0z, ACCMz, K1, ACCMz

	MOVQ productTable+0(FP), pTbl
	MOVQ dst+8(FP), ptx
	MOVQ src_base+32(FP), ctx
	MOVQ src_len+40(FP), ptxLen
	MOVQ ctr+56(FP), ctrPtr
	MOVQ T+64(FP), tPtr
	MOVQ ks_base+72(FP), ks
	MOVQ ks_len+80(FP), NR

	SHRQ $2, NR
	DECQ NR

  MOVOU gcmPoly<>(SB), T1x
	VMOVUPS T1v, POLYz
	VPXORQ ACC0z, ACC0z, ACC0z
	MOVOU (tPtr), ACC0
	VPXORQ ACC1z, ACC1z, ACC1z
	VPXORQ ACCMz, ACCMz, ACCMz

	CMPQ ptxLen, $512
	JB gcmAesDecSingles

  // init ctr in Z10..Z17
	VMOVUPS ·vmsswapMask<>(SB), Z3         
	KXNORQ K2, K2, K2 
	VBROADCASTI64X2 0(ctrPtr), K2, Z0 
	VBROADCASTI64X2 (0*16)(ks), K2, Z18   
 	VMOVUPS Z0, Z2                     
	VPXORQ Z18, Z0, Z0                 
	VMOVUPS Z18, Z1                    
	VPSHUFB Z3, Z1, Z1                 
	KMOVW ·ivBlendMask<>(SB), K1       
	VPBLENDMD Z0, Z1, K1, Z1           
	VPXORQ Z0, Z0, Z0                  
	VPSHUFB Z3, Z2, Z2                  
	VPBLENDMD Z0, Z2, K1, Z0              
	VMOVUPS ·ctr_init<>(SB), Z2        
	VPADDD Z2, Z0, Z0                   
	VPXORQ Z0, Z1, Z10                  
  VPSHUFB Z3, Z10, Z10                
	VMOVUPS ·ctr_incr<>(SB), Z2        
  VMOVUPS ·vbswapMask<>(SB), VBSWAP  
	VBROADCASTI64X2 (9*16)(ks), K2, Z26   

gcmAesDecOctetsLoop:
	CMPQ ptxLen, $512
	JB gcmAesDecEndOctets
	SUBQ $512, ptxLen

	MOVOU (16*0)(ctx), T0x
	VPSHUFB VBSWAP, T0z, T0z
	PXOR ACC0, T0x
	PSHUFD $78, T0x, T1x
	PXOR T0x, T1x

	VPXORQ ACC0z, ACC0z, ACC0z
	VPXORQ ACCMz, ACCMz, ACCMz
	MOVOU (16*0)(pTbl), ACC0
	MOVOU (512+16*0)(pTbl), ACCMx
  vincrement(Z11)
  VMOVUPS ACC0z, ACC1z
  vincrement(Z12)
	PCLMULQDQ $0x00, T0x, ACC0
  vincrement(Z13)
	PCLMULQDQ $0x11, T0x, ACC1x
  vincrement(Z14)
 	PCLMULQDQ $0x00, T1x, ACCMx
  vincrement(Z15)
	vincrement(Z16)
	vincrement(Z17)

  // mask lane 0
	MOVW $0xfc, R12
  KMOVW R12, K1
	vcombinedDecRound(Z18,0)
	// unmask lane 0
  KXNORQ K1, K1, K1                 
	vcombinedDecRound(Z19,1)
	vcombinedDecRound(Z20,2)
	vcombinedDecRound(Z21,3)
	vcombinedDecRound(Z22,4)
	vcombinedDecRound(Z23,5)
	vcombinedDecRound(Z24,6)
	vcombinedDecRound(Z25,7)
  vaesRound(Z26)
  VBROADCASTI64X2 (10*16)(ks), K2, Z28 
	CMPQ NR, $12
	JB decLast1
	VBROADCASTI64X2 (11*16)(ks), K2, Z29   
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (12*16)(ks), K2, Z28   
	VBROADCASTI64X2 (13*16)(ks), K2, Z29   
	JE decLast1
  vaesRound(Z28)
  vaesRound(Z29)
	VBROADCASTI64X2 (14*16)(ks), K2, Z28   

decLast1:
  vaesLastRound(Z28)

  VPXORQ ACC0z, ACCMz, ACCMz
	VPXORQ ACC1z, ACCMz, ACCMz
  xorh512to128(ACCMz)
	MOVOU ACCMx, T0x
	PSRLDQ $8, ACCMx
	PSLLDQ $8, T0x
	// xorh uses T1
  xorh512to128(ACC1z)
	PXOR ACCMx, ACC1x
  xorh512to128(ACC0z)
	PXOR T0x, ACC0
	vreduceRound
	vreduceRound
  PXOR ACC1x, ACC0

	// generate and store plain text 
  VPXORQ Z10, Z18, Z18
	VMOVUPS Z18, (64*0)(ptx)
  vincrement(Z10)
  VPXORQ Z11, Z19, Z19
  VMOVUPS Z19, (64*1)(ptx)
  VPXORQ Z12, Z20, Z20
  VMOVUPS Z20, (64*2)(ptx)
  VPXORQ Z13, Z21, Z21
  VMOVUPS Z21, (64*3)(ptx)
  VPXORQ Z14, Z22, Z22
  VMOVUPS Z22, (64*4)(ptx)
  VPXORQ Z15, Z23, Z23
  VMOVUPS Z23, (64*5)(ptx)
  VPXORQ Z16, Z24, Z24
  VMOVUPS Z24, (64*6)(ptx)
  VPXORQ Z17, Z25, Z25
  VMOVUPS Z25, (64*7)(ptx)

  LEAQ 512(ptx), ptx
  LEAQ 512(ctx), ctx

  JMP gcmAesDecOctetsLoop

gcmAesDecEndOctets:

	SUBQ $7, aluCTR

// singles uses scalar ctx
	VEXTRACTPS $3, X0, aluCTR
	MOVL (3*4)(ks), aluK
	BSWAPL aluK
	VEXTRACTPS $3, X1, aluTMP
	XORL aluCTR, aluTMP
	BSWAPL aluTMP
	MOVOU X1, (0*16)(SP)
	MOVL aluTMP, (0*16 + 12)(SP)
	JMP gcmAesDecSingles1

gcmAesDecSingles:
  MOVOU (ctrPtr), B0
  MOVL (3*4)(ctrPtr), aluCTR
  MOVOU (ks), T0
  MOVL (3*4)(ks), aluK
  BSWAPL aluCTR
  BSWAPL aluK
	PXOR B0, T0
  MOVOU T0, (0*16)(SP)
  increment(0)

gcmAesDecSingles1:
	MOVOU bswapMask<>(SB), BSWAP
	MOVOU gcmPoly<>(SB), POLY
	MOVOU (16*1)(ks), B1
	MOVOU (16*2)(ks), B2
	MOVOU (16*3)(ks), B3
	MOVOU (16*4)(ks), B4
	MOVOU (16*5)(ks), B5
	MOVOU (16*6)(ks), B6
	MOVOU (16*7)(ks), B7

	MOVOU (16*31)(pTbl), T2

gcmAesDecSinglesLoop:

		CMPQ ptxLen, $16
		JB gcmAesDecTail
		SUBQ $16, ptxLen

		MOVOU (ctx), B0
		MOVOU B0, T1
		PSHUFB BSWAP, B0
		PXOR ACC0, B0

		MOVOU T2, ACC0
		MOVOU T2, ACC1
		MOVOU (512+16*31)(pTbl), ACCM

		PCLMULQDQ $0x00, B0, ACC0
		PCLMULQDQ $0x11, B0, ACC1
		PSHUFD $78, B0, T0
		PXOR B0, T0
		PCLMULQDQ $0x00, T0, ACCM

		PXOR ACC0, ACCM
		PXOR ACC1, ACCM
		MOVOU ACCM, T0
		PSRLDQ $8, ACCM
		PSLLDQ $8, T0
		PXOR ACCM, ACC1
		PXOR T0, ACC0

		reduceRound(ACC0)
		reduceRound(ACC0)
		PXOR ACC1, ACC0

		MOVOU (0*16)(SP), B0
		increment(0)
		AESENC B1, B0
		AESENC B2, B0
		AESENC B3, B0
		AESENC B4, B0
		AESENC B5, B0
		AESENC B6, B0
		AESENC B7, B0
		MOVOU (16*8)(ks), T0
		AESENC T0, B0
		MOVOU (16*9)(ks), T0
		AESENC T0, B0
		MOVOU (16*10)(ks), T0
		CMPQ NR, $12
		JB decLast2
		AESENC T0, B0
		MOVOU (16*11)(ks), T0
		AESENC T0, B0
		MOVOU (16*12)(ks), T0
		JE decLast2
		AESENC T0, B0
		MOVOU (16*13)(ks), T0
		AESENC T0, B0
		MOVOU (16*14)(ks), T0
decLast2:
		AESENCLAST T0, B0

		PXOR T1, B0
		MOVOU B0, (ptx)

		LEAQ (16*1)(ptx), ptx
		LEAQ (16*1)(ctx), ctx

	JMP gcmAesDecSinglesLoop

gcmAesDecTail:

	TESTQ ptxLen, ptxLen
	JE gcmAesDecDone

	MOVQ ptxLen, aluTMP
	SHLQ $4, aluTMP
	LEAQ andMask<>(SB), aluCTR
	MOVOU -16(aluCTR)(aluTMP*1), T1

	MOVOU (ctx), B0	// I assume there is TAG attached to the ctx, and there is no read overflow
	PAND T1, B0

	MOVOU B0, T1
	PSHUFB BSWAP, B0
	PXOR ACC0, B0

	MOVOU (16*31)(pTbl), ACC0
	MOVOU (512+16*31)(pTbl), ACCM
	MOVOU ACC0, ACC1

	PCLMULQDQ $0x00, B0, ACC0
	PCLMULQDQ $0x11, B0, ACC1
	PSHUFD $78, B0, T0
	PXOR B0, T0
	PCLMULQDQ $0x00, T0, ACCM

	PXOR ACC0, ACCM
	PXOR ACC1, ACCM
	MOVOU ACCM, T0
	PSRLDQ $8, ACCM
	PSLLDQ $8, T0
	PXOR ACCM, ACC1
	PXOR T0, ACC0

	reduceRound(ACC0)
	reduceRound(ACC0)
	PXOR ACC1, ACC0

	MOVOU (0*16)(SP), B0
	increment(0)
	AESENC B1, B0
	AESENC B2, B0
	AESENC B3, B0
	AESENC B4, B0
	AESENC B5, B0
	AESENC B6, B0
	AESENC B7, B0
	MOVOU (16*8)(ks), T0
	AESENC T0, B0
	MOVOU (16*9)(ks), T0
	AESENC T0, B0
	MOVOU (16*10)(ks), T0
	CMPQ NR, $12
	JB decLast3
	AESENC T0, B0
	MOVOU (16*11)(ks), T0
	AESENC T0, B0
	MOVOU (16*12)(ks), T0
	JE decLast3
	AESENC T0, B0
	MOVOU (16*13)(ks), T0
	AESENC T0, B0
	MOVOU (16*14)(ks), T0
decLast3:
	AESENCLAST T0, B0
	PXOR T1, B0

ptxStoreLoop:
		PEXTRB $0, B0, (ptx)
		PSRLDQ $1, B0
		LEAQ 1(ptx), ptx
		DECQ ptxLen

	JNE ptxStoreLoop

gcmAesDecDone:

	MOVOU ACC0, (tPtr)
	RET
