// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define PosInf 0x7FF0000000000000

#define ROUNDFN(NAME, MODE) 	\
TEXT NAME(SB),NOSPLIT,$0; 	\
	MOVD	x+0(FP), F0; 	\
	FRFLAGS	X5;		\
	FEQD	F0, F0, X6;	\
	BNEZ	X6, 5(PC);	\
	FADDD	F0, F0, F0;	\
	MOVD	F0, ret+8(FP); 	\
	FSFLAGS	X5;		\
	RET;			\
	MOV	$PosInf, X6;	\
	FMVDX	X6, F1;		\
	FABSD	F0, F2;		\
	FLTD	F1, F2, X6;	\
	/* zero ret */		\
	BEQZ	X6, 4(PC);	\
	FCVTLD.MODE	F0, X6;	\
	FCVTDL	X6, F1;		\
	FSGNJD	F0, F1, F0;	\
	/* ret */		\
	MOVD	F0, ret+8(FP); 	\
	FSFLAGS	X5;		\
	RET

// func archFloor(x float64) float64
ROUNDFN(·archFloor, RDN)

// func archCeil(x float64) float64
ROUNDFN(·archCeil, RUP)

// func archTrunc(x float64) float64
ROUNDFN(·archTrunc, RTZ)
