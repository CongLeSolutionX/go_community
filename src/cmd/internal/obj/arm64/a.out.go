// cmd/7c/7.out.h  from Vita Nuova.
// https://code.google.com/p/ken-cc/source/browse/src/cmd/7c/7.out.h
//
// 	Copyright © 1994-1999 Lucent Technologies Inc. All rights reserved.
// 	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
// 	Portions Copyright © 1997-1999 Vita Nuova Limited
// 	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
// 	Portions Copyright © 2004,2006 Bruce Ellis
// 	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
// 	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
// 	Portions Copyright © 2009 The Go Authors. All rights reserved.
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

package arm64

import "cmd/internal/obj"

const (
	NSNAME = 8
	NSYM   = 50
	NREG   = 32 /* number of general registers */
	NFREG  = 32 /* number of floating point registers */
)

// General purpose registers, kept in the low bits of Prog.Reg.
const (
	// integer
	REG_R0 = obj.RBaseARM64 + iota
	REG_R1
	REG_R2
	REG_R3
	REG_R4
	REG_R5
	REG_R6
	REG_R7
	REG_R8
	REG_R9
	REG_R10
	REG_R11
	REG_R12
	REG_R13
	REG_R14
	REG_R15
	REG_R16
	REG_R17
	REG_R18
	REG_R19
	REG_R20
	REG_R21
	REG_R22
	REG_R23
	REG_R24
	REG_R25
	REG_R26
	REG_R27
	REG_R28
	REG_R29
	REG_R30
	REG_R31

	// scalar floating point
	REG_F0
	REG_F1
	REG_F2
	REG_F3
	REG_F4
	REG_F5
	REG_F6
	REG_F7
	REG_F8
	REG_F9
	REG_F10
	REG_F11
	REG_F12
	REG_F13
	REG_F14
	REG_F15
	REG_F16
	REG_F17
	REG_F18
	REG_F19
	REG_F20
	REG_F21
	REG_F22
	REG_F23
	REG_F24
	REG_F25
	REG_F26
	REG_F27
	REG_F28
	REG_F29
	REG_F30
	REG_F31

	// SIMD
	REG_V0
	REG_V1
	REG_V2
	REG_V3
	REG_V4
	REG_V5
	REG_V6
	REG_V7
	REG_V8
	REG_V9
	REG_V10
	REG_V11
	REG_V12
	REG_V13
	REG_V14
	REG_V15
	REG_V16
	REG_V17
	REG_V18
	REG_V19
	REG_V20
	REG_V21
	REG_V22
	REG_V23
	REG_V24
	REG_V25
	REG_V26
	REG_V27
	REG_V28
	REG_V29
	REG_V30
	REG_V31

	// SVE scalable vector registers
	REG_Z0
	REG_Z1
	REG_Z2
	REG_Z3
	REG_Z4
	REG_Z5
	REG_Z6
	REG_Z7
	REG_Z8
	REG_Z9
	REG_Z10
	REG_Z11
	REG_Z12
	REG_Z13
	REG_Z14
	REG_Z15
	REG_Z16
	REG_Z17
	REG_Z18
	REG_Z19
	REG_Z20
	REG_Z21
	REG_Z22
	REG_Z23
	REG_Z24
	REG_Z25
	REG_Z26
	REG_Z27
	REG_Z28
	REG_Z29
	REG_Z30
	REG_Z31

	// SME ZA tile name
	REG_ZA0
	REG_ZA1
	REG_ZA2
	REG_ZA3
	REG_ZA4
	REG_ZA5
	REG_ZA6
	REG_ZA7

	// SVE scalable predicate registers
	REG_P0
	REG_P1
	REG_P2
	REG_P3
	REG_P4
	REG_P5
	REG_P6
	REG_P7
	REG_P8
	REG_P9
	REG_P10
	REG_P11
	REG_P12
	REG_P13
	REG_P14
	REG_P15

	// SVE scalable predicate registers, with predicate-as-counter encoding.
	// These are actually P registers, but encoded differently.
	// In order to distinguish with P registers, define them as PN registers.
	REG_PN0
	REG_PN1
	REG_PN2
	REG_PN3
	REG_PN4
	REG_PN5
	REG_PN6
	REG_PN7
	REG_PN8
	REG_PN9
	REG_PN10
	REG_PN11
	REG_PN12
	REG_PN13
	REG_PN14
	REG_PN15

	REG_ZT0

	REG_RSP = REG_ZT0 + 23 // to differentiate ZR/SP, REG_RSP&0x1f = 31
)

// All register types
const (
	RTYP_NORMAL     = iota // Rn, Vn.B
	RTYP_ARNG_INDEX        // Vn.<T>[index]
	RTYP_REG_INDEX         // Rn[index]
	RTYP_EXT_UXTB          // Rn.UXTB<<num
	RTYP_EXT_UXTH          // Rn.UXTH<<num
	RTYP_EXT_UXTW          // Rn.UXTW<<num
	RTYP_EXT_UXTX          // Rn.UXTX<<num
	RTYP_EXT_SXTB          // Rn.SXTB<<num
	RTYP_EXT_SXTH          // Rn.SXTH<<num
	RTYP_EXT_SXTW          // Rn.SXTW<<num
	RTYP_EXT_SXTX          // Rn.SXTX<<num
	RTYP_EXT_LSL           // Rm<<num (extension type)
	RTYP_MEM_ROFF          // (Rn)(Rm)
	RTYP_MEM_IMMEXT        // (const*VL)(Rn)
	RTYP_SVE_PM            // Pg/M
	RTYP_SVE_PZ            // Pg/Z
)

// All shift type
const (
	SHIFT_LL  = iota // <<, logic shift left
	SHIFT_LR         // >>, logic shift right
	SHIFT_AR         // ->, arithmetic shift right
	SHIFT_ROR        // @>, rotate shift right
)

// Arrangement types
const (
	ARNG_NONE = iota // no arrangement: Rn/Vn
	ARNG_8B
	ARNG_16B
	ARNG_1D
	ARNG_4H
	ARNG_8H
	ARNG_2S
	ARNG_4S
	ARNG_2D
	ARNG_1Q
	ARNG_B
	ARNG_H
	ARNG_S
	ARNG_D
	ARNG_Q
)

// Special registers, after subtracting obj.RBaseARM64, bit 12 indicates
// a special register and the low bits select the register.
// SYSREG_END is the last item in the automatically generated system register
// declaration, and it is defined in the sysRegEnc.go file.
// Define the special register after REG_SPECIAL, the first value of it should be
// REG_{name} = SYSREG_END + iota.
const (
	REG_SPECIAL = obj.RBaseARM64 + 1<<12
)

// Register assignments:
//
// compiler allocates R0 up as temps
// compiler allocates register variables R7-R25
// compiler allocates external registers R26 down
//
// compiler allocates register variables F7-F26
// compiler allocates external registers F26 down
const (
	REGMIN = REG_R7  // register variables allocated from here to REGMAX
	REGRT1 = REG_R16 // ARM64 IP0, external linker may use as a scratch register in trampoline
	REGRT2 = REG_R17 // ARM64 IP1, external linker may use as a scratch register in trampoline
	REGPR  = REG_R18 // ARM64 platform register, unused in the Go toolchain
	REGMAX = REG_R25

	REGCTXT = REG_R26 // environment for closures
	REGTMP  = REG_R27 // reserved for liblink
	REGG    = REG_R28 // G
	REGFP   = REG_R29 // frame pointer
	REGLINK = REG_R30

	// ARM64 uses R31 as both stack pointer and zero register,
	// depending on the instruction. To differentiate RSP from ZR,
	// we use a different numeric value for REGZERO and REGSP.
	REGZERO = REG_R31
	REGSP   = REG_RSP

	FREGRET = REG_F0
	FREGMIN = REG_F7  // first register variable
	FREGMAX = REG_F26 // last register variable for 7g only
	FREGEXT = REG_F26 // first external register
)

// http://infocenter.arm.com/help/topic/com.arm.doc.ecm0665627/abi_sve_aadwarf_100985_0000_00_en.pdf
var ARM64DWARFRegisters = map[int16]int16{
	REG_R0:  0,
	REG_R1:  1,
	REG_R2:  2,
	REG_R3:  3,
	REG_R4:  4,
	REG_R5:  5,
	REG_R6:  6,
	REG_R7:  7,
	REG_R8:  8,
	REG_R9:  9,
	REG_R10: 10,
	REG_R11: 11,
	REG_R12: 12,
	REG_R13: 13,
	REG_R14: 14,
	REG_R15: 15,
	REG_R16: 16,
	REG_R17: 17,
	REG_R18: 18,
	REG_R19: 19,
	REG_R20: 20,
	REG_R21: 21,
	REG_R22: 22,
	REG_R23: 23,
	REG_R24: 24,
	REG_R25: 25,
	REG_R26: 26,
	REG_R27: 27,
	REG_R28: 28,
	REG_R29: 29,
	REG_R30: 30,

	// SVE predicate registers
	REG_P0:  48,
	REG_P1:  49,
	REG_P2:  50,
	REG_P3:  51,
	REG_P4:  52,
	REG_P5:  53,
	REG_P6:  54,
	REG_P7:  55,
	REG_P8:  56,
	REG_P9:  57,
	REG_P10: 58,
	REG_P11: 59,
	REG_P12: 60,
	REG_P13: 61,
	REG_P14: 62,
	REG_P15: 63,

	// floating point
	REG_F0:  64,
	REG_F1:  65,
	REG_F2:  66,
	REG_F3:  67,
	REG_F4:  68,
	REG_F5:  69,
	REG_F6:  70,
	REG_F7:  71,
	REG_F8:  72,
	REG_F9:  73,
	REG_F10: 74,
	REG_F11: 75,
	REG_F12: 76,
	REG_F13: 77,
	REG_F14: 78,
	REG_F15: 79,
	REG_F16: 80,
	REG_F17: 81,
	REG_F18: 82,
	REG_F19: 83,
	REG_F20: 84,
	REG_F21: 85,
	REG_F22: 86,
	REG_F23: 87,
	REG_F24: 88,
	REG_F25: 89,
	REG_F26: 90,
	REG_F27: 91,
	REG_F28: 92,
	REG_F29: 93,
	REG_F30: 94,
	REG_F31: 95,

	// SIMD
	REG_V0:  64,
	REG_V1:  65,
	REG_V2:  66,
	REG_V3:  67,
	REG_V4:  68,
	REG_V5:  69,
	REG_V6:  70,
	REG_V7:  71,
	REG_V8:  72,
	REG_V9:  73,
	REG_V10: 74,
	REG_V11: 75,
	REG_V12: 76,
	REG_V13: 77,
	REG_V14: 78,
	REG_V15: 79,
	REG_V16: 80,
	REG_V17: 81,
	REG_V18: 82,
	REG_V19: 83,
	REG_V20: 84,
	REG_V21: 85,
	REG_V22: 86,
	REG_V23: 87,
	REG_V24: 88,
	REG_V25: 89,
	REG_V26: 90,
	REG_V27: 91,
	REG_V28: 92,
	REG_V29: 93,
	REG_V30: 94,
	REG_V31: 95,

	// SVE vector registers
	REG_Z0:  96,
	REG_Z1:  97,
	REG_Z2:  98,
	REG_Z3:  99,
	REG_Z4:  100,
	REG_Z5:  101,
	REG_Z6:  102,
	REG_Z7:  103,
	REG_Z8:  104,
	REG_Z9:  105,
	REG_Z10: 106,
	REG_Z11: 107,
	REG_Z12: 108,
	REG_Z13: 109,
	REG_Z14: 110,
	REG_Z15: 111,
	REG_Z16: 112,
	REG_Z17: 113,
	REG_Z18: 114,
	REG_Z19: 115,
	REG_Z20: 116,
	REG_Z21: 117,
	REG_Z22: 118,
	REG_Z23: 119,
	REG_Z24: 120,
	REG_Z25: 121,
	REG_Z26: 122,
	REG_Z27: 123,
	REG_Z28: 124,
	REG_Z29: 125,
	REG_Z30: 126,
	REG_Z31: 127,
}

const (
	BIG = 2048 - 8
)

const (
	/* mark flags */
	LABEL = 1 << iota
	LEAF
	FLOAT
	BRANCH
	LOAD
	FCMP
	SYNC
	LIST
	FOLL
	NOSCHED
)

const (
	// optab is sorted based on the order of these constants
	// and the first match is chosen.
	// The more specific class needs to come earlier.
	C_NONE   = iota
	C_REG    // R0..R30
	C_ZREG   // R0..R30, ZR
	C_RSP    // R0..R30, RSP
	C_FREG   // F0..F31
	C_VREG   // V0..V31
	C_PAIR   // (Rn, Rm)
	C_SHIFT  // Rn<<2
	C_EXTREG // Rn.UXTB[<<3]
	C_SPR    // REG_NZCV
	C_COND   // condition code, EQ, NE, etc.
	C_SPOP   // special operand, PLDL1KEEP, VMALLE1IS, etc.
	C_ARNG   // Vn.<T>
	C_ELEM   // Vn.<T>[index]
	C_LIST   // [V1, V2, V3]

	C_ZCON     // $0
	C_ABCON0   // could be C_ADDCON0 or C_BITCON
	C_ADDCON0  // 12-bit unsigned, unshifted
	C_ABCON    // could be C_ADDCON or C_BITCON
	C_AMCON    // could be C_ADDCON or C_MOVCON
	C_ADDCON   // 12-bit unsigned, shifted left by 0 or 12
	C_MBCON    // could be C_MOVCON or C_BITCON
	C_MOVCON   // generated by a 16-bit constant, optionally inverted and/or shifted by multiple of 16
	C_BITCON   // bitfield and logical immediate masks
	C_ADDCON2  // 24-bit constant
	C_LCON     // 32-bit constant
	C_MOVCON2  // a constant that can be loaded with one MOVZ/MOVN and one MOVK
	C_MOVCON3  // a constant that can be loaded with one MOVZ/MOVN and two MOVKs
	C_VCON     // 64-bit constant
	C_FCON     // floating-point constant
	C_VCONADDR // 64-bit memory address

	C_AACON  // ADDCON offset in auto constant $a(FP)
	C_AACON2 // 24-bit offset in auto constant $a(FP)
	C_LACON  // 32-bit offset in auto constant $a(FP)
	C_AECON  // ADDCON offset in extern constant $e(SB)

	// TODO(aram): only one branch class should be enough
	C_SBRA // for TYPE_BRANCH
	C_LBRA

	C_ZAUTO       // 0(RSP)
	C_NSAUTO_16   // -256 <= x < 0, 0 mod 16
	C_NSAUTO_8    // -256 <= x < 0, 0 mod 8
	C_NSAUTO_4    // -256 <= x < 0, 0 mod 4
	C_NSAUTO      // -256 <= x < 0
	C_NPAUTO_16   // -512 <= x < 0, 0 mod 16
	C_NPAUTO      // -512 <= x < 0, 0 mod 8
	C_NQAUTO_16   // -1024 <= x < 0, 0 mod 16
	C_NAUTO4K     // -4095 <= x < 0
	C_PSAUTO_16   // 0 to 255, 0 mod 16
	C_PSAUTO_8    // 0 to 255, 0 mod 8
	C_PSAUTO_4    // 0 to 255, 0 mod 4
	C_PSAUTO      // 0 to 255
	C_PPAUTO_16   // 0 to 504, 0 mod 16
	C_PPAUTO      // 0 to 504, 0 mod 8
	C_PQAUTO_16   // 0 to 1008, 0 mod 16
	C_UAUTO4K_16  // 0 to 4095, 0 mod 16
	C_UAUTO4K_8   // 0 to 4095, 0 mod 8
	C_UAUTO4K_4   // 0 to 4095, 0 mod 4
	C_UAUTO4K_2   // 0 to 4095, 0 mod 2
	C_UAUTO4K     // 0 to 4095
	C_UAUTO8K_16  // 0 to 8190, 0 mod 16
	C_UAUTO8K_8   // 0 to 8190, 0 mod 8
	C_UAUTO8K_4   // 0 to 8190, 0 mod 4
	C_UAUTO8K     // 0 to 8190, 0 mod 2  + C_PSAUTO
	C_UAUTO16K_16 // 0 to 16380, 0 mod 16
	C_UAUTO16K_8  // 0 to 16380, 0 mod 8
	C_UAUTO16K    // 0 to 16380, 0 mod 4 + C_PSAUTO
	C_UAUTO32K_16 // 0 to 32760, 0 mod 16 + C_PSAUTO
	C_UAUTO32K    // 0 to 32760, 0 mod 8 + C_PSAUTO
	C_UAUTO64K    // 0 to 65520, 0 mod 16 + C_PSAUTO
	C_LAUTO       // any other 32-bit constant

	C_SEXT1  // 0 to 4095, direct
	C_SEXT2  // 0 to 8190
	C_SEXT4  // 0 to 16380
	C_SEXT8  // 0 to 32760
	C_SEXT16 // 0 to 65520
	C_LEXT

	C_ZOREG     // 0(R)
	C_NSOREG_16 // must mirror C_NSAUTO_16, etc
	C_NSOREG_8
	C_NSOREG_4
	C_NSOREG
	C_NPOREG_16
	C_NPOREG
	C_NQOREG_16
	C_NOREG4K
	C_PSOREG_16
	C_PSOREG_8
	C_PSOREG_4
	C_PSOREG
	C_PPOREG_16
	C_PPOREG
	C_PQOREG_16
	C_UOREG4K_16
	C_UOREG4K_8
	C_UOREG4K_4
	C_UOREG4K_2
	C_UOREG4K
	C_UOREG8K_16
	C_UOREG8K_8
	C_UOREG8K_4
	C_UOREG8K
	C_UOREG16K_16
	C_UOREG16K_8
	C_UOREG16K
	C_UOREG32K_16
	C_UOREG32K
	C_UOREG64K
	C_LOREG

	C_MEMIMMEXT // (const*VL)(Rn)
	C_REGINDEX  // Rn[index]

	C_ADDR // TODO(aram): explain difference from C_VCONADDR

	// The GOT slot for a symbol in -dynlink mode.
	C_GOTADDR

	// TLS "var" in local exec mode: will become a constant offset from
	// thread local base that is ultimately chosen by the program linker.
	C_TLS_LE

	// TLS "var" in initial exec mode: will become a memory address (chosen
	// by the program linker) that the dynamic linker will fill with the
	// offset from the thread local base.
	C_TLS_IE

	C_ROFF // register offset (including register extended)

	C_GOK
	C_TEXTSIZE
	C_NCLASS // must be last
)

type oprType uint16

// The classification table below will eventually replace the classification table above.
//
//go:generate stringer -type oprType -trimprefix AC_
const (
	AC_NONE   oprType = iota
	AC_REG            // general purpose registers R0..R30 and ZR
	AC_RSP            // general purpose registers R0..R30 and RSP
	AC_FREG           // floating point registers, such as F1
	AC_VREG           // vector registers, such as V1
	AC_ZREG           // the scalable vector registers, such as Z1
	AC_ZAREG          // the name of the ZA tile, defined as registers, such as ZA0
	AC_ZTREG          // the ZT0 register
	AC_PREG           // the scalable predicate registers, such as P1
	AC_PNREG          // the scalable predicate registers, with predicate-as-counter encoding, such as PN1
	AC_PREGM          // Pg/M
	AC_PREGZ          // Pg/Z
	AC_SPR            // special register, such as REG_NZCV, system registers
	AC_REGIDX         // P8[1]
	AC_PAIR           // register pair, such as (R1, R3)
	AC_REGEXT         // general purpose register with extend, such as R7.SXTW<<1

	AC_REGSHIFT // general purpose register with shift, such as R1<<2

	AC_COND  // conditional flags, such as CS
	AC_SPOP  // special operands, such as DAIFSet
	AC_LABEL // branch labels

	AC_IMM // constants

	AC_REGLIST1  // list of 1 vector register, such as [V1]
	AC_REGLIST2  // list of 2 vector registers, such as [V1, V2]
	AC_REGLIST2C // list of 2 consecutive vector registers, such as [V1, V2]
	AC_REGLIST3  // list of 3 vector registers, such as [V1, V2, V3]
	AC_REGLIST4  // list of 4 vector registers, such as [V1, V2, V3, V4]
	AC_REGLIST4C // list of 4 consecutive vector registers, such as [V1, V2, V3, V4]
	AC_LISTIDX   // list with index, such as [V1.B, V2.B][2]

	AC_ARNG    // vector register with arrangement, such as V11.D2
	AC_ARNGIDX // vector register with arrangement and index, such as V12.D[1]

	AC_MEMIMM     // address with optional offset, the offset is an immediate, such as 4(R1)
	AC_MEMIMMEXT  // address with optional offset, the offset is an immediate with extension, such as (2*VL)(R1)
	AC_MEMEXT     // address with extend offset, such as (R2)(R5.SXTX<<1)
	AC_MEMPOSTIMM // address of the post-index class, offset is an immediate
	AC_MEMPOSTREG // address of the post-index class, offset is a register
	AC_MEMPREIMM  // address of the pre-index class, offset is an immediate

	AC_ZAHVTILEIDX // <ZAd><HV>.D[<Ws>, <offs>]
	AC_ZAHVTILESEL // <ZAd><HV>.D[<Ws>, <offsf>:<offsl>]

	AC_ZAVECTORIDX    // ZA[<Wv>, <offs>]
	AC_ZAVECTORIDXVG2 // ZA.<T>[<Wv>, <offs>{, VGx2}]
	AC_ZAVECTORIDXVG4 // ZA.<T>[<Wv>, <offs>{, VGx4}]

	AC_ZAVECTORSEL    // ZA.<T>[<Wv>, <offsf>:<offsl>]
	AC_ZAVECTORSELVG2 // ZA.<T>[<Wv>, <offsf>:<offsl>{, VGx2}]
	AC_ZAVECTORSELVG4 // ZA.<T>[<Wv>, <offsf>:<offsl>{, VGx4}]

	AC_TEXTSIZE
	AC_ANY // any other operand format
)

