// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//go:build wasm32

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

TEXT runtime·rt0_go(SB), NOSPLIT|NOFRAME|TOPFRAME, $0
	// save m->g0 = g0
	MOVW $runtime·g0(SB), runtime·m0+m_g0(SB)
	// save m0 to g0->m
	MOVW $runtime·m0(SB), runtime·g0+g_m(SB)
	// set g to g0
	MOVW $runtime·g0(SB), g
	CALLNORESUME runtime·check(SB)
#ifdef GOOS_js
	CALLNORESUME runtime·args(SB)
#endif
	CALLNORESUME runtime·osinit(SB)
	CALLNORESUME runtime·schedinit(SB)
	MOVW $runtime·mainPC(SB), 0(SP)
	CALLNORESUME runtime·newproc(SB)
	CALL runtime·mstart(SB) // WebAssembly stack will unwind when switching to another goroutine
	UNDEF

TEXT runtime·mstart(SB),NOSPLIT|TOPFRAME,$0
	CALL	runtime·mstart0(SB)
	RET // not reached

DATA  runtime·mainPC+0(SB)/4,$runtime·main(SB)
GLOBL runtime·mainPC(SB),RODATA,$4

// func checkASM() bool
TEXT ·checkASM(SB), NOSPLIT, $0-1
	MOVB $1, ret+0(FP)
	RET

TEXT runtime·checkcomp(SB), NOSPLIT, $0-8
	MOVB $1, ret+0(FP)
	RET

TEXT runtime·gogo(SB), NOSPLIT, $0-8
	MOVW buf+0(FP), R0
	MOVW gobuf_g(R0), R1
	MOVW 0(R1), R2	// make sure g != nil
	MOVW R1, g
	MOVW gobuf_sp(R0), SP

	// Put target PC at -8(SP), wasm_pc_f_loop will pick it up
	Get SP
	I32Const $4
	I32Sub
	I32Load gobuf_pc(R0)
	I32Store $0

	MOVW gobuf_ret(R0), RET0
	MOVW gobuf_ctxt(R0), CTXT
	// clear to help garbage collector
	MOVW $0, gobuf_sp(R0)
	MOVW $0, gobuf_ret(R0)
	MOVW $0, gobuf_ctxt(R0)

	I32Const $1
	Return

// func mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return. It should gogo(&g->sched)
// to keep running g.
TEXT runtime·mcall(SB), NOSPLIT, $0-8
	// CTXT = fn
	MOVW fn+0(FP), CTXT
	// R1 = g.m
	MOVW g_m(g), R1
	// R2 = g0
	MOVW m_g0(R1), R2

	// save state in g->sched
	MOVW 0(SP), g_sched+gobuf_pc(g)     // caller's PC
	MOVW $fn+0(FP), g_sched+gobuf_sp(g) // caller's SP

	// if g == g0 call badmcall
	Get g
	Get R2
	I32Eq
	If
		JMP runtime·badmcall(SB)
	End

	// switch to g0's stack
	I32Load (g_sched+gobuf_sp)(R2)
	I32Const $8
	I32Sub
	Set SP

	// set arg to current g
	MOVW g, 0(SP)

	// switch to g0
	MOVW R2, g

	// call fn
	Get CTXT
	I32Load $0
	CALL

	Get SP
	I32Const $4
	I32Add
	Set SP

	JMP runtime·badmcall2(SB)

// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
	// R0 = fn
	MOVW fn+0(FP), R0
	// R1 = g.m
	MOVW g_m(g), R1
	// R2 = g0
	MOVW m_g0(R1), R2

	// if g == g0
	Get g
	Get R2
	I32Eq
	If
		// no switch:
		MOVW R0, CTXT

		Get CTXT
		I32Load $0
		JMP
	End

	// if g != m.curg
	Get g
	I32Load m_curg(R1)
	I32Ne
	If
		CALLNORESUME runtime·badsystemstack(SB)
		CALLNORESUME runtime·abort(SB)
	End

	// switch:

	// save state in g->sched. Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOVW $runtime·systemstack_switch(SB), g_sched+gobuf_pc(g)

	MOVW SP, g_sched+gobuf_sp(g)

	// switch to g0
	MOVW R2, g

	// make it look like mstart called systemstack on g0, to stop traceback
	I32Load (g_sched+gobuf_sp)(R2)
	I32Const $8
	I32Sub
	Set R3

	MOVW $runtime·mstart(SB), 0(R3)
	MOVW R3, SP

	// call fn
	MOVW R0, CTXT

	Get CTXT
	I32Load $0
	CALL

	// switch back to g
	MOVW g_m(g), R1
	MOVW m_curg(R1), R2
	MOVW R2, g
	MOVW g_sched+gobuf_sp(R2), SP
	MOVW $0, g_sched+gobuf_sp(R2)
	RET

TEXT runtime·systemstack_switch(SB), NOSPLIT, $0-0
	RET

TEXT runtime·abort(SB),NOSPLIT|NOFRAME,$0-0
	UNDEF

// AES hashing not implemented for wasm
TEXT runtime·memhash(SB),NOSPLIT|NOFRAME,$0-32
	JMP	runtime·memhashFallback(SB)
TEXT runtime·strhash(SB),NOSPLIT|NOFRAME,$0-24
	JMP	runtime·strhashFallback(SB)
TEXT runtime·memhash32(SB),NOSPLIT|NOFRAME,$0-24
	JMP	runtime·memhash32Fallback(SB)
TEXT runtime·memhash64(SB),NOSPLIT|NOFRAME,$0-24
	JMP	runtime·memhash64Fallback(SB)

TEXT runtime·return0(SB), NOSPLIT, $0-0
	MOVW $0, RET0
	RET

TEXT runtime·asminit(SB), NOSPLIT, $0-0
	// No per-thread init.
	RET

TEXT ·publicationBarrier(SB), NOSPLIT, $0-0
	RET

TEXT runtime·procyield(SB), NOSPLIT, $0-0 // FIXME
	RET

TEXT runtime·breakpoint(SB), NOSPLIT, $0-0
	UNDEF

// func switchToCrashStack0(fn func())
TEXT runtime·switchToCrashStack0(SB), NOSPLIT, $0-8
	MOVW fn+0(FP), CTXT	// context register
	MOVW	g_m(g), R2	// curm

	// set g to gcrash
	MOVW	$runtime·gcrash(SB), g	// g = &gcrash
	MOVW	R2, g_m(g)	// g.m = curm
	MOVW	g, m_g0(R2)	// curm.g0 = g

	// switch to crashstack
	I32Load (g_stack+stack_hi)(g)
	I32Const $(-4*8)
	I32Add
	Set SP

	// call target function
	Get CTXT
	I32Load $0
	CALL

	// should never return
	CALL	runtime·abort(SB)
	UNDEF

// Called during function prolog when more stack is needed.
//
// The traceback routines see morestack on a g0 as being
// the top of a stack (for example, morestack calling newstack
// calling the scheduler calling newm calling gc), so we must
// record an argument size. For that purpose, it has no arguments.
TEXT runtime·morestack(SB), NOSPLIT, $0-0
	// R1 = g.m
	MOVW g_m(g), R1

	// R2 = g0
	MOVW m_g0(R1), R2

	// Set g->sched to context in f.
	NOP	SP	// tell vet SP changed - stop checking offsets
	MOVW 0(SP), g_sched+gobuf_pc(g)
	MOVW $4(SP), g_sched+gobuf_sp(g) // f's SP
	MOVW CTXT, g_sched+gobuf_ctxt(g)

	// Cannot grow scheduler stack (m->g0).
	Get g
	Get R2
	I32Eq
	If
		CALLNORESUME runtime·badmorestackg0(SB)
		CALLNORESUME runtime·abort(SB)
	End

	// Cannot grow signal stack (m->gsignal).
	Get g
	I32Load m_gsignal(R1)
	I32Eq
	If
		CALLNORESUME runtime·badmorestackgsignal(SB)
		CALLNORESUME runtime·abort(SB)
	End

	// Called from f.
	// Set m->morebuf to f's caller.
	MOVW 4(SP), m_morebuf+gobuf_pc(R1)
	MOVW $8(SP), m_morebuf+gobuf_sp(R1) // f's caller's SP
	MOVW g, m_morebuf+gobuf_g(R1)

	// Call newstack on m->g0's stack.
	MOVW R2, g
	MOVW g_sched+gobuf_sp(R2), SP
	CALL runtime·newstack(SB)
	UNDEF // crash if newstack returns

