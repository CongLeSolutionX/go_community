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
// The frame contains a stack-allocated RegArgs after the space saved for
// arguments to callReflect. This RegArgs is visible to the GC via a special
// case in the runtime's stack map code, so its offset is known to the runtime.
//
// TODO(mknyszek): Modify this function to call callReflect with
// ABIInternal.
TEXT ·makeFuncStub<ABIInternal>(SB),(NOSPLIT|WRAPPER),$312
	NO_LOCAL_POINTERS
	LEAQ	40(SP), R12
	CALL	runtime·spillArgs<ABIInternal>(SB)
	MOVQ	DX, 0(SP)
	MOVQ	R12, 8(SP)
	CALL	·moveMakeFuncArgPtrs(SB)
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)
	MOVB	$0, 32(SP)
	LEAQ	32(SP), AX
	MOVQ	AX, 16(SP)
	LEAQ	40(SP), AX
	MOVQ	AX, 24(SP)
	CALL	·callReflect<ABIInternal>(SB)
	LEAQ	40(SP), R12
	CALL	runtime·unspillArgs<ABIInternal>(SB)
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
// methodValueCall must be ABIInternal because it is placed directly
// in function values.
// The frame contains a stack-allocated RegArgs after the space saved for
// arguments to callMethod. This RegArgs is visible to the GC via a special
// case in the runtime's stack map code, so its offset is known to the runtime.
//
// TODO(mknyszek): Modify this function to call callMethod with
// ABIInternal.
TEXT ·methodValueCall<ABIInternal>(SB),(NOSPLIT|WRAPPER),$312
	NO_LOCAL_POINTERS
	LEAQ	40(SP), R12
	CALL	runtime·spillArgs<ABIInternal>(SB)
	MOVQ	DX, 0(SP)
	MOVQ	R12, 8(SP)
	CALL	·moveMakeFuncArgPtrs(SB)
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)
	MOVB	$0, 32(SP)
	LEAQ	32(SP), AX
	MOVQ	AX, 16(SP)
	LEAQ	40(SP), AX
	MOVQ	AX, 24(SP)
	CALL	·callMethod<ABIInternal>(SB)
	LEAQ	40(SP), R12
	CALL	runtime·unspillArgs<ABIInternal>(SB)
	RET
