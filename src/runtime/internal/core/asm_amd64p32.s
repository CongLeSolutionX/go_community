// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "Funcdata.h"
#include "textflag.h"

TEXT runtime∕internal∕schedinit·rt0_go(SB),NOSPLIT,$0
	// copy arguments forward on an even stack
	MOVL	Argc+0(FP), AX
	MOVL	Argv+4(FP), BX
	MOVL	SP, CX
	SUBL	$128, SP		// plenty of scratch
	ANDL	$~15, CX
	MOVL	CX, SP

	MOVL	AX, 16(SP)
	MOVL	BX, 24(SP)
	
	// create istack out of the given (operating system) stack.
	MOVL	$runtime∕internal∕core·g0(SB), DI
	LEAL	(-64*1024+104)(SP), BX
	MOVL	BX, G_Stackguard0(DI)
	MOVL	BX, G_Stackguard1(DI)
	MOVL	BX, (G_Stack+Stack_Lo)(DI)
	MOVL	SP, (G_Stack+Stack_Hi)(DI)

	// find out information about the processor we're on
	MOVQ	$0, AX
	CPUID
	CMPQ	AX, $0
	JE	nocpuinfo
	MOVQ	$1, AX
	CPUID
	MOVL	CX, runtime∕internal∕hash·cpuid_ecx(SB)
	MOVL	DX, runtime·cpuid_edx(SB)
nocpuinfo:	
	
needtls:
	LEAL	runtime·tls0(SB), DI
	CALL	runtime·settls(SB)

	// store through it, to make sure it works
	get_tls(BX)
	MOVQ	$0x123, g(BX)
	MOVQ	runtime·tls0(SB), AX
	CMPQ	AX, $0x123
	JEQ 2(PC)
	MOVL	AX, 0	// abort
ok:
	// set the per-goroutine and per-mach "registers"
	get_tls(BX)
	LEAL	runtime∕internal∕core·g0(SB), CX
	MOVL	CX, g(BX)
	LEAL	runtime∕internal∕core·M0(SB), AX

	// Save m->g0 = g0
	MOVL	CX, M_G0(AX)
	// Save M0 to g0->m
	MOVL	AX, G_M(CX)

	CLD				// convention is D is always left cleared
	CALL	runtime∕internal∕check·check(SB)

	MOVL	16(SP), AX		// copy Argc
	MOVL	AX, 0(SP)
	MOVL	24(SP), AX		// copy Argv
	MOVL	AX, 4(SP)
	CALL	runtime∕internal∕vdso·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime∕internal∕schedinit·schedinit(SB)

	// create a new goroutine to start program
	MOVL	$runtime·main·f(SB), AX	// entry
	MOVL	$0, 0(SP)
	MOVL	AX, 4(SP)
	CALL	runtime·newproc(SB)

	// start this M
	CALL	runtime∕internal∕sched·Mstart(SB)

	MOVL	$0xf1, 0xf1  // Crash
	RET

DATA	runtime·main·f+0(SB)/4,$runtime·main(SB)
GLOBL	runtime·main·f(SB),RODATA,$4

TEXT runtime·breakpoint(SB),NOSPLIT,$0-0
	INT $3
	RET

TEXT runtime∕internal∕core·Asminit(SB),NOSPLIT,$0-0
	// No per-thread init.
	RET

/*
 *  go-routine
 */

// void gosave(Gobuf*)
// Save state in Gobuf; setjmp
TEXT runtime∕internal∕sched·gosave(SB), NOSPLIT, $0-4
	MOVL	Buf+0(FP), AX	// gobuf
	LEAL	Buf+0(FP), BX	// caller's SP
	MOVL	BX, Gobuf_Sp(AX)
	MOVL	0(SP), BX		// caller's PC
	MOVL	BX, Gobuf_Pc(AX)
	MOVL	$0, Gobuf_Ctxt(AX)
	MOVQ	$0, Gobuf_Ret(AX)
	get_tls(CX)
	MOVL	g(CX), BX
	MOVL	BX, Gobuf_G(AX)
	RET

