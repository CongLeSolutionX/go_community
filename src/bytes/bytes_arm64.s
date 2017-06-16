// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// countByte(s []byte, c byte) int
TEXT bytes·countByte(SB),NOSPLIT,$0-40
	MOVD	s_base+0(FP), R0
	MOVD	s_len+8(FP), R2
	MOVBU	c+24(FP), R1
	ADD	$40, RSP, R8
	B	bytes·countByte<>(SB)

// input:
//   R0: data
//   R1: byte to search
//   R2: data len
//   R8: address to put result
TEXT bytes·countByte<>(SB),NOSPLIT,$0
	MOVD	$0, R11
	// short path to handle 0-byte case
	CBZ	R2, done
	CMP	$0x20, R2
	// jump directly to tail if length < 32
	BLO	tail
	ANDS	$0x1f, R0, R9
	BEQ	chunk
	// Work with not 32-byte aligned head
	BIC	$0x1f, R0, R3
	ADD	$0x20, R3
head_loop:
	MOVBU.P	1(R0), R5
	CMP	R5, R1
	CINC	EQ, R11, R11
	SUB	$1, R2, R2
	CBZ	R2, done
	CMP	R0, R3
	BNE	head_loop
	// Work with 32-byte aligned chunks
chunk:
	BIC	$0x1f, R2, R9
	// The first chunk can also be the last
	CBZ	R9, tail
	ADD	R0, R9, R3
	MOVD	$1, R5
	VMOV	R5, V5.B16
	SUB	R9, R2, R2
	VMOV	R1, V0.B16
	// Count the target byte in 32-byte chunk
chunk_loop:
	VLD1.P	(R0), [V1.B16, V2.B16]
	CMP	R0, R3
	VCMEQ	V0.B16, V1.B16, V3.B16
	VCMEQ	V0.B16, V2.B16, V4.B16
	// Clear the higher 7 bits
	VAND	V5.B16, V3.B16, V3.B16
	VAND	V5.B16, V4.B16, V4.B16
	// Count lanes match the requested byte
	VADDP	V4.B16, V3.B16, V6.B16 // 32B->16B
	VUADDLV	V6.B16, V6
	VMOV	V6.B[0], R6
	ADD	R6, R11, R11
	BNE	chunk_loop
	CBZ	R2, done
tail:
	// Work with tail shorter than 32 bytes
	MOVBU.P	1(R0), R5
	SUB	$1, R2, R2
	CMP	R5, R1
	CINC	EQ, R11, R11
	CBNZ	R2, tail
done:
	MOVD	R11, (R8)
	RET
