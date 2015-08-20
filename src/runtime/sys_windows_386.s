// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// void runtime·asmstdcall(void *c);
TEXT runtime·asmstdcall(SB),NOSPLIT,$0
	MOVL	fn+0(FP), BX

	// SetLastError(0).
	MOVL	$0, 0x34(FS)

	// Copy args to the stack.
	MOVL	SP, BP
	MOVL	libcall_n(BX), CX	// words
	MOVL	CX, AX
	SALL	$2, AX
	SUBL	AX, SP			// room for args
	MOVL	SP, DI
	MOVL	libcall_args(BX), SI
	CLD
	REP; MOVSL

	// Call stdcall or cdecl function.
	// DI SI BP BX are preserved, SP is not
	CALL	libcall_fn(BX)
	MOVL	BP, SP

	// Return result.
	MOVL	fn+0(FP), BX
	MOVL	AX, libcall_r1(BX)
	MOVL	DX, libcall_r2(BX)

	// GetLastError().
	MOVL	0x34(FS), AX
	MOVL	AX, libcall_err(BX)

	RET

TEXT	runtime·badsignal2(SB),NOSPLIT,$24
	// stderr
	MOVL	$-12, 0(SP)
	MOVL	SP, BP
	CALL	*runtime·_GetStdHandle(SB)
	MOVL	BP, SP

	MOVL	AX, 0(SP)	// handle
	MOVL	$runtime·badsignalmsg(SB), DX // pointer
	MOVL	DX, 4(SP)
	MOVL	runtime·badsignallen(SB), DX // count
	MOVL	DX, 8(SP)
	LEAL	20(SP), DX  // written count
	MOVL	$0, 0(DX)
	MOVL	DX, 12(SP)
	MOVL	$0, 16(SP) // overlapped
	CALL	*runtime·_WriteFile(SB)
	MOVL	BP, SI
	RET

// faster get/set last error
TEXT runtime·getlasterror(SB),NOSPLIT,$0
	MOVL	0x34(FS), AX
	MOVL	AX, ret+0(FP)
	RET

TEXT runtime·setlasterror(SB),NOSPLIT,$0
	MOVL	err+0(FP), AX
	MOVL	AX, 0x34(FS)
	RET

// Called by Windows as a Vectored Exception Handler (VEH).
// First argument is pointer to struct containing
// exception record and context pointers.
// Handler function is stored in AX.
// Return 0 for 'not handled', -1 for handled.
TEXT runtime∕internal∕base·sigtramp(SB),NOSPLIT,$0-0
	MOVL	ptrs+0(FP), CX
	SUBL	$40, SP

	// Save callee-saved registers
	MOVL	BX, 28(SP)
	MOVL	BP, 16(SP)
	MOVL	SI, 20(SP)
	MOVL	DI, 24(SP)

	MOVL	AX, SI	// Save handler address

	// find g
	get_tls(DX)
	CMPL	DX, $0
	JNE	3(PC)
	MOVL	$0, AX // continue
	JMP	done
	MOVL	g(DX), DX
	CMPL	DX, $0
	JNE	2(PC)
	CALL	runtime·badsignal2(SB)

	// Save g and SP in case of stack switch
	MOVL	DX, 32(SP)	// g
	MOVL	SP, 36(SP)

	// do we need to switch to the g0 stack?
	MOVL	G_M(DX), BX
	MOVL	M_G0(BX), BX
	CMPL	DX, BX
	JEQ	g0

	// switch to the g0 stack
	get_tls(BP)
	MOVL	BX, g(BP)
	MOVL	(G_Sched+Gobuf_Sp)(BX), DI
	// make it look like Mstart called us on g0, to stop Traceback
	SUBL	$4, DI
	MOVL	$runtime∕internal∕base·Mstart(SB), 0(DI)
	// Traceback will think that we've done SUBL
	// on this stack, so subtract them here to match.
	// (we need room for Sighandler arguments anyway).
	// and re-Save old SP for restoring later.
	SUBL	$40, DI
	MOVL	SP, 36(DI)
	MOVL	DI, SP

g0:
	MOVL	0(CX), BX // ExceptionRecord*
	MOVL	4(CX), CX // Context*
	MOVL	BX, 0(SP)
	MOVL	CX, 4(SP)
	MOVL	DX, 8(SP)
	CALL	SI	// call handler
	// AX is set to report result back to Windows
	MOVL	12(SP), AX

	// switch back to original stack and g
	// no-op if we never left.
	MOVL	36(SP), SP
	MOVL	32(SP), DX
	get_tls(BP)
	MOVL	DX, g(BP)

done:
	// restore callee-saved registers
	MOVL	24(SP), DI
	MOVL	20(SP), SI
	MOVL	16(SP), BP
	MOVL	28(SP), BX

	ADDL	$40, SP
	// RET 4 (return and pop 4 bytes parameters)
	BYTE $0xC2; WORD $4
	RET // unreached; make assembler happy
 
