// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// _rt0_386 is common startup code for most 386 systems when using
// internal linking. This is the entry point for the program from the
// kernel for an ordinary -buildmode=exe program. The stack holds the
// number of arguments and the C-style argv.
TEXT _rt0_386(SB),NOSPLIT,$8
	MOVL	8(SP), AX	// argc
	LEAL	12(SP), BX	// argv
	MOVL	AX, 0(SP)
	MOVL	BX, 4(SP)
	JMP	·rt0_go(SB)

// _rt0_386_lib is common startup code for most 386 systems when
// using -buildmode=c-archive or -buildmode=c-shared. The linker will
// arrange to invoke this function as a global constructor (for
// c-archive) or when the shared library is loaded (for c-shared).
// We expect argc and argv to be passed on the stack following the
// usual C ABI.
TEXT _rt0_386_lib(SB),NOSPLIT,$0
	PUSHL	BP
	MOVL	SP, BP
	PUSHL	BX
	PUSHL	SI
	PUSHL	DI

	MOVL	8(BP), AX
	MOVL	AX, _rt0_386_lib_argc<>(SB)
	MOVL	12(BP), AX
	MOVL	AX, _rt0_386_lib_argv<>(SB)

	// Synchronous initialization.
	CALL	·libpreinit(SB)

	SUBL	$8, SP

	// Create a new thread to do the runtime initialization.
	MOVL	_cgo_sys_thread_create(SB), AX
	TESTL	AX, AX
	JZ	nocgo

	// Align stack to call C function.
	// We moved SP to BP above, but BP was clobbered by the libpreinit call.
	MOVL	SP, BP
	ANDL	$~15, SP

	MOVL	$_rt0_386_lib_go(SB), BX
	MOVL	BX, 0(SP)
	MOVL	$0, 4(SP)

	CALL	AX

	MOVL	BP, SP

	JMP	restore

nocgo:
	MOVL	$0x800000, 0(SP)                    // stacksize = 8192KB
	MOVL	$_rt0_386_lib_go(SB), AX
	MOVL	AX, 4(SP)                           // fn
	CALL	·newosproc0(SB)

restore:
	ADDL	$8, SP
	POPL	DI
	POPL	SI
	POPL	BX
	POPL	BP
	RET

// _rt0_386_lib_go initializes the Go runtime.
// This is started in a separate thread by _rt0_386_lib.
TEXT _rt0_386_lib_go(SB),NOSPLIT,$8
	MOVL	_rt0_386_lib_argc<>(SB), AX
	MOVL	AX, 0(SP)
	MOVL	_rt0_386_lib_argv<>(SB), AX
	MOVL	AX, 4(SP)
	JMP	·rt0_go(SB)

DATA _rt0_386_lib_argc<>(SB)/4, $0
GLOBL _rt0_386_lib_argc<>(SB),NOPTR, $4
DATA _rt0_386_lib_argv<>(SB)/4, $0
GLOBL _rt0_386_lib_argv<>(SB),NOPTR, $4

TEXT ·rt0_go(SB),NOSPLIT|NOFRAME|TOPFRAME,$0
	// Copy arguments forward on an even stack.
	// Users of this function jump to it, they don't call it.
	MOVL	0(SP), AX
	MOVL	4(SP), BX
	SUBL	$128, SP		// plenty of scratch
	ANDL	$~15, SP
	MOVL	AX, 120(SP)		// save argc, argv away
	MOVL	BX, 124(SP)

	// set default stack bounds.
	// _cgo_init may update stackguard.
	MOVL	$·g0(SB), BP
	LEAL	(-64*1024+104)(SP), BX
	MOVL	BX, g_stackguard0(BP)
	MOVL	BX, g_stackguard1(BP)
	MOVL	BX, (g_stack+stack_lo)(BP)
	MOVL	SP, (g_stack+stack_hi)(BP)

	// find out information about the processor we're on
	// first see if CPUID instruction is supported.
	PUSHFL
	PUSHFL
	XORL	$(1<<21), 0(SP) // flip ID bit
	POPFL
	PUSHFL
	POPL	AX
	XORL	0(SP), AX
	POPFL	// restore EFLAGS
	TESTL	$(1<<21), AX
	JNE 	has_cpuid

bad_proc: // show that the program requires MMX.
	MOVL	$2, 0(SP)
	MOVL	$bad_proc_msg<>(SB), 4(SP)
	MOVL	$0x3d, 8(SP)
	CALL	·write(SB)
	MOVL	$1, 0(SP)
	CALL	·exit(SB)
	CALL	·abort(SB)

has_cpuid:
	MOVL	$0, AX
	CPUID
	MOVL	AX, SI
	CMPL	AX, $0
	JE	nocpuinfo

	CMPL	BX, $0x756E6547  // "Genu"
	JNE	notintel
	CMPL	DX, $0x49656E69  // "ineI"
	JNE	notintel
	CMPL	CX, $0x6C65746E  // "ntel"
	JNE	notintel
	MOVB	$1, ·isIntel(SB)
notintel:

	// Load EAX=1 cpuid flags
	MOVL	$1, AX
	CPUID
	MOVL	CX, DI // Move to global variable clobbers CX when generating PIC
	MOVL	AX, ·processorVersionInfo(SB)

	// Check for MMX support
	TESTL	$(1<<23), DX // MMX
	JZ	bad_proc

nocpuinfo:
	// if there is an _cgo_init, call it to let it
	// initialize and to set up GS.  if not,
	// we set up GS ourselves.
	MOVL	_cgo_init(SB), AX
	TESTL	AX, AX
	JZ	needtls
#ifdef GOOS_android
	// arg 4: TLS base, stored in slot 0 (Android's TLS_SLOT_SELF).
	// Compensate for tls_g (+8).
	MOVL	-8(TLS), BX
	MOVL	BX, 12(SP)
	MOVL	$·tls_g(SB), 8(SP)	// arg 3: &tls_g
#else
	MOVL	$0, BX
	MOVL	BX, 12(SP)	// arg 4: not used when using platform's TLS
#ifdef GOOS_windows
	MOVL	$·tls_g(SB), 8(SP)	// arg 3: &tls_g
#else
	MOVL	BX, 8(SP)	// arg 3: not used when using platform's TLS
