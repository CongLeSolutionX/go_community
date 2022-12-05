// Original source:
//	http://www.zorinaq.com/papers/md5-amd64.html
//	http://www.zorinaq.com/papers/md5-amd64.tar.bz2
//
// Translated from Perl generating GNU assembly into
// #defines generating 6a assembly by the Go Authors.

#include "textflag.h"

// MD5 optimized for AMD64.
//
// Author: Marc Bevand <bevand_m (at) epita.fr>
// Licence: I hereby disclaim the copyright on this code and place it
// in the public domain.

TEXT	·block(SB),NOSPLIT,$8-32
        CMPB     ·useAVX512(SB), $1
       JE avx

	MOVQ	dig+0(FP),	BP
	MOVQ	p+8(FP),	SI
	MOVQ	p_len+16(FP), DX
	SHRQ	$6,		DX
	SHLQ	$6,		DX

	LEAQ	(SI)(DX*1),	DI
	MOVL	(0*4)(BP),	AX
	MOVL	(1*4)(BP),	BX
	MOVL	(2*4)(BP),	CX
	MOVL	(3*4)(BP),	DX

	CMPQ	SI,		DI
	JEQ	end

loop:
	MOVL	AX,		R12
	MOVL	BX,		R13
	MOVL	CX,		R14
	MOVL	DX,		R15

	MOVL	(0*4)(SI),	R8
	MOVL	DX,		R9

#define ROUND1(a, b, c, d, index, const, shift) \
	XORL	c, R9; \
        ADDL $const, a; \
	ANDL	b, R9; \
        ADDL  R8, a; \
	XORL d, R9; \
        MOVL (index*4)(SI), R8; \
	ADDL R9, a; \
	ROLL $shift, a; \
	MOVL c, R9; \
	ADDL b, a

	ROUND1(AX,BX,CX,DX, 1,0xd76aa478, 7);
	ROUND1(DX,AX,BX,CX, 2,0xe8c7b756,12);
	ROUND1(CX,DX,AX,BX, 3,0x242070db,17);
	ROUND1(BX,CX,DX,AX, 4,0xc1bdceee,22);
	ROUND1(AX,BX,CX,DX, 5,0xf57c0faf, 7);
	ROUND1(DX,AX,BX,CX, 6,0x4787c62a,12);
	ROUND1(CX,DX,AX,BX, 7,0xa8304613,17);
	ROUND1(BX,CX,DX,AX, 8,0xfd469501,22);
	ROUND1(AX,BX,CX,DX, 9,0x698098d8, 7);
	ROUND1(DX,AX,BX,CX,10,0x8b44f7af,12);
	ROUND1(CX,DX,AX,BX,11,0xffff5bb1,17);
	ROUND1(BX,CX,DX,AX,12,0x895cd7be,22);
	ROUND1(AX,BX,CX,DX,13,0x6b901122, 7);
	ROUND1(DX,AX,BX,CX,14,0xfd987193,12);
	ROUND1(CX,DX,AX,BX,15,0xa679438e,17);
	ROUND1(BX,CX,DX,AX, 0,0x49b40821,22);

	MOVL	(1*4)(SI),	R8
	MOVL	DX,		R9
	MOVL	DX,		R10

#define ROUND2(a, b, c, d, index, const, shift) \
	NOTL	R9; \
	ADDL $const, a; \
        ANDL    c,   R9; \
        ADDL    R8, a; \
        MOVL    d,   R10; \
        ADDL    R9,  a ; \
	ANDL	b,		R10; \
        ADDL    R10, a; \
	MOVL	(index*4)(SI),R8; \
	MOVL	c,		R9; \
	MOVL	c,		R10; \
	ROLL	$shift,	a; \
	ADDL	b,		a

	ROUND2(AX,BX,CX,DX, 6,0xf61e2562, 5);
	ROUND2(DX,AX,BX,CX,11,0xc040b340, 9);
	ROUND2(CX,DX,AX,BX, 0,0x265e5a51,14);
	ROUND2(BX,CX,DX,AX, 5,0xe9b6c7aa,20);
	ROUND2(AX,BX,CX,DX,10,0xd62f105d, 5);
	ROUND2(DX,AX,BX,CX,15, 0x2441453, 9);
	ROUND2(CX,DX,AX,BX, 4,0xd8a1e681,14);
	ROUND2(BX,CX,DX,AX, 9,0xe7d3fbc8,20);
	ROUND2(AX,BX,CX,DX,14,0x21e1cde6, 5);
	ROUND2(DX,AX,BX,CX, 3,0xc33707d6, 9);
	ROUND2(CX,DX,AX,BX, 8,0xf4d50d87,14);
	ROUND2(BX,CX,DX,AX,13,0x455a14ed,20);
	ROUND2(AX,BX,CX,DX, 2,0xa9e3e905, 5);
	ROUND2(DX,AX,BX,CX, 7,0xfcefa3f8, 9);
	ROUND2(CX,DX,AX,BX,12,0x676f02d9,14);
	ROUND2(BX,CX,DX,AX, 0,0x8d2a4c8a,20);

	MOVL	(5*4)(SI),	R8
	MOVL	CX,		R9

