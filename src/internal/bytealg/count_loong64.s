// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·Count<ABIInternal>(SB),NOSPLIT,$0-40
	// R4 = b_base
	// R5 = b_len
	// R6 = b_cap (unused)
	// R7 = byte to count (want in R6)
	AND	$0xff, R7, R6
	MOVV	R0, R8	// count
	ADDV	R4, R5	// end
	PCALIGN	$16
loop:
	BEQ	R5, R4, done
	MOVBU	(R4), R9
	ADDV	$1, R4
	BNE	R6, R9, loop
	ADDV	$1, R8
	JMP	loop
done:
	MOVV	R8, R4
	RET

TEXT ·CountString<ABIInternal>(SB),NOSPLIT,$0-32
	// R4 = s_base
	// R5 = s_len
	// R6 = byte to count
	AND	$0xff, R6
	MOVV	R0, R7	// count
	ADDV	R4, R5	// end
	PCALIGN	$16
loop:
	BEQ	R4, R5, done
	MOVBU	(R4), R8
	ADDV	$1, R4
	BNE	R6, R8, loop
	ADDV	$1, R7
	JMP	loop
done:
	MOVV	R7, R4
	RET
