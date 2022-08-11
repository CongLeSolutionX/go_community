// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package arm64 implements an ARM64 assembler. Go assembly syntax is different from GNU ARM64
syntax, but we can still follow the general rules to map between them.

# Instructions mnemonics mapping rules

1. Most non-vector instructions use width suffixes of instruction names to indicate operand width
rather than using different register names.
General instructions:

	32-bit: "W";
	64-bit: no suffix;

Scalar SIMD&FP, SVE and SME instructions:

	8-bit: "B"
	16-bit: "H"
	32-bit: "S"
	64-bit: "D"
	128-bit: "Q"

Examples:

	ADC R24, R14, R12          <=>     adc x12, x14, x24
	ADDW R26->24, R21, R15     <=>     add w15, w21, w26, asr #24
	FCMPS F2, F3               <=>     fcmp s3, s2
	FCMPD F2, F3               <=>     fcmp d3, d2
	FCVTDH F2, F3              <=>     fcvt h3, d2

2. Go uses .P and .W suffixes to indicate post-increment and pre-increment.

Examples:

	MOVD.P -8(R10), R8         <=>      ldr x8, [x10],#-8
	MOVB.W 16(R16), R10        <=>      ldrsb x10, [x16,#16]!
	MOVBU.W 16(R16), R10       <=>      ldrb x10, [x16,#16]!

3. Go moves conditions into opcode suffix, like BLT.

4. Vector instructions have no suffix.

5. 16-bit(half-precision) non-vector SIMD&FP instructions have a "H" suffix.

6. For advanced SIMD instructions, add a "V" prefix. For floating-point SIMD instructions that do not
start with "F", add an "F" prefix. For cryptographic extension instructions and float instructions, don't
add a prefix.

Examples:

	VADD V5.H8, V18.H8, V9.H8         <=>      add v9.8h, v18.8h, v5.8h
	VLD1.P (R6)(R11), [V31.D1]        <=>      ld1 {v31.1d}, [x6], x11
	VFMLA V29.S2, V20.S2, V14.S2      <=>      fmla v14.2s, v20.2s, v29.2s
	FLDPD (R0), (F1, F2)              <=>      ldp d1, d2, [x0]
	FABSD F1, F2                      <=>      fabs d2, d1
	AESD V22.B16, V19.B16             <=>      aesd v19.16b, v22.16b

7. For a SVE and SME instruction, if it contains Z register, add a "Z" prefix, else if it contains P register,
add a "P" prefix; else no prefix.

8. Go uses a series of MOV instructions as load and store.

64-bit variant ldr, str, stur => MOVD;
32-bit variant str, stur, ldrsw => MOVW;
32-bit variant ldr => MOVWU;
ldrb, ldurb => MOVBU; ldrh, ldurh => MOVHU;
ldrsb, sturb, strb => MOVB (64-bit), MOVBW (32-bit);
ldrsh, sturh, strh =>  MOVH (64-bit), MOVHW (32-bit);
ldrsw => MOVW
SVE/SME ldr, str with Z registers => ZMOV;
SVE/SME ldr, str with P registers => PMOV;

9. Align directive

Go asm supports the PCALIGN directive, which indicates that the next instruction should be aligned
to a specified boundary by padding with NOOP instruction. The alignment value supported on arm64
must be a power of 2 and in the range of [8, 2048].

Examples:

	PCALIGN $16
	MOVD $2, R0          // This instruction is aligned with 16 bytes.
	PCALIGN $1024
	MOVD $3, R1          // This instruction is aligned with 1024 bytes.

PCALIGN also changes the function alignment. If a function has one or more PCALIGN directives,
its address will be aligned to the same or coarser boundary, which is the maximum of all the
alignment values.

In the following example, the function Add is aligned with 128 bytes.

Examples:

	TEXT ·Add(SB),$40-16
	MOVD $2, R0
	PCALIGN $32
	MOVD $4, R1
	PCALIGN $128
	MOVD $8, R2
	RET

On arm64, functions in Go are aligned to 16 bytes by default, we can also use PCALIGN to set the
function alignment. The functions that need to be aligned are preferably using NOFRAME and NOSPLIT
to avoid the impact of the prologues inserted by the assembler, so that the function address will
have the same alignment as the first hand-written instruction.

In the following example, PCALIGN at the entry of the function Add will align its address to 2048 bytes.

Examples:

	TEXT ·Add(SB),NOSPLIT|NOFRAME,$0
	  PCALIGN $2048
	  MOVD $1, R0
	  MOVD $1, R1
	  RET

10. Move large constants to vector registers.

Go asm uses VMOVQ/VMOVD/VMOVS to move 128-bit, 64-bit and 32-bit constants into vector registers, respectively.
And for a 128-bit integer, it take two 64-bit operands, for the low and high parts separately.

Examples:

	VMOVS $0x11223344, V0
	VMOVD $0x1122334455667788, V1
	VMOVQ $0x1122334455667788, $0x99aabbccddeeff00, V2   // V2=0x99aabbccddeeff001122334455667788

11. Move an optionally-shifted 16-bit immediate value to a register.

The instructions are MOVK(W), MOVZ(W) and MOVN(W), the assembly syntax is "op $(uimm16<<shift), <Rd>". The <uimm16>
is the 16-bit unsigned immediate, in the range 0 to 65535; For the 32-bit variant, the <shift> is 0 or 16, for the
64-bit variant, the <shift> is 0, 16, 32 or 48.

The current Go assembler does not accept zero shifts, such as "op $0, Rd" and "op $(0<<(16|32|48)), Rd" instructions.

Examples:

	MOVK $(10<<32), R20     <=>      movk x20, #10, lsl #32
	MOVZW $(20<<16), R8     <=>      movz w8, #20, lsl #16
	MOVK $(0<<16), R10 will be reported as an error by the assembler.

12. Some register conversion instructions have a two-character suffix, the suffix of the source register plus the suffix
of the destination register. Mapping rules:

	Source register type    Destination register type    Suffix
	Half-precision          Half-precision               H
	Single-precision        Single-precision             S
	Double-precision        Double-precision             D
	Half-precision          16-bit                       HH
	Half-precision          32-bit                       HW
	Half-precision          64-bit                       H
	Single-precision        32-bit                       SW
	Single-precision        64-bit                       S
	Double-precision        32-bit                       DW
	Double-precision        64-bit                       D
	16-bit                  Half-precision               HH
	32-bit                  Half-precision               WH
	32-bit                  Single-precision             WS
	32-bit                  Double-precision             WD
	64-bit                  Half-precision               H
	64-bit                  Single-precision             S
	64-bit                  Double-precision             D
	Half-precision          Single-precision             HS
	Half-precision          Double-precision             HD
	Single-precision        Half-precision               SH
	Single-precision        Double-precision             SD
	Double-precision        Half-precision               DH
	Double-precision        Double-precision             DS

Example:

	SCVTFWS R3, F16                   <=>      scvtf s17, w6
	VSCVTFH F1, F2                    <=>      scvtf h2, h1

Special Cases.

(1) umov is written as VMOV.

(2) br is renamed JMP, blr is renamed CALL.

(3) No need to add "W" suffix: LDARB, LDARH, LDAXRB, LDAXRH, LDTRH, LDXRB, LDXRH.

(4) In Go assembly syntax, NOP is a zero-width pseudo-instruction serves generic purpose, nothing
related to real ARM64 instruction. NOOP serves for the hardware nop instruction. NOOP is an alias of
HINT $0.

(5) Atomic instructions ldset* are renamed as LDOR*.

(6) Add a "D" and "W" suffix for 64-bit and 32-bit atomic instructions respectively.

(7) Add a "W" suffix to 32-bit while{cond}, CTERMEQ and CTERMNE instructions.

(8) MOVW, MOVH, MOVB are alias of sxtw, sxth and sxtb. MOVWU is 32-bit mov instruction,
UXTW, MOVHU and MOVBU are alias of ubfx. UXTHW and UXTBW correspond to uxth and uxtb.
Note that Go also supports instructions such as SXTW, SXTH, SXTHW, UXTH, UXTHW etc.

Examples:

		VMOV V13.B[1], R20      <=>      mov x20, v13.b[1]
		VMOV V13.H[1], R20      <=>      mov w20, v13.h[1]
		JMP (R3)                <=>      br x3
		CALL (R17)              <=>      blr x17
		LDAXRB (R19), R16       <=>      ldaxrb w16, [x19]
		NOOP                    <=>      nop
		LDORD R5, (R6), ZR      <=>      ldset x5, xzr, [x6]
		MOVW R1, R2             <=>      sxtw x2, w1
	    MOVWU R1, R2            <=>      mov w2, w1
		MOVH R1, R2             <=>      sxth x2, w1
		SXTHW R1, R2            <=>      sxth w2, w1
		MOVHU R1, R2            <=>      ubfx x2, x1, #0, #16
		UXTHW R1, R2            <=>      uxth w2, w1

(9) No need to add a "D" suffix: VADD/VSUB <Vm>, <Vn>, <Vd>

(10) FMOVS <Fd>, <Rn>  <=> fmov <Wd>, <Sn>

(11) FMOVS <Rn>, <Fd>  <=> fmov <Sd>, <Wn>

# Register mapping rules

1. All basic register names are written as Rn.

2. Go uses ZR as the zero register and RSP as the stack pointer.

3. For vector SIMD&FP instructions, floating-pointer register Vn is written as Vn, such as V1.
For non-vector SIMD&FP instructions, Vn is written as Fn, such as F2.

4. SVE and SME Z registers are written as Zn, P registers are written as Pn.

Special Cases.

(1) VADD/VSUB V1, V2, V3    <=>    add/sub d1, d2, d3

(2) SHA1H V1, V2            <=>    sha1h s2, s1

# Argument mapping rules

1. The operands appear in left-to-right assignment order.

Go reverses the arguments of most instructions.

Examples:

	ADD R11.SXTB<<1, RSP, R25      <=>      add x25, sp, w11, sxtb #1
	VADD V16, V19, V14             <=>      add d14, d19, d16

2. Treat the optional pattern specifier <pattern> as a separate operand.
The <pattern> is the optional pattern specifier, defaulting to ALL.
It can be POW2, VL1, VL2, VL3, VL4, VL6, VL7, VL8, VL16, VL32, VL64, VL128, VL256, MUL4, MUL3 or ALL.

The gnu argument "<Xdn>{, <pattern>{, mul #<imm>}}" mapping to go syntax is "{<pattern>{*$<imm>}, }<Rdn>".
It can be two operands "<pattern>{*$<imm>}, <Rdn>" or one operand "<Rdn>"

Examples:

	INCH VL1, R1              <=>    inch x1, vl1
	INCH ALL*$2, R1           <=>    inch x1, all, mul #2
	INCB R1                   <=>    incb, x1
	ZUQDECD VL1*$12, Z2.D     <=>    uqdecd z2.d, vl1, mul #12
	PTRUE VL1, P1.B           <=>    ptrue p1.b, vl1
	PTRUE P1.B                <=>    ptrue p1.b

Special Cases.

(1) Argument order is the same as in the GNU ARM64 syntax: cbz, cbnz and some store instructions,
such as str, stur, strb, sturb, strh, sturh stlr, stlrb. stlrh, st1.

Examples:

	MOVD R29, 384(R19)    <=>    str x29, [x19,#384]
	MOVB.P R30, 30(R4)    <=>    strb w30, [x4],#30
	STLRH R21, (R19)      <=>    stlrh w21, [x19]

(2) MADD, MADDW, MSUB, MSUBW, SMADDL, SMSUBL, UMADDL, UMSUBL <Rm>, <Ra>, <Rn>, <Rd>

Examples:

	MADD R2, R30, R22, R6       <=>    madd x6, x22, x2, x30
	SMSUBL R10, R3, R17, R27    <=>    smsubl x27, w17, w10, x3

(3) FMADDD, FMADDS, FMSUBD, FMSUBS, FNMADDD, FNMADDS, FNMSUBD, FNMSUBS <Fm>, <Fa>, <Fn>, <Fd>

Examples:

	FMADDD F30, F20, F3, F29    <=>    fmadd d29, d3, d30, d20
	FNMSUBS F7, F25, F7, F22    <=>    fnmsub s22, s7, s7, s25

(4) BFI, BFM, BFXIL, SBFM, SBFIZ, SBFX, UBFM, UBFIZ, UBFX $<lsb>, <Rn>, $<width>, <Rd>

Examples:

	BFIW $16, R20, $6, R0      <=>    bfi w0, w20, #16, #6
	UBFIZ $34, R26, $5, R20    <=>    ubfiz x20, x26, #34, #5

(5) FCCMPD, FCCMPS, FCCMPED, FCCMPES <cond>, Fm. Fn, $<nzcv>

Examples:

	FCCMPD AL, F8, F26, $0     <=>    fccmp d26, d8, #0x0, al
	FCCMPS VS, F29, F4, $4     <=>    fccmp s4, s29, #0x4, vs
	FCCMPED LE, F20, F5, $13   <=>    fccmpe d5, d20, #0xd, le
	FCCMPES NE, F26, F10, $0   <=>    fccmpe s10, s26, #0x0, ne

(6) CCMN, CCMNW, CCMP, CCMPW <cond>, <Rn>, $<imm>, $<nzcv>

Examples:

	CCMP MI, R22, $12, $13     <=>    ccmp x22, #0xc, #0xd, mi
	CCMNW AL, R1, $11, $8      <=>    ccmn w1, #0xb, #0x8, al

(7) CCMN, CCMNW, CCMP, CCMPW <cond>, <Rn>, <Rm>, $<nzcv>

Examples:

	CCMN VS, R13, R22, $10     <=>    ccmn x13, x22, #0xa, vs
	CCMPW HS, R19, R14, $11    <=>    ccmp w19, w14, #0xb, cs

(9) CSEL, CSELW, CSNEG, CSNEGW, CSINC, CSINCW, CSINV, CSINVW <cond>, <Rn>, <Rm>, <Rd> ;
FCSELD, FCSELS <cond>, <Fn>, <Fm>, <Fd>

Examples:

	CSEL GT, R0, R19, R1        <=>    csel x1, x0, x19, gt
	CSNEGW GT, R7, R17, R8      <=>    csneg w8, w7, w17, gt
	FCSELD EQ, F15, F18, F16    <=>    fcsel d16, d15, d18, eq

(10) TBNZ, TBZ $<imm>, <Rt>, <label>

(11) STLXR, STLXRW, STXR, STXRW, STLXRB, STLXRH, STXRB, STXRH  <Rf>, (<Rn|RSP>), <Rs>

Examples:

	STLXR ZR, (R15), R16    <=>    stlxr w16, xzr, [x15]
	STXRB R9, (R21), R19    <=>    stxrb w19, w9, [x21]

(12) STLXP, STLXPW, STXP, STXPW (<Rf1>, <Rf2>), (<Rn|RSP>), <Rs>

Examples:

	STLXP (R17, R19), (R4), R5      <=>    stlxp w5, x17, x19, [x4]
	STXPW (R30, R25), (R22), R13    <=>    stxp w13, w30, w25, [x22]

(13) Atomic instructions with 3 operands, ST64BV and ST64BV0.

Examples:

	SWPAD R5, (R6), R7  <=> swpa x5, x7, [x6]
	CASW R5, (R7), R6   <=> cas w5, w6, [x7]

(14) MOVI, MVNI $<imm8>, MSL $<amount>, <Vd>.<T>

Examples:

	MOVI $255, MSL $8, V1.S2 <=> movi v1.2s, #255, MSL #8

2. Expressions for special arguments.

#<immediate> is written as $<immediate>.

Optionally-shifted immediate.

Examples:

	ADD $(3151<<12), R14, R20     <=>    add x20, x14, #0xc4f, lsl #12
	ADDW $1864, R25, R6           <=>    add w6, w25, #0x748

Optionally-shifted registers are written as <Rm>{<shift><amount>}.
The <shift> can be <<(lsl), >>(lsr), ->(asr), @>(ror).

Examples:

	ADD R19>>30, R10, R24     <=>    add x24, x10, x19, lsr #30
	ADDW R26->24, R21, R15    <=>    add w15, w21, w26, asr #24

Extended registers are written as <Rm>{.<extend>{<<<amount>}}.
<extend> can be UXTB, UXTH, UXTW, UXTX, SXTB, SXTH, SXTW or SXTX.

Examples:

	ADDS R19.UXTB<<4, R9, R26     <=>    adds x26, x9, w19, uxtb #4
	ADDSW R14.SXTX, R14, R6       <=>    adds w6, w14, w14, sxtx

Memory references: [<Xn|SP>{,#0}] is written as (Rn|RSP), a base register and an immediate
offset is written as imm(Rn|RSP), a base register and an offset register is written as (Rn|RSP)(Rm).

Examples:

	LDAR (R22), R9                  <=>    ldar x9, [x22]
	LDP 28(R17), (R15, R23)         <=>    ldp x15, x23, [x17,#28]
	MOVWU (R4)(R12<<2), R8          <=>    ldr w8, [x4, x12, lsl #2]
	MOVD (R7)(R11.UXTW<<3), R25     <=>    ldr x25, [x7,w11,uxtw #3]
	MOVBU (R27)(R23), R14           <=>    ldrb w14, [x27,x23]

Register pairs are written as (Rt1, Rt2).

Examples:

	LDP.P -240(R11), (R12, R26)    <=>    ldp x12, x26, [x11],#-240

Register with arrangement and register with arrangement and index.

Examples:

	VADD V5.H8, V18.H8, V9.H8                     <=>    add v9.8h, v18.8h, v5.8h
	VLD1 (R2), [V21.B16]                          <=>    ld1 {v21.16b}, [x2]
	VST1.P V9.S[1], (R16)(R21)                    <=>    st1 {v9.s}[1], [x16], x28
	VST1.P [V13.H8, V14.H8, V15.H8], (R3)(R14)    <=>    st1 {v13.8h-v15.8h}, [x3], x14
	VST1.P [V14.D1, V15.D1], (R7)(R23)            <=>    st1 {v14.1d, v15.1d}, [x7], x23
*/
package arm64
