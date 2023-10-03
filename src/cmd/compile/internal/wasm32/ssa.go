// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm32

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/objw"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/wasm32"
	"fmt"
	"internal/buildcfg"
	"os"
)

/*

   Wasm implementation
   -------------------

   Wasm is a strange Go port because the machine isn't
   a register-based machine, threads are different, code paths
   are different, etc. We outline those differences here.

   See the design doc for some additional info on this topic.
   https://docs.google.com/document/d/131vjr4DH6JFnb-blm_uRdaC0_Nv3OUwjEY5qVCxCup4/edit#heading=h.mjo1bish3xni

   PCs:

   Wasm doesn't have PCs in the normal sense that you can jump
   to or call to. Instead, we simulate these PCs using our own construct.

   A PC in the Wasm implementation is the combination of a function
   ID and a block ID within that function. The function ID is an index
   into a function table which transfers control to the start of the
   function in question, and the block ID is a sequential integer
   indicating where in the function we are.

   Every function starts with a branch table which transfers control
   to the place in the function indicated by the block ID. The block
   ID is provided to the function as the sole Wasm argument.

   Block IDs do not encode every possible PC. They only encode places
   in the function where it might be suspended. Typically these places
   are call sites.

   Sometimes we encode the function ID and block ID separately. When
   recorded together as a single integer, we use the value F<<16+B.

   Threads:

   Wasm doesn't (yet) have threads. We have to simulate threads by
   keeping goroutine stacks in linear memory and unwinding
   the Wasm stack each time we want to switch goroutines.

   To support unwinding a stack, each function call returns on the Wasm
   stack a boolean that tells the function whether it should return
   immediately or not. When returning immediately, a return address
   is left on the top of the Go stack indicating where the goroutine
   should be resumed.

   Stack pointer:

   There is a single global stack pointer which records the stack pointer
   used by the currently active goroutine. This is just an address in
   linear memory where the Go runtime is maintaining the stack for that
   goroutine.

   Functions cache the global stack pointer in a local variable for
   faster access, but any changes must be spilled to the global variable
   before any call and restored from the global variable after any call.

   Calling convention:

   All Go arguments and return values are passed on the Go stack, not
   the wasm stack. In addition, return addresses are pushed on the
   Go stack at every call point. Return addresses are not used during
   normal execution, they are used only when resuming goroutines.
   (So they are not really a "return address", they are a "resume address".)

   All Go functions have the Wasm type (i32)->i32. The argument
   is the block ID and the return value is the exit immediately flag.

   Callsite:
    - write arguments to the Go stack (starting at SP+0)
    - push return address to Go stack (4 bytes)
    - write local SP to global SP
    - push 0 (type i32) to Wasm stack
    - issue Call
    - restore local SP from global SP
    - pop int32 from top of Wasm stack. If nonzero, exit function immediately.
    - use results from Go stack (starting at SP+sizeof(args))
       - note that the callee will have popped the return address

   Prologue:
    - initialize local SP from global SP
    - jump to the location indicated by the block ID argument
      (which appears in local variable 0)
    - at block 0
      - check for Go stack overflow, call morestack if needed
      - subtract frame size from SP
      - note that arguments now start at SP+framesize+4

   Normal epilogue:
    - pop frame from Go stack
    - pop return address from Go stack
    - push 0 (type i32) on the Wasm stack
    - return
   Exit immediately epilogue:
    - push 1 (type i32) on the Wasm stack
    - return
    - note that the return address and stack frame are left on the Go stack

   The main loop that executes goroutines is wasm_pc_f_loop, in
   runtime/rt0_js_wasm.s. It grabs the saved return address from
   the top of the Go stack (actually SP-4?), splits it up into F
   and B parts, then calls F with its Wasm argument set to B.

   Note that when resuming a goroutine, only the most recent function
   invocation of that goroutine appears on the Wasm stack. When that
   Wasm function returns normally, the next most recent frame will
   then be started up by wasm_pc_f_loop.

   Global 0 is SP (stack pointer)
   Global 1 is CTXT (closure pointer)
   Global 2 is GP (goroutine pointer)
*/

