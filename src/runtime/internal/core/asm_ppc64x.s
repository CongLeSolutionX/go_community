// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

#include "go_asm.h"
#include "go_tls.h"
#include "Funcdata.h"
#include "textflag.h"

TEXT runtime∕internal∕schedinit·rt0_go(SB),NOSPLIT,$0
	// initialize essential registers
	BL	runtime·reginit(SB)

	SUB	$24, R1
	MOVW	R3, 8(R1) // Argc
	MOVD	R4, 16(R1) // Argv

	// create istack out of the given (operating system) stack.
	// _cgo_init may update stackguard.
	MOVD	$runtime∕internal∕core·g0(SB), g
	MOVD	$(-64*1024), R31
	ADD	R31, R1, R3
	MOVD	R3, G_Stackguard0(g)
	MOVD	R3, G_Stackguard1(g)
	MOVD	R3, (G_Stack+Stack_Lo)(g)
	MOVD	R1, (G_Stack+Stack_Hi)(g)

	// TODO: if there is a _cgo_init, call it.
	// TODO: add TLS

	// set the per-goroutine and per-mach "registers"
	MOVD	$runtime∕internal∕core·M0(SB), R3

	// Save m->g0 = g0
	MOVD	g, M_G0(R3)
	// Save M0 to g0->m
	MOVD	R3, G_M(g)

	BL	runtime∕internal∕check·check(SB)

	// args are already prepared
	BL	runtime∕internal∕vdso·args(SB)
	BL	runtime·osinit(SB)
	BL	runtime∕internal∕schedinit·schedinit(SB)

	// create a new goroutine to start program
	MOVD	$runtime·main·f(SB), R3		// entry
	MOVDU	R3, -8(R1)
	MOVDU	R0, -8(R1)
	MOVDU	R0, -8(R1)
	BL	runtime·newproc(SB)
	ADD	$24, R1

	// start this M
	BL	runtime∕internal∕sched·Mstart(SB)

	MOVD	R0, 1(R0)
	RETURN

DATA	runtime·main·f+0(SB)/8,$runtime·main(SB)
GLOBL	runtime·main·f(SB),RODATA,$8

TEXT runtime·breakpoint(SB),NOSPLIT,$-8-0
	MOVD	R0, 2(R0) // TODO: TD
	RETURN

TEXT runtime∕internal∕core·Asminit(SB),NOSPLIT,$-8-0
	RETURN

TEXT runtime·reginit(SB),NOSPLIT,$-8-0
	// set R0 to zero, it's expected by the toolchain
	XOR R0, R0
	// initialize essential FP registers
	FMOVD	$4503601774854144.0, F27
	FMOVD	$0.5, F29
	FSUB	F29, F29, F28
	FADD	F29, F29, F30
	FADD	F30, F30, F31
	RETURN

/*
 *  go-routine
 */

// void gosave(Gobuf*)
// Save state in Gobuf; setjmp
TEXT runtime∕internal∕sched·gosave(SB), NOSPLIT, $-8-8
	MOVD	Buf+0(FP), R3
	MOVD	R1, Gobuf_Sp(R3)
	MOVD	LR, R31
	MOVD	R31, Gobuf_Pc(R3)
	MOVD	g, Gobuf_G(R3)
	MOVD	R0, Gobuf_Lr(R3)
	MOVD	R0, Gobuf_Ret(R3)
	MOVD	R0, Gobuf_Ctxt(R3)
	RETURN

// void Gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT runtime∕internal∕sched·Gogo(SB), NOSPLIT, $-8-8
	MOVD	Buf+0(FP), R5
	MOVD	Gobuf_G(R5), g	// make sure g is not nil
	MOVD	0(g), R4
	MOVD	Gobuf_Sp(R5), R1
	MOVD	Gobuf_Lr(R5), R31
	MOVD	R31, LR
	MOVD	Gobuf_Ret(R5), R3
	MOVD	Gobuf_Ctxt(R5), R11
	MOVD	R0, Gobuf_Sp(R5)
	MOVD	R0, Gobuf_Ret(R5)
	MOVD	R0, Gobuf_Lr(R5)
	MOVD	R0, Gobuf_Ctxt(R5)
	CMP	R0, R0 // set condition codes for == test, needed by stack split
	MOVD	Gobuf_Pc(R5), R31
	MOVD	R31, CTR
	BR	(CTR)

