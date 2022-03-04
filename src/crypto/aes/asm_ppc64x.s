// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on CRYPTOGAMS code with the following comment:
// # ====================================================================
// # Written by Andy Polyakov <appro@openssl.org> for the OpenSSL
// # project. The module is, however, dual licensed under OpenSSL and
// # CRYPTOGAMS licenses depending on where you obtain it. For further
// # details see http://www.openssl.org/~appro/cryptogams/.
// # ====================================================================

// Original code can be found at the link below:
// https://github.com/dot-asm/cryptogams/blob/master/ppc/aesp8-ppc.pl

// Some function names were changed to be consistent with Go function
// names. For instance, function aes_p8_set_{en,de}crypt_key become
// set{En,De}cryptKeyAsm. I also split setEncryptKeyAsm in two parts
// and a new session was created (doEncryptKeyAsm). This was necessary to
// avoid arguments overwriting when setDecryptKeyAsm calls setEncryptKeyAsm.
// There were other modifications as well but kept the same functionality.

//go:build ppc64 || ppc64le

#include "textflag.h"

// For expandKeyAsm
#define INP     R3
#define BITS    R4
#define OUT     R5
#define PTR     R6
#define CNT     R7
#define ROUNDS  R8
#define OUTDEC  R9
#define OUTDECP R10
#define ZERO    V0
#define IN0     V1
#define IN1     V2
#define KEY     V3
#define RCON    V4
#define MASK    V5
#define TMP     V6
#define STAGE   V7

// For P9 instruction emulation.
#define PERMV   V21
#define TMP2    V22

DATA ·rcon+0x00(SB)/8, $0x0f0e0d0c0b0a0908 // Permute for vector doubleword endian swap
DATA ·rcon+0x08(SB)/8, $0x0706050403020100
DATA ·rcon+0x10(SB)/8, $0x0100000001000000 // RCON
DATA ·rcon+0x18(SB)/8, $0x0100000001000000 // RCON
DATA ·rcon+0x20(SB)/8, $0x1b0000001b000000
DATA ·rcon+0x28(SB)/8, $0x1b0000001b000000
DATA ·rcon+0x30(SB)/8, $0x0d0e0f0c0d0e0f0c // MASK
DATA ·rcon+0x38(SB)/8, $0x0d0e0f0c0d0e0f0c // MASK
DATA ·rcon+0x40(SB)/8, $0x0000000000000000
DATA ·rcon+0x48(SB)/8, $0x0000000000000000
GLOBL ·rcon(SB), RODATA, $80


// Emulate unaligned BE vector loads on LE targets 
#ifdef GOARCH_ppc64le

#define P8_LXVB16X(RA,RB,VT) \
	LXVD2X (RA+RB), VT \
	VPERM VT, VT, PERMV, VT

#define P8_STXVB16X(VS,RA,RB) \
	VPERM VS, VS, PERMV, TMP2 \
	STXVD2X TMP2, (RA+RB)

#define P8_XXBRD(VA,VT) \
	VPERM VA, VA, PERMV, VT

#else

#define P8_LXVB16X(RA,RB,VT) \
	LXVD2X (RA+RB), VT

#define P8_STXVB16X(VS,RA,RB) \
	STXVD2X VS, (RA+RB)

// nop for BE
#define P8_XXBRD(VA,VT)

#endif

// func expandKeyAsm(nr int, key *byte, enc, dec *uint32) {
TEXT ·expandKeyAsm(SB), NOSPLIT|NOFRAME, $0
	MOVD nr+0(FP), BITS
	MOVD key+8(FP), INP
	MOVD enc+16(FP), OUT
	MOVD dec+24(FP), OUTDECP

	// Arguments are checked prior to entry.
	// BITS is either 10, 12, 14 (128, 196, or 256)
	// INP/OUT/OUTDEC are non-null pointers

	MOVD $·rcon(SB), PTR // PTR point to rcon addr
#ifdef GOARCH_ppc64le
	LVX (PTR), PERMV // Load LE to BE permute.
#endif
	ADD $0x10, PTR

	// Get key from memory and write aligned into VR
	P8_LXVB16X(INP, R0, IN0)
	ADD      $16, INP, INP
	CMPW     BITS, $12
	VSPLTISB $0x0f, MASK
	LVX      (PTR)(R0), RCON
	MOVD     $32, R11
	LXVD2X   (PTR)(R11), MASK
	ADD      $0x10, PTR, PTR
	MOVD     $8, CNT
	VXOR     ZERO, ZERO, ZERO
	MOVD     CNT, CTR // Set the counter to 8 (rounds)

	ADD $160, OUTDECP, OUTDEC
	BLT loop128
	ADD $192, OUTDECP, OUTDEC
	BEQ l192
	ADD $224, OUTDECP, OUTDEC
	JMP l256

