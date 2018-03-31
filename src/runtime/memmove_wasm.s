// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// void runtime·memmove(void*, void*, uintptr)
TEXT runtime·memmove(SB), NOSPLIT, $0-24
	MOVD to+0(FP), I0
	MOVD from+8(FP), I1
	MOVD n+16(FP), I2

	Get I0
	Get I1
	I64LtU
	If // forward
exit_forward_64:
		Block
loop_forward_64:
			Loop
				Get I2
				I64Const $8
				I64LtU
				BrIf exit_forward_64

				MOVD 0(I1), 0(I0)

				Get I0
				I64Const $8
				I64Add
				Set I0

				Get I1
				I64Const $8
				I64Add
				Set I1

				Get I2
				I64Const $8
				I64Sub
				Set I2

				Br loop_forward_64
			End
		End

loop_forward_8:
		Loop
			Get I2
			I64Eqz
			If
				RET
			End

			Get I0
			I32WrapI64
			I64Load8U (I1)
			I64Store8 $0

			Get I0
			I64Const $1
			I64Add
			Set I0

			Get I1
			I64Const $1
			I64Add
			Set I1

			Get I2
			I64Const $1
			I64Sub
			Set I2

			Br loop_forward_8
		End

	Else
		// backward
		Get I0
		Get I2
		I64Add
		Set I0

		Get I1
		Get I2
		I64Add
		Set I1

exit_backward_64:
		Block
loop_backward_64:
			Loop
				Get I2
				I64Const $8
				I64LtU
				BrIf exit_backward_64

				Get I0
				I64Const $8
				I64Sub
				Set I0

				Get I1
				I64Const $8
				I64Sub
				Set I1

				Get I2
				I64Const $8
				I64Sub
				Set I2

				MOVD 0(I1), 0(I0)

				Br loop_backward_64
			End
		End

loop_backward_8:
		Loop
			Get I2
			I64Eqz
			If
				RET
			End

			Get I0
			I64Const $1
			I64Sub
			Set I0

			Get I1
			I64Const $1
			I64Sub
			Set I1

			Get I2
			I64Const $1
			I64Sub
			Set I2

			Get I0
			I32WrapI64
			I64Load8U (I1)
			I64Store8 $0

			Br loop_backward_8
		End
	End

	UNDEF
