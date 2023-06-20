// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define PosInf 0x7FF0000000000000

#define ROUNDFN(NAME, MODE) 	\
TEXT NAME(SB),NOSPLIT,$0; 	\
	MOVD	x+0(FP), F0; 	\
	FEQD	F0, F0, T1;	\
	BNEZ	T1, 4(PC);	\
	FADDD	F0, F0, F0;	\
	MOVD	F0, ret+8(FP); 	\
	RET;			\
	MOV	$PosInf, T1;	\
	FMVDX	T1, F1;		\
	FABSD	F0, F2;		\
	FLTD	F1, F2, T1;	\
	/* zero ret */		\
	BEQZ	T1, 3(PC);	\
	FCVTLD.MODE	F0, T1;	\
	FCVTDL	T1, F1;		\
	FSGNJD	F0, F1, F0;	\
	/* ret */		\
	MOVD	F0, ret+8(FP); 	\
	RET

// func archFloor(x float64) float64
ROUNDFN(·archFloor, RDN)

// func archCeil(x float64) float64
ROUNDFN(·archCeil, RUP)

// func archTrunc(x float64) float64
ROUNDFN(·archTrunc, RTZ)
