// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// RISCV64 version of md5block.go
// derived from crypto/md5/md5block_amd64.s

#include "textflag.h"

TEXT	·block(SB),NOSPLIT,$8-32
	MOV 	dig+8(SP),	X8
	MOV 	p+16(SP),	X10
	MOV 	p_len+24(SP), X28
	SRLI   	$6,	X28,X28
	SLLI	$6,	X28,X28

	ADD 	X10,X28,X11

	MOVW	(0*4)(X8),	X5
	MOVW	(1*4)(X8),	X6
	MOVW	(2*4)(X8),	X7
	MOVW	(3*4)(X8),	X28
	BEQ     X10,X11,end


loop:
	MOVW	X5,	X15
	MOVW	X6,	X16
	MOVW	X7,	X17
	MOVW	X28,X30

	MOVW	(0*4)(X10),	X12
	MOVW	X28,X13

#define ROUND1(a, b, c, d, index, const, shift) \
	XOR 	c, X13,X13; \
	MOV    $const,X29;\
	ADDW    X12,a,a;\
	AND 	b, X13,X13; \
	ADDW    X29,a,a;\
	XOR     d, X13,X13; \
	MOVW    (index*4)(X10), X12; \
	ADD     X13, a,a; \
	SRLIW	$(32-shift),a,X29;\
	SLLIW	$shift,a,a;\
	OR		X29,a,a;\
	MOV      c, X13; \
	ADDW     b, a,a

	ROUND1(X5,X6,X7,X28, 1,0xd76aa478, 7);
	ROUND1(X28,X5,X6,X7, 2,0xe8c7b756,12);
	ROUND1(X7,X28,X5,X6, 3,0x242070db,17);
	ROUND1(X6,X7,X28,X5, 4,0xc1bdceee,22);
	ROUND1(X5,X6,X7,X28, 5,0xf57c0faf, 7);
	ROUND1(X28,X5,X6,X7, 6,0x4787c62a,12);
	ROUND1(X7,X28,X5,X6, 7,0xa8304613,17);
	ROUND1(X6,X7,X28,X5, 8,0xfd469501,22);
	ROUND1(X5,X6,X7,X28, 9,0x698098d8, 7);
	ROUND1(X28,X5,X6,X7,10,0x8b44f7af,12);
	ROUND1(X7,X28,X5,X6,11,0xffff5bb1,17);
	ROUND1(X6,X7,X28,X5,12,0x895cd7be,22);
	ROUND1(X5,X6,X7,X28,13,0x6b901122, 7);
	ROUND1(X28,X5,X6,X7,14,0xfd987193,12);
	ROUND1(X7,X28,X5,X6,15,0xa679438e,17);
	ROUND1(X6,X7,X28,X5, 0,0x49b40821,22);

	MOVW	(1*4)(X10),	X12
	MOV    	X28,	X13
	MOV		X28,	X14

#define ROUND2(a, b, c, d, index, const, shift) \
	NOT 	X13,X13; \
	MOV    $const,X29;\
	ADDW    X12,a,a;\
	AND     b,X14,X14; \
	AND 	c,X13,X13; \
	ADDW    X29,a,a;\
	MOVW	(index*4)(X10),X12; \
	OR  	X13,X14,X14; \
	MOV    	c,X13; \
	ADDW 	X14,a,a; \
	MOV		c,X14; \
	SRLIW	$(32-shift),a,X29;\
	SLLIW	$shift,a,a;\
	OR		X29,a,a;\
	ADDW 	b,	a,a

	ROUND2(X5,X6,X7,X28, 6,0xf61e2562, 5);
	ROUND2(X28,X5,X6,X7,11,0xc040b340, 9);
	ROUND2(X7,X28,X5,X6, 0,0x265e5a51,14);
	ROUND2(X6,X7,X28,X5, 5,0xe9b6c7aa,20);
	ROUND2(X5,X6,X7,X28,10,0xd62f105d, 5);
	ROUND2(X28,X5,X6,X7,15, 0x2441453, 9);
	ROUND2(X7,X28,X5,X6, 4,0xd8a1e681,14);
	ROUND2(X6,X7,X28,X5, 9,0xe7d3fbc8,20);
	ROUND2(X5,X6,X7,X28,14,0x21e1cde6, 5);
	ROUND2(X28,X5,X6,X7, 3,0xc33707d6, 9);
	ROUND2(X7,X28,X5,X6, 8,0xf4d50d87,14);
	ROUND2(X6,X7,X28,X5,13,0x455a14ed,20);
	ROUND2(X5,X6,X7,X28, 2,0xa9e3e905, 5);
	ROUND2(X28,X5,X6,X7, 7,0xfcefa3f8, 9);
	ROUND2(X7,X28,X5,X6,12,0x676f02d9,14);
	ROUND2(X6,X7,X28,X5, 0,0x8d2a4c8a,20);

	MOVW	(5*4)(X10),	X12
	MOV	    X7,		X13

