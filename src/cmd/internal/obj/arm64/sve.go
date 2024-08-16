// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
)

type encoding struct {
	// Concatenation of all fixed bits in the encoding.
	base uint32
	// Which operand format does this encoding require? Key into format table.
	format []int
	// Which encoder does this encoding use? Key into encoder table.
	encoder int
}

// List of symbol types, one type for each operand. E.g. REG_Z | EXT_B, REG_R
type format = []int

type rule struct {
	// Indices of operands to pass to the rule.
	// E.g. to pass <Zd> and <Zn> from '<Zd>, <Pg>, <Zn>, <Zm>', create []int{0,2}
	symbolRefs []int
	// Rule function, returns 32-bit number to OR with the base, and a success flag.
	encode func(...*obj.Addr) (uint32, bool)
}

type encoder struct {
	// List of rules performed by this encoder
	rules []rule
}

func (e *encoding) assemble(p *obj.Prog) (uint32, bool) {
	args, ok := e.getFormattedArgs(p)
	if !ok {
		return 0, ok
	}

	result := e.base
	for _, r := range encoders[e.encoder].rules {
		selectedArgs := []*obj.Addr{}
		for _, sym := range r.symbolRefs {
			selectedArgs = append(selectedArgs, args[sym])
		}
		enc, ok := r.encode(selectedArgs...)
		if !ok {
			return 0, false
		}
		result |= enc
	}

	return result, true
}

func assembleSVE(p *obj.Prog) (uint32, error) {
	encodings, ok := instructionTable[p.As]

	if !ok {
		return 0, fmt.Errorf("encoding not found for: %s", p)
	}

	var result uint32
	for _, enc := range encodings {
		result, ok = enc.assemble(p)
		if ok {
			return result, nil
		}
	}
	return 0, fmt.Errorf("no encoding matches format of: %s", p)
}

// Unpacks ordered operands from the Prog structure. This reverses the packing process
// performed by ARM64AsmInstruction, see cmd/asm/internal/arch/arm64.go. The operands
// are validated against the format of this encoding instance, if there is a match it
// will return true as the second return value, otherwise false.
func (e *encoding) getFormattedArgs(prog *obj.Prog) ([]*obj.Addr, bool) {
	for _, format := range e.format {
		fmt := formats[format]

		args := []*obj.Addr{}

		if prog.To.Type != obj.TYPE_NONE {
			args = append(args, &prog.To)
		}

		if prog.RegTo2 != obj.REG_NONE {
			args = append(args, &obj.Addr{Type: obj.TYPE_REG, Reg: prog.RegTo2})
		}

		for _, arg := range prog.RestArgs {
			if arg.Pos == obj.Destination {
				args = append(args, &arg.Addr)
			}
		}

		if prog.From.Type != obj.TYPE_NONE {
			args = append(args, &prog.From)
		}

		if prog.Reg != obj.REG_NONE {
			args = append(args, &obj.Addr{Type: obj.TYPE_REG, Reg: prog.Reg})
		}

		for _, arg := range prog.RestArgs {
			if arg.Pos == obj.Source {
				args = append(args, &arg.Addr)
			}
		}

		if len(args) != len(fmt) {
			return nil, false
		}

		ok := true
		for i, arg := range args {
			ok = validateArg(arg, fmt[i])
			if !ok {
				break
			}
		}
		if !ok {
			continue
		}

		return args, true
	}

	return nil, false
}

func validateArg(arg *obj.Addr, format int) bool {
	switch getType(format) {
	case REG_R, REG_V, REG_Z, REG_P:
		if arg.Type != obj.TYPE_REG {
			return false
		}

		if !IsSVERegister(arg.Reg) {
			r := NewSVERegister(arg.Reg, EXT_NONE)
			arg.Reg = r.ToInt16()
		}

		reg := AsSVERegister(arg.Reg)
		if reg.Format() != format {
			return false
		}
	case REG_V_INDEXED, REG_Z_INDEXED, REG_P_INDEXED:
		if arg.Type != obj.TYPE_REGINDEX {
			return false
		}

		reg := AsSVERegister(arg.Reg)

		if int(reg.Ext()) != getSubtype(format) {
			return false
		}

		// Rebase type on REG_V to get base register type without index
		group := getType(format) - REG_V_INDEXED + REG_V

		if reg.Group() != group {
			return false
		}
	case MEM_ADDR:
		if arg.Type != obj.TYPE_ADDR {
			return false
		}

        addr := AsAddress(arg)

        if addr.Format() != format {
			return false
		}
	default:
		return false
	}

	return true
}
