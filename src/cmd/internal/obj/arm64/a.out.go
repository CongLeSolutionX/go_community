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

	// The EQ in
	// 	CSET	EQ, R0
	// is encoded as TYPE_REG, even though it's not really a register.
	COND_EQ
	COND_NE
	COND_HS
	COND_LO
	COND_MI
	COND_PL
	COND_VS
	COND_VC
	COND_HI
	COND_LS
	COND_GE
	COND_LT
	COND_GT
	COND_LE
	COND_AL
	COND_NV

	REG_RSP = REG_V31 + 32 // to differentiate ZR/SP, REG_RSP&0x1f = 31
)

// bits 0-4 indicates register: Vn
// bits 5-8 indicates arrangement: <T>
const (
	REG_ARNG = obj.RBaseARM64 + 1<<10 + iota<<9 // Vn.<T>
	REG_ELEM                                    // Vn.<T>[index]
	REG_ELEM_END
)

// Not registers, but flags that can be combined with regular register
// constants to indicate extended register conversion. When checking,
// you should subtract obj.RBaseARM64 first. From this difference, bit 11
// indicates extended register, bits 8-10 select the conversion mode.
// REG_LSL is the index shift specifier, bit 9 indicates shifted offset register.
const REG_LSL = obj.RBaseARM64 + 1<<9
const REG_EXT = obj.RBaseARM64 + 1<<11

const (
	REG_UXTB = REG_EXT + iota<<8
	REG_UXTH
	REG_UXTW
	REG_UXTX
	REG_SXTB
	REG_SXTH
	REG_SXTW
	REG_SXTX
)

// Special registers, after subtracting obj.RBaseARM64, bit 12 indicates
// a special register and the low bits select the register.
// SYSREG_END is the last item in the automatically generated system register
// declaration, and it is defined in the sysRegEnc.go file.
const (
	REG_SPECIAL = obj.RBaseARM64 + 1<<12
	REG_DAIFSet = SYSREG_END + iota
	REG_DAIFClr
	REG_PLDL1KEEP
	REG_PLDL1STRM
	REG_PLDL2KEEP
	REG_PLDL2STRM
	REG_PLDL3KEEP
	REG_PLDL3STRM
	REG_PLIL1KEEP
	REG_PLIL1STRM
	REG_PLIL2KEEP
	REG_PLIL2STRM
	REG_PLIL3KEEP
	REG_PLIL3STRM
	REG_PSTL1KEEP
	REG_PSTL1STRM
	REG_PSTL2KEEP
	REG_PSTL2STRM
	REG_PSTL3KEEP
	REG_PSTL3STRM
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
	REGRT1 = REG_R16 // ARM64 IP0, external linker may use as a scrach register in trampoline
	REGRT2 = REG_R17 // ARM64 IP1, external linker may use as a scrach register in trampoline
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
}

const (
	/* mark flags */
	LEAF         = 1 << iota
	LFROM        // p.From uses constant pool
	LFROM128     // p.From3<<64+p.From forms a 128-bit constant in literal pool
	LTO          // p.To uses constant pool
	NOTUSETMP    // p expands to multiple instructions, but does NOT use REGTMP
	BRANCH14BITS // branch instruction encodes 14 bits
	BRANCH19BITS // branch instruction encodes 19 bits
	BRANCH26BITS // branch instruction encodes 26 bits
)