loop128:
	// Key schedule (Round 1 to 8)
	VPERM       IN0, IN0, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	STXVD2X     IN0, (OUT)
	STXVD2X     IN0, (OUTDEC)
	VCIPHERLAST KEY, RCON, KEY
	ADD         $16, OUT, OUT
	ADD         $-16, OUTDEC, OUTDEC

	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VADDUWM RCON, RCON, RCON
	VXOR    IN0, KEY, IN0
	BC      0x10, 0, loop128

	LXVD2X (PTR)(R0), RCON // Last two round keys

	// Key schedule (Round 9)
	VPERM       IN0, IN0, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	STXVD2X     IN0, (OUT)
	STXVD2X     IN0, (OUTDEC)
	VCIPHERLAST KEY, RCON, KEY
	ADD         $16, OUT, OUT
	ADD         $-16, OUTDEC, OUTDEC

	// Key schedule (Round 10)
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VADDUWM RCON, RCON, RCON
	VXOR    IN0, KEY, IN0

	VPERM       IN0, IN0, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	STXVD2X     IN0, (OUT)
	STXVD2X     IN0, (OUTDEC)
	VCIPHERLAST KEY, RCON, KEY
	ADD         $16, OUT, OUT
	ADD         $-16, OUTDEC, OUTDEC

	// Key schedule (Round 11)
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VXOR    IN0, KEY, IN0
	VSLDOI  $12, ZERO, IN0, TMP
	STXVD2X IN0, (OUT)
	STXVD2X IN0, (OUTDEC)
	RET

l192:
	LXSDX (INP)(R0), IN1
	P8_XXBRD(IN1, IN1) // Load next 8 bytes into the upper-half of VSR, and swap to BE ordering.
	MOVD     $4, CNT
	STXVD2X  IN0, (OUT)
	STXVD2X  IN0, (OUTDEC)
	ADD      $16, OUT, OUT
	ADD      $-16, OUTDEC, OUTDEC
	VSPLTISB $8, KEY
	MOVD     CNT, CTR
	VSUBUBM  MASK, KEY, MASK

loop192:
	VPERM       IN1, IN1, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	VCIPHERLAST KEY, RCON, KEY

	VXOR   IN0, TMP, IN0
	VSLDOI $12, ZERO, TMP, TMP
	VXOR   IN0, TMP, IN0
	VSLDOI $12, ZERO, TMP, TMP
	VXOR   IN0, TMP, IN0

	VSLDOI  $8, ZERO, IN1, STAGE
	VSPLTW  $3, IN0, TMP
	VXOR    TMP, IN1, TMP
	VSLDOI  $12, ZERO, IN1, IN1
	VADDUWM RCON, RCON, RCON
	VXOR    IN1, TMP, IN1
	VXOR    IN0, KEY, IN0
	VXOR    IN1, KEY, IN1
	VSLDOI  $8, STAGE, IN0, STAGE

	VPERM       IN1, IN1, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	STXVD2X     STAGE, (OUT)
	STXVD2X     STAGE, (OUTDEC)
	VCIPHERLAST KEY, RCON, KEY
	ADD         $16, OUT, OUT
	ADD         $-16, OUTDEC, OUTDEC

	VSLDOI  $8, IN0, IN1, STAGE
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	STXVD2X STAGE, (OUT)
	STXVD2X STAGE, (OUTDEC)
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	ADD     $16, OUT, OUT
	ADD     $-16, OUTDEC, OUTDEC

	VSPLTW  $3, IN0, TMP
	VXOR    TMP, IN1, TMP
	VSLDOI  $12, ZERO, IN1, IN1
	VADDUWM RCON, RCON, RCON
	VXOR    IN1, TMP, IN1
	VXOR    IN0, KEY, IN0
	VXOR    IN1, KEY, IN1
	STXVD2X IN0, (OUT)
	STXVD2X IN0, (OUTDEC)
	ADD     $15, OUT, INP
	ADD     $16, OUT, OUT
	ADD     $-16, OUTDEC, OUTDEC
	BC      0x10, 0, loop192
	RET

