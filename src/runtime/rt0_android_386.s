// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_386_android(SB),NOSPLIT,$8
	MOVL	8(SP), AX  // argc
	LEAL	12(SP), BX  // argv
	MOVL	AX, 0(SP)
	MOVL	BX, 4(SP)
	CALL	main(SB)
	INT	$3

TEXT _rt0_386_android_lib(SB),NOSPLIT,$0
	PUSHL	$1  // argc
	PUSHL	$_rt0_386_android_argv(SB)  // argv
	CALL	_rt0_386_linux_lib(SB)
	POPL	AX
	POPL	AX
	RET

DATA _rt0_386_android_argv+0x00(SB)/4,$_rt0_386_android_argv0(SB)
DATA _rt0_386_android_argv+0x04(SB)/4,$0
DATA _rt0_386_android_argv+0x08(SB)/4,$0
DATA _rt0_386_android_argv+0x0c(SB)/4,$15  // AT_PLATFORM  (auxvec.h)
DATA _rt0_386_android_argv+0x10(SB)/4,$_rt0_386_android_auxv0(SB)
DATA _rt0_386_android_argv+0x14(SB)/4,$0
DATA _rt0_386_android_argv+0x18(SB)/4,$0
DATA _rt0_386_android_argv+0x1c(SB)/4,$0
GLOBL _rt0_386_android_argv(SB),NOPTR,$0x20

// TODO: AT_HWCAP necessary? If so, what value?

DATA _rt0_386_android_argv0(SB)/8, $"gojni"
GLOBL _rt0_386_android_argv0(SB),RODATA,$8

DATA _rt0_386_android_auxv0(SB)/8, $"i386"
GLOBL _rt0_386_android_auxv0(SB),RODATA,$8