#define ROUND3(a, b, c, d, index, const, shift) \
	ADDL	$const,a; \
        ADDL    R8, a; \
	MOVL	(index*4)(SI),R8; \
	XORL	d,		R9; \
	XORL	b,		R9; \
	ADDL	R9,		a; \
	ROLL	$shift,		a; \
	MOVL	b,		R9; \
	ADDL	b,		a

	ROUND3(AX,BX,CX,DX, 8,0xfffa3942, 4);
	ROUND3(DX,AX,BX,CX,11,0x8771f681,11);
	ROUND3(CX,DX,AX,BX,14,0x6d9d6122,16);
	ROUND3(BX,CX,DX,AX, 1,0xfde5380c,23);
	ROUND3(AX,BX,CX,DX, 4,0xa4beea44, 4);
	ROUND3(DX,AX,BX,CX, 7,0x4bdecfa9,11);
	ROUND3(CX,DX,AX,BX,10,0xf6bb4b60,16);
	ROUND3(BX,CX,DX,AX,13,0xbebfbc70,23);
	ROUND3(AX,BX,CX,DX, 0,0x289b7ec6, 4);
	ROUND3(DX,AX,BX,CX, 3,0xeaa127fa,11);
	ROUND3(CX,DX,AX,BX, 6,0xd4ef3085,16);
	ROUND3(BX,CX,DX,AX, 9, 0x4881d05,23);
	ROUND3(AX,BX,CX,DX,12,0xd9d4d039, 4);
	ROUND3(DX,AX,BX,CX,15,0xe6db99e5,11);
	ROUND3(CX,DX,AX,BX, 2,0x1fa27cf8,16);
	ROUND3(BX,CX,DX,AX, 0,0xc4ac5665,23);

	MOVL	(0*4)(SI),	R8
	MOVL	$0xffffffff,	R9
	XORL	DX,		R9

#define ROUND4(a, b, c, d, index, const, shift) \
        ADDL    $const, a; \
        ADDL    R8, a; \
	ORL	b,		R9; \
	XORL	c,		R9; \
	ADDL	R9,		a; \
	MOVL	(index*4)(SI),R8; \
	MOVL	$0xffffffff,	R9; \
	ROLL	$shift,		a; \
	XORL	c,		R9; \
	ADDL	b,		a

	ROUND4(AX,BX,CX,DX, 7,0xf4292244, 6);
	ROUND4(DX,AX,BX,CX,14,0x432aff97,10);
	ROUND4(CX,DX,AX,BX, 5,0xab9423a7,15);
	ROUND4(BX,CX,DX,AX,12,0xfc93a039,21);
	ROUND4(AX,BX,CX,DX, 3,0x655b59c3, 6);
	ROUND4(DX,AX,BX,CX,10,0x8f0ccc92,10);
	ROUND4(CX,DX,AX,BX, 1,0xffeff47d,15);
	ROUND4(BX,CX,DX,AX, 8,0x85845dd1,21);
	ROUND4(AX,BX,CX,DX,15,0x6fa87e4f, 6);
	ROUND4(DX,AX,BX,CX, 6,0xfe2ce6e0,10);
	ROUND4(CX,DX,AX,BX,13,0xa3014314,15);
	ROUND4(BX,CX,DX,AX, 4,0x4e0811a1,21);
	ROUND4(AX,BX,CX,DX,11,0xf7537e82, 6);
	ROUND4(DX,AX,BX,CX, 2,0xbd3af235,10);
	ROUND4(CX,DX,AX,BX, 9,0x2ad7d2bb,15);
	ROUND4(BX,CX,DX,AX, 0,0xeb86d391,21);

	ADDL	R12,	AX
	ADDL	R13,	BX
	ADDL	R14,	CX
	ADDL	R15,	DX

	ADDQ	$64,		SI
	CMPQ	SI,		DI
	JB	loop

