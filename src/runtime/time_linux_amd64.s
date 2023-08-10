// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !faketime

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

#define SYS_clock_gettime	228

#define CLOCK_REALTIME			0
#define CLOCK_MONOTONIC			1
#define CLOCK_PROCESS_CPUTIME_ID	2
#define CLOCK_THREAD_CPUTIME_ID		3
#define CLOCK_MONOTONIC_RAW		4
#define CLOCK_REALTIME_COARSE		5
#define CLOCK_MONOTONIC_COARSE		6
#define CLOCK_BOOTTIME			7
#define CLOCK_REALTIME_ALARM		8
#define CLOCK_BOOTTIME_ALARM		9

#define _1e9 1000000000

// func time.now() (sec int64, nsec int32, mono int64)
TEXT time·now<ABIInternal>(SB),NOSPLIT,$16-24
	MOVQ	SP, R12 // Save old SP; R12 unchanged by C code.

	MOVQ	g_m(R14), BX // BX unchanged by C code.

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	// Save the old values on stack and restore them on exit,
	// so this function is reentrant.
	MOVQ	m_vdsoPC(BX), CX
	MOVQ	m_vdsoSP(BX), DX
	MOVQ	CX, 0(SP)
	MOVQ	DX, 8(SP)

	LEAQ	sec+0(FP), DX
	MOVQ	-8(DX), CX	// Sets CX to function return address.
	MOVQ	CX, m_vdsoPC(BX)
	MOVQ	DX, m_vdsoSP(BX)

	CMPQ	R14, m_curg(BX)	// Only switch if on curg.
	JNE	noswitch

	MOVQ	m_g0(BX), DX
	MOVQ	(g_sched+gobuf_sp)(DX), SP	// Set SP to g0 stack

noswitch:
	// Six time results:
	//
	//	 0(SP) - CLOCK_MONOTONIC
	//	16(SP) - CLOCK_REALTIME
	//	32(SP) - CLOCK_MONOTONIC_COARSE (before)
	//	48(SP) - CLOCK_REALTIME_COARSE (before)
	//	64(SP) - CLOCK_MONOTONIC_COARSE (after)
	//	80(SP) - CLOCK_REALTIME_COARSE (after)
	//
	SUBQ	$96, SP		// Space for six time results
	ANDQ	$~15, SP	// Align for C code

	MOVL	$CLOCK_MONOTONIC_COARSE, DI
	LEAQ	32(SP), SI
	MOVQ	runtime·vdsoClockgettimeSym(SB), R13
	CMPQ	R13, $0
	JEQ	fallback
	CALL	R13

	MOVL	$CLOCK_REALTIME_COARSE, DI
	LEAQ	48(SP), SI
	CALL	R13

	MOVL	$CLOCK_REALTIME, DI
	LEAQ	16(SP), SI
	CALL	R13

	MOVL	$CLOCK_MONOTONIC_COARSE, DI
	LEAQ	64(SP), SI
	CALL	R13

	MOVQ	32(SP), AX // monotonic coarse sec
	MOVQ	40(SP), R8 // monotonic coarse nsec
	MOVQ	64(SP), CX // monotonic coarse sec
	MOVQ	72(SP), R10 // monotonic coarse nsec
	CMPQ	AX, CX
	JNE	mismatch
	CMPQ	R8, R10
	JNE	mismatch
	IMULQ	$_1e9, AX
	ADDQ	AX, R8

	MOVQ	48(SP), AX // realtime coarse sec
	MOVQ	56(SP), R9 // realtime coarse nsec
	IMULQ	$_1e9, AX
	ADDQ	AX, R9

	MOVQ	16(SP), CX // realtime actual sec
	MOVQ	CX, AX
	MOVQ	24(SP), DI // realtime actual nsec
	IMULQ	$_1e9, CX
	ADDQ	DI, CX

	SUBQ	R9, CX
	ADDQ	R8, CX
	MOVQ	$0, DX
	JMP 	regret

mismatch:
	MOVL	$CLOCK_MONOTONIC, DI
	LEAQ	0(SP), SI
	CALL	R13

ret:
	MOVQ	16(SP), AX	// realtime sec
	MOVQ	24(SP), DI	// realtime nsec (moved to BX below)
	MOVQ	0(SP), CX	// monotonic sec
	IMULQ	$_1e9, CX
	MOVQ	8(SP), DX	// monotonic nsec

regret:
	MOVQ	R12, SP		// Restore real SP

	// Restore vdsoPC, vdsoSP
	// We don't worry about being signaled between the two stores.
	// If we are not in a signal handler, we'll restore vdsoSP to 0,
	// and no one will care about vdsoPC. If we are in a signal handler,
	// we cannot receive another signal.
	MOVQ	8(SP), SI
	MOVQ	SI, m_vdsoSP(BX)
	MOVQ	0(SP), SI
	MOVQ	SI, m_vdsoPC(BX)

	// set result registers; AX is already correct
	MOVQ	DI, BX
	ADDQ	DX, CX
	RET

fallback:
	MOVL	$CLOCK_REALTIME, DI
	LEAQ	16(SP), SI
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL

	MOVL	$CLOCK_MONOTONIC, DI
	LEAQ	0(SP), SI
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL

	JMP	ret
