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

// For set{En,De}cryptKeyAsm
#define INP     R3
#define BITS    R4
#define OUT     R5
#define PTR     R6
#define CNT     R7
#define ROUNDS  R8
#define OUTDEC  R9
#define OUTDECP R10
#define TEMP    R19
#define ZERO    V0
#define IN0     V1
#define IN1     V2
#define KEY     V3
#define RCON    V4
#define MASK    V5
#define TMP     V6
#define STAGE   V7
#define PERMV   V8
#define TMP2    V9

// For {en,de}cryptBlockAsm
#define BLK_INP    R3
#define BLK_OUT    R4
#define BLK_KEY    R5
#define BLK_ROUNDS R6
#define BLK_IDX    R7

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
	VPERM VT, VT, V8, VT

#define P8_STXVB16X(VS,RA,RB) \
	VPERM VS, VS, V8, TMP2 \
	STXVD2X TMP2, (RA+RB)

#define P8_XXBRD(VA,VT) \
	VPERM VA, VA, V8, VT

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
	LVX (PTR), V8 // Load LE to BE permute.
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

// func encryptBlockAsm(nr int, xk *uint32, dst, src *byte)
TEXT ·encryptBlockAsm(SB), NOSPLIT|NOFRAME, $0
	MOVD nr+0(FP), R6   // Round count/Key size
	MOVD xk+8(FP), R5   // Key pointer
	MOVD dst+16(FP), R3 // Dest pointer
	MOVD src+24(FP), R4 // Src pointer
	MOVD $·rcon(SB), R7 // PTR point to rcon addr
	LVX  (R7), V8       // Load doubleword endian-swap permute

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
	LVX  (R7), V8       // Load LE to BE permute.
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


// Remove defines from above so they can be defined here
#undef INP
#undef OUT
#undef ROUNDS
#undef OUTDEC
#undef OUTDECP
#undef KEY
#undef TMP
#undef STAGE
#undef PERMV
#undef TMP2

// CBC encrypt or decrypt
// R3 src
// R4 dst
// R5 len
// R6 key
// R7 iv
// R8 enc=1 dec=0

#define INP R3
#define OUTP R4
#define LEN R5
#define KEYP R6
#define IVP R7
#define ENC R8
#define ROUNDS R9
#define IDX R10

#define RNDKEY0 V0
#define RNDKEY1 V1
#define INOUT V2
#define TMP V3

#define IVEC V4
#define TMP2 V5

// Do the final two cipher operations, xor, and store the
// result, and branch to LABEL if more work, or save the
// IV.
#define CBC_ENC_TAIL(Va, Vb, LABEL) \
	VCIPHER     INOUT, Va, INOUT \
	VCIPHERLAST INOUT, Vb, INOUT \
	VOR         INOUT, INOUT, IVEC \
	P8_STXVB16X(INOUT, OUTP, R0) \
	ADD         $16, OUTP \
	BGE         LABEL \
	P8_STXVB16X(INOUT, IVP, R0)

// Do the final two cipher/xor/or operations on cbc decrypt.
#define CBC_DEC_TAIL(Va, Vb, LABEL) \
	VNCIPHER     INOUT, Va, INOUT \
	VNCIPHERLAST INOUT, Vb, INOUT \
	VXOR         INOUT, IVEC, INOUT \
	VOR          TMP, TMP, IVEC \
	P8_STXVB16X(INOUT, OUTP, R0) \
	ADD          $16, OUTP \
	BGE          Lcbc_dec \
	P8_STXVB16X(IVEC, IVP, R0)

// Run 10,12,14 vcipher operations on Vinout. The expanded key is held in
// V6-V7,V9-V21 (V18-V19 are only valid for keysize 12, V18-V21 for keysize 14).
// One of CR1EQ-CR3EQ is true for keysize 10,12,14 respectively.
#define CIPHER_BLOCK(Vinout, vcipher, label10, label12, label14) \
	vcipher Vinout, V7, Vinout \
	vcipher Vinout, V9, Vinout \
	vcipher Vinout, V10, Vinout \
	vcipher Vinout, V11, Vinout \
	vcipher Vinout, V12, Vinout \
	vcipher Vinout, V13, Vinout \
	vcipher Vinout, V14, Vinout \
	vcipher Vinout, V15, Vinout \
	BEQ     CR1, label10 \
	vcipher Vinout, V16, Vinout \
	vcipher Vinout, V17, Vinout \
	BEQ     CR2, label12 \
	vcipher Vinout, V18, Vinout \
	vcipher Vinout, V19, Vinout \
	BEQ     CR3, label14 