end:
	MOVL	AX,		(0*4)(BP)
	MOVL	BX,		(1*4)(BP)
	MOVL	CX,		(2*4)(BP)
	MOVL	DX,		(3*4)(BP)
	RET


avx:
	MOVQ	dig+0(FP),	BP
	MOVQ	p+8(FP),	SI
	MOVQ	p_len+16(FP), DX
	SHRQ	$6,		DX
	SHLQ	$6,		DX
        MOVQ    $K256<>(SB), CX

        VMOVDQA32 (CX), X16
 
 
	LEAQ	(SI)(DX*1),	DI
	MOVL	(0*4)(BP),	X0
	MOVL	(1*4)(BP),	X1
	MOVL	(2*4)(BP),	X2
	MOVL	(3*4)(BP),	X3

	CMPQ	SI,		DI
	JEQ	end1

loop1:

	MOVD 	X0,		X12
	MOVD	X1,		X13
	MOVD	X2,		X14
	MOVD	X3,		X15

	MOVL	(0*4)(SI),	X8
	MOVD	X3,		X9

#define ROUND11(a, b, c, d, index, const, shift) \
	VPXORD	X9, c, X9; \
        VPADDD const(CX), a, a; \
        VPTERNLOGD $0x6c, b, d, X9; \
        VPADDD  a, X8, a; \
        MOVD (index*4)(SI), X8; \
	VPADDD a, X9, a; \
	VPROLD $shift, a, a; \
	MOVD c, X9; \
	VPADDD a, b, a

	ROUND11(X0,X1,X2,X3, 1,0, 7);
	ROUND11(X3,X0,X1,X2, 2,4,12);
	ROUND11(X2,X3,X0,X1, 3,8,17);
	ROUND11(X1,X2,X3,X0, 4,12,22);
	ROUND11(X0,X1,X2,X3, 5,16, 7);
	ROUND11(X3,X0,X1,X2, 6,20,12);
	ROUND11(X2,X3,X0,X1, 7,24,17);
	ROUND11(X1,X2,X3,X0, 8,28,22);
	ROUND11(X0,X1,X2,X3, 9,32, 7);
	ROUND11(X3,X0,X1,X2,10,36,12);
	ROUND11(X2,X3,X0,X1,11,40,17);
	ROUND11(X1,X2,X3,X0,12,44,22);
	ROUND11(X0,X1,X2,X3,13,48, 7);
	ROUND11(X3,X0,X1,X2,14,52,12);
	ROUND11(X2,X3,X0,X1,15,56,17);
	ROUND11(X1,X2,X3,X0, 0,60,22);

	MOVD	(1*4)(SI),	X8
	MOVD	X3,		X9
	MOVD	X3,		X10
        MOVD    $0xffffffff,      BX
        MOVD    BX, X4

#define ROUND21(a, b, c, d, index, const, shift) \
        VPANDN X4, X9, X9; \
        VPADDD  const(CX), a, a; \  
        VPANDD    X9, c,   X9; \
        VPADDD    a, X8, a; \
        MOVD    d,   X10; \
        VPTERNLOGD $0xec, b, X9, X10; \
        VPADDD  a,  X10, a; \
	MOVD	(index*4)(SI),X8; \
	MOVD	c,		X9; \
	MOVD	c,		X10; \
	VPROLD	$shift,	a, a; \
	VPADDD	a, b,		a

	ROUND21(X0,X1,X2,X3, 6,64, 5);
        ROUND21(X3,X0,X1,X2,11,68, 9);
        ROUND21(X2,X3,X0,X1, 0,72,14);
        ROUND21(X1,X2,X3,X0, 5,76,20);
        ROUND21(X0,X1,X2,X3,10,80, 5);
        ROUND21(X3,X0,X1,X2,15,84, 9);
        ROUND21(X2,X3,X0,X1, 4,88,14);
        ROUND21(X1,X2,X3,X0, 9,92,20);
        ROUND21(X0,X1,X2,X3,14,96, 5);
        ROUND21(X3,X0,X1,X2, 3,100, 9);
        ROUND21(X2,X3,X0,X1, 8,104,14);
        ROUND21(X1,X2,X3,X0,13,108,20);
        ROUND21(X0,X1,X2,X3, 2,112, 5);
        ROUND21(X3,X0,X1,X2, 7,116, 9);
        ROUND21(X2,X3,X0,X1,12,120,14);
        ROUND21(X1,X2,X3,X0, 0,124,20);

	MOVD	(5*4)(SI),	X8
	MOVD	X2,		X9