func Init(arch *ssagen.ArchInfo) {
	arch.LinkArch = &wasm32.Linkwasm
	arch.REGSP = wasm32.REG_SP
	arch.MAXWIDTH = 1 << 50

	arch.ZeroRange = zeroRange
	arch.Ginsnop = ginsnop

	arch.SSAMarkMoves = ssaMarkMoves
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}

func zeroRange(pp *objw.Progs, p *obj.Prog, off, cnt int64, state *uint32) *obj.Prog {
	if cnt == 0 {
		return p
	}
	if cnt%4 != 0 {
		base.Fatalf("zerorange count not a multiple of widthptr %d", cnt)
	}

	for i := int64(0); i < cnt; i += 8 {
		p = pp.Append(p, wasm32.AGet, obj.TYPE_REG, wasm32.REG_SP, 0, 0, 0, 0)
		p = pp.Append(p, wasm32.AI32Const, obj.TYPE_CONST, 0, 0, 0, 0, 0)
		p = pp.Append(p, wasm32.AI32Store, 0, 0, 0, obj.TYPE_CONST, 0, off+i)
	}

	return p
}

func ginsnop(pp *objw.Progs) *obj.Prog {
	return pp.Prog(wasm32.ANop)
}

func ssaMarkMoves(s *ssagen.State, b *ssa.Block) {
}

func ssaGenBlock(s *ssagen.State, b, next *ssa.Block) {
	switch b.Kind {
	case ssa.BlockPlain:
		if next != b.Succs[0].Block() {
			s.Br(obj.AJMP, b.Succs[0].Block())
		}

	case ssa.BlockIf:
		switch next {
		case b.Succs[0].Block():
			// if false, jump to b.Succs[1]
			getValue32(s, b.Controls[0])
			s.Prog(wasm32.AI32Eqz)
			s.Prog(wasm32.AIf)
			s.Br(obj.AJMP, b.Succs[1].Block())
			s.Prog(wasm32.AEnd)
		case b.Succs[1].Block():
			// if true, jump to b.Succs[0]
			getValue32(s, b.Controls[0])
			s.Prog(wasm32.AIf)
			s.Br(obj.AJMP, b.Succs[0].Block())
			s.Prog(wasm32.AEnd)
		default:
			// if true, jump to b.Succs[0], else jump to b.Succs[1]
			getValue32(s, b.Controls[0])
			s.Prog(wasm32.AIf)
			s.Br(obj.AJMP, b.Succs[0].Block())
			s.Prog(wasm32.AEnd)
			s.Br(obj.AJMP, b.Succs[1].Block())
		}

	case ssa.BlockRet:
		s.Prog(obj.ARET)

	case ssa.BlockExit, ssa.BlockRetJmp:

	case ssa.BlockDefer:
		p := s.Prog(wasm32.AGet)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: wasm32.REG_RET0}
		s.Prog(wasm32.AI32Eqz)
		s.Prog(wasm32.AI32Eqz)
		s.Prog(wasm32.AIf)
		s.Br(obj.AJMP, b.Succs[1].Block())
		s.Prog(wasm32.AEnd)
		if next != b.Succs[0].Block() {
			s.Br(obj.AJMP, b.Succs[0].Block())
		}

	default:
		panic("unexpected block")
	}

	// Entry point for the next block. Used by the JMP in goToBlock.
	s.Prog(wasm32.ARESUMEPOINT)

	if s.OnWasmStackSkipped != 0 {
		panic(fmt.Sprintf("wasm: bad stack: %d", s.OnWasmStackSkipped))
	}
}

