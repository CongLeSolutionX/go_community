// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// void runtime·memclrNoHeapPointers(void*, uintptr)
TEXT runtime·memclrNoHeapPointers(SB), NOSPLIT, $0-16
    I64Load ptr+0(FP)
  Set I0

    I64Load n+8(FP)
  Set I1

  Block
    Loop
          Get I1
        I64Eqz
      BrIf $1

          Get I0
        I32WrapI64
        I64Const $0
      I64Store8 $0

          Get I0
          I64Const $1
        I64Add
      Set I0

          Get I1
          I64Const $1
        I64Sub
      Set I1

      Br $0
    End
  End

  RET
