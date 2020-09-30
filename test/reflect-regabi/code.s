// func addInt(a, b int) int
TEXT ·addInt(SB),$0-0
	ADDQ BX, AX
	RET

// func addFloat64(a, b float64) float64
TEXT ·addFloat64(SB),$0-0
	ADDSD X1, X0
	RET

// func sumSpillInt(a, b, c, d, e, f, g, h, i, j int) int
TEXT ·sumSpillInt(SB),$0-8
	ADDQ BX, AX
	ADDQ CX, AX
	ADDQ DI, AX
	ADDQ SI, AX
	ADDQ R8, AX
	ADDQ R9, AX
	ADDQ R10, AX
	ADDQ R11, AX
	MOVQ j+0(FP), R12
	ADDQ R12, AX
	RET

// func sumSpillFloat64(a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p float64) float64
TEXT ·sumSpillFloat64(SB),$0-8
	ADDSD X1, X0
	ADDSD X2, X0
	ADDSD X3, X0
	ADDSD X4, X0
	ADDSD X5, X0
	ADDSD X6, X0
	ADDSD X7, X0
	ADDSD X8, X0
	ADDSD X9, X0
	ADDSD X10, X0
	ADDSD X11, X0
	ADDSD X12, X0
	ADDSD X13, X0
	ADDSD X14, X0
	MOVQ p+0(FP), X1
	ADDSD X1, X0
	RET

// func sumSpillMix(a, b, c, d, e, f, g, h, i, j int, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z float64) (int, float64)
TEXT ·sumSpillMix(SB),$0-16
	ADDQ BX, AX
	ADDQ CX, AX
	ADDQ DI, AX
	ADDQ SI, AX
	ADDQ R8, AX
	ADDQ R9, AX
	ADDQ R10, AX
	ADDQ R11, AX
	MOVQ j+0(FP), R12
	ADDQ R12, AX
	ADDSD X1, X0
	ADDSD X2, X0
	ADDSD X3, X0
	ADDSD X4, X0
	ADDSD X5, X0
	ADDSD X6, X0
	ADDSD X7, X0
	ADDSD X8, X0
	ADDSD X9, X0
	ADDSD X10, X0
	ADDSD X11, X0
	ADDSD X12, X0
	ADDSD X13, X0
	ADDSD X14, X0
	MOVQ z+8(FP), X1
	ADDSD X1, X0
	RET

// func splitSpillInt(a int) (b, c, d, e, f, g, h, i, j, k int)
TEXT ·splitSpillInt(SB),$0-8
	SUBQ $9, AX
	MOVQ $1, BX
	MOVQ $1, CX
	MOVQ $1, DI
	MOVQ $1, SI
	MOVQ $1, R8
	MOVQ $1, R9
	MOVQ $1, R10
	MOVQ $1, R11
	MOVQ $1, k+0(FP)
	RET

// func passArray(a [1]uint32) [1]uint32
TEXT ·passArray(SB),$0-8
	MOVLQZX a+0(FP), AX
	MOVL AX, out+8(FP)
	RET

// func passArrayMix(f int, a [1]uint32, g float64) (int, [1]uint32, float64)
TEXT ·passArrayMix(SB),$0-8
	// f and g are passed and returned in the same registers.
	MOVLQZX a+0(FP), R12
	MOVL R12, out+8(FP)
	RET