func ssaGenValue(s *ssagen.State, v *ssa.Value) {
	switch v.Op {
	case ssa.OpWasm32LoweredStaticCall, ssa.OpWasm32LoweredClosureCall, ssa.OpWasm32LoweredInterCall, ssa.OpWasm32LoweredTailCall:
		s.PrepareCall(v)
		if call, ok := v.Aux.(*ssa.AuxCall); ok && call.Fn == ir.Syms.Deferreturn {
			// The runtime needs to inject jumps to
			// deferreturn calls using the address in
			// _func.deferreturn. Hence, the call to
			// deferreturn must itself be a resumption
			// point so it gets a target PC.
			s.Prog(wasm32.ARESUMEPOINT)
		}
		if v.Op == ssa.OpWasm32LoweredClosureCall {
			getValue32(s, v.Args[1])
			setReg(s, wasm32.REG_CTXT)
		}
		if call, ok := v.Aux.(*ssa.AuxCall); ok && call.Fn != nil {
			sym := call.Fn
			p := s.Prog(obj.ACALL)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: sym}
			p.Pos = v.Pos
			if v.Op == ssa.OpWasm32LoweredTailCall {
				p.As = obj.ARET
			}
		} else {
			getValue32(s, v.Args[0])
			p := s.Prog(obj.ACALL)
			p.To = obj.Addr{Type: obj.TYPE_NONE}
			p.Pos = v.Pos
		}

	case ssa.OpWasm32LoweredMove:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		i32Const(s, int32(v.AuxInt))
		s.Prog(wasm32.AMemoryCopy)

	case ssa.OpWasm32LoweredZero:
		getValue32(s, v.Args[0])
		i32Const(s, 0)
		i32Const(s, int32(v.AuxInt))
		s.Prog(wasm32.AMemoryFill)

	case ssa.OpWasm32LoweredNilCheck:
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI32Eqz)
		s.Prog(wasm32.AIf)
		p := s.Prog(wasm32.ACALLNORESUME)
		p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.SigPanic}
		s.Prog(wasm32.AEnd)
		if logopt.Enabled() {
			logopt.LogOpt(v.Pos, "nilcheck", "genssa", v.Block.Func.Name)
		}
		if base.Debug.Nil != 0 && v.Pos.Line() > 1 { // v.Pos.Line()==1 in generated wrappers
			base.WarnfAt(v.Pos, "generated nil check")
		}
	case ssa.OpWasm32LoweredPanicExtendA, ssa.OpWasm32LoweredPanicExtendB, ssa.OpWasm32LoweredPanicExtendC:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		getValue32(s, v.Args[2])
		p := s.Prog(obj.ACALL)
		p.To.Type = obj.TYPE_MEM
		p.To.Name = obj.NAME_EXTERN
		p.To.Sym = ssagen.ExtendCheckFunc[v.AuxInt]
		s.UseArgs(12) // space used in callee args area by assembly stubs

	case ssa.OpWasm32LoweredWB:
		p := s.Prog(wasm32.ACall)
		// AuxInt encodes how many buffer entries we need.
		p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.GCWriteBarrier[v.AuxInt-1]}
		setReg(s, v.Reg0()) // move result from wasm stack to register local

	case ssa.OpWasm32I64Store8, ssa.OpWasm32I64Store16, ssa.OpWasm32I64Store32, ssa.OpWasm32I64Store, ssa.OpWasm32F32Store, ssa.OpWasm32F64Store:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		p := s.Prog(v.Op.Asm())
		p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}

	case ssa.OpWasm32I32Store:
		//fmt.Printf("i32store: %s (%s)\n", v.Args[0].Op, v.Args[0])
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		p := s.Prog(v.Op.Asm())
		p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}
	case ssa.OpWasm32I32Store8, ssa.OpWasm32I32Store16:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		p := s.Prog(v.Op.Asm())
		p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}

	case ssa.OpStoreReg:
		getReg(s, wasm32.REG_SP)
		getValue32(s, v.Args[0])
		p := s.Prog(storeOp(v.Type))
		ssagen.AddrAuto(&p.To, v)

	case ssa.OpWasm32Mul64Decomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64Mul)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Div64Decomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64DivS)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Div64uDecomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64DivU)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Mod64Decomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64RemS)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Mod64uDecomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64RemU)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Add64Decomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64Add)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpWasm32Sub64Decomp:
		// compose x
		getValue32(s, v.Args[0])
		s.Prog(wasm32.AI64ExtendI32S)
		p := s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[1])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		// compose y
		getValue32(s, v.Args[2])
		s.Prog(wasm32.AI64ExtendI32S)
		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64Shl)

		getValue32(s, v.Args[3])
		s.Prog(wasm32.AI64ExtendI32U)
		s.Prog(wasm32.AI64Or)

		s.Prog(wasm32.AI64Sub)

		p = s.Prog(wasm32.ATee)
		p.To = obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  wasm32.REG_X0,
		}

		p = s.Prog(wasm32.AI64Const)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 32}
		s.Prog(wasm32.AI64ShrS)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg0())

		getReg(s, wasm32.REG_X0)
		s.Prog(wasm32.AI32WrapI64)
		setReg(s, v.Reg1())
	case ssa.OpClobber, ssa.OpClobberReg:
		// TODO: implement for clobberdead experiment. Nop is ok for now.

	default:
		if v.Type.IsMemory() {
			return
		}
		if v.OnWasmStack {
			s.OnWasmStackSkipped++
			/*
				if s.FuncInfo().Text.Ctxt.Pkgpath == "runtime/internal/sys" {
					fmt.Printf("s> %s (%d, %s) %d\n", v.Op, v.Op, v, s.OnWasmStackSkipped)
					if v.Op.String() == "LoweredAdd32carry" {
						panic("on ladd32 carry")
					}
				}
			*/
			// If a Value is marked OnWasmStack, we don't generate the value and store it to a register now.
			// Instead, we delay the generation to when the value is used and then directly generate it on the WebAssembly stack.
			return
		}
		set := ssaGenValueOnStack(s, v, true)
		if s.OnWasmStackSkipped != 0 {
			name := s.FuncInfo().Text.Ctxt.Pkgpath
			panic(fmt.Sprintf("wasm: bad stack: %d (%s)", s.OnWasmStackSkipped, name))
		}
		if set {
			if !v.HasReg() {
				fmt.Fprintln(os.Stderr, v.LongString())
				panic("bad news reg")
			}

			if v.Reg() == 16392 {
				panic(fmt.Sprintf("%s had bad reg", v.Op))
			}

			setReg(s, v.Reg())
		}
	}
}

