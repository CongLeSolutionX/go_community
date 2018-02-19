// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !nacl

#include "textflag.h"

// func xorKeyStream(dst, src *byte, n int, state *[256]byte, i, j *uint8)
TEXT ·xorKeyStream(SB),NOSPLIT,$0
	MOVD   dst+0(FP), R3    // dst addr
	MOVD   src+8(FP), R4    // src addr
	MOVD   n+16(FP), R5     //  n
	MOVD   state+24(FP), R6	// state addr
	MOVD   i+32(FP), R7     // i addr
	MOVD   j+40(FP), R8     // j addr
	MOVBZ  (R7), R9         // local i
	MOVBZ  (R8), R10        // local j

	MOVD    R5,CTR
	CMP     R5,R0           // skip if <= 0
	BLE     done
loop:
	MOVBZ   (R4),R5         // src v
	ADD     $1,R9           // i += 1
	ANDCC   $0xff,R9        // uint8
	SLD     $2,R9,R14
	ADD     R6,R14,R14
	MOVWZ   (R14),R15       // c.s[i]
	ADD     R15,R10,R10	// j += c.s[i]
	ANDCC   $0xff,R10       // trunc to byte
	SLD     $2,R10,R16
	ADD     R6,R16,R16      // &c.s[j]
	MOVWZ   (R16),R17       // c.s[j]
	MOVW    R15,(R16)       // swap
	ADD     R15,R17,R18     // c.s[i] + c.s[j]
	ANDCC   $0xff,R18       // trunc to byte
	SLD     $2,R18,R18
	ADD     R18,R6,R18      // &c.s[c.s[i]+c.s[j]]
	MOVW    R17,(R14)
	MOVWZ   (R18),R16       // get value
	XOR     R5,R16,R5       // v ^ uint8(...)
	MOVB    R5,(R3)         // store to dst
	ADD     $1,R3           // next dst uint32
	ADD     $1,R4           // next src uint32
	BC      16, 0, loop     // iterate bndz

done:
	MOVW    R9,(R7)         // c.i = i
	MOVW    R10,(R8)        // c.j = j
	RET