// void Gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT runtime∕internal∕sched·Gogo(SB), NOSPLIT, $0-4
	MOVL	Buf+0(FP), BX		// gobuf
	MOVL	Gobuf_G(BX), DX
	MOVL	0(DX), CX		// make sure g != nil
	get_tls(CX)
	MOVL	DX, g(CX)
	MOVL	Gobuf_Sp(BX), SP	// restore SP
	MOVL	Gobuf_Ctxt(BX), DX
	MOVQ	Gobuf_Ret(BX), AX
	MOVL	$0, Gobuf_Sp(BX)	// clear to help garbage collector
	MOVQ	$0, Gobuf_Ret(BX)
	MOVL	$0, Gobuf_Ctxt(BX)
	MOVL	Gobuf_Pc(BX), BX
	JMP	BX

// func Mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return.  It should Gogo(&g->Sched)
// to keep running g.
TEXT runtime∕internal∕sched·Mcall(SB), NOSPLIT, $0-4
	MOVL	fn+0(FP), DI
	
	get_tls(CX)
	MOVL	g(CX), AX	// Save state in g->Sched
	MOVL	0(SP), BX	// caller's PC
	MOVL	BX, (G_Sched+Gobuf_Pc)(AX)
	LEAL	fn+0(FP), BX	// caller's SP
	MOVL	BX, (G_Sched+Gobuf_Sp)(AX)
	MOVL	AX, (G_Sched+Gobuf_G)(AX)

	// switch to m->g0 & its stack, call fn
	MOVL	g(CX), BX
	MOVL	G_M(BX), BX
	MOVL	M_G0(BX), SI
	CMPL	SI, AX	// if g == m->g0 call badmcall
	JNE	3(PC)
	MOVL	$runtime·badmcall(SB), AX
	JMP	AX
	MOVL	SI, g(CX)	// g = m->g0
	MOVL	(G_Sched+Gobuf_Sp)(SI), SP	// sp = m->g0->Sched.sp
	PUSHQ	AX
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI
	POPQ	AX
	MOVL	$runtime·badmcall2(SB), AX
	JMP	AX
	RET

// systemstack_switch is a dummy routine that Systemstack leaves at the bottom
// of the G stack.  We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime∕internal∕schedinit·systemstack_switch(SB), NOSPLIT, $0-0
	RET

// func Systemstack(fn func())
TEXT runtime∕internal∕lock·Systemstack(SB), NOSPLIT, $0-4
	MOVL	fn+0(FP), DI	// DI = fn
	get_tls(CX)
	MOVL	g(CX), AX	// AX = g
	MOVL	G_M(AX), BX	// BX = m

	MOVL	M_Gsignal(BX), DX	// DX = gsignal
	CMPL	AX, DX
	JEQ	noswitch

	MOVL	M_G0(BX), DX	// DX = g0
	CMPL	AX, DX
	JEQ	noswitch

	MOVL	M_Curg(BX), R8
	CMPL	AX, R8
	JEQ	switch
	
	// Not g0, not curg. Must be gsignal, but that's not allowed.
	// Hide call from linker nosplit analysis.
	MOVL	$runtime·badsystemstack(SB), AX
	CALL	AX

switch:
	// Save our state in g->Sched.  Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVL	$runtime∕internal∕schedinit·systemstack_switch(SB), SI
	MOVL	SI, (G_Sched+Gobuf_Pc)(AX)
	MOVL	SP, (G_Sched+Gobuf_Sp)(AX)
	MOVL	AX, (G_Sched+Gobuf_G)(AX)

	// switch to g0
	MOVL	DX, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(DX), SP

	// call target function
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI

	// switch back to g
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	G_M(AX), BX
	MOVL	M_Curg(BX), AX
	MOVL	AX, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(AX), SP
	MOVL	$0, (G_Sched+Gobuf_Sp)(AX)
	RET