func ssaGenValueOnStack(s *ssagen.State, v *ssa.Value, extend bool) bool {
	switch v.Op {
	case ssa.OpWasm32LoweredGetClosurePtr:
		getReg(s, wasm32.REG_CTXT)

	case ssa.OpWasm32LoweredGetCallerPC:
		p := s.Prog(wasm32.AI32Load)
		// Caller PC is stored 4 bytes below first parameter.
		p.From = obj.Addr{
			Type:   obj.TYPE_MEM,
			Name:   obj.NAME_PARAM,
			Offset: -4,
		}

	case ssa.OpWasm32LoweredGetCallerSP:
		p := s.Prog(wasm32.AGet)
		// Caller SP is the address of the first parameter.
		p.From = obj.Addr{
			Type:   obj.TYPE_ADDR,
			Name:   obj.NAME_PARAM,
			Reg:    wasm32.REG_SP,
			Offset: 0,
		}

	case ssa.OpWasm32LoweredAddr:
		if v.Aux == nil { // address of off(SP), no symbol
			getValue32(s, v.Args[0])
			i64Const(s, v.AuxInt)
			s.Prog(wasm32.AI32Add)
			break
		}
		p := s.Prog(wasm32.AGet)
		p.From.Type = obj.TYPE_ADDR
		switch v.Aux.(type) {
		case *obj.LSym:
			ssagen.AddAux(&p.From, v)
		case *ir.Name:
			p.From.Reg = v.Args[0].Reg()
			ssagen.AddAux(&p.From, v)
		default:
			panic("wasm: bad LoweredAddr")
		}

	case ssa.OpWasm32LoweredConvert:
		getValue32(s, v.Args[0])

	case ssa.OpWasm32Select:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		getValue32(s, v.Args[2])
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I64AddConst:
		getValue32(s, v.Args[0])
		i64Const(s, v.AuxInt)
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I32AddConst:
		getValue32(s, v.Args[0])
		i32Const(s, int32(v.AuxInt))
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I64Const:
		i64Const(s, v.AuxInt)

	case ssa.OpWasm32I32Const:
		i32Const(s, int32(v.AuxInt))

	case ssa.OpWasm32F32Const:
		f32Const(s, v.AuxFloat())

	case ssa.OpWasm32F64Const:
		f64Const(s, v.AuxFloat())

	case
		ssa.OpWasm32I32Load8U, ssa.OpWasm32I32Load8S, ssa.OpWasm32I32Load16U, ssa.OpWasm32I32Load16S,
		ssa.OpWasm32I32Load,
		ssa.OpWasm32I64Load8U, ssa.OpWasm32I64Load8S, ssa.OpWasm32I64Load16U, ssa.OpWasm32I64Load16S,
		ssa.OpWasm32I64Load32U, ssa.OpWasm32I64Load32S, ssa.OpWasm32I64Load,
		ssa.OpWasm32F32Load, ssa.OpWasm32F64Load:
		getValue32(s, v.Args[0])
		p := s.Prog(v.Op.Asm())
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}

		/*
			case ssa.OpWasm32I64Eqz:
				getValue32(s, v.Args[0])
				s.Prog(v.Op.Asm())
				if extend {
					s.Prog(wasm32.AI64ExtendI32U)
				}
		*/

	case ssa.OpWasm32I32Eqz:
		getValue32(s, v.Args[0])
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I32Eq, ssa.OpWasm32I32Ne, ssa.OpWasm32I32LtS, ssa.OpWasm32I32LtU, ssa.OpWasm32I32GtS, ssa.OpWasm32I32GtU, ssa.OpWasm32I32LeS, ssa.OpWasm32I32LeU, ssa.OpWasm32I32GeS, ssa.OpWasm32I32GeU,
		ssa.OpWasm32F32Eq, ssa.OpWasm32F32Ne, ssa.OpWasm32F32Lt, ssa.OpWasm32F32Gt, ssa.OpWasm32F32Le, ssa.OpWasm32F32Ge:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		s.Prog(v.Op.Asm())

	/*
		case ssa.OpWasm32I64Eq, ssa.OpWasm32I64Ne, ssa.OpWasm32I64LtS, ssa.OpWasm32I64LtU, ssa.OpWasm32I64GtS, ssa.OpWasm32I64GtU, ssa.OpWasm32I64LeS, ssa.OpWasm32I64LeU, ssa.OpWasm32I64GeS, ssa.OpWasm32I64GeU,
	*/
	case ssa.OpWasm32F64Eq, ssa.OpWasm32F64Ne, ssa.OpWasm32F64Lt, ssa.OpWasm32F64Gt, ssa.OpWasm32F64Le, ssa.OpWasm32F64Ge:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		s.Prog(v.Op.Asm())
		/*
			if extend {
				s.Prog(wasm32.AI64ExtendI32U)
			}
		*/

	case ssa.OpWasm32I32Add, ssa.OpWasm32I32Sub, ssa.OpWasm32I32Mul, ssa.OpWasm32I32DivU, ssa.OpWasm32I32RemS, ssa.OpWasm32I32RemU, ssa.OpWasm32I32And, ssa.OpWasm32I32Or, ssa.OpWasm32I32Xor, ssa.OpWasm32I32Shl, ssa.OpWasm32I32ShrS, ssa.OpWasm32I32ShrU, ssa.OpWasm32I32Rotl,
		ssa.OpWasm32F32Add, ssa.OpWasm32F32Sub, ssa.OpWasm32F32Mul, ssa.OpWasm32F32Div, ssa.OpWasm32F32Copysign:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I64Add, ssa.OpWasm32I64Sub, ssa.OpWasm32I64Mul, ssa.OpWasm32I64DivU, ssa.OpWasm32I64RemS, ssa.OpWasm32I64RemU, ssa.OpWasm32I64And, ssa.OpWasm32I64Or, ssa.OpWasm32I64Xor, ssa.OpWasm32I64Shl, ssa.OpWasm32I64ShrS, ssa.OpWasm32I64ShrU, ssa.OpWasm32I64Rotl,
		ssa.OpWasm32F64Add, ssa.OpWasm32F64Sub, ssa.OpWasm32F64Mul, ssa.OpWasm32F64Div, ssa.OpWasm32F64Copysign:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32I32DivS:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		if v.Type.Size() == 8 {
			// Division of int64 needs helper function wasmDiv to handle the MinInt64 / -1 case.
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmDiv}
			break
		}
		s.Prog(wasm32.AI32DivS)

	case ssa.OpWasm32I64DivS:
		getValue32(s, v.Args[0])
		getValue32(s, v.Args[1])
		if v.Type.Size() == 8 {
			// Division of int64 needs helper function wasmDiv to handle the MinInt64 / -1 case.
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmDiv}
			break
		}
		s.Prog(wasm32.AI32DivS)

	case ssa.OpWasm32I32TruncSatF32S, ssa.OpWasm32I32TruncSatF64S:
		getValue32(s, v.Args[0])
		if buildcfg.GOWASM.SatConv {
			s.Prog(v.Op.Asm())
		} else {
			if v.Op == ssa.OpWasm32I32TruncSatF32S {
				s.Prog(wasm32.AF64PromoteF32)
			}
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmTruncS}
		}

	case ssa.OpWasm32I64TruncSatF32S, ssa.OpWasm32I64TruncSatF64S:
		getValue32(s, v.Args[0])
		if buildcfg.GOWASM.SatConv {
			s.Prog(v.Op.Asm())
		} else {
			if v.Op == ssa.OpWasm32I64TruncSatF32S {
				s.Prog(wasm32.AF64PromoteF32)
			}
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmTruncS}
		}

	case ssa.OpWasm32I32TruncSatF32U, ssa.OpWasm32I32TruncSatF64U:
		getValue32(s, v.Args[0])
		if buildcfg.GOWASM.SatConv {
			s.Prog(v.Op.Asm())
		} else {
			if v.Op == ssa.OpWasm32I32TruncSatF32U {
				s.Prog(wasm32.AF64PromoteF32)
			}
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmTruncU}
		}

	case ssa.OpWasm32I64TruncSatF32U, ssa.OpWasm32I64TruncSatF64U:
		getValue32(s, v.Args[0])
		if buildcfg.GOWASM.SatConv {
			s.Prog(v.Op.Asm())
		} else {
			if v.Op == ssa.OpWasm32I64TruncSatF32U {
				s.Prog(wasm32.AF64PromoteF32)
			}
			p := s.Prog(wasm32.ACall)
			p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.WasmTruncU}
		}

	case ssa.OpWasm32F32DemoteF64:
		getValue32(s, v.Args[0])
		s.Prog(v.Op.Asm())

	case ssa.OpWasm32F64PromoteF32:
		getValue32(s, v.Args[0])
		s.Prog(v.Op.Asm())

	case
		ssa.OpWasm32F32ConvertI64S, ssa.OpWasm32F32ConvertI64U,
		ssa.OpWasm32F64ConvertI64S, ssa.OpWasm32F64ConvertI64U,
		ssa.OpWasm32F32ConvertI32S, ssa.OpWasm32F32ConvertI32U,
		ssa.OpWasm32F64ConvertI32S, ssa.OpWasm32F64ConvertI32U,
		ssa.OpWasm32I64Extend8S, ssa.OpWasm32I64Extend16S, ssa.OpWasm32I64Extend32S,
		ssa.OpWasm32I32Extend8S, ssa.OpWasm32I32Extend16S,
		ssa.OpWasm32F32Neg, ssa.OpWasm32F32Sqrt, ssa.OpWasm32F32Trunc, ssa.OpWasm32F32Ceil, ssa.OpWasm32F32Floor, ssa.OpWasm32F32Nearest, ssa.OpWasm32F32Abs,
		ssa.OpWasm32F64Neg, ssa.OpWasm32F64Sqrt, ssa.OpWasm32F64Trunc, ssa.OpWasm32F64Ceil, ssa.OpWasm32F64Floor, ssa.OpWasm32F64Nearest, ssa.OpWasm32F64Abs,
		ssa.OpWasm32I64Ctz, ssa.OpWasm32I64Clz, ssa.OpWasm32I64Popcnt,
		ssa.OpWasm32I32Ctz, ssa.OpWasm32I32Clz, ssa.OpWasm32I32Popcnt:
		getValue32(s, v.Args[0])
		s.Prog(v.Op.Asm())

	case ssa.OpLoadReg:
		p := s.Prog(loadOp(v.Type))
		ssagen.AddrAuto(&p.From, v.Args[0])

	case ssa.OpCopy:
		getValue32(s, v.Args[0])

	default:
		v.Fatalf("unexpected op: %s", v.Op)
	}

	return true
}

