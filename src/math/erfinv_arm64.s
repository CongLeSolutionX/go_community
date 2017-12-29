// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

	// Coefficients for approximation to erf in |x| <= 0.85
DATA ·erfinvdataab<> + 0(SB)/8, $8.8709406962545514830200e2	// a7
DATA ·erfinvdataab<> + 8(SB)/8, $5.2264952788528545610e3	// b7
DATA ·erfinvdataab<> + 16(SB)/8, $1.1819493347062294404278e4	// a6
DATA ·erfinvdataab<> + 24(SB)/8, $2.8729085735721942674e4	// b6
DATA ·erfinvdataab<> + 32(SB)/8, $2.3782041382114385731252e4	// a5
DATA ·erfinvdataab<> + 40(SB)/8, $3.9307895800092710610e4	// b5
DATA ·erfinvdataab<> + 48(SB)/8, $1.6235862515167575384252e4	// a4
DATA ·erfinvdataab<> + 56(SB)/8, $2.1213794301586595867e4	// b4
DATA ·erfinvdataab<> + 64(SB)/8, $4.8548868893843886794648e3	// a3
DATA ·erfinvdataab<> + 72(SB)/8, $5.3941960214247511077e3	// b3
DATA ·erfinvdataab<> + 80(SB)/8, $6.9706266534389598238465e2	// a2
DATA ·erfinvdataab<> + 88(SB)/8, $6.8718700749205790830e2	// b2
DATA ·erfinvdataab<> + 96(SB)/8, $4.7072688112383978012285e1	// a1
DATA ·erfinvdataab<> + 104(SB)/8, $4.2313330701600911252e1	// b1
DATA ·erfinvdataab<> + 112(SB)/8, $1.1975323115670912564578e0	// a0
DATA ·erfinvdataab<> + 120(SB)/8, $1.0000000000000000000e0	// b0
GLOBL ·erfinvdataab<> + 0(SB), RODATA, $128

	// Coefficients for approximation to erf in 0.85 < |x| <= 1-2*exp(-25)
DATA ·erfinvdatacd<> + 0(SB)/8, $7.74545014278341407640e-4		// c7
DATA ·erfinvdatacd<> + 8(SB)/8, $1.4859850019840355905497876e-9	// d7
DATA ·erfinvdatacd<> + 16(SB)/8, $2.27238449892691845833e-2		// c6
DATA ·erfinvdatacd<> + 24(SB)/8, $7.7441459065157709165577218e-4	// d6
DATA ·erfinvdatacd<> + 32(SB)/8, $2.41780725177450611770e-1		// c5
DATA ·erfinvdatacd<> + 40(SB)/8, $2.1494160384252876777097297e-2	// d5
DATA ·erfinvdatacd<> + 48(SB)/8, $1.27045825245236838258e0		// c4
DATA ·erfinvdatacd<> + 56(SB)/8, $2.0945065210512749128288442e-1	// d4
DATA ·erfinvdatacd<> + 64(SB)/8, $3.64784832476320460504e0		// c3
DATA ·erfinvdatacd<> + 72(SB)/8, $9.7547832001787427186894837e-1	// d3
DATA ·erfinvdatacd<> + 80(SB)/8, $5.76949722146069140550e0		// c2
DATA ·erfinvdatacd<> + 88(SB)/8, $2.3707661626024532365971225e0	// d2
DATA ·erfinvdatacd<> + 96(SB)/8, $4.63033784615654529590e0		// c1
DATA ·erfinvdatacd<> + 104(SB)/8, $2.9036514445419946173133295e0	// d1
DATA ·erfinvdatacd<> + 112(SB)/8, $1.42343711074968357734e0		// c0
DATA ·erfinvdatacd<> + 120(SB)/8, $1.4142135623730950488016887e0	// d0
GLOBL ·erfinvdatacd<> + 0(SB), RODATA, $128

	// Coefficients for approximation to erf in 1-2*exp(-25) < |x| < 1
DATA ·erfinvdataef<> + 0(SB)/8, $2.01033439929228813265e-7		// e7
DATA ·erfinvdataef<> + 8(SB)/8, $2.891024605872965461538222e-15	// f7
DATA ·erfinvdataef<> + 16(SB)/8, $2.71155556874348757815e-5		// e6
DATA ·erfinvdataef<> + 24(SB)/8, $2.010321207683943062279931e-7	// f6
DATA ·erfinvdataef<> + 32(SB)/8, $1.24266094738807843860e-3		// e5
DATA ·erfinvdataef<> + 40(SB)/8, $2.611088405080593625138020e-5	// f5
DATA ·erfinvdataef<> + 48(SB)/8, $2.65321895265761230930e-2		// e4
DATA ·erfinvdataef<> + 56(SB)/8, $1.112800997078859844711555e-3	// f4
DATA ·erfinvdataef<> + 64(SB)/8, $2.96560571828504891230e-1		// e3
DATA ·erfinvdataef<> + 72(SB)/8, $2.103693768272068968719679e-2	// f3
DATA ·erfinvdataef<> + 80(SB)/8, $1.78482653991729133580e0		// e2
DATA ·erfinvdataef<> + 88(SB)/8, $1.936480946950659106176712e-1	// f2
DATA ·erfinvdataef<> + 96(SB)/8, $5.46378491116411436990e0		// e1
DATA ·erfinvdataef<> + 104(SB)/8, $8.482908416595164588112026e-1	// f1
DATA ·erfinvdataef<> + 112(SB)/8, $6.65790464350110377720e0		// e0
DATA ·erfinvdataef<> + 120(SB)/8, $1.414213562373095048801689e0	// f0
GLOBL ·erfinvdataef<> + 0(SB), RODATA, $128

