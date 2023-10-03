// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//go:build wasm32

#include "go_asm.h"
#include "textflag.h"

// memequal(p, q unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal(SB), NOSPLIT, $0-25
	Get SP
	I32Load a+0(FP)
	I32Load b+4(FP)
	I32Load size+8(FP)
	Call memeqbody<>(SB)
	I32Store8 ret+12(FP)
	RET

// memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB), NOSPLIT, $0-17
	Get SP
	I32Load a+0(FP)
	I32Load b+4(FP)
	I32Load 8(CTXT) // compiler stores size at offset 8 in the closure
	Call memeqbody<>(SB)
	I32Store8 ret+8(FP)
	RET

// params: a, b, len
// ret: 0/1
TEXT memeqbody<>(SB), NOSPLIT, $0-0
	Get R0
	Get R1
	I32Eq
	If
		I32Const $1
		Return
	End

loop:
	Loop
		Get R2
		I32Eqz
		If
			I32Const $1
			Return
		End

		Get R0
		I32Load8U $0
		Get R1
		I32Load8U $0
		I32Ne
		If
			I32Const $0
			Return
		End

		Get R0
		I32Const $1
		I32Add
		Set R0

		Get R1
		I32Const $1
		I32Add
		Set R1

		Get R2
		I32Const $1
		I32Sub
		Set R2

		Br loop
	End
	UNDEF
