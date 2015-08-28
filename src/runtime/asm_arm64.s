// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "tls_arm64.h"
#include "Funcdata.h"
#include "textflag.h"

TEXT runtime·rt0_go(SB),NOSPLIT,$0
	// SP = stack; R0 = argc; R1 = argv

	// initialize essential registers
	BL	runtime·reginit(SB)

	SUB	$32, RSP
	MOVW	R0, 8(RSP) // argc
	MOVD	R1, 16(RSP) // argv

	// create istack out of the given (operating system) stack.
	// _cgo_init may update stackguard.
	MOVD	$runtime·g0(SB), g
	MOVD RSP, R7
	MOVD	$(-64*1024)(R7), R0
	MOVD	R0, G_Stackguard0(g)
	MOVD	R0, G_stackguard1(g)
	MOVD	R0, (G_Stack+Stack_Lo)(g)
	MOVD	R7, (G_Stack+Stack_Hi)(g)

	// if there is a _cgo_init, call it using the gcc ABI.
	MOVD	_cgo_init(SB), R12
	CMP	$0, R12
	BEQ	nocgo

	MRS_TPIDR_R0			// load TLS base pointer
	MOVD	R0, R3			// arg 3: TLS base pointer
#ifdef TLSG_IS_VARIABLE
	MOVD	$runtime·tls_g(SB), R2 	// arg 2: tlsg
#else
	MOVD	$0x10, R2		// arg 2: tlsg TODO(minux): hardcoded for linux
#endif
	MOVD	$setg_gcc<>(SB), R1	// arg 1: setg
	MOVD	g, R0			// arg 0: G
	BL	(R12)
	MOVD	_cgo_init(SB), R12
	CMP	$0, R12
	BEQ	nocgo

nocgo:
	// update stackguard after _cgo_init
	MOVD	(G_Stack+Stack_Lo)(g), R0
	ADD	$const_StackGuard, R0
	MOVD	R0, G_Stackguard0(g)
	MOVD	R0, G_stackguard1(g)

	// set the per-goroutine and per-mach "registers"
	MOVD	$runtime∕internal∕base·M0(SB), R0

	// Save m->g0 = g0
	MOVD	g, M_G0(R0)
	// Save M0 to g0->m
	MOVD	R0, G_M(g)

	BL	runtime·check(SB)

	MOVW	8(RSP), R0	// copy argc
	MOVW	R0, -8(RSP)
	MOVD	16(RSP), R0		// copy argv
	MOVD	R0, 0(RSP)
	BL	runtime·args(SB)
	BL	runtime·osinit(SB)
	BL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOVD	$runtime·mainPC(SB), R0		// entry
	MOVD	RSP, R7
	MOVD.W	$0, -8(R7)
	MOVD.W	R0, -8(R7)
	MOVD.W	$0, -8(R7)
	MOVD.W	$0, -8(R7)
	MOVD	R7, RSP
	BL	runtime·newproc(SB)
	ADD	$32, RSP

	// start this M
	BL	runtime∕internal∕base·Mstart(SB)

	MOVD	$0, R0
	MOVD	R0, (R0)	// boom
	UNDEF

DATA	runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOBL	runtime·mainPC(SB),RODATA,$8

TEXT runtime·breakpoint(SB),NOSPLIT,$-8-0
	BRK
	RET

TEXT runtime∕internal∕base·Asminit(SB),NOSPLIT,$-8-0
	RET

TEXT runtime·reginit(SB),NOSPLIT,$-8-0
	// initialize essential FP registers
	FMOVD	$4503601774854144.0, F27
	FMOVD	$0.5, F29
	FSUBD	F29, F29, F28
	FADDD	F29, F29, F30
	FADDD	F30, F30, F31
	RET

/*
 *  go-routine
 */

// void gosave(Gobuf*)
// Save state in Gobuf; setjmp
TEXT runtime∕internal∕base·gosave(SB), NOSPLIT, $-8-8
	MOVD	buf+0(FP), R3
	MOVD	RSP, R0
	MOVD	R0, Gobuf_Sp(R3)
	MOVD	LR, Gobuf_Pc(R3)
	MOVD	g, Gobuf_G(R3)
	MOVD	ZR, Gobuf_Lr(R3)
	MOVD	ZR, Gobuf_Ret(R3)
	MOVD	ZR, Gobuf_Ctxt(R3)
	RET

