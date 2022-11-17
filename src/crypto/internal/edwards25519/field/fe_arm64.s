// Copyright (c) 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm64,gc,!purego

#include "textflag.h"

// func feMul(out *Element, a *Element, b *Element)
TEXT ·feMul(SB), NOSPLIT, $0-24
	MOVD	a+8(FP), R1  // a
	MOVD	b+16(FP), R2 // b

	LDP	(R1), (R10, R11) // a0, a1
	LDP	16(R1), (R12, R13) // a2, a3
	MOVD	32(R1), R14 // a4

	LDP	(R2), (R20, R21) // b0, b1
	LDP	16(R2), (R22, R23) // b2, b3
	MOVD	32(R2), R24 // b4

	MOVD	$19, R5
	MUL	R11, R5, R6 // a1_19
	MUL	R12, R5, R7 // a2_19
	MUL	R13, R5, R8 // a3_19
	MUL	R14, R5, R9 // a4_19

	// r0 = a0xb0
	MUL	R10, R20, R1
	UMULH	R10, R20, R2
	// r0 += 19xa1xb4
	MUL	R6, R24, R25
	UMULH	R6, R24, R26
	ADDS	R1, R25, R1
	ADC	R2, R26, R2
	// r0 += 19xa2xb3
	MUL	R7, R23, R25
	UMULH	R7, R23, R26
	ADDS	R1, R25, R1
	ADC	R2, R26, R2
	// r0 += 19xa3xb2
	MUL	R8, R22, R25
	UMULH	R8, R22, R26
	ADDS	R1, R25, R1
	ADC	R2, R26, R2
	// r0 += 19xa4xb1
	MUL	R9, R21, R25
	UMULH	R9, R21, R26
	ADDS	R1, R25, R1 // r0.lo
	ADC	R2, R26, R2 // r0.hi

	// r1 = a0xb1
	MUL	R10, R21, R3
	UMULH	R10, R21, R4
	// r1 += a1xb0
	MUL	R11, R20, R25
	UMULH	R11, R20, R26
	ADDS	R3, R25, R3
	ADC	R4, R26, R4
	// r1 += 19xa2xb4
	MUL	R7, R24, R25
	UMULH	R7, R24, R26
	ADDS	R3, R25, R3
	ADC	R4, R26, R4
	// r1 += 19xa3xb3
	MUL	R8, R23, R25
	UMULH	R8, R23, R26
	ADDS	R3, R25, R3
	ADC	R4, R26, R4
	// r1 += 19xa4xb2
	MUL	R9, R22, R25
	UMULH	R9, R22, R26
	ADDS	R3, R25, R3 // r1.lo
	ADC	R4, R26, R4 // r1.hi

	// r2 = a0xb2
	MUL	R10, R22, R6
	UMULH	R10, R22, R7
	// r2 += a1xb1
	MUL	R11, R21, R25
	UMULH	R11, R21, R26
	ADDS	R6, R25, R6
	ADC	R7, R26, R7
	// r2 += a2xb0
	MUL	R12, R20, R25
	UMULH	R12, R20, R26
	ADDS	R6, R25, R6
	ADC	R7, R26, R7
	// r2 += 19xa3xb4
	MUL	R8, R24, R25
	UMULH	R8, R24, R26
	ADDS	R6, R25, R6
	ADC	R7, R26, R7
	// r2 += 19xa4xb3
	MUL	R9, R23, R25
	UMULH	R9, R23, R26
	ADDS	R6, R25, R6 // r2.lo
	ADC	R7, R26, R7 // r2.hi

	// r3 = a0xb3
	MUL	R10, R23, R15
	UMULH	R10, R23, R16
	// r3 += a1xb2
	MUL	R11, R22, R25
	UMULH	R11, R22, R26
	ADDS	R15, R25, R15
	ADC	R16, R26, R16
	// r3 += a2xb1
	MUL	R12, R21, R25
	UMULH	R12, R21, R26
	ADDS	R15, R25, R15
	ADC	R16, R26, R16
	// r3 += a3xb0
	MUL	R13, R20, R25
	UMULH	R13, R20, R26
	ADDS	R15, R25, R15
	ADC	R16, R26, R16
	// r3 += 19xa4xb4
	MUL	R9, R24, R25
	UMULH	R9, R24, R26
	ADDS	R15, R25, R15 // r3.lo
	ADC	R16, R26, R16 // r3.hi

	// r4 = a0xb4
	MUL	R10, R24, R8
	UMULH	R10, R24, R9
	// r4 += a1xb3
	MUL	R11, R23, R25
	UMULH	R11, R23, R26
	ADDS	R8, R25, R8
	ADC	R9, R26, R9
	// r4 += a2xb2
	MUL	R12, R22, R25
	UMULH	R12, R22, R26
	ADDS	R8, R25, R8
	ADC	R9, R26, R9
	// r4 += a3xb1
	MUL	R13, R21, R25
	UMULH	R13, R21, R26
	ADDS	R8, R25, R8
	ADC	R9, R26, R9
	// r4 += a4xb0
	MUL	R14, R20, R25
	UMULH	R14, R20, R26
	ADDS	R8, R25, R8 // r4.lo
	ADC	R9, R26, R9 // r4.hi

	EXTR $51, R1, R2, R10 // c0
	EXTR $51, R3, R4, R11 // c1
	EXTR $51, R6, R7, R12 // c2
	EXTR $51, R15, R16, R13 // c3
	EXTR $51, R8, R9, R14 // c4

	AND	$0X7FFFFFFFFFFFF, R1
	MADD	R14, R1, R5, R1 // rr0
	AND	$0X7FFFFFFFFFFFF, R3
	ADD	R10, R3, R2 // rr1
	AND	$0X7FFFFFFFFFFFF, R6
	ADD	R11, R6, R3 // rr2
	AND	$0X7FFFFFFFFFFFF, R15
	ADD	R12, R15, R4 // rr3
	AND	$0X7FFFFFFFFFFFF, R8
	ADD	R13, R8, R6 // rr4

	LSR	$51, R6, R14 // c4
	// v.l0 = v.l0&maskLow51Bits + c4*19
	AND	$0X7FFFFFFFFFFFF, R1, R7
	MADD	R14, R7, R5, R10 // v.l0

	// v.l1 = v.l1&maskLow51Bits + c0
	AND	$0X7FFFFFFFFFFFF, R2, R8
	ADD	R1>>51, R8, R11 // v.l1

	// v.l2 = v.l2&maskLow51Bits + c1
	AND	$0X7FFFFFFFFFFFF, R3, R9
	ADD	R2>>51, R9, R12 // v.l2

	// v.l3 = v.l3&maskLow51Bits + c2
	AND	$0X7FFFFFFFFFFFF, R4, R7
	ADD	R3>>51, R7, R13 // v.l3

	// v.l4 = v.l4&maskLow51Bits + c3
	AND	$0X7FFFFFFFFFFFF, R6, R8
	ADD	R4>>51, R8, R14 // v.l3

	// Store output
	MOVD	out+0(FP), R0 // out
	STP	(R10, R11), (R0)
	STP	(R12, R13), 16(R0)
	MOVD	R14, 32(R0)

	RET

