// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"bytes"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"encoding/binary"
	"fmt"
	"math"
)

var Register = []string{
	"PC_F",
	"PC_B",
	"SP",
	"CTX",
	"g",
	"RET0",
	"RET1",
	"RET2",
	"RET3",

	"I0",
	"I1",
	"I2",
	"I3",
	"I4",
	"I5",
	"I6",
	"I7",
	"I8",
	"I9",
	"I10",
	"I11",
	"I12",
	"I13",
	"I14",
	"I15",

	"F0",
	"F1",
	"F2",
	"F3",
	"F4",
	"F5",
	"F6",
	"F7",
	"F8",
	"F9",
	"F10",
	"F11",
	"F12",
	"F13",
	"F14",
	"F15",

	"MAXREG",
}

func init() {
	obj.RegisterRegister(REG_PC_F, REG_PC_F+len(Register), rconv)
	obj.RegisterOpcode(obj.ABaseWASM, Anames)
}

func rconv(r int) string {
	if REG_PC_F <= r && r-REG_PC_F < len(Register) {
		return Register[r-REG_PC_F]
	}
	return fmt.Sprintf("Rgok(%d)", r-obj.RBaseWASM)
}

var unaryDst = map[obj.As]bool{
	ASet:          true,
	ATee:          true,
	ACall:         true,
	ACallIndirect: true,
	ACallImport:   true,
	ABr:           true,
	ABrIf:         true,
	AI32Store:     true,
	AI64Store:     true,
	AF32Store:     true,
	AF64Store:     true,
	AI32Store8:    true,
	AI32Store16:   true,
	AI64Store8:    true,
	AI64Store16:   true,
	AI64Store32:   true,
}

var Linkwasm = obj.LinkArch{
	Arch:       sys.ArchWASM,
	Init:       instinit,
	Preprocess: preprocess,
	Assemble:   assemble,
	UnaryDst:   unaryDst,
}

var wasmzero *obj.LSym
var morestack *obj.LSym
var morestackNoCtxt *obj.LSym
var sigpanic *obj.LSym
var jmpdefer *obj.LSym

const (
	/* mark flags */
	DoesNotBlock = 1 << 0
	WasmImport   = 1 << 1
)

func instinit(ctxt *obj.Link) {
	wasmzero = ctxt.Lookup("runtime.wasmzero")
	morestack = ctxt.Lookup("runtime.morestack")
	morestackNoCtxt = ctxt.Lookup("runtime.morestack_noctxt")
	sigpanic = ctxt.Lookup("runtime.sigpanic")
	jmpdefer = ctxt.Lookup("runtime.jmpdefer")
}

