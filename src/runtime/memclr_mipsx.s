// Inferno's libkern/memset-mips.s
// https://bitbucket.org/inferno-os/inferno-os/raw/68f98ca17effa92be9e79f0294c221706624dea5/libkern/memset-mips.s
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

// void runtime·memclr(void*, uintptr)
TEXT runtime·memclr(SB),NOSPLIT,$0-8
	MOVW	ptr+0(FP), R4 /* R4 is pointer */
	MOVW	n+4(FP), R3 /* R3 is count */
	ADDU	R3, R4, R6 /* R6 is end pointer */

	/*
	 * if not at least 4 chars,
	 * dont even mess around.
	 * 3 chars to guarantee any
	 * rounding up to a word
	 * boundary and 4 characters
	 * to get at least maybe one
	 * full word store.
	 */
	SGT	$4, R3, R1
	BNE	R1, out

/*
 * store one byte at a time until pointer
 * is alligned on a word boundary
 */
l1:
	AND	$3, R4, R1
	BEQ	R1, l2
	MOVB	R0, 0(R4)
	ADDU	$1, R4
	JMP	l1

/*
 * turn R3 into end pointer-15
 * store 16 at a time while theres room
 */
l2:
	ADDU	$-15, R6, R3
l3:
	SGTU	R3, R4, R1
	BEQ	R1, l4
	MOVW	R0, 0(R4)
	MOVW	R0, 4(R4)
	ADDU	$16, R4
	MOVW	R0, -8(R4)
	MOVW	R0, -4(R4)
	JMP	l3

/*
 * turn R3 into end pointer-3
 * store 4 at a time while theres room
 */
l4:
	ADDU	$-3, R6, R3
l5:
	SGTU	R3, R4, R1
	BEQ	R1, out
	MOVW	R0, 0(R4)
	ADDU	$4, R4
	JMP	l5

/*
 * last loop, store byte at a time
 */
out:
	SGTU	R6, R4 , R1
	BEQ	R1, ret
	MOVB	R0, 0(R4)
	ADDU	$1, R4
	JMP	out

ret:
	RET