#endif
#endif
	MOVL	$setg_gcc<>(SB), BX
	MOVL	BX, 4(SP)	// arg 2: setg_gcc
	MOVL	BP, 0(SP)	// arg 1: g0
	CALL	AX

	// update stackguard after _cgo_init
	MOVL	$·g0(SB), CX
	MOVL	(g_stack+stack_lo)(CX), AX
	ADDL	$const_stackGuard, AX
	MOVL	AX, g_stackguard0(CX)
	MOVL	AX, g_stackguard1(CX)

#ifndef GOOS_windows
	// skip ·ldt0setup(SB) and tls test after _cgo_init for non-windows
	JMP ok
#endif
needtls:
#ifdef GOOS_openbsd
	// skip ·ldt0setup(SB) and tls test on OpenBSD in all cases
	JMP	ok
#endif
#ifdef GOOS_plan9
	// skip ·ldt0setup(SB) and tls test on Plan 9 in all cases
	JMP	ok
#endif

	// set up %gs
	CALL	ldt0setup<>(SB)

	// store through it, to make sure it works
	get_tls(BX)
	MOVL	$0x123, g(BX)
	MOVL	·m0+m_tls(SB), AX
	CMPL	AX, $0x123
	JEQ	ok
	MOVL	AX, 0	// abort
ok:
	// set up m and g "registers"
	get_tls(BX)
	LEAL	·g0(SB), DX
	MOVL	DX, g(BX)
	LEAL	·m0(SB), AX

	// save m->g0 = g0
	MOVL	DX, m_g0(AX)
	// save g0->m = m0
	MOVL	AX, g_m(DX)

	CALL	·emptyfunc(SB)	// fault if stack check is wrong

	// convention is D is always cleared
	CLD

	CALL	·check(SB)

	// saved argc, argv
	MOVL	120(SP), AX
	MOVL	AX, 0(SP)
	MOVL	124(SP), AX
	MOVL	AX, 4(SP)
	CALL	·args(SB)
	CALL	·osinit(SB)
	CALL	·schedinit(SB)

	// create a new goroutine to start program
	PUSHL	$·mainPC(SB)	// entry
	CALL	·newproc(SB)
	POPL	AX

	// start this M
	CALL	·mstart(SB)

	CALL	·abort(SB)
	RET

DATA	bad_proc_msg<>+0x00(SB)/61, $"This program can only be run on processors with MMX support.\n"
GLOBL	bad_proc_msg<>(SB), RODATA, $61

DATA	·mainPC+0(SB)/4,$·main(SB)
GLOBL	·mainPC(SB),RODATA,$4

TEXT ·breakpoint(SB),NOSPLIT,$0-0
	INT $3
	RET

TEXT ·asminit(SB),NOSPLIT,$0-0
	// Linux and MinGW start the FPU in extended double precision.
	// Other operating systems use double precision.
	// Change to double precision to match them,
	// and to match other hardware that only has double.
	FLDCW	·controlWord64(SB)
	RET

TEXT ·mstart(SB),NOSPLIT|TOPFRAME,$0
	CALL	·mstart0(SB)
	RET // not reached

/*
 *  go-routine
 */

// void gogo(Gobuf*)
// restore state from Gobuf; longjmp
TEXT ·gogo(SB), NOSPLIT, $0-4
	MOVL	buf+0(FP), BX		// gobuf
	MOVL	gobuf_g(BX), DX
	MOVL	0(DX), CX		// make sure g != nil
	JMP	gogo<>(SB)

TEXT gogo<>(SB), NOSPLIT, $0
	get_tls(CX)
	MOVL	DX, g(CX)
	MOVL	gobuf_sp(BX), SP	// restore SP
	MOVL	gobuf_ret(BX), AX
	MOVL	gobuf_ctxt(BX), DX
	MOVL	$0, gobuf_sp(BX)	// clear to help garbage collector
	MOVL	$0, gobuf_ret(BX)
	MOVL	$0, gobuf_ctxt(BX)
	MOVL	gobuf_pc(BX), BX
	JMP	BX

// func mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return. It should gogo(&g->sched)
// to keep running g.
TEXT ·mcall(SB), NOSPLIT, $0-4
	MOVL	fn+0(FP), DI

	get_tls(DX)
	MOVL	g(DX), AX	// save state in g->sched
	MOVL	0(SP), BX	// caller's PC
	MOVL	BX, (g_sched+gobuf_pc)(AX)
	LEAL	fn+0(FP), BX	// caller's SP
	MOVL	BX, (g_sched+gobuf_sp)(AX)

	// switch to m->g0 & its stack, call fn
	MOVL	g(DX), BX
	MOVL	g_m(BX), BX
	MOVL	m_g0(BX), SI
	CMPL	SI, AX	// if g == m->g0 call badmcall
	JNE	3(PC)
	MOVL	$·badmcall(SB), AX
	JMP	AX
	MOVL	SI, g(DX)	// g = m->g0
	MOVL	(g_sched+gobuf_sp)(SI), SP	// sp = m->g0->sched.sp
	PUSHL	AX
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI
	POPL	AX
	MOVL	$·badmcall2(SB), AX
	JMP	AX
	RET

// systemstack_switch is a dummy routine that systemstack leaves at the bottom
// of the G stack. We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT ·systemstack_switch(SB), NOSPLIT, $0-0
	RET

// func systemstack(fn func())
TEXT ·systemstack(SB), NOSPLIT, $0-4
	MOVL	fn+0(FP), DI	// DI = fn
	get_tls(CX)
	MOVL	g(CX), AX	// AX = g
	MOVL	g_m(AX), BX	// BX = m

	CMPL	AX, m_gsignal(BX)
	JEQ	noswitch

	MOVL	m_g0(BX), DX	// DX = g0
	CMPL	AX, DX
	JEQ	noswitch

	CMPL	AX, m_curg(BX)
	JNE	bad

	// switch stacks
	// save our state in g->sched. Pretend to
	// be systemstack_switch if the G stack is scanned.
	CALL	gosave_systemstack_switch<>(SB)

	// switch to g0
	get_tls(CX)
	MOVL	DX, g(CX)
	MOVL	(g_sched+gobuf_sp)(DX), BX
	MOVL	BX, SP

	// call target function
	MOVL	DI, DX
	MOVL	0(DI), DI
	CALL	DI

	// switch back to g
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	g_m(AX), BX
	MOVL	m_curg(BX), AX
	MOVL	AX, g(CX)
	MOVL	(g_sched+gobuf_sp)(AX), SP
	MOVL	$0, (g_sched+gobuf_sp)(AX)
	RET