// func feSquare(out *Element, a *Element)
TEXT ·feSquare(SB), NOSPLIT, $0-16
	MOVD	a+8(FP), R1  // a

	LDP	(R1), (R10, R11) // l0, l1
	LDP	16(R1), (R12, R13) // l2, l3
	MOVD	32(R1), R14 // l4

	LSL $1, R10, R1 // l0_2
	LSL	$1, R11, R2 // l1_2
	MOVD	$38, R9
	MUL	R9, R11, R3 // l1_38
	MUL	R9, R12, R4 // l2_38
	MUL	R9, R13, R5 // l3_38
	MOVD	$19, R9
	MUL	R9, R13, R6 // l3_19
	MUL	R9, R14, R7 // l4_19

	// r0 = l0xl0
	MUL	R10, R10, R15
	UMULH	R10, R10, R16
	// r0 += 19x2xl1xl4
	MUL	R3, R14, R25
	UMULH	R3, R14, R26
	ADDS	R25, R15
	ADC	R26, R16
	// r0 += 19x2xl2xl3
	MUL	R4, R13, R25
	UMULH	R4, R13, R26
	ADDS	R25, R15 // r0.lo
	ADC	R26, R16 // r0.hi

	// r1 = 2xl0xl1
	MUL	R1, R11, R19
	UMULH	R1, R11, R20
	// r1 += 19x2xl2xl4
	MUL	R4, R14, R25
	UMULH	R4, R14, R26
	ADDS	R25, R19
	ADC	R26, R20
	// r1 += 19xl3xl3
	MUL	R6, R13, R25
	UMULH	R6, R13, R26
	ADDS	R25, R19 // r1.lo
	ADC	R26, R20 // r1.hi

	// r2 = 2xl0xl2
	MUL	R1, R12, R21
	UMULH	R1, R12, R22
	// r2 += l1xl1
	MUL	R11, R11, R25
	UMULH	R11, R11, R26
	ADDS	R25, R21
	ADC	R26, R22
	// r2 += 19x2xl3xl4
	MUL	R5, R14, R25
	UMULH	R5, R14, R26
	ADDS	R25, R21 // r2.lo
	ADC	R26, R22 // r2.hi

	// r3 = 2xl0xl3
	MUL	R1, R13, R23
	UMULH	R1, R13, R24
	// r3 += 2xl1xl2
	MUL	R2, R12, R25
	UMULH	R2, R12, R26
	ADDS	R25, R23
	ADC	R26, R24
	// r3 += 19xl4xl4
	MUL	R7, R14, R25
	UMULH	R7, R14, R26
	ADDS	R25, R23 // r3.lo
	ADC	R26, R24 // r3.hi

	// r4 = 2xl0xl4
	MUL	R1, R14, R25
	UMULH	R1, R14, R26
	// r4 += 2xl1xl3
	MUL	R2, R13, R3
	UMULH	R2, R13, R4
	ADDS	R3, R25
	ADC	R4, R26
	// r4 += l2xl2
	MUL	R12, R12, R3
	UMULH	R12, R12, R4
	ADDS	R3, R25 // r4.lo
	ADC	R4, R26 // r4.hi

	EXTR $51, R15, R16, R10 // c0
	EXTR $51, R19, R20, R11 // c1
	EXTR $51, R21, R22, R12 // c2
	EXTR $51, R23, R24, R13 // c3
	EXTR $51, R25, R26, R14 // c4

	AND	$0X7FFFFFFFFFFFF, R15
	MADD	R14, R15, R9, R1 // rr0
	AND	$0X7FFFFFFFFFFFF, R19
	ADD	R10, R19, R2 // rr1
	AND	$0X7FFFFFFFFFFFF, R21
	ADD	R11, R21, R3 // rr2
	AND	$0X7FFFFFFFFFFFF, R23
	ADD	R12, R23, R4 // rr3
	AND	$0X7FFFFFFFFFFFF, R25
	ADD	R13, R25, R5 // rr4

	LSR	$51, R5, R14 // c4
	// v.l0 = v.l0&maskLow51Bits + c4*19
	AND	$0X7FFFFFFFFFFFF, R1, R7
	MADD	R14, R7, R9, R10 // v.l0

	// v.l1 = v.l1&maskLow51Bits + c0
	AND	$0X7FFFFFFFFFFFF, R2, R8
	ADD	R1>>51, R8, R11 // v.l1

	// v.l2 = v.l2&maskLow51Bits + c1
	AND	$0X7FFFFFFFFFFFF, R3, R9
	ADD	R2>>51, R9, R12 // v.l2

	// v.l3 = v.l3&maskLow51Bits + c2
	AND	$0X7FFFFFFFFFFFF, R4, R7
	ADD	R3>>51, R7, R13 // v.l3

	// v.l4 = v.l4&maskLow51Bits + c3
	AND	$0X7FFFFFFFFFFFF, R5, R8
	ADD	R4>>51, R8, R14 // v.l3

	// Store output
	MOVD	out+0(FP), R0 // out
	STP	(R10, R11), (R0)
	STP	(R12, R13), 16(R0)
	MOVD	R14, 32(R0)

	RET