// Load the crypt key into V6-V7,V9-V21. V18-V21 are only populated for
// keysize 12 or 14.  This also clobbers R12,R14-24,Rtmp.
// CR1EQ-CR2EQ are used to infer key sizes 10/12/14.
#define LOAD_KEY(Rkeyp, Rtmp) \
	MOVD   $16, R12 \
	MOVD   $32, R14 \
	MOVD   $48, R15 \
	MOVD   $64, R16 \
	MOVD   $80, R17 \
	MOVD   $96, R18 \
	MOVD   $112, R19 \
	MOVD   $128, R20 \
	MOVD   $144, R21 \
	MOVD   $160, R22 \
	MOVD   $176, R23 \
	MOVD   $192, R24 \
	ADD    $32, Rkeyp, Rtmp \
	LXVD2X (KEYP)(R0), V6 \
	LXVD2X (KEYP)(R12), V7 \
	LXVD2X (Rtmp)(R0), V9 \
	LXVD2X (Rtmp)(R12), V10 \
	LXVD2X (Rtmp)(R14), V11 \
	LXVD2X (Rtmp)(R15), V12 \
	LXVD2X (Rtmp)(R16), V13 \
	LXVD2X (Rtmp)(R17), V14 \
	LXVD2X (Rtmp)(R18), V15 \
	LXVD2X (Rtmp)(R19), V16 \
	LXVD2X (Rtmp)(R20), V17 \
	BEQ    CR1, Lstart \
	LXVD2X (Rtmp)(R21), V18 \
	LXVD2X (Rtmp)(R22), V19 \
	BEQ    CR2, Lstart \
	LXVD2X (Rtmp)(R23), V20 \
	LXVD2X (Rtmp)(R24), V21 \
	Lstart:

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
	LVX  (R11), V8       // Load LE to BE permute.
#endif

	// Assume len > 0 && len % blockSize == 0.
	CMPW ENC, $0
	P8_LXVB16X(IVP, R0, IVEC)
	CMPU ROUNDS, $10, CR1
	CMPU ROUNDS, $12, CR2
	CMPU ROUNDS, $14, CR3 // This ASM kernel only supports 10/12/14 key sizes (limited by earlier calls)

	LOAD_KEY(KEYP, R11)

	BEQ Lcbc_dec

	PCALIGN $32
Lcbc_enc:
	P8_LXVB16X(INP, R0, INOUT)
	ADD  $16, INP
	ADD  $-16, LEN
	CMPU LEN, $16

	ADD  $32, KEYP, R11
	VXOR INOUT, V6, INOUT
	VXOR INOUT, IVEC, INOUT
	CIPHER_BLOCK(INOUT, VCIPHER, Lcbc_enc10, Lcbc_enc12, Lcbc_enc14)
	MOVD R0, 0(R0) // Keysize != 10,12,14. Segfault.

	PCALIGN $32
Lcbc_enc10:
	CBC_ENC_TAIL(V16, V17, Lcbc_enc)
	RET

	PCALIGN $32
Lcbc_enc12:
	CBC_ENC_TAIL(V18, V19, Lcbc_enc)
	RET

	PCALIGN $32
Lcbc_enc14:
	CBC_ENC_TAIL(V20, V21, Lcbc_enc)
	RET

	PCALIGN $32
Lcbc_dec:
	P8_LXVB16X(INP, R0, TMP)
	ADD  $16, INP
	ADD  $-16, LEN
	CMPU LEN, $16

	VXOR TMP, V6, INOUT
	CIPHER_BLOCK(INOUT, VNCIPHER, Lcbc_dec10, Lcbc_dec12, Lcbc_dec14)
	MOVD R0, 0(R0) // Keysize != 10,12,14. Segfault.

	PCALIGN $32
Lcbc_dec10:
	CBC_DEC_TAIL(V16, V17, Lcbc_dec)
	RET

	PCALIGN $32
Lcbc_dec12:
	CBC_DEC_TAIL(V18, V19, Lcbc_dec)
	RET

	PCALIGN $32
Lcbc_dec14:
	CBC_DEC_TAIL(V20, V21, Lcbc_dec)
	RET
