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

#include "textflag.h"

// For set{En,De}cryptKeyAsm
#define INP     R3
#define BITS    R4
#define OUT     R5
#define PTR     R6
#define CNT     R7
#define ROUNDS  R8
#define TEMP    R19
#define ZERO    V0
#define IN0     V1
#define IN1     V2
#define KEY     V3
#define RCON    V4
#define MASK    V5
#define TMP     V6
#define STAGE   V7
#define OUTPERM V8
#define OUTMASK V9
#define OUTHEAD V10
#define OUTTAIL V11

// For {en,de}cryptBlockAsm
#define BLK_INP    R3
#define BLK_OUT    R4
#define BLK_KEY    R5
#define BLK_ROUNDS R6
#define BLK_IDX    R7

DATA ·rcon+0x00(SB)/8, $0x0100000001000000 // RCON
DATA ·rcon+0x08(SB)/8, $0x0100000001000000 // RCON
DATA ·rcon+0x10(SB)/8, $0x1b0000001b000000
DATA ·rcon+0x18(SB)/8, $0x1b0000001b000000
DATA ·rcon+0x20(SB)/8, $0x0d0e0f0c0d0e0f0c // MASK
DATA ·rcon+0x28(SB)/8, $0x0d0e0f0c0d0e0f0c // MASK
DATA ·rcon+0x30(SB)/8, $0x0000000000000000
DATA ·rcon+0x38(SB)/8, $0x0000000000000000
GLOBL ·rcon(SB), RODATA, $64

// func setEncryptKeyAsm(key *byte, keylen int, enc *uint32) int
TEXT ·setEncryptKeyAsm(SB), NOSPLIT|NOFRAME, $0
	// Load the arguments inside the registers
	MOVD key+0(FP), INP
	MOVD keylen+8(FP), BITS
	MOVD enc+16(FP), OUT
	JMP  ·doEncryptKeyAsm(SB)

// This text is used both setEncryptKeyAsm and setDecryptKeyAsm
TEXT ·doEncryptKeyAsm(SB), NOSPLIT|NOFRAME, $0
	// Do not change R10 since it's storing the LR value in setDecryptKeyAsm

	// Check arguments
	MOVD  $-1, PTR               // li    6,-1       exit code to -1 (255)
	CMPU  INP, $0                // cmpldi r3,0      input key pointer set?
	BC    0x0E, 2, enc_key_abort // beq-  .Lenc_key_abort
	CMPU  OUT, $0                // cmpldi r5,0      output key pointer set?
	BC    0x0E, 2, enc_key_abort // beq-  .Lenc_key_abort
	MOVD  $-2, PTR               // li    6,-2       exit code to -2 (254)
	CMPW  BITS, $128             // cmpwi 4,128      greater or equal to 128
	BC    0x0E, 0, enc_key_abort // blt-  .Lenc_key_abort
	CMPW  BITS, $256             // cmpwi 4,256      lesser or equal to 256
	BC    0x0E, 1, enc_key_abort // bgt-  .Lenc_key_abort
	ANDCC $0x3f, BITS, TEMP      // andi. 0,4,0x3f   multiple of 64
	BC    0x06, 2, enc_key_abort // bne-  .Lenc_key_abort

	MOVD $·rcon(SB), PTR // PTR point to rcon addr

	// Get key from memory and write aligned into VR
	NEG      INP, R9            // neg   9,3        R9 is ~INP + 1
	LVX      (INP)(R0), IN0     // lvx   1,0,3      Load key inside IN0
	ADD      $15, INP, INP      // addi  3,3,15     Add 15B to INP addr
	LVSR     (R9)(R0), KEY      // lvsr  3,0,9
	MOVD     $0x20, R8          // li    8,0x20     R8 = 32
	CMPW     BITS, $192         // cmpwi 4,192      Key size == 192?
	LVX      (INP)(R0), IN1     // lvx   2,0,3
	VSPLTISB $0x0f, MASK        // vspltisb 5,0x0f  0x0f0f0f0f... mask
	LVX      (PTR)(R0), RCON    // lvx   4,0,6      Load first 16 bytes into RCON
	VXOR     KEY, MASK, KEY     // vxor  3,3,5      Adjust for byte swap
	LVX      (PTR)(R8), MASK    // lvx   5,8,6
	ADD      $0x10, PTR, PTR    // addi  6,6,0x10   PTR to next 16 bytes of RCON
	VPERM    IN0, IN1, KEY, IN0 // vperm 1,1,2,3    Align
	MOVD     $8, CNT            // li    7,8        CNT = 8
	VXOR     ZERO, ZERO, ZERO   // vxor  0,0,0      Zero to be zero :)
	MOVD     CNT, CTR           // mtctr 7          Set the counter to 8 (rounds)

	LVSL     (OUT)(R0), OUTPERM              // lvsl  8,0,5
	VSPLTISB $-1, OUTMASK                    // vspltisb      9,-1
	LVX      (OUT)(R0), OUTHEAD              // lvx   10,0,5
	VPERM    OUTMASK, ZERO, OUTPERM, OUTMASK // vperm 9,9,0,8

	BLT loop128      // blt   .Loop128
	ADD $8, INP, INP // addi  3,3,8
	BEQ l192         // beq   .L192
	ADD $8, INP, INP // addi  3,3,8
	JMP l256         // b     .L256

