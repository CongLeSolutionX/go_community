// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// The method is based on a paper by Naoki Shibata: "Efficient evaluation
// methods of elementary functions suitable for SIMD computation", Proc.
// of International Supercomputing Conference 2010 (ISC'10), pp. 25 -- 32
// (May 2010). The paper is available at
// http://www.springerlink.com/content/340228x165742104/
//
// The original code and the constants below are from the author's
// implementation available at http://freshmeat.net/projects/sleef.
// The README file says, "The software is in public domain.
// You can use the software without any obligation."
//
// This code is a simplified version of the original.

#define PosOne 0x3FF0000000000000
#define PosInf 0x7FF0000000000000
#define NaN    0x7FF8000000000001
#define PI4A 0.7853981554508209228515625 // pi/4 split into three parts
#define PI4B 0.794662735614792836713604629039764404296875e-8
#define PI4C 0.306161699786838294306516483068750264552437361480769e-16
#define M4PI 1.273239544735162542821171882678754627704620361328125 // 4/pi
#define T0 1.0
#define T1 -8.33333333333333333333333e-02 // (-1.0/12)
#define T2 2.77777777777777777777778e-03 // (+1.0/360)
#define T3 -4.96031746031746031746032e-05 // (-1.0/20160)
#define T4 5.51146384479717813051146e-07 // (+1.0/1814400)

// func Sincos(d float64) (sin, cos float64)
TEXT ·Sincos(SB),NOSPLIT,$0
	JMP	·sincos(SB)
