// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "Funcdata.h"
#include "textflag.h"

// using frame size $-4 means do not Save LR on stack.
TEXT runtime∕internal∕schedinit·rt0_go(SB),NOSPLIT,$-4
	MOVW	$0xcafebabe, R12

	// copy arguments forward on an even stack
	// use R13 instead of SP to avoid linker rewriting the offsets
	MOVW	0(R13), R0		// Argc
	MOVW	4(R13), R1		// Argv
	SUB	$64, R13		// plenty of scratch
	AND	$~7, R13
	MOVW	R0, 60(R13)		// Save Argc, Argv away
	MOVW	R1, 64(R13)

	// set up g register
	// g is R10
	MOVW	$runtime∕internal∕core·g0(SB), g
	MOVW	$runtime∕internal∕core·M0(SB), R8

	// Save m->g0 = g0
	MOVW	g, M_G0(R8)
	// Save g->m = M0
	MOVW	R8, G_M(g)

	// create istack out of the OS stack
	MOVW	$(-8192+104)(R13), R0
	MOVW	R0, G_Stackguard0(g)
	MOVW	R0, G_Stackguard1(g)
	MOVW	R0, (G_Stack+Stack_Lo)(g)
	MOVW	R13, (G_Stack+Stack_Hi)(g)

	BL	runtime·emptyfunc(SB)	// fault if stack check is wrong

#ifndef GOOS_nacl
	// if there is an _cgo_init, call it.
	MOVW	_cgo_init(SB), R4
	CMP	$0, R4
	B.EQ	nocgo
	MRC     15, 0, R0, C13, C0, 3 	// load TLS base pointer
	MOVW 	R0, R3 			// arg 3: TLS base pointer
	MOVW 	$runtime·tlsg(SB), R2 	// arg 2: tlsg
	MOVW	$setg_gcc<>(SB), R1 	// arg 1: Setg
	MOVW	g, R0 			// arg 0: G
	BL	(R4) // will clobber R0-R3
#endif

nocgo:
	// update stackguard after _cgo_init
	MOVW	(G_Stack+Stack_Lo)(g), R0
	ADD	$const_StackGuard, R0
	MOVW	R0, G_Stackguard0(g)
	MOVW	R0, G_Stackguard1(g)

	BL	runtime·checkgoarm(SB)
	BL	runtime∕internal∕check·check(SB)

	// saved Argc, Argv
	MOVW	60(R13), R0
	MOVW	R0, 4(R13)
	MOVW	64(R13), R1
	MOVW	R1, 8(R13)
	BL	runtime∕internal∕vdso·args(SB)
	BL	runtime·osinit(SB)
	BL	runtime∕internal∕schedinit·schedinit(SB)

	// create a new goroutine to start program
	MOVW	$runtime·main·f(SB), R0
	MOVW.W	R0, -4(R13)
	MOVW	$8, R0
	MOVW.W	R0, -4(R13)
	MOVW	$0, R0
	MOVW.W	R0, -4(R13)	// push $0 as guard
	BL	runtime·newproc(SB)
	MOVW	$12(R13), R13	// pop args and LR

	// start this M
	BL	runtime∕internal∕sched·Mstart(SB)

	MOVW	$1234, R0
	MOVW	$1000, R1
	MOVW	R0, (R1)	// fail hard

DATA	runtime·main·f+0(SB)/4,$runtime·main(SB)
GLOBL	runtime·main·f(SB),RODATA,$4

TEXT runtime·breakpoint(SB),NOSPLIT,$0-0
	// gdb won't skip this breakpoint instruction automatically,
	// so you must manually "set $pc+=4" to skip it and continue.
#ifdef GOOS_nacl
	WORD	$0xe125be7f	// BKPT 0x5bef, NACL_INSTR_ARM_BREAKPOINT
#else
	WORD	$0xe7f001f0	// undefined instruction that gdb understands is a software breakpoint
#endif
	RET

TEXT runtime∕internal∕core·Asminit(SB),NOSPLIT,$0-0
	// disable runfast (flush-to-zero) mode of vfp if runtime.goarm > 5
	MOVB	runtime·goarm(SB), R11
	CMP	$5, R11
	BLE	4(PC)
	WORD	$0xeef1ba10	// vmrs r11, fpscr
	BIC	$(1<<24), R11
	WORD	$0xeee1ba10	// vmsr fpscr, r11
	RET

/*
 *  go-routine
 */