// Big classes of the instruction argument.
const (
	C_NONE argtype = iota
	// register class
	C_REG    // R0..R30
	C_RSP    // R0..R30, RSP
	C_FREG   // F0..F31
	C_VREG   // V0..V31
	C_PAIR   // (Rn, Rm)
	C_SHIFT  // Rn<<2
	C_EXTREG // Rn.UXTB[<<3]
	C_SPR    // special register, such as REG_DAIFSet
	C_COND   // EQ, NE, etc
	C_ARNG   // Vn.<T>
	C_ELEM   // Vn.<T>[index]
	C_LIST   // [V1, V2, V3]

	// constant class
	C_ZCON // $0 or ZR
	C_LCON // 32-bit integer constant
	C_VCON // 64-bit integer constant
	C_FCON // floating-point constant
	// a constant that can be loaded with one MOVZ, optionally shifted by multiple of 16,
	// and zero or several MOVKs.
	C_MOVCONZ
	// a constant that can be loaded with one MOVN, optionally shifted by multiple of 16,
	// and zero or several MOVKs.
	C_MOVCONN

	// address class
	C_VCONADDR // 64-bit memory address constant $foo+imm(SB), will be tranlated to adrp+add(reloc)
	C_LACON    // 64-bit offset in auto constant $imm(R)

	// branch class
	C_SBRA // for TYPE_BRANCH

	// memory class
	C_ZOREG   // (R), 0(R), (RSP), 0(RSP)
	C_LOREG   // 32-bit register offset, imm32(R)
	C_VOREG   // 64-bit register offset, imm64(R)
	C_ROFF    // register offset (including register extended), [<Xn|SP>, (<Wm>|<Xm>){, <extend> {<amount>}}]
	C_ADDR    // 64-bit memory address foo(SB), will be tranlated to adrp+add+ldr(reloc)
	C_GOTADDR // The GOT slot for a symbol in -dynlink mode
	// TLS "var" in local exec mode: will become a constant offset from
	// thread local base that is ultimately chosen by the program linker.
	C_TLS_LE
	// TLS "var" in initial exec mode: will become a memory address (chosen
	// by the program linker) that the dynamic linker will fill with the
	// offset from the thread local base.
	C_TLS_IE

	C_GOK
	C_TEXTSIZE
	C_NCLASS // must be last
)

const (
	C_XPRE  = 1 << 6 // match arm.C_WBIT, so Prog.String know how to print it
	C_XPOST = 1 << 5 // match arm.C_PBIT, so Prog.String know how to print it
)

//go:generate go run ../stringer.go -i $GOFILE -o anames.go -p arm64