l256:
	P8_LXVB16X(INP,R0,IN1)
	MOVD    $7, CNT
	MOVD    $14, ROUNDS
	STXVD2X IN0, (OUT)
	STXVD2X IN0, (OUTDEC)
	ADD     $16, OUT, OUT
	ADD     $-16, OUTDEC, OUTDEC
	MOVD    CNT, CTR

loop256:
	VPERM       IN1, IN1, MASK, KEY
	VSLDOI      $12, ZERO, IN0, TMP
	STXVD2X     IN1, (OUT)
	STXVD2X     IN1, (OUTDEC)
	VCIPHERLAST KEY, RCON, KEY
	ADD         $16, OUT, OUT
	ADD         $-16, OUTDEC, OUTDEC

	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VSLDOI  $12, ZERO, TMP, TMP
	VXOR    IN0, TMP, IN0
	VADDUWM RCON, RCON, RCON
	VXOR    IN0, KEY, IN0
	STXVD2X IN0, (OUT)
	STXVD2X IN0, (OUTDEC)
	ADD     $15, OUT, INP
	ADD     $16, OUT, OUT
	ADD     $-16, OUTDEC, OUTDEC
	BC      0x12, 0, done

	VSPLTW $3, IN0, KEY
	VSLDOI $12, ZERO, IN1, TMP
	VSBOX  KEY, KEY

	VXOR   IN1, TMP, IN1
	VSLDOI $12, ZERO, TMP, TMP
	VXOR   IN1, TMP, IN1
	VSLDOI $12, ZERO, TMP, TMP
	VXOR   IN1, TMP, IN1

	VXOR IN1, KEY, IN1
	JMP  loop256

done:
	RET

// Cleanup expandKeyAsm register definitions
#undef INP
#undef OUT
#undef ROUNDS
#undef OUTDEC
#undef OUTDECP
#undef KEY
#undef TMP
#undef STAGE

// func encryptBlockAsm(nr int, xk *uint32, dst, src *byte)
TEXT ·encryptBlockAsm(SB), NOSPLIT|NOFRAME, $0
	MOVD nr+0(FP), R6   // Round count/Key size
	MOVD xk+8(FP), R5   // Key pointer
	MOVD dst+16(FP), R3 // Dest pointer
	MOVD src+24(FP), R4 // Src pointer
	MOVD $·rcon(SB), R7 // PTR point to rcon addr
	LVX  (R7), PERMV    // Load doubleword endian-swap permute

	CMPU R6, $10, CR1
	CMPU R6, $12, CR2
	CMPU R6, $14, CR3

	MOVD $16, R6
	MOVD $32, R7
	MOVD $48, R8
	MOVD $64, R9
	MOVD $80, R10
	MOVD $96, R11
	MOVD $112, R12

	P8_LXVB16X(R4, R0, V0)
	LXVD2X (R5)(R0), V1
	VXOR V0, V1, V0

	LXVD2X  (R5)(R6), V1
	LXVD2X  (R5)(R7), V2
	VCIPHER V0, V1, V0
	VCIPHER V0, V2, V0

	LXVD2X  (R5)(R8), V1
	LXVD2X  (R5)(R9), V2
	VCIPHER V0, V1, V0
	VCIPHER V0, V2, V0

	LXVD2X  (R5)(R10), V1
	LXVD2X  (R5)(R11), V2
	VCIPHER V0, V1, V0
	VCIPHER V0, V2, V0

	ADD $112, R5

	LXVD2X (R5)(R0), V1
	LXVD2X (R5)(R6), V2
	VCIPHER  V0, V1, V0
	VCIPHER  V0, V2, V0

	LXVD2X  (R5)(R7), V1
	LXVD2X  (R5)(R8), V2
	BEQ     CR1, Ldec_tail // Key size 10?
	VCIPHER V0, V1, V0
	VCIPHER V0, V2, V0

	LXVD2X  (R5)(R9), V1
	LXVD2X  (R5)(R10), V2
	BEQ     CR2, Ldec_tail // Key size 12?
	VCIPHER V0, V1, V0
	VCIPHER V0, V2, V0

	LXVD2X (R5)(R11), V1
	LXVD2X (R5)(R12), V2
	BNE    CR3, Linvalid_key_len // Not key size 14?

Ldec_tail:
	VCIPHER     V0, V1, V1
	VCIPHERLAST V1, V2, V2
	P8_STXVB16X(V2, R3, R0)
	RET

