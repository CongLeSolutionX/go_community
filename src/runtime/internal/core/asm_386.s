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
	SUBL	$128, SP		// plenty of scratch
	ANDL	$~15, SP
	MOVL	AX, 120(SP)		// Save Argc, Argv away
	MOVL	BX, 124(SP)

	// set default stack bounds.
	// _cgo_init may update stackguard.
	MOVL	$runtime∕internal∕core·g0(SB), BP
	LEAL	(-64*1024+104)(SP), BX
	MOVL	BX, G_Stackguard0(BP)
	MOVL	BX, G_Stackguard1(BP)
	MOVL	BX, (G_Stack+Stack_Lo)(BP)
	MOVL	SP, (G_Stack+Stack_Hi)(BP)
	
	// find out information about the processor we're on
	MOVL	$0, AX
	CPUID
	CMPL	AX, $0
	JE	nocpuinfo
	MOVL	$1, AX
	CPUID
	MOVL	CX, runtime∕internal∕hash·cpuid_ecx(SB)
	MOVL	DX, runtime·cpuid_edx(SB)
nocpuinfo:	

	// if there is an _cgo_init, call it to let it
	// initialize and to set up GS.  if not,
	// we set up GS ourselves.
	MOVL	_cgo_init(SB), AX
	TESTL	AX, AX
	JZ	needtls
	MOVL	$setg_gcc<>(SB), BX
	MOVL	BX, 4(SP)
	MOVL	BP, 0(SP)
	CALL	AX

	// update stackguard after _cgo_init
	MOVL	$runtime∕internal∕core·g0(SB), CX
	MOVL	(G_Stack+Stack_Lo)(CX), AX
	ADDL	$const_StackGuard, AX
	MOVL	AX, G_Stackguard0(CX)
	MOVL	AX, G_Stackguard1(CX)

	// skip runtime·ldt0setup(SB) and tls test after _cgo_init for non-windows
	CMPL runtime·iswindows(SB), $0
	JEQ ok
needtls:
	// skip runtime·ldt0setup(SB) and tls test on Plan 9 in all cases
	CMPL	runtime·isplan9(SB), $1
	JEQ	ok

	// set up %gs
	CALL	runtime·ldt0setup(SB)

	// store through it, to make sure it works
	get_tls(BX)
	MOVL	$0x123, g(BX)
	MOVL	runtime·tls0(SB), AX
	CMPL	AX, $0x123
	JEQ	ok
	MOVL	AX, 0	// abort
ok:
	// set up m and g "registers"
	get_tls(BX)
	LEAL	runtime∕internal∕core·g0(SB), CX
	MOVL	CX, g(BX)
	LEAL	runtime∕internal∕core·M0(SB), AX

	// Save m->g0 = g0
	MOVL	CX, M_G0(AX)
	// Save g0->m = M0
	MOVL	AX, G_M(CX)

	CALL	runtime·emptyfunc(SB)	// fault if stack check is wrong

	// convention is D is always cleared
	CLD

	CALL	runtime∕internal∕check·check(SB)

	// saved Argc, Argv
	MOVL	120(SP), AX
	MOVL	AX, 0(SP)
	MOVL	124(SP), AX
	MOVL	AX, 4(SP)
	CALL	runtime∕internal∕vdso·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime∕internal∕schedinit·schedinit(SB)

	// create a new goroutine to start program
	PUSHL	$runtime·main·f(SB)	// entry
	PUSHL	$0	// arg size
	CALL	runtime·newproc(SB)
	POPL	AX
	POPL	AX

	// start this M
	CALL	runtime∕internal∕sched·Mstart(SB)

	INT $3
	RET

DATA	runtime·main·f+0(SB)/4,$runtime·main(SB)
GLOBL	runtime·main·f(SB),RODATA,$4

TEXT runtime·breakpoint(SB),NOSPLIT,$0-0
	INT $3
	RET

TEXT runtime∕internal∕core·Asminit(SB),NOSPLIT,$0-0
	// Linux and MinGW start the FPU in extended double precision.
	// Other operating systems use double precision.
	// Change to double precision to match them,
	// and to match other hardware that only has double.
	PUSHL $0x27F
	FLDCW	0(SP)
	POPL AX
	RET

/*
 *  go-routine
 */

// void gosave(Gobuf*)
// Save state in Gobuf; setjmp
TEXT runtime∕internal∕sched·gosave(SB), NOSPLIT, $0-4
	MOVL	Buf+0(FP), AX		// gobuf
	LEAL	Buf+0(FP), BX		// caller's SP
	MOVL	BX, Gobuf_Sp(AX)
	MOVL	0(SP), BX		// caller's PC
	MOVL	BX, Gobuf_Pc(AX)
	MOVL	$0, Gobuf_Ret(AX)
	MOVL	$0, Gobuf_Ctxt(AX)
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
	MOVL	Gobuf_Ret(BX), AX
	MOVL	Gobuf_Ctxt(BX), DX
	MOVL	$0, Gobuf_Sp(BX)	// clear to help garbage collector
	MOVL	$0, Gobuf_Ret(BX)
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
	PUSHL	AX
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI
	POPL	AX
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

	MOVL	M_Curg(BX), BP
	CMPL	AX, BP
	JEQ	switch
	
	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOVL	$runtime·badsystemstack(SB), AX
	CALL	AX

