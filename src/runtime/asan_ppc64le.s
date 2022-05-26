// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build asan

#include "go_asm.h"
#include "textflag.h"

#define RARG0 R3
#define RARG1 R4
#define RARG2 R5
#define RARG3 R6
#define FARG R8

// Called from instrumented code.
// func runtime·doasanread(addr unsafe.Pointer, sz, sp, pc uintptr)
TEXT	runtime·doasanread(SB), NOSPLIT, $0-32
	MOVD	addr+0(FP), RARG0
	MOVD	size+8(FP), RARG1
	MOVD	sp+16(FP), RARG2
	MOVD	pc+24(FP), RARG3
	// void __asan_read_go(void *addr, uintptr_t sz, void *sp, void *pc);
	MOVD	$__asan_read_go(SB), FARG
	BR	asancall<>(SB)

// func runtime·doasanwrite(addr unsafe.Pointer, sz, sp, pc uintptr)
TEXT	runtime·doasanwrite(SB), NOSPLIT, $0-32
	MOVD	addr+0(FP), RARG0
	MOVD	size+8(FP), RARG1
	MOVD	sp+16(FP), RARG2
	MOVD	pc+24(FP), RARG3
	// void __asan_write_go(void *addr, uintptr_t sz, void *sp, void *pc);
	MOVD	$__asan_write_go(SB), FARG
	BR	asancall<>(SB)

// func runtime·asanunpoison(addr unsafe.Pointer, sz uintptr)
TEXT	runtime·asanunpoison(SB), NOSPLIT, $0-16
	MOVD	addr+0(FP), RARG0
	MOVD	size+8(FP), RARG1
	// void __asan_unpoison_go(void *addr, uintptr_t sz);
	MOVD	$__asan_unpoison_go(SB), FARG
	BR	asancall<>(SB)

// func runtime·asanpoison(addr unsafe.Pointer, sz uintptr)
TEXT	runtime·asanpoison(SB), NOSPLIT, $0-16
	MOVD	addr+0(FP), RARG0
	MOVD	size+8(FP), RARG1
	// void __asan_poison_go(void *addr, uintptr_t sz);
	MOVD	$__asan_poison_go(SB), FARG
	BR	asancall<>(SB)

// func runtime·asanregisterglobals(addr unsafe.Pointer, n uintptr)
TEXT	runtime·asanregisterglobals(SB), NOSPLIT, $0-16
	MOVD	addr+0(FP), RARG0
	MOVD	size+8(FP), RARG1
	// void __asan_register_globals_go(void *addr, uintptr_t n);
	MOVD	$__asan_register_globals_go(SB), FARG
	BR	asancall<>(SB)

// Switches SP to g0 stack and calls (FARG). Arguments already set.
TEXT	asancall<>(SB), NOSPLIT, $0-0
	// Set the LR slot for the ppc64 ABI
	MOVD	LR, R10
	MOVD	R10, 0(R1)	// Go expectation
	MOVD	R10, 16(R1)	// C ABI
	// Get info from the current goroutine
	MOVD	runtime·tls_g(SB), R10	// g offset in TLS
	MOVD	0(R10), g
	MOVD	g_m(g), R7		// m for g
	MOVD	R1, R16			// callee-saved, preserved across C call
	MOVD	m_g0(R7), R10		// g0 for m
	CMP	R10, g			// same g0?
	BEQ	call			// already on g0
	MOVD	(g_sched+gobuf_sp)(R10), R1 // switch R1
call:
	// prepare frame for C ABI
	SUB	$32, R1			// create frame for callee saving LR, CR, R2 etc.
	RLDCR	$0, R1, $~15, R1	// align SP to 16 bytes
	MOVD	FARG, CTR		// R8 = caller addr
	MOVD	FARG, R12		// expected by PPC64 ABI
	BL	(CTR)
	XOR	R0, R0			// clear R0 on return from C/C++ section
	MOVD	R16, R1			// restore R1; R16 nonvol in C/C++ section
	MOVD	runtime·tls_g(SB), R10	// find correct g
	MOVD	0(R10), g
	MOVD	16(R1), R10		// LR was saved away, restore for return
	MOVD	R10, LR
	RET

// tls_g, g value for each thread in TLS
GLOBL runtime·tls_g+0(SB), TLSBSS+DUPOK, $8
