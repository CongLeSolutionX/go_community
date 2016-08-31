// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the
// 32-bit MIPS (MIPS32) instruction set, to minimize its interaction
// with the core of the assembler.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
)

func jumpMIPS(word string) bool {
	switch word {
	case "BEQ", "BFPF", "BFPT", "BGEZ", "BGEZAL", "BGTZ", "BLEZ", "BLTZ", "BLTZAL", "BNE", "JMP", "JAL", "CALL":
		return true
	}
	return false
}

// IsMIPSCMP reports whether the op (as defined by an mips32.A* constant) is
// one of the CMP instructions that require special handling.
func IsMIPSCMP(op obj.As) bool {
	switch op {
	case mips32.ACMPEQF, mips32.ACMPEQD, mips32.ACMPGEF, mips32.ACMPGED,
		mips32.ACMPGTF, mips32.ACMPGTD:
		return true
	}
	return false
}

// IsMIPSMUL reports whether the op (as defined by an mips32.A* constant) is
// one of the MUL/DIV/REM instructions that require special handling.
func IsMIPSMUL(op obj.As) bool {
	switch op {
	case mips32.AMUL, mips32.AMULU, mips32.ADIV, mips32.ADIVU,
		mips32.AREM, mips32.AREMU:
		return true
	}
	return false
}

func mips32RegisterNumber(name string, n int16) (int16, bool) {
	switch name {
	case "F":
		if 0 <= n && n <= 31 {
			return mips32.REG_F0 + n, true
		}
	case "FCR":
		if 0 <= n && n <= 31 {
			return mips32.REG_FCR0 + n, true
		}
	case "M":
		if 0 <= n && n <= 31 {
			return mips32.REG_M0 + n, true
		}
	case "R":
		if 0 <= n && n <= 31 {
			return mips32.REG_R0 + n, true
		}
	}
	return 0, false
}
