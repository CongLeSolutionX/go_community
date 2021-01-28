// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// makeFuncStub is the code half of the function returned by MakeFunc.
// See the comment on the declaration of makeFuncStub in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
// makeFuncStub must be ABIInternal because it is placed directly
// in function values.
// The frame size is enormous because this function calls into callReflect
// which expects an internal/abi.RegArgs and we set that up in the argument
// space so any pointers passed through are visible to the GC via callReflect's
// stack map. callReflect then returns another internal/abi.RegArgs value
// (modifying the argument and returning is unsafe, because the compiler
// could optimize that away). That's then 32 bytes for the first few arguments
// to callReflect, then 276 bytes for the calling RegArgs, and 276 bytes for
// the returned RegArgs.
//
// TODO(mknyszek): Modify this function to call callReflect with
// ABIInternal.
TEXT ·makeFuncStub<ABIInternal>(SB),(NOSPLIT|WRAPPER),$584
	NO_LOCAL_POINTERS
	LEAQ    32(SP), R12
	CALL    ·runtimeSpillArgs<>(SB)
	MOVQ	DX, 0(SP)
	MOVQ    R12, 8(SP)
	CALL    ·moveMakeFuncArgPtrs<>(SB)
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)
	MOVB	$0, 24(SP)
	LEAQ	24(SP), AX
	MOVQ	AX, 16(SP)
	CALL	·callReflect<ABIInternal>(SB)
	LEAQ    304(SP), R12
	CALL    ·runtimeUnspillArgs<>(SB)
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
// methodValueCall must be ABIInternal because it is placed directly
// in function values.
// The frame size is enormous because this function calls into callMethod
// which expects an internal/abi.RegArgs and we set that up in the argument
// space so any pointers passed through are visible to the GC via callMethod's
// stack map. callMethod then returns another internal/abi.RegArgs value
// (modifying the argument and returning is unsafe, because the compiler
// could optimize that away). That's then 32 bytes for the first few arguments
// to callMethod, then 276 bytes for the calling RegArgs, and 276 bytes for
// the returned RegArgs.
//
// TODO(mknyszek): Modify this function to call callMethod with
// ABIInternal.
TEXT ·methodValueCall<ABIInternal>(SB),(NOSPLIT|WRAPPER),$584
	NO_LOCAL_POINTERS
	LEAQ    32(SP), R12
	CALL    ·runtimeSpillArgs<>(SB)
	MOVQ	DX, 0(SP)
	MOVQ    R12, 8(SP)
	CALL    ·moveMakeFuncArgPtrs<>(SB)
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)
	MOVB	$0, 24(SP)
	LEAQ	24(SP), AX
	MOVQ	AX, 16(SP)
	CALL	·callMethod<ABIInternal>(SB)
	LEAQ    304(SP), R12
	CALL    ·runtimeUnspillArgs<>(SB)
	RET