#define ln2	0.693147180559945309417232121458176568075500134360255254120680009 // https://oeis.org/A002162
#define COMPUTEZ \
\	// compute z1 and z2 at the same time, only comment z1
	VMOV	V12.D[0], V12.D[1] \
	VLD1.P	64(R0), [V0.D2, V1.D2, V2.D2, V3.D2] \
	VLD1	(R0), [V4.D2, V5.D2, V6.D2, V7.D2] \
	VFMLA	V0.D2, V12.D2, V1.D2 \	// a7*r+a6
	VFMLA	V1.D2, V12.D2, V2.D2 \	// (a7*r+a6)*r+a5
	VFMLA	V2.D2, V12.D2, V3.D2 \	// ((a7*r+a6)*r+a5)*r+a4
	VFMLA	V3.D2, V12.D2, V4.D2 \	// (((a7*r+a6)*r+a5)*r+a4)*r+a3
	VFMLA	V4.D2, V12.D2, V5.D2 \	// ((((a7*r+a6)*r+a5)*r+a4)*r+a3)*r+a2
	VFMLA	V5.D2, V12.D2, V6.D2 \	// (((((a7*r+a6)*r+a5)*r+a4)*r+a3)*r+a2)*r+a1
	VFMLA	V6.D2, V12.D2, V7.D2 \  // z1=((((((a7*r+a6)*r+a5)*r+a4)*r+a3)*r+a2)*r+a1)*r+a0

// Erfinv returns the inverse error function of x
// func Erfinv(x float64) float64
TEXT ·Erfinv(SB),$24-16
	MOVD	x+0(FP), R1
	AND	$(1<<63), R1, R2	// sign
	AND	$~(1<<63), R1, R3	// abs
	CMP	$0X3FF0000000000000, R3	// 0x3ff0000000000000 = 1.0
	BEQ	eq1	// x = +/-1, return +/-Inf
	BHI	gt1	// x > 1 || x < -1, return NaN

	FMOVD	R3, F10		// F10 = abs(x)
	FMOVD	$0.85, F11
	FMOVD	$0.180625, F12
	FMOVD	$0.25, F13
	FMOVD	$1.0, F14
	FCMPD	F11, F10
	BGT	gt085
	// r := 0.180625 - 0.25*x*x
	FMULD	F10, F13
	FMULD	F10, F13
	FSUBD	F13, F12	// F12 = r
	MOVD	$·erfinvdataab<>+0(SB), R0
	COMPUTEZ	// compute z1 and z2
	// ans = (x * z1) / z2
	VMOV	V7.D[1], V9
	FMULD	F7, F10
	FDIVD	F9, F10
	FMOVD	F10, R4
	ORR	R2, R4
	MOVD	R4, ret+8(FP)
	RET
gt085:
	// r := Sqrt(Ln2 - Log(1.0-x))
	FSUBD	F10, F14
	MOVD	R2, sign-8(SP)	// save sign
	FMOVD	F14, f-24(SP)
	BL	·Log(SB)
	FMOVD	ret-16(SP), F17
	MOVD	sign-8(SP), R2	// restore sign
	FMOVD	$ln2, F18
	FSUBD	F17, F18
	FSQRTD	F18, F12
	FMOVD	$5.0, F15	// 5.0
	FMOVD	$1.6, F16	// 1.6
	FCMPD	F15, F12
	BGT	gt5
	FSUBD	F16, F12	// r -= 1.6
	MOVD	$·erfinvdatacd<>+0(SB), R0
	B	computeZ
gt5:
	FSUBD	F15, F12	// r -= 5.0
	MOVD	$·erfinvdataef<>+0(SB), R0
computeZ:
	COMPUTEZ	// compute z1 and z2
	// ans = z1 / z2
	VMOV	V7.D[1], V9
	FDIVD	F9, F7
	FMOVD	F7, R4
	ORR	R2, R4
	MOVD	R4, ret+8(FP)
	RET
gt1:
	ADD	$1, R2	// for NaN
eq1:
	ORR	$0X7FF0000000000000, R2
	MOVD	R2, ret+8(FP)
	RET
