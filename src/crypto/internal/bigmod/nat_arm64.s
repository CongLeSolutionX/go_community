//go:build arm64 && gc && !purego

// #include "textflag.h"

// func montgomeryLoop(d []uint, a []uint, b []uint, m []uint, m0inv uint) uint
// TEXT ·montgomeryLoop(SB),NOSPLIT|NOFRAME,$0
TEXT ·montgomeryLoop(SB),$32-112
	// R0:  d_base
	// R1:  d_len
	// R2:  ai
	// R3:  b_base
	// R4:  m_base
	// R5:  m0inv
	// R6:  overflow
	// R7:  i
	// R8:  j
	// R9:  z_lo
	// R10: z_hi
	// R11: f
	// R12: scratch (m[j]*f lo)
	// R13: scratch (m[j]*f hi)
	// R14: carry
	// R15: scratch (various register derefs)
	// R20: d_len-1

	MOVD d_base+0(FP), R0
	MOVD d_len+8(FP), R1
	MOVD b_base+48(FP), R3
	MOVD m_base+72(FP), R4
	MOVD m0inv+96(FP), R5
	
	EOR R6, R6 // overflow
	EOR R7, R7 // i

	// MOVD R1, R20
	SUB $1, R1, R20 // size-1
	
	PCALIGN $16
outerLoop:

	// reuse R2 for ai
	MOVD a_base+24(FP), R2
	MOVD (R2)(R7<<3), R2 // ai
	
	// z = b[0] * ai
	MOVD (R3), R15
	UMULH R15, R2, R10
	MUL R15, R2, R9

	// z += d[0]
	MOVD (R0), R15
	ADDS R15, R9
	ADC ZR, R10

	// f = m0inv * z.lo
	MUL R5, R9, R11

	// f &= _MASK
	BFM $1, ZR, $0, R11

	// z += m[0] * f
	MOVD (R4), R15
	MUL R15, R11, R12
	ADDS R12, R9
	UMULH R15, R11, R13
	ADC R13, R10
	
	// carry = z_hi<<1 | z_lo>>_W
	EXTR $63, R9, R10, R14

	// j = 1
	MOVD $1, R8
	JMP innerLoopCondition

	PCALIGN $16
innerLoop:

	// z = d[j] + a[i] * b[j] + f * m[j] + carry

	// z = a[i] * b[j]
	MOVD (R3)(R8<<3), R15
	UMULH R15, R2, R10
	MUL R15, R2, R9
	
	// z += z + m[j] * f
	MOVD (R4)(R8<<3), R15
	MUL R15, R11, R12
	ADDS R12, R9
	UMULH R15, R11, R13
	ADC R13, R10

	// z += z + d[j]
	MOVD (R0)(R8<<3), R15
	ADDS R15, R9
	ADC ZR, R10

	// z += carry
	ADDS R14, R9
	ADC ZR, R10
	
	// d[j-1] = z_lo & _MASK
	MOVD R9, R15 
	BFM $1, ZR, $0, R15
	SUB $1, R8, R12
	MOVD R15, (R0)(R12<<3) // mask and store in a single op?
	
	// carry = z_hi<<1 | z_lo>>_W
	EXTR $63, R9, R10, R14
	
	// j++
	ADD $1, R8
	
	PCALIGN $16
innerLoopCondition:

	// d_len > j, jump to innerLoop
	CMP R8, R1
	BGT innerLoop
	
	// overflow += carry
	ADD R14, R6
	
	// d[size-1] = overflow & _MASK
	MOVD R6, R15
	BFM $1, ZR, $0, R15
	MOVD R15, (R0)(R20<<3)
	
	// overflow = overflow >> 63
	LSR $63, R6, R6

	// i++
	ADD $1, R7
	CMP R7, R1
	BGT outerLoop

	// return overflow
	MOVD R6, ret+104(FP)
	RET