noswitch:
	// already on m stack, just call directly
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI
	RET

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
//
// The Traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime∕internal∕schedinit·morestack(SB),NOSPLIT,$0-0
	get_tls(CX)
	MOVL	g(CX), BX
	MOVL	G_M(BX), BX

	// Cannot grow scheduler stack (m->g0).
	MOVL	M_G0(BX), SI
	CMPL	g(CX), SI
	JNE	2(PC)
	MOVL	0, AX

	// Cannot grow signal stack (m->gsignal).
	MOVL	M_Gsignal(BX), SI
	CMPL	g(CX), SI
	JNE	2(PC)
	MOVL	0, AX

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVL	8(SP), AX	// f's caller's PC
	MOVL	AX, (M_Morebuf+Gobuf_Pc)(BX)
	LEAL	16(SP), AX	// f's caller's SP
	MOVL	AX, (M_Morebuf+Gobuf_Sp)(BX)
	get_tls(CX)
	MOVL	g(CX), SI
	MOVL	SI, (M_Morebuf+Gobuf_G)(BX)

	// Set g->Sched to context in f.
	MOVL	0(SP), AX // f's PC
	MOVL	AX, (G_Sched+Gobuf_Pc)(SI)
	MOVL	SI, (G_Sched+Gobuf_G)(SI)
	LEAL	8(SP), AX // f's SP
	MOVL	AX, (G_Sched+Gobuf_Sp)(SI)
	MOVL	DX, (G_Sched+Gobuf_Ctxt)(SI)

	// Call newstack on m->g0's stack.
	MOVL	M_G0(BX), BX
	MOVL	BX, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(BX), SP
	CALL	runtime·newstack(SB)
	MOVL	$0, 0x1003	// Crash if newstack returns
	RET

// morestack trampolines
TEXT runtime·morestack_noctxt(SB),NOSPLIT,$0
	MOVL	$0, DX
	JMP	runtime∕internal∕schedinit·morestack(SB)

// Reflectcall: call a function with the given argument list
// func call(argtype *_type, f *FuncVal, arg *byte, argsize, retoffset uint32).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)		\
	CMPL	CX, $MAXSIZE;		\
	JA	3(PC);			\
	MOVL	$NAME(SB), AX;		\
	JMP	AX
// Note: can't just "JMP NAME(SB)" - const_bad inlining results.

TEXT reflect·call(SB), NOSPLIT, $0-0
	JMP	runtime∕internal∕finalize·Reflectcall(SB)

TEXT runtime∕internal∕finalize·Reflectcall(SB), NOSPLIT, $0-20
	MOVLQZX argsize+12(FP), CX
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
	MOVL	$runtime·badreflectcall(SB), AX
	JMP	AX

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-20;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVL	argptr+8(FP), SI;		\
	MOVL	argsize+12(FP), CX;		\
	MOVL	SP, DI;				\
	REP;MOVSB;				\
	/* call function */			\
	MOVL	f+4(FP), DX;			\
	MOVL	(DX), AX;			\
	CALL	AX;				\
	/* copy return values back */		\
	MOVL	argptr+8(FP), DI;		\
	MOVL	argsize+12(FP), CX;		\
	MOVL	retoffset+16(FP), BX;		\
	MOVL	SP, SI;				\
	ADDL	BX, DI;				\
	ADDL	BX, SI;				\
	SUBL	BX, CX;				\
	REP;MOVSB;				\
	/* execute Write barrier updates */	\
	MOVL	argtype+0(FP), DX;		\
	MOVL	argptr+8(FP), DI;		\
	MOVL	argsize+12(FP), CX;		\
	MOVL	retoffset+16(FP), BX;		\
	MOVL	DX, 0(SP);			\
	MOVL	DI, 4(SP);			\
	MOVL	CX, 8(SP);			\
	MOVL	BX, 12(SP);			\
	CALL	runtime·callwritebarrier(SB);	\
	RET

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