loop128:
	// Key schedule (Round 1 to 8)
	VPERM       IN0, IN0, MASK, KEY              // vperm 3,1,1,5         Rotate-n-splat
	VSLDOI      $12, ZERO, IN0, TMP              // vsldoi 6,0,1,12
	VPERM       IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8    Rotate
	VSEL        OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR         OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VCIPHERLAST KEY, RCON, KEY                   // vcipherlast 3,3,4
	STVX        STAGE, (OUT+R0)                  // stvx 7,0,5        Write to output
	ADD         $16, OUT, OUT                    // addi 5,5,16       Point to the next round

	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VADDUWM RCON, RCON, RCON    // vadduwm 4,4,4
	VXOR    IN0, KEY, IN0       // vxor 1,1,3
	BC      0x10, 0, loop128    // bdnz .Loop128

	LVX (PTR)(R0), RCON // lvx 4,0,6     Last two round keys

	// Key schedule (Round 9)
	VPERM       IN0, IN0, MASK, KEY              // vperm 3,1,1,5   Rotate-n-spat
	VSLDOI      $12, ZERO, IN0, TMP              // vsldoi 6,0,1,12
	VPERM       IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8  Rotate
	VSEL        OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR         OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VCIPHERLAST KEY, RCON, KEY                   // vcipherlast 3,3,4
	STVX        STAGE, (OUT+R0)                  // stvx 7,0,5   Round 9
	ADD         $16, OUT, OUT                    // addi 5,5,16

	// Key schedule (Round 10)
	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0       // vxor 1,1,6
	VADDUWM RCON, RCON, RCON    // vadduwm 4,4,4
	VXOR    IN0, KEY, IN0       // vxor 1,1,3

	VPERM       IN0, IN0, MASK, KEY              // vperm 3,1,1,5   Rotate-n-splat
	VSLDOI      $12, ZERO, IN0, TMP              // vsldoi 6,0,1,12
	VPERM       IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8  Rotate
	VSEL        OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR         OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VCIPHERLAST KEY, RCON, KEY                   // vcipherlast 3,3,4
	STVX        STAGE, (OUT+R0)                  // stvx 7,0,5    Round 10
	ADD         $16, OUT, OUT                    // addi 5,5,16

	// Key schedule (Round 11)
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	VXOR   IN0, KEY, IN0                    // vxor 1,1,3
	VPERM  IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8
	VSEL   OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR    OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	STVX   STAGE, (OUT+R0)                  // stvx 7,0,5  Round 11

	ADD $15, OUT, INP   // addi  3,5,15
	ADD $0x50, OUT, OUT // addi  5,5,0x50

	MOVD $10, ROUNDS // li    8,10
	JMP  done        // b     .Ldone

l192:
	LVX      (INP)(R0), TMP                   // lvx 6,0,3
	MOVD     $4, CNT                          // li 7,4
	VPERM    IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8
	VSEL     OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR      OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	STVX     STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD      $16, OUT, OUT                    // addi 5,5,16
	VPERM    IN1, TMP, KEY, IN1               // vperm 2,2,6,3
	VSPLTISB $8, KEY                          // vspltisb 3,8
	MOVD     CNT, CTR                         // mtctr 7
	VSUBUBM  MASK, KEY, MASK                  // vsububm 5,5,3

loop192:
	VPERM       IN1, IN1, MASK, KEY // vperm 3,2,2,5
	VSLDOI      $12, ZERO, IN0, TMP // vsldoi 6,0,1,12
	VCIPHERLAST KEY, RCON, KEY      // vcipherlast 3,3,4

	VXOR   IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR   IN0, TMP, IN0       // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR   IN0, TMP, IN0       // vxor 1,1,6

	VSLDOI  $8, ZERO, IN1, STAGE  // vsldoi 7,0,2,8
	VSPLTW  $3, IN0, TMP          // vspltw 6,1,3
	VXOR    TMP, IN1, TMP         // vxor 6,6,2
	VSLDOI  $12, ZERO, IN1, IN1   // vsldoi 2,0,2,12
	VADDUWM RCON, RCON, RCON      // vadduwm 4,4,4
	VXOR    IN1, TMP, IN1         // vxor 2,2,6
	VXOR    IN0, KEY, IN0         // vxor 1,1,3
	VXOR    IN1, KEY, IN1         // vxor 2,2,3
	VSLDOI  $8, STAGE, IN0, STAGE // vsldoi 7,7,1,8

	VPERM       IN1, IN1, MASK, KEY              // vperm 3,2,2,5
	VSLDOI      $12, ZERO, IN0, TMP              // vsldoi 6,0,1,12
	VPERM       STAGE, STAGE, OUTPERM, OUTTAIL   // vperm 11,7,7,8
	VSEL        OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR         OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VCIPHERLAST KEY, RCON, KEY                   // vcipherlast 3,3,4
	STVX        STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD         $16, OUT, OUT                    // addi 5,5,16

	VSLDOI $8, IN0, IN1, STAGE              // vsldoi 7,1,2,8
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VPERM  STAGE, STAGE, OUTPERM, OUTTAIL   // vperm 11,7,7,8
	VSEL   OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR    OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VXOR   IN0, TMP, IN0                    // vxor 1,1,6
	STVX   STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD    $16, OUT, OUT                    // addi 5,5,16

	VSPLTW  $3, IN0, TMP                     // vspltw 6,1,3
	VXOR    TMP, IN1, TMP                    // vxor 6,6,2
	VSLDOI  $12, ZERO, IN1, IN1              // vsldoi 2,0,2,12
	VADDUWM RCON, RCON, RCON                 // vadduwm 4,4,4
	VXOR    IN1, TMP, IN1                    // vxor 2,2,6
	VXOR    IN0, KEY, IN0                    // vxor 1,1,3
	VXOR    IN1, KEY, IN1                    // vxor 2,2,3
	VPERM   IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8
	VSEL    OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR     OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	STVX    STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD     $15, OUT, INP                    // addi 3,5,15
	ADD     $16, OUT, OUT                    // addi 5,5,16
	BC      0x10, 0, loop192                 // bdnz .Loop192

	MOVD $12, ROUNDS     // li 8,12
	ADD  $0x20, OUT, OUT // addi 5,5,0x20
	BR   done            // b .Ldone

