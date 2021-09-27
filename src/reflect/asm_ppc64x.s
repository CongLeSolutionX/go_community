// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64 || ppc64le
// +build ppc64 ppc64le

#include "textflag.h"
#include "funcdata.h"
#include "asm_ppc64x.h"

// FIXED_FRAME+32 is a save spot
// FIXED_FRAME+40 is return information
// FIXED_FRAME+48 is start of register args

#define SAVER11 FIXED_FRAME
#define LOCAL_RETVALID 40+FIXED_FRAME
#define LOCAL_REGARGS 48+FIXED_FRAME

// The frame size of the functions below is
// 32 fixed frame + 32 (args of callReflect) + 8 (bool + padding) + 288 (abi.RegArgs) = 360.

// makeFuncStub is the code half of the function returned by MakeFunc.
// See the comment on the declaration of makeFuncStub in makefunc.go
// for more details.
// No arg size here, runtime pulls arg map out of the func value.
TEXT ·makeFuncStub(SB),(NOSPLIT|WRAPPER),$360
	NO_LOCAL_POINTERS
	ADD	$LOCAL_REGARGS, R1, R20
	CALL	runtime·spillArgs(SB)
	MOVD	R11, FIXED_FRAME+32(R1)	// TODO: determine offset to save
	MOVD	R11, FIXED_FRAME+0(R1)	// arg for moveMakeFuncArgPtrs
	MOVD	R20, FIXED_FRAME+8(R1)	// arg for local args
	CALL	·moveMakeFuncArgPtrs(SB)
	MOVD	FIXED_FRAME+32(R1), R11	// restore R11 ctxt
	// previous
	MOVD	R11, FIXED_FRAME+0(R1)	// ctxt (arg0)
	MOVD	$argframe+0(FP), R3	// ??
	MOVD	R3, FIXED_FRAME+8(R1)	// frame (arg1)
	ADD	$LOCAL_RETVALID, R1, R3 // addr of return flag
	MOVB	R0, (R3)		// clear flag
	MOVD	R3, FIXED_FRAME+16(R1)	// addr retvalid (arg2)
	ADD     $LOCAL_REGARGS, R1, R3
	MOVD	R3, FIXED_FRAME+24(R1)	// abiregargs (arg3)
	BL	·callReflect(SB)
	ADD	$LOCAL_REGARGS, R1, R20	// set address of spill area
	CALL	runtime·unspillArgs(SB)
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$360
	NO_LOCAL_POINTERS
	ADD	$LOCAL_REGARGS, R1, R20
	CALL	runtime·spillArgs(SB)
	MOVD	R11, FIXED_FRAME+0(R1) // arg0 ctxt
	MOVD	R11, FIXED_FRAME+32(R1) // save for later
	MOVD	R20, FIXED_FRAME+8(R1) // arg1 abiregargs
	CALL	·moveMakeFuncArgPtrs(SB)
	MOVD	FIXED_FRAME+32(R1), R11 // restore ctxt
	MOVD	R11, FIXED_FRAME+0(R1) // set as arg0
	MOVD	$argframe+0(FP), R3	// frame pointer
	MOVD	R3, FIXED_FRAME+8(R1)	// set as arg1
	ADD	$LOCAL_RETVALID, R1, R3
	MOVB	$0, (R3)		// clear ret flag
	MOVD	R3, FIXED_FRAME+16(R1)	// addr of return flag
	ADD	$LOCAL_REGARGS, R1, R3	// addr of abiregargs
	MOVD	R3, FIXED_FRAME+24(R1)	// set as arg3
	BL	·callMethod(SB)
	ADD     $LOCAL_REGARGS, R1, R20
	CALL	runtime·unspillArgs(SB)
	RET
