#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"
#include "core_asm.h"

TEXT runtime∕internal∕schedinit·rt0_go(SB),NOSPLIT,$0
	// copy arguments forward on an even stack
	MOVQ	DI, AX		// argc
	MOVQ	SI, BX		// argv
	SUBQ	$(4*8+7), SP		// 2args 2auto
	ANDQ	$~15, SP
	MOVQ	AX, 16(SP)
	MOVQ	BX, 24(SP)
	
	// create istack out of the given (operating system) stack.
	// _cgo_init may update stackguard.
	MOVQ	$runtime∕internal∕core·g0(SB), DI
	LEAQ	(-64*1024+104)(SP), BX
	MOVQ	BX, G_Stackguard0(DI)
	MOVQ	BX, G_Stackguard1(DI)
	MOVQ	BX, (G_Stack+Stack_Lo)(DI)
	MOVQ	SP, (G_Stack+Stack_Hi)(DI)

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
	
	// if there is an _cgo_init, call it.
	MOVQ	_cgo_init(SB), AX
	TESTQ	AX, AX
	JZ	needtls
	// g0 already in DI
	MOVQ	DI, CX	// Win64 uses CX for first parameter
	MOVQ	$setg_gcc<>(SB), SI
	CALL	AX

	// update stackguard after _cgo_init
	MOVQ	$runtime∕internal∕core·g0(SB), CX
	MOVQ	(G_Stack+Stack_Lo)(CX), AX
	ADDQ	$const_StackGuard, AX
	MOVQ	AX, G_Stackguard0(CX)
	MOVQ	AX, G_Stackguard1(CX)

	CMPL	runtime·iswindows(SB), $0
	JEQ ok
needtls:
	// skip TLS setup on Plan 9
	CMPL	runtime·isplan9(SB), $1
	JEQ ok
	// skip TLS setup on Solaris
	CMPL	runtime·issolaris(SB), $1
	JEQ ok

	LEAQ	runtime·tls0(SB), DI
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
	LEAQ	runtime∕internal∕core·g0(SB), CX
	MOVQ	CX, g(BX)
	LEAQ	runtime∕internal∕core·g0(SB), AX

	// save m->g0 = g0
	MOVQ	CX, M_G0(AX)
	// save m0 to g0->m
	MOVQ	AX, G_M(CX)

	CLD				// convention is D is always left cleared
	CALL	runtime∕internal∕check·check(SB)

	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP)
	CALL	runtime∕internal∕vdso·args(SB)
	CALL	runtime·osinit(SB)
	CALL    runtime∕internal∕schedinit·schedinit(SB)

	// create a new goroutine to start program
	MOVQ	$runtime·main·f(SB), BP		// entry
	PUSHQ	BP
	PUSHQ	$0			// arg size
	CALL	runtime·newproc(SB)
	POPQ	AX
	POPQ	AX

	// start this M
	CALL	runtime∕internal∕sched·Mstart(SB)

	MOVL	$0xf1, 0xf1  // crash
	RET

// XXX this is copied...
TEXT setg_gcc<>(SB),NOSPLIT,$0
	get_tls(AX)
	MOVQ	DI, g(AX)
	RET
