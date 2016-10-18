// Inferno's libkern/memmove-mips.s
// https://bitbucket.org/inferno-os/inferno-os/raw/68f98ca17effa92be9e79f0294c221706624dea5/libkern/memmove-mips.s
//
//         Copyright © 1994-1999 Lucent Technologies Inc. All rights reserved.
//         Revisions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com).  All rights reserved.
//         Portions Copyright 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// +build mips mipsle

#include "textflag.h"

#ifdef GOARCH_mips
#define MOVWHI  MOVWL
#define MOVWLO  MOVWR
#else
#define MOVWHI  MOVWR
#define MOVWLO  MOVWL
#endif

TEXT runtime·memmove(SB),NOSPLIT,$-0-12
	MOVW	to+0(FP), R4
	MOVW	from+4(FP), R5
	MOVW	n+8(FP), R3

	ADDU	R3, R5, R7 /* R7 is end from-pointer */
	ADDU	R3, R4, R6 /* R6 is end to-pointer */

	/*
	 * easiest test is copy backwards if
	 * destination string has higher mem address
	 */
	SGTU	$4, R3, R2
	SGTU	R4, R5, R1
	BNE	R1, back

	/*
	 * if not at least 4 chars,
	 * don't even mess around.
	 * 3 chars to guarantee any
	 * rounding up to a word
	 * boundary and 4 characters
	 * to get at least maybe one
	 * full word store.
	 */
	BNE	R2, fout


/*
 * byte at a time to word align destination
 */
f1:
	AND	$3, R4, R1
	BEQ	R1, f2
	MOVB	0(R5), R8
	ADDU	$1, R5
	MOVB	R8, 0(R4)
	ADDU	$1, R4
	JMP	f1

/*
 * test if source is now word aligned
 */
f2:
	AND	$3, R5, R1
	BNE	R1, fun2
	/*
	 * turn R3 into to-end pointer-31
	 * copy 32 at a time while theres room.
	 * R6 is smaller than R7 --
	 * there are problems if R7 is 0.
	 */
	ADDU	$-31, R6, R3
f3:
	SGTU	R3, R4, R1
	BEQ	R1, f4
	MOVW	0(R5), R8
	MOVW	4(R5), R9
	MOVW	8(R5), R10
	MOVW	12(R5), R11
	MOVW	16(R5), R12
	MOVW	20(R5), R13
	MOVW	24(R5), R14
	MOVW	28(R5), R15
	ADDU	$32, R5
	MOVW	R8, 0(R4)
	MOVW	R9, 4(R4)
	MOVW	R10, 8(R4)
	MOVW	R11, 12(R4)
	MOVW	R12, 16(R4)
	MOVW	R13, 20(R4)
	MOVW	R14, 24(R4)
	MOVW	R15, 28(R4)
	ADDU	$32, R4
	JMP	f3

/*
 * turn R3 into to-end pointer-3
 * copy 4 at a time while theres room
 */
f4:
	ADDU	$-3, R6, R3
f5:
	SGTU	R3, R4, R1
	BEQ	R1, fout
	MOVW	0(R5), R8
	ADDU	$4, R5
	MOVW	R8, 0(R4)
	ADDU	$4, R4
	JMP	f5

/*
 * forward copy, unaligned
 * turn R3 into to-end pointer-31
 * copy 32 at a time while theres room.
 * R6 is smaller than R7 --
 * there are problems if R7 is 0.
 */
fun2:
	ADDU	$-31, R6, R3
fun3:
	SGTU	R3, R4, R1
	BEQ	R1, fun4
	MOVWHI	0(R5), R8
	MOVWLO	3(R5), R8
	MOVWHI	4(R5), R9
	MOVWLO	7(R5), R9
	MOVW	R8, 0(R4)
	MOVWHI	8(R5), R8
	MOVWLO	11(R5), R8
	MOVW	R9, 4(R4)
	MOVWHI	12(R5), R9
	MOVWLO	15(R5), R9
	MOVW	R8, 8(R4)
	MOVW	R9, 12(R4)
	MOVWHI	16(R5), R8
	MOVWLO	19(R5), R8
	MOVWHI	20(R5), R9
	MOVWLO	23(R5), R9
	MOVW	R8, 16(R4)
	MOVWHI	24(R5), R8
	MOVWLO	27(R5), R8
	MOVW	R9, 20(R4)
	MOVWHI	28(R5), R9
	MOVWLO	31(R5), R9
	ADDU	$32, R5
	MOVW	R8, 24(R4)
	MOVW	R9, 28(R4)
	ADDU	$32, R4
	JMP	fun3