// void Gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT runtime∕internal∕base·Gogo(SB), NOSPLIT, $-8-8
	MOVD	buf+0(FP), R5
	MOVD	Gobuf_G(R5), g
	BL	runtime·save_g(SB)

	MOVD	0(g), R4	// make sure g is not nil
	MOVD	Gobuf_Sp(R5), R0
	MOVD	R0, RSP
	MOVD	Gobuf_Lr(R5), LR
	MOVD	Gobuf_Ret(R5), R0
	MOVD	Gobuf_Ctxt(R5), R26
	MOVD	$0, Gobuf_Sp(R5)
	MOVD	$0, Gobuf_Ret(R5)
	MOVD	$0, Gobuf_Lr(R5)
	MOVD	$0, Gobuf_Ctxt(R5)
	CMP	ZR, ZR // set condition codes for == test, needed by stack split
	MOVD	Gobuf_Pc(R5), R6
	B	(R6)

// void Mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return.  It should Gogo(&g->Sched)
// to keep running g.
TEXT runtime∕internal∕base·Mcall(SB), NOSPLIT, $-8-8
	// Save caller state in g->Sched
	MOVD	RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(g)
	MOVD	LR, (G_Sched+Gobuf_Pc)(g)
	MOVD	$0, (G_Sched+Gobuf_Lr)(g)
	MOVD	g, (G_Sched+Gobuf_G)(g)

	// Switch to m->g0 & its stack, call fn.
	MOVD	g, R3
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), g
	BL	runtime·save_g(SB)
	CMP	g, R3
	BNE	2(PC)
	B	runtime·badmcall(SB)
	MOVD	fn+0(FP), R26			// context
	MOVD	0(R26), R4			// code pointer
	MOVD	(G_Sched+Gobuf_Sp)(g), R0
	MOVD	R0, RSP	// sp = m->g0->Sched.sp
	MOVD	R3, -8(RSP)
	MOVD	$0, -16(RSP)
	SUB	$16, RSP
	BL	(R4)
	B	runtime·badmcall2(SB)

// systemstack_switch is a dummy routine that Systemstack leaves at the bottom
// of the G stack.  We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime·systemstack_switch(SB), NOSPLIT, $0-0
	UNDEF
	BL	(LR)	// make sure this function is not leaf
	RET

// func Systemstack(fn func())
TEXT runtime∕internal∕base·Systemstack(SB), NOSPLIT, $0-8
	MOVD	fn+0(FP), R3	// R3 = fn
	MOVD	R3, R26		// context
	MOVD	G_M(g), R4	// R4 = m

	MOVD	M_Gsignal(R4), R5	// R5 = gsignal
	CMP	g, R5
	BEQ	noswitch

	MOVD	M_G0(R4), R5	// R5 = g0
	CMP	g, R5
	BEQ	noswitch

	MOVD	M_Curg(R4), R6
	CMP	g, R6
	BEQ	switch

	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOVD	$runtime·badsystemstack(SB), R3
	BL	(R3)

switch:
	// Save our state in g->Sched.  Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVD	$runtime·systemstack_switch(SB), R6
	ADD	$8, R6	// get past prologue
	MOVD	R6, (G_Sched+Gobuf_Pc)(g)
	MOVD	RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(g)
	MOVD	$0, (G_Sched+Gobuf_Lr)(g)
	MOVD	g, (G_Sched+Gobuf_G)(g)

	// switch to g0
	MOVD	R5, g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R3
	// make it look like Mstart called Systemstack on g0, to stop Traceback
	SUB	$16, R3
	AND	$~15, R3
	MOVD	$runtime∕internal∕base·Mstart(SB), R4
	MOVD	R4, 0(R3)
	MOVD	R3, RSP

	// call target function
	MOVD	0(R26), R3	// code pointer
	BL	(R3)

	// switch back to g
	MOVD	G_M(g), R3
	MOVD	M_Curg(R3), g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R0
	MOVD	R0, RSP
	MOVD	$0, (G_Sched+Gobuf_Sp)(g)
	RET