const (
	AADC = obj.ABaseARM64 + obj.A_ARCHSPECIFIC + iota
	AADCS
	AADCSW
	AADCW
	AADD
	AADDS
	AADDSW
	AADDW
	AADR
	AADRP
	AAND
	AANDS
	AANDSW
	AANDW
	AASR
	AASRW
	AAT
	ABFI
	ABFIW
	ABFM
	ABFMW
	ABFXIL
	ABFXILW
	ABIC
	ABICS
	ABICSW
	ABICW
	ABRK
	ACBNZ
	ACBNZW
	ACBZ
	ACBZW
	ACCMN
	ACCMNW
	ACCMP
	ACCMPW
	ACINC
	ACINCW
	ACINV
	ACINVW
	ACLREX
	ACLS
	ACLSW
	ACLZ
	ACLZW
	ACMN
	ACMNW
	ACMP
	ACMPW
	ACNEG
	ACNEGW
	ACRC32B
	ACRC32CB
	ACRC32CH
	ACRC32CW
	ACRC32CX
	ACRC32H
	ACRC32W
	ACRC32X
	ACSEL
	ACSELW
	ACSET
	ACSETM
	ACSETMW
	ACSETW
	ACSINC
	ACSINCW
	ACSINV
	ACSINVW
	ACSNEG
	ACSNEGW
	ADC
	ADCPS1
	ADCPS2
	ADCPS3
	ADMB
	ADRPS
	ADSB
	AEON
	AEONW
	AEOR
	AEORW
	AERET
	AEXTR
	AEXTRW
	AHINT
	AHLT
	AHVC
	AIC
	AISB
	ALDADDAB
	ALDADDAD
	ALDADDAH
	ALDADDAW
	ALDADDALB
	ALDADDALD
	ALDADDALH
	ALDADDALW
	ALDADDB
	ALDADDD
	ALDADDH
	ALDADDW
	ALDADDLB
	ALDADDLD
	ALDADDLH
	ALDADDLW
	ALDAR
	ALDARB
	ALDARH
	ALDARW
	ALDAXP
	ALDAXPW
	ALDAXR
	ALDAXRB
	ALDAXRH
	ALDAXRW
	ALDCLRAB
	ALDCLRAD
	ALDCLRAH
	ALDCLRAW
	ALDCLRALB
	ALDCLRALD
	ALDCLRALH
	ALDCLRALW
	ALDCLRB
	ALDCLRD
	ALDCLRH
	ALDCLRW
	ALDCLRLB
	ALDCLRLD
	ALDCLRLH
	ALDCLRLW
	ALDEORAB
	ALDEORAD
	ALDEORAH
	ALDEORAW
	ALDEORALB
	ALDEORALD
	ALDEORALH
	ALDEORALW
	ALDEORB
	ALDEORD
	ALDEORH
	ALDEORW
	ALDEORLB
	ALDEORLD
	ALDEORLH
	ALDEORLW
	ALDORAB
	ALDORAD
	ALDORAH
	ALDORAW
	ALDORALB
	ALDORALD
	ALDORALH
	ALDORALW
	ALDORB
	ALDORD
	ALDORH
	ALDORW
	ALDORLB
	ALDORLD
	ALDORLH
	ALDORLW
	ALDP
	ALDPW
	ALDPSW
	ALDXR
	ALDXRB
	ALDXRH
	ALDXRW
	ALDXP
	ALDXPW
	ALSL
	ALSLW
	ALSR
	ALSRW
	AMADD
	AMADDW
	AMNEG
	AMNEGW
	AMOVK
	AMOVKW
	AMOVN
	AMOVNW
	AMOVZ
	AMOVZW
	AMRS
	AMSR
	AMSUB
	AMSUBW
	AMUL
	AMULW
	AMVN
	AMVNW
	ANEG
	ANEGS
	ANEGSW
	ANEGW
	ANGC
	ANGCS
	ANGCSW
	ANGCW
	ANOOP
	AORN
	AORNW
	AORR
	AORRW
	APRFM
	APRFUM
	ARBIT
	ARBITW
	AREM
	AREMW
	AREV
	AREV16
	AREV16W
	AREV32
	AREVW
	AROR
	ARORW
	ASBC
	ASBCS
	ASBCSW
	ASBCW
	ASBFIZ
	ASBFIZW
	ASBFM
	ASBFMW
	ASBFX
	ASBFXW
	ASDIV
	ASDIVW
	ASEV
	ASEVL
	ASMADDL
	ASMC
	ASMNEGL
	ASMSUBL
	ASMULH
	ASMULL
	ASTXR
	ASTXRB
	ASTXRH
	ASTXP
	ASTXPW
	ASTXRW
	ASTLP
	ASTLPW
	ASTLR
	ASTLRB
	ASTLRH
	ASTLRW
	ASTLXP
	ASTLXPW
	ASTLXR
	ASTLXRB
	ASTLXRH
	ASTLXRW
	ASTP
	ASTPW
	ASUB
	ASUBS
	ASUBSW
	ASUBW
	ASVC
	ASXTB
	ASXTBW
	ASXTH
	ASXTHW
	ASXTW
	ASYS
	ASYSL
	ATBNZ
	ATBZ
	ATLBI
	ATST
	ATSTW
	AUBFIZ
	AUBFIZW
	AUBFM
	AUBFMW
	AUBFX
	AUBFXW
	AUDIV
	AUDIVW
	AUMADDL
	AUMNEGL
	AUMSUBL
	AUMULH
	AUMULL
	AUREM
	AUREMW
	AUXTB
	AUXTH
	AUXTW
	AUXTBW
	AUXTHW
	AWFE
	AWFI
	AYIELD
	AMOVB
	AMOVBU
	AMOVH
	AMOVHU
	AMOVW
	AMOVWU
	AMOVD
	AMOVNP
	AMOVNPW
	AMOVP
	AMOVPD
	AMOVPQ
	AMOVPS
	AMOVPSW
	AMOVPW
	ASWPAD
	ASWPAW
	ASWPAH
	ASWPAB
	ASWPALD
	ASWPALW
	ASWPALH
	ASWPALB
	ASWPD
	ASWPW
	ASWPH
	ASWPB
	ASWPLD
	ASWPLW
	ASWPLH
	ASWPLB
	ACASD
	ACASW
	ACASH
	ACASB
	ACASAD
	ACASAW
	ACASLD
	ACASLW
	ACASALD
	ACASALW
	ACASALH
	ACASALB
	ACASPD
	ACASPW
	ABEQ
	ABNE
	ABCS
	ABHS
	ABCC
	ABLO
	ABMI
	ABPL
	ABVS
	ABVC
	ABHI
	ABLS
	ABGE
	ABLT
	ABGT
	ABLE
	AFABSD
	AFABSS
	AFADDD
	AFADDS
	AFCCMPD
	AFCCMPED
	AFCCMPS
	AFCCMPES
	AFCMPD
	AFCMPED
	AFCMPES
	AFCMPS
	AFCVTSD
	AFCVTDS
	AFCVTZSD
	AFCVTZSDW
	AFCVTZSS
	AFCVTZSSW
	AFCVTZUD
	AFCVTZUDW
	AFCVTZUS
	AFCVTZUSW
	AFDIVD
	AFDIVS
	AFLDPD
	AFLDPQ
	AFLDPS
	AFMOVQ
	AFMOVD
	AFMOVS
	AVMOVQ
	AVMOVD
	AVMOVS
	AFMULD
	AFMULS
	AFNEGD
	AFNEGS
	AFSQRTD
	AFSQRTS
	AFSTPD
	AFSTPQ
	AFSTPS
	AFSUBD
	AFSUBS
	ASCVTFD
	ASCVTFS
	ASCVTFWD
	ASCVTFWS
	AUCVTFD
	AUCVTFS
	AUCVTFWD
	AUCVTFWS
	AWORD
	ADWORD
	AFCSELS
	AFCSELD
	AFMAXS
	AFMINS
	AFMAXD
	AFMIND
	AFMAXNMS
	AFMAXNMD
	AFNMULS
	AFNMULD
	AFRINTNS
	AFRINTND
	AFRINTPS
	AFRINTPD
	AFRINTMS
	AFRINTMD
	AFRINTZS
	AFRINTZD
	AFRINTAS
	AFRINTAD
	AFRINTXS
	AFRINTXD
	AFRINTIS
	AFRINTID
	AFMADDS
	AFMADDD
	AFMSUBS
	AFMSUBD
	AFNMADDS
	AFNMADDD
	AFNMSUBS
	AFNMSUBD
	AFMINNMS
	AFMINNMD
	AFCVTDH
	AFCVTHS
	AFCVTHD
	AFCVTSH
	AAESD
	AAESE
	AAESIMC
	AAESMC
	ASHA1C
	ASHA1H
	ASHA1M
	ASHA1P
	ASHA1SU0
	ASHA1SU1
	ASHA256H
	ASHA256H2
	ASHA256SU0
	ASHA256SU1
	ASHA512H
	ASHA512H2
	ASHA512SU0
	ASHA512SU1
	AVADD
	AVADDP
	AVAND
	AVBIF
	AVBCAX
	AVCMEQ
	AVCNT
	AVEOR
	AVEOR3
	AVMOV
	AVLD1
	AVLD2
	AVLD3
	AVLD4
	AVLD1R
	AVLD2R
	AVLD3R
	AVLD4R
	AVORR
	AVREV16
	AVREV32
	AVREV64
	AVST1
	AVST2
	AVST3
	AVST4
	AVDUP
	AVADDV
	AVMOVI
	AVUADDLV
	AVSUB
	AVFMLA
	AVFMLS
	AVPMULL
	AVPMULL2
	AVEXT
	AVRBIT
	AVRAX1
	AVUMAX
	AVUMIN
	AVUSHR
	AVUSHLL
	AVUSHLL2
	AVUXTL
	AVUXTL2
	AVUZP1
	AVUZP2
	AVSHL
	AVSRI
	AVSLI
	AVBSL
	AVBIT
	AVTBL
	AVXAR
	AVZIP1
	AVZIP2
	AVCMTST
	AVUADDW2
	AVUADDW
	AVUSRA
	ALAST
	AB  = obj.AJMP
	ABL = obj.ACALL
)

const (
	// shift types
	SHIFT_LL  = 0 << 22
	SHIFT_LR  = 1 << 22
	SHIFT_AR  = 2 << 22
	SHIFT_ROR = 3 << 22
)

// Arrangement for ARM64 SIMD instructions
const (
	// arrangement types
	ARNG_8B = iota
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
)