noswitch:
	// already on system stack; tail call the function
	// Using a tail call here cleans up tracebacks since we won't stop
	// at an intermediate systemstack.
	MOVL	DI, DX
	MOVL	0(DI), DI
	JMP	DI

bad:
	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOVL	$·badsystemstack(SB), AX
	CALL	AX
	INT	$3

/*
 * support for morestack
 */

// Called during function prolog when more stack is needed.
//
// The traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT ·morestack(SB),NOSPLIT,$0-0
	// Cannot grow scheduler stack (m->g0).
	get_tls(CX)
	MOVL	g(CX), BX
	MOVL	g_m(BX), BX
	MOVL	m_g0(BX), SI
	CMPL	g(CX), SI
	JNE	3(PC)
	CALL	·badmorestackg0(SB)
	CALL	·abort(SB)

	// Cannot grow signal stack.
	MOVL	m_gsignal(BX), SI
	CMPL	g(CX), SI
	JNE	3(PC)
	CALL	·badmorestackgsignal(SB)
	CALL	·abort(SB)

	// Called from f.
	// Set m->morebuf to f's caller.
	NOP	SP	// tell vet SP changed - stop checking offsets
	MOVL	4(SP), DI	// f's caller's PC
	MOVL	DI, (m_morebuf+gobuf_pc)(BX)
	LEAL	8(SP), CX	// f's caller's SP
	MOVL	CX, (m_morebuf+gobuf_sp)(BX)
	get_tls(CX)
	MOVL	g(CX), SI
	MOVL	SI, (m_morebuf+gobuf_g)(BX)

	// Set g->sched to context in f.
	MOVL	0(SP), AX	// f's PC
	MOVL	AX, (g_sched+gobuf_pc)(SI)
	LEAL	4(SP), AX	// f's SP
	MOVL	AX, (g_sched+gobuf_sp)(SI)
	MOVL	DX, (g_sched+gobuf_ctxt)(SI)

	// Call newstack on m->g0's stack.
	MOVL	m_g0(BX), BP
	MOVL	BP, g(CX)
	MOVL	(g_sched+gobuf_sp)(BP), AX
	MOVL	-4(AX), BX	// fault if CALL would, before smashing SP
	MOVL	AX, SP
	CALL	·newstack(SB)
	CALL	·abort(SB)	// crash if newstack returns
	RET

TEXT ·morestack_noctxt(SB),NOSPLIT,$0-0
	MOVL	$0, DX
	JMP ·morestack(SB)

// reflectcall: call a function with the given argument list
// func call(stackArgsType *_type, f *FuncVal, stackArgs *byte, stackArgsSize, stackRetOffset, frameSize uint32, regArgs *abi.RegArgs).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)		\
	CMPL	CX, $MAXSIZE;		\
	JA	3(PC);			\
	MOVL	$NAME(SB), AX;		\
	JMP	AX
// Note: can't just "JMP NAME(SB)" - bad inlining results.

TEXT ·reflectcall(SB), NOSPLIT, $0-28
	MOVL	frameSize+20(FP), CX
	DISPATCH(·call16, 16)
	DISPATCH(·call32, 32)
	DISPATCH(·call64, 64)
	DISPATCH(·call128, 128)
	DISPATCH(·call256, 256)
	DISPATCH(·call512, 512)
	DISPATCH(·call1024, 1024)
	DISPATCH(·call2048, 2048)
	DISPATCH(·call4096, 4096)
	DISPATCH(·call8192, 8192)
	DISPATCH(·call16384, 16384)
	DISPATCH(·call32768, 32768)
	DISPATCH(·call65536, 65536)
	DISPATCH(·call131072, 131072)
	DISPATCH(·call262144, 262144)
	DISPATCH(·call524288, 524288)
	DISPATCH(·call1048576, 1048576)
	DISPATCH(·call2097152, 2097152)
	DISPATCH(·call4194304, 4194304)
	DISPATCH(·call8388608, 8388608)
	DISPATCH(·call16777216, 16777216)
	DISPATCH(·call33554432, 33554432)
	DISPATCH(·call67108864, 67108864)
	DISPATCH(·call134217728, 134217728)
	DISPATCH(·call268435456, 268435456)
	DISPATCH(·call536870912, 536870912)
	DISPATCH(·call1073741824, 1073741824)
	MOVL	$·badreflectcall(SB), AX
	JMP	AX

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-28;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOVL	stackArgs+8(FP), SI;		\
	MOVL	stackArgsSize+12(FP), CX;		\
	MOVL	SP, DI;				\
	REP;MOVSB;				\
	/* call function */			\
	MOVL	f+4(FP), DX;			\
	MOVL	(DX), AX; 			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	CALL	AX;				\
	/* copy return values back */		\
	MOVL	stackArgsType+0(FP), DX;		\
	MOVL	stackArgs+8(FP), DI;		\
	MOVL	stackArgsSize+12(FP), CX;		\
	MOVL	stackRetOffset+16(FP), BX;		\
	MOVL	SP, SI;				\
	ADDL	BX, DI;				\
	ADDL	BX, SI;				\
	SUBL	BX, CX;				\
	CALL	callRet<>(SB);			\
	RET

// callRet copies return values back at the end of call*. This is a
// separate function so it can allocate stack space for the arguments
// to reflectcallmove. It does not follow the Go ABI; it expects its
// arguments in registers.
TEXT callRet<>(SB), NOSPLIT, $20-0
	MOVL	DX, 0(SP)
	MOVL	DI, 4(SP)
	MOVL	SI, 8(SP)
	MOVL	CX, 12(SP)
	MOVL	$0, 16(SP)
	CALL	·reflectcallmove(SB)
	RET