l256:
	LVX   (INP)(R0), TMP                   // lvx 6,0,3
	MOVD  $7, CNT                          // li 7,7
	MOVD  $14, ROUNDS                      // li 8,14
	VPERM IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8
	VSEL  OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR   OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	STVX  STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD   $16, OUT, OUT                    // addi 5,5,16
	VPERM IN1, TMP, KEY, IN1               // vperm 2,2,6,3
	MOVD  CNT, CTR                         // mtctr 7

loop256:
	VPERM       IN1, IN1, MASK, KEY              // vperm 3,2,2,5
	VSLDOI      $12, ZERO, IN0, TMP              // vsldoi 6,0,1,12
	VPERM       IN1, IN1, OUTPERM, OUTTAIL       // vperm 11,2,2,8
	VSEL        OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR         OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	VCIPHERLAST KEY, RCON, KEY                   // vcipherlast 3,3,4
	STVX        STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD         $16, OUT, OUT                    // addi 5,5,16

	VXOR    IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0                    // vxor 1,1,6
	VSLDOI  $12, ZERO, TMP, TMP              // vsldoi 6,0,6,12
	VXOR    IN0, TMP, IN0                    // vxor 1,1,6
	VADDUWM RCON, RCON, RCON                 // vadduwm 4,4,4
	VXOR    IN0, KEY, IN0                    // vxor 1,1,3
	VPERM   IN0, IN0, OUTPERM, OUTTAIL       // vperm 11,1,1,8
	VSEL    OUTHEAD, OUTTAIL, OUTMASK, STAGE // vsel 7,10,11,9
	VOR     OUTTAIL, OUTTAIL, OUTHEAD        // vor 10,11,11
	STVX    STAGE, (OUT+R0)                  // stvx 7,0,5
	ADD     $15, OUT, INP                    // addi 3,5,15
	ADD     $16, OUT, OUT                    // addi 5,5,16
	BC      0x12, 0, done                    // bdz .Ldone

	VSPLTW $3, IN0, KEY        // vspltw 3,1,3
	VSLDOI $12, ZERO, IN1, TMP // vsldoi 6,0,2,12
	VSBOX  KEY, KEY            // vsbox 3,3

	VXOR   IN1, TMP, IN1       // vxor 2,2,6
	VSLDOI $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR   IN1, TMP, IN1       // vxor 2,2,6
	VSLDOI $12, ZERO, TMP, TMP // vsldoi 6,0,6,12
	VXOR   IN1, TMP, IN1       // vxor 2,2,6

	VXOR IN1, KEY, IN1 // vxor 2,2,3
	JMP  loop256       // b .Loop256

done:
	LVX  (INP)(R0), IN1             // lvx   2,0,3
	VSEL OUTHEAD, IN1, OUTMASK, IN1 // vsel 2,10,2,9
	STVX IN1, (INP+R0)              // stvx  2,0,3
	MOVD $0, PTR                    // li    6,0    set PTR to 0 (exit code 0)
	MOVW ROUNDS, 0(OUT)             // stw   8,0(5)

enc_key_abort:
	MOVD PTR, INP        // mr    3,6    set exit code with PTR value
	MOVD INP, ret+24(FP) // Put return value into the FP
	RET                  // blr

