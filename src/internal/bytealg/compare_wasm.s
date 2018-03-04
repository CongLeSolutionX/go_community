// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·Compare(SB), NOSPLIT, $0-56
    Get SP
      I64Load s1_base+0(FP)
      I64Load s1_len+8(FP)
      I64Load s2_base+24(FP)
      I64Load s2_len+32(FP)
    Call cmpbody<>(SB)
  I64Store ret+48(FP)
  RET

TEXT bytes·Compare(SB), NOSPLIT, $0-56
    Get SP
      I64Load s1_base+0(FP)
      I64Load s1_len+8(FP)
      I64Load s2_base+24(FP)
      I64Load s2_len+32(FP)
    Call cmpbody<>(SB)
  I64Store ret+48(FP)
  RET

TEXT runtime·cmpstring(SB), NOSPLIT, $0-40
    Get SP
      I64Load s1_base+0(FP)
      I64Load s1_len+8(FP)
      I64Load s2_base+16(FP)
      I64Load s2_len+24(FP)
    Call cmpbody<>(SB)
  I64Store ret+32(FP)
  RET

// params: a, alen, b, blen
// ret: -1/0/1
TEXT cmpbody<>(SB), NOSPLIT, $0-0
  // len = min(alen, blen)
      Get $1
      Get $3
        Get $1
        Get $3
      I64LtU
    Select
  Set I4

          Get $0
        I32WrapI64
          Get $2
        I32WrapI64
          Get I4
        I32WrapI64
      Call memcmp<>(SB)
    I64ExtendSI32
  Set I5

      Get I5
    I64Eqz
  If
    // check length
        Get $1
        Get $3
      I64Sub
    Set I5
  End

    I64Const $0
      I64Const $-1
      I64Const $1
        Get I5
        I64Const $0
      I64LtS
    Select
      Get I5
    I64Eqz
  Select

// compiled with emscripten
TEXT memcmp<>(SB), NOSPLIT, $0-0
  Get $2
  If $1
  Loop
  Get $0
  I32Load8S $0
  Tee $3
  Get $1
  I32Load8S $0
  Tee $4
  I32Eq
  If
  Get $0
  I32Const $1
  I32Add
  Set $0
  Get $1
  I32Const $1
  I32Add
  Set $1
  I32Const $0
  Get $2
  I32Const $-1
  I32Add
  Tee $2
  I32Eqz
  BrIf $3
  Drop
  Br $1
  End
  End
  Get $3
  I32Const $255
  I32And
  Get $4
  I32Const $255
  I32And
  I32Sub
  Else
  I32Const $0
  End
