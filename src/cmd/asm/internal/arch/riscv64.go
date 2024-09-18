// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the RISCV64
// instruction set, to minimize its interaction with the core of the
// assembler.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
)

// IsRISCV64AMO reports whether the op (as defined by a riscv.A*
// constant) is one of the AMO instructions that requires special
// handling.
func IsRISCV64AMO(op obj.As) bool {
	switch op {
	case riscv.ASCW, riscv.ASCD, riscv.AAMOSWAPW, riscv.AAMOSWAPD, riscv.AAMOADDW, riscv.AAMOADDD,
		riscv.AAMOANDW, riscv.AAMOANDD, riscv.AAMOORW, riscv.AAMOORD, riscv.AAMOXORW, riscv.AAMOXORD,
		riscv.AAMOMINW, riscv.AAMOMIND, riscv.AAMOMINUW, riscv.AAMOMINUD,
		riscv.AAMOMAXW, riscv.AAMOMAXD, riscv.AAMOMAXUW, riscv.AAMOMAXUD:
		return true
	}
	return false
}

// IsRISCV64CSRO reports whether the op is an instruction that uses
// CSR symbolic names and whether that instruction expects a register
// or an immediate source operand.
func IsRISCV64CSRO(op obj.As) (imm bool, ok bool) {
	switch op {
	case riscv.ACSRRWI, riscv.ACSRRSI, riscv.ACSRRCI:
		imm = true
		fallthrough
	case riscv.ACSRRW, riscv.ACSRRS, riscv.ACSRRC:
		ok = true
	}
	return
}

var riscv64SpecialOperand map[string]riscv.SpecialOperand

// GetRISCV64SpecialOperand returns the internal representation of a special operand.
func GetRISCV64SpecialOperand(name string) riscv.SpecialOperand {
	if riscv64SpecialOperand == nil {
		// Generate the mapping automatically when the first time the function is called.
		riscv64SpecialOperand = map[string]riscv.SpecialOperand{}

		// Add the CSRs
		for csrCode, csrName := range riscv.CSRs {
			riscv64SpecialOperand[csrName] = riscv.SpecialOperand(int64(csrCode) + int64(riscv.SPOP_CSR_BEGIN))
		}
	}
	if opd, ok := riscv64SpecialOperand[name]; ok {
		return opd
	}
	return riscv.SPOP_END
}