// func setDecryptKeyAsm(key *byte, keylen int, dec *uint32) int
TEXT ·setDecryptKeyAsm(SB), NOSPLIT|NOFRAME, $0
	// Load the arguments inside the registers
	MOVD key+0(FP), INP
	MOVD keylen+8(FP), BITS
	MOVD dec+16(FP), OUT

	MOVD LR, R10              // mflr 10
	CALL ·doEncryptKeyAsm(SB)
	MOVD R10, LR              // mtlr 10

	CMPW INP, $0                // cmpwi 3,0  exit 0 = ok
	BC   0x06, 2, dec_key_abort // bne- .Ldec_key_abort

	// doEncryptKeyAsm set ROUNDS (R8) with the proper value for each mode
	SLW  $4, ROUNDS, CNT    // slwi 7,8,4
	SUB  $240, OUT, INP     // subi 3,5,240
	SRW  $1, ROUNDS, ROUNDS // srwi 8,8,1
	ADD  R7, INP, OUT       // add 5,3,7
	MOVD ROUNDS, CTR        // mtctr 8

	// dec_key will invert the key sequence in order to be used for decrypt
dec_key:
	MOVWZ 0(INP), TEMP     // lwz 0, 0(3)
	MOVWZ 4(INP), R6       // lwz 6, 4(3)
	MOVWZ 8(INP), R7       // lwz 7, 8(3)
	MOVWZ 12(INP), R8      // lwz 8, 12(3)
	ADD   $16, INP, INP    // addi 3,3,16
	MOVWZ 0(OUT), R9       // lwz 9, 0(5)
	MOVWZ 4(OUT), R10      // lwz 10,4(5)
	MOVWZ 8(OUT), R11      // lwz 11,8(5)
	MOVWZ 12(OUT), R12     // lwz 12,12(5)
	MOVW  TEMP, 0(OUT)     // stw 0, 0(5)
	MOVW  R6, 4(OUT)       // stw 6, 4(5)
	MOVW  R7, 8(OUT)       // stw 7, 8(5)
	MOVW  R8, 12(OUT)      // stw 8, 12(5)
	SUB   $16, OUT, OUT    // subi 5,5,16
	MOVW  R9, -16(INP)     // stw 9, -16(3)
	MOVW  R10, -12(INP)    // stw 10,-12(3)
	MOVW  R11, -8(INP)     // stw 11,-8(3)
	MOVW  R12, -4(INP)     // stw 12,-4(3)
	BC    0x10, 0, dec_key // bdnz .Ldeckey

	XOR R3, R3, R3 // xor 3,3,3      Clean R3

dec_key_abort:
	MOVD R3, ret+24(FP) // Put return value into the FP
	RET                 // blr

// func encryptBlockAsm(dst, src *byte, enc *uint32)
TEXT ·encryptBlockAsm(SB), NOSPLIT|NOFRAME, $0
	// Load the arguments inside the registers
	MOVD dst+0(FP), BLK_OUT
	MOVD src+8(FP), BLK_INP
	MOVD enc+16(FP), BLK_KEY

	MOVWZ 240(BLK_KEY), BLK_ROUNDS // lwz 6,240(5)
	MOVD  $15, BLK_IDX             // li 7,15

	LVX      (BLK_INP)(R0), ZERO        // lvx 0,0,3
	NEG      BLK_OUT, R11               // neg 11,4
	LVX      (BLK_INP)(BLK_IDX), IN0    // lvx 1,7,3
	LVSL     (BLK_INP)(R0), IN1         // lvsl 2,0,3
	VSPLTISB $0x0f, RCON                // vspltisb 4,0x0f
	LVSR     (R11)(R0), KEY             // lvsr 3,0,11
	VXOR     IN1, RCON, IN1             // vxor 2,2,4
	MOVD     $16, BLK_IDX               // li 7,16
	VPERM    ZERO, IN0, IN1, ZERO       // vperm 0,0,1,2
	LVX      (BLK_KEY)(R0), IN0         // lvx 1,0,5
	LVSR     (BLK_KEY)(R0), MASK        // lvsr 5,0,5
	SRW      $1, BLK_ROUNDS, BLK_ROUNDS // srwi 6,6,1
	LVX      (BLK_KEY)(BLK_IDX), IN1    // lvx 2,7,5
	ADD      $16, BLK_IDX, BLK_IDX      // addi 7,7,16
	SUB      $1, BLK_ROUNDS, BLK_ROUNDS // subi 6,6,1
	VPERM    IN1, IN0, MASK, IN0        // vperm 1,2,1,5

	VXOR ZERO, IN0, ZERO         // vxor 0,0,1
	LVX  (BLK_KEY)(BLK_IDX), IN0 // lvx 1,7,5
	ADD  $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	MOVD BLK_ROUNDS, CTR         // mtctr 6

