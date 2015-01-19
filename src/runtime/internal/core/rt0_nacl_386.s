// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// NaCl entry has:
//	0(FP) - arg block == SP+8
//	4(FP) - cleanup function pointer, always 0
//	8(FP) - envc
//	12(FP) - Argc
//	16(FP) - Argv, then 0, then envv, then 0, then auxv
TEXT _rt0_386_nacl(SB),NOSPLIT,$8
	MOVL	Argc+12(FP), AX
	LEAL	Argv+16(FP), BX
	MOVL	AX, 0(SP)
	MOVL	BX, 4(SP)
	CALL	main(SB)
	INT	$3

TEXT main(SB),NOSPLIT,$0
	JMP	runtime∕internal∕schedinit·rt0_go(SB)