CALLFN(·call16, 16)
CALLFN(·call32, 32)
CALLFN(·call64, 64)
CALLFN(·call128, 128)
CALLFN(·call256, 256)
CALLFN(·call512, 512)
CALLFN(·call1024, 1024)
CALLFN(·call2048, 2048)
CALLFN(·call4096, 4096)
CALLFN(·call8192, 8192)
CALLFN(·call16384, 16384)
CALLFN(·call32768, 32768)
CALLFN(·call65536, 65536)
CALLFN(·call131072, 131072)
CALLFN(·call262144, 262144)
CALLFN(·call524288, 524288)
CALLFN(·call1048576, 1048576)
CALLFN(·call2097152, 2097152)
CALLFN(·call4194304, 4194304)
CALLFN(·call8388608, 8388608)
CALLFN(·call16777216, 16777216)
CALLFN(·call33554432, 33554432)
CALLFN(·call67108864, 67108864)
CALLFN(·call134217728, 134217728)
CALLFN(·call268435456, 268435456)
CALLFN(·call536870912, 536870912)
CALLFN(·call1073741824, 1073741824)

TEXT ·procyield(SB),NOSPLIT,$0-0
	MOVL	cycles+0(FP), AX
again:
	PAUSE
	SUBL	$1, AX
	JNZ	again
	RET

TEXT ·publicationBarrier(SB),NOSPLIT,$0-0
	// Stores are already ordered on x86, so this is just a
	// compile barrier.
	RET

// Save state of caller into g->sched,
// but using fake PC from systemstack_switch.
// Must only be called from functions with no locals ($0)
// or else unwinding from systemstack_switch is incorrect.
TEXT gosave_systemstack_switch<>(SB),NOSPLIT,$0
	PUSHL	AX
	PUSHL	BX
	get_tls(BX)
	MOVL	g(BX), BX
	LEAL	arg+0(FP), AX
	MOVL	AX, (g_sched+gobuf_sp)(BX)
	MOVL	$·systemstack_switch(SB), AX
	MOVL	AX, (g_sched+gobuf_pc)(BX)
	MOVL	$0, (g_sched+gobuf_ret)(BX)
	// Assert ctxt is zero. See func save.
	MOVL	(g_sched+gobuf_ctxt)(BX), AX
	TESTL	AX, AX
	JZ	2(PC)
	CALL	·abort(SB)
	POPL	BX
	POPL	AX
	RET

// func asmcgocall_no_g(fn, arg unsafe.Pointer)
// Call fn(arg) aligned appropriately for the gcc ABI.
// Called on a system stack, and there may be no g yet (during needm).
TEXT ·asmcgocall_no_g(SB),NOSPLIT,$0-8
	MOVL	fn+0(FP), AX
	MOVL	arg+4(FP), BX
	MOVL	SP, DX
	SUBL	$32, SP
	ANDL	$~15, SP	// alignment, perhaps unnecessary
	MOVL	DX, 8(SP)	// save old SP
	MOVL	BX, 0(SP)	// first argument in x86-32 ABI
	CALL	AX
	MOVL	8(SP), DX
	MOVL	DX, SP
	RET

// func asmcgocall(fn, arg unsafe.Pointer) int32
// Call fn(arg) on the scheduler stack,
// aligned appropriately for the gcc ABI.
// See cgocall.go for more details.
TEXT ·asmcgocall(SB),NOSPLIT,$0-12
	MOVL	fn+0(FP), AX
	MOVL	arg+4(FP), BX

	MOVL	SP, DX

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already. Or we might already
	// be on the m->gsignal stack.
	get_tls(CX)
	MOVL	g(CX), DI
	CMPL	DI, $0
	JEQ	nosave	// Don't even have a G yet.
	MOVL	g_m(DI), BP
	CMPL	DI, m_gsignal(BP)
	JEQ	noswitch
	MOVL	m_g0(BP), SI
	CMPL	DI, SI
	JEQ	noswitch
	CALL	gosave_systemstack_switch<>(SB)
	get_tls(CX)
	MOVL	SI, g(CX)
	MOVL	(g_sched+gobuf_sp)(SI), SP

noswitch:
	// Now on a scheduling stack (a pthread-created stack).
	SUBL	$32, SP
	ANDL	$~15, SP	// alignment, perhaps unnecessary
	MOVL	DI, 8(SP)	// save g
	MOVL	(g_stack+stack_hi)(DI), DI
	SUBL	DX, DI
	MOVL	DI, 4(SP)	// save depth in stack (can't just save SP, as stack might be copied during a callback)
	MOVL	BX, 0(SP)	// first argument in x86-32 ABI
	CALL	AX

	// Restore registers, g, stack pointer.
	get_tls(CX)
	MOVL	8(SP), DI
	MOVL	(g_stack+stack_hi)(DI), SI
	SUBL	4(SP), SI
	MOVL	DI, g(CX)
	MOVL	SI, SP

	MOVL	AX, ret+8(FP)
	RET
nosave:
	// Now on a scheduling stack (a pthread-created stack).
	SUBL	$32, SP
	ANDL	$~15, SP	// alignment, perhaps unnecessary
	MOVL	DX, 4(SP)	// save original stack pointer
	MOVL	BX, 0(SP)	// first argument in x86-32 ABI
	CALL	AX

	MOVL	4(SP), CX	// restore original stack pointer
	MOVL	CX, SP
	MOVL	AX, ret+8(FP)
	RET

// cgocallback(fn, frame unsafe.Pointer, ctxt uintptr)
// See cgocall.go for more details.
TEXT ·cgocallback(SB),NOSPLIT,$12-12  // Frame size must match commented places below
	NO_LOCAL_POINTERS

	// Skip cgocallbackg, just dropm when fn is nil, and frame is the saved g.
	// It is used to dropm while thread is exiting.
	MOVL	fn+0(FP), AX
	CMPL	AX, $0
	JNE	loadg
	// Restore the g from frame.
	get_tls(CX)
	MOVL	frame+4(FP), BX
	MOVL	BX, g(CX)
	JMP	dropm

loadg:
	// If g is nil, Go did not create the current thread,
	// or if this thread never called into Go on pthread platforms.
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
	MOVL	g_m(BP), BP
	MOVL	BP, savedm-4(SP) // saved copy of oldm
	JMP	havem
