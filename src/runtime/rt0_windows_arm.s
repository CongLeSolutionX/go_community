// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// This is the entry point for the program from the
// kernel for an ordinary -buildmode=exe program.
TEXT _rt0_arm_windows(SB),NOSPLIT|NOFRAME,$0
	B	·rt0_go(SB)

TEXT _rt0_arm_windows_lib(SB),NOSPLIT|NOFRAME,$0
	MOVW	$_rt0_arm_windows_lib_go(SB), R0
	MOVW	$0, R1
	MOVW	_cgo_sys_thread_create(SB), R2
	B	(R2)

TEXT _rt0_arm_windows_lib_go(SB),NOSPLIT|NOFRAME,$0
	MOVW	$0, R0
	MOVW	$0, R1
	MOVW	$runtime·rt0_go(SB), R2
	B	(R2)
