// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This input was created by taking the instruction productions in
// the old assembler's (6a's) grammar and hand-writing complete
// instructions for each rule, to guarantee we cover the same space.

TEXT	foo(SB), 7, $0

// LTYPE1 nonrem	{ outcode($1, &$2); }
	NEGQ	R11
	NEGQ	4(R11)
	NEGQ	foo+4(SB)

// LTYPE2 rimnon	{ outcode($1, &$2); }
	INT	$4
	DIVB	R11
	DIVB	4(R11)
	DIVB	foo+4(SB)

// LTYPE3 rimrem	{ outcode($1, &$2); }
	SUBQ $4, DI
	SUBQ R11, DI
	SUBQ 4(R11), DI
	SUBQ foo+4(SB), DI
	SUBQ $4, 8(R12)
	SUBQ R11, 8(R12)
	SUBQ R11, foo+4(SB)

// LTYPE4 remrim	{ outcode($1, &$2); }
	CMPB	CX, $4

// LTYPER nonrel	{ outcode($1, &$2); }
label:
	JB	-4(PC)
	JB	label

// LTYPEC spec3	{ outcode($1, &$2); }
	JB	2(PC)
	JMP	-4(PC)
	JB	2(PC)
	JMP	label
	JB	2(PC)
	JMP	foo+4(SB)
	JB	2(PC)
	JMP	bar<>+4(SB)
	JB	2(PC)
	JMP	bar<>+4(SB)(R11*4)
	JB	2(PC)
	JMP	*4(SP)
	JB	2(PC)
	JMP	*(R12)
	JB	2(PC)
	JMP	*(R12*4)
	JB	2(PC)
	JMP	*(R12)(R13*4)
	JB	2(PC)
	JMP	*(AX)
	JB	2(PC)
	JMP	*(SP)
	JB	2(PC)
	JMP	*(AX*4)
	JB	2(PC)
	JMP	*(AX)(AX*4)
	JB	2(PC)
	JMP	4(SP)
	JB	2(PC)
	JMP	(R12)
	JB	2(PC)
	JMP	(R12*4)
	JB	2(PC)
	JMP	(R12)(R13*4)
	JB	2(PC)
	JMP	(AX)
	JB	2(PC)
	JMP	(SP)
	JB	2(PC)
	JMP	(AX*4)
	JB	2(PC)
	JMP	(AX)(AX*4)
	JB	2(PC)
	JMP	R13

// LTYPEN spec4	{ outcode($1, &$2); }
	NOP
	NOP	AX
	NOP	foo+4(SB)

// LTYPES spec5	{ outcode($1, &$2); }
	SHLL	CX, R12
	SHLL	CX, foo+4(SB)
	SHLL	CX, R11:AX // Old syntax, still accepted.

// LTYPEM spec6	{ outcode($1, &$2); }
	MOVL	AX, R11
	MOVL	$4, R11
//	MOVL	AX, 0(AX):DS // no longer works - did it ever?

// LTYPEI spec7	{ outcode($1, &$2); }
	IMULB	DX
	IMULW	DX, BX
	IMULL	R11, R12
	IMULQ	foo+4(SB), R11

// LTYPEXC spec8	{ outcode($1, &$2); }
	CMPPD	X1, X2, 4
	CMPPD	foo+4(SB), X2, 4

// LTYPEX spec9	{ outcode($1, &$2); }
	PINSRW	$4, AX, X2
	PINSRW	$4, foo+4(SB), X2

// LTYPERT spec10	{ outcode($1, &$2); }
	JB	2(PC)
	RETFL	$4

// Was bug: LOOP is a branch instruction.
	JB	2(PC)
loop:
	LOOP	loop

// LTYPE0 nonnon	{ outcode($1, &$2); }
	JB	2(PC)
	RET

	MOVOU	(AX), X0
	LDOU	(AX), X0
	LDOU	(BX), X1
	PSLLO	$7, X0
	PSLLO	$7, X1
	PSRLO	$7, X3
	PSRLO	$4, X3

	PABSB	X3, X2
	PABSD	X3, X2
	PABSW	X3, X2
	PBLENDVB	X3, X2
	PHADDD	X3, X2
	PHADDSW	X3, X2
	PHADDW	X3, X2
	PHMINPOSUW	X3, X2
	PHSUBD	X3, X2
	PHSUBSW	X3, X2
	PHSUBW	X3, X2
	PMADDUBSW	X3, X2
	PMAXSB	X3, X2
	PMAXSD	X3, X2
	PMAXUD	X3, X2
	PMAXUW	X3, X2
	PMINSB	X3, X2
	PMINSD	X3, X2
	PMINUD	X3, X2
	PMINUW	X3, X2
	PMOVSXBD	4(AX), X2
	PMOVSXBD	X3, X2
	PMOVSXBQ	X3, X2
	PMOVSXBW	X3, X2
	PMOVSXDQ	X3, X2
	PMOVSXWD	X3, X2
	PMOVSXWQ	X3, X2
	PMOVZXBD	4(AX), X2
	PMOVZXBD	X3, X2
	PMOVZXBQ	X3, X2
	PMOVZXBW	X3, X2
	PMOVZXDQ	X3, X2
	PMOVZXWD	X3, X2
	PMOVZXWQ	X3, X2
	PMULHRSW	X3, X2
	PMULLD	X3, X2
	PMULO	X3, X2
	PSHUFB	X3, X2
	PSIGNB	X3, X2
	PSIGND	X3, X2
	PSIGNW	X3, X2

// END
	END
