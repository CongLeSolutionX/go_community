// cmd/7l/list.c and cmd/7l/sub.c from Vita Nuova.
// https://bitbucket.org/plan9-from-bell-labs/9-cc/src/master/
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
	obj.RegisterOpcode(obj.ABaseARM64, Anames)
	obj.RegisterOpSuffix("arm64", obj.CConvARM)
	obj.RegisterSpecialOperands(int64(SPOP_BEGIN), int64(SPOP_END), SPCconv)
}

func arrange(a int) string {
	switch a {
	case ARNG_8B:
		return "B8"
	case ARNG_16B:
		return "B16"
	case ARNG_4H:
		return "H4"
	case ARNG_8H:
		return "H8"
	case ARNG_2S:
		return "S2"
	case ARNG_4S:
		return "S4"
	case ARNG_1D:
		return "D1"
	case ARNG_2D:
		return "D2"
	case ARNG_B:
		return "B"
	case ARNG_H:
		return "H"
	case ARNG_S:
		return "S"
	case ARNG_D:
		return "D"
	case ARNG_1Q:
		return "Q1"
	default:
		return ""
	}
}

func ParseNEONArrangement(data string) int {
	switch data {
	case "B8":
		return ARNG_8B
	case "B16":
		return ARNG_16B
	case "H4":
		return ARNG_4H
	case "H8":
		return ARNG_8H
	case "S2":
		return ARNG_2S
	case "S4":
		return ARNG_4S
	case "D1":
		return ARNG_1D
	case "D2":
		return ARNG_2D
	case "B":
		return ARNG_B
	case "H":
		return ARNG_H
	case "S":
		return ARNG_S
	case "D":
		return ARNG_D
	case "Q1":
		return ARNG_1Q
	default:
		return ARNG_INVALID
	}
}

func ParseIntegerExtension(data string) int {
	switch data {
	case "UXTB":
		return REG_UXTB
	case "UXTH":
		return REG_UXTH
	case "UXTW":
		return REG_UXTW
	case "UXTX":
		return REG_UXTX
	case "SXTB":
		return REG_SXTB
	case "SXTH":
		return REG_SXTH
	case "SXTW":
		return REG_SXTW
	case "SXTX":
		return REG_SXTX
	case "LSL":
		return REG_LSL
	default:
		return obj.REG_NONE
	}
}

func Rconv(r int) string {
	if IsSVECompatibleRegister(int16(r)) {
		reg := AsRegister(int16(r))
		return reg.String()
	}

	ext := (r >> 5) & 7
	if r == REGG {
		return "g"
	}
	switch {
	case REG_R0 <= r && r <= REG_R30:
		return fmt.Sprintf("R%d", r-REG_R0)
	case r == REG_R31:
		return "ZR"
	case REG_F0 <= r && r <= REG_F31:
		return fmt.Sprintf("F%d", r-REG_F0)
	case REG_V0 <= r && r <= REG_V31:
		return fmt.Sprintf("V%d", r-REG_V0)
	case REG_Z0 <= r && r <= REG_Z31:
		return fmt.Sprintf("Z%d", r-REG_Z0)
	case REG_P0 <= r && r <= REG_P15:
		return fmt.Sprintf("P%d", r-REG_P0)
	case r == REGSP:
		return "RSP"
	case REG_UXTB <= r && r < REG_UXTH:
		if ext != 0 {
			return fmt.Sprintf("%s.UXTB<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.UXTB", regname(r))
		}
	case REG_UXTH <= r && r < REG_UXTW:
		if ext != 0 {
			return fmt.Sprintf("%s.UXTH<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.UXTH", regname(r))
		}
	case REG_UXTW <= r && r < REG_UXTX:
		if ext != 0 {
			return fmt.Sprintf("%s.UXTW<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.UXTW", regname(r))
		}
	case REG_UXTX <= r && r < REG_SXTB:
		if ext != 0 {
			return fmt.Sprintf("%s.UXTX<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.UXTX", regname(r))
		}
	case REG_SXTB <= r && r < REG_SXTH:
		if ext != 0 {
			return fmt.Sprintf("%s.SXTB<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.SXTB", regname(r))
		}
	case REG_SXTH <= r && r < REG_SXTW:
		if ext != 0 {
			return fmt.Sprintf("%s.SXTH<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.SXTH", regname(r))
		}
	case REG_SXTW <= r && r < REG_SXTX:
		if ext != 0 {
			return fmt.Sprintf("%s.SXTW<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.SXTW", regname(r))
		}
	case REG_SXTX <= r && r < REG_SPECIAL:
		if ext != 0 {
			return fmt.Sprintf("%s.SXTX<<%d", regname(r), ext)
		} else {
			return fmt.Sprintf("%s.SXTX", regname(r))
		}
	// bits 0-4 indicate register, bits 5-7 indicate shift amount, bit 8 equals to 0.
	case REG_LSL <= r && r < (REG_LSL+1<<8):
		return fmt.Sprintf("R%d<<%d", r&31, (r>>5)&7)
	case REG_ARNG <= r && r < REG_UXTB:
		return fmt.Sprintf("V%d.%s", r&31, arrange((r>>5)&15))
	}
	// Return system register name.
	name, _, _ := SysRegEnc(int16(r))
	if name != "" {
		return name
	}
	return fmt.Sprintf("badreg(%d)", r)
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
		return fmt.Sprintf("%s", spc)
	}
	return "SPC_??"
}

func rlconv(list int64) string {
	rlist := ARM64RegisterList{uint64(list)}

	str := Rconv(int(rlist.Base()))

	for i := 1; i < int(rlist.Count()); i++ {
		str += fmt.Sprintf(",%s", Rconv(int(rlist.GetRegisterAtIndex(i))))
	}

	return "[" + str + "]"
}

func regname(r int) string {
	if r&31 == 31 {
		return "ZR"
	}
	return fmt.Sprintf("R%d", r&31)
}

var RegisterList map[string]int16

func InitRegisterList() {
	RegisterList = make(map[string]int16)
	// Create maps for easy lookup of instruction names etc.
	// Note that there is no list of names as there is for 386 and amd64.
	RegisterList[Rconv(REGSP)] = int16(REGSP)
	for i := REG_R0; i <= REG_R31; i++ {
		RegisterList[Rconv(i)] = int16(i)
	}
	// Rename R18 to R18_PLATFORM to avoid accidental use.
	RegisterList["R18_PLATFORM"] = RegisterList["R18"]
	delete(RegisterList, "R18")
	for i := REG_F0; i <= REG_F31; i++ {
		RegisterList[Rconv(i)] = int16(i)
	}
	for i := REG_V0; i <= REG_V31; i++ {
		RegisterList[Rconv(i)] = int16(i)
	}
	for i := REG_Z0; i <= REG_Z31; i++ {
		RegisterList[Rconv(i)] = int16(i)
	}
	for i := REG_P0; i <= REG_P15; i++ {
		RegisterList[Rconv(i)] = int16(i)
	}

	// System registers.
	for i := 0; i < len(SystemReg); i++ {
		RegisterList[SystemReg[i].Name] = SystemReg[i].Reg
	}

	RegisterList["LR"] = REGLINK

	// Avoid unintentionally clobbering g using R28.
	delete(RegisterList, "R28")
	RegisterList["g"] = REG_R28
}
