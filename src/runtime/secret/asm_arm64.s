// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These functions are for testing only.
// (Assembly in _arm64_test.s files doesn't work.)

TEXT ·loadRegisters(SB),0,$0-8
	MOVD	p+0(FP), R0

	MOVD	(R0), R10
	MOVD	(R0), R11
	MOVD	(R0), R12
	MOVD	(R0), R13

	FMOVD	(R0), F15
	FMOVD	(R0), F16
	FMOVD	(R0), F17
	FMOVD	(R0), F18

	VLD1	(R0), [V20.B16]
	VLD1	(R0), [V21.H8]
	VLD1	(R0), [V22.S4]
	VLD1	(R0), [V23.D2]

	RET

TEXT ·spillRegisters(SB),0,$0-16
	MOVD	p+0(FP), R0
	MOVD	R0, R1

	MOVD	R10, (R0)
	MOVD	R11, 8(R0)
	MOVD	R12, 16(R0)
	MOVD	R13, 24(R0)
	ADD	$32, R0

	FMOVD	F15, (R0)
	FMOVD	F16, 16(R0)
	FMOVD	F17, 32(R0)
	FMOVD	F18, 64(R0)
	ADD	$64, R0

	VST1.P	[V20.B16], (R0)
	VST1.P	[V21.H8], (R0)
	VST1.P	[V22.S4], (R0)
	VST1.P	[V23.D2], (R0)

	SUB	R1, R0, R0
	MOVD	R0, ret+8(FP)
	RET
