// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func rawVforkSyscall(trap, a1 uintptr) (r1, err uintptr)
TEXT ·rawVforkSyscall(SB),NOSPLIT,$0-32
	MOVD	a1+8(FP), R0
	MOVD	$0, R1
	MOVD	$0, R2
	MOVD	$0, R3
	MOVD	$0, R4
	MOVD	$0, R5
	MOVD	trap+0(FP), R8	// syscall entry
	SVC
	CMN	$4095, R0
	BCC	ok
	MOVD	$-1, R4
	MOVD	R4, r1+16(FP)	// r1
	NEG	R0, R0
	MOVD	R0, err+24(FP)	// errno
	RET
ok:
	MOVD	R0, r1+16(FP)	// r1
	MOVD	ZR, err+24(FP)	// errno
	RET