TEXT runtime·exceptiontramp(SB),NOSPLIT,$0
	MOVL	$runtime·exceptionhandler(SB), AX
	JMP	runtime∕internal∕base·sigtramp(SB)

TEXT runtime·firstcontinuetramp(SB),NOSPLIT,$0-0
	// is never called
	INT	$3

TEXT runtime·lastcontinuetramp(SB),NOSPLIT,$0-0
	MOVL	$runtime·lastcontinuehandler(SB), AX
	JMP	runtime∕internal∕base·sigtramp(SB)

TEXT runtime·ctrlhandler(SB),NOSPLIT,$0
	PUSHL	$runtime·ctrlhandler1(SB)
	CALL	runtime·externalthreadhandler(SB)
	MOVL	4(SP), CX
	ADDL	$12, SP
	JMP	CX

TEXT runtime·profileloop(SB),NOSPLIT,$0
	PUSHL	$runtime·profileloop1(SB)
	CALL	runtime·externalthreadhandler(SB)
	MOVL	4(SP), CX
	ADDL	$12, SP
	JMP	CX

TEXT runtime·externalthreadhandler(SB),NOSPLIT,$0
	PUSHL	BP
	MOVL	SP, BP
	PUSHL	BX
	PUSHL	SI
	PUSHL	DI
	PUSHL	0x14(FS)
	MOVL	SP, DX

	// setup dummy m, g
	SUBL	$m__size, SP		// space for M
	MOVL	SP, 0(SP)
	MOVL	$m__size, 4(SP)
	CALL	runtime∕internal∕base·Memclr(SB)	// smashes AX,BX,CX

	LEAL	M_tls(SP), CX
	MOVL	CX, 0x14(FS)
	MOVL	SP, BX
	SUBL	$g__size, SP		// space for G
	MOVL	SP, g(CX)
	MOVL	SP, M_G0(BX)

	MOVL	SP, 0(SP)
	MOVL	$g__size, 4(SP)
	CALL	runtime∕internal∕base·Memclr(SB)	// smashes AX,BX,CX
	LEAL	g__size(SP), BX
	MOVL	BX, G_M(SP)
	LEAL	-8192(SP), CX
	MOVL	CX, (G_Stack+Stack_Lo)(SP)
	ADDL	$const_StackGuard, CX
	MOVL	CX, G_Stackguard0(SP)
	MOVL	CX, G_stackguard1(SP)
	MOVL	DX, (G_Stack+Stack_Hi)(SP)

	PUSHL	AX			// room for return value
	PUSHL	16(BP)			// arg for handler
	CALL	8(BP)
	POPL	CX
	POPL	AX			// pass return value to Windows in AX

	get_tls(CX)
	MOVL	g(CX), CX
	MOVL	(G_Stack+Stack_Hi)(CX), SP
	POPL	0x14(FS)
	POPL	DI
	POPL	SI
	POPL	BX
	POPL	BP
	RET

GLOBL runtime·cbctxts(SB), NOPTR, $4

TEXT runtime·callbackasm1+0(SB),NOSPLIT,$0
  	MOVL	0(SP), AX	// will use to find our callback context

	// remove return address from stack, we are not returning there
	ADDL	$4, SP

	// address to callback parameters into CX
	LEAL	4(SP), CX

	// Save registers as required for windows callback
	PUSHL	DI
	PUSHL	SI
	PUSHL	BP
	PUSHL	BX

	// determine Index into runtime·cbctxts table
	SUBL	$runtime·callbackasm(SB), AX
	MOVL	$0, DX
	MOVL	$5, BX	// divide by 5 because each call instruction in runtime·callbacks is 5 bytes long
	DIVL	BX

	// find correspondent runtime·cbctxts table entry
	MOVL	runtime·cbctxts(SB), BX
	MOVL	-4(BX)(AX*4), BX

	// extract callback context
	MOVL	wincallbackcontext_gobody(BX), AX
	MOVL	wincallbackcontext_argsize(BX), DX

	// preserve whatever's at the memory location that
	// the callback will use to store the return value
	PUSHL	0(CX)(DX*1)

	// extend argsize by size of return value
	ADDL	$4, DX

	// remember how to restore stack on return
	MOVL	wincallbackcontext_restorestack(BX), BX
	PUSHL	BX

	// call target Go function
	PUSHL	DX			// argsize (including return value)
	PUSHL	CX			// callback parameters
	PUSHL	AX			// address of target Go function
	CLD
	CALL	runtime·cgocallback_gofunc(SB)
	POPL	AX
	POPL	CX
	POPL	DX

	// how to restore stack on return
	POPL	BX

	// return value into AX (as per Windows spec)
	// and restore previously preserved value
	MOVL	-4(CX)(DX*1), AX
	POPL	-4(CX)(DX*1)

	MOVL	BX, CX			// cannot use BX anymore

	// restore registers as required for windows callback
	POPL	BX
	POPL	BP
	POPL	SI
	POPL	DI

	// remove callback parameters before return (as per Windows spec)
	POPL	DX
	ADDL	CX, SP
	PUSHL	DX

	CLD

	RET