needm:
	MOVL	$·needAndBindM(SB), AX
	CALL	AX
	MOVL	$0, savedm-4(SP)
	get_tls(CX)
	MOVL	g(CX), BP
	MOVL	g_m(BP), BP

	// Set m->sched.sp = SP, so that if a panic happens
	// during the function we are about to execute, it will
	// have a valid SP to run on the g0 stack.
	// The next few lines (after the havem label)
	// will save this SP onto the stack and then write
	// the same SP back to m->sched.sp. That seems redundant,
	// but if an unrecovered panic happens, unwindm will
	// restore the g->sched.sp from the stack location
	// and then systemstack will try to use it. If we don't set it here,
	// that restored SP will be uninitialized (typically 0) and
	// will not be usable.
	MOVL	m_g0(BP), SI
	MOVL	SP, (g_sched+gobuf_sp)(SI)

havem:
	// Now there's a valid m, and we're running on its m->g0.
	// Save current m->g0->sched.sp on stack and then set it to SP.
	// Save current sp in m->g0->sched.sp in preparation for
	// switch back to m->curg stack.
	// NOTE: unwindm knows that the saved g->sched.sp is at 0(SP).
	MOVL	m_g0(BP), SI
	MOVL	(g_sched+gobuf_sp)(SI), AX
	MOVL	AX, 0(SP)
	MOVL	SP, (g_sched+gobuf_sp)(SI)

	// Switch to m->curg stack and call runtime.cgocallbackg.
	// Because we are taking over the execution of m->curg
	// but *not* resuming what had been running, we need to
	// save that information (m->curg->sched) so we can restore it.
	// We can restore m->curg->sched.sp easily, because calling
	// runtime.cgocallbackg leaves SP unchanged upon return.
	// To save m->curg->sched.pc, we push it onto the curg stack and
	// open a frame the same size as cgocallback's g0 frame.
	// Once we switch to the curg stack, the pushed PC will appear
	// to be the return PC of cgocallback, so that the traceback
	// will seamlessly trace back into the earlier calls.
	MOVL	m_curg(BP), SI
	MOVL	SI, g(CX)
	MOVL	(g_sched+gobuf_sp)(SI), DI // prepare stack as DI
	MOVL	(g_sched+gobuf_pc)(SI), BP
	MOVL	BP, -4(DI)  // "push" return PC on the g stack
	// Gather our arguments into registers.
	MOVL	fn+0(FP), AX
	MOVL	frame+4(FP), BX
	MOVL	ctxt+8(FP), CX
	LEAL	-(4+12)(DI), SP  // Must match declared frame size
	MOVL	AX, 0(SP)
	MOVL	BX, 4(SP)
	MOVL	CX, 8(SP)
	CALL	·cgocallbackg(SB)

	// Restore g->sched (== m->curg->sched) from saved values.
	get_tls(CX)
	MOVL	g(CX), SI
	MOVL	12(SP), BP  // Must match declared frame size
	MOVL	BP, (g_sched+gobuf_pc)(SI)
	LEAL	(12+4)(SP), DI  // Must match declared frame size
	MOVL	DI, (g_sched+gobuf_sp)(SI)

	// Switch back to m->g0's stack and restore m->g0->sched.sp.
	// (Unlike m->curg, the g0 goroutine never uses sched.pc,
	// so we do not have to restore it.)
	MOVL	g(CX), BP
	MOVL	g_m(BP), BP
	MOVL	m_g0(BP), SI
	MOVL	SI, g(CX)
	MOVL	(g_sched+gobuf_sp)(SI), SP
	MOVL	0(SP), AX
	MOVL	AX, (g_sched+gobuf_sp)(SI)

	// If the m on entry was nil, we called needm above to borrow an m,
	// 1. for the duration of the call on non-pthread platforms,
	// 2. or the duration of the C thread alive on pthread platforms.
	// If the m on entry wasn't nil,
	// 1. the thread might be a Go thread,
	// 2. or it wasn't the first call from a C thread on pthread platforms,
	//    since then we skip dropm to reuse the m in the first call.
	MOVL	savedm-4(SP), DX
	CMPL	DX, $0
	JNE	droppedm

	// Skip dropm to reuse it in the next call, when a pthread key has been created.
	MOVL	_cgo_pthread_key_created(SB), DX
	// It means cgo is disabled when _cgo_pthread_key_created is a nil pointer, need dropm.
	CMPL	DX, $0
	JEQ	dropm
	CMPL	(DX), $0
	JNE	droppedm

dropm:
	MOVL	$·dropm(SB), AX
	CALL	AX
droppedm:

	// Done!
	RET

// void setg(G*); set g. for use by needm.
TEXT ·setg(SB), NOSPLIT, $0-4
	MOVL	gg+0(FP), BX
#ifdef GOOS_windows
	MOVL	·tls_g(SB), CX
	CMPL	BX, $0
	JNE	settls
	MOVL	$0, 0(CX)(FS)
	RET
settls:
	MOVL	g_m(BX), AX
	LEAL	m_tls(AX), AX
	MOVL	AX, 0(CX)(FS)
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

TEXT ·abort(SB),NOSPLIT,$0-0
	INT	$3
loop:
	JMP	loop

// check that SP is in range [g->stack.lo, g->stack.hi)
TEXT ·stackcheck(SB), NOSPLIT, $0-0
	get_tls(CX)
	MOVL	g(CX), AX
	CMPL	(g_stack+stack_hi)(AX), SP
	JHI	2(PC)
	CALL	·abort(SB)
	CMPL	SP, (g_stack+stack_lo)(AX)
	JHI	2(PC)
	CALL	·abort(SB)
	RET

// func cputicks() int64
TEXT ·cputicks(SB),NOSPLIT,$0-8
	// LFENCE/MFENCE instruction support is dependent on SSE2.
	// When no SSE2 support is present do not enforce any serialization
	// since using CPUID to serialize the instruction stream is
	// very costly.
#ifdef GO386_softfloat
	JMP	rdtsc  // no fence instructions available
#endif
	CMPB	internal∕cpu·X86+const_offsetX86HasRDTSCP(SB), $1
	JNE	fences
	// Instruction stream serializing RDTSCP is supported.
	// RDTSCP is supported by Intel Nehalem (2008) and
	// AMD K8 Rev. F (2006) and newer.
	RDTSCP