#define ROUND31(a, b, c, d, index, const, shift) \
        VPADDD  const(CX),a , a; \
        VPADDD    a,X8, a; \
        VPTERNLOGD $0x96, b,d, X9; \
	MOVD    (index*4)(SI), X8; \
	VPADDD	a, X9,		a; \
	VPROLD	$shift,	a,	a; \
	MOVD	b,		X9; \
	VPADDD	a, b,		a

	ROUND31(X0,X1,X2,X3, 8,128, 4);
	ROUND31(X3,X0,X1,X2,11,132,11);
	ROUND31(X2,X3,X0,X1,14,136,16);
	ROUND31(X1,X2,X3,X0, 1,140,23);
	ROUND31(X0,X1,X2,X3, 4,144, 4);
	ROUND31(X3,X0,X1,X2, 7,148,11);
	ROUND31(X2,X3,X0,X1,10,152,16);
	ROUND31(X1,X2,X3,X0,13,156,23);
	ROUND31(X0,X1,X2,X3, 0,160, 4);
	ROUND31(X3,X0,X1,X2, 3,164,11);
	ROUND31(X2,X3,X0,X1, 6,168,16);
	ROUND31(X1,X2,X3,X0, 9,172,23);
	ROUND31(X0,X1,X2,X3,12,176, 4);
	ROUND31(X3,X0,X1,X2,15,180,11);
	ROUND31(X2,X3,X0,X1, 2,184,16);
	ROUND31(X1,X2,X3,X0, 0,188,23);

	MOVD	(0*4)(SI),	X8
        MOVD    BX, X9
        VPXORD    X3, X9, X9

#define ROUND41(a, b, c, d, index, const, shift) \
        VPADDD    const(CX), a, a; \
        VPADDD    a, X8, a; \
        VPTERNLOGD $0x36, b, c, X9; \
	VPADDD	X9,a,		a; \
	MOVD	(index*4)(SI),X8; \
	VPROLD	$shift,a,	a	; \
	VPXORD	X4, c, 		X9; \
	VPADDD	a, b,		a

	ROUND41(X0,X1,X2,X3, 7,192, 6);
	ROUND41(X3,X0,X1,X2,14,196,10);
	ROUND41(X2,X3,X0,X1, 5,200,15);
	ROUND41(X1,X2,X3,X0,12,204,21);
	ROUND41(X0,X1,X2,X3, 3,208, 6);
	ROUND41(X3,X0,X1,X2,10,212,10);
	ROUND41(X2,X3,X0,X1, 1,216,15);
	ROUND41(X1,X2,X3,X0, 8,220,21);
	ROUND41(X0,X1,X2,X3,15,224, 6);
	ROUND41(X3,X0,X1,X2, 6,228,10);
	ROUND41(X2,X3,X0,X1,13,232,15);
	ROUND41(X1,X2,X3,X0, 4,236,21);
	ROUND41(X0,X1,X2,X3,11,240, 6);
	ROUND41(X3,X0,X1,X2, 2,244,10);
	ROUND41(X2,X3,X0,X1, 9,248,15);
	ROUND41(X1,X2,X3,X0, 0,252,21);

	VPADDD	X0, X12,	X0
	VPADDD	X1, X13,	X1
	VPADDD	X2, X14,	X2
	VPADDD	X3, X15,	X3

	ADDQ	$64,		SI
	CMPQ	SI,		DI
	JB	loop1

end1:
	MOVL	X0,		(0*4)(BP)
	MOVL	X1,		(1*4)(BP)
	MOVL	X2,		(2*4)(BP)
	MOVL	X3,		(3*4)(BP)

	RET


// Round specific constants
DATA K256<>+0x00(SB)/4, $0xd76aa478 // k1
DATA K256<>+0x04(SB)/4, $0xe8c7b756 // k2
DATA K256<>+0x08(SB)/4, $0x242070db // k3
DATA K256<>+0x0c(SB)/4, $0xc1bdceee // k4
DATA K256<>+0x10(SB)/4, $0xf57c0faf // k1
DATA K256<>+0x14(SB)/4, $0x4787c62a // k2
DATA K256<>+0x18(SB)/4, $0xa8304613 // k3
DATA K256<>+0x1c(SB)/4, $0xfd469501 // k4