// void Mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return.  It should Gogo(&g->Sched)
// to keep running g.
TEXT runtime∕internal∕sched·Mcall(SB), NOSPLIT, $-8-8
	// Save caller state in g->Sched
	MOVD	R1, (G_Sched+Gobuf_Sp)(g)
	MOVD	LR, R31
	MOVD	R31, (G_Sched+Gobuf_Pc)(g)
	MOVD	R0, (G_Sched+Gobuf_Lr)(g)
	MOVD	g, (G_Sched+Gobuf_G)(g)

	// Switch to m->g0 & its stack, call fn.
	MOVD	g, R3
	MOVD	G_M(g), R8
	MOVD	M_G0(R8), g
	CMP	g, R3
	BNE	2(PC)
	BR	runtime·badmcall(SB)
	MOVD	fn+0(FP), R11			// context
	MOVD	0(R11), R4			// code pointer
	MOVD	R4, CTR
	MOVD	(G_Sched+Gobuf_Sp)(g), R1	// sp = m->g0->Sched.sp
	MOVDU	R3, -8(R1)
	MOVDU	R0, -8(R1)
	BL	(CTR)
	BR	runtime·badmcall2(SB)

// systemstack_switch is a dummy routine that Systemstack leaves at the bottom
// of the G stack.  We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime∕internal∕schedinit·systemstack_switch(SB), NOSPLIT, $0-0
	UNDEF
	BL	(LR)	// make sure this function is not leaf
	RETURN

// func Systemstack(fn func())
TEXT runtime∕internal∕lock·Systemstack(SB), NOSPLIT, $0-8
	MOVD	fn+0(FP), R3	// R3 = fn
	MOVD	R3, R11		// context
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
	MOVD	R3, CTR
	BL	(CTR)

switch:
	// Save our state in g->Sched.  Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVD	$runtime∕internal∕schedinit·systemstack_switch(SB), R6
	ADD	$8, R6	// get past prologue
	MOVD	R6, (G_Sched+Gobuf_Pc)(g)
	MOVD	R1, (G_Sched+Gobuf_Sp)(g)
	MOVD	R0, (G_Sched+Gobuf_Lr)(g)
	MOVD	g, (G_Sched+Gobuf_G)(g)

	// switch to g0
	MOVD	R5, g
	MOVD	(G_Sched+Gobuf_Sp)(g), R3
	// make it look like Mstart called Systemstack on g0, to stop Traceback
	SUB	$8, R3
	MOVD	$runtime∕internal∕sched·Mstart(SB), R4
	MOVD	R4, 0(R3)
	MOVD	R3, R1

	// call target function
	MOVD	0(R11), R3	// code pointer
	MOVD	R3, CTR
	BL	(CTR)

	// switch back to g
	MOVD	G_M(g), R3
	MOVD	M_Curg(R3), g
	MOVD	(G_Sched+Gobuf_Sp)(g), R1
	MOVD	R0, (G_Sched+Gobuf_Sp)(g)
	RETURN

