// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !math_big_pure_go && riscv64

#include "textflag.h"

// This file provides fast assembly versions for the elementary
// arithmetic operations on vectors implemented in arith.go.

TEXT ·addVV(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30

	MOV	$4, X28
	MOV	$0, X29		// c = 0

	BEQZ	X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	0(X6), X9	// y[0]
	MOV	8(X5), X11	// x[1]
	MOV	8(X6), X12	// y[1]
	MOV	16(X5), X14	// x[2]
	MOV	16(X6), X15	// y[2]
	MOV	24(X5), X17	// x[3]
	MOV	24(X6), X18	// y[3]

	ADD	X8, X9, X21	// z[0] = x[0] + y[0]
	SLTU	X8, X21, X22
	ADD	X21, X29, X10	// z[0] = x[0] + y[0] + c
	SLTU	X21, X10, X23
	ADD	X22, X23, X29	// next c

	ADD	X11, X12, X24	// z[1] = x[1] + y[1]
	SLTU	X11, X24, X25
	ADD	X24, X29, X13	// z[1] = x[1] + y[1] + c
	SLTU	X24, X13, X26
	ADD	X25, X26, X29	// next c

	ADD	X14, X15, X21	// z[2] = x[2] + y[2]
	SLTU	X14, X21, X22
	ADD	X21, X29, X16	// z[2] = x[2] + y[2] + c
	SLTU	X21, X16, X23
	ADD	X22, X23, X29	// next c

	ADD	X17, X18, X21	// z[3] = x[3] + y[3]
	SLTU	X17, X21, X22
	ADD	X21, X29, X19	// z[3] = x[3] + y[3] + c
	SLTU	X21, X19, X23
	ADD	X22, X23, X29	// next c

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X6
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x
	MOV	0(X6), X11	// y

	ADD	X10, X11, X12	// z = x + y
	SLTU	X10, X12, X14
	ADD	X12, X29, X13	// z = x + y + c
	SLTU	X12, X13, X15
	ADD	X14, X15, X29	// next c

	MOV	X13, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X6
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+72(FP)	// return c
	RET

TEXT ·subVV(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30

	MOV	$4, X28
	MOV	$0, X29		// b = 0

	BEQZ	X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	0(X6), X9	// y[0]
	MOV	8(X5), X11	// x[1]
	MOV	8(X6), X12	// y[1]
	MOV	16(X5), X14	// x[2]
	MOV	16(X6), X15	// y[2]
	MOV	24(X5), X17	// x[3]
	MOV	24(X6), X18	// y[3]

	SUB	X9, X8, X21	// z[0] = x[0] - y[0]
	SLTU	X21, X8, X22
	SUB	X29, X21, X10	// z[0] = x[0] - y[0] - b
	SLTU	X10, X21, X23
	ADD	X22, X23, X29	// next b

	SUB	X12, X11, X24	// z[1] = x[1] - y[1]
	SLTU	X24, X11, X25
	SUB	X29, X24, X13	// z[1] = x[1] - y[1] - b
	SLTU	X13, X24, X26
	ADD	X25, X26, X29	// next b

	SUB	X15, X14, X21	// z[2] = x[2] - y[2]
	SLTU	X21, X14, X22
	SUB	X29, X21, X16	// z[2] = x[2] - y[2] - b
	SLTU	X16, X21, X23
	ADD	X22, X23, X29	// next b

	SUB	X18, X17, X21	// z[3] = x[3] - y[3]
	SLTU	X21, X17, X22
	SUB	X29, X21, X19	// z[3] = x[3] - y[3] - b
	SLTU	X19, X21, X23
	ADD	X22, X23, X29	// next b

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X6
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x
	MOV	0(X6), X11	// y

	SUB	X11, X10, X12	// z = x - y
	SLTU	X12, X10, X14
	SUB	X29, X12, X13	// z = x - y - b
	SLTU	X13, X12, X15
	ADD	X14, X15, X29	// next b

	MOV	X13, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X6
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+72(FP)	// return b
	RET

TEXT ·addVW(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30

	MOV	$4, X28
	MOV	X6, X29		// c = y

	BEQZ	X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	8(X5), X11	// x[1]
	MOV	16(X5), X14	// x[2]
	MOV	24(X5), X17	// x[3]

	ADD	X8, X29, X10	// z[0] = x[0] + c
	SLTU	X8, X10, X29	// next c

	ADD	X11, X29, X13	// z[1] = x[1] + c
	SLTU	X11, X13, X29	// next c

	ADD	X14, X29, X16	// z[2] = x[2] + c
	SLTU	X14, X16, X29	// next c

	ADD	X17, X29, X19	// z[3] = x[3] + c
	SLTU	X17, X19, X29	// next c

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x

	ADD	X10, X29, X12	// z = x + c
	SLTU	X10, X12, X29	// next c

	MOV	X12, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+56(FP)	// return c
	RET

TEXT ·subVW(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30

	MOV	$4, X28
	MOV	X6, X29		// b = y

	BEQZ	X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	8(X5), X11	// x[1]
	MOV	16(X5), X14	// x[2]
	MOV	24(X5), X17	// x[3]

	SUB	X29, X8, X10	// z[0] = x[0] - b
	SLTU	X10, X8, X29	// next b

	SUB	X29, X11, X13	// z[1] = x[1] - b
	SLTU	X13, X11, X29	// next b

	SUB	X29, X14, X16	// z[2] = x[2] - b
	SLTU	X16, X14, X29	// next b

	SUB	X29, X17, X19	// z[3] = x[3] - b
	SLTU	X19, X17, X29	// next b

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x

	SUB	X29, X10, X12	// z = x - b
	SLTU	X12, X10, X29	// next b

	MOV	X12, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+56(FP)	// return b
	RET

TEXT ·shlVU(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	s+48(FP), X25
	MOV	z+0(FP), X6
	MOV	z_len+8(FP), X30

	MOV	$0, X29		// c = 0
	BEQZ	X30, done

	AND	$63, X25
	MOV	$64, X26
	SUB	X25, X26, X26	// sinv
	AND	$63, X26

	SUB	$1, X30
	MOV	$8, X10
	MUL	X10, X30, X10
	ADD	X10, X5		// x[len(z)-1]
	ADD	X10, X6		// z[len(z)-1]

	MOV	(X5), X24	// c = x[len(z)-1] >> sinv
	SRL	X26, X24, X29
	SEQZ	X25, X10	// c = 0, if s == 0
	SUB	$1, X10, X19
	AND	X19, X29

	BEQZ	X30, last

	MOV	$4, X28
	BLTU	X30, X28, loop1

loop4:
	MOV	X24, X20	// x[i-0]
	MOV	(-1*8)(X5), X21	// x[i-1]
	MOV	(-2*8)(X5), X22	// x[i-2]
	MOV	(-3*8)(X5), X23	// x[i-3]
	MOV	(-4*8)(X5), X24	// x[i-4]

	SLL	X25, X20, X10	// x[i-0] << s
	SLL	X25, X21, X12	// x[i-1] << s
	SLL	X25, X22, X14	// x[i-2] << s
	SLL	X25, X23, X16	// x[i-3] << s

	SRL	X26, X21, X11	// x[i-1] >> sinv
	SRL	X26, X22, X13	// x[i-2] >> sinv
	SRL	X26, X23, X15	// x[i-3] >> sinv
	SRL	X26, X24, X17	// x[i-4] >> sinv

	AND	X19, X11
	AND	X19, X13
	AND	X19, X15
	AND	X19, X17

	OR	X10, X11, X10	// x[i-0]<<s | x[i-1]>>sinv
	OR	X12, X13, X12	// x[i-1]<<s | x[i-2]>>sinv
	OR	X14, X15, X14	// x[i-2]<<s | x[i-3]>>sinv
	OR	X16, X17, X16	// x[i-3]<<s | x[i-4]>>sinv

	MOV	X10, (-0*8)(X6)	// z[i-0] = x[i-0]<<s | x[i-1]>>sinv
	MOV	X12, (-1*8)(X6)	// z[i-1] = x[i-1]<<s | x[i-2]>>sinv
	MOV	X14, (-2*8)(X6)	// z[i-2] = x[i-2]<<s | x[i-3]>>sinv
	MOV	X16, (-3*8)(X6)	// z[i-3] = x[i-3]<<s | x[i-4]>>sinv

	SUB	$32, X5
	SUB	$32, X6
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, last

loop1:
	MOV	X24, X20	// x[i]
	MOV	(-1*8)(X5), X24	// x[i-1]
	SLL	X25, X20, X10	// x[i] << s
	SRL	X26, X24, X11	// x[i-1] >> sinv
	AND	X19, X11
	OR	X10, X11, X12	// x[i]<<s | x[i-1]>>sinv
	MOV	X12, (X6)	// z[i] = x[i]<<s | x[i-1]>>sinv

	SUB	$8, X5
	SUB	$8, X6
	SUB	$1, X30

	BNEZ	X30, loop1

last:
	MOV	(X5), X10	// x[0]
	SLL	X25, X10	// x[0] << s
	MOV	X10, (X6)	// z[0] = x[0]<<s

done:
	MOV	X29, c+56(FP)	// return c
	RET

TEXT ·shrVU(SB),NOSPLIT,$0
	JMP ·shrVU_g(SB)

TEXT ·mulAddVWW(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30
	MOV	r+56(FP), X29

	MOV	$4, X28

	BEQ	ZERO, X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	8(X5), X11	// x[1]
	MOV	16(X5), X14	// x[2]
	MOV	24(X5), X17	// x[3]

	MULHU	X8, X6, X9	// z_hi[0] = x * y
	MUL	X8, X6, X8	// z_lo[0] = x * y
	ADD	X8, X29, X10	// z_lo[0] = z_lo[0] + c
	SLTU	X8, X10, X23
	ADD	X23, X9, X29	// next c

	MULHU	X11, X6, X12	// z_hi[0] = x * y
	MUL	X11, X6, X11	// z_lo[0] = x * y
	ADD	X11, X29, X13	// z_lo[0] = z_lo[0] + c
	SLTU	X11, X13, X23
	ADD	X23, X12, X29	// next c

	MULHU	X14, X6, X15	// z_hi[0] = x * y
	MUL	X14, X6, X14	// z_lo[0] = x * y
	ADD	X14, X29, X16	// z_lo[0] = z_lo[0] + c
	SLTU	X14, X16, X23
	ADD	X23, X15, X29	// next c

	MULHU	X17, X6, X18	// z_hi[0] = x * y
	MUL	X17, X6, X17	// z_lo[0] = x * y
	ADD	X17, X29, X19	// z_lo[0] = z_lo[0] + c
	SLTU	X17, X19, X23
	ADD	X23, X18, X29	// next c

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x

	MULHU	X10, X6, X12	// z_hi = x * y
	MUL	X10, X6, X10	// z_lo = x * y
	ADD	X10, X29, X13	// z_lo + c
	SLTU	X10, X13, X15
	ADD	X12, X15, X29	// next c

	MOV	X13, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+64(FP)	// return c
	RET

TEXT ·addMulVVW(SB),NOSPLIT,$0
	MOV	x+24(FP), X5
	MOV	y+48(FP), X6
	MOV	z+0(FP), X7
	MOV	z_len+8(FP), X30

	MOV	$4, X28
	MOV	$0, X29		// c = 0

	BEQZ	X30, done
	BLTU	X30, X28, loop1

loop4:
	MOV	0(X5), X8	// x[0]
	MOV	0(X7), X10	// z[0]
	MOV	8(X5), X11	// x[1]
	MOV	8(X7), X13	// z[1]
	MOV	16(X5), X14	// x[2]
	MOV	16(X7), X16	// z[2]
	MOV	24(X5), X17	// x[3]
	MOV	24(X7), X19	// z[3]

	MULHU	X8, X6, X9	// z_hi[0] = x[0] * y
	MUL	X8, X6, X8	// z_lo[0] = x[0] * y
	ADD	X8, X10, X21	// z_lo[0] = x[0] * y + z[0]
	SLTU	X8, X21, X22
	ADD	X9, X22, X9	// z_hi[0] = x[0] * y + z[0]
	ADD	X21, X29, X10	// z_lo[0] = x[0] * y + z[0] + c
	SLTU	X21, X10, X22
	ADD	X9, X22, X29	// next c

	MULHU	X11, X6, X12	// z_hi[1] = x[1] * y
	MUL	X11, X6, X11	// z_lo[1] = x[1] * y
	ADD	X11, X13, X21	// z_lo[1] = x[1] * y + z[1]
	SLTU	X11, X21, X22
	ADD	X12, X22, X12	// z_hi[1] = x[1] * y + z[1]
	ADD	X21, X29, X13	// z_lo[1] = x[1] * y + z[1] + c
	SLTU	X21, X13, X22
	ADD	X12, X22, X29	// next c

	MULHU	X14, X6, X15	// z_hi[0] = x[0] * y
	MUL	X14, X6, X14	// z_lo[0] = x[0] * y
	ADD	X14, X16, X21	// z_lo[0] = x[0] * y + z[0]
	SLTU	X14, X21, X22
	ADD	X15, X22, X15	// z_hi[0] = x[0] * y + z[0]
	ADD	X21, X29, X16	// z_lo[0] = x[0] * y + z[0] + c
	SLTU	X21, X16, X22
	ADD	X15, X22, X29	// next c

	MULHU	X17, X6, X18	// z_hi[0] = x[0] * y
	MUL	X17, X6, X17	// z_lo[0] = x[0] * y
	ADD	X17, X19, X21	// z_lo[0] = x[0] * y + z[0]
	SLTU	X17, X21, X22
	ADD	X18, X22, X18	// z_hi[0] = x[0] * y + z[0]
	ADD	X21, X29, X19	// z_lo[0] = x[0] * y + z[0] + c
	SLTU	X21, X19, X22
	ADD	X18, X22, X29	// next c

	MOV	X10, 0(X7)	// z[0]
	MOV	X13, 8(X7)	// z[1]
	MOV	X16, 16(X7)	// z[2]
	MOV	X19, 24(X7)	// z[3]

	ADD	$32, X5
	ADD	$32, X7
	SUB	$4, X30

	BGEU	X30, X28, loop4
	BEQZ	X30, done

loop1:
	MOV	0(X5), X10	// x
	MOV	0(X7), X11	// z

	MULHU	X10, X6, X12	// z_hi = x * y
	MUL	X10, X6, X10	// z_lo = x * y
	ADD	X10, X11, X13	// z_lo = x * y + z
	SLTU	X10, X13, X15
	ADD	X12, X15, X12	// z_hi = x * y + z
	ADD	X13, X29, X10	// z_lo = x * y + z + c
	SLTU	X13, X10, X15
	ADD	X12, X15, X29	// next c

	MOV	X10, 0(X7)	// z

	ADD	$8, X5
	ADD	$8, X7
	SUB	$1, X30

	BNEZ	X30, loop1

done:
	MOV	X29, c+56(FP)	// return c
	RET