// void gosave(Gobuf*)
// Save state in Gobuf; setjmp
TEXT runtime∕internal∕sched·gosave(SB),NOSPLIT,$-4-4
	MOVW	0(FP), R0		// gobuf
	MOVW	SP, Gobuf_Sp(R0)
	MOVW	LR, Gobuf_Pc(R0)
	MOVW	g, Gobuf_G(R0)
	MOVW	$0, R11
	MOVW	R11, Gobuf_Lr(R0)
	MOVW	R11, Gobuf_Ret(R0)
	MOVW	R11, Gobuf_Ctxt(R0)
	RET

// void Gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT runtime∕internal∕sched·Gogo(SB),NOSPLIT,$-4-4
	MOVW	0(FP), R1		// gobuf
	MOVW	Gobuf_G(R1), R0
	BL	Setg<>(SB)

	// NOTE: We updated g above, and we are about to update SP.
	// Until LR and PC are also updated, the g/SP/LR/PC quadruple
	// are out of sync and must not be used as the basis of a Traceback.
	// Sigprof skips the Traceback when SP is not within g's bounds,
	// and when the PC is inside this function, runtime.Gogo.
	// Since we are about to update SP, until we complete runtime.Gogo
	// we must not leave this function. In particular, no calls
	// after this point: it must be straight-line code until the
	// final B instruction.
	// See large comment in sigprof for more details.
	MOVW	Gobuf_Sp(R1), SP	// restore SP
	MOVW	Gobuf_Lr(R1), LR
	MOVW	Gobuf_Ret(R1), R0
	MOVW	Gobuf_Ctxt(R1), R7
	MOVW	$0, R11
	MOVW	R11, Gobuf_Sp(R1)	// clear to help garbage collector
	MOVW	R11, Gobuf_Ret(R1)
	MOVW	R11, Gobuf_Lr(R1)
	MOVW	R11, Gobuf_Ctxt(R1)
	MOVW	Gobuf_Pc(R1), R11
	CMP	R11, R11 // set condition codes for == test, needed by stack split
	B	(R11)

// func Mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return.  It should Gogo(&g->Sched)
// to keep running g.
TEXT runtime∕internal∕sched·Mcall(SB),NOSPLIT,$-4-4
	// Save caller state in g->Sched.
	MOVW	SP, (G_Sched+Gobuf_Sp)(g)
	MOVW	LR, (G_Sched+Gobuf_Pc)(g)
	MOVW	$0, R11
	MOVW	R11, (G_Sched+Gobuf_Lr)(g)
	MOVW	g, (G_Sched+Gobuf_G)(g)

	// Switch to m->g0 & its stack, call fn.
	MOVW	g, R1
	MOVW	G_M(g), R8
	MOVW	M_G0(R8), R0
	BL	Setg<>(SB)
	CMP	g, R1
	B.NE	2(PC)
	B	runtime·badmcall(SB)
	MOVB	runtime∕internal∕sched·Iscgo(SB), R11
	CMP	$0, R11
	BL.NE	runtime·save_g(SB)
	MOVW	fn+0(FP), R0
	MOVW	(G_Sched+Gobuf_Sp)(g), SP
	SUB	$8, SP
	MOVW	R1, 4(SP)
	MOVW	R0, R7
	MOVW	0(R0), R0
	BL	(R0)
	B	runtime·badmcall2(SB)
	RET

// systemstack_switch is a dummy routine that Systemstack leaves at the bottom
// of the G stack.  We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime∕internal∕schedinit·systemstack_switch(SB),NOSPLIT,$0-0
	MOVW	$0, R0
	BL	(R0) // clobber lr to ensure push {lr} is kept
	RET

// func Systemstack(fn func())
TEXT runtime∕internal∕lock·Systemstack(SB),NOSPLIT,$0-4
	MOVW	fn+0(FP), R0	// R0 = fn
	MOVW	G_M(g), R1	// R1 = m

	MOVW	M_Gsignal(R1), R2	// R2 = gsignal
	CMP	g, R2
	B.EQ	noswitch

	MOVW	M_G0(R1), R2	// R2 = g0
	CMP	g, R2
	B.EQ	noswitch

	MOVW	M_Curg(R1), R3
	CMP	g, R3
	B.EQ	switch

	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOVW	$runtime·badsystemstack(SB), R0
	BL	(R0)