noswitch:
	// already on m stack, just call directly
	MOVD	0(R26), R3	// code pointer
	BL	(R3)
	RET

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
// Caller has already loaded:
// R3 prolog's LR (R30)
//
// The Traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling Newm calling Gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime·morestack(SB),NOSPLIT,$-8-0
	// Cannot grow scheduler stack (m->g0).
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), R4
	CMP	g, R4
	BNE	2(PC)
	B	runtime·abort(SB)

	// Cannot grow signal stack (m->gsignal).
	MOVD	M_Gsignal(R8), R4
	CMP	g, R4
	BNE	2(PC)
	B	runtime·abort(SB)

	// Called from f.
	// Set g->Sched to context in f
	MOVD	R26, (G_Sched+Gobuf_Ctxt)(g)
	MOVD	RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(g)
	MOVD	LR, (G_Sched+Gobuf_Pc)(g)
	MOVD	R3, (G_Sched+Gobuf_Lr)(g)

	// Called from f.
	// Set m->morebuf to f's Callers.
	MOVD	R3, (M_Morebuf+Gobuf_Pc)(R8)	// f's caller's PC
	MOVD	RSP, R0
	MOVD	R0, (M_Morebuf+Gobuf_Sp)(R8)	// f's caller's RSP
	MOVD	g, (M_Morebuf+Gobuf_G)(R8)

	// Call newstack on m->g0's stack.
	MOVD	M_G0(R8), g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R0
	MOVD	R0, RSP
	BL	runtime·newstack(SB)

	// Not reached, but make sure the return PC from the call to newstack
	// is still in this function, and not the beginning of the next.
	UNDEF

TEXT runtime·morestack_noctxt(SB),NOSPLIT,$-4-0
	MOVW	$0, R26
	B runtime·morestack(SB)

TEXT runtime·stackBarrier(SB),NOSPLIT,$0
	// We came here via a RET to an overwritten LR.
	// R0 may be live (see return0). Other registers are available.

	// Get the original return PC, g.stkbar[g.stkbarPos].savedLRVal.
	MOVD	(G_Stkbar+Slice_Array)(g), R4
	MOVD	G_StkbarPos(g), R5
	MOVD	$stkbar__size, R6
	MUL	R5, R6
	ADD	R4, R6
	MOVD	Stkbar_SavedLRVal(R6), R6
	// Record that this stack barrier was hit.
	ADD	$1, R5
	MOVD	R5, G_StkbarPos(g)
	// Jump to the original return PC.
	B	(R6)

// reflectcall: call a function with the given argument list
// func call(argtype *_type, f *FuncVal, arg *byte, argsize, retoffset uint32).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)		\
	MOVD	$MAXSIZE, R27;		\
	CMP	R27, R16;		\
	BGT	3(PC);			\
	MOVD	$NAME(SB), R27;	\
	B	(R27)
// Note: can't just "B NAME(SB)" - const_bad inlining results.

TEXT reflect·call(SB), NOSPLIT, $0-0
	B	runtime·reflectcall(SB)

TEXT ·reflectcall(SB), NOSPLIT, $-8-32
	MOVWU argsize+24(FP), R16
	// NOTE(rsc): No call16, because CALLFN needs four words
	// of argument space to invoke callwritebarrier.
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
	MOVD	$runtime·badreflectcall(SB), R0
	B	(R0)

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-24;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVD	arg+16(FP), R3;			\
	MOVWU	argsize+24(FP), R4;			\
	MOVD	RSP, R5;				\
	ADD	$(8-1), R5;			\
	SUB	$1, R3;				\
	ADD	R5, R4;				\
	CMP	R5, R4;				\
	BEQ	4(PC);				\
	MOVBU.W	1(R3), R6;			\
	MOVBU.W	R6, 1(R5);			\
	B	-4(PC);				\
	/* call function */			\
	MOVD	f+8(FP), R26;			\
	MOVD	(R26), R0;			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	BL	(R0);				\
	/* copy return values back */		\
	MOVD	arg+16(FP), R3;			\
	MOVWU	n+24(FP), R4;			\
	MOVWU	retoffset+28(FP), R6;		\
	MOVD	RSP, R5;				\
	ADD	R6, R5; 			\
	ADD	R6, R3;				\
	SUB	R6, R4;				\
	ADD	$(8-1), R5;			\
	SUB	$1, R3;				\
	ADD	R5, R4;				\