func isCmp(v *ssa.Value) bool {
	switch v.Op {
	case ssa.OpWasm32I64Eqz, ssa.OpWasm32I64Eq, ssa.OpWasm32I64Ne, ssa.OpWasm32I64LtS, ssa.OpWasm32I64LtU, ssa.OpWasm32I64GtS, ssa.OpWasm32I64GtU, ssa.OpWasm32I64LeS, ssa.OpWasm32I64LeU, ssa.OpWasm32I64GeS, ssa.OpWasm32I64GeU,
		ssa.OpWasm32F32Eq, ssa.OpWasm32F32Ne, ssa.OpWasm32F32Lt, ssa.OpWasm32F32Gt, ssa.OpWasm32F32Le, ssa.OpWasm32F32Ge,
		ssa.OpWasm32F64Eq, ssa.OpWasm32F64Ne, ssa.OpWasm32F64Lt, ssa.OpWasm32F64Gt, ssa.OpWasm32F64Le, ssa.OpWasm32F64Ge:
		return true
	default:
		return false
	}
}

func getValue32(s *ssagen.State, v *ssa.Value) {
	if v.OnWasmStack {
		/*
			if s.FuncInfo().Text.Ctxt.Pkgpath == "runtime/internal/sys" {
				fmt.Printf("s< %s %d\n", v.Op, s.OnWasmStackSkipped)
			}
		*/
		s.OnWasmStackSkipped--
		ssaGenValueOnStack(s, v, false)
		return
	}

	reg := v.Reg()
	getReg(s, reg)
}