// func passString(a string) string
TEXT ·passString(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passInterface(a interface{}) interface{}
TEXT ·passInterface(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passSlice(a []byte) []byte
TEXT ·passSlice(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func setPointer(a *byte) *byte
TEXT ·setPointer(SB),$0-0
	MOVB $231, 0(AX)
	RET

// func passStruct1(a Struct1) Struct1
TEXT ·passStruct1(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct2(a Struct2) Struct2
TEXT ·passStruct2(SB),$0-64
	MOVQ a_a+0(FP), AX
	MOVQ AX, out_a+32(FP)
	MOVQ a_b+8(FP), AX
	MOVQ AX, out_b+40(FP)
	MOVQ a_c+16(FP), AX
	MOVQ AX, out_c+48(FP)
	MOVL a_d_0+24(FP), AX
	MOVL AX, out_d_0+56(FP)
	MOVL a_d_1+28(FP), AX
	MOVL AX, out_d_1+60(FP)
	RET

// func passStruct3(a Struct3) Struct3
TEXT ·passStruct3(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct4(a Struct4) Struct4
TEXT ·passStruct4(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct5(a Struct5) Struct5
TEXT ·passStruct5(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct6(a Struct6) Struct6
TEXT ·passStruct6(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct7(a Struct7) Struct7
TEXT ·passStruct7(SB),$0-112
	MOVQ a_Struct1_a+0(FP), AX
	MOVQ AX, out_Struct1_a+56(FP)
	MOVQ a_Struct1_b+8(FP), AX
	MOVQ AX, out_Struct1_b+64(FP)
	MOVQ a_Struct1_c+16(FP), AX
	MOVQ AX, out_Struct1_c+72(FP)
	MOVQ a_Struct2_a+24(FP), AX
	MOVQ AX, out_Struct2_a+80(FP)
	MOVQ a_Struct2_b+32(FP), AX
	MOVQ AX, out_Struct1_b+88(FP)
	MOVQ a_Struct2_c+40(FP), AX
	MOVQ AX, out_Struct2_c+96(FP)
	MOVL a_Struct2_d_0+48(FP), AX
	MOVL AX, out_Struct2_d_0+104(FP)
	MOVL a_Struct2_d_1+52(FP), AX
	MOVL AX, out_Struct2_d_1+108(FP)
	RET

// func passStruct8(a Struct8) Struct8
TEXT ·passStruct8(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct9(a Struct9) Struct9
TEXT ·passStruct9(SB),$0-160
	MOVQ a_Struct1_a+0(FP), AX
	MOVQ AX, out_Struct1_a+80(FP)
	MOVQ a_Struct1_b+8(FP), AX
	MOVQ AX, out_Struct1_b+88(FP)
	MOVQ a_Struct1_c+16(FP), AX
	MOVQ AX, out_Struct1_c+96(FP)
	MOVQ a_Struct7_Struct1_a+24(FP), AX
	MOVQ AX, out_Struct7_Struct1_a+104(FP)
	MOVQ a_Struct7_Struct1_b+32(FP), AX
	MOVQ AX, out_Struct7_Struct1_b+112(FP)
	MOVQ a_Struct7_Struct1_c+40(FP), AX
	MOVQ AX, out_Struct7_Struct1_c+120(FP)
	MOVQ a_Struct7_Struct2_a+48(FP), AX
	MOVQ AX, out_Struct7_Struct2_a+128(FP)
	MOVQ a_Struct7_Struct2_b+56(FP), AX
	MOVQ AX, out_Struct7_Struct1_b+136(FP)
	MOVQ a_Struct7_Struct2_c+64(FP), AX
	MOVQ AX, out_Struct7_Struct2_c+144(FP)
	MOVL a_Struct7_Struct2_d_0+72(FP), AX
	MOVL AX, out_Struct7_Struct2_d_0+152(FP)
	MOVL a_Struct7_Struct2_d_1+76(FP), AX
	MOVL AX, out_Struct7_Struct2_d_1+156(FP)
	RET

// func passStruct10(a Struct10) Struct10
TEXT ·passStruct10(SB),$0-224
	MOVW a_Struct5_a+0(FP), AX
	MOVW AX, out_Struct5_a+112(FP)
	MOVW a_Struct5_b+2(FP), AX
	MOVW AX, out_Struct5_b+114(FP)
	MOVL a_Struct5_c+4(FP), AX
	MOVL AX, out_Struct5_c+116(FP)
	MOVL a_Struct5_d+8(FP), AX
	MOVL AX, out_Struct5_d+120(FP)
	MOVL a_Struct5_e+12(FP), AX
	MOVL AX, out_Struct5_e+124(FP)
	MOVL a_Struct5_f+16(FP), AX
	MOVL AX, out_Struct5_f+128(FP)
	MOVL a_Struct5_g+20(FP), AX
	MOVL AX, out_Struct5_g+132(FP)
	MOVL a_Struct5_h+24(FP), AX
	MOVL AX, out_Struct5_h+136(FP)
	MOVL a_Struct5_i+28(FP), AX
	MOVL AX, out_Struct5_i+140(FP)
	MOVL a_Struct5_j+32(FP), AX
	MOVL AX, out_Struct5_j+144(FP)

	MOVW a_Struct8_Struct5_a+40(FP), AX
	MOVW AX, out_Struct8_Struct5_a+152(FP)
	MOVW a_Struct8_Struct5_b+42(FP), AX
	MOVW AX, out_Struct8_Struct5_b+154(FP)
	MOVL a_Struct8_Struct5_c+44(FP), AX
	MOVL AX, out_Struct8_Struct5_c+156(FP)
	MOVL a_Struct8_Struct5_d+48(FP), AX
	MOVL AX, out_Struct8_Struct5_d+160(FP)
	MOVL a_Struct8_Struct5_e+52(FP), AX
	MOVL AX, out_Struct8_Struct5_e+164(FP)
	MOVL a_Struct8_Struct5_f+56(FP), AX
	MOVL AX, out_Struct8_Struct5_f+168(FP)
	MOVL a_Struct8_Struct5_g+60(FP), AX
	MOVL AX, out_Struct8_Struct5_g+172(FP)
	MOVL a_Struct8_Struct5_h+64(FP), AX
	MOVL AX, out_Struct8_Struct5_h+176(FP)
	MOVL a_Struct8_Struct5_i+68(FP), AX
	MOVL AX, out_Struct8_Struct5_i+180(FP)
	MOVL a_Struct8_Struct5_j+72(FP), AX
	MOVL AX, out_Struct8_Struct5_j+184(FP)

	MOVQ a_Struct8_Struct3_a+80(FP), AX
	MOVQ AX, out_Struct8_Struct3_a+192(FP)
	MOVQ a_Struct8_Struct3_b+88(FP), AX
	MOVQ AX, out_Struct8_Struct3_b+200(FP)
	MOVQ a_Struct8_Struct3_c+96(FP), AX
	MOVQ AX, out_Struct8_Struct3_c+208(FP)
	RET

// func passStruct11(a Struct11) Struct11
TEXT ·passStruct11(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passStruct12(a Struct12) Struct12
TEXT ·passStruct12(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func incStruct13(a Struct13) Struct13
TEXT ·incStruct13(SB),$0-0
	// Passed and returned in the same registers.
	ADDQ $1, AX
	ADDQ $1, BX
	RET

// func pass2Struct1(a, b Struct1) (x, y Struct1)
TEXT ·pass2Struct1(SB),$0-0
	// Passed and returned in the same registers.
	RET

// func passEmptyStruct(a int, b struct{}, c float64) (int, struct{}, float64)
TEXT ·passEmptyStruct(SB),$0-0
	// Passed and returned in the same registers.
	RET
