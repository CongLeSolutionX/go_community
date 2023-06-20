// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define PosInf 0x7FF0000000000000

// The rounding mode of RISC-V is different from Go spec.

#define ROUNDFN(NAME, MODE) 	\
TEXT NAME(SB),NOSPLIT,$0; 	\
	MOVD	x+0(FP), F0; 	\
	/* whether x is NaN */; \
	FEQD	F0, F0, T1;	\
	BNEZ	T1, 4(PC);	\
	/* return NaN if x is NaN */; \
	FADDD	F0, F0, F0;	\
	MOVD	F0, ret+8(FP); 	\
	RET;			\
	MOV	$PosInf, T1;	\
	FMVDX	T1, F1;		\
	FABSD	F0, F2;		\
	/* if abs(x) > +Inf, return Inf instead of round(x) */; \
	FLTD	F1, F2, T1;	\
	 /* Inf should keep same signed with x then return */;	\
	BEQZ	T1, 3(PC); \
	FCVTLD.MODE	F0, T1;	\
	FCVTDL	T1, F1;		\
	/* rounding will drop signed bit in RISCV, restore it */; \
	FSGNJD	F0, F1, F0;	\
	MOVD	F0, ret+8(FP); 	\
	RET

// func archFloor(x float64) float64
ROUNDFN(·archFloor, RDN)

// func archCeil(x float64) float64
ROUNDFN(·archCeil, RUP)

// func archTrunc(x float64) float64
ROUNDFN(·archTrunc, RTZ)
