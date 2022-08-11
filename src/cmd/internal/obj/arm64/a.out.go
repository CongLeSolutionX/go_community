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
	RTYP_INDEX             // Vn.<T>[index], Rn[index]
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
	ARNG_4B
	ARNG_8B
	ARNG_16B
	ARNG_1D
	ARNG_2H
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
	C_NONE   = iota + 1 // starting from 1, leave unclassified Addr's class as 0
	C_REG               // R0..R30
	C_ZREG              // R0..R30, ZR
	C_RSP               // R0..R30, RSP
	C_FREG              // F0..F31
	C_VREG              // V0..V31
	C_PAIR              // (Rn, Rm)
	C_SHIFT             // Rn<<2
	C_EXTREG            // Rn.UXTB[<<3]
	C_SPR               // REG_NZCV
	C_COND              // condition code, EQ, NE, etc.
	C_SPOP              // special operand, PLDL1KEEP, VMALLE1IS, etc.
	C_ARNG              // Vn.<T>
	C_ELEM              // Vn.<T>[index]
	C_LIST              // [V1, V2, V3]

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
	C_LAUTOPOOL   // any other constant up to 64 bits (needs pool literal)
	C_LAUTO       // any other constant up to 64 bits

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
	C_LOREGPOOL
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
// instTab is sorted based on the order of these constants and the first match is chosen.
//
//go:generate stringer -type oprType -trimprefix AC_
const (
	AC_NONE     oprType = iota
	AC_REG              // general purpose registers R0..R30 and ZR
	AC_RSP              // general purpose registers R0..R30 and RSP
	AC_FREG             // floating point registers, such as F1
	AC_VREG             // vector registers, such as V1
	AC_ZREG             // the scalable vector registers, such as Z1
	AC_ZAREG            // the name of the ZA tile, defined as registers, such as ZA0
	AC_ZTREG            // the ZT0 register
	AC_PREG             // the scalable predicate registers, such as P1
	AC_PNREG            // the scalable predicate registers, with predicate-as-counter encoding, such as PN1
	AC_PREGM            // Pg/M
	AC_PREGZ            // Pg/Z
	AC_REGIDX           // P8[1]
	AC_PAIR             // register pair, such as (R1, R3)
	AC_REGSHIFT         // general purpose register with shift, such as R1<<2
	AC_REGEXT           // general purpose register with extend, such as R7.SXTW<<1
	AC_ARNG             // vector register with arrangement, such as V11.D2
	AC_ARNGIDX          // vector register with arrangement and index, such as V12.D[1]

	AC_COND  // conditional flags, such as CS
	AC_SPR   // special register, such as REG_NZCV, system registers
	AC_SPOP  // special operands, such as DAIFSet
	AC_LABEL // branch labels
	AC_IMM   // constants

	AC_REGLIST1 // list of 1 vector register, such as [V1]
	AC_REGLIST2 // list of 2 vector registers, such as [V1, V2], [Z0, Z8]
	AC_REGLIST3 // list of 3 vector registers, such as [V1, V2, V3]
	AC_REGLIST4 // list of 4 vector registers, such as [V1, V2, V3, V4], [Z0, Z4, Z8, Z12]
	AC_LISTIDX  // list with index, such as [V1.B, V2.B][2]

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
	AADC - obj.ABaseARM64:       true,
	AADCS - obj.ABaseARM64:      true,
	AADCSW - obj.ABaseARM64:     true,
	AADCW - obj.ABaseARM64:      true,
	AADD - obj.ABaseARM64:       true,
	AADDS - obj.ABaseARM64:      true,
	AADDSW - obj.ABaseARM64:     true,
	AADDW - obj.ABaseARM64:      true,
	AADR - obj.ABaseARM64:       true,
	AADRP - obj.ABaseARM64:      true,
	AAESD - obj.ABaseARM64:      true,
	AAESE - obj.ABaseARM64:      true,
	AAESIMC - obj.ABaseARM64:    true,
	AAESMC - obj.ABaseARM64:     true,
	AAND - obj.ABaseARM64:       true,
	AANDS - obj.ABaseARM64:      true,
	AANDSW - obj.ABaseARM64:     true,
	AANDW - obj.ABaseARM64:      true,
	AASR - obj.ABaseARM64:       true,
	AASRW - obj.ABaseARM64:      true,
	AAT - obj.ABaseARM64:        true,
	ABCC - obj.ABaseARM64:       true,
	ABCS - obj.ABaseARM64:       true,
	ABEQ - obj.ABaseARM64:       true,
	ABFI - obj.ABaseARM64:       true,
	ABFIW - obj.ABaseARM64:      true,
	ABFM - obj.ABaseARM64:       true,
	ABFMW - obj.ABaseARM64:      true,
	ABFXIL - obj.ABaseARM64:     true,
	ABFXILW - obj.ABaseARM64:    true,
	ABGE - obj.ABaseARM64:       true,
	ABGT - obj.ABaseARM64:       true,
	ABHI - obj.ABaseARM64:       true,
	ABHS - obj.ABaseARM64:       true,
	ABIC - obj.ABaseARM64:       true,
	ABICS - obj.ABaseARM64:      true,
	ABICSW - obj.ABaseARM64:     true,
	ABICW - obj.ABaseARM64:      true,
	ABLE - obj.ABaseARM64:       true,
	ABLO - obj.ABaseARM64:       true,
	ABLS - obj.ABaseARM64:       true,
	ABLT - obj.ABaseARM64:       true,
	ABMI - obj.ABaseARM64:       true,
	ABNE - obj.ABaseARM64:       true,
	ABPL - obj.ABaseARM64:       true,
	ABRK - obj.ABaseARM64:       true,
	ABVC - obj.ABaseARM64:       true,
	ABVS - obj.ABaseARM64:       true,
	ACASAD - obj.ABaseARM64:     true,
	ACASALB - obj.ABaseARM64:    true,
	ACASALD - obj.ABaseARM64:    true,
	ACASALH - obj.ABaseARM64:    true,
	ACASALW - obj.ABaseARM64:    true,
	ACASAW - obj.ABaseARM64:     true,
	ACASB - obj.ABaseARM64:      true,
	ACASD - obj.ABaseARM64:      true,
	ACASH - obj.ABaseARM64:      true,
	ACASLD - obj.ABaseARM64:     true,
	ACASLW - obj.ABaseARM64:     true,
	ACASPD - obj.ABaseARM64:     true,
	ACASPW - obj.ABaseARM64:     true,
	ACASW - obj.ABaseARM64:      true,
	ACBNZ - obj.ABaseARM64:      true,
	ACBNZW - obj.ABaseARM64:     true,
	ACBZ - obj.ABaseARM64:       true,
	ACBZW - obj.ABaseARM64:      true,
	ACCMN - obj.ABaseARM64:      true,
	ACCMNW - obj.ABaseARM64:     true,
	ACCMP - obj.ABaseARM64:      true,
	ACCMPW - obj.ABaseARM64:     true,
	ACINC - obj.ABaseARM64:      true,
	ACINCW - obj.ABaseARM64:     true,
	ACINV - obj.ABaseARM64:      true,
	ACINVW - obj.ABaseARM64:     true,
	ACLREX - obj.ABaseARM64:     true,
	ACLS - obj.ABaseARM64:       true,
	ACLSW - obj.ABaseARM64:      true,
	ACLZ - obj.ABaseARM64:       true,
	ACLZW - obj.ABaseARM64:      true,
	ACMN - obj.ABaseARM64:       true,
	ACMNW - obj.ABaseARM64:      true,
	ACMP - obj.ABaseARM64:       true,
	ACMPW - obj.ABaseARM64:      true,
	ACNEG - obj.ABaseARM64:      true,
	ACNEGW - obj.ABaseARM64:     true,
	ACRC32B - obj.ABaseARM64:    true,
	ACRC32CB - obj.ABaseARM64:   true,
	ACRC32CH - obj.ABaseARM64:   true,
	ACRC32CW - obj.ABaseARM64:   true,
	ACRC32CX - obj.ABaseARM64:   true,
	ACRC32H - obj.ABaseARM64:    true,
	ACRC32W - obj.ABaseARM64:    true,
	ACRC32X - obj.ABaseARM64:    true,
	ACSEL - obj.ABaseARM64:      true,
	ACSELW - obj.ABaseARM64:     true,
	ACSET - obj.ABaseARM64:      true,
	ACSETM - obj.ABaseARM64:     true,
	ACSETMW - obj.ABaseARM64:    true,
	ACSETW - obj.ABaseARM64:     true,
	ACSINC - obj.ABaseARM64:     true,
	ACSINCW - obj.ABaseARM64:    true,
	ACSINV - obj.ABaseARM64:     true,
	ACSINVW - obj.ABaseARM64:    true,
	ACSNEG - obj.ABaseARM64:     true,
	ACSNEGW - obj.ABaseARM64:    true,
	ADC - obj.ABaseARM64:        true,
	ADCPS1 - obj.ABaseARM64:     true,
	ADCPS2 - obj.ABaseARM64:     true,
	ADCPS3 - obj.ABaseARM64:     true,
	ADMB - obj.ABaseARM64:       true,
	ADRPS - obj.ABaseARM64:      true,
	ADSB - obj.ABaseARM64:       true,
	ADWORD - obj.ABaseARM64:     true,
	AEON - obj.ABaseARM64:       true,
	AEONW - obj.ABaseARM64:      true,
	AEOR - obj.ABaseARM64:       true,
	AEORW - obj.ABaseARM64:      true,
	AERET - obj.ABaseARM64:      true,
	AEXTR - obj.ABaseARM64:      true,
	AEXTRW - obj.ABaseARM64:     true,
	AFABSD - obj.ABaseARM64:     true,
	AFABSS - obj.ABaseARM64:     true,
	AFADDD - obj.ABaseARM64:     true,
	AFADDS - obj.ABaseARM64:     true,
	AFCCMPD - obj.ABaseARM64:    true,
	AFCCMPED - obj.ABaseARM64:   true,
	AFCCMPES - obj.ABaseARM64:   true,
	AFCCMPS - obj.ABaseARM64:    true,
	AFCMPD - obj.ABaseARM64:     true,
	AFCMPED - obj.ABaseARM64:    true,
	AFCMPES - obj.ABaseARM64:    true,
	AFCMPS - obj.ABaseARM64:     true,
	AFCSELD - obj.ABaseARM64:    true,
	AFCSELS - obj.ABaseARM64:    true,
	AFCVTDH - obj.ABaseARM64:    true,
	AFCVTDS - obj.ABaseARM64:    true,
	AFCVTHD - obj.ABaseARM64:    true,
	AFCVTHS - obj.ABaseARM64:    true,
	AFCVTSD - obj.ABaseARM64:    true,
	AFCVTSH - obj.ABaseARM64:    true,
	AFCVTZSD - obj.ABaseARM64:   true,
	AFCVTZSDW - obj.ABaseARM64:  true,
	AFCVTZSS - obj.ABaseARM64:   true,
	AFCVTZSSW - obj.ABaseARM64:  true,
	AFCVTZUD - obj.ABaseARM64:   true,
	AFCVTZUDW - obj.ABaseARM64:  true,
	AFCVTZUS - obj.ABaseARM64:   true,
	AFCVTZUSW - obj.ABaseARM64:  true,
	AFDIVD - obj.ABaseARM64:     true,
	AFDIVS - obj.ABaseARM64:     true,
	AFLDPD - obj.ABaseARM64:     true,
	AFLDPQ - obj.ABaseARM64:     true,
	AFLDPS - obj.ABaseARM64:     true,
	AFMADDD - obj.ABaseARM64:    true,
	AFMADDS - obj.ABaseARM64:    true,
	AFMAXD - obj.ABaseARM64:     true,
	AFMAXNMD - obj.ABaseARM64:   true,
	AFMAXNMS - obj.ABaseARM64:   true,
	AFMAXS - obj.ABaseARM64:     true,
	AFMIND - obj.ABaseARM64:     true,
	AFMINNMD - obj.ABaseARM64:   true,
	AFMINNMS - obj.ABaseARM64:   true,
	AFMINS - obj.ABaseARM64:     true,
	AFMOVD - obj.ABaseARM64:     true,
	AFMOVQ - obj.ABaseARM64:     true,
	AFMOVS - obj.ABaseARM64:     true,
	AFMSUBD - obj.ABaseARM64:    true,
	AFMSUBS - obj.ABaseARM64:    true,
	AFMULD - obj.ABaseARM64:     true,
	AFMULS - obj.ABaseARM64:     true,
	AFNEGD - obj.ABaseARM64:     true,
	AFNEGS - obj.ABaseARM64:     true,
	AFNMADDD - obj.ABaseARM64:   true,
	AFNMADDS - obj.ABaseARM64:   true,
	AFNMSUBD - obj.ABaseARM64:   true,
	AFNMSUBS - obj.ABaseARM64:   true,
	AFNMULD - obj.ABaseARM64:    true,
	AFNMULS - obj.ABaseARM64:    true,
	AFRINTAD - obj.ABaseARM64:   true,
	AFRINTAS - obj.ABaseARM64:   true,
	AFRINTID - obj.ABaseARM64:   true,
	AFRINTIS - obj.ABaseARM64:   true,
	AFRINTMD - obj.ABaseARM64:   true,
	AFRINTMS - obj.ABaseARM64:   true,
	AFRINTND - obj.ABaseARM64:   true,
	AFRINTNS - obj.ABaseARM64:   true,
	AFRINTPD - obj.ABaseARM64:   true,
	AFRINTPS - obj.ABaseARM64:   true,
	AFRINTXD - obj.ABaseARM64:   true,
	AFRINTXS - obj.ABaseARM64:   true,
	AFRINTZD - obj.ABaseARM64:   true,
	AFRINTZS - obj.ABaseARM64:   true,
	AFSQRTD - obj.ABaseARM64:    true,
	AFSQRTS - obj.ABaseARM64:    true,
	AFSTPD - obj.ABaseARM64:     true,
	AFSTPQ - obj.ABaseARM64:     true,
	AFSTPS - obj.ABaseARM64:     true,
	AFSUBD - obj.ABaseARM64:     true,
	AFSUBS - obj.ABaseARM64:     true,
	AHINT - obj.ABaseARM64:      true,
	AHLT - obj.ABaseARM64:       true,
	AHVC - obj.ABaseARM64:       true,
	AIC - obj.ABaseARM64:        true,
	AISB - obj.ABaseARM64:       true,
	ALDADDAB - obj.ABaseARM64:   true,
	ALDADDAD - obj.ABaseARM64:   true,
	ALDADDAH - obj.ABaseARM64:   true,
	ALDADDALB - obj.ABaseARM64:  true,
	ALDADDALD - obj.ABaseARM64:  true,
	ALDADDALH - obj.ABaseARM64:  true,
	ALDADDALW - obj.ABaseARM64:  true,
	ALDADDAW - obj.ABaseARM64:   true,
	ALDADDB - obj.ABaseARM64:    true,
	ALDADDD - obj.ABaseARM64:    true,
	ALDADDH - obj.ABaseARM64:    true,
	ALDADDLB - obj.ABaseARM64:   true,
	ALDADDLD - obj.ABaseARM64:   true,
	ALDADDLH - obj.ABaseARM64:   true,
	ALDADDLW - obj.ABaseARM64:   true,
	ALDADDW - obj.ABaseARM64:    true,
	ALDAR - obj.ABaseARM64:      true,
	ALDARB - obj.ABaseARM64:     true,
	ALDARH - obj.ABaseARM64:     true,
	ALDARW - obj.ABaseARM64:     true,
	ALDAXP - obj.ABaseARM64:     true,
	ALDAXPW - obj.ABaseARM64:    true,
	ALDAXR - obj.ABaseARM64:     true,
	ALDAXRB - obj.ABaseARM64:    true,
	ALDAXRH - obj.ABaseARM64:    true,
	ALDAXRW - obj.ABaseARM64:    true,
	ALDCLRAB - obj.ABaseARM64:   true,
	ALDCLRAD - obj.ABaseARM64:   true,
	ALDCLRAH - obj.ABaseARM64:   true,
	ALDCLRALB - obj.ABaseARM64:  true,
	ALDCLRALD - obj.ABaseARM64:  true,
	ALDCLRALH - obj.ABaseARM64:  true,
	ALDCLRALW - obj.ABaseARM64:  true,
	ALDCLRAW - obj.ABaseARM64:   true,
	ALDCLRB - obj.ABaseARM64:    true,
	ALDCLRD - obj.ABaseARM64:    true,
	ALDCLRH - obj.ABaseARM64:    true,
	ALDCLRLB - obj.ABaseARM64:   true,
	ALDCLRLD - obj.ABaseARM64:   true,
	ALDCLRLH - obj.ABaseARM64:   true,
	ALDCLRLW - obj.ABaseARM64:   true,
	ALDCLRW - obj.ABaseARM64:    true,
	ALDEORAB - obj.ABaseARM64:   true,
	ALDEORAD - obj.ABaseARM64:   true,
	ALDEORAH - obj.ABaseARM64:   true,
	ALDEORALB - obj.ABaseARM64:  true,
	ALDEORALD - obj.ABaseARM64:  true,
	ALDEORALH - obj.ABaseARM64:  true,
	ALDEORALW - obj.ABaseARM64:  true,
	ALDEORAW - obj.ABaseARM64:   true,
	ALDEORB - obj.ABaseARM64:    true,
	ALDEORD - obj.ABaseARM64:    true,
	ALDEORH - obj.ABaseARM64:    true,
	ALDEORLB - obj.ABaseARM64:   true,
	ALDEORLD - obj.ABaseARM64:   true,
	ALDEORLH - obj.ABaseARM64:   true,
	ALDEORLW - obj.ABaseARM64:   true,
	ALDEORW - obj.ABaseARM64:    true,
	ALDORAB - obj.ABaseARM64:    true,
	ALDORAD - obj.ABaseARM64:    true,
	ALDORAH - obj.ABaseARM64:    true,
	ALDORALB - obj.ABaseARM64:   true,
	ALDORALD - obj.ABaseARM64:   true,
	ALDORALH - obj.ABaseARM64:   true,
	ALDORALW - obj.ABaseARM64:   true,
	ALDORAW - obj.ABaseARM64:    true,
	ALDORB - obj.ABaseARM64:     true,
	ALDORD - obj.ABaseARM64:     true,
	ALDORH - obj.ABaseARM64:     true,
	ALDORLB - obj.ABaseARM64:    true,
	ALDORLD - obj.ABaseARM64:    true,
	ALDORLH - obj.ABaseARM64:    true,
	ALDORLW - obj.ABaseARM64:    true,
	ALDORW - obj.ABaseARM64:     true,
	ALDP - obj.ABaseARM64:       true,
	ALDPSW - obj.ABaseARM64:     true,
	ALDPW - obj.ABaseARM64:      true,
	ALDXP - obj.ABaseARM64:      true,
	ALDXPW - obj.ABaseARM64:     true,
	ALDXR - obj.ABaseARM64:      true,
	ALDXRB - obj.ABaseARM64:     true,
	ALDXRH - obj.ABaseARM64:     true,
	ALDXRW - obj.ABaseARM64:     true,
	ALSL - obj.ABaseARM64:       true,
	ALSLW - obj.ABaseARM64:      true,
	ALSR - obj.ABaseARM64:       true,
	ALSRW - obj.ABaseARM64:      true,
	AMADD - obj.ABaseARM64:      true,
	AMADDW - obj.ABaseARM64:     true,
	AMNEG - obj.ABaseARM64:      true,
	AMNEGW - obj.ABaseARM64:     true,
	AMOVB - obj.ABaseARM64:      true,
	AMOVBU - obj.ABaseARM64:     true,
	AMOVD - obj.ABaseARM64:      true,
	AMOVH - obj.ABaseARM64:      true,
	AMOVHU - obj.ABaseARM64:     true,
	AMOVK - obj.ABaseARM64:      true,
	AMOVKW - obj.ABaseARM64:     true,
	AMOVN - obj.ABaseARM64:      true,
	AMOVNW - obj.ABaseARM64:     true,
	AMOVW - obj.ABaseARM64:      true,
	AMOVWU - obj.ABaseARM64:     true,
	AMOVZ - obj.ABaseARM64:      true,
	AMOVZW - obj.ABaseARM64:     true,
	AMRS - obj.ABaseARM64:       true,
	AMSR - obj.ABaseARM64:       true,
	AMSUB - obj.ABaseARM64:      true,
	AMSUBW - obj.ABaseARM64:     true,
	AMUL - obj.ABaseARM64:       true,
	AMULW - obj.ABaseARM64:      true,
	AMVN - obj.ABaseARM64:       true,
	AMVNW - obj.ABaseARM64:      true,
	ANEG - obj.ABaseARM64:       true,
	ANEGS - obj.ABaseARM64:      true,
	ANEGSW - obj.ABaseARM64:     true,
	ANEGW - obj.ABaseARM64:      true,
	ANGC - obj.ABaseARM64:       true,
	ANGCS - obj.ABaseARM64:      true,
	ANGCSW - obj.ABaseARM64:     true,
	ANGCW - obj.ABaseARM64:      true,
	ANOOP - obj.ABaseARM64:      true,
	AORN - obj.ABaseARM64:       true,
	AORNW - obj.ABaseARM64:      true,
	AORR - obj.ABaseARM64:       true,
	AORRW - obj.ABaseARM64:      true,
	APRFM - obj.ABaseARM64:      true,
	ARBIT - obj.ABaseARM64:      true,
	ARBITW - obj.ABaseARM64:     true,
	AREM - obj.ABaseARM64:       true,
	AREMW - obj.ABaseARM64:      true,
	AREV - obj.ABaseARM64:       true,
	AREV16 - obj.ABaseARM64:     true,
	AREV16W - obj.ABaseARM64:    true,
	AREV32 - obj.ABaseARM64:     true,
	AREVW - obj.ABaseARM64:      true,
	AROR - obj.ABaseARM64:       true,
	ARORW - obj.ABaseARM64:      true,
	ASBC - obj.ABaseARM64:       true,
	ASBCS - obj.ABaseARM64:      true,
	ASBCSW - obj.ABaseARM64:     true,
	ASBCW - obj.ABaseARM64:      true,
	ASBFIZ - obj.ABaseARM64:     true,
	ASBFIZW - obj.ABaseARM64:    true,
	ASBFM - obj.ABaseARM64:      true,
	ASBFMW - obj.ABaseARM64:     true,
	ASBFX - obj.ABaseARM64:      true,
	ASBFXW - obj.ABaseARM64:     true,
	ASCVTFD - obj.ABaseARM64:    true,
	ASCVTFS - obj.ABaseARM64:    true,
	ASCVTFWD - obj.ABaseARM64:   true,
	ASCVTFWS - obj.ABaseARM64:   true,
	ASDIV - obj.ABaseARM64:      true,
	ASDIVW - obj.ABaseARM64:     true,
	ASEV - obj.ABaseARM64:       true,
	ASEVL - obj.ABaseARM64:      true,
	ASHA1C - obj.ABaseARM64:     true,
	ASHA1H - obj.ABaseARM64:     true,
	ASHA1M - obj.ABaseARM64:     true,
	ASHA1P - obj.ABaseARM64:     true,
	ASHA1SU0 - obj.ABaseARM64:   true,
	ASHA1SU1 - obj.ABaseARM64:   true,
	ASHA256H - obj.ABaseARM64:   true,
	ASHA256H2 - obj.ABaseARM64:  true,
	ASHA256SU0 - obj.ABaseARM64: true,
	ASHA256SU1 - obj.ABaseARM64: true,
	ASHA512H - obj.ABaseARM64:   true,
	ASHA512H2 - obj.ABaseARM64:  true,
	ASHA512SU0 - obj.ABaseARM64: true,
	ASHA512SU1 - obj.ABaseARM64: true,
	ASMADDL - obj.ABaseARM64:    true,
	ASMC - obj.ABaseARM64:       true,
	ASMNEGL - obj.ABaseARM64:    true,
	ASMSUBL - obj.ABaseARM64:    true,
	ASMULH - obj.ABaseARM64:     true,
	ASMULL - obj.ABaseARM64:     true,
	ASTLR - obj.ABaseARM64:      true,
	ASTLRB - obj.ABaseARM64:     true,
	ASTLRH - obj.ABaseARM64:     true,
	ASTLRW - obj.ABaseARM64:     true,
	ASTLXP - obj.ABaseARM64:     true,
	ASTLXPW - obj.ABaseARM64:    true,
	ASTLXR - obj.ABaseARM64:     true,
	ASTLXRB - obj.ABaseARM64:    true,
	ASTLXRH - obj.ABaseARM64:    true,
	ASTLXRW - obj.ABaseARM64:    true,
	ASTP - obj.ABaseARM64:       true,
	ASTPW - obj.ABaseARM64:      true,
	ASTXP - obj.ABaseARM64:      true,
	ASTXPW - obj.ABaseARM64:     true,
	ASTXR - obj.ABaseARM64:      true,
	ASTXRB - obj.ABaseARM64:     true,
	ASTXRH - obj.ABaseARM64:     true,
	ASTXRW - obj.ABaseARM64:     true,
	ASUB - obj.ABaseARM64:       true,
	ASUBS - obj.ABaseARM64:      true,
	ASUBSW - obj.ABaseARM64:     true,
	ASUBW - obj.ABaseARM64:      true,
	ASVC - obj.ABaseARM64:       true,
	ASWPAB - obj.ABaseARM64:     true,
	ASWPAD - obj.ABaseARM64:     true,
	ASWPAH - obj.ABaseARM64:     true,
	ASWPALB - obj.ABaseARM64:    true,
	ASWPALD - obj.ABaseARM64:    true,
	ASWPALH - obj.ABaseARM64:    true,
	ASWPALW - obj.ABaseARM64:    true,
	ASWPAW - obj.ABaseARM64:     true,
	ASWPB - obj.ABaseARM64:      true,
	ASWPD - obj.ABaseARM64:      true,
	ASWPH - obj.ABaseARM64:      true,
	ASWPLB - obj.ABaseARM64:     true,
	ASWPLD - obj.ABaseARM64:     true,
	ASWPLH - obj.ABaseARM64:     true,
	ASWPLW - obj.ABaseARM64:     true,
	ASWPW - obj.ABaseARM64:      true,
	ASXTB - obj.ABaseARM64:      true,
	ASXTBW - obj.ABaseARM64:     true,
	ASXTH - obj.ABaseARM64:      true,
	ASXTHW - obj.ABaseARM64:     true,
	ASXTW - obj.ABaseARM64:      true,
	ASYS - obj.ABaseARM64:       true,
	ASYSL - obj.ABaseARM64:      true,
	ATBNZ - obj.ABaseARM64:      true,
	ATBZ - obj.ABaseARM64:       true,
	ATLBI - obj.ABaseARM64:      true,
	ATST - obj.ABaseARM64:       true,
	ATSTW - obj.ABaseARM64:      true,
	AUBFIZ - obj.ABaseARM64:     true,
	AUBFIZW - obj.ABaseARM64:    true,
	AUBFM - obj.ABaseARM64:      true,
	AUBFMW - obj.ABaseARM64:     true,
	AUBFX - obj.ABaseARM64:      true,
	AUBFXW - obj.ABaseARM64:     true,
	AUCVTFD - obj.ABaseARM64:    true,
	AUCVTFS - obj.ABaseARM64:    true,
	AUCVTFWD - obj.ABaseARM64:   true,
	AUCVTFWS - obj.ABaseARM64:   true,
	AUDIV - obj.ABaseARM64:      true,
	AUDIVW - obj.ABaseARM64:     true,
	AUMADDL - obj.ABaseARM64:    true,
	AUMNEGL - obj.ABaseARM64:    true,
	AUMSUBL - obj.ABaseARM64:    true,
	AUMULH - obj.ABaseARM64:     true,
	AUMULL - obj.ABaseARM64:     true,
	AUREM - obj.ABaseARM64:      true,
	AUREMW - obj.ABaseARM64:     true,
	AUXTB - obj.ABaseARM64:      true,
	AUXTBW - obj.ABaseARM64:     true,
	AUXTH - obj.ABaseARM64:      true,
	AUXTHW - obj.ABaseARM64:     true,
	AUXTW - obj.ABaseARM64:      true,
	AVADD - obj.ABaseARM64:      true,
	AVADDP - obj.ABaseARM64:     true,
	AVADDV - obj.ABaseARM64:     true,
	AVAND - obj.ABaseARM64:      true,
	AVBCAX - obj.ABaseARM64:     true,
	AVBIF - obj.ABaseARM64:      true,
	AVBIT - obj.ABaseARM64:      true,
	AVBSL - obj.ABaseARM64:      true,
	AVCMEQ - obj.ABaseARM64:     true,
	AVCMTST - obj.ABaseARM64:    true,
	AVCNT - obj.ABaseARM64:      true,
	AVDUP - obj.ABaseARM64:      true,
	AVEOR - obj.ABaseARM64:      true,
	AVEOR3 - obj.ABaseARM64:     true,
	AVEXT - obj.ABaseARM64:      true,
	AVFMLA - obj.ABaseARM64:     true,
	AVFMLS - obj.ABaseARM64:     true,
	AVLD1 - obj.ABaseARM64:      true,
	AVLD1R - obj.ABaseARM64:     true,
	AVLD2 - obj.ABaseARM64:      true,
	AVLD2R - obj.ABaseARM64:     true,
	AVLD3 - obj.ABaseARM64:      true,
	AVLD3R - obj.ABaseARM64:     true,
	AVLD4 - obj.ABaseARM64:      true,
	AVLD4R - obj.ABaseARM64:     true,
	AVMOV - obj.ABaseARM64:      true,
	AVMOVD - obj.ABaseARM64:     true,
	AVMOVI - obj.ABaseARM64:     true,
	AVMOVQ - obj.ABaseARM64:     true,
	AVMOVS - obj.ABaseARM64:     true,
	AVORR - obj.ABaseARM64:      true,
	AVPMULL - obj.ABaseARM64:    true,
	AVPMULL2 - obj.ABaseARM64:   true,
	AVRAX1 - obj.ABaseARM64:     true,
	AVRBIT - obj.ABaseARM64:     true,
	AVREV16 - obj.ABaseARM64:    true,
	AVREV32 - obj.ABaseARM64:    true,
	AVREV64 - obj.ABaseARM64:    true,
	AVSHL - obj.ABaseARM64:      true,
	AVSLI - obj.ABaseARM64:      true,
	AVSRI - obj.ABaseARM64:      true,
	AVST1 - obj.ABaseARM64:      true,
	AVST2 - obj.ABaseARM64:      true,
	AVST3 - obj.ABaseARM64:      true,
	AVST4 - obj.ABaseARM64:      true,
	AVSUB - obj.ABaseARM64:      true,
	AVTBL - obj.ABaseARM64:      true,
	AVTBX - obj.ABaseARM64:      true,
	AVTRN1 - obj.ABaseARM64:     true,
	AVTRN2 - obj.ABaseARM64:     true,
	AVUADDLV - obj.ABaseARM64:   true,
	AVUADDW - obj.ABaseARM64:    true,
	AVUADDW2 - obj.ABaseARM64:   true,
	AVUMAX - obj.ABaseARM64:     true,
	AVUMIN - obj.ABaseARM64:     true,
	AVUSHLL - obj.ABaseARM64:    true,
	AVUSHLL2 - obj.ABaseARM64:   true,
	AVUSHR - obj.ABaseARM64:     true,
	AVUSRA - obj.ABaseARM64:     true,
	AVUXTL - obj.ABaseARM64:     true,
	AVUXTL2 - obj.ABaseARM64:    true,
	AVUZP1 - obj.ABaseARM64:     true,
	AVUZP2 - obj.ABaseARM64:     true,
	AVXAR - obj.ABaseARM64:      true,
	AVZIP1 - obj.ABaseARM64:     true,
	AVZIP2 - obj.ABaseARM64:     true,
	AWFE - obj.ABaseARM64:       true,
	AWFI - obj.ABaseARM64:       true,
	AWORD - obj.ABaseARM64:      true,
	AYIELD - obj.ABaseARM64:     true,
}
