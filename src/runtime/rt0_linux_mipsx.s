// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux
// +build mips mipsle

#include "textflag.h"

TEXT _rt0_mips_linux(SB),NOSPLIT,$0
	JMP	_main<>(SB)

TEXT _rt0_mipsle_linux(SB),NOSPLIT,$0
	JMP	_main<>(SB)

TEXT _main<>(SB),NOSPLIT|NOFRAME,$0
	// In a statically linked binary, the stack contains argc,
	// argv as argc string pointers followed by a NULL, envv as a
	// sequence of string pointers followed by a NULL, and auxv.
	// There is no TLS base pointer.
	MOVW	0(R29), R4	// argc
	ADD	$4, R29, R5	// argv
	JMP	main(SB)

TEXT main(SB),NOSPLIT|NOFRAME,$0
        // In external linking, libc jumps to main with argc in R4, argv in R5
	MOVW	$runtime·rt0_go(SB), R1
	JMP	(R1)

// When building with -buildmode=c-shared, this symbol is called when the shared
// library is loaded.
TEXT _rt0_mips_linux_lib(SB),NOSPLIT,$-4
#ifdef GOBUILDMODE_shared
	CPLOAD	R25, RSB
#endif
	MOVW $_rt0_lib<>(SB), R1
	JMP (R1)

TEXT _rt0_mipsle_linux_lib(SB),NOSPLIT,$-4
#ifdef GOBUILDMODE_shared
	CPLOAD	R25, RSB
#endif
	MOVW $_rt0_lib<>(SB), R1
	JMP (R1)

TEXT _rt0_lib<>(SB),NOSPLIT,$84

	// Preserve callee-save registers.
	MOVW	R16, (4)(R29)
	MOVW	R17, (4+4*1)(R29)
	MOVW	R18, (4+4*2)(R29)
	MOVW	R19, (4+4*3)(R29)
	MOVW	R20, (4+4*4)(R29)
	MOVW	R21, (4+4*5)(R29)
	MOVW	R22, (4+4*6)(R29)
	MOVW	R23, (4+4*7)(R29)
	MOVW	g, (4+4*8)(R29)

	MOVD	F20, (40)(R29)
	MOVD	F22, (40+8*1)(R29)
	MOVD	F24, (40+8*2)(R29)
	MOVD	F26, (40+8*3)(R29)
	MOVD	F28, (40+8*4)(R29)
	MOVD	F30, (40+8*5)(R29)

	// Save argc/argv.
	MOVW	R4, _rt0_lib_argc<>(SB)
	MOVW	R5, _rt0_lib_argv<>(SB)

	// Synchronous initialization.
	MOVW	$runtime·libpreinit(SB), R2
	JAL	(R2)

	// Create a new thread to do the runtime initialization.
	// We setup call frame to _cgo_sys_thread_create following O32 ABI
	MOVW	_cgo_sys_thread_create(SB), R25
	ADDU	$-16, R29
	MOVW	$_rt0_lib_go<>(SB), R4
	MOVW	R0, R5
	JAL	(R25)
	ADDU	$16, R29

	MOVW	(4)(R29), R16
	MOVW	(4+4*1)(R29), R17
	MOVW	(4+4*2)(R29), R18
	MOVW	(4+4*3)(R29), R19
	MOVW	(4+4*4)(R29), R20
	MOVW	(4+4*5)(R29), R21
	MOVW	(4+4*6)(R29), R22
	MOVW	(4+4*7)(R29), R23
	MOVW	(4+4*8)(R29), g

	MOVD	(40)(R29), F20
	MOVD	(40+8*1)(R29), F22
	MOVD	(40+8*2)(R29), F24
	MOVD	(40+8*3)(R29), F26
	MOVD	(40+8*4)(R29), F28
	MOVD	(40+8*5)(R29), F30
	RET

TEXT _rt0_lib_go<>(SB),NOSPLIT,$-4
	// This is called from external pthread code so we need to setup RSB on entry
#ifdef GOBUILDMODE_shared
	CPLOAD	R25, RSB
#endif	
	MOVW	_rt0_lib_argc<>(SB), R4
	MOVW	_rt0_lib_argv<>(SB), R5
	JMP	runtime·rt0_go(SB)

DATA _rt0_lib_argc<>(SB)/4,$0
GLOBL _rt0_lib_argc<>(SB),NOPTR,$4
DATA _rt0_lib_argv<>(SB)/4,$0
GLOBL _rt0_lib_argv<>(SB),NOPTR,$4