// morestack but not preserving ctxt.
TEXT runtime·morestack_noctxt(SB),NOSPLIT,$0
	MOVW $0, CTXT
	JMP runtime·morestack(SB)

TEXT ·asmcgocall(SB), NOSPLIT, $0-0
	UNDEF

#define DISPATCH(NAME, MAXSIZE) \
	Get R0; \
	I32Const $MAXSIZE; \
	I32LeU; \
	If; \
		JMP NAME(SB); \
	End

// func reflectcall(stackArgsType *_type, fn, stackArgs unsafe.Pointer, stackArgsSize, stackRetOffset, frameSize uint32, regArgs *abi.RegArgs)
TEXT ·reflectcall(SB), NOSPLIT, $0-48
	I32Load fn+4(FP)
	I32Eqz
	If
		CALLNORESUME runtime·sigpanic<ABIInternal>(SB)
	End

	MOVW frameSize+16(FP), R0

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
	JMP runtime·badreflectcall(SB)

#define CALLFN(NAME, MAXSIZE) \
TEXT NAME(SB), WRAPPER, $MAXSIZE-48; \
	NO_LOCAL_POINTERS; \
	MOVW stackArgsSize+8(FP), R0; \
	\
	Get R0; \
	I32Eqz; \
	Not; \
	If; \
		Get SP; \
		I32Load stackArgs+4(FP); \
		I32Load stackArgsSize+8(FP); \
		MemoryCopy; \
	End; \
	\
	MOVW f+4(FP), CTXT; \
	Get CTXT; \
	I32Load $0; \
	CALL; \
	\
	I32Load stackRetOffset+10(FP); \
	Set R0; \
	\
	MOVW stackArgsType+0(FP), RET0; \
	\
	I32Load stackArgs+4(FP); \
	Get R0; \
	I32Add; \
	Set RET1; \
	\
	Get SP; \
	Get R0; \
	I32Add; \
	Set RET2; \
	\
	I32Load stackArgsSize+8(FP); \
	Get R0; \
	I32Sub; \
	Set RET3; \
	\
	CALL callRet<>(SB); \
	RET

// callRet copies return values back at the end of call*. This is a
// separate function so it can allocate stack space for the arguments
// to reflectcallmove. It does not follow the Go ABI; it expects its
// arguments in registers.
TEXT callRet<>(SB), NOSPLIT, $40-0
	NO_LOCAL_POINTERS
	MOVW RET0, 0(SP)
	MOVW RET1, 4(SP)
	MOVW RET2, 8(SP)
	MOVW RET3, 12(SP)
	MOVW $0,   16(SP)
	CALL runtime·reflectcallmove(SB)
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

TEXT runtime·goexit(SB), NOSPLIT|TOPFRAME, $0-0
	NOP // first PC of goexit is skipped
	CALL runtime·goexit1(SB) // does not return
	UNDEF

TEXT runtime·cgocallback(SB), NOSPLIT, $0-24
	UNDEF