// bool Cas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕sched·Cas(SB), NOSPLIT, $0-17
	MOVL	Ptr+0(FP), BX
	MOVL	old+4(FP), AX
	MOVL	new+8(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	SETEQ	ret+16(FP)
	RET

TEXT runtime∕internal∕core·Casuintptr(SB), NOSPLIT, $0-17
	JMP	runtime∕internal∕sched·Cas(SB)

TEXT runtime∕internal∕core·Atomicloaduintptr(SB), NOSPLIT, $0-12
	JMP	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕channels·Atomicloaduint(SB), NOSPLIT, $0-12
	JMP	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕core·atomicstoreuintptr(SB), NOSPLIT, $0-12
	JMP	runtime∕internal∕lock·Atomicstore(SB)

// bool	runtime∕internal∕sched·Cas64(uint64 *val, uint64 old, uint64 new)
// Atomically:
//	if(*val == *old){
//		*val = new;
//		return 1;
//	} else {
//		return 0;
//	}
TEXT runtime∕internal∕sched·Cas64(SB), NOSPLIT, $0-25
	MOVL	Ptr+0(FP), BX
	MOVQ	old+8(FP), AX
	MOVQ	new+16(FP), CX
	LOCK
	CMPXCHGQ	CX, 0(BX)
	SETEQ	ret+24(FP)
	RET

// bool casp(void **val, void *old, void *new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕check·casp1(SB), NOSPLIT, $0-17
	MOVL	Ptr+0(FP), BX
	MOVL	old+4(FP), AX
	MOVL	new+8(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	SETEQ	ret+16(FP)
	RET

// uint32 Xadd(uint32 volatile *val, int32 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT runtime∕internal∕lock·Xadd(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), BX
	MOVL	delta+4(FP), AX
	MOVL	AX, CX
	LOCK
	XADDL	AX, 0(BX)
	ADDL	CX, AX
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime∕internal∕lock·Xadd64(SB), NOSPLIT, $0-24
	MOVL	Ptr+0(FP), BX
	MOVQ	delta+8(FP), AX
	MOVQ	AX, CX
	LOCK
	XADDQ	AX, 0(BX)
	ADDQ	CX, AX
	MOVQ	AX, ret+16(FP)
	RET

TEXT runtime·xchg(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), BX
	MOVL	new+4(FP), AX
	XCHGL	AX, 0(BX)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime∕internal∕sched·Xchg64(SB), NOSPLIT, $0-24
	MOVL	Ptr+0(FP), BX
	MOVQ	new+8(FP), AX
	XCHGQ	AX, 0(BX)
	MOVQ	AX, ret+16(FP)
	RET

TEXT runtime·xchgp1(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), BX
	MOVL	new+4(FP), AX
	XCHGL	AX, 0(BX)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime∕internal∕netpoll·xchguintptr(SB), NOSPLIT, $0-12
	JMP	runtime·xchg(SB)

TEXT runtime∕internal∕lock·Procyield(SB),NOSPLIT,$0-0
	MOVL	cycles+0(FP), AX
again:
	PAUSE
	SUBL	$1, AX
	JNZ	again
	RET

TEXT runtime∕internal∕sched·Atomicstorep1(SB), NOSPLIT, $0-8
	MOVL	Ptr+0(FP), BX
	MOVL	val+4(FP), AX
	XCHGL	AX, 0(BX)
	RET

TEXT runtime∕internal∕lock·Atomicstore(SB), NOSPLIT, $0-8
	MOVL	Ptr+0(FP), BX
	MOVL	val+4(FP), AX
	XCHGL	AX, 0(BX)
	RET

TEXT runtime∕internal∕sched·Atomicstore64(SB), NOSPLIT, $0-16
	MOVL	Ptr+0(FP), BX
	MOVQ	val+8(FP), AX
	XCHGQ	AX, 0(BX)
	RET

// void	runtime∕internal∕sched·Atomicor8(byte volatile*, byte);
TEXT runtime∕internal∕sched·Atomicor8(SB), NOSPLIT, $0-5
	MOVL	Ptr+0(FP), BX
	MOVB	val+4(FP), AX
	LOCK
	ORB	AX, 0(BX)
	RET

// void Jmpdefer(fn, sp);
// called from deferreturn.
// 1. pop the caller
// 2. sub 5 Bytes from the Callers return
// 3. jmp to the argument
TEXT runtime∕internal∕schedinit·Jmpdefer(SB), NOSPLIT, $0-8
	MOVL	fv+0(FP), DX
	MOVL	argp+4(FP), BX
	LEAL	-8(BX), SP	// caller sp after CALL
	SUBL	$5, (SP)	// return to CALL again
	MOVL	0(DX), BX
	JMP	BX	// but first run the deferred function

// Asmcgocall(void(*fn)(void*), void *arg)
// Not implemented.
TEXT runtime∕internal∕sched·Asmcgocall(SB),NOSPLIT,$0-8
	MOVL	0, AX
	RET

// Asmcgocall(void(*fn)(void*), void *arg)
// Not implemented.
TEXT runtime∕internal∕cgo·asmcgocall_errno(SB),NOSPLIT,$0-12
	MOVL	0, AX
	RET

// cgocallback(void (*fn)(void*), void *frame, uintptr framesize)
// Not implemented.
TEXT runtime·cgocallback(SB),NOSPLIT,$0-12
	MOVL	0, AX
	RET

// void Setg(G*); set g. for use by needm.
// Not implemented.
TEXT runtime∕internal∕core·Setg(SB), NOSPLIT, $0-4
	MOVL	0, AX
	RET

// check that SP is in range [g->stack.lo, g->stack.hi)
TEXT runtime·stackcheck(SB), NOSPLIT, $0-0
	get_tls(CX)
	MOVL	g(CX), AX
	CMPL	(G_Stack+Stack_Hi)(AX), SP
	JHI	2(PC)
	MOVL	0, AX
	CMPL	SP, (G_Stack+Stack_Lo)(AX)
	JHI	2(PC)
	MOVL	0, AX
	RET

TEXT runtime∕internal∕core·Memclr(SB),NOSPLIT,$0-8
	MOVL	Ptr+0(FP), DI
	MOVL	n+4(FP), CX
	MOVQ	CX, BX
	ANDQ	$7, BX
	SHRQ	$3, CX
	MOVQ	$0, AX
	CLD
	REP
	STOSQ
	MOVQ	BX, CX
	REP
	STOSB
	RET

TEXT runtime∕internal∕lock·Getcallerpc(SB),NOSPLIT,$0-12
	MOVL	argp+0(FP),AX		// addr of first arg
	MOVL	-8(AX),AX		// get calling pc
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·gogetcallerpc(SB),NOSPLIT,$0-12
	MOVL	p+0(FP),AX		// addr of first arg
	MOVL	-8(AX),AX		// get calling pc
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime∕internal∕channels·setcallerpc(SB),NOSPLIT,$0-8
	MOVL	argp+0(FP),AX		// addr of first arg
	MOVL	pc+4(FP), BX		// pc to set
	MOVQ	BX, -8(AX)		// set calling pc
	RET

TEXT runtime∕internal∕lock·Getcallersp(SB),NOSPLIT,$0-12
	MOVL	argp+0(FP), AX
	MOVL	AX, ret+8(FP)
	RET

// func gogetcallersp(p unsafe.Pointer) uintptr
TEXT runtime·gogetcallersp(SB),NOSPLIT,$0-12
	MOVL	p+0(FP),AX		// addr of first arg
	MOVL	AX, ret+8(FP)
	RET

// int64 runtime∕internal∕sched·Cputicks(void)
TEXT runtime∕internal∕sched·Cputicks(SB),NOSPLIT,$0-0
	RDTSC
	SHLQ	$32, DX
	ADDQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

// memhash_varlen(p unsafe.Pointer, h seed) uintptr
// redirects to Memhash(p, h, size) using the size
// stored in the closure.
TEXT runtime·memhash_varlen(SB),NOSPLIT,$24-12
	GO_ARGS
	NO_LOCAL_POINTERS
	MOVL	p+0(FP), AX
	MOVL	h+4(FP), BX
	MOVL	4(DX), CX
	MOVL	AX, 0(SP)
	MOVL	BX, 4(SP)
	MOVL	CX, 8(SP)
	CALL	runtime∕internal∕sched·Memhash(SB)
	MOVL	16(SP), AX
	MOVL	AX, ret+8(FP)
	RET

// Hash function using AES hardware instructions
// For now, our one amd64p32 system (NaCl) does not
// support using AES instructions, so have not bothered to
// Write the implementations. Can copy and adjust the ones
// in asm_amd64.s when the time comes.

TEXT runtime∕internal∕sched·aeshash(SB),NOSPLIT,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime∕internal∕hash·aeshashstr(SB),NOSPLIT,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime∕internal∕hash·aeshash32(SB),NOSPLIT,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime∕internal∕hash·aeshash64(SB),NOSPLIT,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime∕internal∕hash·Memeq(SB),NOSPLIT,$0-17
	MOVL	a+0(FP), SI
	MOVL	b+4(FP), DI
	MOVL	size+8(FP), BX
	CALL	runtime·memeqbody(SB)
	MOVB	AX, ret+16(FP)
	RET

// memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB),NOSPLIT,$0-9
	MOVL    a+0(FP), SI
	MOVL    b+4(FP), DI
	CMPL    SI, DI
	JEQ     eq
	MOVL    4(DX), BX    // compiler stores size at offset 4 in the closure
	CALL    runtime·memeqbody(SB)
	MOVB    AX, ret+8(FP)
	RET
eq:
	MOVB    $1, ret+8(FP)
	RET

// eqstring tests whether two strings are equal.
// The compiler guarantees that strings passed
// to eqstring have equal length.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.
TEXT runtime·eqstring(SB),NOSPLIT,$0-17
	MOVL	s1str+0(FP), SI
	MOVL	s2str+8(FP), DI
	CMPL	SI, DI
	JEQ	same
	MOVL	s1len+4(FP), BX
	CALL	runtime·memeqbody(SB)
	MOVB	AX, v+16(FP)
	RET
same:
	MOVB	$1, v+16(FP)
	RET

// a in SI
// b in DI
// count in BX
TEXT runtime·memeqbody(SB),NOSPLIT,$0-0
	XORQ	AX, AX

	CMPQ	BX, $8
	JB	small
	
	// 64 Bytes at a time using xmm registers
hugeloop:
	CMPQ	BX, $64
	JB	bigloop
	MOVOU	(SI), X0
	MOVOU	(DI), X1
	MOVOU	16(SI), X2
	MOVOU	16(DI), X3
	MOVOU	32(SI), X4
	MOVOU	32(DI), X5
	MOVOU	48(SI), X6
	MOVOU	48(DI), X7
	PCMPEQB	X1, X0
	PCMPEQB	X3, X2
	PCMPEQB	X5, X4
	PCMPEQB	X7, X6
	PAND	X2, X0
	PAND	X6, X4
	PAND	X4, X0
	PMOVMSKB X0, DX
	ADDQ	$64, SI
	ADDQ	$64, DI
	SUBQ	$64, BX
	CMPL	DX, $0xffff
	JEQ	hugeloop
	RET

	// 8 Bytes at a time using 64-bit register
bigloop:
	CMPQ	BX, $8
	JBE	leftover
	MOVQ	(SI), CX
	MOVQ	(DI), DX
	ADDQ	$8, SI
	ADDQ	$8, DI
	SUBQ	$8, BX
	CMPQ	CX, DX
	JEQ	bigloop
	RET

	// remaining 0-8 Bytes
leftover:
	ADDQ	BX, SI
	ADDQ	BX, DI
	MOVQ	-8(SI), CX
	MOVQ	-8(DI), DX
	CMPQ	CX, DX
	SETEQ	AX
	RET

small:
	CMPQ	BX, $0
	JEQ	equal

	LEAQ	0(BX*8), CX
	NEGQ	CX

	CMPB	SI, $0xf8
	JA	si_high

	// load at SI won't cross a page boundary.
	MOVQ	(SI), SI
	JMP	si_finish
si_high:
	// address ends in 11111xxx.  Load up to Bytes we want, move to correct position.
	MOVQ	BX, DX
	ADDQ	SI, DX
	MOVQ	-8(DX), SI
	SHRQ	CX, SI
si_finish:

	// same for DI.
	CMPB	DI, $0xf8
	JA	di_high
	MOVQ	(DI), DI
	JMP	di_finish
di_high:
	MOVQ	BX, DX
	ADDQ	DI, DX
	MOVQ	-8(DX), DI
	SHRQ	CX, DI
di_finish:

	SUBQ	SI, DI
	SHLQ	CX, DI
equal:
	SETEQ	AX
	RET

TEXT runtime·cmpstring(SB),NOSPLIT,$0-20
	MOVL	s1_base+0(FP), SI
	MOVL	s1_len+4(FP), BX
	MOVL	s2_base+8(FP), DI
	MOVL	s2_len+12(FP), DX
	CALL	runtime·cmpbody(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT bytes·Compare(SB),NOSPLIT,$0-28
	MOVL	s1+0(FP), SI
	MOVL	s1+4(FP), BX
	MOVL	s2+12(FP), DI
	MOVL	s2+16(FP), DX
	CALL	runtime·cmpbody(SB)
	MOVQ	AX, res+24(FP)
	RET

// input:
//   SI = a
//   DI = b
//   BX = alen
//   DX = blen
// output:
//   AX = 1/0/-1
TEXT runtime·cmpbody(SB),NOSPLIT,$0-0
	CMPQ	SI, DI
	JEQ	allsame
	CMPQ	BX, DX
	MOVQ	DX, R8
	CMOVQLT	BX, R8 // R8 = min(alen, blen) = # of Bytes to compare
	CMPQ	R8, $8
	JB	small

loop:
	CMPQ	R8, $16
	JBE	_0through16
	MOVOU	(SI), X0
	MOVOU	(DI), X1
	PCMPEQB X0, X1
	PMOVMSKB X1, AX
	XORQ	$0xffff, AX	// convert EQ to NE
	JNE	diff16	// branch if at least one byte is not equal
	ADDQ	$16, SI
	ADDQ	$16, DI
	SUBQ	$16, R8
	JMP	loop
	
	// AX = bit mask of differences
diff16:
	BSFQ	AX, BX	// Index of first byte that differs
	XORQ	AX, AX
	ADDQ	BX, SI
	MOVB	(SI), CX
	ADDQ	BX, DI
	CMPB	CX, (DI)
	SETHI	AX
	LEAQ	-1(AX*2), AX	// convert 1/0 to +1/-1
	RET

	// 0 through 16 Bytes left, alen>=8, blen>=8
_0through16:
	CMPQ	R8, $8
	JBE	_0through8
	MOVQ	(SI), AX
	MOVQ	(DI), CX
	CMPQ	AX, CX
	JNE	diff8
_0through8:
	ADDQ	R8, SI
	ADDQ	R8, DI
	MOVQ	-8(SI), AX
	MOVQ	-8(DI), CX
	CMPQ	AX, CX
	JEQ	allsame

	// AX and CX contain parts of a and b that differ.
diff8:
	BSWAPQ	AX	// reverse order of Bytes
	BSWAPQ	CX
	XORQ	AX, CX
	BSRQ	CX, CX	// Index of highest bit difference
	SHRQ	CX, AX	// move a's bit to bottom
	ANDQ	$1, AX	// mask bit
	LEAQ	-1(AX*2), AX // 1/0 => +1/-1
	RET

	// 0-7 Bytes in common
small:
	LEAQ	(R8*8), CX	// Bytes left -> bits left
	NEGQ	CX		//  - bits lift (== 64 - bits left mod 64)
	JEQ	allsame

	// load Bytes of a into high Bytes of AX
	CMPB	SI, $0xf8
	JA	si_high
	MOVQ	(SI), SI
	JMP	si_finish
si_high:
	ADDQ	R8, SI
	MOVQ	-8(SI), SI
	SHRQ	CX, SI
si_finish:
	SHLQ	CX, SI

	// load Bytes of b in to high Bytes of BX
	CMPB	DI, $0xf8
	JA	di_high
	MOVQ	(DI), DI
	JMP	di_finish
di_high:
	ADDQ	R8, DI
	MOVQ	-8(DI), DI
	SHRQ	CX, DI
di_finish:
	SHLQ	CX, DI

	BSWAPQ	SI	// reverse order of Bytes
	BSWAPQ	DI
	XORQ	SI, DI	// find bit differences
	JEQ	allsame
	BSRQ	DI, CX	// Index of highest bit difference
	SHRQ	CX, SI	// move a's bit to bottom
	ANDQ	$1, SI	// mask bit
	LEAQ	-1(SI*2), AX // 1/0 => +1/-1
	RET

allsame:
	XORQ	AX, AX
	XORQ	CX, CX
	CMPQ	BX, DX
	SETGT	AX	// 1 if alen > blen
	SETEQ	CX	// 1 if alen == blen
	LEAQ	-1(CX)(AX*2), AX	// 1,0,-1 result
	RET

TEXT bytes·IndexByte(SB),NOSPLIT,$0
	MOVL s+0(FP), SI
	MOVL s_len+4(FP), BX
	MOVB c+12(FP), AL
	CALL runtime·indexbytebody(SB)
	MOVL AX, ret+16(FP)
	RET

TEXT strings·IndexByte(SB),NOSPLIT,$0
	MOVL s+0(FP), SI
	MOVL s_len+4(FP), BX
	MOVB c+8(FP), AL
	CALL runtime·indexbytebody(SB)
	MOVL AX, ret+16(FP)
	RET

// input:
//   SI: Data
//   BX: Data len
//   AL: byte sought
// output:
//   AX
TEXT runtime·indexbytebody(SB),NOSPLIT,$0
	MOVL SI, DI

	CMPL BX, $16
	JLT small

	// Round up to first 16-byte boundary
	TESTL $15, SI
	JZ aligned
	MOVL SI, CX
	ANDL $~15, CX
	ADDL $16, CX

	// search the beginning
	SUBL SI, CX
	REPN; SCASB
	JZ success

// DI is 16-byte aligned; get Ready to search using SSE instructions
aligned:
	// Round down to last 16-byte boundary
	MOVL BX, R11
	ADDL SI, R11
	ANDL $~15, R11

	// shuffle X0 around so that each byte contains c
	MOVD AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0
	PSHUFL $0, X0, X0
	JMP condition

sse:
	// move the Next 16-byte chunk of the buffer into X1
	MOVO (DI), X1
	// compare Bytes in X0 to X1
	PCMPEQB X0, X1
	// take the top bit of each byte in X1 and put the result in DX
	PMOVMSKB X1, DX
	TESTL DX, DX
	JNZ ssesuccess
	ADDL $16, DI

condition:
	CMPL DI, R11
	JLT sse

	// search the end
	MOVL SI, CX
	ADDL BX, CX
	SUBL R11, CX
	// if CX == 0, the zero flag will be set and we'll end up
	// returning a false success
	JZ failure
	REPN; SCASB
	JZ success

failure:
	MOVL $-1, AX
	RET

// handle for lengths < 16
small:
	MOVL BX, CX
	REPN; SCASB
	JZ success
	MOVL $-1, AX
	RET

// we've found the chunk containing the byte
// now just figure out which specific byte it is
ssesuccess:
	// get the Index of the least significant set bit
	BSFW DX, DX
	SUBL SI, DI
	ADDL DI, DX
	MOVL DX, AX
	RET

success:
	SUBL SI, DI
	SUBL $1, DI
	MOVL DI, AX
	RET

TEXT bytes·Equal(SB),NOSPLIT,$0-25
	MOVL	a_len+4(FP), BX
	MOVL	b_len+16(FP), CX
	XORL	AX, AX
	CMPL	BX, CX
	JNE	eqret
	MOVL	a+0(FP), SI
	MOVL	b+12(FP), DI
	CALL	runtime·memeqbody(SB)
eqret:
	MOVB	AX, ret+24(FP)
	RET

TEXT runtime∕internal∕lock·Fastrand1(SB), NOSPLIT, $0-4
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	G_M(AX), AX
	MOVL	M_Fastrand(AX), DX
	ADDL	DX, DX
	MOVL	DX, BX
	XORL	$0x88888eef, DX
	CMOVLMI	BX, DX
	MOVL	DX, M_Fastrand(AX)
	MOVL	DX, ret+0(FP)
	RET

TEXT runtime·return0(SB), NOSPLIT, $0
	MOVL	$0, AX
	RET

// The top-most function running on a goroutine
// returns to Goexit+PCQuantum.
TEXT runtime∕internal∕schedinit·Goexit(SB),NOSPLIT,$0-0
	BYTE	$0x90	// NOP
	CALL	runtime·goexit1(SB)	// does not return

TEXT runtime∕internal∕core·Getg(SB),NOSPLIT,$0-4
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	AX, ret+0(FP)
	RET

TEXT runtime∕internal∕check·prefetcht0(SB),NOSPLIT,$0-4
	MOVL	addr+0(FP), AX
	PREFETCHT0	(AX)
	RET

TEXT runtime∕internal∕check·prefetcht1(SB),NOSPLIT,$0-4
	MOVL	addr+0(FP), AX
	PREFETCHT1	(AX)
	RET


TEXT runtime∕internal∕check·prefetcht2(SB),NOSPLIT,$0-4
	MOVL	addr+0(FP), AX
	PREFETCHT2	(AX)
	RET

TEXT runtime∕internal∕check·prefetchnta(SB),NOSPLIT,$0-4
	MOVL	addr+0(FP), AX
	PREFETCHNTA	(AX)
	RET
