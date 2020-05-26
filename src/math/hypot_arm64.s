// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// ARM64 version of hypot.go
// derived from math/hypot_amd64.s

#include "textflag.h"

#define PosInf 0x7FF0000000000000
#define NaN 0x7FF8000000000001

// func Hypot(p, q float64) float64
TEXT ·Hypot(SB),NOSPLIT,$0
        // special cases
        MOVD    p+0(FP), R1
        MOVD    $~(1<<63), R0
        AND     R0, R1, R1  //  p = |p|
        MOVD    q+8(FP), R2
        AND     R0, R2, R2  //  q = |q|
        MOVD    $PosInf, R0
        CMP     R0, R1
        BEQ     isInfOrNaN
        CMP     R0, R2
        BEQ     isInfOrNaN
        // hypot = max * sqrt(1 + (min/max)**2)
        ORR     R1, R2, R4
        CBZ     R4, isZero
        FMOVD   R1, F0
        FMOVD   R2, F1
        FMAXD   F0, F1, F2
        FMIND   F0, F1, F3
        FDIVD   F2, F3
        FMULD   F3, F3
        FMOVD   $1.0, F4
        FADDD   F4, F3
        FSQRTD  F3, F3
        FMULD   F2, F3
        FMOVD   F3, ret+16(FP)
        RET
isInfOrNaN:
        CMP     R0, R1
        BEQ     isInf
        CMP     R0, R2
        BEQ     isInf
        MOVD    $NaN, R0
        MOVD    R0, ret+16(FP) // neturn NaN
        RET
isInf:
        MOVD    R0, ret+16(FP) // return +Inf
        RET
isZero:
        MOVD    $0, R0
        MOVD    R0, ret+16(FP) // return 0
        RET