done:
	MOVL	AX, ret_lo+0(FP)
	MOVL	DX, ret_hi+4(FP)
	RET
fences:
	// MFENCE is instruction stream serializing and flushes the
	// store buffers on AMD. The serialization semantics of LFENCE on AMD
	// are dependent on MSR C001_1029 and CPU generation.
	// LFENCE on Intel does wait for all previous instructions to have executed.
	// Intel recommends MFENCE;LFENCE in its manuals before RDTSC to have all
	// previous instructions executed and all previous loads and stores to globally visible.
	// Using MFENCE;LFENCE here aligns the serializing properties without
	// runtime detection of CPU manufacturer.
	MFENCE
	LFENCE
rdtsc:
	RDTSC
	JMP done

TEXT ldt0setup<>(SB),NOSPLIT,$16-0
#ifdef GOOS_windows
	CALL	·wintls(SB)
#endif
	// set up ldt 7 to point at m0.tls
	// ldt 1 would be fine on Linux, but on OS X, 7 is as low as we can go.
	// the entry number is just a hint.  setldt will set up GS with what it used.
	MOVL	$7, 0(SP)
	LEAL	·m0+m_tls(SB), AX
	MOVL	AX, 4(SP)
	MOVL	$32, 8(SP)	// sizeof(tls array)
	CALL	·setldt(SB)
	RET

TEXT ·emptyfunc(SB),0,$0-0
	RET

// hash function using AES hardware instructions
TEXT ·memhash(SB),NOSPLIT,$0-16
	CMPB	·useAeshash(SB), $0
	JEQ	noaes
	MOVL	p+0(FP), AX	// ptr to data
	MOVL	s+8(FP), BX	// size
	LEAL	ret+12(FP), DX
	JMP	aeshashbody<>(SB)
noaes:
	JMP	·memhashFallback(SB)

TEXT ·strhash(SB),NOSPLIT,$0-12
	CMPB	·useAeshash(SB), $0
	JEQ	noaes
	MOVL	p+0(FP), AX	// ptr to string object
	MOVL	4(AX), BX	// length of string
	MOVL	(AX), AX	// string data
	LEAL	ret+8(FP), DX
	JMP	aeshashbody<>(SB)
noaes:
	JMP	·strhashFallback(SB)

// AX: data
// BX: length
// DX: address to put return value
TEXT aeshashbody<>(SB),NOSPLIT,$0-0
	MOVL	h+4(FP), X0	            // 32 bits of per-table hash seed
	PINSRW	$4, BX, X0	            // 16 bits of length
	PSHUFHW	$0, X0, X0	            // replace size with its low 2 bytes repeated 4 times
	MOVO	X0, X1                      // save unscrambled seed
	PXOR	·aeskeysched(SB), X0 // xor in per-process seed
	AESENC	X0, X0                      // scramble seed

	CMPL	BX, $16
	JB	aes0to15
	JE	aes16
	CMPL	BX, $32
	JBE	aes17to32
	CMPL	BX, $64
	JBE	aes33to64
	JMP	aes65plus

aes0to15:
	TESTL	BX, BX
	JE	aes0

	ADDL	$16, AX
	TESTW	$0xff0, AX
	JE	endofpage

	// 16 bytes loaded at this address won't cross
	// a page boundary, so we can load it directly.
	MOVOU	-16(AX), X1
	ADDL	BX, BX
	PAND	masks<>(SB)(BX*8), X1

final1:
	PXOR	X0, X1	// xor data with seed
	AESENC	X1, X1  // scramble combo 3 times
	AESENC	X1, X1
	AESENC	X1, X1
	MOVL	X1, (DX)
	RET

endofpage:
	// address ends in 1111xxxx. Might be up against
	// a page boundary, so load ending at last byte.
	// Then shift bytes down using pshufb.
	MOVOU	-32(AX)(BX*1), X1
	ADDL	BX, BX
	PSHUFB	shifts<>(SB)(BX*8), X1
	JMP	final1

aes0:
	// Return scrambled input seed
	AESENC	X0, X0
	MOVL	X0, (DX)
	RET

aes16:
	MOVOU	(AX), X1
	JMP	final1

aes17to32:
	// make second starting seed
	PXOR	·aeskeysched+16(SB), X1
	AESENC	X1, X1

	// load data to be hashed
	MOVOU	(AX), X2
	MOVOU	-16(AX)(BX*1), X3

	// xor with seed
	PXOR	X0, X2
	PXOR	X1, X3

	// scramble 3 times
	AESENC	X2, X2
	AESENC	X3, X3
	AESENC	X2, X2
	AESENC	X3, X3
	AESENC	X2, X2
	AESENC	X3, X3

	// combine results
	PXOR	X3, X2
	MOVL	X2, (DX)
	RET

aes33to64:
	// make 3 more starting seeds
	MOVO	X1, X2
	MOVO	X1, X3
	PXOR	·aeskeysched+16(SB), X1
	PXOR	·aeskeysched+32(SB), X2
	PXOR	·aeskeysched+48(SB), X3
	AESENC	X1, X1
	AESENC	X2, X2
	AESENC	X3, X3

	MOVOU	(AX), X4
	MOVOU	16(AX), X5
	MOVOU	-32(AX)(BX*1), X6
	MOVOU	-16(AX)(BX*1), X7

	PXOR	X0, X4
	PXOR	X1, X5
	PXOR	X2, X6
	PXOR	X3, X7

	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	PXOR	X6, X4
	PXOR	X7, X5
	PXOR	X5, X4
	MOVL	X4, (DX)
	RET

aes65plus:
	// make 3 more starting seeds
	MOVO	X1, X2
	MOVO	X1, X3
	PXOR	·aeskeysched+16(SB), X1
	PXOR	·aeskeysched+32(SB), X2
	PXOR	·aeskeysched+48(SB), X3
	AESENC	X1, X1
	AESENC	X2, X2
	AESENC	X3, X3

	// start with last (possibly overlapping) block
	MOVOU	-64(AX)(BX*1), X4
	MOVOU	-48(AX)(BX*1), X5
	MOVOU	-32(AX)(BX*1), X6
	MOVOU	-16(AX)(BX*1), X7

	// scramble state once
	AESENC	X0, X4
	AESENC	X1, X5
	AESENC	X2, X6
	AESENC	X3, X7

	// compute number of remaining 64-byte blocks
	DECL	BX
	SHRL	$6, BX

