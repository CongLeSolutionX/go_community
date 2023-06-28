// cmd/7l/list.c and cmd/7l/sub.c from Vita Nuova.
// https://code.google.com/p/ken-cc/source/browse/
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

import (
	"bytes"
	"cmd/internal/obj"
	"fmt"
)

var strcond = [16]string{
	"EQ",
	"NE",
	"HS",
	"LO",
	"MI",
	"PL",
	"VS",
	"VC",
	"HI",
	"LS",
	"GE",
	"LT",
	"GT",
	"LE",
	"AL",
	"NV",
}

func init() {
	obj.RegisterRegister(obj.RBaseARM64, REG_SPECIAL+1024, rconv)
	obj.RegisterAconvFunc("arm64", aconvReg, aconvMem, aconvShift, aconvRegList)
	obj.RegisterOpcode(obj.ABaseARM64, Anames)
	obj.RegisterOpSuffix("arm64", obj.CConvARM)
	obj.RegisterSpecialOperands(int64(SPOP_BEGIN), int64(SPOP_END), SPCconv)
}

func EncodeIndex(arng, types, amount int16) int16 {
	return arng<<11 | types<<6 | amount
}

func DecodeIndex(index int16) (arng, types, amount uint32) {
	arng = uint32(index>>11) & 0x1f
	types = uint32(index>>6) & 0x1f
	amount = uint32(index) & 0x3f
	return
}

func arrange(a int16) string {
	switch a {
	case ARNG_4B:
		return ".B4"
	case ARNG_8B:
		return ".B8"
	case ARNG_16B:
		return ".B16"
	case ARNG_2H:
		return ".H2"
	case ARNG_4H:
		return ".H4"
	case ARNG_8H:
		return ".H8"
	case ARNG_2S:
		return ".S2"
	case ARNG_4S:
		return ".S4"
	case ARNG_1D:
		return ".D1"
	case ARNG_2D:
		return ".D2"
	case ARNG_B:
		return ".B"
	case ARNG_H:
		return ".H"
	case ARNG_S:
		return ".S"
	case ARNG_D:
		return ".D"
	case ARNG_1Q:
		return ".Q1"
	default:
		return ""
	}
}

func rconv(r int) string {
	switch {
	case REG_R0 <= r && r <= REG_R30:
		if r == REGG {
			return "g"
		}
		return fmt.Sprintf("R%d", r-REG_R0)
	case r == REG_R31:
		return "ZR"
	case REG_F0 <= r && r <= REG_F31:
		return fmt.Sprintf("F%d", r-REG_F0)
	case REG_V0 <= r && r <= REG_V31:
		return fmt.Sprintf("V%d", r-REG_V0)
	case r == REGSP:
		return "RSP"
	default:
		return fmt.Sprintf("badreg(%d)", r)
	}
}

func formatReg(reg, arng, typ, amount int16) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s%s", rconv(int(reg)), arrange(arng))

	switch typ {
	case RTYP_INDEX:
		fmt.Fprintf(buf, "[%d]", amount)
	case RTYP_EXT_LSL:
		fmt.Fprintf(buf, "<<%d", amount)
	case RTYP_EXT_UXTB:
		fmt.Fprintf(buf, ".UXTB")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_UXTH:
		fmt.Fprintf(buf, ".UXTH")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_UXTW:
		fmt.Fprintf(buf, ".UXTW")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_UXTX:
		fmt.Fprintf(buf, ".UXTX")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_SXTB:
		fmt.Fprintf(buf, ".SXTB")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_SXTH:
		fmt.Fprintf(buf, ".SXTH")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_SXTW:
		fmt.Fprintf(buf, ".SXTW")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	case RTYP_EXT_SXTX:
		fmt.Fprintf(buf, ".SXTX")
		if amount != 0 {
			fmt.Fprintf(buf, "<<%d", amount)
		}
	}
	return buf.String()
}

func aconvReg(a *obj.Addr) string {
	if a.Reg > REG_SPECIAL {
		name, _, _ := SysRegEnc(int16(a.Reg))
		if name != "" {
			return name
		}
		return fmt.Sprintf("badreg(%d)", a.Reg)
	}
	arng, typ, amount := DecodeIndex(a.Index)
	return formatReg(a.Reg, int16(arng), int16(typ), int16(amount))
}

func aconvShift(a *obj.Addr) string {
	_, shift, amount := DecodeIndex(a.Index)
	switch shift {
	case SHIFT_LL:
		return fmt.Sprintf("%s<<%d", rconv(int(a.Reg)), amount)
	case SHIFT_LR:
		return fmt.Sprintf("%s>>%d", rconv(int(a.Reg)), amount)
	case SHIFT_AR:
		return fmt.Sprintf("%s->%d", rconv(int(a.Reg)), amount)
	case SHIFT_ROR:
		return fmt.Sprintf("%s@>%d", rconv(int(a.Reg)), amount)
	default:
		return "R???"
	}
}

func aconvMem(a *obj.Addr) string {
	buf := new(bytes.Buffer)
	if a.Name != obj.NAME_NONE {
		a.WriteNameTo(buf)
	} else {
		arng, typ, amount := DecodeIndex(a.Index)
		reg := formatReg(a.Reg, int16(arng), int16(typ), int16(amount))
		if typ != RTYP_MEM_ROFF {
			// const offset
			if a.Offset != 0 {
				fmt.Fprintf(buf, "%d(%s)", a.Offset, reg)
			} else {
				fmt.Fprintf(buf, "(%s)", reg)
			}
		} else {
			// register offset
			arng, typ, amount := DecodeIndex(int16(a.Offset >> 16))
			fmt.Fprintf(buf, "(%s)(%s)", reg, formatReg(int16(a.Offset&0xffff), int16(arng), int16(typ), int16(amount)))
		}
	}
	return buf.String()
}

func DRconv(a int) string {
	if a >= C_NONE && a <= C_NCLASS {
		return cnames7[a]
	}
	return "C_??"
}

func SPCconv(a int64) string {
	spc := SpecialOperand(a)
	if spc >= SPOP_BEGIN && spc < SPOP_END {
		return spc.String()
	}
	return "SPC_??"
}

func aconvRegList(a *obj.Addr) string {
	str := ""
	firstReg := a.Reg
	arng, typ, index := DecodeIndex(a.Index)
	regCnt := a.Offset
	scale := a.Scale

	regbase := int(firstReg &^ 0x1f)
	r := int(firstReg & 0x1f)
	for i := 0; i < int(regCnt); i++ {
		if str == "" {
			str += "["
		} else {
			str += ","
		}
		reg := (r+int(scale)*i)&31 + regbase
		str += rconv(reg) + arrange(int16(arng))
	}
	str += "]"
	if typ == RTYP_INDEX {
		return fmt.Sprintf("%s[%d]", str, index)
	}
	return str
}
