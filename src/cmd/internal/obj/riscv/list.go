//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2008 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2008 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
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

package riscv

import (
	"fmt"

	"cmd/internal/obj"
)

// regNames is a map for register IDs to names. We use the ABI name.
var regNames = map[int16]string{
	0: "NONE",

	// General registers with ABI names.
	REG_ZERO: "ZERO",
	REG_RA:   "RA",
	REG_SP:   "SP",
	REG_GP:   "GP",
	// REG_TP is REG_G
	REG_T0: "T0",
	REG_T1: "T1",
	REG_T2: "T2",
	REG_S0: "S0",
	REG_S1: "S1",
	REG_A0: "A0",
	REG_A1: "A1",
	REG_A2: "A2",
	REG_A3: "A3",
	REG_A4: "A4",
	REG_A5: "A5",
	REG_A6: "A6",
	REG_A7: "A7",
	REG_S2: "S2",
	REG_S3: "S3",
	// REG_S4 is REG_CTXT.
	REG_S5:  "S5",
	REG_S6:  "S6",
	REG_S7:  "S7",
	REG_S8:  "S8",
	REG_S9:  "S9",
	REG_S10: "S10",
	REG_S11: "S11",
	REG_T3:  "T3",
	REG_T4:  "T4",
	REG_T5:  "T5",
	// REG_T6 is REG_TMP.

	// Go runtime register names.
	REG_G:    "g",
	REG_CTXT: "CTXT",
	REG_TMP:  "TMP",

	// ABI names for floating point registers.
	REG_FT0:  "FT0",
	REG_FT1:  "FT1",
	REG_FT2:  "FT2",
	REG_FT3:  "FT3",
	REG_FT4:  "FT4",
	REG_FT5:  "FT5",
	REG_FT6:  "FT6",
	REG_FT7:  "FT7",
	REG_FS0:  "FS0",
	REG_FS1:  "FS1",
	REG_FA0:  "FA0",
	REG_FA1:  "FA1",
	REG_FA2:  "FA2",
	REG_FA3:  "FA3",
	REG_FA4:  "FA4",
	REG_FA5:  "FA5",
	REG_FA6:  "FA6",
	REG_FA7:  "FA7",
	REG_FS2:  "FS2",
	REG_FS3:  "FS3",
	REG_FS4:  "FS4",
	REG_FS5:  "FS5",
	REG_FS6:  "FS6",
	REG_FS7:  "FS7",
	REG_FS8:  "FS8",
	REG_FS9:  "FS9",
	REG_FS10: "FS10",
	REG_FS11: "FS11",
	REG_FT8:  "FT8",
	REG_FT9:  "FT9",
	REG_FT10: "FT10",
	REG_FT11: "FT11",
}

// checkRegNames asserts that regNames includes names for all registers.
func checkRegNames() {
	for i := REG_X0; i <= REG_X31; i++ {
		if _, ok := regNames[int16(i)]; !ok {
			panic(fmt.Sprintf("REG_X%d missing from regNames", i))
		}
	}
	for i := REG_F0; i <= REG_F31; i++ {
		if _, ok := regNames[int16(i)]; !ok {
			panic(fmt.Sprintf("REG_F%d missing from regNames", i))
		}
	}
}

func init() {
	checkRegNames()

	obj.RegisterRegister(obj.RBaseRISCV, REG_END, regName)
	obj.RegisterOpcode(obj.ABaseRISCV, Anames)
}

func regName(r int) string {
	if name, ok := regNames[int16(r)]; ok {
		return name
	}
	return fmt.Sprintf("R???%d", r) // Similar format to As.
}