loop:						\
	CMP	R5, R4;				\
	BEQ	end;				\
	MOVBU.W	1(R5), R6;			\
	MOVBU.W	R6, 1(R3);			\
	B	loop;				\
end:						\
	/* execute Write barrier updates */	\
	MOVD	argtype+0(FP), R7;		\
	MOVD	arg+16(FP), R3;			\
	MOVWU	n+24(FP), R4;			\
	MOVWU	retoffset+28(FP), R6;		\
	MOVD	R7, 8(RSP);			\
	MOVD	R3, 16(RSP);			\
	MOVD	R4, 24(RSP);			\
	MOVD	R6, 32(RSP);			\
	BL	runtime·callwritebarrier(SB);	\
	RET

// These have 8 added to make the overall frame size a multiple of 16,
// as required by the ABI. (There is another +8 for the saved LR.)
CALLFN(runtime·call16, 24 )
CALLFN(runtime·call32, 40 )
CALLFN(runtime·call64, 72 )
CALLFN(runtime·call128, 136 )
CALLFN(runtime·call256, 264 )
CALLFN(runtime·call512, 520 )
CALLFN(runtime·call1024, 1032 )
CALLFN(runtime·call2048, 2056 )
CALLFN(runtime·call4096, 4104 )
CALLFN(runtime·call8192, 8200 )
CALLFN(runtime·call16384, 16392 )
CALLFN(runtime·call32768, 32776 )
CALLFN(runtime·call65536, 65544 )
CALLFN(runtime·call131072, 131080 )
CALLFN(runtime·call262144, 262152 )
CALLFN(runtime·call524288, 524296 )
CALLFN(runtime·call1048576, 1048584 )
CALLFN(runtime·call2097152, 2097160 )
CALLFN(runtime·call4194304, 4194312 )
CALLFN(runtime·call8388608, 8388616 )
CALLFN(runtime·call16777216, 16777224 )
CALLFN(runtime·call33554432, 33554440 )
CALLFN(runtime·call67108864, 67108872 )
CALLFN(runtime·call134217728, 134217736 )
CALLFN(runtime·call268435456, 268435464 )
CALLFN(runtime·call536870912, 536870920 )
CALLFN(runtime·call1073741824, 1073741832 )

// bool Cas(uint32 *ptr, uint32 old, uint32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕base·Cas(SB), NOSPLIT, $0-17
	MOVD	ptr+0(FP), R0
	MOVW	old+8(FP), R1
	MOVW	new+12(FP), R2
again:
	LDAXRW	(R0), R3
	CMPW	R1, R3
	BNE	ok
	STLXRW	R2, (R0), R3
	CBNZ	R3, again
ok:
	CSET	EQ, R0
	MOVB	R0, ret+16(FP)
	RET

TEXT runtime∕internal∕base·Casuintptr(SB), NOSPLIT, $0-25
	B	runtime∕internal∕base·Cas64(SB)

TEXT runtime∕internal∕base·Atomicloaduintptr(SB), NOSPLIT, $-8-16
	B	runtime∕internal∕base·Atomicload64(SB)

TEXT runtime∕internal∕iface·Atomicloaduint(SB), NOSPLIT, $-8-16
	B	runtime∕internal∕base·Atomicload64(SB)

TEXT runtime∕internal∕base·atomicstoreuintptr(SB), NOSPLIT, $0-16
	B	runtime∕internal∕base·Atomicstore64(SB)

// AES hashing not implemented for ARM64, issue #10109.
TEXT runtime∕internal∕base·aeshash(SB),NOSPLIT,$-8-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime·aeshash32(SB),NOSPLIT,$-8-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime·aeshash64(SB),NOSPLIT,$-8-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT runtime·aeshashstr(SB),NOSPLIT,$-8-0
	MOVW	$0, R0
	MOVW	(R0), R1

// bool casp(void **val, void *old, void *new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime·casp1(SB), NOSPLIT, $0-25
	B runtime∕internal∕base·Cas64(SB)

TEXT runtime∕internal∕base·Procyield(SB),NOSPLIT,$0-0
	MOVWU	cycles+0(FP), R0
again:
	YIELD
	SUBW	$1, R0
	CBNZ	R0, again
	RET

