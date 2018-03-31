// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT runtime·wasmmove(SB), NOSPLIT, $0-0
loop:
	Loop
		// *dst = *src
		Get $0
		Get $1
		I64Load $0
		I64Store $0

		// n--
		Get $2
		I32Const $1
		I32Sub
		Set $2

		// n == 0
		Get $2
		I32Eqz
		If
			Return
		End

		// dst += 8
		Get $0
		I32Const $8
		I32Add
		Set $0

		// src += 8
		Get $1
		I32Const $8
		I32Add
		Set $1

		Br loop
	End
	UNDEF

TEXT runtime·wasmzero(SB), NOSPLIT, $0-0
loop:
	Loop
		// *dst = 0
		Get $0
		I64Const $0
		I64Store $0

		// n--
		Get $1
		I32Const $1
		I32Sub
		Set $1

		// n == 0
		Get $1
		I32Eqz
		If
			Return
		End

		// dst += 8
		Get $0
		I32Const $8
		I32Add
		Set $0

		Br loop
	End
	UNDEF

TEXT runtime·wasmdiv(SB), NOSPLIT, $0-0
	Get $0
	I64Const $-0x8000000000000000
	I64Eq
	If
		Get $1
		I64Const $-1
		I64Eq
		If
			I64Const $-0x8000000000000000
			Return
		End
	End
	Get $0
	Get $1
	I64DivS

TEXT runtime·wasmtrunc(SB), NOSPLIT, $0-0
	Get $0
	Get $0
	F64Ne // NaN
	If
		I64Const $0x8000000000000000
		Return
	End

	Get $0
	F64Const $9223372036854775807.
	F64Gt
	If
		I64Const $0x8000000000000000
		Return
	End

	Get $0
	F64Const $-9223372036854775808.
	F64Lt
	If
		I64Const $0x8000000000000000
		Return
	End

	Get $0
	I64TruncSF64

TEXT runtime·exit(SB), NOSPLIT, $0-8
	Call runtime·wasmexit(SB)
	Drop
	I32Const $0
	Set SP
	I32Const $1

TEXT runtime·exitThread(SB), NOSPLIT, $0-0
	UNDEF

TEXT runtime·getclosureptr(SB), NOSPLIT, $0-0
	UNDEF

TEXT runtime·osyield(SB), NOSPLIT, $0-0
	UNDEF

TEXT runtime·usleep(SB), NOSPLIT, $0-0
	RET // FIXME

TEXT runtime·currentMemory(SB), NOSPLIT, $0
	Get SP
	CurrentMemory
	I32Store ret+0(FP)
	RET

TEXT runtime·growMemory(SB), NOSPLIT, $0
	Get SP
	I32Load pages+0(FP)
	GrowMemory
	I32Store ret+8(FP)
	RET

TEXT ·wasmexit(SB), NOSPLIT, $0
	CallImport
	RET

TEXT ·wasmwrite(SB), NOSPLIT, $0
	CallImport
	RET

TEXT ·nanotime(SB), NOSPLIT, $0
	CallImport
	RET

TEXT ·walltime(SB), NOSPLIT, $0
	CallImport
	RET