aesloop:
	// scramble state, xor in a block
	MOVOU	(AX), X0
	MOVOU	16(AX), X1
	MOVOU	32(AX), X2
	MOVOU	48(AX), X3
	AESENC	X0, X4
	AESENC	X1, X5
	AESENC	X2, X6
	AESENC	X3, X7

	// scramble state
	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	ADDL	$64, AX
	DECL	BX
	JNE	aesloop

	// 3 more scrambles to finish
	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	AESENC	X4, X4
	AESENC	X5, X5
	AESENC	X6, X6
	AESENC	X7, X7

	PXOR	X6, X4
	PXOR	X7, X5
	PXOR	X5, X4
	MOVL	X4, (DX)
	RET

TEXT ·memhash32(SB),NOSPLIT,$0-12
	CMPB	·useAeshash(SB), $0
	JEQ	noaes
	MOVL	p+0(FP), AX	// ptr to data
	MOVL	h+4(FP), X0	// seed
	PINSRD	$1, (AX), X0	// data
	AESENC	·aeskeysched+0(SB), X0
	AESENC	·aeskeysched+16(SB), X0
	AESENC	·aeskeysched+32(SB), X0
	MOVL	X0, ret+8(FP)
	RET
noaes:
	JMP	·memhash32Fallback(SB)

TEXT ·memhash64(SB),NOSPLIT,$0-12
	CMPB	·useAeshash(SB), $0
	JEQ	noaes
	MOVL	p+0(FP), AX	// ptr to data
	MOVQ	(AX), X0	// data
	PINSRD	$2, h+4(FP), X0	// seed
	AESENC	·aeskeysched+0(SB), X0
	AESENC	·aeskeysched+16(SB), X0
	AESENC	·aeskeysched+32(SB), X0
	MOVL	X0, ret+8(FP)
	RET
noaes:
	JMP	·memhash64Fallback(SB)

// simple mask to get rid of data in the high part of the register.
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

// these are arguments to pshufb. They move data down from
// the high bytes of the register to the low bytes of the register.
// index is how many bytes to move.
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

TEXT ·checkASM(SB),NOSPLIT,$0-1
	// check that masks<>(SB) and shifts<>(SB) are aligned to 16-byte
	MOVL	$masks<>(SB), AX
	MOVL	$shifts<>(SB), BX
	ORL	BX, AX
	TESTL	$15, AX
	SETEQ	ret+0(FP)
	RET

TEXT ·return0(SB), NOSPLIT, $0
	MOVL	$0, AX
	RET

// Called from cgo wrappers, this function returns g->m->curg.stack.hi.
// Must obey the gcc calling convention.
TEXT _cgo_topofstack(SB),NOSPLIT,$0
	get_tls(CX)
	MOVL	g(CX), AX
	MOVL	g_m(AX), AX
	MOVL	m_curg(AX), AX
	MOVL	(g_stack+stack_hi)(AX), AX
	RET

// The top-most function running on a goroutine
// returns to goexit+PCQuantum.
TEXT ·goexit(SB),NOSPLIT|TOPFRAME,$0-0
	BYTE	$0x90	// NOP
	CALL	·goexit1(SB)	// does not return
	// traceback from goexit1 must hit code range of goexit
	BYTE	$0x90	// NOP

// Add a module's moduledata to the linked list of moduledata objects. This
// is called from .init_array by a function generated in the linker and so
// follows the platform ABI wrt register preservation -- it only touches AX,
// CX (implicitly) and DX, but it does not follow the ABI wrt arguments:
// instead the pointer to the moduledata is passed in AX.
TEXT ·addmoduledata(SB),NOSPLIT,$0-0
	MOVL	·lastmoduledatap(SB), DX
	MOVL	AX, moduledata_next(DX)
	MOVL	AX, ·lastmoduledatap(SB)
	RET

TEXT ·uint32tofloat64(SB),NOSPLIT,$8-12
	MOVL	a+0(FP), AX
	MOVL	AX, 0(SP)
	MOVL	$0, 4(SP)
	FMOVV	0(SP), F0
	FMOVDP	F0, ret+4(FP)
	RET

TEXT ·float64touint32(SB),NOSPLIT,$12-12
	FMOVD	a+0(FP), F0
	FSTCW	0(SP)
	FLDCW	·controlWord64trunc(SB)
	FMOVVP	F0, 4(SP)
	FLDCW	0(SP)
	MOVL	4(SP), AX
	MOVL	AX, ret+8(FP)
	RET

// gcWriteBarrier informs the GC about heap pointer writes.
//
// gcWriteBarrier returns space in a write barrier buffer which
// should be filled in by the caller.
// gcWriteBarrier does NOT follow the Go ABI. It accepts the
// number of bytes of buffer needed in DI, and returns a pointer
// to the buffer space in DI.
// It clobbers FLAGS. It does not clobber any general-purpose registers,
// but may clobber others (e.g., SSE registers).
// Typical use would be, when doing *(CX+88) = AX
//     CMPL    $0, runtime.writeBarrier(SB)
//     JEQ     dowrite
//     CALL    runtime.gcBatchBarrier2(SB)
//     MOVL    AX, (DI)
//     MOVL    88(CX), DX
//     MOVL    DX, 4(DI)
// dowrite:
//     MOVL    AX, 88(CX)
TEXT gcWriteBarrier<>(SB),NOSPLIT,$28
	// Save the registers clobbered by the fast path. This is slightly
	// faster than having the caller spill these.
	MOVL	CX, 20(SP)
	MOVL	BX, 24(SP)
