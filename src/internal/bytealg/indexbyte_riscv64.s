// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·IndexByte<ABIInternal>(SB),NOSPLIT,$0-40
#ifndef GOEXPERIMENT_regabiargs
	MOV	b_base+0(FP), A0
	MOV	b_len+8(FP), A1
	MOVBU	c+24(FP), A2	// byte to find
#endif
	MOV	A0, A4		// store base for later
	ADD	A0, A1		// end
	ADD	$-1, A0

loop:
	ADD	$1, A0
	BEQ	A0, A1, notfound
	MOVBU	(A0), A5
	BNE	A2, A5, loop

	SUB	A4, A0		// remove base
#ifndef GOEXPERIMENT_regabiargs
	MOV	A0, ret+32(FP)
#endif
	RET

notfound:
	MOV	$-1, A0
#ifndef GOEXPERIMENT_regabiargs
	MOV	A0, ret+32(FP)
#endif
	RET

TEXT ·IndexByteString<ABIInternal>(SB),NOSPLIT,$0-32
#ifndef GOEXPERIMENT_regabiargs
	MOV	s_base+0(FP), A0
	MOV	s_len+8(FP), A1
	MOVBU	c+16(FP), A2	// byte to find
#endif
	MOV	A0, A4		// store base for later
	ADD	A0, A1		// end
	ADD	$-1, A0

loop:
	ADD	$1, A0
	BEQ	A0, A1, notfound
	MOVBU	(A0), A5
	BNE	A2, A5, loop

	SUB	A4, A0		// remove base
#ifndef GOEXPERIMENT_regabiargs
	MOV	A0, ret+24(FP)
#endif
	RET

notfound:
	MOV	$-1, A0
#ifndef GOEXPERIMENT_regabiargs
	MOV	A0, ret+24(FP)
#endif
	RET