// void jmpdefer(fv, sp);
// called from deferreturn.
// 1. grab stored LR for caller
// 2. sub 4 bytes to get back to BL deferreturn
// 3. BR to fn
TEXT runtime·jmpdefer(SB), NOSPLIT, $-8-16
	MOVD	0(RSP), R0
	SUB	$4, R0
	MOVD	R0, LR

	MOVD	fv+0(FP), R26
	MOVD	argp+8(FP), R0
	MOVD	R0, RSP
	SUB	$8, RSP
	MOVD	0(R26), R3
	B	(R3)

// Save state of caller into g->Sched. Smashes R0.
TEXT gosave<>(SB),NOSPLIT,$-8
	MOVD	LR, (G_Sched+Gobuf_Pc)(g)
	MOVD RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(g)
	MOVD	$0, (G_Sched+Gobuf_Lr)(g)
	MOVD	$0, (G_Sched+Gobuf_Ret)(g)
	MOVD	$0, (G_Sched+Gobuf_Ctxt)(g)
	RET

// func Asmcgocall(fn, arg unsafe.Pointer) int32
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.go for more details.
TEXT runtime∕internal∕base·Asmcgocall(SB),NOSPLIT,$0-20
	MOVD	fn+0(FP), R1
	MOVD	arg+8(FP), R0

	MOVD	RSP, R2		// Save original stack pointer
	MOVD	g, R4

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already.
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), R3
	CMP	R3, g
	BEQ	g0
	MOVD	R0, R9	// gosave<> and save_g might clobber R0
	BL	gosave<>(SB)
	MOVD	R3, g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R0
	MOVD	R0, RSP
	MOVD	R9, R0

	// Now on a scheduling stack (a pthread-created stack).
g0:
	// Save room for two of our pointers /*, plus 32 bytes of callee
	// Save area that lives on the caller stack. */
	MOVD	RSP, R13
	SUB	$16, R13
	MOVD	R13, RSP
	MOVD	R4, 0(RSP)	// Save old g on stack
	MOVD	(G_Stack+Stack_Hi)(R4), R4
	SUB	R2, R4
	MOVD	R4, 8(RSP)	// Save depth in old g stack (can't just Save SP, as stack might be copied during a callback)
	BL	(R1)
	MOVD	R0, R9

	// Restore g, stack pointer.  R0 is errno, so don't touch it
	MOVD	0(RSP), g
	BL	runtime·save_g(SB)
	MOVD	(G_Stack+Stack_Hi)(g), R5
	MOVD	8(RSP), R6
	SUB	R6, R5
	MOVD	R9, R0
	MOVD	R5, RSP

	MOVW	R0, ret+16(FP)
	RET

// cgocallback(void (*fn)(void*), void *frame, uintptr framesize)
// Turn the fn into a Go func (by taking its address) and call
// cgocallback_gofunc.
TEXT runtime·cgocallback(SB),NOSPLIT,$24-24
	MOVD	$fn+0(FP), R0
	MOVD	R0, 8(RSP)
	MOVD	frame+8(FP), R0
	MOVD	R0, 16(RSP)
	MOVD	framesize+16(FP), R0
	MOVD	R0, 24(RSP)
	MOVD	$runtime·cgocallback_gofunc(SB), R0
	BL	(R0)
	RET

// cgocallback_gofunc(FuncVal*, void *frame, uintptr framesize)
// See cgocall.go for more details.
TEXT ·cgocallback_gofunc(SB),NOSPLIT,$24-24
	NO_LOCAL_POINTERS

	// Load g from thread-local storage.
	MOVB	runtime∕internal∕base·Iscgo(SB), R3
	CMP	$0, R3
	BEQ	nocgo
	BL	runtime·load_g(SB)
nocgo:

	// If g is nil, Go did not create the current thread.
	// Call needm to obtain one for temporary use.
	// In this case, we're running on the thread stack, so there's
	// lots of space, but the linker doesn't know. Hide the call from
	// the linker analysis by using an indirect call.
	CMP	$0, g
	BNE	havem
	MOVD	g, savedm-8(SP) // g is zero, so is m.
	MOVD	$runtime·needm(SB), R0
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
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), R3
	MOVD	RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(R3)