switch:
	// Save our state in g->Sched.  Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVW	$runtime∕internal∕schedinit·systemstack_switch(SB), R3
	ADD	$4, R3, R3 // get past push {lr}
	MOVW	R3, (G_Sched+Gobuf_Pc)(g)
	MOVW	SP, (G_Sched+Gobuf_Sp)(g)
	MOVW	LR, (G_Sched+Gobuf_Lr)(g)
	MOVW	g, (G_Sched+Gobuf_G)(g)

	// switch to g0
	MOVW	R0, R5
	MOVW	R2, R0
	BL	Setg<>(SB)
	MOVW	R5, R0
	MOVW	(G_Sched+Gobuf_Sp)(R2), R3
	// make it look like Mstart called Systemstack on g0, to stop Traceback
	SUB	$4, R3, R3
	MOVW	$runtime∕internal∕sched·Mstart(SB), R4
	MOVW	R4, 0(R3)
	MOVW	R3, SP

	// call target function
	MOVW	R0, R7
	MOVW	0(R0), R0
	BL	(R0)

	// switch back to g
	MOVW	G_M(g), R1
	MOVW	M_Curg(R1), R0
	BL	Setg<>(SB)
	MOVW	(G_Sched+Gobuf_Sp)(g), SP
	MOVW	$0, R3
	MOVW	R3, (G_Sched+Gobuf_Sp)(g)
	RET

noswitch:
	MOVW	R0, R7
	MOVW	0(R0), R0
	BL	(R0)
	RET

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
// R1 frame size
// R2 arg size
// R3 prolog's LR
// NB. we do not Save R0 because we've forced 5c to pass all arguments
// on the stack.
// using frame size $-4 means do not Save LR on stack.
//
// The Traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime∕internal∕schedinit·morestack(SB),NOSPLIT,$-4-0
	// Cannot grow scheduler stack (m->g0).
	MOVW	G_M(g), R8
	MOVW	M_G0(R8), R4
	CMP	g, R4
	BL.EQ	runtime·abort(SB)

	// Cannot grow signal stack (m->gsignal).
	MOVW	M_Gsignal(R8), R4
	CMP	g, R4
	BL.EQ	runtime·abort(SB)

	// Called from f.
	// Set g->Sched to context in f.
	MOVW	R7, (G_Sched+Gobuf_Ctxt)(g)
	MOVW	SP, (G_Sched+Gobuf_Sp)(g)
	MOVW	LR, (G_Sched+Gobuf_Pc)(g)
	MOVW	R3, (G_Sched+Gobuf_Lr)(g)

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVW	R3, (M_Morebuf+Gobuf_Pc)(R8)	// f's caller's PC
	MOVW	SP, (M_Morebuf+Gobuf_Sp)(R8)	// f's caller's SP
	MOVW	$4(SP), R3			// f's argument pointer
	MOVW	g, (M_Morebuf+Gobuf_G)(R8)

	// Call newstack on m->g0's stack.
	MOVW	M_G0(R8), R0
	BL	Setg<>(SB)
	MOVW	(G_Sched+Gobuf_Sp)(g), SP
	BL	runtime·newstack(SB)

	// Not reached, but make sure the return PC from the call to newstack
	// is still in this function, and not the beginning of the next.
	RET

TEXT runtime·morestack_noctxt(SB),NOSPLIT,$-4-0
	MOVW	$0, R7
	B runtime∕internal∕schedinit·morestack(SB)

// Reflectcall: call a function with the given argument list
// func call(f *FuncVal, arg *byte, argsize, retoffset uint32).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)		\
	CMP	$MAXSIZE, R0;		\
	B.HI	3(PC);			\
	MOVW	$NAME(SB), R1;		\
	B	(R1)

TEXT reflect·call(SB), NOSPLIT, $0-0
	B	runtime∕internal∕finalize·Reflectcall(SB)