noswitch:
	// already on m stack, just call directly
	MOVD	0(R11), R3	// code pointer
	MOVD	R3, CTR
	BL	(CTR)
	RETURN

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
// Caller has already loaded:
// R3: framesize, R4: argsize, R5: LR
//
// The Traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime∕internal∕schedinit·morestack(SB),NOSPLIT,$-8-0
	// Cannot grow scheduler stack (m->g0).
	MOVD	G_M(g), R7
	MOVD	M_G0(R7), R8
	CMP	g, R8
	BNE	2(PC)
	BL	runtime·abort(SB)

	// Cannot grow signal stack (m->gsignal).
	MOVD	M_Gsignal(R7), R8
	CMP	g, R8
	BNE	2(PC)
	BL	runtime·abort(SB)

	// Called from f.
	// Set g->Sched to context in f.
	MOVD	R11, (G_Sched+Gobuf_Ctxt)(g)
	MOVD	R1, (G_Sched+Gobuf_Sp)(g)
	MOVD	LR, R8
	MOVD	R8, (G_Sched+Gobuf_Pc)(g)
	MOVD	R5, (G_Sched+Gobuf_Lr)(g)

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVD	R5, (M_Morebuf+Gobuf_Pc)(R7)	// f's caller's PC
	MOVD	R1, (M_Morebuf+Gobuf_Sp)(R7)	// f's caller's SP
	MOVD	g, (M_Morebuf+Gobuf_G)(R7)

	// Call newstack on m->g0's stack.
	MOVD	M_G0(R7), g
	MOVD	(G_Sched+Gobuf_Sp)(g), R1
	BL	runtime·newstack(SB)

	// Not reached, but make sure the return PC from the call to newstack
	// is still in this function, and not the beginning of the next.
	UNDEF

TEXT runtime·morestack_noctxt(SB),NOSPLIT,$-8-0
	MOVD	R0, R11
	BR	runtime∕internal∕schedinit·morestack(SB)

// Reflectcall: call a function with the given argument list
// func call(f *FuncVal, arg *byte, argsize, retoffset uint32).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)		\
	MOVD	$MAXSIZE, R31;		\
	CMP	R3, R31;		\
	BGT	4(PC);			\
	MOVD	$NAME(SB), R31;	\
	MOVD	R31, CTR;		\
	BR	(CTR)
// Note: can't just "BR NAME(SB)" - const_bad inlining results.

TEXT reflect·call(SB), NOSPLIT, $0-0
	BR	runtime∕internal∕finalize·Reflectcall(SB)

TEXT runtime∕internal∕finalize·Reflectcall(SB), NOSPLIT, $-8-24
	MOVWZ n+16(FP), R3
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
	MOVD	$runtime·badreflectcall(SB), R31
	MOVD	R31, CTR
	BR	(CTR)

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-24;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVD	arg+8(FP), R3;			\
	MOVWZ	n+16(FP), R4;			\
	MOVD	R1, R5;				\
	ADD	$(8-1), R5;			\
	SUB	$1, R3;				\
	ADD	R5, R4;				\
	CMP	R5, R4;				\
	BEQ	4(PC);				\
	MOVBZU	1(R3), R6;			\
	MOVBZU	R6, 1(R5);			\
	BR	-4(PC);				\
	/* call function */			\
	MOVD	f+0(FP), R11;			\
	MOVD	(R11), R31;			\
	MOVD	R31, CTR;			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	BL	(CTR);				\
	/* copy return values back */		\
	MOVD	arg+8(FP), R3;			\
	MOVWZ	n+16(FP), R4;			\
	MOVWZ	retoffset+20(FP), R6;		\
	MOVD	R1, R5;				\
	ADD	R6, R5; 			\
	ADD	R6, R3;				\
	SUB	R6, R4;				\
	ADD	$(8-1), R5;			\
	SUB	$1, R3;				\
	ADD	R5, R4;				\
	CMP	R5, R4;				\
	BEQ	4(PC);				\
	MOVBZU	1(R5), R6;			\
	MOVBZU	R6, 1(R3);			\
	BR	-4(PC);				\
	RETURN

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

// bool Cas(uint32 *Ptr, uint32 old, uint32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕sched·Cas(SB), NOSPLIT, $0-17
	MOVD	Ptr+0(FP), R3
	MOVWZ	old+8(FP), R4
	MOVWZ	new+12(FP), R5
cas_again:
	SYNC
	LWAR	(R3), R6
	CMPW	R6, R4
	BNE	cas_fail
	STWCCC	R5, (R3)
	BNE	cas_again
	MOVD	$1, R3
	SYNC
	ISYNC
	MOVB	R3, ret+16(FP)
	RETURN