loop_enc:
	VPERM   IN0, IN1, MASK, IN1     // vperm 2,1,2,5
	VCIPHER ZERO, IN1, ZERO         // vcipher 0,0,2
	LVX     (BLK_KEY)(BLK_IDX), IN1 // lvx 2,7,5
	ADD     $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	VPERM   IN1, IN0, MASK, IN0     // vperm 1,2,1,5
	VCIPHER ZERO, IN0, ZERO         // vcipher 0,0,1
	LVX     (BLK_KEY)(BLK_IDX), IN0 // lvx 1,7,5
	ADD     $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	BC      0x10, 0, loop_enc       // bdnz .Loop_enc

	VPERM       IN0, IN1, MASK, IN1     // vperm 2,1,2,5
	VCIPHER     ZERO, IN1, ZERO         // vcipher 0,0,2
	LVX         (BLK_KEY)(BLK_IDX), IN1 // lvx 2,7,5
	VPERM       IN1, IN0, MASK, IN0     // vperm 1,2,1,5
	VCIPHERLAST ZERO, IN0, ZERO         // vcipherlast 0,0,1

	VSPLTISB $-1, IN1                 // vspltisb 2,-1
	VXOR     IN0, IN0, IN0            // vxor 1,1,1
	MOVD     $15, BLK_IDX             // li 7,15
	VPERM    IN1, IN0, KEY, IN1       // vperm 2,2,1,3
	VXOR     KEY, RCON, KEY           // vxor 3,3,4
	LVX      (BLK_OUT)(R0), IN0       // lvx 1,0,4
	VPERM    ZERO, ZERO, KEY, ZERO    // vperm 0,0,0,3
	VSEL     IN0, ZERO, IN1, IN0      // vsel 1,1,0,2
	LVX      (BLK_OUT)(BLK_IDX), RCON // lvx 4,7,4
	STVX     IN0, (BLK_OUT+R0)        // stvx 1,0,4
	VSEL     ZERO, RCON, IN1, ZERO    // vsel 0,0,4,2
	STVX     ZERO, (BLK_OUT+BLK_IDX)  // stvx 0,7,4

	RET // blr

// func decryptBlockAsm(dst, src *byte, dec *uint32)
TEXT ·decryptBlockAsm(SB), NOSPLIT|NOFRAME, $0
	// Load the arguments inside the registers
	MOVD dst+0(FP), BLK_OUT
	MOVD src+8(FP), BLK_INP
	MOVD dec+16(FP), BLK_KEY

	MOVWZ 240(BLK_KEY), BLK_ROUNDS // lwz 6,240(5)
	MOVD  $15, BLK_IDX             // li 7,15

	LVX      (BLK_INP)(R0), ZERO        // lvx 0,0,3
	NEG      BLK_OUT, R11               // neg 11,4
	LVX      (BLK_INP)(BLK_IDX), IN0    // lvx 1,7,3
	LVSL     (BLK_INP)(R0), IN1         // lvsl 2,0,3
	VSPLTISB $0x0f, RCON                // vspltisb 4,0x0f
	LVSR     (R11)(R0), KEY             // lvsr 3,0,11
	VXOR     IN1, RCON, IN1             // vxor 2,2,4
	MOVD     $16, BLK_IDX               // li 7,16
	VPERM    ZERO, IN0, IN1, ZERO       // vperm 0,0,1,2
	LVX      (BLK_KEY)(R0), IN0         // lvx 1,0,5
	LVSR     (BLK_KEY)(R0), MASK        // lvsr 5,0,5
	SRW      $1, BLK_ROUNDS, BLK_ROUNDS // srwi 6,6,1
	LVX      (BLK_KEY)(BLK_IDX), IN1    // lvx 2,7,5
	ADD      $16, BLK_IDX, BLK_IDX      // addi 7,7,16
	SUB      $1, BLK_ROUNDS, BLK_ROUNDS // subi 6,6,1
	VPERM    IN1, IN0, MASK, IN0        // vperm 1,2,1,5

	VXOR ZERO, IN0, ZERO         // vxor 0,0,1
	LVX  (BLK_KEY)(BLK_IDX), IN0 // lvx 1,7,5
	ADD  $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	MOVD BLK_ROUNDS, CTR         // mtctr 6

loop_dec:
	VPERM    IN0, IN1, MASK, IN1     // vperm 2,1,2,5
	VNCIPHER ZERO, IN1, ZERO         // vncipher 0,0,2
	LVX      (BLK_KEY)(BLK_IDX), IN1 // lvx 2,7,5
	ADD      $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	VPERM    IN1, IN0, MASK, IN0     // vperm 1,2,1,5
	VNCIPHER ZERO, IN0, ZERO         // vncipher 0,0,1
	LVX      (BLK_KEY)(BLK_IDX), IN0 // lvx 1,7,5
	ADD      $16, BLK_IDX, BLK_IDX   // addi 7,7,16
	BC       0x10, 0, loop_dec       // bdnz .Loop_dec

	VPERM        IN0, IN1, MASK, IN1     // vperm 2,1,2,5
	VNCIPHER     ZERO, IN1, ZERO         // vncipher 0,0,2
	LVX          (BLK_KEY)(BLK_IDX), IN1 // lvx 2,7,5
	VPERM        IN1, IN0, MASK, IN0     // vperm 1,2,1,5
	VNCIPHERLAST ZERO, IN0, ZERO         // vncipherlast 0,0,1

	VSPLTISB $-1, IN1                 // vspltisb 2,-1
	VXOR     IN0, IN0, IN0            // vxor 1,1,1
	MOVD     $15, BLK_IDX             // li 7,15
	VPERM    IN1, IN0, KEY, IN1       // vperm 2,2,1,3
	VXOR     KEY, RCON, KEY           // vxor 3,3,4
	LVX      (BLK_OUT)(R0), IN0       // lvx 1,0,4
	VPERM    ZERO, ZERO, KEY, ZERO    // vperm 0,0,0,3
	VSEL     IN0, ZERO, IN1, IN0      // vsel 1,1,0,2
	LVX      (BLK_OUT)(BLK_IDX), RCON // lvx 4,7,4
	STVX     IN0, (BLK_OUT+R0)        // stvx 1,0,4
	VSEL     ZERO, RCON, IN1, ZERO    // vsel 0,0,4,2
	STVX     ZERO, (BLK_OUT+BLK_IDX)  // stvx 0,7,4

	RET // blr