func i32Const(s *ssagen.State, val int32) {
	p := s.Prog(wasm32.AI32Const)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(val)}
}

func i64Const(s *ssagen.State, val int64) {
	p := s.Prog(wasm32.AI32Const)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: val}
}

func f32Const(s *ssagen.State, val float64) {
	p := s.Prog(wasm32.AF32Const)
	p.From = obj.Addr{Type: obj.TYPE_FCONST, Val: val}
}

func f64Const(s *ssagen.State, val float64) {
	p := s.Prog(wasm32.AF64Const)
	p.From = obj.Addr{Type: obj.TYPE_FCONST, Val: val}
}

func getReg(s *ssagen.State, reg int16) {
	p := s.Prog(wasm32.AGet)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: reg}
}

func setReg(s *ssagen.State, reg int16) {
	p := s.Prog(wasm32.ASet)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: reg}
	if reg == 16392 {
		panic("nope")
	}
}

func loadOp(t *types.Type) obj.As {
	if t.IsFloat() {
		switch t.Size() {
		case 4:
			return wasm32.AF32Load
		case 8:
			return wasm32.AF64Load
		default:
			panic("bad load type")
		}
	}

	switch t.Size() {
	case 1:
		if t.IsSigned() {
			return wasm32.AI32Load8S
		}
		return wasm32.AI32Load8U
	case 2:
		if t.IsSigned() {
			return wasm32.AI32Load16S
		}
		return wasm32.AI32Load16U
	case 4:
		return wasm32.AI32Load
	default:
		panic("bad load type")
	}
}

func storeOp(t *types.Type) obj.As {
	if t.IsFloat() {
		switch t.Size() {
		case 4:
			return wasm32.AF32Store
		case 8:
			return wasm32.AF64Store
		default:
			panic("bad store type")
		}
	}

	switch t.Size() {
	case 1:
		return wasm32.AI32Store8
	case 2:
		return wasm32.AI32Store16
	case 4:
		return wasm32.AI32Store
	default:
		panic("bad store type")
	}
}
