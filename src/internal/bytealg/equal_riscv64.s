// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define	CTXT	X26

// func memequal(a, b unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-25
#ifndef GOEXPERIMENT_regabiargs
	MOV	a+0(FP), A0
	MOV	b+8(FP), A1
	BEQ	A0, A1, eq
	MOV	size+16(FP), A2
#else
	BEQ	A0, A1, eq
#endif
	ADD	A0, A2, A4
loop:
	BEQ	A0, A4, eq

	MOVBU	(A0), A6
	ADD	$1, A0
	MOVBU	(A1), A7
	ADD	$1, A1
	BEQ	A6, A7, loop

#ifndef GOEXPERIMENT_regabiargs
	MOVB	ZERO, ret+24(FP)
#else
	MOVB	ZERO, A0
#endif
	RET
eq:
	MOV	$1, A0
#ifndef GOEXPERIMENT_regabiargs
	MOVB	A0, ret+24(FP)
#endif
	RET

// func memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen<ABIInternal>(SB),NOSPLIT,$40-17
#ifndef GOEXPERIMENT_regabiargs
	MOV	a+0(FP), A0
	MOV	b+8(FP), A1
	BEQ	A0, A1, eq
	MOV	8(CTXT), A2    // compiler stores size at offset 8 in the closure
	MOV	A0, 8(X2)
	MOV	A1, 16(X2)
	MOV	A2, 24(X2)
#else
	BEQ	A0, A1, eq
	MOV	8(CTXT), A2
#endif
	CALL	runtime·memequal(SB)
#ifndef GOEXPERIMENT_regabiargs
	MOVBU	32(X2), A0
	MOVB	A0, ret+16(FP)
#endif
	RET
eq:
	MOV	$1, A0
#ifndef GOEXPERIMENT_regabiargs
	MOVB	A0, ret+16(FP)
#endif
	RET