const (
	C_XPRE  = 1 << 6 // match arm.C_WBIT, so Prog.String know how to print it
	C_XPOST = 1 << 5 // match arm.C_PBIT, so Prog.String know how to print it
)

//go:generate stringer -type SpecialOperand -trimprefix SPOP_
type SpecialOperand int

const (
	// PRFM
	SPOP_PLDL1KEEP SpecialOperand = iota     // must be the first one
	SPOP_BEGIN     SpecialOperand = iota - 1 // set as the lower bound
	SPOP_PLDL1STRM
	SPOP_PLDL2KEEP
	SPOP_PLDL2STRM
	SPOP_PLDL3KEEP
	SPOP_PLDL3STRM
	SPOP_PLIL1KEEP
	SPOP_PLIL1STRM
	SPOP_PLIL2KEEP
	SPOP_PLIL2STRM
	SPOP_PLIL3KEEP
	SPOP_PLIL3STRM
	SPOP_PSTL1KEEP
	SPOP_PSTL1STRM
	SPOP_PSTL2KEEP
	SPOP_PSTL2STRM
	SPOP_PSTL3KEEP
	SPOP_PSTL3STRM

	// TLBI
	SPOP_VMALLE1IS
	SPOP_VAE1IS
	SPOP_ASIDE1IS
	SPOP_VAAE1IS
	SPOP_VALE1IS
	SPOP_VAALE1IS
	SPOP_VMALLE1
	SPOP_VAE1
	SPOP_ASIDE1
	SPOP_VAAE1
	SPOP_VALE1
	SPOP_VAALE1
	SPOP_IPAS2E1IS
	SPOP_IPAS2LE1IS
	SPOP_ALLE2IS
	SPOP_VAE2IS
	SPOP_ALLE1IS
	SPOP_VALE2IS
	SPOP_VMALLS12E1IS
	SPOP_IPAS2E1
	SPOP_IPAS2LE1
	SPOP_ALLE2
	SPOP_VAE2
	SPOP_ALLE1
	SPOP_VALE2
	SPOP_VMALLS12E1
	SPOP_ALLE3IS
	SPOP_VAE3IS
	SPOP_VALE3IS
	SPOP_ALLE3
	SPOP_VAE3
	SPOP_VALE3
	SPOP_VMALLE1OS
	SPOP_VAE1OS
	SPOP_ASIDE1OS
	SPOP_VAAE1OS
	SPOP_VALE1OS
	SPOP_VAALE1OS
	SPOP_RVAE1IS
	SPOP_RVAAE1IS
	SPOP_RVALE1IS
	SPOP_RVAALE1IS
	SPOP_RVAE1OS
	SPOP_RVAAE1OS
	SPOP_RVALE1OS
	SPOP_RVAALE1OS
	SPOP_RVAE1
	SPOP_RVAAE1
	SPOP_RVALE1
	SPOP_RVAALE1
	SPOP_RIPAS2E1IS
	SPOP_RIPAS2LE1IS
	SPOP_ALLE2OS
	SPOP_VAE2OS
	SPOP_ALLE1OS
	SPOP_VALE2OS
	SPOP_VMALLS12E1OS
	SPOP_RVAE2IS
	SPOP_RVALE2IS
	SPOP_IPAS2E1OS
	SPOP_RIPAS2E1
	SPOP_RIPAS2E1OS
	SPOP_IPAS2LE1OS
	SPOP_RIPAS2LE1
	SPOP_RIPAS2LE1OS
	SPOP_RVAE2OS
	SPOP_RVALE2OS
	SPOP_RVAE2
	SPOP_RVALE2
	SPOP_ALLE3OS
	SPOP_VAE3OS
	SPOP_VALE3OS
	SPOP_RVAE3IS
	SPOP_RVALE3IS
	SPOP_RVAE3OS
	SPOP_RVALE3OS
	SPOP_RVAE3
	SPOP_RVALE3

	// DC
	SPOP_IVAC
	SPOP_ISW
	SPOP_CSW
	SPOP_CISW
	SPOP_ZVA
	SPOP_CVAC
	SPOP_CVAU
	SPOP_CIVAC
	SPOP_IGVAC
	SPOP_IGSW
	SPOP_IGDVAC
	SPOP_IGDSW
	SPOP_CGSW
	SPOP_CGDSW
	SPOP_CIGSW
	SPOP_CIGDSW
	SPOP_GVA
	SPOP_GZVA
	SPOP_CGVAC
	SPOP_CGDVAC
	SPOP_CGVAP
	SPOP_CGDVAP
	SPOP_CGVADP
	SPOP_CGDVADP
	SPOP_CIGVAC
	SPOP_CIGDVAC
	SPOP_CVAP
	SPOP_CVADP

	// PSTATE fields
	SPOP_DAIFSet
	SPOP_DAIFClr

	// Condition code, EQ, NE, etc. Their relative order to EQ is matter.
	SPOP_EQ
	SPOP_NE
	SPOP_HS
	SPOP_LO
	SPOP_MI
	SPOP_PL
	SPOP_VS
	SPOP_VC
	SPOP_HI
	SPOP_LS
	SPOP_GE
	SPOP_LT
	SPOP_GT
	SPOP_LE
	SPOP_AL
	SPOP_NV
	// Condition code end.

	SPOP_END
)