func preprocess(ctxt *obj.Link, s *obj.LSym, newprog obj.ProgAlloc) {
	appendp := func(p *obj.Prog, as obj.As) *obj.Prog {
		p2 := obj.Appendp(p, newprog)
		p2.As = as
		p2.Pc = p.Pc
		return p2
	}

	appendpConst := func(p *obj.Prog, as obj.As, value int64) *obj.Prog {
		p = appendp(p, as)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: value}
		return p
	}

	framesize := s.Func.Text.To.Offset
	if framesize < 0 {
		panic("unexpected")
	}
	s.Func.Args = s.Func.Text.To.Val.(int32)
	s.Func.Locals = int32(framesize)

	bpsize := 0
	pcsize := 8

	if s.Func.Text.From.Sym.Wrapper() {
		// if g._panic != nil && g._panic.argp == FP {
		//   g._panic.argp = bottom-of-frame
		// }
		//
		//   I64Load g_panic(g)
		// Set I0
		//       Get I0
		//     I64Eqz
		//   I32Eqz
		// If
		//           Get SP
		//         I64ExtendUI32
		//         I64Const $framesize+8
		//       I64Add
		//       I64Load panic_argp(I0)
		//     I64Eq
		//   If
		//         Get I0
		//       I32WrapI64
		//         Get SP
		//       I64ExtendUI32
		//     I64Store $panic_argp
		//   End
		// End

		gpanic := obj.Addr{
			Type:   obj.TYPE_MEM,
			Reg:    REGG,
			Offset: 4 * 8, // g_panic
		}

		panicargp := obj.Addr{
			Type:   obj.TYPE_MEM,
			Reg:    REG_I0,
			Offset: 0, // panic.argp
		}

		p := s.Func.Text

		p = appendp(p, AI64Load)
		p.From = gpanic
		p = appendp(p, ASet)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_I0}

		p = appendp(p, AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_I0}
		p = appendp(p, AI64Eqz)
		p = appendp(p, AI32Eqz)
		p = appendp(p, AIf)

		p = appendp(p, AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		p = appendp(p, AI64ExtendUI32)
		p = appendpConst(p, AI64Const, framesize+8)
		p = appendp(p, AI64Add)
		p = appendp(p, AI64Load)
		p.From = panicargp

		p = appendp(p, AI64Eq)
		p = appendp(p, AIf)

		p = appendp(p, AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_I0}
		p = appendp(p, AI32WrapI64)
		p = appendp(p, AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		p = appendp(p, AI64ExtendUI32)
		p = appendp(p, AI64Store)
		p.To = panicargp

		p = appendp(p, AEnd)
		p = appendp(p, AEnd)
	}

	if framesize > 0 {
		p := s.Func.Text

		p = appendp(p, AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

		p = appendpConst(p, AI32Const, framesize)

		p = appendp(p, AI32Sub)

		p = appendp(p, ASet)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		p.Spadj = int32(framesize)
	}

	if !s.Func.Text.From.Sym.NoSplit() {
		p := s.Func.Text

		p = appendp(p, ABlock)

		if framesize <= objabi.StackSmall {
			// small stack: SP <= stackguard
			//   Get SP
			//       Get g
			//     I32WrapI64
			//   I32Load $stackguard0
			// I32GtU

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REGG}

			p = appendp(p, AI32WrapI64)

			p = appendpConst(p, AI32Load, 2*int64(ctxt.Arch.PtrSize)) // G.stackguard0

			p = appendp(p, AI32GtU)
		} else {
			// large stack: SP-framesize <= stackguard-StackSmall
			//              SP <= stackguard+(framesize-StackSmall)
			//   Get SP
			//         Get g
			//       I32WrapI64
			//     I32Load $stackguard0
			//     I32Const $(framesize-StackSmall)
			//   I32Add
			// I32GtU

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REGG}

			p = appendp(p, AI32WrapI64)

			p = appendpConst(p, AI32Load, 2*int64(ctxt.Arch.PtrSize)) // G.stackguard0

			p = appendpConst(p, AI32Const, int64(framesize)-objabi.StackSmall)

			p = appendp(p, AI32Add)

			p = appendp(p, AI32GtU)
		}
		// TODO(neelance): handle wraparound case

		p = appendpConst(p, ABrIf, 0)

		p = appendp(p, obj.ACALL)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
		if s.Func.Text.From.Sym.NeedCtxt() {
			p.To = obj.Addr{Type: obj.TYPE_MEM, Sym: morestack}
		} else {
			p.To = obj.Addr{Type: obj.TYPE_MEM, Sym: morestackNoCtxt}
		}

		p = appendp(p, AEnd)
	}

	numEntryBlocks := 0
	additionalBlockDepth := 0
	pc := int64(0)
	var tableIdxs []uint64
	tablePC := int64(0)
	incAllPC := false // TODO remove
	for p := s.Func.Text; p != nil; p = p.Link {
		p.Pc = pc
		if p.Spadj != 0 || incAllPC {
			pc++
		}
		switch p.As {
		case ABlock, ALoop, AIf:
			additionalBlockDepth++
		case AEnd:
			if additionalBlockDepth > 0 {
				additionalBlockDepth--
				break
			}
			for tablePC <= pc {
				tableIdxs = append(tableIdxs, uint64(numEntryBlocks))
				tablePC++
			}
			numEntryBlocks++
			pc++
		}
	}
	s.Size = pc + 1

	if numEntryBlocks > 0 { // not assembly function
		p := s.Func.Text

		p = appendp(p, ABlock)

		p = appendp(p, ALoop)

		if numEntryBlocks > 1 {
			for i := 0; i < numEntryBlocks; i++ {
				p = appendp(p, ABlock)
			}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_B}

			p = appendp(p, ABrTable)
			p.To = obj.Addr{Val: tableIdxs}

			p = appendp(p, AEnd)
		}

		for p.Link != nil {
			p = p.Link
		}

		p = appendp(p, AEnd)

		p = appendp(p, obj.ARET)
	}

	p := s.Func.Text
	blockDepth := 0
	for p != nil {
		switch p.As {
		case ABlock, ALoop, AIf:
			blockDepth++
		case AEnd:
			blockDepth--
		}

		switch p.From.Name {
		case obj.NAME_AUTO:
			p.From.Offset += int64(framesize) - int64(bpsize)
		case obj.NAME_PARAM:
			p.From.Reg = REG_SP
			p.From.Offset += int64(framesize) + int64(pcsize)
		}

		switch p.To.Name {
		case obj.NAME_AUTO:
			p.To.Offset += int64(framesize) - int64(bpsize)
		case obj.NAME_PARAM:
			p.From.Reg = REG_SP
			p.To.Offset += int64(framesize) + int64(pcsize)
		}

		switch p.As {
		case AAddrOf:
			addrOf := p

			switch p.From.Name {
			case obj.NAME_AUTO, obj.NAME_PARAM:
				if p.From.Reg != REG_SP {
					panic("unexpected")
				}

				p = appendp(p, AGet)
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

				p = appendp(p, AI64ExtendUI32)

				p = appendpConst(p, AI64Const, addrOf.From.Offset)

				p = appendp(p, AI64Add)

			case obj.NAME_EXTERN:
				p = appendp(p, AI64Const)
				p.From = obj.Addr{Type: obj.TYPE_MEM, Sym: addrOf.From.Sym}

			default:
				panic("unexpected")
			}

		case obj.AJMP, obj.ACALL:
			jmp := p

			if jmp.As == obj.AJMP && jmp.To.Type == obj.TYPE_BRANCH {
				p = appendpConst(p, AI32Const, jmp.To.Val.(*obj.Prog).Pc)

				p = appendp(p, ASet)
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_B}

				p = appendp(p, ABr)
				p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(blockDepth - 2)}
				break
			}

			switch jmp.As {
			case obj.AJMP:
				p = appendpConst(p, AI32Const, 0)
				p = appendp(p, ASet)
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_B}

				if jmp.To.Type == obj.TYPE_MEM {
					p = appendp(p, ACall)
					p.To = obj.Addr{Type: obj.TYPE_MEM, Sym: jmp.To.Sym}
				} else {
					p = appendp(p, AI32WrapI64)
					p = appendpConst(p, AI32Const, 16)
					p = appendp(p, AI32ShrU)
					p = appendp(p, ACallIndirect)

					// TODO it should be possible to remove this by putting a loop around deferreturn
					if s == jmpdefer {
						p = appendp(p, ADrop)
						p = appendpConst(p, AI32Const, 1)
					}
				}

				p = appendp(p, ABr)
				p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(blockDepth)}

			case obj.ACALL:
				if jmp.Link.As != AEnd && jmp.Mark&DoesNotBlock == 0 {
					panic("end expected")
				}

				pcAfterCall := p.Link.Link.Pc
				if jmp.To.Sym == sigpanic {
					pcAfterCall-- // sigpanic expects to be called without advancing the pc
				}

				p = appendp(p, AGet)
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				p = appendpConst(p, AI32Const, 8)
				p = appendp(p, AI32Sub)
				p = appendp(p, ASet)
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

				p = appendp(p, AGet)
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				p = appendpConst(p, AI64Const, pcAfterCall)
				p = appendp(p, AI64Store)
				p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}

				p = appendp(p, AGet)
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				p = appendp(p, AI32Const)
				p.From = obj.Addr{Type: obj.TYPE_BRANCH, Sym: s}
				p = appendp(p, AI32Store16)
				p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: 2}

				p = appendpConst(p, AI32Const, 0)
				p = appendp(p, ASet)
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_B}

				if jmp.To.Type == obj.TYPE_MEM {
					p = appendp(p, ACall)
					p.To = obj.Addr{Type: obj.TYPE_MEM, Sym: jmp.To.Sym}
				} else {
					p = appendp(p, AI32WrapI64)
					p = appendpConst(p, AI32Const, 16)
					p = appendp(p, AI32ShrU)
					p = appendp(p, ACallIndirect)
				}

				p = appendp(p, AIf)
				if jmp.Mark&DoesNotBlock != 0 {
					p = appendp(p, AUnreachable)
				} else {
					p = appendpConst(p, AI32Const, 1)
					p = appendp(p, ABr)
					p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(blockDepth + 1)}
				}
				p = appendp(p, AEnd)
			}

		case obj.ARET:
			if framesize > 0 {
				p = appendp(p, AGet)
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

				p = appendpConst(p, AI32Const, framesize)

				p = appendp(p, AI32Add)

				p = appendp(p, ASet)
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				// p.Spadj = int32(-framesize)
			}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
			p = appendpConst(p, AI32Load16U, 0)
			p = appendp(p, ASet)
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_B}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
			p = appendpConst(p, AI32Load16U, 2)
			p = appendp(p, ASet)
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_PC_F}

			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
			p = appendpConst(p, AI32Const, 8)
			p = appendp(p, AI32Add)
			p = appendp(p, ASet)
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

			p = appendpConst(p, AI32Const, 0)

		case ABr:
			if p.To.Offset == -2 {
				p.To.Offset = int64(blockDepth - 1)
			}

		case AI32Load, AI64Load, AF32Load, AF64Load, AI32Load8S, AI32Load8U, AI32Load16S, AI32Load16U, AI64Load8S, AI64Load8U, AI64Load16S, AI64Load16U, AI64Load32S, AI64Load32U:
			if p.From.Type == obj.TYPE_MEM {
				as := p.As
				from := p.From

				p.As = AGet
				p.From = obj.Addr{Type: obj.TYPE_REG, Reg: from.Reg}

				if from.Reg != REG_SP {
					p = appendp(p, AI32WrapI64)
				}

				p = appendpConst(p, as, from.Offset)
			}

		case ACallImport:
			p = appendp(p, AGet)
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}

			p = appendp(p, ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: s}
			p.Mark = WasmImport
		}

		p = p.Link
	}
}