#define ROUND3(a, b, c, d, index, const, shift) \
	MOV    $const,X29;\
	ADDW    X12,a,a;\
	MOVW	(index*4)(X10),X12; \
	XOR 	d,X13,X13; \
	XOR 	b,X13,X13; \
	ADDW    X29,a,a;\
	ADDW 	X13,a,a; \
	SRLIW	$(32-shift),a,X29;\
	SLLIW	$shift,a,a;\
	OR		X29,a,a;\
	MOV	    b,X13; \
	ADDW	b,	a,a

	ROUND3(X5,X6,X7,X28, 8,0xfffa3942, 4);
	ROUND3(X28,X5,X6,X7,11,0x8771f681,11);
	ROUND3(X7,X28,X5,X6,14,0x6d9d6122,16);
	ROUND3(X6,X7,X28,X5, 1,0xfde5380c,23);
	ROUND3(X5,X6,X7,X28, 4,0xa4beea44, 4);
	ROUND3(X28,X5,X6,X7, 7,0x4bdecfa9,11);
	ROUND3(X7,X28,X5,X6,10,0xf6bb4b60,16);
	ROUND3(X6,X7,X28,X5,13,0xbebfbc70,23);
	ROUND3(X5,X6,X7,X28, 0,0x289b7ec6, 4);
	ROUND3(X28,X5,X6,X7, 3,0xeaa127fa,11);
	ROUND3(X7,X28,X5,X6, 6,0xd4ef3085,16);
	ROUND3(X6,X7,X28,X5, 9, 0x4881d05,23);
	ROUND3(X5,X6,X7,X28,12,0xd9d4d039, 4);
	ROUND3(X28,X5,X6,X7,15,0xe6db99e5,11);
	ROUND3(X7,X28,X5,X6, 2,0x1fa27cf8,16);
	ROUND3(X6,X7,X28,X5, 0,0xc4ac5665,23);

	MOVW	(0*4)(X10),	X12
	MOV 	$(0-1),X13
	XOR 	X28,X13,X13

#define ROUND4(a, b, c, d, index, const, shift) \
	MOV    $const,X29;\
	ADDW    X12,a,a;\
	OR  	b,	X13,X13; \
	XOR 	c,	X13,X13; \
	ADDW    X29,a,a;\
	ADDW	X13,a,a; \
	MOVW	(index*4)(X10),X12; \
	MOV 	$(0-1),	X13; \
	SRLIW	$(32-shift),a,X29;\
	SLLIW	$shift,a,a;\
	OR		X29,a,a;\
	XOR 	c,	X13,X13; \
	ADDW	b,a,a

	ROUND4(X5,X6,X7,X28, 7,0xf4292244, 6);
	ROUND4(X28,X5,X6,X7,14,0x432aff97,10);
	ROUND4(X7,X28,X5,X6, 5,0xab9423a7,15);
	ROUND4(X6,X7,X28,X5,12,0xfc93a039,21);
	ROUND4(X5,X6,X7,X28, 3,0x655b59c3, 6);
	ROUND4(X28,X5,X6,X7,10,0x8f0ccc92,10);
	ROUND4(X7,X28,X5,X6, 1,0xffeff47d,15);
	ROUND4(X6,X7,X28,X5, 8,0x85845dd1,21);
	ROUND4(X5,X6,X7,X28,15,0x6fa87e4f, 6);
	ROUND4(X28,X5,X6,X7, 6,0xfe2ce6e0,10);
	ROUND4(X7,X28,X5,X6,13,0xa3014314,15);
	ROUND4(X6,X7,X28,X5, 4,0x4e0811a1,21);
	ROUND4(X5,X6,X7,X28,11,0xf7537e82, 6);
	ROUND4(X28,X5,X6,X7, 2,0xbd3af235,10);
	ROUND4(X7,X28,X5,X6, 9,0x2ad7d2bb,15);
	ROUND4(X6,X7,X28,X5, 0,0xeb86d391,21);

	ADDW	X15,X5,X5
	ADDW	X16,X6,X6
	ADDW	X17,X7,X7
	ADDW	X30,X28,X28

	ADD 	$64,X10
	BLT 	X10,X11,loop

end:
	MOVW	X5,		(0*4)(X8)
	MOVW	X6,		(1*4)(X8)
	MOVW	X7,		(2*4)(X8)
	MOVW	X28,	(3*4)(X8)

	RET

