// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func featureCheck() bool
TEXT ·featureCheck(SB),NOSPLIT,$16-1
	LA	tmp-16(SP), R1
	XOR	R0, R0         // query function code is 0
	WORD    $0xB93E0006    // KIMD (R6 is ignored)
	MOVBZ	tmp-16(SP), R4 // get the first byte
	AND	$0x20, R4      // bit 2 (big endian) for SHA256
	CMPBEQ	R4, $0, nosha256
	MOVB	$1, ret+0(FP)
	RET
nosha256:
	MOVB	$0, ret+0(FP)
	RET

// func blockString(dig *digest, s string)
TEXT	·blockString(SB),NOSPLIT|NOFRAME,$0-24
	LMG	dig+0(FP), R1, R3 // R2 = &p[0], R3 = len(p)
	BR	sha256block<>(SB)

// func block(dig *digest, p []byte)
TEXT	·block(SB),NOSPLIT|NOFRAME,$0-32
	LMG	dig+0(FP), R1, R3 // R2 = &p[0], R3 = len(p)
	BR	sha256block<>(SB)

// Message block routine, expects:
//   R1: pointer to digest
//   R2: pointer to input bytes to hash
//   R3: length of input
TEXT	sha256block<>(SB),NOSPLIT|NOFRAME,$0-0
	MOVBZ	·useAsm(SB), R4
	CMPBNE	R4, $1, generic
	MOVBZ	$2, R0        // SHA256 function code
loop:
	WORD	$0xB93E0002   // KIMD R2
	BVS	loop          // continue if interrupted
done:
	XOR	R0, R0        // restore R0
	RET
generic:
	BR	·blockGeneric(SB)
