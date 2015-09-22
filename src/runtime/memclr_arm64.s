// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// void runtime·memclr(void*, uintptr)
TEXT runtime·memclr(SB),NOSPLIT,$0-16
	MOVD	ptr+0(FP), R3
	MOVD	n+8(FP), R4
	AND	$~7, R4, R5	// R5 is N&~7
	SUB	R5, R4, R6	// R6 is N&7

	CMP	$0, R5
	BEQ	nowords

	ADD	R3, R5, R5

	MOVD.P	$0, 8(R3)
	CMP	R3, R5
	BNE	-2(PC)
nowords:
        CMP	$0, R6
        BEQ	done

	ADD	R3, R6, R6

	MOVBU.P	$0, 1(R3)
	CMP	R3, R6
	BNE	-2(PC)
done:
	RET
