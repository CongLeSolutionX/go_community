// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// void runtime·memmove(void*, void*, uintptr)
TEXT runtime·memmove(SB), NOSPLIT, $0-24
    I64Load to+0(FP)
  Set I0

    I64Load from+8(FP)
  Set I1

    I64Load n+16(FP)
  Set I2

      Get I0
      Get I1
    I64LtU
  If // forward
    Block
      Loop
            Get I2
            I64Const $8
          I64LtU
        BrIf $1

            Get I0
          I32WrapI64
          I64Load (I1)
        I64Store $0

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

        Br $0
      End
    End

    Loop
          Get I2
        I64Eqz
      BrIf $1

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

      Br $0
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

    Block
      Loop
            Get I2
            I64Const $8
          I64LtU
        BrIf $1

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

            Get I0
          I32WrapI64
          I64Load (I1)
        I64Store $0

        Br $0
      End
    End

    Loop
          Get I2
        I64Eqz
      BrIf $1

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

      Br $0
    End
  End

  RET