retry:
	// TODO: Consider passing g.m.p in as an argument so they can be shared
	// across a sequence of write barriers.
	get_tls(BX)
	MOVL	g(BX), BX
	MOVL	g_m(BX), BX
	MOVL	m_p(BX), BX
	// Get current buffer write position.
	MOVL	(p_wbBuf+wbBuf_next)(BX), CX	// original next position
	ADDL	DI, CX				// new next position
	// Is the buffer full?
	CMPL	CX, (p_wbBuf+wbBuf_end)(BX)
	JA	flush
	// Commit to the larger buffer.
	MOVL	CX, (p_wbBuf+wbBuf_next)(BX)
	// Make return value (the original next position)
	SUBL	DI, CX
	MOVL	CX, DI
	// Restore registers.
	MOVL	20(SP), CX
	MOVL	24(SP), BX
	RET

flush:
	// Save all general purpose registers since these could be
	// clobbered by wbBufFlush and were not saved by the caller.
	MOVL	DI, 0(SP)
	MOVL	AX, 4(SP)
	// BX already saved
	// CX already saved
	MOVL	DX, 8(SP)
	MOVL	BP, 12(SP)
	MOVL	SI, 16(SP)
	// DI already saved

	CALL	·wbBufFlush(SB)

	MOVL	0(SP), DI
	MOVL	4(SP), AX
	MOVL	8(SP), DX
	MOVL	12(SP), BP
	MOVL	16(SP), SI
	JMP	retry

TEXT ·gcWriteBarrier1<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$4, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier2<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$8, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier3<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$12, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier4<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$16, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier5<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$20, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier6<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$24, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier7<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$28, DI
	JMP	gcWriteBarrier<>(SB)
TEXT ·gcWriteBarrier8<ABIInternal>(SB),NOSPLIT,$0
	MOVL	$32, DI
	JMP	gcWriteBarrier<>(SB)

// Note: these functions use a special calling convention to save generated code space.
// Arguments are passed in registers, but the space for those arguments are allocated
// in the caller's stack frame. These stubs write the args into that stack space and
// then tail call to the corresponding runtime handler.
// The tail call makes these stubs disappear in backtraces.
TEXT ·panicIndex(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicIndex(SB)
TEXT ·panicIndexU(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicIndexU(SB)
TEXT ·panicSliceAlen(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSliceAlen(SB)
TEXT ·panicSliceAlenU(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSliceAlenU(SB)
TEXT ·panicSliceAcap(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSliceAcap(SB)
TEXT ·panicSliceAcapU(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSliceAcapU(SB)
TEXT ·panicSliceB(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicSliceB(SB)
TEXT ·panicSliceBU(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicSliceBU(SB)
TEXT ·panicSlice3Alen(SB),NOSPLIT,$0-8
	MOVL	DX, x+0(FP)
	MOVL	BX, y+4(FP)
	JMP	·goPanicSlice3Alen(SB)
TEXT ·panicSlice3AlenU(SB),NOSPLIT,$0-8
	MOVL	DX, x+0(FP)
	MOVL	BX, y+4(FP)
	JMP	·goPanicSlice3AlenU(SB)
TEXT ·panicSlice3Acap(SB),NOSPLIT,$0-8
	MOVL	DX, x+0(FP)
	MOVL	BX, y+4(FP)
	JMP	·goPanicSlice3Acap(SB)
TEXT ·panicSlice3AcapU(SB),NOSPLIT,$0-8
	MOVL	DX, x+0(FP)
	MOVL	BX, y+4(FP)
	JMP	·goPanicSlice3AcapU(SB)
TEXT ·panicSlice3B(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSlice3B(SB)
TEXT ·panicSlice3BU(SB),NOSPLIT,$0-8
	MOVL	CX, x+0(FP)
	MOVL	DX, y+4(FP)
	JMP	·goPanicSlice3BU(SB)
TEXT ·panicSlice3C(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicSlice3C(SB)
TEXT ·panicSlice3CU(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	·goPanicSlice3CU(SB)
TEXT ·panicSliceConvert(SB),NOSPLIT,$0-8
	MOVL	DX, x+0(FP)
	MOVL	BX, y+4(FP)
	JMP	·goPanicSliceConvert(SB)

// Extended versions for 64-bit indexes.
TEXT ·panicExtendIndex(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendIndex(SB)
TEXT ·panicExtendIndexU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendIndexU(SB)
TEXT ·panicExtendSliceAlen(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSliceAlen(SB)
TEXT ·panicExtendSliceAlenU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSliceAlenU(SB)
TEXT ·panicExtendSliceAcap(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSliceAcap(SB)
TEXT ·panicExtendSliceAcapU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSliceAcapU(SB)
TEXT ·panicExtendSliceB(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendSliceB(SB)
TEXT ·panicExtendSliceBU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendSliceBU(SB)
TEXT ·panicExtendSlice3Alen(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	DX, lo+4(FP)
	MOVL	BX, y+8(FP)
	JMP	·goPanicExtendSlice3Alen(SB)
TEXT ·panicExtendSlice3AlenU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	DX, lo+4(FP)
	MOVL	BX, y+8(FP)
	JMP	·goPanicExtendSlice3AlenU(SB)
TEXT ·panicExtendSlice3Acap(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	DX, lo+4(FP)
	MOVL	BX, y+8(FP)
	JMP	·goPanicExtendSlice3Acap(SB)
TEXT ·panicExtendSlice3AcapU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	DX, lo+4(FP)
	MOVL	BX, y+8(FP)
	JMP	·goPanicExtendSlice3AcapU(SB)
TEXT ·panicExtendSlice3B(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSlice3B(SB)
TEXT ·panicExtendSlice3BU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	CX, lo+4(FP)
	MOVL	DX, y+8(FP)
	JMP	·goPanicExtendSlice3BU(SB)
TEXT ·panicExtendSlice3C(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendSlice3C(SB)
TEXT ·panicExtendSlice3CU(SB),NOSPLIT,$0-12
	MOVL	SI, hi+0(FP)
	MOVL	AX, lo+4(FP)
	MOVL	CX, y+8(FP)
	JMP	·goPanicExtendSlice3CU(SB)

#ifdef GOOS_android
// Use the free TLS_SLOT_APP slot #2 on Android Q.
// Earlier androids are set up in gcc_android.c.
DATA ·tls_g+0(SB)/4, $8
GLOBL ·tls_g+0(SB), NOPTR, $4
#endif
#ifdef GOOS_windows
GLOBL ·tls_g+0(SB), NOPTR, $4
#endif
