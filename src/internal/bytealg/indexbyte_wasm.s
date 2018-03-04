// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·IndexByte(SB), NOSPLIT, $0-40
	I64Load s+0(FP)
	I32WrapI64
	I32Load8U c+24(FP)
	I64Load s_len+8(FP)
	I32WrapI64
	Call memchr<>(SB)
	I64ExtendSI32
	Set R0

	Get SP
	I64Const $-1
	Get R0
	I64Load s+0(FP)
	I64Sub
	Get R0
	I64Eqz $0
	Select
	I64Store ret+32(FP)

	RET

TEXT ·IndexByteString(SB), NOSPLIT, $0-32
	Get SP
	I64Load s+0(FP)
	I32WrapI64
	I32Load8U c+16(FP)
	I64Load s_len+8(FP)
	I32WrapI64
	Call memchr<>(SB)
	I64ExtendSI32
	Set R0

	I64Const $-1
	Get R0
	I64Load s+0(FP)
	I64Sub
	Get R0
	I64Eqz $0
	Select
	I64Store ret+24(FP)

	RET

TEXT bytes·IndexByte(SB), NOSPLIT, $0-40
	Get SP
	I64Load s+0(FP)
	I32WrapI64
	I32Load8U c+24(FP)
	I64Load s_len+8(FP)
	I32WrapI64
	Call memchr<>(SB)
	I64ExtendSI32
	Set R0

	I64Const $-1
	Get R0
	I64Load s+0(FP)
	I64Sub
	Get R0
	I64Eqz $0
	Select
	I64Store ret+32(FP)

	RET

TEXT strings·IndexByte(SB), NOSPLIT, $0-32
	Get SP
	I64Load s+0(FP)
	I32WrapI64
	I32Load8U c+16(FP)
	I64Load s_len+8(FP)
	I32WrapI64
	Call memchr<>(SB)
	I64ExtendSI32
	Set R0

	I64Const $-1
	Get R0
	I64Load s+0(FP)
	I64Sub
	Get R0
	I64Eqz $0
	Select
	I64Store ret+24(FP)
	RET

// compiled with emscripten
// params: s, c, len
// ret: index
TEXT memchr<>(SB), NOSPLIT, $0
	Get $1
	I32Const $255
	I32And
	Set $4
	Block
	Block
	Get $2
	I32Const $0
	I32Ne
	Tee $3
	Get $0
	I32Const $3
	I32And
	I32Const $0
	I32Ne
	I32And
	If
	Get $1
	I32Const $255
	I32And
	Set $5
	Loop
	Get $0
	I32Load8U $0
	Get $5
	I32Eq
	BrIf $2
	Get $2
	I32Const $-1
	I32Add
	Tee $2
	I32Const $0
	I32Ne
	Tee $3
	Get $0
	I32Const $1
	I32Add
	Tee $0
	I32Const $3
	I32And
	I32Const $0
	I32Ne
	I32And
	BrIf $0
	End
	End
	Get $3
	BrIf $0
	I32Const $0
	Set $1
	Br $1
	End
	Get $0
	I32Load8U $0
	Get $1
	I32Const $255
	I32And
	Tee $3
	I32Eq
	If
	Get $2
	Set $1
	Else
	Get $4
	I32Const $16843009
	I32Mul
	Set $4
	Block
	Block
	Get $2
	I32Const $3
	I32GtU
	If
	Get $2
	Set $1
	Loop
	Get $0
	I32Load $0
	Get $4
	I32Xor
	Tee $2
	I32Const $-2139062144
	I32And
	I32Const $-2139062144
	I32Xor
	Get $2
	I32Const $-16843009
	I32Add
	I32And
	I32Eqz
	If
	Get $0
	I32Const $4
	I32Add
	Set $0
	Get $1
	I32Const $-4
	I32Add
	Tee $1
	I32Const $3
	I32GtU
	BrIf $1
	Br $3
	End
	End
	Else
	Get $2
	Set $1
	Br $1
	End
	Br $1
	End
	Get $1
	I32Eqz
	If
	I32Const $0
	Set $1
	Br $3
	End
	End
	Loop
	Get $0
	I32Load8U $0
	Get $3
	I32Eq
	BrIf $2
	Get $0
	I32Const $1
	I32Add
	Set $0
	Get $1
	I32Const $-1
	I32Add
	Tee $1
	BrIf $0
	I32Const $0
	Set $1
	End
	End
	End
	Get $0
	I32Const $0
	Get $1
	Select
	Return