Linvalid_key_len:
	// Segfault, this should never happen. Only 3 keys sizes are created/used.
	MOVD R0, 0(R0)
	RET

// func decryptBlockAsm(nr int, xk *uint32, dst, src *byte)
TEXT ·decryptBlockAsm(SB), NOSPLIT|NOFRAME, $0
	MOVD nr+0(FP), R6   // Round count/Key size
	MOVD xk+8(FP), R5   // Key pointer
	MOVD dst+16(FP), R3 // Dest pointer
	MOVD src+24(FP), R4 // Src pointer

#ifdef GOARCH_ppc64le
	MOVD $·rcon(SB), R7 // PTR point to rcon addr
	LVX  (R7), PERMV    // Load LE to BE permute.
#endif

	CMPU R6, $10, CR1
	CMPU R6, $12, CR2
	CMPU R6, $14, CR3

	MOVD $16, R6
	MOVD $32, R7
	MOVD $48, R8
	MOVD $64, R9
	MOVD $80, R10
	MOVD $96, R11
	MOVD $112, R12

	P8_LXVB16X(R4, R0, V0)
	LXVD2X (R5)(R0), V1
	VXOR   V0, V1, V0

	LXVD2X   (R5)(R6), V1
	LXVD2X   (R5)(R7), V2
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	LXVD2X   (R5)(R8), V1
	LXVD2X   (R5)(R9), V2
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	LXVD2X   (R5)(R10), V1
	LXVD2X   (R5)(R11), V2
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	ADD $112, R5

	LXVD2X   (R5)(R0), V1
	LXVD2X   (R5)(R6), V2
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	LXVD2X   (R5)(R7), V1
	LXVD2X   (R5)(R8), V2
	BEQ      CR1, Ldec_tail // Key size 10?
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	LXVD2X   (R5)(R9), V1
	LXVD2X   (R5)(R10), V2
	BEQ      CR2, Ldec_tail // Key size 12?
	VNCIPHER V0, V1, V0
	VNCIPHER V0, V2, V0

	LXVD2X   (R5)(R11), V1
	LXVD2X   (R5)(R12), V2
	BNE      CR3, Linvalid_key_len // Not key size 14?

Ldec_tail:
	VNCIPHER     V0, V1, V1
	VNCIPHERLAST V1, V2, V2
	P8_STXVB16X(V2, R3, R0)
	RET

Linvalid_key_len:
	// Segfault, this should never happen. Only 3 keys sizes are created/used.
	MOVD R0, 0(R0)
	RET

#define INP R3
#define OUTP R4
#define LEN R5
#define KEYP R6
#define ROUNDS R7
#define IVP R8
#define ENC R9

#define INOUT V2
#define TMP V3
#define IVEC V4

// Load the crypt key into VSRs.
// Rkeyp holds the key pointer. It is clobbered.
// R12,R14-R21 are scratch registers.
// For keyp of 10, V6, V11-V20 hold the expanded key.
// For keyp of 12, V6, V9-V20 hold the expanded key.
// For keyp of 14, V6, V7-V20 hold the expanded key.
#define LOAD_KEY(Rkeyp) \
	MOVD   $16, R12 \
	MOVD   $32, R14 \
	MOVD   $48, R15 \
	MOVD   $64, R16 \
	MOVD   $80, R17 \
	MOVD   $96, R18 \
	MOVD   $112, R19 \
	MOVD   $128, R20 \
	MOVD   $144, R21 \
	LXVD2X (Rkeyp)(R0), V6 \
	ADD    $16, Rkeyp \
	BEQ    CR1, L_start10 \
	BEQ    CR2, L_start12 \
	LXVD2X (Rkeyp)(R0), V7 \
	LXVD2X (Rkeyp)(R12), V8 \
	ADD $32, Rkeyp \
	L_start12: \
	LXVD2X (Rkeyp)(R0), V9 \
	LXVD2X (Rkeyp)(R12), V10 \
	ADD $32, Rkeyp \
	L_start10: \
	LXVD2X (Rkeyp)(R0), V11 \
	LXVD2X (Rkeyp)(R12), V12 \
	LXVD2X (Rkeyp)(R14), V13 \
	LXVD2X (Rkeyp)(R15), V14 \
	LXVD2X (Rkeyp)(R16), V15 \
	LXVD2X (Rkeyp)(R17), V16 \
	LXVD2X (Rkeyp)(R18), V17 \
	LXVD2X (Rkeyp)(R19), V18 \
	LXVD2X (Rkeyp)(R20), V19 \
	LXVD2X (Rkeyp)(R21), V20 \
	Lstart:

// Perform aes cipher operation for keysize 10/12/14 based on
// VSRs loaded by LOAD_KEY, and CR1/CR2 for 14 and 12.
// One of CR1EQ-CR3EQ is true for keysize 10,12,14 respectively.
#define CIPHER_BLOCK(Vin, Vout, vcipher, vciphel, label10, label12) \
	VXOR    Vin, V6, Vout \
	BEQ     CR1, label10 \
	BEQ     CR2, label12 \
	vcipher Vout, V7, Vout \
	vcipher Vout, V8, Vout \
	label12: \
	vcipher Vout, V9, Vout \
	vcipher Vout, V10, Vout \
	label10: \
	vcipher Vout, V11, Vout \
	vcipher Vout, V12, Vout \
	vcipher Vout, V13, Vout \
	vcipher Vout, V14, Vout \
	vcipher Vout, V15, Vout \
	vcipher Vout, V16, Vout \
	vcipher Vout, V17, Vout \
	vcipher Vout, V18, Vout \
	vcipher Vout, V19, Vout \
	vciphel Vout, V20, Vout \

#define CLEAR_KEYS() \
	VXOR V6, V6, V6 \
	VXOR V7, V7, V7 \
	VXOR V8, V8, V8 \
	VXOR V9, V9, V9 \
	VXOR V10, V10, V10 \
	VXOR V11, V11, V11 \
	VXOR V12, V12, V12 \
	VXOR V13, V13, V13 \
	VXOR V14, V14, V14 \
	VXOR V15, V15, V15 \
	VXOR V16, V16, V16 \
	VXOR V17, V17, V17 \
	VXOR V18, V18, V18 \
	VXOR V19, V19, V19 \
	VXOR V20, V20, V20

//func cryptBlocksChain(src, dst *byte, length int, key *uint32, keylen int, iv *byte, enc int)
TEXT ·cryptBlocksChain(SB), NOSPLIT|NOFRAME, $0
	MOVD src+0(FP), INP
	MOVD dst+8(FP), OUTP
	MOVD length+16(FP), LEN
	MOVD key+24(FP), KEYP
	MOVD keylen+32(FP), ROUNDS
	MOVD iv+40(FP), IVP
	MOVD enc+48(FP), ENC

#ifdef GOARCH_ppc64le
	MOVD $·rcon(SB), R11 // PTR point to rcon addr
	LVX  (R11), PERMV    // Load LE to BE permute.
#endif

	// Assume len > 0 && len % blockSize == 0.
	CMPW ENC, $0
	P8_LXVB16X(IVP, R0, IVEC)
	CMPU ROUNDS, $10, CR1
	CMPU ROUNDS, $12, CR2 // Only sizes 10/12/14 are supported.

	// Setup key in VSRs, and set loop count in CTR.
	LOAD_KEY(KEYP)
	SRD $4, LEN
	MOVD LEN, CTR

	BEQ Lcbc_dec

	PCALIGN $32
Lcbc_enc:
	P8_LXVB16X(INP, R0, INOUT)
	ADD  $16, INP
	VXOR INOUT, IVEC, INOUT
	CIPHER_BLOCK(INOUT, INOUT, VCIPHER, VCIPHERLAST, Lcbc_enc10, Lcbc_enc12)
	VOR  INOUT, INOUT, IVEC // ciphertext (INOUT) is IVEC for next block.
	P8_STXVB16X(INOUT, OUTP, R0)
	ADD  $16, OUTP
	BC 16, 0, Lcbc_enc // bdnz

	P8_STXVB16X(INOUT, IVP, R0)
	CLEAR_KEYS()
	RET

	PCALIGN $32
Lcbc_dec:
	P8_LXVB16X(INP, R0, TMP)
	ADD  $16, INP
	CIPHER_BLOCK(TMP, INOUT, VNCIPHER, VNCIPHERLAST, Lcbc_dec10, Lcbc_dec12)
	VXOR INOUT, IVEC, INOUT
	VOR  TMP, TMP, IVEC // TMP is IVEC for next block.
	P8_STXVB16X(INOUT, OUTP, R0)
	ADD  $16, OUTP
	BC 16, 0, Lcbc_dec // bdnz

	P8_STXVB16X(IVEC, IVP, R0)
	CLEAR_KEYS()
	RET