TEXT runtime∕internal∕finalize·Reflectcall(SB),NOSPLIT,$-4-16
	MOVW	argsize+8(FP), R0
	DISPATCH(runtime·call16, 16)
	DISPATCH(runtime·call32, 32)
	DISPATCH(runtime·call64, 64)
	DISPATCH(runtime·call128, 128)
	DISPATCH(runtime·call256, 256)
	DISPATCH(runtime·call512, 512)
	DISPATCH(runtime·call1024, 1024)
	DISPATCH(runtime·call2048, 2048)
	DISPATCH(runtime·call4096, 4096)
	DISPATCH(runtime·call8192, 8192)
	DISPATCH(runtime·call16384, 16384)
	DISPATCH(runtime·call32768, 32768)
	DISPATCH(runtime·call65536, 65536)
	DISPATCH(runtime·call131072, 131072)
	DISPATCH(runtime·call262144, 262144)
	DISPATCH(runtime·call524288, 524288)
	DISPATCH(runtime·call1048576, 1048576)
	DISPATCH(runtime·call2097152, 2097152)
	DISPATCH(runtime·call4194304, 4194304)
	DISPATCH(runtime·call8388608, 8388608)
	DISPATCH(runtime·call16777216, 16777216)
	DISPATCH(runtime·call33554432, 33554432)
	DISPATCH(runtime·call67108864, 67108864)
	DISPATCH(runtime·call134217728, 134217728)
	DISPATCH(runtime·call268435456, 268435456)
	DISPATCH(runtime·call536870912, 536870912)
	DISPATCH(runtime·call1073741824, 1073741824)
	MOVW	$runtime·badreflectcall(SB), R1
	B	(R1)

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-16;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVW	argptr+4(FP), R0;		\
	MOVW	argsize+8(FP), R2;		\
	ADD	$4, SP, R1;			\
	CMP	$0, R2;				\
	B.EQ	5(PC);				\
	MOVBU.P	1(R0), R5;			\
	MOVBU.P R5, 1(R1);			\
	SUB	$1, R2, R2;			\
	B	-5(PC);				\
	/* call function */			\
	MOVW	f+0(FP), R7;			\
	MOVW	(R7), R0;			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	BL	(R0);				\
	/* copy return values back */		\
	MOVW	argptr+4(FP), R0;		\
	MOVW	argsize+8(FP), R2;		\
	MOVW	retoffset+12(FP), R3;		\
	ADD	$4, SP, R1;			\
	ADD	R3, R1;				\
	ADD	R3, R0;				\
	SUB	R3, R2;				\
	CMP	$0, R2;				\
	RET.EQ	;				\
	MOVBU.P	1(R1), R5;			\
	MOVBU.P R5, 1(R0);			\
	SUB	$1, R2, R2;			\
	B	-5(PC)				\

CALLFN(runtime·call16, 16)
CALLFN(runtime·call32, 32)
CALLFN(runtime·call64, 64)
CALLFN(runtime·call128, 128)
CALLFN(runtime·call256, 256)
CALLFN(runtime·call512, 512)
CALLFN(runtime·call1024, 1024)
CALLFN(runtime·call2048, 2048)
CALLFN(runtime·call4096, 4096)
CALLFN(runtime·call8192, 8192)
CALLFN(runtime·call16384, 16384)
CALLFN(runtime·call32768, 32768)
CALLFN(runtime·call65536, 65536)
CALLFN(runtime·call131072, 131072)
CALLFN(runtime·call262144, 262144)
CALLFN(runtime·call524288, 524288)
CALLFN(runtime·call1048576, 1048576)
CALLFN(runtime·call2097152, 2097152)
CALLFN(runtime·call4194304, 4194304)
CALLFN(runtime·call8388608, 8388608)
CALLFN(runtime·call16777216, 16777216)
CALLFN(runtime·call33554432, 33554432)
CALLFN(runtime·call67108864, 67108864)
CALLFN(runtime·call134217728, 134217728)
CALLFN(runtime·call268435456, 268435456)
CALLFN(runtime·call536870912, 536870912)
CALLFN(runtime·call1073741824, 1073741824)

// void Jmpdefer(fn, sp);
// called from deferreturn.
// 1. grab stored LR for caller
// 2. sub 4 Bytes to get back to BL deferreturn
// 3. B to fn
// TODO(rsc): Push things on stack and then use pop
// to load all registers simultaneously, so that a profiling
// interrupt can never see mismatched SP/LR/PC.
// (And double-check that pop is atomic in that way.)
TEXT runtime∕internal∕schedinit·Jmpdefer(SB),NOSPLIT,$0-8
	MOVW	0(SP), LR
	MOVW	$-4(LR), LR	// BL deferreturn
	MOVW	fv+0(FP), R7
	MOVW	argp+4(FP), SP
	MOVW	$-4(SP), SP	// SP is 4 below argp, due to saved LR
	MOVW	0(R7), R1
	B	(R1)

// Save state of caller into g->Sched. Smashes R11.
TEXT gosave<>(SB),NOSPLIT,$0
	MOVW	LR, (G_Sched+Gobuf_Pc)(g)
	MOVW	R13, (G_Sched+Gobuf_Sp)(g)
	MOVW	$0, R11
	MOVW	R11, (G_Sched+Gobuf_Lr)(g)
	MOVW	R11, (G_Sched+Gobuf_Ret)(g)
	MOVW	R11, (G_Sched+Gobuf_Ctxt)(g)
	RET