/*
 * turn R3 into to-end pointer-3
 * copy 4 at a time while theres room
 */
fun4:
	ADDU	$-3, R6, R3
fun5:
	SGTU	R3, R4, R1
	BEQ	R1, fout
	MOVWHI	0(R5), R8
	MOVWLO	3(R5), R8
	ADDU	$4, R5
	MOVW	R8, 0(R4)
	ADDU	$4, R4
	JMP	fun5

/*
 * last loop, copy byte at a time
 */
fout:
	BEQ	R7, R5, ret
	MOVB	0(R5), R8
	ADDU	$1, R5
	MOVB	R8, 0(R4)
	ADDU	$1, R4
	JMP	fout

/*
 * whole thing repeated for backwards
 */
back:
	BNE	R2, bout
b1:
	AND	$3, R6, R1
	BEQ	R1, b2
	MOVB	-1(R7), R8
	ADDU	$-1, R7
	MOVB	R8, -1(R6)
	ADDU	$-1, R6
	JMP	b1

b2:
	AND	$3, R7, R1
	BNE	R1, bun2

	ADDU	$31, R5, R3
b3:
	SGTU	R7, R3, R1
	BEQ	R1, b4
	MOVW	-4(R7), R8
	MOVW	-8(R7), R9
	MOVW	-12(R7), R10
	MOVW	-16(R7), R11
	MOVW	-20(R7), R12
	MOVW	-24(R7), R13
	MOVW	-28(R7), R14
	MOVW	-32(R7), R15
	ADDU	$-32, R7
	MOVW	R8, -4(R6)
	MOVW	R9, -8(R6)
	MOVW	R10, -12(R6)
	MOVW	R11, -16(R6)
	MOVW	R12, -20(R6)
	MOVW	R13, -24(R6)
	MOVW	R14, -28(R6)
	MOVW	R15, -32(R6)
	ADDU	$-32, R6
	JMP	b3
b4:
	ADDU	$3, R5, R3
b5:
	SGTU	R7, R3, R1
	BEQ	R1, bout
	MOVW	-4(R7), R8
	ADDU	$-4, R7
	MOVW	R8, -4(R6)
	ADDU	$-4, R6
	JMP	b5

bun2:
	ADDU	$31, R5, R3
bun3:
	SGTU	R7, R3, R1
	BEQ	R1, bun4
	MOVWHI	-4(R7), R8
	MOVWLO	-1(R7), R8
	MOVWHI	-8(R7), R9
	MOVWLO	-5(R7), R9
	MOVW	R8, -4(R6)
	MOVWHI	-12(R7), R8
	MOVWLO	-9(R7), R8
	MOVW	R9, -8(R6)
	MOVWHI	-16(R7), R9
	MOVWLO	-13(R7), R9
	ADDU	$-16, R7
	MOVW	R8, -12(R6)
	MOVW	R9, -16(R6)
	ADDU	$-16, R6
	MOVWHI	-4(R7), R8
	MOVWLO	-1(R7), R8
	MOVWHI	-8(R7), R9
	MOVWLO	-5(R7), R9
	MOVW	R8, -4(R6)
	MOVWHI	-12(R7), R8
	MOVWLO	-9(R7), R8
	MOVW	R9, -8(R6)
	MOVWHI	-16(R7), R9
	MOVWLO	-13(R7), R9
	ADDU	$-16, R7
	MOVW	R8, -12(R6)
	MOVW	R9, -16(R6)
	ADDU	$-16, R6
	JMP	bun3

bun4:
	ADDU	$3, R5, R3
bun5:
	SGTU	R7, R3, R1
	BEQ	R1, bout
	MOVWHI	-4(R7), R8
	MOVWLO	-1(R7), R8
	ADDU	$-4, R7
	MOVW	R8, -4(R6)
	ADDU	$-4, R6
	JMP	bun5

bout:
	BEQ	R7, R5, ret
	MOVB	-1(R7), R8
	ADDU	$-1, R7
	MOVB	R8, -1(R6)
	ADDU	$-1, R6
	JMP	bout

ret:
	RET

