// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips64 mips64le

#include "textflag.h"

// void runtime·memmove(void*, void*, uintptr)
TEXT runtime·memmove(SB), NOSPLIT, $-8-24
	MOVV	to+0(FP), R1
	MOVV	from+8(FP), R2
	MOVV	n+16(FP), R3
	BNE	R3, check
	RET

check:
	SGTU	R1, R2, R4
	BNE	R4, backward

	ADDV	R1, R3
loop:
	MOVB	(R2), R4
	ADDV	$1, R2
	MOVB	R4, (R1)
	ADDV	$1, R1
	BNE	R1, R3, loop
	RET

backward:
	ADDV	R3, R2
	ADDV	R1, R3
loop1:
	ADDV	$-1, R2
	MOVB	(R2), R4
	ADDV	$-1, R3
	MOVB	R4, (R3)
	BNE	R1, R3, loop1
	RET