cas_fail:
	MOVD	$0, R3
	BR	-5(PC)

// bool	runtime∕internal∕sched·Cas64(uint64 *Ptr, uint64 old, uint64 new)
// Atomically:
//	if(*val == *old){
//		*val = new;
//		return 1;
//	} else {
//		return 0;
//	}
TEXT runtime∕internal∕sched·Cas64(SB), NOSPLIT, $0-25
	MOVD	Ptr+0(FP), R3
	MOVD	old+8(FP), R4
	MOVD	new+16(FP), R5
cas64_again:
	SYNC
	LDAR	(R3), R6
	CMP	R6, R4
	BNE	cas64_fail
	STDCCC	R5, (R3)
	BNE	cas64_again
	MOVD	$1, R3
	SYNC
	ISYNC
	MOVB	R3, ret+24(FP)
	RETURN
cas64_fail:
	MOVD	$0, R3
	BR	-5(PC)

TEXT runtime∕internal∕core·Casuintptr(SB), NOSPLIT, $0-25
	BR	runtime∕internal∕sched·Cas64(SB)

TEXT runtime∕internal∕core·Atomicloaduintptr(SB), NOSPLIT, $-8-16
	BR	runtime∕internal∕sched·Atomicload64(SB)

TEXT runtime∕internal∕channels·atomicloaduint(SB), NOSPLIT, $-8-16
	BR	runtime∕internal∕sched·Atomicload64(SB)