// gcWriteBarrier informs the GC about heap pointer writes.
//
// gcWriteBarrier does NOT follow the Go ABI. It accepts the
// number of bytes of buffer needed as a wasm argument
// (put on the TOS by the caller, lives in local R0 in this body)
// and returns a pointer to the buffer space as a wasm result
// (left on the TOS in this body, appears on the wasm stack
// in the caller).
TEXT gcWriteBarrier<>(SB), NOSPLIT, $0
	Loop
		// R3 = g.m
		MOVW g_m(g), R3
		// R4 = p
		MOVW m_p(R3), R4
		// R5 = wbBuf.next
		MOVW p_wbBuf+wbBuf_next(R4), R5

		// Increment wbBuf.next
		Get R5
		Get R0
		I32Add
		Set R5

		// Is the buffer full?
		Get R5
		I32Load (p_wbBuf+wbBuf_end)(R4)
		I32LeU
		If
			// Commit to the larger buffer.
			MOVW R5, p_wbBuf+wbBuf_next(R4)

			// Make return value (the original next position)
			Get R5
			Get R0
			I32Sub

			Return
		End

		// Flush
		CALLNORESUME runtime·wbBufFlush(SB)

		// Retry
		Br $0
	End

TEXT runtime·gcWriteBarrier1<ABIInternal>(SB),NOSPLIT,$0
	I32Const $8
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier2<ABIInternal>(SB),NOSPLIT,$0
	I32Const $16
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier3<ABIInternal>(SB),NOSPLIT,$0
	I32Const $24
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier4<ABIInternal>(SB),NOSPLIT,$0
	I32Const $32
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier5<ABIInternal>(SB),NOSPLIT,$0
	I32Const $40
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier6<ABIInternal>(SB),NOSPLIT,$0
	I32Const $48
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier7<ABIInternal>(SB),NOSPLIT,$0
	I32Const $56
	Call	gcWriteBarrier<>(SB)
	Return
TEXT runtime·gcWriteBarrier8<ABIInternal>(SB),NOSPLIT,$0
	I32Const $64
	Call	gcWriteBarrier<>(SB)
	Return

TEXT wasm_pc_f_loop(SB),NOSPLIT,$0
// Call the function for the current PC_F. Repeat until PAUSE != 0 indicates pause or exit.
// The WebAssembly stack may unwind, e.g. when switching goroutines.
// The Go stack on the linear memory is then used to jump to the correct functions
// with this loop, without having to restore the full WebAssembly stack.
// It is expected to have a pending call before entering the loop, so check PAUSE first.
	Get PAUSE
	I32Eqz
	If
	loop:
		Loop
			// Get PC_B & PC_F from -4(SP)
			Get SP
			I32Const $4
			I32Sub
			I32Load16U $0 // PC_B

			Get SP
			I32Const $4
			I32Sub
			I32Load16U $2 // PC_F

			CallIndirect $0
			Drop

			Get PAUSE
			I32Eqz
			BrIf loop
		End
	End

	I32Const $0
	Set PAUSE

	Return

TEXT wasm_export_lib(SB),NOSPLIT,$0
	UNDEF

// Extended versions for 64-bit indexes.
TEXT runtime·panicExtendIndex(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALLNORESUME	runtime·goPanicExtendIndex(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendIndexU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALLNORESUME	runtime·goPanicExtendIndexU(SB)
	RET
TEXT runtime·panicExtendSliceAlen(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceAlen(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSliceAlenU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceAlenU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSliceAcap(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceAcap(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSliceAcapU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceAcapU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSliceB(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceB(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSliceBU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSliceBU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3Alen(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3Alen(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3AlenU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3AlenU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3Acap(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3Acap(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3AcapU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3AcapU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3B(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3B(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3BU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3BU(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3C(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3C(SB)
	Set R0 // basically drop
	RET
TEXT runtime·panicExtendSlice3CU(SB),NOSPLIT,$0-12
	MOVW	R0, hi+0(FP)
	MOVW	R1, lo+4(FP)
	MOVW	R2, y+8(FP)
	CALL	runtime·goPanicExtendSlice3CU(SB)
	Set R0 // basically drop
	RET