// CBC encrypt or decrypt
// R3 src
// R4 dst
// R5 len
// R6 key
// R7 iv
// R8 enc=1 dec=0
// Ported from: aes_p8_cbc_encrypt
// Register usage:
// R9: ROUNDS
// R10: Index
// V0: initialized to 0
// V3: initialized to mask
// V4: IV
// V5: SRC
// V6: IV perm mask
// V7: DST
// V10: KEY perm mask

// Vector loads are done using LVX followed by
// a VPERM using a mask which is generated from
// an earlier LVSL or LVSR instruction. The VPERM
// is done to select the correct bytes for the final
// value in case the addresses are unaligned.

// Encryption is done with VCIPHER and VCIPHERLAST
// Decryption is done with VNCIPHER and VNCIPHERLAST

// Encrypt and decypt is done as follows:
// - The initial encrypted/decrypted value is set up
// before the outer loop as an XOR of the SRC with IV.
// - The loop counter for the inner loop is set up
// as ROUNDS/2 since 2 encrypt/decrypt operations
// are done per iteration.
// - The outer loop stores the DST value and checks
// if there are more SRC values to load and process.
TEXT ·cryptBlocksChain(SB), NOSPLIT|NOFRAME, $0
	MOVD src+0(FP), R3
	MOVD dst+8(FP), R4
	MOVD length+16(FP), R5
	MOVD key+24(FP), R6
	MOVD iv+32(FP), R7
	MOVD enc+40(FP), R8

	CMPU     R5, $16    // cmpldi r5,16 check len
	BC       14, 0, LR  // bltlr- exit if len == 0
	CMPW     R8, $0     // cmpwi r8,0 check ENC or DEC
	MOVD     $15, R10   // li r10,15 INDEX
	VXOR     V0, V0, V0 // vxor v0,v0,v0 clear V0
	VSPLTISB $0xf, V3   // vspltisb $0xf,v3 set up mask

	LVX   (R7)(R0), V4   // lvx v4,r0,r7 load IV
	LVSL  (R7)(R0), V6   // lvsl v6,r0,r7 load IV perm mask
	LVX   (R7)(R10), V5  // lvx v5,r10,r7 load IV+15 used by VPERM
	VXOR  V6, V3, V6     // vxor v3, v6, v6 reverse byte order
	VPERM V4, V5, V6, V4 // vperm v4,v4,v5,v6 loaded+permed IV
	NEG   R3, R11        // neg r11,r3 used for LVSR for 2nd vector
	LVSR  (R6)(R0), V10  // lvsr v10,r0,r6 KEY perm mask
	MOVWZ 240(R6), R9    // lwz r9,240(r6) ROUNDS
	LVSR  (R11)(R0), V6  // lvsr v6,r0,r11 perm mask for 2nd
	LVX   (R3)(R0), V5   // lvx v5,r0,r3 load SRC
	ADD   $15, R3        // addi r3,r3,15 SRC+15
	VXOR  V6, V3, V6     // vxor v6, v3, v6 reverse byte order of mask

	LVSL     (R4)(R0), V8   // lvsl v8,r0,r4 DST perm mask
	VSPLTISB $-1, V9        // vspltisb v9,-1 mask of 1s
	LVX      (R4)(R0), V7   // lvx v7,r0,r4 Load initial DST
	VPERM    V9, V0, V8, V9 // vperm v9,v9,v0,v8 Set up mask
	VXOR     V8, V3, V8     // vxor v8, v3, v8 swap byte order of mask
	SRW      $1, R9         // rlwinm r9,r9,31,1,31 ROUNDS/2

	MOVD $16, R10 // li r10,16 Set up index
	ADD  $-1, R9  // addi r9,r9,-1 LAST done specially
	BEQ  Lcbc_dec // beq decrypt code

// Outer cbc enc loop; done once per SRC length
// Save initial SRC in V2
// Load next SRC in V5, update SRC ptr, dec length
// Load KEY, adjust with VPERM, load next KEY, adjust with VPERM
// Inc KEY index; XOR SRC and KEY for encryption VALUE
// Load next KEY; inc INDEX; XOR IV with encryption VALUE