// supportedInsts contains the Go instructions that we have supported now.
// Since many instructions share the same elements, when we support the encoding
// and decoding of a certain element, we may unknowingly support an instructions,
// but the format of the instruction may be unreasonable. To prevent this from
// happening, explicitly control supported instructions through this slice.
var supportedInsts = [ALAST - obj.ABaseARM64]bool{
	// Should we just list those supported ones ?
	AABS - obj.ABaseARM64:         false,
	AABSW - obj.ABaseARM64:        false,
	AADC - obj.ABaseARM64:         true,
	AADCS - obj.ABaseARM64:        true,
	AADCSW - obj.ABaseARM64:       true,
	AADCW - obj.ABaseARM64:        true,
	AADD - obj.ABaseARM64:         true,
	AADDG - obj.ABaseARM64:        false,
	AADDPL - obj.ABaseARM64:       false,
	AADDS - obj.ABaseARM64:        true,
	AADDSPL - obj.ABaseARM64:      false,
	AADDSVL - obj.ABaseARM64:      false,
	AADDSW - obj.ABaseARM64:       true,
	AADDVL - obj.ABaseARM64:       false,
	AADDW - obj.ABaseARM64:        true,
	AADR - obj.ABaseARM64:         true,
	AADRP - obj.ABaseARM64:        true,
	AAESD - obj.ABaseARM64:        true,
	AAESE - obj.ABaseARM64:        true,
	AAESIMC - obj.ABaseARM64:      true,
	AAESMC - obj.ABaseARM64:       true,
	AAND - obj.ABaseARM64:         true,
	AANDS - obj.ABaseARM64:        true,
	AANDSW - obj.ABaseARM64:       true,
	AANDW - obj.ABaseARM64:        true,
	AASR - obj.ABaseARM64:         true,
	AASRV - obj.ABaseARM64:        false,
	AASRVW - obj.ABaseARM64:       false,
	AASRW - obj.ABaseARM64:        true,
	AAT - obj.ABaseARM64:          true,
	AAUTDA - obj.ABaseARM64:       false,
	AAUTDB - obj.ABaseARM64:       false,
	AAUTDZA - obj.ABaseARM64:      false,
	AAUTDZB - obj.ABaseARM64:      false,
	AAUTIA - obj.ABaseARM64:       false,
	AAUTIA1716 - obj.ABaseARM64:   false,
	AAUTIASP - obj.ABaseARM64:     false,
	AAUTIAZ - obj.ABaseARM64:      false,
	AAUTIB - obj.ABaseARM64:       false,
	AAUTIB1716 - obj.ABaseARM64:   false,
	AAUTIBSP - obj.ABaseARM64:     false,
	AAUTIBZ - obj.ABaseARM64:      false,
	AAUTIZA - obj.ABaseARM64:      false,
	AAUTIZB - obj.ABaseARM64:      false,
	AAXFLAG - obj.ABaseARM64:      false,
	ABCC - obj.ABaseARM64:         true,
	ABCCC - obj.ABaseARM64:        false,
	ABCCS - obj.ABaseARM64:        false,
	ABCEQ - obj.ABaseARM64:        false,
	ABCGE - obj.ABaseARM64:        false,
	ABCGT - obj.ABaseARM64:        false,
	ABCHI - obj.ABaseARM64:        false,
	ABCHS - obj.ABaseARM64:        false,
	ABCLE - obj.ABaseARM64:        false,
	ABCLO - obj.ABaseARM64:        false,
	ABCLS - obj.ABaseARM64:        false,
	ABCLT - obj.ABaseARM64:        false,
	ABCMI - obj.ABaseARM64:        false,
	ABCNE - obj.ABaseARM64:        false,
	ABCPL - obj.ABaseARM64:        false,
	ABCS - obj.ABaseARM64:         true,
	ABCVC - obj.ABaseARM64:        false,
	ABCVS - obj.ABaseARM64:        false,
	ABEQ - obj.ABaseARM64:         true,
	ABFC - obj.ABaseARM64:         false,
	ABFCVTS - obj.ABaseARM64:      false,
	ABFCW - obj.ABaseARM64:        false,
	ABFI - obj.ABaseARM64:         true,
	ABFIW - obj.ABaseARM64:        true,
	ABFM - obj.ABaseARM64:         true,
	ABFMW - obj.ABaseARM64:        true,
	ABFXIL - obj.ABaseARM64:       true,
	ABFXILW - obj.ABaseARM64:      true,
	ABGE - obj.ABaseARM64:         true,
	ABGT - obj.ABaseARM64:         true,
	ABHI - obj.ABaseARM64:         true,
	ABHS - obj.ABaseARM64:         true,
	ABIC - obj.ABaseARM64:         true,
	ABICS - obj.ABaseARM64:        true,
	ABICSW - obj.ABaseARM64:       true,
	ABICW - obj.ABaseARM64:        true,
	ABLE - obj.ABaseARM64:         true,
	ABLO - obj.ABaseARM64:         true,
	ABLRAA - obj.ABaseARM64:       false,
	ABLRAAZ - obj.ABaseARM64:      false,
	ABLRAB - obj.ABaseARM64:       false,
	ABLRABZ - obj.ABaseARM64:      false,
	ABLS - obj.ABaseARM64:         true,
	ABLT - obj.ABaseARM64:         true,
	ABMI - obj.ABaseARM64:         true,
	ABNE - obj.ABaseARM64:         true,
	ABPL - obj.ABaseARM64:         true,
	ABRAA - obj.ABaseARM64:        false,
	ABRAAZ - obj.ABaseARM64:       false,
	ABRAB - obj.ABaseARM64:        false,
	ABRABZ - obj.ABaseARM64:       false,
	ABRB - obj.ABaseARM64:         false,
	ABRK - obj.ABaseARM64:         true,
	ABTI - obj.ABaseARM64:         false,
	ABVC - obj.ABaseARM64:         true,
	ABVS - obj.ABaseARM64:         true,
	ACASAB - obj.ABaseARM64:       false,
	ACASAD - obj.ABaseARM64:       true,
	ACASAH - obj.ABaseARM64:       false,
	ACASALB - obj.ABaseARM64:      true,
	ACASALD - obj.ABaseARM64:      true,
	ACASALH - obj.ABaseARM64:      true,
	ACASALW - obj.ABaseARM64:      true,
	ACASAW - obj.ABaseARM64:       true,
	ACASB - obj.ABaseARM64:        true,
	ACASD - obj.ABaseARM64:        true,
	ACASH - obj.ABaseARM64:        true,
	ACASLB - obj.ABaseARM64:       false,
	ACASLD - obj.ABaseARM64:       true,
	ACASLH - obj.ABaseARM64:       false,
	ACASLW - obj.ABaseARM64:       true,
	ACASPAD - obj.ABaseARM64:      false,
	ACASPALD - obj.ABaseARM64:     false,
	ACASPALW - obj.ABaseARM64:     false,
	ACASPAW - obj.ABaseARM64:      false,
	ACASPD - obj.ABaseARM64:       true,
	ACASPLD - obj.ABaseARM64:      false,
	ACASPLW - obj.ABaseARM64:      false,
	ACASPW - obj.ABaseARM64:       true,
	ACASW - obj.ABaseARM64:        true,
	ACBNZ - obj.ABaseARM64:        true,
	ACBNZW - obj.ABaseARM64:       true,
	ACBZ - obj.ABaseARM64:         true,
	ACBZW - obj.ABaseARM64:        true,
	ACCMN - obj.ABaseARM64:        true,
	ACCMNW - obj.ABaseARM64:       true,
	ACCMP - obj.ABaseARM64:        true,
	ACCMPW - obj.ABaseARM64:       true,
	ACFINV - obj.ABaseARM64:       false,
	ACFP - obj.ABaseARM64:         false,
	ACHKFEAT - obj.ABaseARM64:     false,
	ACINC - obj.ABaseARM64:        true,
	ACINCW - obj.ABaseARM64:       true,
	ACINV - obj.ABaseARM64:        true,
	ACINVW - obj.ABaseARM64:       true,
	ACLRBHB - obj.ABaseARM64:      false,
	ACLREX - obj.ABaseARM64:       true,
	ACLS - obj.ABaseARM64:         true,
	ACLSW - obj.ABaseARM64:        true,
	ACLZ - obj.ABaseARM64:         true,
	ACLZW - obj.ABaseARM64:        true,
	ACMN - obj.ABaseARM64:         true,
	ACMNW - obj.ABaseARM64:        true,
	ACMP - obj.ABaseARM64:         true,
	ACMPP - obj.ABaseARM64:        false,
	ACMPW - obj.ABaseARM64:        true,
	ACNEG - obj.ABaseARM64:        true,
	ACNEGW - obj.ABaseARM64:       true,
	ACNT - obj.ABaseARM64:         false,
	ACNTB - obj.ABaseARM64:        false,
	ACNTD - obj.ABaseARM64:        false,
	ACNTH - obj.ABaseARM64:        false,
	ACNTW - obj.ABaseARM64:        false,
	ACOSP - obj.ABaseARM64:        false,
	ACPP - obj.ABaseARM64:         false,
	ACPYE - obj.ABaseARM64:        false,
	ACPYEN - obj.ABaseARM64:       false,
	ACPYERN - obj.ABaseARM64:      false,
	ACPYERT - obj.ABaseARM64:      false,
	ACPYERTN - obj.ABaseARM64:     false,
	ACPYERTRN - obj.ABaseARM64:    false,
	ACPYERTWN - obj.ABaseARM64:    false,
	ACPYET - obj.ABaseARM64:       false,
	ACPYETN - obj.ABaseARM64:      false,
	ACPYETRN - obj.ABaseARM64:     false,
	ACPYETWN - obj.ABaseARM64:     false,
	ACPYEWN - obj.ABaseARM64:      false,
	ACPYEWT - obj.ABaseARM64:      false,
	ACPYEWTN - obj.ABaseARM64:     false,
	ACPYEWTRN - obj.ABaseARM64:    false,
	ACPYEWTWN - obj.ABaseARM64:    false,
	ACPYFE - obj.ABaseARM64:       false,
	ACPYFEN - obj.ABaseARM64:      false,
	ACPYFERN - obj.ABaseARM64:     false,
	ACPYFERT - obj.ABaseARM64:     false,
	ACPYFERTN - obj.ABaseARM64:    false,
	ACPYFERTRN - obj.ABaseARM64:   false,
	ACPYFERTWN - obj.ABaseARM64:   false,
	ACPYFET - obj.ABaseARM64:      false,
	ACPYFETN - obj.ABaseARM64:     false,
	ACPYFETRN - obj.ABaseARM64:    false,
	ACPYFETWN - obj.ABaseARM64:    false,
	ACPYFEWN - obj.ABaseARM64:     false,
	ACPYFEWT - obj.ABaseARM64:     false,
	ACPYFEWTN - obj.ABaseARM64:    false,
	ACPYFEWTRN - obj.ABaseARM64:   false,
	ACPYFEWTWN - obj.ABaseARM64:   false,
	ACPYFM - obj.ABaseARM64:       false,
	ACPYFMN - obj.ABaseARM64:      false,
	ACPYFMRN - obj.ABaseARM64:     false,
	ACPYFMRT - obj.ABaseARM64:     false,
	ACPYFMRTN - obj.ABaseARM64:    false,
	ACPYFMRTRN - obj.ABaseARM64:   false,
	ACPYFMRTWN - obj.ABaseARM64:   false,
	ACPYFMT - obj.ABaseARM64:      false,
	ACPYFMTN - obj.ABaseARM64:     false,
	ACPYFMTRN - obj.ABaseARM64:    false,
	ACPYFMTWN - obj.ABaseARM64:    false,
	ACPYFMWN - obj.ABaseARM64:     false,
	ACPYFMWT - obj.ABaseARM64:     false,
	ACPYFMWTN - obj.ABaseARM64:    false,
	ACPYFMWTRN - obj.ABaseARM64:   false,
	ACPYFMWTWN - obj.ABaseARM64:   false,
	ACPYFP - obj.ABaseARM64:       false,
	ACPYFPN - obj.ABaseARM64:      false,
	ACPYFPRN - obj.ABaseARM64:     false,
	ACPYFPRT - obj.ABaseARM64:     false,
	ACPYFPRTN - obj.ABaseARM64:    false,
	ACPYFPRTRN - obj.ABaseARM64:   false,
	ACPYFPRTWN - obj.ABaseARM64:   false,
	ACPYFPT - obj.ABaseARM64:      false,
	ACPYFPTN - obj.ABaseARM64:     false,
	ACPYFPTRN - obj.ABaseARM64:    false,
	ACPYFPTWN - obj.ABaseARM64:    false,
	ACPYFPWN - obj.ABaseARM64:     false,
	ACPYFPWT - obj.ABaseARM64:     false,
	ACPYFPWTN - obj.ABaseARM64:    false,
	ACPYFPWTRN - obj.ABaseARM64:   false,
	ACPYFPWTWN - obj.ABaseARM64:   false,
	ACPYM - obj.ABaseARM64:        false,
	ACPYMN - obj.ABaseARM64:       false,
	ACPYMRN - obj.ABaseARM64:      false,
	ACPYMRT - obj.ABaseARM64:      false,
	ACPYMRTN - obj.ABaseARM64:     false,
	ACPYMRTRN - obj.ABaseARM64:    false,
	ACPYMRTWN - obj.ABaseARM64:    false,
	ACPYMT - obj.ABaseARM64:       false,
	ACPYMTN - obj.ABaseARM64:      false,
	ACPYMTRN - obj.ABaseARM64:     false,
	ACPYMTWN - obj.ABaseARM64:     false,
	ACPYMWN - obj.ABaseARM64:      false,
	ACPYMWT - obj.ABaseARM64:      false,
	ACPYMWTN - obj.ABaseARM64:     false,
	ACPYMWTRN - obj.ABaseARM64:    false,
	ACPYMWTWN - obj.ABaseARM64:    false,
	ACPYP - obj.ABaseARM64:        false,
	ACPYPN - obj.ABaseARM64:       false,
	ACPYPRN - obj.ABaseARM64:      false,
	ACPYPRT - obj.ABaseARM64:      false,
	ACPYPRTN - obj.ABaseARM64:     false,
	ACPYPRTRN - obj.ABaseARM64:    false,
	ACPYPRTWN - obj.ABaseARM64:    false,
	ACPYPT - obj.ABaseARM64:       false,
	ACPYPTN - obj.ABaseARM64:      false,
	ACPYPTRN - obj.ABaseARM64:     false,
	ACPYPTWN - obj.ABaseARM64:     false,
	ACPYPWN - obj.ABaseARM64:      false,
	ACPYPWT - obj.ABaseARM64:      false,
	ACPYPWTN - obj.ABaseARM64:     false,
	ACPYPWTRN - obj.ABaseARM64:    false,
	ACPYPWTWN - obj.ABaseARM64:    false,
	ACRC32B - obj.ABaseARM64:      true,
	ACRC32CB - obj.ABaseARM64:     true,
	ACRC32CH - obj.ABaseARM64:     true,
	ACRC32CW - obj.ABaseARM64:     true,
	ACRC32CX - obj.ABaseARM64:     true,
	ACRC32H - obj.ABaseARM64:      true,
	ACRC32W - obj.ABaseARM64:      true,
	ACRC32X - obj.ABaseARM64:      true,
	ACSDB - obj.ABaseARM64:        false,
	ACSEL - obj.ABaseARM64:        true,
	ACSELW - obj.ABaseARM64:       true,
	ACSET - obj.ABaseARM64:        true,
	ACSETM - obj.ABaseARM64:       true,
	ACSETMW - obj.ABaseARM64:      true,
	ACSETW - obj.ABaseARM64:       true,
	ACSINC - obj.ABaseARM64:       true,
	ACSINCW - obj.ABaseARM64:      true,
	ACSINV - obj.ABaseARM64:       true,
	ACSINVW - obj.ABaseARM64:      true,
	ACSNEG - obj.ABaseARM64:       true,
	ACSNEGW - obj.ABaseARM64:      true,
	ACTERMEQ - obj.ABaseARM64:     false,
	ACTERMEQW - obj.ABaseARM64:    false,
	ACTERMNE - obj.ABaseARM64:     false,
	ACTERMNEW - obj.ABaseARM64:    false,
	ACTZ - obj.ABaseARM64:         false,
	ACTZW - obj.ABaseARM64:        false,
	ADC - obj.ABaseARM64:          true,
	ADCPS1 - obj.ABaseARM64:       true,
	ADCPS2 - obj.ABaseARM64:       true,
	ADCPS3 - obj.ABaseARM64:       true,
	ADECB - obj.ABaseARM64:        false,
	ADECD - obj.ABaseARM64:        false,
	ADECH - obj.ABaseARM64:        false,
	ADECW - obj.ABaseARM64:        false,
	ADGH - obj.ABaseARM64:         false,
	ADMB - obj.ABaseARM64:         true,
	ADRPS - obj.ABaseARM64:        true,
	ADSB - obj.ABaseARM64:         true,
	ADVP - obj.ABaseARM64:         false,
	ADWORD - obj.ABaseARM64:       true,
	AEON - obj.ABaseARM64:         true,
	AEONW - obj.ABaseARM64:        true,
	AEOR - obj.ABaseARM64:         true,
	AEORW - obj.ABaseARM64:        true,
	AERET - obj.ABaseARM64:        true,
	AERETAA - obj.ABaseARM64:      false,
	AERETAB - obj.ABaseARM64:      false,
	AESB - obj.ABaseARM64:         false,
	AEXTR - obj.ABaseARM64:        true,
	AEXTRW - obj.ABaseARM64:       true,
	AFABSD - obj.ABaseARM64:       true,
	AFABSH - obj.ABaseARM64:       false,
	AFABSS - obj.ABaseARM64:       true,
	AFADDD - obj.ABaseARM64:       true,
	AFADDH - obj.ABaseARM64:       false,
	AFADDS - obj.ABaseARM64:       true,
	AFCCMPD - obj.ABaseARM64:      true,
	AFCCMPED - obj.ABaseARM64:     true,
	AFCCMPEH - obj.ABaseARM64:     false,
	AFCCMPES - obj.ABaseARM64:     true,
	AFCCMPH - obj.ABaseARM64:      false,
	AFCCMPS - obj.ABaseARM64:      true,
	AFCMPD - obj.ABaseARM64:       true,
	AFCMPED - obj.ABaseARM64:      true,
	AFCMPEH - obj.ABaseARM64:      false,
	AFCMPES - obj.ABaseARM64:      true,
	AFCMPH - obj.ABaseARM64:       false,
	AFCMPS - obj.ABaseARM64:       true,
	AFCSELD - obj.ABaseARM64:      true,
	AFCSELH - obj.ABaseARM64:      false,
	AFCSELS - obj.ABaseARM64:      true,
	AFCVTASD - obj.ABaseARM64:     false,
	AFCVTASDW - obj.ABaseARM64:    false,
	AFCVTASH - obj.ABaseARM64:     false,
	AFCVTASHW - obj.ABaseARM64:    false,
	AFCVTASS - obj.ABaseARM64:     false,
	AFCVTASSW - obj.ABaseARM64:    false,
	AFCVTAUD - obj.ABaseARM64:     false,
	AFCVTAUDW - obj.ABaseARM64:    false,
	AFCVTAUH - obj.ABaseARM64:     false,
	AFCVTAUHW - obj.ABaseARM64:    false,
	AFCVTAUS - obj.ABaseARM64:     false,
	AFCVTAUSW - obj.ABaseARM64:    false,
	AFCVTDH - obj.ABaseARM64:      true,
	AFCVTDS - obj.ABaseARM64:      true,
	AFCVTHD - obj.ABaseARM64:      true,
	AFCVTHS - obj.ABaseARM64:      true,
	AFCVTMSD - obj.ABaseARM64:     false,
	AFCVTMSDW - obj.ABaseARM64:    false,
	AFCVTMSH - obj.ABaseARM64:     false,
	AFCVTMSHW - obj.ABaseARM64:    false,
	AFCVTMSS - obj.ABaseARM64:     false,
	AFCVTMSSW - obj.ABaseARM64:    false,
	AFCVTMUD - obj.ABaseARM64:     false,
	AFCVTMUDW - obj.ABaseARM64:    false,
	AFCVTMUH - obj.ABaseARM64:     false,
	AFCVTMUHW - obj.ABaseARM64:    false,
	AFCVTMUS - obj.ABaseARM64:     false,
	AFCVTMUSW - obj.ABaseARM64:    false,
	AFCVTNSD - obj.ABaseARM64:     false,
	AFCVTNSDW - obj.ABaseARM64:    false,
	AFCVTNSH - obj.ABaseARM64:     false,
	AFCVTNSHW - obj.ABaseARM64:    false,
	AFCVTNSS - obj.ABaseARM64:     false,
	AFCVTNSSW - obj.ABaseARM64:    false,
	AFCVTNUD - obj.ABaseARM64:     false,
	AFCVTNUDW - obj.ABaseARM64:    false,
	AFCVTNUH - obj.ABaseARM64:     false,
	AFCVTNUHW - obj.ABaseARM64:    false,
	AFCVTNUS - obj.ABaseARM64:     false,
	AFCVTNUSW - obj.ABaseARM64:    false,
	AFCVTPSD - obj.ABaseARM64:     false,
	AFCVTPSDW - obj.ABaseARM64:    false,
	AFCVTPSH - obj.ABaseARM64:     false,
	AFCVTPSHW - obj.ABaseARM64:    false,
	AFCVTPSS - obj.ABaseARM64:     false,
	AFCVTPSSW - obj.ABaseARM64:    false,
	AFCVTPUD - obj.ABaseARM64:     false,
	AFCVTPUDW - obj.ABaseARM64:    false,
	AFCVTPUH - obj.ABaseARM64:     false,
	AFCVTPUHW - obj.ABaseARM64:    false,
	AFCVTPUS - obj.ABaseARM64:     false,
	AFCVTPUSW - obj.ABaseARM64:    false,
	AFCVTSD - obj.ABaseARM64:      true,
	AFCVTSH - obj.ABaseARM64:      true,
	AFCVTZSD - obj.ABaseARM64:     true,
	AFCVTZSDW - obj.ABaseARM64:    true,
	AFCVTZSH - obj.ABaseARM64:     false,
	AFCVTZSHW - obj.ABaseARM64:    false,
	AFCVTZSS - obj.ABaseARM64:     true,
	AFCVTZSSW - obj.ABaseARM64:    true,
	AFCVTZUD - obj.ABaseARM64:     true,
	AFCVTZUDW - obj.ABaseARM64:    true,
	AFCVTZUH - obj.ABaseARM64:     false,
	AFCVTZUHW - obj.ABaseARM64:    false,
	AFCVTZUS - obj.ABaseARM64:     true,
	AFCVTZUSW - obj.ABaseARM64:    true,
	AFDIVD - obj.ABaseARM64:       true,
	AFDIVH - obj.ABaseARM64:       false,
	AFDIVS - obj.ABaseARM64:       true,
	AFJCVTZSDW - obj.ABaseARM64:   false,
	AFLDAPURB - obj.ABaseARM64:    false,
	AFLDAPURD - obj.ABaseARM64:    false,
	AFLDAPURH - obj.ABaseARM64:    false,
	AFLDAPURQ - obj.ABaseARM64:    false,
	AFLDAPURS - obj.ABaseARM64:    false,
	AFLDNPD - obj.ABaseARM64:      false,
	AFLDNPQ - obj.ABaseARM64:      false,
	AFLDNPS - obj.ABaseARM64:      false,
	AFLDPD - obj.ABaseARM64:       true,
	AFLDPQ - obj.ABaseARM64:       true,
	AFLDPS - obj.ABaseARM64:       true,
	AFMADDD - obj.ABaseARM64:      true,
	AFMADDH - obj.ABaseARM64:      false,
	AFMADDS - obj.ABaseARM64:      true,
	AFMAXD - obj.ABaseARM64:       true,
	AFMAXH - obj.ABaseARM64:       false,
	AFMAXNMD - obj.ABaseARM64:     true,
	AFMAXNMH - obj.ABaseARM64:     false,
	AFMAXNMS - obj.ABaseARM64:     true,
	AFMAXS - obj.ABaseARM64:       true,
	AFMIND - obj.ABaseARM64:       true,
	AFMINH - obj.ABaseARM64:       false,
	AFMINNMD - obj.ABaseARM64:     true,
	AFMINNMH - obj.ABaseARM64:     false,
	AFMINNMS - obj.ABaseARM64:     true,
	AFMINS - obj.ABaseARM64:       true,
	AFMOV - obj.ABaseARM64:        false,
	AFMOVB - obj.ABaseARM64:       false,
	AFMOVD - obj.ABaseARM64:       true,
	AFMOVH - obj.ABaseARM64:       false,
	AFMOVHW - obj.ABaseARM64:      false,
	AFMOVQ - obj.ABaseARM64:       true,
	AFMOVS - obj.ABaseARM64:       true,
	AFMOVSW - obj.ABaseARM64:      false,
	AFMOVWH - obj.ABaseARM64:      false,
	AFMOVWS - obj.ABaseARM64:      false,
	AFMSUBD - obj.ABaseARM64:      true,
	AFMSUBH - obj.ABaseARM64:      false,
	AFMSUBS - obj.ABaseARM64:      true,
	AFMULD - obj.ABaseARM64:       true,
	AFMULH - obj.ABaseARM64:       false,
	AFMULS - obj.ABaseARM64:       true,
	AFNEGD - obj.ABaseARM64:       true,
	AFNEGH - obj.ABaseARM64:       false,
	AFNEGS - obj.ABaseARM64:       true,
	AFNMADDD - obj.ABaseARM64:     true,
	AFNMADDH - obj.ABaseARM64:     false,
	AFNMADDS - obj.ABaseARM64:     true,
	AFNMSUBD - obj.ABaseARM64:     true,
	AFNMSUBH - obj.ABaseARM64:     false,
	AFNMSUBS - obj.ABaseARM64:     true,
	AFNMULD - obj.ABaseARM64:      true,
	AFNMULH - obj.ABaseARM64:      false,
	AFNMULS - obj.ABaseARM64:      true,
	AFRINT32XD - obj.ABaseARM64:   false,
	AFRINT32XS - obj.ABaseARM64:   false,
	AFRINT32ZD - obj.ABaseARM64:   false,
	AFRINT32ZS - obj.ABaseARM64:   false,
	AFRINT64XD - obj.ABaseARM64:   false,
	AFRINT64XS - obj.ABaseARM64:   false,
	AFRINT64ZD - obj.ABaseARM64:   false,
	AFRINT64ZS - obj.ABaseARM64:   false,
	AFRINTAD - obj.ABaseARM64:     true,
	AFRINTAH - obj.ABaseARM64:     false,
	AFRINTAS - obj.ABaseARM64:     true,
	AFRINTID - obj.ABaseARM64:     true,
	AFRINTIH - obj.ABaseARM64:     false,
	AFRINTIS - obj.ABaseARM64:     true,
	AFRINTMD - obj.ABaseARM64:     true,
	AFRINTMH - obj.ABaseARM64:     false,
	AFRINTMS - obj.ABaseARM64:     true,
	AFRINTND - obj.ABaseARM64:     true,
	AFRINTNH - obj.ABaseARM64:     false,
	AFRINTNS - obj.ABaseARM64:     true,
	AFRINTPD - obj.ABaseARM64:     true,
	AFRINTPH - obj.ABaseARM64:     false,
	AFRINTPS - obj.ABaseARM64:     true,
	AFRINTXD - obj.ABaseARM64:     true,
	AFRINTXH - obj.ABaseARM64:     false,
	AFRINTXS - obj.ABaseARM64:     true,
	AFRINTZD - obj.ABaseARM64:     true,
	AFRINTZH - obj.ABaseARM64:     false,
	AFRINTZS - obj.ABaseARM64:     true,
	AFSQRTD - obj.ABaseARM64:      true,
	AFSQRTH - obj.ABaseARM64:      false,
	AFSQRTS - obj.ABaseARM64:      true,
	AFSTLURB - obj.ABaseARM64:     false,
	AFSTLURD - obj.ABaseARM64:     false,
	AFSTLURH - obj.ABaseARM64:     false,
	AFSTLURQ - obj.ABaseARM64:     false,
	AFSTLURS - obj.ABaseARM64:     false,
	AFSTNPD - obj.ABaseARM64:      false,
	AFSTNPQ - obj.ABaseARM64:      false,
	AFSTNPS - obj.ABaseARM64:      false,
	AFSTPD - obj.ABaseARM64:       true,
	AFSTPQ - obj.ABaseARM64:       true,
	AFSTPS - obj.ABaseARM64:       true,
	AFSUBD - obj.ABaseARM64:       true,
	AFSUBH - obj.ABaseARM64:       false,
	AFSUBS - obj.ABaseARM64:       true,
	AGCSB - obj.ABaseARM64:        false,
	AGCSPOPCX - obj.ABaseARM64:    false,
	AGCSPOPM - obj.ABaseARM64:     false,
	AGCSPOPX - obj.ABaseARM64:     false,
	AGCSPUSHM - obj.ABaseARM64:    false,
	AGCSPUSHX - obj.ABaseARM64:    false,
	AGCSSS1 - obj.ABaseARM64:      false,
	AGCSSS2 - obj.ABaseARM64:      false,
	AGCSSTR - obj.ABaseARM64:      false,
	AGCSSTTR - obj.ABaseARM64:     false,
	AGMI - obj.ABaseARM64:         false,
	AHINT - obj.ABaseARM64:        true,
	AHLT - obj.ABaseARM64:         true,
	AHVC - obj.ABaseARM64:         true,
	AIC - obj.ABaseARM64:          true,
	AINCB - obj.ABaseARM64:        false,
	AINCD - obj.ABaseARM64:        false,
	AINCH - obj.ABaseARM64:        false,
	AINCW - obj.ABaseARM64:        false,
	AIRG - obj.ABaseARM64:         false,
	AISB - obj.ABaseARM64:         true,
	ALD64B - obj.ABaseARM64:       false,
	ALDADDAB - obj.ABaseARM64:     true,
	ALDADDAD - obj.ABaseARM64:     true,
	ALDADDAH - obj.ABaseARM64:     true,
	ALDADDALB - obj.ABaseARM64:    true,
	ALDADDALD - obj.ABaseARM64:    true,
	ALDADDALH - obj.ABaseARM64:    true,
	ALDADDALW - obj.ABaseARM64:    true,
	ALDADDAW - obj.ABaseARM64:     true,
	ALDADDB - obj.ABaseARM64:      true,
	ALDADDD - obj.ABaseARM64:      true,
	ALDADDH - obj.ABaseARM64:      true,
	ALDADDLB - obj.ABaseARM64:     true,
	ALDADDLD - obj.ABaseARM64:     true,
	ALDADDLH - obj.ABaseARM64:     true,
	ALDADDLW - obj.ABaseARM64:     true,
	ALDADDW - obj.ABaseARM64:      true,
	ALDAPR - obj.ABaseARM64:       false,
	ALDAPRB - obj.ABaseARM64:      false,
	ALDAPRH - obj.ABaseARM64:      false,
	ALDAPRW - obj.ABaseARM64:      false,
	ALDAPUR - obj.ABaseARM64:      false,
	ALDAPURB - obj.ABaseARM64:     false,
	ALDAPURH - obj.ABaseARM64:     false,
	ALDAPURSB - obj.ABaseARM64:    false,
	ALDAPURSBW - obj.ABaseARM64:   false,
	ALDAPURSH - obj.ABaseARM64:    false,
	ALDAPURSHW - obj.ABaseARM64:   false,
	ALDAPURSW - obj.ABaseARM64:    false,
	ALDAPURW - obj.ABaseARM64:     false,
	ALDAR - obj.ABaseARM64:        true,
	ALDARB - obj.ABaseARM64:       true,
	ALDARH - obj.ABaseARM64:       true,
	ALDARW - obj.ABaseARM64:       true,
	ALDAXP - obj.ABaseARM64:       true,
	ALDAXPW - obj.ABaseARM64:      true,
	ALDAXR - obj.ABaseARM64:       true,
	ALDAXRB - obj.ABaseARM64:      true,
	ALDAXRH - obj.ABaseARM64:      true,
	ALDAXRW - obj.ABaseARM64:      true,
	ALDCLRAB - obj.ABaseARM64:     true,
	ALDCLRAD - obj.ABaseARM64:     true,
	ALDCLRAH - obj.ABaseARM64:     true,
	ALDCLRALB - obj.ABaseARM64:    true,
	ALDCLRALD - obj.ABaseARM64:    true,
	ALDCLRALH - obj.ABaseARM64:    true,
	ALDCLRALW - obj.ABaseARM64:    true,
	ALDCLRAW - obj.ABaseARM64:     true,
	ALDCLRB - obj.ABaseARM64:      true,
	ALDCLRD - obj.ABaseARM64:      true,
	ALDCLRH - obj.ABaseARM64:      true,
	ALDCLRLB - obj.ABaseARM64:     true,
	ALDCLRLD - obj.ABaseARM64:     true,
	ALDCLRLH - obj.ABaseARM64:     true,
	ALDCLRLW - obj.ABaseARM64:     true,
	ALDCLRP - obj.ABaseARM64:      false,
	ALDCLRPA - obj.ABaseARM64:     false,
	ALDCLRPAL - obj.ABaseARM64:    false,
	ALDCLRPL - obj.ABaseARM64:     false,
	ALDCLRW - obj.ABaseARM64:      true,
	ALDEORAB - obj.ABaseARM64:     true,
	ALDEORAD - obj.ABaseARM64:     true,
	ALDEORAH - obj.ABaseARM64:     true,
	ALDEORALB - obj.ABaseARM64:    true,
	ALDEORALD - obj.ABaseARM64:    true,
	ALDEORALH - obj.ABaseARM64:    true,
	ALDEORALW - obj.ABaseARM64:    true,
	ALDEORAW - obj.ABaseARM64:     true,
	ALDEORB - obj.ABaseARM64:      true,
	ALDEORD - obj.ABaseARM64:      true,
	ALDEORH - obj.ABaseARM64:      true,
	ALDEORLB - obj.ABaseARM64:     true,
	ALDEORLD - obj.ABaseARM64:     true,
	ALDEORLH - obj.ABaseARM64:     true,
	ALDEORLW - obj.ABaseARM64:     true,
	ALDEORW - obj.ABaseARM64:      true,
	ALDG - obj.ABaseARM64:         false,
	ALDGM - obj.ABaseARM64:        false,
	ALDIAPP - obj.ABaseARM64:      false,
	ALDIAPPW - obj.ABaseARM64:     false,
	ALDLAR - obj.ABaseARM64:       false,
	ALDLARB - obj.ABaseARM64:      false,
	ALDLARH - obj.ABaseARM64:      false,
	ALDLARW - obj.ABaseARM64:      false,
	ALDNP - obj.ABaseARM64:        false,
	ALDNPW - obj.ABaseARM64:       false,
	ALDORAB - obj.ABaseARM64:      true,
	ALDORAD - obj.ABaseARM64:      true,
	ALDORAH - obj.ABaseARM64:      true,
	ALDORALB - obj.ABaseARM64:     true,
	ALDORALD - obj.ABaseARM64:     true,
	ALDORALH - obj.ABaseARM64:     true,
	ALDORALW - obj.ABaseARM64:     true,
	ALDORAW - obj.ABaseARM64:      true,
	ALDORB - obj.ABaseARM64:       true,
	ALDORD - obj.ABaseARM64:       true,
	ALDORH - obj.ABaseARM64:       true,
	ALDORLB - obj.ABaseARM64:      true,
	ALDORLD - obj.ABaseARM64:      true,
	ALDORLH - obj.ABaseARM64:      true,
	ALDORLW - obj.ABaseARM64:      true,
	ALDORP - obj.ABaseARM64:       false,
	ALDORW - obj.ABaseARM64:       true,
	ALDP - obj.ABaseARM64:         true,
	ALDPSW - obj.ABaseARM64:       true,
	ALDPW - obj.ABaseARM64:        true,
	ALDRAA - obj.ABaseARM64:       false,
	ALDRAB - obj.ABaseARM64:       false,
	ALDSETPA - obj.ABaseARM64:     false,
	ALDSETPAL - obj.ABaseARM64:    false,
	ALDSETPL - obj.ABaseARM64:     false,
	ALDSMAXAB - obj.ABaseARM64:    false,
	ALDSMAXAD - obj.ABaseARM64:    false,
	ALDSMAXAH - obj.ABaseARM64:    false,
	ALDSMAXALB - obj.ABaseARM64:   false,
	ALDSMAXALD - obj.ABaseARM64:   false,
	ALDSMAXALH - obj.ABaseARM64:   false,
	ALDSMAXALW - obj.ABaseARM64:   false,
	ALDSMAXAW - obj.ABaseARM64:    false,
	ALDSMAXB - obj.ABaseARM64:     false,
	ALDSMAXD - obj.ABaseARM64:     false,
	ALDSMAXH - obj.ABaseARM64:     false,
	ALDSMAXLB - obj.ABaseARM64:    false,
	ALDSMAXLD - obj.ABaseARM64:    false,
	ALDSMAXLH - obj.ABaseARM64:    false,
	ALDSMAXLW - obj.ABaseARM64:    false,
	ALDSMAXW - obj.ABaseARM64:     false,
	ALDSMINAB - obj.ABaseARM64:    false,
	ALDSMINAD - obj.ABaseARM64:    false,
	ALDSMINAH - obj.ABaseARM64:    false,
	ALDSMINALB - obj.ABaseARM64:   false,
	ALDSMINALD - obj.ABaseARM64:   false,
	ALDSMINALH - obj.ABaseARM64:   false,
	ALDSMINALW - obj.ABaseARM64:   false,
	ALDSMINAW - obj.ABaseARM64:    false,
	ALDSMINB - obj.ABaseARM64:     false,
	ALDSMIND - obj.ABaseARM64:     false,
	ALDSMINH - obj.ABaseARM64:     false,
	ALDSMINLB - obj.ABaseARM64:    false,
	ALDSMINLD - obj.ABaseARM64:    false,
	ALDSMINLH - obj.ABaseARM64:    false,
	ALDSMINLW - obj.ABaseARM64:    false,
	ALDSMINW - obj.ABaseARM64:     false,
	ALDTR - obj.ABaseARM64:        false,
	ALDTRB - obj.ABaseARM64:       false,
	ALDTRH - obj.ABaseARM64:       false,
	ALDTRSB - obj.ABaseARM64:      false,
	ALDTRSBW - obj.ABaseARM64:     false,
	ALDTRSH - obj.ABaseARM64:      false,
	ALDTRSHW - obj.ABaseARM64:     false,
	ALDTRSW - obj.ABaseARM64:      false,
	ALDTRW - obj.ABaseARM64:       false,
	ALDUMAXAB - obj.ABaseARM64:    false,
	ALDUMAXAD - obj.ABaseARM64:    false,
	ALDUMAXAH - obj.ABaseARM64:    false,
	ALDUMAXALB - obj.ABaseARM64:   false,
	ALDUMAXALD - obj.ABaseARM64:   false,
	ALDUMAXALH - obj.ABaseARM64:   false,
	ALDUMAXALW - obj.ABaseARM64:   false,
	ALDUMAXAW - obj.ABaseARM64:    false,
	ALDUMAXB - obj.ABaseARM64:     false,
	ALDUMAXD - obj.ABaseARM64:     false,
	ALDUMAXH - obj.ABaseARM64:     false,
	ALDUMAXLB - obj.ABaseARM64:    false,
	ALDUMAXLD - obj.ABaseARM64:    false,
	ALDUMAXLH - obj.ABaseARM64:    false,
	ALDUMAXLW - obj.ABaseARM64:    false,
	ALDUMAXW - obj.ABaseARM64:     false,
	ALDUMINAB - obj.ABaseARM64:    false,
	ALDUMINAD - obj.ABaseARM64:    false,
	ALDUMINAH - obj.ABaseARM64:    false,
	ALDUMINALB - obj.ABaseARM64:   false,
	ALDUMINALD - obj.ABaseARM64:   false,
	ALDUMINALH - obj.ABaseARM64:   false,
	ALDUMINALW - obj.ABaseARM64:   false,
	ALDUMINAW - obj.ABaseARM64:    false,
	ALDUMINB - obj.ABaseARM64:     false,
	ALDUMIND - obj.ABaseARM64:     false,
	ALDUMINH - obj.ABaseARM64:     false,
	ALDUMINLB - obj.ABaseARM64:    false,
	ALDUMINLD - obj.ABaseARM64:    false,
	ALDUMINLH - obj.ABaseARM64:    false,
	ALDUMINLW - obj.ABaseARM64:    false,
	ALDUMINW - obj.ABaseARM64:     false,
	ALDURSB - obj.ABaseARM64:      false,
	ALDURSBW - obj.ABaseARM64:     false,
	ALDURSH - obj.ABaseARM64:      false,
	ALDURSHW - obj.ABaseARM64:     false,
	ALDURSW - obj.ABaseARM64:      false,
	ALDXP - obj.ABaseARM64:        true,
	ALDXPW - obj.ABaseARM64:       true,
	ALDXR - obj.ABaseARM64:        true,
	ALDXRB - obj.ABaseARM64:       true,
	ALDXRH - obj.ABaseARM64:       true,
	ALDXRW - obj.ABaseARM64:       true,
	ALSL - obj.ABaseARM64:         true,
	ALSLV - obj.ABaseARM64:        false,
	ALSLVW - obj.ABaseARM64:       false,
	ALSLW - obj.ABaseARM64:        true,
	ALSR - obj.ABaseARM64:         true,
	ALSRV - obj.ABaseARM64:        false,
	ALSRVW - obj.ABaseARM64:       false,
	ALSRW - obj.ABaseARM64:        true,
	AMADD - obj.ABaseARM64:        true,
	AMADDW - obj.ABaseARM64:       true,
	AMNEG - obj.ABaseARM64:        true,
	AMNEGW - obj.ABaseARM64:       true,
	AMOVB - obj.ABaseARM64:        true,
	AMOVBU - obj.ABaseARM64:       true,
	AMOVBW - obj.ABaseARM64:       false,
	AMOVD - obj.ABaseARM64:        true,
	AMOVH - obj.ABaseARM64:        true,
	AMOVHU - obj.ABaseARM64:       true,
	AMOVHW - obj.ABaseARM64:       false,
	AMOVK - obj.ABaseARM64:        true,
	AMOVKW - obj.ABaseARM64:       true,
	AMOVN - obj.ABaseARM64:        true,
	AMOVNW - obj.ABaseARM64:       true,
	AMOVT - obj.ABaseARM64:        false,
	AMOVW - obj.ABaseARM64:        true,
	AMOVWU - obj.ABaseARM64:       true,
	AMOVZ - obj.ABaseARM64:        true,
	AMOVZW - obj.ABaseARM64:       true,
	AMRRS - obj.ABaseARM64:        false,
	AMRS - obj.ABaseARM64:         true,
	AMSR - obj.ABaseARM64:         true,
	AMSRR - obj.ABaseARM64:        false,
	AMSUB - obj.ABaseARM64:        true,
	AMSUBW - obj.ABaseARM64:       true,
	AMUL - obj.ABaseARM64:         true,
	AMULW - obj.ABaseARM64:        true,
	AMVN - obj.ABaseARM64:         true,
	AMVNW - obj.ABaseARM64:        true,
	ANEG - obj.ABaseARM64:         true,
	ANEGS - obj.ABaseARM64:        true,
	ANEGSW - obj.ABaseARM64:       true,
	ANEGW - obj.ABaseARM64:        true,
	ANGC - obj.ABaseARM64:         true,
	ANGCS - obj.ABaseARM64:        true,
	ANGCSW - obj.ABaseARM64:       true,
	ANGCW - obj.ABaseARM64:        true,
	ANOOP - obj.ABaseARM64:        true,
	AORN - obj.ABaseARM64:         true,
	AORNW - obj.ABaseARM64:        true,
	AORR - obj.ABaseARM64:         true,
	AORRW - obj.ABaseARM64:        true,
	APACDA - obj.ABaseARM64:       false,
	APACDB - obj.ABaseARM64:       false,
	APACDZA - obj.ABaseARM64:      false,
	APACDZB - obj.ABaseARM64:      false,
	APACGA - obj.ABaseARM64:       false,
	APACIA - obj.ABaseARM64:       false,
	APACIA1716 - obj.ABaseARM64:   false,
	APACIASP - obj.ABaseARM64:     false,
	APACIAZ - obj.ABaseARM64:      false,
	APACIB - obj.ABaseARM64:       false,
	APACIB1716 - obj.ABaseARM64:   false,
	APACIBSP - obj.ABaseARM64:     false,
	APACIBZ - obj.ABaseARM64:      false,
	APACIZA - obj.ABaseARM64:      false,
	APACIZB - obj.ABaseARM64:      false,
	APAND - obj.ABaseARM64:        false,
	APANDS - obj.ABaseARM64:       false,
	APBIC - obj.ABaseARM64:        false,
	APBICS - obj.ABaseARM64:       false,
	APBRKAS - obj.ABaseARM64:      false,
	APBRKBS - obj.ABaseARM64:      false,
	APBRKN - obj.ABaseARM64:       false,
	APBRKNS - obj.ABaseARM64:      false,
	APBRKPA - obj.ABaseARM64:      false,
	APBRKPAS - obj.ABaseARM64:     false,
	APBRKPB - obj.ABaseARM64:      false,
	APBRKPBS - obj.ABaseARM64:     false,
	APCNTP - obj.ABaseARM64:       false,
	APDECP - obj.ABaseARM64:       false,
	APEOR - obj.ABaseARM64:        false,
	APEORS - obj.ABaseARM64:       false,
	APEXT - obj.ABaseARM64:        false,
	APFALSE - obj.ABaseARM64:      false,
	APFIRST - obj.ABaseARM64:      false,
	APINCP - obj.ABaseARM64:       false,
	APLD1B - obj.ABaseARM64:       false,
	APLDR - obj.ABaseARM64:        false,
	APMOV - obj.ABaseARM64:        false,
	APMOVS - obj.ABaseARM64:       false,
	APNAND - obj.ABaseARM64:       false,
	APNANDS - obj.ABaseARM64:      false,
	APNEXT - obj.ABaseARM64:       false,
	APNOR - obj.ABaseARM64:        false,
	APNORS - obj.ABaseARM64:       false,
	APNOT - obj.ABaseARM64:        false,
	APNOTS - obj.ABaseARM64:       false,
	APORN - obj.ABaseARM64:        false,
	APORNS - obj.ABaseARM64:       false,
	APORR - obj.ABaseARM64:        false,
	APORRS - obj.ABaseARM64:       false,
	APPRFB - obj.ABaseARM64:       false,
	APPRFD - obj.ABaseARM64:       false,
	APPRFH - obj.ABaseARM64:       false,
	APPRFW - obj.ABaseARM64:       false,
	APRDFFR - obj.ABaseARM64:      false,
	APRDFFRS - obj.ABaseARM64:     false,
	APREV - obj.ABaseARM64:        false,
	APRFM - obj.ABaseARM64:        true,
	APRFUM - obj.ABaseARM64:       false,
	APSB - obj.ABaseARM64:         false,
	APSEL - obj.ABaseARM64:        false,
	APSQDECPD - obj.ABaseARM64:    false,
	APSQDECPS - obj.ABaseARM64:    false,
	APSQINCPD - obj.ABaseARM64:    false,
	APSQINCPS - obj.ABaseARM64:    false,
	APSSBB - obj.ABaseARM64:       false,
	APST1B - obj.ABaseARM64:       false,
	APSTR - obj.ABaseARM64:        false,
	APTEST - obj.ABaseARM64:       false,
	APTRN1 - obj.ABaseARM64:       false,
	APTRN2 - obj.ABaseARM64:       false,
	APTRUE - obj.ABaseARM64:       false,
	APTRUES - obj.ABaseARM64:      false,
	APUNPKHI - obj.ABaseARM64:     false,
	APUNPKLO - obj.ABaseARM64:     false,
	APUQDECPD - obj.ABaseARM64:    false,
	APUQDECPS - obj.ABaseARM64:    false,
	APUQINCPD - obj.ABaseARM64:    false,
	APUQINCPS - obj.ABaseARM64:    false,
	APUZP1 - obj.ABaseARM64:       false,
	APUZP2 - obj.ABaseARM64:       false,
	APWHILERW - obj.ABaseARM64:    false,
	APWHILEWR - obj.ABaseARM64:    false,
	APWRFFR - obj.ABaseARM64:      false,
	APZIP1 - obj.ABaseARM64:       false,
	APZIP2 - obj.ABaseARM64:       false,
	ARBIT - obj.ABaseARM64:        true,
	ARBITW - obj.ABaseARM64:       true,
	ARCWCAS - obj.ABaseARM64:      false,
	ARCWCASA - obj.ABaseARM64:     false,
	ARCWCASAL - obj.ABaseARM64:    false,
	ARCWCASL - obj.ABaseARM64:     false,
	ARCWCASP - obj.ABaseARM64:     false,
	ARCWCASPA - obj.ABaseARM64:    false,
	ARCWCASPAL - obj.ABaseARM64:   false,
	ARCWCASPL - obj.ABaseARM64:    false,
	ARCWCLR - obj.ABaseARM64:      false,
	ARCWCLRA - obj.ABaseARM64:     false,
	ARCWCLRAL - obj.ABaseARM64:    false,
	ARCWCLRL - obj.ABaseARM64:     false,
	ARCWCLRP - obj.ABaseARM64:     false,
	ARCWCLRPA - obj.ABaseARM64:    false,
	ARCWCLRPAL - obj.ABaseARM64:   false,
	ARCWCLRPL - obj.ABaseARM64:    false,
	ARCWSCAS - obj.ABaseARM64:     false,
	ARCWSCASA - obj.ABaseARM64:    false,
	ARCWSCASAL - obj.ABaseARM64:   false,
	ARCWSCASL - obj.ABaseARM64:    false,
	ARCWSCASP - obj.ABaseARM64:    false,
	ARCWSCASPA - obj.ABaseARM64:   false,
	ARCWSCASPAL - obj.ABaseARM64:  false,
	ARCWSCASPL - obj.ABaseARM64:   false,
	ARCWSCLR - obj.ABaseARM64:     false,
	ARCWSCLRA - obj.ABaseARM64:    false,
	ARCWSCLRAL - obj.ABaseARM64:   false,
	ARCWSCLRL - obj.ABaseARM64:    false,
	ARCWSCLRP - obj.ABaseARM64:    false,
	ARCWSCLRPA - obj.ABaseARM64:   false,
	ARCWSCLRPAL - obj.ABaseARM64:  false,
	ARCWSCLRPL - obj.ABaseARM64:   false,
	ARCWSET - obj.ABaseARM64:      false,
	ARCWSETA - obj.ABaseARM64:     false,
	ARCWSETAL - obj.ABaseARM64:    false,
	ARCWSETL - obj.ABaseARM64:     false,
	ARCWSETP - obj.ABaseARM64:     false,
	ARCWSETPA - obj.ABaseARM64:    false,
	ARCWSETPAL - obj.ABaseARM64:   false,
	ARCWSETPL - obj.ABaseARM64:    false,
	ARCWSSET - obj.ABaseARM64:     false,
	ARCWSSETA - obj.ABaseARM64:    false,
	ARCWSSETAL - obj.ABaseARM64:   false,
	ARCWSSETL - obj.ABaseARM64:    false,
	ARCWSSETP - obj.ABaseARM64:    false,
	ARCWSSETPA - obj.ABaseARM64:   false,
	ARCWSSETPAL - obj.ABaseARM64:  false,
	ARCWSSETPL - obj.ABaseARM64:   false,
	ARCWSSWP - obj.ABaseARM64:     false,
	ARCWSSWPA - obj.ABaseARM64:    false,
	ARCWSSWPAL - obj.ABaseARM64:   false,
	ARCWSSWPL - obj.ABaseARM64:    false,
	ARCWSSWPP - obj.ABaseARM64:    false,
	ARCWSSWPPA - obj.ABaseARM64:   false,
	ARCWSSWPPAL - obj.ABaseARM64:  false,
	ARCWSSWPPL - obj.ABaseARM64:   false,
	ARCWSWP - obj.ABaseARM64:      false,
	ARCWSWPA - obj.ABaseARM64:     false,
	ARCWSWPAL - obj.ABaseARM64:    false,
	ARCWSWPL - obj.ABaseARM64:     false,
	ARCWSWPP - obj.ABaseARM64:     false,
	ARCWSWPPA - obj.ABaseARM64:    false,
	ARCWSWPPAL - obj.ABaseARM64:   false,
	ARCWSWPPL - obj.ABaseARM64:    false,
	ARDSVL - obj.ABaseARM64:       false,
	ARDVL - obj.ABaseARM64:        false,
	AREM - obj.ABaseARM64:         true,
	AREMW - obj.ABaseARM64:        true,
	ARETAA - obj.ABaseARM64:       false,
	ARETAB - obj.ABaseARM64:       false,
	AREV - obj.ABaseARM64:         true,
	AREV16 - obj.ABaseARM64:       true,
	AREV16W - obj.ABaseARM64:      true,
	AREV32 - obj.ABaseARM64:       true,
	AREV64 - obj.ABaseARM64:       false,
	AREVW - obj.ABaseARM64:        true,
	ARMIF - obj.ABaseARM64:        false,
	AROR - obj.ABaseARM64:         true,
	ARORV - obj.ABaseARM64:        false,
	ARORVW - obj.ABaseARM64:       false,
	ARORW - obj.ABaseARM64:        true,
	ARPRFM - obj.ABaseARM64:       false,
	ASB - obj.ABaseARM64:          false,
	ASBC - obj.ABaseARM64:         true,
	ASBCS - obj.ABaseARM64:        true,
	ASBCSW - obj.ABaseARM64:       true,
	ASBCW - obj.ABaseARM64:        true,
	ASBFIZ - obj.ABaseARM64:       true,
	ASBFIZW - obj.ABaseARM64:      true,
	ASBFM - obj.ABaseARM64:        true,
	ASBFMW - obj.ABaseARM64:       true,
	ASBFX - obj.ABaseARM64:        true,
	ASBFXW - obj.ABaseARM64:       true,
	ASCVTFD - obj.ABaseARM64:      true,
	ASCVTFH - obj.ABaseARM64:      false,
	ASCVTFS - obj.ABaseARM64:      true,
	ASCVTFWD - obj.ABaseARM64:     true,
	ASCVTFWH - obj.ABaseARM64:     false,
	ASCVTFWS - obj.ABaseARM64:     true,
	ASDIV - obj.ABaseARM64:        true,
	ASDIVW - obj.ABaseARM64:       true,
	ASETE - obj.ABaseARM64:        false,
	ASETEN - obj.ABaseARM64:       false,
	ASETET - obj.ABaseARM64:       false,
	ASETETN - obj.ABaseARM64:      false,
	ASETF16 - obj.ABaseARM64:      false,
	ASETF8 - obj.ABaseARM64:       false,
	ASETFFR - obj.ABaseARM64:      false,
	ASETGE - obj.ABaseARM64:       false,
	ASETGEN - obj.ABaseARM64:      false,
	ASETGET - obj.ABaseARM64:      false,
	ASETGETN - obj.ABaseARM64:     false,
	ASETGM - obj.ABaseARM64:       false,
	ASETGMN - obj.ABaseARM64:      false,
	ASETGMT - obj.ABaseARM64:      false,
	ASETGMTN - obj.ABaseARM64:     false,
	ASETGP - obj.ABaseARM64:       false,
	ASETGPN - obj.ABaseARM64:      false,
	ASETGPT - obj.ABaseARM64:      false,
	ASETGPTN - obj.ABaseARM64:     false,
	ASETM - obj.ABaseARM64:        false,
	ASETMN - obj.ABaseARM64:       false,
	ASETMT - obj.ABaseARM64:       false,
	ASETMTN - obj.ABaseARM64:      false,
	ASETP - obj.ABaseARM64:        false,
	ASETPN - obj.ABaseARM64:       false,
	ASETPT - obj.ABaseARM64:       false,
	ASETPTN - obj.ABaseARM64:      false,
	ASEV - obj.ABaseARM64:         true,
	ASEVL - obj.ABaseARM64:        true,
	ASHA1C - obj.ABaseARM64:       true,
	ASHA1H - obj.ABaseARM64:       true,
	ASHA1M - obj.ABaseARM64:       true,
	ASHA1P - obj.ABaseARM64:       true,
	ASHA1SU0 - obj.ABaseARM64:     true,
	ASHA1SU1 - obj.ABaseARM64:     true,
	ASHA256H - obj.ABaseARM64:     true,
	ASHA256H2 - obj.ABaseARM64:    true,
	ASHA256SU0 - obj.ABaseARM64:   true,
	ASHA256SU1 - obj.ABaseARM64:   true,
	ASHA512H - obj.ABaseARM64:     true,
	ASHA512H2 - obj.ABaseARM64:    true,
	ASHA512SU0 - obj.ABaseARM64:   true,
	ASHA512SU1 - obj.ABaseARM64:   true,
	ASM3PARTW1 - obj.ABaseARM64:   false,
	ASM3PARTW2 - obj.ABaseARM64:   false,
	ASM3SS1 - obj.ABaseARM64:      false,
	ASM3TT1A - obj.ABaseARM64:     false,
	ASM3TT1B - obj.ABaseARM64:     false,
	ASM3TT2A - obj.ABaseARM64:     false,
	ASM3TT2B - obj.ABaseARM64:     false,
	ASM4E - obj.ABaseARM64:        false,
	ASM4EKEY - obj.ABaseARM64:     false,
	ASMADDL - obj.ABaseARM64:      true,
	ASMAX - obj.ABaseARM64:        false,
	ASMAXW - obj.ABaseARM64:       false,
	ASMC - obj.ABaseARM64:         true,
	ASMIN - obj.ABaseARM64:        false,
	ASMINW - obj.ABaseARM64:       false,
	ASMNEGL - obj.ABaseARM64:      true,
	ASMSTART - obj.ABaseARM64:     false,
	ASMSTOP - obj.ABaseARM64:      false,
	ASMSUBL - obj.ABaseARM64:      true,
	ASMULH - obj.ABaseARM64:       true,
	ASMULL - obj.ABaseARM64:       true,
	ASQDECBD - obj.ABaseARM64:     false,
	ASQDECBS - obj.ABaseARM64:     false,
	ASQDECDD - obj.ABaseARM64:     false,
	ASQDECDS - obj.ABaseARM64:     false,
	ASQDECHD - obj.ABaseARM64:     false,
	ASQDECHS - obj.ABaseARM64:     false,
	ASQDECWD - obj.ABaseARM64:     false,
	ASQDECWS - obj.ABaseARM64:     false,
	ASQINCBD - obj.ABaseARM64:     false,
	ASQINCBS - obj.ABaseARM64:     false,
	ASQINCDD - obj.ABaseARM64:     false,
	ASQINCDS - obj.ABaseARM64:     false,
	ASQINCHD - obj.ABaseARM64:     false,
	ASQINCHS - obj.ABaseARM64:     false,
	ASQINCWD - obj.ABaseARM64:     false,
	ASQINCWS - obj.ABaseARM64:     false,
	ASSBB - obj.ABaseARM64:        false,
	AST2G - obj.ABaseARM64:        false,
	AST64B - obj.ABaseARM64:       false,
	AST64BV - obj.ABaseARM64:      false,
	AST64BV0 - obj.ABaseARM64:     false,
	ASTADDB - obj.ABaseARM64:      false,
	ASTADDD - obj.ABaseARM64:      false,
	ASTADDH - obj.ABaseARM64:      false,
	ASTADDLB - obj.ABaseARM64:     false,
	ASTADDLD - obj.ABaseARM64:     false,
	ASTADDLH - obj.ABaseARM64:     false,
	ASTADDLW - obj.ABaseARM64:     false,
	ASTADDW - obj.ABaseARM64:      false,
	ASTCLRB - obj.ABaseARM64:      false,
	ASTCLRD - obj.ABaseARM64:      false,
	ASTCLRH - obj.ABaseARM64:      false,
	ASTCLRLB - obj.ABaseARM64:     false,
	ASTCLRLD - obj.ABaseARM64:     false,
	ASTCLRLH - obj.ABaseARM64:     false,
	ASTCLRLW - obj.ABaseARM64:     false,
	ASTCLRW - obj.ABaseARM64:      false,
	ASTEORB - obj.ABaseARM64:      false,
	ASTEORD - obj.ABaseARM64:      false,
	ASTEORH - obj.ABaseARM64:      false,
	ASTEORLB - obj.ABaseARM64:     false,
	ASTEORLD - obj.ABaseARM64:     false,
	ASTEORLH - obj.ABaseARM64:     false,
	ASTEORLW - obj.ABaseARM64:     false,
	ASTEORW - obj.ABaseARM64:      false,
	ASTG - obj.ABaseARM64:         false,
	ASTGM - obj.ABaseARM64:        false,
	ASTGP - obj.ABaseARM64:        false,
	ASTILP - obj.ABaseARM64:       false,
	ASTILPW - obj.ABaseARM64:      false,
	ASTLLR - obj.ABaseARM64:       false,
	ASTLLRB - obj.ABaseARM64:      false,
	ASTLLRH - obj.ABaseARM64:      false,
	ASTLLRW - obj.ABaseARM64:      false,
	ASTLR - obj.ABaseARM64:        true,
	ASTLRB - obj.ABaseARM64:       true,
	ASTLRH - obj.ABaseARM64:       true,
	ASTLRW - obj.ABaseARM64:       true,
	ASTLUR - obj.ABaseARM64:       false,
	ASTLURB - obj.ABaseARM64:      false,
	ASTLURH - obj.ABaseARM64:      false,
	ASTLURW - obj.ABaseARM64:      false,
	ASTLXP - obj.ABaseARM64:       true,
	ASTLXPW - obj.ABaseARM64:      true,
	ASTLXR - obj.ABaseARM64:       true,
	ASTLXRB - obj.ABaseARM64:      true,
	ASTLXRH - obj.ABaseARM64:      true,
	ASTLXRW - obj.ABaseARM64:      true,
	ASTNP - obj.ABaseARM64:        false,
	ASTNPW - obj.ABaseARM64:       false,
	ASTORB - obj.ABaseARM64:       false,
	ASTORD - obj.ABaseARM64:       false,
	ASTORH - obj.ABaseARM64:       false,
	ASTORLB - obj.ABaseARM64:      false,
	ASTORLD - obj.ABaseARM64:      false,
	ASTORLH - obj.ABaseARM64:      false,
	ASTORLW - obj.ABaseARM64:      false,
	ASTORW - obj.ABaseARM64:       false,
	ASTP - obj.ABaseARM64:         true,
	ASTPW - obj.ABaseARM64:        true,
	ASTSMAXB - obj.ABaseARM64:     false,
	ASTSMAXD - obj.ABaseARM64:     false,
	ASTSMAXH - obj.ABaseARM64:     false,
	ASTSMAXLB - obj.ABaseARM64:    false,
	ASTSMAXLD - obj.ABaseARM64:    false,
	ASTSMAXLH - obj.ABaseARM64:    false,
	ASTSMAXLW - obj.ABaseARM64:    false,
	ASTSMAXW - obj.ABaseARM64:     false,
	ASTSMINB - obj.ABaseARM64:     false,
	ASTSMIND - obj.ABaseARM64:     false,
	ASTSMINH - obj.ABaseARM64:     false,
	ASTSMINLB - obj.ABaseARM64:    false,
	ASTSMINLD - obj.ABaseARM64:    false,
	ASTSMINLH - obj.ABaseARM64:    false,
	ASTSMINLW - obj.ABaseARM64:    false,
	ASTSMINW - obj.ABaseARM64:     false,
	ASTTR - obj.ABaseARM64:        false,
	ASTTRB - obj.ABaseARM64:       false,
	ASTTRH - obj.ABaseARM64:       false,
	ASTTRW - obj.ABaseARM64:       false,
	ASTUMAXB - obj.ABaseARM64:     false,
	ASTUMAXD - obj.ABaseARM64:     false,
	ASTUMAXH - obj.ABaseARM64:     false,
	ASTUMAXLB - obj.ABaseARM64:    false,
	ASTUMAXLD - obj.ABaseARM64:    false,
	ASTUMAXLH - obj.ABaseARM64:    false,
	ASTUMAXLW - obj.ABaseARM64:    false,
	ASTUMAXW - obj.ABaseARM64:     false,
	ASTUMINB - obj.ABaseARM64:     false,
	ASTUMIND - obj.ABaseARM64:     false,
	ASTUMINH - obj.ABaseARM64:     false,
	ASTUMINLB - obj.ABaseARM64:    false,
	ASTUMINLD - obj.ABaseARM64:    false,
	ASTUMINLH - obj.ABaseARM64:    false,
	ASTUMINLW - obj.ABaseARM64:    false,
	ASTUMINW - obj.ABaseARM64:     false,
	ASTXP - obj.ABaseARM64:        true,
	ASTXPW - obj.ABaseARM64:       true,
	ASTXR - obj.ABaseARM64:        true,
	ASTXRB - obj.ABaseARM64:       true,
	ASTXRH - obj.ABaseARM64:       true,
	ASTXRW - obj.ABaseARM64:       true,
	ASTZ2G - obj.ABaseARM64:       false,
	ASTZG - obj.ABaseARM64:        false,
	ASTZGM - obj.ABaseARM64:       false,
	ASUB - obj.ABaseARM64:         true,
	ASUBG - obj.ABaseARM64:        false,
	ASUBP - obj.ABaseARM64:        false,
	ASUBPS - obj.ABaseARM64:       false,
	ASUBS - obj.ABaseARM64:        true,
	ASUBSW - obj.ABaseARM64:       true,
	ASUBW - obj.ABaseARM64:        true,
	ASVC - obj.ABaseARM64:         true,
	ASWPAB - obj.ABaseARM64:       true,
	ASWPAD - obj.ABaseARM64:       true,
	ASWPAH - obj.ABaseARM64:       true,
	ASWPALB - obj.ABaseARM64:      true,
	ASWPALD - obj.ABaseARM64:      true,
	ASWPALH - obj.ABaseARM64:      true,
	ASWPALW - obj.ABaseARM64:      true,
	ASWPAW - obj.ABaseARM64:       true,
	ASWPB - obj.ABaseARM64:        true,
	ASWPD - obj.ABaseARM64:        true,
	ASWPH - obj.ABaseARM64:        true,
	ASWPLB - obj.ABaseARM64:       true,
	ASWPLD - obj.ABaseARM64:       true,
	ASWPLH - obj.ABaseARM64:       true,
	ASWPLW - obj.ABaseARM64:       true,
	ASWPP - obj.ABaseARM64:        false,
	ASWPPA - obj.ABaseARM64:       false,
	ASWPPAL - obj.ABaseARM64:      false,
	ASWPPL - obj.ABaseARM64:       false,
	ASWPW - obj.ABaseARM64:        true,
	ASXTB - obj.ABaseARM64:        true,
	ASXTBW - obj.ABaseARM64:       true,
	ASXTH - obj.ABaseARM64:        true,
	ASXTHW - obj.ABaseARM64:       true,
	ASXTW - obj.ABaseARM64:        true,
	ASYS - obj.ABaseARM64:         true,
	ASYSL - obj.ABaseARM64:        true,
	ASYSP - obj.ABaseARM64:        false,
	ATBNZ - obj.ABaseARM64:        true,
	ATBZ - obj.ABaseARM64:         true,
	ATCANCEL - obj.ABaseARM64:     false,
	ATCOMMIT - obj.ABaseARM64:     false,
	ATLBI - obj.ABaseARM64:        true,
	ATLBIP - obj.ABaseARM64:       false,
	ATRCIT - obj.ABaseARM64:       false,
	ATSB - obj.ABaseARM64:         false,
	ATST - obj.ABaseARM64:         true,
	ATSTART - obj.ABaseARM64:      false,
	ATSTW - obj.ABaseARM64:        true,
	ATTEST - obj.ABaseARM64:       false,
	AUBFIZ - obj.ABaseARM64:       true,
	AUBFIZW - obj.ABaseARM64:      true,
	AUBFM - obj.ABaseARM64:        true,
	AUBFMW - obj.ABaseARM64:       true,
	AUBFX - obj.ABaseARM64:        true,
	AUBFXW - obj.ABaseARM64:       true,
	AUCVTFD - obj.ABaseARM64:      true,
	AUCVTFH - obj.ABaseARM64:      false,
	AUCVTFS - obj.ABaseARM64:      true,
	AUCVTFWD - obj.ABaseARM64:     true,
	AUCVTFWH - obj.ABaseARM64:     false,
	AUCVTFWS - obj.ABaseARM64:     true,
	AUDF - obj.ABaseARM64:         false,
	AUDIV - obj.ABaseARM64:        true,
	AUDIVW - obj.ABaseARM64:       true,
	AUMADDL - obj.ABaseARM64:      true,
	AUMAX - obj.ABaseARM64:        false,
	AUMAXW - obj.ABaseARM64:       false,
	AUMIN - obj.ABaseARM64:        false,
	AUMINW - obj.ABaseARM64:       false,
	AUMNEGL - obj.ABaseARM64:      true,
	AUMSUBL - obj.ABaseARM64:      true,
	AUMULH - obj.ABaseARM64:       true,
	AUMULL - obj.ABaseARM64:       true,
	AUQDECBD - obj.ABaseARM64:     false,
	AUQDECBS - obj.ABaseARM64:     false,
	AUQDECDD - obj.ABaseARM64:     false,
	AUQDECDS - obj.ABaseARM64:     false,
	AUQDECHD - obj.ABaseARM64:     false,
	AUQDECHS - obj.ABaseARM64:     false,
	AUQDECWD - obj.ABaseARM64:     false,
	AUQDECWS - obj.ABaseARM64:     false,
	AUQINCBD - obj.ABaseARM64:     false,
	AUQINCBS - obj.ABaseARM64:     false,
	AUQINCDD - obj.ABaseARM64:     false,
	AUQINCDS - obj.ABaseARM64:     false,
	AUQINCHD - obj.ABaseARM64:     false,
	AUQINCHS - obj.ABaseARM64:     false,
	AUQINCWD - obj.ABaseARM64:     false,
	AUQINCWS - obj.ABaseARM64:     false,
	AUREM - obj.ABaseARM64:        true,
	AUREMW - obj.ABaseARM64:       true,
	AUXTB - obj.ABaseARM64:        true,
	AUXTBW - obj.ABaseARM64:       true,
	AUXTH - obj.ABaseARM64:        true,
	AUXTHW - obj.ABaseARM64:       true,
	AUXTW - obj.ABaseARM64:        true,
	AVABS - obj.ABaseARM64:        false,
	AVADD - obj.ABaseARM64:        true,
	AVADDHN - obj.ABaseARM64:      false,
	AVADDHN2 - obj.ABaseARM64:     false,
	AVADDP - obj.ABaseARM64:       true,
	AVADDV - obj.ABaseARM64:       true,
	AVAND - obj.ABaseARM64:        true,
	AVBCAX - obj.ABaseARM64:       true,
	AVBFCVTN - obj.ABaseARM64:     false,
	AVBFCVTN2 - obj.ABaseARM64:    false,
	AVBFDOT - obj.ABaseARM64:      false,
	AVBFMLALB - obj.ABaseARM64:    false,
	AVBFMLALT - obj.ABaseARM64:    false,
	AVBFMMLA - obj.ABaseARM64:     false,
	AVBIC - obj.ABaseARM64:        false,
	AVBICH - obj.ABaseARM64:       false,
	AVBICS - obj.ABaseARM64:       false,
	AVBIF - obj.ABaseARM64:        true,
	AVBIT - obj.ABaseARM64:        true,
	AVBSL - obj.ABaseARM64:        true,
	AVCLS - obj.ABaseARM64:        false,
	AVCLZ - obj.ABaseARM64:        false,
	AVCMEQ - obj.ABaseARM64:       true,
	AVCMGE - obj.ABaseARM64:       false,
	AVCMGT - obj.ABaseARM64:       false,
	AVCMHI - obj.ABaseARM64:       false,
	AVCMHS - obj.ABaseARM64:       false,
	AVCMLE - obj.ABaseARM64:       false,
	AVCMLT - obj.ABaseARM64:       false,
	AVCMTST - obj.ABaseARM64:      true,
	AVCNT - obj.ABaseARM64:        true,
	AVDUP - obj.ABaseARM64:        true,
	AVEOR - obj.ABaseARM64:        true,
	AVEOR3 - obj.ABaseARM64:       true,
	AVEXT - obj.ABaseARM64:        true,
	AVFABD - obj.ABaseARM64:       false,
	AVFABDH - obj.ABaseARM64:      false,
	AVFABSH - obj.ABaseARM64:      false,
	AVFABSS - obj.ABaseARM64:      false,
	AVFACGE - obj.ABaseARM64:      false,
	AVFACGEH - obj.ABaseARM64:     false,
	AVFACGT - obj.ABaseARM64:      false,
	AVFACGTH - obj.ABaseARM64:     false,
	AVFADDH - obj.ABaseARM64:      false,
	AVFADDPH - obj.ABaseARM64:     false,
	AVFADDPS - obj.ABaseARM64:     false,
	AVFADDS - obj.ABaseARM64:      false,
	AVFCADD - obj.ABaseARM64:      false,
	AVFCMEQ - obj.ABaseARM64:      false,
	AVFCMEQH - obj.ABaseARM64:     false,
	AVFCMGE - obj.ABaseARM64:      false,
	AVFCMGEH - obj.ABaseARM64:     false,
	AVFCMGT - obj.ABaseARM64:      false,
	AVFCMGTH - obj.ABaseARM64:     false,
	AVFCMLA - obj.ABaseARM64:      false,
	AVFCMLE - obj.ABaseARM64:      false,
	AVFCMLEH - obj.ABaseARM64:     false,
	AVFCMLT - obj.ABaseARM64:      false,
	AVFCMLTH - obj.ABaseARM64:     false,
	AVFCVTAS - obj.ABaseARM64:     false,
	AVFCVTASH - obj.ABaseARM64:    false,
	AVFCVTAU - obj.ABaseARM64:     false,
	AVFCVTAUH - obj.ABaseARM64:    false,
	AVFCVTL - obj.ABaseARM64:      false,
	AVFCVTL2 - obj.ABaseARM64:     false,
	AVFCVTMS - obj.ABaseARM64:     false,
	AVFCVTMSH - obj.ABaseARM64:    false,
	AVFCVTMU - obj.ABaseARM64:     false,
	AVFCVTMUH - obj.ABaseARM64:    false,
	AVFCVTN - obj.ABaseARM64:      false,
	AVFCVTN2 - obj.ABaseARM64:     false,
	AVFCVTNS - obj.ABaseARM64:     false,
	AVFCVTNSH - obj.ABaseARM64:    false,
	AVFCVTNU - obj.ABaseARM64:     false,
	AVFCVTNUH - obj.ABaseARM64:    false,
	AVFCVTPS - obj.ABaseARM64:     false,
	AVFCVTPSH - obj.ABaseARM64:    false,
	AVFCVTPU - obj.ABaseARM64:     false,
	AVFCVTPUH - obj.ABaseARM64:    false,
	AVFCVTXN - obj.ABaseARM64:     false,
	AVFCVTXN2 - obj.ABaseARM64:    false,
	AVFCVTZS - obj.ABaseARM64:     false,
	AVFCVTZSH - obj.ABaseARM64:    false,
	AVFCVTZU - obj.ABaseARM64:     false,
	AVFCVTZUH - obj.ABaseARM64:    false,
	AVFDIVH - obj.ABaseARM64:      false,
	AVFDIVS - obj.ABaseARM64:      false,
	AVFMAXH - obj.ABaseARM64:      false,
	AVFMAXNMH - obj.ABaseARM64:    false,
	AVFMAXNMPH - obj.ABaseARM64:   false,
	AVFMAXNMPS - obj.ABaseARM64:   false,
	AVFMAXNMS - obj.ABaseARM64:    false,
	AVFMAXNMVH - obj.ABaseARM64:   false,
	AVFMAXNMVS - obj.ABaseARM64:   false,
	AVFMAXPH - obj.ABaseARM64:     false,
	AVFMAXPS - obj.ABaseARM64:     false,
	AVFMAXS - obj.ABaseARM64:      false,
	AVFMAXVH - obj.ABaseARM64:     false,
	AVFMAXVS - obj.ABaseARM64:     false,
	AVFMINH - obj.ABaseARM64:      false,
	AVFMINNMH - obj.ABaseARM64:    false,
	AVFMINNMPH - obj.ABaseARM64:   false,
	AVFMINNMPS - obj.ABaseARM64:   false,
	AVFMINNMS - obj.ABaseARM64:    false,
	AVFMINNMVH - obj.ABaseARM64:   false,
	AVFMINNMVS - obj.ABaseARM64:   false,
	AVFMINPH - obj.ABaseARM64:     false,
	AVFMINPS - obj.ABaseARM64:     false,
	AVFMINS - obj.ABaseARM64:      false,
	AVFMINVH - obj.ABaseARM64:     false,
	AVFMINVS - obj.ABaseARM64:     false,
	AVFMLA - obj.ABaseARM64:       true,
	AVFMLAH - obj.ABaseARM64:      false,
	AVFMLAL - obj.ABaseARM64:      false,
	AVFMLAL2 - obj.ABaseARM64:     false,
	AVFMLAS - obj.ABaseARM64:      false,
	AVFMLS - obj.ABaseARM64:       true,
	AVFMLSH - obj.ABaseARM64:      false,
	AVFMLSL - obj.ABaseARM64:      false,
	AVFMLSL2 - obj.ABaseARM64:     false,
	AVFMLSS - obj.ABaseARM64:      false,
	AVFMOVD - obj.ABaseARM64:      false,
	AVFMOVH - obj.ABaseARM64:      false,
	AVFMOVS - obj.ABaseARM64:      false,
	AVFMUL - obj.ABaseARM64:       false,
	AVFMULH - obj.ABaseARM64:      false,
	AVFMULS - obj.ABaseARM64:      false,
	AVFMULX - obj.ABaseARM64:      false,
	AVFMULXH - obj.ABaseARM64:     false,
	AVFNEGH - obj.ABaseARM64:      false,
	AVFNEGS - obj.ABaseARM64:      false,
	AVFRECPE - obj.ABaseARM64:     false,
	AVFRECPEH - obj.ABaseARM64:    false,
	AVFRECPS - obj.ABaseARM64:     false,
	AVFRECPSH - obj.ABaseARM64:    false,
	AVFRECPXH - obj.ABaseARM64:    false,
	AVFRECPXS - obj.ABaseARM64:    false,
	AVFRINT32X - obj.ABaseARM64:   false,
	AVFRINT32Z - obj.ABaseARM64:   false,
	AVFRINT64X - obj.ABaseARM64:   false,
	AVFRINT64Z - obj.ABaseARM64:   false,
	AVFRINTAH - obj.ABaseARM64:    false,
	AVFRINTAS - obj.ABaseARM64:    false,
	AVFRINTIH - obj.ABaseARM64:    false,
	AVFRINTIS - obj.ABaseARM64:    false,
	AVFRINTMH - obj.ABaseARM64:    false,
	AVFRINTMS - obj.ABaseARM64:    false,
	AVFRINTNH - obj.ABaseARM64:    false,
	AVFRINTNS - obj.ABaseARM64:    false,
	AVFRINTPH - obj.ABaseARM64:    false,
	AVFRINTPS - obj.ABaseARM64:    false,
	AVFRINTXH - obj.ABaseARM64:    false,
	AVFRINTXS - obj.ABaseARM64:    false,
	AVFRINTZH - obj.ABaseARM64:    false,
	AVFRINTZS - obj.ABaseARM64:    false,
	AVFRSQRTE - obj.ABaseARM64:    false,
	AVFRSQRTEH - obj.ABaseARM64:   false,
	AVFRSQRTS - obj.ABaseARM64:    false,
	AVFRSQRTSH - obj.ABaseARM64:   false,
	AVFSQRTH - obj.ABaseARM64:     false,
	AVFSQRTS - obj.ABaseARM64:     false,
	AVFSUBH - obj.ABaseARM64:      false,
	AVFSUBS - obj.ABaseARM64:      false,
	AVINS - obj.ABaseARM64:        false,
	AVLD1 - obj.ABaseARM64:        true,
	AVLD1B - obj.ABaseARM64:       false,
	AVLD1D - obj.ABaseARM64:       false,
	AVLD1H - obj.ABaseARM64:       false,
	AVLD1R - obj.ABaseARM64:       true,
	AVLD1S - obj.ABaseARM64:       false,
	AVLD2 - obj.ABaseARM64:        true,
	AVLD2B - obj.ABaseARM64:       false,
	AVLD2D - obj.ABaseARM64:       false,
	AVLD2H - obj.ABaseARM64:       false,
	AVLD2R - obj.ABaseARM64:       true,
	AVLD2S - obj.ABaseARM64:       false,
	AVLD3 - obj.ABaseARM64:        true,
	AVLD3B - obj.ABaseARM64:       false,
	AVLD3D - obj.ABaseARM64:       false,
	AVLD3H - obj.ABaseARM64:       false,
	AVLD3R - obj.ABaseARM64:       true,
	AVLD3S - obj.ABaseARM64:       false,
	AVLD4 - obj.ABaseARM64:        true,
	AVLD4B - obj.ABaseARM64:       false,
	AVLD4D - obj.ABaseARM64:       false,
	AVLD4H - obj.ABaseARM64:       false,
	AVLD4R - obj.ABaseARM64:       true,
	AVLD4S - obj.ABaseARM64:       false,
	AVLDAP1D - obj.ABaseARM64:     false,
	AVMLA - obj.ABaseARM64:        false,
	AVMLS - obj.ABaseARM64:        false,
	AVMOV - obj.ABaseARM64:        true,
	AVMOVD - obj.ABaseARM64:       true,
	AVMOVI - obj.ABaseARM64:       true,
	AVMOVQ - obj.ABaseARM64:       true,
	AVMOVS - obj.ABaseARM64:       true,
	AVMUL - obj.ABaseARM64:        false,
	AVMVN - obj.ABaseARM64:        false,
	AVMVNIH - obj.ABaseARM64:      false,
	AVMVNIS - obj.ABaseARM64:      false,
	AVNEG - obj.ABaseARM64:        false,
	AVNOT - obj.ABaseARM64:        false,
	AVORN - obj.ABaseARM64:        false,
	AVORR - obj.ABaseARM64:        true,
	AVORRH - obj.ABaseARM64:       false,
	AVORRS - obj.ABaseARM64:       false,
	AVPMUL - obj.ABaseARM64:       false,
	AVPMULL - obj.ABaseARM64:      true,
	AVPMULL2 - obj.ABaseARM64:     true,
	AVRADDHN - obj.ABaseARM64:     false,
	AVRADDHN2 - obj.ABaseARM64:    false,
	AVRAX1 - obj.ABaseARM64:       true,
	AVRBIT - obj.ABaseARM64:       true,
	AVREV16 - obj.ABaseARM64:      true,
	AVREV32 - obj.ABaseARM64:      true,
	AVREV64 - obj.ABaseARM64:      true,
	AVRSHRN - obj.ABaseARM64:      false,
	AVRSHRN2 - obj.ABaseARM64:     false,
	AVRSUBHN - obj.ABaseARM64:     false,
	AVRSUBHN2 - obj.ABaseARM64:    false,
	AVSABA - obj.ABaseARM64:       false,
	AVSABAL - obj.ABaseARM64:      false,
	AVSABAL2 - obj.ABaseARM64:     false,
	AVSABD - obj.ABaseARM64:       false,
	AVSABDL - obj.ABaseARM64:      false,
	AVSABDL2 - obj.ABaseARM64:     false,
	AVSADALP - obj.ABaseARM64:     false,
	AVSADDL - obj.ABaseARM64:      false,
	AVSADDL2 - obj.ABaseARM64:     false,
	AVSADDLP - obj.ABaseARM64:     false,
	AVSADDLV - obj.ABaseARM64:     false,
	AVSADDW - obj.ABaseARM64:      false,
	AVSADDW2 - obj.ABaseARM64:     false,
	AVSCVTF - obj.ABaseARM64:      false,
	AVSCVTFH - obj.ABaseARM64:     false,
	AVSDOT - obj.ABaseARM64:       false,
	AVSHADD - obj.ABaseARM64:      false,
	AVSHL - obj.ABaseARM64:        true,
	AVSHLL - obj.ABaseARM64:       false,
	AVSHLL2 - obj.ABaseARM64:      false,
	AVSHRN - obj.ABaseARM64:       false,
	AVSHRN2 - obj.ABaseARM64:      false,
	AVSHSUB - obj.ABaseARM64:      false,
	AVSLI - obj.ABaseARM64:        true,
	AVSMAX - obj.ABaseARM64:       false,
	AVSMAXP - obj.ABaseARM64:      false,
	AVSMAXV - obj.ABaseARM64:      false,
	AVSMIN - obj.ABaseARM64:       false,
	AVSMINP - obj.ABaseARM64:      false,
	AVSMINV - obj.ABaseARM64:      false,
	AVSMLAL - obj.ABaseARM64:      false,
	AVSMLAL2 - obj.ABaseARM64:     false,
	AVSMLSL - obj.ABaseARM64:      false,
	AVSMLSL2 - obj.ABaseARM64:     false,
	AVSMMLA - obj.ABaseARM64:      false,
	AVSMOVD - obj.ABaseARM64:      false,
	AVSMOVS - obj.ABaseARM64:      false,
	AVSMULL - obj.ABaseARM64:      false,
	AVSMULL2 - obj.ABaseARM64:     false,
	AVSQABS - obj.ABaseARM64:      false,
	AVSQADD - obj.ABaseARM64:      false,
	AVSQDMLAL - obj.ABaseARM64:    false,
	AVSQDMLAL2 - obj.ABaseARM64:   false,
	AVSQDMLSL - obj.ABaseARM64:    false,
	AVSQDMLSL2 - obj.ABaseARM64:   false,
	AVSQDMULH - obj.ABaseARM64:    false,
	AVSQDMULL - obj.ABaseARM64:    false,
	AVSQDMULL2 - obj.ABaseARM64:   false,
	AVSQNEG - obj.ABaseARM64:      false,
	AVSQRDMLAH - obj.ABaseARM64:   false,
	AVSQRDMLSH - obj.ABaseARM64:   false,
	AVSQRDMULH - obj.ABaseARM64:   false,
	AVSQRSHL - obj.ABaseARM64:     false,
	AVSQRSHRN - obj.ABaseARM64:    false,
	AVSQRSHRN2 - obj.ABaseARM64:   false,
	AVSQRSHRUN - obj.ABaseARM64:   false,
	AVSQRSHRUN2 - obj.ABaseARM64:  false,
	AVSQSHL - obj.ABaseARM64:      false,
	AVSQSHLU - obj.ABaseARM64:     false,
	AVSQSHRN - obj.ABaseARM64:     false,
	AVSQSHRN2 - obj.ABaseARM64:    false,
	AVSQSHRUN - obj.ABaseARM64:    false,
	AVSQSHRUN2 - obj.ABaseARM64:   false,
	AVSQSUB - obj.ABaseARM64:      false,
	AVSQXTN - obj.ABaseARM64:      false,
	AVSQXTN2 - obj.ABaseARM64:     false,
	AVSQXTUN - obj.ABaseARM64:     false,
	AVSQXTUN2 - obj.ABaseARM64:    false,
	AVSRHADD - obj.ABaseARM64:     false,
	AVSRI - obj.ABaseARM64:        true,
	AVSRSHL - obj.ABaseARM64:      false,
	AVSRSHR - obj.ABaseARM64:      false,
	AVSRSRA - obj.ABaseARM64:      false,
	AVSSHL - obj.ABaseARM64:       false,
	AVSSHLL - obj.ABaseARM64:      false,
	AVSSHLL2 - obj.ABaseARM64:     false,
	AVSSHR - obj.ABaseARM64:       false,
	AVSSRA - obj.ABaseARM64:       false,
	AVSSUBL - obj.ABaseARM64:      false,
	AVSSUBL2 - obj.ABaseARM64:     false,
	AVSSUBW - obj.ABaseARM64:      false,
	AVSSUBW2 - obj.ABaseARM64:     false,
	AVST1 - obj.ABaseARM64:        true,
	AVST1B - obj.ABaseARM64:       false,
	AVST1D - obj.ABaseARM64:       false,
	AVST1H - obj.ABaseARM64:       false,
	AVST1S - obj.ABaseARM64:       false,
	AVST2 - obj.ABaseARM64:        true,
	AVST2B - obj.ABaseARM64:       false,
	AVST2D - obj.ABaseARM64:       false,
	AVST2H - obj.ABaseARM64:       false,
	AVST2S - obj.ABaseARM64:       false,
	AVST3 - obj.ABaseARM64:        true,
	AVST3B - obj.ABaseARM64:       false,
	AVST3D - obj.ABaseARM64:       false,
	AVST3H - obj.ABaseARM64:       false,
	AVST3S - obj.ABaseARM64:       false,
	AVST4 - obj.ABaseARM64:        true,
	AVST4B - obj.ABaseARM64:       false,
	AVST4D - obj.ABaseARM64:       false,
	AVST4H - obj.ABaseARM64:       false,
	AVST4S - obj.ABaseARM64:       false,
	AVSTL1D - obj.ABaseARM64:      false,
	AVSUB - obj.ABaseARM64:        true,
	AVSUBHN - obj.ABaseARM64:      false,
	AVSUBHN2 - obj.ABaseARM64:     false,
	AVSUDOT - obj.ABaseARM64:      false,
	AVSUQADD - obj.ABaseARM64:     false,
	AVSXTL - obj.ABaseARM64:       false,
	AVSXTL2 - obj.ABaseARM64:      false,
	AVTBL - obj.ABaseARM64:        true,
	AVTBX - obj.ABaseARM64:        true,
	AVTRN1 - obj.ABaseARM64:       true,
	AVTRN2 - obj.ABaseARM64:       true,
	AVUABA - obj.ABaseARM64:       false,
	AVUABAL - obj.ABaseARM64:      false,
	AVUABAL2 - obj.ABaseARM64:     false,
	AVUABD - obj.ABaseARM64:       false,
	AVUABDL - obj.ABaseARM64:      false,
	AVUABDL2 - obj.ABaseARM64:     false,
	AVUADALP - obj.ABaseARM64:     false,
	AVUADDL - obj.ABaseARM64:      false,
	AVUADDL2 - obj.ABaseARM64:     false,
	AVUADDLP - obj.ABaseARM64:     false,
	AVUADDLV - obj.ABaseARM64:     true,
	AVUADDW - obj.ABaseARM64:      true,
	AVUADDW2 - obj.ABaseARM64:     true,
	AVUCVTF - obj.ABaseARM64:      false,
	AVUCVTFH - obj.ABaseARM64:     false,
	AVUDOT - obj.ABaseARM64:       false,
	AVUHADD - obj.ABaseARM64:      false,
	AVUHSUB - obj.ABaseARM64:      false,
	AVUMAX - obj.ABaseARM64:       true,
	AVUMAXP - obj.ABaseARM64:      false,
	AVUMAXV - obj.ABaseARM64:      false,
	AVUMIN - obj.ABaseARM64:       true,
	AVUMINP - obj.ABaseARM64:      false,
	AVUMINV - obj.ABaseARM64:      false,
	AVUMLAL - obj.ABaseARM64:      false,
	AVUMLAL2 - obj.ABaseARM64:     false,
	AVUMLSL - obj.ABaseARM64:      false,
	AVUMLSL2 - obj.ABaseARM64:     false,
	AVUMMLA - obj.ABaseARM64:      false,
	AVUMULL - obj.ABaseARM64:      false,
	AVUMULL2 - obj.ABaseARM64:     false,
	AVUQADD - obj.ABaseARM64:      false,
	AVUQRSHL - obj.ABaseARM64:     false,
	AVUQRSHRN - obj.ABaseARM64:    false,
	AVUQRSHRN2 - obj.ABaseARM64:   false,
	AVUQSHL - obj.ABaseARM64:      false,
	AVUQSHRN - obj.ABaseARM64:     false,
	AVUQSHRN2 - obj.ABaseARM64:    false,
	AVUQSUB - obj.ABaseARM64:      false,
	AVUQXTN - obj.ABaseARM64:      false,
	AVUQXTN2 - obj.ABaseARM64:     false,
	AVURECPE - obj.ABaseARM64:     false,
	AVURHADD - obj.ABaseARM64:     false,
	AVURSHL - obj.ABaseARM64:      false,
	AVURSHR - obj.ABaseARM64:      false,
	AVURSQRTE - obj.ABaseARM64:    false,
	AVURSRA - obj.ABaseARM64:      false,
	AVUSDOT - obj.ABaseARM64:      false,
	AVUSHL - obj.ABaseARM64:       false,
	AVUSHLL - obj.ABaseARM64:      true,
	AVUSHLL2 - obj.ABaseARM64:     true,
	AVUSHR - obj.ABaseARM64:       true,
	AVUSMMLA - obj.ABaseARM64:     false,
	AVUSQADD - obj.ABaseARM64:     false,
	AVUSRA - obj.ABaseARM64:       true,
	AVUSUBL - obj.ABaseARM64:      false,
	AVUSUBL2 - obj.ABaseARM64:     false,
	AVUSUBW - obj.ABaseARM64:      false,
	AVUSUBW2 - obj.ABaseARM64:     false,
	AVUXTL - obj.ABaseARM64:       true,
	AVUXTL2 - obj.ABaseARM64:      true,
	AVUZP1 - obj.ABaseARM64:       true,
	AVUZP2 - obj.ABaseARM64:       true,
	AVXAR - obj.ABaseARM64:        true,
	AVXTN - obj.ABaseARM64:        false,
	AVXTN2 - obj.ABaseARM64:       false,
	AVZIP1 - obj.ABaseARM64:       true,
	AVZIP2 - obj.ABaseARM64:       true,
	AWFE - obj.ABaseARM64:         true,
	AWFET - obj.ABaseARM64:        false,
	AWFI - obj.ABaseARM64:         true,
	AWFIT - obj.ABaseARM64:        false,
	AWHILEGE - obj.ABaseARM64:     false,
	AWHILEGEW - obj.ABaseARM64:    false,
	AWHILEGT - obj.ABaseARM64:     false,
	AWHILEGTW - obj.ABaseARM64:    false,
	AWHILEHI - obj.ABaseARM64:     false,
	AWHILEHIW - obj.ABaseARM64:    false,
	AWHILEHS - obj.ABaseARM64:     false,
	AWHILEHSW - obj.ABaseARM64:    false,
	AWHILELE - obj.ABaseARM64:     false,
	AWHILELEW - obj.ABaseARM64:    false,
	AWHILELO - obj.ABaseARM64:     false,
	AWHILELOW - obj.ABaseARM64:    false,
	AWHILELS - obj.ABaseARM64:     false,
	AWHILELSW - obj.ABaseARM64:    false,
	AWHILELT - obj.ABaseARM64:     false,
	AWHILELTW - obj.ABaseARM64:    false,
	AWORD - obj.ABaseARM64:        true,
	AXAFLAG - obj.ABaseARM64:      false,
	AXPACD - obj.ABaseARM64:       false,
	AXPACI - obj.ABaseARM64:       false,
	AXPACLRI - obj.ABaseARM64:     false,
	AYIELD - obj.ABaseARM64:       true,
	AZABS - obj.ABaseARM64:        false,
	AZADCLB - obj.ABaseARM64:      false,
	AZADCLT - obj.ABaseARM64:      false,
	AZADD - obj.ABaseARM64:        false,
	AZADDHAD - obj.ABaseARM64:     false,
	AZADDHAS - obj.ABaseARM64:     false,
	AZADDHNB - obj.ABaseARM64:     false,
	AZADDHNT - obj.ABaseARM64:     false,
	AZADDP - obj.ABaseARM64:       false,
	AZADDQV - obj.ABaseARM64:      false,
	AZADDVAD - obj.ABaseARM64:     false,
	AZADDVAS - obj.ABaseARM64:     false,
	AZADR - obj.ABaseARM64:        false,
	AZAND - obj.ABaseARM64:        false,
	AZANDQV - obj.ABaseARM64:      false,
	AZANDV - obj.ABaseARM64:       false,
	AZASR - obj.ABaseARM64:        false,
	AZASRD - obj.ABaseARM64:       false,
	AZASRR - obj.ABaseARM64:       false,
	AZBCAX - obj.ABaseARM64:       false,
	AZBDEP - obj.ABaseARM64:       false,
	AZBEXT - obj.ABaseARM64:       false,
	AZBFADD - obj.ABaseARM64:      false,
	AZBFCLAMP - obj.ABaseARM64:    false,
	AZBFCVT - obj.ABaseARM64:      false,
	AZBFCVTN - obj.ABaseARM64:     false,
	AZBFCVTNT - obj.ABaseARM64:    false,
	AZBFDOT - obj.ABaseARM64:      false,
	AZBFMAX - obj.ABaseARM64:      false,
	AZBFMAXNM - obj.ABaseARM64:    false,
	AZBFMIN - obj.ABaseARM64:      false,
	AZBFMINNM - obj.ABaseARM64:    false,
	AZBFMLA - obj.ABaseARM64:      false,
	AZBFMLAL - obj.ABaseARM64:     false,
	AZBFMLALB - obj.ABaseARM64:    false,
	AZBFMLALT - obj.ABaseARM64:    false,
	AZBFMLS - obj.ABaseARM64:      false,
	AZBFMLSL - obj.ABaseARM64:     false,
	AZBFMLSLB - obj.ABaseARM64:    false,
	AZBFMLSLT - obj.ABaseARM64:    false,
	AZBFMMLA - obj.ABaseARM64:     false,
	AZBFMOPA - obj.ABaseARM64:     false,
	AZBFMOPS - obj.ABaseARM64:     false,
	AZBFMUL - obj.ABaseARM64:      false,
	AZBFSUB - obj.ABaseARM64:      false,
	AZBFVDOT - obj.ABaseARM64:     false,
	AZBGRP - obj.ABaseARM64:       false,
	AZBIC - obj.ABaseARM64:        false,
	AZBMOPA - obj.ABaseARM64:      false,
	AZBMOPS - obj.ABaseARM64:      false,
	AZBRKA - obj.ABaseARM64:       false,
	AZBRKB - obj.ABaseARM64:       false,
	AZBSL - obj.ABaseARM64:        false,
	AZBSL1N - obj.ABaseARM64:      false,
	AZBSL2N - obj.ABaseARM64:      false,
	AZCADD - obj.ABaseARM64:       false,
	AZCDOT - obj.ABaseARM64:       false,
	AZCDOTD - obj.ABaseARM64:      false,
	AZCDOTS - obj.ABaseARM64:      false,
	AZCLASTA - obj.ABaseARM64:     false,
	AZCLASTB - obj.ABaseARM64:     false,
	AZCLS - obj.ABaseARM64:        false,
	AZCLZ - obj.ABaseARM64:        false,
	AZCMLA - obj.ABaseARM64:       false,
	AZCMLAH - obj.ABaseARM64:      false,
	AZCMLAS - obj.ABaseARM64:      false,
	AZCMPEQ - obj.ABaseARM64:      false,
	AZCMPGE - obj.ABaseARM64:      false,
	AZCMPGT - obj.ABaseARM64:      false,
	AZCMPHI - obj.ABaseARM64:      false,
	AZCMPHS - obj.ABaseARM64:      false,
	AZCMPLE - obj.ABaseARM64:      false,
	AZCMPLO - obj.ABaseARM64:      false,
	AZCMPLS - obj.ABaseARM64:      false,
	AZCMPLT - obj.ABaseARM64:      false,
	AZCMPNE - obj.ABaseARM64:      false,
	AZCNOT - obj.ABaseARM64:       false,
	AZCNT - obj.ABaseARM64:        false,
	AZCOMPACT - obj.ABaseARM64:    false,
	AZCPY - obj.ABaseARM64:        false,
	AZDECD - obj.ABaseARM64:       false,
	AZDECH - obj.ABaseARM64:       false,
	AZDECP - obj.ABaseARM64:       false,
	AZDECW - obj.ABaseARM64:       false,
	AZDUP - obj.ABaseARM64:        false,
	AZDUPM - obj.ABaseARM64:       false,
	AZDUPQ - obj.ABaseARM64:       false,
	AZEON - obj.ABaseARM64:        false,
	AZEOR - obj.ABaseARM64:        false,
	AZEOR3 - obj.ABaseARM64:       false,
	AZEORBT - obj.ABaseARM64:      false,
	AZEORQV - obj.ABaseARM64:      false,
	AZEORTB - obj.ABaseARM64:      false,
	AZEORV - obj.ABaseARM64:       false,
	AZERO - obj.ABaseARM64:        false,
	AZEXT - obj.ABaseARM64:        false,
	AZEXTQ - obj.ABaseARM64:       false,
	AZFABD - obj.ABaseARM64:       false,
	AZFABS - obj.ABaseARM64:       false,
	AZFACGE - obj.ABaseARM64:      false,
	AZFACGT - obj.ABaseARM64:      false,
	AZFACLE - obj.ABaseARM64:      false,
	AZFACLT - obj.ABaseARM64:      false,
	AZFADD - obj.ABaseARM64:       false,
	AZFADDA - obj.ABaseARM64:      false,
	AZFADDH - obj.ABaseARM64:      false,
	AZFADDP - obj.ABaseARM64:      false,
	AZFADDQV - obj.ABaseARM64:     false,
	AZFADDV - obj.ABaseARM64:      false,
	AZFCADD - obj.ABaseARM64:      false,
	AZFCLAMP - obj.ABaseARM64:     false,
	AZFCMEQ - obj.ABaseARM64:      false,
	AZFCMGE - obj.ABaseARM64:      false,
	AZFCMGT - obj.ABaseARM64:      false,
	AZFCMLA - obj.ABaseARM64:      false,
	AZFCMLAH - obj.ABaseARM64:     false,
	AZFCMLAS - obj.ABaseARM64:     false,
	AZFCMLE - obj.ABaseARM64:      false,
	AZFCMLT - obj.ABaseARM64:      false,
	AZFCMNE - obj.ABaseARM64:      false,
	AZFCMUO - obj.ABaseARM64:      false,
	AZFCPY - obj.ABaseARM64:       false,
	AZFCVT - obj.ABaseARM64:       false,
	AZFCVTDH - obj.ABaseARM64:     false,
	AZFCVTDS - obj.ABaseARM64:     false,
	AZFCVTHD - obj.ABaseARM64:     false,
	AZFCVTHS - obj.ABaseARM64:     false,
	AZFCVTL - obj.ABaseARM64:      false,
	AZFCVTLTHS - obj.ABaseARM64:   false,
	AZFCVTLTSD - obj.ABaseARM64:   false,
	AZFCVTN - obj.ABaseARM64:      false,
	AZFCVTNTDS - obj.ABaseARM64:   false,
	AZFCVTNTSH - obj.ABaseARM64:   false,
	AZFCVTSD - obj.ABaseARM64:     false,
	AZFCVTSH - obj.ABaseARM64:     false,
	AZFCVTXDS - obj.ABaseARM64:    false,
	AZFCVTXNTDS - obj.ABaseARM64:  false,
	AZFCVTZS - obj.ABaseARM64:     false,
	AZFCVTZSD - obj.ABaseARM64:    false,
	AZFCVTZSDW - obj.ABaseARM64:   false,
	AZFCVTZSH - obj.ABaseARM64:    false,
	AZFCVTZSHH - obj.ABaseARM64:   false,
	AZFCVTZSHW - obj.ABaseARM64:   false,
	AZFCVTZSS - obj.ABaseARM64:    false,
	AZFCVTZSSW - obj.ABaseARM64:   false,
	AZFCVTZU - obj.ABaseARM64:     false,
	AZFCVTZUD - obj.ABaseARM64:    false,
	AZFCVTZUDW - obj.ABaseARM64:   false,
	AZFCVTZUH - obj.ABaseARM64:    false,
	AZFCVTZUHH - obj.ABaseARM64:   false,
	AZFCVTZUHW - obj.ABaseARM64:   false,
	AZFCVTZUS - obj.ABaseARM64:    false,
	AZFCVTZUSW - obj.ABaseARM64:   false,
	AZFDIV - obj.ABaseARM64:       false,
	AZFDIVR - obj.ABaseARM64:      false,
	AZFDOT - obj.ABaseARM64:       false,
	AZFDUP - obj.ABaseARM64:       false,
	AZFEXPA - obj.ABaseARM64:      false,
	AZFLOGB - obj.ABaseARM64:      false,
	AZFMAD - obj.ABaseARM64:       false,
	AZFMAX - obj.ABaseARM64:       false,
	AZFMAXNM - obj.ABaseARM64:     false,
	AZFMAXNMP - obj.ABaseARM64:    false,
	AZFMAXNMQV - obj.ABaseARM64:   false,
	AZFMAXNMV - obj.ABaseARM64:    false,
	AZFMAXP - obj.ABaseARM64:      false,
	AZFMAXQV - obj.ABaseARM64:     false,
	AZFMAXV - obj.ABaseARM64:      false,
	AZFMIN - obj.ABaseARM64:       false,
	AZFMINNM - obj.ABaseARM64:     false,
	AZFMINNMP - obj.ABaseARM64:    false,
	AZFMINNMQV - obj.ABaseARM64:   false,
	AZFMINNMV - obj.ABaseARM64:    false,
	AZFMINP - obj.ABaseARM64:      false,
	AZFMINQV - obj.ABaseARM64:     false,
	AZFMINV - obj.ABaseARM64:      false,
	AZFMLA - obj.ABaseARM64:       false,
	AZFMLAD - obj.ABaseARM64:      false,
	AZFMLAH - obj.ABaseARM64:      false,
	AZFMLAL - obj.ABaseARM64:      false,
	AZFMLALB - obj.ABaseARM64:     false,
	AZFMLALBS - obj.ABaseARM64:    false,
	AZFMLALT - obj.ABaseARM64:     false,
	AZFMLALTS - obj.ABaseARM64:    false,
	AZFMLAS - obj.ABaseARM64:      false,
	AZFMLS - obj.ABaseARM64:       false,
	AZFMLSD - obj.ABaseARM64:      false,
	AZFMLSH - obj.ABaseARM64:      false,
	AZFMLSL - obj.ABaseARM64:      false,
	AZFMLSLB - obj.ABaseARM64:     false,
	AZFMLSLBS - obj.ABaseARM64:    false,
	AZFMLSLT - obj.ABaseARM64:     false,
	AZFMLSLTS - obj.ABaseARM64:    false,
	AZFMLSS - obj.ABaseARM64:      false,
	AZFMMLAD - obj.ABaseARM64:     false,
	AZFMMLAS - obj.ABaseARM64:     false,
	AZFMOPA - obj.ABaseARM64:      false,
	AZFMOPAD - obj.ABaseARM64:     false,
	AZFMOPAH - obj.ABaseARM64:     false,
	AZFMOPAS - obj.ABaseARM64:     false,
	AZFMOPS - obj.ABaseARM64:      false,
	AZFMOPSD - obj.ABaseARM64:     false,
	AZFMOPSH - obj.ABaseARM64:     false,
	AZFMOPSS - obj.ABaseARM64:     false,
	AZFMOV - obj.ABaseARM64:       false,
	AZFMSB - obj.ABaseARM64:       false,
	AZFMUL - obj.ABaseARM64:       false,
	AZFMULD - obj.ABaseARM64:      false,
	AZFMULH - obj.ABaseARM64:      false,
	AZFMULS - obj.ABaseARM64:      false,
	AZFMULX - obj.ABaseARM64:      false,
	AZFNEG - obj.ABaseARM64:       false,
	AZFNMAD - obj.ABaseARM64:      false,
	AZFNMLA - obj.ABaseARM64:      false,
	AZFNMLS - obj.ABaseARM64:      false,
	AZFNMSB - obj.ABaseARM64:      false,
	AZFRECPE - obj.ABaseARM64:     false,
	AZFRECPS - obj.ABaseARM64:     false,
	AZFRECPX - obj.ABaseARM64:     false,
	AZFRINTA - obj.ABaseARM64:     false,
	AZFRINTI - obj.ABaseARM64:     false,
	AZFRINTM - obj.ABaseARM64:     false,
	AZFRINTN - obj.ABaseARM64:     false,
	AZFRINTP - obj.ABaseARM64:     false,
	AZFRINTX - obj.ABaseARM64:     false,
	AZFRINTZ - obj.ABaseARM64:     false,
	AZFRSQRTE - obj.ABaseARM64:    false,
	AZFRSQRTS - obj.ABaseARM64:    false,
	AZFSCALE - obj.ABaseARM64:     false,
	AZFSQRT - obj.ABaseARM64:      false,
	AZFSUB - obj.ABaseARM64:       false,
	AZFSUBH - obj.ABaseARM64:      false,
	AZFSUBR - obj.ABaseARM64:      false,
	AZFTMAD - obj.ABaseARM64:      false,
	AZFTSMUL - obj.ABaseARM64:     false,
	AZFTSSEL - obj.ABaseARM64:     false,
	AZFVDOT - obj.ABaseARM64:      false,
	AZHISTCNT - obj.ABaseARM64:    false,
	AZHISTSEG - obj.ABaseARM64:    false,
	AZINCD - obj.ABaseARM64:       false,
	AZINCH - obj.ABaseARM64:       false,
	AZINCP - obj.ABaseARM64:       false,
	AZINCW - obj.ABaseARM64:       false,
	AZINDEX - obj.ABaseARM64:      false,
	AZINSR - obj.ABaseARM64:       false,
	AZLASTA - obj.ABaseARM64:      false,
	AZLASTB - obj.ABaseARM64:      false,
	AZLD1B - obj.ABaseARM64:       false,
	AZLD1BB - obj.ABaseARM64:      false,
	AZLD1BD - obj.ABaseARM64:      false,
	AZLD1BH - obj.ABaseARM64:      false,
	AZLD1BS - obj.ABaseARM64:      false,
	AZLD1D - obj.ABaseARM64:       false,
	AZLD1DD - obj.ABaseARM64:      false,
	AZLD1DS - obj.ABaseARM64:      false,
	AZLD1H - obj.ABaseARM64:       false,
	AZLD1HD - obj.ABaseARM64:      false,
	AZLD1HH - obj.ABaseARM64:      false,
	AZLD1HS - obj.ABaseARM64:      false,
	AZLD1Q - obj.ABaseARM64:       false,
	AZLD1RBB - obj.ABaseARM64:     false,
	AZLD1RBD - obj.ABaseARM64:     false,
	AZLD1RBH - obj.ABaseARM64:     false,
	AZLD1RBS - obj.ABaseARM64:     false,
	AZLD1RD - obj.ABaseARM64:      false,
	AZLD1RHD - obj.ABaseARM64:     false,
	AZLD1RHH - obj.ABaseARM64:     false,
	AZLD1RHS - obj.ABaseARM64:     false,
	AZLD1ROB - obj.ABaseARM64:     false,
	AZLD1ROD - obj.ABaseARM64:     false,
	AZLD1ROH - obj.ABaseARM64:     false,
	AZLD1ROW - obj.ABaseARM64:     false,
	AZLD1RQB - obj.ABaseARM64:     false,
	AZLD1RQD - obj.ABaseARM64:     false,
	AZLD1RQH - obj.ABaseARM64:     false,
	AZLD1RQW - obj.ABaseARM64:     false,
	AZLD1RSBD - obj.ABaseARM64:    false,
	AZLD1RSBH - obj.ABaseARM64:    false,
	AZLD1RSBS - obj.ABaseARM64:    false,
	AZLD1RSHD - obj.ABaseARM64:    false,
	AZLD1RSHS - obj.ABaseARM64:    false,
	AZLD1RSW - obj.ABaseARM64:     false,
	AZLD1RWD - obj.ABaseARM64:     false,
	AZLD1RWS - obj.ABaseARM64:     false,
	AZLD1SBD - obj.ABaseARM64:     false,
	AZLD1SBH - obj.ABaseARM64:     false,
	AZLD1SBS - obj.ABaseARM64:     false,
	AZLD1SHD - obj.ABaseARM64:     false,
	AZLD1SHS - obj.ABaseARM64:     false,
	AZLD1SW - obj.ABaseARM64:      false,
	AZLD1SWD - obj.ABaseARM64:     false,
	AZLD1SWS - obj.ABaseARM64:     false,
	AZLD1W - obj.ABaseARM64:       false,
	AZLD1WD - obj.ABaseARM64:      false,
	AZLD1WQ - obj.ABaseARM64:      false,
	AZLD1WS - obj.ABaseARM64:      false,
	AZLD2B - obj.ABaseARM64:       false,
	AZLD2D - obj.ABaseARM64:       false,
	AZLD2H - obj.ABaseARM64:       false,
	AZLD2Q - obj.ABaseARM64:       false,
	AZLD2W - obj.ABaseARM64:       false,
	AZLD3B - obj.ABaseARM64:       false,
	AZLD3D - obj.ABaseARM64:       false,
	AZLD3H - obj.ABaseARM64:       false,
	AZLD3Q - obj.ABaseARM64:       false,
	AZLD3W - obj.ABaseARM64:       false,
	AZLD4B - obj.ABaseARM64:       false,
	AZLD4D - obj.ABaseARM64:       false,
	AZLD4H - obj.ABaseARM64:       false,
	AZLD4Q - obj.ABaseARM64:       false,
	AZLD4W - obj.ABaseARM64:       false,
	AZLDFF1BB - obj.ABaseARM64:    false,
	AZLDFF1BD - obj.ABaseARM64:    false,
	AZLDFF1BH - obj.ABaseARM64:    false,
	AZLDFF1BS - obj.ABaseARM64:    false,
	AZLDFF1D - obj.ABaseARM64:     false,
	AZLDFF1DD - obj.ABaseARM64:    false,
	AZLDFF1DS - obj.ABaseARM64:    false,
	AZLDFF1HD - obj.ABaseARM64:    false,
	AZLDFF1HH - obj.ABaseARM64:    false,
	AZLDFF1HS - obj.ABaseARM64:    false,
	AZLDFF1SBD - obj.ABaseARM64:   false,
	AZLDFF1SBH - obj.ABaseARM64:   false,
	AZLDFF1SBS - obj.ABaseARM64:   false,
	AZLDFF1SHD - obj.ABaseARM64:   false,
	AZLDFF1SHS - obj.ABaseARM64:   false,
	AZLDFF1SW - obj.ABaseARM64:    false,
	AZLDFF1SWD - obj.ABaseARM64:   false,
	AZLDFF1SWS - obj.ABaseARM64:   false,
	AZLDFF1WD - obj.ABaseARM64:    false,
	AZLDFF1WS - obj.ABaseARM64:    false,
	AZLDNF1BB - obj.ABaseARM64:    false,
	AZLDNF1BD - obj.ABaseARM64:    false,
	AZLDNF1BH - obj.ABaseARM64:    false,
	AZLDNF1BS - obj.ABaseARM64:    false,
	AZLDNF1D - obj.ABaseARM64:     false,
	AZLDNF1HD - obj.ABaseARM64:    false,
	AZLDNF1HH - obj.ABaseARM64:    false,
	AZLDNF1HS - obj.ABaseARM64:    false,
	AZLDNF1SBD - obj.ABaseARM64:   false,
	AZLDNF1SBH - obj.ABaseARM64:   false,
	AZLDNF1SBS - obj.ABaseARM64:   false,
	AZLDNF1SHD - obj.ABaseARM64:   false,
	AZLDNF1SHS - obj.ABaseARM64:   false,
	AZLDNF1SW - obj.ABaseARM64:    false,
	AZLDNF1WD - obj.ABaseARM64:    false,
	AZLDNF1WS - obj.ABaseARM64:    false,
	AZLDNT1B - obj.ABaseARM64:     false,
	AZLDNT1BD - obj.ABaseARM64:    false,
	AZLDNT1BS - obj.ABaseARM64:    false,
	AZLDNT1D - obj.ABaseARM64:     false,
	AZLDNT1H - obj.ABaseARM64:     false,
	AZLDNT1HD - obj.ABaseARM64:    false,
	AZLDNT1HS - obj.ABaseARM64:    false,
	AZLDNT1SBD - obj.ABaseARM64:   false,
	AZLDNT1SBS - obj.ABaseARM64:   false,
	AZLDNT1SHD - obj.ABaseARM64:   false,
	AZLDNT1SHS - obj.ABaseARM64:   false,
	AZLDNT1SW - obj.ABaseARM64:    false,
	AZLDNT1W - obj.ABaseARM64:     false,
	AZLDNT1WD - obj.ABaseARM64:    false,
	AZLDNT1WS - obj.ABaseARM64:    false,
	AZLDR - obj.ABaseARM64:        false,
	AZLSL - obj.ABaseARM64:        false,
	AZLSLR - obj.ABaseARM64:       false,
	AZLSR - obj.ABaseARM64:        false,
	AZLSRR - obj.ABaseARM64:       false,
	AZLUTI2 - obj.ABaseARM64:      false,
	AZLUTI4 - obj.ABaseARM64:      false,
	AZMAD - obj.ABaseARM64:        false,
	AZMATCH - obj.ABaseARM64:      false,
	AZMLA - obj.ABaseARM64:        false,
	AZMLAD - obj.ABaseARM64:       false,
	AZMLAH - obj.ABaseARM64:       false,
	AZMLAS - obj.ABaseARM64:       false,
	AZMLS - obj.ABaseARM64:        false,
	AZMLSD - obj.ABaseARM64:       false,
	AZMLSH - obj.ABaseARM64:       false,
	AZMLSS - obj.ABaseARM64:       false,
	AZMOV - obj.ABaseARM64:        false,
	AZMOVA - obj.ABaseARM64:       false,
	AZMOVAB - obj.ABaseARM64:      false,
	AZMOVAD - obj.ABaseARM64:      false,
	AZMOVAH - obj.ABaseARM64:      false,
	AZMOVAQ - obj.ABaseARM64:      false,
	AZMOVAS - obj.ABaseARM64:      false,
	AZMOVAZ - obj.ABaseARM64:      false,
	AZMOVAZB - obj.ABaseARM64:     false,
	AZMOVAZD - obj.ABaseARM64:     false,
	AZMOVAZH - obj.ABaseARM64:     false,
	AZMOVAZQ - obj.ABaseARM64:     false,
	AZMOVAZS - obj.ABaseARM64:     false,
	AZMOVB - obj.ABaseARM64:       false,
	AZMOVD - obj.ABaseARM64:       false,
	AZMOVH - obj.ABaseARM64:       false,
	AZMOVPRFX - obj.ABaseARM64:    false,
	AZMOVQ - obj.ABaseARM64:       false,
	AZMOVS - obj.ABaseARM64:       false,
	AZMSB - obj.ABaseARM64:        false,
	AZMUL - obj.ABaseARM64:        false,
	AZMULD - obj.ABaseARM64:       false,
	AZMULH - obj.ABaseARM64:       false,
	AZMULS - obj.ABaseARM64:       false,
	AZNBSL - obj.ABaseARM64:       false,
	AZNEG - obj.ABaseARM64:        false,
	AZNMATCH - obj.ABaseARM64:     false,
	AZNOT - obj.ABaseARM64:        false,
	AZORN - obj.ABaseARM64:        false,
	AZORQV - obj.ABaseARM64:       false,
	AZORR - obj.ABaseARM64:        false,
	AZORV - obj.ABaseARM64:        false,
	AZPMUL - obj.ABaseARM64:       false,
	AZPMULLBH - obj.ABaseARM64:    false,
	AZPMULLBQ - obj.ABaseARM64:    false,
	AZPMULLTH - obj.ABaseARM64:    false,
	AZPMULLTQ - obj.ABaseARM64:    false,
	AZPRFBD - obj.ABaseARM64:      false,
	AZPRFBS - obj.ABaseARM64:      false,
	AZPRFDD - obj.ABaseARM64:      false,
	AZPRFDS - obj.ABaseARM64:      false,
	AZPRFHD - obj.ABaseARM64:      false,
	AZPRFHS - obj.ABaseARM64:      false,
	AZPRFWD - obj.ABaseARM64:      false,
	AZPRFWS - obj.ABaseARM64:      false,
	AZRADDHNB - obj.ABaseARM64:    false,
	AZRADDHNT - obj.ABaseARM64:    false,
	AZRAX1 - obj.ABaseARM64:       false,
	AZRBIT - obj.ABaseARM64:       false,
	AZREV - obj.ABaseARM64:        false,
	AZREVB - obj.ABaseARM64:       false,
	AZREVD - obj.ABaseARM64:       false,
	AZREVH - obj.ABaseARM64:       false,
	AZREVW - obj.ABaseARM64:       false,
	AZRSHRNB - obj.ABaseARM64:     false,
	AZRSHRNT - obj.ABaseARM64:     false,
	AZRSUBHNB - obj.ABaseARM64:    false,
	AZRSUBHNT - obj.ABaseARM64:    false,
	AZSABA - obj.ABaseARM64:       false,
	AZSABALB - obj.ABaseARM64:     false,
	AZSABALT - obj.ABaseARM64:     false,
	AZSABD - obj.ABaseARM64:       false,
	AZSABDLB - obj.ABaseARM64:     false,
	AZSABDLT - obj.ABaseARM64:     false,
	AZSADALP - obj.ABaseARM64:     false,
	AZSADDLB - obj.ABaseARM64:     false,
	AZSADDLBT - obj.ABaseARM64:    false,
	AZSADDLT - obj.ABaseARM64:     false,
	AZSADDV - obj.ABaseARM64:      false,
	AZSADDWB - obj.ABaseARM64:     false,
	AZSADDWT - obj.ABaseARM64:     false,
	AZSBCLB - obj.ABaseARM64:      false,
	AZSBCLT - obj.ABaseARM64:      false,
	AZSCLAMP - obj.ABaseARM64:     false,
	AZSCVTF - obj.ABaseARM64:      false,
	AZSCVTFD - obj.ABaseARM64:     false,
	AZSCVTFH - obj.ABaseARM64:     false,
	AZSCVTFHH - obj.ABaseARM64:    false,
	AZSCVTFS - obj.ABaseARM64:     false,
	AZSCVTFWD - obj.ABaseARM64:    false,
	AZSCVTFWH - obj.ABaseARM64:    false,
	AZSCVTFWS - obj.ABaseARM64:    false,
	AZSDIV - obj.ABaseARM64:       false,
	AZSDIVR - obj.ABaseARM64:      false,
	AZSDOT - obj.ABaseARM64:       false,
	AZSDOTD - obj.ABaseARM64:      false,
	AZSDOTS - obj.ABaseARM64:      false,
	AZSEL - obj.ABaseARM64:        false,
	AZSHADD - obj.ABaseARM64:      false,
	AZSHRNB - obj.ABaseARM64:      false,
	AZSHRNT - obj.ABaseARM64:      false,
	AZSHSUB - obj.ABaseARM64:      false,
	AZSHSUBR - obj.ABaseARM64:     false,
	AZSLI - obj.ABaseARM64:        false,
	AZSMAX - obj.ABaseARM64:       false,
	AZSMAXP - obj.ABaseARM64:      false,
	AZSMAXQV - obj.ABaseARM64:     false,
	AZSMAXV - obj.ABaseARM64:      false,
	AZSMIN - obj.ABaseARM64:       false,
	AZSMINP - obj.ABaseARM64:      false,
	AZSMINQV - obj.ABaseARM64:     false,
	AZSMINV - obj.ABaseARM64:      false,
	AZSMLAL - obj.ABaseARM64:      false,
	AZSMLALB - obj.ABaseARM64:     false,
	AZSMLALBD - obj.ABaseARM64:    false,
	AZSMLALBS - obj.ABaseARM64:    false,
	AZSMLALL - obj.ABaseARM64:     false,
	AZSMLALLD - obj.ABaseARM64:    false,
	AZSMLALLS - obj.ABaseARM64:    false,
	AZSMLALT - obj.ABaseARM64:     false,
	AZSMLALTD - obj.ABaseARM64:    false,
	AZSMLALTS - obj.ABaseARM64:    false,
	AZSMLSL - obj.ABaseARM64:      false,
	AZSMLSLB - obj.ABaseARM64:     false,
	AZSMLSLBD - obj.ABaseARM64:    false,
	AZSMLSLBS - obj.ABaseARM64:    false,
	AZSMLSLL - obj.ABaseARM64:     false,
	AZSMLSLLD - obj.ABaseARM64:    false,
	AZSMLSLLS - obj.ABaseARM64:    false,
	AZSMLSLT - obj.ABaseARM64:     false,
	AZSMLSLTD - obj.ABaseARM64:    false,
	AZSMLSLTS - obj.ABaseARM64:    false,
	AZSMMLA - obj.ABaseARM64:      false,
	AZSMOPA - obj.ABaseARM64:      false,
	AZSMOPAD - obj.ABaseARM64:     false,
	AZSMOPAS - obj.ABaseARM64:     false,
	AZSMOPS - obj.ABaseARM64:      false,
	AZSMOPSD - obj.ABaseARM64:     false,
	AZSMOPSS - obj.ABaseARM64:     false,
	AZSMULH - obj.ABaseARM64:      false,
	AZSMULLB - obj.ABaseARM64:     false,
	AZSMULLBD - obj.ABaseARM64:    false,
	AZSMULLBS - obj.ABaseARM64:    false,
	AZSMULLT - obj.ABaseARM64:     false,
	AZSMULLTD - obj.ABaseARM64:    false,
	AZSMULLTS - obj.ABaseARM64:    false,
	AZSPLICE - obj.ABaseARM64:     false,
	AZSQABS - obj.ABaseARM64:      false,
	AZSQADD - obj.ABaseARM64:      false,
	AZSQCADD - obj.ABaseARM64:     false,
	AZSQCVT - obj.ABaseARM64:      false,
	AZSQCVTN - obj.ABaseARM64:     false,
	AZSQCVTU - obj.ABaseARM64:     false,
	AZSQCVTUN - obj.ABaseARM64:    false,
	AZSQDECD - obj.ABaseARM64:     false,
	AZSQDECH - obj.ABaseARM64:     false,
	AZSQDECP - obj.ABaseARM64:     false,
	AZSQDECW - obj.ABaseARM64:     false,
	AZSQDMLALB - obj.ABaseARM64:   false,
	AZSQDMLALBD - obj.ABaseARM64:  false,
	AZSQDMLALBS - obj.ABaseARM64:  false,
	AZSQDMLALBT - obj.ABaseARM64:  false,
	AZSQDMLALT - obj.ABaseARM64:   false,
	AZSQDMLALTD - obj.ABaseARM64:  false,
	AZSQDMLALTS - obj.ABaseARM64:  false,
	AZSQDMLSLB - obj.ABaseARM64:   false,
	AZSQDMLSLBD - obj.ABaseARM64:  false,
	AZSQDMLSLBS - obj.ABaseARM64:  false,
	AZSQDMLSLBT - obj.ABaseARM64:  false,
	AZSQDMLSLT - obj.ABaseARM64:   false,
	AZSQDMLSLTD - obj.ABaseARM64:  false,
	AZSQDMLSLTS - obj.ABaseARM64:  false,
	AZSQDMULH - obj.ABaseARM64:    false,
	AZSQDMULHD - obj.ABaseARM64:   false,
	AZSQDMULHH - obj.ABaseARM64:   false,
	AZSQDMULHS - obj.ABaseARM64:   false,
	AZSQDMULLB - obj.ABaseARM64:   false,
	AZSQDMULLBD - obj.ABaseARM64:  false,
	AZSQDMULLBS - obj.ABaseARM64:  false,
	AZSQDMULLT - obj.ABaseARM64:   false,
	AZSQDMULLTD - obj.ABaseARM64:  false,
	AZSQDMULLTS - obj.ABaseARM64:  false,
	AZSQINCD - obj.ABaseARM64:     false,
	AZSQINCH - obj.ABaseARM64:     false,
	AZSQINCP - obj.ABaseARM64:     false,
	AZSQINCW - obj.ABaseARM64:     false,
	AZSQNEG - obj.ABaseARM64:      false,
	AZSQRDCMLAH - obj.ABaseARM64:  false,
	AZSQRDCMLAHH - obj.ABaseARM64: false,
	AZSQRDCMLAHS - obj.ABaseARM64: false,
	AZSQRDMLAH - obj.ABaseARM64:   false,
	AZSQRDMLAHD - obj.ABaseARM64:  false,
	AZSQRDMLAHH - obj.ABaseARM64:  false,
	AZSQRDMLAHS - obj.ABaseARM64:  false,
	AZSQRDMLSH - obj.ABaseARM64:   false,
	AZSQRDMLSHD - obj.ABaseARM64:  false,
	AZSQRDMLSHH - obj.ABaseARM64:  false,
	AZSQRDMLSHS - obj.ABaseARM64:  false,
	AZSQRDMULH - obj.ABaseARM64:   false,
	AZSQRDMULHD - obj.ABaseARM64:  false,
	AZSQRDMULHH - obj.ABaseARM64:  false,
	AZSQRDMULHS - obj.ABaseARM64:  false,
	AZSQRSHL - obj.ABaseARM64:     false,
	AZSQRSHLR - obj.ABaseARM64:    false,
	AZSQRSHR - obj.ABaseARM64:     false,
	AZSQRSHRN - obj.ABaseARM64:    false,
	AZSQRSHRNB - obj.ABaseARM64:   false,
	AZSQRSHRNT - obj.ABaseARM64:   false,
	AZSQRSHRU - obj.ABaseARM64:    false,
	AZSQRSHRUN - obj.ABaseARM64:   false,
	AZSQRSHRUNB - obj.ABaseARM64:  false,
	AZSQRSHRUNT - obj.ABaseARM64:  false,
	AZSQSHL - obj.ABaseARM64:      false,
	AZSQSHLR - obj.ABaseARM64:     false,
	AZSQSHLU - obj.ABaseARM64:     false,
	AZSQSHRNB - obj.ABaseARM64:    false,
	AZSQSHRNT - obj.ABaseARM64:    false,
	AZSQSHRUNB - obj.ABaseARM64:   false,
	AZSQSHRUNT - obj.ABaseARM64:   false,
	AZSQSUB - obj.ABaseARM64:      false,
	AZSQSUBR - obj.ABaseARM64:     false,
	AZSQXTNB - obj.ABaseARM64:     false,
	AZSQXTNT - obj.ABaseARM64:     false,
	AZSQXTUNB - obj.ABaseARM64:    false,
	AZSQXTUNT - obj.ABaseARM64:    false,
	AZSRHADD - obj.ABaseARM64:     false,
	AZSRI - obj.ABaseARM64:        false,
	AZSRSHL - obj.ABaseARM64:      false,
	AZSRSHLR - obj.ABaseARM64:     false,
	AZSRSHR - obj.ABaseARM64:      false,
	AZSRSRA - obj.ABaseARM64:      false,
	AZSSHLLB - obj.ABaseARM64:     false,
	AZSSHLLT - obj.ABaseARM64:     false,
	AZSSRA - obj.ABaseARM64:       false,
	AZSSUBLB - obj.ABaseARM64:     false,
	AZSSUBLBT - obj.ABaseARM64:    false,
	AZSSUBLT - obj.ABaseARM64:     false,
	AZSSUBLTB - obj.ABaseARM64:    false,
	AZSSUBWB - obj.ABaseARM64:     false,
	AZSSUBWT - obj.ABaseARM64:     false,
	AZST1B - obj.ABaseARM64:       false,
	AZST1BD - obj.ABaseARM64:      false,
	AZST1BS - obj.ABaseARM64:      false,
	AZST1D - obj.ABaseARM64:       false,
	AZST1DD - obj.ABaseARM64:      false,
	AZST1DS - obj.ABaseARM64:      false,
	AZST1H - obj.ABaseARM64:       false,
	AZST1HD - obj.ABaseARM64:      false,
	AZST1HS - obj.ABaseARM64:      false,
	AZST1Q - obj.ABaseARM64:       false,
	AZST1W - obj.ABaseARM64:       false,
	AZST1WD - obj.ABaseARM64:      false,
	AZST1WS - obj.ABaseARM64:      false,
	AZST2B - obj.ABaseARM64:       false,
	AZST2D - obj.ABaseARM64:       false,
	AZST2H - obj.ABaseARM64:       false,
	AZST2Q - obj.ABaseARM64:       false,
	AZST2W - obj.ABaseARM64:       false,
	AZST3B - obj.ABaseARM64:       false,
	AZST3D - obj.ABaseARM64:       false,
	AZST3H - obj.ABaseARM64:       false,
	AZST3Q - obj.ABaseARM64:       false,
	AZST3W - obj.ABaseARM64:       false,
	AZST4B - obj.ABaseARM64:       false,
	AZST4D - obj.ABaseARM64:       false,
	AZST4H - obj.ABaseARM64:       false,
	AZST4Q - obj.ABaseARM64:       false,
	AZST4W - obj.ABaseARM64:       false,
	AZSTNT1B - obj.ABaseARM64:     false,
	AZSTNT1BD - obj.ABaseARM64:    false,
	AZSTNT1BS - obj.ABaseARM64:    false,
	AZSTNT1D - obj.ABaseARM64:     false,
	AZSTNT1H - obj.ABaseARM64:     false,
	AZSTNT1HD - obj.ABaseARM64:    false,
	AZSTNT1HS - obj.ABaseARM64:    false,
	AZSTNT1W - obj.ABaseARM64:     false,
	AZSTNT1WD - obj.ABaseARM64:    false,
	AZSTNT1WS - obj.ABaseARM64:    false,
	AZSTR - obj.ABaseARM64:        false,
	AZSUB - obj.ABaseARM64:        false,
	AZSUBHNB - obj.ABaseARM64:     false,
	AZSUBHNT - obj.ABaseARM64:     false,
	AZSUBR - obj.ABaseARM64:       false,
	AZSUDOT - obj.ABaseARM64:      false,
	AZSUMLALL - obj.ABaseARM64:    false,
	AZSUMOPAD - obj.ABaseARM64:    false,
	AZSUMOPAS - obj.ABaseARM64:    false,
	AZSUMOPSD - obj.ABaseARM64:    false,
	AZSUMOPSS - obj.ABaseARM64:    false,
	AZSUNPK - obj.ABaseARM64:      false,
	AZSUNPKHI - obj.ABaseARM64:    false,
	AZSUNPKLO - obj.ABaseARM64:    false,
	AZSUQADD - obj.ABaseARM64:     false,
	AZSUVDOT - obj.ABaseARM64:     false,
	AZSVDOT - obj.ABaseARM64:      false,
	AZSVDOTD - obj.ABaseARM64:     false,
	AZSVDOTS - obj.ABaseARM64:     false,
	AZSXTB - obj.ABaseARM64:       false,
	AZSXTH - obj.ABaseARM64:       false,
	AZSXTW - obj.ABaseARM64:       false,
	AZTBL - obj.ABaseARM64:        false,
	AZTBLQ - obj.ABaseARM64:       false,
	AZTBX - obj.ABaseARM64:        false,
	AZTBXQ - obj.ABaseARM64:       false,
	AZTRN1 - obj.ABaseARM64:       false,
	AZTRN2 - obj.ABaseARM64:       false,
	AZUABA - obj.ABaseARM64:       false,
	AZUABALB - obj.ABaseARM64:     false,
	AZUABALT - obj.ABaseARM64:     false,
	AZUABD - obj.ABaseARM64:       false,
	AZUABDLB - obj.ABaseARM64:     false,
	AZUABDLT - obj.ABaseARM64:     false,
	AZUADALP - obj.ABaseARM64:     false,
	AZUADDLB - obj.ABaseARM64:     false,
	AZUADDLT - obj.ABaseARM64:     false,
	AZUADDV - obj.ABaseARM64:      false,
	AZUADDWB - obj.ABaseARM64:     false,
	AZUADDWT - obj.ABaseARM64:     false,
	AZUCLAMP - obj.ABaseARM64:     false,
	AZUCVTF - obj.ABaseARM64:      false,
	AZUCVTFD - obj.ABaseARM64:     false,
	AZUCVTFH - obj.ABaseARM64:     false,
	AZUCVTFHH - obj.ABaseARM64:    false,
	AZUCVTFS - obj.ABaseARM64:     false,
	AZUCVTFWD - obj.ABaseARM64:    false,
	AZUCVTFWH - obj.ABaseARM64:    false,
	AZUCVTFWS - obj.ABaseARM64:    false,
	AZUDIV - obj.ABaseARM64:       false,
	AZUDIVR - obj.ABaseARM64:      false,
	AZUDOT - obj.ABaseARM64:       false,
	AZUDOTD - obj.ABaseARM64:      false,
	AZUDOTS - obj.ABaseARM64:      false,
	AZUHADD - obj.ABaseARM64:      false,
	AZUHSUB - obj.ABaseARM64:      false,
	AZUHSUBR - obj.ABaseARM64:     false,
	AZUMAX - obj.ABaseARM64:       false,
	AZUMAXP - obj.ABaseARM64:      false,
	AZUMAXQV - obj.ABaseARM64:     false,
	AZUMAXV - obj.ABaseARM64:      false,
	AZUMIN - obj.ABaseARM64:       false,
	AZUMINP - obj.ABaseARM64:      false,
	AZUMINQV - obj.ABaseARM64:     false,
	AZUMINV - obj.ABaseARM64:      false,
	AZUMLAL - obj.ABaseARM64:      false,
	AZUMLALB - obj.ABaseARM64:     false,
	AZUMLALBD - obj.ABaseARM64:    false,
	AZUMLALBS - obj.ABaseARM64:    false,
	AZUMLALL - obj.ABaseARM64:     false,
	AZUMLALLD - obj.ABaseARM64:    false,
	AZUMLALLS - obj.ABaseARM64:    false,
	AZUMLALT - obj.ABaseARM64:     false,
	AZUMLALTD - obj.ABaseARM64:    false,
	AZUMLALTS - obj.ABaseARM64:    false,
	AZUMLSL - obj.ABaseARM64:      false,
	AZUMLSLB - obj.ABaseARM64:     false,
	AZUMLSLBD - obj.ABaseARM64:    false,
	AZUMLSLBS - obj.ABaseARM64:    false,
	AZUMLSLL - obj.ABaseARM64:     false,
	AZUMLSLLD - obj.ABaseARM64:    false,
	AZUMLSLLS - obj.ABaseARM64:    false,
	AZUMLSLT - obj.ABaseARM64:     false,
	AZUMLSLTD - obj.ABaseARM64:    false,
	AZUMLSLTS - obj.ABaseARM64:    false,
	AZUMMLA - obj.ABaseARM64:      false,
	AZUMOPA - obj.ABaseARM64:      false,
	AZUMOPAD - obj.ABaseARM64:     false,
	AZUMOPAS - obj.ABaseARM64:     false,
	AZUMOPS - obj.ABaseARM64:      false,
	AZUMOPSD - obj.ABaseARM64:     false,
	AZUMOPSS - obj.ABaseARM64:     false,
	AZUMULH - obj.ABaseARM64:      false,
	AZUMULLB - obj.ABaseARM64:     false,
	AZUMULLBD - obj.ABaseARM64:    false,
	AZUMULLBS - obj.ABaseARM64:    false,
	AZUMULLT - obj.ABaseARM64:     false,
	AZUMULLTD - obj.ABaseARM64:    false,
	AZUMULLTS - obj.ABaseARM64:    false,
	AZUQADD - obj.ABaseARM64:      false,
	AZUQCVT - obj.ABaseARM64:      false,
	AZUQCVTN - obj.ABaseARM64:     false,
	AZUQDECD - obj.ABaseARM64:     false,
	AZUQDECH - obj.ABaseARM64:     false,
	AZUQDECP - obj.ABaseARM64:     false,
	AZUQDECW - obj.ABaseARM64:     false,
	AZUQINCD - obj.ABaseARM64:     false,
	AZUQINCH - obj.ABaseARM64:     false,
	AZUQINCP - obj.ABaseARM64:     false,
	AZUQINCW - obj.ABaseARM64:     false,
	AZUQRSHL - obj.ABaseARM64:     false,
	AZUQRSHLR - obj.ABaseARM64:    false,
	AZUQRSHR - obj.ABaseARM64:     false,
	AZUQRSHRN - obj.ABaseARM64:    false,
	AZUQRSHRNB - obj.ABaseARM64:   false,
	AZUQRSHRNT - obj.ABaseARM64:   false,
	AZUQSHL - obj.ABaseARM64:      false,
	AZUQSHLR - obj.ABaseARM64:     false,
	AZUQSHRNB - obj.ABaseARM64:    false,
	AZUQSHRNT - obj.ABaseARM64:    false,
	AZUQSUB - obj.ABaseARM64:      false,
	AZUQSUBR - obj.ABaseARM64:     false,
	AZUQXTNB - obj.ABaseARM64:     false,
	AZUQXTNT - obj.ABaseARM64:     false,
	AZURECPE - obj.ABaseARM64:     false,
	AZURHADD - obj.ABaseARM64:     false,
	AZURSHL - obj.ABaseARM64:      false,
	AZURSHLR - obj.ABaseARM64:     false,
	AZURSHR - obj.ABaseARM64:      false,
	AZURSQRTE - obj.ABaseARM64:    false,
	AZURSRA - obj.ABaseARM64:      false,
	AZUSDOT - obj.ABaseARM64:      false,
	AZUSHLLB - obj.ABaseARM64:     false,
	AZUSHLLT - obj.ABaseARM64:     false,
	AZUSMLALL - obj.ABaseARM64:    false,
	AZUSMMLA - obj.ABaseARM64:     false,
	AZUSMOPAD - obj.ABaseARM64:    false,
	AZUSMOPAS - obj.ABaseARM64:    false,
	AZUSMOPSD - obj.ABaseARM64:    false,
	AZUSMOPSS - obj.ABaseARM64:    false,
	AZUSQADD - obj.ABaseARM64:     false,
	AZUSRA - obj.ABaseARM64:       false,
	AZUSUBLB - obj.ABaseARM64:     false,
	AZUSUBLT - obj.ABaseARM64:     false,
	AZUSUBWB - obj.ABaseARM64:     false,
	AZUSUBWT - obj.ABaseARM64:     false,
	AZUSVDOT - obj.ABaseARM64:     false,
	AZUUNPK - obj.ABaseARM64:      false,
	AZUUNPKHI - obj.ABaseARM64:    false,
	AZUUNPKLO - obj.ABaseARM64:    false,
	AZUVDOT - obj.ABaseARM64:      false,
	AZUVDOTD - obj.ABaseARM64:     false,
	AZUVDOTS - obj.ABaseARM64:     false,
	AZUXTB - obj.ABaseARM64:       false,
	AZUXTH - obj.ABaseARM64:       false,
	AZUXTW - obj.ABaseARM64:       false,
	AZUZP1 - obj.ABaseARM64:       false,
	AZUZP2 - obj.ABaseARM64:       false,
	AZUZPB - obj.ABaseARM64:       false,
	AZUZPQ - obj.ABaseARM64:       false,
	AZUZPQ1 - obj.ABaseARM64:      false,
	AZUZPQ2 - obj.ABaseARM64:      false,
	AZXAR - obj.ABaseARM64:        false,
	AZZIP1 - obj.ABaseARM64:       false,
	AZZIP2 - obj.ABaseARM64:       false,
	AZZIPB - obj.ABaseARM64:       false,
	AZZIPQ - obj.ABaseARM64:       false,
	AZZIPQ1 - obj.ABaseARM64:      false,
	AZZIPQ2 - obj.ABaseARM64:      false,
}