TEXT runtime∕internal∕core·atomicstoreuintptr(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕sched·Atomicstore64(SB)

// bool casp(void **val, void *old, void *new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕check·casp1(SB), NOSPLIT, $0-25
	BR runtime∕internal∕sched·Cas64(SB)

// uint32 Xadd(uint32 volatile *Ptr, int32 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT runtime∕internal∕lock·Xadd(SB), NOSPLIT, $0-20
	MOVD	Ptr+0(FP), R4
	MOVW	delta+8(FP), R5
	SYNC
	LWAR	(R4), R3
	ADD	R5, R3
	STWCCC	R3, (R4)
	BNE	-4(PC)
	SYNC
	ISYNC
	MOVW	R3, ret+16(FP)
	RETURN

TEXT runtime∕internal∕lock·Xadd64(SB), NOSPLIT, $0-24
	MOVD	Ptr+0(FP), R4
	MOVD	delta+8(FP), R5
	SYNC
	LDAR	(R4), R3
	ADD	R5, R3
	STDCCC	R3, (R4)
	BNE	-4(PC)
	SYNC
	ISYNC
	MOVD	R3, ret+16(FP)
	RETURN

TEXT runtime·xchg(SB), NOSPLIT, $0-20
	MOVD	Ptr+0(FP), R4
	MOVW	new+8(FP), R5
	SYNC
	LWAR	(R4), R3
	STWCCC	R5, (R4)
	BNE	-3(PC)
	SYNC
	ISYNC
	MOVW	R3, ret+16(FP)
	RETURN

TEXT runtime∕internal∕sched·Xchg64(SB), NOSPLIT, $0-24
	MOVD	Ptr+0(FP), R4
	MOVD	new+8(FP), R5
	SYNC
	LDAR	(R4), R3
	STDCCC	R5, (R4)
	BNE	-3(PC)
	SYNC
	ISYNC
	MOVD	R3, ret+16(FP)
	RETURN

TEXT runtime·xchgp1(SB), NOSPLIT, $0-24
	BR	runtime∕internal∕sched·Xchg64(SB)

TEXT runtime∕internal∕netpoll·xchguintptr(SB), NOSPLIT, $0-24
	BR	runtime∕internal∕sched·Xchg64(SB)

TEXT runtime∕internal∕lock·Procyield(SB),NOSPLIT,$0-0
	RETURN

TEXT runtime∕internal∕sched·atomicstorep1(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕sched·Atomicstore64(SB)

TEXT runtime∕internal∕lock·Atomicstore(SB), NOSPLIT, $0-12
	MOVD	Ptr+0(FP), R3
	MOVW	val+8(FP), R4
	SYNC
	MOVW	R4, 0(R3)
	RETURN

TEXT runtime∕internal∕sched·Atomicstore64(SB), NOSPLIT, $0-16
	MOVD	Ptr+0(FP), R3
	MOVD	val+8(FP), R4
	SYNC
	MOVD	R4, 0(R3)
	RETURN

// void	runtime∕internal∕sched·Atomicor8(byte volatile*, byte);
TEXT runtime∕internal∕sched·Atomicor8(SB), NOSPLIT, $0-9
	MOVD	Ptr+0(FP), R3
	MOVBZ	val+8(FP), R4
	// Align Ptr down to 4 Bytes so we can use 32-bit load/store.
	// R5 = (R3 << 0) & ~3
	RLDCR	$0, R3, $~3, R5
	// Compute val const_Shift.
#ifdef GOARCH_ppc64
	// Big endian.  Ptr = Ptr ^ 3
	XOR	$3, R3
#endif
	// R6 = ((Ptr & 3) * 8) = (Ptr << 3) & (3*8)
	RLDC	$3, R3, $(3*8), R6
	// Shift val for aligned Ptr.  R4 = val << R6
	SLD	R6, R4, R4

atomicor8_again:
	SYNC
	LWAR	(R5), R6
	OR	R4, R6
	STWCCC	R6, (R5)
	BNE	atomicor8_again
	SYNC
	ISYNC
	RETURN

// void Jmpdefer(fv, sp);
// called from deferreturn.
// 1. grab stored LR for caller
// 2. sub 4 Bytes to get back to BL deferreturn
// 3. BR to fn
TEXT runtime∕internal∕schedinit·Jmpdefer(SB), NOSPLIT, $-8-16
	MOVD	0(R1), R31
	SUB	$4, R31
	MOVD	R31, LR

	MOVD	fv+0(FP), R11
	MOVD	argp+8(FP), R1
	SUB	$8, R1
	MOVD	0(R11), R3
	MOVD	R3, CTR
	BR	(CTR)

// Save state of caller into g->Sched. Smashes R31.
TEXT gosave<>(SB),NOSPLIT,$-8
	MOVD	LR, R31
	MOVD	R31, (G_Sched+Gobuf_Pc)(g)
	MOVD	R1, (G_Sched+Gobuf_Sp)(g)
	MOVD	R0, (G_Sched+Gobuf_Lr)(g)
	MOVD	R0, (G_Sched+Gobuf_Ret)(g)
	MOVD	R0, (G_Sched+Gobuf_Ctxt)(g)
	RETURN

// Asmcgocall(void(*fn)(void*), void *arg)
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.c for more details.
TEXT runtime∕internal∕sched·Asmcgocall(SB),NOSPLIT,$0-16
	MOVD	R0, 21(R0)

// cgocallback(void (*fn)(void*), void *frame, uintptr framesize)
// Turn the fn into a Go func (by taking its address) and call
// cgocallback_gofunc.
TEXT runtime·cgocallback(SB),NOSPLIT,$24-24
	MOVD	R0, 22(R0)

// cgocallback_gofunc(FuncVal*, void *frame, uintptr framesize)
// See cgocall.c for more details.
TEXT runtime·cgocallback_gofunc(SB),NOSPLIT,$8-24
	MOVD	R0, 23(R0)

// void Setg(G*); set g. for use by needm.
TEXT runtime∕internal∕core·Setg(SB), NOSPLIT, $0-8
	MOVD	R0, 24(R0)

// void setg_gcc(G*); set g called from gcc.
TEXT setg_gcc<>(SB),NOSPLIT,$0
	MOVD	R0, 25(R0)

TEXT runtime∕internal∕lock·Getcallerpc(SB),NOSPLIT,$-8-16
	MOVD	0(R1), R3
	MOVD	R3, ret+8(FP)
	RETURN

TEXT runtime·gogetcallerpc(SB),NOSPLIT,$-8-16
	MOVD	0(R1), R3
	MOVD	R3,ret+8(FP)
	RETURN

TEXT runtime∕internal∕channels·setcallerpc(SB),NOSPLIT,$-8-16
	MOVD	pc+8(FP), R3
	MOVD	R3, 0(R1)		// set calling pc
	RETURN

TEXT runtime∕internal∕lock·Getcallersp(SB),NOSPLIT,$0-16
	MOVD	argp+0(FP), R3
	SUB	$8, R3
	MOVD	R3, ret+8(FP)
	RETURN

// func gogetcallersp(p unsafe.Pointer) uintptr
TEXT runtime·gogetcallersp(SB),NOSPLIT,$0-16
	MOVD	sp+0(FP), R3
	SUB	$8, R3
	MOVD	R3,ret+8(FP)
	RETURN

TEXT runtime·abort(SB),NOSPLIT,$-8-0
	MOVW	(R0), R0
	UNDEF

#define	TBRL	268
#define	TBRU	269		/* Time base Upper/Lower */

// int64 runtime∕internal∕sched·Cputicks(void)
TEXT runtime∕internal∕sched·Cputicks(SB),NOSPLIT,$0-8
	MOVW	SPR(TBRU), R4
	MOVW	SPR(TBRL), R3
	MOVW	SPR(TBRU), R5
	CMPW	R4, R5
	BNE	-4(PC)
	SLD	$32, R5
	OR	R5, R3
	MOVD	R3, ret+0(FP)
	RETURN

// AES hashing not implemented for ppc64
TEXT runtime∕internal∕hash·aeshash(SB),NOSPLIT,$-8-0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshash32(SB),NOSPLIT,$-8-0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshash64(SB),NOSPLIT,$-8-0
	MOVW	(R0), R1
TEXT runtime∕internal∕hash·aeshashstr(SB),NOSPLIT,$-8-0
	MOVW	(R0), R1

TEXT runtime∕internal∕hash·Memeq(SB),NOSPLIT,$-8-25
	MOVD	a+0(FP), R3
	MOVD	b+8(FP), R4
	MOVD	size+16(FP), R5
	SUB	$1, R3
	SUB	$1, R4
	ADD	R3, R5, R8
loop:
	CMP	R3, R8
	BNE	test
	MOVD	$1, R3
	MOVB	R3, ret+24(FP)
	RETURN
test:
	MOVBZU	1(R3), R6
	MOVBZU	1(R4), R7
	CMP	R6, R7
	BEQ	loop

	MOVB	R0, ret+24(FP)
	RETURN

// eqstring tests whether two strings are equal.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.
TEXT runtime·eqstring(SB),NOSPLIT,$0-33
	MOVD	s1len+8(FP), R4
	MOVD	s2len+24(FP), R5
	CMP	R4, R5
	BNE	noteq

	MOVD	s1str+0(FP), R3
	MOVD	s2str+16(FP), R4
	SUB	$1, R3
	SUB	$1, R4
	ADD	R3, R5, R8
loop:
	CMP	R3, R8
	BNE	4(PC)
	MOVD	$1, R3
	MOVB	R3, ret+32(FP)
	RETURN
	MOVBZU	1(R3), R6
	MOVBZU	1(R4), R7
	CMP	R6, R7
	BEQ	loop
noteq:
	MOVB	R0, ret+32(FP)
	RETURN

// TODO: share code with Memeq?
TEXT bytes·Equal(SB),NOSPLIT,$0-49
	MOVD	a_len+8(FP), R3
	MOVD	b_len+32(FP), R4

	CMP	R3, R4		// unequal lengths are not equal
	BNE	noteq

	MOVD	a+0(FP), R5
	MOVD	b+24(FP), R6
	SUB	$1, R5
	SUB	$1, R6
	ADD	R5, R3		// end-1

loop:
	CMP	R5, R3
	BEQ	equal		// reached the end
	MOVBZU	1(R5), R4
	MOVBZU	1(R6), R7
	CMP	R4, R7
	BEQ	loop

noteq:
	MOVBZ	R0, ret+48(FP)
	RETURN

equal:
	MOVD	$1, R3
	MOVBZ	R3, ret+48(FP)
	RETURN

TEXT bytes·IndexByte(SB),NOSPLIT,$0-40
	MOVD	s+0(FP), R3
	MOVD	s_len+8(FP), R4
	MOVBZ	c+24(FP), R5	// byte to find
	MOVD	R3, R6		// store base for later
	SUB	$1, R3
	ADD	R3, R4		// end-1

loop:
	CMP	R3, R4
	BEQ	notfound
	MOVBZU	1(R3), R7
	CMP	R7, R5
	BNE	loop

	SUB	R6, R3		// remove base
	MOVD	R3, ret+32(FP)
	RETURN

notfound:
	MOVD	$-1, R3
	MOVD	R3, ret+32(FP)
	RETURN

TEXT strings·IndexByte(SB),NOSPLIT,$0
	MOVD	p+0(FP), R3
	MOVD	b_len+8(FP), R4
	MOVBZ	c+16(FP), R5	// byte to find
	MOVD	R3, R6		// store base for later
	SUB	$1, R3
	ADD	R3, R4		// end-1

loop:
	CMP	R3, R4
	BEQ	notfound
	MOVBZU	1(R3), R7
	CMP	R7, R5
	BNE	loop

	SUB	R6, R3		// remove base
	MOVD	R3, ret+24(FP)
	RETURN

notfound:
	MOVD	$-1, R3
	MOVD	R3, ret+24(FP)
	RETURN


// A Duff's device for zeroing memory.
// The compiler jumps to computed addresses within
// this routine to zero chunks of memory.  Do not
// change this code without also changing the code
// in ../../cmd/9g/ggen.c:/^clearfat.
// R0: always zero
// R3 (aka REGRT1): Ptr to memory to be zeroed - 8
// On return, R3 points to the last zeroed dword.
TEXT runtime·duffzero(SB), NOSPLIT, $-8-0
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	MOVDU	R0, 8(R3)
	RETURN

TEXT runtime∕internal∕lock·Fastrand1(SB), NOSPLIT, $0-4
	MOVD	G_M(g), R4
	MOVWZ	M_Fastrand(R4), R3
	ADD	R3, R3
	CMPW	R3, $0
	BGE	2(PC)
	XOR	$0x88888eef, R3
	MOVW	R3, M_Fastrand(R4)
	MOVW	R3, ret+0(FP)
	RETURN

TEXT runtime·return0(SB), NOSPLIT, $0
	MOVW	$0, R3
	RETURN

// Called from cgo wrappers, this function returns g->m->curg.stack.hi.
// Must obey the gcc calling convention.
TEXT _cgo_topofstack(SB),NOSPLIT,$0
	MOVD	R0, 26(R0)

// The top-most function running on a goroutine
// returns to Goexit+PCQuantum.
TEXT runtime∕internal∕schedinit·Goexit(SB),NOSPLIT,$-8-0
	MOVD	R0, R0	// NOP
	BL	runtime·goexit1(SB)	// does not return

TEXT runtime∕internal∕core·Getg(SB),NOSPLIT,$-8-8
	MOVD	g, ret+0(FP)
	RETURN

TEXT runtime∕internal∕check·prefetcht0(SB),NOSPLIT,$0-8
	RETURN

TEXT runtime∕internal∕check·prefetcht1(SB),NOSPLIT,$0-8
	RETURN

TEXT runtime∕internal∕check·prefetcht2(SB),NOSPLIT,$0-8
	RETURN

TEXT runtime∕internal∕check·prefetchnta(SB),NOSPLIT,$0-8
	RETURN