// KEY values loaded in V0, V1; VPERM done on V0
// Initial encryption value to VCIPHER in V2
Lcbc_enc:
	VOR   V5, V5, V2      // vor v2,v5,v5 MOVE previous SRC V2
	LVX   (R3)(R0), V5    // lvx v5,r0,r3 LOAD SRC V5
	ADD   $16, R3         // addi r3,r3,16 POINT SRC to NEXT
	MOVD  R9, CTR         // mtctr r9 SET up ROUNDS in CTR
	ADD   $-16, R5        // addi r5,r5,-16 SUB LEN
	LVX   (R6)(R0), V0    // lvx v0,r0,r6 LOAD KEY
	VPERM V2, V5, V6, V2  // vperm v2,v2,v5,v6 SRC permed
	LVX   (R6)(R10), V1   // lvx v1,r10,r6 LOAD next KEY
	ADD   $16, R10        // addi r10,r10,16 NEXT index
	VPERM V1, V0, V10, V0 // vperm v0,v1,v0,v10 KEY
	VXOR  V2, V0, V2      // vxor v2,v2,v0 SRC xor KEY is ENC VALUE
	LVX   (R6)(R10), V0   // lvx v0,r10,r6 LOAD next KEY
	ADD   $16, R10        // addi r10,r10,16 INC index
	VXOR  V2, V4, V2      // vxor v2,v2,v4 SRC xor IV is ENC VALUE

// Loop counter set to ROUNDS/2 since 2 encryptions per loop
// VPERM KEY in V1; encrypt initial KEY with VALUE
// Load and VPERM next KEY and encrypt with VALUE
// Load next KEY in V1 for next iteration or final encyption
Loop_cbc_enc:
	VPERM   V0, V1, V10, V1     // vperm v1,v1,v0,v10 KEY VPERMed
	VCIPHER V2, V1, V2          // vcipher v2,v2,v1 ENCRYPT KEY and VALUE
	LVX     (R6)(R10), V1       // lvx v1,r10,r6 LOAD next KEY
	ADD     $16, R10            // addi r10,r10,16 inc INDEX
	VPERM   V1, V0, V10, V0     // vperm v0,v0,v1,v10 KEY VPERMed
	VCIPHER V2, V0, V2          // vcipher v2,v2,v0 ENCRYPT
	LVX     (R6)(R10), V0       // lvx v0,r10,r6 next KEY
	ADD     $16, R10            // addi r10,r10,16 inc KEY INDEX
	BC      16, 0, Loop_cbc_enc // bdnz Loop_cbc_enc Loop per ROUNDS/2

// VPERM next to last KEY and encrypt with VALUE
// VCIPHERLAST on last KEY with encrypted VALUE
// VPERM and VSEL value to prepare VALUE for store to DST
// Increment DST pointer to continue
// Check length for more iterations of outer loop
	VPERM       V0, V1, V10, V1 // vperm v1,v1,v0,v10 VPERM next to last KEY
	VCIPHER     V2, V1, V2      // vcipher v2,v2,v1 ENCRYPT into VALUE
	LVX         (R6)(R10), V1   // lvx v1,r10,r6 LOAD last KEY
	MOVD        $16, R10        // li r10,16 SET UP index back to 16
	VPERM       V1, V0, V10, V0 // vperm v0,v0,v1,v10 VPERM last KEY 
	VCIPHERLAST V2, V0, V4      // vcipherlast v4,v2,v0 ENCRYPT last KEY
	CMPU        R5, $16         // cmpldi r5,16 CMP length
	VPERM       V4, V4, V8, V3  // vperm v3,v4,v4,v8 Prepare for DST store
	VSEL        V7, V3, V9, V2  // vsel v2,v7,v3,v9 Adjust value for store
	VOR         V3, V3, V7      // vor v7,v3,v3 Save V3 to V7 
	STVX        V2, (R4)(R0)    // stvx v2,r0,r4 Store VSELed DST
	ADD         $16, R4         // addi r4,r4,16 Pointer to Next DST
	BGE         Lcbc_enc        // bge Lcbc_enc Continue if more to process
	BR          Lcbc_done       // b Lcbc_done All done

// Outer cbc dec loop; done once per SRC length
// Save initial SRC in V2
// Load next SRC in V5, update SRC ptr, dec length
// Load KEY, adjust with VPERM, load next KEY, adjust with VPERM
// Inc KEY index; XOR SRC and KEY for decryption VALUE
// Load next KEY; inc INDEX; XOR IV with decryption VALUE

// KEY values loaded in V0, V1; VPERM done on V0
// Initial decryption value to VCIPHER in V2

Lcbc_dec:
	VOR   V5, V5, V3      // vor v3,v5,v5 Save V5 in V3
	LVX   (R3)(R0), V5    // lvx v5,r0,r3 Load SRC in V3
	ADD   $16, R3         // addi r3,r3,16 Inc SRC pointer
	MOVD  R9, CTR         // mtctr r9 ROUNDS/2 in loop counter
	ADD   $-16, R5        // addi r5,r5,-16 Dec length
	LVX   (R6)(R0), V0    // lvx v0,r0,r6 Load KEY
	VPERM V3, V5, V6, V3  // vperm v3,v3,v5,v6 VPERM SRC
	LVX   (R6)(R10), V1   // lvx v1,r10,r6 Load next KEY
	ADD   $16, R10        // addi r10,r10,16 INC Index
	VPERM V1, V0, V10, V0 // vperm v0,v1,v0,v10 VPERM KEY 
	VXOR  V3, V0, V2      // vxor v2,v3,v0 XOR SRC KEY
	LVX   (R6)(R10), V0   // lvx v0,r10,r6 Load next KEY
	ADD   $16, R10        // addi r10,r10,16 INC Index

