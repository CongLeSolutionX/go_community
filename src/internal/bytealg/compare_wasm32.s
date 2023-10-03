// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//go:build wasm32

#include "go_asm.h"
#include "textflag.h"

TEXT ·Compare(SB), NOSPLIT, $0-56
	Get SP
	I32Load a_base+0(FP)
	I32Load a_len+4(FP)
	I32Load b_base+12(FP)
	I32Load b_len+16(FP)
	Call cmpbody<>(SB)
	I32Store ret+24(FP)
	RET

TEXT runtime·cmpstring(SB), NOSPLIT, $0-40
	Get SP
	I32Load a_base+0(FP)
	I32Load a_len+4(FP)
	I32Load b_base+8(FP)
	I32Load b_len+16(FP)
	Call cmpbody<>(SB)
	I32Store ret+24(FP)
	RET

// params: a, alen, b, blen
// ret: -1/0/1
TEXT cmpbody<>(SB), NOSPLIT, $0-0
	// len = min(alen, blen)
	Get R1
	Get R3
	Get R1
	Get R3
	I32LtU
	Select
	Set R4

	Get R0
	Get R2
	Get R4
	Call memcmp<>(SB)
	Tee R5

	I32Eqz
	If
		// check length
		Get R1
		Get R3
		I32Sub
		Set R5
	End

	I32Const $0
	I32Const $-1
	I32Const $1
	Get R5
	I32Const $0
	I32LtS
	Select
	Get R5
	I32Eqz
	Select
	Return

// compiled with emscripten
// params: a, b, len
// ret: <0/0/>0
TEXT memcmp<>(SB), NOSPLIT, $0-0
	Get R2
	If $1
	Loop
	Get R0
	I32Load8S $0
	Tee R3
	Get R1
	I32Load8S $0
	Tee R4
	I32Eq
	If
	Get R0
	I32Const $1
	I32Add
	Set R0
	Get R1
	I32Const $1
	I32Add
	Set R1
	I32Const $0
	Get R2
	I32Const $-1
	I32Add
	Tee R2
	I32Eqz
	BrIf $3
	Drop
	Br $1
	End
	End
	Get R3
	I32Const $255
	I32And
	Get R4
	I32Const $255
	I32And
	I32Sub
	Else
	I32Const $0
	End
	Return