// void tstart(M *Newm);
TEXT runtime·tstart(SB),NOSPLIT,$0
	MOVL	Newm+4(SP), CX		// m
	MOVL	M_G0(CX), DX		// g

	// Layout new m scheduler stack on os stack.
	MOVL	SP, AX
	MOVL	AX, (G_Stack+Stack_Hi)(DX)
	SUBL	$(64*1024), AX		// stack size
	MOVL	AX, (G_Stack+Stack_Lo)(DX)
	ADDL	$const_StackGuard, AX
	MOVL	AX, G_Stackguard0(DX)
	MOVL	AX, G_stackguard1(DX)

	// Set up tls.
	LEAL	M_tls(CX), SI
	MOVL	SI, 0x14(FS)
	MOVL	CX, G_M(DX)
	MOVL	DX, g(SI)

	// Someday the convention will be D is always cleared.
	CLD

	CALL	runtime·stackcheck(SB)	// clobbers AX,CX
	CALL	runtime∕internal∕base·Mstart(SB)

	RET

// uint32 tstart_stdcall(M *Newm);
TEXT runtime·tstart_stdcall(SB),NOSPLIT,$0
	MOVL	Newm+4(SP), BX

	PUSHL	BX
	CALL	runtime·tstart(SB)
	POPL	BX

	// Adjust stack for stdcall to return properly.
	MOVL	(SP), AX		// Save return address
	ADDL	$4, SP			// remove single parameter
	MOVL	AX, (SP)		// restore return address

	XORL	AX, AX			// return 0 == success

	RET

// setldt(int entry, int address, int limit)
TEXT runtime·setldt(SB),NOSPLIT,$0
	MOVL	address+4(FP), CX
	MOVL	CX, 0x14(FS)
	RET

// Sleep duration is in 100ns units.
TEXT runtime·usleep1(SB),NOSPLIT,$0
	MOVL	usec+0(FP), BX
	MOVL	$runtime·usleep2(SB), AX // to hide from 8l

	// Execute call on m->g0 stack, in case we are not actually
	// calling a system call wrapper, like when running under WINE.
	get_tls(CX)
	CMPL	CX, $0
	JNE	3(PC)
	// Not a Go-managed thread. Do not switch stack.
	CALL	AX
	RET

	MOVL	g(CX), BP
	MOVL	G_M(BP), BP

	// leave pc/sp for cpu profiler
	MOVL	(SP), SI
	MOVL	SI, M_libcallpc(BP)
	MOVL	g(CX), SI
	MOVL	SI, M_libcallg(BP)
	// sp must be the last, because once async cpu profiler finds
	// all three values to be non-zero, it will use them
	LEAL	usec+0(FP), SI
	MOVL	SI, M_Libcallsp(BP)

	MOVL	M_G0(BP), SI
	CMPL	g(CX), SI
	JNE	switch
	// executing on m->g0 already
	CALL	AX
	JMP	ret

switch:
	// Switch to m->g0 stack and back.
	MOVL	(G_Sched+Gobuf_Sp)(SI), SI
	MOVL	SP, -4(SI)
	LEAL	-4(SI), SP
	CALL	AX
	MOVL	0(SP), SP

ret:
	get_tls(CX)
	MOVL	g(CX), BP
	MOVL	G_M(BP), BP
	MOVL	$0, M_Libcallsp(BP)
	RET

// Runs on OS stack. duration (in 100ns units) is in BX.
TEXT runtime·usleep2(SB),NOSPLIT,$20
	// Want negative 100ns units.
	NEGL	BX
	MOVL	$-1, hi-4(SP)
	MOVL	BX, lo-8(SP)
	LEAL	lo-8(SP), BX
	MOVL	BX, ptime-12(SP)
	MOVL	$0, alertable-16(SP)
	MOVL	$-1, handle-20(SP)
	MOVL	SP, BP
	MOVL	runtime·_NtWaitForSingleObject(SB), AX
	CALL	AX
	MOVL	BP, SP
	RET

// func now() (sec int64, nsec int32)
TEXT time·now(SB),NOSPLIT,$8-12
	CALL	runtime·unixnano(SB)
	MOVL	0(SP), AX
	MOVL	4(SP), DX

	MOVL	$1000000000, CX
	DIVL	CX
	MOVL	AX, sec+0(FP)
	MOVL	$0, sec+4(FP)
	MOVL	DX, nsec+8(FP)
	RET