var globals = map[int16]uint64{
	REG_PC_F: 0,
	REG_PC_B: 1,
	REG_SP:   2,
	REG_CTX:  3,
	REGG:     4,
	REG_RET0: 5,
	REG_RET1: 6,
	REG_RET2: 7,
	REG_RET3: 8,
}

func assemble(ctxt *obj.Link, s *obj.LSym, newprog obj.ProgAlloc) {
	if s.P != nil {
		return
	}

	var localRegs = []int16{
		REG_I0, REG_I1, REG_I2, REG_I3, REG_I4, REG_I5, REG_I6, REG_I7, REG_I8, REG_I9, REG_I10, REG_I11, REG_I12, REG_I13, REG_I14, REG_I15,
		REG_F0, REG_F1, REG_F2, REG_F3, REG_F4, REG_F5, REG_F6, REG_F7, REG_F8, REG_F9, REG_F10, REG_F11, REG_F12, REG_F13, REG_F14, REG_F15,
	}
	var locals = make(map[int16]uint64)
	for i, r := range localRegs {
		locals[r] = uint64(i)
	}

	w := new(bytes.Buffer)
	switch s.Name {
	case "runtime.memchr":
		writeUleb128(w, 1) // number of sets of locals
		writeUleb128(w, 3) // number of locals
		w.WriteByte(0x7F)  // i32
	case "runtime.memcmp":
		writeUleb128(w, 1) // number of sets of locals
		writeUleb128(w, 2) // number of locals
		w.WriteByte(0x7F)  // i32
	default:
		writeUleb128(w, 2)  // number of sets of locals
		writeUleb128(w, 16) // number of locals
		w.WriteByte(0x7E)   // i64
		writeUleb128(w, 16) // number of locals
		w.WriteByte(0x7C)   // f64
	}

	var r []obj.Reloc
	for p := s.Func.Text; p != nil; p = p.Link {
		switch p.As {
		case AGet:
			switch p.From.Type {
			case obj.TYPE_CONST:
				w.WriteByte(0x20) // get_local
				writeUleb128(w, uint64(p.From.Offset))
				continue
			case obj.TYPE_REG:
				if idx, ok := locals[p.From.Reg]; ok {
					w.WriteByte(0x20) // get_local
					writeUleb128(w, idx)
					continue
				}
				if idx, ok := globals[p.From.Reg]; ok {
					w.WriteByte(0x23) // get_global
					writeUleb128(w, idx)
					continue
				}
				panic("bad reg")
			}
			panic("bad type")

		case ASet:
			switch p.To.Type {
			case obj.TYPE_CONST:
				w.WriteByte(0x21) // set_local
				writeUleb128(w, uint64(p.To.Offset))
				continue
			case obj.TYPE_REG:
				if idx, ok := locals[p.To.Reg]; ok {
					if p.Link.As == AGet && p.Link.From.Reg == p.To.Reg {
						w.WriteByte(0x22) // tee_local
						p = p.Link
					} else {
						w.WriteByte(0x21) // set_local
					}
					writeUleb128(w, idx)
					continue
				}
				if idx, ok := globals[p.To.Reg]; ok {
					w.WriteByte(0x24) // set_global
					writeUleb128(w, idx)
					continue
				}
				panic("bad reg")
			}
			panic("bad type")

		case ATee:
			switch p.To.Type {
			case obj.TYPE_CONST:
				w.WriteByte(0x22) // tee_local
				writeUleb128(w, uint64(p.To.Offset))
				continue
			case obj.TYPE_REG:
				if idx, ok := locals[p.To.Reg]; ok {
					w.WriteByte(0x22) // tee_local
					writeUleb128(w, idx)
					continue
				}
				panic("bad reg")
			}
			panic("bad type")

		case obj.ATEXT, AAddrOf, ACallImport, obj.AJMP, obj.ACALL, obj.ARET, obj.AFUNCDATA, obj.APCDATA, AWORD:
			// ignore
			continue
		}

		switch {
		case p.As < AUnreachable:
			panic(fmt.Sprintf("unexpected assembler op: %s", p.As))
		case p.As < AEnd:
			w.WriteByte(byte(p.As - AUnreachable + 0x00))
		case p.As < ADrop:
			w.WriteByte(byte(p.As - AEnd + 0x0B))
		case p.As < AI32Load:
			w.WriteByte(byte(p.As - ADrop + 0x1A))
		default:
			w.WriteByte(byte(p.As - AI32Load + 0x28))
		}

		switch p.As {
		case ABlock, ALoop, AIf:
			if p.From.Offset != 0 {
				// block type, rarely used, e.g. for code compiled with emscripten
				w.WriteByte(0x80 - byte(p.From.Offset))
				continue
			}
			w.WriteByte(0x40)

		case ABr, ABrIf:
			writeUleb128(w, uint64(p.To.Offset))

		case ABrTable:
			idxs := p.To.Val.([]uint64)
			writeUleb128(w, uint64(len(idxs)-1))
			for _, idx := range idxs {
				writeUleb128(w, idx)
			}

		case ACall:
			switch p.To.Type {
			case obj.TYPE_CONST:
				writeUleb128(w, uint64(p.To.Offset))

			case obj.TYPE_MEM:
				typ := objabi.R_CALL
				if p.Mark&WasmImport != 0 {
					typ = objabi.R_WASMIMPORT
				}
				r = append(r, obj.Reloc{
					Off:  int32(w.Len()),
					Type: typ,
					Sym:  p.To.Sym,
				})

			default:
				panic("unexpected")
			}

		case ACallIndirect:
			writeUleb128(w, uint64(p.To.Offset))
			w.WriteByte(0x00) // reserved value

		case AI32Const, AI64Const:
			switch p.From.Type {
			case obj.TYPE_CONST, obj.TYPE_ADDR:
				writeSleb128(w, p.From.Offset)

			case obj.TYPE_MEM:
				r = append(r, obj.Reloc{
					Off:  int32(w.Len()),
					Type: objabi.R_ADDR,
					Sym:  p.From.Sym,
				})

			case obj.TYPE_BRANCH:
				r = append(r, obj.Reloc{
					Off:  int32(w.Len()),
					Type: objabi.R_CALL,
					Sym:  p.From.Sym,
				})

			default:
				panic("unexpected")
			}

		case AF64Const:
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, math.Float64bits(p.From.Val.(float64)))
			w.Write(b)

		case AI32Load, AI64Load, AF32Load, AF64Load, AI32Load8S, AI32Load8U, AI32Load16S, AI32Load16U, AI64Load8S, AI64Load8U, AI64Load16S, AI64Load16U, AI64Load32S, AI64Load32U:
			if p.From.Offset < 0 {
				panic("negative offset")
			}
			if p.From.Type != obj.TYPE_CONST {
				panic("unexpected")
			}
			writeUleb128(w, align(p.As))
			writeUleb128(w, uint64(p.From.Offset))

		case AI32Store, AI64Store, AF32Store, AF64Store, AI32Store8, AI32Store16, AI64Store8, AI64Store16, AI64Store32:
			if p.To.Offset < 0 {
				panic("negative offset")
			}
			writeUleb128(w, align(p.As))
			writeUleb128(w, uint64(p.To.Offset))

		case ACurrentMemory, AGrowMemory:
			w.WriteByte(0x00)

		}
	}

	s.P = w.Bytes()
	s.R = r
}

func align(as obj.As) uint64 {
	switch as {
	case AI32Load8S, AI32Load8U, AI64Load8S, AI64Load8U, AI32Store8, AI64Store8:
		return 0
	case AI32Load16S, AI32Load16U, AI64Load16S, AI64Load16U, AI32Store16, AI64Store16:
		return 1
	case AI32Load, AF32Load, AI64Load32S, AI64Load32U, AI32Store, AF32Store, AI64Store32:
		return 2
	case AI64Load, AF64Load, AI64Store, AF64Store:
		return 3
	default:
		panic("unexpected")
	}
}

func writeUleb128(w *bytes.Buffer, v uint64) {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		w.WriteByte(c)
		if c&0x80 == 0 {
			break
		}
	}
}

func writeSleb128(w *bytes.Buffer, v int64) {
	for {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		if (v != -1 || s == 0) && (v != 0 || s != 0) {
			c |= 0x80
		}
		w.WriteByte(c)
		if c&0x80 == 0 {
			break
		}
	}
}
