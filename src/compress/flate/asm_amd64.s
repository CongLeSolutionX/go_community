// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func haveSSE42() bool
TEXT ·haveSSE42(SB), NOSPLIT, $0
	MOVQ $1, AX
	CPUID
	SHRQ $20, CX
	ANDQ $1, CX
	MOVB CX, ret+0(FP)
	RET

// func crc32SSE(a []byte) uint32
TEXT ·crc32SSE(SB), NOSPLIT, $0
	MOVQ a+0(FP), R10
	MOVQ $0, BX

	// CRC32L (R10), BX
	BYTE $0xF2; BYTE $0x41; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0x1a

	MOVL BX, ret+24(FP)
	RET

// func crc32SSEAll(a []byte, dst []uint32)
TEXT ·crc32SSEAll(SB), NOSPLIT, $0
	MOVQ  a+0(FP), R8      // R8: src
	MOVQ  a_len+8(FP), R10
	MOVQ  dst+24(FP), R9   // R9: dst
	SUBQ  $4, R10          // R10: inputLength; len(a)-4
	JS    end
	JZ    one_crc
	MOVQ  R10, R13
	SHRQ  $2, R10          // inputLength/4
	ANDQ  $3, R13          // inputLength&3
	MOVQ  $0, BX
	ADDQ  $1, R13
	TESTQ R10, R10
	JZ    rem_loop

crc_loop:
	MOVQ (R8), R11
	MOVQ $0, BX
	MOVQ $0, DX
	MOVQ $0, DI
	MOVQ R11, R12
	SHRQ $8, R11
	MOVQ R12, AX
	MOVQ R11, CX
	SHRQ $16, R12
	SHRQ $16, R11
	MOVQ R12, SI

	// CRC32L AX, BX
	BYTE $0xF2; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0xd8

	// CRC32L CX, DX
	BYTE $0xF2; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0xd1

	// CRC32L SI, DI
	BYTE $0xF2; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0xfe
	MOVL BX, (R9)
	MOVL DX, 4(R9)
	MOVL DI, 8(R9)

	MOVQ $0, BX
	MOVL R11, AX

	// CRC32L AX, BX
	BYTE $0xF2; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0xd8
	MOVL BX, 12(R9)

	ADDQ $16, R9
	ADDQ $4, R8
	MOVQ $0, BX
	SUBQ $1, R10
	JNZ  crc_loop

rem_loop:
	MOVL (R8), AX

	// CRC32L AX, BX
	BYTE $0xF2; BYTE $0x0f
	BYTE $0x38; BYTE $0xf1; BYTE $0xd8

	MOVL BX, (R9)
	ADDQ $4, R9
	ADDQ $1, R8
	MOVQ $0, BX
	SUBQ $1, R13
	JNZ  rem_loop

end:
	RET

one_crc:
	MOVQ $1, R13
	MOVQ $0, BX
	JMP  rem_loop

// func matchLenSSE4(a, b []byte, max int) int
TEXT ·matchLenSSE4(SB), NOSPLIT, $0
	MOVQ  a+0(FP), SI        // SI: &a[0]
	MOVQ  b+24(FP), DI       // DI: &b[0]
	MOVQ  $0, R11            // R11: match length, initially zero
	MOVQ  max+48(FP), R10
	MOVQ  R10, R12
	SHRQ  $4, R10            // R10: max/16
	ANDQ  $15, R12           // R12: max&15
	TESTQ R10, R10
	JZ    matchlen_verysmall

	// Check 16 bytes at a time using PCMPESTRI.
	//
	// If mismatch is found CX contains the first mismatched index.
	MOVQ $16, AX // PCMPESTRI 1st input length.
	MOVQ $16, DX // PCMPESTRI 2nd input length.

loopback_matchlen:
	MOVOU (SI), X0 // a[x]
	MOVOU (DI), X1 // b[x]

	// PCMPESTRI $0x18, X1, X0
	// 0x18 = _SIDD_UBYTE_OPS (0x0) | _SIDD_CMP_EQUAL_EACH (0x8) | _SIDD_NEGATIVE_POLARITY (0x10)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a
	BYTE $0x61; BYTE $0xc1; BYTE $0x18

	JC match_ended

	ADDQ $16, SI
	ADDQ $16, DI
	ADDQ $16, R11

	SUBQ $1, R10
	JNZ  loopback_matchlen

	// Check the remainder using REP CMPSB.
matchlen_verysmall:
	TESTQ R12, R12
	JZ    done_matchlen
	MOVQ  R12, CX
	ADDQ  R12, R11

	// Compare CX bytes at [SI] [DI]
	// Subtract one from CX for every match.
	// Terminates when CX is zero (checked pre-compare).
	REP; CMPSB

	// Check if last was a match.
	JZ done_matchlen

	// Subtract remaining bytes.
	SUBQ CX, R11
	SUBQ $1, R11
	MOVQ R11, ret+56(FP)
	RET

match_ended:
	ADDQ CX, R11

done_matchlen:
	MOVQ R11, ret+56(FP)
	RET

// func histogram(b []byte, h []int32)
TEXT ·histogram(SB), NOSPLIT, $0
	MOVQ b+0(FP), SI     // SI: &b[0]
	MOVQ b_len+8(FP), R9 // R9: len(b)
	MOVQ h+24(FP), DI    // DI: Histogram
	MOVQ R9, R8
	SHRQ $3, R8
	JZ   hist1
	MOVQ $0, R11

loop_hist8:
	MOVQ (SI), R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	MOVB R10, R11
	INCL (DI)(R11*4)
	SHRQ $8, R10

	INCL (DI)(R10*4)

	ADDQ $8, SI
	DECQ R8
	JNZ  loop_hist8

hist1:
	ANDQ $7, R9
	JZ   end_hist
	MOVQ $0, R10

loop_hist1:
	MOVB (SI), R10
	INCL (DI)(R10*4)
	INCQ SI
	DECQ R9
	JNZ  loop_hist1

end_hist:
	RET
