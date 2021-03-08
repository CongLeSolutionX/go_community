// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build race

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

// The following thunks allow calling the clang-compiled race runtime directly.
// An recap of the mips64 calling convention(N64).
// The R0 which is constant zero.
// The R1 for Assembler temporary.
// Values for function returns and expression evaluation in R2, R3.
// Arguments are passed in R4...R11, the rest is on stack.
// Call-preserverd registers are: R16...R23, R28..R31
// Call-clobbered registers are: R12..R15, R24, R25.
// OS Kernel reserved: R26, R27
// Global Pointer R28
// SP (R29)
// FP (R30)
// Return address (R31)

// The R25 is target address.
// g is alias of R30
// callstack

// func runtime·raceread(addr uintptr)
// Called from instrumented code.
TEXT	runtime·raceread(SB), NOSPLIT, $0-8
	MOVV	addr+0(FP), R5
	MOVV	R31, R6
	// void __tsan_read(ThreadState *thr, void *addr, void *pc);
	MOVV	$__tsan_read(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·RaceRead(addr uintptr)
TEXT	runtime·RaceRead(SB), NOSPLIT, $0-8
	// This needs to be a tail call, because raceread reads caller pc.
	JMP	runtime·raceread(SB)

// func runtime·racereadpc(void *addr, void *callpc, void *pc)
TEXT	runtime·racereadpc(SB), NOSPLIT, $0-24
	MOVV	addr+0(FP), R5
	MOVV	callpc+8(FP), R6
	MOVV	pc+16(FP), R7
	// void __tsan_read_pc(ThreadState *thr, void *addr, void *callpc, void *pc);
	MOVV	$__tsan_read_pc(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·racewrite(addr uintptr)
// Called from instrumented code.
TEXT	runtime·racewrite(SB), NOSPLIT, $0-8
	MOVV	addr+0(FP), R5
	MOVV	R31, R6
	// void __tsan_write(ThreadState *thr, void *addr, void *pc);
	MOVV	$__tsan_write(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·RaceWrite(addr uintptr)
TEXT	runtime·RaceWrite(SB), NOSPLIT, $0-8
	// This needs to be a tail call, because racewrite reads caller pc.
	JMP	runtime·racewrite(SB)

// func runtime·racewritepc(void *addr, void *callpc, void *pc)
TEXT	runtime·racewritepc(SB), NOSPLIT, $0-24
	MOVV	addr+0(FP), R5
	MOVV	callpc+8(FP), R6
	MOVV	pc+16(FP), R7
	// void __tsan_write_pc(ThreadState *thr, void *addr, void *callpc, void *pc);
	MOVV	$__tsan_write_pc(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·racereadrange(addr, size uintptr)
// Called from instrumented code.
TEXT	runtime·racereadrange(SB), NOSPLIT, $0-16
	MOVV	addr+0(FP), R5
	MOVV	size+8(FP), R6
	MOVV	R31, R7
	// void __tsan_read_range(ThreadState *thr, void *addr, uintptr size, void *pc);
	MOVV	$__tsan_read_range(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·RaceReadRange(addr, size uintptr)
TEXT	runtime·RaceReadRange(SB), NOSPLIT, $0-16
	// This needs to be a tail call, because racereadrange reads caller pc.
	JMP	runtime·racereadrange(SB)

// func runtime·racereadrangepc1(void *addr, uintptr sz, void *pc)
TEXT	runtime·racereadrangepc1(SB), NOSPLIT, $0-24
	MOVV	addr+0(FP), R5
	MOVV	size+8(FP), R6
	MOVV	pc+16(FP), R7
	ADDV	$8, R7	// pc is function start, tsan wants return address. 2x pc quantum for delay slot.
	// void __tsan_read_range(ThreadState *thr, void *addr, uintptr size, void *pc);
	MOVV	$__tsan_read_range(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·racewriterange(addr, size uintptr)
// Called from instrumented code.
TEXT	runtime·racewriterange(SB), NOSPLIT, $0-16
	MOVV	addr+0(FP), R5
	MOVV	size+8(FP), R6
	MOVV	R31, R7
	// void __tsan_write_range(ThreadState *thr, void *addr, uintptr size, void *pc);
	MOVV	$__tsan_write_range(SB), R25
	JMP	racecalladdr<>(SB)

// func runtime·RaceWriteRange(addr, size uintptr)
TEXT	runtime·RaceWriteRange(SB), NOSPLIT, $0-16
	// This needs to be a tail call, because racewriterange reads caller pc.
	JMP	runtime·racewriterange(SB)

// func runtime·racewriterangepc1(void *addr, uintptr sz, void *pc)
TEXT	runtime·racewriterangepc1(SB), NOSPLIT, $0-24
	MOVV	addr+0(FP), R5
	MOVV	size+8(FP), R6
	MOVV	pc+16(FP), R7
	ADDV	$8, R7	// pc is function start, tsan wants return address. 2x pc quantum for delay slot.
	// void __tsan_write_range(ThreadState *thr, void *addr, uintptr size, void *pc);
	MOVV	$__tsan_write_range(SB), R25
	JMP	racecalladdr<>(SB)

// If addr (R5) is out of range, do nothing.
// Otherwise, setup goroutine context and invoke racecall. Other arguments already set.
TEXT	racecalladdr<>(SB), NOSPLIT, $0-0
	JAL     runtime·load_g(SB)
	MOVV	g_racectx(g), R4
	// Check that addr is within [arenastart, arenaend) or within [racedatastart, racedataend).
	MOVV	runtime·racearenastart(SB), R12
	SGTU	R5, R12, R13
	BEQ	R13, R0, data

	MOVV	runtime·racearenaend(SB), R12
	SGTU	R5, R12, R13
	BEQ	R13, R0, call
data:
	MOVV	runtime·racedatastart(SB), R12
	SGTU	R5, R12, R13
	BEQ	R13, R0, ret

	MOVV	runtime·racedataend(SB), R12
	SGTU	R5, R12, R13
	BNE	R13, R0, ret
call:
	JAL	racecall<>(SB)
ret:
	RET

// func runtime·racefuncenterfp(fp uintptr)
// Called from instrumented code.
// Like racefuncenter but doesn't passes an arg, uses the caller pc
// from the first slot on the stack
TEXT	runtime·racefuncenterfp(SB), NOSPLIT, $0-0
	MOVV	0(R29), R5
	JMP	racefuncenter<>(SB)

// func runtime·racefuncenter(pc uintptr)
// Called from instrumented code.
TEXT	runtime·racefuncenter(SB), NOSPLIT, $0-8
	MOVV	callpc+0(FP), R5
	JMP	racefuncenter<>(SB)

// Common code for racefuncenter/racefuncenterfp
// R25 = caller's return address
TEXT	racefuncenter<>(SB), NOSPLIT, $0-0
	JAL     runtime·load_g(SB)
	MOVV	g_racectx(g), R4	// goroutine racectx
	// void __tsan_func_enter(ThreadState *thr, void *pc);
	MOVV	$__tsan_func_enter(SB), R25
	JAL	racecall<>(SB)
	RET

// func runtime·racefuncexit()
// Called from instrumented code.
TEXT	runtime·racefuncexit(SB), NOSPLIT, $0-0
	JAL	runtime·load_g(SB)
	MOVV	g_racectx(g), R4	// race context
	// void __tsan_func_exit(ThreadState *thr);
	MOVV	$__tsan_func_exit(SB), R25
	JAL	racecall<>(SB)
	RET

// Atomic operations for sync/atomic package.
// R25 = addr of arguments passed to this function, it can
// be fetched at 40(R29) in racecallatomic after two times JAL
// R4, R5, R6 set in racecallatomic

// Load
TEXT	sync∕atomic·LoadInt32(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic32_load(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·LoadInt64(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic64_load(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·LoadUint32(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·LoadInt32(SB)

TEXT	sync∕atomic·LoadUint64(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·LoadInt64(SB)

TEXT	sync∕atomic·LoadUintptr(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·LoadInt64(SB)

TEXT	sync∕atomic·LoadPointer(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·LoadInt64(SB)

// Store
TEXT	sync∕atomic·StoreInt32(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic32_store(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·StoreInt64(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic64_store(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·StoreUint32(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·StoreInt32(SB)

TEXT	sync∕atomic·StoreUint64(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·StoreInt64(SB)

TEXT	sync∕atomic·StoreUintptr(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·StoreInt64(SB)

// Swap
TEXT	sync∕atomic·SwapInt32(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic32_exchange(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·SwapInt64(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic64_exchange(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·SwapUint32(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·SwapInt32(SB)

TEXT	sync∕atomic·SwapUint64(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·SwapInt64(SB)

TEXT	sync∕atomic·SwapUintptr(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·SwapInt64(SB)

// Add
TEXT	sync∕atomic·AddInt32(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic32_fetch_add(SB), R25
	JAL	racecallatomic<>(SB)
	MOVV	add+8(FP), R4	// convert fetch_add to add_fetch
	MOVV	ret+16(FP), R5
	ADDV    R4, R5, R4
	MOVV	R4, ret+16(FP)
	RET

TEXT	sync∕atomic·AddInt64(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic64_fetch_add(SB), R25
	JAL	racecallatomic<>(SB)
	MOVV	add+8(FP), R4	// convert fetch_add to add_fetch
	MOVV	ret+16(FP), R5
	ADDV	R4, R5, R4
	MOVV	R4, ret+16(FP)
	RET

TEXT	sync∕atomic·AddUint32(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·AddInt32(SB)

TEXT	sync∕atomic·AddUint64(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·AddInt64(SB)

TEXT	sync∕atomic·AddUintptr(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·AddInt64(SB)

// CompareAndSwap
TEXT	sync∕atomic·CompareAndSwapInt32(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic32_compare_exchange(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·CompareAndSwapInt64(SB), NOSPLIT, $0
	GO_ARGS
	MOVV	$__tsan_go_atomic64_compare_exchange(SB), R25
	JAL	racecallatomic<>(SB)
	RET

TEXT	sync∕atomic·CompareAndSwapUint32(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·CompareAndSwapInt32(SB)

TEXT	sync∕atomic·CompareAndSwapUint64(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·CompareAndSwapInt64(SB)

TEXT	sync∕atomic·CompareAndSwapUintptr(SB), NOSPLIT, $0
	GO_ARGS
	JMP	sync∕atomic·CompareAndSwapInt64(SB)

// Generic atomic operation implementation.
// R25 = addr of target function
TEXT	racecallatomic<>(SB), NOSPLIT, $0
	// Set up these registers
	// R4 = *ThreadState
	// R5 = caller pc
	// R6 = pc
	// R7 = addr of incoming arg list

	// Trigger SIGSEGV early.
	MOVV	24(R29), R7	// 1st arg is addr. after two times JAL, get it at 24(R29)
	MOVV	(R7), R0	// segv here if addr is bad
	// Check that addr is within [arenastart, arenaend) or within [racedatastart, racedataend).
	MOVV	runtime·racearenastart(SB), R12
	SGTU	R7, R12, R13 // addr < start ?
	BEQ	R13, R0, racecallatomic_data
	MOVV	runtime·racearenaend(SB), R12
	SGTU	R12, R7, R13
	BNE	R13, R0, racecallatomic_ok

racecallatomic_data:
	MOVV	runtime·racedatastart(SB), R12
	SGTU	R7, R12, R13
	BEQ	R13, R0, racecallatomic_ignore

	MOVV	runtime·racedataend(SB), R12
	SGTU	R12, R7, R13
	BEQ	R13, R0, racecallatomic_ignore

racecallatomic_ok:
	// Addr is within the good range, call the atomic function.
	JAL	runtime·load_g(SB)
	MOVV	g_racectx(g), R4	// goroutine context
	MOVV	16(R29), R5		// caller pc
	MOVV	R25, R6			// pc
	ADDV	$24, R29, R7		// 1st argument address
	JAL	racecall<>(SB)
	RET

racecallatomic_ignore:
	// Addr is outside the good range.
	// Call __tsan_go_ignore_sync_begin to ignore synchronization during the atomic op.
	// An attempt to synchronize on the address would cause crash.
	MOVV	R25, R22	// remember the original function
	MOVV	$__tsan_go_ignore_sync_begin(SB), R25
	JAL     runtime·load_g(SB)
	MOVV	g_racectx(g), R4	// goroutine context
	JAL	racecall<>(SB)
	MOVV	R22, R25	// restore the original function

	// Call the atomic function.
	// racecall will call LLVM race code which might clobber R28 (g)
	JAL     runtime·load_g(SB)
	MOVV	g_racectx(g), R4	// goroutine context
	MOVV	16(R29), R5		// caller pc
	MOVV	R25, R6			// pc
	ADDV	$24, R29, R7		// arguments
	JAL	racecall<>(SB)

	// Call __tsan_go_ignore_sync_end.
	MOVV	$__tsan_go_ignore_sync_end(SB), R25
	MOVV	g_racectx(g), R4	// goroutine context
	JAL	racecall<>(SB)
	RET

// func runtime·racecall(void(*f)(...), ...)
// Calls C function f from race runtime and passes up to 4 arguments to it.
// The arguments are never heap-object-preserving pointers, so we pretend there are no arguments.
TEXT	runtime·racecall(SB), NOSPLIT, $0-40
	MOVV	fn+0(FP), R25
	MOVV	arg0+8(FP), R4
	MOVV	arg1+16(FP), R5
	MOVV	arg2+24(FP), R6
	MOVV	arg3+32(FP), R7
	JMP	racecall<>(SB)

// Switches SP to g0 stack and calls (R25). Arguments already set.
// No stack space is needed
TEXT	racecall<>(SB), NOSPLIT, $0

	MOVV	R29, R17
	JAL	runtime·load_g(SB)
	MOVV	g_m(g), R3
	MOVV	m_g0(R3), R3
	BEQ	g, R3, call
	MOVV	(g_sched+gobuf_sp)(R3), R29
call:

	JAL	(R25)
	MOVV	R17, R29

	RET

// C->Go callback thunk that allows to call runtime·racesymbolize from C code.
// Direct Go->C race call has only switched SP, finish g->g0 switch by setting correct g.
// The overall effect of Go->C->Go call chain is similar to that of mcall.
// R4 contains command code. R5 contains command-specific context.
// See racecallback for command codes.
TEXT	runtime·racecallbackthunk(SB), NOSPLIT, $0
	BNE	R4, R0, rest

	// load_g will clobbered R3 and llvm treat g(R30) as fp
	MOVV	R3, R14
	MOVV	g, R15
	JAL	runtime·load_g(SB)

	MOVV	g_m(g), R4
	MOVV	m_p(R4), R4
	MOVV	p_raceprocctx(R4), R4
	MOVV	R4, (R5)

	MOVV	R14, R3
	MOVV	R15, g // restore g

	RET

rest:
#ifndef GOMIPS64_softfloat
	ADDV	$(-8*23), R29
#else
	ADDV	$(-8*15), R29
#endif
	MOVV	R31, (8*0)(R29)
	MOVV	R16, (8*5)(R29)
	MOVV	R17, (8*6)(R29)
	MOVV	R18, (8*7)(R29)
	MOVV	R19, (8*8)(R29)
	MOVV	R20, (8*9)(R29)
	MOVV	R21, (8*10)(R29)
	MOVV	R22, (8*11)(R29)
	MOVV	R23, (8*12)(R29)
	MOVV	RSB, (8*13)(R29)
	MOVV	g,   (8*14)(R29)
#ifndef GOMIPS64_softfloat
	MOVD	F24, (8*15)(R29)
	MOVD	F25, (8*16)(R29)
	MOVD	F26, (8*17)(R29)
	MOVD	F27, (8*18)(R29)
	MOVD	F28, (8*19)(R29)
	MOVD	F29, (8*20)(R29)
	MOVD	F30, (8*21)(R29)
	MOVD	F31, (8*22)(R29)
#endif

	JAL	runtime·load_g(SB)
	MOVV	g_m(g), R3
	MOVV	m_g0(R3), R3
	BEQ	R3, g, noswitch // branch if already on g0
	MOVV	R3, g

noswitch:
	MOVV	R4, (8*1)(R29) // cmd
	MOVV	R5, (8*2)(R29) // *ctx

	BGEZAL  R0, 1(PC) // just like mcall
	SRLV    $32, R31, RSB
	SLLV    $32, RSB
	JAL	runtime·racecallback(SB)
	JAL	runtime·load_g(SB)

ret:
	// restore callee saved register
	MOVV	(8*0)(R29), R31
	MOVV	(8*5)(R29), R16
	MOVV	(8*6)(R29), R17
	MOVV	(8*7)(R29), R18
	MOVV	(8*8)(R29), R19
	MOVV	(8*9)(R29), R20
	MOVV	(8*10)(R29), R21
	MOVV	(8*11)(R29), R22
	MOVV	(8*12)(R29), R23
	MOVV	(8*13)(R29), RSB
	MOVV	(8*14)(R29), g
#ifndef GOMIPS64_softfloat
	MOVD	(8*15)(R29), F24
	MOVD	(8*16)(R29), F25
	MOVD	(8*17)(R29), F26
	MOVD	(8*18)(R29), F27
	MOVD	(8*19)(R29), F28
	MOVD	(8*20)(R29), F29
	MOVD	(8*21)(R29), F30
	MOVD	(8*22)(R29), F31
	ADDV	$(8*23), R29
#else
	ADDV	$(8*15), R29
#endif
	RET
