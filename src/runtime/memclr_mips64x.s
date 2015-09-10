// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips64 mips64le

#include "textflag.h"

// void runtime·memclr(void*, uintptr)
TEXT runtime·memclr(SB),NOSPLIT,$0-16
	MOVV	ptr+0(FP), R1
	MOVV	n+8(FP), R2
	ADDV	R1, R2
	BEQ	R1, R2, done
	MOVB	R0, (R1)
	ADDV	$1, R1
	JMP	-3(PC)
done:
	RET