DATA K256<>+0x20(SB)/4, $0x698098d8 // k5 - k8
DATA K256<>+0x24(SB)/4, $0x8b44f7af
DATA K256<>+0x28(SB)/4, $0xffff5bb1
DATA K256<>+0x2c(SB)/4, $0x895cd7be
DATA K256<>+0x30(SB)/4, $0x6b901122
DATA K256<>+0x34(SB)/4, $0xfd987193
DATA K256<>+0x38(SB)/4, $0xa679438e
DATA K256<>+0x3c(SB)/4, $0x49b40821


DATA K256<>+0x40(SB)/4, $0xf61e2562 // k1
DATA K256<>+0x44(SB)/4, $0xc040b340 // k2
DATA K256<>+0x48(SB)/4, $0x265e5a51 // k3
DATA K256<>+0x4c(SB)/4, $0xe9b6c7aa // k4
DATA K256<>+0x50(SB)/4, $0xd62f105d // k1
DATA K256<>+0x54(SB)/4, $0x02441453 // k2
DATA K256<>+0x58(SB)/4, $0xd8a1e681 // k3
DATA K256<>+0x5c(SB)/4, $0xe7d3fbc8 // k4

DATA K256<>+0x60(SB)/4, $0x21e1cde6 // k5 - k8
DATA K256<>+0x64(SB)/4, $0xc33707d6
DATA K256<>+0x68(SB)/4, $0xf4d50d87
DATA K256<>+0x6c(SB)/4, $0x455a14ed
DATA K256<>+0x70(SB)/4, $0xa9e3e905
DATA K256<>+0x74(SB)/4, $0xfcefa3f8
DATA K256<>+0x78(SB)/4, $0x676f02d9
DATA K256<>+0x7c(SB)/4, $0x8d2a4c8a

DATA K256<>+0x80(SB)/4, $0xfffa3942 // k1
DATA K256<>+0x84(SB)/4, $0x8771f681 // k2
DATA K256<>+0x88(SB)/4, $0x6d9d6122 // k3
DATA K256<>+0x8c(SB)/4, $0xfde5380c // k4
DATA K256<>+0x90(SB)/4, $0xa4beea44 // k1
DATA K256<>+0x94(SB)/4, $0x4bdecfa9 // k2
DATA K256<>+0x98(SB)/4, $0xf6bb4b60 // k3
DATA K256<>+0x9c(SB)/4, $0xbebfbc70 // k4

DATA K256<>+0xa0(SB)/4, $0x289b7ec6 // k5 - k8
DATA K256<>+0xa4(SB)/4, $0xeaa127fa
DATA K256<>+0xa8(SB)/4, $0xd4ef3085
DATA K256<>+0xac(SB)/4, $0x04881d05
DATA K256<>+0xb0(SB)/4, $0xd9d4d039
DATA K256<>+0xb4(SB)/4, $0xe6db99e5
DATA K256<>+0xb8(SB)/4, $0x1fa27cf8
DATA K256<>+0xbc(SB)/4, $0xc4ac5665

DATA K256<>+0xc0(SB)/4, $0xf4292244 // k1
DATA K256<>+0xc4(SB)/4, $0x432aff97 // k2
DATA K256<>+0xc8(SB)/4, $0xab9423a7 // k3
DATA K256<>+0xcc(SB)/4, $0xfc93a039 // k4
DATA K256<>+0xd0(SB)/4, $0x655b59c3 // k1
DATA K256<>+0xd4(SB)/4, $0x8f0ccc92 // k2
DATA K256<>+0xd8(SB)/4, $0xffeff47d // k3
DATA K256<>+0xdc(SB)/4, $0x85845dd1 // k4

DATA K256<>+0xe0(SB)/4, $0x6fa87e4f // k5 - k8
DATA K256<>+0xe4(SB)/4, $0xfe2ce6e0
DATA K256<>+0xe8(SB)/4, $0xa3014314
DATA K256<>+0xec(SB)/4, $0x4e0811a1
DATA K256<>+0xf0(SB)/4, $0xf7537e82
DATA K256<>+0xf4(SB)/4, $0xbd3af235
DATA K256<>+0xf8(SB)/4, $0x2ad7d2bb
DATA K256<>+0xfc(SB)/4, $0xeb86d391



GLOBL K256<>(SB), (NOPTR + RODATA), $256

