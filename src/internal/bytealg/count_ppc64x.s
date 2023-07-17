// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64le || ppc64

#include "go_asm.h"
#include "textflag.h"

TEXT ·Count<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-40
	// R3 = byte array pointer
	// R4 = length
	// R6 = byte to count
	MTVRD	R6, V1		// move compare byte
	MOVD	R6, R5
	VSPLTB	$7, V1, V1	// replicate byte across V1
	BR	countbytebody<>(SB)

TEXT ·CountString<ABIInternal>(SB), NOSPLIT|NOFRAME, $0-32
	// R3 = byte array pointer
	// R4 = length
	// R5 = byte to count
	MTVRD	R5, V1		// move compare byte
	VSPLTB	$7, V1, V1	// replicate byte across V1
	BR	countbytebody<>(SB)

// R3: addr of string
// R4: len of string
// R5: byte to count
// V1: byte to count, splatted.
// On exit:
// R3: return value
TEXT countbytebody<>(SB), NOSPLIT|NOFRAME, $0-0
	MOVD	$0, R18 // byte count

#ifndef GOPPC64_power10
	RLDIMI	$8, R5, $48, R5
	RLDIMI	$16, R5, $32, R5
	RLDIMI	$32, R5, $0, R5	// fill reg with the byte to count
#endif

	CMPU	R4, $32		// Check if it's a small string (<32 bytes)
	BLT	tail		// Jump to the small string case
	SRD	$5, R4, R20
	MOVD	R20, CTR
	MOVD	$16, R21
	XXLXOR	V4, V4, V4
	XXLXOR	V5, V5, V5

cmploop:
	LXVD2X	(R0)(R3), V0
	LXVD2X	(R21)(R3), V2
	VCMPEQUB V2, V1, V2
	VCMPEQUB V0, V1, V0
	VPOPCNTD V2, V2
	VPOPCNTD V0, V0
	VADDUDM	V0, V4, V4
	VADDUDM	V2, V5, V5
	ADD	$32, R3
	BDNZ	cmploop

	VADDUDM	V4, V5, V5
	MFVSRD	V5, R18
	VSLDOI	$8, V5, V5, V5
	MFVSRD	V5, R21
	ADD	R21, R18, R18
	ANDCC	$31, R4, R4
	// Skip the tail processing if no bytes remaining.
	BEQ	tail_0

#ifdef GOPPC64_power10
	SRD	$3, R18, R18
tail:
	CMP	R4, $16
	BLT	small_tail_p10
	LXV	0(R3), V0
	VCMPEQUB V0, V1, V0
	VCNTMBB	V0, $1, R14
	SRD	$56, R14, R14
	ADD	R14, R18, R18
	ADD	$16, R3, R3
	ANDCC	$15, R4, R4

small_tail_p10:
	SLD	$56, R4, R6
	LXVLL	R3, R6, V0
	VCMPEQUB V0, V1, V0
	VCLRRB	V0, R4, V0	// Zero out non-participating matches.
	VCNTMBB	V0, $1, R14
	SRD	$56, R14, R14
	ADD	R14, R18, R3
	RET

#else
tail:	// Count the last 0 - 31 bytes.
	CMP	R4, $16
	BLT	tail_8
	MOVD	(R3), R12
	MOVD	8(R3), R14
	CMPB	R12, R5, R12
	CMPB	R14, R5, R14
	POPCNTD	R12, R12
	POPCNTD	R14, R14
	ADD	R12, R18, R18
	ADD	R14, R18, R18
	ADD	$16, R3, R3
	ADD	$-16, R4, R4

tail_8:
	CMP	R4, $8
	BLT	tail_4
	MOVD	(R3), R12
	CMPB	R12, R5, R12
	POPCNTD	R12, R12
	ADD	R12, R18, R18
	ADD	$8, R3, R3
	ADD	$-8, R4, R4

tail_4:
	CMP	R4, $4
	BLT	tail_2
	MOVWZ	(R3), R12
	CMPB	R12, R5, R12
	SLD	$32, R12, R12	// Remove non-participating matches.
	POPCNTD	R12, R12
	ADD	R12, R18, R18
	ADD	$4, R3, R3
	ADD	$-4, R4, R4

tail_2:
	CMP	R4, $2
	BLT	tail_1
	MOVHZ	(R3), R12
	CMPB	R12, R5, R12
	SLD	$48, R12, R12	// Remove non-participating matches.
	POPCNTD	R12, R12
	ADD	R12, R18, R18
	ADD	$2, R3, R3
	ADD	$-2, R4, R4

tail_1:
	CMP	R4, $1
	BLT	tail_0
	MOVBZ	(R3), R12
	CMPB	R12, R5, R12
	ANDCC	$0x8, R12, R12
	ADD	R12, R18, R18
#endif

tail_0:
	SRD	$3, R18, R3	// Fixup count, it is off by 8x.
	RET