havem:
	MOVD	G_M(g), R8
	MOVD	R8, savedm-8(SP)
	// Now there's a valid m, and we're running on its m->g0.
	// Save current m->g0->Sched.sp on stack and then set it to SP.
	// Save current sp in m->g0->Sched.sp in preparation for
	// switch back to m->curg stack.
	// NOTE: unwindm knows that the saved g->Sched.sp is at 16(RSP) aka savedsp-16(SP).
	// Beware that the frame size is actually 32.
	MOVD	M_G0(R8), R3
	MOVD	(G_Sched+Gobuf_Sp)(R3), R4
	MOVD	R4, savedsp-16(SP)
	MOVD	RSP, R0
	MOVD	R0, (G_Sched+Gobuf_Sp)(R3)

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
	// so that the Traceback will seamlessly Trace back into
	// the earlier calls.
	//
	// In the new goroutine, -16(SP) and -8(SP) are unused.
	MOVD	M_Curg(R8), g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R4 // prepare stack as R4
	MOVD	(G_Sched+Gobuf_Pc)(g), R5
	MOVD	R5, -(24+8)(R4)	// maintain 16-byte SP alignment
	MOVD	$-(24+8)(R4), R0
	MOVD	R0, RSP
	BL	runtime·cgocallbackg(SB)

	// Restore g->Sched (== m->curg->Sched) from saved values.
	MOVD	0(RSP), R5
	MOVD	R5, (G_Sched+Gobuf_Pc)(g)
	MOVD	RSP, R4
	ADD	$(24+8), R4, R4
	MOVD	R4, (G_Sched+Gobuf_Sp)(g)

	// Switch back to m->g0's stack and restore m->g0->Sched.sp.
	// (Unlike m->curg, the g0 goroutine never uses Sched.pc,
	// so we do not have to restore it.)
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), g
	BL	runtime·save_g(SB)
	MOVD	(G_Sched+Gobuf_Sp)(g), R0
	MOVD	R0, RSP
	MOVD	savedsp-16(SP), R4
	MOVD	R4, (G_Sched+Gobuf_Sp)(g)

	// If the m on entry was nil, we called needm above to borrow an m
	// for the duration of the call. Since the call is over, return it with dropm.
	MOVD	savedm-8(SP), R6
	CMP	$0, R6
	BNE	droppedm
	MOVD	$runtime·dropm(SB), R0
	BL	(R0)
droppedm:

	// Done!
	RET

// Called from cgo wrappers, this function returns g->m->curg.stack.hi.
// Must obey the gcc calling convention.
TEXT _cgo_topofstack(SB),NOSPLIT,$24
	// g (R28) and REGTMP (R27)  might be clobbered by load_g. They
	// are callee-Save in the gcc calling convention, so Save them.
	MOVD	R27, savedR27-8(SP)
	MOVD	g, saveG-16(SP)

	BL	runtime·load_g(SB)
	MOVD	G_M(g), R0
	MOVD	M_Curg(R0), R0
	MOVD	(G_Stack+Stack_Hi)(R0), R0

	MOVD	saveG-16(SP), g
	MOVD	savedR28-8(SP), R27
	RET

// void setg(G*); set g. for use by needm.
TEXT runtime·setg(SB), NOSPLIT, $0-8
	MOVD	gg+0(FP), g
	// This only happens if Iscgo, so jump straight to save_g
	BL	runtime·save_g(SB)
	RET

// void setg_gcc(G*); set g called from gcc
TEXT setg_gcc<>(SB),NOSPLIT,$8
	MOVD	R0, g
	MOVD	R27, savedR27-8(SP)
	BL	runtime·save_g(SB)
	MOVD	savedR27-8(SP), R27
	RET

TEXT runtime∕internal∕base·Getcallerpc(SB),NOSPLIT,$8-16
	MOVD	16(RSP), R0		// LR saved by caller
	MOVD	runtime∕internal∕base·StackBarrierPC(SB), R1
	CMP	R0, R1
	BNE	nobar
	// Get original return PC.
	BL	runtime·nextBarrierPC(SB)
	MOVD	8(RSP), R0
nobar:
	MOVD	R0, ret+8(FP)
	RET

TEXT runtime·setcallerpc(SB),NOSPLIT,$8-16
	MOVD	pc+8(FP), R0
	MOVD	16(RSP), R1
	MOVD	runtime∕internal∕base·StackBarrierPC(SB), R2
	CMP	R1, R2
	BEQ	setbar
	MOVD	R0, 16(RSP)		// set LR in caller
	RET
