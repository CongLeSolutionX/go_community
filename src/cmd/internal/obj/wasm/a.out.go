// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import "cmd/internal/obj"

//go:generate go run ../stringer.go -i $GOFILE -o anames.go -p wasm

const (
	/* mark flags */
	DONE          = 1 << iota
	PRESERVEFLAGS // not allowed to clobber flags
)

/*
 *	wasm
 */
const (
	AAddrOf = obj.ABaseWasm + obj.A_ARCHSPECIFIC + iota
	ACallImport
	AGet
	ASet
	ATee

	AUnreachable // 0x00
	ANop
	ABlock
	ALoop
	AIf
	AElse

	AEnd // 0x0B
	ABr
	ABrIf
	ABrTable
	AReturn
	ACall
	ACallIndirect

	ADrop // 0x1A
	ASelect

	AI32Load // 0x28
	AI64Load
	AF32Load
	AF64Load
	AI32Load8S
	AI32Load8U
	AI32Load16S
	AI32Load16U
	AI64Load8S
	AI64Load8U
	AI64Load16S
	AI64Load16U
	AI64Load32S
	AI64Load32U
	AI32Store
	AI64Store
	AF32Store
	AF64Store
	AI32Store8
	AI32Store16
	AI64Store8
	AI64Store16
	AI64Store32
	ACurrentMemory
	AGrowMemory

	AI32Const
	AI64Const
	AF32Const
	AF64Const

	AI32Eqz
	AI32Eq
	AI32Ne
	AI32LtS
	AI32LtU
	AI32GtS
	AI32GtU
	AI32LeS
	AI32LeU
	AI32GeS
	AI32GeU

	AI64Eqz
	AI64Eq
	AI64Ne
	AI64LtS
	AI64LtU
	AI64GtS
	AI64GtU
	AI64LeS
	AI64LeU
	AI64GeS
	AI64GeU

	AF32Eq
	AF32Ne
	AF32Lt
	AF32Gt
	AF32Le
	AF32Ge

	AF64Eq
	AF64Ne
	AF64Lt
	AF64Gt
	AF64Le
	AF64Ge

	AI32Clz
	AI32Ctz
	AI32Popcnt
	AI32Add
	AI32Sub
	AI32Mul
	AI32DivS
	AI32DivU
	AI32RemS
	AI32RemU
	AI32And
	AI32Or
	AI32Xor
	AI32Shl
	AI32ShrS
	AI32ShrU
	AI32Rotl
	AI32Rotr

	AI64Clz
	AI64Ctz
	AI64Popcnt
	AI64Add
	AI64Sub
	AI64Mul
	AI64DivS
	AI64DivU
	AI64RemS
	AI64RemU
	AI64And
	AI64Or
	AI64Xor
	AI64Shl
	AI64ShrS
	AI64ShrU
	AI64Rotl
	AI64Rotr

	AF32Abs
	AF32Neg
	AF32Ceil
	AF32Floor
	AF32Trunc
	AF32Nearest
	AF32Sqrt
	AF32Add
	AF32Sub
	AF32Mul
	AF32Div
	AF32Min
	AF32Max
	AF32Copysign

	AF64Abs
	AF64Neg
	AF64Ceil
	AF64Floor
	AF64Trunc
	AF64Nearest
	AF64Sqrt
	AF64Add
	AF64Sub
	AF64Mul
	AF64Div
	AF64Min
	AF64Max
	AF64Copysign

	AI32WrapI64
	AI32TruncSF32
	AI32TruncUF32
	AI32TruncSF64
	AI32TruncUF64
	AI64ExtendSI32
	AI64ExtendUI32
	AI64TruncSF32
	AI64TruncUF32
	AI64TruncSF64
	AI64TruncUF64
	AF32ConvertSI32
	AF32ConvertUI32
	AF32ConvertSI64
	AF32ConvertUI64
	AF32DemoteF64
	AF64ConvertSI32
	AF64ConvertUI32
	AF64ConvertSI64
	AF64ConvertUI64
	AF64PromoteF32
	AI32ReinterpretF32
	AI64ReinterpretF64
	AF32ReinterpretI32
	AF64ReinterpretI64

	AMOVB
	AMOVH
	AMOVW
	AMOVD

	AWORD
	ALAST
)

const (
	REG_NONE = 0
)

const (
	REG_PC_F = obj.RBaseWasm + iota
	REG_PC_B
	REG_SP
	REG_CTX
	REGG
	REG_RET0
	REG_RET1
	REG_RET2
	REG_RET3

	REG_I0
	REG_I1
	REG_I2
	REG_I3
	REG_I4
	REG_I5
	REG_I6
	REG_I7
	REG_I8
	REG_I9
	REG_I10
	REG_I11
	REG_I12
	REG_I13
	REG_I14
	REG_I15

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

	MAXREG

	REGARG = -1
	// REGRET   = REG_AX
	// FREGRET  = REG_X0
	REGSP = REG_SP
	// REGCTXT  = REG_DX
	// REGEXT   = REG_R15     /* compiler allocates external registers R15 down */
	// FREGMIN  = REG_X0 + 5  /* first register variable */
	// FREGEXT  = REG_X0 + 15 /* first external register */
	T_TYPE   = 1 << 0
	T_INDEX  = 1 << 1
	T_OFFSET = 1 << 2
	T_FCONST = 1 << 3
	T_SYM    = 1 << 4
	T_SCONST = 1 << 5
	T_64     = 1 << 6
	T_GOTYPE = 1 << 7
)