switch:
	// Save our state in g->Sched.  Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVL	$runtime∕internal∕schedinit·systemstack_switch(SB), (G_Sched+Gobuf_Pc)(AX)
	MOVL	SP, (G_Sched+Gobuf_Sp)(AX)
	MOVL	AX, (G_Sched+Gobuf_G)(AX)

	// switch to g0
	MOVL	DX, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(DX), BX
	// make it look like Mstart called Systemstack on g0, to stop Traceback
	SUBL	$4, BX
	MOVL	$runtime∕internal∕sched·Mstart(SB), DX
	MOVL	DX, 0(BX)
	MOVL	BX, SP

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
	// already on system stack, just call directly
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
	// Cannot grow scheduler stack (m->g0).
	get_tls(CX)
	MOVL	g(CX), BX
	MOVL	G_M(BX), BX
	MOVL	M_G0(BX), SI
	CMPL	g(CX), SI
	JNE	2(PC)
	INT	$3

	// Cannot grow signal stack.
	MOVL	M_Gsignal(BX), SI
	CMPL	g(CX), SI
	JNE	2(PC)
	INT	$3

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVL	4(SP), DI	// f's caller's PC
	MOVL	DI, (M_Morebuf+Gobuf_Pc)(BX)
	LEAL	8(SP), CX	// f's caller's SP
	MOVL	CX, (M_Morebuf+Gobuf_Sp)(BX)
	get_tls(CX)
	MOVL	g(CX), SI
	MOVL	SI, (M_Morebuf+Gobuf_G)(BX)

	// Set g->Sched to context in f.
	MOVL	0(SP), AX	// f's PC
	MOVL	AX, (G_Sched+Gobuf_Pc)(SI)
	MOVL	SI, (G_Sched+Gobuf_G)(SI)
	LEAL	4(SP), AX	// f's SP
	MOVL	AX, (G_Sched+Gobuf_Sp)(SI)
	MOVL	DX, (G_Sched+Gobuf_Ctxt)(SI)

	// Call newstack on m->g0's stack.
	MOVL	M_G0(BX), BP
	MOVL	BP, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(BP), AX
	MOVL	-4(AX), BX	// fault if CALL would, before smashing SP
	MOVL	AX, SP
	CALL	runtime·newstack(SB)
	MOVL	$0, 0x1003	// Crash if newstack returns
	RET

TEXT runtime·morestack_noctxt(SB),NOSPLIT,$0-0
	MOVL	$0, DX
	JMP runtime∕internal∕schedinit·morestack(SB)

// Reflectcall: call a function with the given argument list
// func call(f *FuncVal, arg *byte, argsize, retoffset uint32).
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