// Asmcgocall(void(*fn)(void*), void *arg)
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.c for more details.
TEXT	runtime∕internal∕sched·Asmcgocall(SB),NOSPLIT,$0-8
	MOVW	fn+0(FP), R1
	MOVW	arg+4(FP), R0
	BL	Asmcgocall<>(SB)
	RET

TEXT runtime∕internal∕cgo·asmcgocall_errno(SB),NOSPLIT,$0-12
	MOVW	fn+0(FP), R1
	MOVW	arg+4(FP), R0
	BL	Asmcgocall<>(SB)
	MOVW	R0, ret+8(FP)
	RET

TEXT Asmcgocall<>(SB),NOSPLIT,$0-0
	// fn in R1, arg in R0.
	MOVW	R13, R2
	MOVW	g, R4

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already.
	MOVW	G_M(g), R8
	MOVW	M_G0(R8), R3
	CMP	R3, g
	BEQ	g0
	BL	gosave<>(SB)
	MOVW	R0, R5
	MOVW	R3, R0
	BL	Setg<>(SB)
	MOVW	R5, R0
	MOVW	(G_Sched+Gobuf_Sp)(g), R13

	// Now on a scheduling stack (a pthread-created stack).
g0:
	SUB	$24, R13
	BIC	$0x7, R13	// alignment for gcc ABI
	MOVW	R4, 20(R13) // Save old g
	MOVW	(G_Stack+Stack_Hi)(R4), R4
	SUB	R2, R4
	MOVW	R4, 16(R13)	// Save depth in stack (can't just Save SP, as stack might be copied during a callback)
	BL	(R1)

	// Restore registers, g, stack pointer.
	MOVW	R0, R5
	MOVW	20(R13), R0
	BL	Setg<>(SB)
	MOVW	(G_Stack+Stack_Hi)(g), R1
	MOVW	16(R13), R2
	SUB	R2, R1
	MOVW	R5, R0
	MOVW	R1, R13
	RET

// cgocallback(void (*fn)(void*), void *frame, uintptr framesize)
// Turn the fn into a Go func (by taking its address) and call
// cgocallback_gofunc.
TEXT runtime·cgocallback(SB),NOSPLIT,$12-12
	MOVW	$fn+0(FP), R0
	MOVW	R0, 4(R13)
	MOVW	frame+4(FP), R0
	MOVW	R0, 8(R13)
	MOVW	framesize+8(FP), R0
	MOVW	R0, 12(R13)
	MOVW	$runtime·cgocallback_gofunc(SB), R0
	BL	(R0)
	RET

// cgocallback_gofunc(void (*fn)(void*), void *frame, uintptr framesize)
// See cgocall.c for more details.
TEXT	runtime·cgocallback_gofunc(SB),NOSPLIT,$8-12
	NO_LOCAL_POINTERS
	
	// Load m and g from thread-local storage.
	MOVB	runtime∕internal∕sched·Iscgo(SB), R0
	CMP	$0, R0
	BL.NE	runtime·load_g(SB)

	// If g is nil, Go did not create the current thread.
	// Call needm to obtain one for temporary use.
	// In this case, we're running on the thread stack, so there's
	// lots of space, but the linker doesn't know. Hide the call from
	// the linker analysis by using an indirect call.
	CMP	$0, g
	B.NE	havem
	MOVW	g, savedm-4(SP) // g is zero, so is m.
	MOVW	$runtime∕internal∕core·needm(SB), R0
	BL	(R0)

	// Set m->Sched.sp = SP, so that if a panic happens
	// during the function we are about to execute, it will
	// have a valid SP to run on the g0 stack.
	// The next few lines (after the havem label)
	// will Save this SP onto the stack and then Write
	// the same SP back to m->Sched.sp. That seems redundant,
	// but if an unrecovered panic happens, unwindm will
	// restore the g->Sched.sp from the stack location
	// and then Systemstack will try to use it. If we don't set it here,
	// that restored SP will be uninitialized (typically 0) and
	// will not be usable.
	MOVW	G_M(g), R8
	MOVW	M_G0(R8), R3
	MOVW	R13, (G_Sched+Gobuf_Sp)(R3)

