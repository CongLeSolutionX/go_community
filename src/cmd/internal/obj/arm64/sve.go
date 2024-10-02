// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import (
	"cmd/internal/obj"
	"fmt"
)

var debugging = false

func Debug(format string, args ...interface{}) {
	if debugging {
		fmt.Println("(sve-debug) " + fmt.Sprintf(format, args...))
	}
}

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
	for i, r := range encoders[e.encoder].rules {
		selectedArgs := []*obj.Addr{}
		for _, sym := range r.symbolRefs {
			selectedArgs = append(selectedArgs, args[sym])
		}
		enc, ok := r.encode(selectedArgs...)
		if !ok {
			Debug("failed to encode '%v'", p)
			return 0, false
		}
		Debug("rule %d output = %08x | %032b", i, enc, enc)
		result |= enc
	}

	return result, true
}

func assembleSVE(p *obj.Prog) (uint32, error) {
	Debug("assemble: %s", p)
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

func isSveStore2(op obj.As) bool {
	switch op {
	case APSTR, AZSTR:
		return true
	}
	return false
}

func isSveStore3(op obj.As) bool {
	switch op {
	case AZST2B, AZST3B, AZST4B, AZST2D, AZST3D, AZST4D,
		AZST2H, AZST3H, AZST4H, AZST2W, AZST3W, AZST4W,
		AZST1B, AZST1H, AZST1W, AZST1D,
		AZSTNT1B, AZSTNT1D, AZSTNT1H, AZSTNT1W:
		return true
	}
	return false
}

// Disassembles operands from the Prog structure. This reverses the process
// performed by ARM64AsmInstruction, see cmd/asm/internal/arch/arm64.go.
func (e *encoding) disassembleProg(prog *obj.Prog) []*obj.Addr {
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

	if isSveStore2(prog.As) {
		// Swap source and destination here to produce the Arm operand order.
		args[0], args[1] = args[1], args[0]
	} else if isSveStore3(prog.As) {
		// When the vector stores are assembled, we'll end up with
		// From: <reglist> From2: Pg To: <addr>
		// which won't produce the order we want in Arm syntax using the
		// unpacking routine above. Arm operand order will always have the
		// transfer register list on the left, regardless of whether it is a
		// source or a destination.
		// We'll have [<addr>, <reglist>, <Pg>], so we just need two swaps
		// to get to [<reglist>, <Pg>, <addr>].
		args[0], args[1] = args[1], args[0]
		args[1], args[2] = args[2], args[1]
	}

	return args
}

// Unpack the Prog and validate against the format of this encoding instance, if there
// is a match it will return true as the second return value, otherwise false.
func (e *encoding) getFormattedArgs(prog *obj.Prog) ([]*obj.Addr, bool) {
	for _, format := range e.format {
		fmt := formats[format]

		args := e.disassembleProg(prog)

		if len(args) != len(fmt) {
			return nil, false
		}

		ok := true
		for i, arg := range args {
			ok = validateArg(arg, fmt[i])
			if !ok {
				Debug("mismatch on arg %d, need %v, addr: %v format: %v", i, fmt[i], arg, fmt)
				break
			}
		}
		if !ok {
			continue
		}

		Debug("matched: %v", fmt)
		return args, true
	}

	return nil, false
}

func validateArg(arg *obj.Addr, format int) bool {
	switch getType(format) {
	case REG_R, REG_F, REG_V, REG_Z, REG_P:
		if arg.Type != obj.TYPE_REG {
			return false
		}

		if !IsSVECompatibleRegister(arg.Reg) {
			r := NewRegister(arg.Reg, EXT_NONE)
			arg.Reg = r.ToInt16()
		}

		reg := AsRegister(arg.Reg)
		if reg.Format() != format {
			Debug("want register format %d, got %d", format, reg.Format())
			return false
		}
	case REG_V_INDEXED, REG_Z_INDEXED, REG_P_INDEXED:
		if arg.Type != obj.TYPE_REGINDEX {
			return false
		}

		reg := AsRegister(arg.Reg)

		if int(reg.Ext()) != getSubtype(format) {
			return false
		}

		// Rebase type on REG_V to get base register type without index
		group := getType(format) - REG_V_INDEXED + REG_V

		if reg.Group() != group {
			Debug("want indexed register %d, got %d", group, reg.Group())
			return false
		}
	case MEM_ADDR:
		if arg.Type != obj.TYPE_MEM {
			return false
		}

		addr := AsAddress(arg)

		if addr.Format() != format {
			Debug("want address format %d, got %d", format, addr.Format())
			return false
		}
		return true
	case IMM:
		immFormat := IMM
		if arg.Type == obj.TYPE_CONST {
			immFormat |= IMM_INT
		} else if arg.Type == obj.TYPE_FCONST {
			immFormat |= IMM_FLOAT
		} else {
			return false
		}
		return immFormat == format
	case PRFOP:
		if arg.Type == obj.TYPE_SPECIAL {
			switch SpecialOperand(arg.Offset) {
			case SPOP_PLDL1KEEP, SPOP_PLDL1STRM,
				SPOP_PLDL2KEEP, SPOP_PLDL2STRM,
				SPOP_PLDL3KEEP, SPOP_PLDL3STRM,
				SPOP_PSTL1KEEP, SPOP_PSTL1STRM,
				SPOP_PSTL2KEEP, SPOP_PSTL2STRM,
				SPOP_PSTL3KEEP, SPOP_PSTL3STRM:
				return true
			default:
				return false
			}
		} else if arg.Type == obj.TYPE_CONST &&
			// Prefetch operands allow 4-bit immediates where
			// bits 1 and 2 are set.
			arg.Offset > 0 && arg.Offset < 16 &&
			(arg.Offset&0b0110) == 0b0110 {
			return true
		}
	case PATTERN:
		op := SpecialOperand(arg.Offset)
		if arg.Type == obj.TYPE_SPECIAL && SPOP_POW2 <= op && op <= SPOP_ALL {
			return true
		} else if arg.Type == obj.TYPE_CONST {
			// Let the encoding rule validate the actual value
			return true
		}
		return false
	case REGLIST_R, REGLIST_F, REGLIST_V, REGLIST_Z, REGLIST_P:
		if arg.Type != obj.TYPE_REGLIST {
			return false
		}
		rl := ARM64RegisterList{uint64(arg.Offset)}
		return rl.Format() == format
	default:
		panic(fmt.Sprintf("invalid argument type '%v'", format))
	}

	return true
}