TEXT runtime∕internal∕finalize·Reflectcall(SB), NOSPLIT, $0-16
	MOVL	argsize+8(FP), CX
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
TEXT NAME(SB), WRAPPER, $MAXSIZE-16;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVL	argptr+4(FP), SI;		\
	MOVL	argsize+8(FP), CX;		\
	MOVL	SP, DI;				\
	REP;MOVSB;				\
	/* call function */			\
	MOVL	f+0(FP), DX;			\
	MOVL	(DX), AX; 			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	CALL	AX;				\
	/* copy return values back */		\
	MOVL	argptr+4(FP), DI;		\
	MOVL	argsize+8(FP), CX;		\
	MOVL	retoffset+12(FP), BX;		\
	MOVL	SP, SI;				\
	ADDL	BX, DI;				\
	ADDL	BX, SI;				\
	SUBL	BX, CX;				\
	REP;MOVSB;				\
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
//	}else
//		return 0;
TEXT runtime∕internal∕sched·Cas(SB), NOSPLIT, $0-13
	MOVL	Ptr+0(FP), BX
	MOVL	old+4(FP), AX
	MOVL	new+8(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	JZ 4(PC)
	MOVL	$0, AX
	MOVB	AX, ret+12(FP)
	RET
	MOVL	$1, AX
	MOVB	AX, ret+12(FP)
	RET

TEXT runtime∕internal∕core·Casuintptr(SB), NOSPLIT, $0-13
	JMP	runtime∕internal∕sched·Cas(SB)

TEXT runtime∕internal∕core·Atomicloaduintptr(SB), NOSPLIT, $0-8
	JMP	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕channels·atomicloaduint(SB), NOSPLIT, $0-8
	JMP	runtime∕internal∕lock·Atomicload(SB)

TEXT runtime∕internal∕core·atomicstoreuintptr(SB), NOSPLIT, $0-8
	JMP	runtime∕internal∕lock·Atomicstore(SB)

// bool runtime∕internal∕sched·Cas64(uint64 *val, uint64 old, uint64 new)
// Atomically:
//	if(*val == *old){
//		*val = new;
//		return 1;
//	} else {
//		return 0;
//	}
TEXT runtime∕internal∕sched·Cas64(SB), NOSPLIT, $0-21
	MOVL	Ptr+0(FP), BP
	MOVL	old_lo+4(FP), AX
	MOVL	old_hi+8(FP), DX
	MOVL	new_lo+12(FP), BX
	MOVL	new_hi+16(FP), CX
	LOCK
	CMPXCHG8B	0(BP)
	JNZ	fail
	MOVL	$1, AX
	MOVB	AX, ret+20(FP)
	RET
fail:
	MOVL	$0, AX
	MOVB	AX, ret+20(FP)
	RET

// bool casp(void **p, void *old, void *new)
// Atomically:
//	if(*p == old){
//		*p = new;
//		return 1;
//	}else
//		return 0;
TEXT runtime∕internal∕check·casp1(SB), NOSPLIT, $0-13
	MOVL	Ptr+0(FP), BX
	MOVL	old+4(FP), AX
	MOVL	new+8(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	JZ 4(PC)
	MOVL	$0, AX
	MOVB	AX, ret+12(FP)
	RET
	MOVL	$1, AX
	MOVB	AX, ret+12(FP)
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

TEXT runtime·xchg(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), BX
	MOVL	new+4(FP), AX
	XCHGL	AX, 0(BX)
	MOVL	AX, ret+8(FP)
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

TEXT runtime∕internal∕sched·atomicstorep1(SB), NOSPLIT, $0-8
	MOVL	Ptr+0(FP), BX
	MOVL	val+4(FP), AX
	XCHGL	AX, 0(BX)
	RET

TEXT runtime∕internal∕lock·Atomicstore(SB), NOSPLIT, $0-8
	MOVL	Ptr+0(FP), BX
	MOVL	val+4(FP), AX
	XCHGL	AX, 0(BX)
	RET

// uint64 Atomicload64(uint64 volatile* addr);
TEXT runtime∕internal∕sched·Atomicload64(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), AX
	LEAL	ret_lo+4(FP), BX
	// MOVQ (%EAX), %MM0
	BYTE $0x0f; BYTE $0x6f; BYTE $0x00
	// MOVQ %MM0, 0(%EBX)
	BYTE $0x0f; BYTE $0x7f; BYTE $0x03
	// EMMS
	BYTE $0x0F; BYTE $0x77
	RET

// void runtime∕internal∕sched·Atomicstore64(uint64 volatile* addr, uint64 v);
TEXT runtime∕internal∕sched·Atomicstore64(SB), NOSPLIT, $0-12
	MOVL	Ptr+0(FP), AX
	// MOVQ and EMMS were introduced on the Pentium MMX.
	// MOVQ 0x8(%ESP), %MM0
	BYTE $0x0f; BYTE $0x6f; BYTE $0x44; BYTE $0x24; BYTE $0x08
	// MOVQ %MM0, (%EAX)
	BYTE $0x0f; BYTE $0x7f; BYTE $0x00 
	// EMMS
	BYTE $0x0F; BYTE $0x77
	// This is essentially a no-op, but it provides required memory fencing.
	// It can be replaced with MFENCE, but MFENCE was introduced only on the Pentium4 (SSE2).
	MOVL	$0, AX
	LOCK
	XADDL	AX, (SP)
	RET

// void	runtime∕internal∕sched·Atomicor8(byte volatile*, byte);
TEXT runtime∕internal∕sched·Atomicor8(SB), NOSPLIT, $0-5
	MOVL	Ptr+0(FP), AX
	MOVB	val+4(FP), BX
	LOCK
	ORB	BX, (AX)
	RET

// void Jmpdefer(fn, sp);
// called from deferreturn.
// 1. pop the caller
// 2. sub 5 Bytes from the Callers return
// 3. jmp to the argument
TEXT runtime∕internal∕schedinit·Jmpdefer(SB), NOSPLIT, $0-8
	MOVL	fv+0(FP), DX	// fn
	MOVL	argp+4(FP), BX	// caller sp
	LEAL	-4(BX), SP	// caller sp after CALL
	SUBL	$5, (SP)	// return to CALL again
	MOVL	0(DX), BX
	JMP	BX	// but first run the deferred function

// Save state of caller into g->Sched.
TEXT gosave<>(SB),NOSPLIT,$0
	PUSHL	AX
	PUSHL	BX
	get_tls(BX)
	MOVL	g(BX), BX
	LEAL	arg+0(FP), AX
	MOVL	AX, (G_Sched+Gobuf_Sp)(BX)
	MOVL	-4(AX), AX
	MOVL	AX, (G_Sched+Gobuf_Pc)(BX)
	MOVL	$0, (G_Sched+Gobuf_Ret)(BX)
	MOVL	$0, (G_Sched+Gobuf_Ctxt)(BX)
	POPL	BX
	POPL	AX
	RET

// Asmcgocall(void(*fn)(void*), void *arg)
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.c for more details.
TEXT runtime∕internal∕sched·Asmcgocall(SB),NOSPLIT,$0-8
	MOVL	fn+0(FP), AX
	MOVL	arg+4(FP), BX
	CALL	Asmcgocall<>(SB)
	RET

TEXT runtime∕internal∕cgo·asmcgocall_errno(SB),NOSPLIT,$0-12
	MOVL	fn+0(FP), AX
	MOVL	arg+4(FP), BX
	CALL	Asmcgocall<>(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT Asmcgocall<>(SB),NOSPLIT,$0-0
	// fn in AX, arg in BX
	MOVL	SP, DX

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already.
	get_tls(CX)
	MOVL	g(CX), BP
	MOVL	G_M(BP), BP
	MOVL	M_G0(BP), SI
	MOVL	g(CX), DI
	CMPL	SI, DI
	JEQ	4(PC)
	CALL	gosave<>(SB)
	MOVL	SI, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(SI), SP

	// Now on a scheduling stack (a pthread-created stack).
	SUBL	$32, SP
	ANDL	$~15, SP	// alignment, perhaps unnecessary
	MOVL	DI, 8(SP)	// Save g
	MOVL	(G_Stack+Stack_Hi)(DI), DI
	SUBL	DX, DI
	MOVL	DI, 4(SP)	// Save depth in stack (can't just Save SP, as stack might be copied during a callback)
	MOVL	BX, 0(SP)	// first argument in x86-32 ABI
	CALL	AX

	// Restore registers, g, stack pointer.
	get_tls(CX)
	MOVL	8(SP), DI
	MOVL	(G_Stack+Stack_Hi)(DI), SI
	SUBL	4(SP), SI
	MOVL	DI, g(CX)
	MOVL	SI, SP
	RET

// cgocallback(void (*fn)(void*), void *frame, uintptr framesize)
// Turn the fn into a Go func (by taking its address) and call
// cgocallback_gofunc.
TEXT runtime·cgocallback(SB),NOSPLIT,$12-12
	LEAL	fn+0(FP), AX
	MOVL	AX, 0(SP)
	MOVL	frame+4(FP), AX
	MOVL	AX, 4(SP)
	MOVL	framesize+8(FP), AX
	MOVL	AX, 8(SP)
	MOVL	$runtime·cgocallback_gofunc(SB), AX
	CALL	AX
	RET

// cgocallback_gofunc(FuncVal*, void *frame, uintptr framesize)
// See cgocall.c for more details.
TEXT runtime·cgocallback_gofunc(SB),NOSPLIT,$12-12
	NO_LOCAL_POINTERS

	// If g is nil, Go did not create the current thread.
	// Call needm to obtain one for temporary use.
	// In this case, we're running on the thread stack, so there's
	// lots of space, but the linker doesn't know. Hide the call from
	// the linker analysis by using an indirect call through AX.
	get_tls(CX)
#ifdef GOOS_windows
	MOVL	$0, BP
	CMPL	CX, $0
	JEQ	2(PC) // TODO
#endif
	MOVL	g(CX), BP
	CMPL	BP, $0
	JEQ	needm
	MOVL	G_M(BP), BP
	MOVL	BP, DX // saved copy of oldm
	JMP	havem
needm:
	MOVL	$0, 0(SP)
	MOVL	$runtime∕internal∕core·needm(SB), AX
	CALL	AX
	MOVL	0(SP), DX
	get_tls(CX)
	MOVL	g(CX), BP
	MOVL	G_M(BP), BP

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
	MOVL	M_G0(BP), SI
	MOVL	SP, (G_Sched+Gobuf_Sp)(SI)

havem:
	// Now there's a valid m, and we're running on its m->g0.
	// Save current m->g0->Sched.sp on stack and then set it to SP.
	// Save current sp in m->g0->Sched.sp in preparation for
	// switch back to m->curg stack.
	// NOTE: unwindm knows that the saved g->Sched.sp is at 0(SP).
	MOVL	M_G0(BP), SI
	MOVL	(G_Sched+Gobuf_Sp)(SI), AX
	MOVL	AX, 0(SP)
	MOVL	SP, (G_Sched+Gobuf_Sp)(SI)

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
	// In the new goroutine, 0(SP) holds the saved oldm (DX) register.
	// 4(SP) and 8(SP) are unused.
	MOVL	M_Curg(BP), SI
	MOVL	SI, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(SI), DI // prepare stack as DI
	MOVL	(G_Sched+Gobuf_Pc)(SI), BP
	MOVL	BP, -4(DI)
	LEAL	-(4+12)(DI), SP
	MOVL	DX, 0(SP)
	CALL	runtime∕internal∕cgo·cgocallbackg(SB)
	MOVL	0(SP), DX

	// Restore g->Sched (== m->curg->Sched) from saved values.
	get_tls(CX)
	MOVL	g(CX), SI
	MOVL	12(SP), BP
	MOVL	BP, (G_Sched+Gobuf_Pc)(SI)
	LEAL	(12+4)(SP), DI
	MOVL	DI, (G_Sched+Gobuf_Sp)(SI)

	// Switch back to m->g0's stack and restore m->g0->Sched.sp.
	// (Unlike m->curg, the g0 goroutine never uses Sched.pc,
	// so we do not have to restore it.)
	MOVL	g(CX), BP
	MOVL	G_M(BP), BP
	MOVL	M_G0(BP), SI
	MOVL	SI, g(CX)
	MOVL	(G_Sched+Gobuf_Sp)(SI), SP
	MOVL	0(SP), AX
	MOVL	AX, (G_Sched+Gobuf_Sp)(SI)
	
	// If the m on entry was nil, we called needm above to borrow an m
	// for the duration of the call. Since the call is over, return it with dropm.
	CMPL	DX, $0
	JNE 3(PC)
	MOVL	$runtime·dropm(SB), AX
	CALL	AX

	// Done!
	RET

// void Setg(G*); set g. for use by needm.
TEXT runtime∕internal∕core·Setg(SB), NOSPLIT, $0-4
	MOVL	gg+0(FP), BX
#ifdef GOOS_windows
	CMPL	BX, $0
	JNE	settls
	MOVL	$0, 0x14(FS)
	RET
settls:
	MOVL	G_M(BX), AX
	LEAL	M_Tls(AX), AX
	MOVL	AX, 0x14(FS)
#endif
	get_tls(CX)
	MOVL	BX, g(CX)
	RET

// void setg_gcc(G*); set g. for use by gcc
TEXT setg_gcc<>(SB), NOSPLIT, $0
	get_tls(AX)
	MOVL	gg+0(FP), DX
	MOVL	DX, g(AX)
	RET

// check that SP is in range [g->stack.lo, g->stack.hi)
TEXT runtime·stackcheck(SB), NOSPLIT, $0-0
	get_tls(CX)
	MOVL	g(CX), AX
	CMPL	(G_Stack+Stack_Hi)(AX), SP
	JHI	2(PC)
	INT	$3
	CMPL	SP, (G_Stack+Stack_Lo)(AX)
	JHI	2(PC)
	INT	$3
	RET

TEXT runtime∕internal∕lock·Getcallerpc(SB),NOSPLIT,$0-8
	MOVL	argp+0(FP),AX		// addr of first arg
	MOVL	-4(AX),AX		// get calling pc
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·gogetcallerpc(SB),NOSPLIT,$0-8
	MOVL	p+0(FP),AX		// addr of first arg
	MOVL	-4(AX),AX		// get calling pc
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime∕internal∕channels·setcallerpc(SB),NOSPLIT,$0-8
	MOVL	argp+0(FP),AX		// addr of first arg
	MOVL	pc+4(FP), BX
	MOVL	BX, -4(AX)		// set calling pc
	RET

TEXT runtime∕internal∕lock·Getcallersp(SB), NOSPLIT, $0-8
	MOVL	argp+0(FP), AX
	MOVL	AX, ret+4(FP)
	RET

// func gogetcallersp(p unsafe.Pointer) uintptr
TEXT runtime·gogetcallersp(SB),NOSPLIT,$0-8
	MOVL	p+0(FP),AX		// addr of first arg
	MOVL	AX, ret+4(FP)
	RET

// int64 runtime∕internal∕sched·Cputicks(void), so really
// void runtime∕internal∕sched·Cputicks(int64 *ticks)
TEXT runtime∕internal∕sched·Cputicks(SB),NOSPLIT,$0-8
	RDTSC
	MOVL	AX, ret_lo+0(FP)
	MOVL	DX, ret_hi+4(FP)
	RET

TEXT runtime·ldt0setup(SB),NOSPLIT,$16-0
	// set up ldt 7 to point at tls0
	// ldt 1 would be fine on Linux, but on OS X, 7 is as low as we can go.
	// the entry number is just a hint.  setldt will set up GS with what it used.
	MOVL	$7, 0(SP)
	LEAL	runtime·tls0(SB), AX
	MOVL	AX, 4(SP)
	MOVL	$32, 8(SP)	// sizeof(tls array)
	CALL	runtime·setldt(SB)
	RET

TEXT runtime·emptyfunc(SB),0,$0-0
	RET

TEXT runtime·abort(SB),NOSPLIT,$0-0
	INT $0x3

// Hash function using AES hardware instructions
TEXT runtime∕internal∕hash·aeshash(SB),NOSPLIT,$0-16
	MOVL	p+0(FP), AX	// Ptr to Data
	MOVL	s+4(FP), CX	// size
	JMP	runtime·aeshashbody(SB)

TEXT runtime∕internal∕hash·aeshashstr(SB),NOSPLIT,$0-16
	MOVL	p+0(FP), AX	// Ptr to string object
	// s+4(FP) is ignored, it is always sizeof(String)
	MOVL	4(AX), CX	// length of string
	MOVL	(AX), AX	// string Data
	JMP	runtime·aeshashbody(SB)

// AX: Data
// CX: length
TEXT runtime·aeshashbody(SB),NOSPLIT,$0-16
	MOVL	h+8(FP), X6	// seed to low 64 bits of xmm6
	PINSRD	$2, CX, X6	// size to high 64 bits of xmm6
	PSHUFHW	$0, X6, X6	// replace size with its low 2 Bytes repeated 4 times
	MOVO	runtime∕internal∕hash·aeskeysched(SB), X7
	CMPL	CX, $16
	JB	aes0to15
	JE	aes16
	CMPL	CX, $32
	JBE	aes17to32
	CMPL	CX, $64
	JBE	aes33to64
	JMP	aes65plus
	
aes0to15:
	TESTL	CX, CX
	JE	aes0

	ADDL	$16, AX
	TESTW	$0xff0, AX
	JE	endofpage

	// 16 Bytes loaded at this address won't cross
	// a page boundary, so we can load it directly.
	MOVOU	-16(AX), X0
	ADDL	CX, CX
	PAND	masks<>(SB)(CX*8), X0

	// scramble 3 times
	AESENC	X6, X0
	AESENC	X7, X0
	AESENC	X7, X0
	MOVL	X0, ret+12(FP)
	RET

endofpage:
	// address ends in 1111xxxx.  Might be up against
	// a page boundary, so load ending at last byte.
	// Then const_Shift Bytes down using pshufb.
	MOVOU	-32(AX)(CX*1), X0
	ADDL	CX, CX
	PSHUFB	shifts<>(SB)(CX*8), X0
	AESENC	X6, X0
	AESENC	X7, X0
	AESENC	X7, X0
	MOVL	X0, ret+12(FP)
	RET

aes0:
	// return input seed
	MOVL	h+8(FP), AX
	MOVL	AX, ret+12(FP)
	RET

aes16:
	MOVOU	(AX), X0
	AESENC	X6, X0
	AESENC	X7, X0
	AESENC	X7, X0
	MOVL	X0, ret+12(FP)
	RET


aes17to32:
	// load Data to be hashed
	MOVOU	(AX), X0
	MOVOU	-16(AX)(CX*1), X1

	// scramble 3 times
	AESENC	X6, X0
	AESENC	runtime∕internal∕hash·aeskeysched+16(SB), X1
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X0
	AESENC	X7, X1

	// combine results
	PXOR	X1, X0
	MOVL	X0, ret+12(FP)
	RET

aes33to64:
	MOVOU	(AX), X0
	MOVOU	16(AX), X1
	MOVOU	-32(AX)(CX*1), X2
	MOVOU	-16(AX)(CX*1), X3
	
	AESENC	X6, X0
	AESENC	runtime∕internal∕hash·aeskeysched+16(SB), X1
	AESENC	runtime∕internal∕hash·aeskeysched+32(SB), X2
	AESENC	runtime∕internal∕hash·aeskeysched+48(SB), X3
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X2
	AESENC	X7, X3
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X2
	AESENC	X7, X3

	PXOR	X2, X0
	PXOR	X3, X1
	PXOR	X1, X0
	MOVL	X0, ret+12(FP)
	RET

aes65plus:
	// start with last (possibly overlapping) block
	MOVOU	-64(AX)(CX*1), X0
	MOVOU	-48(AX)(CX*1), X1
	MOVOU	-32(AX)(CX*1), X2
	MOVOU	-16(AX)(CX*1), X3

	// scramble state once
	AESENC	X6, X0
	AESENC	runtime∕internal∕hash·aeskeysched+16(SB), X1
	AESENC	runtime∕internal∕hash·aeskeysched+32(SB), X2
	AESENC	runtime∕internal∕hash·aeskeysched+48(SB), X3

	// compute number of remaining 64-byte blocks
	DECL	CX
	SHRL	$6, CX
	
aesloop:
	// scramble state, xor in a block
	MOVOU	(AX), X4
	MOVOU	16(AX), X5
	AESENC	X4, X0
	AESENC	X5, X1
	MOVOU	32(AX), X4
	MOVOU	48(AX), X5
	AESENC	X4, X2
	AESENC	X5, X3

	// scramble state
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X2
	AESENC	X7, X3

	ADDL	$64, AX
	DECL	CX
	JNE	aesloop

	// 2 more scrambles to finish
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X2
	AESENC	X7, X3
	AESENC	X7, X0
	AESENC	X7, X1
	AESENC	X7, X2
	AESENC	X7, X3

	PXOR	X2, X0
	PXOR	X3, X1
	PXOR	X1, X0
	MOVL	X0, ret+12(FP)
	RET

TEXT runtime∕internal∕hash·aeshash32(SB),NOSPLIT,$0-16
	MOVL	p+0(FP), AX	// Ptr to Data
	// s+4(FP) is ignored, it is always sizeof(int32)
	MOVL	h+8(FP), X0	// seed
	PINSRD	$1, (AX), X0	// Data
	AESENC	runtime∕internal∕hash·aeskeysched+0(SB), X0
	AESENC	runtime∕internal∕hash·aeskeysched+16(SB), X0
	AESENC	runtime∕internal∕hash·aeskeysched+32(SB), X0
	MOVL	X0, ret+12(FP)
	RET

TEXT runtime∕internal∕hash·aeshash64(SB),NOSPLIT,$0-16
	MOVL	p+0(FP), AX	// Ptr to Data
	// s+4(FP) is ignored, it is always sizeof(int64)
	MOVQ	(AX), X0	// Data
	PINSRD	$2, h+8(FP), X0	// seed
	AESENC	runtime∕internal∕hash·aeskeysched+0(SB), X0
	AESENC	runtime∕internal∕hash·aeskeysched+16(SB), X0
	AESENC	runtime∕internal∕hash·aeskeysched+32(SB), X0
	MOVL	X0, ret+12(FP)
	RET

// simple const_Mask to get rid of Data in the high part of the register.
DATA masks<>+0x00(SB)/4, $0x00000000
DATA masks<>+0x04(SB)/4, $0x00000000
DATA masks<>+0x08(SB)/4, $0x00000000
DATA masks<>+0x0c(SB)/4, $0x00000000
	
DATA masks<>+0x10(SB)/4, $0x000000ff
DATA masks<>+0x14(SB)/4, $0x00000000
DATA masks<>+0x18(SB)/4, $0x00000000
DATA masks<>+0x1c(SB)/4, $0x00000000
	
DATA masks<>+0x20(SB)/4, $0x0000ffff
DATA masks<>+0x24(SB)/4, $0x00000000
DATA masks<>+0x28(SB)/4, $0x00000000
DATA masks<>+0x2c(SB)/4, $0x00000000
	
DATA masks<>+0x30(SB)/4, $0x00ffffff
DATA masks<>+0x34(SB)/4, $0x00000000
DATA masks<>+0x38(SB)/4, $0x00000000
DATA masks<>+0x3c(SB)/4, $0x00000000
	
DATA masks<>+0x40(SB)/4, $0xffffffff
DATA masks<>+0x44(SB)/4, $0x00000000
DATA masks<>+0x48(SB)/4, $0x00000000
DATA masks<>+0x4c(SB)/4, $0x00000000
	
DATA masks<>+0x50(SB)/4, $0xffffffff
DATA masks<>+0x54(SB)/4, $0x000000ff
DATA masks<>+0x58(SB)/4, $0x00000000
DATA masks<>+0x5c(SB)/4, $0x00000000
	
DATA masks<>+0x60(SB)/4, $0xffffffff
DATA masks<>+0x64(SB)/4, $0x0000ffff
DATA masks<>+0x68(SB)/4, $0x00000000
DATA masks<>+0x6c(SB)/4, $0x00000000
	
DATA masks<>+0x70(SB)/4, $0xffffffff
DATA masks<>+0x74(SB)/4, $0x00ffffff
DATA masks<>+0x78(SB)/4, $0x00000000
DATA masks<>+0x7c(SB)/4, $0x00000000
	
DATA masks<>+0x80(SB)/4, $0xffffffff
DATA masks<>+0x84(SB)/4, $0xffffffff
DATA masks<>+0x88(SB)/4, $0x00000000
DATA masks<>+0x8c(SB)/4, $0x00000000
	
DATA masks<>+0x90(SB)/4, $0xffffffff
DATA masks<>+0x94(SB)/4, $0xffffffff
DATA masks<>+0x98(SB)/4, $0x000000ff
DATA masks<>+0x9c(SB)/4, $0x00000000
	
DATA masks<>+0xa0(SB)/4, $0xffffffff
DATA masks<>+0xa4(SB)/4, $0xffffffff
DATA masks<>+0xa8(SB)/4, $0x0000ffff
DATA masks<>+0xac(SB)/4, $0x00000000
	
DATA masks<>+0xb0(SB)/4, $0xffffffff
DATA masks<>+0xb4(SB)/4, $0xffffffff
DATA masks<>+0xb8(SB)/4, $0x00ffffff
DATA masks<>+0xbc(SB)/4, $0x00000000
	
DATA masks<>+0xc0(SB)/4, $0xffffffff
DATA masks<>+0xc4(SB)/4, $0xffffffff
DATA masks<>+0xc8(SB)/4, $0xffffffff
DATA masks<>+0xcc(SB)/4, $0x00000000
	
DATA masks<>+0xd0(SB)/4, $0xffffffff
DATA masks<>+0xd4(SB)/4, $0xffffffff
DATA masks<>+0xd8(SB)/4, $0xffffffff
DATA masks<>+0xdc(SB)/4, $0x000000ff
	
DATA masks<>+0xe0(SB)/4, $0xffffffff
DATA masks<>+0xe4(SB)/4, $0xffffffff
DATA masks<>+0xe8(SB)/4, $0xffffffff
DATA masks<>+0xec(SB)/4, $0x0000ffff
	
DATA masks<>+0xf0(SB)/4, $0xffffffff
DATA masks<>+0xf4(SB)/4, $0xffffffff
DATA masks<>+0xf8(SB)/4, $0xffffffff
DATA masks<>+0xfc(SB)/4, $0x00ffffff

GLOBL masks<>(SB),RODATA,$256

// these are arguments to pshufb.  They move Data down from
// the high Bytes of the register to the low Bytes of the register.
// Index is how many Bytes to move.
DATA shifts<>+0x00(SB)/4, $0x00000000
DATA shifts<>+0x04(SB)/4, $0x00000000
DATA shifts<>+0x08(SB)/4, $0x00000000
DATA shifts<>+0x0c(SB)/4, $0x00000000
	
DATA shifts<>+0x10(SB)/4, $0xffffff0f
DATA shifts<>+0x14(SB)/4, $0xffffffff
DATA shifts<>+0x18(SB)/4, $0xffffffff
DATA shifts<>+0x1c(SB)/4, $0xffffffff
	
DATA shifts<>+0x20(SB)/4, $0xffff0f0e
DATA shifts<>+0x24(SB)/4, $0xffffffff
DATA shifts<>+0x28(SB)/4, $0xffffffff
DATA shifts<>+0x2c(SB)/4, $0xffffffff
	
DATA shifts<>+0x30(SB)/4, $0xff0f0e0d
DATA shifts<>+0x34(SB)/4, $0xffffffff
DATA shifts<>+0x38(SB)/4, $0xffffffff
DATA shifts<>+0x3c(SB)/4, $0xffffffff
	
DATA shifts<>+0x40(SB)/4, $0x0f0e0d0c
DATA shifts<>+0x44(SB)/4, $0xffffffff
DATA shifts<>+0x48(SB)/4, $0xffffffff
DATA shifts<>+0x4c(SB)/4, $0xffffffff
	
DATA shifts<>+0x50(SB)/4, $0x0e0d0c0b
DATA shifts<>+0x54(SB)/4, $0xffffff0f
DATA shifts<>+0x58(SB)/4, $0xffffffff
DATA shifts<>+0x5c(SB)/4, $0xffffffff
	
DATA shifts<>+0x60(SB)/4, $0x0d0c0b0a
DATA shifts<>+0x64(SB)/4, $0xffff0f0e
DATA shifts<>+0x68(SB)/4, $0xffffffff
DATA shifts<>+0x6c(SB)/4, $0xffffffff
	
DATA shifts<>+0x70(SB)/4, $0x0c0b0a09
DATA shifts<>+0x74(SB)/4, $0xff0f0e0d
DATA shifts<>+0x78(SB)/4, $0xffffffff
DATA shifts<>+0x7c(SB)/4, $0xffffffff
	
DATA shifts<>+0x80(SB)/4, $0x0b0a0908
DATA shifts<>+0x84(SB)/4, $0x0f0e0d0c
DATA shifts<>+0x88(SB)/4, $0xffffffff
DATA shifts<>+0x8c(SB)/4, $0xffffffff
	
DATA shifts<>+0x90(SB)/4, $0x0a090807
DATA shifts<>+0x94(SB)/4, $0x0e0d0c0b
DATA shifts<>+0x98(SB)/4, $0xffffff0f
DATA shifts<>+0x9c(SB)/4, $0xffffffff
	
DATA shifts<>+0xa0(SB)/4, $0x09080706
DATA shifts<>+0xa4(SB)/4, $0x0d0c0b0a
DATA shifts<>+0xa8(SB)/4, $0xffff0f0e
DATA shifts<>+0xac(SB)/4, $0xffffffff
	
DATA shifts<>+0xb0(SB)/4, $0x08070605
DATA shifts<>+0xb4(SB)/4, $0x0c0b0a09
DATA shifts<>+0xb8(SB)/4, $0xff0f0e0d
DATA shifts<>+0xbc(SB)/4, $0xffffffff
	
DATA shifts<>+0xc0(SB)/4, $0x07060504
DATA shifts<>+0xc4(SB)/4, $0x0b0a0908
DATA shifts<>+0xc8(SB)/4, $0x0f0e0d0c
DATA shifts<>+0xcc(SB)/4, $0xffffffff
	
DATA shifts<>+0xd0(SB)/4, $0x06050403
DATA shifts<>+0xd4(SB)/4, $0x0a090807
DATA shifts<>+0xd8(SB)/4, $0x0e0d0c0b
DATA shifts<>+0xdc(SB)/4, $0xffffff0f
	
DATA shifts<>+0xe0(SB)/4, $0x05040302
DATA shifts<>+0xe4(SB)/4, $0x09080706
DATA shifts<>+0xe8(SB)/4, $0x0d0c0b0a
DATA shifts<>+0xec(SB)/4, $0xffff0f0e
	
DATA shifts<>+0xf0(SB)/4, $0x04030201
DATA shifts<>+0xf4(SB)/4, $0x08070605
DATA shifts<>+0xf8(SB)/4, $0x0c0b0a09
DATA shifts<>+0xfc(SB)/4, $0xff0f0e0d

GLOBL shifts<>(SB),RODATA,$256

TEXT runtime∕internal∕hash·Memeq(SB),NOSPLIT,$0-13
	MOVL	a+0(FP), SI
	MOVL	b+4(FP), DI
	MOVL	size+8(FP), BX
	CALL	runtime·memeqbody(SB)
	MOVB	AX, ret+12(FP)
	RET

// eqstring tests whether two strings are equal.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.
TEXT runtime·eqstring(SB),NOSPLIT,$0-17
	MOVL	s1len+4(FP), AX
	MOVL	s2len+12(FP), BX
	CMPL	AX, BX
	JNE	different
	MOVL	s1str+0(FP), SI
	MOVL	s2str+8(FP), DI
	CMPL	SI, DI
	JEQ	same
	CALL	runtime·memeqbody(SB)
	MOVB	AX, v+16(FP)
	RET
same:
	MOVB	$1, v+16(FP)
	RET
different:
	MOVB	$0, v+16(FP)
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

// a in SI
// b in DI
// count in BX
TEXT runtime·memeqbody(SB),NOSPLIT,$0-0
	XORL	AX, AX

	CMPL	BX, $4
	JB	small

	// 64 Bytes at a time using xmm registers
hugeloop:
	CMPL	BX, $64
	JB	bigloop
	TESTL	$0x4000000, runtime·cpuid_edx(SB) // check for sse2
	JE	bigloop
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
	ADDL	$64, SI
	ADDL	$64, DI
	SUBL	$64, BX
	CMPL	DX, $0xffff
	JEQ	hugeloop
	RET

	// 4 Bytes at a time using 32-bit register
bigloop:
	CMPL	BX, $4
	JBE	leftover
	MOVL	(SI), CX
	MOVL	(DI), DX
	ADDL	$4, SI
	ADDL	$4, DI
	SUBL	$4, BX
	CMPL	CX, DX
	JEQ	bigloop
	RET

	// remaining 0-4 Bytes
leftover:
	MOVL	-4(SI)(BX*1), CX
	MOVL	-4(DI)(BX*1), DX
	CMPL	CX, DX
	SETEQ	AX
	RET

small:
	CMPL	BX, $0
	JEQ	equal

	LEAL	0(BX*8), CX
	NEGL	CX

	MOVL	SI, DX
	CMPB	DX, $0xfc
	JA	si_high

	// load at SI won't cross a page boundary.
	MOVL	(SI), SI
	JMP	si_finish
si_high:
	// address ends in 111111xx.  Load up to Bytes we want, move to correct position.
	MOVL	-4(SI)(BX*1), SI
	SHRL	CX, SI
si_finish:

	// same for DI.
	MOVL	DI, DX
	CMPB	DX, $0xfc
	JA	di_high
	MOVL	(DI), DI
	JMP	di_finish
di_high:
	MOVL	-4(DI)(BX*1), DI
	SHRL	CX, DI
di_finish:

	SUBL	SI, DI
	SHLL	CX, DI
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
	MOVL	AX, ret+24(FP)
	RET

TEXT bytes·IndexByte(SB),NOSPLIT,$0
	MOVL	s+0(FP), SI
	MOVL	s_len+4(FP), CX
	MOVB	c+12(FP), AL
	MOVL	SI, DI
	CLD; REPN; SCASB
	JZ 3(PC)
	MOVL	$-1, ret+16(FP)
	RET
	SUBL	SI, DI
	SUBL	$1, DI
	MOVL	DI, ret+16(FP)
	RET

TEXT strings·IndexByte(SB),NOSPLIT,$0
	MOVL	s+0(FP), SI
	MOVL	s_len+4(FP), CX
	MOVB	c+8(FP), AL
	MOVL	SI, DI
	CLD; REPN; SCASB
	JZ 3(PC)
	MOVL	$-1, ret+12(FP)
	RET
	SUBL	SI, DI
	SUBL	$1, DI
	MOVL	DI, ret+12(FP)
	RET

// input:
//   SI = a
//   DI = b
//   BX = alen
//   DX = blen
// output:
//   AX = 1/0/-1
TEXT runtime·cmpbody(SB),NOSPLIT,$0-0
	CMPL	SI, DI
	JEQ	allsame
	CMPL	BX, DX
	MOVL	DX, BP
	CMOVLLT	BX, BP // BP = min(alen, blen)
	CMPL	BP, $4
	JB	small
	TESTL	$0x4000000, runtime·cpuid_edx(SB) // check for sse2
	JE	mediumloop
largeloop:
	CMPL	BP, $16
	JB	mediumloop
	MOVOU	(SI), X0
	MOVOU	(DI), X1
	PCMPEQB X0, X1
	PMOVMSKB X1, AX
	XORL	$0xffff, AX	// convert EQ to NE
	JNE	diff16	// branch if at least one byte is not equal
	ADDL	$16, SI
	ADDL	$16, DI
	SUBL	$16, BP
	JMP	largeloop

diff16:
	BSFL	AX, BX	// Index of first byte that differs
	XORL	AX, AX
	MOVB	(SI)(BX*1), CX
	CMPB	CX, (DI)(BX*1)
	SETHI	AX
	LEAL	-1(AX*2), AX	// convert 1/0 to +1/-1
	RET

mediumloop:
	CMPL	BP, $4
	JBE	_0through4
	MOVL	(SI), AX
	MOVL	(DI), CX
	CMPL	AX, CX
	JNE	diff4
	ADDL	$4, SI
	ADDL	$4, DI
	SUBL	$4, BP
	JMP	mediumloop

_0through4:
	MOVL	-4(SI)(BP*1), AX
	MOVL	-4(DI)(BP*1), CX
	CMPL	AX, CX
	JEQ	allsame

diff4:
	BSWAPL	AX	// reverse order of Bytes
	BSWAPL	CX
	XORL	AX, CX	// find bit differences
	BSRL	CX, CX	// Index of highest bit difference
	SHRL	CX, AX	// move a's bit to bottom
	ANDL	$1, AX	// const_Mask bit
	LEAL	-1(AX*2), AX // 1/0 => +1/-1
	RET

	// 0-3 Bytes in common
small:
	LEAL	(BP*8), CX
	NEGL	CX
	JEQ	allsame

	// load si
	CMPB	SI, $0xfc
	JA	si_high
	MOVL	(SI), SI
	JMP	si_finish
si_high:
	MOVL	-4(SI)(BP*1), SI
	SHRL	CX, SI
si_finish:
	SHLL	CX, SI

	// same for di
	CMPB	DI, $0xfc
	JA	di_high
	MOVL	(DI), DI
	JMP	di_finish
di_high:
	MOVL	-4(DI)(BP*1), DI
	SHRL	CX, DI
di_finish:
	SHLL	CX, DI

	BSWAPL	SI	// reverse order of Bytes
	BSWAPL	DI
	XORL	SI, DI	// find bit differences
	JEQ	allsame
	BSRL	DI, CX	// Index of highest bit difference
	SHRL	CX, SI	// move a's bit to bottom
	ANDL	$1, SI	// const_Mask bit
	LEAL	-1(SI*2), AX // 1/0 => +1/-1
	RET

	// all the Bytes in common are the same, so we just need
	// to compare the lengths.
allsame:
	XORL	AX, AX
	XORL	CX, CX
	CMPL	BX, DX
	SETGT	AX	// 1 if alen > blen
	SETEQ	CX	// 1 if alen == blen
	LEAL	-1(CX)(AX*2), AX	// 1,0,-1 result
	RET

// A Duff's device for zeroing memory.
// The compiler jumps to computed addresses within
// this routine to zero chunks of memory.  Do not
// change this code without also changing the code
// in ../../cmd/8g/ggen.c:clearfat.
// AX: zero
// DI: Ptr to memory to be zeroed
// DI is updated as a side effect.
TEXT runtime·duffzero(SB), NOSPLIT, $0-0
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	STOSL
	RET

// A Duff's device for copying memory.
// The compiler jumps to computed addresses within
// this routine to copy chunks of memory.  Source
// and destination must not overlap.  Do not
// change this code without also changing the code
// in ../../cmd/6g/cgen.c:sgen.
// SI: Ptr to source memory
// DI: Ptr to destination memory
// SI and DI are updated as a side effect.

// NOTE: this is equivalent to a sequence of MOVSL but
// for some reason MOVSL is really slow.
TEXT runtime·duffcopy(SB), NOSPLIT, $0-0
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
	MOVL	(SI),CX
	ADDL	$4,SI
	MOVL	CX,(DI)
	ADDL	$4,DI
	
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

// Called from cgo wrappers, this function returns g->m->curg.stack.hi.
// Must obey the gcc calling convention.
TEXT _cgo_topofstack(SB),NOSPLIT,$0
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	G_M(AX), AX
	MOVL	M_Curg(AX), AX
	MOVL	(G_Stack+Stack_Hi)(AX), AX
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