havem:
	MOVW	G_M(g), R8
	MOVW	R8, savedm-4(SP)
	// Now there's a valid m, and we're running on its m->g0.
	// Save current m->g0->Sched.sp on stack and then set it to SP.
	// Save current sp in m->g0->Sched.sp in preparation for
	// switch back to m->curg stack.
	// NOTE: unwindm knows that the saved g->Sched.sp is at 4(R13) aka savedsp-8(SP).
	MOVW	M_G0(R8), R3
	MOVW	(G_Sched+Gobuf_Sp)(R3), R4
	MOVW	R4, savedsp-8(SP)
	MOVW	R13, (G_Sched+Gobuf_Sp)(R3)

	// Switch to m->curg stack and call runtime.cgocallbackg.
	// Because we are taking over the execution of m->curg
	// but *not* resuming what had been running, we need to
	// Save that information (m->curg->Sched) so we can restore it.
	// We can restore m->curg->Sched.sp easily, because calling
	// runtime.cgocallbackg leaves SP unchanged upon return.
	// To Save m->curg->Sched.pc, we push it onto the stack.
	// This has the added benefit that it looks to the Traceback
	// routine like cgocallbackg is going to return to that
	// PC (because the frame we allocate below has the same
	// size as cgocallback_gofunc's frame declared above)
	// so that the Traceback will seamlessly trace back into
	// the earlier calls.
	//
	// In the new goroutine, -8(SP) and -4(SP) are unused.
	MOVW	M_Curg(R8), R0
	BL	Setg<>(SB)
	MOVW	(G_Sched+Gobuf_Sp)(g), R4 // prepare stack as R4
	MOVW	(G_Sched+Gobuf_Pc)(g), R5
	MOVW	R5, -12(R4)
	MOVW	$-12(R4), R13
	BL	runtime∕internal∕cgo·cgocallbackg(SB)

	// Restore g->Sched (== m->curg->Sched) from saved values.
	MOVW	0(R13), R5
	MOVW	R5, (G_Sched+Gobuf_Pc)(g)
	MOVW	$12(R13), R4
	MOVW	R4, (G_Sched+Gobuf_Sp)(g)

	// Switch back to m->g0's stack and restore m->g0->Sched.sp.
	// (Unlike m->curg, the g0 goroutine never uses Sched.pc,
	// so we do not have to restore it.)
	MOVW	G_M(g), R8
	MOVW	M_G0(R8), R0
	BL	Setg<>(SB)
	MOVW	(G_Sched+Gobuf_Sp)(g), R13
	MOVW	savedsp-8(SP), R4
	MOVW	R4, (G_Sched+Gobuf_Sp)(g)

	// If the m on entry was nil, we called needm above to borrow an m
	// for the duration of the call. Since the call is over, return it with dropm.
	MOVW	savedm-4(SP), R6
	CMP	$0, R6
	B.NE	3(PC)
	MOVW	$runtime·dropm(SB), R0
	BL	(R0)

	// Done!
	RET

// void Setg(G*); set g. for use by needm.
TEXT runtime∕internal∕core·Setg(SB),NOSPLIT,$-4-4
	MOVW	gg+0(FP), R0
	B	Setg<>(SB)

TEXT Setg<>(SB),NOSPLIT,$-4-0
	MOVW	R0, g

	// Save g to thread-local storage.
	MOVB	runtime∕internal∕sched·Iscgo(SB), R0
	CMP	$0, R0
	B.EQ	2(PC)
	B	runtime·save_g(SB)

	MOVW	g, R0
	RET

TEXT runtime∕internal∕lock·Getcallerpc(SB),NOSPLIT,$-4-4
	MOVW	0(SP), R0
	MOVW	R0, ret+4(FP)
	RET

TEXT runtime·gogetcallerpc(SB),NOSPLIT,$-4-8
	MOVW	R14, ret+4(FP)
	RET

TEXT runtime∕internal∕channels·setcallerpc(SB),NOSPLIT,$-4-8
	MOVW	pc+4(FP), R0
	MOVW	R0, 0(SP)
	RET

TEXT runtime∕internal∕lock·Getcallersp(SB),NOSPLIT,$-4-4
	MOVW	0(FP), R0
	MOVW	$-4(R0), R0
	MOVW	R0, ret+4(FP)
	RET

// func gogetcallersp(p unsafe.Pointer) uintptr
TEXT runtime·gogetcallersp(SB),NOSPLIT,$-4-8
	MOVW	0(FP), R0
	MOVW	$-4(R0), R0
	MOVW	R0, ret+4(FP)
	RET