setbar:
	// Set the stack barrier return PC.
	MOVD	R0, 8(RSP)
	BL	runtime·setNextBarrierPC(SB)
	RET

TEXT runtime∕internal∕base·Getcallersp(SB),NOSPLIT,$0-16
	MOVD	argp+0(FP), R0
	SUB	$8, R0
	MOVD	R0, ret+8(FP)
	RET

TEXT runtime·abort(SB),NOSPLIT,$-8-0
	B	(ZR)
	UNDEF

// memhash_varlen(p unsafe.Pointer, h seed) uintptr
// redirects to Memhash(p, h, size) using the size
// stored in the closure.
TEXT runtime·memhash_varlen(SB),NOSPLIT,$40-24
	GO_ARGS
	NO_LOCAL_POINTERS
	MOVD	p+0(FP), R3
	MOVD	h+8(FP), R4
	MOVD	8(R26), R5
	MOVD	R3, 8(RSP)
	MOVD	R4, 16(RSP)
	MOVD	R5, 24(RSP)
	BL	runtime∕internal∕base·Memhash(SB)
	MOVD	32(RSP), R3
	MOVD	R3, ret+16(FP)
	RET

TEXT runtime·memeq(SB),NOSPLIT,$-8-25
	MOVD	a+0(FP), R1
	MOVD	b+8(FP), R2
	MOVD	size+16(FP), R3
	ADD	R1, R3, R6
	MOVD	$1, R0
	MOVB	R0, ret+24(FP)
loop:
	CMP	R1, R6
	BEQ	done
	MOVBU.P	1(R1), R4
	MOVBU.P	1(R2), R5
	CMP	R4, R5
	BEQ	loop

	MOVB	$0, ret+24(FP)
done:
	RET

// memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB),NOSPLIT,$40-17
	MOVD	a+0(FP), R3
	MOVD	b+8(FP), R4
	CMP	R3, R4
	BEQ	eq
	MOVD	8(R26), R5    // compiler stores size at offset 8 in the closure
	MOVD	R3, 8(RSP)
	MOVD	R4, 16(RSP)
	MOVD	R5, 24(RSP)
	BL	runtime·memeq(SB)
	MOVBU	32(RSP), R3
	MOVB	R3, ret+16(FP)
	RET
eq:
	MOVD	$1, R3
	MOVB	R3, ret+16(FP)
	RET

TEXT runtime·cmpstring(SB),NOSPLIT,$-4-40
	MOVD	s1_base+0(FP), R2
	MOVD	s1_len+8(FP), R0
	MOVD	s2_base+16(FP), R3
	MOVD	s2_len+24(FP), R1
	ADD	$40, RSP, R7
	B	runtime·cmpbody<>(SB)

TEXT bytes·Compare(SB),NOSPLIT,$-4-56
	MOVD	s1+0(FP), R2
	MOVD	s1+8(FP), R0
	MOVD	s2+24(FP), R3
	MOVD	s2+32(FP), R1
	ADD	$56, RSP, R7
	B	runtime·cmpbody<>(SB)

// On entry:
// R0 is the length of s1
// R1 is the length of s2
// R2 points to the start of s1
// R3 points to the start of s2
// R7 points to return value (-1/0/1 will be written here)
//
// On Exit:
// R4, R5, and R6 are clobbered
TEXT runtime·cmpbody<>(SB),NOSPLIT,$-4-0
	CMP	R0, R1
	CSEL    LT, R1, R0, R6 // R6 is min(R0, R1)

	ADD	R2, R6	// R2 is current byte in s1, R6 is last byte in s1 to compare
loop:
	CMP	R2, R6
	BEQ	samebytes // all compared bytes were the same; compare lengths
	MOVBU.P	1(R2), R4
	MOVBU.P	1(R3), R5
	CMP	R4, R5
	BEQ	loop
	// bytes differed
	MOVD	$1, R4
	CSNEG	LT, R4, R4, R4
	MOVD	R4, (R7)
	RET
samebytes:
	MOVD	$1, R4
	CMP	R0, R1
	CSNEG	LT, R4, R4, R4
	CSEL	EQ, ZR, R4, R4
	MOVD	R4, (R7)
	RET