// VPERM first KEY
// VNCIPHER KEY with decryption VALUE
// Load next KEY; INC key ptr; VPERM key
// VNCIPHER KEY with decryption VALUE
// Load next KEY; INC key ptr
Loop_cbc_dec:
	VPERM    V0, V1, V10, V1      // vperm v1,v0,v1,v10 VPERM KEY
	VNCIPHER V2, V1, V2           // vncipher v2,v2,v1 Decrypt KEY with value
	LVX      (R6)(R10), V1        // lvx v1,r10,r6 Load next KEY
	ADD      $16, R10             // addi r10,r10,16 Inc KEY index
	VPERM    V1, V0, V10, V0      // vperm v0,v1,v0,v10 VPERM KEY
	VNCIPHER V2, V0, V2           // vncipher v2,v2,v0 Decrypt KEY with value
	LVX      (R6)(R10), V0        // lvx v0,r10,r6 Load next key
	ADD      $16, R10             // addi r10,r10,16 Inc KEY index
	BC       16, LT, Loop_cbc_dec // bdnz

// VPERM final KEY
// Decrypt next to last KEY with VALUE
// Load next KEY; INC key ptr; VPERM KEY
// VNCIPHERLAST last KEY with decryption VALUE
// Check length to continue
// XOR encrypted value; Move 
// Reset index
// 

	VPERM        V0, V1, V10, V1 // vperm v1,v0,v1,v10 VPERM KEY
	VNCIPHER     V2, V1, V2      // vncipher v2,v2,v1 Decrypt KEY with value
	LVX          (R6)(R10), V1   // lvx v1,r10,r6 Load next KEY
	MOVD         $16, R10        // li r10,16 Initialize index to 16
	VPERM        V1, V0, V10, V0 // vperm v0,v1,v0,v10 VPERM KEY
	VNCIPHERLAST V2, V0, V2      // vncipherlast v2,v2,v0 Decrypt last KEY with value
	CMPU         R5, $16         // cmpldi r5,16 Check length
	VXOR         V2, V4, V2      // vxor v2,v2,v4 ?? Not sure why this is needed
	VOR          V3, V3, V4      // vor v4,v3,v3 Save DST value in V4
	VPERM        V2, V2, V8, V3  // vperm v3,v2,v2,v8 VPERM bytes for DST store
	VSEL         V7, V3, V9, V2  // vsel v2,v7,v3,v9  VSEL bytes for result
	VOR          V3, V3, V7      // vor v7,v3,v3 Save VPERMed DST to V7
	STVX         V2, (R4)(R0)    // stvx v2,r0,r4 Store VSELed bytes
	ADD          $16, R4         // addi r4,r4,16 Increment DST address
	BGE          Lcbc_dec        // bge Continue if length > 16

Lcbc_done:
	ADD      $-1, R4        // addi r4,r4,-1 WHY?
	LVX      (R4)(R0), V2   // lvx v2,r0,r4 load DST 
	VSEL     V7, V2, V9, V2 // vsel v2,v7,v2,v9 VSEL DST with??
	STVX     V2, (R4)(R0)   // stvx v2,r0,r4 Store VSELed value into DST
	NEG      R7, R8         // neg r8,r7 NEG R7?
	MOVD     $15, R10       // li r10,15 Initialize IDX
	VXOR     V0, V0, V0     // vxor v0,v0,v0 CLEAR
	VSPLTISB $-1, V9        // vspltisb v9,-1 ONES
	VSPLTISB $0xf, V3       // vspltisb v3, 0xf
	LVSR     (R8)(R0), V8   // lvsl v8,r0,r8 VPERM mask based on R7
	VPERM    V9, V0, V8, V9 // vperm v9,v9,v0,v8 Creating new MASK??
	VXOR     V8, V3, V8     // vxor v9, v3, v9 ??
	LVX      (R7)(R0), V7   // lvx v7,r0,r7 IV
	VPERM    V4, V4, V8, V4 // vperm v4,v4,v4,v8 ??
	VSEL     V7, V4, V9, V2 // vsel v2,v7,v4,v9
	LVX      (R7)(R10), V5  // lvx v5,r10,r7 IV+16
	STVX     V2, (R7)(R0)   // stvx v2,r0,r7   SAVE IV
	VSEL     V4, V5, V9, V2 // vsel v2,v4,v5,v9
	STVX     V2, (R7)(R10)  // stvx v2,r10,r7  SAVE IV+16
	RET                     // bclr 20,lt,0
	WORD $0 // .long
	WORD $0 // .long