TEXT runtime·emptyfunc(SB),0,$0-0
	RET

TEXT runtime·abort(SB),NOSPLIT,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1

// bool armcas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	}else
//		return 0;
//
// To implement runtime∕internal∕sched·Cas in sys_$GOOS_arm.s
// using the native instructions, use:
//
//	TEXT runtime∕internal∕sched·Cas(SB),NOSPLIT,$0
//		B	runtime·armcas(SB)
//
TEXT runtime·armcas(SB),NOSPLIT,$0-13
	MOVW	valptr+0(FP), R1
	MOVW	old+4(FP), R2
	MOVW	new+8(FP), R3
casl:
	LDREX	(R1), R0
	CMP	R0, R2
	BNE	casfail
	STREX	R3, (R1), R0
	CMP	$0, R0
	BNE	casl
	MOVW	$1, R0
	MOVB	R0, ret+12(FP)
	RET
casfail:
	MOVW	$0, R0
	MOVB	R0, ret+12(FP)
	RET

TEXT runtime∕internal∕core·Casuintptr(SB),NOSPLIT,$0-13
	B	runtime∕internal∕sched·Cas(SB)

TEXT runtime∕internal∕core·Atomicloaduintptr(SB),NOSPLIT,$0-8
	B	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕channels·atomicloaduint(SB),NOSPLIT,$0-8
	B	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕core·atomicstoreuintptr(SB),NOSPLIT,$0-8
	B	runtime∕internal∕lock·Atomicstore(SB)

// AES hashing not implemented for ARM
TEXT runtime∕internal∕hash·aeshash(SB),NOSPLIT,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshash32(SB),NOSPLIT,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshash64(SB),NOSPLIT,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshashstr(SB),NOSPLIT,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1

TEXT runtime∕internal∕hash·Memeq(SB),NOSPLIT,$-4-13
	MOVW	a+0(FP), R1
	MOVW	b+4(FP), R2
	MOVW	size+8(FP), R3
	ADD	R1, R3, R6
	MOVW	$1, R0
	MOVB	R0, ret+12(FP)
loop:
	CMP	R1, R6
	RET.EQ
	MOVBU.P	1(R1), R4
	MOVBU.P	1(R2), R5
	CMP	R4, R5
	BEQ	loop

	MOVW	$0, R0
	MOVB	R0, ret+12(FP)
	RET

// eqstring tests whether two strings are equal.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.
TEXT runtime·eqstring(SB),NOSPLIT,$-4-17
	MOVW	s1len+4(FP), R0
	MOVW	s2len+12(FP), R1
	MOVW	$0, R7
	CMP	R0, R1
	MOVB.NE R7, v+16(FP)
	RET.NE
	MOVW	s1str+0(FP), R2
	MOVW	s2str+8(FP), R3
	MOVW	$1, R8
	MOVB	R8, v+16(FP)
	CMP	R2, R3
	RET.EQ
	ADD	R2, R0, R6
loop:
	CMP	R2, R6
	RET.EQ
	MOVBU.P	1(R2), R4
	MOVBU.P	1(R3), R5
	CMP	R4, R5
	BEQ	loop
	MOVB	R7, v+16(FP)
	RET

// void setg_gcc(G*); set g called from gcc.
TEXT setg_gcc<>(SB),NOSPLIT,$0
	MOVW	R0, g
	B		runtime·save_g(SB)

// TODO: share code with Memeq?
TEXT bytes·Equal(SB),NOSPLIT,$0
	MOVW	a_len+4(FP), R1
	MOVW	b_len+16(FP), R3
	
	CMP	R1, R3		// unequal lengths are not equal
	B.NE	notequal

	MOVW	a+0(FP), R0
	MOVW	b+12(FP), R2
	ADD	R0, R1		// end

loop:
	CMP	R0, R1
	B.EQ	equal		// reached the end
	MOVBU.P	1(R0), R4
	MOVBU.P	1(R2), R5
	CMP	R4, R5
	B.EQ	loop

notequal:
	MOVW	$0, R0
	MOVBU	R0, ret+24(FP)
	RET

equal:
	MOVW	$1, R0
	MOVBU	R0, ret+24(FP)
	RET

TEXT bytes·IndexByte(SB),NOSPLIT,$0
	MOVW	s+0(FP), R0
	MOVW	s_len+4(FP), R1
	MOVBU	c+12(FP), R2	// byte to find
	MOVW	R0, R4		// store base for later
	ADD	R0, R1		// end 