// eqstring tests whether two strings are equal.
// The compiler guarantees that strings passed
// to eqstring have equal length.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.
TEXT runtime·eqstring(SB),NOSPLIT,$0-33
	MOVD	s1str+0(FP), R0
	MOVD	s1len+8(FP), R1
	MOVD	s2str+16(FP), R2
	ADD	R0, R1		// end
loop:
	CMP	R0, R1
	BEQ	equal		// reaches the end
	MOVBU.P	1(R0), R4
	MOVBU.P	1(R2), R5
	CMP	R4, R5
	BEQ	loop
notequal:
	MOVB	ZR, ret+32(FP)
	RET
equal:
	MOVD	$1, R0
	MOVB	R0, ret+32(FP)
	RET

//
// functions for other packages
//
TEXT bytes·IndexByte(SB),NOSPLIT,$0-40
	MOVD	b+0(FP), R0
	MOVD	b_len+8(FP), R1
	MOVBU	c+24(FP), R2	// byte to find
	MOVD	R0, R4		// store base for later
	ADD	R0, R1		// end
loop:
	CMP	R0, R1
	BEQ	notfound
	MOVBU.P	1(R0), R3
	CMP	R2, R3
	BNE	loop

	SUB	$1, R0		// R0 will be one beyond the position we want
	SUB	R4, R0		// remove base
	MOVD	R0, ret+32(FP)
	RET

notfound:
	MOVD	$-1, R0
	MOVD	R0, ret+32(FP)
	RET

TEXT strings·IndexByte(SB),NOSPLIT,$0-32
	MOVD	s+0(FP), R0
	MOVD	s_len+8(FP), R1
	MOVBU	c+16(FP), R2	// byte to find
	MOVD	R0, R4		// store base for later
	ADD	R0, R1		// end
loop:
	CMP	R0, R1
	BEQ	notfound
	MOVBU.P	1(R0), R3
	CMP	R2, R3
	BNE	loop

	SUB	$1, R0		// R0 will be one beyond the position we want
	SUB	R4, R0		// remove base
	MOVD	R0, ret+24(FP)
	RET

notfound:
	MOVD	$-1, R0
	MOVD	R0, ret+24(FP)
	RET

// TODO: share code with memeq?
TEXT bytes·Equal(SB),NOSPLIT,$0-49
	MOVD	a_len+8(FP), R1
	MOVD	b_len+32(FP), R3
	CMP	R1, R3		// unequal lengths are not equal
	BNE	notequal
	MOVD	a+0(FP), R0
	MOVD	b+24(FP), R2
	ADD	R0, R1		// end
loop:
	CMP	R0, R1
	BEQ	equal		// reaches the end
	MOVBU.P	1(R0), R4
	MOVBU.P	1(R2), R5
	CMP	R4, R5
	BEQ	loop
notequal:
	MOVB	ZR, ret+48(FP)
	RET
equal:
	MOVD	$1, R0
	MOVB	R0, ret+48(FP)
	RET

TEXT runtime∕internal∕base·Fastrand1(SB),NOSPLIT,$-8-4
	MOVD	G_M(g), R1
	MOVWU	M_fastrand(R1), R0
	ADD	R0, R0
	CMPW	$0, R0
	BGE	notneg
	EOR	$0x88888eef, R0
notneg:
	MOVW	R0, M_fastrand(R1)
	MOVW	R0, ret+0(FP)
	RET

TEXT runtime·return0(SB), NOSPLIT, $0
	MOVW	$0, R0
	RET

// The top-most function running on a goroutine
// returns to Goexit+PCQuantum.
TEXT runtime∕internal∕base·Goexit(SB),NOSPLIT,$-8-0
	MOVD	R0, R0	// NOP
	BL	runtime·goexit1(SB)	// does not return

// TODO(aram): use PRFM here.
TEXT runtime·prefetcht0(SB),NOSPLIT,$0-8
	RET

TEXT runtime·prefetcht1(SB),NOSPLIT,$0-8
	RET

TEXT runtime·prefetcht2(SB),NOSPLIT,$0-8
	RET

TEXT runtime∕internal∕base·Prefetchnta(SB),NOSPLIT,$0-8
	RET

