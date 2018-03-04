// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·Equal(SB), NOSPLIT, $0-49
	MOVD a_len+8(FP), I0
	MOVD b_len+32(FP), I1
	Get I0
	Get I1
	I64Eq
	If
		Get SP
		I64Load a+0(FP)
		I64Load b+24(FP)
		Get I0
		Call memeqbody<>(SB)
		I64Store8 ret+48(FP)
	Else
		Get SP
		I64Const $0
		I64Store8 ret+48(FP)
	End
	RET

TEXT bytes·Equal(SB), NOSPLIT, $0-49
	MOVD a_len+8(FP), I0
	MOVD b_len+32(FP), I1
	Get I0
	Get I1
	I64Eq
	If
		Get SP
		I64Load a+0(FP)
		I64Load b+24(FP)
		Get I0
		Call memeqbody<>(SB)
		I64Store8 ret+48(FP)
	Else
		Get SP
		I64Const $0
		I64Store8 ret+48(FP)
	End
	RET

// memequal(p, q unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal(SB), NOSPLIT, $0-25
	Get SP
	I64Load a+0(FP)
	I64Load b+8(FP)
	I64Load size+16(FP)
	Call memeqbody<>(SB)
	I64Store8 ret+24(FP)
	RET

// memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB), NOSPLIT, $0-17
	Get SP
	I64Load a+0(FP)
	I64Load b+8(FP)
	I64Load 8(CTX)
	Call memeqbody<>(SB)
	I64Store8 ret+16(FP)
	RET

// params: a, b, len
// ret: 0/1
TEXT memeqbody<>(SB), NOSPLIT, $0-0
	Get $0
	Get $1
	I64Eq
	If $0
		I64Const $1
		Br $1
	End

	Block
		Loop
			Get $2
			I64Eqz
			BrIf $1

			Get $0
			I32WrapI64
			I64Load8U $0
			Get $1
			I32WrapI64
			I64Load8U $0
			I64Ne
			If
				I64Const $0
				Br $3
			End

			Get $0
			I64Const $1
			I64Add
			Set $0

			Get $1
			I64Const $1
			I64Add
			Set $1

			Get $2
			I64Const $1
			I64Sub
			Set $2

			Br $0
		End
	End

	I64Const $1
