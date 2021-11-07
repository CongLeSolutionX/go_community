// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·Count<ABIInternal>(SB),NOSPLIT,$0-40
#ifndef GOEXPERIMENT_regabiargs
	MOV	b_base+0(FP), A0
	MOV	b_len+8(FP), A1
	MOVBU	c+24(FP), A2	// byte to count
#endif
	MOV	ZERO, A4	// count
	ADD	A0, A1		// end

loop:
	BEQ	A0, A1, done
	MOVBU	(A0), A5
	ADD	$1, A0
	BNE	A2, A5, loop
	ADD	$1, A4
	JMP	loop

done:
#ifndef GOEXPERIMENT_regabiargs
	MOV	A4, ret+32(FP)
#else
	MOV	A4, A0
#endif
	RET

TEXT ·CountString<ABIInternal>(SB),NOSPLIT,$0-32
#ifndef GOEXPERIMENT_regabiargs
	MOV	s_base+0(FP), A0
	MOV	s_len+8(FP), A1
	MOVBU	c+16(FP), A2	// byte to count
#endif
	MOV	ZERO, A4	// count
	ADD	A0, A1		// end

loop:
	BEQ	A0, A1, done
	MOVBU	(A0), A5
	ADD	$1, A0
	BNE	A2, A5, loop
	ADD	$1, A4
	JMP	loop

done:
#ifndef GOEXPERIMENT_regabiargs
	MOV	A4, ret+24(FP)
#else
	MOV	A4, A0
#endif
	RET