_loop:
	CMP	R0, R1
	B.EQ	_notfound
	MOVBU.P	1(R0), R3
	CMP	R2, R3
	B.NE	_loop

	SUB	$1, R0		// R0 will be one beyond the position we want
	SUB	R4, R0		// remove base
	MOVW    R0, ret+16(FP) 
	RET

_notfound:
	MOVW	$-1, R0
	MOVW	R0, ret+16(FP)
	RET

TEXT strings·IndexByte(SB),NOSPLIT,$0
	MOVW	s+0(FP), R0
	MOVW	s_len+4(FP), R1
	MOVBU	c+8(FP), R2	// byte to find
	MOVW	R0, R4		// store base for later
	ADD	R0, R1		// end 

_sib_loop:
	CMP	R0, R1
	B.EQ	_sib_notfound
	MOVBU.P	1(R0), R3
	CMP	R2, R3
	B.NE	_sib_loop

	SUB	$1, R0		// R0 will be one beyond the position we want
	SUB	R4, R0		// remove base
	MOVW	R0, ret+12(FP) 
	RET

_sib_notfound:
	MOVW	$-1, R0
	MOVW	R0, ret+12(FP)
	RET

// A Duff's device for zeroing memory.
// The compiler jumps to computed addresses within
// this routine to zero chunks of memory.  Do not
// change this code without also changing the code
// in ../../cmd/5g/ggen.c:clearfat.
// R0: zero
// R1: Ptr to memory to be zeroed
// R1 is updated as a side effect.
TEXT runtime·duffzero(SB),NOSPLIT,$0-0
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	MOVW.P	R0, 4(R1)
	RET

// A Duff's device for copying memory.
// The compiler jumps to computed addresses within
// this routine to copy chunks of memory.  Source
// and destination must not overlap.  Do not
// change this code without also changing the code
// in ../../cmd/5g/cgen.c:sgen.
// R0: scratch space
// R1: Ptr to source memory
// R2: Ptr to destination memory
// R1 and R2 are updated as a side effect
TEXT runtime·duffcopy(SB),NOSPLIT,$0-0
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	MOVW.P	4(R1), R0
	MOVW.P	R0, 4(R2)
	RET

TEXT runtime∕internal∕lock·Fastrand1(SB),NOSPLIT,$-4-4
	MOVW	G_M(g), R1
	MOVW	M_Fastrand(R1), R0
	ADD.S	R0, R0
	EOR.MI	$0x88888eef, R0
	MOVW	R0, M_Fastrand(R1)
	MOVW	R0, ret+0(FP)
	RET

TEXT runtime·return0(SB),NOSPLIT,$0
	MOVW	$0, R0
	RET

TEXT runtime∕internal∕lock·Procyield(SB),NOSPLIT,$-4
	MOVW	cycles+0(FP), R1
	MOVW	$0, R0
yieldloop:
	CMP	R0, R1
	B.NE	2(PC)
	RET
	SUB	$1, R1
	B yieldloop

// Called from cgo wrappers, this function returns g->m->curg.stack.hi.
// Must obey the gcc calling convention.
TEXT _cgo_topofstack(SB),NOSPLIT,$8
	// R11 and g register are clobbered by load_g.  They are
	// callee-Save in the gcc calling convention, so Save them here.
	MOVW	R11, saveR11-4(SP)
	MOVW	g, saveG-8(SP)
	
	BL	runtime·load_g(SB)
	MOVW	G_M(g), R0
	MOVW	M_Curg(R0), R0
	MOVW	(G_Stack+Stack_Hi)(R0), R0
	
	MOVW	saveG-8(SP), g
	MOVW	saveR11-4(SP), R11
	RET

// The top-most function running on a goroutine
// returns to Goexit+PCQuantum.
TEXT runtime∕internal∕schedinit·Goexit(SB),NOSPLIT,$-4-0
	MOVW	R0, R0	// NOP
	BL	runtime·goexit1(SB)	// does not return

TEXT runtime∕internal∕core·Getg(SB),NOSPLIT,$-4-4
	MOVW	g, ret+0(FP)
	RET

TEXT runtime∕internal∕check·prefetcht0(SB),NOSPLIT,$0-4
	RET

TEXT runtime∕internal∕check·prefetcht1(SB),NOSPLIT,$0-4
	RET

TEXT runtime∕internal∕check·prefetcht2(SB),NOSPLIT,$0-4
	RET

TEXT runtime∕internal∕check·prefetchnta(SB),NOSPLIT,$0-4
	RET
